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

package redis

import (
	"context"
	"regexp"
	"runtime/trace"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	errInvalidFieldmask   = errors.DefineInvalidArgument("invalid_fieldmask", "invalid fieldmask")
	errInvalidIdentifiers = errors.DefineInvalidArgument("invalid_identifiers", "invalid identifiers")
	errReadOnlyField      = errors.DefineInvalidArgument("read_only_field", "read-only field `{field}`")
)

// appendImplicitAssociationGetPaths appends implicit ttnpb.ApplicationPackageAssociation get paths to paths.
func appendImplicitAssociationGetPaths(paths ...string) []string {
	return append(append(make([]string, 0, 4+len(paths)),
		"created_at",
		"ids",
		"package_name",
		"updated_at",
	), paths...)
}

func applyAssociationFieldMask(dst, src *ttnpb.ApplicationPackageAssociation, paths ...string) (*ttnpb.ApplicationPackageAssociation, error) {
	if dst == nil {
		dst = &ttnpb.ApplicationPackageAssociation{}
	}
	return dst, dst.SetFields(src, paths...)
}

func applyDefaultAssociationFieldMask(dst, src *ttnpb.ApplicationPackageDefaultAssociation, paths ...string) (*ttnpb.ApplicationPackageDefaultAssociation, error) {
	if dst == nil {
		dst = &ttnpb.ApplicationPackageDefaultAssociation{}
	}
	return dst, dst.SetFields(src, paths...)
}

// ApplicationPackagesRegistry is a Redis application packages registry.
type ApplicationPackagesRegistry struct {
	Redis   *ttnredis.Client
	LockTTL time.Duration
}

// NewApplicationPackagesRegistry creates, initializes and returns a new ApplicationPackagesRegistry.
func NewApplicationPackagesRegistry(
	ctx context.Context,
	cl *ttnredis.Client,
	lockTTL time.Duration,
) (packages.Registry, error) {
	reg := &ApplicationPackagesRegistry{
		Redis:   cl,
		LockTTL: lockTTL,
	}
	if err := reg.Init(ctx); err != nil {
		return nil, err
	}
	return reg, nil
}

// Init initializes the ApplicationPackagesRegistry.
func (r *ApplicationPackagesRegistry) Init(ctx context.Context) error {
	if err := ttnredis.InitMutex(ctx, r.Redis); err != nil {
		return err
	}
	return nil
}

func (r *ApplicationPackagesRegistry) uidKey(uid string) string {
	return r.Redis.Key("uid", uid)
}

func (r *ApplicationPackagesRegistry) fPortStr(fPort uint32) string {
	if fPort > 255 {
		panic("FPort cannot be higher than 255")
	}
	return strconv.FormatUint(uint64(fPort), 10)
}

func (r *ApplicationPackagesRegistry) associationKey(uid string, fPort string) string {
	return r.Redis.Key("uid", uid, fPort)
}

func (r *ApplicationPackagesRegistry) makeAssociationKeyFunc(uid string) func(port string) string {
	return func(port string) string {
		return r.associationKey(uid, port)
	}
}

func (r *ApplicationPackagesRegistry) transactionKey(uid string, fPort string, packageName string) string {
	return r.Redis.Key("transaction", uid, fPort, packageName)
}

func packagesRegex(uid string) (*regexp.Regexp, error) {
	keyRegex := strings.ReplaceAll(uid, ":", "\\:")
	keyRegex = strings.ReplaceAll(keyRegex, "*", ".[^\\:]*")
	keyRegex = keyRegex + "\\:\\d*$"
	return regexp.Compile(keyRegex)
}

// GetAssociation implements applicationpackages.AssociationRegistry.
func (r ApplicationPackagesRegistry) GetAssociation(ctx context.Context, ids *ttnpb.ApplicationPackageAssociationIdentifiers, paths []string) (*ttnpb.ApplicationPackageAssociation, error) {
	pb := &ttnpb.ApplicationPackageAssociation{}
	defer trace.StartRegion(ctx, "get application package association by id").End()
	if err := ttnredis.GetProto(ctx, r.Redis, r.associationKey(unique.ID(ctx, ids.EndDeviceIds), r.fPortStr(ids.FPort))).ScanProto(pb); err != nil {
		return nil, err
	}
	return applyAssociationFieldMask(nil, pb, appendImplicitAssociationGetPaths(paths...)...)
}

// ListAssociations implements applicationpackages.AssociationRegistry.
func (r ApplicationPackagesRegistry) ListAssociations(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, paths []string) ([]*ttnpb.ApplicationPackageAssociation, error) {
	var pbs []*ttnpb.ApplicationPackageAssociation
	devUID := unique.ID(ctx, ids)
	uidKey := r.uidKey(devUID)

	opts := []ttnredis.FindProtosOption{}
	limit, offset := ttnredis.PaginationLimitAndOffsetFromContext(ctx)
	if limit != 0 {
		opts = append(opts,
			ttnredis.FindProtosSorted(false),
			ttnredis.FindProtosWithOffsetAndCount(offset, limit),
		)
	}

	rangeProtos := func(c redis.Cmdable) error {
		return ttnredis.FindProtos(ctx, c, uidKey, r.makeAssociationKeyFunc(devUID), opts...).Range(func() (proto.Message, func() (bool, error)) {
			pb := &ttnpb.ApplicationPackageAssociation{}
			return pb, func() (bool, error) {
				pb, err := applyAssociationFieldMask(nil, pb, appendImplicitAssociationGetPaths(paths...)...)
				if err != nil {
					return false, err
				}
				pbs = append(pbs, pb)
				return true, nil
			}
		})
	}

	defer trace.StartRegion(ctx, "list application package associations by device id").End()

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

// SetAssociation implements applicationpackages.AssociationRegistry.
func (r ApplicationPackagesRegistry) SetAssociation(ctx context.Context, ids *ttnpb.ApplicationPackageAssociationIdentifiers, gets []string, f func(*ttnpb.ApplicationPackageAssociation) (*ttnpb.ApplicationPackageAssociation, []string, error)) (*ttnpb.ApplicationPackageAssociation, error) {
	devUID := unique.ID(ctx, ids.EndDeviceIds)
	fPort := r.fPortStr(ids.FPort)
	uidKey := r.uidKey(devUID)
	associationkey := r.associationKey(devUID, fPort)

	lockerID, err := ttnredis.GenerateLockerID()
	if err != nil {
		return nil, err
	}

	defer trace.StartRegion(ctx, "set application package association by id").End()

	var pb *ttnpb.ApplicationPackageAssociation
	err = ttnredis.LockedWatch(ctx, r.Redis, associationkey, lockerID, r.LockTTL, func(tx *redis.Tx) error {
		cmd := ttnredis.GetProto(ctx, tx, associationkey)
		stored := &ttnpb.ApplicationPackageAssociation{}
		if err := cmd.ScanProto(stored); errors.IsNotFound(err) {
			stored = nil
		} else if err != nil {
			return err
		}

		gets = appendImplicitAssociationGetPaths(gets...)

		var err error
		if stored != nil {
			pb = &ttnpb.ApplicationPackageAssociation{}
			if err := cmd.ScanProto(pb); err != nil {
				return err
			}
			pb, err = applyAssociationFieldMask(nil, pb, gets...)
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
			pb, err = applyAssociationFieldMask(nil, stored, gets...)
			return err
		}

		var pipelined func(redis.Pipeliner) error
		if pb == nil && len(sets) == 0 {
			pipelined = func(p redis.Pipeliner) error {
				p.Del(ctx, associationkey)
				p.SRem(ctx, uidKey, ids.FPort)
				return nil
			}
		} else {
			if pb == nil {
				pb = &ttnpb.ApplicationPackageAssociation{}
			}

			pb.UpdatedAt = timestamppb.Now()
			sets = append(append(sets[:0:0], sets...),
				"updated_at",
			)

			updated := &ttnpb.ApplicationPackageAssociation{}
			if stored == nil {
				if err := ttnpb.RequireFields(sets,
					"ids.end_device_ids.application_ids",
					"ids.end_device_ids.device_id",
					"ids.f_port",
				); err != nil {
					return errInvalidFieldmask.WithCause(err)
				}

				pb.CreatedAt = pb.UpdatedAt
				sets = append(sets, "created_at")

				updated, err = applyAssociationFieldMask(updated, pb, sets...)
				if err != nil {
					return err
				}
				if updated.Ids.EndDeviceIds.ApplicationIds.ApplicationId != ids.EndDeviceIds.ApplicationIds.ApplicationId || updated.Ids.EndDeviceIds.DeviceId != ids.EndDeviceIds.DeviceId || updated.Ids.FPort != ids.FPort {
					return errInvalidIdentifiers.New()
				}
			} else {
				if ttnpb.HasAnyField(sets, "ids.end_device_ids.application_ids.application_id") && pb.Ids.EndDeviceIds.ApplicationIds.ApplicationId != stored.Ids.EndDeviceIds.ApplicationIds.ApplicationId {
					return errReadOnlyField.WithAttributes("field", "ids.end_device_ids.application_ids.application_id")
				}
				if ttnpb.HasAnyField(sets, "ids.end_device_ids.device_id") && pb.Ids.EndDeviceIds.DeviceId != stored.Ids.EndDeviceIds.DeviceId {
					return errReadOnlyField.WithAttributes("field", "ids.end_device_ids.device_id")
				}
				if ttnpb.HasAnyField(sets, "ids.f_port") && pb.Ids.FPort != stored.Ids.FPort {
					return errReadOnlyField.WithAttributes("field", "ids.f_port")
				}
				if err := cmd.ScanProto(updated); err != nil {
					return err
				}
				updated, err = applyAssociationFieldMask(updated, pb, sets...)
				if err != nil {
					return err
				}
			}
			if err := updated.ValidateFields(); err != nil {
				return err
			}

			pipelined = func(p redis.Pipeliner) error {
				if _, err := ttnredis.SetProto(ctx, p, associationkey, updated, 0); err != nil {
					return err
				}
				p.SAdd(ctx, uidKey, ids.FPort)
				return nil
			}

			pb, err = applyAssociationFieldMask(nil, updated, gets...)
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

// GetDefaultAssociation implements applicationpackages.AssociationRegistry.
func (r ApplicationPackagesRegistry) GetDefaultAssociation(ctx context.Context, ids *ttnpb.ApplicationPackageDefaultAssociationIdentifiers, paths []string) (*ttnpb.ApplicationPackageDefaultAssociation, error) {
	pb := &ttnpb.ApplicationPackageDefaultAssociation{}
	defer trace.StartRegion(ctx, "get application package default association by id").End()
	if err := ttnredis.GetProto(ctx, r.Redis, r.associationKey(unique.ID(ctx, ids.ApplicationIds), r.fPortStr(ids.FPort))).ScanProto(pb); err != nil {
		return nil, err
	}
	return applyDefaultAssociationFieldMask(nil, pb, appendImplicitAssociationGetPaths(paths...)...)
}

// ListDefaultAssociations implements applicationpackages.AssociationRegistry.
func (r ApplicationPackagesRegistry) ListDefaultAssociations(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, paths []string) ([]*ttnpb.ApplicationPackageDefaultAssociation, error) {
	var pbs []*ttnpb.ApplicationPackageDefaultAssociation
	appUID := unique.ID(ctx, ids)
	uidKey := r.uidKey(appUID)

	opts := []ttnredis.FindProtosOption{}
	limit, offset := ttnredis.PaginationLimitAndOffsetFromContext(ctx)
	if limit != 0 {
		opts = append(opts,
			ttnredis.FindProtosSorted(false),
			ttnredis.FindProtosWithOffsetAndCount(offset, limit),
		)
	}

	rangeProtos := func(c redis.Cmdable) error {
		return ttnredis.FindProtos(ctx, c, uidKey, r.makeAssociationKeyFunc(appUID), opts...).Range(func() (proto.Message, func() (bool, error)) {
			pb := &ttnpb.ApplicationPackageDefaultAssociation{}
			return pb, func() (bool, error) {
				pb, err := applyDefaultAssociationFieldMask(nil, pb, appendImplicitAssociationGetPaths(paths...)...)
				if err != nil {
					return false, err
				}
				pbs = append(pbs, pb)
				return true, nil
			}
		})
	}

	defer trace.StartRegion(ctx, "list application package default associations by application id").End()

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

// SetDefaultAssociation implements applicationpackages.AssociationRegistry.
func (r ApplicationPackagesRegistry) SetDefaultAssociation(ctx context.Context, ids *ttnpb.ApplicationPackageDefaultAssociationIdentifiers, gets []string, f func(*ttnpb.ApplicationPackageDefaultAssociation) (*ttnpb.ApplicationPackageDefaultAssociation, []string, error)) (*ttnpb.ApplicationPackageDefaultAssociation, error) {
	appUID := unique.ID(ctx, ids.ApplicationIds)
	fPort := r.fPortStr(ids.FPort)
	uidKey := r.uidKey(appUID)
	associationkey := r.associationKey(appUID, fPort)

	lockerID, err := ttnredis.GenerateLockerID()
	if err != nil {
		return nil, err
	}

	defer trace.StartRegion(ctx, "set application package default association by id").End()

	var pb *ttnpb.ApplicationPackageDefaultAssociation
	err = ttnredis.LockedWatch(ctx, r.Redis, associationkey, lockerID, r.LockTTL, func(tx *redis.Tx) error {
		cmd := ttnredis.GetProto(ctx, tx, associationkey)
		stored := &ttnpb.ApplicationPackageDefaultAssociation{}
		if err := cmd.ScanProto(stored); errors.IsNotFound(err) {
			stored = nil
		} else if err != nil {
			return err
		}

		gets = appendImplicitAssociationGetPaths(gets...)

		var err error
		if stored != nil {
			pb = &ttnpb.ApplicationPackageDefaultAssociation{}
			if err := cmd.ScanProto(pb); err != nil {
				return err
			}
			pb, err = applyDefaultAssociationFieldMask(nil, pb, gets...)
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
			pb, err = applyDefaultAssociationFieldMask(nil, stored, gets...)
			return err
		}

		var pipelined func(redis.Pipeliner) error
		if pb == nil && len(sets) == 0 {
			pipelined = func(p redis.Pipeliner) error {
				p.Del(ctx, associationkey)
				p.SRem(ctx, uidKey, ids.FPort)
				return nil
			}
		} else {
			if pb == nil {
				pb = &ttnpb.ApplicationPackageDefaultAssociation{}
			}

			pb.UpdatedAt = timestamppb.Now()
			sets = append(append(sets[:0:0], sets...),
				"updated_at",
			)

			updated := &ttnpb.ApplicationPackageDefaultAssociation{}
			if stored == nil {
				if err := ttnpb.RequireFields(sets,
					"ids.application_ids",
					"ids.f_port",
				); err != nil {
					return errInvalidFieldmask.WithCause(err)
				}

				pb.CreatedAt = pb.UpdatedAt
				sets = append(sets, "created_at")

				updated, err = applyDefaultAssociationFieldMask(updated, pb, sets...)
				if err != nil {
					return err
				}
				if updated.Ids.ApplicationIds.ApplicationId != ids.ApplicationIds.ApplicationId || updated.Ids.FPort != ids.FPort {
					return errInvalidIdentifiers.New()
				}
			} else {
				if ttnpb.HasAnyField(sets, "ids.application_ids.application_id") && pb.Ids.ApplicationIds.ApplicationId != stored.Ids.ApplicationIds.ApplicationId {
					return errReadOnlyField.WithAttributes("field", "ids.application_ids.application_id")
				}
				if ttnpb.HasAnyField(sets, "ids.f_port") && pb.Ids.FPort != stored.Ids.FPort {
					return errReadOnlyField.WithAttributes("field", "ids.f_port")
				}
				if err := cmd.ScanProto(updated); err != nil {
					return err
				}
				updated, err = applyDefaultAssociationFieldMask(updated, pb, sets...)
				if err != nil {
					return err
				}
			}
			if err := updated.ValidateFields(); err != nil {
				return err
			}

			pipelined = func(p redis.Pipeliner) error {
				if _, err := ttnredis.SetProto(ctx, p, associationkey, updated, 0); err != nil {
					return err
				}
				p.SAdd(ctx, uidKey, ids.FPort)
				return nil
			}

			pb, err = applyDefaultAssociationFieldMask(nil, updated, gets...)
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

// WithPagination implements applicationpackages.AssociationRegistry.
func (r ApplicationPackagesRegistry) WithPagination(ctx context.Context, limit, page uint32, total *int64) context.Context {
	return ttnredis.NewContextWithPagination(ctx, int64(limit), int64(page), total)
}

// EndDeviceTransaction implements applicationpackages.TransactionRegistry.
func (r *ApplicationPackagesRegistry) EndDeviceTransaction(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, fPort uint32, packageName string, fn func(ctx context.Context) error) error {
	k := r.transactionKey(unique.ID(ctx, ids), r.fPortStr(fPort), packageName)

	lockerID, err := ttnredis.GenerateLockerID()
	if err != nil {
		return err
	}

	defer trace.StartRegion(ctx, "run end device transaction").End()

	if err := ttnredis.LockMutex(ctx, r.Redis, k, lockerID, r.LockTTL); err != nil {
		return err
	}
	defer func() {
		if err := ttnredis.UnlockMutex(ctx, r.Redis, k, lockerID, r.LockTTL); err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to unlock mutex")
		}
	}()
	return fn(ctx)
}

// Range ranges over the application packages and calls the appropriate callback function, until false is returned.
func (r ApplicationPackagesRegistry) Range(
	ctx context.Context, paths []string,
	devFunc func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.ApplicationPackageAssociation) bool,
	appFunc func(context.Context, *ttnpb.ApplicationIdentifiers, *ttnpb.ApplicationPackageDefaultAssociation) bool,
) error {
	associationEntityRegex, err := packagesRegex(r.uidKey(unique.GenericID(ctx, "*")))
	if err != nil {
		return err
	}
	return ttnredis.RangeRedisKeys(ctx, r.Redis, r.associationKey(unique.GenericID(ctx, "*"), "*"), ttnredis.DefaultRangeCount, func(key string) (bool, error) {
		if !associationEntityRegex.MatchString(key) {
			return true, nil
		}
		if strings.Contains(key, ".") {
			assoc := &ttnpb.ApplicationPackageAssociation{}
			if err := ttnredis.GetProto(ctx, r.Redis, key).ScanProto(assoc); err != nil {
				return false, err
			}
			assoc, err := applyAssociationFieldMask(nil, assoc, paths...)
			if err != nil {
				return false, err
			}
			if !devFunc(ctx, assoc.GetIds().GetEndDeviceIds(), assoc) {
				return false, nil
			}
		} else {
			defAssoc := &ttnpb.ApplicationPackageDefaultAssociation{}
			if err := ttnredis.GetProto(ctx, r.Redis, key).ScanProto(defAssoc); err != nil {
				return false, err
			}
			defAssoc, err := applyDefaultAssociationFieldMask(nil, defAssoc, paths...)
			if err != nil {
				return false, err
			}
			if !appFunc(ctx, defAssoc.GetIds().GetApplicationIds(), defAssoc) {
				return false, nil
			}
		}
		return true, nil
	})
}

func (r ApplicationPackagesRegistry) clearAssociations(ctx context.Context, id ttnpb.IDStringer) error {
	uid := unique.ID(ctx, id)
	uidKey := r.uidKey(uid)

	return r.Redis.Watch(ctx, func(tx *redis.Tx) error {
		// Retrieve the list of fPorts from the uidKey set.
		fPorts, err := tx.SMembers(ctx, uidKey).Result()
		if err != nil {
			return err
		}
		// Build all the association keys.
		keys := make([]string, 0, len(fPorts))
		for _, fPort := range fPorts {
			keys = append(keys, r.associationKey(uid, fPort))
		}
		keys = append(keys, uidKey)

		if _, err := tx.TxPipelined(ctx, func(p redis.Pipeliner) error {
			p.Del(ctx, keys...)
			return nil
		}); err != nil {
			return err
		}
		return nil
	}, uidKey)
}

// ClearAssociations clears all the associations for an end device.
func (r ApplicationPackagesRegistry) ClearAssociations(
	ctx context.Context, ids *ttnpb.EndDeviceIdentifiers,
) error {
	return r.clearAssociations(ctx, ids)
}

// ClearDefaultAssociations clears all package associations for an application.
func (r ApplicationPackagesRegistry) ClearDefaultAssociations(
	ctx context.Context, ids *ttnpb.ApplicationIdentifiers,
) error {
	return r.clearAssociations(ctx, ids)
}
