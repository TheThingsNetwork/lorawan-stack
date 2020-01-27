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

package networkserver

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// DeviceRegistry is a registry, containing devices.
type DeviceRegistry interface {
	GetByEUI(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, context.Context, error)
	GetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, context.Context, error)
	RangeByAddr(ctx context.Context, devAddr types.DevAddr, paths []string, f func(context.Context, *ttnpb.EndDevice) bool) error
	SetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error)
}
