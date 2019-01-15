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

package joinserver

import (
	"context"
	"fmt"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/joinserver/provisioning"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/warning"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

type jsEndDeviceRegistryServer struct {
	Registry DeviceRegistry
}

// Get implements ttnpb.JsEndDeviceRegistryServer.
func (s jsEndDeviceRegistryServer) Get(ctx context.Context, req *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if req.JoinEUI == nil || req.JoinEUI.IsZero() {
		return nil, errNoJoinEUI
	}
	if req.DevEUI == nil || req.DevEUI.IsZero() {
		return nil, errNoDevEUI
	}
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_READ); err != nil {
		return nil, err
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "root_keys") {
		if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS); err != nil {
			return nil, err
		}
	}
	// TODO: Validate field mask (https://github.com/TheThingsIndustries/lorawan-stack/issues/1226)
	dev, err := s.Registry.GetByEUI(ctx, *req.EndDeviceIdentifiers.JoinEUI, *req.EndDeviceIdentifiers.DevEUI, req.FieldMask.Paths)
	if errors.IsNotFound(err) {
		return nil, errDeviceNotFound
	}
	if err != nil {
		return nil, err
	}
	if !dev.ApplicationIdentifiers.Equal(req.ApplicationIdentifiers) {
		return nil, errDeviceNotFound
	}
	return dev, nil
}

// Set implements ttnpb.JsEndDeviceRegistryServer.
func (s jsEndDeviceRegistryServer) Set(ctx context.Context, req *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if req.Device.JoinEUI == nil || req.Device.JoinEUI.IsZero() {
		return nil, errNoJoinEUI
	}
	if req.Device.DevEUI == nil || req.Device.DevEUI.IsZero() {
		return nil, errNoDevEUI
	}
	if err := rights.RequireApplication(ctx, req.Device.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "root_keys") {
		if err := rights.RequireApplication(ctx, req.Device.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS); err != nil {
			return nil, err
		}
	}
	// TODO: Validate field mask (https://github.com/TheThingsIndustries/lorawan-stack/issues/1226)
	return s.Registry.SetByEUI(ctx, *req.Device.JoinEUI, *req.Device.DevEUI, req.FieldMask.Paths, func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		if dev != nil && !dev.ApplicationIdentifiers.Equal(req.Device.ApplicationIdentifiers) {
			return nil, nil, errInvalidIdentifiers
		}
		return &req.Device, req.FieldMask.Paths, nil
	})
}

func (s jsEndDeviceRegistryServer) Provision(ctx context.Context, req *ttnpb.ProvisionEndDevicesRequest) (*ttnpb.EndDevices, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS); err != nil {
		return nil, err
	}
	for _, dev := range req.EndDeviceIDs {
		if dev.ApplicationIdentifiers != req.ApplicationIdentifiers {
			return nil, errInvalidIdentifiers
		}
		if dev.JoinEUI == nil || dev.JoinEUI.IsZero() {
			return nil, errNoJoinEUI
		}
		if dev.DevEUI == nil || dev.DevEUI.IsZero() {
			return nil, errNoDevEUI
		}
	}
	provisioner := provisioning.Get(req.Provisioner)
	if provisioner == nil {
		return nil, errProvisionerNotFound.WithAttributes("id", req.Provisioner)
	}
	entries, err := provisioner.Decode(req.Data)
	if err != nil {
		return nil, errProvisionerDecode.WithCause(err)
	}
	if len(entries) != len(req.EndDeviceIDs) {
		return nil, errProvisionEntryCount.WithAttributes(
			"expected", len(req.EndDeviceIDs),
			"actual", len(entries),
		)
	}
	res := &ttnpb.EndDevices{
		EndDevices: make([]*ttnpb.EndDevice, 0, len(entries)),
	}
	for i, entry := range entries {
		ids := req.EndDeviceIDs[i]
		dev, err := s.Registry.SetByEUI(ctx, *ids.JoinEUI, *ids.DevEUI,
			[]string{
				"provisioner",
				"provisioning_data",
			},
			func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
				if dev == nil {
					return nil, nil, errDeviceNotFound.WithAttributes("device_uid", unique.ID(ctx, ids))
				}
				var paths []string
				dev.Provisioner = req.Provisioner
				dev.ProvisioningData = entry
				paths = append(paths, "provisioner", "provisioning_data")
				return dev, paths, nil
			},
		)
		if err != nil {
			warning.Add(ctx, fmt.Sprintf("%d: %s", i+1, err.Error()))
		} else {
			res.EndDevices = append(res.EndDevices, dev)
		}
	}
	if len(res.EndDevices) == 0 {
		return nil, errProvisioning
	}
	return res, nil
}

// Delete implements ttnpb.JsEndDeviceRegistryServer.
func (s jsEndDeviceRegistryServer) Delete(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	if ids.JoinEUI == nil || ids.JoinEUI.IsZero() {
		return nil, errNoJoinEUI
	}
	if ids.DevEUI == nil || ids.DevEUI.IsZero() {
		return nil, errNoDevEUI
	}
	if err := rights.RequireApplication(ctx, ids.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	_, err := s.Registry.SetByEUI(ctx, *ids.JoinEUI, *ids.DevEUI, nil, func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		if dev == nil || !dev.ApplicationIdentifiers.Equal(ids.ApplicationIdentifiers) {
			return nil, nil, errDeviceNotFound
		}
		return nil, nil, nil
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, err
}
