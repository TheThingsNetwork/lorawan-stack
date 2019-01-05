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
	"time"

	"github.com/go-redis/redis"
	"github.com/gogo/protobuf/proto"
	"go.thethings.network/lorawan-stack/pkg/errors"
	ttnredis "go.thethings.network/lorawan-stack/pkg/redis"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

const webhookKey = "webhook"

func applyWebhookFieldMask(dst, src *ttnpb.ApplicationWebhook, paths ...string) (*ttnpb.ApplicationWebhook, error) {
	if dst == nil {
		dst = &ttnpb.ApplicationWebhook{}
	}
	return dst, dst.SetFields(src, append(paths, "ids")...)
}

// WebhookRegistry is a Redis webhook registry.
type WebhookRegistry struct {
	Redis *ttnredis.Client
}

// Get implements WebhookRegistry.
func (r WebhookRegistry) Get(ctx context.Context, ids ttnpb.ApplicationWebhookIdentifiers, paths []string) (*ttnpb.ApplicationWebhook, error) {
	k := r.Redis.Key(webhookKey, unique.ID(ctx, ids.ApplicationIdentifiers), ids.WebhookID)
	pb := &ttnpb.ApplicationWebhook{}
	if err := ttnredis.GetProto(r.Redis, k).ScanProto(pb); err != nil {
		return nil, err
	}
	return applyWebhookFieldMask(nil, pb, paths...)
}

// List implements WebhookRegistry.
func (r WebhookRegistry) List(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) ([]*ttnpb.ApplicationWebhook, error) {
	var pbs []*ttnpb.ApplicationWebhook
	k := r.Redis.Key(webhookKey, unique.ID(ctx, ids))
	keyCmd := func(ks ...string) string {
		return r.Redis.Key(append([]string{webhookKey, unique.ID(ctx, ids)}, ks...)...)
	}
	err := ttnredis.FindProtos(r.Redis, k, keyCmd).Range(func() (proto.Message, func() (bool, error)) {
		pb := &ttnpb.ApplicationWebhook{}
		return pb, func() (bool, error) {
			pb, err := applyWebhookFieldMask(nil, pb, paths...)
			if err != nil {
				return false, err
			}
			pbs = append(pbs, pb)
			return true, nil
		}
	})
	if err != nil {
		return nil, err
	}
	return pbs, nil
}

// Set implements WebhookRegistry.
func (r WebhookRegistry) Set(ctx context.Context, ids ttnpb.ApplicationWebhookIdentifiers, gets []string, f func(*ttnpb.ApplicationWebhook) (*ttnpb.ApplicationWebhook, []string, error)) (*ttnpb.ApplicationWebhook, error) {
	k := r.Redis.Key(webhookKey, unique.ID(ctx, ids.ApplicationIdentifiers), ids.WebhookID)
	var pb *ttnpb.ApplicationWebhook
	err := r.Redis.Watch(func(tx *redis.Tx) error {
		var create bool
		cmd := ttnredis.GetProto(tx, k)
		stored := &ttnpb.ApplicationWebhook{}
		if err := cmd.ScanProto(stored); errors.IsNotFound(err) {
			create = true
			stored = nil
		} else if err != nil {
			return err
		}

		var err error
		if stored != nil {
			pb, err = applyWebhookFieldMask(nil, stored, gets...)
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
				p.SRem(r.Redis.Key(webhookKey, unique.ID(ctx, ids.ApplicationIdentifiers)), ids.WebhookID)
				return nil
			}
		} else {
			pb.ApplicationWebhookIdentifiers = ids
			pb.UpdatedAt = time.Now().UTC()
			sets = append(sets, "updated_at")
			if create {
				pb.CreatedAt = pb.UpdatedAt
				sets = append(sets, "created_at")
			}
			stored = &ttnpb.ApplicationWebhook{}
			if err := cmd.ScanProto(stored); err != nil && !errors.IsNotFound(err) {
				return err
			}
			stored, err = applyWebhookFieldMask(stored, pb, sets...)
			if err != nil {
				return err
			}
			pb, err = applyWebhookFieldMask(nil, stored, gets...)
			if err != nil {
				return err
			}
			f = func(p redis.Pipeliner) error {
				_, err := ttnredis.SetProto(p, k, stored, 0)
				if err != nil {
					return err
				}
				p.SAdd(r.Redis.Key(webhookKey, unique.ID(ctx, ids.ApplicationIdentifiers)), ids.WebhookID)
				return nil
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
