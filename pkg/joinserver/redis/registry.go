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

package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	ttnredis "go.thethings.network/lorawan-stack/pkg/redis"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

var (
	errInvalidIdentifiers   = errors.DefineInvalidArgument("invalid_identifiers", "invalid identifiers")
	errDuplicateIdentifiers = errors.DefineAlreadyExists("duplicate_identifiers", "duplicate identifiers")
)

type DeviceRegistry struct {
	Redis *ttnredis.Client
}

func (r *DeviceRegistry) GetByEUI(ctx context.Context, joinEUI types.EUI64, devEUI types.EUI64) (*ttnpb.EndDevice, error) {
	if joinEUI.IsZero() || devEUI.IsZero() {
		return nil, errInvalidIdentifiers
	}

	pb := &ttnpb.EndDevice{}
	if err := ttnredis.GetProto(r.Redis, r.Redis.Key(joinEUI.String(), devEUI.String())).ScanProto(pb); err != nil {
		return nil, err
	}
	return pb, nil
}

func (r *DeviceRegistry) SetByEUI(ctx context.Context, joinEUI types.EUI64, devEUI types.EUI64, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, error)) (*ttnpb.EndDevice, error) {
	if joinEUI.IsZero() || devEUI.IsZero() {
		return nil, errInvalidIdentifiers
	}

	k := r.Redis.Key(joinEUI.String(), devEUI.String())

	var pb *ttnpb.EndDevice
	err := r.Redis.Watch(func(tx *redis.Tx) error {
		var create bool
		pb = &ttnpb.EndDevice{}
		if err := ttnredis.GetProto(tx, k).ScanProto(pb); errors.IsNotFound(err) {
			create = true
			pb = nil
		} else if err != nil {
			return err
		}

		createdAt := pb.GetCreatedAt()

		var err error
		pb, err = f(pb)
		if err != nil {
			return err
		}

		var f func(redis.Pipeliner) error
		if pb == nil {
			f = func(p redis.Pipeliner) error {
				p.Del(k)
				return nil
			}
		} else {
			pb.UpdatedAt = time.Now().UTC()
			if create {
				pb.CreatedAt = pb.UpdatedAt
			} else {
				pb.CreatedAt = createdAt
			}

			f = func(p redis.Pipeliner) error {
				ttnredis.SetProto(p, k, pb, 0)
				return nil
			}
		}

		cmds, err := tx.Pipelined(f)
		if err != nil {
			return err
		}
		for _, cmd := range cmds {
			if err := cmd.Err(); err != nil {
				return err
			}
		}
		return nil
	}, k)
	if err != nil {
		return nil, err
	}
	return pb, nil
}

type KeyRegistry struct {
	Redis *ttnredis.Client
}

func (r *KeyRegistry) GetByID(ctx context.Context, devEUI types.EUI64, id string) (*ttnpb.SessionKeys, error) {
	if devEUI.IsZero() || id == "" {
		return nil, errInvalidIdentifiers
	}

	pb := &ttnpb.SessionKeys{}
	if err := ttnredis.GetProto(r.Redis, r.Redis.Key(devEUI.String(), id)).ScanProto(pb); err != nil {
		return nil, err
	}
	return pb, nil
}

func (r *KeyRegistry) SetByID(ctx context.Context, devEUI types.EUI64, id string, f func(*ttnpb.SessionKeys) (*ttnpb.SessionKeys, error)) (*ttnpb.SessionKeys, error) {
	if devEUI.IsZero() || id == "" {
		return nil, errInvalidIdentifiers
	}

	k := r.Redis.Key(devEUI.String(), id)

	var pb *ttnpb.SessionKeys
	err := r.Redis.Watch(func(tx *redis.Tx) error {
		pb = &ttnpb.SessionKeys{}
		if err := ttnredis.GetProto(tx, k).ScanProto(pb); errors.IsNotFound(err) {
			pb = nil
		} else if err != nil {
			return err
		}

		var err error
		pb, err = f(pb)
		if err != nil {
			return err
		}

		var f func(redis.Pipeliner) error
		if pb == nil {
			f = func(p redis.Pipeliner) error {
				p.Del(k)
				return nil
			}
		} else {
			f = func(p redis.Pipeliner) error {
				ttnredis.SetProto(p, k, pb, 0)
				return nil
			}
		}

		cmds, err := tx.Pipelined(f)
		if err != nil {
			return err
		}
		for _, cmd := range cmds {
			if err := cmd.Err(); err != nil {
				return err
			}
		}
		return nil
	}, k)
	if err != nil {
		return nil, err
	}
	return pb, nil
}
