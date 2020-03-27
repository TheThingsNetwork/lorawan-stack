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
	if dev.MACState == nil || dev.Session == nil {
		logger.Debug("Avoid updating downlink task queue for device with no MAC state or session")
		return nil
	}

	if t := timeNow().UTC().Add(nsScheduleWindow()); earliestAt.Before(t) {
		earliestAt = t
	}
	var t time.Time
	_, phy, err := getDeviceBandVersion(dev, ns.FrequencyPlans)
	if err != nil {
		logger.WithError(err).Warn("Failed to determine device band")
		t = earliestAt
	} else {
		delay := dev.MACState.CurrentParameters.Rx1Delay.Duration() / 2
		var ok bool
		t, _, ok = nextDataDownlinkAt(ctx, dev, phy, ns.defaultMACSettings, earliestAt.Add(delay))
		if !ok {
			return nil
		}
		if t = t.Add(-nsScheduleWindow() - delay); t.Before(earliestAt) {
			t = earliestAt
		}
	}
	logger.WithField("start_at", t).Debug("Add downlink task")
	return ns.downlinkTasks.Add(ctx, dev.EndDeviceIdentifiers, t, true)
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
				genState.Events = append(genState.Events, ev(ctx, dev.EndDeviceIdentifiers))
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
				ADR: deviceUseADR(dev, ns.defaultMACSettings),
			},
		},
	}
	logger = logger.WithFields(log.Fields(
		"ack", pld.FHDR.FCtrl.Ack,
		"adr", pld.FHDR.FCtrl.ADR,
	))
	ctx = log.NewContext(ctx, logger)

	if len(cmdBuf) <= fOptsCapacity {
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
				invalid, rest := partitionDownlinksBySessionKeyIDEquality(dev.Session.SessionKeyID, dev.Session.QueuedApplicationDownlinks[i:]...)
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
		pld.FHDR.FCnt = dev.Session.LastNFCntDown + 1

	default:
		return nil, genState, errNoDownlink.New()
	}

	logger = logger.WithFields(log.Fields(
		"f_cnt", pld.FHDR.FCnt,
		"f_port", pld.FPort,
		"m_type", mType,
	))
	ctx = log.NewContext(ctx, logger)

	if len(cmdBuf) > 0 && (pld.FPort == 0 || dev.MACState.LoRaWANVersion.EncryptFOpts()) {
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
		cmdBuf, err = crypto.EncryptDownlink(key, dev.Session.DevAddr, fCnt, cmdBuf)
		if err != nil {
			return nil, genState, errEncryptMAC.WithCause(err)
		}
	}
	if pld.FPort == 0 {
		pld.FRMPayload = cmdBuf
	} else {
		pld.FHDR.FOpts = cmdBuf
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
	if needsAck {
		var confirmedAt time.Time
		switch class {
		case ttnpb.CLASS_B:
			confirmedAt, _ = nextConfirmedClassBDownlinkAt(ctx, dev, ns.defaultMACSettings, transmitAt)

		case ttnpb.CLASS_C:
			confirmedAt, _ = nextConfirmedClassCDownlinkAt(ctx, dev, ns.defaultMACSettings, transmitAt)
		}
		if confirmedAt.After(transmitAt) {
			logger.WithField("confirmed_at", confirmedAt).Debug("Confirmed class B/C downlink attempt performed too soon")
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

// classAWindowsAvailableAt returns whether class A downlink can be made following up
// in either window considering the given Rx delay.
func classAWindowsAvailableAt(up *ttnpb.UplinkMessage, rxDelay time.Duration, earliestAt time.Time) (rx1, rx2 time.Time) {
	rx1, rx2 = classAWindows(up, rxDelay)
	switch {
	case !earliestAt.After(rx1):
		return rx1, rx2
	case !earliestAt.After(rx2):
		return time.Time{}, rx2
	}
	return time.Time{}, time.Time{}
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
		return nil, errNoPath.New()
	}

	logger := log.FromContext(ctx)

	type attempt struct {
		peer  cluster.Peer
		paths []*ttnpb.DownlinkPath
	}
	attempts := make([]*attempt, 0, len(paths))
	lastAttempt := func() *attempt {
		return attempts[len(attempts)-1]
	}

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
		if len(attempts) > 0 && lastAttempt().peer == p {
			a = lastAttempt()
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
		transmitAt := timeNow().Add(res.Delay)
		logger.WithFields(log.Fields(
			"transmission_delay", res.Delay,
			"transmit_at", transmitAt,
		)).Debug("Scheduled downlink")
		return &scheduledDownlink{
			Message:    down,
			TransmitAt: transmitAt,
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

// txRequestFromUplink return the Class A TxRequest, which can be used to answer up.
// txRequestFromUplink does not set the priority.
func txRequestFromUplink(phy band.Band, macState *ttnpb.MACState, rx1, rx2 bool, rxDelay time.Duration, up *ttnpb.UplinkMessage) (*ttnpb.TxRequest, error) {
	if !rx1 && !rx2 {
		return nil, errNoPath.New()
	}
	req := &ttnpb.TxRequest{
		Class:    ttnpb.CLASS_A,
		Rx1Delay: ttnpb.RxDelay(rxDelay / time.Second),
	}
	if rx1 {
		if up.DeviceChannelIndex > math.MaxUint8 {
			return nil, errInvalidChannelIndex.New()
		}
		rx1ChIdx, err := phy.Rx1Channel(uint8(up.DeviceChannelIndex))
		if err != nil {
			return nil, err
		}
		if uint(rx1ChIdx) >= uint(len(macState.CurrentParameters.Channels)) ||
			macState.CurrentParameters.Channels[int(rx1ChIdx)].GetDownlinkFrequency() == 0 {
			return nil, errCorruptedMACState.New()
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
		return 0, errDataRateNotFound.New()
	}
	return dr.MaxMACPayloadSize(fp.DwellTime.GetUplinks()), nil
}

// downlinkRetryInterval is the time interval, which defines the interval between downlink task retries.
const downlinkRetryInterval = 2 * time.Second

func recordDataDownlink(dev *ttnpb.EndDevice, genDown *generatedDownlink, genState generateDownlinkState, down *scheduledDownlink, defaults ttnpb.MACSettings) {
	if genState.ApplicationDownlink == nil || dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 && genDown.FCnt > dev.Session.LastNFCntDown {
		dev.Session.LastNFCntDown = genDown.FCnt
	}
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

type downlinkAttemptResult struct {
	applicationUpAppender func([]*ttnpb.ApplicationUp, bool) []*ttnpb.ApplicationUp

	SetPaths     []string
	TransmitAt   time.Time
	Scheduled    bool
	QueuedEvents []events.Event
}

func (res downlinkAttemptResult) AppendApplicationUplinks(ups ...*ttnpb.ApplicationUp) []*ttnpb.ApplicationUp {
	if res.applicationUpAppender == nil {
		return ups
	}
	return res.applicationUpAppender(ups, res.Scheduled)
}

func (ns *NetworkServer) attemptClassADataDownlink(ctx context.Context, dev *ttnpb.EndDevice, phy band.Band, fp *frequencyplans.FrequencyPlan, maxUpLength uint16) downlinkAttemptResult {
	var sets []string
	logger := log.FromContext(ctx)
	if len(dev.MACState.RecentUplinks) == 0 {
		logger.Warn("Rx windows available, but no uplink present, skip class A downlink slot")
		dev.MACState.QueuedResponses = nil
		dev.MACState.RxWindowsAvailable = false
		return downlinkAttemptResult{
			SetPaths: ttnpb.AddFields(sets,
				"mac_state.queued_responses",
				"mac_state.rx_windows_available",
			),
		}
	}

	var rxDelay time.Duration
	up := lastUplink(dev.MACState.RecentUplinks...)
	switch up.Payload.MHDR.MType {
	case ttnpb.MType_CONFIRMED_UP, ttnpb.MType_UNCONFIRMED_UP:
		rxDelay = dev.MACState.CurrentParameters.Rx1Delay.Duration()

	case ttnpb.MType_REJOIN_REQUEST:
		rxDelay = phy.JoinAcceptDelay1

	default:
		logger.Warn("Last uplink is neither data uplink, nor rejoin-request, skip class A downlink slot")
		dev.MACState.QueuedResponses = nil
		dev.MACState.RxWindowsAvailable = false
		return downlinkAttemptResult{
			SetPaths: ttnpb.AddFields(sets,
				"mac_state.queued_responses",
				"mac_state.rx_windows_available",
			),
		}
	}
	ctx = events.ContextWithCorrelationID(ctx, up.CorrelationIDs...)

	paths := downlinkPathsFromRecentUplinks(dev.MACState.RecentUplinks...)
	if len(paths) == 0 {
		logger.Warn("No downlink path available, skip class A downlink slot")
		dev.MACState.QueuedResponses = nil
		dev.MACState.RxWindowsAvailable = false
		return downlinkAttemptResult{
			SetPaths: ttnpb.AddFields(sets,
				"mac_state.queued_responses",
				"mac_state.rx_windows_available",
			),
		}
	}

	rx1, rx2 := classAWindowsAvailableAt(up, rxDelay, timeNow().Add(infrastructureDelay/2))
	attemptRx1 := !rx1.IsZero()
	attemptRx2 := !rx2.IsZero()
	if !attemptRx1 && !attemptRx2 {
		logger.Warn("Rx1 and Rx2 are expired, skip class A downlink slot")
		dev.MACState.QueuedResponses = nil
		dev.MACState.RxWindowsAvailable = false
		return downlinkAttemptResult{
			SetPaths: ttnpb.AddFields(sets,
				"mac_state.queued_responses",
				"mac_state.rx_windows_available",
			),
		}
	}

	req, err := txRequestFromUplink(phy, dev.MACState, attemptRx1, attemptRx2, rxDelay, up)
	if err != nil {
		logger.WithError(err).Warn("Failed to generate Tx request from uplink, skip class A downlink slot")
		dev.MACState.QueuedResponses = nil
		dev.MACState.RxWindowsAvailable = false
		return downlinkAttemptResult{
			SetPaths: ttnpb.AddFields(sets,
				"mac_state.queued_responses",
				"mac_state.rx_windows_available",
			),
		}
	}

	// transmitAt is the earliest time.Time when downlink will be transmitted to the device.
	var transmitAt time.Time
	var maxDRIdx ttnpb.DataRateIndex
	switch {
	case attemptRx1 && attemptRx2:
		transmitAt = rx1
		maxDRIdx = req.Rx1DataRateIndex
		if req.Rx2DataRateIndex > maxDRIdx {
			maxDRIdx = req.Rx2DataRateIndex
		}

	case attemptRx1:
		transmitAt = rx1
		maxDRIdx = req.Rx1DataRateIndex

	case attemptRx2:
		transmitAt = rx2
		maxDRIdx = req.Rx2DataRateIndex
	}

	maxDR, ok := phy.DataRates[maxDRIdx]
	if !ok {
		logger.WithField("data_rate_index", maxDRIdx).Error("Data rate not found")
		return downlinkAttemptResult{
			SetPaths: ttnpb.AddFields(sets,
				"mac_state.queued_responses",
				"mac_state.rx_windows_available",
			),
		}
	}
	genDown, genState, err := ns.generateDataDownlink(ctx, dev, phy, ttnpb.CLASS_A, transmitAt,
		maxDR.MaxMACPayloadSize(fp.DwellTime.GetDownlinks()),
		maxUpLength,
	)
	if genState.NeedsDownlinkQueueUpdate {
		sets = ttnpb.AddFields(sets,
			"session.queued_application_downlinks",
		)
	}
	if err != nil {
		switch {
		case errors.Resemble(err, errNoDownlink):
			logger.Debug("No class A downlink to send, skip class A downlink slot")

		default:
			logger.WithError(err).Warn("Failed to generate class A downlink, skip class A downlink slot")
		}
		if genState.ApplicationDownlink != nil {
			dev.Session.QueuedApplicationDownlinks = append([]*ttnpb.ApplicationDownlink{genState.ApplicationDownlink}, dev.Session.QueuedApplicationDownlinks...)
		}
		return downlinkAttemptResult{
			SetPaths:              sets,
			applicationUpAppender: genState.appendApplicationUplinks,
		}
	}

	if attemptRx1 && attemptRx2 {
		dr1, ok := phy.DataRates[req.Rx1DataRateIndex]
		if !ok {
			logger.WithField("data_rate_index", req.Rx1DataRateIndex).Error("Rx1 data rate not found")
		}
		dr2, ok := phy.DataRates[req.Rx2DataRateIndex]
		if !ok {
			logger.WithField("data_rate_index", req.Rx2DataRateIndex).Error("Rx2 data rate not found")
		}
		attemptRx1 = len(genDown.Payload) <= int(dr1.MaxMACPayloadSize(fp.DwellTime.GetDownlinks()))
		attemptRx2 = len(genDown.Payload) <= int(dr2.MaxMACPayloadSize(fp.DwellTime.GetDownlinks()))
		if !attemptRx1 && !attemptRx2 {
			logger.Error("Generated downlink payload size does not fit neither Rx1, nor Rx2, skip class A downlink slot")
			dev.MACState.QueuedResponses = nil
			dev.MACState.RxWindowsAvailable = false
			return downlinkAttemptResult{
				SetPaths: ttnpb.AddFields(sets,
					"mac_state.queued_responses",
					"mac_state.rx_windows_available",
				),
				applicationUpAppender: genState.appendApplicationUplinks,
			}
		}
		// NOTE: It may be possible that RX1 is dropped at this point and DevStatusReq can be scheduled in RX2 due to the downlink being
		// transmitted later, but that's micro-optimization, which we don't need to make.
		req, err = txRequestFromUplink(phy, dev.MACState, attemptRx1, attemptRx2, rxDelay, up)
		if err != nil {
			logger.WithError(err).Warn("Failed to generate Tx request from uplink, skip class A downlink slot")
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
				applicationUpAppender: genState.appendApplicationUplinks,
			}
		}
	}

	if genState.ApplicationDownlink != nil {
		ctx = events.ContextWithCorrelationID(ctx, genState.ApplicationDownlink.CorrelationIDs...)
	}
	req.FrequencyPlanID = dev.FrequencyPlanID
	req.Priority = genDown.Priority

	down, err := ns.scheduleDownlinkByPaths(
		log.NewContext(ctx, loggerWithTxRequestFields(logger, req, attemptRx1, attemptRx2).WithField("rx1_delay", req.Rx1Delay)),
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
			applicationUpAppender: genState.appendApplicationUplinks,
		}
	}
	if genState.ApplicationDownlink != nil {
		sets = ttnpb.AddFields(sets, "session.queued_application_downlinks")
	}
	recordDataDownlink(dev, genDown, genState, down, ns.defaultMACSettings)
	return downlinkAttemptResult{
		SetPaths: ttnpb.AddFields(sets,
			"mac_state.last_confirmed_downlink_at",
			"mac_state.pending_application_downlink",
			"mac_state.pending_requests",
			"mac_state.queued_responses",
			"mac_state.recent_downlinks",
			"mac_state.rx_windows_available",
			"recent_downlinks",
			"session",
		),
		applicationUpAppender: genState.appendApplicationUplinks,
		TransmitAt:            down.TransmitAt,
		QueuedEvents:          genState.Events,
		Scheduled:             true,
	}
}

// processDownlinkTask processes the most recent downlink task ready for execution, if such is available or wait until it is before processing it.
// NOTE: ctx.Done() is not guaranteed to be respected by processDownlinkTask.
func (ns *NetworkServer) processDownlinkTask(ctx context.Context) error {
	var setErr bool
	var addErr bool
	err := ns.downlinkTasks.Pop(ctx, func(ctx context.Context, devID ttnpb.EndDeviceIdentifiers, t time.Time) error {
		logger := log.FromContext(ctx).WithFields(log.Fields(
			"device_uid", unique.ID(ctx, devID),
			"started_at", timeNow().UTC(),
		))
		ctx = log.NewContext(ctx, logger)
		logger.WithField("start_at", t).Debug("Process downlink task")

		var queuedApplicationUplinks []*ttnpb.ApplicationUp
		var queuedEvents []events.Event
		var retryTask bool
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
					retryTask = true
					logger.WithError(err).Error("Failed to get frequency plan of the device, retry downlink slot")
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
					up := lastUplink(dev.RecentUplinks...)
					switch up.Payload.MHDR.MType {
					case ttnpb.MType_JOIN_REQUEST, ttnpb.MType_REJOIN_REQUEST:
					default:
						logger.Warn("Last uplink is neither join-request, nor rejoin-request, skip downlink slot")
						return dev, nil, nil
					}
					ctx := events.ContextWithCorrelationID(ctx, up.CorrelationIDs...)

					paths := downlinkPathsFromRecentUplinks(up)
					if len(paths) == 0 {
						logger.Warn("No downlink path available, skip join-accept downlink slot")
						dev.PendingMACState.RxWindowsAvailable = false
						return dev, []string{
							"pending_mac_state.rx_windows_available",
						}, nil
					}

					// Be more optimistic when scheduling join-accepts and assume minimum transmission delay.
					rx1, rx2 := classAWindowsAvailableAt(up, phy.JoinAcceptDelay1, timeNow().Add(infrastructureDelay/2))
					attemptRx1 := !rx1.IsZero()
					attemptRx2 := !rx2.IsZero()
					if !attemptRx1 && !attemptRx2 {
						logger.Warn("Rx1 and Rx2 are expired, skip join-accept downlink slot")
						dev.PendingMACState.RxWindowsAvailable = false
						return dev, []string{
							"pending_mac_state.rx_windows_available",
						}, nil
					}

					req, err := txRequestFromUplink(phy, dev.PendingMACState, attemptRx1, attemptRx2, phy.JoinAcceptDelay1, up)
					if err != nil {
						logger.WithError(err).Warn("Failed to generate Tx request from uplink, skip downlink slot")
						return dev, nil, nil
					}
					req.Priority = ns.downlinkPriorities.JoinAccept
					req.FrequencyPlanID = dev.FrequencyPlanID

					down, err := ns.scheduleDownlinkByPaths(
						log.NewContext(ctx, loggerWithTxRequestFields(logger, req, attemptRx1, attemptRx2).WithField("rx1_delay", req.Rx1Delay)),
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
				if dev.Session == nil {
					logger.Warn("Unknown session, skip downlink slot")
					return dev, nil, nil
				}

				logger = logger.WithFields(log.Fields(
					"dev_addr", dev.Session.DevAddr,
					"device_class", dev.MACState.DeviceClass,
				))
				ctx = log.NewContext(ctx, logger)

				if !dev.MACState.RxWindowsAvailable {
					logger.Debug("Rx windows not available, skip class A downlink slot")
					if dev.MACState.DeviceClass == ttnpb.CLASS_A {
						return dev, nil, nil
					}
				}

				var maxUpLength uint16 = math.MaxUint16
				if dev.MACState.LoRaWANVersion == ttnpb.MAC_V1_1 {
					maxUpLength, err = maximumUplinkLength(fp, phy, dev.MACState.RecentUplinks...)
					if err != nil {
						logger.WithError(err).Error("Failed to determine maximum uplink length")
						return dev, nil, nil
					}
				}

				var sets []string
				if dev.MACState.RxWindowsAvailable {
					a := ns.attemptClassADataDownlink(ctx, dev, phy, fp, maxUpLength)
					sets = ttnpb.AddFields(sets, a.SetPaths...)
					queuedEvents = append(queuedEvents, a.QueuedEvents...)
					queuedApplicationUplinks = a.AppendApplicationUplinks(queuedApplicationUplinks...)
					if a.Scheduled {
						return dev, sets, nil
					}
				}
				if dev.MACState.DeviceClass == ttnpb.CLASS_A {
					return dev, sets, nil
				}

				transmissionDelay := dev.MACState.CurrentParameters.Rx1Delay.Duration() / 2

				// Class B/C data downlink
				transmitAt, class, ok := nextDataDownlinkAt(ctx, dev, phy, ns.defaultMACSettings, timeNow().UTC().Add(transmissionDelay))
				if !ok || class == ttnpb.CLASS_A {
					logger.Debug("No class B/C downlink available, skip class B/C downlink slot")
					return dev, sets, nil
				}
				if delay := transmitAt.Sub(timeNow()); delay > 2*transmissionDelay+2*nsScheduleWindow() {
					logger.WithFields(log.Fields(
						"delay", delay,
						"transmit_at", transmitAt,
					)).Info("Class B/C downlink scheduling attempt performed too soon, retry attempt")
					return dev, sets, nil
				}

				var drIdx ttnpb.DataRateIndex
				var freq uint64
				switch class {
				case ttnpb.CLASS_B:
					if dev.MACState.CurrentParameters.PingSlotDataRateIndexValue == nil {
						logger.Error("Device is in class B mode, but ping slot data rate index is not known")
						return dev, sets, nil
					}
					drIdx = dev.MACState.CurrentParameters.PingSlotDataRateIndexValue.Value
					freq = dev.MACState.CurrentParameters.PingSlotFrequency

				case ttnpb.CLASS_C:
					drIdx = dev.MACState.CurrentParameters.Rx2DataRateIndex
					freq = dev.MACState.CurrentParameters.Rx2Frequency

				default:
					panic(fmt.Sprintf("unmatched downlink class: '%s'", class))
				}

				dr, ok := phy.DataRates[drIdx]
				if !ok {
					logger.WithField("data_rate_index", drIdx).Error("Rx2 data rate not found")
					return dev, sets, nil
				}
				genDown, genState, err := ns.generateDataDownlink(ctx, dev, phy, class, transmitAt,
					dr.MaxMACPayloadSize(fp.DwellTime.GetDownlinks()),
					maxUpLength,
				)
				if genState.NeedsDownlinkQueueUpdate {
					sets = ttnpb.AddFields(sets, "session.queued_application_downlinks")
				}
				if err != nil {
					logger.WithError(err).Warn("Failed to generate class B/C downlink, skip class B/C downlink slot")
					queuedApplicationUplinks = genState.appendApplicationUplinks(queuedApplicationUplinks, false)
					if genState.ApplicationDownlink != nil && ttnpb.HasAnyField(sets, "session.queued_application_downlinks") {
						dev.Session.QueuedApplicationDownlinks = append([]*ttnpb.ApplicationDownlink{genState.ApplicationDownlink}, dev.Session.QueuedApplicationDownlinks...)
					}
					return dev, sets, nil
				}

				if genState.ApplicationDownlink != nil {
					ctx = events.ContextWithCorrelationID(ctx, genState.ApplicationDownlink.CorrelationIDs...)
				}

				var paths []downlinkPath
				if fixedPaths := genState.ApplicationDownlink.GetClassBC().GetGateways(); len(fixedPaths) > 0 {
					paths = make([]downlinkPath, 0, len(fixedPaths))
					for i := range fixedPaths {
						paths = append(paths, downlinkPath{
							GatewayIdentifiers: fixedPaths[i].GatewayIdentifiers,
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
						logger.Warn("No downlink path available, skip class B/C downlink slot")
						queuedApplicationUplinks = genState.appendApplicationUplinks(queuedApplicationUplinks, false)
						if genState.ApplicationDownlink != nil && ttnpb.HasAnyField(sets, "session.queued_application_downlinks") {
							dev.Session.QueuedApplicationDownlinks = append([]*ttnpb.ApplicationDownlink{genState.ApplicationDownlink}, dev.Session.QueuedApplicationDownlinks...)
						}
						return dev, sets, nil
					}
				}

				req := &ttnpb.TxRequest{
					Class:            class,
					Priority:         genDown.Priority,
					Rx2DataRateIndex: drIdx,
					Rx2Frequency:     freq,
					FrequencyPlanID:  dev.FrequencyPlanID,
				}
				switch {
				case req.Class == ttnpb.CLASS_B && transmitAt.IsZero():
					logger.Error("Class B downlink with no absolute time generated")
					return dev, sets, nil

				case genState.ApplicationDownlink.GetClassBC().GetAbsoluteTime() != nil:
					req.AbsoluteTime = genState.ApplicationDownlink.GetClassBC().GetAbsoluteTime()

				case req.Class == ttnpb.CLASS_B, req.Class == ttnpb.CLASS_C && transmitAt.After(timeNow().Add(transmissionDelay)):
					req.AbsoluteTime = &transmitAt
				}

				down, err := ns.scheduleDownlinkByPaths(
					log.NewContext(ctx, loggerWithTxRequestFields(logger, req, false, true)),
					req,
					genDown.Payload,
					paths...,
				)
				if err != nil {
					retryTask = true
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
								logger.Warn("Absolute time invalid, fail downlink and retry attempt")
								retryTask = true
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
									sets = ttnpb.AddFields(sets, "session.queued_application_downlinks")
								}
								return dev, sets, nil
							}

							if len(genState.ApplicationDownlink.GetClassBC().GetGateways()) > 0 &&
								allErrors(nonRetryableFixedPathGatewayError, pathErrs...) {
								logger.Warn("Fixed paths invalid, fail application downlink and retry attempt")
								retryTask = true
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
									sets = ttnpb.AddFields(sets, "session.queued_application_downlinks")
								}
								return dev, sets, nil
							}
						}
					}
					logger.Warn("All Gateway Servers failed to schedule downlink, skip class B/C downlink slot")
					if genState.NeedsDownlinkQueueUpdate {
						dev.Session.QueuedApplicationDownlinks = append([]*ttnpb.ApplicationDownlink{genState.ApplicationDownlink}, dev.Session.QueuedApplicationDownlinks...)
					}
					return dev, sets, nil
				}

				recordDataDownlink(dev, genDown, genState, down, ns.defaultMACSettings)
				queuedEvents = append(queuedEvents, genState.Events...)
				queuedApplicationUplinks = genState.appendApplicationUplinks(queuedApplicationUplinks, true)
				if genState.ApplicationDownlink != nil {
					sets = ttnpb.AddFields(sets, "session.queued_application_downlinks")
				}
				return dev, ttnpb.AddFields(sets,
					"mac_state.last_confirmed_downlink_at",
					"mac_state.last_network_initiated_downlink_at",
					"mac_state.pending_application_downlink",
					"mac_state.pending_requests",
					"mac_state.queued_responses",
					"mac_state.recent_downlinks",
					"mac_state.rx_windows_available",
					"recent_downlinks",
					"session",
				), nil
			},
		)
		if len(queuedApplicationUplinks) > 0 {
			if err := ns.applicationUplinks.Add(ctx, queuedApplicationUplinks...); err != nil {
				logger.WithError(err).Warn("Failed to queue application uplinks for sending to Application Server")
			}
		}
		if len(queuedEvents) > 0 {
			for _, ev := range queuedEvents {
				events.Publish(ev)
			}
		}
		if err != nil {
			setErr = true
			logger.WithError(err).Error("Failed to update device in registry")
			return err
		}

		if retryTask {
			if err := ns.updateDataDownlinkTask(ctx, dev, timeNow().Add(downlinkRetryInterval)); err != nil {
				addErr = true
				logger.WithError(err).Error("Failed to update downlink task queue after downlink attempt")
				return err
			}
		} else if dev != nil {
			if err := ns.updateDataDownlinkTask(ctx, dev, time.Time{}); err != nil {
				addErr = true
				logger.WithError(err).Error("Failed to update downlink task queue after downlink attempt")
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
