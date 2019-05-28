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
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoservices"
	"go.thethings.network/lorawan-stack/pkg/crypto/cryptoutil"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/joinserver/provisioning"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
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
			RootKeyID: rootKeysEnc.RootKeyID,
		}
		cs := srv.JS.GetPeer(ctx, ttnpb.PeerInfo_CRYPTO_SERVER, dev.EndDeviceIdentifiers)
		if ttnpb.HasAnyField(req.FieldMask.Paths, "root_keys.nwk_key") {
			var networkCryptoService cryptoservices.Network
			if rootKeysEnc.GetNwkKey() != nil {
				nwkKey, err := cryptoutil.UnwrapAES128Key(*rootKeysEnc.NwkKey, srv.JS.KeyVault)
				if err != nil {
					return nil, err
				}
				networkCryptoService = cryptoservices.NewMemory(&nwkKey, nil)
			} else if cs != nil {
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
			} else if cs != nil {
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
	gets := append(req.FieldMask.Paths[:0:0], req.FieldMask.Paths...)
	return srv.JS.devices.SetByID(ctx, req.EndDevice.ApplicationIdentifiers, req.EndDevice.DeviceID, gets, func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		if dev != nil {
			return &req.EndDevice, req.FieldMask.Paths, nil
		}
		sets := append(req.FieldMask.Paths,
			"ids.application_ids",
			"ids.dev_eui",
			"ids.device_id",
			"ids.join_eui",
		)
		if req.EndDevice.DevAddr != nil {
			sets = append(sets,
				"ids.dev_addr",
			)
		}
		return &req.EndDevice, sets, nil
	})
}

func (srv jsEndDeviceRegistryServer) Provision(req *ttnpb.ProvisionEndDevicesRequest, stream ttnpb.JsEndDeviceRegistry_ProvisionServer) error {
	if err := rights.RequireApplication(stream.Context(), req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE_KEYS); err != nil {
		return err
	}

	provisioner := provisioning.Get(req.ProvisionerID)
	if provisioner == nil {
		return errProvisionerNotFound.WithAttributes("id", req.ProvisionerID)
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
			if ids.JoinEUI == nil {
				ids.JoinEUI = devices.List.JoinEUI
			}
			return &ttnpb.EndDevice{
				EndDeviceIdentifiers: ids,
			}, nil
		}
	case *ttnpb.ProvisionEndDevicesRequest_Range:
		devEUIInt := binary.BigEndian.Uint64(devices.Range.StartDevEUI[:])
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
			deviceID, err := provisioner.DefaultDeviceID(joinEUI, devEUI, entry)
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
	case *ttnpb.ProvisionEndDevicesRequest_FromData:
		next = func(entry *pbtypes.Struct) (*ttnpb.EndDevice, error) {
			var joinEUI types.EUI64
			if devices.FromData.JoinEUI != nil {
				joinEUI = *devices.FromData.JoinEUI
			} else {
				var err error
				if joinEUI, err = provisioner.DefaultJoinEUI(entry); err != nil {
					return nil, err
				}
			}
			devEUI, err := provisioner.DefaultDevEUI(entry)
			if err != nil {
				return nil, err
			}
			deviceID, err := provisioner.DefaultDeviceID(joinEUI, devEUI, entry)
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

	entries, err := provisioner.Decode(req.ProvisioningData)
	if err != nil {
		return errProvisionerDecode.WithCause(err)
	}
	for _, entry := range entries {
		dev, err := next(entry)
		if err != nil {
			return err
		}
		if err := dev.EndDeviceIdentifiers.ValidateContext(stream.Context()); err != nil {
			return err
		}
		if dev.JoinEUI == nil || dev.JoinEUI.IsZero() {
			return errNoJoinEUI
		}
		if dev.DevEUI == nil || dev.DevEUI.IsZero() {
			return errNoDevEUI
		}
		dev.ProvisionerID = req.ProvisionerID
		dev.ProvisioningData = entry
		if err := stream.Send(dev); err != nil {
			return err
		}
	}
	return nil
}

// Delete implements ttnpb.JsEndDeviceRegistryServer.
func (srv jsEndDeviceRegistryServer) Delete(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, ids.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	_, err := srv.JS.devices.SetByID(ctx, ids.ApplicationIdentifiers, ids.DeviceID, nil, func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
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
