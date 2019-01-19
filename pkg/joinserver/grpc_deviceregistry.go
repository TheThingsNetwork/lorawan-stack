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
	"encoding/binary"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/joinserver/provisioning"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
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

func (s jsEndDeviceRegistryServer) Provision(req *ttnpb.ProvisionEndDevicesRequest, stream ttnpb.JsEndDeviceRegistry_ProvisionServer) error {
	if err := rights.RequireApplication(stream.Context(), req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS); err != nil {
		return err
	}

	provisioner := provisioning.Get(req.Provisioner)
	if provisioner == nil {
		return errProvisionerNotFound.WithAttributes("id", req.Provisioner)
	}

	var next func(*pbtypes.Struct) (*ttnpb.EndDevice, error)
	switch devices := req.EndDevices.(type) {
	case *ttnpb.ProvisionEndDevicesRequest_List:
		i := 0
		next = func(*pbtypes.Struct) (*ttnpb.EndDevice, error) {
			if i == len(devices.List.EndDeviceIDs) {
				return nil, errProvisionEntryCount.WithAttributes(
					"expected", len(devices.List.EndDeviceIDs),
					"actual", i+1,
				)
			}
			ids := devices.List.EndDeviceIDs[i]
			i++
			if ids.ApplicationIdentifiers != req.ApplicationIdentifiers {
				return nil, errInvalidIdentifiers
			}
			if ids.JoinEUI == nil || ids.JoinEUI.IsZero() {
				return nil, errNoJoinEUI
			}
			if ids.DevEUI == nil || ids.DevEUI.IsZero() {
				return nil, errNoDevEUI
			}
			return &ttnpb.EndDevice{
				EndDeviceIdentifiers: ids,
			}, nil
		}
	case *ttnpb.ProvisionEndDevicesRequest_Range:
		if devices.Range.FromDevEUI.IsZero() {
			return errNoDevEUI
		}
		devEUIInt := binary.BigEndian.Uint64(devices.Range.FromDevEUI[:])
		next = func(entry *pbtypes.Struct) (*ttnpb.EndDevice, error) {
			var devEUI types.EUI64
			binary.BigEndian.PutUint64(devEUI[:], devEUIInt)
			devEUIInt++
			var joinEUI types.EUI64
			if devices.Range.JoinEUI != nil {
				joinEUI = *devices.Range.JoinEUI
			} else {
				var err error
				if joinEUI, err = provisioner.DefaultJoinEUI(entry); err != nil {
					return nil, err
				}
			}
			deviceID, err := provisioner.DeviceID(joinEUI, devEUI, entry)
			if err != nil {
				return nil, err
			}
			return &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					ApplicationIdentifiers: req.ApplicationIdentifiers,
					DeviceID:               deviceID,
					JoinEUI:                &joinEUI,
					DevEUI:                 &devEUI,
				},
			}, nil
		}
	default:
		return errInvalidIdentifiers
	}

	entries, err := provisioner.Decode(req.Data)
	if err != nil {
		return errProvisionerDecode.WithCause(err)
	}
	for _, entry := range entries {
		dev, err := next(entry)
		if err != nil {
			return err
		}
		dev.Provisioner = req.Provisioner
		dev.ProvisioningData = entry
		if err := stream.Send(dev); err != nil {
			return err
		}
	}
	return nil
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
