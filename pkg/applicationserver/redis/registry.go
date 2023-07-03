// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

// Package redis provides interfaces to Redis.
package redis

import (
	"bytes"
	"context"
	"runtime/trace"
	"time"

	"github.com/redis/go-redis/v9"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/internal/registry"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	errInvalidFieldmask   = errors.DefineInvalidArgument("invalid_fieldmask", "invalid fieldmask")
	errInvalidIdentifiers = errors.DefineInvalidArgument("invalid_identifiers", "invalid identifiers")
	errReadOnlyField      = errors.DefineInvalidArgument("read_only_field", "read-only field `{field}`")
)

// SchemaVersion is the Application Server database schema version. Bump when a migration is required.
const SchemaVersion = 1

// DeviceRegistry is a Redis device registry.
type DeviceRegistry struct {
	Redis   *ttnredis.Client
	LockTTL time.Duration
}

// Init initializes the DeviceRegistry.
func (r *DeviceRegistry) Init(ctx context.Context) error {
	if err := ttnredis.InitMutex(ctx, r.Redis); err != nil {
		return err
	}
	return nil
}

func (r *DeviceRegistry) uidKey(uid string) string {
	return r.Redis.Key("uid", uid)
}

func (r *DeviceRegistry) euiKey(joinEUI, devEUI types.EUI64) string {
	return r.Redis.Key("eui", devEUI.String(), joinEUI.String())
}

// Get returns the end device by its identifiers.
func (r *DeviceRegistry) Get(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, paths []string) (*ttnpb.EndDevice, error) {
	if err := ids.ValidateContext(ctx); err != nil {
		return nil, err
	}

	defer trace.StartRegion(ctx, "get end device").End()

	pb := &ttnpb.EndDevice{}
	if err := ttnredis.GetProto(ctx, r.Redis, r.uidKey(unique.ID(ctx, ids))).ScanProto(pb); err != nil {
		return nil, err
	}
	return ttnpb.FilterGetEndDevice(pb, paths...)
}

func equalEUI64(x, y *types.EUI64) bool {
	if x == nil || y == nil {
		return x == y
	}
	return x.Equal(*y)
}

// Set creates, updates or deletes the end device by its identifiers.
func (r *DeviceRegistry) Set(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
	if err := ids.ValidateContext(ctx); err != nil {
		return nil, err
	}
	uid := unique.ID(ctx, ids)
	uk := r.uidKey(uid)

	lockerID, err := ttnredis.GenerateLockerID()
	if err != nil {
		return nil, err
	}

	defer trace.StartRegion(ctx, "set end device").End()

	var pb *ttnpb.EndDevice
	err = ttnredis.LockedWatch(ctx, r.Redis, uk, lockerID, r.LockTTL, func(tx *redis.Tx) error {
		cmd := ttnredis.GetProto(ctx, tx, uk)
		stored := &ttnpb.EndDevice{}
		if err := cmd.ScanProto(stored); errors.IsNotFound(err) {
			stored = nil
		} else if err != nil {
			return err
		}

		var err error
		if stored != nil {
			pb = &ttnpb.EndDevice{}
			if err := cmd.ScanProto(pb); err != nil {
				return err
			}
			pb, err = ttnpb.FilterGetEndDevice(pb, gets...)
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
			pb, err = ttnpb.FilterGetEndDevice(stored, gets...)
			return err
		}

		var pipelined func(redis.Pipeliner) error
		if pb == nil && len(sets) == 0 {
			pipelined = func(p redis.Pipeliner) error {
				p.Del(ctx, uk)
				if stored.Ids.JoinEui != nil && stored.Ids.DevEui != nil {
					p.Del(
						ctx,
						r.euiKey(
							types.MustEUI64(stored.Ids.JoinEui).OrZero(),
							types.MustEUI64(stored.Ids.DevEui).OrZero(),
						),
					)
				}
				return nil
			}
		} else {
			if pb == nil {
				pb = &ttnpb.EndDevice{}
			}

			if pb.Ids.ApplicationIds.ApplicationId != ids.ApplicationIds.ApplicationId || pb.Ids.DeviceId != ids.DeviceId {
				return errInvalidIdentifiers.New()
			}

			pb.UpdatedAt = timestamppb.Now()
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

				updated, err = ttnpb.ApplyEndDeviceFieldMask(updated, pb, sets...)
				if err != nil {
					return err
				}
				if updated.Ids.ApplicationIds.ApplicationId != ids.ApplicationIds.ApplicationId || updated.Ids.DeviceId != ids.DeviceId {
					return errInvalidIdentifiers.New()
				}
			} else {
				if ttnpb.HasAnyField(sets, "ids.application_ids.application_id") && pb.Ids.ApplicationIds.ApplicationId != stored.Ids.ApplicationIds.ApplicationId {
					return errReadOnlyField.WithAttributes("field", "ids.application_ids.application_id")
				}
				if ttnpb.HasAnyField(sets, "ids.device_id") && pb.Ids.DeviceId != stored.Ids.DeviceId {
					return errReadOnlyField.WithAttributes("field", "ids.device_id")
				}
				if ttnpb.HasAnyField(sets, "ids.join_eui") && !bytes.Equal(pb.Ids.JoinEui, stored.Ids.JoinEui) {
					return errReadOnlyField.WithAttributes("field", "ids.join_eui")
				}
				if ttnpb.HasAnyField(sets, "ids.dev_eui") && !bytes.Equal(pb.Ids.DevEui, stored.Ids.DevEui) {
					return errReadOnlyField.WithAttributes("field", "ids.dev_eui")
				}
				if err := cmd.ScanProto(updated); err != nil {
					return err
				}
				updated, err = ttnpb.ApplyEndDeviceFieldMask(updated, pb, sets...)
				if err != nil {
					return err
				}
			}
			if err := updated.ValidateFields(); err != nil {
				return err
			}

			pipelined = func(p redis.Pipeliner) error {
				if stored == nil && updated.Ids.JoinEui != nil && updated.Ids.DevEui != nil {
					joinEUI := types.MustEUI64(updated.Ids.JoinEui).OrZero()
					devEUI := types.MustEUI64(updated.Ids.DevEui).OrZero()
					ek := r.euiKey(joinEUI, devEUI)
					if err := tx.Watch(ctx, ek).Err(); err != nil {
						return err
					}

					storedUIDStr, err := tx.Get(ctx, ek).Result()
					if errors.Is(err, redis.Nil) {
						p.SetNX(ctx, ek, uid, 0)
					} else if err != nil {
						return err
					} else {
						return registry.UniqueEUIViolationErr(ctx, joinEUI, devEUI, storedUIDStr)
					}
				}

				if _, err := ttnredis.SetProto(ctx, p, uk, updated, 0); err != nil {
					return err
				}
				return nil
			}
			pb, err = ttnpb.FilterGetEndDevice(updated, gets...)
			if err != nil {
				return err
			}
		}
		_, err = tx.TxPipelined(ctx, pipelined)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, ttnredis.ConvertError(err)
	}
	return pb, nil
}

// Range ranges over the end devices and calls the callback function, until false is returned.
func (r *DeviceRegistry) Range(ctx context.Context, paths []string, f func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.EndDevice) bool) error {
	deviceEntityRegex, err := ttnredis.EntityRegex((r.uidKey(unique.GenericID(ctx, "*"))))
	if err != nil {
		return err
	}
	return ttnredis.RangeRedisKeys(ctx, r.Redis, r.uidKey(unique.GenericID(ctx, "*")), ttnredis.DefaultRangeCount, func(key string) (bool, error) {
		if !deviceEntityRegex.MatchString(key) {
			return true, nil
		}
		dev := &ttnpb.EndDevice{}
		if err := ttnredis.GetProto(ctx, r.Redis, key).ScanProto(dev); err != nil {
			return false, err
		}
		dev, err := ttnpb.FilterGetEndDevice(dev, paths...)
		if err != nil {
			return false, err
		}
		if !f(ctx, dev.Ids, dev) {
			return false, nil
		}
		return true, nil
	})
}

func applyLinkFieldMask(dst, src *ttnpb.ApplicationLink, paths ...string) (*ttnpb.ApplicationLink, error) {
	if dst == nil {
		dst = &ttnpb.ApplicationLink{}
	}
	return dst, dst.SetFields(src, paths...)
}

// LinkRegistry is a store for application links.
type LinkRegistry struct {
	Redis   *ttnredis.Client
	LockTTL time.Duration
}

// Init initializes the LinkRegistry.
func (r *LinkRegistry) Init(ctx context.Context) error {
	if err := ttnredis.InitMutex(ctx, r.Redis); err != nil {
		return err
	}
	return nil
}

func (r *LinkRegistry) allKey(ctx context.Context) string {
	return r.Redis.Key("all")
}

func (r *LinkRegistry) appKey(uid string) string {
	return r.Redis.Key("uid", uid)
}

// Get returns the link by the application identifiers.
func (r *LinkRegistry) Get(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, paths []string) (*ttnpb.ApplicationLink, error) {
	defer trace.StartRegion(ctx, "get link").End()

	pb := &ttnpb.ApplicationLink{}
	if err := ttnredis.GetProto(ctx, r.Redis, r.appKey(unique.ID(ctx, ids))).ScanProto(pb); err != nil {
		return nil, err
	}
	return applyLinkFieldMask(nil, pb, paths...)
}

var errApplicationUID = errors.DefineCorruption("application_uid", "invalid application UID `{application_uid}`")

// Range ranges the links and calls the callback function, until false is returned.
func (r *LinkRegistry) Range(ctx context.Context, paths []string, f func(context.Context, *ttnpb.ApplicationIdentifiers, *ttnpb.ApplicationLink) bool) error {
	defer trace.StartRegion(ctx, "range links").End()

	uids, err := r.Redis.SMembers(ctx, r.allKey(ctx)).Result()
	if err != nil {
		return ttnredis.ConvertError(err)
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
		if err := ttnredis.GetProto(ctx, r.Redis, r.appKey(uid)).ScanProto(pb); err != nil {
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
func (r *LinkRegistry) Set(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, gets []string, f func(*ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, []string, error)) (*ttnpb.ApplicationLink, error) {
	uid := unique.ID(ctx, ids)
	uk := r.appKey(uid)

	lockerID, err := ttnredis.GenerateLockerID()
	if err != nil {
		return nil, err
	}

	defer trace.StartRegion(ctx, "set link").End()

	var pb *ttnpb.ApplicationLink
	err = ttnredis.LockedWatch(ctx, r.Redis, uk, lockerID, r.LockTTL, func(tx *redis.Tx) error {
		cmd := ttnredis.GetProto(ctx, tx, uk)
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
				p.Del(ctx, uk)
				p.SRem(ctx, r.allKey(ctx), uid)
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

			if err := updated.ValidateFields(); err != nil {
				return err
			}

			pipelined = func(p redis.Pipeliner) error {
				_, err := ttnredis.SetProto(ctx, p, uk, updated, 0)
				if err != nil {
					return err
				}
				p.SAdd(ctx, r.allKey(ctx), uid)
				return nil
			}
			pb, err = applyLinkFieldMask(nil, updated, gets...)
			if err != nil {
				return err
			}
		}
		_, err = tx.TxPipelined(ctx, pipelined)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, ttnredis.ConvertError(err)
	}
	return pb, nil
}

// BatchDelete implements Registry.
func (r *DeviceRegistry) BatchDelete(
	ctx context.Context,
	appIDs *ttnpb.ApplicationIdentifiers,
	deviceIDs []string,
) ([]*ttnpb.EndDeviceIdentifiers, error) {
	var (
		uidKeys = make([]string, 0, len(deviceIDs))
		ret     = make([]*ttnpb.EndDeviceIdentifiers, 0)
	)
	for _, devID := range deviceIDs {
		uidKeys = append(
			uidKeys,
			r.uidKey(
				unique.ID(
					ctx,
					&ttnpb.EndDeviceIdentifiers{
						ApplicationIds: appIDs,
						DeviceId:       devID,
					}),
			),
		)
	}

	// Delete the devices in a single transaction.
	transaction := func(tx *redis.Tx) error {
		// Read the devices to delete.
		raw, err := tx.MGet(ctx, uidKeys...).Result()
		if err != nil {
			return err
		}
		euiKeys := make([]string, 0, len(raw))
		for _, raw := range raw {
			switch val := raw.(type) {
			case nil:
				continue
			case string:
				dev := &ttnpb.EndDevice{}
				if err := ttnredis.UnmarshalProto(val, dev); err != nil {
					log.FromContext(ctx).WithError(err).Warn("Failed to decode stored end device")
					continue
				}
				ret = append(ret, dev.Ids)
				if dev.Ids.JoinEui != nil && dev.Ids.DevEui != nil {
					euiKeys = append(euiKeys, r.euiKey(
						types.MustEUI64(dev.GetIds().GetJoinEui()).OrZero(),
						types.MustEUI64(dev.GetIds().GetDevEui()).OrZero(),
					))
				}
			}
		}
		if err := tx.Watch(ctx, euiKeys...).Err(); err != nil {
			return err
		}
		if _, err := tx.TxPipelined(ctx, func(p redis.Pipeliner) error {
			p.Del(ctx, append(uidKeys, euiKeys...)...)
			return nil
		}); err != nil {
			return err
		}
		return nil
	}

	defer trace.StartRegion(ctx, "batch delete end device").End()
	err := r.Redis.Watch(ctx, transaction, uidKeys...)
	if err != nil {
		return nil, ttnredis.ConvertError(err)
	}
	return ret, nil
}
