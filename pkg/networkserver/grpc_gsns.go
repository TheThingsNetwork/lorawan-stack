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
	"runtime/trace"
	"slices"

	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/specification/macspec"
	"go.thethings.network/lorawan-stack/v3/pkg/specification/relayspec"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/toa"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	// recentUplinkCount is the maximum amount of recent uplinks stored per device.
	recentUplinkCount = 20

	// retransmissionWindow is the maximum delay between Rx2 end and an uplink retransmission.
	retransmissionWindow = 10 * time.Second

	// maxConfNbTrans is the maximum number of confirmed uplink retransmissions for pre-1.0.4 devices.
	maxConfNbTrans = 5

	// joinRequestCollectionWindow is the duration for which duplicated JoinRequests are collected.
	// This parameter is separated from the uplink collection period since the JoinRequest may have to be
	// served by a Join Server which is either geographically far away, or simply slow to respond.
	joinRequestCollectionWindow = 6 * time.Second

	// DeduplicationLimit is the number of metadata to deduplicate for a single transmission.
	deduplicationLimit = 50
)

// UplinkDeduplicator represents an entity, that deduplicates uplinks and accumulates metadata.
type UplinkDeduplicator interface {
	// DeduplicateUplink deduplicates an uplink message for specified time.Duration, in the provided round.
	// DeduplicateUplink returns true if the uplink is not a duplicate or false and error, if any, otherwise.
	DeduplicateUplink(
		ctx context.Context, up *ttnpb.UplinkMessage, window time.Duration, limit int, round uint64,
	) (first bool, err error)
	// AccumulatedMetadata returns accumulated metadata for specified uplink message in the provided round and error, if any.
	AccumulatedMetadata(ctx context.Context, up *ttnpb.UplinkMessage, round uint64) (mds []*ttnpb.RxMetadata, err error)
}

func (ns *NetworkServer) deduplicateUplink(
	ctx context.Context, up *ttnpb.UplinkMessage, window time.Duration, limit int, round uint64,
) (bool, error) {
	ok, err := ns.uplinkDeduplicator.DeduplicateUplink(ctx, up, window, limit, round)
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
	if macspec.LimitConfirmedTransmissions(ver) {
		return maxConfNbTrans
	}
	return nbTrans
}

func maxRetransmissionDelay(rxDelay ttnpb.RxDelay) time.Duration {
	return rxDelay.Duration() + time.Second + retransmissionWindow
}

func matchCmacF(ctx context.Context, fNwkSIntKey types.AES128Key, macVersion ttnpb.MACVersion, fCnt uint32, up *ttnpb.UplinkMessage) ([4]byte, bool) {
	trace.Log(ctx, "ns", "compute mic")
	registerMICComputation(ctx)
	cmacF, err := crypto.ComputeLegacyUplinkMIC(
		fNwkSIntKey,
		types.MustDevAddr(up.Payload.GetMacPayload().FHdr.DevAddr).OrZero(),
		fCnt,
		up.RawPayload[:len(up.RawPayload)-4],
	)
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to compute cmacF")
		return [4]byte{}, false
	}
	var micMatch bool
	if macspec.UseLegacyMIC(macVersion) {
		micMatch = bytes.Equal(up.Payload.Mic, cmacF[:])
	} else {
		micMatch = bytes.Equal(up.Payload.Mic[2:], cmacF[:2])
	}
	if !micMatch {
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
	queuedLength := len(dev.MacState.QueuedResponses)
	return func(ctx context.Context, dev *ttnpb.EndDevice, up *ttnpb.UplinkMessage) (events.Builders, error) {
		switch n := len(dev.MacState.QueuedResponses); {
		case n < queuedLength:
			return nil, ErrCorruptedMACState.
				WithAttributes(
					"queued_length", queuedLength,
					"n", n,
				).
				WithCause(ErrMACHandler)
		case n == queuedLength:
			return f(ctx, dev, up)
		default:
			tail := append(dev.MacState.QueuedResponses[:0:0], dev.MacState.QueuedResponses[queuedLength:]...)
			dev.MacState.QueuedResponses = dev.MacState.QueuedResponses[:queuedLength]
			evs, err := f(ctx, dev, up)
			dev.MacState.QueuedResponses = append(dev.MacState.QueuedResponses, tail...)
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

func appendRecentMACCommandIdentifier(
	cids []ttnpb.MACCommandIdentifier,
	cid ttnpb.MACCommandIdentifier,
) []ttnpb.MACCommandIdentifier {
	switch {
	case mac.ContainsStickyMACCommand(cid):
		return append(cids, cid)
	default:
		return cids
	}
}

// matchAndHandleDataUplink handles and matches a device prematched by CMACF check.
func (ns *NetworkServer) matchAndHandleDataUplink(ctx context.Context, dev *ttnpb.EndDevice, up *ttnpb.UplinkMessage, deduplicated bool, cmacFMatchResult cmacFMatchingResult) (*matchResult, bool, error) {
	defer trace.StartRegion(ctx, "match and handle data uplink").End()

	fps, err := ns.FrequencyPlansStore(ctx)
	if err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to get frequency plans store")
		return nil, false, nil
	}
	phy, err := DeviceBand(dev, fps)
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
		"data_rate", up.Settings.DataRate,
	))

	pld := up.Payload.GetMacPayload()
	devAddr := types.MustDevAddr(pld.FHdr.DevAddr).OrZero()
	pendingAppDown := dev.MacState.GetPendingApplicationDownlink()

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

	// Pending session match
	if !pld.FHdr.FCtrl.Ack &&
		cmacFMatchResult.IsPending &&
		dev.PendingSession != nil &&
		dev.PendingMacState != nil &&
		devAddr.Equal(types.MustDevAddr(dev.PendingSession.DevAddr).OrZero()) &&
		macspec.UseLegacyMIC(cmacFMatchResult.LoRaWANVersion) == macspec.UseLegacyMIC(dev.PendingMacState.LorawanVersion) {
		fNwkSIntKey, err := cryptoutil.UnwrapAES128Key(ctx, dev.PendingSession.Keys.FNwkSIntKey, ns.KeyService())
		if err != nil {
			log.FromContext(ctx).WithError(err).WithField("kek_label", dev.PendingSession.Keys.FNwkSIntKey.KekLabel).Warn("Failed to unwrap FNwkSIntKey")
			return nil, false, nil
		}
		if cmacFMatchResult.FNwkSIntKey.Equal(fNwkSIntKey) {
			ctx = log.NewContextWithField(ctx, "mac_version", dev.PendingMacState.LorawanVersion)
			if dev.PendingMacState.PendingJoinRequest == nil {
				log.FromContext(ctx).Warn("Pending join-request missing")
				return nil, false, nil
			}
			dev.PendingMacState.CurrentParameters.Rx1Delay = dev.PendingMacState.PendingJoinRequest.RxDelay
			dev.PendingMacState.CurrentParameters.Rx1DataRateOffset = dev.PendingMacState.PendingJoinRequest.DownlinkSettings.Rx1DrOffset
			dev.PendingMacState.CurrentParameters.Rx2DataRateIndex = dev.PendingMacState.PendingJoinRequest.DownlinkSettings.Rx2Dr
			if dev.PendingMacState.PendingJoinRequest.DownlinkSettings.OptNeg && macspec.UseRekeyInd(dev.LorawanVersion) {
				// The version will be further negotiated via RekeyInd/RekeyConf.
				dev.PendingMacState.LorawanVersion = macspec.RekeyPeriodVersion(dev.LorawanVersion)
			}
			chs, ok := applyCFList(dev.PendingMacState.PendingJoinRequest.CfList, phy, dev.PendingMacState.CurrentParameters.Channels...)
			if !ok {
				log.FromContext(ctx).Debug("Failed to apply CFList")
				return nil, false, nil
			}
			dev.PendingMacState.CurrentParameters.Channels = chs

			dev.MacState = dev.PendingMacState
			dev.PendingSession.StartedAt = up.ReceivedAt

			trace.Log(ctx, "ns", "pending session match")
			matchType = pendingMatch
		}
	}

	// Current session match
	if matchType != pendingMatch &&
		dev.Session != nil &&
		dev.MacState != nil &&
		devAddr.Equal(types.MustDevAddr(dev.Session.DevAddr).OrZero()) &&
		macspec.UseLegacyMIC(cmacFMatchResult.LoRaWANVersion) == macspec.UseLegacyMIC(dev.MacState.LorawanVersion) &&
		(cmacFMatchResult.FullFCnt == FullFCnt(uint16(pld.FHdr.FCnt), dev.Session.LastFCntUp, mac.DeviceSupports32BitFCnt(dev, ns.defaultMACSettings)) ||
			cmacFMatchResult.FullFCnt == pld.FHdr.FCnt) {
		fNwkSIntKey, err := cryptoutil.UnwrapAES128Key(ctx, dev.Session.Keys.FNwkSIntKey, ns.KeyService())
		if err != nil {
			log.FromContext(ctx).WithError(err).WithField("kek_label", dev.Session.Keys.FNwkSIntKey.KekLabel).Warn("Failed to unwrap FNwkSIntKey")
			return nil, false, nil
		}
		if cmacFMatchResult.FNwkSIntKey.Equal(fNwkSIntKey) {
			ctx = log.NewContextWithFields(ctx, log.Fields(
				"last_f_cnt_up", dev.Session.LastFCntUp,
				"mac_version", dev.MacState.LorawanVersion,
				"pending_session", false,
			))
			switch {
			case cmacFMatchResult.FullFCnt < dev.Session.LastFCntUp:
				if pld.FHdr.FCtrl.Ack || dev.Session.LastFCntUp != cmacFMatchResult.LastFCnt || !mac.DeviceResetsFCnt(dev, ns.defaultMACSettings) {
					return nil, false, nil
				}
				ctx = log.NewContextWithField(ctx, "f_cnt_reset", true)

				macState, err := mac.NewState(dev, fps, ns.defaultMACSettings)
				if err != nil {
					log.FromContext(ctx).WithError(err).Warn("Failed to generate new MAC state")
					return nil, false, nil
				}

				dev.MacState = macState
				dev.Session.StartedAt = up.ReceivedAt

				trace.Log(ctx, "ns", "current session match with reset")
				matchType = currentResetMatch

			case cmacFMatchResult.FullFCnt > dev.Session.LastFCntUp,
				dev.Session.LastFCntUp == 0 && dev.SupportsJoin && len(dev.MacState.RecentUplinks) == 1,
				dev.Session.LastFCntUp == 0 && !dev.SupportsJoin && len(dev.MacState.RecentUplinks) == 0:
				ctx = log.NewContextWithField(ctx, "f_cnt_reset", false)

				fCntGap := cmacFMatchResult.FullFCnt - dev.Session.LastFCntUp
				if macspec.HasMaxFCntGap(dev.MacState.LorawanVersion) && uint(fCntGap) > phy.MaxFCntGap {
					log.FromContext(ctx).WithFields(log.Fields(
						"f_cnt_gap", fCntGap,
						"max_f_cnt_gap", phy.MaxFCntGap,
					)).Debug("FCnt gap exceeds maximum after reset")
					return nil, false, nil
				}

				trace.Log(ctx, "ns", "current session match")
				matchType = currentOriginalMatch

			default: // cmacFMatchResult.FullFCnt == dev.Session.LastFCntUp
				ctx = log.NewContextWithField(ctx, "f_cnt_reset", false)

				maxNbTrans := maxTransmissionNumber(dev.MacState.LorawanVersion, up.Payload.MHdr.MType == ttnpb.MType_CONFIRMED_UP, dev.MacState.CurrentParameters.AdrNbTrans)
				if maxNbTrans < 1 {
					panic(fmt.Sprintf("invalid maximum transmission number %d", maxNbTrans))
				}
				ctx = log.NewContextWithField(ctx, "max_transmissions", maxNbTrans)

				nbTrans := uint32(1)
				var (
					lastAt             time.Time
					recentUpPHYPayload []byte
				)
				for i := len(dev.MacState.RecentUplinks) - 1; i >= 0; i-- {
					recentUp := dev.MacState.RecentUplinks[i]
					recentUpPHYPayload, err = lorawan.AppendMessage(recentUpPHYPayload[:0], recentUp.Payload)
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
					if recvAt := ttnpb.StdTime(recentUp.ReceivedAt); recvAt.After(lastAt) {
						lastAt = *recvAt
					}
				}
				if nbTrans < 2 || lastAt.IsZero() {
					log.FromContext(ctx).Debug("Repeated FCnt value, but frame is not a retransmission")
					return nil, false, nil
				}
				maxDelay := maxRetransmissionDelay(dev.MacState.CurrentParameters.Rx1Delay)
				delay := ttnpb.StdTime(up.ReceivedAt).Sub(lastAt)
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

				trace.Log(ctx, "ns", "current session match with retransmission")
				matchType = currentRetransmissionMatch
			}
		} else {
			return nil, false, nil
		}
	} else if matchType != pendingMatch {
		return nil, false, nil
	}

	// NOTE: We assume no dwell time if current value unknown.
	if macspec.IgnoreUplinksExceedingLengthLimit(dev.MacState.LorawanVersion) && len(up.RawPayload)-5 > int(dr.MaxMACPayloadSize(dev.MacState.CurrentParameters.UplinkDwellTime.GetValue())) {
		log.FromContext(ctx).Debug("Uplink length exceeds maximum")
		return nil, false, nil
	}

	cmdBuf := pld.FHdr.FOpts
	if pld.FPort == 0 && len(pld.FrmPayload) > 0 {
		cmdBuf = pld.FrmPayload
	}
	cmdsInFOpts := len(pld.FHdr.FOpts) > 0
	if len(cmdBuf) > 0 && (!cmdsInFOpts || macspec.EncryptFOpts(dev.MacState.LorawanVersion)) {
		session := dev.Session
		if matchType == pendingMatch {
			session = dev.PendingSession
		}
		if session.GetKeys().GetNwkSEncKey() == nil {
			log.FromContext(ctx).Warn("Device missing NwkSEncKey in registry")
			return nil, false, nil
		}
		key, err := cryptoutil.UnwrapAES128Key(ctx, session.Keys.NwkSEncKey, ns.KeyService())
		if err != nil {
			log.FromContext(ctx).WithField("kek_label", session.Keys.NwkSEncKey.KekLabel).WithError(err).Warn("Failed to unwrap NwkSEncKey")
			return nil, false, nil
		}
		encOpts := macspec.EncryptionOptions(dev.MacState.LorawanVersion, macspec.UplinkFrame, pld.FPort, cmdsInFOpts)
		cmdBuf, err = crypto.DecryptUplink(key, devAddr, cmacFMatchResult.FullFCnt, cmdBuf, encOpts...)
		if err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to decrypt uplink")
			return nil, false, nil
		}
	}

	logger := log.FromContext(ctx)
	if matchType == currentRetransmissionMatch {
		dev.MacState.PendingRequests = nil
	}
	var queuedEventBuilders []events.Builder
	var cmds []*ttnpb.MACCommand
	for r := bytes.NewReader(cmdBuf); r.Len() > 0; {
		cmd := &ttnpb.MACCommand{}
		if err := lorawan.DefaultMACCommands.ReadUplink(*phy, r, cmd); err != nil {
			log.FromContext(ctx).WithFields(log.Fields(
				"bytes_left", r.Len(),
				"mac_count", len(cmds),
			)).WithError(err).Debug("Failed to read MAC command")
			queuedEventBuilders = append(queuedEventBuilders, mac.EvtParseMACCommandFail.BindData(err))
			break
		}
		logger := logger.WithField("command", cmd)
		logger.Debug("Read MAC command")
		def, ok := lorawan.DefaultMACCommands[cmd.Cid]
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
	trace.Logf(ctx, "ns", "read %d MAC commands", len(cmds))
	ctx = log.NewContext(ctx, logger)

	if pld.FHdr.FCtrl.ClassB {
		switch {
		case !dev.SupportsClassB:
			logger.Debug("Ignore class B bit in uplink, since device does not support class B")

		case dev.MacState.CurrentParameters.PingSlotFrequency == 0 && len(phy.PingSlotFrequencies) == 0:
			logger.Debug("Ignore class B bit in uplink, since ping slot frequency is not known")

		case dev.MacState.CurrentParameters.PingSlotDataRateIndexValue == nil:
			logger.Debug("Ignore class B bit in uplink, since ping slot data rate index is not known")

		case dev.MacState.PingSlotPeriodicity == nil:
			logger.Debug("Ignore class B bit in uplink, since ping slot periodicity is not known")

		case dev.MacState.DeviceClass != ttnpb.Class_CLASS_B:
			logger.WithField("previous_class", dev.MacState.DeviceClass).Debug("Switch device class to class B")
			queuedEventBuilders = append(queuedEventBuilders, mac.EvtClassBSwitch.BindData(dev.MacState.DeviceClass))
			dev.MacState.DeviceClass = ttnpb.Class_CLASS_B
		}
	} else if dev.MacState.DeviceClass == ttnpb.Class_CLASS_B {
		if !macspec.UseDeviceModeInd(dev.MacState.LorawanVersion) && dev.SupportsClassC {
			queuedEventBuilders = append(queuedEventBuilders, mac.EvtClassCSwitch.BindData(ttnpb.Class_CLASS_B))
			dev.MacState.DeviceClass = ttnpb.Class_CLASS_C
		} else {
			queuedEventBuilders = append(queuedEventBuilders, mac.EvtClassASwitch.BindData(ttnpb.Class_CLASS_B))
			dev.MacState.DeviceClass = ttnpb.Class_CLASS_A
		}
	}

	var deferredMACHandlers []macHandler
	if len(cmds) > 0 && !deduplicated {
		deferredMACHandlers = make([]macHandler, 0, 2)
	}
	var setPaths []string
	recentMACCommandIdentifiers := make([]ttnpb.MACCommandIdentifier, 0, 1)
	dev.MacState.QueuedResponses = dev.MacState.QueuedResponses[:0]
macLoop:
	for len(cmds) > 0 {
		var cmd *ttnpb.MACCommand
		cmd, cmds = cmds[0], cmds[1:]
		logger := logger.WithField("command", cmd)
		logger.Debug("Handle MAC command")
		ctx := log.NewContext(ctx, logger)

		var evs events.Builders
		var err error
		switch cmd.Cid {
		case ttnpb.MACCommandIdentifier_CID_RESET:
			evs, err = mac.HandleResetInd(ctx, dev, cmd.GetResetInd(), fps, ns.defaultMACSettings)
		case ttnpb.MACCommandIdentifier_CID_LINK_CHECK:
			if !deduplicated {
				deferredMACHandlers = append(deferredMACHandlers, makeDeferredMACHandler(dev, mac.HandleLinkCheckReq))
				continue macLoop
			}
			evs, err = mac.HandleLinkCheckReq(ctx, dev, up)
		case ttnpb.MACCommandIdentifier_CID_LINK_ADR:
			pld := cmd.GetLinkAdrAns()
			dupCount := 0
			if macspec.AllowDuplicateLinkADRAns(dev.MacState.LorawanVersion) {
				for _, dup := range cmds {
					if dup.Cid != ttnpb.MACCommandIdentifier_CID_LINK_ADR {
						break
					}
					if !proto.Equal(dup.GetLinkAdrAns(), pld) {
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
			evs, err = mac.HandleLinkADRAns(ctx, dev, pld, uint(dupCount), cmacFMatchResult.FullFCnt, fps)
		case ttnpb.MACCommandIdentifier_CID_DUTY_CYCLE:
			evs, err = mac.HandleDutyCycleAns(ctx, dev)
		case ttnpb.MACCommandIdentifier_CID_RX_PARAM_SETUP:
			evs, err = mac.HandleRxParamSetupAns(ctx, dev, cmd.GetRxParamSetupAns())
		case ttnpb.MACCommandIdentifier_CID_DEV_STATUS:
			evs, err = mac.HandleDevStatusAns(ctx, dev, cmd.GetDevStatusAns(), cmacFMatchResult.FullFCnt, *ttnpb.StdTime(up.ReceivedAt))
			if err == nil {
				setPaths = ttnpb.AddFields(setPaths,
					"battery_percentage",
					"downlink_margin",
					"last_dev_status_received_at",
					"power_state",
				)
			}
		case ttnpb.MACCommandIdentifier_CID_NEW_CHANNEL:
			evs, err = mac.HandleNewChannelAns(ctx, dev, cmd.GetNewChannelAns())
		case ttnpb.MACCommandIdentifier_CID_RX_TIMING_SETUP:
			evs, err = mac.HandleRxTimingSetupAns(ctx, dev)
		case ttnpb.MACCommandIdentifier_CID_TX_PARAM_SETUP:
			evs, err = mac.HandleTxParamSetupAns(ctx, dev)
		case ttnpb.MACCommandIdentifier_CID_DL_CHANNEL:
			evs, err = mac.HandleDLChannelAns(ctx, dev, cmd.GetDlChannelAns())
		case ttnpb.MACCommandIdentifier_CID_REKEY:
			evs, err = mac.HandleRekeyInd(ctx, dev, cmd.GetRekeyInd(), devAddr)
		case ttnpb.MACCommandIdentifier_CID_ADR_PARAM_SETUP:
			evs, err = mac.HandleADRParamSetupAns(ctx, dev)
		case ttnpb.MACCommandIdentifier_CID_DEVICE_TIME:
			if !deduplicated {
				m := makeDeferredMACHandler(dev, mac.HandleDeviceTimeReq)
				deferredMACHandlers = append(deferredMACHandlers, m)
				continue macLoop
			}
			evs, err = mac.HandleDeviceTimeReq(ctx, dev, up)
		case ttnpb.MACCommandIdentifier_CID_REJOIN_PARAM_SETUP:
			evs, err = mac.HandleRejoinParamSetupAns(ctx, dev, cmd.GetRejoinParamSetupAns())
		case ttnpb.MACCommandIdentifier_CID_PING_SLOT_INFO:
			evs, err = mac.HandlePingSlotInfoReq(ctx, dev, cmd.GetPingSlotInfoReq())
		case ttnpb.MACCommandIdentifier_CID_PING_SLOT_CHANNEL:
			evs, err = mac.HandlePingSlotChannelAns(ctx, dev, cmd.GetPingSlotChannelAns())
		case ttnpb.MACCommandIdentifier_CID_BEACON_TIMING:
			evs, err = mac.HandleBeaconTimingReq(ctx, dev)
		case ttnpb.MACCommandIdentifier_CID_BEACON_FREQ:
			evs, err = mac.HandleBeaconFreqAns(ctx, dev, cmd.GetBeaconFreqAns())
		case ttnpb.MACCommandIdentifier_CID_DEVICE_MODE:
			evs, err = mac.HandleDeviceModeInd(ctx, dev, cmd.GetDeviceModeInd())
		case ttnpb.MACCommandIdentifier_CID_RELAY_CONF:
			evs, err = mac.HandleRelayConfAns(ctx, dev, cmd.GetRelayConfAns())
		case ttnpb.MACCommandIdentifier_CID_RELAY_END_DEVICE_CONF:
			evs, err = mac.HandleRelayEndDeviceConfAns(ctx, dev, cmd.GetRelayEndDeviceConfAns())
		case ttnpb.MACCommandIdentifier_CID_RELAY_UPDATE_UPLINK_LIST:
			evs, err = mac.HandleRelayUpdateUplinkListAns(ctx, dev, cmd.GetRelayUpdateUplinkListAns())
		case ttnpb.MACCommandIdentifier_CID_RELAY_CTRL_UPLINK_LIST:
			evs, err = mac.HandleRelayCtrlUplinkListAns(ctx, dev, cmd.GetRelayCtrlUplinkListAns())
		case ttnpb.MACCommandIdentifier_CID_RELAY_CONFIGURE_FWD_LIMIT:
			evs, err = mac.HandleRelayConfigureFwdLimitAns(ctx, dev, cmd.GetRelayConfigureFwdLimitAns())
		case ttnpb.MACCommandIdentifier_CID_RELAY_NOTIFY_NEW_END_DEVICE:
			evs, err = mac.HandleRelayNotifyNewEndDeviceReq(ctx, dev, cmd.GetRelayNotifyNewEndDeviceReq())
		default:
			_, known := lorawan.DefaultMACCommands[cmd.Cid]
			logger.WithField("known", known).Debug("Unknown MAC command received")
			queuedEventBuilders = append(queuedEventBuilders, mac.EvtUnknownMACCommand.BindData(cmd))
			break macLoop
		}
		if err != nil {
			logger.WithError(err).Debug("Failed to process MAC command")
			queuedEventBuilders = append(queuedEventBuilders, mac.EvtProcessMACCommandFail.BindData(err))
			break macLoop
		}
		queuedEventBuilders = append(queuedEventBuilders, evs...)
		recentMACCommandIdentifiers = appendRecentMACCommandIdentifier(recentMACCommandIdentifiers, cmd.Cid)
	}
	dev.MacState.RecentMacCommandIdentifiers = recentMACCommandIdentifiers
	if n := len(dev.MacState.PendingRequests); n > 0 {
		logger.WithField("unanswered_request_count", n).Debug("MAC command buffer not fully answered")
		queuedEventBuilders = append(queuedEventBuilders, mac.EvtUnansweredMACCommand.BindData(&ttnpb.MACCommands{
			Commands: slices.Clone(dev.MacState.PendingRequests),
		}))
		dev.MacState.PendingRequests = dev.MacState.PendingRequests[:0]
	}

	if matchType == pendingMatch {
		if !macspec.UseRekeyInd(dev.MacState.LorawanVersion) {
			dev.Ids.DevAddr = devAddr.Bytes()
			dev.Session = dev.PendingSession
		} else if dev.PendingSession != nil || dev.PendingMacState != nil || dev.MacState.PendingJoinRequest != nil {
			logger.Debug("No RekeyInd received for LoRaWAN 1.1+ device")
			return nil, false, nil
		}
		setPaths = ttnpb.AddFields(setPaths, "ids.dev_addr")
	} else if dev.PendingSession != nil || dev.PendingMacState != nil {
		// TODO: Notify AS of session recovery(https://github.com/TheThingsNetwork/lorawan-stack/issues/594)
	}
	dev.MacState.PendingJoinRequest = nil
	dev.MacState.PendingRelayDownlink = nil
	dev.PendingMacState = nil
	dev.PendingSession = nil

	chIdx, err := searchUplinkChannel(up.Settings.Frequency, dev.MacState)
	if err != nil {
		logger.WithError(err).Debug("Failed to determine channel index of uplink")
		return nil, false, nil
	}
	logger = logger.WithField("device_channel_index", chIdx)
	ctx = log.NewContext(ctx, logger)

	// NOTE: Legacy MIC check is already performed.
	if !macspec.UseLegacyMIC(dev.MacState.LorawanVersion) {
		sNwkSIntKey, err := cryptoutil.UnwrapAES128Key(ctx, dev.Session.Keys.SNwkSIntKey, ns.KeyService())
		if err != nil {
			logger.WithField("kek_label", dev.Session.Keys.SNwkSIntKey.GetKekLabel()).WithError(err).Warn("Failed to unwrap SNwkSIntKey")
			return nil, false, nil
		}

		var confFCnt uint32
		if pld.FHdr.FCtrl.Ack {
			confFCnt = dev.Session.LastConfFCntDown
		}
		trace.Log(ctx, "ns", "compute mic")
		registerMICComputation(ctx)
		fullMIC, err := crypto.ComputeUplinkMICFromLegacy(
			cmacFMatchResult.CmacF,
			sNwkSIntKey,
			confFCnt,
			uint8(drIdx),
			chIdx,
			devAddr,
			cmacFMatchResult.FullFCnt,
			up.RawPayload[:len(up.RawPayload)-4],
		)
		if err != nil {
			logger.WithError(err).Error("Failed to compute 1.1 MIC")
			return nil, false, nil
		}
		if !bytes.Equal(up.Payload.Mic, fullMIC[:]) {
			trace.Log(ctx, "ns", "no mic match")
			logger.Debug("Full MIC mismatch")
			return nil, false, nil
		}
	}
	dev.MacState.RxWindowsAvailable = true
	dev.Session.LastFCntUp = cmacFMatchResult.FullFCnt

	var queuedApplicationUplinks []*ttnpb.ApplicationUp
	if pendingAppDown != nil {
		if pld.FHdr.FCtrl.Ack {
			queuedApplicationUplinks = []*ttnpb.ApplicationUp{
				{
					EndDeviceIds: dev.Ids,
					Up: &ttnpb.ApplicationUp_DownlinkAck{
						DownlinkAck: pendingAppDown,
					},
					CorrelationIds: append(pendingAppDown.CorrelationIds, up.CorrelationIds...),
				},
			}
		} else {
			queuedApplicationUplinks = []*ttnpb.ApplicationUp{
				{
					EndDeviceIds: dev.Ids,
					Up: &ttnpb.ApplicationUp_DownlinkNack{
						DownlinkNack: pendingAppDown,
					},
					CorrelationIds: append(pendingAppDown.CorrelationIds, up.CorrelationIds...),
				},
			}
		}
		if dev.MacState != nil {
			dev.MacState.PendingApplicationDownlink = nil
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
		SetPaths: ttnpb.AddFields(setPaths,
			"mac_state",
			"pending_mac_state",
			"pending_session",
			"session",
		),
	}, true, nil
}

func toMACStateRxMetadata(mds []*ttnpb.RxMetadata) []*ttnpb.MACState_UplinkMessage_RxMetadata {
	if len(mds) == 0 {
		return nil
	}
	recentMDs := make([]*ttnpb.MACState_UplinkMessage_RxMetadata, 0, len(mds))
	for _, md := range mds {
		var pbMD *ttnpb.MACState_UplinkMessage_RxMetadata_PacketBrokerMetadata
		if md.PacketBroker != nil {
			pbMD = &ttnpb.MACState_UplinkMessage_RxMetadata_PacketBrokerMetadata{}
		}
		var relayMD *ttnpb.MACState_UplinkMessage_RxMetadata_RelayMetadata
		if md.Relay != nil {
			relayMD = &ttnpb.MACState_UplinkMessage_RxMetadata_RelayMetadata{}
		}
		recentMDs = append(recentMDs, &ttnpb.MACState_UplinkMessage_RxMetadata{
			GatewayIds:             md.GatewayIds,
			PacketBroker:           pbMD,
			Relay:                  relayMD,
			ChannelRssi:            md.ChannelRssi,
			Snr:                    md.Snr,
			DownlinkPathConstraint: md.DownlinkPathConstraint,
			UplinkToken:            md.UplinkToken,
		})
	}
	return recentMDs
}

func toMACStateTxSettings(settings *ttnpb.TxSettings) *ttnpb.MACState_UplinkMessage_TxSettings {
	if settings == nil {
		return nil
	}
	return &ttnpb.MACState_UplinkMessage_TxSettings{
		DataRate: settings.DataRate,
	}
}

func toMACStateUplinkMessages(ups ...*ttnpb.UplinkMessage) []*ttnpb.MACState_UplinkMessage {
	if len(ups) == 0 {
		return nil
	}
	recentUps := make([]*ttnpb.MACState_UplinkMessage, 0, len(ups))
	for _, up := range ups {
		recentUps = append(recentUps, &ttnpb.MACState_UplinkMessage{
			Payload:            up.Payload,
			Settings:           toMACStateTxSettings(up.Settings),
			RxMetadata:         toMACStateRxMetadata(up.RxMetadata),
			ReceivedAt:         up.ReceivedAt,
			CorrelationIds:     up.CorrelationIds,
			DeviceChannelIndex: up.DeviceChannelIndex,
		})
	}
	return recentUps
}

func appendRecentUplink(
	recent []*ttnpb.MACState_UplinkMessage,
	up *ttnpb.UplinkMessage,
	window int,
) []*ttnpb.MACState_UplinkMessage {
	ups := toMACStateUplinkMessages(up)
	if n := len(recent); n > 0 {
		recent[n-1].CorrelationIds = nil
		if len(downlinkPathsFromRecentUplinks(ups...)) > 0 {
			for _, md := range recent[n-1].RxMetadata {
				md.UplinkToken = nil
				md.DownlinkPathConstraint = ttnpb.DownlinkPathConstraint_DOWNLINK_PATH_CONSTRAINT_NEVER
			}
		}
	}
	recent = append(recent, ups...)
	if extra := len(recent) - window; extra > 0 {
		recent = recent[extra:]
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
// mergeMetadata mutates up.RxMetadata.
func (ns *NetworkServer) mergeMetadata(ctx context.Context, up *ttnpb.UplinkMessage, round uint64) {
	mds, err := ns.uplinkDeduplicator.AccumulatedMetadata(ctx, up, round)
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to merge metadata")
		return
	}
	if len(mds) == 0 {
		log.FromContext(ctx).Warn("No metadata to merge, keep uplink message metadata")
		return
	}
	up.RxMetadata = mds
	log.FromContext(ctx).WithField("metadata_count", len(up.RxMetadata)).Debug("Merged metadata")
	registerMergeMetadata(ctx, up)
}

// filterMetadata filters the collected metadata.
// filterMetadata removes metadata from Packet Broker that has been received from a forwarder that identifies like this
// Network Server identifies itself. This is to avoid that failed downlink attempts through Gateway Server lead to
// downlink scheduling attempts through Packet Broker, ending up on the same Gateway Server that already failed to schedule.
// filterMetadata mutates up.RxMetadata.
func (ns *NetworkServer) filterMetadata(ctx context.Context, up *ttnpb.UplinkMessage) {
	mds := make([]*ttnpb.RxMetadata, 0, len(up.RxMetadata))
	for _, md := range up.RxMetadata {
		if pbMD := md.GetPacketBroker(); pbMD != nil {
			if types.MustNetID(pbMD.ForwarderNetId).OrZero().Equal(ns.netID(ctx)) &&
				pbMD.ForwarderClusterId == ns.clusterID {
				continue
			}
		}
		mds = append(mds, md)
	}
	up.RxMetadata = mds
	if d := cap(mds) - len(mds); d > 0 {
		log.FromContext(ctx).WithFields(log.Fields(
			"metadata_count", len(mds),
			"filtered_count", d,
		)).Debug("Filtered metadata")
	}
}

const (
	initialDeduplicationRound = iota
)

func (ns *NetworkServer) handleDataUplink(ctx context.Context, up *ttnpb.UplinkMessage) (err error) {
	defer trace.StartRegion(ctx, "handle data uplink").End()

	if len(up.RawPayload) < 4 {
		return errRawPayloadTooShort.New()
	}
	pld := up.Payload.GetMacPayload()
	ctx = log.NewContextWithFields(ctx, log.Fields(
		"ack", pld.FHdr.FCtrl.Ack,
		"adr", pld.FHdr.FCtrl.Adr,
		"adr_ack_req", pld.FHdr.FCtrl.AdrAckReq,
		"class_b", pld.FHdr.FCtrl.ClassB,
		"dev_addr", types.MustDevAddr(pld.FHdr.DevAddr).OrZero(),
		"f_opts_len", len(pld.FHdr.FOpts),
		"f_port", pld.FPort,
		"uplink_f_cnt", pld.FHdr.FCnt,
	))

	ok, err := ns.deduplicateUplink(ctx, up, ns.collectionWindow(ctx), deduplicationLimit, initialDeduplicationRound)
	if err != nil {
		return err
	}
	if !ok {
		trace.Log(ctx, "ns", "message is duplicate (initial round)")
		return errDuplicateUplink.New()
	}
	trace.Log(ctx, "ns", "message is original (initial round)")

	ctx, flushMatchStats := newContextWithMatchStats(ctx)
	defer flushMatchStats()

	var matched *matchResult
	if err := ns.devices.RangeByUplinkMatches(ctx, up,
		func(ctx context.Context, match *UplinkMatch) (bool, error) {
			defer trace.StartRegion(ctx, "iterate uplink match").End()
			registerMatchCandidate(ctx)

			ctx = log.NewContextWithFields(ctx, log.Fields(
				"mac_version", match.LoRaWANVersion,
				"pending_session", match.IsPending,
			))

			fNwkSIntKey, err := cryptoutil.UnwrapAES128Key(ctx, match.FNwkSIntKey, ns.KeyService())
			if err != nil {
				log.FromContext(ctx).WithError(err).WithField("kek_label", match.FNwkSIntKey.KekLabel).Warn("Failed to unwrap FNwkSIntKey")
				return false, nil
			}
			fCnt := FullFCnt(uint16(pld.FHdr.FCnt), match.LastFCnt, mac.DeviceSupports32BitFCnt(&ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					Supports_32BitFCnt: match.Supports32BitFCnt,
				},
			}, ns.defaultMACSettings))

			var cmacF [4]byte
			cmacF, ok = matchCmacF(ctx, fNwkSIntKey, match.LoRaWANVersion, fCnt, up)
			if !ok && fCnt != pld.FHdr.FCnt && !pld.FHdr.FCtrl.Ack && !match.IsPending && mac.DeviceResetsFCnt(&ttnpb.EndDevice{
				MacSettings: &ttnpb.MACSettings{
					ResetsFCnt: match.ResetsFCnt,
				},
			}, ns.defaultMACSettings) {
				// FCnt reset
				fCnt = pld.FHdr.FCnt
				cmacF, ok = matchCmacF(ctx, fNwkSIntKey, match.LoRaWANVersion, fCnt, up)
			}
			if !ok {
				trace.Log(ctx, "ns", "no mic match")
				return false, nil
			}
			trace.Log(ctx, "ns", "mic match")

			ctx = log.NewContextWithField(ctx, "full_f_cnt_up", fCnt)
			dev, ctx, err := ns.devices.GetByID(ctx, match.ApplicationIdentifiers, match.DeviceID, handleDataUplinkGetPaths[:])
			if err != nil {
				log.FromContext(ctx).WithError(err).Warn("Failed to get device after cmacF matching")
				return macspec.UseLegacyMIC(match.LoRaWANVersion), nil
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
			return ok || macspec.UseLegacyMIC(match.LoRaWANVersion), nil
		},
	); err != nil {
		logRegistryRPCError(ctx, err, "Failed to find devices in registry by DevAddr")
		return errDeviceNotFound.WithCause(err)
	}
	if !ok {
		return errDeviceNotFound.New()
	}

	pld.FullFCnt = matched.FullFCnt
	up.DeviceChannelIndex = uint32(matched.ChannelIndex)
	ctx = matched.Context

	queuedEvents := []events.Event{
		evtReceiveDataUplink.NewWithIdentifiersAndData(ctx, matched.Device.Ids, up),
	}
	defer func(ids *ttnpb.EndDeviceIdentifiers) {
		if err != nil {
			queuedEvents = append(queuedEvents, evtDropDataUplink.NewWithIdentifiersAndData(ctx, ids, err))
		}
		publishEvents(ctx, queuedEvents...)
	}(matched.Device.Ids)

	publishEvents(ctx, queuedEvents...)
	queuedEvents = nil
	up = ttnpb.Clone(up)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ns.deduplicationDone(ctx, up):
	}
	ns.mergeMetadata(ctx, up, initialDeduplicationRound)
	ns.filterMetadata(ctx, up)

	for _, f := range matched.DeferredMACHandlers {
		evs, err := f(ctx, matched.Device, up)
		if err != nil {
			log.FromContext(ctx).WithError(err).Debug("Failed to process MAC command after deduplication")
			matched.QueuedEventBuilders = append(
				matched.QueuedEventBuilders, mac.EvtProcessMACCommandFail.BindData(err),
			)
			break
		}
		matched.QueuedEventBuilders = append(matched.QueuedEventBuilders, evs...)
	}

	var queuedApplicationUplinks []*ttnpb.ApplicationUp
	defer func() { ns.submitApplicationUplinks(ctx, queuedApplicationUplinks...) }()

	stored, _, err := ns.devices.SetByID(ctx, matched.Device.Ids.ApplicationIds, matched.Device.Ids.DeviceId, handleDataUplinkGetPaths[:],
		func(ctx context.Context, stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			defer trace.StartRegion(ctx, "update stored device").End()

			if stored == nil {
				log.FromContext(ctx).Warn("Device deleted during uplink handling, drop")
				return nil, nil, errOutdatedData.New()
			}

			if !proto.Equal(matched.Device.CreatedAt, stored.CreatedAt) || !proto.Equal(matched.Device.UpdatedAt, stored.UpdatedAt) {
				matched, ok, err = ns.matchAndHandleDataUplink(ctx, stored, up, true, matched.cmacFMatchingResult)
				if err != nil {
					return nil, nil, err
				}
				if !ok {
					return nil, nil, errOutdatedData.New()
				}
				pld.FullFCnt = matched.FullFCnt
				up.DeviceChannelIndex = uint32(matched.ChannelIndex)
				ctx = matched.Context
			}

			queuedApplicationUplinks = append(queuedApplicationUplinks, matched.QueuedApplicationUplinks...)
			queuedEvents = append(queuedEvents, matched.QueuedEventBuilders.New(ctx, events.WithIdentifiers(matched.Device.Ids))...)

			stored = matched.Device
			paths := ttnpb.AddFields(matched.SetPaths,
				"mac_state.recent_uplinks",
			)
			stored.MacState.RecentUplinks = appendRecentUplink(stored.MacState.RecentUplinks, up, recentUplinkCount)

			if matched.DataRateIndex < stored.MacState.CurrentParameters.AdrDataRateIndex {
				// Device lowers TX power index before lowering data rate index according to the spec.
				stored.MacState.CurrentParameters.AdrTxPowerIndex = 0
				paths = ttnpb.AddFields(paths,
					"mac_state.current_parameters.adr_tx_power_index",
				)
			}
			stored.MacState.CurrentParameters.AdrDataRateIndex = matched.DataRateIndex
			paths = ttnpb.AddFields(paths,
				"mac_state.current_parameters.adr_data_rate_index",
			)

			adaptDataRate, resetDesiredParameters, staticADRSettings := mac.DeviceShouldAdaptDataRate(stored, ns.defaultMACSettings, matched.phy)
			if resetDesiredParameters || staticADRSettings != nil {
				paths = ttnpb.AddFields(paths,
					"mac_state.desired_parameters.adr_data_rate_index",
					"mac_state.desired_parameters.adr_nb_trans",
					"mac_state.desired_parameters.adr_tx_power_index",
				)
			}
			if currentParameters, desiredParameters := stored.MacState.CurrentParameters, stored.MacState.DesiredParameters; resetDesiredParameters {
				desiredParameters.AdrDataRateIndex = currentParameters.AdrDataRateIndex
				desiredParameters.AdrTxPowerIndex = currentParameters.AdrTxPowerIndex
				desiredParameters.AdrNbTrans = currentParameters.AdrNbTrans
			}
			if desiredParameters := stored.MacState.DesiredParameters; staticADRSettings != nil {
				desiredParameters.AdrDataRateIndex = staticADRSettings.DataRateIndex
				desiredParameters.AdrTxPowerIndex = staticADRSettings.TxPowerIndex
				desiredParameters.AdrNbTrans = staticADRSettings.NbTrans
			}
			if !pld.FHdr.FCtrl.Adr || !adaptDataRate {
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
		var frmPayload []byte
		switch pld.FPort {
		case 0:
		case relayspec.FPort:
			relayUp, relayEvents, err := handleRelayForwardingProtocol(
				ctx, matched.Device, matched.FullFCnt, matched.phy, up, ns.KeyService(),
			)
			queuedEvents = append(queuedEvents, relayEvents...)
			if err != nil {
				log.FromContext(ctx).WithError(err).Warn("Failed to handle relay forwarding protocol")
			} else {
				ns.StartTask(&task.Config{
					Context: ns.FromRequestContext(ctx),
					ID:      "loopback_relay_uplink",
					Func:    relayLoopbackFunc(ns.LoopbackConn(), relayUp, ns.WithClusterAuth()),
					Restart: task.RestartNever,
					Backoff: task.DefaultBackoffConfig,
				})
			}
		default:
			frmPayload = pld.FrmPayload
		}
		queuedApplicationUplinks = append(queuedApplicationUplinks, &ttnpb.ApplicationUp{
			EndDeviceIds:   stored.Ids,
			CorrelationIds: up.CorrelationIds,
			Up: &ttnpb.ApplicationUp_UplinkMessage{
				UplinkMessage: &ttnpb.ApplicationUplink{
					Confirmed:       up.Payload.MHdr.MType == ttnpb.MType_CONFIRMED_UP,
					FCnt:            pld.FullFCnt,
					FPort:           pld.FPort,
					FrmPayload:      frmPayload,
					RxMetadata:      up.RxMetadata,
					SessionKeyId:    stored.Session.Keys.SessionKeyId,
					Settings:        up.Settings,
					ReceivedAt:      up.ReceivedAt,
					ConsumedAirtime: up.ConsumedAirtime,
					NetworkIds:      ns.networkIdentifiers(ctx),
				},
			},
		})
	}
	queuedEvents = append(queuedEvents, evtProcessDataUplink.NewWithIdentifiersAndData(ctx, matched.Device.Ids, up))
	registerProcessUplink(ctx, up)
	return nil
}

func joinResponseWithoutKeys(resp *ttnpb.JoinResponse) *ttnpb.JoinResponse {
	return &ttnpb.JoinResponse{
		RawPayload: resp.RawPayload,
		SessionKeys: &ttnpb.SessionKeys{
			SessionKeyId: resp.SessionKeys.SessionKeyId,
		},
		Lifetime:       resp.Lifetime,
		CorrelationIds: resp.CorrelationIds,
	}
}

func (ns *NetworkServer) sendJoinRequest(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, req *ttnpb.JoinRequest) (*ttnpb.JoinResponse, []events.Event, error) {
	defer trace.StartRegion(ctx, "send join request").End()

	var queuedEvents []events.Event
	logger := log.FromContext(ctx)
	cc, err := ns.GetPeerConn(ctx, ttnpb.ClusterRole_JOIN_SERVER, nil)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.WithError(err).Debug("Join Server peer not found")
		} else {
			logger.WithError(err).Error("Join Server peer connection lookup failed")
		}
	} else {
		queuedEvents = append(queuedEvents, evtClusterJoinAttempt.NewWithIdentifiersAndData(ctx, ids, req))
		resp, err := ttnpb.NewNsJsClient(cc).HandleJoin(ctx, req, ns.WithClusterAuth())
		if err == nil {
			logger.Debug("Join-request accepted by cluster-local Join Server")
			queuedEvents = append(queuedEvents, evtClusterJoinSuccess.NewWithIdentifiersAndData(ctx, ids, joinResponseWithoutKeys(resp)))
			return resp, queuedEvents, nil
		}
		logger.WithError(err).Info("Cluster-local Join Server did not accept join-request")
		queuedEvents = append(queuedEvents, evtClusterJoinFail.NewWithIdentifiersAndData(ctx, ids, err))
		if !errors.IsNotFound(err) {
			return nil, queuedEvents, err
		}
	}
	if ns.interopClient != nil {
		queuedEvents = append(queuedEvents, evtInteropJoinAttempt.NewWithIdentifiersAndData(ctx, ids, req))
		resp, err := ns.interopClient.HandleJoinRequest(ctx, ns.netID(ctx), ns.nsID(ctx), req)
		if err == nil {
			logger.Debug("Join-request accepted by interop Join Server")
			queuedEvents = append(queuedEvents, evtInteropJoinSuccess.NewWithIdentifiersAndData(ctx, ids, joinResponseWithoutKeys(resp)))
			return resp, queuedEvents, nil
		}
		logger.WithError(err).Warn("Interop Join Server did not accept join-request")
		queuedEvents = append(queuedEvents, evtInteropJoinFail.NewWithIdentifiersAndData(ctx, ids, err))
		if !errors.IsNotFound(err) {
			return nil, queuedEvents, err
		}
	}
	return nil, queuedEvents, errJoinServerNotFound.New()
}

func (ns *NetworkServer) deduplicationDone(ctx context.Context, up *ttnpb.UplinkMessage) <-chan time.Time {
	return time.After(time.Until(ttnpb.StdTime(up.ReceivedAt).Add(ns.deduplicationWindow(ctx))))
}

func (ns *NetworkServer) handleJoinRequest(ctx context.Context, up *ttnpb.UplinkMessage) (err error) {
	defer trace.StartRegion(ctx, "handle join request").End()

	pld := up.Payload.GetJoinRequestPayload()
	ctx = log.NewContextWithFields(ctx, log.Fields(
		"dev_eui", types.MustEUI64(pld.DevEui).OrZero(),
		"join_eui", types.MustEUI64(pld.JoinEui).OrZero(),
	))

	ok, err := ns.deduplicateUplink(ctx, up, joinRequestCollectionWindow, deduplicationLimit, initialDeduplicationRound)
	if err != nil {
		return err
	}
	if !ok {
		trace.Log(ctx, "ns", "message is duplicate")
		return errDuplicateUplink.New()
	}
	trace.Log(ctx, "ns", "message is original")

	joinEUI, devEUI := types.MustEUI64(pld.JoinEui).OrZero(), types.MustEUI64(pld.DevEui).OrZero()
	matched, matchedCtx, err := ns.devices.GetByEUI(ctx, joinEUI, devEUI,
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
		return errDeviceNotFound.WithCause(err)
	}
	ctx = matchedCtx
	ctx = log.NewContextWithField(ctx, "device_uid", unique.ID(ctx, matched.Ids))

	queuedEvents := []events.Event{
		evtReceiveJoinRequest.NewWithIdentifiersAndData(ctx, matched.Ids, up),
	}
	defer func() {
		if err != nil {
			queuedEvents = append(queuedEvents, evtDropJoinRequest.NewWithIdentifiersAndData(ctx, matched.Ids, err))
		}
		publishEvents(ctx, queuedEvents...)
	}()

	if !matched.SupportsJoin {
		log.FromContext(ctx).Warn("ABP device sent a join-request, drop")
		queuedEvents = append(queuedEvents, evtDropJoinRequest.NewWithIdentifiersAndData(ctx, matched.Ids, errABPJoinRequest))
		return nil
	}

	fps, err := ns.FrequencyPlansStore(ctx)
	if err != nil {
		return err
	}
	fp, phy, err := DeviceFrequencyPlanAndBand(matched, fps)
	if err != nil {
		return err
	}
	ctx = log.NewContextWithField(ctx,
		"data_rate", up.Settings.DataRate,
	)

	macState, err := mac.NewState(matched, fps, ns.defaultMACSettings)
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
		"device_channel_index", chIdx,
	)

	devAddr := ns.newDevAddr(ctx)
	const maxDevAddrGenerationRetries = 5
	for i := 0; i < maxDevAddrGenerationRetries && matched.Session != nil && devAddr.Equal(types.MustDevAddr(matched.Session.DevAddr).OrZero()); i++ {
		devAddr = ns.newDevAddr(ctx)
	}
	ctx = log.NewContextWithField(ctx, "dev_addr", devAddr)
	if matched.Session != nil && devAddr.Equal(types.MustDevAddr(matched.Session.DevAddr).OrZero()) {
		log.FromContext(ctx).Error("Reusing the DevAddr used for current session")
	}

	maxMACPayloadSize, err := computeMaxMACDownlinkPayloadSize(
		macState,
		phy,
		fp,
		uint32(chIdx),
		up.Settings.DataRate,
	)
	if err != nil {
		return err
	}
	// NOTE: The CFList is an optional part of a JoinAccept message. A JoinAccept containing a CFList has size
	// 33, while a JoinAccept without a CFList has size 17. Bands which are susceptible to downlink dwell time
	// limitations such as AS923 and its variants may need to generate messages whose MAC payload is smaller
	// than 19 bytes. For such bands, we need to always omit the CFList.
	// NOTE: The 5 bytes added to the maximum represent the MHDR (1 byte) and the MIC (4 bytes).
	var cfList *ttnpb.CFList
	if maxMACPayloadSize+5 >= lorawan.JoinAcceptWithCFListLength {
		cfList = mac.CFList(phy, macState.DesiredParameters.Channels...)
	}
	dlSettings := &ttnpb.DLSettings{
		Rx1DrOffset: macState.DesiredParameters.Rx1DataRateOffset,
		Rx2Dr:       macState.DesiredParameters.Rx2DataRateIndex,
		OptNeg:      macspec.UseRekeyInd(matched.LorawanVersion),
	}

	resp, joinEvents, err := ns.sendJoinRequest(ctx, matched.Ids, &ttnpb.JoinRequest{
		Payload:            up.Payload,
		CfList:             cfList,
		CorrelationIds:     events.CorrelationIDsFromContext(ctx),
		DevAddr:            devAddr.Bytes(),
		NetId:              ns.netID(ctx).Bytes(),
		RawPayload:         up.RawPayload,
		RxDelay:            macState.DesiredParameters.Rx1Delay,
		SelectedMacVersion: matched.LorawanVersion, // Assume NS version is always higher than the version of the device
		ConsumedAirtime:    up.ConsumedAirtime,
		DownlinkSettings:   dlSettings,
	})

	queuedEvents = append(queuedEvents, joinEvents...)
	if err != nil {
		return err
	}
	registerForwardJoinRequest(ctx, up)

	keys := resp.SessionKeys
	keyEnvelopes := []*ttnpb.KeyEnvelope{keys.FNwkSIntKey, keys.NwkSEncKey, keys.SNwkSIntKey}
	if !dlSettings.OptNeg {
		keys.NwkSEncKey = keys.FNwkSIntKey
		keys.SNwkSIntKey = keys.FNwkSIntKey
		keyEnvelopes = keyEnvelopes[:1]
	}
	for _, keyEnvelope := range keyEnvelopes {
		unwrappedKey, err := cryptoutil.UnwrapAES128Key(ctx, keyEnvelope, ns.KeyService())
		if err != nil {
			return err
		}
		wrappedEnvelope, err := cryptoutil.WrapAES128Key(ctx, unwrappedKey, ns.deviceKEKLabel, ns.KeyService())
		if err != nil {
			return err
		}
		if err := keyEnvelope.SetFields(wrappedEnvelope, ttnpb.KeyEnvelopeFieldPathsTopLevel...); err != nil {
			return err
		}
	}
	macState.QueuedJoinAccept = &ttnpb.MACState_JoinAccept{
		CorrelationIds: resp.CorrelationIds,
		Keys:           keys,
		Payload:        resp.RawPayload,
		DevAddr:        devAddr.Bytes(),
		NetId:          ns.netID(ctx).Bytes(),
		Request: &ttnpb.MACState_JoinRequest{
			RxDelay:          macState.DesiredParameters.Rx1Delay,
			CfList:           cfList,
			DownlinkSettings: dlSettings,
		},
	}
	macState.RxWindowsAvailable = true
	ctx = events.ContextWithCorrelationID(ctx, resp.CorrelationIds...)

	if err := ns.deliverRelaySessionKeys(ctx, matched, keys.SessionKeyId); err != nil {
		return err
	}

	publishEvents(ctx, queuedEvents...)
	queuedEvents = nil
	up = ttnpb.Clone(up)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ns.deduplicationDone(ctx, up):
	}
	ns.mergeMetadata(ctx, up, initialDeduplicationRound)
	ns.filterMetadata(ctx, up)
	macState.RecentUplinks = appendRecentUplink(nil, up, recentUplinkCount)

	logger := log.FromContext(ctx)
	stored, storedCtx, err := ns.devices.SetByID(ctx, matched.Ids.ApplicationIds, matched.Ids.DeviceId,
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
			stored.PendingMacState = macState
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
	downAt := ttnpb.StdTime(up.ReceivedAt).Add(-infrastructureDelay/2 + phy.JoinAcceptDelay1 - macState.DesiredParameters.Rx1Delay.Duration()/2 - nsScheduleWindow())
	if earliestAt := time.Now().Add(nsScheduleWindow()); downAt.Before(earliestAt) {
		downAt = earliestAt
	}
	logger.WithField("start_at", downAt).Debug("Add downlink task")
	if err := ns.downlinkTasks.Add(ctx, stored.Ids, downAt, true); err != nil {
		logger.WithError(err).Error("Failed to add downlink task after join-request")
	}
	queuedEvents = append(queuedEvents, evtProcessJoinRequest.NewWithIdentifiersAndData(ctx, matched.Ids, up))
	registerProcessUplink(ctx, up)
	return nil
}

var errRejoinRequest = errors.DefineUnavailable("rejoin_request", "rejoin-request handling is not implemented")

func (ns *NetworkServer) handleRejoinRequest(ctx context.Context, up *ttnpb.UplinkMessage) error {
	defer trace.StartRegion(ctx, "handle rejoin request").End()

	// TODO: Implement https://github.com/TheThingsNetwork/lorawan-stack/issues/8
	return errRejoinRequest.New()
}

// HandleUplink is called by the Gateway Server when an uplink message arrives.
func (ns *NetworkServer) HandleUplink(ctx context.Context, up *ttnpb.UplinkMessage) (_ *emptypb.Empty, err error) {
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}

	ctx = events.ContextWithCorrelationID(ctx, up.CorrelationIds...)
	ctx = appendUplinkCorrelationID(ctx)
	up.CorrelationIds = events.CorrelationIDsFromContext(ctx)

	registerUplinkLatency(ctx, up)
	up.ReceivedAt = timestamppb.New(time.Now()) // NOTE: This is not equivalent to timestamppb.Now().

	up.Payload = &ttnpb.Message{}
	if err := lorawan.UnmarshalMessage(up.RawPayload, up.Payload); err != nil {
		return nil, errDecodePayload.WithCause(err)
	}
	if err := up.Payload.ValidateFields(); err != nil {
		return nil, errDecodePayload.WithCause(err)
	}
	registerReceiveUplink(ctx, up)
	defer func() {
		if errors.Is(err, errDuplicateUplink) {
			registerReceiveDuplicateUplink(ctx, up)
			return
		}
		if err != nil {
			registerDropUplink(ctx, up, err)
		}
	}()

	logger := log.FromContext(ctx).WithFields(log.Fields(
		"m_type", up.Payload.MHdr.MType,
		"major", up.Payload.MHdr.Major,
		"phy_payload_len", len(up.RawPayload),
		"received_at", ttnpb.StdTime(up.ReceivedAt),
		"frequency", up.Settings.Frequency,
	))
	switch dr := up.Settings.DataRate.Modulation.(type) {
	case *ttnpb.DataRate_Fsk:
		logger = logger.WithField(
			"bit_rate", dr.Fsk.GetBitRate(),
		)
	case *ttnpb.DataRate_Lora:
		logger = logger.WithFields(log.Fields(
			"bandwidth", dr.Lora.GetBandwidth(),
			"spreading_factor", dr.Lora.GetSpreadingFactor(),
		))
	case *ttnpb.DataRate_Lrfhss:
		logger = logger.WithFields(log.Fields(
			"modulation_type", dr.Lrfhss.GetModulationType(),
			"coding_rate", dr.Lrfhss.GetCodingRate(),
			"ocw", dr.Lrfhss.GetOperatingChannelWidth(),
		))
	default:
		return nil, errDataRateNotFound.WithAttributes("data_rate", up.Settings.DataRate)
	}
	ctx = log.NewContext(ctx, logger)

	if t, err := toa.Compute(len(up.RawPayload), up.Settings); err != nil {
		log.FromContext(ctx).WithError(err).Debug("Failed to compute time-on-air")
	} else {
		up.ConsumedAirtime = durationpb.New(t)
	}
	switch up.Payload.MHdr.MType {
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

var errTransmission = errors.Define("transmission", "downlink transmission failed with result `{result}`")

// ReportTxAcknowledgment is called by the Gateway Server when a tx acknowledgment arrives.
func (ns *NetworkServer) ReportTxAcknowledgment(
	ctx context.Context, txAck *ttnpb.GatewayTxAcknowledgment,
) (*emptypb.Empty, error) {
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}

	ack := txAck.GetTxAck()
	ctx = events.ContextWithCorrelationID(ctx, ack.CorrelationIds...)
	ctx = appendTxAckCorrelationID(ctx)

	down, err := ns.scheduledDownlinkMatcher.Match(ctx, ack)
	if err != nil {
		if errors.IsNotFound(err) {
			log.FromContext(ctx).Debug("Received TxAck but did not match scheduled downlink")
			return ttnpb.Empty, nil
		}
		return nil, err
	}

	ctx = events.ContextWithCorrelationID(ctx, down.CorrelationIds...)
	ack.CorrelationIds = events.CorrelationIDsFromContext(ctx)
	down.CorrelationIds = events.CorrelationIDsFromContext(ctx)

	queuedEvents := make([]events.Event, 0, 1)
	defer func() { events.Publish(queuedEvents...) }()

	ids := down.GetEndDeviceIds()
	var transmissionError *errors.Error
	switch ack.Result {
	case ttnpb.TxAcknowledgment_SUCCESS:
		queuedEvents = append(queuedEvents, evtTransmissionSuccess.NewWithIdentifiersAndData(ctx, ids, down))
	default:
		transmissionError = errTransmission.WithAttributes("result", ack.GetResult().String())
		queuedEvents = append(queuedEvents, evtTransmissionFail.NewWithIdentifiersAndData(ctx, ids, transmissionError))
	}

	macPayload := down.GetPayload().GetMacPayload()
	if macPayload.GetFPort() == 0 {
		return ttnpb.Empty, nil
	}
	appUp := &ttnpb.ApplicationUp{
		EndDeviceIds:   ids,
		CorrelationIds: ack.GetCorrelationIds(),
	}
	appDown := &ttnpb.ApplicationDownlink{
		SessionKeyId:   down.GetSessionKeyId(),
		FPort:          macPayload.GetFPort(),
		FCnt:           macPayload.GetFullFCnt(),
		FrmPayload:     macPayload.GetFrmPayload(),
		Confirmed:      down.GetPayload().GetMHdr().GetMType() == ttnpb.MType_CONFIRMED_DOWN,
		Priority:       down.GetRequest().GetPriority(),
		CorrelationIds: down.GetCorrelationIds(),
	}
	if transmissionError != nil {
		appUp.Up = &ttnpb.ApplicationUp_DownlinkFailed{
			DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
				Downlink: appDown,
				Error:    ttnpb.ErrorDetailsToProto(transmissionError),
			},
		}
	} else {
		appUp.Up = &ttnpb.ApplicationUp_DownlinkSent{DownlinkSent: appDown}
	}
	ns.submitApplicationUplinks(ctx, appUp)

	return ttnpb.Empty, nil
}
