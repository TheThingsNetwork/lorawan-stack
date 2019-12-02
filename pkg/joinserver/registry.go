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

// DeviceRegistry is a registry, containing devices.
type DeviceRegistry interface {
	GetByEUI(ctx context.Context, joinEUI types.EUI64, devEUI types.EUI64, paths []string) (*ttnpb.ContextualEndDevice, error)
	GetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error)
	SetByEUI(ctx context.Context, joinEUI types.EUI64, devEUI types.EUI64, paths []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.ContextualEndDevice, error)
	SetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
}

// DeleteDevice deletes device identified by joinEUI, devEUI from r.
func DeleteDevice(ctx context.Context, r DeviceRegistry, appID ttnpb.ApplicationIdentifiers, devID string) error {
	_, err := r.SetByID(ctx, appID, devID, nil, func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) { return nil, nil, nil })
	return err
}

// KeyRegistry is a registry, containing session keys.
type KeyRegistry interface {
	GetByID(ctx context.Context, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.ContextualSessionKeys, error)
	SetByID(ctx context.Context, devEUI types.EUI64, id []byte, paths []string, f func(*ttnpb.ContextualSessionKeys) (*ttnpb.SessionKeys, []string, error)) (*ttnpb.ContextualSessionKeys, error)
}

// DeleteKeys deletes session keys identified by devEUI, id pair from r.
func DeleteKeys(ctx context.Context, r KeyRegistry, devEUI types.EUI64, id []byte) error {
	_, err := r.SetByID(ctx, devEUI, id, nil, func(keys *ttnpb.ContextualSessionKeys) (*ttnpb.SessionKeys, []string, error) { return nil, nil, nil })
	return err
}
