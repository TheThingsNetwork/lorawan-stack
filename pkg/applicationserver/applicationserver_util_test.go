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

package applicationserver_test

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

var (
	Timeout          = (1 << 6) * test.Delay
	EventsBufferSize = (1 << 6)
)

// MockDeviceRegistry is a mock DeviceRegistry used for testing.
type MockDeviceRegistry struct {
	GetFunc func(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, paths []string) (*ttnpb.EndDevice, error)
	SetFunc func(
		ctx context.Context,
		ids *ttnpb.EndDeviceIdentifiers,
		paths []string,
		f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error),
	) (*ttnpb.EndDevice, error)
	BatchDeleteFunc func(context.Context, *ttnpb.ApplicationIdentifiers, []string) ([]*ttnpb.EndDeviceIdentifiers, error)
}

// Get calls GetFunc if set and panics otherwise.
func (r MockDeviceRegistry) Get(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, paths []string) (*ttnpb.EndDevice, error) {
	if r.GetFunc == nil {
		panic("Get called, but not set")
	}
	return r.GetFunc(ctx, ids, paths)
}

// Set calls SetFunc if set and panics otherwise.
func (r MockDeviceRegistry) Set(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
	if r.SetFunc == nil {
		panic("Set called, but not set")
	}
	return r.SetFunc(ctx, ids, paths, f)
}

// Range is a no-op.
func (r MockDeviceRegistry) Range(ctx context.Context, paths []string, f func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.EndDevice) bool) error {
	return nil
}

// MockLinkRegistry is a mock LinkRegistry used for testing.
type MockLinkRegistry struct {
	GetFunc   func(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, paths []string) (*ttnpb.ApplicationLink, error)
	RangeFunc func(ctx context.Context, paths []string, f func(context.Context, *ttnpb.ApplicationIdentifiers, *ttnpb.ApplicationLink) bool) error
	SetFunc   func(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, paths []string, f func(*ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, []string, error)) (*ttnpb.ApplicationLink, error)
}

// Get calls GetFunc if set and panics otherwise.
func (m MockLinkRegistry) Get(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, paths []string) (*ttnpb.ApplicationLink, error) {
	if m.GetFunc == nil {
		panic("Get called, but not set")
	}
	return m.GetFunc(ctx, ids, paths)
}

// Range calls RangeFunc if set and panics otherwise.
func (m MockLinkRegistry) Range(ctx context.Context, paths []string, f func(context.Context, *ttnpb.ApplicationIdentifiers, *ttnpb.ApplicationLink) bool) error {
	if m.RangeFunc == nil {
		panic("Range called, but not set")
	}
	return m.RangeFunc(ctx, paths, f)
}

// Set calls SetFunc if set and panics otherwise.
func (m MockLinkRegistry) Set(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, paths []string, f func(*ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, []string, error)) (*ttnpb.ApplicationLink, error) {
	if m.SetFunc == nil {
		panic("Set called, but not set")
	}
	return m.SetFunc(ctx, ids, paths, f)
}

// BatchDelete calls BatchDeleteFunc if set and panics otherwise.
func (r MockDeviceRegistry) BatchDelete(
	ctx context.Context,
	appIDs *ttnpb.ApplicationIdentifiers,
	deviceIDs []string,
) ([]*ttnpb.EndDeviceIdentifiers, error) {
	if r.BatchDeleteFunc == nil {
		panic("BatchDelete called, but not set")
	}
	return r.BatchDeleteFunc(ctx, appIDs, deviceIDs)
}
