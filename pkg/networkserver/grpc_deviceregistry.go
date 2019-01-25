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
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Get implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) Get(ctx context.Context, req *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_READ); err != nil {
		return nil, err
	}
	return ns.devices.GetByEUI(ctx, *req.EndDeviceIdentifiers.JoinEUI, *req.EndDeviceIdentifiers.DevEUI, req.FieldMask.Paths)
}

// Set implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) Set(ctx context.Context, req *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if err := rights.RequireApplication(ctx, req.Device.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	var addDownlinkTask bool
	dev, err := ns.devices.SetByID(ctx, req.Device.EndDeviceIdentifiers.ApplicationIdentifiers, req.Device.EndDeviceIdentifiers.DeviceID, req.FieldMask.Paths, func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		paths := req.FieldMask.Paths
		if dev == nil && !req.Device.SupportsJoin {
			if req.Device.Session == nil {
				return nil, nil, errEmptySession
			}
			if err := resetMACState(&req.Device, ns.FrequencyPlans); err != nil {
				return nil, nil, err
			}
			if req.Device.MACState.DeviceClass != ttnpb.CLASS_A {
				addDownlinkTask = len(req.Device.QueuedApplicationDownlinks) > 0 || !dev.MACState.CurrentParameters.Equal(dev.MACState.DesiredParameters)
			}
			paths = append(paths, "mac_state")
		} else if ttnpb.HasAnyField(paths, "mac_state.device_class") && dev.MACState.DeviceClass == ttnpb.CLASS_A && req.Device.MACState.DeviceClass != ttnpb.CLASS_A {
			addDownlinkTask = true
		}
		return &req.Device, paths, nil
	})
	if err != nil {
		return nil, err
	}
	if addDownlinkTask {
		if err = ns.downlinkTasks.Add(ctx, dev.EndDeviceIdentifiers, time.Now()); err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to add downlink task for device after set")
		}
	}
	return dev, nil
}

// Delete implements NsEndDeviceRegistryServer.
func (ns *NetworkServer) Delete(ctx context.Context, req *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	_, err := ns.devices.SetByID(ctx, req.ApplicationIdentifiers, req.DeviceID, nil, func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		return nil, nil, nil
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, err
}
