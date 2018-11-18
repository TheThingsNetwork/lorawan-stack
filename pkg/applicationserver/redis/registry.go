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
	"go.thethings.network/lorawan-stack/pkg/unique"
)

// DeviceRegistry is a Redis device registry.
type DeviceRegistry struct {
	Redis *ttnredis.Client
}

// Get returns the end device by its identifiers.
func (r *DeviceRegistry) Get(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, paths []string) (*ttnpb.EndDevice, error) {
	k := r.Redis.Key(unique.ID(ctx, ids))
	pb := &ttnpb.EndDevice{}
	if err := ttnredis.GetProto(r.Redis, k).ScanProto(pb); err != nil {
		return nil, err
	}
	// TODO: Apply field mask once generator is ready (https://github.com/TheThingsIndustries/lorawan-stack/issues/1212)
	return pb, nil
}

// Set creates, updates or deletes the end device by its identifiers.
func (r *DeviceRegistry) Set(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
	k := r.Redis.Key(unique.ID(ctx, ids))
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
		// TODO: Apply field mask once generator is ready (https://github.com/TheThingsIndustries/lorawan-stack/issues/1212)
		createdAt := pb.GetCreatedAt()
		var err error
		pb, _, err = f(pb)
		if err != nil {
			return err
		}
		// TODO: Apply field mask once generator is ready (https://github.com/TheThingsIndustries/lorawan-stack/issues/1212)
		var f func(redis.Pipeliner) error
		if pb == nil {
			f = func(p redis.Pipeliner) error {
				p.Del(k)
				return nil
			}
		} else {
			pb.EndDeviceIdentifiers = ids
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

// LinkRegistry is a store for application links.
type LinkRegistry struct {
	Redis *ttnredis.Client
}

const (
	allKey  = "all"
	linkKey = "link"
)

// Get returns the link by the application identifiers.
func (r *LinkRegistry) Get(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) (*ttnpb.ApplicationLink, error) {
	k := r.Redis.Key(linkKey, unique.ID(ctx, ids))
	pb := &ttnpb.ApplicationLink{}
	if err := ttnredis.GetProto(r.Redis, k).ScanProto(pb); err != nil {
		return nil, err
	}
	// TODO: Apply field mask once generator is ready (https://github.com/TheThingsIndustries/lorawan-stack/issues/1212)
	return pb, nil
}

var errApplicationUID = errors.DefineCorruption("application_uid", "invalid application UID `{application_uid}`")

// Range ranges the links and calls the callback function, until false is returned.
func (r *LinkRegistry) Range(ctx context.Context, paths []string, f func(context.Context, ttnpb.ApplicationIdentifiers, *ttnpb.ApplicationLink) bool) error {
	uids, err := r.Redis.SMembers(r.Redis.Key(allKey)).Result()
	if err != nil {
		return err
	}
	for _, uid := range uids {
		ctx, err := unique.WithContext(ctx, uid)
		if err != nil {
			return errApplicationUID.WithCause(err).WithAttributes("application_uid", uid)
		}
		ids, err := unique.ToApplicationID(uid)
		if err != nil {
			return errApplicationUID.WithCause(err).WithAttributes("application_uid", uid)
		}
		pb := &ttnpb.ApplicationLink{}
		if err := ttnredis.GetProto(r.Redis, r.Redis.Key(linkKey, uid)).ScanProto(pb); err != nil {
			return err
		}
		// TODO: Apply field mask once generator is ready (https://github.com/TheThingsIndustries/lorawan-stack/issues/1212)
		if !f(ctx, ids, pb) {
			break
		}
	}
	return nil
}

// Set creates, updates or deletes the link by the application identifiers.
func (r *LinkRegistry) Set(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string, f func(*ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, []string, error)) (*ttnpb.ApplicationLink, error) {
	uid := unique.ID(ctx, ids)
	k := r.Redis.Key(linkKey, uid)
	var pb *ttnpb.ApplicationLink
	err := r.Redis.Watch(func(tx *redis.Tx) error {
		pb = &ttnpb.ApplicationLink{}
		if err := ttnredis.GetProto(tx, k).ScanProto(pb); errors.IsNotFound(err) {
			pb = nil
		} else if err != nil {
			return err
		}
		// TODO: Apply field mask once generator is ready (https://github.com/TheThingsIndustries/lorawan-stack/issues/1212)
		var err error
		pb, _, err = f(pb)
		if err != nil {
			return err
		}
		// TODO: Apply field mask once generator is ready (https://github.com/TheThingsIndustries/lorawan-stack/issues/1212)
		var f func(redis.Pipeliner) error
		if pb == nil {
			f = func(p redis.Pipeliner) error {
				p.Del(k)
				p.SRem(r.Redis.Key(allKey), uid)
				return nil
			}
		} else {
			f = func(p redis.Pipeliner) error {
				ttnredis.SetProto(p, k, pb, 0)
				p.SAdd(r.Redis.Key(allKey), uid)
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
