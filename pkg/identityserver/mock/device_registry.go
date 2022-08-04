// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package mockis

import (
	"context"
	"sync"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

var errFetchDevice = errors.DefineInternal("fetch_device", "failed to fetch device")

type mockISEndDeviceRegistry struct {
	ttnpb.EndDeviceRegistryServer
	endDevices sync.Map
}

func (m *mockISEndDeviceRegistry) Add(ctx context.Context, dev *ttnpb.EndDevice) {
	m.endDevices.Store(unique.ID(ctx, dev.Ids), dev)
}

func (m *mockISEndDeviceRegistry) load(id string) (*ttnpb.EndDevice, error) {
	v, ok := m.endDevices.Load(id)
	if !ok || v == nil {
		return nil, errNotFound.New()
	}
	dev, ok := v.(*ttnpb.EndDevice)
	if !ok {
		return nil, errFetchDevice.New()
	}
	return dev, nil
}

func (m *mockISEndDeviceRegistry) Get(ctx context.Context, in *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	return m.load(unique.ID(ctx, in.GetEndDeviceIds()))
}

func (m *mockISEndDeviceRegistry) Update(
	ctx context.Context,
	in *ttnpb.UpdateEndDeviceRequest,
) (*ttnpb.EndDevice, error) {
	dev, err := m.load(unique.ID(ctx, in.GetEndDevice().GetIds()))
	if err != nil {
		return nil, err
	}
	if err := dev.SetFields(in.EndDevice, in.GetFieldMask().GetPaths()...); err != nil {
		return nil, err
	}
	m.Add(ctx, dev)
	return dev, nil
}
