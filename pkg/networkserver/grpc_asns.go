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
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

type applicationUpStream struct {
	ttnpb.AsNs_LinkApplicationServer
	closeCh chan struct{}
}

func (s applicationUpStream) Close() error {
	close(s.closeCh)
	return nil
}

// LinkApplication is called by the Application Server to subscribe to application events.
func (ns *NetworkServer) LinkApplication(link ttnpb.AsNs_LinkApplicationServer) error {
	ctx := link.Context()

	ids := ttnpb.ApplicationIdentifiers{
		ApplicationID: rpcmetadata.FromIncomingContext(ctx).ID,
	}
	var err error
	if err = ids.ValidateContext(ctx); err != nil {
		return err
	}
	if err = rights.RequireApplication(ctx, ids, ttnpb.RIGHT_APPLICATION_LINK); err != nil {
		return err
	}

	ws := &applicationUpStream{
		AsNs_LinkApplicationServer: link,
		closeCh:                    make(chan struct{}),
	}

	uid := unique.ID(ctx, ids)

	logger := log.FromContext(ctx).WithField("application_uid", uid)

	ns.applicationServersMu.Lock()
	cl, ok := ns.applicationServers[uid]
	ns.applicationServers[uid] = ws
	if ok {
		if err := cl.Close(); err != nil {
			ns.applicationServersMu.Unlock()
			logger.WithError(err).Warn("Failed to link application")
			return err
		}
	}
	ns.applicationServersMu.Unlock()

	logger.Debug("Linked application")

	events.Publish(evtBeginApplicationLink(ctx, ids, nil))
	defer events.Publish(evtEndApplicationLink(ctx, ids, err))

	select {
	case <-ctx.Done():
		err := ctx.Err()
		ns.applicationServersMu.Lock()
		cl, ok := ns.applicationServers[uid]
		if ok && cl == ws {
			delete(ns.applicationServers, uid)
		}
		ns.applicationServersMu.Unlock()
		return err
	case <-ws.closeCh:
		return errDuplicateSubscription
	}
}

// DownlinkQueueReplace is called by the Application Server to completely replace the downlink queue for a device.
func (ns *NetworkServer) DownlinkQueueReplace(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_LINK); err != nil {
		return nil, err
	}

	logger := log.FromContext(ctx).WithField("device_uid", unique.ID(ctx, req.EndDeviceIdentifiers))

	dev, err := ns.devices.SetByID(ctx, req.EndDeviceIdentifiers.ApplicationIdentifiers, req.EndDeviceIdentifiers.DeviceID,
		[]string{
			"queued_application_downlinks",
			"mac_state.device_class",
		},
		func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if dev == nil {
				return nil, nil, errDeviceNotFound
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
	if dev.MACState != nil {
		logger = logger.WithField("device_class", dev.MACState.DeviceClass)
	}
	logger.Debug("Replaced application downlink queue")
	if dev.MACState != nil && dev.MACState.DeviceClass != ttnpb.CLASS_A && len(dev.QueuedApplicationDownlinks) > 0 {
		startAt := time.Now().UTC()
		logger.WithField("start_at", startAt).Debug("Add downlink task with application downlink pending")
		return ttnpb.Empty, ns.downlinkTasks.Add(ctx, req.EndDeviceIdentifiers, startAt, false)
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
			"queued_application_downlinks",
			"mac_state.device_class",
		},
		func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			if dev == nil {
				return nil, nil, errDeviceNotFound
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
	if dev.MACState != nil {
		logger = logger.WithField("device_class", dev.MACState.DeviceClass)
	}
	logger.Debug("Pushed application downlink to queue")
	if dev.MACState != nil && dev.MACState.DeviceClass != ttnpb.CLASS_A {
		startAt := time.Now().UTC()
		logger.WithField("start_at", startAt).Debug("Add downlink task with application downlink pending")
		return ttnpb.Empty, ns.downlinkTasks.Add(ctx, req.EndDeviceIdentifiers, startAt, false)
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
