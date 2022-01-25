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

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

type mockISEndDeviceRegistry struct {
	ttnpb.EndDeviceRegistryServer

	endDevicesMu sync.RWMutex
	endDevices   map[string]*ttnpb.EndDevice
}

func (m *mockISEndDeviceRegistry) Add(ctx context.Context, dev *ttnpb.EndDevice) {
	m.endDevicesMu.Lock()
	defer m.endDevicesMu.Unlock()
	m.endDevices[unique.ID(ctx, dev.Ids)] = dev
}

func (m *mockISEndDeviceRegistry) Get(ctx context.Context, in *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	m.endDevicesMu.RLock()
	defer m.endDevicesMu.RUnlock()
	if dev, ok := m.endDevices[unique.ID(ctx, in.EndDeviceIds)]; ok {
		return dev, nil
	}
	return nil, errNotFound.New()
}

func (m *mockISEndDeviceRegistry) Update(ctx context.Context, in *ttnpb.UpdateEndDeviceRequest) (*ttnpb.EndDevice, error) {
	m.endDevicesMu.Lock()
	defer m.endDevicesMu.Unlock()
	dev, ok := m.endDevices[unique.ID(ctx, in.EndDevice.Ids)]
	if !ok {
		return nil, errNotFound.New()
	}
	if err := dev.SetFields(in.EndDevice, in.GetFieldMask().GetPaths()...); err != nil {
		return nil, err
	}
	m.endDevices[unique.ID(ctx, in.EndDevice.Ids)] = dev
	return dev, nil
}
