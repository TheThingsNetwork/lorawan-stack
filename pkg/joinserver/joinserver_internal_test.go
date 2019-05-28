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

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

var (
	ErrAddressNotAuthorized = errAddressNotAuthorized
	ErrDevNonceTooSmall     = errDevNonceTooSmall
	ErrNoAppSKey            = errNoAppSKey
	ErrNoFNwkSIntKey        = errNoFNwkSIntKey
	ErrNoNwkSEncKey         = errNoNwkSEncKey
	ErrNoSNwkSIntKey        = errNoSNwkSIntKey
	ErrRegistryOperation    = errRegistryOperation
	ErrReuseDevNonce        = errReuseDevNonce
)

func KeyToBytes(key types.AES128Key) []byte { return key[:] }

func KeyPtr(key types.AES128Key) *types.AES128Key { return &key }

type AsJsServer = asJsServer
type NsJsServer = nsJsServer
type JsDeviceServer = jsEndDeviceRegistryServer

type MockDeviceRegistry struct {
	GetByEUIFunc func(context.Context, types.EUI64, types.EUI64, []string) (*ttnpb.EndDevice, error)
	GetByIDFunc  func(context.Context, ttnpb.ApplicationIdentifiers, string, []string) (*ttnpb.EndDevice, error)
	SetByEUIFunc func(context.Context, types.EUI64, types.EUI64, []string, func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
	SetByIDFunc  func(context.Context, ttnpb.ApplicationIdentifiers, string, []string, func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
}

// GetByEUI calls GetByEUIFunc if set and panics otherwise.
func (m MockDeviceRegistry) GetByEUI(ctx context.Context, joinEUI types.EUI64, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error) {
	if m.GetByEUIFunc == nil {
		panic("GetByEUI called, but not set")
	}
	return m.GetByEUIFunc(ctx, joinEUI, devEUI, paths)
}

// GetByID calls GetByIDFunc if set and panics otherwise.
func (m MockDeviceRegistry) GetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
	if m.GetByIDFunc == nil {
		panic("GetByID called, but not set")
	}
	return m.GetByIDFunc(ctx, appID, devID, paths)
}

// SetByEUI calls SetByEUIFunc if set and panics otherwise.
func (m MockDeviceRegistry) SetByEUI(ctx context.Context, joinEUI types.EUI64, devEUI types.EUI64, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
	if m.SetByEUIFunc == nil {
		panic("SetByEUI called, but not set")
	}
	return m.SetByEUIFunc(ctx, joinEUI, devEUI, paths, f)
}

// SetByID calls SetByIDFunc if set and panics otherwise.
func (m MockDeviceRegistry) SetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
	if m.SetByIDFunc == nil {
		panic("SetByID called, but not set")
	}
	return m.SetByIDFunc(ctx, appID, devID, paths, f)
}

type MockKeyRegistry struct {
	GetByIDFunc func(context.Context, types.EUI64, []byte, []string) (*ttnpb.SessionKeys, error)
	SetByIDFunc func(context.Context, types.EUI64, []byte, []string, func(*ttnpb.SessionKeys) (*ttnpb.SessionKeys, []string, error)) (*ttnpb.SessionKeys, error)
}

// GetByID calls GetByIDFunc if set and panics otherwise.
func (m MockKeyRegistry) GetByID(ctx context.Context, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
	if m.GetByIDFunc == nil {
		panic("GetByID called, but not set")
	}
	return m.GetByIDFunc(ctx, devEUI, id, paths)
}

// SetByID calls SetByIDFunc if set and panics otherwise.
func (m MockKeyRegistry) SetByID(ctx context.Context, devEUI types.EUI64, id []byte, paths []string, f func(*ttnpb.SessionKeys) (*ttnpb.SessionKeys, []string, error)) (*ttnpb.SessionKeys, error) {
	if m.SetByIDFunc == nil {
		panic("SetByID called, but not set")
	}
	return m.SetByIDFunc(ctx, devEUI, id, paths, f)
}
