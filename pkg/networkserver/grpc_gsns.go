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
	"github.com/mohae/deepcopy"
	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
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

	// maxTransmissionDelay is the maximum delay between uplink retransmissions.
	maxTransmissionDelay = 10 * time.Second
)

var (
	// appQueueUpdateTimeout represents the time interval, within which AS
	// shall update the application queue after receiving the uplink.
	appQueueUpdateTimeout = 200 * time.Millisecond
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

type matchedDevice struct {
	Device   *ttnpb.EndDevice
	Pending  bool
	Session  *ttnpb.Session
	MACState *ttnpb.MACState
	FCnt     uint32
}

// matchDevice tries to match the uplink message with a device and returns the matched device.
func (ns *NetworkServer) matchDevice(ctx context.Context, up *ttnpb.UplinkMessage) (*matchedDevice, error) {
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

	var addrMatches []matchedDevice
	if err := ns.devices.RangeByAddr(ctx, pld.DevAddr,
		[]string{
			"frequency_plan_id",
			"lorawan_phy_version",
			"mac_settings.resets_f_cnt",
			"mac_settings.supports_32_bit_f_cnt",
			"mac_state",
			"pending_mac_state",
			"pending_session",
			"recent_downlinks",
			"recent_uplinks",
			"session",
		},
		func(dev *ttnpb.EndDevice) bool {
			if dev.MACState == nil && dev.PendingMACState == nil {
				return true
			}

			if dev.Session != nil && dev.Session.DevAddr == pld.DevAddr {
				addrMatches = append(addrMatches, matchedDevice{
					Device:   dev,
					Session:  dev.Session,
					MACState: dev.MACState,
				})
			}
			if dev.PendingSession != nil && dev.PendingSession.DevAddr == pld.DevAddr {
				if dev.Session != nil && dev.Session.DevAddr == pld.DevAddr {
					logger.Warn("Same DevAddr was assigned to a device in two consecutive sessions")
				}
				addrMatches = append(addrMatches, matchedDevice{
					Device:   dev,
					Session:  dev.PendingSession,
					MACState: dev.PendingMACState,
					Pending:  true,
				})
			}
			return true

		}); err != nil {
		logger.WithError(err).Warn("Failed to find devices in registry by DevAddr")
		return nil, err
	}
	if len(addrMatches) == 0 {
		logger.Debug("No device matched DevAddr")
		return nil, errDeviceNotFound
	}

	type device struct {
		matchedDevice
		gap       uint32
		fCntReset bool
		logger    log.Interface
	}
	matched := make([]device, 0, len(addrMatches))

	for _, match := range addrMatches {
		if pld.FCnt == 0 && match.Session.LastFCntUp == 0 && (len(match.Device.RecentUplinks) == 0 || match.Device.PendingSession != nil) {
			matched = append(matched, device{
				matchedDevice: match,
			})
		}

		supports32BitFCnt := true
		if match.Device.GetMACSettings().GetSupports32BitFCnt() != nil {
			supports32BitFCnt = match.Device.MACSettings.Supports32BitFCnt.Value
		} else if ns.defaultMACSettings.GetSupports32BitFCnt() != nil {
			supports32BitFCnt = ns.defaultMACSettings.Supports32BitFCnt.Value
		}

		fCnt := pld.FCnt
		switch {
		case !supports32BitFCnt, fCnt >= match.Session.LastFCntUp, fCnt == 0:
		case fCnt > match.Session.LastFCntUp&0xffff:
			fCnt |= match.Session.LastFCntUp &^ 0xffff
		case match.Session.LastFCntUp < 0xffff0000:
			fCnt |= (match.Session.LastFCntUp + 0x10000) &^ 0xffff
		}

		logger := logger.WithFields(log.Fields(
			"device_uid", unique.ID(ctx, match.Device.EndDeviceIdentifiers),
			"last_f_cnt_up", match.Session.LastFCntUp,
			"mac_version", match.MACState.LoRaWANVersion,
			"nb_trans", match.MACState.CurrentParameters.ADRNbTrans,
			"pending_session", match.Pending,
		))

		switch {
		case fCnt == match.Session.LastFCntUp:
			trans, lastAt := transmissionNumber(macPayloadBytes, match.Device.RecentUplinks...)
			if !lastAt.IsZero() {
				logger = logger.WithField("last_transmission_at", lastAt)
			}

			logger = logger.WithFields(log.Fields(
				"f_cnt_gap", 0,
				"full_f_cnt_up", match.Session.LastFCntUp,
				"transmission", trans,
			))

			if trans > match.MACState.CurrentParameters.ADRNbTrans ||
				!lastAt.IsZero() && up.ReceivedAt.Sub(lastAt) > match.MACState.CurrentParameters.Rx1Delay.Duration()+time.Second+maxTransmissionDelay {
				logger.Debug("Possible replay attack or malfunctioning device, skip")
				continue
			}
			match.FCnt = match.Session.LastFCntUp
			matched = append(matched, device{
				matchedDevice: match,
				logger:        logger,
			})

		case fCnt < match.Session.LastFCntUp:
			logger = logger.WithFields(log.Fields(
				"full_f_cnt_up", pld.FCnt,
				"transmission", 1,
			))

			if !resetsFCnt(match.Device, ns.defaultMACSettings) {
				logger.Debug("FCnt too low, skip")
				continue
			}

			_, phy, err := getDeviceBandVersion(match.Device, ns.FrequencyPlans)
			if err != nil {
				logger.WithError(err).Warn("Failed to get device's versioned band, skip")
				continue
			}
			if match.MACState.LoRaWANVersion.HasMaxFCntGap() && uint(pld.FCnt) > phy.MaxFCntGap {
				continue
			}

			var gap uint32
			if math.MaxUint32-match.Session.LastFCntUp < pld.FCnt {
				gap = match.Session.LastFCntUp + pld.FCnt
			} else {
				gap = math.MaxUint32
			}
			match.FCnt = pld.FCnt
			matched = append(matched, device{
				fCntReset:     true,
				gap:           gap,
				matchedDevice: match,
				logger:        logger.WithField("f_cnt_gap", gap),
			})

		default:
			logger = logger.WithField("transmission", 1)

			_, phy, err := getDeviceBandVersion(match.Device, ns.FrequencyPlans)
			if err != nil {
				logger.WithField("full_f_cnt_up", fCnt).WithError(err).Warn("Failed to get device's versioned band, skip")
				continue
			}

			if fCnt != pld.FCnt &&
				resetsFCnt(match.Device, ns.defaultMACSettings) &&
				(!match.MACState.LoRaWANVersion.HasMaxFCntGap() || uint(pld.FCnt) <= phy.MaxFCntGap) {

				var gap uint32
				if math.MaxUint32-match.Session.LastFCntUp < pld.FCnt {
					gap = match.Session.LastFCntUp + pld.FCnt
				} else {
					gap = math.MaxUint32
				}
				match.FCnt = pld.FCnt
				matched = append(matched, device{
					fCntReset:     true,
					gap:           gap,
					matchedDevice: match,
					logger: logger.WithFields(log.Fields(
						"f_cnt_gap", gap,
						"full_f_cnt_up", pld.FCnt,
					)),
				})
			}

			gap := fCnt - match.Session.LastFCntUp
			logger = logger.WithFields(log.Fields(
				"f_cnt_gap", gap,
				"full_f_cnt_up", fCnt,
			))
			if match.MACState.LoRaWANVersion.HasMaxFCntGap() && uint(gap) > phy.MaxFCntGap {
				logger.Debug("FCnt gap too high, skip")
				continue
			}
			match.FCnt = fCnt
			matched = append(matched, device{
				gap:           gap,
				matchedDevice: match,
				logger:        logger,
			})
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		if matched[i].gap != matched[j].gap {
			return matched[i].gap < matched[j].gap
		}
		if matched[i].fCntReset != matched[j].fCntReset {
			return matched[j].fCntReset
		}
		return matched[i].Session.LastFCntUp < matched[j].Session.LastFCntUp
	})

	logger.WithField("device_count", len(matched)).Debug("Perform MIC checks on devices with matching frame counters")
	for _, match := range matched {
		logger := match.logger

		if pld.Ack {
			if len(match.Device.RecentDownlinks) == 0 {
				// Uplink acknowledges a downlink, but no downlink was sent to the device,
				// hence it must be the wrong device.
				logger.Debug("Uplink contains ACK, but no downlink was sent to device, skip")
				continue
			}
		}

		if match.Session.FNwkSIntKey == nil || len(match.Session.FNwkSIntKey.Key) == 0 {
			logger.Warn("Device missing FNwkSIntKey in registry, skip")
			continue
		}

		fNwkSIntKey, err := cryptoutil.UnwrapAES128Key(*match.Session.FNwkSIntKey, ns.KeyVault)
		if err != nil {
			logger.WithField("kek_label", match.Session.FNwkSIntKey.KEKLabel).WithError(err).Warn("Failed to unwrap FNwkSIntKey, skip")
			continue
		}

		if match.fCntReset {
			// TODO: Handle MAC state reset(https://github.com/TheThingsNetwork/lorawan-stack/issues/505)
		}

		var computedMIC [4]byte
		if match.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
			computedMIC, err = crypto.ComputeLegacyUplinkMIC(
				fNwkSIntKey,
				pld.DevAddr,
				match.FCnt,
				macPayloadBytes,
			)
		} else {
			if match.Session.SNwkSIntKey == nil || len(match.Session.SNwkSIntKey.Key) == 0 {
				logger.Warn("Device missing SNwkSIntKey in registry, skip")
				continue
			}

			drIdx, err := searchDataRate(up.Settings.DataRate, match.Device, ns.FrequencyPlans)
			if err != nil {
				logger.WithError(err).Debug("Failed to determine data rate index of uplink")
				continue
			}

			chIdx, err := searchUplinkChannel(up.Settings.Frequency, match.MACState)
			if err != nil {
				logger.WithError(err).Debug("Failed to determine channel index of uplink")
				continue
			}

			sNwkSIntKey, err := cryptoutil.UnwrapAES128Key(*match.Session.SNwkSIntKey, ns.KeyVault)
			if err != nil {
				logger.WithField("kek_label", match.Session.SNwkSIntKey.KEKLabel).WithError(err).Warn("Failed to unwrap SNwkSIntKey, skip")
				continue
			}

			var confFCnt uint32
			if pld.Ack {
				confFCnt = match.Session.LastConfFCntDown
			}
			computedMIC, err = crypto.ComputeUplinkMIC(
				sNwkSIntKey,
				fNwkSIntKey,
				confFCnt,
				uint8(drIdx),
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

		if match.FCnt == math.MaxUint32 {
			return nil, errFCntTooHigh
		}
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

func (ns *NetworkServer) handleUplink(ctx context.Context, up *ttnpb.UplinkMessage, acc *metadataAccumulator) (err error) {
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
	matched, err := ns.matchDevice(ctx, up)
	if err != nil {
		registerDropDataUplink(ctx, nil, up, err)
		logger.WithError(err).Debug("Failed to match device")
		return errDeviceNotFound.WithCause(err)
	}

	logger = logger.WithField("device_uid", unique.ID(ctx, matched.Device.EndDeviceIdentifiers))
	ctx = log.NewContext(ctx, logger)

	logger.Debug("Matched device")

	if matched.MACState.PendingApplicationDownlink != nil {
		asUp := &ttnpb.ApplicationUp{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				DevAddr:                &pld.DevAddr,
				JoinEUI:                matched.Device.JoinEUI,
				DevEUI:                 matched.Device.DevEUI,
				ApplicationIdentifiers: matched.Device.ApplicationIdentifiers,
				DeviceID:               matched.Device.DeviceID,
			},
			CorrelationIDs: matched.MACState.PendingApplicationDownlink.CorrelationIDs,
			ReceivedAt:     &up.ReceivedAt,
		}

		if pld.Ack {
			asUp.Up = &ttnpb.ApplicationUp_DownlinkAck{
				DownlinkAck: matched.MACState.PendingApplicationDownlink,
			}
		} else {
			asUp.Up = &ttnpb.ApplicationUp_DownlinkNack{
				DownlinkNack: matched.MACState.PendingApplicationDownlink,
			}
		}
		asUp.CorrelationIDs = append(asUp.CorrelationIDs, up.CorrelationIDs...)

		matched.MACState.PendingApplicationDownlink = nil

		asCtx, cancel := context.WithTimeout(ctx, appQueueUpdateTimeout)
		defer cancel()

		logger.Debug("Send downlink (n)ack to Application Server")
		ok, err := ns.handleASUplink(asCtx, matched.Device.EndDeviceIdentifiers.ApplicationIdentifiers, asUp)
		if err != nil {
			return err
		}
		if !ok {
			logger.Warn("Application Server not found, downlink (n)ack not sent")
		}
	}

	macBuf := pld.FOpts
	if len(macBuf) == 0 && pld.FPort == 0 {
		macBuf = pld.FRMPayload
	}

	if len(macBuf) > 0 && (len(pld.FOpts) == 0 || matched.MACState != nil && matched.MACState.LoRaWANVersion.EncryptFOpts()) {
		if matched.Session.NwkSEncKey == nil || len(matched.Session.NwkSEncKey.Key) == 0 {
			return errUnknownNwkSEncKey
		}
		key, err := cryptoutil.UnwrapAES128Key(*matched.Session.NwkSEncKey, ns.KeyVault)
		if err != nil {
			logger.WithField("kek_label", matched.Session.NwkSEncKey.KEKLabel).WithError(err).Warn("Failed to unwrap NwkSEncKey")
			return err
		}

		macBuf, err = crypto.DecryptUplink(key, pld.DevAddr, pld.FCnt, macBuf)
		if err != nil {
			return errDecrypt.WithCause(err)
		}
	}

	_, phy, err := getDeviceBandVersion(matched.Device, ns.FrequencyPlans)
	if err != nil {
		return errUnknownBand.WithCause(err)
	}
	var cmds []*ttnpb.MACCommand
	for r := bytes.NewReader(macBuf); r.Len() > 0; {
		cmd := &ttnpb.MACCommand{}
		if err := lorawan.DefaultMACCommands.ReadUplink(phy, r, cmd); err != nil {
			logger.WithFields(log.Fields(
				"bytes_left", r.Len(),
				"mac_count", len(cmds),
			)).WithError(err).Warn("Failed to unmarshal MAC command")
			break
		}
		logger.WithField("cid", cmd.CID).Debug("Read MAC command")
		cmds = append(cmds, cmd)
	}
	logger = logger.WithField("mac_count", len(cmds))
	ctx = log.NewContext(ctx, logger)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ns.deduplicationDone(ctx, up):
	}

	up.RxMetadata = acc.Accumulated()
	logger = logger.WithField("metadata_count", len(up.RxMetadata))
	logger.Debug("Merged metadata")
	ctx = log.NewContext(ctx, logger)
	registerMergeMetadata(ctx, &matched.Device.EndDeviceIdentifiers, up)

	var nbTrans uint32
	var handleErr bool
	stored, err := ns.devices.SetByID(ctx, matched.Device.ApplicationIdentifiers, matched.Device.DeviceID,
		[]string{
			"downlink_margin",
			"frequency_plan_id",
			"last_dev_status_received_at",
			"lorawan_phy_version",
			"lorawan_version",
			"mac_settings",
			"mac_state",
			"pending_mac_state",
			"pending_session",
			"recent_uplinks",
			"session",
			"supports_class_b",
			"supports_class_c",
			"supports_join",
		},
		func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if stored == nil {
				logger.Warn("Device deleted during uplink handling, drop")
				handleErr = true
				return nil, nil, errOutdatedData
			}

			storedSession := stored.Session
			if matched.Pending {
				storedSession = stored.PendingSession
			}

			if !bytes.Equal(storedSession.GetSessionKeyID(), matched.Session.SessionKeyID) {
				logger.Warn("Device changed session during uplink handling, drop")
				handleErr = true
				return nil, nil, errOutdatedData
			}
			if storedSession.GetLastFCntUp() != matched.Session.LastFCntUp && !resetsFCnt(stored, ns.defaultMACSettings) {
				logger.WithFields(log.Fields(
					"stored_f_cnt", storedSession.GetLastFCntUp(),
					"got_f_cnt", matched.Session.LastFCntUp,
				)).Warn("A more recent uplink was received by device during uplink handling, drop")
				handleErr = true
				return nil, nil, errOutdatedData
			}

			var paths []string

			if matched.Pending {
				stored.MACState = stored.PendingMACState
				stored.PendingSession.LastFCntUp = matched.FCnt
				stored.PendingSession.StartedAt = up.ReceivedAt

				if stored.SupportsClassC {
					stored.MACState.DeviceClass = ttnpb.CLASS_C
				} else if stored.SupportsClassB {
					stored.MACState.DeviceClass = ttnpb.CLASS_B
				}
			} else {
				stored.Session.LastFCntUp = matched.FCnt
			}
			stored.PendingMACState = nil
			paths = append(paths,
				"mac_state",
				"pending_mac_state",
				"pending_session",
				"session",
			)

			upChIdx, err := searchUplinkChannel(up.Settings.Frequency, stored.MACState)
			if err != nil {
				return nil, nil, err
			}
			up.DeviceChannelIndex = uint32(upChIdx)

			upDRIdx, err := searchDataRate(up.Settings.DataRate, stored, ns.Component.FrequencyPlans)
			if err != nil {
				return nil, nil, err
			}
			up.Settings.DataRateIndex = upDRIdx

			var lastAt time.Time
			nbTrans, lastAt = transmissionNumber(up.RawPayload[:len(up.RawPayload)-4], stored.RecentUplinks...)
			if nbTrans > stored.MACState.CurrentParameters.ADRNbTrans ||
				!lastAt.IsZero() && up.ReceivedAt.Sub(lastAt) > stored.MACState.CurrentParameters.Rx1Delay.Duration()+time.Second+maxTransmissionDelay {
				logger.Warn("A more recent uplink retransmission was received during uplink handling, drop")
				return nil, nil, errOutdatedData
			}

			stored.RecentUplinks = appendRecentUplink(stored.RecentUplinks, up, recentUplinkCount)
			paths = append(paths, "recent_uplinks")

			if nbTrans == 1 {
				stored.MACState.QueuedResponses = stored.MACState.QueuedResponses[:0]
			outer:
				for len(cmds) > 0 {
					var cmd *ttnpb.MACCommand
					cmd, cmds = cmds[0], cmds[1:]
					logger := logger.WithField("cid", cmd.CID)
					logger.Debug("Handle MAC command")
					switch cmd.CID {
					case ttnpb.CID_RESET:
						err = handleResetInd(ctx, stored, cmd.GetResetInd(), ns.FrequencyPlans, ns.defaultMACSettings)
					case ttnpb.CID_LINK_CHECK:
						err = handleLinkCheckReq(ctx, stored, up)
					case ttnpb.CID_LINK_ADR:
						pld := cmd.GetLinkADRAns()
						dupCount := 0
						if stored.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_0_2) >= 0 && stored.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
							logger.Debug("Count duplicates")
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
							logger.WithField("duplicate_count", dupCount).Debug("Counted duplicates")
						}
						if err != nil {
							break
						}
						cmds = cmds[dupCount:]
						err = handleLinkADRAns(ctx, stored, pld, uint(dupCount), ns.FrequencyPlans)
					case ttnpb.CID_DUTY_CYCLE:
						err = handleDutyCycleAns(ctx, stored)
					case ttnpb.CID_RX_PARAM_SETUP:
						err = handleRxParamSetupAns(ctx, stored, cmd.GetRxParamSetupAns())
					case ttnpb.CID_DEV_STATUS:
						err = handleDevStatusAns(ctx, stored, cmd.GetDevStatusAns(), matched.Session.LastFCntUp, up.ReceivedAt)
						paths = append(paths,
							"battery_percentage",
							"downlink_margin",
							"last_dev_status_received_at",
							"power_state",
						)
					case ttnpb.CID_NEW_CHANNEL:
						err = handleNewChannelAns(ctx, stored, cmd.GetNewChannelAns())
					case ttnpb.CID_RX_TIMING_SETUP:
						err = handleRxTimingSetupAns(ctx, stored)
					case ttnpb.CID_TX_PARAM_SETUP:
						err = handleTxParamSetupAns(ctx, stored)
					case ttnpb.CID_DL_CHANNEL:
						err = handleDLChannelAns(ctx, stored, cmd.GetDLChannelAns())
					case ttnpb.CID_REKEY:
						err = handleRekeyInd(ctx, stored, cmd.GetRekeyInd())
					case ttnpb.CID_ADR_PARAM_SETUP:
						err = handleADRParamSetupAns(ctx, stored)
					case ttnpb.CID_DEVICE_TIME:
						err = handleDeviceTimeReq(ctx, stored, up)
					case ttnpb.CID_REJOIN_PARAM_SETUP:
						err = handleRejoinParamSetupAns(ctx, stored, cmd.GetRejoinParamSetupAns())
					case ttnpb.CID_PING_SLOT_INFO:
						err = handlePingSlotInfoReq(ctx, stored, cmd.GetPingSlotInfoReq())
					case ttnpb.CID_PING_SLOT_CHANNEL:
						err = handlePingSlotChannelAns(ctx, stored, cmd.GetPingSlotChannelAns())
					case ttnpb.CID_BEACON_TIMING:
						err = handleBeaconTimingReq(ctx, stored)
					case ttnpb.CID_BEACON_FREQ:
						err = handleBeaconFreqAns(ctx, stored, cmd.GetBeaconFreqAns())
					case ttnpb.CID_DEVICE_MODE:
						err = handleDeviceModeInd(ctx, stored, cmd.GetDeviceModeInd())
					default:
						h, ok := ns.macHandlers.Load(cmd.CID)
						if !ok {
							logger.WithField("cid", cmd.CID).Warn("Unknown MAC command received, skip the rest")
							break outer
						}
						err = h.(MACHandler)(ctx, stored, cmd.GetRawPayload(), up)
					}
					if err != nil {
						logger.WithField("cid", cmd.CID).WithError(err).Warn("Failed to process MAC command")
					}
				}
			}

			if matched.Pending {
				if stored.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
					stored.Session = stored.PendingSession
					stored.EndDeviceIdentifiers.DevAddr = &pld.DevAddr
				} else if stored.PendingSession != nil {
					handleErr = true
					return nil, nil, errNoRekey
				}
				paths = append(paths, "ids.dev_addr")
			}
			if !matched.Pending && stored.PendingSession != nil {
				// TODO: Notify AS of session recovery(https://github.com/TheThingsNetwork/lorawan-stack/issues/594)
			}

			stored.PendingSession = nil
			stored.MACState.PendingApplicationDownlink = nil
			stored.MACState.PendingJoinRequest = nil
			stored.MACState.RxWindowsAvailable = true

			paths = append(paths, "recent_adr_uplinks")
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
		registerDropDataUplink(ctx, &matched.Device.EndDeviceIdentifiers, up, err)
	}
	if err != nil {
		registerDropDataUplink(ctx, &matched.Device.EndDeviceIdentifiers, up, err)
		return err
	}

	asCtx, cancel := context.WithTimeout(ctx, appQueueUpdateTimeout)
	defer cancel()

	if nbTrans == 1 {
		logger.Debug("Send uplink to Application Server")
		ok, err := ns.handleASUplink(asCtx, stored.ApplicationIdentifiers, &ttnpb.ApplicationUp{
			EndDeviceIdentifiers: stored.EndDeviceIdentifiers,
			CorrelationIDs:       up.CorrelationIDs,
			ReceivedAt:           &up.ReceivedAt,
			Up: &ttnpb.ApplicationUp_UplinkMessage{UplinkMessage: &ttnpb.ApplicationUplink{
				FCnt:         stored.Session.LastFCntUp,
				FPort:        pld.FPort,
				FRMPayload:   pld.FRMPayload,
				RxMetadata:   up.RxMetadata,
				SessionKeyID: stored.Session.SessionKeyID,
				Settings:     up.Settings,
			}},
		})
		if err != nil {
			logger.WithError(err).Error("Failed to forward uplink to AS")
		} else if !ok {
			logger.Warn("Application Server not found, not forwarding uplink")
		} else {
			registerForwardDataUplink(ctx, &stored.EndDeviceIdentifiers, up)
		}
	}
	startAt := time.Now().UTC()
	logger.WithField("start_at", startAt).Debug("Add downlink task for class A downlink")
	return ns.downlinkTasks.Add(ctx, stored.EndDeviceIdentifiers, startAt, true)
}

// newDevAddr generates a DevAddr for specified EndDevice.
func (ns *NetworkServer) newDevAddr(context.Context, *ttnpb.EndDevice) types.DevAddr {
	var devAddr types.DevAddr
	random.Read(devAddr[:])
	prefix := ns.devAddrPrefixes[random.Intn(len(ns.devAddrPrefixes))]
	return devAddr.WithPrefix(prefix)
}

func (ns *NetworkServer) handleJoin(ctx context.Context, up *ttnpb.UplinkMessage, acc *metadataAccumulator) (err error) {
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
		},
	)
	if err != nil {
		registerDropJoinRequest(ctx, nil, up, err)
		logger.WithError(err).Warn("Failed to load device from registry")
		return err
	}

	defer func(dev *ttnpb.EndDevice) {
		if err != nil {
			registerDropJoinRequest(ctx, &dev.EndDeviceIdentifiers, up, err)
		}
	}(dev)

	logger = logger.WithField("device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers))
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
				macState.CurrentParameters.Channels[i].EnableUplink = m
			}
		}
	}

	js, err := ns.jsClient(ctx, dev.EndDeviceIdentifiers)
	if err != nil {
		logger.WithError(err).Debug("Could not get Join Server")
		return err
	}

	logger.Debug("Send join-request to Join Server")
	resp, err := js.HandleJoin(ctx, req, ns.WithClusterAuth())
	if err != nil {
		logger.WithError(err).Warn("Join Server failed to handle join-request")
		return err
	}
	logger.Debug("Join-accept received from Join Server")

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

	registerForwardJoinRequest(ctx, &dev.EndDeviceIdentifiers, up)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ns.deduplicationDone(ctx, up):
	}

	up.RxMetadata = acc.Accumulated()
	registerMergeMetadata(ctx, &dev.EndDeviceIdentifiers, up)

	var invalidatedQueue []*ttnpb.ApplicationDownlink
	var resetErr bool
	dev, err = ns.devices.SetByID(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers, dev.EndDeviceIdentifiers.DeviceID,
		[]string{
			"frequency_plan_id",
			"lorawan_phy_version",
			"queued_application_downlinks",
			"recent_uplinks",
		},
		func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
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
	if err != nil && !resetErr {
		logger.WithError(err).Warn("Failed to update device in registry")
		// TODO: Retry transaction. (https://github.com/TheThingsNetwork/lorawan-stack/issues/33)
	}
	if err != nil {
		return err
	}

	logger = logger.WithField(
		"application_uid", unique.ID(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers),
	)
	logger.Debug("Send join-accept to AS")
	_, err = ns.handleASUplink(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers, &ttnpb.ApplicationUp{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: dev.EndDeviceIdentifiers.ApplicationIdentifiers,
			DeviceID:               dev.EndDeviceIdentifiers.DeviceID,
			DevEUI:                 dev.EndDeviceIdentifiers.DevEUI,
			JoinEUI:                dev.EndDeviceIdentifiers.JoinEUI,
			DevAddr:                &devAddr,
		},
		CorrelationIDs: events.CorrelationIDsFromContext(ctx),
		ReceivedAt:     &up.ReceivedAt,
		Up: &ttnpb.ApplicationUp_JoinAccept{JoinAccept: &ttnpb.ApplicationJoinAccept{
			AppSKey:              resp.SessionKeys.AppSKey,
			InvalidatedDownlinks: invalidatedQueue,
			SessionKeyID:         resp.SessionKeys.SessionKeyID,
		}},
	})
	if err != nil {
		logger.WithError(err).Errorf("Failed to send join-accept to AS")
		return err
	}

	startAt := time.Now().UTC()
	logger.WithField("start_at", startAt).Debug("Add downlink task for join-accept")
	if err := ns.downlinkTasks.Add(ctx, dev.EndDeviceIdentifiers, startAt, true); err != nil {
		return err
	}
	return nil
}

func (ns *NetworkServer) handleRejoin(ctx context.Context, up *ttnpb.UplinkMessage, acc *metadataAccumulator) (err error) {
	defer func() {
		if err != nil {
			registerDropRejoinRequest(ctx, nil, up, err)
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
	up.ReceivedAt = time.Now().UTC()
	up.Payload = &ttnpb.Message{}
	if err := lorawan.UnmarshalMessage(up.RawPayload, up.Payload); err != nil {
		return nil, errDecodePayload.WithCause(err)
	}

	if up.Payload.Major != ttnpb.Major_LORAWAN_R1 {
		return nil, errUnsupportedLoRaWANVersion.WithAttributes(
			"major", up.Payload.Major,
		)
	}

	logger := log.FromContext(ctx).WithFields(log.Fields(
		"m_type", up.Payload.MType,
		"major", up.Payload.Major,
		"received_at", up.ReceivedAt,
	))
	ctx = log.NewContext(ctx, logger)

	logger.Debug("Deduplicate uplink")
	acc, stopDedup, ok := ns.deduplicateUplink(ctx, up)
	if ok {
		logger.Debug("Dropped duplicate uplink")
		registerReceiveUplinkDuplicate(ctx, up)
		return ttnpb.Empty, nil
	}
	registerReceiveUplink(ctx, up)

	defer func(up *ttnpb.UplinkMessage) {
		<-ns.collectionDone(ctx, up)
		stopDedup()
		logger.Debug("Done deduplicating uplink")
	}(up)

	up = deepcopy.Copy(up).(*ttnpb.UplinkMessage)
	switch up.Payload.MType {
	case ttnpb.MType_CONFIRMED_UP, ttnpb.MType_UNCONFIRMED_UP:
		logger.Debug("Handle data uplink")
		return ttnpb.Empty, ns.handleUplink(ctx, up, acc)
	case ttnpb.MType_JOIN_REQUEST:
		logger.Debug("Handle join-request")
		return ttnpb.Empty, ns.handleJoin(ctx, up, acc)
	case ttnpb.MType_REJOIN_REQUEST:
		logger.Debug("Handle rejoin-request")
		return ttnpb.Empty, ns.handleRejoin(ctx, up, acc)
	default:
		logger.Warn("Unmatched MType")
		return ttnpb.Empty, nil
	}
}
