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

func validateApplicationDownlinks(macState ttnpb.MACState, session ttnpb.Session, downs ...*ttnpb.ApplicationDownlink) error {
	var lastFCnt *uint32
	if macState.LoRaWANVersion.Compare(ttnpb.MAC_V1_1) < 0 {
		lastFCnt = &session.LastNFCntDown
	}
	for _, down := range downs {
		if down.ClassBC != nil && macState.DeviceClass == ttnpb.CLASS_A {
			return errClassBCForClassA
		}
		if !bytes.Equal(down.SessionKeyID, session.SessionKeyID) {
			return errUnknownSession
		}
		if lastFCnt != nil && down.FCnt <= *lastFCnt {
			return errFCntTooLow
		}
		lastFCnt = &down.FCnt
	}
	return nil
}

// validateQueuedApplicationDownlinks validates the given end device's application downlink queue.
// This function returns an error if any of the following checks fail:
// - The device has neither MACState and Session, nor PendingMACState and PendingSession set.
// - Items belong to different sessions;
// - An item has ClassBC set, but device is in Class A mode.
// - An item's FRMPayload is longer than 250.
// - An item's session is neither the device's session or pending session;
// - An item's FCnt is not higher than the previous for the corresponding session;
// - The LoRaWAN version is 1.0.x and an item's FCnt is not higher than the session's NFCntDown.
func validateQueuedApplicationDownlinks(dev *ttnpb.EndDevice) error {
	for _, down := range dev.QueuedApplicationDownlinks {
		if dev.Multicast {
			if down.Confirmed {
				return errConfirmedMulticastDownlink
			}
			if len(down.GetClassBC().GetGateways()) == 0 {
				return errNoPath
			}
		}
		if absTime := down.GetClassBC().GetAbsoluteTime(); absTime != nil && absTime.Before(timeNow()) {
			return errExpiredDownlink
		}
		if len(down.FRMPayload) > 250 {
			return errInvalidPayload
		}
	}
	hasActiveSession := dev.MACState != nil && dev.Session != nil
	hasPendingSession := dev.PendingMACState != nil && dev.PendingSession != nil
	switch {
	case !hasActiveSession && !hasPendingSession:
		return errUnknownMACState
	case !hasPendingSession:
		return validateApplicationDownlinks(*dev.MACState, *dev.Session, dev.QueuedApplicationDownlinks...)
	case !hasActiveSession:
		return validateApplicationDownlinks(*dev.PendingMACState, *dev.PendingSession, dev.QueuedApplicationDownlinks...)
	default:
		if validateApplicationDownlinks(*dev.PendingMACState, *dev.PendingSession, dev.QueuedApplicationDownlinks...) == nil {
			return nil
		}
		return validateApplicationDownlinks(*dev.MACState, *dev.Session, dev.QueuedApplicationDownlinks...)
	}
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
			dev.QueuedApplicationDownlinks = req.Downlinks
			if err := validateQueuedApplicationDownlinks(dev); err != nil {
				return nil, nil, err
			}
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
			dev.QueuedApplicationDownlinks = append(dev.QueuedApplicationDownlinks, req.Downlinks...)
			if err := validateQueuedApplicationDownlinks(dev); err != nil {
				return nil, nil, err
			}
			return dev, []string{"queued_application_downlinks"}, nil
		},
	)
	if err != nil {
		logger.WithError(err).Warn("Failed to push application downlink to queue")
		return nil, err
	}

	logger = logger.WithField("queue_length", len(dev.QueuedApplicationDownlinks))
	logger.Debug("Pushed application downlink to queue")

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
