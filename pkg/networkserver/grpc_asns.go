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
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

type ApplicationUplinkQueueDrainFunc func(limit int, f func(...*ttnpb.ApplicationUp) error) error

type ApplicationUplinkQueue interface {
	// Add adds application uplinks ups to queue.
	// Implementations must ensure that Add returns fast.
	Add(ctx context.Context, ups ...*ttnpb.ApplicationUp) error

	// PopApplication calls f on the most recent application uplink task in the schedule, for which timestamp is in range [0, time.Now()],
	// if such is available, otherwise it blocks until it is.
	// Context passed to f must be derived from ctx.
	// Implementations must respect ctx.Done() value on best-effort basis.
	Pop(ctx context.Context, consumerID string, f func(context.Context, ttnpb.ApplicationIdentifiers, ApplicationUplinkQueueDrainFunc) (time.Time, error)) error
}

func applicationJoinAcceptWithoutAppSKey(pld *ttnpb.ApplicationJoinAccept) *ttnpb.ApplicationJoinAccept {
	return &ttnpb.ApplicationJoinAccept{
		SessionKeyId:         pld.SessionKeyId,
		InvalidatedDownlinks: pld.InvalidatedDownlinks,
		PendingSession:       pld.PendingSession,
		ReceivedAt:           pld.ReceivedAt,
	}
}

const (
	applicationUplinkTaskRetryInterval = time.Minute
	applicationUplinkLimit             = 100
)

func (ns *NetworkServer) sendApplicationUplinks(ctx context.Context, cl ttnpb.NsAsClient, ups ...*ttnpb.ApplicationUp) error {
	if _, err := cl.HandleUplink(ctx, &ttnpb.NsAsHandleUplinkRequest{
		ApplicationUps: ups,
	}, ns.WithClusterAuth()); err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to send application uplinks")
		return err
	}
	for _, up := range ups {
		ctx := events.ContextWithCorrelationID(ctx, up.CorrelationIds...)
		switch pld := up.Up.(type) {
		case *ttnpb.ApplicationUp_UplinkMessage:
			registerForwardDataUplink(ctx, pld.UplinkMessage)
			events.Publish(evtForwardDataUplink.NewWithIdentifiersAndData(ctx, up.EndDeviceIds, up))
		case *ttnpb.ApplicationUp_JoinAccept:
			events.Publish(evtForwardJoinAccept.NewWithIdentifiersAndData(ctx, up.EndDeviceIds, &ttnpb.ApplicationUp{
				EndDeviceIds:   up.EndDeviceIds,
				CorrelationIds: up.CorrelationIds,
				Up: &ttnpb.ApplicationUp_JoinAccept{
					JoinAccept: applicationJoinAcceptWithoutAppSKey(pld.JoinAccept),
				},
			}))
		}
	}
	return nil
}

func (ns *NetworkServer) createProcessApplicationUplinkTask(consumerID string) func(context.Context) error {
	return func(ctx context.Context) error {
		return ns.processApplicationUplinkTask(ctx, consumerID)
	}
}

func (ns *NetworkServer) processApplicationUplinkTask(ctx context.Context, consumerID string) error {
	return ns.applicationUplinks.Pop(ctx, consumerID, func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, drain ApplicationUplinkQueueDrainFunc) (time.Time, error) {
		conn, err := ns.GetPeerConn(ctx, ttnpb.ClusterRole_APPLICATION_SERVER, nil)
		if err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to get Application Server peer")
			return time.Now().Add(applicationUplinkTaskRetryInterval), nil
		}

		cl := ttnpb.NewNsAsClient(conn)
		var sendErr bool
		if err := drain(applicationUplinkLimit, func(ups ...*ttnpb.ApplicationUp) error {
			err := ns.sendApplicationUplinks(ctx, cl, ups...)
			if err != nil {
				sendErr = true
			}
			return err
		}); err != nil {
			if !sendErr {
				log.FromContext(ctx).WithError(err).Error("Failed to drain application uplinks")
			}
			return time.Now().Add(applicationUplinkTaskRetryInterval), nil
		}
		return time.Time{}, nil
	})
}

func minAFCntDown(session *ttnpb.Session, macState *ttnpb.MACState) (uint32, error) {
	if session == nil || macState == nil {
		return 0, nil
	}
	var minFCnt uint32
	if macState.LorawanVersion.Compare(ttnpb.MACVersion_MAC_V1_1) >= 0 {
	outer:
		for i := len(macState.RecentDownlinks) - 1; i >= 0; i-- {
			pld := macState.RecentDownlinks[i].Payload
			switch pld.MHdr.MType {
			case ttnpb.MType_UNCONFIRMED_DOWN, ttnpb.MType_CONFIRMED_DOWN:
				macPayload := pld.GetMacPayload()
				if macPayload == nil {
					return 0, errInvalidPayload.New()
				}
				if macPayload.FPort > 0 && macPayload.FHdr.FCnt >= minFCnt {
					// NOTE: In an unlikely case all len(recentDowns) downlinks are FPort==0 or something unmatched in the switch (e.g. a proprietary downlink) minFCnt will
					// not reflect the correct AFCntDown - that is fine, because this is AS's responsibility and FCnt checking here is essentially just a sanity check.
					minFCnt = macPayload.FHdr.FCnt + 1
					break outer
				}
			case ttnpb.MType_JOIN_ACCEPT:
				// TODO: Support rejoins (https://github.com/TheThingsNetwork/lorawan-stack/issues/8).
				minFCnt = 0
				break outer
			case ttnpb.MType_PROPRIETARY:
			default:
				panic(fmt.Sprintf("invalid downlink MType: %s", pld.MHdr.MType))
			}
		}
	} else if session.LastNFCntDown > 0 || session.LastNFCntDown == 0 && len(macState.RecentDownlinks) > 0 {
		minFCnt = session.LastNFCntDown + 1
	}
	if len(session.QueuedApplicationDownlinks) > 0 {
		if fCnt := session.QueuedApplicationDownlinks[len(session.QueuedApplicationDownlinks)-1].FCnt; fCnt >= minFCnt {
			minFCnt = fCnt + 1
		}
	}
	return minFCnt, nil
}

// matchApplicationDownlinks validates downs and adds them to session.QueuedApplicationDownlinks.
// matchApplicationDownlinks returns downs, nil if session == nil.
func matchApplicationDownlinks(session *ttnpb.Session, macState *ttnpb.MACState, multicast bool, maxDownLen uint16, minFCnt uint32, makeQueueOperationErrorDetails func() *ttnpb.DownlinkQueueOperationErrorDetails, downs ...*ttnpb.ApplicationDownlink) (unmatched []*ttnpb.ApplicationDownlink, err error) {
	if session == nil {
		return downs, nil
	}
	downs, unmatched = ttnpb.PartitionDownlinksBySessionKeyIDEquality(session.Keys.SessionKeyId, downs...)
	switch {
	case len(downs) == 0:
		return unmatched, nil
	case len(downs) > 0 && macState == nil:
		return unmatched, errUnknownMACState.New()
	}

	for _, down := range downs {
		switch {
		case len(down.FrmPayload) > int(maxDownLen):
			return unmatched, errApplicationDownlinkTooLong.WithAttributes("length", len(down.FrmPayload), "max", maxDownLen)

		case down.FCnt < minFCnt:
			return unmatched, errFCntTooLow.WithAttributes("f_cnt", down.FCnt, "min_f_cnt", minFCnt).WithDetails(makeQueueOperationErrorDetails())

		case !bytes.Equal(down.SessionKeyId, session.Keys.SessionKeyId):
			return unmatched, errUnknownSession.WithDetails(makeQueueOperationErrorDetails())

		case multicast && down.Confirmed:
			return unmatched, errConfirmedMulticastDownlink.New()

		case multicast && len(down.GetClassBC().GetGateways()) == 0:
			return unmatched, errNoPath.New()

		case down.GetClassBC().GetAbsoluteTime() != nil && ttnpb.StdTime(down.GetClassBC().GetAbsoluteTime()).Before(time.Now().Add(macState.CurrentParameters.Rx1Delay.Duration()/2)):
			return unmatched, errExpiredDownlink.New()
		}
		minFCnt = down.FCnt + 1
		session.QueuedApplicationDownlinks = append(session.QueuedApplicationDownlinks, down)
	}
	return unmatched, nil
}

// matchQueuedApplicationDownlinks validates the given end device's application downlinks and adds them to appropriate session's queue.
// This function returns an error if any of the following checks fail:
// - An item's FCnt is not higher than the previous for the corresponding session;
// - An item's FRMPayload is longer than 250 bytes;
// - An item's session is neither the device's active session, nor device's pending session;
// - An item's session matches device's session, but corresponding MACState is missing;
// - The LoRaWAN version is 1.0.x and an item's FCnt is not higher than the session's NFCntDown.
func matchQueuedApplicationDownlinks(ctx context.Context, dev *ttnpb.EndDevice, fps *frequencyplans.Store, downs ...*ttnpb.ApplicationDownlink) error {
	if len(downs) == 0 {
		return nil
	}

	fp, phy, err := DeviceFrequencyPlanAndBand(dev, fps)
	if err != nil {
		return err
	}
	var maxDownLen uint16 = 0
	for _, dr := range phy.DataRates {
		if n := dr.MaxMACPayloadSize(fp.DwellTime.GetDownlinks()); n > maxDownLen {
			maxDownLen = n
		}
	}
	if maxDownLen < 8 {
		log.FromContext(ctx).Error("Data rate MAC payload size limits too low for application downlink to be scheduled")
		maxDownLen = 0
	} else {
		maxDownLen -= 8
	}

	minCurrentFCntDown, err := minAFCntDown(dev.Session, dev.MacState)
	if err != nil {
		return err
	}
	minPendingFCntDown, err := minAFCntDown(dev.PendingSession, dev.PendingMacState)
	if err != nil {
		return err
	}

	makeDownlinkQueueOperationErrorDetails := func() *ttnpb.DownlinkQueueOperationErrorDetails {
		d := &ttnpb.DownlinkQueueOperationErrorDetails{}
		if dev.Session != nil {
			d.DevAddr = &dev.Session.DevAddr
			d.SessionKeyId = dev.Session.Keys.SessionKeyId
			d.MinFCntDown = minCurrentFCntDown
		}
		if dev.PendingSession != nil {
			d.PendingDevAddr = &dev.PendingSession.DevAddr
			d.PendingSessionKeyId = dev.PendingSession.Keys.SessionKeyId
			d.PendingMinFCntDown = minPendingFCntDown
		}
		return d
	}

	unmatched, err := matchApplicationDownlinks(dev.Session, dev.MacState, dev.Multicast, maxDownLen, minCurrentFCntDown, makeDownlinkQueueOperationErrorDetails, downs...)
	if err != nil {
		return err
	}
	unmatched, err = matchApplicationDownlinks(dev.PendingSession, dev.PendingMacState, dev.Multicast, maxDownLen, minPendingFCntDown, makeDownlinkQueueOperationErrorDetails, unmatched...)
	if err != nil {
		return err
	}
	if len(unmatched) > 0 {
		return errUnknownSession.WithDetails(makeDownlinkQueueOperationErrorDetails())
	}
	return nil
}

var errDownlinkQueueCapacity = errors.DefineResourceExhausted("downlink_queue_capacity", "downlink queue capacity exceeded")

// DownlinkQueueReplace is called by the Application Server to completely replace the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueReplace(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*pbtypes.Empty, error) {
	if n := len(req.Downlinks); n > ns.downlinkQueueCapacity*2 {
		return nil, errDownlinkQueueCapacity.New()
	} else if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}

	ctx = log.NewContextWithField(ctx, "device_uid", unique.ID(ctx, req.EndDeviceIds))

	gets := []string{
		"mac_state",
		"multicast",
		"pending_mac_state",
		"pending_session",
		"session",
	}
	if len(req.Downlinks) > 0 {
		gets = append(gets,
			"frequency_plan_id",
			"last_dev_status_received_at",
			"lorawan_phy_version",
			"mac_settings",
		)
	}

	log.FromContext(ctx).WithField("downlink_count", len(req.Downlinks)).Debug("Replace downlink queue")
	dev, ctx, err := ns.devices.SetByID(ctx, req.EndDeviceIds.ApplicationIds, req.EndDeviceIds.DeviceId, gets,
		func(ctx context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if dev == nil {
				return nil, nil, errDeviceNotFound.New()
			}
			if dev.Session != nil {
				dev.Session.QueuedApplicationDownlinks = nil
			}
			if dev.PendingSession != nil {
				dev.PendingSession.QueuedApplicationDownlinks = nil
			}
			fps, err := ns.FrequencyPlansStore(ctx)
			if err != nil {
				return nil, nil, err
			}
			if err := matchQueuedApplicationDownlinks(ctx, dev, fps, req.Downlinks...); err != nil {
				return nil, nil, err
			}
			if len(dev.Session.GetQueuedApplicationDownlinks()) > ns.downlinkQueueCapacity || len(dev.PendingSession.GetQueuedApplicationDownlinks()) > ns.downlinkQueueCapacity {
				return nil, nil, errDownlinkQueueCapacity.New()
			}
			return dev, []string{
				"session.queued_application_downlinks",
				"pending_session.queued_application_downlinks",
			}, nil
		},
	)
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to replace application downlink queue")
		return nil, err
	}

	ctx = log.NewContextWithFields(ctx, log.Fields(
		"active_session_queue_length", len(dev.Session.GetQueuedApplicationDownlinks()),
		"pending_session_queue_length", len(dev.PendingSession.GetQueuedApplicationDownlinks()),
	))
	log.FromContext(ctx).Debug("Replaced application downlink queue")

	if len(req.Downlinks) > 0 {
		if err := ns.updateDataDownlinkTask(ctx, dev, time.Time{}); err != nil {
			log.FromContext(ctx).WithError(err).Error("Failed to update downlink task queue after downlink queue replace")
		}
	}
	return ttnpb.Empty, nil
}

// DownlinkQueuePush is called by the Application Server to push a downlink to queue for a device.
func (ns *NetworkServer) DownlinkQueuePush(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*pbtypes.Empty, error) {
	if n := len(req.Downlinks); n > ns.downlinkQueueCapacity*2 {
		return nil, errDownlinkQueueCapacity.New()
	} else if n == 0 {
		return ttnpb.Empty, nil
	} else if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}

	ctx = log.NewContextWithField(ctx, "device_uid", unique.ID(ctx, req.EndDeviceIds))

	log.FromContext(ctx).WithField("downlink_count", len(req.Downlinks)).Debug("Push application downlink to queue")
	dev, ctx, err := ns.devices.SetByID(ctx, req.EndDeviceIds.ApplicationIds, req.EndDeviceIds.DeviceId,
		[]string{
			"frequency_plan_id",
			"last_dev_status_received_at",
			"lorawan_phy_version",
			"mac_settings",
			"mac_state",
			"multicast",
			"pending_mac_state",
			"pending_session",
			"session",
		},
		func(ctx context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if dev == nil {
				return nil, nil, errDeviceNotFound.New()
			}
			fps, err := ns.FrequencyPlansStore(ctx)
			if err != nil {
				return nil, nil, err
			}
			if err := matchQueuedApplicationDownlinks(ctx, dev, fps, req.Downlinks...); err != nil {
				return nil, nil, err
			}
			if len(dev.Session.GetQueuedApplicationDownlinks()) > ns.downlinkQueueCapacity || len(dev.PendingSession.GetQueuedApplicationDownlinks()) > ns.downlinkQueueCapacity {
				return nil, nil, errDownlinkQueueCapacity.New()
			}
			return dev, []string{
				"session.queued_application_downlinks",
				"pending_session.queued_application_downlinks",
			}, nil
		},
	)
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to push application downlink to queue")
		return nil, err
	}

	ctx = log.NewContextWithFields(ctx, log.Fields(
		"active_session_queue_length", len(dev.Session.GetQueuedApplicationDownlinks()),
		"pending_session_queue_length", len(dev.PendingSession.GetQueuedApplicationDownlinks()),
	))
	log.FromContext(ctx).Debug("Pushed application downlink to queue")

	if err := ns.updateDataDownlinkTask(ctx, dev, time.Time{}); err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to update downlink task queue after downlink queue push")
	}
	return ttnpb.Empty, nil
}

// DownlinkQueueList is called by the Application Server to get the current state of the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueList(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (*ttnpb.ApplicationDownlinks, error) {
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}
	dev, ctx, err := ns.devices.GetByID(ctx, ids.ApplicationIds, ids.DeviceId, []string{
		"session.queued_application_downlinks",
		"pending_session.queued_application_downlinks",
	})
	if err != nil {
		logRegistryRPCError(ctx, err, "Failed to list application downlink queue")
		return nil, err
	}
	return &ttnpb.ApplicationDownlinks{
		Downlinks: append(dev.Session.GetQueuedApplicationDownlinks(), dev.PendingSession.GetQueuedApplicationDownlinks()...),
	}, nil
}
