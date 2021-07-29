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

	pbtypes "github.com/gogo/protobuf/types"
	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/toa"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

const (
	// recentUplinkCount is the maximum amount of recent uplinks stored per device.
	recentUplinkCount = 20

	// retransmissionWindow is the maximum delay between Rx2 end and an uplink retransmission.
	retransmissionWindow = 10 * time.Second

	// maxConfNbTrans is the maximum number of confirmed uplink retransmissions for pre-1.0.4 devices.
	maxConfNbTrans = 5
)

// UplinkDeduplicator represents an entity, that deduplicates uplinks and accumulates metadata.
type UplinkDeduplicator interface {
	// DeduplicateUplink deduplicates an uplink message for specified time.Duration.
	// DeduplicateUplink returns true if the uplink is not a duplicate or false and error, if any, otherwise.
	DeduplicateUplink(context.Context, *ttnpb.UplinkMessage, time.Duration) (bool, error)
	// AccumulatedMetadata returns accumulated metadata for specified uplink message and error, if any.
	AccumulatedMetadata(context.Context, *ttnpb.UplinkMessage) ([]*ttnpb.RxMetadata, error)
}

func (ns *NetworkServer) deduplicateUplink(ctx context.Context, up *ttnpb.UplinkMessage) (bool, error) {
	ok, err := ns.uplinkDeduplicator.DeduplicateUplink(ctx, up, ns.collectionWindow(ctx))
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to deduplicate uplink")
		return false, err
	}
	if !ok {
		log.FromContext(ctx).Debug("Dropped duplicate uplink")
		return false, nil
	}
	return true, nil
}

func maxTransmissionNumber(ver ttnpb.MACVersion, confirmed bool, nbTrans uint32) uint32 {
	if !confirmed {
		return nbTrans
	}
	if ver.Compare(ttnpb.MAC_V1_0_4) < 0 {
		return maxConfNbTrans
	}
	return nbTrans
}

func maxRetransmissionDelay(rxDelay ttnpb.RxDelay) time.Duration {
	return rxDelay.Duration() + time.Second + retransmissionWindow
}

func matchCmacF(ctx context.Context, fNwkSIntKey types.AES128Key, macVersion ttnpb.MACVersion, fCnt uint32, up *ttnpb.UplinkMessage) ([4]byte, bool) {
	registerMICComputation(ctx)
	cmacF, err := crypto.ComputeLegacyUplinkMIC(fNwkSIntKey, up.Payload.GetMACPayload().DevAddr, fCnt, up.RawPayload[:len(up.RawPayload)-4])
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to compute cmacF")
		return [4]byte{}, false
	}
	var micMatch bool
	if macVersion.UseLegacyMIC() {
		micMatch = bytes.Equal(up.Payload.MIC, cmacF[:])
	} else {
		micMatch = bytes.Equal(up.Payload.MIC[2:], cmacF[:2])
	}
	if !micMatch {
		registerMICMismatch(ctx)
		return [4]byte{}, false
	}
	return cmacF, true
}

type cmacFMatchingResult struct {
	LastFCnt       uint32
	IsPending      bool
	FNwkSIntKey    types.AES128Key
	LoRaWANVersion ttnpb.MACVersion
	FullFCnt       uint32
	CmacF          [4]byte
}

type macHandler func(context.Context, *ttnpb.EndDevice, *ttnpb.UplinkMessage) (events.Builders, error)

func makeDeferredMACHandler(dev *ttnpb.EndDevice, f macHandler) macHandler {
	queuedLength := len(dev.MACState.QueuedResponses)
	return func(ctx context.Context, dev *ttnpb.EndDevice, up *ttnpb.UplinkMessage) (events.Builders, error) {
		switch n := len(dev.MACState.QueuedResponses); {
		case n < queuedLength:
			return nil, errCorruptedMACState.New()
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

type matchResult struct {
	cmacFMatchingResult

	phy *band.Band

	Context                  context.Context
	Device                   *ttnpb.EndDevice
	ChannelIndex             uint8
	DataRateIndex            ttnpb.DataRateIndex
	DeferredMACHandlers      []macHandler
	IsRetransmission         bool
	QueuedApplicationUplinks []*ttnpb.ApplicationUp
	QueuedEventBuilders      events.Builders
	SetPaths                 []string
}

func applyCFList(cfList *ttnpb.CFList, phy *band.Band, chs ...*ttnpb.MACParameters_Channel) ([]*ttnpb.MACParameters_Channel, bool) {
	if cfList == nil {
		return chs, true
	}
	switch cfList.Type {
	case ttnpb.CFListType_FREQUENCIES:
		for _, freq := range cfList.Freq {
			if freq == 0 {
				break
			}
			chs = append(chs, &ttnpb.MACParameters_Channel{
				UplinkFrequency:   uint64(freq) * phy.FreqMultiplier,
				DownlinkFrequency: uint64(freq) * phy.FreqMultiplier,
				MaxDataRateIndex:  phy.MaxADRDataRateIndex,
				EnableUplink:      true,
			})
		}

	case ttnpb.CFListType_CHANNEL_MASKS:
		if len(chs) != len(cfList.ChMasks) {
			return nil, false
		}
		for i, m := range cfList.ChMasks {
			if m {
				continue
			}
			if chs[i] == nil {
				return nil, false
			}
			chs[i].EnableUplink = m
		}
	}
	return chs, true
}

// matchAndHandleDataUplink handles and matches a device prematched by CMACF check.
func (ns *NetworkServer) matchAndHandleDataUplink(ctx context.Context, dev *ttnpb.EndDevice, up *ttnpb.UplinkMessage, deduplicated bool, cmacFMatchResult cmacFMatchingResult) (*matchResult, bool, error) {
	phy, err := DeviceBand(dev, ns.FrequencyPlans)
	if err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to get device's versioned band")
		return nil, false, nil
	}
	drIdx, dr, ok := phy.FindUplinkDataRate(up.Settings.DataRate)
	if !ok {
		log.FromContext(ctx).Debug("Data rate not found in PHY")
		return nil, false, nil
	}
	ctx = log.NewContextWithFields(ctx, log.Fields(
		"band_id", phy.ID,
		"data_rate_index", up.Settings.DataRateIndex,
	))

	pld := up.Payload.GetMACPayload()
	pendingAppDown := dev.MACState.GetPendingApplicationDownlink()

	// NOTE: Device might have changed session since the CMACF match.
	// E.g. We could have matched pending session by CMACF and device might
	// have activated it meanwhile and made that matched session current.
	type sessionMatchType uint8
	const (
		currentOriginalMatch sessionMatchType = iota
		currentRetransmissionMatch
		currentResetMatch
		pendingMatch
	)
	var matchType sessionMatchType
	switch {
	// Pending session match
	case !pld.Ack &&
		cmacFMatchResult.IsPending &&
		dev.PendingSession != nil &&
		dev.PendingMACState != nil &&
		pld.DevAddr.Equal(dev.PendingSession.DevAddr) &&
		cmacFMatchResult.LoRaWANVersion.UseLegacyMIC() == dev.PendingMACState.LorawanVersion.UseLegacyMIC():
		fNwkSIntKey, err := cryptoutil.UnwrapAES128Key(ctx, dev.PendingSession.FNwkSIntKey, ns.KeyVault)
		if err != nil {
			log.FromContext(ctx).WithError(err).WithField("kek_label", dev.PendingSession.FNwkSIntKey.KEKLabel).Warn("Failed to unwrap FNwkSIntKey")
			return nil, false, nil
		}
		if cmacFMatchResult.FNwkSIntKey.Equal(fNwkSIntKey) {
			ctx = log.NewContextWithField(ctx, "mac_version", dev.PendingMACState.LorawanVersion)
			if dev.PendingMACState.PendingJoinRequest == nil {
				log.FromContext(ctx).Warn("Pending join-request missing")
				return nil, false, nil
			}
			dev.PendingMACState.CurrentParameters.Rx1Delay = dev.PendingMACState.PendingJoinRequest.RxDelay
			dev.PendingMACState.CurrentParameters.Rx1DataRateOffset = dev.PendingMACState.PendingJoinRequest.DownlinkSettings.Rx1DROffset
			dev.PendingMACState.CurrentParameters.Rx2DataRateIndex = dev.PendingMACState.PendingJoinRequest.DownlinkSettings.Rx2DR
			if dev.PendingMACState.PendingJoinRequest.DownlinkSettings.OptNeg && dev.LorawanVersion.Compare(ttnpb.MAC_V1_1) >= 0 {
				// The version will be further negotiated via RekeyInd/RekeyConf
				dev.PendingMACState.LorawanVersion = ttnpb.MAC_V1_1
			}
			chs, ok := applyCFList(dev.PendingMACState.PendingJoinRequest.CFList, phy, dev.PendingMACState.CurrentParameters.Channels...)
			if !ok {
				log.FromContext(ctx).Debug("Failed to apply CFList")
				return nil, false, nil
			}
			dev.PendingMACState.CurrentParameters.Channels = chs

			dev.MACState = dev.PendingMACState
			dev.PendingSession.StartedAt = up.ReceivedAt

			matchType = pendingMatch
			break
		}
		// Key mismatch, attempt to match current session.
		fallthrough

	// Current session match
	case dev.Session != nil &&
		dev.MACState != nil &&
		pld.DevAddr.Equal(dev.Session.DevAddr) &&
		cmacFMatchResult.LoRaWANVersion.UseLegacyMIC() == dev.MACState.LorawanVersion.UseLegacyMIC() &&
		(cmacFMatchResult.FullFCnt == FullFCnt(uint16(pld.FCnt), dev.Session.LastFCntUp, mac.DeviceSupports32BitFCnt(dev, ns.defaultMACSettings)) ||
			cmacFMatchResult.FullFCnt == pld.FCnt):
		fNwkSIntKey, err := cryptoutil.UnwrapAES128Key(ctx, dev.Session.FNwkSIntKey, ns.KeyVault)
		if err != nil {
			log.FromContext(ctx).WithError(err).WithField("kek_label", dev.Session.FNwkSIntKey.KEKLabel).Warn("Failed to unwrap FNwkSIntKey")
			return nil, false, nil
		}
		if cmacFMatchResult.FNwkSIntKey.Equal(fNwkSIntKey) {
			ctx = log.NewContextWithFields(ctx, log.Fields(
				"last_f_cnt_up", dev.Session.LastFCntUp,
				"mac_version", dev.MACState.LorawanVersion,
				"pending_session", false,
			))
			switch {
			case cmacFMatchResult.FullFCnt < dev.Session.LastFCntUp:
				if pld.Ack || dev.Session.LastFCntUp != cmacFMatchResult.LastFCnt || !mac.DeviceResetsFCnt(dev, ns.defaultMACSettings) {
					return nil, false, nil
				}
				ctx = log.NewContextWithField(ctx, "f_cnt_reset", true)

				macState, err := mac.NewState(dev, ns.FrequencyPlans, ns.defaultMACSettings)
				if err != nil {
					log.FromContext(ctx).WithError(err).Warn("Failed to generate new MAC state")
					return nil, false, nil
				}

				dev.MACState = macState
				dev.Session.StartedAt = up.ReceivedAt

				matchType = currentResetMatch

			case cmacFMatchResult.FullFCnt > dev.Session.LastFCntUp,
				dev.Session.LastFCntUp == 0 && dev.SupportsJoin && len(dev.MACState.RecentUplinks) == 1,
				dev.Session.LastFCntUp == 0 && !dev.SupportsJoin && len(dev.MACState.RecentUplinks) == 0:
				ctx = log.NewContextWithField(ctx, "f_cnt_reset", false)

				fCntGap := cmacFMatchResult.FullFCnt - dev.Session.LastFCntUp
				if dev.MACState.LorawanVersion.HasMaxFCntGap() && uint(fCntGap) > phy.MaxFCntGap {
					log.FromContext(ctx).WithFields(log.Fields(
						"f_cnt_gap", fCntGap,
						"max_f_cnt_gap", phy.MaxFCntGap,
					)).Debug("FCnt gap exceeds maximum after reset")
					return nil, false, nil
				}

				matchType = currentOriginalMatch

			default: // cmacFMatchResult.FullFCnt == dev.Session.LastFCntUp
				ctx = log.NewContextWithField(ctx, "f_cnt_reset", false)

				maxNbTrans := maxTransmissionNumber(dev.MACState.LorawanVersion, up.Payload.MType == ttnpb.MType_CONFIRMED_UP, dev.MACState.CurrentParameters.AdrNbTrans)
				if maxNbTrans < 1 {
					panic(fmt.Sprintf("invalid maximum transmission number %d", maxNbTrans))
				}
				ctx = log.NewContextWithField(ctx, "max_transmissions", maxNbTrans)

				nbTrans := uint32(1)
				var (
					lastAt             time.Time
					recentUpPHYPayload []byte
				)
				for i := len(dev.MACState.RecentUplinks) - 1; i >= 0; i-- {
					recentUp := dev.MACState.RecentUplinks[i]
					recentUpPHYPayload, err = lorawan.AppendMessage(recentUpPHYPayload[:0], *recentUp.Payload)
					if err != nil {
						log.FromContext(ctx).WithError(err).Error("Failed to marshal recent uplink payload")
						return nil, false, nil
					}
					if len(recentUpPHYPayload) < 4 {
						log.FromContext(ctx).Error("Length of marshaled recent uplink payload is too short")
						return nil, false, nil
					}
					if !bytes.Equal(up.RawPayload[:len(up.RawPayload)-4], recentUpPHYPayload[:len(recentUpPHYPayload)-4]) {
						break
					}
					if nbTrans >= maxNbTrans {
						log.FromContext(ctx).Info("Transmission number exceeds maximum")
						return nil, false, nil
					}
					nbTrans++
					if recentUp.ReceivedAt.After(lastAt) {
						lastAt = recentUp.ReceivedAt
					}
				}
				if nbTrans < 2 || lastAt.IsZero() {
					log.FromContext(ctx).Debug("Repeated FCnt value, but frame is not a retransmission")
					return nil, false, nil
				}
				maxDelay := maxRetransmissionDelay(dev.MACState.CurrentParameters.Rx1Delay)
				delay := up.ReceivedAt.Sub(lastAt)
				ctx = log.NewContextWithFields(ctx, log.Fields(
					"last_transmission_at", lastAt,
					"max_retransmission_delay", maxDelay,
					"retransmission_delay", delay,
					"retransmission_number", nbTrans,
				))
				if delay > maxDelay {
					log.FromContext(ctx).Warn("Retransmission delay exceeds maximum")
					return nil, false, nil
				}

				matchType = currentRetransmissionMatch
			}
			break
		}
		// Key mismatch
		return nil, false, nil

	default:
		return nil, false, nil
	}

	// NOTE: We assume no dwell time if current value unknown.
	if dev.MACState.LorawanVersion.IgnoreUplinksExceedingLengthLimit() && len(up.RawPayload)-5 > int(dr.MaxMACPayloadSize(dev.MACState.CurrentParameters.UplinkDwellTime.GetValue())) {
		log.FromContext(ctx).Debug("Uplink length exceeds maximum")
		return nil, false, nil
	}

	cmdBuf := pld.FOpts
	if pld.FPort == 0 && len(pld.FRMPayload) > 0 {
		cmdBuf = pld.FRMPayload
	}
	if len(cmdBuf) > 0 && (len(pld.FOpts) == 0 || dev.MACState.LorawanVersion.EncryptFOpts()) {
		session := dev.Session
		if matchType == pendingMatch {
			session = dev.PendingSession
		}
		if session.NwkSEncKey == nil || len(session.NwkSEncKey.Key) == 0 {
			log.FromContext(ctx).Warn("Device missing NwkSEncKey in registry")
			return nil, false, nil
		}
		key, err := cryptoutil.UnwrapAES128Key(ctx, session.NwkSEncKey, ns.KeyVault)
		if err != nil {
			log.FromContext(ctx).WithField("kek_label", session.NwkSEncKey.KEKLabel).WithError(err).Warn("Failed to unwrap NwkSEncKey")
			return nil, false, nil
		}
		cmdBuf, err = crypto.DecryptUplink(key, pld.DevAddr, cmacFMatchResult.FullFCnt, cmdBuf, len(pld.FOpts) > 0)
		if err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to decrypt uplink")
			return nil, false, nil
		}
	}

	logger := log.FromContext(ctx)
	if matchType == currentRetransmissionMatch {
		dev.MACState.PendingRequests = nil
	}
	var cmds []*ttnpb.MACCommand
	for r := bytes.NewReader(cmdBuf); r.Len() > 0; {
		cmd := &ttnpb.MACCommand{}
		if err := lorawan.DefaultMACCommands.ReadUplink(*phy, r, cmd); err != nil {
			log.FromContext(ctx).WithFields(log.Fields(
				"bytes_left", r.Len(),
				"mac_count", len(cmds),
			)).WithError(err).Warn("Failed to read MAC command")
			break
		}
		logger := logger.WithField("command", cmd)
		logger.Debug("Read MAC command")
		def, ok := lorawan.DefaultMACCommands[cmd.CID]
		if ok && !def.InitiatedByDevice {
			switch matchType {
			case currentResetMatch, pendingMatch:
				logger.Debug("Received MAC command answer after MAC state reset")
				return nil, false, nil

			case currentRetransmissionMatch:
				logger.Debug("Skip processing of MAC command not initiated by the device in a retransmission")
				continue
			}
		}
		cmds = append(cmds, cmd)
	}
	logger = logger.WithField("mac_count", len(cmds))
	ctx = log.NewContext(ctx, logger)

	var queuedEventBuilders []events.Builder
	if pld.ClassB {
		switch {
		case !dev.SupportsClassB:
			logger.Debug("Ignore class B bit in uplink, since device does not support class B")

		case dev.MACState.CurrentParameters.PingSlotFrequency == 0:
			logger.Debug("Ignore class B bit in uplink, since ping slot frequency is not known")

		case dev.MACState.CurrentParameters.PingSlotDataRateIndexValue == nil:
			logger.Debug("Ignore class B bit in uplink, since ping slot data rate index is not known")

		case dev.MACState.PingSlotPeriodicity == nil:
			logger.Debug("Ignore class B bit in uplink, since ping slot periodicity is not known")

		case dev.MACState.DeviceClass != ttnpb.CLASS_B:
			logger.WithField("previous_class", dev.MACState.DeviceClass).Debug("Switch device class to class B")
			queuedEventBuilders = append(queuedEventBuilders, mac.EvtClassBSwitch.BindData(dev.MACState.DeviceClass))
			dev.MACState.DeviceClass = ttnpb.CLASS_B
		}
	} else if dev.MACState.DeviceClass == ttnpb.CLASS_B {
		if dev.MACState.LorawanVersion.Compare(ttnpb.MAC_V1_1) < 0 && dev.SupportsClassC {
			queuedEventBuilders = append(queuedEventBuilders, mac.EvtClassCSwitch.BindData(ttnpb.CLASS_B))
			dev.MACState.DeviceClass = ttnpb.CLASS_C
		} else {
			queuedEventBuilders = append(queuedEventBuilders, mac.EvtClassASwitch.BindData(ttnpb.CLASS_B))
			dev.MACState.DeviceClass = ttnpb.CLASS_A
		}
	}

	var deferredMACHandlers []macHandler
	if len(cmds) > 0 && !deduplicated {
		deferredMACHandlers = make([]macHandler, 0, 1)
	}
	var setPaths []string
	dev.MACState.QueuedResponses = dev.MACState.QueuedResponses[:0]
macLoop:
	for len(cmds) > 0 {
		var cmd *ttnpb.MACCommand
		cmd, cmds = cmds[0], cmds[1:]
		logger := logger.WithField("command", cmd)
		logger.Debug("Handle MAC command")
		ctx := log.NewContext(ctx, logger)

		var evs events.Builders
		var err error
		switch cmd.CID {
		case ttnpb.CID_RESET:
			evs, err = mac.HandleResetInd(ctx, dev, cmd.GetResetInd(), ns.FrequencyPlans, ns.defaultMACSettings)
		case ttnpb.CID_LINK_CHECK:
			if !deduplicated {
				deferredMACHandlers = append(deferredMACHandlers, makeDeferredMACHandler(dev, mac.HandleLinkCheckReq))
				continue macLoop
			}
			evs, err = mac.HandleLinkCheckReq(ctx, dev, up)
		case ttnpb.CID_LINK_ADR:
			pld := cmd.GetLinkAdrAns()
			dupCount := 0
			if dev.MACState.LorawanVersion.Compare(ttnpb.MAC_V1_0_2) >= 0 && dev.MACState.LorawanVersion.Compare(ttnpb.MAC_V1_1) < 0 {
				for _, dup := range cmds {
					if dup.CID != ttnpb.CID_LINK_ADR {
						break
					}
					if !dup.GetLinkAdrAns().Equal(pld) {
						err = errInvalidPayload.New()
						break
					}
					dupCount++
				}
			}
			if err != nil {
				break
			}
			cmds = cmds[dupCount:]
			evs, err = mac.HandleLinkADRAns(ctx, dev, pld, uint(dupCount), cmacFMatchResult.FullFCnt, ns.FrequencyPlans)
		case ttnpb.CID_DUTY_CYCLE:
			evs, err = mac.HandleDutyCycleAns(ctx, dev)
		case ttnpb.CID_RX_PARAM_SETUP:
			evs, err = mac.HandleRxParamSetupAns(ctx, dev, cmd.GetRxParamSetupAns())
		case ttnpb.CID_DEV_STATUS:
			evs, err = mac.HandleDevStatusAns(ctx, dev, cmd.GetDevStatusAns(), cmacFMatchResult.FullFCnt, up.ReceivedAt)
			if err == nil {
				setPaths = append(setPaths,
					"battery_percentage",
					"downlink_margin",
					"last_dev_status_received_at",
					"power_state",
				)
			}
		case ttnpb.CID_NEW_CHANNEL:
			evs, err = mac.HandleNewChannelAns(ctx, dev, cmd.GetNewChannelAns())
		case ttnpb.CID_RX_TIMING_SETUP:
			evs, err = mac.HandleRxTimingSetupAns(ctx, dev)
		case ttnpb.CID_TX_PARAM_SETUP:
			evs, err = mac.HandleTxParamSetupAns(ctx, dev)
		case ttnpb.CID_DL_CHANNEL:
			evs, err = mac.HandleDLChannelAns(ctx, dev, cmd.GetDLChannelAns())
		case ttnpb.CID_REKEY:
			evs, err = mac.HandleRekeyInd(ctx, dev, cmd.GetRekeyInd(), pld.DevAddr)
		case ttnpb.CID_ADR_PARAM_SETUP:
			evs, err = mac.HandleADRParamSetupAns(ctx, dev)
		case ttnpb.CID_DEVICE_TIME:
			evs, err = mac.HandleDeviceTimeReq(ctx, dev, up)
		case ttnpb.CID_REJOIN_PARAM_SETUP:
			evs, err = mac.HandleRejoinParamSetupAns(ctx, dev, cmd.GetRejoinParamSetupAns())
		case ttnpb.CID_PING_SLOT_INFO:
			evs, err = mac.HandlePingSlotInfoReq(ctx, dev, cmd.GetPingSlotInfoReq())
		case ttnpb.CID_PING_SLOT_CHANNEL:
			evs, err = mac.HandlePingSlotChannelAns(ctx, dev, cmd.GetPingSlotChannelAns())
		case ttnpb.CID_BEACON_TIMING:
			evs, err = mac.HandleBeaconTimingReq(ctx, dev)
		case ttnpb.CID_BEACON_FREQ:
			evs, err = mac.HandleBeaconFreqAns(ctx, dev, cmd.GetBeaconFreqAns())
		case ttnpb.CID_DEVICE_MODE:
			evs, err = mac.HandleDeviceModeInd(ctx, dev, cmd.GetDeviceModeInd())
		default:
			logger.Warn("Unknown MAC command received, skip the rest")
			break macLoop
		}
		if err != nil {
			logger.WithError(err).Debug("Failed to process MAC command")
			break macLoop
		}
		queuedEventBuilders = append(queuedEventBuilders, evs...)
	}
	if n := len(dev.MACState.PendingRequests); n > 0 {
		logger.WithField("unanswered_request_count", n).Warn("MAC command buffer not fully answered")
		dev.MACState.PendingRequests = dev.MACState.PendingRequests[:0]
	}

	if matchType == pendingMatch {
		if dev.MACState.LorawanVersion.Compare(ttnpb.MAC_V1_1) < 0 {
			dev.EndDeviceIdentifiers.DevAddr = &pld.DevAddr
			dev.Session = dev.PendingSession
		} else if dev.PendingSession != nil || dev.PendingMACState != nil || dev.MACState.PendingJoinRequest != nil {
			logger.Debug("No RekeyInd received for LoRaWAN 1.1+ device")
			return nil, false, nil
		}
		setPaths = append(setPaths, "ids.dev_addr")
	} else if dev.PendingSession != nil || dev.PendingMACState != nil {
		// TODO: Notify AS of session recovery(https://github.com/TheThingsNetwork/lorawan-stack/issues/594)
	}
	dev.MACState.PendingJoinRequest = nil
	dev.PendingMACState = nil
	dev.PendingSession = nil

	chIdx, err := searchUplinkChannel(up.Settings.Frequency, dev.MACState)
	if err != nil {
		logger.WithError(err).Debug("Failed to determine channel index of uplink")
		return nil, false, nil
	}
	logger = logger.WithField("device_channel_index", chIdx)
	ctx = log.NewContext(ctx, logger)

	// NOTE: Legacy MIC check is already performed.
	if !dev.MACState.LorawanVersion.UseLegacyMIC() {
		sNwkSIntKey, err := cryptoutil.UnwrapAES128Key(ctx, dev.Session.SNwkSIntKey, ns.KeyVault)
		if err != nil {
			logger.WithField("kek_label", dev.Session.SNwkSIntKey.GetKEKLabel()).WithError(err).Warn("Failed to unwrap SNwkSIntKey")
			return nil, false, nil
		}

		var confFCnt uint32
		if pld.Ack {
			confFCnt = dev.Session.LastConfFCntDown
		}
		registerMICComputation(ctx)
		fullMIC, err := crypto.ComputeUplinkMICFromLegacy(
			cmacFMatchResult.CmacF,
			sNwkSIntKey,
			confFCnt,
			uint8(drIdx),
			chIdx,
			pld.DevAddr,
			cmacFMatchResult.FullFCnt,
			up.RawPayload[:len(up.RawPayload)-4],
		)
		if err != nil {
			logger.WithError(err).Error("Failed to compute 1.1 MIC")
			return nil, false, nil
		}
		if !bytes.Equal(up.Payload.MIC, fullMIC[:]) {
			logger.Debug("Full MIC mismatch")
			registerMICMismatch(ctx)
			return nil, false, nil
		}
	}
	dev.MACState.RxWindowsAvailable = true
	dev.Session.LastFCntUp = cmacFMatchResult.FullFCnt

	var queuedApplicationUplinks []*ttnpb.ApplicationUp
	if pendingAppDown != nil {
		if pld.Ack {
			queuedApplicationUplinks = []*ttnpb.ApplicationUp{
				{
					EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
					Up: &ttnpb.ApplicationUp_DownlinkAck{
						DownlinkAck: pendingAppDown,
					},
					CorrelationIDs: append(pendingAppDown.CorrelationIDs, up.CorrelationIDs...),
				},
			}
		} else {
			queuedApplicationUplinks = []*ttnpb.ApplicationUp{
				{
					EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
					Up: &ttnpb.ApplicationUp_DownlinkNack{
						DownlinkNack: pendingAppDown,
					},
					CorrelationIDs: append(pendingAppDown.CorrelationIDs, up.CorrelationIDs...),
				},
			}
		}
		if dev.MACState != nil {
			dev.MACState.PendingApplicationDownlink = nil
		}
	}
	return &matchResult{
		cmacFMatchingResult:      cmacFMatchResult,
		phy:                      phy,
		Context:                  ctx,
		Device:                   dev,
		ChannelIndex:             chIdx,
		DataRateIndex:            drIdx,
		DeferredMACHandlers:      deferredMACHandlers,
		IsRetransmission:         matchType == currentRetransmissionMatch,
		QueuedApplicationUplinks: queuedApplicationUplinks,
		QueuedEventBuilders:      queuedEventBuilders,
		SetPaths: append(setPaths,
			"mac_state",
			"pending_mac_state",
			"pending_session",
			"session",
		),
	}, true, nil
}

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
	"session",
	"supports_class_b",
	"supports_class_c",
	"supports_join",
}

// mergeMetadata merges the metadata collected for up.
// mergeMetadata mutates up.RxMetadata discarding any existing up.RxMetadata value.
// NOTE: Since events are published async we need ensure that up passed to an event earlier is not mutated,
// hence up is taken by value here.
func (ns *NetworkServer) mergeMetadata(ctx context.Context, up *ttnpb.UplinkMessage) {
	mds, err := ns.uplinkDeduplicator.AccumulatedMetadata(ctx, up)
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to merge metadata")
		return
	}
	up.RxMetadata = mds
	log.FromContext(ctx).WithField("metadata_count", len(up.RxMetadata)).Debug("Merged metadata")
	registerMergeMetadata(ctx, up)
}

func (ns *NetworkServer) handleDataUplink(ctx context.Context, up *ttnpb.UplinkMessage) (err error) {
	if len(up.RawPayload) < 4 {
		return errRawPayloadTooShort.New()
	}
	pld := up.Payload.GetMACPayload()
	ctx = log.NewContextWithFields(ctx, log.Fields(
		"ack", pld.Ack,
		"adr", pld.Adr,
		"adr_ack_req", pld.AdrAckReq,
		"class_b", pld.ClassB,
		"dev_addr", pld.DevAddr,
		"f_opts_len", len(pld.FOpts),
		"f_port", pld.FPort,
		"uplink_f_cnt", pld.FCnt,
	))

	var (
		matched *matchResult
		ok      bool
	)
	const matchTTL = time.Minute
	if err := ns.devices.RangeByUplinkMatches(ctx, up, matchTTL,
		func(ctx context.Context, match *UplinkMatch) (bool, error) {
			ctx = log.NewContextWithFields(ctx, log.Fields(
				"mac_version", match.LoRaWANVersion,
				"pending_session", match.IsPending,
			))

			fNwkSIntKey, err := cryptoutil.UnwrapAES128Key(ctx, match.FNwkSIntKey, ns.KeyVault)
			if err != nil {
				log.FromContext(ctx).WithError(err).WithField("kek_label", match.FNwkSIntKey.KEKLabel).Warn("Failed to unwrap FNwkSIntKey")
				return false, nil
			}
			fCnt := FullFCnt(uint16(pld.FCnt), match.LastFCnt, mac.DeviceSupports32BitFCnt(&ttnpb.EndDevice{
				MACSettings: &ttnpb.MACSettings{
					Supports32BitFCnt: match.Supports32BitFCnt,
				},
			}, ns.defaultMACSettings))

			var cmacF [4]byte
			cmacF, ok = matchCmacF(ctx, fNwkSIntKey, match.LoRaWANVersion, fCnt, up)
			if !ok && fCnt != pld.FCnt && !pld.Ack && !match.IsPending && mac.DeviceResetsFCnt(&ttnpb.EndDevice{
				MACSettings: &ttnpb.MACSettings{
					ResetsFCnt: match.ResetsFCnt,
				},
			}, ns.defaultMACSettings) {
				// FCnt reset
				fCnt = pld.FCnt
				cmacF, ok = matchCmacF(ctx, fNwkSIntKey, match.LoRaWANVersion, fCnt, up)
			}
			if !ok {
				return false, nil
			}

			ctx = log.NewContextWithField(ctx, "full_f_cnt_up", fCnt)
			dev, ctx, err := ns.devices.GetByID(ctx, match.ApplicationIdentifiers, match.DeviceID, handleDataUplinkGetPaths[:])
			if err != nil {
				log.FromContext(ctx).WithError(err).Warn("Failed to get device after cmacF matching")
				return match.LoRaWANVersion.UseLegacyMIC(), nil
			}
			matched, ok, err = ns.matchAndHandleDataUplink(ctx, dev, up, false, cmacFMatchingResult{
				LastFCnt:       match.LastFCnt,
				IsPending:      match.IsPending,
				FNwkSIntKey:    fNwkSIntKey,
				LoRaWANVersion: match.LoRaWANVersion,
				FullFCnt:       fCnt,
				CmacF:          cmacF,
			})
			if err != nil {
				return false, err
			}
			return ok || match.LoRaWANVersion.UseLegacyMIC(), nil
		},
	); err != nil {
		logRegistryRPCError(ctx, err, "Failed to find devices in registry by DevAddr")
		return errDeviceNotFound.WithCause(err)
	}
	if !ok {
		return errDeviceNotFound
	}

	pld.FullFCnt = matched.FullFCnt
	up.DeviceChannelIndex = uint32(matched.ChannelIndex)
	up.Settings.DataRateIndex = matched.DataRateIndex
	ctx = matched.Context

	queuedEvents := []events.Event{
		evtReceiveDataUplink.NewWithIdentifiersAndData(ctx, &matched.Device.EndDeviceIdentifiers, up),
	}
	defer func(ids ttnpb.EndDeviceIdentifiers) {
		if err != nil {
			queuedEvents = append(queuedEvents, evtDropDataUplink.NewWithIdentifiersAndData(ctx, &ids, err))
		}
		publishEvents(ctx, queuedEvents...)
	}(matched.Device.EndDeviceIdentifiers)

	ok, err = ns.deduplicateUplink(ctx, up)
	if err != nil {
		return err
	}
	if !ok {
		queuedEvents = append(queuedEvents, evtDropDataUplink.NewWithIdentifiersAndData(ctx, &matched.Device.EndDeviceIdentifiers, errDuplicate))
		registerReceiveDuplicateUplink(ctx, up)
		return nil
	}

	publishEvents(ctx, queuedEvents...)
	queuedEvents = nil
	up = CopyUplinkMessage(up)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ns.deduplicationDone(ctx, up):
	}
	ns.mergeMetadata(ctx, up)

	for _, f := range matched.DeferredMACHandlers {
		evs, err := f(ctx, matched.Device, up)
		if err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to process MAC command after deduplication")
			break
		}
		matched.QueuedEventBuilders = append(matched.QueuedEventBuilders, evs...)
	}

	var queuedApplicationUplinks []*ttnpb.ApplicationUp
	defer func() { ns.enqueueApplicationUplinks(ctx, queuedApplicationUplinks...) }()

	stored, _, err := ns.devices.SetByID(ctx, matched.Device.ApplicationIdentifiers, matched.Device.DeviceId, handleDataUplinkGetPaths[:],
		func(ctx context.Context, stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if stored == nil {
				log.FromContext(ctx).Warn("Device deleted during uplink handling, drop")
				return nil, nil, errOutdatedData.New()
			}

			if !matched.Device.CreatedAt.Equal(stored.CreatedAt) || !matched.Device.UpdatedAt.Equal(stored.UpdatedAt) {
				matched, ok, err = ns.matchAndHandleDataUplink(ctx, stored, up, true, matched.cmacFMatchingResult)
				if err != nil {
					return nil, nil, err
				}
				if !ok {
					return nil, nil, errOutdatedData.New()
				}
				pld.FullFCnt = matched.FullFCnt
				up.DeviceChannelIndex = uint32(matched.ChannelIndex)
				up.Settings.DataRateIndex = matched.DataRateIndex
				ctx = matched.Context
			}

			queuedApplicationUplinks = append(queuedApplicationUplinks, matched.QueuedApplicationUplinks...)
			queuedEvents = append(queuedEvents, matched.QueuedEventBuilders.New(ctx, events.WithIdentifiers(&matched.Device.EndDeviceIdentifiers))...)

			stored = matched.Device
			paths := ttnpb.AddFields(matched.SetPaths,
				"mac_state.recent_uplinks",
			)
			stored.MACState.RecentUplinks = appendRecentUplink(stored.MACState.RecentUplinks, &ttnpb.UplinkMessage{
				Payload:            up.Payload,
				Settings:           up.Settings,
				RxMetadata:         up.RxMetadata,
				ReceivedAt:         up.ReceivedAt,
				CorrelationIDs:     up.CorrelationIDs,
				DeviceChannelIndex: up.DeviceChannelIndex,
				ConsumedAirtime:    up.ConsumedAirtime,
			}, recentUplinkCount)

			if up.Settings.DataRateIndex < stored.MACState.CurrentParameters.AdrDataRateIndex {
				// Device lowers TX power index before lowering data rate index according to the spec.
				stored.MACState.CurrentParameters.AdrTxPowerIndex = 0
				paths = ttnpb.AddFields(paths,
					"mac_state.current_parameters.adr_tx_power_index",
				)
			}
			stored.MACState.CurrentParameters.AdrDataRateIndex = up.Settings.DataRateIndex
			paths = ttnpb.AddFields(paths,
				"mac_state.current_parameters.adr_data_rate_index",
			)

			useADR := mac.DeviceUseADR(stored, ns.defaultMACSettings, matched.phy)
			if useADR {
				paths = ttnpb.AddFields(paths,
					"mac_state.desired_parameters.adr_data_rate_index",
					"mac_state.desired_parameters.adr_nb_trans",
					"mac_state.desired_parameters.adr_tx_power_index",
				)
				stored.MACState.DesiredParameters.AdrDataRateIndex = stored.MACState.CurrentParameters.AdrDataRateIndex
				stored.MACState.DesiredParameters.AdrTxPowerIndex = stored.MACState.CurrentParameters.AdrTxPowerIndex
				stored.MACState.DesiredParameters.AdrNbTrans = stored.MACState.CurrentParameters.AdrNbTrans
			}
			if !pld.FHDR.Adr || !useADR {
				return stored, paths, nil
			}
			if err := mac.AdaptDataRate(ctx, stored, matched.phy, ns.defaultMACSettings); err != nil {
				log.FromContext(ctx).WithError(err).Info("Failed to adapt data rate, avoid ADR")
			}
			return stored, paths, nil
		})
	if err != nil {
		// TODO: Retry transaction. (https://github.com/TheThingsNetwork/lorawan-stack/issues/33)
		logRegistryRPCError(ctx, err, "Failed to update device in registry")
		return err
	}
	matched.Device = stored
	ctx = matched.Context

	if err := ns.updateDataDownlinkTask(ctx, stored, time.Time{}); err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to update downlink task queue after data uplink")
	}
	if !matched.IsRetransmission {
		queuedApplicationUplinks = append(queuedApplicationUplinks, &ttnpb.ApplicationUp{
			EndDeviceIdentifiers: stored.EndDeviceIdentifiers,
			CorrelationIDs:       up.CorrelationIDs,
			Up: &ttnpb.ApplicationUp_UplinkMessage{
				UplinkMessage: &ttnpb.ApplicationUplink{
					Confirmed:       up.Payload.MType == ttnpb.MType_CONFIRMED_UP,
					FCnt:            pld.FullFCnt,
					FPort:           pld.FPort,
					FRMPayload:      pld.FRMPayload,
					RxMetadata:      up.RxMetadata,
					SessionKeyID:    stored.Session.SessionKeyID,
					Settings:        up.Settings,
					ReceivedAt:      up.ReceivedAt,
					ConsumedAirtime: up.ConsumedAirtime,
					NetworkIds: &ttnpb.NetworkIdentifiers{
						NetId:     &ns.netID,
						ClusterId: ns.clusterID,
					},
				},
			},
		})
	}
	queuedEvents = append(queuedEvents, evtProcessDataUplink.NewWithIdentifiersAndData(ctx, &matched.Device.EndDeviceIdentifiers, up))
	registerProcessUplink(ctx, up)
	return nil
}

func joinResponseWithoutKeys(resp *ttnpb.JoinResponse) *ttnpb.JoinResponse {
	return &ttnpb.JoinResponse{
		RawPayload: resp.RawPayload,
		SessionKeys: ttnpb.SessionKeys{
			SessionKeyID: resp.SessionKeys.SessionKeyID,
		},
		Lifetime:       resp.Lifetime,
		CorrelationIDs: resp.CorrelationIDs,
	}
}

func (ns *NetworkServer) sendJoinRequest(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, req *ttnpb.JoinRequest) (*ttnpb.JoinResponse, []events.Event, error) {
	var queuedEvents []events.Event
	logger := log.FromContext(ctx)
	cc, err := ns.GetPeerConn(ctx, ttnpb.ClusterRole_JOIN_SERVER, &ids)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.WithError(err).Debug("Join Server peer not found")
		} else {
			logger.WithError(err).Error("Join Server peer connection lookup failed")
		}
	} else {
		queuedEvents = append(queuedEvents, evtClusterJoinAttempt.NewWithIdentifiersAndData(ctx, &ids, req))
		resp, err := ttnpb.NewNsJsClient(cc).HandleJoin(ctx, req, ns.WithClusterAuth())
		if err == nil {
			logger.Debug("Join-request accepted by cluster-local Join Server")
			queuedEvents = append(queuedEvents, evtClusterJoinSuccess.NewWithIdentifiersAndData(ctx, &ids, joinResponseWithoutKeys(resp)))
			return resp, queuedEvents, nil
		}
		logger.WithError(err).Info("Cluster-local Join Server did not accept join-request")
		queuedEvents = append(queuedEvents, evtClusterJoinFail.NewWithIdentifiersAndData(ctx, &ids, err))
		if !errors.IsNotFound(err) {
			return nil, queuedEvents, err
		}
	}
	if ns.interopClient != nil {
		queuedEvents = append(queuedEvents, evtInteropJoinAttempt.NewWithIdentifiersAndData(ctx, &ids, req))
		resp, err := ns.interopClient.HandleJoinRequest(ctx, ns.netID, req)
		if err == nil {
			logger.Debug("Join-request accepted by interop Join Server")
			queuedEvents = append(queuedEvents, evtInteropJoinSuccess.NewWithIdentifiersAndData(ctx, &ids, joinResponseWithoutKeys(resp)))
			return resp, queuedEvents, nil
		}
		logger.WithError(err).Warn("Interop Join Server did not accept join-request")
		queuedEvents = append(queuedEvents, evtInteropJoinFail.NewWithIdentifiersAndData(ctx, &ids, err))
		if !errors.IsNotFound(err) {
			return nil, queuedEvents, err
		}
	}
	return nil, queuedEvents, errJoinServerNotFound.New()
}

func (ns *NetworkServer) deduplicationDone(ctx context.Context, up *ttnpb.UplinkMessage) <-chan time.Time {
	return time.After(time.Until(up.ReceivedAt.Add(ns.deduplicationWindow(ctx))))
}

func (ns *NetworkServer) handleJoinRequest(ctx context.Context, up *ttnpb.UplinkMessage) (err error) {
	pld := up.Payload.GetJoinRequestPayload()
	ctx = log.NewContextWithFields(ctx, log.Fields(
		"dev_eui", pld.DevEui,
		"join_eui", pld.JoinEui,
	))

	matched, matchedCtx, err := ns.devices.GetByEUI(ctx, pld.JoinEui, pld.DevEui,
		[]string{
			"frequency_plan_id",
			"lorawan_phy_version",
			"lorawan_version",
			"mac_settings",
			"session.dev_addr",
			"supports_class_b",
			"supports_class_c",
			"supports_join",
		},
	)
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to load device from registry by EUIs")
		return err
	}
	ctx = matchedCtx
	ctx = log.NewContextWithField(ctx, "device_uid", unique.ID(ctx, matched.EndDeviceIdentifiers))

	queuedEvents := []events.Event{
		evtReceiveJoinRequest.NewWithIdentifiersAndData(ctx, &matched.EndDeviceIdentifiers, up),
	}
	defer func() {
		if err != nil {
			queuedEvents = append(queuedEvents, evtDropJoinRequest.NewWithIdentifiersAndData(ctx, &matched.EndDeviceIdentifiers, err))
		}
		publishEvents(ctx, queuedEvents...)
	}()

	if !matched.SupportsJoin {
		log.FromContext(ctx).Warn("ABP device sent a join-request, drop")
		queuedEvents = append(queuedEvents, evtDropJoinRequest.NewWithIdentifiersAndData(ctx, &matched.EndDeviceIdentifiers, errABPJoinRequest))
		return nil
	}

	fp, phy, err := DeviceFrequencyPlanAndBand(matched, ns.FrequencyPlans)
	if err != nil {
		return err
	}
	drIdx, _, ok := phy.FindUplinkDataRate(up.Settings.DataRate)
	if !ok {
		return errDataRateNotFound.New()
	}
	up.Settings.DataRateIndex = drIdx
	ctx = log.NewContextWithField(ctx,
		"data_rate_index", drIdx,
	)

	macState, err := mac.NewState(matched, ns.FrequencyPlans, ns.defaultMACSettings)
	if err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to reset device's MAC state")
		return err
	}

	chIdx, err := searchUplinkChannel(up.Settings.Frequency, macState)
	if err != nil {
		return err
	}
	up.DeviceChannelIndex = uint32(chIdx)
	ctx = log.NewContextWithField(ctx,
		"device_channel_index", drIdx,
	)

	ok, err = ns.deduplicateUplink(ctx, up)
	if err != nil {
		return err
	}
	if !ok {
		queuedEvents = append(queuedEvents, evtDropJoinRequest.NewWithIdentifiersAndData(ctx, &matched.EndDeviceIdentifiers, errDuplicate))
		registerReceiveDuplicateUplink(ctx, up)
		return nil
	}

	devAddr := ns.newDevAddr(ctx, matched)
	const maxDevAddrGenerationRetries = 5
	for i := 0; i < maxDevAddrGenerationRetries && matched.Session != nil && devAddr.Equal(matched.Session.DevAddr); i++ {
		devAddr = ns.newDevAddr(ctx, matched)
	}
	ctx = log.NewContextWithField(ctx, "dev_addr", devAddr)
	if matched.Session != nil && devAddr.Equal(matched.Session.DevAddr) {
		log.FromContext(ctx).Error("Reusing the DevAddr used for current session")
	}

	cfList := frequencyplans.CFList(*fp, matched.LorawanPhyVersion)
	dlSettings := ttnpb.DLSettings{
		Rx1DROffset: macState.DesiredParameters.Rx1DataRateOffset,
		Rx2DR:       macState.DesiredParameters.Rx2DataRateIndex,
		OptNeg:      matched.LorawanVersion.Compare(ttnpb.MAC_V1_1) >= 0,
	}
	resp, joinEvents, err := ns.sendJoinRequest(ctx, matched.EndDeviceIdentifiers, &ttnpb.JoinRequest{
		Payload:            up.Payload,
		CFList:             cfList,
		CorrelationIDs:     events.CorrelationIDsFromContext(ctx),
		DevAddr:            devAddr,
		NetId:              ns.netID,
		RawPayload:         up.RawPayload,
		RxDelay:            macState.DesiredParameters.Rx1Delay,
		SelectedMACVersion: matched.LorawanVersion, // Assume NS version is always higher than the version of the device
		DownlinkSettings:   dlSettings,
		ConsumedAirtime:    up.ConsumedAirtime,
	})
	queuedEvents = append(queuedEvents, joinEvents...)
	if err != nil {
		return err
	}
	registerForwardJoinRequest(ctx, up)

	keys := resp.SessionKeys
	if !dlSettings.OptNeg {
		keys.NwkSEncKey = keys.FNwkSIntKey
		keys.SNwkSIntKey = keys.FNwkSIntKey
	}
	macState.QueuedJoinAccept = &ttnpb.MACState_JoinAccept{
		CorrelationIDs: resp.CorrelationIDs,
		Keys:           keys,
		Payload:        resp.RawPayload,
		DevAddr:        devAddr,
		NetId:          ns.netID,
		Request: ttnpb.MACState_JoinRequest{
			RxDelay:          macState.DesiredParameters.Rx1Delay,
			CFList:           cfList,
			DownlinkSettings: dlSettings,
		},
	}
	macState.RxWindowsAvailable = true
	ctx = events.ContextWithCorrelationID(ctx, resp.CorrelationIDs...)

	publishEvents(ctx, queuedEvents...)
	queuedEvents = nil
	up = CopyUplinkMessage(up)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ns.deduplicationDone(ctx, up):
	}
	ns.mergeMetadata(ctx, up)
	macState.RecentUplinks = []*ttnpb.UplinkMessage{{
		Payload:            up.Payload,
		Settings:           up.Settings,
		RxMetadata:         up.RxMetadata,
		ReceivedAt:         up.ReceivedAt,
		CorrelationIDs:     up.CorrelationIDs,
		DeviceChannelIndex: up.DeviceChannelIndex,
		ConsumedAirtime:    up.ConsumedAirtime,
	}}

	logger := log.FromContext(ctx)
	stored, storedCtx, err := ns.devices.SetByID(ctx, matched.EndDeviceIdentifiers.ApplicationIdentifiers, matched.EndDeviceIdentifiers.DeviceId,
		[]string{
			"frequency_plan_id",
			"lorawan_phy_version",
			"pending_session.queued_application_downlinks",
			"session.queued_application_downlinks",
		},
		func(ctx context.Context, stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if stored == nil {
				logger.Warn("Device deleted during join-request handling, drop")
				return nil, nil, errOutdatedData.New()
			}
			stored.PendingMACState = macState
			return stored, []string{
				"pending_mac_state",
			}, nil
		})
	if err != nil {
		// TODO: Retry transaction. (https://github.com/TheThingsNetwork/lorawan-stack/issues/33)
		logRegistryRPCError(ctx, err, "Failed to update device in registry")
		return err
	}
	matched = stored
	ctx = storedCtx

	// TODO: Extract this into a utility function shared with mac.HandleRejoinRequest. (https://github.com/TheThingsNetwork/lorawan-stack/issues/8)
	downAt := up.ReceivedAt.Add(-infrastructureDelay/2 + phy.JoinAcceptDelay1 - macState.DesiredParameters.Rx1Delay.Duration()/2 - nsScheduleWindow())
	if earliestAt := time.Now().Add(nsScheduleWindow()); downAt.Before(earliestAt) {
		downAt = earliestAt
	}
	logger.WithField("start_at", downAt).Debug("Add downlink task")
	if err := ns.downlinkTasks.Add(ctx, stored.EndDeviceIdentifiers, downAt, true); err != nil {
		logger.WithError(err).Error("Failed to add downlink task after join-request")
	}
	queuedEvents = append(queuedEvents, evtProcessJoinRequest.NewWithIdentifiersAndData(ctx, &matched.EndDeviceIdentifiers, up))
	registerProcessUplink(ctx, up)
	return nil
}

var errRejoinRequest = errors.DefineUnimplemented("rejoin_request", "rejoin-request handling is not implemented")

func (ns *NetworkServer) handleRejoinRequest(ctx context.Context, up *ttnpb.UplinkMessage) error {
	// TODO: Implement https://github.com/TheThingsNetwork/lorawan-stack/issues/8
	return errRejoinRequest.New()
}

// HandleUplink is called by the Gateway Server when an uplink message arrives.
func (ns *NetworkServer) HandleUplink(ctx context.Context, up *ttnpb.UplinkMessage) (_ *pbtypes.Empty, err error) {
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}

	ctx = events.ContextWithCorrelationID(ctx, append(
		up.CorrelationIDs,
		fmt.Sprintf("ns:uplink:%s", events.NewCorrelationID()),
	)...)
	up.CorrelationIDs = events.CorrelationIDsFromContext(ctx)

	registerUplinkLatency(ctx, up)

	up.ReceivedAt = time.Now().UTC()

	up.Payload = &ttnpb.Message{}
	if err := lorawan.UnmarshalMessage(up.RawPayload, up.Payload); err != nil {
		return nil, errDecodePayload.WithCause(err)
	}
	registerReceiveUplink(ctx, up)
	defer func() {
		if err != nil {
			registerDropUplink(ctx, up, err)
		}
	}()
	if up.Payload.Major != ttnpb.Major_LORAWAN_R1 {
		return nil, errUnsupportedLoRaWANVersion.WithAttributes(
			"version", up.Payload.Major,
		)
	}

	logger := log.FromContext(ctx).WithFields(log.Fields(
		"m_type", up.Payload.MType,
		"major", up.Payload.Major,
		"phy_payload_len", len(up.RawPayload),
		"received_at", up.ReceivedAt,
		"frequency", up.Settings.Frequency,
	))
	switch dr := up.Settings.DataRate.Modulation.(type) {
	case *ttnpb.DataRate_FSK:
		logger = logger.WithField(
			"bit_rate", dr.FSK.GetBitRate(),
		)
	case *ttnpb.DataRate_Lora:
		logger = logger.WithFields(log.Fields(
			"bandwidth", dr.Lora.GetBandwidth(),
			"spreading_factor", dr.Lora.GetSpreadingFactor(),
		))
	default:
		return nil, errDataRateNotFound.New()
	}
	ctx = log.NewContext(ctx, logger)

	if t, err := toa.Compute(len(up.RawPayload), up.Settings); err != nil {
		log.FromContext(ctx).WithError(err).Debug("Failed to compute time-on-air")
	} else {
		up.ConsumedAirtime = &t
	}
	switch up.Payload.MType {
	case ttnpb.MType_CONFIRMED_UP, ttnpb.MType_UNCONFIRMED_UP:
		return ttnpb.Empty, ns.handleDataUplink(ctx, up)
	case ttnpb.MType_JOIN_REQUEST:
		return ttnpb.Empty, ns.handleJoinRequest(ctx, up)
	case ttnpb.MType_REJOIN_REQUEST:
		return ttnpb.Empty, ns.handleRejoinRequest(ctx, up)
	}
	logger.Debug("Unmatched MType")
	return ttnpb.Empty, nil
}

// ReportTxAcknowledgment is called by the Gateway Server when a tx acknowledgment arrives.
func (ns *NetworkServer) ReportTxAcknowledgment(ctx context.Context, up *ttnpb.GatewayTxAcknowledgment) (_ *pbtypes.Empty, err error) {
	ack := up.GetTxAck()
	if ack.GetResult() != ttnpb.TxAcknowledgment_SUCCESS {
		return ttnpb.Empty, nil
	}
	down, err := ns.scheduledDownlinkMatcher.Match(ctx, ack)
	if err != nil {
		if errors.IsNotFound(err) {
			log.FromContext(ctx).Debug("Received TxAck but did not match scheduled downlink")
			return ttnpb.Empty, nil
		}
		return nil, err
	}
	ids := down.GetEndDeviceIds()
	if ids == nil {
		return ttnpb.Empty, nil
	}
	pld := down.GetPayload().GetMACPayload()
	if pld == nil {
		return ttnpb.Empty, nil
	}
	if pld.GetFPort() == 0 {
		return ttnpb.Empty, nil
	}
	ns.enqueueApplicationUplinks(ctx, &ttnpb.ApplicationUp{
		EndDeviceIdentifiers: *ids,
		Up: &ttnpb.ApplicationUp_DownlinkSent{
			DownlinkSent: &ttnpb.ApplicationDownlink{
				SessionKeyID:   down.GetSessionKeyId(),
				FRMPayload:     pld.GetFRMPayload(),
				FPort:          pld.GetFPort(),
				FCnt:           pld.GetFullFCnt(),
				CorrelationIDs: ack.GetCorrelationIDs(),
				Priority:       down.GetRequest().GetPriority(),
			},
		},
	})
	return ttnpb.Empty, nil
}
