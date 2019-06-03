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

package redis

import (
	"context"
	"runtime/trace"
	"time"

	"github.com/go-redis/redis"
	"go.thethings.network/lorawan-stack/pkg/errors"
	ttnredis "go.thethings.network/lorawan-stack/pkg/redis"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

var (
	errInvalidFieldmask     = errors.DefineInvalidArgument("invalid_fieldmask", "invalid fieldmask")
	errInvalidIdentifiers   = errors.DefineInvalidArgument("invalid_identifiers", "invalid identifiers")
	errDuplicateIdentifiers = errors.DefineAlreadyExists("duplicate_identifiers", "duplicate identifiers")
)

// appendImplicitDeviceGetPaths appends implicit ttnpb.EndDevice get paths to paths.
func appendImplicitDeviceGetPaths(paths ...string) []string {
	return append(append(make([]string, 0, 3+len(paths)),
		"created_at",
		"ids",
		"updated_at",
	), paths...)
}

func applyDeviceFieldMask(dst, src *ttnpb.EndDevice, paths ...string) (*ttnpb.EndDevice, error) {
	if dst == nil {
		dst = &ttnpb.EndDevice{}
	}
	return dst, dst.SetFields(src, paths...)
}

// DeviceRegistry is a Redis device registry.
type DeviceRegistry struct {
	Redis *ttnredis.Client
}

func (r *DeviceRegistry) uidKey(uid string) string {
	return r.Redis.Key("uid", uid)
}

func (r *DeviceRegistry) euiKey(devEUI, joinEUI types.EUI64) string {
	return r.Redis.Key("eui", joinEUI.String(), devEUI.String())
}

// Get returns the end device by its identifiers.
func (r *DeviceRegistry) Get(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, paths []string) (*ttnpb.EndDevice, error) {
	if err := ids.ValidateContext(ctx); err != nil {
		return nil, err
	}

	defer trace.StartRegion(ctx, "get end device").End()

	pb := &ttnpb.EndDevice{}
	if err := ttnredis.GetProto(r.Redis, r.uidKey(unique.ID(ctx, ids))).ScanProto(pb); err != nil {
		return nil, err
	}
	return applyDeviceFieldMask(nil, pb, appendImplicitDeviceGetPaths(paths...)...)
}

// Set creates, updates or deletes the end device by its identifiers.
func (r *DeviceRegistry) Set(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
	if err := ids.ValidateContext(ctx); err != nil {
		return nil, err
	}
	uid := unique.ID(ctx, ids)
	uk := r.uidKey(uid)

	defer trace.StartRegion(ctx, "set end device").End()

	var pb *ttnpb.EndDevice
	err := r.Redis.Watch(func(tx *redis.Tx) error {
		cmd := ttnredis.GetProto(tx, uk)
		stored := &ttnpb.EndDevice{}
		if err := cmd.ScanProto(stored); errors.IsNotFound(err) {
			stored = nil
		} else if err != nil {
			return err
		}

		gets = appendImplicitDeviceGetPaths(gets...)

		var err error
		if stored != nil {
			pb = &ttnpb.EndDevice{}
			if err := cmd.ScanProto(pb); err != nil {
				return err
			}
			pb, err = applyDeviceFieldMask(nil, pb, gets...)
			if err != nil {
				return err
			}
		}

		var sets []string
		pb, sets, err = f(pb)
		if err != nil {
			return err
		}
		if err := ttnpb.ProhibitFields(sets,
			"created_at",
			"updated_at",
		); err != nil {
			return errInvalidFieldmask.WithCause(err)
		}

		if stored == nil && pb == nil {
			return nil
		}
		if pb != nil && len(sets) == 0 {
			pb, err = applyDeviceFieldMask(nil, stored, gets...)
			return err
		}

		var pipelined func(redis.Pipeliner) error
		if pb == nil && len(sets) == 0 {
			pipelined = func(p redis.Pipeliner) error {
				p.Del(uk)
				if stored.JoinEUI != nil && stored.DevEUI != nil {
					p.Del(r.euiKey(*stored.JoinEUI, *stored.DevEUI))
				}
				return nil
			}
		} else {
			if pb == nil {
				pb = &ttnpb.EndDevice{}
			}

			if pb.ApplicationIdentifiers != ids.ApplicationIdentifiers || pb.DeviceID != ids.DeviceID {
				return errInvalidIdentifiers
			}

			pb.UpdatedAt = time.Now().UTC()
			sets = append(append(sets[:0:0], sets...),
				"updated_at",
			)

			updated := &ttnpb.EndDevice{}
			if stored == nil {
				if err := ttnpb.RequireFields(sets,
					"ids.application_ids",
					"ids.device_id",
				); err != nil {
					return errInvalidFieldmask.WithCause(err)
				}

				pb.CreatedAt = pb.UpdatedAt
				sets = append(sets, "created_at")

				updated, err = applyDeviceFieldMask(updated, pb, sets...)
				if err != nil {
					return err
				}
				if updated.ApplicationIdentifiers != ids.ApplicationIdentifiers || updated.DeviceID != ids.DeviceID {
					return errInvalidIdentifiers
				}
			} else {
				if err := ttnpb.ProhibitFields(sets,
					"ids.application_ids",
					"ids.dev_eui",
					"ids.device_id",
					"ids.join_eui",
				); err != nil {
					return errInvalidFieldmask.WithCause(err)
				}

				if err := cmd.ScanProto(updated); err != nil {
					return err
				}
				updated, err = applyDeviceFieldMask(updated, pb, sets...)
				if err != nil {
					return err
				}
			}
			if err := updated.ValidateFields(sets...); err != nil {
				return err
			}

			pipelined = func(p redis.Pipeliner) error {
				if stored == nil && updated.JoinEUI != nil && updated.DevEUI != nil {
					ek := r.euiKey(*updated.JoinEUI, *updated.DevEUI)
					if err := tx.Watch(ek).Err(); err != nil {
						return err
					}
					i, err := tx.Exists(ek).Result()
					if err != nil {
						return ttnredis.ConvertError(err)
					}
					if i != 0 {
						return errDuplicateIdentifiers
					}
					p.SetNX(ek, uid, 0)
				}

				if _, err := ttnredis.SetProto(p, uk, updated, 0); err != nil {
					return err
				}
				return nil
			}
			pb, err = applyDeviceFieldMask(nil, updated, gets...)
			if err != nil {
				return err
			}
		}
		_, err = tx.Pipelined(pipelined)
		if err != nil {
			return err
		}
		return nil
	}, uk)
	if err != nil {
		return nil, err
	}
	return pb, nil
}

func applyLinkFieldMask(dst, src *ttnpb.ApplicationLink, paths ...string) (*ttnpb.ApplicationLink, error) {
	if dst == nil {
		dst = &ttnpb.ApplicationLink{}
	}
	return dst, dst.SetFields(src, paths...)
}

// LinkRegistry is a store for application links.
type LinkRegistry struct {
	Redis *ttnredis.Client
}

func (r *LinkRegistry) allKey() string {
	return r.Redis.Key("all")
}

func (r *LinkRegistry) appKey(uid string) string {
	return r.Redis.Key("uid", uid)
}

// Get returns the link by the application identifiers.
func (r *LinkRegistry) Get(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) (*ttnpb.ApplicationLink, error) {
	defer trace.StartRegion(ctx, "get link").End()

	pb := &ttnpb.ApplicationLink{}
	if err := ttnredis.GetProto(r.Redis, r.appKey(unique.ID(ctx, ids))).ScanProto(pb); err != nil {
		return nil, err
	}
	return applyLinkFieldMask(nil, pb, paths...)
}

var errApplicationUID = errors.DefineCorruption("application_uid", "invalid application UID `{application_uid}`")

// Range ranges the links and calls the callback function, until false is returned.
func (r *LinkRegistry) Range(ctx context.Context, paths []string, f func(context.Context, ttnpb.ApplicationIdentifiers, *ttnpb.ApplicationLink) bool) error {
	defer trace.StartRegion(ctx, "range links").End()

	uids, err := r.Redis.SMembers(r.allKey()).Result()
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
		if err := ttnredis.GetProto(r.Redis, r.appKey(uid)).ScanProto(pb); err != nil {
			return err
		}
		pb, err = applyLinkFieldMask(nil, pb, paths...)
		if err != nil {
			return err
		}
		if !f(ctx, ids, pb) {
			break
		}
	}
	return nil
}

// Set creates, updates or deletes the link by the application identifiers.
func (r *LinkRegistry) Set(ctx context.Context, ids ttnpb.ApplicationIdentifiers, gets []string, f func(*ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, []string, error)) (*ttnpb.ApplicationLink, error) {
	defer trace.StartRegion(ctx, "set link").End()

	uid := unique.ID(ctx, ids)
	uk := r.appKey(uid)

	var pb *ttnpb.ApplicationLink
	err := r.Redis.Watch(func(tx *redis.Tx) error {
		cmd := ttnredis.GetProto(tx, uk)
		stored := &ttnpb.ApplicationLink{}
		if err := cmd.ScanProto(stored); errors.IsNotFound(err) {
			stored = nil
		} else if err != nil {
			return err
		}

		var err error
		if pb != nil {
			pb = &ttnpb.ApplicationLink{}
			if err := cmd.ScanProto(pb); err != nil {
				return err
			}
			pb, err = applyLinkFieldMask(nil, pb, gets...)
			if err != nil {
				return err
			}
		}

		var sets []string
		pb, sets, err = f(pb)
		if err != nil {
			return err
		}
		if stored == nil && pb == nil {
			return nil
		}
		if pb != nil && len(sets) == 0 {
			pb, err = applyLinkFieldMask(nil, stored, gets...)
			return err
		}

		var pipelined func(redis.Pipeliner) error
		if pb == nil && len(sets) == 0 {
			pipelined = func(p redis.Pipeliner) error {
				p.Del(uk)
				p.SRem(r.allKey(), uid)
				return nil
			}
		} else {
			if pb == nil {
				pb = &ttnpb.ApplicationLink{}
			}

			updated := &ttnpb.ApplicationLink{}
			if stored != nil {
				if err := cmd.ScanProto(updated); err != nil {
					return err
				}
			}
			updated, err = applyLinkFieldMask(updated, pb, sets...)
			if err != nil {
				return err
			}

			if err := updated.ValidateFields(sets...); err != nil {
				return err
			}

			pipelined = func(p redis.Pipeliner) error {
				_, err := ttnredis.SetProto(p, uk, updated, 0)
				if err != nil {
					return err
				}
				p.SAdd(r.allKey(), uid)
				return nil
			}
			pb, err = applyLinkFieldMask(nil, updated, gets...)
			if err != nil {
				return err
			}
		}
		_, err = tx.Pipelined(pipelined)
		if err != nil {
			return err
		}
		return nil
	}, uk)
	if err != nil {
		return nil, err
	}
	return pb, nil
}
