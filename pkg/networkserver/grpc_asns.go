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
	"io"
	"sync"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

type ApplicationUplinkQueue interface {
	// Add adds application uplinks ups to queue.
	// Implementations must ensure that Add returns fast.
	Add(ctx context.Context, ups ...*ttnpb.ApplicationUp) error

	// Subscribe calls f sequentially for each application uplink in the queue.
	// If f returns a non-nil error or ctx is done, Subscribe stops the iteration.
	// TODO: Use ...*ttnpb.ApplicationUp in callback once https://github.com/TheThingsNetwork/lorawan-stack/issues/1523 is implemented.
	Subscribe(ctx context.Context, appID ttnpb.ApplicationIdentifiers, f func(context.Context, *ttnpb.ApplicationUp) error) error
}

type applicationUpStream struct {
	cancel    context.CancelFunc
	waitClose func()
}

func (s applicationUpStream) Close() error {
	s.cancel()
	s.waitClose()
	return nil
}

func applicationJoinAcceptWithoutAppSKey(pld *ttnpb.ApplicationJoinAccept) *ttnpb.ApplicationJoinAccept {
	return &ttnpb.ApplicationJoinAccept{
		SessionKeyID:         pld.SessionKeyID,
		InvalidatedDownlinks: pld.InvalidatedDownlinks,
		PendingSession:       pld.PendingSession,
		ReceivedAt:           pld.ReceivedAt,
	}
}

// LinkApplication is called by the Application Server to subscribe to application events.
func (ns *NetworkServer) LinkApplication(link ttnpb.AsNs_LinkApplicationServer) error {
	ctx := link.Context()

	ids := ttnpb.ApplicationIdentifiers{
		ApplicationID: rpcmetadata.FromIncomingContext(ctx).ID,
	}
	if err := ids.ValidateContext(ctx); err != nil {
		return err
	}
	if err := rights.RequireApplication(ctx, ids, ttnpb.RIGHT_APPLICATION_LINK); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)

	ctx, cancel := context.WithCancel(ctx)
	ws := &applicationUpStream{
		cancel:    cancel,
		waitClose: wg.Wait,
	}
	defer func() {
		wg.Done()
	}()

	uid := unique.ID(ctx, ids)
	logger := log.FromContext(ctx).WithField("application_uid", uid)

	v, ok := ns.applicationServers.LoadOrStore(uid, ws)
	for ok {
		logger.Debug("Close existing application link")
		if err := v.(io.Closer).Close(); err != nil {
			logger.WithError(err).Warn("Failed to close existing application link")
		}
		v, ok = ns.applicationServers.LoadOrStore(uid, ws)
	}
	defer ns.applicationServers.Delete(uid)

	logger.Debug("Linked application")
	events.Publish(evtBeginApplicationLink.NewWithIdentifiersAndData(ctx, ids, nil))
	err := ns.applicationUplinks.Subscribe(ctx, ids, func(ctx context.Context, up *ttnpb.ApplicationUp) error {
		if err := link.Send(up); err != nil {
			return err
		}
		if _, err := link.Recv(); err != nil {
			return err
		}
		switch pld := up.Up.(type) {
		case *ttnpb.ApplicationUp_UplinkMessage:
			registerForwardDataUplink(ctx, pld.UplinkMessage)
			events.Publish(evtForwardDataUplink.NewWithIdentifiersAndData(ctx, up.EndDeviceIdentifiers, up))
		case *ttnpb.ApplicationUp_JoinAccept:
			events.Publish(evtForwardJoinAccept.NewWithIdentifiersAndData(ctx, up.EndDeviceIdentifiers, &ttnpb.ApplicationUp{
				EndDeviceIdentifiers: up.EndDeviceIdentifiers,
				CorrelationIDs:       up.CorrelationIDs,
				Up: &ttnpb.ApplicationUp_JoinAccept{
					JoinAccept: applicationJoinAcceptWithoutAppSKey(pld.JoinAccept),
				},
			}))
		}
		return nil
	})
	logger.WithError(err).Debug("Close application link")
	events.Publish(evtEndApplicationLink.NewWithIdentifiersAndData(ctx, ids, err))
	return err
}

// matchApplicationDownlinks validates downs and adds them to session.QueuedApplicationDownlinks.
// matchApplicationDownlinks returns downs, nil if session == nil.
func matchApplicationDownlinks(session *ttnpb.Session, macState *ttnpb.MACState, multicast bool, maxDownLen uint16, downs ...*ttnpb.ApplicationDownlink) (unmatched []*ttnpb.ApplicationDownlink, err error) {
	if session == nil {
		return downs, nil
	}
	downs, unmatched = ttnpb.PartitionDownlinksBySessionKeyIDEquality(session.SessionKeyID, downs...)
	switch {
	case len(downs) == 0:
		return unmatched, nil
	case len(downs) > 0 && macState == nil:
		return unmatched, errUnknownMACState.New()
	}

	var minFCnt uint32
	if macState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) >= 0 {
	outer:
		for i := len(macState.RecentDownlinks) - 1; i >= 0; i-- {
			var pld ttnpb.Message
			if err := lorawan.UnmarshalMessage(macState.RecentDownlinks[i].RawPayload, &pld); err != nil {
				return unmatched, errDecodePayload.WithCause(err)
			}
			switch pld.MType {
			case ttnpb.MType_UNCONFIRMED_DOWN, ttnpb.MType_CONFIRMED_DOWN:
				macPayload := pld.GetMACPayload()
				if macPayload == nil {
					return unmatched, errInvalidPayload.New()
				}
				if macPayload.FPort > 0 && macPayload.FCnt >= minFCnt {
					// NOTE: In an unlikely case all len(recentDowns) downlinks are FPort==0 or something unmatched in the switch (e.g. a proprietary downlink) minFCnt will
					// not reflect the correct AFCntDown - that is fine, because this is AS's responsibility and FCnt checking here is essentially just a sanity check.
					minFCnt = macPayload.FCnt + 1
					break outer
				}
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

	for _, down := range downs {
		switch {
		case len(down.FRMPayload) > int(maxDownLen):
			return unmatched, errApplicationDownlinkTooLong.WithAttributes("length", len(down.FRMPayload), "max", maxDownLen)

		case down.FCnt < minFCnt:
			return unmatched, errFCntTooLow.WithAttributes("f_cnt", down.FCnt, "min_f_cnt", minFCnt)

		case !bytes.Equal(down.SessionKeyID, session.SessionKeyID):
			return unmatched, errUnknownSession.New()

		case multicast && down.Confirmed:
			return unmatched, errConfirmedMulticastDownlink.New()

		case multicast && len(down.GetClassBC().GetGateways()) == 0:
			return unmatched, errNoPath.New()

		case down.GetClassBC().GetAbsoluteTime() != nil && down.GetClassBC().GetAbsoluteTime().Before(timeNow().Add(macState.CurrentParameters.Rx1Delay.Duration()/2)):
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

	fp, phy, err := deviceFrequencyPlanAndBand(dev, fps)
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

	unmatched, err := matchApplicationDownlinks(dev.Session, dev.MACState, dev.Multicast, maxDownLen, downs...)
	if err != nil {
		return err
	}
	unmatched, err = matchApplicationDownlinks(dev.PendingSession, dev.PendingMACState, dev.Multicast, maxDownLen, unmatched...)
	if err != nil {
		return err
	}
	if len(unmatched) > 0 {
		return errUnknownSession.New()
	}
	return nil
}

var errDownlinkQueueCapacityExceeded = errors.DefineResourceExhausted("downlink_queue_capacity_exceeded", "Downlink queue capacity exceeded")

// DownlinkQueueReplace is called by the Application Server to completely replace the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueReplace(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*pbtypes.Empty, error) {
	if n := len(req.Downlinks); n > ns.downlinkQueueCapacity*2 {
		return nil, errDownlinkQueueCapacityExceeded.New()
	} else if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_LINK); err != nil {
		return nil, err
	}

	ctx = log.NewContextWithField(ctx, "device_uid", unique.ID(ctx, req.EndDeviceIdentifiers))

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
			"recent_uplinks",
		)
	}

	log.FromContext(ctx).WithField("downlink_count", len(req.Downlinks)).Debug("Replace downlink queue")
	dev, ctx, err := ns.devices.SetByID(ctx, req.EndDeviceIdentifiers.ApplicationIdentifiers, req.EndDeviceIdentifiers.DeviceID, gets,
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
			if err := matchQueuedApplicationDownlinks(ctx, dev, ns.FrequencyPlans, req.Downlinks...); err != nil {
				return nil, nil, err
			}
			if len(dev.Session.GetQueuedApplicationDownlinks()) > ns.downlinkQueueCapacity || len(dev.PendingSession.GetQueuedApplicationDownlinks()) > ns.downlinkQueueCapacity {
				return nil, nil, errDownlinkQueueCapacityExceeded.New()
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

	if err := ns.updateDataDownlinkTask(ctx, dev, time.Time{}); err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to update downlink task queue after downlink queue replace")
	}
	return ttnpb.Empty, nil
}

// DownlinkQueuePush is called by the Application Server to push a downlink to queue for a device.
func (ns *NetworkServer) DownlinkQueuePush(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*pbtypes.Empty, error) {
	if n := len(req.Downlinks); n > ns.downlinkQueueCapacity*2 {
		return nil, errDownlinkQueueCapacityExceeded.New()
	} else if n == 0 {
		return ttnpb.Empty, nil
	} else if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_LINK); err != nil {
		return nil, err
	}

	ctx = log.NewContextWithField(ctx, "device_uid", unique.ID(ctx, req.EndDeviceIdentifiers))

	log.FromContext(ctx).WithField("downlink_count", len(req.Downlinks)).Debug("Push application downlink to queue")
	dev, ctx, err := ns.devices.SetByID(ctx, req.EndDeviceIdentifiers.ApplicationIdentifiers, req.EndDeviceIdentifiers.DeviceID,
		[]string{
			"frequency_plan_id",
			"last_dev_status_received_at",
			"lorawan_phy_version",
			"mac_settings",
			"mac_state",
			"multicast",
			"pending_mac_state",
			"pending_session",
			"recent_uplinks",
			"session",
		},
		func(ctx context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if dev == nil {
				return nil, nil, errDeviceNotFound.New()
			}
			if err := matchQueuedApplicationDownlinks(ctx, dev, ns.FrequencyPlans, req.Downlinks...); err != nil {
				return nil, nil, err
			}
			if len(dev.Session.GetQueuedApplicationDownlinks()) > ns.downlinkQueueCapacity || len(dev.PendingSession.GetQueuedApplicationDownlinks()) > ns.downlinkQueueCapacity {
				return nil, nil, errDownlinkQueueCapacityExceeded.New()
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
	if err := rights.RequireApplication(ctx, ids.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_LINK); err != nil {
		return nil, err
	}
	dev, ctx, err := ns.devices.GetByID(ctx, ids.ApplicationIdentifiers, ids.DeviceID, []string{
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
