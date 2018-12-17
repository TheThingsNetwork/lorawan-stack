// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type jsEndDeviceRegistryServer struct {
	JS *JoinServer
}

// Get implements ttnpb.JsEndDeviceRegistryServer.
func (s jsEndDeviceRegistryServer) Get(ctx context.Context, req *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	// TODO: Change JsEndDeviceRegistry to not work with EndDeviceIdentifiers (https://github.com/TheThingsIndustries/lorawan-stack/pull/1374)
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}
	return s.JS.devices.GetByEUI(ctx, *req.EndDeviceIdentifiers.JoinEUI, *req.EndDeviceIdentifiers.DevEUI, req.FieldMask.Paths)
}

// Set implements ttnpb.AsEndDeviceRegistryServer.
func (s jsEndDeviceRegistryServer) Set(ctx context.Context, req *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	// TODO: Change JsEndDeviceRegistry to not work with EndDeviceIdentifiers (https://github.com/TheThingsIndustries/lorawan-stack/pull/1374)
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}
	return s.JS.devices.SetByEUI(ctx, *req.Device.EndDeviceIdentifiers.JoinEUI, *req.Device.EndDeviceIdentifiers.DevEUI, req.FieldMask.Paths, func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		return &req.Device, req.FieldMask.Paths, nil
	})
}

// Delete implements ttnpb.AsEndDeviceRegistryServer.
func (s jsEndDeviceRegistryServer) Delete(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	// TODO: Change JsEndDeviceRegistry to not work with EndDeviceIdentifiers (https://github.com/TheThingsIndustries/lorawan-stack/pull/1374)
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}
	_, err := s.JS.devices.SetByEUI(ctx, *ids.JoinEUI, *ids.DevEUI, nil, func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		return nil, nil, nil
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, err
}
