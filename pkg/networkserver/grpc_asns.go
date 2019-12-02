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
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
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
	events.Publish(evtBeginApplicationLink(ctx, ids, nil))
	err := ns.applicationUplinks.Subscribe(ctx, ids, func(ctx context.Context, up *ttnpb.ApplicationUp) error {
		if err := link.Send(up); err != nil {
			return err
		}
		_, err := link.Recv()
		return err
	})
	logger.WithError(err).Debug("Close application link")
	events.Publish(evtEndApplicationLink(ctx, ids, err))
	return err
}

func validateApplicationDownlinks(session ttnpb.Session, macState *ttnpb.MACState, multicast bool, queue []*ttnpb.ApplicationDownlink, downs ...*ttnpb.ApplicationDownlink) (unmatchedQueue, unmatchedDowns []*ttnpb.ApplicationDownlink, err error) {
	queue, unmatchedQueue = partitionDownlinksBySessionKeyIDEquality(session.SessionKeyID, queue...)
	downs, unmatchedDowns = partitionDownlinksBySessionKeyIDEquality(session.SessionKeyID, downs...)
	switch {
	case len(downs) == 0:
		return unmatchedQueue, unmatchedDowns, nil
	case len(downs) > 0 && macState == nil:
		return unmatchedQueue, unmatchedDowns, errUnknownMACState
	}

	var minFCnt uint32
	if macState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) >= 0 {
	outer:
		for i := len(macState.RecentDownlinks) - 1; i >= 0; i-- {
			pld := macState.RecentDownlinks[i].Payload
			switch pld.MType {
			case ttnpb.MType_UNCONFIRMED_DOWN, ttnpb.MType_CONFIRMED_DOWN:
				macPayload := pld.GetMACPayload()
				if macPayload.FPort > 0 && macPayload.FCnt >= minFCnt {
					// NOTE: In an unlikely case all len(recentDowns) downlinks are FPort==0 or something unmatched in the switch (e.g. a proprietary downlink) minFCnt will
					// not reflect the correct AFCntDown - that is fine, because this is AS's responsibilty and FCnt checking here is essentially just a sanity check.
					minFCnt = macPayload.FCnt + 1
					break outer
				}
			}
		}
	} else if session.LastNFCntDown > 0 || session.LastNFCntDown == 0 && len(macState.RecentDownlinks) > 0 {
		minFCnt = session.LastNFCntDown + 1
	}
	if len(queue) > 0 {
		if fCnt := queue[len(queue)-1].FCnt; fCnt >= minFCnt {
			minFCnt = fCnt + 1
		}
	}
	for _, down := range downs {
		switch {
		case len(down.FRMPayload) > 250:
			return unmatchedQueue, unmatchedDowns, errApplicationDownlinkTooLong

		case down.FCnt < minFCnt:
			return unmatchedQueue, unmatchedDowns, errFCntTooLow

		case !bytes.Equal(down.SessionKeyID, session.SessionKeyID):
			return unmatchedQueue, unmatchedDowns, errUnknownSession

		case multicast && down.Confirmed:
			return unmatchedQueue, unmatchedDowns, errConfirmedMulticastDownlink

		case multicast && len(down.GetClassBC().GetGateways()) == 0:
			return unmatchedQueue, unmatchedDowns, errNoPath

		case down.GetClassBC().GetAbsoluteTime() != nil && down.GetClassBC().GetAbsoluteTime().Before(timeNow()):
			return unmatchedQueue, unmatchedDowns, errExpiredDownlink
		}
		minFCnt = down.FCnt + 1
	}
	return unmatchedQueue, unmatchedDowns, nil
}

// validateQueuedApplicationDownlinks validates the given end device's application downlink queue.
// This function returns an error if any of the following checks fail:
// - An item's FCnt is not higher than the previous for the corresponding session;
// - An item's FRMPayload is longer than 250 bytes;
// - An item's session is neither the device's active session, nor device's pending session;
// - An item's session matches device's session, but corresponding MACState is missing;
// - The LoRaWAN version is 1.0.x and an item's FCnt is not higher than the session's NFCntDown.
func validateQueuedApplicationDownlinks(dev *ttnpb.EndDevice, downs ...*ttnpb.ApplicationDownlink) error {
	if len(downs) == 0 {
		return nil
	}

	var err error
	unmatchedDowns := downs
	unmatchedQueue := dev.QueuedApplicationDownlinks
	if dev.Session != nil {
		unmatchedDowns, unmatchedQueue, err = validateApplicationDownlinks(*dev.Session, dev.MACState, dev.Multicast, unmatchedQueue, unmatchedDowns...)
		if err != nil {
			return err
		}
	}
	if dev.PendingSession != nil {
		unmatchedDowns, unmatchedQueue, err = validateApplicationDownlinks(*dev.PendingSession, dev.PendingMACState, dev.Multicast, unmatchedQueue, unmatchedDowns...)
		if err != nil {
			return err
		}
	}
	if len(unmatchedDowns) > 0 {
		return errUnknownSession
	}
	return nil
}

// DownlinkQueueReplace is called by the Application Server to completely replace the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueReplace(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_LINK); err != nil {
		return nil, err
	}

	logger := log.FromContext(ctx).WithField("device_uid", unique.ID(ctx, req.EndDeviceIdentifiers))

	gets := []string{
		"mac_state",
		"multicast",
		"pending_mac_state",
		"pending_session",
		"queued_application_downlinks",
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

	dev, err := ns.devices.SetByID(ctx, req.EndDeviceIdentifiers.ApplicationIdentifiers, req.EndDeviceIdentifiers.DeviceID, gets,
		func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if dev == nil {
				return nil, nil, errDeviceNotFound
			}
			dev.QueuedApplicationDownlinks = nil
			if err := validateQueuedApplicationDownlinks(dev, req.Downlinks...); err != nil {
				return nil, nil, err
			}
			dev.QueuedApplicationDownlinks = req.Downlinks
			return dev, []string{"queued_application_downlinks"}, nil
		},
	)
	if err != nil {
		logger.WithError(err).Warn("Failed to replace application downlink queue")
		return nil, err
	}

	logger = logger.WithField("queue_length", len(dev.QueuedApplicationDownlinks))
	logger.Debug("Replaced application downlink queue")

	if len(dev.QueuedApplicationDownlinks) == 0 || dev.MACState == nil {
		return ttnpb.Empty, nil
	}

	var downAt time.Time
	_, phy, err := getDeviceBandVersion(dev, ns.FrequencyPlans)
	if err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to determine device band")
		downAt = timeNow().UTC()
	} else {
		var ok bool
		downAt, ok = nextDataDownlinkAt(ctx, dev, phy, ns.defaultMACSettings)
		if !ok {
			return ttnpb.Empty, nil
		}
	}
	downAt = downAt.Add(-nsScheduleWindow)
	log.FromContext(ctx).WithField("start_at", downAt).Debug("Add downlink task after downlink queue replace")
	if err := ns.downlinkTasks.Add(ctx, dev.EndDeviceIdentifiers, downAt, true); err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to add downlink task after downlink queue replace")
	}
	return ttnpb.Empty, nil
}

// DownlinkQueuePush is called by the Application Server to push a downlink to queue for a device.
func (ns *NetworkServer) DownlinkQueuePush(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*pbtypes.Empty, error) {
	if len(req.Downlinks) == 0 {
		return ttnpb.Empty, nil
	}

	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_LINK); err != nil {
		return nil, err
	}

	logger := log.FromContext(ctx).WithField("device_uid", unique.ID(ctx, req.EndDeviceIdentifiers))

	dev, err := ns.devices.SetByID(ctx, req.EndDeviceIdentifiers.ApplicationIdentifiers, req.EndDeviceIdentifiers.DeviceID,
		[]string{
			"frequency_plan_id",
			"last_dev_status_received_at",
			"lorawan_phy_version",
			"mac_settings",
			"mac_state",
			"multicast",
			"pending_mac_state",
			"pending_session",
			"queued_application_downlinks",
			"recent_uplinks",
			"session",
		},
		func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if dev == nil {
				return nil, nil, errDeviceNotFound
			}
			if err := validateQueuedApplicationDownlinks(dev, req.Downlinks...); err != nil {
				return nil, nil, err
			}
			dev.QueuedApplicationDownlinks = append(dev.QueuedApplicationDownlinks, req.Downlinks...)
			return dev, []string{"queued_application_downlinks"}, nil
		},
	)
	if err != nil {
		logger.WithError(err).Warn("Failed to push application downlink to queue")
		return nil, err
	}

	logger = logger.WithField("queue_length", len(dev.QueuedApplicationDownlinks))
	logger.Debug("Pushed application downlink to queue")

	if dev.MACState == nil {
		return ttnpb.Empty, nil
	}

	var downAt time.Time
	_, phy, err := getDeviceBandVersion(dev, ns.FrequencyPlans)
	if err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to determine device band")
		downAt = timeNow().UTC()
	} else {
		var ok bool
		downAt, ok = nextDataDownlinkAt(ctx, dev, phy, ns.defaultMACSettings)
		if !ok {
			return ttnpb.Empty, nil
		}
	}
	downAt = downAt.Add(-nsScheduleWindow)
	log.FromContext(ctx).WithField("start_at", downAt).Debug("Add downlink task after downlink queue push")
	if err := ns.downlinkTasks.Add(ctx, dev.EndDeviceIdentifiers, downAt, true); err != nil {
		log.FromContext(ctx).WithError(err).Error("Failed to add downlink task after downlink queue push")
	}
	return ttnpb.Empty, nil
}

// DownlinkQueueList is called by the Application Server to get the current state of the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueList(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (*ttnpb.ApplicationDownlinks, error) {
	if err := rights.RequireApplication(ctx, ids.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_LINK); err != nil {
		return nil, err
	}
	dev, err := ns.devices.GetByID(ctx, ids.ApplicationIdentifiers, ids.DeviceID, []string{"queued_application_downlinks"})
	if err != nil {
		return nil, err
	}
	return &ttnpb.ApplicationDownlinks{Downlinks: dev.QueuedApplicationDownlinks}, nil
}
