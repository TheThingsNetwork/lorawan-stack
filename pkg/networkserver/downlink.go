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
	"math"
	"sort"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"google.golang.org/grpc"
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
			pairs = append(pairs, "absolute_time", *down.ClassBC.AbsoluteTime)
		}
		if len(down.ClassBC.GetGateways()) > 0 {
			pairs = append(pairs, "fixed_gateway_count", len(down.ClassBC.Gateways))
		}
	} else {
		pairs = append(pairs, "class_b_c", false)
	}
	return logger.WithFields(log.Fields(pairs...))
}

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

func (ns *NetworkServer) updateDataDownlinkTask(ctx context.Context, dev *ttnpb.EndDevice, earliestAt time.Time) error {
	logger := log.FromContext(ctx)
	if dev.GetMACState() == nil || dev.GetSession() == nil {
		logger.Debug("Avoid updating downlink task queue for device with no MAC state or session")
		return nil
	}

	if t := timeNow().UTC().Add(nsScheduleWindow()); earliestAt.Before(t) {
		earliestAt = t
	}
	var taskAt time.Time
	_, phy, err := getDeviceBandVersion(dev, ns.FrequencyPlans)
	if err != nil {
		logger.WithError(err).Warn("Failed to determine device band")
	} else {
		slot, ok := nextDataDownlinkSlot(ctx, dev, phy, ns.defaultMACSettings, earliestAt)
		if !ok {
			return nil
		}
		from := slot.From()
		switch {
		case slot.IsContinuous():
			// Continuous downlink slot, enqueue at the time it becomes available.
			taskAt = from

		case !from.IsZero():
			// Absolute time downlink slot, enqueue in advance to allow for scheduling.
			taskAt = from.Add(-dev.MACState.CurrentParameters.Rx1Delay.Duration() - nsScheduleWindow())
		}
	}
	if taskAt.Before(earliestAt) {
		taskAt = earliestAt
	}
	logger.WithField("start_at", taskAt).Debug("Add downlink task")
	return ns.downlinkTasks.Add(ctx, dev.EndDeviceIdentifiers, taskAt, true)
}

// generateDataDownlink attempts to generate a downlink.
// generateDataDownlink returns the generated downlink, application uplinks associated with the generation and error, if any.
// generateDataDownlink may mutate the device in order to record the downlink generated.
// maxDownLen and maxUpLen represent the maximum length of MACPayload for the downlink and corresponding uplink respectively.
// If no downlink could be generated errNoDownlink is returned.
// generateDataDownlink does not perform validation of dev.MACState.DesiredParameters,
// hence, it could potentially generate MAC command(s), which are not suported by the
// regional parameters the device operates in.
// For example, a sequence of 'NewChannel' MAC commands could be generated for a
// device operating in a region where a fixed channel plan is defined in case
// dev.MACState.CurrentParameters.Channels is not equal to dev.MACState.DesiredParameters.Channels.
// Note, that generateDataDownlink assumes transmitAt is the earliest possible time a downlink can be transmitted to the device.
func (ns *NetworkServer) generateDataDownlink(ctx context.Context, dev *ttnpb.EndDevice, phy band.Band, class ttnpb.Class, transmitAt time.Time, maxDownLen, maxUpLen uint16) (*generatedDownlink, generateDownlinkState, error) {
	if dev.MACState == nil {
		return nil, generateDownlinkState{}, errUnknownMACState.New()
	}
	if dev.Session == nil {
		return nil, generateDownlinkState{}, errEmptySession.New()
	}

	ctx = log.NewContextWithFields(ctx, log.Fields(
		"device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers),
		"mac_version", dev.MACState.LoRaWANVersion,
		"max_downlink_length", maxDownLen,
		"phy_version", dev.LoRaWANPHYVersion,
		"transmit_at", transmitAt,
	))
	logger := log.FromContext(ctx)

	// NOTE: len(FHDR) + len(FPort) = 7 + 1 = 8
	if maxDownLen < 8 || maxUpLen < 8 {
		log.FromContext(ctx).Error("Data rate MAC payload size limits too low for data downlink to be generated")
		return nil, generateDownlinkState{}, errInvalidDataRate.New()
	}
	maxDownLen, maxUpLen = maxDownLen-8, maxUpLen-8

	var fPending bool
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
			maxDownLen -= 1 + desc.DownlinkLength
		}
	}
	dev.MACState.QueuedResponses = nil
	dev.MACState.PendingRequests = dev.MACState.PendingRequests[:0]

	var genState generateDownlinkState
	cmdBuf := make([]byte, 0, maxDownLen)
	if !dev.Multicast && len(lostResps) == 0 {
		enqueuers := make([]func(context.Context, *ttnpb.EndDevice, uint16, uint16) macCommandEnqueueState, 0, 13)
		if dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_0) >= 0 {
			enqueuers = append(enqueuers,
				enqueueDutyCycleReq,
				enqueueRxParamSetupReq,
				func(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen uint16, maxUpLen uint16) macCommandEnqueueState {
					return enqueueDevStatusReq(ctx, dev, maxDownLen, maxUpLen, ns.defaultMACSettings, transmitAt)
				},
				enqueueNewChannelReq,
				func(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen uint16, maxUpLen uint16) macCommandEnqueueState {
					// NOTE: LinkADRReq must be enqueued after NewChannelReq.
					st, err := enqueueLinkADRReq(ctx, dev, maxDownLen, maxUpLen, ns.defaultMACSettings, phy)
					if err != nil {
						logger.WithError(err).Error("Failed to enqueue LinkADRReq")
						return macCommandEnqueueState{
							MaxDownLen: maxDownLen,
							MaxUpLen:   maxUpLen,
						}
					}
					return st
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
					func(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen uint16, maxUpLen uint16) macCommandEnqueueState {
						return enqueueTxParamSetupReq(ctx, dev, maxDownLen, maxUpLen, phy)
					},
				)
			}
			enqueuers = append(enqueuers,
				enqueueDLChannelReq,
			)
		}
		if dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) >= 0 {
			enqueuers = append(enqueuers,
				func(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen uint16, maxUpLen uint16) macCommandEnqueueState {
					return enqueueADRParamSetupReq(ctx, dev, maxDownLen, maxUpLen, phy)
				},
				enqueueForceRejoinReq,
				enqueueRejoinParamSetupReq,
			)
		}

		for _, f := range enqueuers {
			st := f(ctx, dev, maxDownLen, maxUpLen)
			maxDownLen = st.MaxDownLen
			maxUpLen = st.MaxUpLen
			fPending = fPending || !st.Ok
			for _, ev := range st.QueuedEvents {
				genState.Events = append(genState.Events, ev.New(ctx, events.WithIdentifiers(dev.EndDeviceIdentifiers)))
			}
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
		}
	}
	logger = logger.WithFields(log.Fields(
		"mac_count", len(cmds),
		"mac_len", len(cmdBuf),
	))
	ctx = log.NewContext(ctx, logger)

	var needsDownlink bool
	var up *ttnpb.UplinkMessage
	if dev.MACState.RxWindowsAvailable && len(dev.MACState.RecentUplinks) > 0 {
		up = lastUplink(dev.MACState.RecentUplinks...)
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
				ADR: deviceUseADR(dev, ns.defaultMACSettings, phy),
			},
		},
	}
	logger = logger.WithFields(log.Fields(
		"ack", pld.FHDR.FCtrl.Ack,
		"adr", pld.FHDR.FCtrl.ADR,
	))
	ctx = log.NewContext(ctx, logger)

	cmdsInFOpts := len(cmdBuf) <= fOptsCapacity
	if cmdsInFOpts {
		appDowns := dev.Session.QueuedApplicationDownlinks[:0:0]
	outer:
		for i, down := range dev.Session.QueuedApplicationDownlinks {
			logger := loggerWithApplicationDownlinkFields(logger, down)

			switch {
			case !bytes.Equal(down.SessionKeyID, dev.Session.SessionKeyID):
				if dev.PendingSession != nil && bytes.Equal(down.SessionKeyID, dev.PendingSession.SessionKeyID) {
					logger.Debug("Skip application downlink for pending session")
					appDowns = append(appDowns, down)
				} else {
					logger.Debug("Drop application downlink for unknown session")
					genState.baseApplicationUps = append(genState.baseApplicationUps, &ttnpb.ApplicationUp{
						EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
						CorrelationIDs:       append(events.CorrelationIDsFromContext(ctx), down.CorrelationIDs...),
						Up: &ttnpb.ApplicationUp_DownlinkFailed{
							DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
								ApplicationDownlink: *down,
								Error:               *ttnpb.ErrorDetailsToProto(errUnknownSession),
							},
						},
					})
				}

			case down.FCnt <= dev.Session.LastNFCntDown && dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0:
				logger.WithField("last_f_cnt_down", dev.Session.LastNFCntDown).Debug("Drop application downlink with too low FCnt")
				invalid, rest := ttnpb.PartitionDownlinksBySessionKeyIDEquality(dev.Session.SessionKeyID, dev.Session.QueuedApplicationDownlinks[i:]...)
				genState.baseApplicationUps = append(genState.baseApplicationUps, &ttnpb.ApplicationUp{
					EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
					CorrelationIDs:       events.CorrelationIDsFromContext(ctx),
					Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
						DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{
							Downlinks:    invalid,
							LastFCntDown: dev.Session.LastNFCntDown,
						},
					},
				})
				appDowns = append(appDowns, rest...)
				break outer

			case down.Confirmed && dev.Multicast:
				logger.Debug("Drop confirmed application downlink for multicast device")
				genState.baseApplicationUps = append(genState.baseApplicationUps, &ttnpb.ApplicationUp{
					EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
					CorrelationIDs:       events.CorrelationIDsFromContext(ctx),
					Up: &ttnpb.ApplicationUp_DownlinkFailed{
						DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
							ApplicationDownlink: *down,
							Error:               *ttnpb.ErrorDetailsToProto(errConfirmedMulticastDownlink),
						},
					},
				})
				// TODO: Check if following downlinks must be dropped (https://github.com/TheThingsNetwork/lorawan-stack/issues/1653).

			case down.ClassBC.GetAbsoluteTime() != nil && down.ClassBC.AbsoluteTime.Before(transmitAt):
				logger.Debug("Drop expired downlink")
				genState.baseApplicationUps = append(genState.baseApplicationUps, &ttnpb.ApplicationUp{
					EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
					CorrelationIDs:       append(events.CorrelationIDsFromContext(ctx), down.CorrelationIDs...),
					Up: &ttnpb.ApplicationUp_DownlinkFailed{
						DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
							ApplicationDownlink: *down,
							Error:               *ttnpb.ErrorDetailsToProto(errExpiredDownlink),
						},
					},
				})
				// TODO: Check if following downlinks must be dropped (https://github.com/TheThingsNetwork/lorawan-stack/issues/1653).

			case down.ClassBC != nil && class == ttnpb.CLASS_A:
				appDowns = append(appDowns, dev.Session.QueuedApplicationDownlinks[i:]...)
				logger.Debug("Skip class B/C downlink for class A downlink slot")
				break outer

			case len(down.FRMPayload) > int(maxDownLen):
				if len(down.FRMPayload) <= int(maxDownLen)+len(cmdBuf) {
					logger.Debug("Skip application downlink with payload length exceeding band regulations due to FOpts field being non-empty")
					appDowns = append(appDowns, dev.Session.QueuedApplicationDownlinks[i:]...)
				} else {
					logger.Debug("Drop application downlink with payload length exceeding band regulations")
					genState.baseApplicationUps = append(genState.baseApplicationUps, &ttnpb.ApplicationUp{
						EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
						CorrelationIDs:       append(events.CorrelationIDsFromContext(ctx), down.CorrelationIDs...),
						Up: &ttnpb.ApplicationUp_DownlinkFailed{
							DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
								ApplicationDownlink: *down,
								Error:               *ttnpb.ErrorDetailsToProto(errApplicationDownlinkTooLong.WithAttributes("length", len(down.FRMPayload), "max", maxDownLen)),
							},
						},
					})
					// TODO: Check if following downlinks must be dropped (https://github.com/TheThingsNetwork/lorawan-stack/issues/1653).
				}

			default:
				appDowns = append(appDowns, dev.Session.QueuedApplicationDownlinks[i+1:]...)
				genState.ApplicationDownlink = down
				break outer
			}
		}
		if genState.ApplicationDownlink != nil {
			genState.NeedsDownlinkQueueUpdate = len(appDowns) != len(dev.Session.QueuedApplicationDownlinks)-1
		} else {
			genState.NeedsDownlinkQueueUpdate = len(appDowns) != len(dev.Session.QueuedApplicationDownlinks)
		}
		dev.Session.QueuedApplicationDownlinks = appDowns
	}

	mType := ttnpb.MType_UNCONFIRMED_DOWN
	switch {
	case genState.ApplicationDownlink != nil:
		loggerWithApplicationDownlinkFields(logger, genState.ApplicationDownlink).Debug("Add application downlink to buffer")
		pld.FHDR.FCnt = genState.ApplicationDownlink.FCnt
		pld.FPort = genState.ApplicationDownlink.FPort
		pld.FRMPayload = genState.ApplicationDownlink.FRMPayload
		if genState.ApplicationDownlink.Confirmed {
			mType = ttnpb.MType_CONFIRMED_DOWN
		}

	case len(cmdBuf) > 0, needsDownlink:
		var fCnt uint32
		if dev.Session.LastNFCntDown > 0 || len(dev.MACState.RecentDownlinks) > 0 {
			fCnt = dev.Session.LastNFCntDown + 1
		}
		pld.FHDR.FCnt = fCnt

	default:
		return nil, genState, errNoDownlink.New()
	}

	logger = logger.WithFields(log.Fields(
		"f_cnt", pld.FHDR.FCnt,
		"f_port", pld.FPort,
		"m_type", mType,
	))
	ctx = log.NewContext(ctx, logger)

	if len(cmdBuf) > 0 && (!cmdsInFOpts || dev.MACState.LoRaWANVersion.EncryptFOpts()) {
		if dev.Session.NwkSEncKey == nil || len(dev.Session.NwkSEncKey.Key) == 0 {
			return nil, genState, errUnknownNwkSEncKey.New()
		}
		key, err := cryptoutil.UnwrapAES128Key(ctx, *dev.Session.NwkSEncKey, ns.KeyVault)
		if err != nil {
			logger.WithField("kek_label", dev.Session.NwkSEncKey.KEKLabel).WithError(err).Warn("Failed to unwrap NwkSEncKey")
			return nil, genState, err
		}
		fCnt := pld.FHDR.FCnt
		if pld.FPort != 0 {
			fCnt = dev.Session.LastNFCntDown
		}
		cmdBuf, err = crypto.EncryptDownlink(key, dev.Session.DevAddr, fCnt, cmdBuf, cmdsInFOpts)
		if err != nil {
			return nil, genState, errEncryptMAC.WithCause(err)
		}
	}
	if cmdsInFOpts {
		pld.FHDR.FOpts = cmdBuf
	} else {
		pld.FRMPayload = cmdBuf
	}
	if pld.FPort == 0 && dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
		genState.ifScheduledApplicationUps = append(genState.ifScheduledApplicationUps, &ttnpb.ApplicationUp{
			EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
			CorrelationIDs:       events.CorrelationIDsFromContext(ctx),
			Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
				DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{
					Downlinks:    dev.Session.QueuedApplicationDownlinks,
					LastFCntDown: pld.FHDR.FCnt,
				},
			},
		})
	}
	if class != ttnpb.CLASS_C {
		pld.FHDR.FCtrl.FPending = fPending || len(dev.Session.QueuedApplicationDownlinks) > 0
	}

	logger = logger.WithField("f_pending", pld.FHDR.FCtrl.FPending)
	ctx = log.NewContext(ctx, logger)

	needsAck := mType == ttnpb.MType_CONFIRMED_DOWN || len(dev.MACState.PendingRequests) > 0
	if needsAck && class != ttnpb.CLASS_A {
		confirmedAt, ok := nextConfirmedNetworkInitiatedDownlinkAt(ctx, dev, phy, ns.defaultMACSettings)
		if !ok {
			return nil, genState, errCorruptedMACState.New()
		}
		if confirmedAt.After(transmitAt) {
			// Caller must have checked this already.
			logger.WithField("confirmed_at", confirmedAt).Error("Confirmed class B/C downlink attempt performed too soon")
			return nil, genState, errConfirmedDownlinkTooSoon.New()
		}
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
		return nil, genState, errEncodePayload.WithCause(err)
	}
	// NOTE: It is assumed, that b does not contain MIC.

	if dev.Session.SNwkSIntKey == nil || len(dev.Session.SNwkSIntKey.Key) == 0 {
		return nil, genState, errUnknownSNwkSIntKey.New()
	}
	key, err := cryptoutil.UnwrapAES128Key(ctx, *dev.Session.SNwkSIntKey, ns.KeyVault)
	if err != nil {
		logger.WithField("kek_label", dev.Session.SNwkSIntKey.KEKLabel).WithError(err).Warn("Failed to unwrap SNwkSIntKey")
		return nil, genState, err
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
		return nil, genState, errComputeMIC.New()
	}
	b = append(b, mic[:]...)

	var priority ttnpb.TxSchedulePriority
	if genState.ApplicationDownlink != nil {
		priority = genState.ApplicationDownlink.Priority
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
	}, genState, nil
}

type downlinkPath struct {
	*ttnpb.GatewayIdentifiers
	*ttnpb.DownlinkPath
}

func downlinkPathsFromMetadata(mds ...*ttnpb.RxMetadata) []downlinkPath {
	mds = append(mds[:0:0], mds...)
	sort.SliceStable(mds, func(i, j int) bool {
		// TODO: Improve the sorting algorithm (https://github.com/TheThingsNetwork/lorawan-stack/issues/13)
		return mds[i].SNR > mds[j].SNR
	})
	head := make([]downlinkPath, 0, len(mds))
	body := make([]downlinkPath, 0, len(mds))
	tail := make([]downlinkPath, 0, len(mds))
	for _, md := range mds {
		if len(md.UplinkToken) == 0 || md.DownlinkPathConstraint == ttnpb.DOWNLINK_PATH_CONSTRAINT_NEVER {
			continue
		}
		path := downlinkPath{
			DownlinkPath: &ttnpb.DownlinkPath{
				Path: &ttnpb.DownlinkPath_UplinkToken{
					UplinkToken: md.UplinkToken,
				},
			},
		}
		if md.PacketBroker != nil {
			tail = append(tail, path)
		} else {
			path.GatewayIdentifiers = &md.GatewayIdentifiers
			switch md.DownlinkPathConstraint {
			case ttnpb.DOWNLINK_PATH_CONSTRAINT_NONE:
				head = append(head, path)
			case ttnpb.DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER:
				body = append(body, path)
			}
		}
	}
	res := append(head, body...)
	res = append(res, tail...)
	return res
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

type scheduleRequest struct {
	*ttnpb.TxRequest
	ttnpb.EndDeviceIdentifiers
	PHYPayload   []byte
	AttemptEvent events.Builder
	SuccessEvent events.Builder
	FailEvent    events.Builder
}

func newDataDownlinkScheduleRequest(req *ttnpb.TxRequest, ids ttnpb.EndDeviceIdentifiers, b []byte) *scheduleRequest {
	return &scheduleRequest{
		TxRequest:            req,
		EndDeviceIdentifiers: ids,
		PHYPayload:           b,
		AttemptEvent:         evtScheduleDataDownlinkAttempt,
		SuccessEvent:         evtScheduleDataDownlinkSuccess,
		FailEvent:            evtScheduleDataDownlinkFail,
	}
}

type downlinkTarget interface {
	Equal(downlinkTarget) bool
	Schedule(context.Context, *ttnpb.DownlinkMessage, ...grpc.CallOption) (time.Duration, error)
}

type gatewayServerDownlinkTarget struct {
	peer cluster.Peer
}

func (t *gatewayServerDownlinkTarget) Equal(target downlinkTarget) bool {
	other, ok := target.(*gatewayServerDownlinkTarget)
	if !ok {
		return false
	}
	return other.peer == t.peer
}

func (t *gatewayServerDownlinkTarget) Schedule(ctx context.Context, msg *ttnpb.DownlinkMessage, callOpts ...grpc.CallOption) (time.Duration, error) {
	conn, err := t.peer.Conn()
	if err != nil {
		return 0, err
	}
	res, err := ttnpb.NewNsGsClient(conn).ScheduleDownlink(ctx, msg, callOpts...)
	if err != nil {
		return 0, err
	}
	return res.Delay, nil
}

type packetBrokerDownlinkTarget struct {
	peer cluster.Peer
}

func (t *packetBrokerDownlinkTarget) Equal(target downlinkTarget) bool {
	_, ok := target.(*packetBrokerDownlinkTarget)
	return ok
}

func (t *packetBrokerDownlinkTarget) Schedule(ctx context.Context, msg *ttnpb.DownlinkMessage, callOpts ...grpc.CallOption) (time.Duration, error) {
	conn, err := t.peer.Conn()
	if err != nil {
		return 0, err
	}
	_, err = ttnpb.NewNsPbaClient(conn).PublishDownlink(ctx, msg, callOpts...)
	if err != nil {
		return 0, err
	}
	return peeringScheduleDelay, nil
}

// scheduleDownlinkByPaths attempts to schedule payload b using parameters in req using paths.
// scheduleDownlinkByPaths discards req.TxRequest.DownlinkPaths and mutates it arbitrarily.
// scheduleDownlinkByPaths returns the scheduled downlink or error.
func (ns *NetworkServer) scheduleDownlinkByPaths(ctx context.Context, req *scheduleRequest, paths ...downlinkPath) (*scheduledDownlink, []events.Event, error) {
	if len(paths) == 0 {
		return nil, nil, errNoPath.New()
	}

	logger := log.FromContext(ctx)

	type attempt struct {
		downlinkTarget
		paths []*ttnpb.DownlinkPath
	}

	queuedEvents := make([]events.Event, 0, len(paths))
	attempts := make([]*attempt, 0, len(paths))
	for _, path := range paths {
		var target downlinkTarget
		if path.GatewayIdentifiers != nil {
			logger := logger.WithFields(log.Fields(
				"target", "gateway_server",
				"gateway_uid", unique.ID(ctx, path.GatewayIdentifiers),
			))
			peer, err := ns.GetPeer(ctx, ttnpb.ClusterRole_GATEWAY_SERVER, *path.GatewayIdentifiers)
			if err != nil {
				logger.WithError(err).Warn("Failed to get Gateway Server peer")
				continue
			}
			target = &gatewayServerDownlinkTarget{peer: peer}
		} else {
			logger := logger.WithField("target", "packet_broker_agent")
			peer, err := ns.GetPeer(ctx, ttnpb.ClusterRole_PACKET_BROKER_AGENT, nil)
			if err != nil {
				logger.WithError(err).Warn("Failed to get Packet Broker Agent peer")
				continue
			}
			target = &packetBrokerDownlinkTarget{peer: peer}
		}

		var a *attempt
		if len(attempts) > 0 {
			if last := attempts[len(attempts)-1]; last.Equal(target) {
				a = last
			}
		}
		if a == nil {
			a = &attempt{
				downlinkTarget: target,
			}
			attempts = append(attempts, a)
		}
		a.paths = append(a.paths, path.DownlinkPath)
	}

	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("ns:downlink:%s", events.NewCorrelationID()))
	errs := make([]error, 0, len(attempts))
	for _, a := range attempts {
		req.TxRequest.DownlinkPaths = a.paths
		down := &ttnpb.DownlinkMessage{
			RawPayload:     req.PHYPayload,
			CorrelationIDs: events.CorrelationIDsFromContext(ctx),
			Settings: &ttnpb.DownlinkMessage_Request{
				Request: req.TxRequest,
			},
		}
		queuedEvents = append(queuedEvents, req.AttemptEvent.NewWithIdentifiersAndData(ctx, req.EndDeviceIdentifiers, down))
		logger.WithField("path_count", len(req.DownlinkPaths)).Debug("Schedule downlink")
		delay, err := a.Schedule(ctx, down, ns.WithClusterAuth())
		if err != nil {
			queuedEvents = append(queuedEvents, req.FailEvent.NewWithIdentifiersAndData(ctx, req.EndDeviceIdentifiers, err))
			errs = append(errs, err)
			continue
		}
		transmitAt := timeNow().Add(delay)
		logger.WithFields(log.Fields(
			"transmission_delay", delay,
			"transmit_at", transmitAt,
		)).Debug("Scheduled downlink")
		queuedEvents = append(queuedEvents, req.SuccessEvent.NewWithIdentifiersAndData(ctx, req.EndDeviceIdentifiers, &ttnpb.ScheduleDownlinkResponse{
			Delay: delay,
		}))
		return &scheduledDownlink{
			Message:    down,
			TransmitAt: transmitAt,
		}, queuedEvents, nil
	}
	return nil, queuedEvents, downlinkSchedulingError(errs)
}

func loggerWithTxRequestFields(logger log.Interface, req *ttnpb.TxRequest, rx1, rx2 bool) log.Interface {
	pairs := []interface{}{
		"attempt_rx1", rx1,
		"attempt_rx2", rx2,
		"downlink_class", req.Class,
		"downlink_priority", req.Priority,
		"frequency_plan", req.FrequencyPlanID,
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
	if req.AbsoluteTime != nil {
		pairs = append(pairs,
			"absolute_time", *req.AbsoluteTime,
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

func rx1Parameters(phy band.Band, macState *ttnpb.MACState, up *ttnpb.UplinkMessage) (uint64, ttnpb.DataRateIndex, error) {
	if up.DeviceChannelIndex > math.MaxUint8 {
		return 0, 0, errInvalidChannelIndex.New()
	}
	chIdx, err := phy.Rx1Channel(uint8(up.DeviceChannelIndex))
	if err != nil {
		return 0, 0, err
	}
	if uint(chIdx) >= uint(len(macState.CurrentParameters.Channels)) ||
		macState.CurrentParameters.Channels[int(chIdx)].GetDownlinkFrequency() == 0 {
		return 0, 0, errCorruptedMACState.New()
	}
	drIdx, err := phy.Rx1DataRate(up.Settings.DataRateIndex, macState.CurrentParameters.Rx1DataRateOffset, macState.CurrentParameters.DownlinkDwellTime.GetValue())
	if err != nil {
		return 0, 0, err
	}
	_, ok := phy.DataRates[drIdx]
	if !ok {
		return 0, 0, errDataRateIndexNotFound.WithAttributes("index", drIdx)
	}
	return macState.CurrentParameters.Channels[int(chIdx)].DownlinkFrequency, drIdx, nil
}

// maximumUplinkLength returns the maximum length of the next uplink after ups.
func maximumUplinkLength(fp *frequencyplans.FrequencyPlan, phy band.Band, ups ...*ttnpb.UplinkMessage) (uint16, error) {
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
	dr, ok := phy.DataRates[maxUpDRIdx]
	if !ok {
		return 0, errDataRateIndexNotFound.WithAttributes("index", maxUpDRIdx)
	}
	return dr.MaxMACPayloadSize(fp.DwellTime.GetUplinks()), nil
}

// downlinkRetryInterval is the time interval, which defines the interval between downlink task retries.
const downlinkRetryInterval = 2 * time.Second

func recordDataDownlink(dev *ttnpb.EndDevice, genDown *generatedDownlink, genState generateDownlinkState, down *scheduledDownlink, defaults ttnpb.MACSettings) {
	if genState.ApplicationDownlink == nil || dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 && genDown.FCnt > dev.Session.LastNFCntDown {
		dev.Session.LastNFCntDown = genDown.FCnt
	}
	dev.MACState.LastDownlinkAt = timePtr(down.TransmitAt)
	if genDown.NeedsAck {
		dev.MACState.LastConfirmedDownlinkAt = timePtr(down.TransmitAt)
	}
	if class := down.Message.GetRequest().GetClass(); class == ttnpb.CLASS_B || class == ttnpb.CLASS_C {
		dev.MACState.LastNetworkInitiatedDownlinkAt = timePtr(down.TransmitAt)
	}

	if genState.ApplicationDownlink != nil && genState.ApplicationDownlink.Confirmed {
		dev.MACState.PendingApplicationDownlink = genState.ApplicationDownlink
		dev.Session.LastConfFCntDown = genDown.FCnt
	}
	dev.MACState.QueuedResponses = nil
	dev.MACState.RxWindowsAvailable = false
	dev.MACState.RecentDownlinks = appendRecentDownlink(dev.MACState.RecentDownlinks, down.Message, recentDownlinkCount)
	dev.RecentDownlinks = appendRecentDownlink(dev.RecentDownlinks, down.Message, recentDownlinkCount)
}

type downlinkTaskUpdateStrategy uint8

const (
	nextDownlinkTask downlinkTaskUpdateStrategy = iota
	retryDownlinkTask
	noDownlinkTask
)

type downlinkAttemptResult struct {
	SetPaths                   []string
	QueuedApplicationUplinks   []*ttnpb.ApplicationUp
	QueuedEvents               []events.Event
	DownlinkTaskUpdateStrategy downlinkTaskUpdateStrategy
}

func (ns *NetworkServer) attemptClassADataDownlink(ctx context.Context, dev *ttnpb.EndDevice, phy band.Band, fp *frequencyplans.FrequencyPlan, slot *classADownlinkSlot, maxUpLength uint16) downlinkAttemptResult {
	ctx = events.ContextWithCorrelationID(ctx, slot.Uplink.CorrelationIDs...)
	if !dev.MACState.RxWindowsAvailable {
		log.FromContext(ctx).Error("RX windows not available, skip class A downlink slot")
		dev.MACState.QueuedResponses = nil
		dev.MACState.RxWindowsAvailable = false
		return downlinkAttemptResult{
			SetPaths: []string{
				"mac_state.queued_responses",
				"mac_state.rx_windows_available",
			},
		}
	}

	paths := downlinkPathsFromRecentUplinks(dev.MACState.RecentUplinks...)
	if len(paths) == 0 {
		log.FromContext(ctx).Error("No downlink path available, skip class A downlink slot")
		return downlinkAttemptResult{
			DownlinkTaskUpdateStrategy: noDownlinkTask,
		}
	}

	now := timeNow()
	if slot.RX2().Before(now) {
		log.FromContext(ctx).Debug("RX2 expired, skip class A downlink slot")
		dev.MACState.QueuedResponses = nil
		dev.MACState.RxWindowsAvailable = false
		return downlinkAttemptResult{
			SetPaths: []string{
				"mac_state.queued_responses",
				"mac_state.rx_windows_available",
			},
		}
	}

	var (
		attemptRX1 bool
		rx1Freq    uint64
		rx1DR      band.DataRate
		rx1DRIdx   ttnpb.DataRateIndex

		attemptRX2 bool
	)
	if !slot.RX1().Before(now) {
		freq, drIdx, err := rx1Parameters(phy, dev.MACState, slot.Uplink)
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Failed to compute RX1 parameters")
		} else {
			dr, ok := phy.DataRates[drIdx]
			if !ok {
				log.FromContext(ctx).WithError(errDataRateIndexNotFound.WithAttributes("index", drIdx)).Error("Failed to compute RX1 parameters")
			} else {
				attemptRX1 = true
				rx1Freq = freq
				rx1DRIdx = drIdx
				rx1DR = dr
			}
		}
	}
	rx2DR, ok := phy.DataRates[dev.MACState.CurrentParameters.Rx2DataRateIndex]
	if !ok {
		log.FromContext(ctx).WithError(errDataRateIndexNotFound.WithAttributes("index", dev.MACState.CurrentParameters.Rx2DataRateIndex)).Error("Failed to compute RX2 parameters")
	} else {
		attemptRX2 = true
	}
	if !attemptRX1 && !attemptRX2 {
		dev.MACState.QueuedResponses = nil
		dev.MACState.RxWindowsAvailable = false
		return downlinkAttemptResult{
			SetPaths: []string{
				"mac_state.queued_responses",
				"mac_state.rx_windows_available",
			},
		}
	}

	var (
		// transmitAt is the latest time.Time when downlink will be transmitted to the device.
		transmitAt time.Time
		maxDR      band.DataRate
	)
	if attemptRX1 && rx1DRIdx > dev.MACState.CurrentParameters.Rx2DataRateIndex {
		transmitAt = slot.RX1()
		maxDR = rx1DR
	} else {
		transmitAt = slot.RX2()
		maxDR = rx2DR
	}

	genDown, genState, err := ns.generateDataDownlink(ctx, dev, phy, ttnpb.CLASS_A, transmitAt,
		maxDR.MaxMACPayloadSize(fp.DwellTime.GetDownlinks()),
		maxUpLength,
	)
	var sets []string
	if genState.NeedsDownlinkQueueUpdate {
		sets = ttnpb.AddFields(sets,
			"session.queued_application_downlinks",
		)
	}
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to generate class A downlink, skip class A downlink slot")
		if genState.ApplicationDownlink != nil {
			dev.Session.QueuedApplicationDownlinks = append([]*ttnpb.ApplicationDownlink{genState.ApplicationDownlink}, dev.Session.QueuedApplicationDownlinks...)
		}
		return downlinkAttemptResult{
			DownlinkTaskUpdateStrategy: noDownlinkTask,
			SetPaths:                   sets,
			QueuedApplicationUplinks:   genState.appendApplicationUplinks(nil, false),
		}
	}

	if attemptRX1 && attemptRX2 {
		attemptRX1 = len(genDown.Payload) <= int(rx1DR.MaxMACPayloadSize(fp.DwellTime.GetDownlinks()))
		attemptRX2 = len(genDown.Payload) <= int(rx2DR.MaxMACPayloadSize(fp.DwellTime.GetDownlinks()))
		if !attemptRX1 && !attemptRX2 {
			log.FromContext(ctx).Error("Generated downlink payload size does not fit neither RX1, nor RX2, skip class A downlink slot")
			dev.MACState.QueuedResponses = nil
			dev.MACState.RxWindowsAvailable = false
			return downlinkAttemptResult{
				DownlinkTaskUpdateStrategy: nextDownlinkTask,
				SetPaths: ttnpb.AddFields(sets,
					"mac_state.queued_responses",
					"mac_state.rx_windows_available",
				),
				QueuedApplicationUplinks: genState.appendApplicationUplinks(nil, false),
			}
		}
		// NOTE: It may be possible that RX1 is dropped at this point and DevStatusReq can be scheduled in RX2 due to the downlink being
		// transmitted later, but that's micro-optimization, which we don't need to make.
	}

	if genState.ApplicationDownlink != nil {
		ctx = events.ContextWithCorrelationID(ctx, genState.ApplicationDownlink.CorrelationIDs...)
	}
	logger := log.FromContext(ctx)

	req := &ttnpb.TxRequest{
		Class:           ttnpb.CLASS_A,
		Priority:        genDown.Priority,
		FrequencyPlanID: dev.FrequencyPlanID,
		Rx1Delay:        ttnpb.RxDelay(slot.RxDelay / time.Second),
	}
	if attemptRX1 {
		req.Rx1Frequency = rx1Freq
		req.Rx1DataRateIndex = rx1DRIdx
	}
	if attemptRX2 {
		req.Rx2Frequency = dev.MACState.CurrentParameters.Rx2Frequency
		req.Rx2DataRateIndex = dev.MACState.CurrentParameters.Rx2DataRateIndex
	}
	down, downEvs, err := ns.scheduleDownlinkByPaths(
		log.NewContext(ctx, loggerWithTxRequestFields(logger, req, attemptRX1, attemptRX2).WithField("rx1_delay", req.Rx1Delay)),
		newDataDownlinkScheduleRequest(req, dev.EndDeviceIdentifiers, genDown.Payload),
		paths...,
	)
	queuedEvents := append(genState.Events, downEvs...)
	if err != nil {
		if schedErr, ok := err.(downlinkSchedulingError); ok {
			logger = loggerWithDownlinkSchedulingErrorFields(logger, schedErr)
		} else {
			logger = logger.WithError(err)
		}
		logger.Warn("All Gateway Servers failed to schedule downlink, skip class A downlink slot")
		if genState.ApplicationDownlink != nil {
			dev.Session.QueuedApplicationDownlinks = append([]*ttnpb.ApplicationDownlink{genState.ApplicationDownlink}, dev.Session.QueuedApplicationDownlinks...)
		}
		dev.MACState.QueuedResponses = nil
		dev.MACState.RxWindowsAvailable = false
		return downlinkAttemptResult{
			SetPaths: ttnpb.AddFields(sets,
				"mac_state.queued_responses",
				"mac_state.rx_windows_available",
			),
			QueuedApplicationUplinks: genState.appendApplicationUplinks(nil, false),
			QueuedEvents:             queuedEvents,
		}
	}
	if genState.ApplicationDownlink != nil {
		sets = ttnpb.AddFields(sets, "session.queued_application_downlinks")
	}
	recordDataDownlink(dev, genDown, genState, down, ns.defaultMACSettings)
	return downlinkAttemptResult{
		SetPaths: ttnpb.AddFields(sets,
			"mac_state.last_confirmed_downlink_at",
			"mac_state.last_downlink_at",
			"mac_state.pending_application_downlink",
			"mac_state.pending_requests",
			"mac_state.queued_responses",
			"mac_state.recent_downlinks",
			"mac_state.rx_windows_available",
			"recent_downlinks",
			"session",
		),
		QueuedApplicationUplinks: genState.appendApplicationUplinks(nil, true),
		QueuedEvents:             queuedEvents,
	}
}

func (ns *NetworkServer) attemptNetworkInitiatedDataDownlink(ctx context.Context, dev *ttnpb.EndDevice, phy band.Band, fp *frequencyplans.FrequencyPlan, slot *networkInitiatedDownlinkSlot, maxUpLength uint16) downlinkAttemptResult {
	var drIdx ttnpb.DataRateIndex
	var freq uint64
	switch slot.Class {
	case ttnpb.CLASS_B:
		if dev.MACState.CurrentParameters.PingSlotDataRateIndexValue == nil {
			log.FromContext(ctx).Error("Device is in class B mode, but ping slot data rate index is not known, skip class B/C downlink slot")
			return downlinkAttemptResult{
				DownlinkTaskUpdateStrategy: noDownlinkTask,
			}
		}
		drIdx = dev.MACState.CurrentParameters.PingSlotDataRateIndexValue.Value
		freq = dev.MACState.CurrentParameters.PingSlotFrequency

	case ttnpb.CLASS_C:
		drIdx = dev.MACState.CurrentParameters.Rx2DataRateIndex
		freq = dev.MACState.CurrentParameters.Rx2Frequency

	default:
		panic(fmt.Sprintf("unmatched downlink class: '%s'", slot.Class))
	}
	dr, ok := phy.DataRates[drIdx]
	if !ok {
		log.FromContext(ctx).WithField("data_rate_index", drIdx).Error("RX2 data rate not found")
		return downlinkAttemptResult{
			DownlinkTaskUpdateStrategy: noDownlinkTask,
		}
	}

	genDown, genState, err := ns.generateDataDownlink(ctx, dev, phy, slot.Class, latestTime(slot.Time, timeNow()),
		dr.MaxMACPayloadSize(fp.DwellTime.GetDownlinks()),
		maxUpLength,
	)
	var sets []string
	if genState.NeedsDownlinkQueueUpdate {
		sets = ttnpb.AddFields(sets, "session.queued_application_downlinks")
	}
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to generate class B/C downlink, skip downlink attempt")
		if genState.ApplicationDownlink != nil && ttnpb.HasAnyField(sets, "session.queued_application_downlinks") {
			dev.Session.QueuedApplicationDownlinks = append([]*ttnpb.ApplicationDownlink{genState.ApplicationDownlink}, dev.Session.QueuedApplicationDownlinks...)
		}
		return downlinkAttemptResult{
			DownlinkTaskUpdateStrategy: noDownlinkTask,
			SetPaths:                   sets,
			QueuedApplicationUplinks:   genState.appendApplicationUplinks(nil, false),
		}
	}
	if genState.ApplicationDownlink != nil {
		ctx = events.ContextWithCorrelationID(ctx, genState.ApplicationDownlink.CorrelationIDs...)
	}

	absTime := genState.ApplicationDownlink.GetClassBC().GetAbsoluteTime()
	switch {
	case absTime != nil:

	case slot.IsApplicationTime:
		log.FromContext(ctx).Error("Absolute time application downlink expected, but no absolute time downlink generated, retry downlink attempt")
		return downlinkAttemptResult{
			SetPaths:                 sets,
			QueuedApplicationUplinks: genState.appendApplicationUplinks(nil, false),
		}

	case !slot.Time.Before(timeNow()):
		absTime = &slot.Time

	case slot.Class == ttnpb.CLASS_B:
		log.FromContext(ctx).Error("Class B ping slot expired, retry downlink attempt")
		return downlinkAttemptResult{
			SetPaths:                 sets,
			QueuedApplicationUplinks: genState.appendApplicationUplinks(nil, false),
		}
	}

	var paths []downlinkPath
	if fixedPaths := genState.ApplicationDownlink.GetClassBC().GetGateways(); len(fixedPaths) > 0 {
		paths = make([]downlinkPath, 0, len(fixedPaths))
		for i := range fixedPaths {
			paths = append(paths, downlinkPath{
				GatewayIdentifiers: &fixedPaths[i].GatewayIdentifiers,
				DownlinkPath: &ttnpb.DownlinkPath{
					Path: &ttnpb.DownlinkPath_Fixed{
						Fixed: &fixedPaths[i],
					},
				},
			})
		}
	} else {
		paths = downlinkPathsFromRecentUplinks(dev.MACState.RecentUplinks...)
		if len(paths) == 0 {
			log.FromContext(ctx).Error("No downlink path available, skip class B/C downlink slot")
			if genState.ApplicationDownlink != nil && ttnpb.HasAnyField(sets, "session.queued_application_downlinks") {
				dev.Session.QueuedApplicationDownlinks = append([]*ttnpb.ApplicationDownlink{genState.ApplicationDownlink}, dev.Session.QueuedApplicationDownlinks...)
			}
			return downlinkAttemptResult{
				DownlinkTaskUpdateStrategy: noDownlinkTask,
				SetPaths:                   sets,
				QueuedApplicationUplinks:   genState.appendApplicationUplinks(nil, false),
			}
		}
	}

	req := &ttnpb.TxRequest{
		Class:            slot.Class,
		Priority:         genDown.Priority,
		FrequencyPlanID:  dev.FrequencyPlanID,
		Rx2DataRateIndex: drIdx,
		Rx2Frequency:     freq,
		AbsoluteTime:     absTime,
	}
	logger := log.FromContext(ctx)
	down, downEvs, err := ns.scheduleDownlinkByPaths(
		log.NewContext(ctx, loggerWithTxRequestFields(logger, req, false, true)),
		newDataDownlinkScheduleRequest(req, dev.EndDeviceIdentifiers, genDown.Payload),
		paths...,
	)
	queuedEvents := append(genState.Events, downEvs...)
	if err != nil {
		schedErr, ok := err.(downlinkSchedulingError)
		if ok {
			logger = loggerWithDownlinkSchedulingErrorFields(logger, schedErr)
		} else {
			logger = logger.WithError(err)
		}
		if ok && genState.ApplicationDownlink != nil {
			pathErrs, ok := schedErr.pathErrors()
			if ok {
				if genState.ApplicationDownlink.GetClassBC().GetAbsoluteTime() != nil &&
					allErrors(nonRetryableAbsoluteTimeGatewayError, pathErrs...) {
					logger.Warn("Absolute time invalid, fail application downlink")
					return downlinkAttemptResult{
						SetPaths: ttnpb.AddFields(sets, "session.queued_application_downlinks"),
						QueuedApplicationUplinks: append(genState.appendApplicationUplinks(nil, false), &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
							CorrelationIDs:       events.CorrelationIDsFromContext(ctx),
							Up: &ttnpb.ApplicationUp_DownlinkFailed{
								DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
									ApplicationDownlink: *genState.ApplicationDownlink,
									Error:               *ttnpb.ErrorDetailsToProto(errInvalidAbsoluteTime),
								},
							},
						}),
					}
				}
				if len(genState.ApplicationDownlink.GetClassBC().GetGateways()) > 0 &&
					allErrors(nonRetryableFixedPathGatewayError, pathErrs...) {
					logger.Warn("Fixed paths invalid, fail application downlink")
					return downlinkAttemptResult{
						SetPaths: ttnpb.AddFields(sets, "session.queued_application_downlinks"),
						QueuedApplicationUplinks: append(genState.appendApplicationUplinks(nil, false), &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
							CorrelationIDs:       events.CorrelationIDsFromContext(ctx),
							Up: &ttnpb.ApplicationUp_DownlinkFailed{
								DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
									ApplicationDownlink: *genState.ApplicationDownlink,
									Error:               *ttnpb.ErrorDetailsToProto(errInvalidFixedPaths),
								},
							},
						}),
					}
				}
			}
		}
		logger.Warn("All Gateway Servers failed to schedule downlink, retry attempt")
		if genState.NeedsDownlinkQueueUpdate {
			dev.Session.QueuedApplicationDownlinks = append([]*ttnpb.ApplicationDownlink{genState.ApplicationDownlink}, dev.Session.QueuedApplicationDownlinks...)
		}
		return downlinkAttemptResult{
			SetPaths:                   sets,
			QueuedApplicationUplinks:   genState.appendApplicationUplinks(nil, false),
			QueuedEvents:               queuedEvents,
			DownlinkTaskUpdateStrategy: retryDownlinkTask,
		}
	}

	recordDataDownlink(dev, genDown, genState, down, ns.defaultMACSettings)
	if genState.ApplicationDownlink != nil {
		sets = ttnpb.AddFields(sets, "session.queued_application_downlinks")
	}
	return downlinkAttemptResult{
		SetPaths: ttnpb.AddFields(sets,
			"mac_state.last_confirmed_downlink_at",
			"mac_state.last_downlink_at",
			"mac_state.last_network_initiated_downlink_at",
			"mac_state.pending_application_downlink",
			"mac_state.pending_requests",
			"mac_state.queued_responses",
			"mac_state.recent_downlinks",
			"mac_state.rx_windows_available",
			"recent_downlinks",
			"session",
		),
		QueuedApplicationUplinks: genState.appendApplicationUplinks(nil, true),
		QueuedEvents:             queuedEvents,
	}
}

// processDownlinkTask processes the most recent downlink task ready for execution, if such is available or wait until it is before processing it.
// NOTE: ctx.Done() is not guaranteed to be respected by processDownlinkTask.
func (ns *NetworkServer) processDownlinkTask(ctx context.Context) error {
	var setErr bool
	var addErr bool
	err := ns.downlinkTasks.Pop(ctx, func(ctx context.Context, devID ttnpb.EndDeviceIdentifiers, t time.Time) error {
		ctx = log.NewContextWithFields(ctx, log.Fields(
			"device_uid", unique.ID(ctx, devID),
			"started_at", timeNow().UTC(),
		))
		logger := log.FromContext(ctx)
		logger.WithField("start_at", t).Debug("Process downlink task")

		var queuedEvents []events.Event
		defer func() { publishEvents(ctx, queuedEvents...) }()

		var queuedApplicationUplinks []*ttnpb.ApplicationUp
		defer func() { ns.enqueueApplicationUplinks(ctx, queuedApplicationUplinks...) }()

		taskUpdateStrategy := noDownlinkTask
		dev, ctx, err := ns.devices.SetByID(ctx, devID.ApplicationIdentifiers, devID.DeviceID,
			[]string{
				"frequency_plan_id",
				"last_dev_status_received_at",
				"lorawan_phy_version",
				"mac_settings",
				"mac_state",
				"multicast",
				"pending_mac_state",
				"recent_downlinks",
				"recent_uplinks",
				"session",
			},
			func(ctx context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
				if dev == nil {
					logger.Warn("Device not found")
					return nil, nil, nil
				}

				fp, phy, err := getDeviceBandVersion(dev, ns.FrequencyPlans)
				if err != nil {
					taskUpdateStrategy = retryDownlinkTask
					logger.WithError(err).Error("Failed to get frequency plan of the device, retry downlink slot")
					return dev, nil, nil
				}
				logger = logger.WithFields(log.Fields(
					"band_id", phy.ID,
					"frequency_plan_id", dev.FrequencyPlanID,
				))
				ctx = log.NewContext(ctx, logger)

				if dev.PendingMACState != nil &&
					dev.PendingMACState.PendingJoinRequest == nil &&
					dev.PendingMACState.RxWindowsAvailable &&
					dev.PendingMACState.QueuedJoinAccept != nil {

					logger = logger.WithField("downlink_type", "join-accept")
					ctx = log.NewContext(ctx, logger)

					if len(dev.RecentUplinks) == 0 {
						logger.Error("No recent uplinks found, skip downlink slot")
						return dev, nil, nil
					}
					up := lastUplink(dev.RecentUplinks...)
					switch up.Payload.MHDR.MType {
					case ttnpb.MType_JOIN_REQUEST, ttnpb.MType_REJOIN_REQUEST:
					default:
						logger.Error("Last uplink is neither join-request, nor rejoin-request, skip downlink slot")
						return dev, nil, nil
					}
					ctx := events.ContextWithCorrelationID(ctx, up.CorrelationIDs...)
					ctx = events.ContextWithCorrelationID(ctx, dev.PendingMACState.QueuedJoinAccept.CorrelationIDs...)

					paths := downlinkPathsFromRecentUplinks(up)
					if len(paths) == 0 {
						logger.Warn("No downlink path available, skip join-accept downlink slot")
						dev.PendingMACState.RxWindowsAvailable = false
						taskUpdateStrategy = nextDownlinkTask
						return dev, []string{
							"pending_mac_state.rx_windows_available",
						}, nil
					}

					rx1 := up.ReceivedAt.Add(phy.JoinAcceptDelay1)
					now := timeNow()
					if rx1.Add(time.Second).Before(now) {
						logger.Warn("RX1 and RX2 are expired, skip join-accept downlink slot")
						dev.PendingMACState.RxWindowsAvailable = false
						taskUpdateStrategy = nextDownlinkTask
						return dev, []string{
							"pending_mac_state.rx_windows_available",
						}, nil
					}

					var (
						attemptRX1 bool
						rx1Freq    uint64
						rx1DRIdx   ttnpb.DataRateIndex

						attemptRX2 bool
					)
					if !rx1.Before(now) {
						freq, drIdx, err := rx1Parameters(phy, dev.PendingMACState, up)
						if err != nil {
							log.FromContext(ctx).WithError(err).Error("Failed to compute RX1 parameters")
						} else {
							attemptRX1 = true
							rx1Freq = freq
							rx1DRIdx = drIdx
						}
					}
					_, ok := phy.DataRates[dev.PendingMACState.CurrentParameters.Rx2DataRateIndex]
					if !ok {
						log.FromContext(ctx).WithError(errDataRateIndexNotFound.WithAttributes("index", dev.PendingMACState.CurrentParameters.Rx2DataRateIndex)).Error("Failed to compute RX2 parameters")
					} else {
						attemptRX2 = true
					}
					if !attemptRX1 && !attemptRX2 {
						dev.PendingMACState.RxWindowsAvailable = false
						taskUpdateStrategy = nextDownlinkTask
						return dev, []string{
							"pending_mac_state.rx_windows_available",
						}, nil
					}

					req := &ttnpb.TxRequest{
						Class:           ttnpb.CLASS_A,
						Priority:        ns.downlinkPriorities.JoinAccept,
						FrequencyPlanID: dev.FrequencyPlanID,
						Rx1Delay:        ttnpb.RxDelay(phy.JoinAcceptDelay1 / time.Second),
					}
					if attemptRX1 {
						req.Rx1Frequency = rx1Freq
						req.Rx1DataRateIndex = rx1DRIdx
					}
					if attemptRX2 {
						req.Rx2Frequency = dev.PendingMACState.CurrentParameters.Rx2Frequency
						req.Rx2DataRateIndex = dev.PendingMACState.CurrentParameters.Rx2DataRateIndex
					}
					down, downEvs, err := ns.scheduleDownlinkByPaths(
						log.NewContext(ctx, loggerWithTxRequestFields(logger, req, attemptRX1, attemptRX2).WithField("rx1_delay", req.Rx1Delay)),
						&scheduleRequest{
							TxRequest:            req,
							EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
							PHYPayload:           dev.PendingMACState.QueuedJoinAccept.Payload,
							AttemptEvent:         evtScheduleJoinAcceptAttempt,
							SuccessEvent:         evtScheduleJoinAcceptSuccess,
							FailEvent:            evtScheduleJoinAcceptFail,
						},
						paths...,
					)
					queuedEvents = append(queuedEvents, downEvs...)
					if err != nil {
						if schedErr, ok := err.(downlinkSchedulingError); ok {
							logger = loggerWithDownlinkSchedulingErrorFields(logger, schedErr)
						} else {
							logger = logger.WithError(err)
						}
						logger.Warn("All Gateway Servers failed to schedule downlink, skip join-accept downlink slot")
						dev.PendingMACState.RxWindowsAvailable = false
						taskUpdateStrategy = nextDownlinkTask
						return dev, []string{
							"pending_mac_state.rx_windows_available",
						}, nil
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
				if dev.Session == nil {
					logger.Warn("Unknown session, skip downlink slot")
					return dev, nil, nil
				}
				logger = logger.WithField("dev_addr", dev.Session.DevAddr)

				if dev.MACState == nil {
					logger.Warn("Unknown MAC state, skip downlink slot")
					return dev, nil, nil
				}
				logger = logger.WithField("device_class", dev.MACState.DeviceClass)

				ctx = log.NewContext(ctx, logger)

				var maxUpLength uint16 = math.MaxUint16
				if !dev.Multicast && dev.MACState.LoRaWANVersion == ttnpb.MAC_V1_1 {
					maxUpLength, err = maximumUplinkLength(fp, phy, dev.MACState.RecentUplinks...)
					if err != nil {
						logger.WithError(err).Error("Failed to determine maximum uplink length")
						return dev, nil, nil
					}
				}
				var earliestAt time.Time
				for {
					v, ok := nextDataDownlinkSlot(ctx, dev, phy, ns.defaultMACSettings, earliestAt)
					if !ok {
						return dev, nil, nil
					}
					switch slot := v.(type) {
					case *classADownlinkSlot:
						a := ns.attemptClassADataDownlink(ctx, dev, phy, fp, slot, maxUpLength)
						queuedEvents = append(queuedEvents, a.QueuedEvents...)
						queuedApplicationUplinks = append(queuedApplicationUplinks, a.QueuedApplicationUplinks...)
						taskUpdateStrategy = a.DownlinkTaskUpdateStrategy
						return dev, a.SetPaths, nil

					case *networkInitiatedDownlinkSlot:
						switch {
						case slot.Class == ttnpb.CLASS_B && slot.Time.IsZero(),
							slot.IsApplicationTime && slot.Time.IsZero():
							logger.Error("Invalid downlink slot generated, skip class B/C downlink slot")
							return dev, nil, nil

						case !slot.IsApplicationTime && slot.Class == ttnpb.CLASS_C && timeUntil(slot.Time) > 0:
							logger.WithFields(log.Fields(
								"slot_start", slot.Time,
							)).Info("Class C downlink scheduling attempt performed too soon, retry attempt")
							taskUpdateStrategy = nextDownlinkTask
							return dev, nil, nil

						case timeUntil(slot.Time) > dev.MACState.CurrentParameters.Rx1Delay.Duration()+2*nsScheduleWindow():
							logger.WithFields(log.Fields(
								"slot_start", slot.Time,
							)).Info("Class B/C downlink scheduling attempt performed too soon, retry attempt")
							taskUpdateStrategy = nextDownlinkTask
							return dev, nil, nil

						case !slot.IsApplicationTime && slot.Class == ttnpb.CLASS_B && timeUntil(slot.Time) < dev.MACState.CurrentParameters.Rx1Delay.Duration()/2:
							earliestAt = timeNow().Add(dev.MACState.CurrentParameters.Rx1Delay.Duration() / 2)
							continue
						}
						a := ns.attemptNetworkInitiatedDataDownlink(ctx, dev, phy, fp, slot, maxUpLength)
						queuedEvents = append(queuedEvents, a.QueuedEvents...)
						queuedApplicationUplinks = append(queuedApplicationUplinks, a.QueuedApplicationUplinks...)
						taskUpdateStrategy = a.DownlinkTaskUpdateStrategy
						return dev, a.SetPaths, nil

					default:
						panic(fmt.Errorf("unknown downlink slot type: %T", slot))
					}
				}
			},
		)
		if err != nil {
			setErr = true
			logger.WithError(err).Error("Failed to update device in registry")
			return err
		}

		var earliestAt time.Time
		switch taskUpdateStrategy {
		case nextDownlinkTask:

		case retryDownlinkTask:
			earliestAt = timeNow().Add(downlinkRetryInterval + nsScheduleWindow())

		case noDownlinkTask:
			return nil

		default:
			panic(fmt.Errorf("unmatched downlink task update strategy: %v", taskUpdateStrategy))
		}
		logger.WithField("earliest_at", earliestAt).Debug("Update downlink task queue after downlink attempt")
		if err := ns.updateDataDownlinkTask(ctx, dev, earliestAt); err != nil {
			addErr = true
			logger.WithError(err).Error("Failed to update downlink task queue after downlink attempt")
			return err
		}
		return nil
	})
	if err != nil && !setErr && !addErr {
		log.FromContext(ctx).WithError(err).Error("Failed to pop device from downlink schedule")
	}
	return err
}
