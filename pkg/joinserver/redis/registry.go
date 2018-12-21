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
	"go.thethings.network/lorawan-stack/pkg/errors"
	ttnredis "go.thethings.network/lorawan-stack/pkg/redis"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

var (
	errInvalidIdentifiers   = errors.DefineInvalidArgument("invalid_identifiers", "invalid identifiers")
	errDuplicateIdentifiers = errors.DefineAlreadyExists("duplicate_identifiers", "duplicate identifiers")
)

func applyDeviceFieldMask(dst, src *ttnpb.EndDevice, paths ...string) (*ttnpb.EndDevice, error) {
	if dst == nil {
		dst = &ttnpb.EndDevice{}
	}
	if err := dst.SetFields(src, append(paths, "ids")...); err != nil {
		return nil, err
	}
	if err := dst.EndDeviceIdentifiers.Validate(); err != nil {
		return nil, err
	}
	return dst, nil
}

type DeviceRegistry struct {
	Redis *ttnredis.Client
}

func (r *DeviceRegistry) GetByEUI(ctx context.Context, joinEUI types.EUI64, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error) {
	if joinEUI.IsZero() || devEUI.IsZero() {
		return nil, errInvalidIdentifiers
	}

	pb := &ttnpb.EndDevice{}
	if err := ttnredis.GetProto(r.Redis, r.Redis.Key(joinEUI.String(), devEUI.String())).ScanProto(pb); err != nil {
		return nil, err
	}
	return applyDeviceFieldMask(&ttnpb.EndDevice{}, pb, paths...)
}

func (r *DeviceRegistry) SetByEUI(ctx context.Context, joinEUI types.EUI64, devEUI types.EUI64, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
	if joinEUI.IsZero() || devEUI.IsZero() {
		return nil, errInvalidIdentifiers
	}

	k := r.Redis.Key(joinEUI.String(), devEUI.String())

	var pb *ttnpb.EndDevice
	err := r.Redis.Watch(func(tx *redis.Tx) error {
		var create bool
		cmd := ttnredis.GetProto(tx, k)
		stored := &ttnpb.EndDevice{}
		if err := cmd.ScanProto(stored); errors.IsNotFound(err) {
			create = true
			stored = nil
		} else if err != nil {
			return err
		}

		var err error
		if stored != nil {
			pb, err = applyDeviceFieldMask(nil, stored, gets...)
			if err != nil {
				return err
			}
		}

		var sets []string
		pb, sets, err = f(pb)
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
			pb.JoinEUI = &joinEUI
			pb.DevEUI = &devEUI
			pb.UpdatedAt = time.Now().UTC()
			sets = append(sets, "updated_at")
			if create {
				pb.CreatedAt = pb.UpdatedAt
				sets = append(sets, "created_at")
			}
			stored = &ttnpb.EndDevice{}
			if err := cmd.ScanProto(stored); err != nil && !errors.IsNotFound(err) {
				return err
			}
			stored, err = applyDeviceFieldMask(stored, pb, sets...)
			if err != nil {
				return err
			}
			pb, err = applyDeviceFieldMask(nil, stored, gets...)
			if err != nil {
				return err
			}
			f = func(p redis.Pipeliner) error {
				_, err := ttnredis.SetProto(p, k, stored, 0)
				return err
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

func applyKeyFieldMask(dst, src *ttnpb.SessionKeys, paths ...string) (*ttnpb.SessionKeys, error) {
	if dst == nil {
		dst = &ttnpb.SessionKeys{}
	}
	return dst, dst.SetFields(src, append(paths, "session_key_id")...)
}

type KeyRegistry struct {
	Redis *ttnredis.Client
}

func (r *KeyRegistry) GetByID(ctx context.Context, devEUI types.EUI64, id string, paths []string) (*ttnpb.SessionKeys, error) {
	if devEUI.IsZero() || id == "" {
		return nil, errInvalidIdentifiers
	}

	pb := &ttnpb.SessionKeys{}
	if err := ttnredis.GetProto(r.Redis, r.Redis.Key(devEUI.String(), id)).ScanProto(pb); err != nil {
		return nil, err
	}
	return applyKeyFieldMask(&ttnpb.SessionKeys{}, pb, paths...)
}

func (r *KeyRegistry) SetByID(ctx context.Context, devEUI types.EUI64, id string, gets []string, f func(*ttnpb.SessionKeys) (*ttnpb.SessionKeys, []string, error)) (*ttnpb.SessionKeys, error) {
	if devEUI.IsZero() || id == "" {
		return nil, errInvalidIdentifiers
	}

	k := r.Redis.Key(devEUI.String(), id)

	var pb *ttnpb.SessionKeys
	err := r.Redis.Watch(func(tx *redis.Tx) error {
		cmd := ttnredis.GetProto(tx, k)
		stored := &ttnpb.SessionKeys{}
		if err := cmd.ScanProto(stored); errors.IsNotFound(err) {
			stored = nil
		} else if err != nil {
			return err
		}

		var err error
		if pb != nil {
			pb, err = applyKeyFieldMask(nil, stored, gets...)
			if err != nil {
				return err
			}
		}

		var sets []string
		pb, sets, err = f(pb)
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
			stored = &ttnpb.SessionKeys{}
			if err := cmd.ScanProto(stored); err != nil && !errors.IsNotFound(err) {
				return err
			}
			stored, err = applyKeyFieldMask(stored, pb, sets...)
			if err != nil {
				return err
			}
			pb, err = applyKeyFieldMask(nil, stored, gets...)
			if err != nil {
				return err
			}
			f = func(p redis.Pipeliner) error {
				_, err := ttnredis.SetProto(p, k, stored, 0)
				return err
			}
		}

		_, err = tx.Pipelined(f)
		return err
	}, k)
	if err != nil {
		return nil, err
	}
	return pb, nil
}
