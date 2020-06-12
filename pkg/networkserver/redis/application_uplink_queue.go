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

	"github.com/go-redis/redis/v7"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

type ApplicationUplinkQueue struct {
	redis  *ttnredis.Client
	maxLen int64
	group  string
	id     string

	subscriptions sync.Map
}

// NewApplicationUplinkQueue returns new application uplink queue.
func NewApplicationUplinkQueue(cl *ttnredis.Client, maxLen int64, group, id string) *ApplicationUplinkQueue {
	return &ApplicationUplinkQueue{
		redis:  cl,
		maxLen: maxLen,
		group:  group,
		id:     id,
	}
}

func (q *ApplicationUplinkQueue) uidUplinkKey(uid string) string {
	return q.redis.Key("uid", uid, "uplinks")
}

const payloadKey = "payload"

func (q *ApplicationUplinkQueue) Add(ctx context.Context, ups ...*ttnpb.ApplicationUp) error {
	for _, up := range ups {
		uid := unique.ID(ctx, up.ApplicationIdentifiers)

		s, err := ttnredis.MarshalProto(up)
		if err != nil {
			return err
		}

		if err = q.redis.XAdd(&redis.XAddArgs{
			Stream:       q.uidUplinkKey(uid),
			MaxLenApprox: q.maxLen,
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

// Subscribe ranges over q.uidUplinkKey(unique.ID(ctx, appID)) using f until ctx is done.
// Subscribe assumes that there's at most 1 active consumer in q.group per stream at all times.
func (q *ApplicationUplinkQueue) Subscribe(ctx context.Context, appID ttnpb.ApplicationIdentifiers, f func(context.Context, *ttnpb.ApplicationUp) error) error {
	uid := unique.ID(ctx, appID)
	upStream := q.uidUplinkKey(uid)

	_, err := q.redis.XGroupCreateMkStream(upStream, q.group, "0").Result()
	if err != nil && !ttnredis.IsConsumerGroupExistsErr(err) {
		return ttnredis.ConvertError(err)
	}

	upCh := make(chan *ttnpb.ApplicationUp, 1)
	_, ok := q.subscriptions.LoadOrStore(uid, upCh)
	if ok {
		panic(fmt.Sprintf("duplicate subscription for application %s", uid))
	}
	defer q.subscriptions.Delete(uid)
	defer func() {
		if err := q.redis.XGroupDelConsumer(upStream, q.group, q.id).Err(); err != nil {
			log.FromContext(ctx).WithError(err).WithFields(log.Fields(
				"consumer", q.id,
				"group", q.group,
				"stream", upStream,
			)).Error("Failed to delete application uplink queue redis consumer")
		}
	}()

	for {
		rets, err := q.redis.XReadGroup(&redis.XReadGroupArgs{
			Group:    q.group,
			Consumer: q.id,
			Streams:  []string{upStream, upStream, "0", ">"},
		}).Result()
		if err != nil && err != redis.Nil {
			return ttnredis.ConvertError(err)
		}
		for _, ret := range rets {
			if ret.Stream != upStream {
				panic(fmt.Sprintf("unknown stream read %s", ret.Stream))
			}

			for _, msg := range ret.Messages {
				v, ok := msg.Values[payloadKey]
				if !ok {
					return errMissingPayload.New()
				}
				s, ok := v.(string)
				if !ok {
					return errInvalidPayload.New()
				}
				up := &ttnpb.ApplicationUp{}
				if err = ttnredis.UnmarshalProto(s, up); err != nil {
					return err
				}
				if err = f(ctx, up); err != nil {
					return err
				}
				if err = q.redis.XAck(upStream, q.group, msg.ID).Err(); err != nil {
					return ttnredis.ConvertError(err)
				}
			}
		}

		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-upCh:
		}
	}
}
