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
	"bytes"
	"context"
	"encoding/base64"
	"regexp"
	"runtime/trace"
	"time"

	"github.com/go-redis/redis/v8"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/provisioning"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

var (
	errAlreadyProvisioned   = errors.DefineAlreadyExists("already_provisioned", "device already provisioned")
	errDuplicateIdentifiers = errors.DefineAlreadyExists("duplicate_identifiers", "duplicate identifiers")
	errInvalidFieldmask     = errors.DefineInvalidArgument("invalid_fieldmask", "invalid fieldmask")
	errInvalidIdentifiers   = errors.DefineInvalidArgument("invalid_identifiers", "invalid identifiers")
	errReadOnlyField        = errors.DefineInvalidArgument("read_only_field", "read-only field `{field}`")
	errProvisionerNotFound  = errors.DefineNotFound("provisioner_not_found", "provisioner `{id}` not found")
)

// SchemaVersion is the Network Server database schema version. Bump when a migration is required.
const SchemaVersion = 1

// DeviceRegistry is an implementation of joinserver.DeviceRegistry.
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

func provisionerUniqueID(dev *ttnpb.EndDevice) (string, error) {
	if dev.ProvisionerId == "" {
		return "", nil
	}
	provisioner := provisioning.Get(dev.ProvisionerId)
	if provisioner == nil {
		return "", errProvisionerNotFound.WithAttributes("id", dev.ProvisionerId)
	}
	return provisioner.UniqueID(dev.ProvisioningData)
}

func (r *DeviceRegistry) uidKey(uid string) string {
	return r.Redis.Key("uid", uid)
}

func (r *DeviceRegistry) euiKey(joinEUI, devEUI types.EUI64) string {
	return r.Redis.Key("eui", joinEUI.String(), devEUI.String())
}

func (r *DeviceRegistry) provisionerKey(provisionerID, pid string) string {
	return r.Redis.Key("provisioner", provisionerID, pid)
}

// GetByID gets device by appID, devID.
func (r *DeviceRegistry) GetByID(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
	ids := &ttnpb.EndDeviceIdentifiers{
		ApplicationIds: appID,
		DeviceId:       devID,
	}
	if err := ids.ValidateContext(ctx); err != nil {
		return nil, err
	}

	defer trace.StartRegion(ctx, "get end device by id").End()

	pb := &ttnpb.EndDevice{}
	if err := ttnredis.GetProto(ctx, r.Redis, r.uidKey(unique.ID(ctx, ids))).ScanProto(pb); err != nil {
		return nil, err
	}
	return ttnpb.FilterGetEndDevice(pb, paths...)
}

// GetByEUI gets device by joinEUI, devEUI.
func (r *DeviceRegistry) GetByEUI(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.ContextualEndDevice, error) {
	if devEUI.IsZero() {
		return nil, errInvalidIdentifiers.New()
	}

	defer trace.StartRegion(ctx, "get end device by eui").End()

	pb := &ttnpb.EndDevice{}
	if err := ttnredis.FindProto(ctx, r.Redis, r.euiKey(joinEUI, devEUI), func(uid string) (string, error) {
		var err error
		ctx, err = unique.WithContext(ctx, uid)
		if err != nil {
			return "", err
		}
		return r.uidKey(uid), nil
	}).ScanProto(pb); err != nil {
		return nil, err
	}
	filtered, err := ttnpb.FilterGetEndDevice(pb, paths...)
	if err != nil {
		return nil, err
	}
	return &ttnpb.ContextualEndDevice{
		Context:   ctx,
		EndDevice: filtered,
	}, nil
}

func equalEUI64(x, y *types.EUI64) bool {
	if x == nil || y == nil {
		return x == y
	}
	return x.Equal(*y)
}

func (r *DeviceRegistry) set(ctx context.Context, tx *redis.Tx, uid string, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.ContextualEndDevice, error) {
	ctx, err := unique.WithContext(ctx, uid)
	if err != nil {
		return nil, err
	}
	uk := r.uidKey(uid)

	cmd := ttnredis.GetProto(ctx, tx, uk)
	stored := &ttnpb.EndDevice{}
	if err := cmd.ScanProto(stored); errors.IsNotFound(err) {
		stored = nil
	} else if err != nil {
		return nil, err
	}

	var pb *ttnpb.EndDevice
	if stored != nil {
		pb = &ttnpb.EndDevice{}
		if err := cmd.ScanProto(pb); err != nil {
			return nil, err
		}
		pb, err = ttnpb.FilterGetEndDevice(pb, gets...)
		if err != nil {
			return nil, err
		}
	}

	var sets []string
	pb, sets, err = f(ctx, pb)
	if err != nil {
		return nil, err
	}
	if err := ttnpb.ProhibitFields(sets,
		"created_at",
		"updated_at",
	); err != nil {
		return nil, errInvalidFieldmask.WithCause(err)
	}

	if stored == nil && pb == nil {
		return nil, nil
	}
	if pb != nil && len(sets) == 0 {
		filtered, err := ttnpb.FilterGetEndDevice(stored, gets...)
		if err != nil {
			return nil, err
		}
		return &ttnpb.ContextualEndDevice{
			Context:   ctx,
			EndDevice: filtered,
		}, nil
	}

	var pipelined func(redis.Pipeliner) error
	if pb == nil && len(sets) == 0 {
		pipelined = func(p redis.Pipeliner) error {
			p.Del(ctx, uk)
			if stored.Ids.JoinEui != nil && stored.Ids.DevEui != nil {
				p.Del(ctx, r.euiKey(*stored.Ids.JoinEui, *stored.Ids.DevEui))
			}
			pid, err := provisionerUniqueID(stored)
			if err != nil {
				return err
			}
			if pid != "" {
				p.Del(ctx, r.provisionerKey(stored.ProvisionerId, pid))
			}
			return nil
		}
	} else {
		if pb == nil {
			pb = &ttnpb.EndDevice{}
		}

		pb.UpdatedAt = ttnpb.ProtoTimePtr(time.Now())
		sets = append(append(sets[:0:0], sets...),
			"updated_at",
		)

		updated := &ttnpb.EndDevice{}
		var updatedPID string
		if stored == nil {
			if err := ttnpb.RequireFields(sets,
				"ids.application_ids",
				"ids.dev_eui",
				"ids.device_id",
				"ids.join_eui",
			); err != nil {
				return nil, errInvalidFieldmask.WithCause(err)
			}

			pb.CreatedAt = pb.UpdatedAt
			sets = append(sets, "created_at")

			updated, err = ttnpb.ApplyEndDeviceFieldMask(updated, pb, sets...)
			if err != nil {
				return nil, err
			}
			updatedPID, err = provisionerUniqueID(updated)
			if err != nil {
				return nil, err
			}
			if updated.Ids.JoinEui == nil || updated.Ids.DevEui == nil || updated.Ids.DevEui.IsZero() {
				return nil, errInvalidIdentifiers.New()
			}
		} else {
			if ttnpb.HasAnyField(sets, "ids.application_ids.application_id") && pb.Ids.ApplicationIds.ApplicationId != stored.Ids.ApplicationIds.ApplicationId {
				return nil, errReadOnlyField.WithAttributes("field", "ids.application_ids.application_id")
			}
			if ttnpb.HasAnyField(sets, "ids.device_id") && pb.Ids.DeviceId != stored.Ids.DeviceId {
				return nil, errReadOnlyField.WithAttributes("field", "ids.device_id")
			}
			if ttnpb.HasAnyField(sets, "ids.join_eui") && !equalEUI64(pb.Ids.JoinEui, stored.Ids.JoinEui) {
				return nil, errReadOnlyField.WithAttributes("field", "ids.join_eui")
			}
			if ttnpb.HasAnyField(sets, "ids.dev_eui") && !equalEUI64(pb.Ids.DevEui, stored.Ids.DevEui) {
				return nil, errReadOnlyField.WithAttributes("field", "ids.dev_eui")
			}
			if ttnpb.HasAnyField(sets, "provisioner_id") && pb.ProvisionerId != stored.ProvisionerId {
				return nil, errReadOnlyField.WithAttributes("field", "provisioner_id")
			}
			if ttnpb.HasAnyField(sets, "provisioning_data") && !pb.ProvisioningData.Equal(stored.ProvisioningData) {
				return nil, errReadOnlyField.WithAttributes("field", "provisioning_data")
			}
			if err := cmd.ScanProto(updated); err != nil {
				return nil, err
			}
			updated, err = ttnpb.ApplyEndDeviceFieldMask(updated, pb, sets...)
			if err != nil {
				return nil, err
			}
		}
		if err := updated.ValidateFields(); err != nil {
			return nil, err
		}

		pipelined = func(p redis.Pipeliner) error {
			if stored == nil {
				ek := r.euiKey(*updated.Ids.JoinEui, *updated.Ids.DevEui)
				if err := tx.Watch(ctx, ek).Err(); err != nil {
					return err
				}
				i, err := tx.Exists(ctx, ek).Result()
				if err != nil {
					return err
				}
				if i != 0 {
					return errDuplicateIdentifiers.New()
				}
				p.SetNX(ctx, ek, uid, 0)
			}

			if updatedPID != "" {
				pk := r.provisionerKey(updated.ProvisionerId, updatedPID)
				if err := tx.Watch(ctx, pk).Err(); err != nil {
					return err
				}
				i, err := tx.Exists(ctx, pk).Result()
				if err != nil {
					return err
				}
				if i != 0 {
					return errAlreadyProvisioned.New()
				}
				p.SetNX(ctx, pk, uid, 0)
			}

			_, err := ttnredis.SetProto(ctx, p, uk, updated, 0)
			if err != nil {
				return err
			}
			return nil
		}
		pb, err = ttnpb.FilterGetEndDevice(updated, gets...)
		if err != nil {
			return nil, err
		}
	}
	_, err = tx.TxPipelined(ctx, pipelined)
	if err != nil {
		return nil, ttnredis.ConvertError(err)
	}
	return &ttnpb.ContextualEndDevice{
		Context:   ctx,
		EndDevice: pb,
	}, nil
}

// SetByEUI sets device by joinEUI, devEUI.
// SetByEUI will only succeed if the device is set via SetByID first.
func (r *DeviceRegistry) SetByEUI(ctx context.Context, joinEUI types.EUI64, devEUI types.EUI64, gets []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.ContextualEndDevice, error) {
	if devEUI.IsZero() {
		return nil, errInvalidIdentifiers.New()
	}
	ek := r.euiKey(joinEUI, devEUI)

	lockerID, err := ttnredis.GenerateLockerID()
	if err != nil {
		return nil, err
	}

	defer trace.StartRegion(ctx, "set end device by eui").End()

	var pb *ttnpb.ContextualEndDevice
	err = ttnredis.LockedWatch(ctx, r.Redis, ek, lockerID, r.LockTTL, func(tx *redis.Tx) error {
		uid, err := tx.Get(ctx, ek).Result()
		if err != nil {
			return err
		}
		if err := tx.Watch(ctx, r.uidKey(uid)).Err(); err != nil {
			return err
		}
		pb, err = r.set(ctx, tx, uid, gets, f)
		return err
	})
	if err != nil {
		return nil, ttnredis.ConvertError(err)
	}
	return pb, nil
}

// SetByID sets device by appID, devID.
func (r *DeviceRegistry) SetByID(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, devID string, gets []string, f func(pb *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
	ids := &ttnpb.EndDeviceIdentifiers{
		ApplicationIds: appID,
		DeviceId:       devID,
	}
	if err := ids.ValidateContext(ctx); err != nil {
		return nil, err
	}
	uid := unique.ID(ctx, ids)

	defer trace.StartRegion(ctx, "set end device by id").End()

	var pb *ttnpb.ContextualEndDevice
	err := r.Redis.Watch(ctx, func(tx *redis.Tx) error {
		var err error
		pb, err = r.set(ctx, tx, uid, gets, func(ctx context.Context, stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			updated, sets, err := f(stored)
			if err != nil {
				return nil, nil, err
			}
			if stored == nil && updated != nil && (updated.Ids.ApplicationIds.ApplicationId != appID.ApplicationId || updated.Ids.DeviceId != devID) {
				return nil, nil, errInvalidIdentifiers.New()
			}
			return updated, sets, nil
		})
		return err
	}, r.uidKey(uid))
	if err != nil {
		return nil, ttnredis.ConvertError(err)
	}
	return pb.EndDevice, nil
}

func (r *DeviceRegistry) RangeByID(ctx context.Context, paths []string, f func(context.Context, *ttnpb.EndDeviceIdentifiers, *ttnpb.EndDevice) bool) error {
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

// KeyRegistry is an implementation of joinserver.KeyRegistry.
type KeyRegistry struct {
	Redis   *ttnredis.Client
	LockTTL time.Duration
	// Limit is the maximum number of session keys to store per JoinEUI and DevEUI combination.
	Limit int
}

// Init initializes the KeyRegistry.
func (r *KeyRegistry) Init(ctx context.Context) error {
	if err := ttnredis.InitMutex(ctx, r.Redis); err != nil {
		return err
	}
	return nil
}

func (r *KeyRegistry) idValue(id []byte) string {
	return base64.RawStdEncoding.EncodeToString(id)
}

func (r *KeyRegistry) idKey(joinEUI, devEUI types.EUI64, id string) string {
	return r.Redis.Key("id", joinEUI.String(), devEUI.String(), id)
}

func (r *KeyRegistry) idSetKey(joinEUI, devEUI types.EUI64) string {
	return r.Redis.Key("ids", joinEUI.String(), devEUI.String())
}

// GetByID gets session keys by joinEUI, devEUI, id.
func (r *KeyRegistry) GetByID(ctx context.Context, joinEUI, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
	if devEUI.IsZero() || len(id) == 0 {
		return nil, errInvalidIdentifiers.New()
	}

	defer trace.StartRegion(ctx, "get session keys").End()

	pb := &ttnpb.SessionKeys{}
	if err := ttnredis.GetProto(ctx, r.Redis, r.idKey(joinEUI, devEUI, r.idValue(id))).ScanProto(pb); err != nil {
		return nil, err
	}
	return ttnpb.FilterGetSessionKeys(pb, paths...)
}

// SetByID sets session keys by joinEUI, devEUI, id.
func (r *KeyRegistry) SetByID(ctx context.Context, joinEUI, devEUI types.EUI64, id []byte, gets []string, f func(*ttnpb.SessionKeys) (*ttnpb.SessionKeys, []string, error)) (*ttnpb.SessionKeys, error) {
	if devEUI.IsZero() || len(id) == 0 {
		return nil, errInvalidIdentifiers.New()
	}
	ik, sk := r.idKey(joinEUI, devEUI, r.idValue(id)), r.idSetKey(joinEUI, devEUI)

	lockerID, err := ttnredis.GenerateLockerID()
	if err != nil {
		return nil, err
	}

	defer trace.StartRegion(ctx, "set session keys").End()

	var pb *ttnpb.SessionKeys
	err = ttnredis.LockedWatch(ctx, r.Redis, sk, lockerID, r.LockTTL, func(tx *redis.Tx) error {
		cmd := ttnredis.GetProto(ctx, tx, ik)
		stored := &ttnpb.SessionKeys{}
		if err := cmd.ScanProto(stored); errors.IsNotFound(err) {
			stored = nil
		} else if err != nil {
			return err
		}

		var err error
		if stored != nil {
			pb = &ttnpb.SessionKeys{}
			if err := cmd.ScanProto(pb); err != nil {
				return err
			}
			pb, err = ttnpb.FilterGetSessionKeys(pb, gets...)
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
			pb, err = ttnpb.FilterGetSessionKeys(stored, gets...)
			return err
		}

		if pb == nil && len(sets) == 0 {
			_, err = tx.TxPipelined(ctx, func(p redis.Pipeliner) error {
				p.Del(ctx, ik)
				p.LRem(ctx, sk, 0, r.idValue(id))
				return nil
			})
			if err != nil {
				return err
			}
			return nil
		}

		if pb == nil {
			pb = &ttnpb.SessionKeys{}
		}

		updated := &ttnpb.SessionKeys{}
		if stored == nil {
			if err := ttnpb.RequireFields(sets,
				"session_key_id",
			); err != nil {
				return errInvalidFieldmask.WithCause(err)
			}
			updated, err = ttnpb.ApplySessionKeysFieldMask(updated, pb, sets...)
			if err != nil {
				return err
			}
			if !bytes.Equal(updated.SessionKeyId, id) {
				return errInvalidIdentifiers.New()
			}
			if r.Limit > 0 {
				count, err := tx.RPush(ctx, sk, r.idValue(id)).Result()
				if err != nil {
					return err
				}
				if d := int(count) - r.Limit; d > 0 {
					oldIDs, err := tx.LPopCount(ctx, sk, d).Result()
					if err != nil {
						return err
					}
					if _, err := tx.Pipelined(ctx, func(p redis.Pipeliner) error {
						for _, oldID := range oldIDs {
							p.Del(ctx, r.idKey(joinEUI, devEUI, oldID))
						}
						return nil
					}); err != nil {
						return err
					}
				}
			}
		} else {
			if err := ttnpb.ProhibitFields(sets,
				"session_key_id",
			); err != nil {
				return errInvalidFieldmask.WithCause(err)
			}
			if err := cmd.ScanProto(updated); err != nil {
				return err
			}
			updated, err = ttnpb.ApplySessionKeysFieldMask(updated, pb, sets...)
			if err != nil {
				return err
			}
		}
		if err := updated.ValidateFields(); err != nil {
			return err
		}

		pb, err = ttnpb.FilterGetSessionKeys(updated, gets...)
		if err != nil {
			return err
		}
		_, err = ttnredis.SetProto(ctx, tx, ik, updated, 0)
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

func (r *KeyRegistry) Delete(ctx context.Context, joinEUI, devEUI types.EUI64) error {
	if devEUI.IsZero() {
		return errInvalidIdentifiers.New()
	}
	sk := r.idSetKey(joinEUI, devEUI)

	lockerID, err := ttnredis.GenerateLockerID()
	if err != nil {
		return err
	}

	defer trace.StartRegion(ctx, "delete session keys").End()

	err = ttnredis.LockedWatch(ctx, r.Redis, sk, lockerID, r.LockTTL, func(tx *redis.Tx) error {
		sids, err := tx.LRange(ctx, sk, 0, 1<<24).Result()
		if err != nil {
			return err
		}
		_, err = tx.Pipelined(ctx, func(p redis.Pipeliner) error {
			for _, sid := range sids {
				p.Del(ctx, r.idKey(joinEUI, devEUI, sid))
			}
			p.Del(ctx, sk)
			return nil
		})
		return err
	})
	if err != nil {
		return ttnredis.ConvertError(err)
	}
	return nil
}

// applyApplicationActivationSettingsFieldMask applies fields specified by paths from src to dst and returns the result.
// If dst is nil, a new ApplicationActivationSettings is created.
func applyApplicationActivationSettingsFieldMask(dst, src *ttnpb.ApplicationActivationSettings, paths ...string) (*ttnpb.ApplicationActivationSettings, error) {
	if dst == nil {
		dst = &ttnpb.ApplicationActivationSettings{}
	}
	return dst, dst.SetFields(src, paths...)
}

// filterGetApplicationActivationSettings returns a new ttnpb.ApplicationActivationSettings with only fields specified by paths set.
func filterGetApplicationActivationSettings(pb *ttnpb.ApplicationActivationSettings, paths ...string) (*ttnpb.ApplicationActivationSettings, error) {
	return applyApplicationActivationSettingsFieldMask(nil, pb, paths...)
}

// ApplicationActivationSettingRegistry is an implementation of joinserver.ApplicationActivationSettingRegistry.
type ApplicationActivationSettingRegistry struct {
	Redis   *ttnredis.Client
	LockTTL time.Duration
}

// Init initializes the ApplicationActivationSettingRegistry.
func (r *ApplicationActivationSettingRegistry) Init(ctx context.Context) error {
	if err := ttnredis.InitMutex(ctx, r.Redis); err != nil {
		return err
	}
	return nil
}

func (r *ApplicationActivationSettingRegistry) uidKey(uid string) string {
	return r.Redis.Key("uid", uid)
}

// GetByID gets application activation settings by appID.
func (r *ApplicationActivationSettingRegistry) GetByID(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, paths []string) (*ttnpb.ApplicationActivationSettings, error) {
	if appID.IsZero() {
		return nil, errInvalidIdentifiers.New()
	}

	defer trace.StartRegion(ctx, "get application activation settings").End()

	pb := &ttnpb.ApplicationActivationSettings{}
	if err := ttnredis.GetProto(ctx, r.Redis, r.uidKey(unique.ID(ctx, appID))).ScanProto(pb); err != nil {
		return nil, err
	}
	return filterGetApplicationActivationSettings(pb, paths...)
}

// SetByID sets application activation settings by appID.
func (r *ApplicationActivationSettingRegistry) SetByID(ctx context.Context, appID *ttnpb.ApplicationIdentifiers, gets []string, f func(*ttnpb.ApplicationActivationSettings) (*ttnpb.ApplicationActivationSettings, []string, error)) (*ttnpb.ApplicationActivationSettings, error) {
	if appID.IsZero() {
		return nil, errInvalidIdentifiers.New()
	}
	uk := r.uidKey(unique.ID(ctx, appID))

	lockerID, err := ttnredis.GenerateLockerID()
	if err != nil {
		return nil, err
	}

	defer trace.StartRegion(ctx, "set application activation settings").End()

	var pb *ttnpb.ApplicationActivationSettings
	err = ttnredis.LockedWatch(ctx, r.Redis, uk, lockerID, r.LockTTL, func(tx *redis.Tx) error {
		cmd := ttnredis.GetProto(ctx, tx, uk)
		stored := &ttnpb.ApplicationActivationSettings{}
		if err := cmd.ScanProto(stored); errors.IsNotFound(err) {
			stored = nil
		} else if err != nil {
			return err
		}

		var err error
		if stored != nil {
			pb = &ttnpb.ApplicationActivationSettings{}
			if err := cmd.ScanProto(pb); err != nil {
				return err
			}
			pb, err = filterGetApplicationActivationSettings(pb, gets...)
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
			pb, err = filterGetApplicationActivationSettings(stored, gets...)
			return err
		}

		var pipelined func(redis.Pipeliner) error
		if pb == nil && len(sets) == 0 {
			pipelined = func(p redis.Pipeliner) error {
				p.Del(ctx, uk)
				return nil
			}
		} else {
			if pb == nil {
				pb = &ttnpb.ApplicationActivationSettings{}
			}

			updated := &ttnpb.ApplicationActivationSettings{}
			if stored == nil {
				updated, err = applyApplicationActivationSettingsFieldMask(updated, pb, sets...)
				if err != nil {
					return err
				}
			} else {
				if err := cmd.ScanProto(updated); err != nil {
					return err
				}
				updated, err = applyApplicationActivationSettingsFieldMask(updated, pb, sets...)
				if err != nil {
					return err
				}
			}
			if err := updated.ValidateFields(); err != nil {
				return err
			}

			pipelined = func(p redis.Pipeliner) error {
				_, err := ttnredis.SetProto(ctx, p, uk, updated, 0)
				if err != nil {
					return err
				}
				return nil
			}
			pb, err = filterGetApplicationActivationSettings(updated, gets...)
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

var uniqueIDPattern = regexp.MustCompile(`(.*)\:(.*)`)

func (r *ApplicationActivationSettingRegistry) Range(ctx context.Context, paths []string, f func(context.Context, *ttnpb.ApplicationIdentifiers, *ttnpb.ApplicationActivationSettings) bool) error {
	appKeyRegex, err := ttnredis.EntityRegex((r.uidKey(unique.GenericID(ctx, "*"))))
	if err != nil {
		return err
	}
	return ttnredis.RangeRedisKeys(ctx, r.Redis, r.uidKey(unique.GenericID(ctx, "*")), ttnredis.DefaultRangeCount, func(key string) (bool, error) {
		if !appKeyRegex.MatchString(key) {
			return true, nil
		}
		matches := uniqueIDPattern.FindStringSubmatch(key)
		appUID := matches[len(matches)-1]
		applicationId, err := unique.ToApplicationID(appUID)
		if err != nil {
			return false, err
		}
		ctx, err := unique.WithContext(ctx, appUID)
		if err != nil {
			return false, err
		}
		appAs := &ttnpb.ApplicationActivationSettings{}
		if err := ttnredis.GetProto(ctx, r.Redis, key).ScanProto(appAs); err != nil {
			return false, err
		}
		appAs, err = applyApplicationActivationSettingsFieldMask(nil, appAs, paths...)
		if err != nil {
			return false, err
		}
		if !f(ctx, applicationId, appAs) {
			return false, nil
		}
		return true, nil
	})
}
