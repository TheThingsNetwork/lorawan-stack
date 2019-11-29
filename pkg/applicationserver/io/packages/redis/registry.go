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
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/gogo/protobuf/proto"
	"go.thethings.network/lorawan-stack/pkg/errors"
	ttnredis "go.thethings.network/lorawan-stack/pkg/redis"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
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

// ApplicationPackagesRegistry is a Redis application packages registry.
type ApplicationPackagesRegistry struct {
	Redis *ttnredis.Client
}

func (r *ApplicationPackagesRegistry) devKey(devUID string) string {
	return r.Redis.Key("uid", devUID)
}

func (r *ApplicationPackagesRegistry) fPortStr(fPort uint32) string {
	if fPort > 255 {
		panic("FPort cannot be higher than 255")
	}
	return strconv.FormatUint(uint64(fPort), 10)
}

func (r *ApplicationPackagesRegistry) associationKey(devUID string, fPort string) string {
	return r.Redis.Key("uid", devUID, fPort)
}

func (r *ApplicationPackagesRegistry) makeAssociationKeyFunc(devUID string) func(port string) string {
	return func(port string) string {
		return r.associationKey(devUID, port)
	}
}

// Get implements applicationpackages.AssociationRegistry.
func (r ApplicationPackagesRegistry) Get(ctx context.Context, ids ttnpb.ApplicationPackageAssociationIdentifiers, paths []string) (*ttnpb.ApplicationPackageAssociation, error) {
	pb := &ttnpb.ApplicationPackageAssociation{}
	defer trace.StartRegion(ctx, "get application package association by id").End()
	if err := ttnredis.GetProto(r.Redis, r.associationKey(unique.ID(ctx, ids.EndDeviceIdentifiers), r.fPortStr(ids.FPort))).ScanProto(pb); err != nil {
		return nil, err
	}
	return applyAssociationFieldMask(nil, pb, appendImplicitAssociationGetPaths(paths...)...)
}

// List implements applicationpackages.AssociationRegistry.
func (r ApplicationPackagesRegistry) List(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, paths []string) ([]*ttnpb.ApplicationPackageAssociation, error) {
	var pbs []*ttnpb.ApplicationPackageAssociation
	devUID := unique.ID(ctx, ids)
	dk := r.devKey(devUID)

	defer trace.StartRegion(ctx, "list application package associations by device id").End()

	err := r.Redis.Watch(func(tx *redis.Tx) (err error) {
		opts := []ttnredis.FindProtosOption{ttnredis.FindProtosSorted(false)}

		limit, offset := ttnredis.PaginationLimitAndOffsetFromContext(ctx)
		if limit != 0 {
			total, err := tx.SCard(dk).Result()
			if err != nil {
				return ttnredis.ConvertError(err)
			}
			ttnredis.SetPaginationTotal(ctx, total)
			opts = append(opts, ttnredis.FindProtosWithOffsetAndCount(offset, limit))
		}

		return ttnredis.FindProtos(tx, dk, r.makeAssociationKeyFunc(devUID), opts...).Range(func() (proto.Message, func() (bool, error)) {
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
	}, dk)
	if err != nil {
		return nil, err
	}
	return pbs, nil
}

// Set implements applicationpackages.AssociationRegistry.
func (r ApplicationPackagesRegistry) Set(ctx context.Context, ids ttnpb.ApplicationPackageAssociationIdentifiers, gets []string, f func(*ttnpb.ApplicationPackageAssociation) (*ttnpb.ApplicationPackageAssociation, []string, error)) (*ttnpb.ApplicationPackageAssociation, error) {
	devUID := unique.ID(ctx, ids.EndDeviceIdentifiers)
	fPort := r.fPortStr(ids.FPort)
	dk := r.devKey(devUID)
	ak := r.associationKey(devUID, fPort)

	defer trace.StartRegion(ctx, "set application package association by id").End()

	var pb *ttnpb.ApplicationPackageAssociation
	err := r.Redis.Watch(func(tx *redis.Tx) error {
		cmd := ttnredis.GetProto(tx, ak)
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
				p.Del(ak)
				p.SRem(dk, ids.FPort)
				return nil
			}
		} else {
			if pb == nil {
				pb = &ttnpb.ApplicationPackageAssociation{}
			}

			pb.UpdatedAt = time.Now().UTC()
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
				if updated.ApplicationID != ids.ApplicationID || updated.DeviceID != ids.DeviceID || updated.FPort != ids.FPort {
					return errInvalidIdentifiers
				}
			} else {
				if ttnpb.HasAnyField(sets, "ids.end_device_ids.application_ids.application_id") && pb.ApplicationID != stored.ApplicationID {
					return errReadOnlyField.WithAttributes("field", "ids.end_device_ids.application_ids.application_id")
				}
				if ttnpb.HasAnyField(sets, "ids.end_device_ids.device_id") && pb.DeviceID != stored.DeviceID {
					return errReadOnlyField.WithAttributes("field", "ids.end_device_ids.device_id")
				}
				if ttnpb.HasAnyField(sets, "ids.f_port") && pb.FPort != stored.FPort {
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
			if err := updated.ValidateFields(sets...); err != nil {
				return err
			}

			pipelined = func(p redis.Pipeliner) error {
				if _, err := ttnredis.SetProto(p, ak, updated, 0); err != nil {
					return err
				}
				p.SAdd(dk, ids.FPort)
				return nil
			}

			pb, err = applyAssociationFieldMask(nil, updated, gets...)
			if err != nil {
				return err
			}
		}
		_, err = tx.Pipelined(pipelined)
		if err != nil {
			return err
		}
		return nil
	}, ak)
	if err != nil {
		return nil, err
	}
	return pb, nil
}

// WithPagination implements applicationpackages.AssociationRegistry.
func (r ApplicationPackagesRegistry) WithPagination(ctx context.Context, limit, page uint32, total *int64) context.Context {
	return ttnredis.NewContextWithPagination(ctx, int64(limit), int64(page), total)
}
