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
	"go.thethings.network/lorawan-stack/pkg/unique"
)

const webhookKey = "webhook"

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
	// TODO: Apply field mask once generator is ready (https://github.com/TheThingsIndustries/lorawan-stack/issues/1212)
	return pb, nil
}

// List implements WebhookRegistry.
func (r WebhookRegistry) List(ctx context.Context, ids ttnpb.ApplicationIdentifiers, paths []string) ([]*ttnpb.ApplicationWebhook, error) {
	var pbs []*ttnpb.ApplicationWebhook
	k := r.Redis.Key(webhookKey, unique.ID(ctx, ids))
	keyCmd := func(ks ...string) string {
		return r.Redis.Key(append([]string{webhookKey, unique.ID(ctx, ids)}, ks...)...)
	}
	err := ttnredis.FindProtos(r.Redis, k, keyCmd).Range(func() (proto.Message, func() bool) {
		pb := &ttnpb.ApplicationWebhook{}
		return pb, func() bool {
			// TODO: Apply field mask once generator is ready (https://github.com/TheThingsIndustries/lorawan-stack/issues/1212)
			pbs = append(pbs, pb)
			return true
		}
	})
	if err != nil {
		return nil, err
	}
	return pbs, nil
}

// Set implements WebhookRegistry.
func (r WebhookRegistry) Set(ctx context.Context, ids ttnpb.ApplicationWebhookIdentifiers, paths []string, f func(*ttnpb.ApplicationWebhook) (*ttnpb.ApplicationWebhook, []string, error)) (*ttnpb.ApplicationWebhook, error) {
	k := r.Redis.Key(webhookKey, unique.ID(ctx, ids.ApplicationIdentifiers), ids.WebhookID)
	var pb *ttnpb.ApplicationWebhook
	err := r.Redis.Watch(func(tx *redis.Tx) error {
		var create bool
		pb = &ttnpb.ApplicationWebhook{}
		if err := ttnredis.GetProto(tx, k).ScanProto(pb); errors.IsNotFound(err) {
			create = true
			pb = nil
		} else if err != nil {
			return err
		}
		createdAt := pb.GetCreatedAt()
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
				p.SRem(r.Redis.Key(webhookKey, unique.ID(ctx, ids.ApplicationIdentifiers)), ids.WebhookID)
				return nil
			}
		} else {
			pb.ApplicationWebhookIdentifiers = ids
			pb.UpdatedAt = time.Now().UTC()
			if create {
				pb.CreatedAt = pb.UpdatedAt
			} else {
				pb.CreatedAt = createdAt
			}
			f = func(p redis.Pipeliner) error {
				ttnredis.SetProto(p, k, pb, 0)
				p.SAdd(r.Redis.Key(webhookKey, unique.ID(ctx, ids.ApplicationIdentifiers)), ids.WebhookID)
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
