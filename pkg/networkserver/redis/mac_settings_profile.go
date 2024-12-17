// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

// Package redis implements a Redis-backed MAC settings profile registry.
package redis

import (
	"context"
	"runtime/trace"

	"github.com/redis/go-redis/v9"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"google.golang.org/protobuf/proto"
)

// MACSettingsProfileRegistry is an implementation of networkserver.MACSettingsProfileRegistry.
type MACSettingsProfileRegistry struct {
	Redis   *ttnredis.Client
	LockTTL time.Duration
}

func applyMACSettingsProfileFieldMask(dst, src *ttnpb.MACSettingsProfile, paths ...string,
) (*ttnpb.MACSettingsProfile, error) {
	if dst == nil {
		dst = &ttnpb.MACSettingsProfile{}
	}
	return dst, dst.SetFields(src, paths...)
}

func (r *MACSettingsProfileRegistry) appKey(uid string) string {
	return r.Redis.Key("uid", uid)
}

func (r *MACSettingsProfileRegistry) profileKey(appUID string, id string) string {
	return r.Redis.Key("uid", appUID, id)
}

func (r *MACSettingsProfileRegistry) makeProfileKeyFunc(appUID string) func(string) string {
	return func(id string) string {
		return r.profileKey(appUID, id)
	}
}

// Init initializes the MAC settings profile registry.
func (r *MACSettingsProfileRegistry) Init(ctx context.Context) error {
	return ttnredis.InitMutex(ctx, r.Redis)
}

// Get gets the MAC settings profile by identifiers.
func (r *MACSettingsProfileRegistry) Get(
	ctx context.Context,
	ids *ttnpb.MACSettingsProfileIdentifiers,
	paths []string,
) (*ttnpb.MACSettingsProfile, error) {
	defer trace.StartRegion(ctx, "get mac settings profile").End()

	if err := ids.ValidateContext(ctx); err != nil {
		return nil, err
	}

	pb := &ttnpb.MACSettingsProfile{}
	appUID := unique.ID(ctx, ids.ApplicationIds)
	if err := ttnredis.GetProto(ctx, r.Redis, r.profileKey(appUID, ids.ProfileId)).ScanProto(pb); err != nil {
		return nil, err
	}
	pb, err := applyMACSettingsProfileFieldMask(nil, pb, paths...)
	if err != nil {
		return nil, err
	}
	return pb, nil
}

// Set sets the MAC settings profile by identifiers.
func (r *MACSettingsProfileRegistry) Set( //nolint:gocyclo
	ctx context.Context,
	ids *ttnpb.MACSettingsProfileIdentifiers,
	paths []string,
	f func(context.Context, *ttnpb.MACSettingsProfile) (*ttnpb.MACSettingsProfile, []string, error),
) (*ttnpb.MACSettingsProfile, error) {
	defer trace.StartRegion(ctx, "set mac settings profile").End()

	if err := ids.ValidateContext(ctx); err != nil {
		return nil, err
	}

	appUID := unique.ID(ctx, ids.ApplicationIds)
	pk := r.profileKey(appUID, ids.ProfileId)

	lockerID, err := ttnredis.GenerateLockerID()
	if err != nil {
		return nil, err
	}

	var pb *ttnpb.MACSettingsProfile
	err = ttnredis.LockedWatch(ctx, r.Redis, pk, lockerID, r.LockTTL, func(tx *redis.Tx) error {
		cmd := ttnredis.GetProto(ctx, tx, pk)
		stored := &ttnpb.MACSettingsProfile{}
		if err := cmd.ScanProto(stored); errors.IsNotFound(err) {
			stored = nil
		} else if err != nil {
			return err
		}

		var err error
		if stored != nil {
			pb = &ttnpb.MACSettingsProfile{}
			if err := cmd.ScanProto(pb); err != nil {
				return err
			}
			pb, err = applyMACSettingsProfileFieldMask(nil, pb, paths...)
			if err != nil {
				return err
			}
		}

		var sets []string
		pb, sets, err = f(ctx, pb)
		if err != nil {
			return err
		}
		if stored == nil && pb == nil {
			return nil
		}
		if pb != nil && len(sets) == 0 {
			pb, err = applyMACSettingsProfileFieldMask(nil, stored, paths...)
			return err
		}

		var pipelined func(redis.Pipeliner) error
		if pb == nil && len(sets) == 0 {
			pipelined = func(p redis.Pipeliner) error {
				p.Del(ctx, pk)
				p.SRem(ctx, r.appKey(appUID), stored.Ids.ProfileId)
				return nil
			}
		} else {
			if pb == nil {
				pb = &ttnpb.MACSettingsProfile{}
			}

			if pb.Ids.ApplicationIds.ApplicationId != ids.ApplicationIds.ApplicationId || pb.Ids.ProfileId != ids.ProfileId {
				return errInvalidIdentifiers.New()
			}
			updated := &ttnpb.MACSettingsProfile{}
			if stored == nil {
				if err := ttnpb.RequireFields(sets,
					"ids.application_ids",
					"ids.profile_id",
				); err != nil {
					return errInvalidFieldmask.WithCause(err)
				}
				updated, err = applyMACSettingsProfileFieldMask(updated, pb, sets...)
				if err != nil {
					return err
				}
				if updated.Ids.ApplicationIds.ApplicationId != ids.ApplicationIds.ApplicationId ||
					updated.Ids.ProfileId != ids.ProfileId {
					return errInvalidIdentifiers.New()
				}
			} else {
				if ttnpb.HasAnyField(sets, "ids.application_ids.application_id") &&
					pb.Ids.ApplicationIds.ApplicationId != stored.Ids.ApplicationIds.ApplicationId {
					return errReadOnlyField.WithAttributes("field", "ids.application_ids.application_id")
				}
				if ttnpb.HasAnyField(sets, "ids.profile_id") && pb.Ids.ProfileId != stored.Ids.ProfileId {
					return errReadOnlyField.WithAttributes("field", "ids.profile_id")
				}
				if err := cmd.ScanProto(updated); err != nil {
					return err
				}
				updated, err = applyMACSettingsProfileFieldMask(updated, pb, sets...)
				if err != nil {
					return err
				}
			}
			if err := updated.ValidateFields(); err != nil {
				return err
			}
			pipelined = func(p redis.Pipeliner) error {
				if _, err := ttnredis.SetProto(ctx, p, pk, updated, 0); err != nil {
					return err
				}
				p.SAdd(ctx, r.appKey(appUID), updated.Ids.ProfileId)
				return nil
			}
			pb, err = applyMACSettingsProfileFieldMask(nil, updated, paths...)
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

// List lists MAC settings profiles by application identifiers.
func (r *MACSettingsProfileRegistry) List(
	ctx context.Context,
	ids *ttnpb.ApplicationIdentifiers,
	paths []string,
) ([]*ttnpb.MACSettingsProfile, error) {
	defer trace.StartRegion(ctx, "list mac settings profile").End()

	if err := ids.ValidateContext(ctx); err != nil {
		return nil, err
	}

	var pbs []*ttnpb.MACSettingsProfile
	appUID := unique.ID(ctx, ids)
	uidKey := r.appKey(appUID)

	opts := []ttnredis.FindProtosOption{}
	limit, offset := ttnredis.PaginationLimitAndOffsetFromContext(ctx)
	if limit != 0 {
		opts = append(opts,
			ttnredis.FindProtosSorted(true),
			ttnredis.FindProtosWithOffsetAndCount(offset, limit),
		)
	}

	rangeProtos := func(c redis.Cmdable) error {
		return ttnredis.FindProtos(
			ctx,
			c,
			uidKey,
			r.makeProfileKeyFunc(appUID),
			opts...,
		).Range(func() (proto.Message, func() (bool, error)) {
			pb := &ttnpb.MACSettingsProfile{}
			return pb, func() (bool, error) {
				pb, err := applyMACSettingsProfileFieldMask(nil, pb, paths...)
				if err != nil {
					return false, err
				}
				pbs = append(pbs, pb)
				return true, nil
			}
		})
	}

	var err error
	if limit != 0 {
		var lockerID string
		lockerID, err = ttnredis.GenerateLockerID()
		if err != nil {
			return nil, err
		}
		err = ttnredis.LockedWatch(ctx, r.Redis, uidKey, lockerID, r.LockTTL, func(tx *redis.Tx) (err error) {
			total, err := tx.SCard(ctx, uidKey).Result()
			if err != nil {
				return err
			}
			ttnredis.SetPaginationTotal(ctx, total)
			return rangeProtos(tx)
		})
	} else {
		err = rangeProtos(r.Redis)
	}
	if err != nil {
		return nil, ttnredis.ConvertError(err)
	}
	return pbs, nil
}

// WithPagination returns a new context with pagination parameters.
func (*MACSettingsProfileRegistry) WithPagination(
	ctx context.Context,
	limit uint32,
	page uint32,
	total *int64,
) context.Context {
	return ttnredis.NewContextWithPagination(ctx, int64(limit), int64(page), total)
}
