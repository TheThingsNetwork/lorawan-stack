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

	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/experimental"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"google.golang.org/grpc"
)

// DownlinkTaskQueue represents an entity, that holds downlink tasks sorted by timestamp.
type DownlinkTaskQueue interface {
	// Add adds downlink task for device identified by devID at time t.
	// Implementations must ensure that Add returns fast.
	Add(ctx context.Context, devID *ttnpb.EndDeviceIdentifiers, t time.Time, replace bool) error

	// Pop calls f on the most recent downlink task in the schedule, for which timestamp is in range [0, time.Now()],
	// if such is available, otherwise it blocks until it is.
	// Context passed to f must be derived from ctx.
	// Implementations must respect ctx.Done() value on best-effort basis.
	// consumerID should be a unique ID for this consumer.
	Pop(ctx context.Context, consumerID string, f func(context.Context, *ttnpb.EndDeviceIdentifiers, time.Time) (time.Time, error)) error
}

func loggerWithApplicationDownlinkFields(logger log.Interface, down *ttnpb.ApplicationDownlink) log.Interface {
	pairs := []interface{}{
		"confirmed", down.Confirmed,
		"f_cnt", down.FCnt,
		"f_port", down.FPort,
		"frm_payload_len", len(down.FrmPayload),
		"priority", down.Priority,
		"session_key_id", down.SessionKeyId,
	}
	if down.GetClassBC() != nil {
		pairs = append(pairs, "class_b_c", true)
		if down.ClassBC.GetAbsoluteTime() != nil {
			pairs = append(pairs, "absolute_time", ttnpb.StdTime(down.ClassBC.AbsoluteTime))
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
	Payload        *ttnpb.Message
	RawPayload     []byte
	Priority       ttnpb.TxSchedulePriority
	NeedsMACAnswer bool
	SessionKeyID   []byte
}

type generateDownlinkState struct {
	baseApplicationUps        []*ttnpb.ApplicationUp
	ifScheduledApplicationUps []*ttnpb.ApplicationUp

	ApplicationDownlink           *ttnpb.ApplicationDownlink
	EventBuilders                 events.Builders
	NeedsDownlinkQueueUpdate      bool
	EvictDownlinkQueueIfScheduled bool
}

func (s generateDownlinkState) appendApplicationUplinks(ups []*ttnpb.ApplicationUp, scheduled bool) []*ttnpb.ApplicationUp {
	if !scheduled {
		return append(ups, s.baseApplicationUps...)
	} else {
		return append(append(ups, s.baseApplicationUps...), s.ifScheduledApplicationUps...)
	}
}

func (ns *NetworkServer) nextDataDownlinkTaskAt(ctx context.Context, dev *ttnpb.EndDevice, earliestAt time.Time) (time.Time, error) {
	if dev.GetMacState() == nil || dev.GetSession() == nil {
		log.FromContext(ctx).Debug("Cannot compute next downlink task time for device with no MAC state or session")
		return time.Time{}, nil
	}

	if t := time.Now().UTC().Add(nsScheduleWindow()); earliestAt.Before(t) {
		earliestAt = t
	}

	fps, err := ns.FrequencyPlansStore(ctx)
	if err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to get frequency plans store")
		return time.Time{}, nil
	}
	var taskAt time.Time
	phy, err := DeviceBand(dev, fps)
	if err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to determine device band")
		return time.Time{}, nil
	}
	slot, ok := nextDataDownlinkSlot(ctx, dev, phy, ns.defaultMACSettings, earliestAt)
	if !ok {
		return time.Time{}, nil
	}
	from := slot.From()
	switch {
	case slot.IsContinuous():
		// Continuous downlink slot, enqueue at the time it becomes available.
		taskAt = from

	case !from.IsZero():
		// Absolute time downlink slot, enqueue in advance to allow for scheduling.
		taskAt = from.Add(-absoluteTimeSchedulingDelay - nsScheduleWindow())
	}
	if taskAt.Before(earliestAt) {
		taskAt = earliestAt
	}
	return taskAt, nil
}

func (ns *NetworkServer) updateDataDownlinkTask(ctx context.Context, dev *ttnpb.EndDevice, earliestAt time.Time) error {
	taskAt, err := ns.nextDataDownlinkTaskAt(ctx, dev, earliestAt)
	if err != nil || taskAt.IsZero() {
		return err
	}
	log.FromContext(ctx).WithField("start_at", taskAt).Debug("Add downlink task")
	return ns.downlinkTasks.Add(ctx, dev.Ids, taskAt, true)
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
func (ns *NetworkServer) generateDataDownlink(ctx context.Context, dev *ttnpb.EndDevice, phy *band.Band, class ttnpb.Class, transmitAt time.Time, maxDownLen, maxUpLen uint16) (*generatedDownlink, generateDownlinkState, error) {
	if dev.MacState == nil {
		return nil, generateDownlinkState{}, errUnknownMACState.New()
	}
	if dev.Session == nil {
		return nil, generateDownlinkState{}, errEmptySession.New()
	}

	ctx = log.NewContextWithFields(ctx, log.Fields(
		"device_uid", unique.ID(ctx, dev.Ids),
		"mac_version", dev.MacState.LorawanVersion,
		"max_downlink_length", maxDownLen,
		"phy_version", dev.LorawanPhyVersion,
		"transmit_at", transmitAt,
	))
	logger := log.FromContext(ctx)

	// NOTE: len(FHDR) + len(FPort) = 7 + 1 = 8
	if maxDownLen < 8 || maxUpLen < 8 {
		log.FromContext(ctx).Error("Data rate MAC payload size limits too low for data downlink to be generated")
		return nil, generateDownlinkState{}, errInvalidDataRate.New()
	}
	maxDownLen, maxUpLen = maxDownLen-8, maxUpLen-8

	var (
		fPending bool
		genState generateDownlinkState
		cmdBuf   []byte
	)
	if class == ttnpb.Class_CLASS_A {
		spec := lorawan.DefaultMACCommands
		cmds := make([]*ttnpb.MACCommand, 0, len(dev.MacState.QueuedResponses)+len(dev.MacState.PendingRequests))

		for _, cmd := range dev.MacState.QueuedResponses {
			logger := logger.WithField("cid", cmd.Cid)
			desc, ok := spec[cmd.Cid]
			switch {
			case !ok:
				logger.Error("Unknown MAC command response enqueued, set FPending")
				maxDownLen = 0
				fPending = true
			case desc.DownlinkLength >= maxDownLen:
				logger.WithFields(log.Fields(
					"command_length", 1+desc.DownlinkLength,
					"remaining_downlink_length", maxDownLen,
				)).Warn("MAC command response does not fit in buffer, set FPending")
				maxDownLen = 0
				fPending = true
			default:
				cmds = append(cmds, cmd)
				maxDownLen -= 1 + desc.DownlinkLength
			}
			if !ok || desc.DownlinkLength > maxDownLen {
				break
			}
		}
		dev.MacState.QueuedResponses = nil
		dev.MacState.PendingRequests = dev.MacState.PendingRequests[:0]

		enqueuers := make([]func(context.Context, *ttnpb.EndDevice, uint16, uint16) mac.EnqueueState, 0, 13)
		if dev.MacState.LorawanVersion.Compare(ttnpb.MAC_V1_0) >= 0 {
			enqueuers = append(enqueuers,
				mac.EnqueueDutyCycleReq,
				mac.EnqueueRxParamSetupReq,
				func(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen uint16, maxUpLen uint16) mac.EnqueueState {
					return mac.EnqueueDevStatusReq(ctx, dev, maxDownLen, maxUpLen, ns.defaultMACSettings, transmitAt)
				},
				mac.EnqueueNewChannelReq,
				func(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen uint16, maxUpLen uint16) mac.EnqueueState {
					// NOTE: LinkADRReq must be enqueued after NewChannelReq.
					st, err := mac.EnqueueLinkADRReq(ctx, dev, maxDownLen, maxUpLen, phy)
					if err != nil {
						logger.WithError(err).Error("Failed to enqueue LinkADRReq")
						return mac.EnqueueState{
							MaxDownLen: maxDownLen,
							MaxUpLen:   maxUpLen,
						}
					}
					return st
				},
				mac.EnqueueRxTimingSetupReq,
			)
			if dev.MacState.DeviceClass == ttnpb.Class_CLASS_B {
				if class == ttnpb.Class_CLASS_A {
					enqueuers = append(enqueuers,
						mac.EnqueuePingSlotChannelReq,
					)
				}
				enqueuers = append(enqueuers,
					mac.EnqueueBeaconFreqReq,
				)
			}
		}
		if dev.MacState.LorawanVersion.Compare(ttnpb.MAC_V1_0_2) >= 0 {
			if phy.TxParamSetupReqSupport {
				enqueuers = append(enqueuers,
					func(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen uint16, maxUpLen uint16) mac.EnqueueState {
						return mac.EnqueueTxParamSetupReq(ctx, dev, maxDownLen, maxUpLen, phy)
					},
				)
			}
			enqueuers = append(enqueuers,
				mac.EnqueueDLChannelReq,
			)
		}
		if dev.MacState.LorawanVersion.Compare(ttnpb.MAC_V1_1) >= 0 {
			enqueuers = append(enqueuers,
				func(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen uint16, maxUpLen uint16) mac.EnqueueState {
					return mac.EnqueueADRParamSetupReq(ctx, dev, maxDownLen, maxUpLen, phy)
				},
				mac.EnqueueForceRejoinReq,
				mac.EnqueueRejoinParamSetupReq,
			)
		}

		for _, f := range enqueuers {
			st := f(ctx, dev, maxDownLen, maxUpLen)
			maxDownLen = st.MaxDownLen
			maxUpLen = st.MaxUpLen
			fPending = fPending || !st.Ok
			genState.EventBuilders = append(genState.EventBuilders, st.QueuedEvents...)
		}

		b := make([]byte, 0, maxDownLen)
		cmds = append(cmds, dev.MacState.PendingRequests...)
		for _, cmd := range cmds {
			logger := logger.WithField("cid", cmd.Cid)
			logger.Debug("Add MAC command to buffer")
			var err error
			b, err = spec.AppendDownlink(*phy, b, *cmd)
			if err != nil {
				return nil, generateDownlinkState{}, errEncodeMAC.WithCause(err)
			}
		}
		logger = logger.WithFields(log.Fields(
			"mac_count", len(cmds),
			"mac_length", len(b),
		))
		if len(b) > 0 {
			cmdBuf = b
		}
		ctx = log.NewContext(ctx, logger)
	}

	var needsDownlink bool
	var up *ttnpb.UplinkMessage
	if dev.MacState.RxWindowsAvailable && len(dev.MacState.RecentUplinks) > 0 {
		up = LastUplink(dev.MacState.RecentUplinks...)
		switch up.Payload.MHdr.MType {
		case ttnpb.MType_UNCONFIRMED_UP:
			if up.Payload.GetMacPayload().FHdr.FCtrl.AdrAckReq {
				logger.Debug("Need downlink for ADRAckReq")
				needsDownlink = true
			}

		case ttnpb.MType_CONFIRMED_UP:
			logger.Debug("Need downlink for confirmed uplink")
			needsDownlink = true

		case ttnpb.MType_PROPRIETARY:

		default:
			panic(fmt.Sprintf("invalid uplink MType: %s", up.Payload.MHdr.MType))
		}
	}

	pld := &ttnpb.MACPayload{
		FHdr: &ttnpb.FHDR{
			DevAddr: dev.Session.DevAddr,
			FCtrl: &ttnpb.FCtrl{
				Ack: up != nil && up.Payload.MHdr.MType == ttnpb.MType_CONFIRMED_UP,
				Adr: mac.DeviceUseADR(dev, ns.defaultMACSettings, phy),
			},
		},
	}
	logger = logger.WithFields(log.Fields(
		"ack", pld.FHdr.FCtrl.Ack,
		"adr", pld.FHdr.FCtrl.Adr,
	))
	ctx = log.NewContext(ctx, logger)

	cmdsInFOpts := len(cmdBuf) <= fOptsCapacity
	if cmdsInFOpts {
		appDowns := dev.Session.QueuedApplicationDownlinks[:0:0]
	outer:
		for i, down := range dev.Session.QueuedApplicationDownlinks {
			logger := loggerWithApplicationDownlinkFields(logger, down)

			switch {
			case !bytes.Equal(down.SessionKeyId, dev.Session.Keys.SessionKeyId):
				if dev.PendingSession != nil && bytes.Equal(down.SessionKeyId, dev.PendingSession.Keys.SessionKeyId) {
					logger.Debug("Skip application downlink for pending session")
					appDowns = append(appDowns, down)
				} else {
					logger.Debug("Drop application downlink for unknown session")
					genState.baseApplicationUps = append(genState.baseApplicationUps, &ttnpb.ApplicationUp{
						EndDeviceIds:   dev.Ids,
						CorrelationIds: append(events.CorrelationIDsFromContext(ctx), down.CorrelationIds...),
						Up: &ttnpb.ApplicationUp_DownlinkFailed{
							DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
								Downlink: down,
								Error:    ttnpb.ErrorDetailsToProto(errUnknownSession),
							},
						},
					})
				}

			case down.FCnt <= dev.Session.LastNFCntDown && dev.MacState.LorawanVersion.Compare(ttnpb.MAC_V1_1) < 0:
				logger.WithField("last_f_cnt_down", dev.Session.LastNFCntDown).Debug("Drop application downlink with too low FCnt")
				genState.baseApplicationUps = append(genState.baseApplicationUps, &ttnpb.ApplicationUp{
					EndDeviceIds:   dev.Ids,
					CorrelationIds: events.CorrelationIDsFromContext(ctx),
					Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
						DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{
							Downlinks:    dev.Session.QueuedApplicationDownlinks[i:],
							LastFCntDown: dev.Session.LastNFCntDown,
							SessionKeyId: dev.Session.Keys.SessionKeyId,
						},
					},
				})
				break outer

			case down.Confirmed && dev.Multicast:
				logger.Debug("Drop confirmed application downlink for multicast device")
				genState.baseApplicationUps = append(genState.baseApplicationUps, &ttnpb.ApplicationUp{
					EndDeviceIds:   dev.Ids,
					CorrelationIds: events.CorrelationIDsFromContext(ctx),
					Up: &ttnpb.ApplicationUp_DownlinkFailed{
						DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
							Downlink: down,
							Error:    ttnpb.ErrorDetailsToProto(errConfirmedMulticastDownlink),
						},
					},
				})
				// TODO: Check if following downlinks must be dropped (https://github.com/TheThingsNetwork/lorawan-stack/issues/1653).

			case down.ClassBC.GetAbsoluteTime() != nil && ttnpb.StdTime(down.ClassBC.AbsoluteTime).Before(transmitAt):
				logger.Debug("Drop expired downlink")
				genState.baseApplicationUps = append(genState.baseApplicationUps, &ttnpb.ApplicationUp{
					EndDeviceIds:   dev.Ids,
					CorrelationIds: append(events.CorrelationIDsFromContext(ctx), down.CorrelationIds...),
					Up: &ttnpb.ApplicationUp_DownlinkFailed{
						DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
							Downlink: down,
							Error:    ttnpb.ErrorDetailsToProto(errExpiredDownlink),
						},
					},
				})
				// TODO: Check if following downlinks must be dropped (https://github.com/TheThingsNetwork/lorawan-stack/issues/1653).

			case down.ClassBC != nil && class == ttnpb.Class_CLASS_A:
				appDowns = append(appDowns, dev.Session.QueuedApplicationDownlinks[i:]...)
				logger.Debug("Skip class B/C downlink for class A downlink slot")
				break outer

			case len(down.FrmPayload) > int(maxDownLen):
				if len(down.FrmPayload) <= int(maxDownLen)+len(cmdBuf) {
					logger.Debug("Skip application downlink with payload length exceeding band regulations due to FOpts field being non-empty")
					appDowns = append(appDowns, dev.Session.QueuedApplicationDownlinks[i:]...)
					break outer
				} else {
					logger.Debug("Drop application downlink with payload length exceeding band regulations")
					genState.baseApplicationUps = append(genState.baseApplicationUps, &ttnpb.ApplicationUp{
						EndDeviceIds:   dev.Ids,
						CorrelationIds: append(events.CorrelationIDsFromContext(ctx), down.CorrelationIds...),
						Up: &ttnpb.ApplicationUp_DownlinkFailed{
							DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
								Downlink: down,
								Error:    ttnpb.ErrorDetailsToProto(errApplicationDownlinkTooLong.WithAttributes("length", len(down.FrmPayload), "max", maxDownLen)),
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
		pld.FullFCnt = genState.ApplicationDownlink.FCnt
		pld.FPort = genState.ApplicationDownlink.FPort
		pld.FrmPayload = genState.ApplicationDownlink.FrmPayload
		if genState.ApplicationDownlink.Confirmed {
			mType = ttnpb.MType_CONFIRMED_DOWN
		}

	case len(cmdBuf) > 0, needsDownlink:
		pld.FullFCnt = func() uint32 {
			for i := len(dev.MacState.RecentDownlinks) - 1; i >= 0; i-- {
				down := dev.MacState.RecentDownlinks[i]
				switch {
				case down == nil:
					logger.Error("Empty downlink stored in device's MAC state")
					continue

				case down.Payload == nil:
					logger.Error("Downlink with no payload stored in device's MAC state")
					continue
				}

				switch down.Payload.MHdr.MType {
				case ttnpb.MType_UNCONFIRMED_DOWN, ttnpb.MType_CONFIRMED_DOWN:
					return dev.Session.LastNFCntDown + 1
				case ttnpb.MType_JOIN_ACCEPT:
					// TODO: Support rejoins (https://github.com/TheThingsNetwork/lorawan-stack/issues/8).
					return 0
				case ttnpb.MType_PROPRIETARY:
				default:
					panic(fmt.Sprintf("invalid downlink MType: %s", down.Payload.MHdr.MType))
				}
			}
			return 0
		}()

	default:
		return nil, genState, errNoDownlink.New()
	}
	pld.FHdr.FCnt = pld.FullFCnt & 0xffff

	logger = logger.WithFields(log.Fields(
		"f_cnt", pld.FHdr.FCnt,
		"full_f_cnt", pld.FullFCnt,
		"f_port", pld.FPort,
		"m_type", mType,
	))
	ctx = log.NewContext(ctx, logger)

	if len(cmdBuf) > 0 && (!cmdsInFOpts || dev.MacState.LorawanVersion.EncryptFOpts()) {
		if len(dev.GetSession().GetKeys().GetNwkSEncKey().GetKey()) == 0 {
			return nil, genState, errUnknownNwkSEncKey.New()
		}
		key, err := cryptoutil.UnwrapAES128Key(ctx, dev.Session.Keys.NwkSEncKey, ns.KeyVault)
		if err != nil {
			logger.WithField("kek_label", dev.Session.Keys.NwkSEncKey.KekLabel).WithError(err).Warn("Failed to unwrap NwkSEncKey")
			return nil, genState, err
		}
		fCnt := pld.FullFCnt
		if pld.FPort != 0 {
			fCnt = dev.Session.LastNFCntDown
		}
		cmdBuf, err = crypto.EncryptDownlink(key, dev.Session.DevAddr, fCnt, cmdBuf, cmdsInFOpts)
		if err != nil {
			return nil, genState, errEncryptMAC.WithCause(err)
		}
	}
	if cmdsInFOpts {
		pld.FHdr.FOpts = cmdBuf
	} else {
		pld.FrmPayload = cmdBuf
	}
	if pld.FPort == 0 && dev.MacState.LorawanVersion.Compare(ttnpb.MAC_V1_1) < 0 {
		genState.ifScheduledApplicationUps = append(genState.ifScheduledApplicationUps, &ttnpb.ApplicationUp{
			EndDeviceIds:   dev.Ids,
			CorrelationIds: events.CorrelationIDsFromContext(ctx),
			Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
				DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{
					Downlinks:    dev.Session.QueuedApplicationDownlinks,
					LastFCntDown: pld.FullFCnt,
					SessionKeyId: dev.Session.Keys.SessionKeyId,
				},
			},
		})
		genState.EvictDownlinkQueueIfScheduled = true
	}
	if class != ttnpb.Class_CLASS_C {
		pld.FHdr.FCtrl.FPending = fPending || len(dev.Session.QueuedApplicationDownlinks) > 0
	}

	logger = logger.WithField("f_pending", pld.FHdr.FCtrl.FPending)
	ctx = log.NewContext(ctx, logger)

	if mType == ttnpb.MType_CONFIRMED_DOWN && class != ttnpb.Class_CLASS_A {
		confirmedAt, ok := nextConfirmedNetworkInitiatedDownlinkAt(ctx, dev, phy, ns.defaultMACSettings)
		if !ok {
			return nil, genState, ErrCorruptedMACState.
				WithCause(ErrNetworkDownlinkSlot)
		}
		if confirmedAt.After(transmitAt) {
			// Caller must have checked this already.
			logger.WithField("confirmed_at", confirmedAt).Error("Confirmed class B/C downlink attempt performed too soon")
			return nil, genState, errConfirmedDownlinkTooSoon.New()
		}
	}

	msg := &ttnpb.Message{
		MHdr: &ttnpb.MHDR{
			MType: mType,
			Major: ttnpb.Major_LORAWAN_R1,
		},
		Payload: &ttnpb.Message_MacPayload{
			MacPayload: pld,
		},
	}
	b, err := lorawan.MarshalMessage(*msg)
	if err != nil {
		return nil, genState, errEncodePayload.WithCause(err)
	}
	// NOTE: It is assumed, that b does not contain MIC.

	if len(dev.Session.GetKeys().GetSNwkSIntKey().GetKey()) == 0 {
		return nil, genState, errUnknownSNwkSIntKey.New()
	}
	key, err := cryptoutil.UnwrapAES128Key(ctx, dev.Session.Keys.SNwkSIntKey, ns.KeyVault)
	if err != nil {
		logger.WithField("kek_label", dev.Session.Keys.SNwkSIntKey.KekLabel).WithError(err).Warn("Failed to unwrap SNwkSIntKey")
		return nil, genState, err
	}

	var mic [4]byte
	if dev.MacState.LorawanVersion.Compare(ttnpb.MAC_V1_1) < 0 {
		mic, err = crypto.ComputeLegacyDownlinkMIC(
			key,
			dev.Session.DevAddr,
			pld.FullFCnt,
			b,
		)
	} else {
		var confFCnt uint32
		if pld.FHdr.FCtrl.Ack {
			confFCnt = up.GetPayload().GetMacPayload().GetFullFCnt()
		}
		mic, err = crypto.ComputeDownlinkMIC(
			key,
			dev.Session.DevAddr,
			confFCnt,
			pld.FullFCnt,
			b,
		)
	}
	if err != nil {
		return nil, genState, errComputeMIC.New()
	}
	b = append(b, mic[:]...)
	msg.Mic = mic[:]

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
		Payload:        msg,
		RawPayload:     b,
		Priority:       priority,
		NeedsMACAnswer: len(dev.MacState.PendingRequests) > 0 && class == ttnpb.Class_CLASS_A,
		SessionKeyID:   dev.Session.Keys.SessionKeyId,
	}, genState, nil
}

type downlinkPath struct {
	*ttnpb.GatewayIdentifiers
	*ttnpb.DownlinkPath
}

func computeWantedRSSI(snr float32, channelRSSI float32) float32 {
	wantedRSSI := channelRSSI
	if snr <= -5.0 {
		wantedRSSI += snr
	} else if snr < 10.0 {
		wantedRSSI += snr/3.0 - 10.0/3.0
	}
	return wantedRSSI
}

func snrMetadataComparator(mds []*ttnpb.RxMetadata) func(int, int) bool {
	return func(i, j int) bool {
		return mds[i].Snr >= mds[j].Snr
	}
}

func wantedRSSIMetadataComparator(mds []*ttnpb.RxMetadata) func(int, int) bool {
	wantedRSSI := func(k int) float32 { return computeWantedRSSI(mds[k].Snr, mds[k].ChannelRssi) }
	return func(i, j int) bool {
		return wantedRSSI(i) >= wantedRSSI(j)
	}
}

var (
	ignoreInvalidMDFeatureFlag = experimental.DefineFeature("ns.down.path.ignore_invalid_metadata", false)
	useWantedRSSIFeatureFlag   = experimental.DefineFeature("ns.down.path.use_wanted_rssi", false)
)

func buildMetadataComparator(ctx context.Context, mds []*ttnpb.RxMetadata) func(int, int) bool {
	comparator := snrMetadataComparator(mds)
	if useWantedRSSIFeatureFlag.GetValue(ctx) {
		comparator = wantedRSSIMetadataComparator(mds)
	}
	if ignoreInvalidMDFeatureFlag.GetValue(ctx) {
		invalidMD := func(k int) bool { return mds[k].Snr == 0.0 || mds[k].ChannelRssi == 0.0 }
		return func(i, j int) bool {
			lhsInvalid, rhsInvalid := invalidMD(i), invalidMD(j)
			if lhsInvalid {
				return lhsInvalid == rhsInvalid
			}
			return comparator(i, j)
		}
	}
	return comparator
}

func downlinkPathsFromMetadata(ctx context.Context, mds ...*ttnpb.RxMetadata) []downlinkPath {
	mds = append(mds[:0:0], mds...)
	sort.SliceStable(mds, buildMetadataComparator(ctx, mds))
	head := make([]downlinkPath, 0, len(mds))
	body := make([]downlinkPath, 0, len(mds))
	tail := make([]downlinkPath, 0, len(mds))
	for _, md := range mds {
		if len(md.UplinkToken) == 0 || md.DownlinkPathConstraint == ttnpb.DownlinkPathConstraint_DOWNLINK_PATH_CONSTRAINT_NEVER {
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
			path.GatewayIdentifiers = md.GatewayIds
			switch md.DownlinkPathConstraint {
			case ttnpb.DownlinkPathConstraint_DOWNLINK_PATH_CONSTRAINT_NONE:
				head = append(head, path)
			case ttnpb.DownlinkPathConstraint_DOWNLINK_PATH_CONSTRAINT_PREFER_OTHER:
				body = append(body, path)
			}
		}
	}
	res := append(head, body...)
	res = append(res, tail...)
	return res
}

func downlinkPathsFromRecentUplinks(ctx context.Context, ups ...*ttnpb.UplinkMessage) []downlinkPath {
	for i := len(ups) - 1; i >= 0; i-- {
		if paths := downlinkPathsFromMetadata(ctx, ups[i].RxMetadata...); len(paths) > 0 {
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
	return errors.IsAborted(err) || // e.g. no absolute gateway time or no time sync with the server.
		errors.IsResourceExhausted(err) || // e.g. time-on-air does not fit with duty-cycle already used.
		errors.IsFailedPrecondition(err) || // e.g. no downlink allowed, invalid frequency, too late for transmission.
		errors.IsAlreadyExists(err) // e.g. a downlink has already been scheduled on the given time.
}

func nonRetryableFixedPathGatewayError(err error) bool {
	return errors.IsNotFound(err) || // e.g. gateway is not connected.
		errors.IsDataLoss(err) || // e.g. invalid uplink token.
		errors.IsFailedPrecondition(err) // e.g. no downlink allowed, invalid frequency, too late for transmission.
}

type scheduleRequest struct {
	*ttnpb.TxRequest
	*ttnpb.EndDeviceIdentifiers
	Payload      *ttnpb.Message
	RawPayload   []byte
	SessionKeyID []byte

	// DownlinkEvents are the event builders associated with particular downlink. Only published on success.
	DownlinkEvents events.Builders
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
	return ttnpb.StdDurationOrZero(res.Delay), nil
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
			peer, err := ns.GetPeer(ctx, ttnpb.ClusterRole_GATEWAY_SERVER, path.GatewayIdentifiers)
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

	var (
		attemptEvent    events.Builder
		successEvent    events.Builder
		failEvent       events.Builder
		registerAttempt func(context.Context)
		registerSuccess func(context.Context)
	)
	switch req.Payload.MHdr.MType {
	case ttnpb.MType_UNCONFIRMED_DOWN:
		attemptEvent = evtScheduleDataDownlinkAttempt
		successEvent = evtScheduleDataDownlinkSuccess
		failEvent = evtScheduleDataDownlinkFail
		registerAttempt = registerAttemptUnconfirmedDataDownlink
		registerSuccess = registerForwardUnconfirmedDataDownlink

	case ttnpb.MType_CONFIRMED_DOWN:
		attemptEvent = evtScheduleDataDownlinkAttempt
		successEvent = evtScheduleDataDownlinkSuccess
		failEvent = evtScheduleDataDownlinkFail
		registerAttempt = registerAttemptConfirmedDataDownlink
		registerSuccess = registerForwardConfirmedDataDownlink

	case ttnpb.MType_JOIN_ACCEPT:
		attemptEvent = evtScheduleJoinAcceptAttempt
		successEvent = evtScheduleJoinAcceptSuccess
		failEvent = evtScheduleJoinAcceptFail
		registerAttempt = registerAttemptJoinAcceptDownlink
		registerSuccess = registerForwardJoinAcceptDownlink
	default:
		panic(fmt.Sprintf("attempt to schedule downlink with invalid MType '%s'", req.Payload.MHdr.MType))
	}
	ctx = events.ContextWithCorrelationID(ctx, fmt.Sprintf("ns:downlink:%s", events.NewCorrelationID()))
	errs := make([]error, 0, len(attempts))
	eventIDOpt := events.WithIdentifiers(req.EndDeviceIdentifiers)
	for _, a := range attempts {
		req.TxRequest.DownlinkPaths = a.paths
		down := &ttnpb.DownlinkMessage{
			RawPayload: req.RawPayload,
			Payload:    req.Payload,
			Settings: &ttnpb.DownlinkMessage_Request{
				Request: req.TxRequest,
			},
			CorrelationIds: events.CorrelationIDsFromContext(ctx),
		}
		queuedEvents = append(queuedEvents, attemptEvent.New(ctx, eventIDOpt, events.WithData(down)))
		registerAttempt(ctx)
		logger.WithField("path_count", len(req.DownlinkPaths)).Debug("Schedule downlink")
		delay, err := a.Schedule(ctx, &ttnpb.DownlinkMessage{
			RawPayload:     down.RawPayload,
			Settings:       down.Settings,
			CorrelationIds: down.CorrelationIds,
		}, ns.WithClusterAuth())
		if err != nil {
			queuedEvents = append(queuedEvents, failEvent.New(ctx, eventIDOpt, events.WithData(err)))
			errs = append(errs, err)
			continue
		}
		transmitAt := time.Now().Add(delay)
		if err := ns.scheduledDownlinkMatcher.Add(ctx, &ttnpb.DownlinkMessage{
			Payload:        down.Payload,
			EndDeviceIds:   req.EndDeviceIdentifiers,
			Settings:       down.Settings,
			CorrelationIds: down.CorrelationIds,
			SessionKeyId:   req.SessionKeyID,
		}); err != nil {
			logger.WithError(err).Debug("Failed to store downlink metadata")
		}
		logger.WithFields(log.Fields(
			"transmission_delay", delay,
			"transmit_at", transmitAt,
		)).Debug("Scheduled downlink")
		queuedEvents = append(queuedEvents, events.Builders(append([]events.Builder{
			successEvent.With(
				events.WithData(&ttnpb.ScheduleDownlinkResponse{
					Delay: ttnpb.ProtoDurationPtr(delay),
				}),
			),
		}, []events.Builder(req.DownlinkEvents)...)).New(ctx, eventIDOpt)...)
		registerSuccess(ctx)
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
		"frequency_plan", req.FrequencyPlanId,
	}
	if rx1 {
		pairs = append(pairs,
			"rx1_data_rate", req.Rx1DataRate,
			"rx1_frequency", req.Rx1Frequency,
		)
	}
	if rx2 {
		pairs = append(pairs,
			"rx2_data_rate", req.Rx2DataRate,
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
	if n := len(recent); n > 0 {
		recent[n-1].CorrelationIds = nil
	}
	recent = append(recent, down)
	if extra := len(recent) - window; extra > 0 {
		recent = recent[extra:]
	}
	return recent
}

func rx1Parameters(phy *band.Band, macState *ttnpb.MACState, up *ttnpb.UplinkMessage) (uint64, ttnpb.DataRateIndex, band.DataRate, error) {
	if up.DeviceChannelIndex > math.MaxUint8 {
		return 0, 0, band.DataRate{}, errInvalidChannelIndex.New()
	}
	chIdx, err := phy.Rx1Channel(uint8(up.DeviceChannelIndex))
	if err != nil {
		return 0, 0, band.DataRate{}, err
	}
	if uint(chIdx) >= uint(len(macState.CurrentParameters.Channels)) {
		return 0, 0, band.DataRate{}, ErrCorruptedMACState.
			WithAttributes(
				"channel_id", chIdx,
				"channels_len", len(macState.CurrentParameters.Channels),
			).
			WithCause(ErrUnknownChannel)
	}
	if macState.CurrentParameters.Channels[int(chIdx)].GetDownlinkFrequency() == 0 {
		return 0, 0, band.DataRate{}, ErrCorruptedMACState.
			WithAttributes(
				"channel_id", chIdx,
			).
			WithCause(ErrUplinkChannel)
	}
	drIdx, _, ok := phy.FindUplinkDataRate(up.Settings.DataRate)
	if !ok {
		return 0, 0, band.DataRate{}, errDataRateNotFound.New()
	}
	drIdx, err = phy.Rx1DataRate(drIdx, macState.CurrentParameters.Rx1DataRateOffset, macState.CurrentParameters.DownlinkDwellTime.GetValue())
	if err != nil {
		return 0, 0, band.DataRate{}, err
	}
	dr, ok := phy.DataRates[drIdx]
	if !ok {
		return 0, 0, band.DataRate{}, errDataRateIndexNotFound.WithAttributes("index", drIdx)
	}
	return macState.CurrentParameters.Channels[int(chIdx)].DownlinkFrequency, drIdx, dr, nil
}

// maximumUplinkLength returns the maximum length of the next uplink after ups.
func maximumUplinkLength(fp *frequencyplans.FrequencyPlan, phy *band.Band, ups ...*ttnpb.UplinkMessage) (uint16, error) {
	// NOTE: If no data uplink is found, we assume ADR is off on the device and, hence, data rate index 0 is used in computation.
	maxUpDRIdx := ttnpb.DataRateIndex_DATA_RATE_0
loop:
	for i := len(ups) - 1; i >= 0; i-- {
		switch ups[i].Payload.MHdr.MType {
		case ttnpb.MType_JOIN_REQUEST:
			break loop
		case ttnpb.MType_UNCONFIRMED_UP, ttnpb.MType_CONFIRMED_UP:
			if ups[i].Payload.GetMacPayload().FHdr.FCtrl.Adr {
				drIdx, _, ok := phy.FindUplinkDataRate(ups[i].Settings.DataRate)
				if !ok {
					continue
				}
				maxUpDRIdx = drIdx
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

func recordDataDownlink(dev *ttnpb.EndDevice, genState generateDownlinkState, needsMACAnswer bool, down *scheduledDownlink, defaults ttnpb.MACSettings) {
	macPayload := down.Message.Payload.GetMacPayload()
	if macPayload == nil {
		panic("invalid downlink")
	}
	if genState.ApplicationDownlink == nil || dev.MacState.LorawanVersion.Compare(ttnpb.MAC_V1_1) < 0 && macPayload.FullFCnt > dev.Session.LastNFCntDown {
		dev.Session.LastNFCntDown = macPayload.FullFCnt
	}
	dev.MacState.LastDownlinkAt = ttnpb.ProtoTimePtr(down.TransmitAt)
	if needsMACAnswer || down.Message.Payload.MHdr.MType == ttnpb.MType_CONFIRMED_DOWN {
		dev.MacState.LastConfirmedDownlinkAt = ttnpb.ProtoTimePtr(down.TransmitAt)
	}
	if class := down.Message.GetRequest().GetClass(); class == ttnpb.Class_CLASS_B || class == ttnpb.Class_CLASS_C {
		dev.MacState.LastNetworkInitiatedDownlinkAt = ttnpb.ProtoTimePtr(down.TransmitAt)
	}

	if genState.ApplicationDownlink != nil && genState.ApplicationDownlink.Confirmed {
		dev.MacState.PendingApplicationDownlink = genState.ApplicationDownlink
		dev.Session.LastConfFCntDown = macPayload.FullFCnt
	}
	dev.MacState.RecentDownlinks = appendRecentDownlink(dev.MacState.RecentDownlinks, &ttnpb.DownlinkMessage{
		Payload:        down.Message.Payload,
		Settings:       down.Message.Settings,
		CorrelationIds: down.Message.CorrelationIds,
	}, recentDownlinkCount)
	dev.MacState.RxWindowsAvailable = false
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

func (ns *NetworkServer) attemptClassADataDownlink(ctx context.Context, dev *ttnpb.EndDevice, phy *band.Band, fp *frequencyplans.FrequencyPlan, slot *classADownlinkSlot, maxUpLength uint16) downlinkAttemptResult {
	ctx = events.ContextWithCorrelationID(ctx, slot.Uplink.CorrelationIds...)
	if !dev.MacState.RxWindowsAvailable {
		log.FromContext(ctx).Error("RX windows not available, skip class A downlink slot")
		dev.MacState.QueuedResponses = nil
		dev.MacState.RxWindowsAvailable = false
		return downlinkAttemptResult{
			SetPaths: []string{
				"mac_state.queued_responses",
				"mac_state.rx_windows_available",
			},
		}
	}

	paths := downlinkPathsFromRecentUplinks(ctx, dev.MacState.RecentUplinks...)
	if len(paths) == 0 {
		log.FromContext(ctx).Error("No downlink path available, skip class A downlink slot")
		return downlinkAttemptResult{
			DownlinkTaskUpdateStrategy: noDownlinkTask,
		}
	}

	now := time.Now()
	if slot.RX2().Before(now) {
		log.FromContext(ctx).Debug("RX2 expired, skip class A downlink slot")
		dev.MacState.QueuedResponses = nil
		dev.MacState.RxWindowsAvailable = false
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
		freq, drIdx, dr, err := rx1Parameters(phy, dev.MacState, slot.Uplink)
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Failed to compute RX1 parameters")
		} else {
			attemptRX1 = true
			rx1Freq = freq
			rx1DRIdx = drIdx
			rx1DR = dr
		}
	}
	rx2DR, ok := phy.DataRates[dev.MacState.CurrentParameters.Rx2DataRateIndex]
	if !ok {
		log.FromContext(ctx).WithError(errDataRateIndexNotFound.WithAttributes("index", dev.MacState.CurrentParameters.Rx2DataRateIndex)).Error("Failed to compute RX2 parameters")
	} else {
		attemptRX2 = true
	}
	if !attemptRX1 && !attemptRX2 {
		dev.MacState.QueuedResponses = nil
		dev.MacState.RxWindowsAvailable = false
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
	if attemptRX1 && rx1DRIdx > dev.MacState.CurrentParameters.Rx2DataRateIndex || !attemptRX2 {
		transmitAt = slot.RX1()
		maxDR = rx1DR
	} else {
		transmitAt = slot.RX2()
		maxDR = rx2DR
	}
	downDwellTime := fp.DwellTime.GetDownlinks()

	genDown, genState, err := ns.generateDataDownlink(ctx, dev, phy, ttnpb.Class_CLASS_A, transmitAt,
		maxDR.MaxMACPayloadSize(downDwellTime),
		maxUpLength,
	)
	var sets []string
	if genState.NeedsDownlinkQueueUpdate {
		sets = ttnpb.AddFields(sets,
			"session.queued_application_downlinks",
		)
	}
	if err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to generate class A downlink, skip class A downlink slot")
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
		// NOTE: genDown.RawPayload contains FRMPayload, which consists of:
		// * MHDR - 1 byte
		// * MACPayload - up to 250 bytes, actual value reported by band.DataRate.MaxMACPayloadSize
		// * MIC - 4 bytes
		attemptRX1 = len(genDown.RawPayload) <= int(rx1DR.MaxMACPayloadSize(downDwellTime))+5
		attemptRX2 = len(genDown.RawPayload) <= int(rx2DR.MaxMACPayloadSize(downDwellTime))+5
		if !attemptRX1 && !attemptRX2 {
			log.FromContext(ctx).Error("Generated downlink payload size does not fit neither RX1, nor RX2, skip class A downlink slot")
			dev.MacState.QueuedResponses = nil
			dev.MacState.RxWindowsAvailable = false
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
		ctx = events.ContextWithCorrelationID(ctx, genState.ApplicationDownlink.CorrelationIds...)
	}
	logger := log.FromContext(ctx)

	req := &ttnpb.TxRequest{
		Class:           ttnpb.Class_CLASS_A,
		Priority:        genDown.Priority,
		FrequencyPlanId: dev.FrequencyPlanId,
		Rx1Delay:        ttnpb.RxDelay(slot.RxDelay / time.Second),
	}
	if attemptRX1 {
		req.Rx1Frequency = rx1Freq
		req.Rx1DataRate = rx1DR.Rate
	}
	if attemptRX2 {
		req.Rx2Frequency = dev.MacState.CurrentParameters.Rx2Frequency
		req.Rx2DataRate = rx2DR.Rate
	}
	down, queuedEvents, err := ns.scheduleDownlinkByPaths(
		log.NewContext(ctx, loggerWithTxRequestFields(logger, req, attemptRX1, attemptRX2).WithField("rx1_delay", req.Rx1Delay)),
		&scheduleRequest{
			TxRequest:            req,
			EndDeviceIdentifiers: dev.Ids,
			Payload:              genDown.Payload,
			RawPayload:           genDown.RawPayload,
			SessionKeyID:         genDown.SessionKeyID,
			DownlinkEvents:       genState.EventBuilders,
		},
		paths...,
	)
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
		dev.MacState.QueuedResponses = nil
		dev.MacState.RxWindowsAvailable = false
		return downlinkAttemptResult{
			SetPaths: ttnpb.AddFields(sets,
				"mac_state.queued_responses",
				"mac_state.rx_windows_available",
			),
			QueuedApplicationUplinks: genState.appendApplicationUplinks(nil, false),
			QueuedEvents:             queuedEvents,
		}
	}
	if genState.ApplicationDownlink != nil || genState.EvictDownlinkQueueIfScheduled {
		sets = ttnpb.AddFields(sets, "session.queued_application_downlinks")
	}
	if genState.EvictDownlinkQueueIfScheduled {
		dev.Session.QueuedApplicationDownlinks = dev.Session.QueuedApplicationDownlinks[:0:0]
	}
	recordDataDownlink(dev, genState, genDown.NeedsMACAnswer, down, ns.defaultMACSettings)
	return downlinkAttemptResult{
		SetPaths: ttnpb.AddFields(sets,
			"mac_state.last_confirmed_downlink_at",
			"mac_state.last_downlink_at",
			"mac_state.pending_application_downlink",
			"mac_state.pending_requests",
			"mac_state.queued_responses",
			"mac_state.recent_downlinks",
			"mac_state.rx_windows_available",
			"session",
		),
		QueuedApplicationUplinks: genState.appendApplicationUplinks(nil, true),
		QueuedEvents:             queuedEvents,
	}
}

func (ns *NetworkServer) attemptNetworkInitiatedDataDownlink(ctx context.Context, dev *ttnpb.EndDevice, phy *band.Band, fp *frequencyplans.FrequencyPlan, slot *networkInitiatedDownlinkSlot, maxUpLength uint16) downlinkAttemptResult {
	var drIdx ttnpb.DataRateIndex
	var freq uint64
	switch slot.Class {
	case ttnpb.Class_CLASS_B:
		if dev.MacState.CurrentParameters.PingSlotDataRateIndexValue == nil {
			log.FromContext(ctx).Error("Device is in class B mode, but ping slot data rate index is not known, skip class B/C downlink slot")
			return downlinkAttemptResult{
				DownlinkTaskUpdateStrategy: noDownlinkTask,
			}
		}
		drIdx = dev.MacState.CurrentParameters.PingSlotDataRateIndexValue.Value
		freq = dev.MacState.CurrentParameters.PingSlotFrequency

	case ttnpb.Class_CLASS_C:
		drIdx = dev.MacState.CurrentParameters.Rx2DataRateIndex
		freq = dev.MacState.CurrentParameters.Rx2Frequency

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

	genDown, genState, err := ns.generateDataDownlink(ctx, dev, phy, slot.Class, latestTime(slot.Time, time.Now()),
		dr.MaxMACPayloadSize(fp.DwellTime.GetDownlinks()),
		maxUpLength,
	)
	var sets []string
	if genState.NeedsDownlinkQueueUpdate {
		sets = ttnpb.AddFields(sets, "session.queued_application_downlinks")
	}
	if err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to generate class B/C downlink, skip downlink attempt")
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
		ctx = events.ContextWithCorrelationID(ctx, genState.ApplicationDownlink.CorrelationIds...)
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

	case slot.Time.After(time.Now()):
		log.FromContext(ctx).Debug("Slot starts in the future, set absolute time in downlink request")
		absTime = ttnpb.ProtoTimePtr(slot.Time)

	case slot.Class == ttnpb.Class_CLASS_B:
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
				GatewayIdentifiers: fixedPaths[i].GatewayIds,
				DownlinkPath: &ttnpb.DownlinkPath{
					Path: &ttnpb.DownlinkPath_Fixed{
						Fixed: fixedPaths[i],
					},
				},
			})
		}
	} else {
		paths = downlinkPathsFromRecentUplinks(ctx, dev.MacState.RecentUplinks...)
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
		Class:           slot.Class,
		Priority:        genDown.Priority,
		FrequencyPlanId: dev.FrequencyPlanId,
		Rx2DataRate:     dr.Rate,
		Rx2Frequency:    freq,
		AbsoluteTime:    absTime,
	}
	down, queuedEvents, err := ns.scheduleDownlinkByPaths(
		log.NewContext(ctx, loggerWithTxRequestFields(log.FromContext(ctx), req, false, true)),
		&scheduleRequest{
			TxRequest:            req,
			EndDeviceIdentifiers: dev.Ids,
			Payload:              genDown.Payload,
			RawPayload:           genDown.RawPayload,
			DownlinkEvents:       genState.EventBuilders,
			SessionKeyID:         dev.GetSession().GetKeys().GetSessionKeyId(),
		},
		paths...,
	)
	if err != nil {
		logger := log.FromContext(ctx)
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
							EndDeviceIds:   dev.Ids,
							CorrelationIds: events.CorrelationIDsFromContext(ctx),
							Up: &ttnpb.ApplicationUp_DownlinkFailed{
								DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
									Downlink: genState.ApplicationDownlink,
									Error:    ttnpb.ErrorDetailsToProto(errInvalidAbsoluteTime),
								},
							},
						}),
						QueuedEvents: queuedEvents,
					}
				}
				if len(genState.ApplicationDownlink.GetClassBC().GetGateways()) > 0 &&
					allErrors(nonRetryableFixedPathGatewayError, pathErrs...) {
					logger.Warn("Fixed paths invalid, fail application downlink")
					return downlinkAttemptResult{
						SetPaths: ttnpb.AddFields(sets, "session.queued_application_downlinks"),
						QueuedApplicationUplinks: append(genState.appendApplicationUplinks(nil, false), &ttnpb.ApplicationUp{
							EndDeviceIds:   dev.Ids,
							CorrelationIds: events.CorrelationIDsFromContext(ctx),
							Up: &ttnpb.ApplicationUp_DownlinkFailed{
								DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
									Downlink: genState.ApplicationDownlink,
									Error:    ttnpb.ErrorDetailsToProto(errInvalidFixedPaths),
								},
							},
						}),
						QueuedEvents: queuedEvents,
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

	recordDataDownlink(dev, genState, genDown.NeedsMACAnswer, down, ns.defaultMACSettings)
	if genState.ApplicationDownlink != nil || genState.EvictDownlinkQueueIfScheduled {
		sets = ttnpb.AddFields(sets, "session.queued_application_downlinks")
	}
	if genState.EvictDownlinkQueueIfScheduled {
		dev.Session.QueuedApplicationDownlinks = dev.Session.QueuedApplicationDownlinks[:0:0]
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
			"session",
		),
		QueuedApplicationUplinks: genState.appendApplicationUplinks(nil, true),
		QueuedEvents:             queuedEvents,
	}
}

func (ns *NetworkServer) createProcessDownlinkTask(consumerID string) func(context.Context) error {
	return func(ctx context.Context) error {
		return ns.processDownlinkTask(ctx, consumerID)
	}
}

// processDownlinkTask processes the most recent downlink task ready for execution, if such is available or wait until it is before processing it.
// NOTE: ctx.Done() is not guaranteed to be respected by processDownlinkTask.
// processDownlinkTask receives the consumerID that will be used for popping from the downlink task queue.
func (ns *NetworkServer) processDownlinkTask(ctx context.Context, consumerID string) error {
	var setErr bool
	var computeNextErr bool
	err := ns.downlinkTasks.Pop(ctx, consumerID, func(ctx context.Context, devID *ttnpb.EndDeviceIdentifiers, t time.Time) (time.Time, error) {
		ctx = log.NewContextWithFields(ctx, log.Fields(
			"device_uid", unique.ID(ctx, devID),
			"started_at", time.Now().UTC(),
		))
		logger := log.FromContext(ctx)
		logger.WithField("start_at", t).Debug("Process downlink task")

		var queuedEvents []events.Event
		defer func() { publishEvents(ctx, queuedEvents...) }()

		var queuedApplicationUplinks []*ttnpb.ApplicationUp
		defer func() { ns.submitApplicationUplinks(ctx, queuedApplicationUplinks...) }()

		taskUpdateStrategy := noDownlinkTask
		dev, ctx, err := ns.devices.SetByID(ctx, devID.ApplicationIds, devID.DeviceId,
			[]string{
				"frequency_plan_id",
				"last_dev_status_received_at",
				"lorawan_phy_version",
				"mac_settings",
				"mac_state",
				"multicast",
				"pending_mac_state",
				"session",
			},
			func(ctx context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
				if dev == nil {
					logger.Warn("Device not found")
					return nil, nil, nil
				}

				fps, err := ns.FrequencyPlansStore(ctx)
				if err != nil {
					logger.WithError(err).Error("Failed to get frequency plan store")
					return nil, nil, err
				}

				fp, phy, err := DeviceFrequencyPlanAndBand(dev, fps)
				if err != nil {
					taskUpdateStrategy = retryDownlinkTask
					logger.WithError(err).Error("Failed to get frequency plan of the device, retry downlink slot")
					return dev, nil, nil
				}
				logger = logger.WithFields(log.Fields(
					"band_id", phy.ID,
					"frequency_plan_id", dev.FrequencyPlanId,
				))
				ctx = log.NewContext(ctx, logger)

				if dev.PendingMacState != nil &&
					dev.PendingMacState.PendingJoinRequest == nil &&
					dev.PendingMacState.RxWindowsAvailable &&
					dev.PendingMacState.QueuedJoinAccept != nil {

					logger = logger.WithField("downlink_type", "join-accept")
					ctx = log.NewContext(ctx, logger)

					if len(dev.PendingMacState.RecentUplinks) == 0 {
						logger.Error("No recent uplinks found, skip downlink slot")
						return dev, nil, nil
					}
					up := LastUplink(dev.PendingMacState.RecentUplinks...)
					switch up.Payload.MHdr.MType {
					case ttnpb.MType_JOIN_REQUEST, ttnpb.MType_REJOIN_REQUEST:
					default:
						logger.Error("Last uplink is neither join-request, nor rejoin-request, skip downlink slot")
						return dev, nil, nil
					}
					ctx := events.ContextWithCorrelationID(ctx, up.CorrelationIds...)
					ctx = events.ContextWithCorrelationID(ctx, dev.PendingMacState.QueuedJoinAccept.CorrelationIds...)

					paths := downlinkPathsFromRecentUplinks(ctx, up)
					if len(paths) == 0 {
						logger.Warn("No downlink path available, skip join-accept downlink slot")
						dev.PendingMacState.RxWindowsAvailable = false
						taskUpdateStrategy = nextDownlinkTask
						return dev, []string{
							"pending_mac_state.rx_windows_available",
						}, nil
					}

					rx1 := ttnpb.StdTime(up.ReceivedAt).Add(phy.JoinAcceptDelay1)
					now := time.Now()
					if rx1.Add(time.Second).Before(now) {
						logger.Warn("RX1 and RX2 are expired, skip join-accept downlink slot")
						dev.PendingMacState.RxWindowsAvailable = false
						taskUpdateStrategy = nextDownlinkTask
						return dev, []string{
							"pending_mac_state.rx_windows_available",
						}, nil
					}

					var (
						attemptRX1 bool
						rx1Freq    uint64
						rx1DR      band.DataRate

						attemptRX2 bool
					)
					if !rx1.Before(now) {
						freq, _, dr, err := rx1Parameters(phy, dev.PendingMacState, up)
						if err != nil {
							log.FromContext(ctx).WithError(err).Error("Failed to compute RX1 parameters")
						} else {
							attemptRX1 = true
							rx1Freq = freq
							rx1DR = dr
						}
					}
					rx2DR, ok := phy.DataRates[dev.PendingMacState.CurrentParameters.Rx2DataRateIndex]
					if !ok {
						log.FromContext(ctx).WithError(errDataRateIndexNotFound.WithAttributes("index", dev.PendingMacState.CurrentParameters.Rx2DataRateIndex)).Error("Failed to compute RX2 parameters")
					} else {
						attemptRX2 = true
					}
					if !attemptRX1 && !attemptRX2 {
						dev.PendingMacState.RxWindowsAvailable = false
						taskUpdateStrategy = nextDownlinkTask
						return dev, []string{
							"pending_mac_state.rx_windows_available",
						}, nil
					}

					req := &ttnpb.TxRequest{
						Class:           ttnpb.Class_CLASS_A,
						Priority:        ns.downlinkPriorities.JoinAccept,
						FrequencyPlanId: dev.FrequencyPlanId,
						Rx1Delay:        ttnpb.RxDelay(phy.JoinAcceptDelay1 / time.Second),
					}
					if attemptRX1 {
						req.Rx1Frequency = rx1Freq
						req.Rx1DataRate = rx1DR.Rate
					}
					if attemptRX2 {
						req.Rx2Frequency = dev.PendingMacState.CurrentParameters.Rx2Frequency
						req.Rx2DataRate = rx2DR.Rate
					}
					down, downEvs, err := ns.scheduleDownlinkByPaths(
						log.NewContext(ctx, loggerWithTxRequestFields(logger, req, attemptRX1, attemptRX2).WithField("rx1_delay", req.Rx1Delay)),
						&scheduleRequest{
							TxRequest:            req,
							EndDeviceIdentifiers: dev.Ids,
							RawPayload:           dev.PendingMacState.QueuedJoinAccept.Payload,
							Payload: &ttnpb.Message{
								MHdr: &ttnpb.MHDR{
									MType: ttnpb.MType_JOIN_ACCEPT,
									Major: ttnpb.Major_LORAWAN_R1,
								},
								Payload: &ttnpb.Message_JoinAcceptPayload{
									JoinAcceptPayload: &ttnpb.JoinAcceptPayload{
										NetId:      dev.PendingMacState.QueuedJoinAccept.NetId,
										DevAddr:    dev.PendingMacState.QueuedJoinAccept.DevAddr,
										DlSettings: dev.PendingMacState.QueuedJoinAccept.Request.DownlinkSettings,
										RxDelay:    dev.PendingMacState.QueuedJoinAccept.Request.RxDelay,
										CfList:     dev.PendingMacState.QueuedJoinAccept.Request.CfList,
									},
								},
							},
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
						dev.PendingMacState.RxWindowsAvailable = false
						taskUpdateStrategy = nextDownlinkTask
						return dev, []string{
							"pending_mac_state.rx_windows_available",
						}, nil
					}

					var invalidatedQueue []*ttnpb.ApplicationDownlink
					if dev.Session != nil {
						invalidatedQueue = dev.Session.QueuedApplicationDownlinks
					} else {
						invalidatedQueue = dev.GetPendingSession().GetQueuedApplicationDownlinks()
					}
					queuedApplicationUplinks = append(queuedApplicationUplinks, &ttnpb.ApplicationUp{
						EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
							ApplicationIds: dev.Ids.ApplicationIds,
							DeviceId:       dev.Ids.DeviceId,
							DevEui:         dev.Ids.DevEui,
							JoinEui:        dev.Ids.JoinEui,
							DevAddr:        &dev.PendingMacState.QueuedJoinAccept.DevAddr,
						},
						CorrelationIds: events.CorrelationIDsFromContext(ctx),
						Up: &ttnpb.ApplicationUp_JoinAccept{
							JoinAccept: &ttnpb.ApplicationJoinAccept{
								AppSKey:              dev.PendingMacState.QueuedJoinAccept.Keys.AppSKey,
								InvalidatedDownlinks: invalidatedQueue,
								SessionKeyId:         dev.PendingMacState.QueuedJoinAccept.Keys.SessionKeyId,
								ReceivedAt:           up.ReceivedAt,
							},
						},
					})

					dev.PendingSession = &ttnpb.Session{
						DevAddr: dev.PendingMacState.QueuedJoinAccept.DevAddr,
						Keys: &ttnpb.SessionKeys{
							SessionKeyId: dev.PendingMacState.QueuedJoinAccept.Keys.SessionKeyId,
							FNwkSIntKey:  dev.PendingMacState.QueuedJoinAccept.Keys.FNwkSIntKey,
							SNwkSIntKey:  dev.PendingMacState.QueuedJoinAccept.Keys.SNwkSIntKey,
							NwkSEncKey:   dev.PendingMacState.QueuedJoinAccept.Keys.NwkSEncKey,
						},
						QueuedApplicationDownlinks: nil,
					}
					dev.PendingMacState.PendingJoinRequest = dev.PendingMacState.QueuedJoinAccept.Request
					dev.PendingMacState.QueuedJoinAccept = nil
					dev.PendingMacState.RxWindowsAvailable = false
					dev.PendingMacState.RecentDownlinks = appendRecentDownlink(dev.PendingMacState.RecentDownlinks, &ttnpb.DownlinkMessage{
						Payload:        down.Message.Payload,
						Settings:       down.Message.Settings,
						CorrelationIds: down.Message.CorrelationIds,
					}, recentDownlinkCount)
					return dev, []string{
						"pending_mac_state.pending_join_request",
						"pending_mac_state.queued_join_accept",
						"pending_mac_state.recent_downlinks",
						"pending_mac_state.rx_windows_available",
						"pending_session.dev_addr",
						"pending_session.keys",
						"pending_session.queued_application_downlinks",
					}, nil
				}

				logger = logger.WithField("downlink_type", "data")
				if dev.Session == nil {
					logger.Warn("Unknown session, skip downlink slot")
					return dev, nil, nil
				}
				logger = logger.WithField("dev_addr", dev.Session.DevAddr)

				if dev.MacState == nil {
					logger.Warn("Unknown MAC state, skip downlink slot")
					return dev, nil, nil
				}
				logger = logger.WithField("device_class", dev.MacState.DeviceClass)

				ctx = log.NewContext(ctx, logger)

				var maxUpLength uint16 = math.MaxUint16
				if !dev.Multicast && dev.MacState.LorawanVersion == ttnpb.MAC_V1_1 {
					maxUpLength, err = maximumUplinkLength(fp, phy, dev.MacState.RecentUplinks...)
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
						case slot.Class == ttnpb.Class_CLASS_B && slot.Time.IsZero(),
							slot.IsApplicationTime && slot.Time.IsZero():
							logger.Error("Invalid downlink slot generated, skip class B/C downlink slot")
							return dev, nil, nil

						case !slot.IsApplicationTime && slot.Class == ttnpb.Class_CLASS_C && time.Until(slot.Time) > 0:
							logger.WithFields(log.Fields(
								"slot_start", slot.Time,
							)).Info("Class C downlink scheduling attempt performed too soon, retry attempt")
							taskUpdateStrategy = nextDownlinkTask
							return dev, nil, nil

						case time.Until(slot.Time) > dev.MacState.CurrentParameters.Rx1Delay.Duration()+2*nsScheduleWindow():
							logger.WithFields(log.Fields(
								"slot_start", slot.Time,
							)).Info("Class B/C downlink scheduling attempt performed too soon, retry attempt")
							taskUpdateStrategy = nextDownlinkTask
							return dev, nil, nil

						case !slot.IsApplicationTime && slot.Class == ttnpb.Class_CLASS_B && time.Until(slot.Time) < dev.MacState.CurrentParameters.Rx1Delay.Duration()/2:
							earliestAt = time.Now().Add(dev.MacState.CurrentParameters.Rx1Delay.Duration() / 2)
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
			return time.Time{}, err
		}

		var earliestAt time.Time
		switch taskUpdateStrategy {
		case nextDownlinkTask:

		case retryDownlinkTask:
			earliestAt = time.Now().Add(downlinkRetryInterval + nsScheduleWindow())

		case noDownlinkTask:
			return time.Time{}, nil

		default:
			panic(fmt.Errorf("unmatched downlink task update strategy: %v", taskUpdateStrategy))
		}
		nextTaskAt, err := ns.nextDataDownlinkTaskAt(ctx, dev, earliestAt)
		if err != nil {
			computeNextErr = true
			logger.WithError(err).Error("Failed to compute next downlink task time after downlink attempt")
			return time.Time{}, nil
		}
		return nextTaskAt, nil
	})
	if err != nil && !setErr && !computeNextErr {
		log.FromContext(ctx).WithError(err).Error("Failed to pop entry from downlink task queue")
	}
	return err
}
