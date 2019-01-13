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
	"sort"
	"time"

	"github.com/mohae/deepcopy"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

// DownlinkTaskQueue represents an entity, that holds downlink tasks sorted by timestamp.
type DownlinkTaskQueue interface {
	// Add adds downlink task for device identified by devID at time t.
	// Implementations must ensure that Add returns fast.
	Add(ctx context.Context, devID ttnpb.EndDeviceIdentifiers, t time.Time) error

	// Pop calls f on the most recent downlink task in the schedule, for which timestamp is in range [0, time.Now()],
	// if such is available, otherwise it blocks until it is.
	// Context passed to f must be derived from ctx.
	// Implementations must respect ctx.Deadline() value on best-effort basis, if such is present.
	Pop(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error
}

var errNoDownlink = errors.Define("no_downlink", "no downlink to send")

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
func generateDownlink(ctx context.Context, dev *ttnpb.EndDevice, maxDownLen, maxUpLen uint16, fps *frequencyplans.Store) ([]byte, *ttnpb.ApplicationDownlink, error) {
	if dev.MACState == nil {
		return nil, nil, errUnknownMACState
	}

	if dev.Session == nil {
		return nil, nil, errEmptySession
	}

	// NOTE: len(MHDR) + len(MIC) = 1 + 4 = 5
	if maxDownLen < 5 || maxUpLen < 5 {
		panic("payload length limits too short to generate downlink")
	}
	maxDownLen, maxUpLen = maxDownLen-5, maxUpLen-5

	logger := log.FromContext(ctx).WithField("device_uid", unique.ID(ctx, dev.EndDeviceIdentifiers))

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

	maxDownLen, maxUpLen, ok, err := enqueueLinkADRReq(ctx, dev, maxDownLen, maxUpLen, fps)
	if err != nil {
		return nil, nil, err
	}
	fPending := !ok
	for _, f := range []func(context.Context, *ttnpb.EndDevice, uint16, uint16) (uint16, uint16, bool){
		// LoRaWAN 1.0+
		enqueueNewChannelReq,
		enqueueDutyCycleReq,
		enqueueRxParamSetupReq,
		enqueueDevStatusReq,
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

	cmdBuf := make([]byte, 0, maxDownLen)
	for _, cmd := range cmds {
		var err error
		cmdBuf, err = spec.AppendDownlink(cmdBuf, *cmd)

		if err != nil {
			return nil, nil, errEncodeMAC.WithCause(err)
		}
	}
	if len(cmdBuf) > 0 {
		logger.WithField("buf", cmdBuf).Debug("Generated MAC command buffer")
	}

	var needsDownlink bool
	var up *ttnpb.UplinkMessage
	for i := len(dev.RecentUplinks) - 1; i >= 0; i-- {
		switch dev.RecentUplinks[i].Payload.MHDR.MType {
		case ttnpb.MType_UNCONFIRMED_UP:
			up = dev.RecentUplinks[i]
			needsDownlink = up.Payload.GetMACPayload().FCtrl.ADRAckReq
			break
		case ttnpb.MType_CONFIRMED_UP:
			up = dev.RecentUplinks[i]
			needsDownlink = true
			break
		case ttnpb.MType_JOIN_REQUEST, ttnpb.MType_REJOIN_REQUEST:
			up = dev.RecentUplinks[i]
			break
		default:
			logger.WithField("m_type", up.Payload.MHDR.MType).Warn("Unknown MType stored in RecentUplinks")
		}
	}
	if !needsDownlink &&
		len(cmdBuf) == 0 &&
		len(dev.QueuedApplicationDownlinks) == 0 {
		return nil, nil, errNoDownlink
	}

	pld := &ttnpb.MACPayload{
		FHDR: ttnpb.FHDR{
			DevAddr: *dev.EndDeviceIdentifiers.DevAddr,
			FCtrl: ttnpb.FCtrl{
				Ack: up != nil && up.Payload.MHDR.MType == ttnpb.MType_CONFIRMED_UP,
			},
			FCnt: dev.Session.LastNFCntDown + 1,
		},
	}

	var appDown *ttnpb.ApplicationDownlink
	mType := ttnpb.MType_UNCONFIRMED_DOWN
	if len(cmdBuf) <= fOptsCapacity && len(dev.QueuedApplicationDownlinks) > 0 {
		var down *ttnpb.ApplicationDownlink
		down, dev.QueuedApplicationDownlinks = dev.QueuedApplicationDownlinks[0], dev.QueuedApplicationDownlinks[1:]

		if len(down.FRMPayload) > int(maxDownLen) {
			// TODO: Inform AS that payload is too long(https://github.com/TheThingsIndustries/lorawan-stack/issues/377)
			if !needsDownlink && len(cmdBuf) == 0 {
				return nil, nil, errNoDownlink
			}
		} else {
			appDown = down
			pld.FHDR.FCnt = down.FCnt
			pld.FPort = down.FPort
			pld.FRMPayload = down.FRMPayload
			if down.Confirmed {
				dev.MACState.PendingApplicationDownlink = down
				dev.Session.LastConfFCntDown = pld.FCnt

				mType = ttnpb.MType_CONFIRMED_DOWN
			}
		}
	}

	if len(cmdBuf) > 0 && (pld.FPort == 0 || dev.MACState.LoRaWANVersion.EncryptFOpts()) {
		if dev.Session.NwkSEncKey == nil || len(dev.Session.NwkSEncKey.Key) == 0 {
			return nil, nil, errUnknownNwkSEncKey
		}

		var key types.AES128Key
		if dev.Session.NwkSEncKey.KEKLabel != "" {
			// TODO: (https://github.com/TheThingsIndustries/lorawan-stack/issues/271)
			panic("unsupported")
		}
		copy(key[:], dev.Session.NwkSEncKey.Key[:])

		var err error
		cmdBuf, err = crypto.EncryptDownlink(key, *dev.EndDeviceIdentifiers.DevAddr, pld.FHDR.FCnt, cmdBuf)
		if err != nil {
			return nil, nil, errEncryptMAC.WithCause(err)
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

	switch {
	case dev.MACState.DeviceClass != ttnpb.CLASS_C,
		mType != ttnpb.MType_CONFIRMED_DOWN && len(dev.MACState.PendingRequests) == 0:
		break

	case dev.MACState.LastConfirmedDownlinkAt.Add(classCTimeout).After(time.Now()):
		return nil, nil, errScheduleTooSoon

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
		return nil, nil, errEncodePayload.WithCause(err)
	}
	// NOTE: It is assumed, that b does not contain MIC.

	var key types.AES128Key
	if dev.Session.SNwkSIntKey == nil || len(dev.Session.SNwkSIntKey.Key) == 0 {
		return nil, nil, errUnknownSNwkSIntKey
	}
	if dev.Session.SNwkSIntKey.KEKLabel != "" {
		// TODO: https://github.com/TheThingsIndustries/lorawan-stack/issues/271
		panic("unsupported")
	}
	copy(key[:], dev.Session.SNwkSIntKey.Key[:])

	var mic [4]byte
	if dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
		mic, err = crypto.ComputeLegacyDownlinkMIC(
			key,
			*dev.EndDeviceIdentifiers.DevAddr,
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
			*dev.EndDeviceIdentifiers.DevAddr,
			confFCnt,
			pld.FHDR.FCnt,
			b,
		)
	}
	if err != nil {
		return nil, nil, errComputeMIC
	}
	return append(b, mic[:]...), appDown, nil
}

type downlinkPath struct {
	ttnpb.GatewayIdentifiers
	*ttnpb.DownlinkPath
}

func downlinkPathsFromMetadata(mds ...*ttnpb.RxMetadata) []downlinkPath {
	mds = append(mds[:0:0], mds...)
	sort.SliceStable(mds, func(i, j int) bool {
		// TODO: Improve the sorting algorithm (https://github.com/TheThingsIndustries/ttn/issues/729)
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
func (ns *NetworkServer) scheduleDownlinkByPaths(ctx context.Context, req *ttnpb.TxRequest, devID ttnpb.EndDeviceIdentifiers, b []byte, paths ...downlinkPath) (*ttnpb.DownlinkMessage, error) {
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

		_, err := ttnpb.NewNsGsClient(a.peer.Conn()).ScheduleDownlink(ctx, down, ns.WithClusterAuth())
		if err != nil {
			errs = append(errs, err)
			continue
		}
		return down, nil
	}

	for i, err := range errs {
		logger = logger.WithField(
			fmt.Sprintf("error_%d", i), err,
		)
	}
	logger.Warn("All Gateway Servers failed to schedule downlink")
	return nil, errSchedule
}

func (ns *NetworkServer) sendQueueInvalidationToAS(ctx context.Context, dev *ttnpb.EndDevice) (bool, error) {
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
		log.FromContext(ctx).WithError(err).Warn("Failed to send application downlink queue invalidation to Application Server")
	} else if !ok {
		log.FromContext(ctx).Warn("Application Server not found")
	}
	return ok, err
}

// processDownlinkTask processes the most recent downlink task ready for execution, if such is available or wait until it is before processing it.
// NOTE: ctx.Done() is not guaranteed to be respected by processDownlinkTask.
func (ns *NetworkServer) processDownlinkTask(ctx context.Context) error {
	var scheduleErr bool
	var setErr bool
	var addErr bool
	err := ns.downlinkTasks.Pop(ctx, func(ctx context.Context, devID ttnpb.EndDeviceIdentifiers, t time.Time) error {
		logger := log.FromContext(ctx).WithFields(log.Fields(
			"delay", time.Now().Sub(t),
			"device_uid", unique.ID(ctx, devID),
			"start_at", t,
		))
		ctx = log.NewContext(ctx, logger)
		logger.Debug("Processing downlink task...")

		dev, err := ns.devices.SetByID(ctx, devID.ApplicationIdentifiers, devID.DeviceID,
			[]string{
				"frequency_plan_id",
				"lorawan_phy_version",
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
					// TODO: Support Class B (https://github.com/TheThingsIndustries/lorawan-stack/issues/833).
					logger.Warn("Class B downlink task scheduled, but Rx windows are not available")
					return dev, nil, nil
				}
				logger = logger.WithField("device_class", dev.MACState.DeviceClass)

				fp, band, err := getDeviceBandVersion(ns.FrequencyPlans, dev)
				if err != nil {
					return nil, nil, errUnknownBand.WithCause(err)
				}

				// NOTE: If no data uplink is found, we assume ADR is off on the device and, hence, data rate index 0 is used in computation.
				maxUpLength := band.DataRates[0].DefaultMaxSize.PayloadSize(fp.DwellTime.GetUplinks())
				for i := len(dev.RecentUplinks) - 1; i >= 0; i-- {
					switch dev.RecentUplinks[i].Payload.MHDR.MType {
					case ttnpb.MType_JOIN_REQUEST:
						break

					case ttnpb.MType_UNCONFIRMED_UP, ttnpb.MType_CONFIRMED_UP:
						if dev.RecentUplinks[i].Payload.GetMACPayload().FHDR.FCtrl.ADR {
							maxUpLength = band.DataRates[dev.RecentUplinks[i].Settings.DataRateIndex].DefaultMaxSize.PayloadSize(fp.DwellTime.GetUplinks())
						}
						break
					}
				}

				if dev.MACState.RxWindowsAvailable {
					if len(dev.RecentUplinks) == 0 {
						return nil, nil, errUplinkNotFound
					}
					up := dev.RecentUplinks[len(dev.RecentUplinks)-1]
					ctx = events.ContextWithCorrelationID(ctx, up.CorrelationIDs...)

					rx1ChIdx, err := band.Rx1Channel(up.Settings.ChannelIndex)
					if err != nil {
						return nil, nil, err
					}
					if uint(rx1ChIdx) >= uint(len(dev.MACState.CurrentParameters.Channels)) ||
						dev.MACState.CurrentParameters.Channels[int(rx1ChIdx)] == nil ||
						dev.MACState.CurrentParameters.Channels[int(rx1ChIdx)].DownlinkFrequency == 0 {
						return nil, nil, errCorruptedMACState
					}

					req := &ttnpb.TxRequest{
						Class:        ttnpb.CLASS_A,
						Rx1Delay:     dev.MACState.CurrentParameters.Rx1Delay,
						Rx1Frequency: dev.MACState.CurrentParameters.Channels[int(rx1ChIdx)].DownlinkFrequency,
					}
					req.Rx1DataRateIndex, err = band.Rx1DataRate(up.Settings.DataRateIndex, dev.MACState.CurrentParameters.Rx1DataRateOffset, dev.MACState.CurrentParameters.DownlinkDwellTime)
					if err != nil {
						return nil, nil, err
					}

					switch {
					case len(dev.MACState.QueuedJoinAccept) > 0:
						// Join-accept downlink for Class A/B/C in Rx1/Rx2
						req.Rx1Delay = ttnpb.RxDelay(band.JoinAcceptDelay1 / time.Second)
						req.Rx2DataRateIndex = dev.MACState.CurrentParameters.Rx2DataRateIndex
						req.Rx2Frequency = dev.MACState.CurrentParameters.Rx2Frequency

						down, err := ns.scheduleDownlinkByPaths(
							log.NewContext(ctx, logger.WithFields(log.Fields(
								"downlink_type", "join-accept",
								"attempt_rx1", true,
								"attempt_rx2", true,
							))),
							req,
							dev.EndDeviceIdentifiers,
							dev.MACState.QueuedJoinAccept,
							downlinkPathsFromRecentUplinks(dev.RecentUplinks...)...,
						)
						if err != nil {
							scheduleErr = true
						} else {
							dev.MACState.QueuedJoinAccept = nil
							dev.MACState.RxWindowsAvailable = false
							dev.RecentDownlinks = append(dev.RecentDownlinks, down)
							if len(dev.RecentDownlinks) > recentDownlinkCount {
								dev.RecentDownlinks = append(dev.RecentDownlinks[:0], dev.RecentDownlinks[len(dev.RecentDownlinks)-recentDownlinkCount:]...)
							}
							return dev, []string{
								"mac_state.queued_join_accept",
								"mac_state.rx_windows_available",
								"recent_downlinks",
							}, nil
						}

					case dev.MACState.DeviceClass == ttnpb.CLASS_A:
						// Data downlink for Class A in Rx1/Rx2
						req.Rx2DataRateIndex = dev.MACState.CurrentParameters.Rx2DataRateIndex
						req.Rx2Frequency = dev.MACState.CurrentParameters.Rx2Frequency
						minDR := req.Rx1DataRateIndex
						if req.Rx2DataRateIndex < minDR {
							minDR = req.Rx2DataRateIndex
						}
						b, appDown, err := generateDownlink(ctx, dev,
							band.DataRates[minDR].DefaultMaxSize.PayloadSize(fp.DwellTime.GetDownlinks()),
							maxUpLength,
							ns.FrequencyPlans,
						)
						if err != nil {
							return nil, nil, err
						}
						ctx = events.ContextWithCorrelationID(ctx, appDown.CorrelationIDs...)

						down, err := ns.scheduleDownlinkByPaths(
							log.NewContext(ctx, logger.WithFields(log.Fields(
								"downlink_type", "data",
								"attempt_rx1", true,
								"attempt_rx2", true,
							))),
							req,
							dev.EndDeviceIdentifiers,
							b,
							downlinkPathsFromRecentUplinks(dev.RecentUplinks...)...,
						)
						if err != nil {
							scheduleErr = true
						} else {
							if appDown == nil && len(dev.QueuedApplicationDownlinks) > 0 && dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
								go ns.sendQueueInvalidationToAS(ctx, dev)
							}
							dev.MACState.RxWindowsAvailable = false
							dev.RecentDownlinks = append(dev.RecentDownlinks, down)
							if len(dev.RecentDownlinks) > recentDownlinkCount {
								dev.RecentDownlinks = append(dev.RecentDownlinks[:0], dev.RecentDownlinks[len(dev.RecentDownlinks)-recentDownlinkCount:]...)
							}
							return dev, []string{
								"mac_state",
								"queued_application_downlinks",
								"recent_downlinks",
								"session",
							}, nil
						}

					default:
						// Data downlink for Class B/C in Rx1

						// NOTE: generateDownlink mutates the device, and since we may need to call it twice(Rx1/Rx2),
						// we need to create a deep copy for the first call.
						devCopy := deepcopy.Copy(dev).(*ttnpb.EndDevice)

						b, appDown, err := generateDownlink(ctx, dev,
							band.DataRates[req.Rx1DataRateIndex].DefaultMaxSize.PayloadSize(fp.DwellTime.GetDownlinks()),
							maxUpLength,
							ns.FrequencyPlans,
						)
						if err != nil {
							return nil, nil, err
						}
						ctx = events.ContextWithCorrelationID(ctx, appDown.CorrelationIDs...)

						var paths []downlinkPath
						if appDown.ClassBC != nil && appDown.ClassBC.AbsoluteTime == nil {
							paths = make([]downlinkPath, 0, len(appDown.ClassBC.Gateways))
							for _, gtw := range appDown.ClassBC.Gateways {
								if gtw == nil || gtw.IsZero() {
									continue
								}
								paths = append(paths, downlinkPath{
									GatewayIdentifiers: gtw.GatewayIdentifiers,
									DownlinkPath: &ttnpb.DownlinkPath{
										Path: &ttnpb.DownlinkPath_Fixed{
											Fixed: gtw,
										},
									},
								})
							}
						} else if appDown.ClassBC == nil {
							paths = downlinkPathsFromRecentUplinks(dev.RecentUplinks...)
						}

						if len(paths) > 0 {
							down, err := ns.scheduleDownlinkByPaths(
								log.NewContext(ctx, logger.WithFields(log.Fields(
									"downlink_type", "data",
									"attempt_rx1", true,
									"attempt_rx2", false,
								))),
								req,
								dev.EndDeviceIdentifiers,
								b,
								paths...,
							)
							if err != nil {
								dev = devCopy
								scheduleErr = true
							} else {
								if appDown == nil && len(dev.QueuedApplicationDownlinks) > 0 && dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
									go ns.sendQueueInvalidationToAS(ctx, dev)
								}
								dev.MACState.RxWindowsAvailable = false
								dev.RecentDownlinks = append(dev.RecentDownlinks, down)
								if len(dev.RecentDownlinks) > recentDownlinkCount {
									dev.RecentDownlinks = append(dev.RecentDownlinks[:0], dev.RecentDownlinks[len(dev.RecentDownlinks)-recentDownlinkCount:]...)
								}
								return dev, []string{
									"mac_state",
									"queued_application_downlinks",
									"recent_downlinks",
									"session",
								}, nil
							}
						}
					}
				}
				if len(dev.MACState.QueuedJoinAccept) > 0 {
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

					b, appDown, err := generateDownlink(ctx, dev,
						band.DataRates[req.Rx2DataRateIndex].DefaultMaxSize.PayloadSize(fp.DwellTime.GetDownlinks()),
						maxUpLength,
						ns.FrequencyPlans,
					)
					if err != nil {
						return nil, nil, err
					}
					ctx = events.ContextWithCorrelationID(ctx, appDown.CorrelationIDs...)

					var paths []downlinkPath
					if appDown.ClassBC != nil {
						paths = make([]downlinkPath, 0, len(appDown.ClassBC.Gateways))
						for _, gtw := range appDown.ClassBC.Gateways {
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

					down, err := ns.scheduleDownlinkByPaths(
						log.NewContext(ctx, logger.WithFields(log.Fields(
							"downlink_type", "data",
							"attempt_rx1", false,
							"attempt_rx2", true,
						))),
						req,
						dev.EndDeviceIdentifiers,
						b,
						paths...,
					)
					if err != nil {
						scheduleErr = true
					} else {
						if appDown == nil && len(dev.QueuedApplicationDownlinks) > 0 && dev.MACState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
							go ns.sendQueueInvalidationToAS(ctx, dev)
						}
						dev.RecentDownlinks = append(dev.RecentDownlinks, down)
						if len(dev.RecentDownlinks) > recentDownlinkCount {
							dev.RecentDownlinks = append(dev.RecentDownlinks[:0], dev.RecentDownlinks[len(dev.RecentDownlinks)-recentDownlinkCount:]...)
						}
						return dev, []string{
							"mac_state",
							"queued_application_downlinks",
							"recent_downlinks",
							"session",
						}, nil
					}
				}
				return nil, nil, errSchedule
			})

		switch {
		case scheduleErr:
			return err

		case err != nil && errors.Resemble(err, errNoDownlink):
			return nil

		case err != nil:
			setErr = true
			logger.WithError(err).Error("Failed to update device in registry")
			return err
		}

		if dev.GetMACState().GetDeviceClass() != ttnpb.CLASS_C {
			return nil
		}

		if err := ns.downlinkTasks.Add(ctx, devID, time.Now()); err != nil {
			addErr = true
			logger.WithError(err).Error("Failed to add class C device to downlink schedule")
			return err
		}
		return nil
	})
	if err != nil && !setErr && !addErr && !scheduleErr {
		ns.Logger().WithError(err).Error("Failed to pop device from downlink schedule")
	}
	return err
}
