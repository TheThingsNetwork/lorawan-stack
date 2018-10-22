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
	"github.com/gogo/protobuf/proto"
	"go.thethings.network/lorawan-stack/pkg/errors"
	ttnredis "go.thethings.network/lorawan-stack/pkg/redis"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

const (
	addrKey = "addr"
	euiKey  = "eui"
)

var (
	errInvalidIdentifiers   = errors.DefineInvalidArgument("invalid_identifiers", "invalid identifiers")
	errDuplicateIdentifiers = errors.DefineAlreadyExists("duplicate_identifiers", "duplicate identifiers")
)

type DeviceRegistry struct {
	Redis *ttnredis.Client
}

func (r *DeviceRegistry) GetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string) (*ttnpb.EndDevice, error) {
	ids := ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: appID,
		DeviceID:               devID,
	}
	if err := ids.Validate(); err != nil {
		return nil, err
	}
	dev := &ttnpb.EndDevice{}
	if err := ttnredis.GetProto(r.Redis, r.Redis.Key(unique.ID(ctx, ids))).ScanProto(dev); err != nil {
		return nil, err
	}
	return dev, nil
}

func (r *DeviceRegistry) GetByEUI(_ context.Context, joinEUI, devEUI types.EUI64) (*ttnpb.EndDevice, error) {
	dev := &ttnpb.EndDevice{}
	if err := ttnredis.FindProto(r.Redis, r.Redis.Key(euiKey, joinEUI.String(), devEUI.String()), r.Redis.Key).ScanProto(dev); err != nil {
		return nil, err
	}
	return dev, nil
}

func (r *DeviceRegistry) RangeByAddr(addr types.DevAddr, f func(*ttnpb.EndDevice) bool) error {
	return ttnredis.FindProtos(r.Redis, r.Redis.Key(addrKey, addr.String()), r.Redis.Key).Range(func() (proto.Message, func() bool) {
		dev := &ttnpb.EndDevice{}
		return dev, func() bool { return f(dev) }
	})
}

func getDevAddrsAndIDs(dev *ttnpb.EndDevice) (addrs struct{ current, fallback *types.DevAddr }, ids ttnpb.EndDeviceIdentifiers) {
	if dev == nil {
		return
	}

	if dev.Session != nil {
		var addr types.DevAddr
		copy(addr[:], dev.Session.DevAddr[:])
		addrs.current = &addr
	}
	if dev.PendingSession != nil {
		var addr types.DevAddr
		copy(addr[:], dev.PendingSession.DevAddr[:])
		addrs.fallback = &addr
	}
	return addrs, *dev.EndDeviceIdentifiers.Copy(&ttnpb.EndDeviceIdentifiers{})
}

func equalAddr(x, y *types.DevAddr) bool {
	if x == nil || y == nil {
		return x == y
	}
	return x.Equal(*y)
}

func equalEUI(x, y *types.EUI64) bool {
	if x == nil || y == nil {
		return x == y
	}
	return x.Equal(*y)
}

func (r *DeviceRegistry) SetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, f func(dev *ttnpb.EndDevice) (*ttnpb.EndDevice, error)) (*ttnpb.EndDevice, error) {
	ids := ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: appID,
		DeviceID:               devID,
	}
	if err := ids.Validate(); err != nil {
		return nil, err
	}
	uid := unique.ID(ctx, ids)
	k := r.Redis.Key(uid)

	var dev *ttnpb.EndDevice
	err := r.Redis.Watch(func(tx *redis.Tx) error {
		var create bool
		dev = &ttnpb.EndDevice{}
		if err := ttnredis.GetProto(tx, k).ScanProto(dev); errors.IsNotFound(err) {
			create = true
			dev = nil
		} else if err != nil {
			return err
		}

		createdAt := dev.GetCreatedAt()
		oldAddrs, oldIDs := getDevAddrsAndIDs(dev)

		var err error
		dev, err = f(dev)
		if err != nil {
			return err
		}

		var f func(redis.Pipeliner) error
		if dev == nil {
			f = func(p redis.Pipeliner) error {
				p.Del(k)
				if oldIDs.JoinEUI != nil && oldIDs.DevEUI != nil {
					p.Del(r.Redis.Key(euiKey, oldIDs.JoinEUI.String(), oldIDs.DevEUI.String()))
				}
				if oldAddrs.fallback != nil {
					p.SRem(r.Redis.Key(addrKey, oldAddrs.fallback.String()), uid)
				}
				if oldAddrs.current != nil {
					p.SRem(r.Redis.Key(addrKey, oldAddrs.current.String()), uid)
				}
				return nil
			}
		} else {
			newAddrs, newIDs := getDevAddrsAndIDs(dev)

			if !create && (!equalEUI(oldIDs.JoinEUI, newIDs.JoinEUI) || !equalEUI(oldIDs.DevEUI, newIDs.DevEUI) ||
				oldIDs.ApplicationIdentifiers != newIDs.ApplicationIdentifiers || oldIDs.DeviceID != newIDs.DeviceID) {
				return errInvalidIdentifiers
			}

			dev.UpdatedAt = time.Now().UTC()
			if create {
				dev.CreatedAt = dev.UpdatedAt
			} else {
				dev.CreatedAt = createdAt
			}

			f = func(p redis.Pipeliner) error {
				if create && newIDs.JoinEUI != nil && newIDs.DevEUI != nil {
					ek := r.Redis.Key(euiKey, newIDs.JoinEUI.String(), newIDs.DevEUI.String())
					if err := tx.Watch(ek).Err(); err != nil {
						return err
					}
					s, err := tx.Get(ek).Result()
					if err != nil && err != redis.Nil {
						return ttnredis.ConvertError(err)
					}
					if err == nil && s != "" {
						return errDuplicateIdentifiers
					}
					p.Set(ek, uid, 0)
				}

				ttnredis.SetProto(p, k, dev, 0)

				if oldAddrs.fallback != nil && !equalAddr(oldAddrs.fallback, newAddrs.fallback) && !equalAddr(oldAddrs.fallback, newAddrs.current) {
					p.SRem(r.Redis.Key(addrKey, oldAddrs.fallback.String()), uid)
				}
				if oldAddrs.current != nil && !equalAddr(oldAddrs.current, newAddrs.fallback) && !equalAddr(oldAddrs.current, newAddrs.current) {
					p.SRem(r.Redis.Key(addrKey, oldAddrs.current.String()), uid)
				}
				if newAddrs.fallback != nil && !equalAddr(newAddrs.fallback, oldAddrs.fallback) && !equalAddr(newAddrs.fallback, oldAddrs.current) {
					p.SAdd(r.Redis.Key(addrKey, newAddrs.fallback.String()), uid)
				}
				if newAddrs.current != nil && !equalAddr(newAddrs.current, oldAddrs.fallback) && !equalAddr(newAddrs.current, oldAddrs.current) {
					p.SAdd(r.Redis.Key(addrKey, newAddrs.current.String()), uid)
				}
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
	return dev, nil
}
