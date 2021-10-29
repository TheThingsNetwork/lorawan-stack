// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package commands

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/cmd/internal/shared"
	ns "go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	nsredis "go.thethings.network/lorawan-stack/v3/pkg/networkserver/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/redis"
)

// NewNSDeviceRegistryCleaner returns a new instance of Network Server RegistryCleaner with a local set
// of devices.
func NewNSDeviceRegistryCleaner(ctx context.Context, config *redis.Config) (*ns.RegistryCleaner, error) {
	deviceRegistry := &nsredis.DeviceRegistry{
		Redis:   redis.New(config.WithNamespace("ns", "devices")),
		LockTTL: defaultLockTTL,
	}
	if err := deviceRegistry.Init(ctx); err != nil {
		return nil, shared.ErrInitializeApplicationServer.WithCause(err)
	}
	cleaner := &ns.RegistryCleaner{
		DevRegistry: deviceRegistry,
	}
	err := cleaner.RangeToLocalSet(ctx)
	if err != nil {
		return nil, err
	}
	return cleaner, nil
}
