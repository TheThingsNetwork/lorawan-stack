// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

// Package networkserver provides a LoRaWAN 1.1-compliant Network Server implementation.
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
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/crypto"
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
)

var (
	// appQueueUpdateTimeout represents the time interval, within which AS
	// shall update the application queue after receiving the uplink.
	appQueueUpdateTimeout = 200 * time.Millisecond
)

func resetMACState(fps *frequencyplans.Store, dev *ttnpb.EndDevice) error {
	fp, err := fps.GetByID(dev.FrequencyPlanID)
	if err != nil {
		return err
	}

	band, err := band.GetByID(fp.BandID)
	if err != nil {
		return err
	}

	dev.MACState = &ttnpb.MACState{
		LoRaWANVersion: dev.LoRaWANVersion,
		CurrentParameters: ttnpb.MACParameters{
			ADRAckDelay:       uint32(band.ADRAckDelay),
			ADRAckLimit:       uint32(band.ADRAckLimit),
			ADRNbTrans:        1,
			MaxDutyCycle:      ttnpb.DUTY_CYCLE_1,
			MaxEIRP:           band.DefaultMaxEIRP,
			Rx1Delay:          ttnpb.RxDelay(band.ReceiveDelay1.Seconds()),
			Rx1DataRateOffset: 0,
			Rx2DataRateIndex:  band.DefaultRx2Parameters.DataRateIndex,
			Rx2Frequency:      band.DefaultRx2Parameters.Frequency,
		},
	}

	// NOTE: dev.MACState.CurrentParameters must not contain pointer values at this point.
	dev.MACState.DesiredParameters = dev.MACState.CurrentParameters

	if len(band.DownlinkChannels) > len(band.UplinkChannels) || len(fp.DownlinkChannels) > len(fp.UplinkChannels) {
		// NOTE: In case the spec changes and this assumption is not valid anymore,
		// the implementation of this function won't be valid and has to be changed.
		return errInvalidFrequencyPlan
	}

	dev.MACState.CurrentParameters.Channels = make([]*ttnpb.MACParameters_Channel, len(band.UplinkChannels))
	dev.MACState.DesiredParameters.Channels = make([]*ttnpb.MACParameters_Channel, int(math.Max(float64(len(dev.MACState.CurrentParameters.Channels)), float64(len(fp.UplinkChannels)))))

	for i, upCh := range band.UplinkChannels {
		if len(upCh.DataRateIndexes) == 0 {
			return errInvalidFrequencyPlan
		}

		ch := &ttnpb.MACParameters_Channel{
			UplinkFrequency:  upCh.Frequency,
			MinDataRateIndex: ttnpb.DataRateIndex(upCh.DataRateIndexes[0]),
			MaxDataRateIndex: ttnpb.DataRateIndex(upCh.DataRateIndexes[len(upCh.DataRateIndexes)-1]),
		}
		dev.MACState.CurrentParameters.Channels[i] = ch

		chCopy := *ch
		dev.MACState.DesiredParameters.Channels[i] = &chCopy
	}

	for i, downCh := range band.DownlinkChannels {
		if i >= len(dev.MACState.CurrentParameters.Channels) {
			return errInvalidFrequencyPlan
		}
		dev.MACState.CurrentParameters.Channels[i].DownlinkFrequency = downCh.Frequency
	}

	for i, upCh := range fp.UplinkChannels {
		ch := dev.MACState.DesiredParameters.Channels[i]
		if ch == nil {
			dev.MACState.DesiredParameters.Channels[i] = &ttnpb.MACParameters_Channel{
				MinDataRateIndex: ttnpb.DataRateIndex(upCh.MinDataRate),
				MaxDataRateIndex: ttnpb.DataRateIndex(upCh.MaxDataRate),
				UplinkFrequency:  upCh.Frequency,
			}
			continue
		}

		if ch.MinDataRateIndex > ttnpb.DataRateIndex(upCh.MinDataRate) || ttnpb.DataRateIndex(upCh.MaxDataRate) > ch.MaxDataRateIndex {
			return errInvalidFrequencyPlan
		}
		ch.MinDataRateIndex = ttnpb.DataRateIndex(upCh.MinDataRate)
		ch.MaxDataRateIndex = ttnpb.DataRateIndex(upCh.MaxDataRate)
		ch.UplinkFrequency = upCh.Frequency
	}

	for i, downCh := range fp.DownlinkChannels {
		if i >= len(dev.MACState.DesiredParameters.Channels) {
			return errInvalidFrequencyPlan
		}
		dev.MACState.DesiredParameters.Channels[i].DownlinkFrequency = downCh.Frequency
	}

	dev.MACState.DesiredParameters.UplinkDwellTime = fp.DwellTime.GetUplinks()
	dev.MACState.DesiredParameters.DownlinkDwellTime = fp.DwellTime.GetDownlinks()

	if fp.Rx2 != nil {
		dev.MACState.DesiredParameters.Rx2Frequency = fp.Rx2.Frequency
	}
	if fp.DefaultRx2DataRate != nil {
		dev.MACState.DesiredParameters.Rx2DataRateIndex = ttnpb.DataRateIndex(*fp.DefaultRx2DataRate)
	}

	if fp.PingSlot != nil {
		dev.MACState.DesiredParameters.PingSlotFrequency = fp.PingSlot.Frequency
	}
	if fp.DefaultPingSlotDataRate != nil {
		dev.MACState.DesiredParameters.PingSlotDataRateIndex = ttnpb.DataRateIndex(*fp.DefaultPingSlotDataRate)
	}

	if fp.MaxEIRP != nil && *fp.MaxEIRP > 0 {
		dev.MACState.DesiredParameters.MaxEIRP = float32(math.Min(float64(dev.MACState.CurrentParameters.MaxEIRP), float64(*fp.MaxEIRP)))
	}

	if dev.DefaultMACParameters != nil {
		dev.MACState.CurrentParameters = deepcopy.Copy(*dev.DefaultMACParameters).(ttnpb.MACParameters)
	}

	return nil
}

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

// matchDevice tries to match the uplink message with a device and returns the matched device and session.
// The LastFCntUp in the matched session is updated according to the FCnt in up.
func (ns *NetworkServer) matchDevice(ctx context.Context, up *ttnpb.UplinkMessage) (*ttnpb.EndDevice, *ttnpb.Session, error) {
	pld := up.Payload.GetMACPayload()

	logger := log.FromContext(ctx).WithField("dev_addr", pld.DevAddr)

	type device struct {
		*ttnpb.EndDevice

		matchedSession *ttnpb.Session
		fCnt           uint32
		gap            uint32
	}

	var devs []device
	if err := ns.devices.RangeByAddr(pld.DevAddr,
		[]string{
			"frequency_plan_id",
			"mac_state",
			"pending_session",
			"recent_downlinks",
			"recent_uplinks",
			"resets_f_cnt",
			"session",
			"uses_32_bit_f_cnt",
		},
		func(dev *ttnpb.EndDevice) bool {
			if dev.MACState == nil || (dev.Session == nil && dev.PendingSession == nil) {
				return true
			}

			ses := dev.Session
			if ses == nil {
				ses = dev.PendingSession
			}
			devs = append(devs, device{
				EndDevice:      dev,
				matchedSession: ses,
			})
			return true

		}); err != nil {
		logger.WithError(err).Warn("Failed to find devices in registry by DevAddr")
		return nil, nil, err
	}

	matching := make([]device, 0, len(devs))

outer:
	for _, dev := range devs {
		fCnt := pld.FCnt

		switch {
		case !dev.Uses32BitFCnt, fCnt > dev.matchedSession.LastFCntUp:
		case fCnt > dev.matchedSession.LastFCntUp&0xffff:
			fCnt |= dev.matchedSession.LastFCntUp &^ 0xffff
		case dev.matchedSession.LastFCntUp < 0xffff0000:
			fCnt |= (dev.matchedSession.LastFCntUp + 0x10000) &^ 0xffff
		}

		gap := uint32(math.MaxUint32)
		if fCnt == 0 && dev.matchedSession.LastFCntUp == 0 && len(dev.RecentUplinks) == 0 {
			gap = 0
		} else if !dev.ResetsFCnt {
			if fCnt <= dev.matchedSession.LastFCntUp {
				continue outer
			}

			gap = fCnt - dev.matchedSession.LastFCntUp

			if dev.MACState.LoRaWANVersion.HasMaxFCntGap() {
				fp, err := ns.Component.FrequencyPlans.GetByID(dev.FrequencyPlanID)
				if err != nil {
					logger.WithError(err).WithFields(log.Fields(
						"device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers),
					)).Warn("Failed to get the frequency plan of the device in registry")
					continue
				}

				band, err := band.GetByID(fp.BandID)
				if err != nil {
					logger.WithError(err).WithFields(log.Fields(
						"device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers),
					)).Warn("Failed to get the band of the device in registry")
					continue
				}

				if gap > uint32(band.MaxFCntGap) {
					continue outer
				}
			}
		}

		matching = append(matching, device{
			EndDevice:      dev.EndDevice,
			matchedSession: dev.matchedSession,
			gap:            gap,
			fCnt:           fCnt,
		})
		if dev.ResetsFCnt && fCnt != pld.FCnt {
			matching = append(matching, device{
				EndDevice:      dev.EndDevice,
				matchedSession: dev.matchedSession,
				gap:            gap,
				fCnt:           pld.FCnt,
			})
		}
	}

	sort.Slice(matching, func(i, j int) bool {
		return matching[i].gap < matching[j].gap
	})

	if len(up.RawPayload) < 4 {
		return nil, nil, errRawPayloadTooShort
	}
	b := up.RawPayload[:len(up.RawPayload)-4]

	for _, dev := range matching {
		if pld.Ack {
			if len(dev.RecentDownlinks) == 0 {
				// Uplink acknowledges a downlink, but no downlink was sent to the device,
				// hence it must be the wrong device.
				continue
			}
		}

		if dev.matchedSession.FNwkSIntKey == nil || len(dev.matchedSession.FNwkSIntKey.Key) == 0 {
			logger.WithFields(log.Fields(
				"device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers),
			)).Warn("Device missing FNwkSIntKey in registry")
			continue
		}

		var fNwkSIntKey types.AES128Key
		if dev.matchedSession.FNwkSIntKey.KEKLabel != "" {
			// TODO: https://github.com/TheThingsIndustries/lorawan-stack/issues/271
			panic("Unsupported")
		}
		copy(fNwkSIntKey[:], dev.matchedSession.FNwkSIntKey.Key[:])

		var computedMIC [4]byte
		var err error
		if dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
			computedMIC, err = crypto.ComputeLegacyUplinkMIC(
				fNwkSIntKey,
				pld.DevAddr,
				dev.fCnt,
				b,
			)

		} else {
			if dev.matchedSession.SNwkSIntKey == nil || len(dev.matchedSession.SNwkSIntKey.Key) == 0 {
				logger.WithFields(log.Fields(
					"device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers),
				)).Warn("Device missing SNwkSIntKey in registry")
				continue
			}

			var sNwkSIntKey types.AES128Key
			if dev.matchedSession.SNwkSIntKey.KEKLabel != "" {
				// TODO: https://github.com/TheThingsIndustries/lorawan-stack/issues/271
				panic("Unsupported")
			}
			copy(sNwkSIntKey[:], dev.matchedSession.SNwkSIntKey.Key[:])

			var confFCnt uint32
			if pld.Ack {
				confFCnt = dev.matchedSession.LastConfFCntDown
			}

			computedMIC, err = crypto.ComputeUplinkMIC(
				sNwkSIntKey,
				fNwkSIntKey,
				confFCnt,
				uint8(up.Settings.DataRateIndex),
				uint8(up.Settings.ChannelIndex),
				pld.DevAddr,
				dev.fCnt,
				b,
			)
		}
		if err != nil {
			logger.WithError(err).WithFields(log.Fields(
				"device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers),
			)).Error("Failed to compute MIC")
			continue
		}
		if !bytes.Equal(up.Payload.MIC, computedMIC[:]) {
			continue
		}

		if dev.fCnt == math.MaxUint32 {
			return nil, nil, errFCntTooHigh
		}
		dev.matchedSession.LastFCntUp = dev.fCnt
		return dev.EndDevice, dev.matchedSession, nil
	}
	return nil, nil, errDeviceNotFound
}

// MACHandler defines the behavior of a MAC command on a device.
type MACHandler func(ctx context.Context, dev *ttnpb.EndDevice, pld []byte, up *ttnpb.UplinkMessage) error

func (ns *NetworkServer) handleUplink(ctx context.Context, up *ttnpb.UplinkMessage, acc *metadataAccumulator) (err error) {
	defer func() {
		if err != nil {
			registerDropUplink(ctx, up, err)
		}
	}()

	dev, ses, err := ns.matchDevice(ctx, up)
	if err != nil {
		return errDeviceNotFound.WithCause(err)
	}

	logger := log.FromContext(ctx).WithField("device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers))

	pld := up.Payload.GetMACPayload()
	if pld == nil {
		return errNoPayload
	}

	if dev.MACState != nil && dev.MACState.PendingApplicationDownlink != nil {
		asUp := &ttnpb.ApplicationUp{
			EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
			CorrelationIDs:       dev.MACState.PendingApplicationDownlink.CorrelationIDs,
		}

		if pld.Ack {
			asUp.Up = &ttnpb.ApplicationUp_DownlinkAck{
				DownlinkAck: dev.MACState.PendingApplicationDownlink,
			}
		} else {
			asUp.Up = &ttnpb.ApplicationUp_DownlinkNack{
				DownlinkNack: dev.MACState.PendingApplicationDownlink,
			}
		}
		asUp.CorrelationIDs = append(asUp.CorrelationIDs, up.CorrelationIDs...)

		logger.Debug("Sending downlink (n)ack to Application Server...")
		if ok, err := ns.handleASUplink(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers, asUp); err != nil {
			return err
		} else if !ok {
			logger.Warn("Application Server not found, downlink (n)ack not sent")
		}
	}

	mac := pld.FOpts
	if len(mac) == 0 && pld.FPort == 0 {
		mac = pld.FRMPayload
	}

	if len(mac) > 0 && (len(pld.FOpts) == 0 || dev.MACState.LoRaWANVersion.EncryptFOpts()) {
		if ses.NwkSEncKey == nil || len(ses.NwkSEncKey.Key) == 0 {
			return errUnknownNwkSEncKey
		}

		var key types.AES128Key
		if ses.NwkSEncKey.KEKLabel != "" {
			// TODO: https://github.com/TheThingsIndustries/lorawan-stack/issues/271
			panic("Unsupported")
		}
		copy(key[:], ses.NwkSEncKey.Key[:])

		mac, err = crypto.DecryptUplink(key, *dev.EndDeviceIdentifiers.DevAddr, pld.FCnt, mac)
		if err != nil {
			return errDecrypt.WithCause(err)
		}
	}

	var cmds []*ttnpb.MACCommand
	for r := bytes.NewReader(mac); r.Len() > 0; {
		cmd := &ttnpb.MACCommand{}
		if err := lorawan.DefaultMACCommands.ReadUplink(r, cmd); err != nil {
			logger.
				WithField("unmarshaled", len(cmds)).
				WithError(err).
				Warn("Failed to unmarshal MAC commands")
			break
		}
		cmds = append(cmds, cmd)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ns.deduplicationDone(ctx, up):
	}

	up.RxMetadata = acc.Accumulated()
	registerMergeMetadata(ctx, dev, up)

	var handleErr bool
	dev, err = ns.devices.SetByID(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers, dev.EndDeviceIdentifiers.DeviceID,
		[]string{
			"default_mac_parameters",
			"downlink_margin",
			"frequency_plan_id",
			"last_dev_status_received_at",
			"lorawan_version",
			"mac_settings",
			"mac_state",
			"pending_session",
			"recent_uplinks",
			"resets_f_cnt",
			"session",
			"supports_join",
		},
		func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if stored == nil {
				logger.Warn("Device deleted during uplink handling, dropping...")
				handleErr = true
				return nil, nil, errOutdatedData
			}

			var paths []string

			storedSes := stored.Session
			if ses != dev.Session {
				storedSes = stored.PendingSession
			}

			if storedSes.GetSessionKeyID() != ses.SessionKeyID {
				logger.Warn("Device changed session during uplink handling, dropping...")
				handleErr = true
				return nil, nil, errOutdatedData
			}
			if storedSes.GetLastFCntUp() > ses.LastFCntUp && !stored.ResetsFCnt {
				logger.WithFields(log.Fields(
					"stored_f_cnt", storedSes.GetLastFCntUp(),
					"got_f_cnt", ses.LastFCntUp,
				)).Warn("A more recent uplink was received by device during uplink handling, dropping...")
				handleErr = true
				return nil, nil, errOutdatedData
			}

			if ses == dev.Session {
				stored.Session = ses
				paths = append(paths, "session")
			} else {
				stored.PendingSession = ses
				paths = append(paths, "pending_session")
			}

			stored.RecentUplinks = append(stored.RecentUplinks, up)
			if len(stored.RecentUplinks) >= recentUplinkCount {
				stored.RecentUplinks = stored.RecentUplinks[len(stored.RecentUplinks)-recentUplinkCount+1:]
			}
			paths = append(paths, "recent_uplinks")

			if stored.MACState != nil {
				stored.MACState.PendingApplicationDownlink = nil
			} else if err := resetMACState(ns.Component.FrequencyPlans, stored); err != nil {
				handleErr = true
				return nil, nil, err
			}
			paths = append(paths, "mac_state")

			stored.MACState.QueuedResponses = stored.MACState.QueuedResponses[:0]

		outer:
			for len(cmds) > 0 {
				cmd, cmds := cmds[0], cmds[1:]
				switch cmd.CID {
				case ttnpb.CID_RESET:
					err = handleResetInd(ctx, stored, cmd.GetResetInd(), ns.Component.FrequencyPlans)
				case ttnpb.CID_LINK_CHECK:
					err = handleLinkCheckReq(ctx, stored, up)
				case ttnpb.CID_LINK_ADR:
					pld := cmd.GetLinkADRAns()
					dupCount := 0
					if stored.MACState.LoRaWANVersion == ttnpb.MAC_V1_0_2 {
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
					}
					if err != nil {
						break
					}
					cmds = cmds[dupCount:]
					err = handleLinkADRAns(ctx, stored, pld, uint(dupCount), ns.Component.FrequencyPlans)
				case ttnpb.CID_DUTY_CYCLE:
					err = handleDutyCycleAns(ctx, stored)
				case ttnpb.CID_RX_PARAM_SETUP:
					err = handleRxParamSetupAns(ctx, stored, cmd.GetRxParamSetupAns())
				case ttnpb.CID_DEV_STATUS:
					err = handleDevStatusAns(ctx, stored, cmd.GetDevStatusAns(), ses.LastFCntUp, up.ReceivedAt)
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
						logger.WithField("cid", cmd.CID).Warn("Unknown MAC command received, skipping the rest...")
						break outer
					}
					err = h.(MACHandler)(ctx, stored, cmd.GetRawPayload(), up)
				}
				if err != nil {
					logger.WithField("cid", cmd.CID).WithError(err).Warn("Failed to process MAC command")
					handleErr = true
					return nil, nil, err
				}
			}
			if stored.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
				stored.Session = ses
				stored.PendingSession = nil
			} else if stored.PendingSession != nil {
				handleErr = true
				return nil, nil, errNoRekey
			}
			paths = append(paths,
				"pending_session",
				"session",
			)

			if stored.Session != ses {
				// Sanity check
				panic(fmt.Errorf("Session mismatch"))
			}
			stored.MACState.RxWindowsAvailable = true

			if !pld.FHDR.ADR {
				dev.RecentADRUplinks = nil
				return stored, paths, nil
			}

			dev.RecentADRUplinks = append(dev.RecentADRUplinks, up)
			if len(dev.RecentADRUplinks) > recentDownlinkCount {
				dev.RecentUplinks = append(dev.RecentUplinks[:0], dev.RecentUplinks[len(dev.RecentUplinks)-recentDownlinkCount:]...)
			}
			paths = append(paths, "recent_adr_uplinks")

			if err := adaptDataRate(ns.Component.FrequencyPlans, dev); err != nil {
				handleErr = true
				return nil, nil, err
			}
			return stored, paths, nil
		})
	if err != nil && !handleErr {
		logger.WithError(err).Error("Failed to update device in registry")
		// TODO: Retry transaction. (https://github.com/TheThingsIndustries/lorawan-stack/issues/1163)
	}
	if err != nil {
		return err
	}

	scheduleAt := time.Now().UTC()
	ok, err := ns.handleASUplink(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers, &ttnpb.ApplicationUp{
		EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
		CorrelationIDs:       up.CorrelationIDs,
		Up: &ttnpb.ApplicationUp_UplinkMessage{UplinkMessage: &ttnpb.ApplicationUplink{
			FCnt:         dev.Session.LastFCntUp,
			FPort:        pld.FPort,
			FRMPayload:   pld.FRMPayload,
			RxMetadata:   up.RxMetadata,
			SessionKeyID: dev.Session.SessionKeyID,
			Settings:     up.Settings,
		}},
	})
	if err != nil {
		logger.WithError(err).Error("Failed to forward uplink to AS")
	} else if !ok {
		logger.Warn("Application Server not found, not forwarding uplink")
	} else {
		registerForwardUplink(ctx, dev, up)
		scheduleAt = scheduleAt.Add(appQueueUpdateTimeout)
	}
	return ns.downlinkTasks.Add(ctx, dev.EndDeviceIdentifiers, scheduleAt)
}

// newDevAddr generates a DevAddr for specified EndDevice.
func (ns *NetworkServer) newDevAddr(context.Context, *ttnpb.EndDevice) types.DevAddr {
	nwkAddr := make([]byte, types.NwkAddrLength(ns.NetID))
	random.Read(nwkAddr)
	nwkAddr[0] &= 0xff >> (8 - types.NwkAddrBits(ns.NetID)%8)
	devAddr, err := types.NewDevAddr(ns.NetID, nwkAddr)
	if err != nil {
		panic(errors.New("failed to create new DevAddr").WithCause(err))
	}
	return devAddr
}

func (ns *NetworkServer) handleJoin(ctx context.Context, up *ttnpb.UplinkMessage, acc *metadataAccumulator) (err error) {
	defer func() {
		if err != nil {
			registerDropUplink(ctx, up, err)
		}
	}()

	pld := up.Payload.GetJoinRequestPayload()

	logger := log.FromContext(ctx).WithFields(log.Fields(
		"dev_eui", pld.DevEUI,
		"join_eui", pld.JoinEUI,
	))

	dev, err := ns.devices.GetByEUI(ctx, pld.JoinEUI, pld.DevEUI,
		[]string{
			"frequency_plan_id",
			"lorawan_version",
			"mac_state",
			"session",
		},
	)
	if err != nil {
		logger.WithError(err).Error("Failed to load device from registry")
		return err
	}

	devAddr := ns.newDevAddr(ctx, dev)
	for dev.Session != nil && devAddr.Equal(dev.Session.DevAddr) {
		devAddr = ns.newDevAddr(ctx, dev)
	}

	fp, err := ns.FrequencyPlans.GetByID(dev.FrequencyPlanID)
	if err != nil {
		return errUnknownFrequencyPlan.WithCause(err)
	}

	req := &ttnpb.JoinRequest{
		RawPayload: up.RawPayload,
		Payload:    up.Payload,
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			DevEUI:  &pld.DevEUI,
			JoinEUI: &pld.JoinEUI,
			DevAddr: &devAddr,
		},
		NetID:              ns.NetID,
		SelectedMACVersion: dev.LoRaWANVersion, // Assume NS version is always higher than the version of the device
		RxDelay:            dev.MACState.DesiredParameters.Rx1Delay,
		CFList:             frequencyplans.CFList(*fp, dev.LoRaWANPHYVersion),
		DownlinkSettings: ttnpb.DLSettings{
			Rx1DROffset: dev.MACState.DesiredParameters.Rx1DataRateOffset,
			Rx2DR:       dev.MACState.DesiredParameters.Rx2DataRateIndex,
			OptNeg:      true,
		},
	}

	js, err := ns.jsClient(ctx, dev.EndDeviceIdentifiers)
	if err != nil {
		logger.WithError(err).Debug("Could not get Join Server")
		return err
	}

	resp, err := js.HandleJoin(ctx, req, ns.WithClusterAuth())
	if err != nil {
		logger.WithError(err).Warn("Join Server failed to handle join-request")
		return err
	}
	registerForwardUplink(ctx, dev, up)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ns.deduplicationDone(ctx, up):
	}

	up.RxMetadata = acc.Accumulated()
	registerMergeMetadata(ctx, dev, up)

	var invalidatedQueue []*ttnpb.ApplicationDownlink
	var resetErr bool
	dev, err = ns.devices.SetByID(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers, dev.EndDeviceIdentifiers.DeviceID,
		[]string{
			"default_mac_parameters",
			"frequency_plan_id",
			"lorawan_version",
			"mac_state",
			"queued_application_downlinks",
			"recent_uplinks",
			"session",
		},
		func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			paths := make([]string, 0, 5)

			dev.Session = &ttnpb.Session{
				DevAddr:     devAddr,
				SessionKeys: resp.SessionKeys,
				StartedAt:   time.Now().UTC(),
			}
			paths = append(paths, "session")

			dev.EndDeviceIdentifiers.DevAddr = &devAddr
			paths = append(paths, "ids.dev_addr")

			if err := resetMACState(ns.Component.FrequencyPlans, dev); err != nil {
				resetErr = true
				return nil, nil, err
			}

			dev.MACState.RxWindowsAvailable = true
			dev.MACState.QueuedJoinAccept = resp.RawPayload
			if req.DownlinkSettings.OptNeg && dev.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) > 0 {
				// The version will be further negotiated via RekeyInd/RekeyConf
				dev.MACState.LoRaWANVersion = ttnpb.MAC_V1_1
			}
			dev.MACState.CurrentParameters.Rx1Delay = req.RxDelay
			dev.MACState.CurrentParameters.Rx1DataRateOffset = req.DownlinkSettings.Rx1DROffset
			dev.MACState.CurrentParameters.Rx2DataRateIndex = req.DownlinkSettings.Rx2DR
			dev.MACState.DesiredParameters.Rx1Delay = dev.MACState.CurrentParameters.Rx1Delay
			dev.MACState.DesiredParameters.Rx1DataRateOffset = dev.MACState.CurrentParameters.Rx1DataRateOffset
			dev.MACState.DesiredParameters.Rx2DataRateIndex = dev.MACState.CurrentParameters.Rx2DataRateIndex
			paths = append(paths, "mac_state")

			dev.RecentUplinks = append(dev.RecentUplinks, up)
			if len(dev.RecentUplinks) > recentUplinkCount {
				dev.RecentUplinks = append(dev.RecentUplinks[:0], dev.RecentUplinks[len(dev.RecentUplinks)-recentUplinkCount:]...)
			}
			paths = append(paths, "recent_uplinks")

			invalidatedQueue = dev.QueuedApplicationDownlinks
			dev.QueuedApplicationDownlinks = nil
			paths = append(paths, "queued_application_downlinks")

			return dev, paths, nil
		})
	if err != nil && !resetErr {
		logger.WithError(err).Error("Failed to update device in registry")
		// TODO: Retry transaction. (https://github.com/TheThingsIndustries/lorawan-stack/issues/1163)
	}
	if err != nil {
		return err
	}

	if err := ns.downlinkTasks.Add(ctx, dev.EndDeviceIdentifiers, time.Now().UTC()); err != nil {
		return err
	}

	logger = logger.WithField(
		"application_uid", unique.ID(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers),
	)

	logger.Debug("Sending join-accept to AS...")
	_, err = ns.handleASUplink(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers, &ttnpb.ApplicationUp{
		EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
		CorrelationIDs:       up.CorrelationIDs,
		Up: &ttnpb.ApplicationUp_JoinAccept{JoinAccept: &ttnpb.ApplicationJoinAccept{
			AppSKey:              resp.SessionKeys.AppSKey,
			InvalidatedDownlinks: invalidatedQueue,
			SessionKeyID:         dev.Session.SessionKeyID,
			SessionStartedAt:     dev.Session.StartedAt,
		}},
	})
	if err != nil {
		logger.WithError(err).Errorf("Failed to send join-accept to AS")
		return err
	}

	return nil
}

func (ns *NetworkServer) handleRejoin(ctx context.Context, up *ttnpb.UplinkMessage, acc *metadataAccumulator) (err error) {
	defer func() {
		if err != nil {
			registerDropUplink(ctx, up, err)
		}
	}()
	// TODO: Implement https://github.com/TheThingsIndustries/ttn/issues/557
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

	logger := log.FromContext(ctx)

	if up.Payload.Payload == nil {
		if err := lorawan.UnmarshalMessage(up.RawPayload, up.Payload); err != nil {
			return nil, errDecodePayload.WithCause(err)
		}
	}

	if up.Payload.Major != ttnpb.Major_LORAWAN_R1 {
		return nil, errUnsupportedLoRaWANVersion.WithAttributes(
			"major", up.Payload.Major,
		)
	}

	acc, stopDedup, ok := ns.deduplicateUplink(ctx, up)
	if ok {
		registerReceiveUplinkDuplicate(ctx, up)
		return ttnpb.Empty, nil
	}
	registerReceiveUplink(ctx, up)

	defer func(up *ttnpb.UplinkMessage) {
		<-ns.collectionDone(ctx, up)
		stopDedup()
	}(up)

	up = deepcopy.Copy(up).(*ttnpb.UplinkMessage)
	switch up.Payload.MType {
	case ttnpb.MType_CONFIRMED_UP, ttnpb.MType_UNCONFIRMED_UP:
		return ttnpb.Empty, ns.handleUplink(ctx, up, acc)
	case ttnpb.MType_JOIN_REQUEST:
		return ttnpb.Empty, ns.handleJoin(ctx, up, acc)
	case ttnpb.MType_REJOIN_REQUEST:
		return ttnpb.Empty, ns.handleRejoin(ctx, up, acc)
	default:
		logger.Error("Unmatched MType")
		return ttnpb.Empty, nil
	}
}
