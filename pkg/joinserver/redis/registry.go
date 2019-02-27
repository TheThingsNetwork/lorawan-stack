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
	"bytes"
	"context"
	"encoding/base64"
	"time"

	"github.com/go-redis/redis"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/joinserver/provisioning"
	ttnredis "go.thethings.network/lorawan-stack/pkg/redis"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

const (
	euiKey         = "eui"
	provisionerKey = "provisioner"
	uidKey         = "uid"
)

var (
	errAlreadyProvisioned   = errors.DefineAlreadyExists("already_provisioned", "device already provisioned")
	errDuplicateIdentifiers = errors.DefineAlreadyExists("duplicate_identifiers", "duplicate identifiers")
	errInvalidIdentifiers   = errors.DefineInvalidArgument("invalid_identifiers", "invalid identifiers")
	errProvisionerNotFound  = errors.DefineNotFound("provisioner_not_found", "provisioner `{id}` not found")
)

func applyDeviceFieldMask(dst, src *ttnpb.EndDevice, paths ...string) (*ttnpb.EndDevice, error) {
	if dst == nil {
		dst = &ttnpb.EndDevice{}
	}
	if err := dst.SetFields(src, append(paths, "ids")...); err != nil {
		return nil, err
	}
	// TODO: Validation after applying fieldmasks https://github.com/TheThingsNetwork/lorawan-stack/issues/51
	return dst, nil
}

// DeviceRegistry is an implementation of joinserver.DeviceRegistry.
type DeviceRegistry struct {
	Redis *ttnredis.Client
}

func provisionerUniqueID(dev *ttnpb.EndDevice) (string, error) {
	if dev.ProvisionerID == "" {
		return "", nil
	}
	provisioner := provisioning.Get(dev.ProvisionerID)
	if provisioner == nil {
		return "", errProvisionerNotFound.WithAttributes("id", dev.ProvisionerID)
	}
	return provisioner.UniqueID(dev.ProvisioningData)
}

// GetByEUI gets device by joinEUI, devEUI.
func (r *DeviceRegistry) GetByEUI(ctx context.Context, joinEUI types.EUI64, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error) {
	if joinEUI.IsZero() || devEUI.IsZero() {
		return nil, errInvalidIdentifiers
	}

	pb := &ttnpb.EndDevice{}
	if err := ttnredis.GetProto(r.Redis, r.Redis.Key(euiKey, joinEUI.String(), devEUI.String())).ScanProto(pb); err != nil {
		return nil, err
	}
	return applyDeviceFieldMask(&ttnpb.EndDevice{}, pb, paths...)
}

func equalEUI(x, y *types.EUI64) bool {
	if x == nil || y == nil {
		return x == y
	}
	return x.Equal(*y)
}

// SetByEUI sets device by joinEUI, devEUI.
func (r *DeviceRegistry) SetByEUI(ctx context.Context, joinEUI types.EUI64, devEUI types.EUI64, gets []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
	if joinEUI.IsZero() || devEUI.IsZero() {
		return nil, errInvalidIdentifiers
	}
	ek := r.Redis.Key(euiKey, joinEUI.String(), devEUI.String())

	var pb *ttnpb.EndDevice
	err := r.Redis.Watch(func(tx *redis.Tx) error {
		cmd := ttnredis.GetProto(tx, ek)
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
		if stored == nil && pb == nil {
			return nil
		}

		var f func(redis.Pipeliner) error
		if pb == nil {
			f = func(p redis.Pipeliner) error {
				p.Del(ek)
				if !stored.ApplicationIdentifiers.IsZero() && stored.DeviceID != "" {
					p.Del(r.Redis.Key(uidKey, unique.ID(ctx, stored.EndDeviceIdentifiers)))
				}
				pid, err := provisionerUniqueID(stored)
				if err != nil {
					return err
				}
				if pid != "" {
					p.Del(r.Redis.Key(provisionerKey, stored.ProvisionerID, pid))
				}
				return nil
			}
		} else {
			if pb.JoinEUI == nil || *pb.JoinEUI != joinEUI ||
				pb.DevEUI == nil || *pb.DevEUI != devEUI {
				return errInvalidIdentifiers
			}

			pb.UpdatedAt = time.Now().UTC()
			sets = append(sets, "updated_at")

			updated := &ttnpb.EndDevice{}
			var updatedPID string
			if stored == nil {
				pb.CreatedAt = pb.UpdatedAt
				sets = append(sets, "created_at")

				updated, err = applyDeviceFieldMask(updated, pb, sets...)
				if err != nil {
					return err
				}
				updatedPID, err = provisionerUniqueID(updated)
				if err != nil {
					return err
				}
			} else {
				if err := cmd.ScanProto(updated); err != nil {
					return err
				}
				updated, err = applyDeviceFieldMask(updated, pb, sets...)
				if err != nil {
					return err
				}

				storedPID, err := provisionerUniqueID(stored)
				if err != nil {
					return err
				}
				updatedPID, err = provisionerUniqueID(updated)
				if err != nil {
					return err
				}
				if !equalEUI(stored.JoinEUI, updated.JoinEUI) || !equalEUI(stored.DevEUI, updated.DevEUI) ||
					stored.ApplicationIdentifiers != updated.ApplicationIdentifiers || stored.DeviceID != updated.DeviceID ||
					storedPID != updatedPID {
					return errInvalidIdentifiers
				}
			}
			pb, err = applyDeviceFieldMask(nil, updated, gets...)
			if err != nil {
				return err
			}

			f = func(p redis.Pipeliner) error {
				eid := ttnredis.Key(joinEUI.String(), devEUI.String())

				if stored == nil {
					if !updated.ApplicationIdentifiers.IsZero() && updated.DeviceID != "" {
						uk := r.Redis.Key(uidKey, unique.ID(ctx, updated.EndDeviceIdentifiers))
						if err := tx.Watch(uk).Err(); err != nil {
							return err
						}
						i, err := tx.Exists(uk).Result()
						if err != nil {
							return ttnredis.ConvertError(err)
						}
						if i != 0 {
							return errDuplicateIdentifiers
						}
						p.SetNX(uk, eid, 0)
					}

					if updatedPID != "" {
						pk := r.Redis.Key(provisionerKey, updated.ProvisionerID, updatedPID)
						if err := tx.Watch(pk).Err(); err != nil {
							return err
						}
						i, err := tx.Exists(pk).Result()
						if err != nil {
							return ttnredis.ConvertError(err)
						}
						if i != 0 {
							return errAlreadyProvisioned
						}
						p.SetNX(pk, eid, 0)
					}
				}

				_, err := ttnredis.SetProto(p, ek, updated, 0)
				if err != nil {
					return err
				}
				return nil
			}
		}
		_, err = tx.Pipelined(f)
		if err != nil {
			return err
		}
		return nil
	}, ek)
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

// KeyRegistry is an implementation of joinserver.KeyRegistry.
type KeyRegistry struct {
	Redis *ttnredis.Client
}

// GetByID gets session keys by devEUI, id.
func (r *KeyRegistry) GetByID(ctx context.Context, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
	if devEUI.IsZero() || len(id) == 0 {
		return nil, errInvalidIdentifiers
	}

	pb := &ttnpb.SessionKeys{}
	if err := ttnredis.GetProto(r.Redis, r.Redis.Key(uidKey, devEUI.String(), base64.RawStdEncoding.EncodeToString(id))).ScanProto(pb); err != nil {
		return nil, err
	}
	return applyKeyFieldMask(&ttnpb.SessionKeys{}, pb, paths...)
}

// SetByID sets session keys by devEUI, id.
func (r *KeyRegistry) SetByID(ctx context.Context, devEUI types.EUI64, id []byte, gets []string, f func(*ttnpb.SessionKeys) (*ttnpb.SessionKeys, []string, error)) (*ttnpb.SessionKeys, error) {
	if devEUI.IsZero() || len(id) == 0 {
		return nil, errInvalidIdentifiers
	}

	uk := r.Redis.Key(uidKey, devEUI.String(), base64.RawStdEncoding.EncodeToString(id))

	var pb *ttnpb.SessionKeys
	err := r.Redis.Watch(func(tx *redis.Tx) error {
		cmd := ttnredis.GetProto(tx, uk)
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
			pb, err = applyKeyFieldMask(nil, pb, gets...)
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

		var f func(redis.Pipeliner) error
		if pb == nil {
			f = func(p redis.Pipeliner) error {
				p.Del(uk)
				return nil
			}
		} else {
			if !bytes.Equal(pb.SessionKeyID, id) {
				return errInvalidIdentifiers
			}

			updated := &ttnpb.SessionKeys{}
			if stored != nil {
				if err := cmd.ScanProto(updated); err != nil {
					return err
				}
			}
			updated, err = applyKeyFieldMask(updated, pb, sets...)
			if err != nil {
				return err
			}

			pb, err = applyKeyFieldMask(nil, updated, gets...)
			if err != nil {
				return err
			}
			f = func(p redis.Pipeliner) error {
				_, err := ttnredis.SetProto(p, uk, updated, 0)
				if err != nil {
					return err
				}
				return nil
			}
		}
		_, err = tx.Pipelined(f)
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
