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

package applicationserver

import (
	"context"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type asEndDeviceRegistryServer struct {
	registry DeviceRegistry
}

// Get implements ttnpb.AsEndDeviceRegistryServer.
func (r asEndDeviceRegistryServer) Get(ctx context.Context, req *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_READ); err != nil {
		return nil, err
	}
	return r.registry.Get(ctx, req.EndDeviceIdentifiers, req.FieldMask.Paths)
}

var errInvalidFieldValue = errors.DefineInvalidArgument("field_value", "invalid value of field `{field}`")

// Set implements ttnpb.AsEndDeviceRegistryServer.
func (r asEndDeviceRegistryServer) Set(ctx context.Context, req *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if err := rights.RequireApplication(ctx, req.EndDevice.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}

	if ttnpb.HasAnyField(req.FieldMask.Paths, "ids.dev_addr") && req.EndDevice.DevAddr != nil && !req.EndDevice.DevAddr.IsZero() {
		return nil, errInvalidFieldValue.WithAttributes("field", "ids.dev_addr")
	}

	gets := append(req.FieldMask.Paths[:0:0], req.FieldMask.Paths...)
	return r.registry.Set(ctx, req.EndDevice.EndDeviceIdentifiers, gets, func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		if dev != nil {
			return &req.EndDevice, req.FieldMask.Paths, nil
		}

		sets := append(req.FieldMask.Paths,
			"ids.application_ids",
			"ids.device_id",
		)
		if req.EndDevice.JoinEUI != nil && !req.EndDevice.JoinEUI.IsZero() {
			sets = append(sets,
				"ids.join_eui",
			)
		}
		if req.EndDevice.DevEUI != nil && !req.EndDevice.DevEUI.IsZero() {
			sets = append(sets,
				"ids.dev_eui",
			)
		}
		return &req.EndDevice, sets, nil
	})
}

// Delete implements ttnpb.AsEndDeviceRegistryServer.
func (r asEndDeviceRegistryServer) Delete(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, ids.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	_, err := r.registry.Set(ctx, *ids, nil, func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		return nil, nil, nil
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}
