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
	"fmt"
	"sync"

	"github.com/go-redis/redis"
	"go.thethings.network/lorawan-stack/pkg/errors"
	ttnredis "go.thethings.network/lorawan-stack/pkg/redis"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

type ApplicationUplinkQueue struct {
	Redis  *ttnredis.Client
	MaxLen int64
	Group  string
	ID     string

	subscriptions sync.Map
}

// NewApplicationUplinkQueue returns new application uplink queue.
func NewApplicationUplinkQueue(cl *ttnredis.Client, maxLen int64, group, id string) *ApplicationUplinkQueue {
	return &ApplicationUplinkQueue{
		Redis:  cl,
		MaxLen: maxLen,
		Group:  group,
		ID:     id,
	}
}

func (q *ApplicationUplinkQueue) uidUplinkKey(uid string) string {
	return q.Redis.Key("uid", uid, "uplinks")
}

const payloadKey = "payload"

func (q *ApplicationUplinkQueue) Add(ctx context.Context, ups ...*ttnpb.ApplicationUp) error {
	for _, up := range ups {
		uid := unique.ID(ctx, up.ApplicationIdentifiers)

		s, err := ttnredis.MarshalProto(up)
		if err != nil {
			return err
		}

		if err = q.Redis.XAdd(&redis.XAddArgs{
			Stream:       q.uidUplinkKey(uid),
			MaxLenApprox: q.MaxLen,
			Values: map[string]interface{}{
				payloadKey: s,
			},
		}).Err(); err != nil {
			return ttnredis.ConvertError(err)
		}

		upCh, ok := q.subscriptions.Load(uid)
		if ok {
			select {
			case upCh.(chan *ttnpb.ApplicationUp) <- up:
			default:
			}
		}
	}
	return nil
}

var (
	errInvalidPayload = errors.DefineCorruption("invalid_payload", "invalid payload")
	errMissingPayload = errors.DefineDataLoss("missing_payload", "missing payload")
)

// Subscribe ranges over q.upStream(unique.ID(ctx, appID)) using f until ctx is done.
// Subscribe assumes that there's at most 1 active consumer in q.Group per stream at all times.
func (q *ApplicationUplinkQueue) Subscribe(ctx context.Context, appID ttnpb.ApplicationIdentifiers, f func(context.Context, *ttnpb.ApplicationUp) error) error {
	uid := unique.ID(ctx, appID)
	upStream := q.uidUplinkKey(uid)

	_, err := q.Redis.XGroupCreateMkStream(upStream, q.Group, "0").Result()
	if err != nil && !ttnredis.IsConsumerGroupExistsErr(err) {
		return ttnredis.ConvertError(err)
	}

	upCh := make(chan *ttnpb.ApplicationUp, 1)
	_, ok := q.subscriptions.LoadOrStore(uid, upCh)
	if ok {
		panic(fmt.Sprintf("duplicate subscription for application %s", uid))
	}
	defer q.subscriptions.Delete(uid)

	for {
		rets, err := q.Redis.XReadGroup(&redis.XReadGroupArgs{
			Group:    q.Group,
			Consumer: q.ID,
			Streams:  []string{upStream, upStream, "0", ">"},
		}).Result()
		if err != nil && err != redis.Nil {
			return ttnredis.ConvertError(err)
		}
		for _, ret := range rets {
			switch ret.Stream {
			case upStream:
				for _, msg := range ret.Messages {
					v, ok := msg.Values[payloadKey]
					if !ok {
						return errMissingPayload
					}
					s, ok := v.(string)
					if !ok {
						return errInvalidPayload
					}
					up := &ttnpb.ApplicationUp{}
					if err = ttnredis.UnmarshalProto(s, up); err != nil {
						return err
					}
					if err = f(ctx, up); err != nil {
						return err
					}
					if err = q.Redis.XAck(upStream, q.Group, msg.ID).Err(); err != nil {
						return ttnredis.ConvertError(err)
					}
				}
			default:
				panic(fmt.Sprintf("unknown stream read %s", ret.Stream))
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-upCh:
		}
	}
}
