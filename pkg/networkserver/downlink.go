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
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

// DownlinkTaskQueue represents an entity, that holds downlink tasks sorted by timestamp.
type DownlinkTaskQueue interface {
	// Add adds downlink task for device identified by devID at time t.
	// Implementations must ensure that Add returns fast.
	Add(ctx context.Context, devID ttnpb.EndDeviceIdentifiers, t time.Time, replace bool) error

	// Pop calls f on the most recent downlink task in the schedule, for which timestamp is in range [0, time.Now()],
	// if such is available, otherwise it blocks until it is.
	// Context passed to f must be derived from ctx.
	// Implementations must respect ctx.Deadline() value on best-effort basis, if such is present.
	Pop(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error
}

// DefaultClassCTimeout is the default time-out for the device to respond to class C downlink messages.
// When waiting for a response times out, the downlink message is considered lost, and the downlink task triggers again.
const DefaultClassCTimeout = 15 * time.Second

func deviceClassCTimeout(dev *ttnpb.EndDevice, defaults ttnpb.MACSettings) time.Duration {
	if dev.MACSettings != nil && dev.MACSettings.ClassCTimeout != nil {
		return *dev.MACSettings.ClassCTimeout
	}
	if defaults.ClassCTimeout != nil {
		return *defaults.ClassCTimeout
	}
	return DefaultClassCTimeout
}

func loggerWithApplicationDownlinkFields(logger log.Interface, down *ttnpb.ApplicationDownlink) log.Interface {
	pairs := []interface{}{
		"confirmed", down.Confirmed,
		"f_cnt", down.FCnt,
		"f_port", down.FPort,
		"frm_payload_len", len(down.FRMPayload),
		"priority", down.Priority,
		"session_key_id", down.SessionKeyID,
	}
	if down.GetClassBC() != nil {
		pairs = append(pairs, "class_b_c", true)
		if down.ClassBC.GetAbsoluteTime() != nil {
			pairs = append(pairs, "abs_time", *down.ClassBC.AbsoluteTime)
		}
		if len(down.ClassBC.GetGateways()) > 0 {
			pairs = append(pairs, "fixed_gateway_count", len(down.ClassBC.Gateways))
		}
	} else {
		pairs = append(pairs, "class_b_c", false)
	}
	return logger.WithFields(log.Fields(pairs...))
}

var errApplicationDownlinkTooLong = errors.DefineInvalidArgument("application_downlink_too_long", "application downlink payload is too long")
var errNoDownlink = errors.Define("no_downlink", "no downlink to send")

type generatedDownlink struct {
	Payload  []byte
	FCnt     uint32
	NeedsAck bool
	Priority ttnpb.TxSchedulePriority
}

type generateDownlinkState struct {
	baseApplicationUps        []*ttnpb.ApplicationUp
	ifScheduledApplicationUps []*ttnpb.ApplicationUp

	ApplicationDownlink      *ttnpb.ApplicationDownlink
	NeedsDownlinkQueueUpdate bool
	Events                   []events.Event
}

func (s generateDownlinkState) appendApplicationUplinks(ups []*ttnpb.ApplicationUp, scheduled bool) []*ttnpb.ApplicationUp {
	if !scheduled {
		return append(ups, s.baseApplicationUps...)
	} else {
		return append(append(ups, s.baseApplicationUps...), s.ifScheduledApplicationUps...)
	}
}

// generateDownlink attempts to generate a downlink.
// generateDownlink returns the generated downlink, application uplinks associated with the generation and error, if any.
// generateDownlink may mutate the device in order to record the downlink generated.
// maxDownLen and maxUpLen represent the maximum length of PHYPayload for the downlink and corresponding uplink respectively.
// If no downlink could be generated errNoDownlink is returned.
// generateDownlink does not perform validation of dev.MACState.DesiredParameters,
// hence, it could potentially generate MAC command(s), which are not suported by the
// regional parameters the device operates in.
// For example, a sequence of 'NewChannel' MAC commands could be generated for a
// device operating in a region where a fixed channel plan is defined in case
// dev.MACState.CurrentParameters.Channels is not equal to dev.MACState.DesiredParameters.Channels.
func (ns *NetworkServer) generateDownlink(ctx context.Context, dev *ttnpb.EndDevice, phy band.Band, class ttnpb.Class, maxDownLen, maxUpLen uint16) (*generatedDownlink, generateDownlinkState, error) {
	if dev.MACState == nil {
		return nil, generateDownlinkState{}, errUnknownMACState
	}
	if dev.Session == nil {
		return nil, generateDownlinkState{}, errEmptySession
	}

	ctx = log.NewContextWithFields(ctx, log.Fields(
		"device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers),
		"mac_version", dev.MACState.LoRaWANVersion,
		"phy_version", dev.LoRaWANPHYVersion,
	))
	logger := log.FromContext(ctx)

	// NOTE: len(MHDR) + len(MIC) = 1 + 4 = 5
	if maxDownLen < 5 || maxUpLen < 5 {
		panic("payload length limits too short to generate downlink")
	}
	maxDownLen, maxUpLen = maxDownLen-5, maxUpLen-5

	var fPending bool
	var st generateDownlinkState
	var skipAppDown bool
	var startIdx int
	for _, down := range dev.QueuedApplicationDownlinks {
		logger := loggerWithApplicationDownlinkFields(logger, down)

		switch {
		case len(down.FRMPayload) > int(maxDownLen):
			logger.WithField("max_down_len", maxDownLen).Debug("Drop application downlink with payload length exceeding band regulations")
			st.baseApplicationUps = append(st.baseApplicationUps, &ttnpb.ApplicationUp{
				EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
				CorrelationIDs:       append(events.CorrelationIDsFromContext(ctx), down.CorrelationIDs...),
				Up: &ttnpb.ApplicationUp_DownlinkFailed{
					DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
						ApplicationDownlink: *down,
						Error:               *ttnpb.ErrorDetailsToProto(errApplicationDownlinkTooLong),
					},
				},
			})
			startIdx++
			skipAppDown = true

		case down.FCnt <= dev.Session.LastNFCntDown && dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0:
			logger.WithField("last_f_cnt_down", dev.Session.LastNFCntDown).Debug("Drop application downlink with too low FCnt")
			st.baseApplicationUps = append(st.baseApplicationUps, &ttnpb.ApplicationUp{
				EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
				CorrelationIDs:       events.CorrelationIDsFromContext(ctx),
				Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
					DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{
						Downlinks:    dev.QueuedApplicationDownlinks,
						LastFCntDown: dev.Session.LastNFCntDown + 1,
					},
				},
			})
			startIdx = len(dev.QueuedApplicationDownlinks)
			skipAppDown = true

		case down.Confirmed && dev.Multicast:
			logger.Debug("Drop confirmed application downlink for multicast device")
			st.baseApplicationUps = append(st.baseApplicationUps, &ttnpb.ApplicationUp{
				EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
				CorrelationIDs:       events.CorrelationIDsFromContext(ctx),
				Up: &ttnpb.ApplicationUp_DownlinkFailed{
					DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
						ApplicationDownlink: *down,
						Error:               *ttnpb.ErrorDetailsToProto(errConfirmedMulticastDownlink),
					},
				},
			})
			startIdx++
			skipAppDown = true

		case down.ClassBC != nil:
			if down.ClassBC.AbsoluteTime != nil && down.ClassBC.AbsoluteTime.Before(time.Now()) {
				logger.Debug("Drop expired downlink")
				st.baseApplicationUps = append(st.baseApplicationUps, &ttnpb.ApplicationUp{
					EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
					CorrelationIDs:       append(events.CorrelationIDsFromContext(ctx), down.CorrelationIDs...),
					Up: &ttnpb.ApplicationUp_DownlinkFailed{
						DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
							ApplicationDownlink: *down,
							Error:               *ttnpb.ErrorDetailsToProto(errExpiredDownlink),
						},
					},
				})
				startIdx++
				continue
			}

			if class == ttnpb.CLASS_A {
				logger.Debug("Skip class B/C downlink for class A downlink")
				if dev.MACState.DeviceClass != ttnpb.CLASS_A && len(dev.MACState.QueuedResponses) == 0 {
					dev.QueuedApplicationDownlinks = dev.QueuedApplicationDownlinks[startIdx:]
					st.NeedsDownlinkQueueUpdate = startIdx > 0
					return nil, st, errNoDownlink
				}
				skipAppDown = true
			}
		}
		break
	}
	dev.QueuedApplicationDownlinks = dev.QueuedApplicationDownlinks[startIdx:]
	st.NeedsDownlinkQueueUpdate = startIdx > 0

	spec := lorawan.DefaultMACCommands
	cmds := make([]*ttnpb.MACCommand, 0, len(dev.MACState.QueuedResponses)+len(dev.MACState.PendingRequests))
	var lostResps []*ttnpb.MACCommand
	if class == ttnpb.CLASS_A {
		for i, cmd := range dev.MACState.QueuedResponses {
			desc := spec[cmd.CID]
			if desc == nil {
				lostResps = append(lostResps, dev.MACState.QueuedResponses[i:]...)
				maxDownLen = 0
				fPending = true
				break
			}
			if desc.DownlinkLength > maxDownLen {
				lostResps = append(lostResps, dev.MACState.QueuedResponses[i:]...)
				maxDownLen = 0
				fPending = true
				break
			}
			cmds = append(cmds, cmd)
			maxDownLen -= desc.DownlinkLength
		}
	}
	dev.MACState.QueuedResponses = nil
	dev.MACState.PendingRequests = dev.MACState.PendingRequests[:0]

	mType := ttnpb.MType_UNCONFIRMED_DOWN
	cmdBuf := make([]byte, 0, maxDownLen)
	if !dev.Multicast && len(lostResps) == 0 {
		enqueuers := make([]func(context.Context, *ttnpb.EndDevice, uint16, uint16) (uint16, uint16, bool), 0, 13)
		if dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_0) >= 0 {
			enqueuers = append(enqueuers,
				enqueueDutyCycleReq,
				enqueueRxParamSetupReq,
				func(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen uint16, maxUpLen uint16) (uint16, uint16, bool) {
					return enqueueDevStatusReq(ctx, dev, maxDownLen, maxUpLen, ns.defaultMACSettings)
				},
				enqueueNewChannelReq,
				func(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen uint16, maxUpLen uint16) (uint16, uint16, bool) {
					// NOTE: LinkADRReq must be enqueued after NewChannelReq.
					newMaxDownLen, newMaxUpLen, ok, err := enqueueLinkADRReq(ctx, dev, maxDownLen, maxUpLen, ns.FrequencyPlans)
					if err != nil {
						logger.WithError(err).Error("Failed to enqueue LinkADRReq")
						return maxDownLen, maxUpLen, false
					}
					return newMaxDownLen, newMaxUpLen, ok
				},
				enqueueRxTimingSetupReq,
			)
			if dev.MACState.DeviceClass == ttnpb.CLASS_B {
				if class == ttnpb.CLASS_A {
					enqueuers = append(enqueuers,
						enqueuePingSlotChannelReq,
					)
				}
				enqueuers = append(enqueuers,
					enqueueBeaconFreqReq,
				)
			}
		}
		if dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_0_2) >= 0 {
			if phy.TxParamSetupReqSupport {
				enqueuers = append(enqueuers,
					enqueueTxParamSetupReq,
				)
			}
			enqueuers = append(enqueuers,
				enqueueDLChannelReq,
			)
		}
		if dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) >= 0 {
			enqueuers = append(enqueuers,
				enqueueADRParamSetupReq,
				enqueueForceRejoinReq,
				enqueueRejoinParamSetupReq,
			)
		}

		for _, f := range enqueuers {
			var ok bool
			maxDownLen, maxUpLen, ok = f(ctx, dev, maxDownLen, maxUpLen)
			fPending = fPending || !ok
			// TODO: Buffer events https://github.com/TheThingsNetwork/lorawan-stack/issues/789
		}

		cmds = append(cmds, dev.MACState.PendingRequests...)
		for _, cmd := range cmds {
			logger := logger.WithField("cid", cmd.CID)
			logger.Debug("Add MAC command to buffer")
			var err error
			cmdBuf, err = spec.AppendDownlink(phy, cmdBuf, *cmd)
			if err != nil {
				return nil, generateDownlinkState{}, errEncodeMAC.WithCause(err)
			}
			if mType == ttnpb.MType_UNCONFIRMED_DOWN &&
				spec[cmd.CID].ExpectAnswer &&
				dev.MACState.DeviceClass == ttnpb.CLASS_C &&
				dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
				logger.Debug("Use confirmed downlink to get immediate answer")
				mType = ttnpb.MType_CONFIRMED_DOWN
			}
		}
	}
	logger = logger.WithField("mac_count", len(cmds))

	var needsDownlink bool
	var up *ttnpb.UplinkMessage
	if dev.MACState.RxWindowsAvailable && len(dev.RecentUplinks) > 0 {
		up = dev.RecentUplinks[len(dev.RecentUplinks)-1]
		switch up.Payload.MHDR.MType {
		case ttnpb.MType_UNCONFIRMED_UP:
			if up.Payload.GetMACPayload().FCtrl.ADRAckReq {
				logger.Debug("Need downlink for ADRAckReq")
				needsDownlink = true
			}
		case ttnpb.MType_CONFIRMED_UP:
			logger.Debug("Need downlink for confirmed uplink")
			needsDownlink = true
		}
	}

	pld := &ttnpb.MACPayload{
		FHDR: ttnpb.FHDR{
			DevAddr: dev.Session.DevAddr,
			FCtrl: ttnpb.FCtrl{
				Ack: up != nil && up.Payload.MHDR.MType == ttnpb.MType_CONFIRMED_UP,
				ADR: deviceUseADR(dev, ns.defaultMACSettings),
			},
		},
	}
	logger = logger.WithFields(log.Fields(
		"ack", pld.FHDR.FCtrl.Ack,
		"adr", pld.FHDR.FCtrl.ADR,
	))

	switch {
	case !skipAppDown && len(dev.QueuedApplicationDownlinks) > 0 && len(cmdBuf) <= fOptsCapacity:
		st.ApplicationDownlink = dev.QueuedApplicationDownlinks[0]
		dev.QueuedApplicationDownlinks = dev.QueuedApplicationDownlinks[1:]
		loggerWithApplicationDownlinkFields(logger, st.ApplicationDownlink).Debug("Add application downlink to buffer")

		pld.FHDR.FCnt = st.ApplicationDownlink.FCnt
		pld.FPort = st.ApplicationDownlink.FPort
		pld.FRMPayload = st.ApplicationDownlink.FRMPayload
		if st.ApplicationDownlink.Confirmed {
			mType = ttnpb.MType_CONFIRMED_DOWN
		}

	case len(cmdBuf) > 0, needsDownlink:
		pld.FHDR.FCnt = dev.Session.LastNFCntDown + 1

	default:
		return nil, st, errNoDownlink
	}
	logger = logger.WithFields(log.Fields(
		"f_cnt", pld.FHDR.FCnt,
		"f_port", pld.FPort,
		"m_type", mType,
	))

	if len(cmdBuf) > 0 && (pld.FPort == 0 || dev.MACState.LoRaWANVersion.EncryptFOpts()) {
		if dev.Session.NwkSEncKey == nil || len(dev.Session.NwkSEncKey.Key) == 0 {
			return nil, st, errUnknownNwkSEncKey
		}
		key, err := cryptoutil.UnwrapAES128Key(*dev.Session.NwkSEncKey, ns.KeyVault)
		if err != nil {
			logger.WithField("kek_label", dev.Session.NwkSEncKey.KEKLabel).WithError(err).Warn("Failed to unwrap NwkSEncKey")
			return nil, st, err
		}

		fCnt := pld.FHDR.FCnt
		if pld.FPort != 0 {
			fCnt = dev.Session.LastNFCntDown
		}
		cmdBuf, err = crypto.EncryptDownlink(key, dev.Session.DevAddr, fCnt, cmdBuf)
		if err != nil {
			return nil, st, errEncryptMAC.WithCause(err)
		}
	}
	if pld.FPort == 0 {
		pld.FRMPayload = cmdBuf
	} else {
		pld.FHDR.FOpts = cmdBuf
	}

	if pld.FPort == 0 && dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
		st.ifScheduledApplicationUps = append(st.ifScheduledApplicationUps, &ttnpb.ApplicationUp{
			EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
			CorrelationIDs:       events.CorrelationIDsFromContext(ctx),
			Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
				DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{
					Downlinks:    dev.QueuedApplicationDownlinks,
					LastFCntDown: pld.FHDR.FCnt,
				},
			},
		})
	}
	pld.FHDR.FCtrl.FPending = fPending || len(dev.QueuedApplicationDownlinks) > 0

	logger = logger.WithField("f_pending", pld.FHDR.FCtrl.FPending)

	needsAck := mType == ttnpb.MType_CONFIRMED_DOWN || len(dev.MACState.PendingRequests) > 0
	if needsAck &&
		dev.MACState.LastConfirmedDownlinkAt != nil &&
		dev.MACState.LastConfirmedDownlinkAt.Add(deviceClassCTimeout(dev, ns.defaultMACSettings)).After(time.Now()) {
		return nil, st, errConfirmedDownlinkTooSoon
	}

	b, err := lorawan.MarshalMessage(ttnpb.Message{
		MHDR: ttnpb.MHDR{
			MType: mType,
			Major: ttnpb.Major_LORAWAN_R1,
		},
		Payload: &ttnpb.Message_MACPayload{
			MACPayload: pld,
		},
	})
	if err != nil {
		return nil, st, errEncodePayload.WithCause(err)
	}
	// NOTE: It is assumed, that b does not contain MIC.

	if dev.Session.SNwkSIntKey == nil || len(dev.Session.SNwkSIntKey.Key) == 0 {
		return nil, st, errUnknownSNwkSIntKey
	}
	key, err := cryptoutil.UnwrapAES128Key(*dev.Session.SNwkSIntKey, ns.KeyVault)
	if err != nil {
		logger.WithField("kek_label", dev.Session.SNwkSIntKey.KEKLabel).WithError(err).Warn("Failed to unwrap SNwkSIntKey")
		return nil, st, err
	}

	var mic [4]byte
	if dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
		mic, err = crypto.ComputeLegacyDownlinkMIC(
			key,
			dev.Session.DevAddr,
			pld.FHDR.FCnt,
			b,
		)
	} else {
		var confFCnt uint32
		if pld.Ack {
			confFCnt = up.GetPayload().GetMACPayload().GetFCnt()
		}
		mic, err = crypto.ComputeDownlinkMIC(
			key,
			dev.Session.DevAddr,
			confFCnt,
			pld.FHDR.FCnt,
			b,
		)
	}
	if err != nil {
		return nil, st, errComputeMIC
	}
	b = append(b, mic[:]...)

	var priority ttnpb.TxSchedulePriority
	if st.ApplicationDownlink != nil {
		priority = st.ApplicationDownlink.Priority
		if max := ns.downlinkPriorities.MaxApplicationDownlink; priority > max {
			priority = max
		}
	}
	if (pld.FPort == 0 || len(cmdBuf) > 0) && priority < ns.downlinkPriorities.MACCommands {
		priority = ns.downlinkPriorities.MACCommands
	}

	logger.WithFields(log.Fields(
		"payload_length", len(b),
		"priority", priority,
	)).Debug("Generated downlink")
	return &generatedDownlink{
		Payload:  b,
		FCnt:     pld.FHDR.FCnt,
		NeedsAck: needsAck,
		Priority: priority,
	}, st, nil
}

type downlinkPath struct {
	ttnpb.GatewayIdentifiers
	*ttnpb.DownlinkPath
}

func downlinkPathsFromMetadata(mds ...*ttnpb.RxMetadata) []downlinkPath {
	mds = append(mds[:0:0], mds...)
	sort.SliceStable(mds, func(i, j int) bool {
		// TODO: Improve the sorting algorithm (https://github.com/TheThingsNetwork/lorawan-stack/issues/13)
		return mds[i].SNR > mds[j].SNR
	})
	head := make([]downlinkPath, 0, len(mds))
	tail := make([]downlinkPath, 0, len(mds))
	for _, md := range mds {
		if len(md.UplinkToken) == 0 || md.DownlinkPathConstraint == ttnpb.DOWNLINK_PATH_CONSTRAINT_NEVER {
			continue
		}

		path := downlinkPath{
			GatewayIdentifiers: md.GatewayIdentifiers,
			DownlinkPath: &ttnpb.DownlinkPath{
				Path: &ttnpb.DownlinkPath_UplinkToken{
					UplinkToken: md.UplinkToken,
				},
			},
		}
		switch md.DownlinkPathConstraint {
		case ttnpb.DOWNLINK_PATH_CONSTRAINT_NONE:
			head = append(head, path)

		case ttnpb.DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER:
			tail = append(tail, path)
		}
	}
	return append(head, tail...)
}

// downlinkPathsForClassA returns the last paths, if any, of the given uplink messages.
// This function returns whether class A downlink can be made in either window considering the given Rx delay.
func downlinkPathsForClassA(rxDelay ttnpb.RxDelay, ups ...*ttnpb.UplinkMessage) (rx1, rx2 bool, paths []downlinkPath) {
	if rxDelay == ttnpb.RX_DELAY_0 {
		rxDelay = ttnpb.RX_DELAY_1
	}
	maxDelta := time.Duration(rxDelay) * time.Second
	for i := len(ups) - 1; i >= 0; i-- {
		up := ups[i]
		delta := time.Since(up.ReceivedAt)
		rx1, rx2 := delta < maxDelta, delta < maxDelta+time.Second
		if paths := downlinkPathsFromMetadata(up.RxMetadata...); len(paths) > 0 {
			return rx1, rx2, paths
		}
	}
	return false, false, nil
}

func downlinkPathsFromRecentUplinks(ups ...*ttnpb.UplinkMessage) []downlinkPath {
	for i := len(ups) - 1; i >= 0; i-- {
		if paths := downlinkPathsFromMetadata(ups[i].RxMetadata...); len(paths) > 0 {
			return paths
		}
	}
	return nil
}

type scheduledDownlink struct {
	Message    *ttnpb.DownlinkMessage
	TransmitAt time.Time
}

type downlinkSchedulingError []error

func (errs downlinkSchedulingError) Error() string {
	return errSchedule.Error()
}

// pathErrors returns path errors represented by errs and boolean
// indicating whether all errors in errs represent path errors.
func (errs downlinkSchedulingError) pathErrors() ([]error, bool) {
	pathErrs := make([]error, 0, len(errs))
	allOK := true
	for _, gsErr := range errs {
		ttnErr, ok := errors.From(gsErr)
		if !ok {
			allOK = false
			continue
		}

		var ds []*ttnpb.ScheduleDownlinkErrorDetails
		for _, msg := range ttnErr.Details() {
			d, ok := msg.(*ttnpb.ScheduleDownlinkErrorDetails)
			if !ok {
				continue
			}
			ds = append(ds, d)
		}
		if len(ds) == 0 {
			allOK = false
			continue
		}
		for _, d := range ds {
			for _, pErr := range d.PathErrors {
				pathErrs = append(pathErrs, ttnpb.ErrorDetailsFromProto(pErr))
			}
		}
	}
	return pathErrs, allOK
}

// allErrors returns true if p(err) == true for each err in errs and false otherwise.
func allErrors(p func(error) bool, errs ...error) bool {
	for _, err := range errs {
		if !p(err) {
			return false
		}
	}
	return true
}

func nonRetryableAbsoluteTimeGatewayError(err error) bool {
	return errors.IsAborted(err) || errors.IsResourceExhausted(err) || errors.IsFailedPrecondition(err)
}

func nonRetryableFixedPathGatewayError(err error) bool {
	return errors.IsNotFound(err) || errors.IsDataLoss(err) || errors.IsFailedPrecondition(err)
}

// scheduleDownlinkByPaths attempts to schedule payload b using parameters in req using paths.
// scheduleDownlinkByPaths discards req.DownlinkPaths and mutates it arbitrarily.
// scheduleDownlinkByPaths returns the scheduled downlink or error.
func (ns *NetworkServer) scheduleDownlinkByPaths(ctx context.Context, req *ttnpb.TxRequest, b []byte, paths ...downlinkPath) (*scheduledDownlink, error) {
	if len(paths) == 0 {
		return nil, errNoPath
	}

	logger := log.FromContext(ctx)

	type attempt struct {
		peer  cluster.Peer
		paths []*ttnpb.DownlinkPath
	}
	attempts := make([]*attempt, 0, len(paths))

	for _, path := range paths {
		logger := logger.WithField(
			"gateway_uid", unique.ID(ctx, path.GatewayIdentifiers),
		)

		p, err := ns.GetPeer(ctx, ttnpb.ClusterRole_GATEWAY_SERVER, path.GatewayIdentifiers)
		if err != nil {
			logger.WithError(err).Debug("Could not get Gateway Server")
			continue
		}

		var a *attempt
		if len(attempts) > 0 && attempts[len(attempts)-1].peer == p {
			a = attempts[len(attempts)-1]
		} else {
			a = &attempt{
				peer: p,
			}
			attempts = append(attempts, a)
		}
		a.paths = append(a.paths, path.DownlinkPath)
	}

	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("ns:downlink:%s", events.NewCorrelationID()))
	errs := make([]error, 0, len(attempts))
	for _, a := range attempts {
		req.DownlinkPaths = a.paths
		down := &ttnpb.DownlinkMessage{
			RawPayload:     b,
			CorrelationIDs: events.CorrelationIDsFromContext(ctx),
			Settings: &ttnpb.DownlinkMessage_Request{
				Request: req,
			},
		}

		logger.WithField("path_count", len(req.DownlinkPaths)).Debug("Schedule downlink")
		cc, err := a.peer.Conn()
		if err != nil {
			errs = append(errs, err)
			continue
		}
		res, err := ttnpb.NewNsGsClient(cc).ScheduleDownlink(ctx, down, ns.WithClusterAuth())
		if err != nil {
			errs = append(errs, err)
			continue
		}
		logger.WithField("delay", res.Delay).Debug("Scheduled downlink")
		return &scheduledDownlink{
			Message:    down,
			TransmitAt: time.Now().Add(res.Delay),
		}, nil
	}
	return nil, downlinkSchedulingError(errs)
}

func loggerWithTxRequestFields(logger log.Interface, req *ttnpb.TxRequest, rx1, rx2 bool) log.Interface {
	pairs := []interface{}{
		"attempt_rx1", rx1,
		"attempt_rx2", rx2,
		"downlink_class", req.Class,
		"downlink_priority", req.Priority,
	}
	if rx1 {
		pairs = append(pairs,
			"rx1_data_rate", req.Rx1DataRateIndex,
			"rx1_frequency", req.Rx1Frequency,
		)
	}
	if rx2 {
		pairs = append(pairs,
			"rx2_data_rate", req.Rx2DataRateIndex,
			"rx2_frequency", req.Rx2Frequency,
		)
	}
	return logger.WithFields(log.Fields(pairs...))
}

func loggerWithDownlinkSchedulingErrorFields(logger log.Interface, errs downlinkSchedulingError) log.Interface {
	pairs := []interface{}{
		"attempts", len(errs),
	}
	for i, err := range errs {
		pairs = append(pairs, fmt.Sprintf("error_%d", i), err)
	}
	return logger.WithFields(log.Fields(pairs...))
}

func appendRecentDownlink(recent []*ttnpb.DownlinkMessage, down *ttnpb.DownlinkMessage, window int) []*ttnpb.DownlinkMessage {
	recent = append(recent, down)
	if len(recent) > window {
		recent = recent[len(recent)-window:]
	}
	return recent
}

// txRequestFromUplink return the Class A TxRequest, which can be used to answer up.
// txRequestFromUplink does not set the priority.
func txRequestFromUplink(phy band.Band, macState *ttnpb.MACState, rx1, rx2 bool, rxDelay ttnpb.RxDelay, up *ttnpb.UplinkMessage) (*ttnpb.TxRequest, error) {
	if !rx1 && !rx2 {
		return nil, errNoPath
	}
	req := &ttnpb.TxRequest{
		Class:    ttnpb.CLASS_A,
		Rx1Delay: rxDelay,
	}
	if rx1 {
		if up.DeviceChannelIndex > math.MaxUint8 {
			return nil, errInvalidChannelIndex
		}
		rx1ChIdx, err := phy.Rx1Channel(uint8(up.DeviceChannelIndex))
		if err != nil {
			return nil, err
		}
		if uint(rx1ChIdx) >= uint(len(macState.CurrentParameters.Channels)) ||
			macState.CurrentParameters.Channels[int(rx1ChIdx)] == nil ||
			macState.CurrentParameters.Channels[int(rx1ChIdx)].DownlinkFrequency == 0 {
			return nil, errCorruptedMACState
		}
		rx1DRIdx, err := phy.Rx1DataRate(up.Settings.DataRateIndex, macState.CurrentParameters.Rx1DataRateOffset, macState.CurrentParameters.DownlinkDwellTime.GetValue())
		if err != nil {
			return nil, err
		}
		req.Rx1DataRateIndex = rx1DRIdx
		req.Rx1Frequency = macState.CurrentParameters.Channels[int(rx1ChIdx)].DownlinkFrequency
	}
	if rx2 {
		req.Rx2DataRateIndex = macState.CurrentParameters.Rx2DataRateIndex
		req.Rx2Frequency = macState.CurrentParameters.Rx2Frequency
	}
	return req, nil
}

// maximumUplinkLength returns the maximum length of the next uplink after ups.
func maximumUplinkLength(fp *frequencyplans.FrequencyPlan, phy band.Band, ups ...*ttnpb.UplinkMessage) uint16 {
	// NOTE: If no data uplink is found, we assume ADR is off on the device and, hence, data rate index 0 is used in computation.
	maxUpDRIdx := ttnpb.DATA_RATE_0
loop:
	for i := len(ups) - 1; i >= 0; i-- {
		switch ups[i].Payload.MHDR.MType {
		case ttnpb.MType_JOIN_REQUEST:
			break loop
		case ttnpb.MType_UNCONFIRMED_UP, ttnpb.MType_CONFIRMED_UP:
			if ups[i].Payload.GetMACPayload().FHDR.FCtrl.ADR {
				maxUpDRIdx = ups[i].Settings.DataRateIndex
			}
			break loop
		}
	}
	return phy.DataRates[maxUpDRIdx].DefaultMaxSize.PayloadSize(fp.DwellTime.GetUplinks())
}

// downlinkRetryInterval is the time interval, which defines the interval between downlink task retries.
const downlinkRetryInterval = time.Second

// gsScheduleWindow is the time interval, which is sufficient for GS to ensure downlink is scheduled.
const gsScheduleWindow = 30 * time.Second

// nsScheduleWindow is the time interval, which is sufficient for NS to ensure downlink is scheduled.
var nsScheduleWindow = time.Second

func (ns *NetworkServer) recordDownlink(dev *ttnpb.EndDevice, genDown *generatedDownlink, genState generateDownlinkState, down *scheduledDownlink) (nextDownlinkAt time.Time) {
	if genState.ApplicationDownlink == nil || dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 && genDown.FCnt > dev.Session.LastNFCntDown {
		dev.Session.LastNFCntDown = genDown.FCnt
	}
	if genDown.NeedsAck {
		dev.MACState.LastConfirmedDownlinkAt = timePtr(down.TransmitAt)
	}
	if genState.ApplicationDownlink != nil && genState.ApplicationDownlink.Confirmed {
		dev.MACState.PendingApplicationDownlink = genState.ApplicationDownlink
		dev.Session.LastConfFCntDown = genDown.FCnt
	}
	if dev.MACState.DeviceClass == ttnpb.CLASS_C {
		var nextConfirmedAt time.Time
		if dev.MACState.LastConfirmedDownlinkAt != nil {
			nextConfirmedAt = dev.MACState.LastConfirmedDownlinkAt.Add(deviceClassCTimeout(dev, ns.defaultMACSettings))
		}
		if nextConfirmedAt.After(down.TransmitAt) {
			nextDownlinkAt = nextConfirmedAt
		} else {
			nextDownlinkAt = down.TransmitAt.Add(dev.MACState.CurrentParameters.Rx1Delay.Duration())
		}
	}
	dev.MACState.QueuedResponses = nil
	dev.MACState.RxWindowsAvailable = false
	dev.RecentDownlinks = appendRecentDownlink(dev.RecentDownlinks, down.Message, recentDownlinkCount)
	return nextDownlinkAt
}

// processDownlinkTask processes the most recent downlink task ready for execution, if such is available or wait until it is before processing it.
// NOTE: ctx.Done() is not guaranteed to be respected by processDownlinkTask.
func (ns *NetworkServer) processDownlinkTask(ctx context.Context) error {
	var setErr bool
	var addErr bool
	err := ns.downlinkTasks.Pop(ctx, func(ctx context.Context, devID ttnpb.EndDeviceIdentifiers, t time.Time) error {
		logger := log.FromContext(ctx).WithFields(log.Fields(
			"device_uid", unique.ID(ctx, devID),
			"started_at", time.Now().UTC(),
		))
		ctx = log.NewContext(ctx, logger)
		logger.WithField("start_at", t).Debug("Process downlink task")

		var queuedApplicationUplinks []*ttnpb.ApplicationUp
		var queuedEvents []events.Event
		var nextDownlinkAt time.Time
		_, err := ns.devices.SetByID(ctx, devID.ApplicationIdentifiers, devID.DeviceID,
			[]string{
				"frequency_plan_id",
				"last_dev_status_received_at",
				"lorawan_phy_version",
				"mac_settings",
				"mac_state",
				"multicast",
				"pending_mac_state",
				"queued_application_downlinks",
				"recent_downlinks",
				"recent_uplinks",
				"session",
			},
			func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
				if dev == nil {
					logger.Warn("Device not found")
					return nil, nil, nil
				}

				fp, phy, err := getDeviceBandVersion(dev, ns.FrequencyPlans)
				if err != nil {
					nextDownlinkAt = time.Now().Add(downlinkRetryInterval).UTC()
					logger.WithField("retry_at", nextDownlinkAt).WithError(err).Error("Failed to get frequency plan of the device, retry downlink slot")
					return dev, nil, nil
				}

				if dev.PendingMACState != nil &&
					dev.PendingMACState.PendingJoinRequest == nil &&
					dev.PendingMACState.RxWindowsAvailable &&
					dev.PendingMACState.QueuedJoinAccept != nil {

					logger = logger.WithField("downlink_type", "join-accept")
					if len(dev.RecentUplinks) == 0 {
						logger.Warn("No recent uplinks found, skip downlink slot")
						return dev, nil, nil
					}
					up := dev.RecentUplinks[len(dev.RecentUplinks)-1]
					switch up.Payload.MHDR.MType {
					case ttnpb.MType_JOIN_REQUEST, ttnpb.MType_REJOIN_REQUEST:
					default:
						logger.Warn("Last uplink is neither join-request, nor rejoin-request, skip downlink slot")
						return dev, nil, nil
					}
					ctx := events.ContextWithCorrelationID(ctx, up.CorrelationIDs...)

					rxDelay := ttnpb.RxDelay(phy.JoinAcceptDelay1 / time.Second)

					// Join-accept downlink for Class A/B/C device in Rx1/Rx2
					rx1, rx2, paths := downlinkPathsForClassA(rxDelay, dev.RecentUplinks...)
					if !rx1 && !rx2 {
						logger.Warn("Rx1 and Rx2 are expired, skip downlink slot")
						dev.PendingMACState.RxWindowsAvailable = false
						return dev, []string{
							"pending_mac_state.rx_windows_available",
						}, nil
					}
					if len(paths) == 0 {
						logger.Warn("No downlink path available, skip downlink slot")
						return dev, nil, nil
					}

					req, err := txRequestFromUplink(phy, dev.PendingMACState, rx1, rx2, rxDelay, up)
					if err != nil {
						logger.WithError(err).Warn("Failed to generate Tx request from uplink, skip downlink slot")
						return dev, nil, nil
					}
					req.Priority = ns.downlinkPriorities.JoinAccept

					down, err := ns.scheduleDownlinkByPaths(
						log.NewContext(ctx, loggerWithTxRequestFields(logger, req, rx1, rx2).WithField("rx1_delay", req.Rx1Delay)),
						req,
						dev.PendingMACState.QueuedJoinAccept.Payload,
						paths...,
					)
					if err != nil {
						if schedErr, ok := err.(downlinkSchedulingError); ok {
							logger = loggerWithDownlinkSchedulingErrorFields(logger, schedErr)
						} else {
							logger = logger.WithError(err)
						}
						logger.Warn("All Gateway Servers failed to schedule downlink, skip downlink slot")
						return dev, nil, nil
					}

					dev.PendingSession = &ttnpb.Session{
						DevAddr:     dev.PendingMACState.QueuedJoinAccept.Request.DevAddr,
						SessionKeys: dev.PendingMACState.QueuedJoinAccept.Keys,
					}
					dev.PendingMACState.PendingJoinRequest = &dev.PendingMACState.QueuedJoinAccept.Request
					dev.PendingMACState.QueuedJoinAccept = nil
					dev.PendingMACState.RxWindowsAvailable = false
					dev.RecentDownlinks = appendRecentDownlink(dev.RecentDownlinks, down.Message, recentDownlinkCount)
					return dev, []string{
						"pending_mac_state.pending_join_request",
						"pending_mac_state.queued_join_accept",
						"pending_mac_state.rx_windows_available",
						"pending_session.dev_addr",
						"pending_session.keys",
						"recent_downlinks",
					}, nil
				}

				logger = logger.WithField("downlink_type", "data")

				if dev.MACState == nil {
					logger.Warn("Unknown MAC state, skip downlink slot")
					return dev, nil, nil
				}

				logger = logger.WithField("device_class", dev.MACState.DeviceClass)

				if !dev.MACState.RxWindowsAvailable {
					logger.Debug("Rx windows not available, skip class A downlink slot")
					if dev.MACState.DeviceClass == ttnpb.CLASS_A {
						return dev, nil, nil
					}
				}

				var maxUpLength uint16 = math.MaxUint16
				if dev.MACState.LoRaWANVersion == ttnpb.MAC_V1_1 {
					maxUpLength = maximumUplinkLength(fp, phy, dev.RecentUplinks...)
				}

				var sets []string
			outer:
				for dev.MACState.RxWindowsAvailable {
					if len(dev.RecentUplinks) == 0 {
						logger.Warn("Rx windows available, but no uplink present, skip class A downlink slot")
						dev.MACState.QueuedResponses = nil
						dev.MACState.RxWindowsAvailable = false
						sets = append(sets,
							"mac_state.queued_responses",
							"mac_state.rx_windows_available",
						)
						break
					}

					var rxDelay ttnpb.RxDelay
					up := dev.RecentUplinks[len(dev.RecentUplinks)-1]
					switch up.Payload.MHDR.MType {
					case ttnpb.MType_CONFIRMED_UP, ttnpb.MType_UNCONFIRMED_UP:
						rxDelay = dev.MACState.CurrentParameters.Rx1Delay
						if rxDelay == ttnpb.RX_DELAY_0 {
							rxDelay = ttnpb.RX_DELAY_1
						}

					case ttnpb.MType_REJOIN_REQUEST:
						rxDelay = ttnpb.RxDelay(phy.JoinAcceptDelay1 / time.Second)

					default:
						logger.Warn("Last uplink is neither data uplink, nor rejoin-request, skip class A downlink slot")
						dev.MACState.QueuedResponses = nil
						dev.MACState.RxWindowsAvailable = false
						sets = append(sets,
							"mac_state.queued_responses",
							"mac_state.rx_windows_available",
						)
						break outer
					}
					ctx = events.ContextWithCorrelationID(ctx, up.CorrelationIDs...)

					rx1, rx2, paths := downlinkPathsForClassA(rxDelay, dev.RecentUplinks...)
					if !rx1 && !rx2 {
						logger.Warn("Rx1 and Rx2 are expired, skip class A downlink slot")
						dev.MACState.QueuedResponses = nil
						dev.MACState.RxWindowsAvailable = false
						sets = append(sets,
							"mac_state.queued_responses",
							"mac_state.rx_windows_available",
						)
						break
					}
					if len(paths) == 0 {
						logger.Warn("No downlink path available, skip class A downlink slot")
						break
					}
					if dev.MACState.DeviceClass != ttnpb.CLASS_A && !rx1 && rx2 && len(dev.MACState.QueuedResponses) == 0 {
						break
					}

					req, err := txRequestFromUplink(phy, dev.MACState, rx1, rx2, rxDelay, up)
					if err != nil {
						logger.WithError(err).Warn("Failed to generate Tx request from uplink, skip class A downlink slot")
						break
					}

					var maxDR ttnpb.DataRateIndex
					switch {
					case rx1 && rx2:
						maxDR = req.Rx1DataRateIndex
						if req.Rx2DataRateIndex > maxDR {
							maxDR = req.Rx2DataRateIndex
						}

					case rx1:
						maxDR = req.Rx1DataRateIndex

					case rx2:
						maxDR = req.Rx2DataRateIndex
					}

					genDown, genState, err := ns.generateDownlink(ctx, dev, phy, ttnpb.CLASS_A,
						phy.DataRates[maxDR].DefaultMaxSize.PayloadSize(fp.DwellTime.GetDownlinks()),
						maxUpLength,
					)
					if genState.NeedsDownlinkQueueUpdate {
						sets = append(sets,
							"queued_application_downlinks",
						)
					}
					if err != nil {
						switch {
						case errors.Resemble(err, errNoDownlink):
							logger.Debug("No class A downlink to send, skip class A downlink slot")

						default:
							logger.WithError(err).Warn("Failed to generate class A downlink, skip class A downlink slot")
						}
						queuedApplicationUplinks = genState.appendApplicationUplinks(queuedApplicationUplinks, false)
						if genState.ApplicationDownlink != nil {
							dev.QueuedApplicationDownlinks = append([]*ttnpb.ApplicationDownlink{genState.ApplicationDownlink}, dev.QueuedApplicationDownlinks...)
						}
						break
					}

					if rx1 && rx2 {
						rx1 = len(genDown.Payload) <= int(phy.DataRates[req.Rx1DataRateIndex].DefaultMaxSize.PayloadSize(fp.DwellTime.GetDownlinks()))
						rx2 = len(genDown.Payload) <= int(phy.DataRates[req.Rx2DataRateIndex].DefaultMaxSize.PayloadSize(fp.DwellTime.GetDownlinks()))
						if !rx1 && !rx2 {
							logger.Error("Generated downlink payload size does not fit neither Rx1, nor Rx2, skip class A downlink slot")
							break
						}
						req, err = txRequestFromUplink(phy, dev.MACState, rx1, rx2, rxDelay, up)
						if err != nil {
							logger.WithError(err).Warn("Failed to generate Tx request from uplink, skip class A downlink slot")
							queuedApplicationUplinks = genState.appendApplicationUplinks(queuedApplicationUplinks, false)
							if genState.ApplicationDownlink != nil {
								dev.QueuedApplicationDownlinks = append([]*ttnpb.ApplicationDownlink{genState.ApplicationDownlink}, dev.QueuedApplicationDownlinks...)
							}
							break
						}
					}

					if genState.ApplicationDownlink != nil {
						ctx = events.ContextWithCorrelationID(ctx, genState.ApplicationDownlink.CorrelationIDs...)
					}
					req.Priority = genDown.Priority

					down, err := ns.scheduleDownlinkByPaths(
						log.NewContext(ctx, loggerWithTxRequestFields(logger, req, rx1, rx2).WithField("rx1_delay", req.Rx1Delay)),
						req,
						genDown.Payload,
						paths...,
					)
					if err != nil {
						if schedErr, ok := err.(downlinkSchedulingError); ok {
							logger = loggerWithDownlinkSchedulingErrorFields(logger, schedErr)
						} else {
							logger = logger.WithError(err)
						}
						logger.Warn("All Gateway Servers failed to schedule downlink, skip class A downlink slot")
						queuedApplicationUplinks = genState.appendApplicationUplinks(queuedApplicationUplinks, false)
						if genState.ApplicationDownlink != nil {
							dev.QueuedApplicationDownlinks = append([]*ttnpb.ApplicationDownlink{genState.ApplicationDownlink}, dev.QueuedApplicationDownlinks...)
						}
						break
					}

					if t := ns.recordDownlink(dev, genDown, genState, down); !t.IsZero() {
						nextDownlinkAt = t
					}
					queuedApplicationUplinks = genState.appendApplicationUplinks(queuedApplicationUplinks, true)
					queuedEvents = append(queuedEvents, genState.Events...)
					return dev, append(sets,
						"mac_state",
						"queued_application_downlinks",
						"recent_downlinks",
						"session",
					), nil
				}

				switch dev.MACState.DeviceClass {
				case ttnpb.CLASS_A:
					return dev, sets, nil

				case ttnpb.CLASS_B:
					// TODO: Support Class B (https://github.com/TheThingsNetwork/lorawan-stack/issues/19).
					logger.Warn("Class B downlinks are not supported, skip class B/C downlink slot")
					return dev, sets, nil
				}

				// Class B/C data downlink
				req := &ttnpb.TxRequest{
					Class:            dev.MACState.DeviceClass,
					Rx2DataRateIndex: dev.MACState.CurrentParameters.Rx2DataRateIndex,
					Rx2Frequency:     dev.MACState.CurrentParameters.Rx2Frequency,
				}

				genDown, genState, err := ns.generateDownlink(ctx, dev, phy, dev.MACState.DeviceClass,
					phy.DataRates[req.Rx2DataRateIndex].DefaultMaxSize.PayloadSize(fp.DwellTime.GetDownlinks()),
					maxUpLength,
				)
				if genState.NeedsDownlinkQueueUpdate {
					sets = []string{
						"queued_application_downlinks",
					}
				}
				if err != nil {
					switch {
					case errors.Resemble(err, errNoDownlink):
						logger.Debug("No class B/C downlink to send, skip class B/C downlink slot")

					case errors.Resemble(err, errConfirmedDownlinkTooSoon):
						nextDownlinkAt = dev.MACState.LastConfirmedDownlinkAt.Add(deviceClassCTimeout(dev, ns.defaultMACSettings))
						logger.WithField("retry_at", nextDownlinkAt).Info("Confirmed downlink scheduled too soon, retry")

					default:
						logger.WithError(err).Warn("Failed to generate class B/C downlink, skip class B/C downlink slot")
					}
					queuedApplicationUplinks = genState.appendApplicationUplinks(queuedApplicationUplinks, false)
					if genState.ApplicationDownlink != nil && ttnpb.HasAnyField(sets, "queued_application_downlinks") {
						dev.QueuedApplicationDownlinks = append([]*ttnpb.ApplicationDownlink{genState.ApplicationDownlink}, dev.QueuedApplicationDownlinks...)
					}
					return dev, sets, nil
				}

				if genState.ApplicationDownlink != nil {
					ctx = events.ContextWithCorrelationID(ctx, genState.ApplicationDownlink.CorrelationIDs...)
				}
				req.Priority = genDown.Priority

				var paths []downlinkPath
				fixedPaths := genState.ApplicationDownlink.GetClassBC().GetGateways()
				if len(fixedPaths) > 0 {
					paths = make([]downlinkPath, 0, len(fixedPaths))
					for _, gtw := range fixedPaths {
						paths = append(paths, downlinkPath{
							GatewayIdentifiers: gtw.GatewayIdentifiers,
							DownlinkPath: &ttnpb.DownlinkPath{
								Path: &ttnpb.DownlinkPath_Fixed{
									Fixed: gtw,
								},
							},
						})
					}
				} else {
					paths = downlinkPathsFromRecentUplinks(dev.RecentUplinks...)
					if len(paths) == 0 {
						logger.Warn("No downlink path available, skip class B/C downlink slot")
						queuedApplicationUplinks = genState.appendApplicationUplinks(queuedApplicationUplinks, false)
						if genState.ApplicationDownlink != nil && ttnpb.HasAnyField(sets, "queued_application_downlinks") {
							dev.QueuedApplicationDownlinks = append([]*ttnpb.ApplicationDownlink{genState.ApplicationDownlink}, dev.QueuedApplicationDownlinks...)
						}
						return dev, sets, nil
					}
				}
				if absTime := genState.ApplicationDownlink.GetClassBC().GetAbsoluteTime(); absTime != nil {
					if absTime.After(time.Now().Add(gsScheduleWindow)) {
						nextDownlinkAt = absTime.Add(-gsScheduleWindow)
						logger.WithField("retry_at", nextDownlinkAt).Info("Downlink scheduled too soon, retry attempt")
						queuedApplicationUplinks = genState.appendApplicationUplinks(queuedApplicationUplinks, false)
						if genState.ApplicationDownlink != nil && ttnpb.HasAnyField(sets, "queued_application_downlinks") {
							dev.QueuedApplicationDownlinks = append([]*ttnpb.ApplicationDownlink{genState.ApplicationDownlink}, dev.QueuedApplicationDownlinks...)
						}
						return dev, sets, nil
					}
					req.AbsoluteTime = absTime
				}

				down, err := ns.scheduleDownlinkByPaths(
					log.NewContext(ctx, loggerWithTxRequestFields(logger, req, false, true)),
					req,
					genDown.Payload,
					paths...,
				)
				if err != nil {
					nextDownlinkAt = time.Now().Add(downlinkRetryInterval).UTC()
					schedErr, ok := err.(downlinkSchedulingError)
					if ok {
						logger = loggerWithDownlinkSchedulingErrorFields(logger, schedErr)
					} else {
						logger = logger.WithError(err)
					}
					queuedApplicationUplinks = genState.appendApplicationUplinks(queuedApplicationUplinks, false)
					if ok && genState.ApplicationDownlink != nil {
						pathErrs, ok := schedErr.pathErrors()
						if ok {
							if genState.ApplicationDownlink.GetClassBC().GetAbsoluteTime() != nil &&
								allErrors(nonRetryableAbsoluteTimeGatewayError, pathErrs...) {
								logger.WithField("retry_at", nextDownlinkAt).Warn("Absolute time invalid, fail downlink and retry attempt")
								queuedApplicationUplinks = append(queuedApplicationUplinks, &ttnpb.ApplicationUp{
									EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
									CorrelationIDs:       events.CorrelationIDsFromContext(ctx),
									Up: &ttnpb.ApplicationUp_DownlinkFailed{
										DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
											ApplicationDownlink: *genState.ApplicationDownlink,
											Error:               *ttnpb.ErrorDetailsToProto(errInvalidAbsoluteTime),
										},
									},
								})
								if !genState.NeedsDownlinkQueueUpdate {
									sets = append(sets, "queued_application_downlinks")
								}
								return dev, sets, nil
							}

							if len(genState.ApplicationDownlink.GetClassBC().GetGateways()) > 0 &&
								allErrors(nonRetryableFixedPathGatewayError, pathErrs...) {
								logger.WithField("retry_at", nextDownlinkAt).Warn("Fixed paths invalid, fail application downlink and retry attempt")
								queuedApplicationUplinks = append(queuedApplicationUplinks, &ttnpb.ApplicationUp{
									EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
									CorrelationIDs:       events.CorrelationIDsFromContext(ctx),
									Up: &ttnpb.ApplicationUp_DownlinkFailed{
										DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
											ApplicationDownlink: *genState.ApplicationDownlink,
											Error:               *ttnpb.ErrorDetailsToProto(errInvalidFixedPaths),
										},
									},
								})
								if !genState.NeedsDownlinkQueueUpdate {
									sets = append(sets, "queued_application_downlinks")
								}
								return dev, sets, nil
							}
						}
					}
					logger.Warn("All Gateway Servers failed to schedule downlink, skip class B/C downlink slot")
					if genState.NeedsDownlinkQueueUpdate {
						dev.QueuedApplicationDownlinks = append([]*ttnpb.ApplicationDownlink{genState.ApplicationDownlink}, dev.QueuedApplicationDownlinks...)
					}
					return dev, sets, nil
				}

				if t := ns.recordDownlink(dev, genDown, genState, down); !t.IsZero() {
					nextDownlinkAt = t
				}
				queuedApplicationUplinks = genState.appendApplicationUplinks(queuedApplicationUplinks, true)
				queuedEvents = append(queuedEvents, genState.Events...)
				return dev, []string{
					"mac_state",
					"queued_application_downlinks",
					"recent_downlinks",
					"session",
				}, nil
			},
		)
		if len(queuedApplicationUplinks) > 0 {
			go func() {
				// TODO: Enqueue the uplinks and let a task processor take care of them.
				// (https://github.com/TheThingsNetwork/lorawan-stack/issues/560)
				for _, up := range queuedApplicationUplinks {
					ok, err := ns.handleASUplink(ctx, devID.ApplicationIdentifiers, up)
					if !ok {
						logger.Warn("Application Server not found, drop uplinks associated with downlink task processing")
						return
					}
					if err != nil {
						logger.WithError(err).Warn("Failed to send uplink to Application Server")
					}
				}
			}()
		}
		if len(queuedEvents) > 0 {
			go func() {
				for _, ev := range queuedEvents {
					events.Publish(ev)
				}
			}()
		}
		if err != nil {
			setErr = true
			logger.WithError(err).Error("Failed to update device in registry")
			return err
		}
		if !nextDownlinkAt.IsZero() {
			logger.WithField("start_at", nextDownlinkAt.UTC()).Debug("Add downlink task after downlink attempt")
			if err := ns.downlinkTasks.Add(ctx, devID, nextDownlinkAt, true); err != nil {
				addErr = true
				logger.WithError(err).Error("Failed to add downlink task after downlink attempt")
				return err
			}
		}
		return nil
	})
	if err != nil && !setErr && !addErr {
		log.FromContext(ctx).WithError(err).Error("Failed to pop device from downlink schedule")
	}
	return err
}
