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

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoservices"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtCreateEndDevice = events.Define(
		"js.end_device.create", "create end device",
		ttnpb.RIGHT_APPLICATION_DEVICES_READ,
	)
	evtUpdateEndDevice = events.Define(
		"js.end_device.update", "update end device",
		ttnpb.RIGHT_APPLICATION_DEVICES_READ,
	)
	evtDeleteEndDevice = events.Define(
		"js.end_device.delete", "delete end device",
		ttnpb.RIGHT_APPLICATION_DEVICES_READ,
	)
)

type jsEndDeviceRegistryServer struct {
	JS *JoinServer
}

// Get implements ttnpb.JsEndDeviceRegistryServer.
func (srv jsEndDeviceRegistryServer) Get(ctx context.Context, req *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_READ); err != nil {
		return nil, err
	}
	paths := req.FieldMask.Paths
	if ttnpb.HasAnyField(req.FieldMask.Paths, "root_keys") {
		if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_READ_KEYS); err != nil {
			return nil, err
		}
		paths = append(paths, "provisioner_id", "provisioning_data")
	}
	dev, err := srv.JS.devices.GetByID(ctx, req.ApplicationIdentifiers, req.DeviceID, paths)
	if errors.IsNotFound(err) {
		return nil, errDeviceNotFound
	}
	if err != nil {
		return nil, err
	}
	if !dev.ApplicationIdentifiers.Equal(req.ApplicationIdentifiers) {
		return nil, errDeviceNotFound
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "root_keys") {
		rootKeysEnc := dev.RootKeys
		dev.RootKeys = &ttnpb.RootKeys{
			RootKeyID: rootKeysEnc.GetRootKeyID(),
		}
		cs, _ := srv.JS.GetPeer(ctx, ttnpb.ClusterRole_CRYPTO_SERVER, dev.EndDeviceIdentifiers)
		if ttnpb.HasAnyField(req.FieldMask.Paths, "root_keys.nwk_key") {
			var networkCryptoService cryptoservices.Network
			if rootKeysEnc.GetNwkKey() != nil {
				nwkKey, err := cryptoutil.UnwrapAES128Key(*rootKeysEnc.NwkKey, srv.JS.KeyVault)
				if err != nil {
					return nil, err
				}
				networkCryptoService = cryptoservices.NewMemory(&nwkKey, nil)
			} else if cs != nil && dev.ProvisionerID != "" {
				networkCryptoService = cryptoservices.NewNetworkRPCClient(cs.Conn(), srv.JS.KeyVault, srv.JS.WithClusterAuth())
			}
			if networkCryptoService != nil {
				if nwkKey, err := networkCryptoService.GetNwkKey(ctx, dev); err == nil {
					dev.RootKeys.NwkKey = &ttnpb.KeyEnvelope{
						Key: &nwkKey,
					}
				} else {
					return nil, err
				}
			}
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "root_keys.app_key") {
			var applicationCryptoService cryptoservices.Application
			if rootKeysEnc.GetAppKey() != nil {
				appKey, err := cryptoutil.UnwrapAES128Key(*rootKeysEnc.AppKey, srv.JS.KeyVault)
				if err != nil {
					return nil, err
				}
				applicationCryptoService = cryptoservices.NewMemory(nil, &appKey)
			} else if cs != nil && dev.ProvisionerID != "" {
				applicationCryptoService = cryptoservices.NewApplicationRPCClient(cs.Conn(), srv.JS.KeyVault, srv.JS.WithClusterAuth())
			}
			if applicationCryptoService != nil {
				if appKey, err := applicationCryptoService.GetAppKey(ctx, dev); err == nil {
					dev.RootKeys.AppKey = &ttnpb.KeyEnvelope{
						Key: &appKey,
					}
				} else {
					return nil, err
				}
			}
		}
	}
	return dev, nil
}

var (
	errInvalidFieldMask  = errors.DefineInvalidArgument("field_mask", "invalid field mask")
	errInvalidFieldValue = errors.DefineInvalidArgument("field_value", "invalid value of field `{field}`")
)

// Set implements ttnpb.JsEndDeviceRegistryServer.
func (srv jsEndDeviceRegistryServer) Set(ctx context.Context, req *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	if req.EndDevice.JoinEUI == nil || req.EndDevice.JoinEUI.IsZero() {
		return nil, errNoJoinEUI
	}
	if req.EndDevice.DevEUI == nil || req.EndDevice.DevEUI.IsZero() {
		return nil, errNoDevEUI
	}

	if err := rights.RequireApplication(ctx, req.EndDevice.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "root_keys") {
		if err := rights.RequireApplication(ctx, req.EndDevice.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS); err != nil {
			return nil, err
		}
	}

	var evt events.Event
	dev, err := srv.JS.devices.SetByID(ctx, req.EndDevice.ApplicationIdentifiers, req.EndDevice.DeviceID, req.FieldMask.Paths, func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		sets := req.FieldMask.Paths
		if dev == nil {
			evt = evtCreateEndDevice(ctx, req.EndDevice.EndDeviceIdentifiers, nil)
		} else {
			evt = evtUpdateEndDevice(ctx, req.EndDevice.EndDeviceIdentifiers, req.FieldMask.Paths)
			if err := ttnpb.ProhibitFields(req.FieldMask.Paths, "ids.dev_addr"); err != nil {
				return nil, nil, errInvalidFieldMask.WithCause(err)
			}
			return &req.EndDevice, sets, nil
		}

		if req.EndDevice.DevAddr != nil && !req.EndDevice.DevAddr.IsZero() {
			return nil, nil, errInvalidFieldValue.WithAttributes("field", "ids.dev_addr")
		}
		return &req.EndDevice, append(req.FieldMask.Paths,
			"ids.application_ids",
			"ids.dev_eui",
			"ids.device_id",
			"ids.join_eui",
		), nil
	})
	if err != nil {
		return nil, err
	}
	if evt != nil {
		events.Publish(evt)
	}
	return dev, nil
}

// Provision is deprecated.
// TODO: Remove (https://github.com/TheThingsNetwork/lorawan-stack/issues/999)
func (srv jsEndDeviceRegistryServer) Provision(req *ttnpb.ProvisionEndDevicesRequest, stream ttnpb.JsEndDeviceRegistry_ProvisionServer) error {
	if err := rights.RequireApplication(stream.Context(), req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS); err != nil {
		return err
	}
	return errProvisionerNotFound.WithAttributes("id", req.ProvisionerID)
}

// Delete implements ttnpb.JsEndDeviceRegistryServer.
func (srv jsEndDeviceRegistryServer) Delete(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, ids.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	var evt events.Event
	_, err := srv.JS.devices.SetByID(ctx, ids.ApplicationIdentifiers, ids.DeviceID, nil, func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		if dev == nil || !dev.ApplicationIdentifiers.Equal(ids.ApplicationIdentifiers) {
			return nil, nil, errDeviceNotFound
		}
		evt = evtDeleteEndDevice(ctx, ids, nil)
		return nil, nil, nil
	})
	if err != nil {
		return nil, err
	}
	if evt != nil {
		events.Publish(evt)
	}
	return ttnpb.Empty, err
}
