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

	"github.com/redis/go-redis/v9"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

type ApplicationUplinkQueue struct {
	applicationQueue *ttnredis.TaskQueue

	redis   *ttnredis.Client
	maxLen  int64
	group   string
	key     string
	minIdle time.Duration
}

const (
	payloadKey = "payload"
)

// NewApplicationUplinkQueue returns new application uplink queue.
func NewApplicationUplinkQueue(
	cl *ttnredis.Client,
	maxLen int64,
	group string,
	minIdle time.Duration,
	streamBlockLimit time.Duration,
) *ApplicationUplinkQueue {
	return &ApplicationUplinkQueue{
		applicationQueue: &ttnredis.TaskQueue{
			Redis:            cl,
			MaxLen:           maxLen,
			Group:            group,
			Key:              cl.Key("application"),
			StreamBlockLimit: streamBlockLimit,
		},
		redis:   cl,
		maxLen:  maxLen,
		group:   group,
		key:     cl.Key("application-uplink"),
		minIdle: minIdle,
	}
}

func ApplicationUplinkQueueUIDGenericUplinkKey(r keyer, uid string) string {
	return ttnredis.Key(UIDKey(r, uid), "uplinks")
}

func (q *ApplicationUplinkQueue) uidGenericUplinkKey(uid string) string {
	return ApplicationUplinkQueueUIDGenericUplinkKey(q.redis, uid)
}

func (q *ApplicationUplinkQueue) uidInvalidationKey(uid string) string {
	return ttnredis.Key(q.uidGenericUplinkKey(uid), "invalidation")
}

func (q *ApplicationUplinkQueue) uidJoinAcceptKey(uid string) string {
	return ttnredis.Key(q.uidGenericUplinkKey(uid), "join-accept")
}

// Init initializes the ApplicationUplinkQueue.
func (q *ApplicationUplinkQueue) Init(ctx context.Context) error {
	return q.applicationQueue.Init(ctx)
}

// Close closes the ApplicationUplinkQueue.
func (q *ApplicationUplinkQueue) Close(ctx context.Context) error {
	return q.applicationQueue.Close(ctx)
}

func (q *ApplicationUplinkQueue) Add(ctx context.Context, ups ...*ttnpb.ApplicationUp) error {
	if len(ups) == 0 {
		return nil
	}
	_, err := q.redis.Pipelined(ctx, func(p redis.Pipeliner) error {
		now := time.Now()
		taskMap := map[string]time.Time{}
		for _, up := range ups {
			uid := unique.ID(ctx, up.EndDeviceIds.ApplicationIds)

			s, err := ttnredis.MarshalProto(up)
			if err != nil {
				return err
			}

			var uidStreamID string
			switch up.Up.(type) {
			case *ttnpb.ApplicationUp_JoinAccept:
				uidStreamID = q.uidJoinAcceptKey(uid)
			case *ttnpb.ApplicationUp_DownlinkQueueInvalidated:
				uidStreamID = q.uidInvalidationKey(uid)
			default:
				uidStreamID = q.uidGenericUplinkKey(uid)
			}
			p.XAdd(ctx, &redis.XAddArgs{
				Stream: uidStreamID,
				MaxLen: q.maxLen,
				Approx: true,
				Values: map[string]interface{}{
					payloadKey: s,
				},
			})
			if _, ok := taskMap[uid]; !ok {
				taskMap[uid] = now
			}
		}
		for uid, t := range taskMap {
			if err := q.applicationQueue.Add(ctx, p, uid, t, false); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return ttnredis.ConvertError(err)
	}
	return nil
}

var (
	errInvalidPayload = errors.DefineCorruption("invalid_payload", "invalid payload")
	errMissingPayload = errors.DefineDataLoss("missing_payload", "missing payload")
)

func (q *ApplicationUplinkQueue) Dispatch(ctx context.Context, consumerID string) error {
	return q.applicationQueue.Dispatch(ctx, consumerID, nil)
}

func (q *ApplicationUplinkQueue) Pop(ctx context.Context, consumerID string, f func(context.Context, *ttnpb.ApplicationIdentifiers, networkserver.ApplicationUplinkQueueDrainFunc) (time.Time, error)) error {
	return q.applicationQueue.Pop(ctx, consumerID, nil, func(p redis.Pipeliner, uid string, _ time.Time) error {
		appID, err := unique.ToApplicationID(uid)
		if err != nil {
			return err
		}
		ctx, err := unique.WithContext(ctx, uid)
		if err != nil {
			return err
		}
		joinAcceptUpStream := q.uidJoinAcceptKey(uid)
		invalidationUpStream := q.uidInvalidationKey(uid)
		genericUpStream := q.uidGenericUplinkKey(uid)

		streams := [...]string{
			joinAcceptUpStream,
			invalidationUpStream,
			genericUpStream,
		}

		cmds, err := q.redis.Pipelined(ctx, func(pp redis.Pipeliner) error {
			for _, stream := range streams {
				pp.XGroupCreateMkStream(ctx, stream, q.group, "0")
			}
			return nil
		})
		if err != nil && !ttnredis.IsConsumerGroupExistsErr(err) {
			return ttnredis.ConvertError(err)
		}
		var initErr error
		for i, cmd := range cmds {
			if err := cmd.Err(); err != nil && !ttnredis.IsConsumerGroupExistsErr(err) {
				initErr = err
				continue
			}
			p.XGroupDelConsumer(ctx, streams[i], q.group, consumerID)
		}
		if initErr != nil {
			return ttnredis.ConvertError(initErr)
		}

		t, err := f(ctx, appID, func(limit int, g func(...*ttnpb.ApplicationUp) error) error {
			ups := make([]*ttnpb.ApplicationUp, 0, limit)

			processMessages := func(stream string, msgs ...redis.XMessage) error {
				ups = ups[:0]
				for _, msg := range msgs {
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
					ups = append(ups, up)
				}
				return g(ups...)
			}
			return ttnredis.RangeStreams(ctx, q.redis, q.group, consumerID, int64(limit), q.minIdle, processMessages, streams[:]...)
		})
		if err != nil || t.IsZero() {
			return err
		}
		return q.applicationQueue.Add(ctx, p, uid, t, true)
	})
}
