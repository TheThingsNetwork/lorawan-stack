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

package applicationserver

import (
	"context"

	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/redis"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func NewRedisDeviceRegistry(cl *redis.Client) DeviceRegistry {
	return &redisDeviceRegistry{}
}

type redisDeviceRegistry struct {
}

func (r *redisDeviceRegistry) Get(context.Context, ttnpb.EndDeviceIdentifiers) (*ttnpb.EndDevice, error) {
	return nil, errors.New("not implemented")
}

func (r *redisDeviceRegistry) Set(context.Context, ttnpb.EndDeviceIdentifiers, func(*ttnpb.EndDevice) (*ttnpb.EndDevice, error)) error {
	return errors.New("not implemented")
}

func NewRedisLinkRegistry(cl *redis.Client) LinkRegistry {
	return &redisLinkRegistry{}
}

type redisLinkRegistry struct {
}

func (r *redisLinkRegistry) Get(context.Context, ttnpb.ApplicationIdentifiers) (*ttnpb.ApplicationLink, error) {
	return nil, errors.New("not implemented")
}

func (r *redisLinkRegistry) Set(context.Context, ttnpb.ApplicationIdentifiers, func(*ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, error)) error {
	return errors.New("not implemented")
}

func (r *redisLinkRegistry) Range(context.Context, func(ttnpb.ApplicationIdentifiers, *ttnpb.ApplicationLink) bool) error {
	return errors.New("not implemented")
}
