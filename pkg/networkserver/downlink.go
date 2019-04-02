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

	"github.com/mohae/deepcopy"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
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

var errApplicationDownlinkTooLong = errors.DefineInvalidArgument("application_downlink_too_long", "application downlink payload is too long")
var errNoDownlink = errors.Define("no_downlink", "no downlink to send")

type generatedDownlink struct {
	Payload             []byte
	FCnt                uint32
	ApplicationDownlink *ttnpb.ApplicationDownlink
}

// generateDownlink attempts to generate a downlink.
// generateDownlink returns the marshaled payload of the downlink, application downlink, if included in the payload and error, if any.
// generateDownlink may mutate the device in order to record the downlink generated.
// maxDownLen and maxUpLen represent the maximum length of PHYPayload for the downlink and corresponding uplink respectively.
// If no downlink could be generated - nil, errNoDownlink is returned.
// generateDownlink does not perform validation of dev.MACState.DesiredParameters,
// hence, it could potentially generate MAC command(s), which are not suported by the
// regional parameters the device operates in.
// For example, a sequence of 'NewChannel' MAC commands could be generated for a
// device operating in a region where a fixed channel plan is defined in case
// dev.MACState.CurrentParameters.Channels is not equal to dev.MACState.DesiredParameters.Channels.
func (ns *NetworkServer) generateDownlink(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16) (*generatedDownlink, error) {
	if dev.MACState == nil {
		return nil, errUnknownMACState
	}
	if dev.Session == nil {
		return nil, errEmptySession
	}

	ctx = log.NewContextWithField(ctx, "device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers))
	logger := log.FromContext(ctx)

	// NOTE: len(MHDR) + len(MIC) = 1 + 4 = 5
	if maxDownLen < 5 || maxUpLen < 5 {
		panic("payload length limits too short to generate downlink")
	}
	maxDownLen, maxUpLen = maxDownLen-5, maxUpLen-5

	spec := lorawan.DefaultMACCommands
	cmds := make([]*ttnpb.MACCommand, 0, len(dev.MACState.QueuedResponses)+len(dev.MACState.PendingRequests))
	for _, cmd := range dev.MACState.QueuedResponses {
		desc := spec[cmd.CID]
		if desc == nil {
			maxDownLen = 0
			continue
		}
		if desc.DownlinkLength > maxDownLen {
			continue
		}
		cmds = append(cmds, cmd)
		maxDownLen -= desc.DownlinkLength
	}

	dev.MACState.PendingRequests = dev.MACState.PendingRequests[:0]

	maxDownLen, maxUpLen, ok, err := enqueueLinkADRReq(ctx, dev, maxDownLen, maxUpLen, ns.FrequencyPlans)
	if err != nil {
		return nil, err
	}
	fPending := !ok
	for _, f := range []func(context.Context, *ttnpb.EndDevice, uint16, uint16) (uint16, uint16, bool){
		// LoRaWAN 1.0+
		enqueueNewChannelReq,
		enqueueDutyCycleReq,
		enqueueRxParamSetupReq,
		func(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen uint16, maxUpLen uint16) (uint16, uint16, bool) {
			return enqueueDevStatusReq(ctx, dev, maxDownLen, maxUpLen, ns.defaultMACSettings)
		},
		enqueueRxTimingSetupReq,
		enqueuePingSlotChannelReq,
		enqueueBeaconFreqReq,

		// LoRaWAN 1.0.2+
		enqueueTxParamSetupReq,
		enqueueDLChannelReq,

		// LoRaWAN 1.1+
		enqueueADRParamSetupReq,
		enqueueForceRejoinReq,
		enqueueRejoinParamSetupReq,
	} {
		var ok bool
		maxDownLen, maxUpLen, ok = f(ctx, dev, maxDownLen, maxUpLen)
		fPending = fPending || !ok
	}
	cmds = append(cmds, dev.MACState.PendingRequests...)

	mType := ttnpb.MType_UNCONFIRMED_DOWN
	cmdBuf := make([]byte, 0, maxDownLen)
	for _, cmd := range cmds {
		logger := logger.WithField("cid", cmd.CID)
		logger.Debug("Add MAC command to buffer")
		var err error
		cmdBuf, err = spec.AppendDownlink(cmdBuf, *cmd)
		if err != nil {
			return nil, errEncodeMAC.WithCause(err)
		}
		if mType == ttnpb.MType_UNCONFIRMED_DOWN && spec[cmd.CID].ExpectAnswer && dev.MACState.DeviceClass == ttnpb.CLASS_C {
			logger.Debug("Using confirmed downlink to get immediate answer")
			mType = ttnpb.MType_CONFIRMED_DOWN
		}
	}
	logger = logger.WithField("mac_count", len(cmds))

	var needsDownlink bool
	var up *ttnpb.UplinkMessage
	if dev.MACState.RxWindowsAvailable && len(dev.RecentUplinks) > 0 {
		up = dev.RecentUplinks[len(dev.RecentUplinks)-1]
		switch up.Payload.MHDR.MType {
		case ttnpb.MType_UNCONFIRMED_UP:
			if needsDownlink = up.Payload.GetMACPayload().FCtrl.ADRAckReq; needsDownlink {
				logger.Debug("Need downlink for ADRAckReq")
			}
		case ttnpb.MType_CONFIRMED_UP:
			needsDownlink = true
			logger.Debug("Need downlink for confirmed uplink")
		}
	}
	if !needsDownlink &&
		len(cmdBuf) == 0 &&
		len(dev.QueuedApplicationDownlinks) == 0 {
		return nil, errNoDownlink
	}

	pld := &ttnpb.MACPayload{
		FHDR: ttnpb.FHDR{
			DevAddr: dev.Session.DevAddr,
			FCtrl: ttnpb.FCtrl{
				Ack: up != nil && up.Payload.MHDR.MType == ttnpb.MType_CONFIRMED_UP,
			},
			FCnt: dev.Session.LastNFCntDown + 1,
		},
	}
	logger = logger.WithField("ack", pld.FHDR.FCtrl.Ack)

	var appDown *ttnpb.ApplicationDownlink
	if len(cmdBuf) <= fOptsCapacity && len(dev.QueuedApplicationDownlinks) > 0 {
		down := dev.QueuedApplicationDownlinks[0]
		if len(down.FRMPayload) > int(maxDownLen) {
			logger.Warn("Application downlink present, but the payload is too long, inform Application Server")
			ok, err := ns.handleASUplink(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers, &ttnpb.ApplicationUp{
				EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
				CorrelationIDs:       append(events.CorrelationIDsFromContext(ctx), down.CorrelationIDs...),
				Up: &ttnpb.ApplicationUp_DownlinkFailed{
					DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
						ApplicationDownlink: *down,
						Error:               *ttnpb.ErrorDetailsToProto(errApplicationDownlinkTooLong),
					},
				},
			})
			if err != nil {
				log.FromContext(ctx).WithError(err).Warn("Failed to inform Application Server that application downlink is too long")
			} else if !ok {
				log.FromContext(ctx).Warn("Application Server not found")
			}
			if !needsDownlink && len(cmdBuf) == 0 {
				return nil, errNoDownlink
			}
		} else if down.FCnt <= dev.Session.LastNFCntDown && dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
			logger.Warn("Application downlink FCnt is too low, inform Application Server")
			ok, err := ns.handleASUplink(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers, &ttnpb.ApplicationUp{
				EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
				CorrelationIDs:       events.CorrelationIDsFromContext(ctx),
				Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
					DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{
						Downlinks:    dev.QueuedApplicationDownlinks,
						LastFCntDown: dev.Session.LastNFCntDown,
					},
				},
			})
			if err != nil {
				logger.WithError(err).Warn("Failed to send application downlink queue invalidation to Application Server")
			} else if !ok {
				logger.Warn("Application Server not found")
			}
			if !needsDownlink && len(cmdBuf) == 0 {
				return nil, errNoDownlink
			}
		} else {
			dev.QueuedApplicationDownlinks = dev.QueuedApplicationDownlinks[1:]
			appDown = down
			pld.FHDR.FCnt = down.FCnt
			pld.FPort = down.FPort
			pld.FRMPayload = down.FRMPayload
			if down.Confirmed {
				mType = ttnpb.MType_CONFIRMED_DOWN
				dev.MACState.PendingApplicationDownlink = down
				dev.Session.LastConfFCntDown = pld.FCnt
			}
		}
	}
	logger = logger.WithFields(log.Fields(
		"f_cnt", pld.FHDR.FCnt,
		"f_port", pld.FPort,
		"m_type", mType,
	))

	if len(cmdBuf) > 0 && (pld.FPort == 0 || dev.MACState.LoRaWANVersion.EncryptFOpts()) {
		if dev.Session.NwkSEncKey == nil || len(dev.Session.NwkSEncKey.Key) == 0 {
			return nil, errUnknownNwkSEncKey
		}
		key, err := cryptoutil.UnwrapAES128Key(*dev.Session.NwkSEncKey, ns.KeyVault)
		if err != nil {
			logger.WithField("kek_label", dev.Session.NwkSEncKey.KEKLabel).WithError(err).Warn("Failed to unwrap NwkSEncKey")
			return nil, err
		}

		cmdBuf, err = crypto.EncryptDownlink(key, dev.Session.DevAddr, pld.FHDR.FCnt, cmdBuf)
		if err != nil {
			return nil, errEncryptMAC.WithCause(err)
		}
	}

	if pld.FPort == 0 {
		pld.FRMPayload = cmdBuf
		dev.Session.LastNFCntDown = pld.FCnt
	} else {
		pld.FHDR.FOpts = cmdBuf
	}

	if dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 && pld.FCnt > dev.Session.LastNFCntDown {
		dev.Session.LastNFCntDown = pld.FCnt
	}

	pld.FHDR.FCtrl.FPending = fPending || len(dev.QueuedApplicationDownlinks) > 0
	logger = logger.WithField("f_pending", pld.FHDR.FCtrl.FPending)

	switch {
	case mType != ttnpb.MType_CONFIRMED_DOWN && len(dev.MACState.PendingRequests) == 0:
		break

	case dev.MACState.DeviceClass == ttnpb.CLASS_C &&
		dev.MACState.LastConfirmedDownlinkAt != nil &&
		dev.MACState.LastConfirmedDownlinkAt.Add(deviceClassCTimeout(dev, ns.defaultMACSettings)).After(time.Now()):
		return nil, errScheduleTooSoon

	default:
		dev.MACState.LastConfirmedDownlinkAt = timePtr(time.Now().UTC())
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
		return nil, errEncodePayload.WithCause(err)
	}
	// NOTE: It is assumed, that b does not contain MIC.

	if dev.Session.SNwkSIntKey == nil || len(dev.Session.SNwkSIntKey.Key) == 0 {
		return nil, errUnknownSNwkSIntKey
	}
	key, err := cryptoutil.UnwrapAES128Key(*dev.Session.SNwkSIntKey, ns.KeyVault)
	if err != nil {
		logger.WithField("kek_label", dev.Session.SNwkSIntKey.KEKLabel).WithError(err).Warn("Failed to unwrap SNwkSIntKey")
		return nil, err
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
		return nil, errComputeMIC
	}
	b = append(b, mic[:]...)

	logger.WithField("payload_length", len(b)).Debug("Generated downlink")
	return &generatedDownlink{
		Payload:             b,
		FCnt:                pld.FHDR.FCnt,
		ApplicationDownlink: appDown,
	}, nil
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
		delta := time.Now().Sub(up.ReceivedAt)
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

// scheduleDownlinkByPaths attempts to schedule payload b using parameters in req for devID using paths.
// scheduleDownlinkByPaths discards req.DownlinkPaths and mutates it arbitrarily.
// scheduleDownlinkByPaths returns the scheduled downlink or error.
func (ns *NetworkServer) scheduleDownlinkByPaths(ctx context.Context, req *ttnpb.TxRequest, devID ttnpb.EndDeviceIdentifiers, b []byte, paths ...downlinkPath) (*ttnpb.DownlinkMessage, time.Time, error) {
	if len(paths) == 0 {
		return nil, time.Time{}, errNoPath
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

		p := ns.GetPeer(ctx, ttnpb.PeerInfo_GATEWAY_SERVER, path.GatewayIdentifiers)
		if p == nil {
			logger.Debug("Could not get Gateway Server")
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
			EndDeviceIDs:   &devID,
			CorrelationIDs: events.CorrelationIDsFromContext(ctx),
			Settings: &ttnpb.DownlinkMessage_Request{
				Request: req,
			},
		}

		logger.WithField("path_count", len(req.DownlinkPaths)).Debug("Schedule downlink")
		res, err := ttnpb.NewNsGsClient(a.peer.Conn()).ScheduleDownlink(ctx, down, ns.WithClusterAuth())
		if err != nil {
			errs = append(errs, err)
			continue
		}
		logger.WithField("delay", res.Delay).Debug("Scheduled downlink")
		return down, time.Now().Add(res.Delay), nil
	}

	for i, err := range errs {
		logger = logger.WithField(
			fmt.Sprintf("error_%d", i), err,
		)
	}
	logger.Warn("All Gateway Servers failed to schedule downlink")
	return nil, time.Time{}, errSchedule
}

func appendRecentDownlink(recent []*ttnpb.DownlinkMessage, down *ttnpb.DownlinkMessage, window int) []*ttnpb.DownlinkMessage {
	recent = append(recent, down)
	if len(recent) > window {
		recent = recent[len(recent)-window:]
	}
	return recent
}

// processDownlinkTask processes the most recent downlink task ready for execution, if such is available or wait until it is before processing it.
// NOTE: ctx.Done() is not guaranteed to be respected by processDownlinkTask.
func (ns *NetworkServer) processDownlinkTask(ctx context.Context) error {
	var scheduleErr bool
	var setErr bool
	var addErr bool
	err := ns.downlinkTasks.Pop(ctx, func(ctx context.Context, devID ttnpb.EndDeviceIdentifiers, t time.Time) error {
		logger := log.FromContext(ctx).WithFields(log.Fields(
			"device_uid", unique.ID(ctx, devID),
			"started_at", time.Now().UTC(),
		))
		ctx = log.NewContext(ctx, logger)
		logger.WithField("start_at", t).Debug("Process downlink task")

		var genDown *generatedDownlink
		var nextDownlinkAt time.Time
		dev, err := ns.devices.SetByID(ctx, devID.ApplicationIdentifiers, devID.DeviceID,
			[]string{
				"frequency_plan_id",
				"last_dev_status_received_at",
				"lorawan_phy_version",
				"mac_settings",
				"mac_state",
				"queued_application_downlinks",
				"recent_downlinks",
				"recent_uplinks",
				"session",
			},
			func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
				switch {
				case dev == nil:
					return nil, nil, errDeviceNotFound

				case dev.MACState == nil:
					return nil, nil, errUnknownMACState

				case dev.MACState.DeviceClass == ttnpb.CLASS_A && !dev.MACState.RxWindowsAvailable:
					return dev, nil, nil

				case dev.MACState.DeviceClass == ttnpb.CLASS_B && !dev.MACState.RxWindowsAvailable:
					// TODO: Support Class B (https://github.com/TheThingsNetwork/lorawan-stack/issues/19).
					logger.Warn("Class B downlink task scheduled, but Rx windows are not available")
					return dev, nil, nil
				}
				logger = logger.WithField("device_class", dev.MACState.DeviceClass)

				fp, band, err := getDeviceBandVersion(dev, ns.FrequencyPlans)
				if err != nil {
					return nil, nil, errUnknownBand.WithCause(err)
				}

				// NOTE: If no data uplink is found, we assume ADR is off on the device and, hence, data rate index 0 is used in computation.
				maxUpDRIdx := ttnpb.DATA_RATE_0
			loop:
				for i := len(dev.RecentUplinks) - 1; i >= 0; i-- {
					switch dev.RecentUplinks[i].Payload.MHDR.MType {
					case ttnpb.MType_JOIN_REQUEST:
						break loop
					case ttnpb.MType_UNCONFIRMED_UP, ttnpb.MType_CONFIRMED_UP:
						if dev.RecentUplinks[i].Payload.GetMACPayload().FHDR.FCtrl.ADR {
							maxUpDRIdx = dev.RecentUplinks[i].Settings.DataRateIndex
						}
						break loop
					}
				}
				maxUpLength := band.DataRates[maxUpDRIdx].DefaultMaxSize.PayloadSize(fp.DwellTime.GetUplinks())

				if dev.MACState.RxWindowsAvailable {
					if len(dev.RecentUplinks) == 0 {
						return nil, nil, errUplinkNotFound
					}
					up := dev.RecentUplinks[len(dev.RecentUplinks)-1]
					ctx = events.ContextWithCorrelationID(ctx, up.CorrelationIDs...)

					if up.DeviceChannelIndex > math.MaxUint8 {
						return nil, nil, errInvalidChannelIndex
					}
					rx1ChIdx, err := band.Rx1Channel(uint8(up.DeviceChannelIndex))
					if err != nil {
						return nil, nil, err
					}
					if uint(rx1ChIdx) >= uint(len(dev.MACState.CurrentParameters.Channels)) ||
						dev.MACState.CurrentParameters.Channels[int(rx1ChIdx)] == nil ||
						dev.MACState.CurrentParameters.Channels[int(rx1ChIdx)].DownlinkFrequency == 0 {
						return nil, nil, errCorruptedMACState
					}
					rx1DRIdx, err := band.Rx1DataRate(up.Settings.DataRateIndex, dev.MACState.CurrentParameters.Rx1DataRateOffset, dev.MACState.CurrentParameters.DownlinkDwellTime)
					if err != nil {
						return nil, nil, err
					}
					rx1Freq := dev.MACState.CurrentParameters.Channels[int(rx1ChIdx)].DownlinkFrequency

					req := &ttnpb.TxRequest{
						Class: ttnpb.CLASS_A,
					}
					switch {
					case dev.MACState.QueuedJoinAccept != nil:
						// Join-accept downlink for Class A/B/C in Rx1/Rx2
						req.Rx1Delay = ttnpb.RxDelay(band.JoinAcceptDelay1 / time.Second)
						rx1, rx2, paths := downlinkPathsForClassA(
							ttnpb.RxDelay(band.JoinAcceptDelay1/time.Second),
							dev.RecentUplinks...,
						)
						if rx1 {
							req.Rx1Frequency = rx1Freq
							req.Rx1DataRateIndex = rx1DRIdx
						}
						if rx2 {
							req.Rx2Frequency = dev.MACState.CurrentParameters.Rx2Frequency
							req.Rx2DataRateIndex = dev.MACState.CurrentParameters.Rx2DataRateIndex
						}
						if !rx1 && !rx2 {
							return nil, nil, errNoPath
						}

						down, _, err := ns.scheduleDownlinkByPaths(
							log.NewContext(ctx, logger.WithFields(log.Fields(
								"attempt_rx1", rx1,
								"attempt_rx2", rx2,
								"downlink_class", req.Class,
								"downlink_type", "join-accept",
								"rx1_delay", req.Rx1Delay,
								"rx1_frequency", req.Rx1Frequency,
								"rx2_data_rate", req.Rx2DataRateIndex,
								"rx2_frequency", req.Rx2Frequency,
							))),
							req,
							dev.EndDeviceIdentifiers,
							dev.MACState.QueuedJoinAccept.Payload,
							paths...,
						)
						if err != nil {
							scheduleErr = true
						} else {
							dev.MACState.RxWindowsAvailable = false
							dev.MACState.PendingJoinRequest = &dev.MACState.QueuedJoinAccept.Request
							dev.PendingSession = &ttnpb.Session{
								DevAddr:     dev.MACState.QueuedJoinAccept.Request.DevAddr,
								SessionKeys: dev.MACState.QueuedJoinAccept.Keys,
							}
							dev.MACState.QueuedJoinAccept = nil
							dev.RecentDownlinks = appendRecentDownlink(dev.RecentDownlinks, down, recentDownlinkCount)
							return dev, []string{
								"ids.dev_addr",
								"mac_state.pending_join_request",
								"mac_state.queued_join_accept",
								"mac_state.rx_windows_available",
								"pending_session",
								"recent_downlinks",
							}, nil
						}

					case dev.MACState.DeviceClass == ttnpb.CLASS_A:
						// Data downlink for Class A in Rx1/Rx2
						req.Rx1Delay = dev.MACState.CurrentParameters.Rx1Delay
						rx1, rx2, paths := downlinkPathsForClassA(
							dev.MACState.CurrentParameters.Rx1Delay,
							dev.RecentUplinks...,
						)
						if rx1 {
							req.Rx1Frequency = rx1Freq
							req.Rx1DataRateIndex = rx1DRIdx
						}
						if rx2 {
							req.Rx2Frequency = dev.MACState.CurrentParameters.Rx2Frequency
							req.Rx2DataRateIndex = dev.MACState.CurrentParameters.Rx2DataRateIndex
						}
						if !rx1 && !rx2 {
							return nil, nil, errNoPath
						}

						minDR := req.Rx1DataRateIndex
						if req.Rx2DataRateIndex < minDR {
							minDR = req.Rx2DataRateIndex
						}
						genDown, err = ns.generateDownlink(ctx, dev,
							band.DataRates[minDR].DefaultMaxSize.PayloadSize(fp.DwellTime.GetDownlinks()),
							maxUpLength,
						)
						if err != nil {
							return nil, nil, err
						}
						if genDown.ApplicationDownlink != nil {
							ctx = events.ContextWithCorrelationID(ctx, genDown.ApplicationDownlink.CorrelationIDs...)
						}

						down, _, err := ns.scheduleDownlinkByPaths(
							log.NewContext(ctx, logger.WithFields(log.Fields(
								"attempt_rx1", rx1,
								"attempt_rx2", rx2,
								"downlink_class", req.Class,
								"downlink_type", "data",
								"rx1_delay", req.Rx1Delay,
								"rx1_frequency", req.Rx1Frequency,
								"rx2_data_rate", req.Rx2DataRateIndex,
								"rx2_frequency", req.Rx2Frequency,
							))),
							req,
							dev.EndDeviceIdentifiers,
							genDown.Payload,
							paths...,
						)
						if err != nil {
							scheduleErr = true
						} else {
							dev.MACState.RxWindowsAvailable = false
							dev.RecentDownlinks = appendRecentDownlink(dev.RecentDownlinks, down, recentDownlinkCount)
							return dev, []string{
								"mac_state",
								"queued_application_downlinks",
								"recent_downlinks",
								"session",
							}, nil
						}

					default:
						// Data downlink for Class B/C in Rx1 if available
						req.Rx1Delay = dev.MACState.CurrentParameters.Rx1Delay
						rx1, _, paths := downlinkPathsForClassA(
							dev.MACState.CurrentParameters.Rx1Delay,
							dev.RecentUplinks...,
						)
						if rx1 {
							req.Rx1Frequency = rx1Freq
							req.Rx1DataRateIndex = rx1DRIdx
						} else {
							break
						}

						// NOTE: generateDownlink mutates the device, and since we may need to call it twice(Rx1/Rx2),
						// we need to create a deep copy for the first call.
						devCopy := deepcopy.Copy(dev).(*ttnpb.EndDevice)

						genDown, err = ns.generateDownlink(ctx, dev,
							band.DataRates[req.Rx1DataRateIndex].DefaultMaxSize.PayloadSize(fp.DwellTime.GetDownlinks()),
							maxUpLength,
						)
						if err != nil {
							if errors.Resemble(err, errScheduleTooSoon) {
								nextDownlinkAt = dev.MACState.LastConfirmedDownlinkAt.Add(deviceClassCTimeout(dev, ns.defaultMACSettings))
							}
							return nil, nil, err
						}
						if dev.MACState.DeviceClass == ttnpb.CLASS_C {
							if dev.MACState.LastConfirmedDownlinkAt != nil {
								nextDownlinkAt = dev.MACState.LastConfirmedDownlinkAt.Add(deviceClassCTimeout(dev, ns.defaultMACSettings))
							} else {
								nextDownlinkAt = time.Now()
							}
						}

						if genDown.ApplicationDownlink != nil {
							if len(genDown.ApplicationDownlink.ClassBC.GetGateways()) > 0 ||
								genDown.ApplicationDownlink.ClassBC.GetAbsoluteTime() != nil {
								// Skip Rx1 when a fixed path or an absolute tranmission time is requested by the application.
								// Gateway Server cannot schedule Rx1 on a fixed path as there is no uplink token.
								// Also, it is highly unlikely and not verifiable by Network Server that Rx1 is at ClassBC.AbsoluteTime.
								paths = nil
							} else {
								ctx = events.ContextWithCorrelationID(ctx, genDown.ApplicationDownlink.CorrelationIDs...)
							}
						}

						if len(paths) > 0 {
							down, downAt, err := ns.scheduleDownlinkByPaths(
								log.NewContext(ctx, logger.WithFields(log.Fields(
									"attempt_rx1", true,
									"attempt_rx2", false,
									"downlink_class", req.Class,
									"downlink_type", "data",
									"rx1_delay", req.Rx1Delay,
									"rx1_frequency", req.Rx1Frequency,
								))),
								req,
								dev.EndDeviceIdentifiers,
								genDown.Payload,
								paths...,
							)
							if err != nil {
								dev = devCopy
								scheduleErr = true
							} else {
								if dev.MACState.DeviceClass == ttnpb.CLASS_C && nextDownlinkAt.Before(downAt) {
									nextDownlinkAt = downAt
								}
								dev.MACState.RxWindowsAvailable = false
								dev.RecentDownlinks = appendRecentDownlink(dev.RecentDownlinks, down, recentDownlinkCount)
								return dev, []string{
									"mac_state",
									"queued_application_downlinks",
									"recent_downlinks",
									"session",
								}, nil
							}
						}
						// Rx1 did not get scheduled, so we restore the original state of the device.
						dev = devCopy
					}
				}
				if dev.MACState.QueuedJoinAccept != nil {
					return nil, nil, errSchedule
				}

				dev.MACState.RxWindowsAvailable = false
				if dev.MACState.DeviceClass == ttnpb.CLASS_B || dev.MACState.DeviceClass == ttnpb.CLASS_C {
					// Data downlink for Class B/C in Rx2
					req := &ttnpb.TxRequest{
						Class:            dev.MACState.DeviceClass,
						Rx2DataRateIndex: dev.MACState.CurrentParameters.Rx2DataRateIndex,
						Rx2Frequency:     dev.MACState.CurrentParameters.Rx2Frequency,
					}

					genDown, err = ns.generateDownlink(ctx, dev,
						band.DataRates[req.Rx2DataRateIndex].DefaultMaxSize.PayloadSize(fp.DwellTime.GetDownlinks()),
						maxUpLength,
					)
					if err != nil {
						if errors.Resemble(err, errScheduleTooSoon) {
							nextDownlinkAt = dev.MACState.LastConfirmedDownlinkAt.Add(deviceClassCTimeout(dev, ns.defaultMACSettings))
						}
						return nil, nil, err
					}
					if dev.MACState.DeviceClass == ttnpb.CLASS_C {
						if dev.MACState.LastConfirmedDownlinkAt != nil {
							nextDownlinkAt = dev.MACState.LastConfirmedDownlinkAt.Add(deviceClassCTimeout(dev, ns.defaultMACSettings))
						} else {
							nextDownlinkAt = time.Now()
						}
					}

					if genDown.ApplicationDownlink != nil {
						ctx = events.ContextWithCorrelationID(ctx, genDown.ApplicationDownlink.CorrelationIDs...)
					}

					var paths []downlinkPath
					if genDown.ApplicationDownlink != nil && genDown.ApplicationDownlink.ClassBC != nil {
						paths = make([]downlinkPath, 0, len(genDown.ApplicationDownlink.ClassBC.Gateways))
						for _, gtw := range genDown.ApplicationDownlink.ClassBC.Gateways {
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
						if len(dev.RecentUplinks) == 0 {
							return nil, nil, errUplinkNotFound
						}
						up := dev.RecentUplinks[len(dev.RecentUplinks)-1]
						ctx = events.ContextWithCorrelationID(ctx, up.CorrelationIDs...)
						paths = downlinkPathsFromRecentUplinks(dev.RecentUplinks...)
					}

					down, downAt, err := ns.scheduleDownlinkByPaths(
						log.NewContext(ctx, logger.WithFields(log.Fields(
							"downlink_type", "data",
							"attempt_rx1", false,
							"attempt_rx2", true,
						))),
						req,
						dev.EndDeviceIdentifiers,
						genDown.Payload,
						paths...,
					)
					if err != nil {
						scheduleErr = true
					} else {
						if dev.MACState.DeviceClass == ttnpb.CLASS_C && nextDownlinkAt.Before(downAt) {
							nextDownlinkAt = downAt
						}
						dev.RecentDownlinks = appendRecentDownlink(dev.RecentDownlinks, down, recentDownlinkCount)
						return dev, []string{
							"mac_state",
							"queued_application_downlinks",
							"recent_downlinks",
							"session",
						}, nil
					}
				}
				return nil, nil, errSchedule
			},
		)

		switch {
		case scheduleErr:
			return err

		case err != nil && errors.Resemble(err, errNoDownlink):
			logger.Debug("No downlink to send, skip downlink slot")
			return nil
		}

		if err != nil && errors.Resemble(err, errScheduleTooSoon) {
			logger.Debug("Downlink scheduled too soon, skip downlink slot")
		} else if err != nil {
			setErr = true
			logger.WithError(err).Warn("Failed to update device in registry")
			return err
		}

		if genDown != nil && dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
			go func() {
				logger.Debug("Send downlink queue invalidation to Application Server")
				ok, err := ns.handleASUplink(ctx, dev.EndDeviceIdentifiers.ApplicationIdentifiers, &ttnpb.ApplicationUp{
					EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
					CorrelationIDs:       events.CorrelationIDsFromContext(ctx),
					Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
						DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{
							Downlinks:    dev.QueuedApplicationDownlinks,
							LastFCntDown: genDown.FCnt,
						},
					},
				})
				if err != nil {
					logger.WithError(err).Warn("Failed to send application downlink queue invalidation to Application Server")
				} else if !ok {
					logger.Warn("Application Server not found")
				}
			}()
		}

		if !nextDownlinkAt.IsZero() {
			logger.WithField("start_at", nextDownlinkAt.UTC()).Debug("Add downlink task after downlink")
			if err := ns.downlinkTasks.Add(ctx, devID, nextDownlinkAt, true); err != nil {
				addErr = true
				logger.WithError(err).Error("Failed to add downlink task after downlink")
				return err
			}
		}
		return nil
	})
	if err != nil && !setErr && !addErr && !scheduleErr {
		ns.Logger().WithError(err).Warn("Failed to pop device from downlink schedule")
	}
	return err
}
