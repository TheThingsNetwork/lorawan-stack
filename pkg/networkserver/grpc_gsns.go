// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package networkserver

import (
	"bytes"
	"context"
	"fmt"
	"hash"
	"math"
	"sort"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/random"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// recentUplinkCount is the maximum amount of recent uplinks stored per device.
	recentUplinkCount = 20

	// accumulationCapacity is the initial capacity of the accumulator.
	accumulationCapacity = 20

	// retransmissionWindow is the maximum delay between Rx2 end and an uplink retransmission.
	retransmissionWindow = 10 * time.Second

	// maxConfNbTrans is the maximum number of confirmed uplink retransmissions for pre-1.0.3 devices.
	maxConfNbTrans = 5
)

func (ns *NetworkServer) deduplicateUplink(ctx context.Context, up *ttnpb.UplinkMessage) (*metadataAccumulator, func(), bool) {
	h := ns.hashPool.Get().(hash.Hash64)
	_, _ = h.Write(up.RawPayload)

	k := h.Sum64()

	h.Reset()
	ns.hashPool.Put(h)

	acc := ns.metadataAccumulatorPool.Get().(*metadataAccumulator)
	lv, isDup := ns.metadataAccumulators.LoadOrStore(k, acc)
	lv.(*metadataAccumulator).Add(up.RxMetadata...)

	if isDup {
		ns.metadataAccumulatorPool.Put(acc)
		return nil, nil, true
	}
	return acc, func() {
		ns.metadataAccumulators.Delete(k)
	}, false
}

func resetsFCnt(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) bool {
	if dev.MACSettings != nil && dev.MACSettings.ResetsFCnt != nil {
		return dev.MACSettings.ResetsFCnt.Value
	}
	if defaults.ResetsFCnt != nil {
		return defaults.ResetsFCnt.Value
	}
	return false
}

// transmissionNumber returns the number of the transmission up would represent if appended to ups
// and the time of the last transmission of macPayload in ups, if such is found.
func transmissionNumber(macPayload []byte, ups ...*ttnpb.UplinkMessage) (uint32, time.Time) {
	nb := uint32(1)
	var lastTrans time.Time
	for i := len(ups) - 1; i >= 0; i-- {
		up := ups[i]
		if len(up.RawPayload) < 4 || !bytes.Equal(macPayload, up.RawPayload[:len(up.RawPayload)-4]) {
			break
		}
		nb++
		if up.ReceivedAt.After(lastTrans) {
			lastTrans = up.ReceivedAt
		}
	}
	return nb, lastTrans
}

func maxTransmissionNumber(ver ttnpb.MACVersion, confirmed bool, nbTrans uint32) uint32 {
	if !confirmed {
		return nbTrans
	}
	if ver.Compare(ttnpb.MAC_V1_0_3) < 0 {
		return maxConfNbTrans
	}
	return nbTrans
}

func maxRetransmissionDelay(rxDelay ttnpb.RxDelay) time.Duration {
	return rxDelay.Duration() + time.Second + retransmissionWindow
}

func fCntResetGap(last, recv uint32) uint32 {
	if math.MaxUint32-last < recv {
		return last + recv
	} else {
		return math.MaxUint32
	}
}

type macHandler func(context.Context, *ttnpb.EndDevice, *ttnpb.UplinkMessage) ([]events.DefinitionDataClosure, error)

func makeDeferredMACHandler(dev *ttnpb.EndDevice, f macHandler) macHandler {
	queuedLength := len(dev.MACState.QueuedResponses)
	return func(ctx context.Context, dev *ttnpb.EndDevice, up *ttnpb.UplinkMessage) ([]events.DefinitionDataClosure, error) {
		switch n := len(dev.MACState.QueuedResponses); {
		case n < queuedLength:
			return nil, errCorruptedMACState
		case n == queuedLength:
			return f(ctx, dev, up)
		default:
			tail := append(dev.MACState.QueuedResponses[queuedLength:0:0], dev.MACState.QueuedResponses[queuedLength:]...)
			dev.MACState.QueuedResponses = dev.MACState.QueuedResponses[:queuedLength]
			evs, err := f(ctx, dev, up)
			dev.MACState.QueuedResponses = append(dev.MACState.QueuedResponses, tail...)
			return evs, err
		}
	}
}

type matchedDevice struct {
	logger log.Interface
	phy    band.Band

	ChannelIndex             uint8
	DataRateIndex            ttnpb.DataRateIndex
	DeferredMACHandlers      []macHandler
	Device                   *ttnpb.EndDevice
	FCnt                     uint32
	FCntReset                bool
	NbTrans                  uint32
	Pending                  bool
	QueuedApplicationUplinks []*ttnpb.ApplicationUp
	QueuedEvents             []events.DefinitionDataClosure
	SetPaths                 []string
}

func (d *matchedDevice) deferMACHandler(f macHandler) {
	d.DeferredMACHandlers = append(d.DeferredMACHandlers, makeDeferredMACHandler(d.Device, f))
}

// matchAndHandleDataUplink tries to match the data uplink message with a device and returns the matched device.
func (ns *NetworkServer) matchAndHandleDataUplink(ctx context.Context, up *ttnpb.UplinkMessage, deduplicated bool, devs ...*ttnpb.EndDevice) (*matchedDevice, error) {
	if len(up.RawPayload) < 4 {
		return nil, errRawPayloadTooShort
	}
	macPayloadBytes := up.RawPayload[:len(up.RawPayload)-4]

	pld := up.Payload.GetMACPayload()

	logger := log.FromContext(ctx).WithFields(log.Fields(
		"dev_addr", pld.DevAddr,
		"uplink_f_cnt", pld.FCnt,
		"payload_length", len(macPayloadBytes),
	))

	type device struct {
		matchedDevice
		band                       band.Band
		gap                        uint32
		pendingApplicationDownlink *ttnpb.ApplicationDownlink
	}
	matches := make([]device, 0, len(devs))
	for _, dev := range devs {
		if dev.Multicast {
			continue
		}

		_, phy, err := getDeviceBandVersion(dev, ns.FrequencyPlans)
		if err != nil {
			logger.WithError(err).Warn("Failed to get device's versioned band, skip")
			continue
		}

		drIdx, err := searchDataRate(up.Settings.DataRate, dev, ns.FrequencyPlans)
		if err != nil {
			logger.WithError(err).Debug("Failed to determine data rate index of uplink, skip")
			continue
		}

		logger := logger.WithField("device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers))

		pendingApplicationDownlink := dev.GetMACState().GetPendingApplicationDownlink()

		if !pld.Ack && dev.PendingSession != nil && dev.PendingMACState != nil && dev.PendingSession.DevAddr == pld.DevAddr {
			logger := logger.WithFields(log.Fields(
				"mac_version", dev.PendingMACState.LoRaWANVersion,
				"pending_session", true,
				"f_cnt_gap", pld.FCnt,
				"full_f_cnt_up", pld.FCnt,
				"transmission", 1,
			))

			pendingDev := dev
			if dev.Session != nil && dev.MACState != nil && dev.Session.DevAddr == pld.DevAddr {
				logger.Error("Same DevAddr was assigned to a device in two consecutive sessions")
				pendingDev = copyEndDevice(dev)
			}
			pendingDev.MACState = pendingDev.PendingMACState
			pendingDev.PendingMACState = nil

			matches = append(matches, device{
				matchedDevice: matchedDevice{
					logger:        logger,
					phy:           phy,
					DataRateIndex: drIdx,
					Device:        pendingDev,
					FCnt:          pld.FCnt,
					NbTrans:       1,
					Pending:       true,
				},
				band:                       phy,
				gap:                        pld.FCnt,
				pendingApplicationDownlink: pendingApplicationDownlink,
			})
		}

		if dev.Session == nil || dev.MACState == nil || dev.Session.DevAddr != pld.DevAddr {
			continue
		}
		if pld.Ack && len(dev.MACState.RecentDownlinks) == 0 {
			logger.Debug("Uplink contains ACK, but no downlink was sent to device, skip")
			continue
		}

		supports32BitFCnt := true
		if dev.GetMACSettings().GetSupports32BitFCnt() != nil {
			supports32BitFCnt = dev.MACSettings.Supports32BitFCnt.Value
		} else if ns.defaultMACSettings.GetSupports32BitFCnt() != nil {
			supports32BitFCnt = ns.defaultMACSettings.Supports32BitFCnt.Value
		}

		fCnt := pld.FCnt
		switch {
		case !supports32BitFCnt, fCnt >= dev.Session.LastFCntUp, fCnt == 0:
		case fCnt > dev.Session.LastFCntUp&0xffff:
			fCnt |= dev.Session.LastFCntUp &^ 0xffff
		case dev.Session.LastFCntUp < 0xffff0000:
			fCnt |= (dev.Session.LastFCntUp + 0x10000) &^ 0xffff
		}

		maxNbTrans := maxTransmissionNumber(dev.MACState.LoRaWANVersion, up.Payload.MType == ttnpb.MType_CONFIRMED_UP, dev.MACState.CurrentParameters.ADRNbTrans)
		logger = logger.WithFields(log.Fields(
			"last_f_cnt_up", dev.Session.LastFCntUp,
			"mac_version", dev.MACState.LoRaWANVersion,
			"max_transmissions", maxNbTrans,
			"pending_session", false,
			"supports_32_bit_f_cnt", true,
		))

		if fCnt == dev.Session.LastFCntUp && len(dev.MACState.RecentUplinks) > 0 {
			nbTrans, lastAt := transmissionNumber(macPayloadBytes, dev.MACState.RecentUplinks...)
			logger = logger.WithFields(log.Fields(
				"f_cnt_gap", 0,
				"f_cnt_reset", false,
				"full_f_cnt_up", dev.Session.LastFCntUp,
				"transmission", nbTrans,
			))
			if nbTrans < 2 || lastAt.IsZero() {
				logger.Debug("Repeated FCnt value, but frame is not a retransmission, skip")
				continue
			}

			maxDelay := maxRetransmissionDelay(dev.MACState.CurrentParameters.Rx1Delay)
			delay := up.ReceivedAt.Sub(lastAt)

			logger = logger.WithFields(log.Fields(
				"last_transmission_at", lastAt,
				"max_retransmission_delay", maxDelay,
				"retransmission_delay", delay,
			))

			if delay > maxDelay {
				logger.Warn("Retransmission delay exceeds maximum, skip")
				continue
			}
			if nbTrans > maxNbTrans {
				logger.Warn("Transmission number exceeds maximum, skip")
				continue
			}
			matches = append(matches, device{
				matchedDevice: matchedDevice{
					logger:        logger,
					phy:           phy,
					DataRateIndex: drIdx,
					Device:        dev,
					FCnt:          dev.Session.LastFCntUp,
					NbTrans:       nbTrans,
				},
				band:                       phy,
				pendingApplicationDownlink: pendingApplicationDownlink,
			})
			continue
		}

		if fCnt < dev.Session.LastFCntUp {
			if !resetsFCnt(dev, ns.defaultMACSettings) {
				logger.Debug("FCnt too low, skip")
				continue
			}

			macState, err := newMACState(dev, ns.FrequencyPlans, ns.defaultMACSettings)
			if err != nil {
				logger.WithError(err).Warn("Failed to generate new MAC state")
				continue
			}
			if macState.LoRaWANVersion.HasMaxFCntGap() && uint(pld.FCnt) > phy.MaxFCntGap {
				continue
			}
			dev.MACState = macState

			gap := fCntResetGap(dev.Session.LastFCntUp, pld.FCnt)
			matches = append(matches, device{
				matchedDevice: matchedDevice{
					logger: logger.WithFields(log.Fields(
						"f_cnt_gap", gap,
						"f_cnt_reset", true,
						"full_f_cnt_up", pld.FCnt,
						"transmission", 1,
					)),
					phy:           phy,
					DataRateIndex: drIdx,
					Device:        dev,
					FCnt:          pld.FCnt,
					FCntReset:     true,
					NbTrans:       1,
				},
				band:                       phy,
				gap:                        gap,
				pendingApplicationDownlink: pendingApplicationDownlink,
			})
			continue
		}

		logger = logger.WithField("transmission", 1)

		if fCnt != pld.FCnt && resetsFCnt(dev, ns.defaultMACSettings) {
			macState, err := newMACState(dev, ns.FrequencyPlans, ns.defaultMACSettings)
			if err != nil {
				logger.WithError(err).Warn("Failed to generate new MAC state")
				continue
			}
			if !macState.LoRaWANVersion.HasMaxFCntGap() || uint(pld.FCnt) <= phy.MaxFCntGap {
				dev := copyEndDevice(dev)
				dev.MACState = macState

				gap := fCntResetGap(dev.Session.LastFCntUp, pld.FCnt)
				matches = append(matches, device{
					matchedDevice: matchedDevice{
						logger: logger.WithFields(log.Fields(
							"f_cnt_gap", gap,
							"f_cnt_reset", true,
							"full_f_cnt_up", pld.FCnt,
						)),
						phy:           phy,
						DataRateIndex: drIdx,
						Device:        dev,
						FCnt:          pld.FCnt,
						FCntReset:     true,
						NbTrans:       1,
					},
					band:                       phy,
					gap:                        gap,
					pendingApplicationDownlink: pendingApplicationDownlink,
				})
			}
		}

		gap := fCnt - dev.Session.LastFCntUp
		logger = logger.WithFields(log.Fields(
			"f_cnt_gap", gap,
			"f_cnt_reset", false,
			"full_f_cnt_up", fCnt,
		))

		if fCnt == math.MaxUint32 {
			logger.Debug("FCnt too high, skip")
			continue
		}
		if dev.MACState.LoRaWANVersion.HasMaxFCntGap() && uint(gap) > phy.MaxFCntGap {
			logger.Debug("FCnt gap too high, skip")
			continue
		}
		matches = append(matches, device{
			matchedDevice: matchedDevice{
				logger:        logger,
				phy:           phy,
				DataRateIndex: drIdx,
				Device:        dev,
				FCnt:          fCnt,
				NbTrans:       1,
			},
			band:                       phy,
			gap:                        gap,
			pendingApplicationDownlink: pendingApplicationDownlink,
		})
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].gap != matches[j].gap {
			return matches[i].gap < matches[j].gap
		}
		if matches[i].FCntReset != matches[j].FCntReset {
			return matches[j].FCntReset
		}
		return matches[i].FCnt < matches[j].FCnt
	})

matchLoop:
	for i, match := range matches {
		logger := match.logger.WithField("match_attempt", i)

		session := match.Device.Session
		if match.Pending {
			session = match.Device.PendingSession
		}

		if session.FNwkSIntKey == nil || len(session.FNwkSIntKey.Key) == 0 {
			logger.Warn("Device missing FNwkSIntKey in registry, skip")
			continue
		}
		fNwkSIntKey, err := cryptoutil.UnwrapAES128Key(ctx, *session.FNwkSIntKey, ns.KeyVault)
		if err != nil {
			logger.WithField("kek_label", session.FNwkSIntKey.KEKLabel).WithError(err).Warn("Failed to unwrap FNwkSIntKey, skip")
			continue
		}

		if match.NbTrans > 1 {
			match.Device.MACState.PendingRequests = nil
		}
		macBuf := pld.FOpts
		if len(macBuf) == 0 && pld.FPort == 0 {
			macBuf = pld.FRMPayload
		}
		if len(macBuf) > 0 && (len(pld.FOpts) == 0 || match.Device.MACState.LoRaWANVersion.EncryptFOpts()) {
			if session.NwkSEncKey == nil || len(session.NwkSEncKey.Key) == 0 {
				logger.Warn("Device missing NwkSEncKey in registry, skip")
				continue
			}
			key, err := cryptoutil.UnwrapAES128Key(ctx, *session.NwkSEncKey, ns.KeyVault)
			if err != nil {
				logger.WithField("kek_label", session.NwkSEncKey.KEKLabel).WithError(err).Warn("Failed to unwrap NwkSEncKey, skip")
				continue
			}
			macBuf, err = crypto.DecryptUplink(key, pld.DevAddr, pld.FCnt, macBuf)
			if err != nil {
				logger.WithError(err).Warn("Failed to decrypt uplink, skip")
				continue
			}
		}

		var cmds []*ttnpb.MACCommand
		for r := bytes.NewReader(macBuf); r.Len() > 0; {
			cmd := &ttnpb.MACCommand{}
			if err := lorawan.DefaultMACCommands.ReadUplink(match.band, r, cmd); err != nil {
				logger.WithFields(log.Fields(
					"bytes_left", r.Len(),
					"mac_count", len(cmds),
				)).WithError(err).Warn("Failed to read MAC command")
				break
			}
			logger := logger.WithField("cid", cmd.CID)
			logger.Debug("Read MAC command")
			def, ok := lorawan.DefaultMACCommands[cmd.CID]
			switch {
			case ok && !def.InitiatedByDevice && (match.Pending || match.FCntReset):
				logger.Debug("Received MAC command answer after MAC state reset, skip")
				continue matchLoop
			case ok && match.NbTrans > 1 && !lorawan.DefaultMACCommands[cmd.CID].InitiatedByDevice:
				logger.Debug("Skip processing of MAC command not initiated by the device in a retransmission")
				continue
			}
			cmds = append(cmds, cmd)
		}
		logger = logger.WithField("mac_count", len(cmds))
		ctx = log.NewContext(ctx, logger)

		match.Device.MACState.QueuedResponses = match.Device.MACState.QueuedResponses[:0]
	macLoop:
		for len(cmds) > 0 {
			var cmd *ttnpb.MACCommand
			cmd, cmds = cmds[0], cmds[1:]
			logger := logger.WithField("cid", cmd.CID)
			ctx := log.NewContext(ctx, logger)

			logger.Debug("Handle MAC command")

			var evs []events.DefinitionDataClosure
			var err error
			switch cmd.CID {
			case ttnpb.CID_RESET:
				evs, err = handleResetInd(ctx, match.Device, cmd.GetResetInd(), ns.FrequencyPlans, ns.defaultMACSettings)
			case ttnpb.CID_LINK_CHECK:
				if !deduplicated {
					match.deferMACHandler(handleLinkCheckReq)
					continue macLoop
				}
				evs, err = handleLinkCheckReq(ctx, match.Device, up)
			case ttnpb.CID_LINK_ADR:
				pld := cmd.GetLinkADRAns()
				dupCount := 0
				if match.Device.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_0_2) >= 0 && match.Device.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
					logger.Debug("Count LinkADR duplicates")
					for _, dup := range cmds {
						if dup.CID != ttnpb.CID_LINK_ADR {
							break
						}
						if *dup.GetLinkADRAns() != *pld {
							err = errInvalidPayload
							break
						}
						dupCount++
					}
					logger.WithField("duplicate_count", dupCount).Debug("Counted LinkADR duplicates")
				}
				if err != nil {
					break
				}
				cmds = cmds[dupCount:]
				evs, err = handleLinkADRAns(ctx, match.Device, pld, uint(dupCount), ns.FrequencyPlans)
			case ttnpb.CID_DUTY_CYCLE:
				evs, err = handleDutyCycleAns(ctx, match.Device)
			case ttnpb.CID_RX_PARAM_SETUP:
				evs, err = handleRxParamSetupAns(ctx, match.Device, cmd.GetRxParamSetupAns())
			case ttnpb.CID_DEV_STATUS:
				evs, err = handleDevStatusAns(ctx, match.Device, cmd.GetDevStatusAns(), session.LastFCntUp, up.ReceivedAt)
				if err == nil {
					match.SetPaths = append(match.SetPaths,
						"battery_percentage",
						"downlink_margin",
						"last_dev_status_received_at",
						"power_state",
					)
				}
			case ttnpb.CID_NEW_CHANNEL:
				evs, err = handleNewChannelAns(ctx, match.Device, cmd.GetNewChannelAns())
			case ttnpb.CID_RX_TIMING_SETUP:
				evs, err = handleRxTimingSetupAns(ctx, match.Device)
			case ttnpb.CID_TX_PARAM_SETUP:
				evs, err = handleTxParamSetupAns(ctx, match.Device)
			case ttnpb.CID_DL_CHANNEL:
				evs, err = handleDLChannelAns(ctx, match.Device, cmd.GetDLChannelAns())
			case ttnpb.CID_REKEY:
				evs, err = handleRekeyInd(ctx, match.Device, cmd.GetRekeyInd())
			case ttnpb.CID_ADR_PARAM_SETUP:
				evs, err = handleADRParamSetupAns(ctx, match.Device)
			case ttnpb.CID_DEVICE_TIME:
				if !deduplicated {
					match.deferMACHandler(handleDeviceTimeReq)
					continue macLoop
				}
				evs, err = handleDeviceTimeReq(ctx, match.Device, up)
			case ttnpb.CID_REJOIN_PARAM_SETUP:
				evs, err = handleRejoinParamSetupAns(ctx, match.Device, cmd.GetRejoinParamSetupAns())
			case ttnpb.CID_PING_SLOT_INFO:
				evs, err = handlePingSlotInfoReq(ctx, match.Device, cmd.GetPingSlotInfoReq())
			case ttnpb.CID_PING_SLOT_CHANNEL:
				evs, err = handlePingSlotChannelAns(ctx, match.Device, cmd.GetPingSlotChannelAns())
			case ttnpb.CID_BEACON_TIMING:
				evs, err = handleBeaconTimingReq(ctx, match.Device)
			case ttnpb.CID_BEACON_FREQ:
				evs, err = handleBeaconFreqAns(ctx, match.Device, cmd.GetBeaconFreqAns())
			case ttnpb.CID_DEVICE_MODE:
				evs, err = handleDeviceModeInd(ctx, match.Device, cmd.GetDeviceModeInd())
			default:
				logger.Warn("Unknown MAC command received, skip the rest")
				break macLoop
			}
			if err != nil {
				logger.WithError(err).Debug("Failed to process MAC command")
				break macLoop
			}
			match.QueuedEvents = append(match.QueuedEvents, evs...)
		}
		if n := len(match.Device.MACState.PendingRequests); n > 0 {
			logger.WithField("unanswered_request_count", n).Warn("MAC command buffer not fully answered")
			match.Device.MACState.PendingRequests = match.Device.MACState.PendingRequests[:0]
		}

		if match.Pending {
			if match.Device.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
				match.Device.Session = match.Device.PendingSession
				match.Device.EndDeviceIdentifiers.DevAddr = &pld.DevAddr
			} else if match.Device.PendingSession != nil {
				logger.Debug("No RekeyInd received for LoRaWAN 1.1+ device, skip")
				continue matchLoop
			}
			match.SetPaths = append(match.SetPaths, "ids.dev_addr")
		}

		chIdx, err := searchUplinkChannel(up.Settings.Frequency, match.Device.MACState)
		if err != nil {
			logger.WithError(err).Debug("Failed to determine channel index of uplink, skip")
			continue
		}
		match.ChannelIndex = chIdx

		var computedMIC [4]byte
		if match.Device.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
			computedMIC, err = crypto.ComputeLegacyUplinkMIC(
				fNwkSIntKey,
				pld.DevAddr,
				match.FCnt,
				macPayloadBytes,
			)
		} else {
			if match.Device.Session.SNwkSIntKey == nil || len(match.Device.Session.SNwkSIntKey.Key) == 0 {
				logger.Warn("Device missing SNwkSIntKey in registry, skip")
				continue
			}

			var sNwkSIntKey types.AES128Key
			sNwkSIntKey, err = cryptoutil.UnwrapAES128Key(ctx, *match.Device.Session.SNwkSIntKey, ns.KeyVault)
			if err != nil {
				logger.WithField("kek_label", match.Device.Session.SNwkSIntKey.KEKLabel).WithError(err).Warn("Failed to unwrap SNwkSIntKey, skip")
				continue
			}

			var confFCnt uint32
			if pld.Ack {
				confFCnt = match.Device.Session.LastConfFCntDown
			}
			computedMIC, err = crypto.ComputeUplinkMIC(
				sNwkSIntKey,
				fNwkSIntKey,
				confFCnt,
				uint8(match.DataRateIndex),
				chIdx,
				pld.DevAddr,
				match.FCnt,
				macPayloadBytes,
			)
		}
		if err != nil {
			logger.WithError(err).Error("Failed to compute MIC")
			continue
		}
		if !bytes.Equal(up.Payload.MIC, computedMIC[:]) {
			logger.Debug("MIC mismatch")
			continue
		}

		if match.pendingApplicationDownlink != nil {
			asUp := &ttnpb.ApplicationUp{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DevAddr:                &pld.DevAddr,
					JoinEUI:                match.Device.JoinEUI,
					DevEUI:                 match.Device.DevEUI,
					ApplicationIdentifiers: match.Device.ApplicationIdentifiers,
					DeviceID:               match.Device.DeviceID,
				},
				CorrelationIDs: append(match.pendingApplicationDownlink.CorrelationIDs, up.CorrelationIDs...),
			}
			if pld.Ack && !match.Pending && !match.FCntReset && match.NbTrans == 1 {
				asUp.Up = &ttnpb.ApplicationUp_DownlinkAck{
					DownlinkAck: match.pendingApplicationDownlink,
				}
			} else {
				asUp.Up = &ttnpb.ApplicationUp_DownlinkNack{
					DownlinkNack: match.pendingApplicationDownlink,
				}
			}
			match.QueuedApplicationUplinks = append(match.QueuedApplicationUplinks, asUp)
		}
		if !match.Pending && match.Device.PendingSession != nil {
			// TODO: Notify AS of session recovery(https://github.com/TheThingsNetwork/lorawan-stack/issues/594)
		}
		if match.Pending || match.FCntReset {
			match.Device.Session.StartedAt = up.ReceivedAt
		}
		match.Device.MACState.PendingApplicationDownlink = nil
		match.Device.MACState.PendingJoinRequest = nil
		match.Device.MACState.RxWindowsAvailable = true
		match.Device.PendingMACState = nil
		match.Device.PendingSession = nil
		match.Device.Session.LastFCntUp = match.FCnt
		match.SetPaths = append(match.SetPaths,
			"mac_state",
			"pending_mac_state",
			"pending_session",
			"session",
		)
		return &match.matchedDevice, nil
	}
	return nil, errDeviceNotFound
}

// MACHandler defines the behavior of a MAC command on a device.
type MACHandler func(ctx context.Context, dev *ttnpb.EndDevice, pld []byte, up *ttnpb.UplinkMessage) error

func appendRecentUplink(recent []*ttnpb.UplinkMessage, up *ttnpb.UplinkMessage, window int) []*ttnpb.UplinkMessage {
	recent = append(recent, up)
	if len(recent) > window {
		recent = recent[len(recent)-window:]
	}
	return recent
}

var handleDataUplinkGetPaths = [...]string{
	"frequency_plan_id",
	"last_dev_status_received_at",
	"lorawan_phy_version",
	"lorawan_version",
	"mac_settings",
	"mac_state",
	"multicast",
	"pending_mac_state",
	"pending_session",
	"queued_application_downlinks",
	"recent_downlinks",
	"recent_uplinks",
	"session",
	"supports_class_b",
	"supports_class_c",
	"supports_join",
}

func (ns *NetworkServer) handleDataUplink(ctx context.Context, up *ttnpb.UplinkMessage, acc *metadataAccumulator) (err error) {
	pld := up.Payload.GetMACPayload()

	logger := log.FromContext(ctx).WithFields(log.Fields(
		"ack", pld.Ack,
		"adr", pld.ADR,
		"adr_ack_req", pld.ADRAckReq,
		"class_b", pld.ClassB,
		"dev_addr", pld.DevAddr,
		"f_opts_len", len(pld.FOpts),
		"f_port", pld.FPort,
		"frm_payload_len", len(pld.FRMPayload),
		"uplink_f_cnt", pld.FCnt,
	))
	ctx = log.NewContext(ctx, logger)

	logger.Debug("Match device")

	var addrMatches []*ttnpb.EndDevice
	if err := ns.devices.RangeByAddr(ctx, pld.DevAddr, handleDataUplinkGetPaths[:],
		func(dev *ttnpb.EndDevice) bool {
			addrMatches = append(addrMatches, dev)
			return true
		}); err != nil {
		logger.WithError(err).Warn("Failed to find devices in registry by DevAddr")
		return err
	}

	matched, err := ns.matchAndHandleDataUplink(ctx, up, false, addrMatches...)
	if err != nil {
		registerDropDataUplink(ctx, up, err)
		logger.WithError(err).Debug("Failed to match device")
		return err
	}

	logger = matched.logger
	ctx = log.NewContext(ctx, matched.logger)

	logger.Debug("Matched device")

	var queuedApplicationUplinks []*ttnpb.ApplicationUp
	var queuedEvents []events.DefinitionDataClosure

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ns.deduplicationDone(ctx, up):
	}

	up.RxMetadata = acc.Accumulated()
	logger = logger.WithField("metadata_count", len(up.RxMetadata))
	logger.Debug("Merged metadata")
	ctx = log.NewContext(ctx, logger)
	queuedEvents = append(queuedEvents, evtMergeMetadata.BindData(len(up.RxMetadata)))
	registerMergeMetadata(ctx, up)

	for _, f := range matched.DeferredMACHandlers {
		evs, err := f(ctx, matched.Device, up)
		if err != nil {
			logger.WithError(err).Warn("Failed to process MAC command after deduplication")
			break
		}
		matched.QueuedEvents = append(matched.QueuedEvents, evs...)
	}

	var handleErr bool
	stored, err := ns.devices.SetByID(ctx, matched.Device.ApplicationIdentifiers, matched.Device.DeviceID, handleDataUplinkGetPaths[:],
		func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if stored == nil {
				logger.Warn("Device deleted during uplink handling, drop")
				handleErr = true
				return nil, nil, errOutdatedData
			}

			if !stored.CreatedAt.Equal(matched.Device.CreatedAt) || !stored.UpdatedAt.Equal(matched.Device.UpdatedAt) {
				matched, err = ns.matchAndHandleDataUplink(ctx, up, true, stored)
				if err != nil {
					handleErr = true
					return nil, nil, errOutdatedData.WithCause(err)
				}
			}
			queuedEvents = append(queuedEvents, matched.QueuedEvents...)
			queuedApplicationUplinks = append(queuedApplicationUplinks, matched.QueuedApplicationUplinks...)

			up.DeviceChannelIndex = uint32(matched.ChannelIndex)
			up.Settings.DataRateIndex = matched.DataRateIndex

			stored = matched.Device
			paths := matched.SetPaths

			stored.MACState.RecentUplinks = appendRecentUplink(stored.MACState.RecentUplinks, up, recentUplinkCount)
			paths = ttnpb.AddFields(paths, "mac_state.recent_uplinks")

			stored.RecentUplinks = appendRecentUplink(stored.RecentUplinks, up, recentUplinkCount)
			paths = ttnpb.AddFields(paths, "recent_uplinks")

			paths = ttnpb.AddFields(paths, "recent_adr_uplinks")
			if !pld.FHDR.ADR {
				stored.RecentADRUplinks = nil
				return stored, paths, nil
			}
			stored.RecentADRUplinks = appendRecentUplink(stored.RecentADRUplinks, up, optimalADRUplinkCount)

			if !deviceUseADR(stored, ns.defaultMACSettings) {
				return stored, paths, nil
			}

			if err := adaptDataRate(stored, ns.FrequencyPlans, ns.defaultMACSettings); err != nil {
				handleErr = true
				return nil, nil, err
			}
			return stored, paths, nil
		})
	if err != nil && !handleErr {
		logger.WithError(err).Warn("Failed to update device in registry")
		// TODO: Retry transaction. (https://github.com/TheThingsNetwork/lorawan-stack/issues/33)
	}
	if err != nil {
		events.Publish(evtDropDataUplink(ctx, matched.Device.EndDeviceIdentifiers, err))
		registerDropDataUplink(ctx, up, err)
		return err
	}

	downAt, ok := nextDataDownlinkAt(ctx, stored, matched.phy, ns.defaultMACSettings)
	if !ok {
		logger.Debug("No downlink to send or windows expired, avoid adding downlink task after data uplink")
	} else {
		downAt = downAt.Add(-nsScheduleWindow)
		logger.WithField("start_at", downAt).Debug("Add downlink task after data uplink")
		if err := ns.downlinkTasks.Add(ctx, stored.EndDeviceIdentifiers, downAt, true); err != nil {
			logger.WithError(err).Error("Failed to add downlink task after data uplink")
		}
	}

	if matched.NbTrans == 1 {
		queuedApplicationUplinks = append(queuedApplicationUplinks, &ttnpb.ApplicationUp{
			EndDeviceIdentifiers: stored.EndDeviceIdentifiers,
			CorrelationIDs:       up.CorrelationIDs,
			Up: &ttnpb.ApplicationUp_UplinkMessage{UplinkMessage: &ttnpb.ApplicationUplink{
				FCnt:         stored.Session.LastFCntUp,
				FPort:        pld.FPort,
				FRMPayload:   pld.FRMPayload,
				RxMetadata:   up.RxMetadata,
				SessionKeyID: stored.Session.SessionKeyID,
				Settings:     up.Settings,
				ReceivedAt:   up.ReceivedAt,
			}},
		})
		queuedEvents = append(queuedEvents, evtForwardDataUplink.BindData(nil))
		registerForwardDataUplink(ctx, up)
	}

	if len(queuedApplicationUplinks) > 0 {
		if err := ns.applicationUplinks.Add(ctx, queuedApplicationUplinks...); err != nil {
			logger.WithError(err).Warn("Failed to queue application uplinks for sending to Application Server")
		}
	}
	if len(queuedEvents) > 0 {
		for _, ev := range queuedEvents {
			events.Publish(ev(ctx, stored.EndDeviceIdentifiers))
		}
	}
	return nil
}

// newDevAddr generates a DevAddr for specified EndDevice.
func (ns *NetworkServer) newDevAddr(context.Context, *ttnpb.EndDevice) types.DevAddr {
	var devAddr types.DevAddr
	random.Read(devAddr[:])
	prefix := ns.devAddrPrefixes[random.Intn(len(ns.devAddrPrefixes))]
	return devAddr.WithPrefix(prefix)
}

func (ns *NetworkServer) sendJoinRequest(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, req *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
	logger := log.FromContext(ctx)
	cc, err := ns.GetPeerConn(ctx, ttnpb.ClusterRole_JOIN_SERVER, ids)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.WithError(err).Debug("Join Server peer not found")
		} else {
			logger.WithError(err).Error("Join Server peer connection lookup failed")
		}
	} else {
		resp, err := ttnpb.NewNsJsClient(cc).HandleJoin(ctx, req, ns.WithClusterAuth())
		if err == nil {
			logger.Debug("Join-request accepted by cluster-local Join Server")
			return resp, nil
		}
		logger.WithError(err).Info("Cluster-local Join Server did not accept join-request")
		if !errors.IsNotFound(err) {
			return nil, err
		}
	}
	if ns.interopClient != nil {
		resp, err := ns.interopClient.HandleJoinRequest(ctx, ns.netID, req)
		if err == nil {
			logger.Debug("Join-request accepted by interop Join Server")
			return resp, nil
		}
		logger.WithError(err).Warn("Interop Join Server did not accept join-request")
		if !errors.IsNotFound(err) {
			return nil, err
		}
	}
	return nil, errJoinServerNotFound
}

func (ns *NetworkServer) handleJoinRequest(ctx context.Context, up *ttnpb.UplinkMessage, acc *metadataAccumulator) (err error) {
	pld := up.Payload.GetJoinRequestPayload()

	logger := log.FromContext(ctx).WithFields(log.Fields(
		"dev_eui", pld.DevEUI,
		"join_eui", pld.JoinEUI,
	))
	ctx = log.NewContext(ctx, logger)

	dev, err := ns.devices.GetByEUI(ctx, pld.JoinEUI, pld.DevEUI,
		[]string{
			"frequency_plan_id",
			"lorawan_phy_version",
			"lorawan_version",
			"mac_settings",
			"session",
			"supports_class_b",
			"supports_class_c",
			"supports_join",
		},
	)
	if err != nil {
		registerDropJoinRequest(ctx, up, err)
		logger.WithError(err).Debug("Failed to load device from registry")
		return err
	}

	defer func(dev *ttnpb.EndDevice) {
		if err != nil {
			events.Publish(evtDropJoinRequest(ctx, dev.EndDeviceIdentifiers, err))
			registerDropJoinRequest(ctx, up, err)
		}
	}(dev)

	logger = logger.WithField("device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers))

	if !dev.SupportsJoin {
		logger.Warn("ABP device sent a join-request, drop")
		return errABPJoinRequest
	}

	ctx = log.NewContext(ctx, logger)

	devAddr := ns.newDevAddr(ctx, dev)
	for dev.Session != nil && devAddr.Equal(dev.Session.DevAddr) {
		devAddr = ns.newDevAddr(ctx, dev)
	}
	logger = logger.WithField("dev_addr", devAddr)
	ctx = log.NewContext(ctx, logger)

	macState, err := newMACState(dev, ns.FrequencyPlans, ns.defaultMACSettings)
	if err != nil {
		logger.WithError(err).Warn("Failed to reset device's MAC state")
		return err
	}

	fp, phy, err := getDeviceBandVersion(dev, ns.FrequencyPlans)
	if err != nil {
		return err
	}

	req := &ttnpb.JoinRequest{
		Payload:            up.Payload,
		CFList:             frequencyplans.CFList(*fp, dev.LoRaWANPHYVersion),
		CorrelationIDs:     events.CorrelationIDsFromContext(ctx),
		DevAddr:            devAddr,
		NetID:              ns.netID,
		RawPayload:         up.RawPayload,
		RxDelay:            macState.DesiredParameters.Rx1Delay,
		SelectedMACVersion: dev.LoRaWANVersion, // Assume NS version is always higher than the version of the device
		DownlinkSettings: ttnpb.DLSettings{
			Rx1DROffset: macState.DesiredParameters.Rx1DataRateOffset,
			Rx2DR:       macState.DesiredParameters.Rx2DataRateIndex,
			OptNeg:      dev.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) >= 0,
		},
	}

	resp, err := ns.sendJoinRequest(ctx, dev.EndDeviceIdentifiers, req)
	if err != nil {
		return err
	}
	respRecvAt := timeNow()

	ctx = events.ContextWithCorrelationID(ctx, resp.CorrelationIDs...)

	keys := resp.SessionKeys
	if !req.DownlinkSettings.OptNeg {
		keys.NwkSEncKey = keys.FNwkSIntKey
		keys.SNwkSIntKey = keys.FNwkSIntKey
	}

	macState.QueuedJoinAccept = &ttnpb.MACState_JoinAccept{
		Keys:    keys,
		Payload: resp.RawPayload,
		Request: *req,
	}
	macState.RxWindowsAvailable = true

	macState.CurrentParameters.Rx1Delay = req.RxDelay
	macState.CurrentParameters.Rx1DataRateOffset = req.DownlinkSettings.Rx1DROffset
	macState.CurrentParameters.Rx2DataRateIndex = req.DownlinkSettings.Rx2DR
	if req.DownlinkSettings.OptNeg && dev.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) > 0 {
		// The version will be further negotiated via RekeyInd/RekeyConf
		macState.LoRaWANVersion = ttnpb.MAC_V1_1
	}
	if req.CFList != nil {
		switch req.CFList.Type {
		case ttnpb.CFListType_FREQUENCIES:
			for _, freq := range req.CFList.Freq {
				if freq == 0 {
					break
				}
				macState.CurrentParameters.Channels = append(macState.CurrentParameters.Channels, &ttnpb.MACParameters_Channel{
					UplinkFrequency:   uint64(freq * 100),
					DownlinkFrequency: uint64(freq * 100),
					MaxDataRateIndex:  ttnpb.DataRateIndex(phy.MaxADRDataRateIndex),
					EnableUplink:      true,
				})
			}

		case ttnpb.CFListType_CHANNEL_MASKS:
			if len(macState.CurrentParameters.Channels) != len(req.CFList.ChMasks) {
				return errCorruptedMACState
			}
			for i, m := range req.CFList.ChMasks {
				if m {
					continue
				}
				if macState.CurrentParameters.Channels[i] == nil {
					return errCorruptedMACState
				}
				macState.CurrentParameters.Channels[i].EnableUplink = m
			}
		}
	}

	events.Publish(evtForwardJoinRequest(ctx, dev.EndDeviceIdentifiers, nil))
	registerForwardJoinRequest(ctx, up)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ns.deduplicationDone(ctx, up):
	}

	up.RxMetadata = acc.Accumulated()
	events.Publish(evtMergeMetadata(ctx, dev.EndDeviceIdentifiers, len(up.RxMetadata)))
	registerMergeMetadata(ctx, up)

	var invalidatedQueue []*ttnpb.ApplicationDownlink
	dev, err = ns.devices.SetByID(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers, dev.EndDeviceIdentifiers.DeviceID,
		[]string{
			"frequency_plan_id",
			"lorawan_phy_version",
			"queued_application_downlinks",
			"recent_uplinks",
		},
		func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if stored == nil {
				logger.Warn("Device deleted during join-request handling, drop")
				return nil, nil, errOutdatedData
			}

			var paths []string

			stored.PendingMACState = macState
			paths = append(paths, "pending_mac_state")

			upChIdx, err := searchUplinkChannel(up.Settings.Frequency, macState)
			if err != nil {
				return nil, nil, err
			}
			up.DeviceChannelIndex = uint32(upChIdx)

			upDRIdx, err := searchDataRate(up.Settings.DataRate, stored, ns.Component.FrequencyPlans)
			if err != nil {
				return nil, nil, err
			}
			up.Settings.DataRateIndex = upDRIdx

			stored.RecentUplinks = appendRecentUplink(stored.RecentUplinks, up, recentUplinkCount)
			paths = append(paths, "recent_uplinks")

			invalidatedQueue = stored.QueuedApplicationDownlinks
			stored.QueuedApplicationDownlinks = nil
			paths = append(paths, "queued_application_downlinks")

			return stored, paths, nil
		})
	if err != nil {
		logger.WithError(err).Warn("Failed to update device in registry")
		// TODO: Retry transaction. (https://github.com/TheThingsNetwork/lorawan-stack/issues/33)
	}
	if err != nil {
		return err
	}

	startAt := up.ReceivedAt.Add(phy.JoinAcceptDelay1 - nsScheduleWindow)
	logger.WithField("start_at", startAt).Debug("Add downlink task for join-accept")
	if err := ns.downlinkTasks.Add(ctx, dev.EndDeviceIdentifiers, startAt, true); err != nil {
		logger.WithError(err).Error("Failed to add downlink task for join-accept")
	}
	if err := ns.applicationUplinks.Add(ctx, &ttnpb.ApplicationUp{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: dev.EndDeviceIdentifiers.ApplicationIdentifiers,
			DeviceID:               dev.EndDeviceIdentifiers.DeviceID,
			DevEUI:                 dev.EndDeviceIdentifiers.DevEUI,
			JoinEUI:                dev.EndDeviceIdentifiers.JoinEUI,
			DevAddr:                &devAddr,
		},
		CorrelationIDs: events.CorrelationIDsFromContext(ctx),
		Up: &ttnpb.ApplicationUp_JoinAccept{JoinAccept: &ttnpb.ApplicationJoinAccept{
			AppSKey:              resp.SessionKeys.AppSKey,
			InvalidatedDownlinks: invalidatedQueue,
			SessionKeyID:         resp.SessionKeys.SessionKeyID,
			ReceivedAt:           respRecvAt,
		}},
	}); err != nil {
		logger.WithError(err).Warn("Failed to queue join-accept for sending to Application Server")
	}
	return nil
}

func (ns *NetworkServer) handleRejoinRequest(ctx context.Context, up *ttnpb.UplinkMessage, acc *metadataAccumulator) (err error) {
	defer func() {
		if err != nil {
			registerDropRejoinRequest(ctx, up, err)
		}
	}()
	// TODO: Implement https://github.com/TheThingsNetwork/lorawan-stack/issues/8
	return status.Errorf(codes.Unimplemented, "not implemented")
}

// HandleUplink is called by the Gateway Server when an uplink message arrives.
func (ns *NetworkServer) HandleUplink(ctx context.Context, up *ttnpb.UplinkMessage) (*pbtypes.Empty, error) {
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}

	ctx = events.ContextWithCorrelationID(ctx, append(
		up.CorrelationIDs,
		fmt.Sprintf("ns:uplink:%s", events.NewCorrelationID()),
	)...)
	up.CorrelationIDs = events.CorrelationIDsFromContext(ctx)
	up.ReceivedAt = timeNow().UTC()
	up.Payload = &ttnpb.Message{}
	if err := lorawan.UnmarshalMessage(up.RawPayload, up.Payload); err != nil {
		return nil, errDecodePayload.WithCause(err)
	}

	if up.Payload.Major != ttnpb.Major_LORAWAN_R1 {
		return nil, errUnsupportedLoRaWANVersion.WithAttributes(
			"version", up.Payload.Major,
		)
	}

	logger := log.FromContext(ctx).WithFields(log.Fields(
		"m_type", up.Payload.MType,
		"major", up.Payload.Major,
		"received_at", up.ReceivedAt,
	))
	ctx = log.NewContext(ctx, logger)

	var handle func(context.Context, *ttnpb.UplinkMessage, *metadataAccumulator) error
	switch up.Payload.MType {
	case ttnpb.MType_CONFIRMED_UP, ttnpb.MType_UNCONFIRMED_UP:
		handle = ns.handleDataUplink
	case ttnpb.MType_JOIN_REQUEST:
		handle = ns.handleJoinRequest
	case ttnpb.MType_REJOIN_REQUEST:
		handle = ns.handleRejoinRequest
	default:
		logger.Debug("Unmatched MType")
		return ttnpb.Empty, nil
	}

	logger.Debug("Deduplicate uplink")
	acc, stopDedup, ok := ns.deduplicateUplink(ctx, up)
	if ok {
		logger.Debug("Dropped duplicate uplink")
		registerReceiveUplinkDuplicate(ctx, up)
		return ttnpb.Empty, nil
	}
	registerReceiveUplink(ctx, up)

	defer func() {
		<-ns.collectionDone(ctx, up)
		stopDedup()
		logger.Debug("Done deduplicating uplink")
	}()

	logger.Debug("Handle uplink")
	return ttnpb.Empty, handle(ctx, up, acc)
}
