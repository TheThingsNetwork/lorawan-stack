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
	"crypto/rand"
	"time"

	"github.com/go-redis/redis"
	ulid "github.com/oklog/ulid/v2"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	ttnredis "go.thethings.network/lorawan-stack/pkg/redis"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

type ApplicationUplinkQueue struct {
	Redis  *ttnredis.Client
	MaxLen int64
	Group  string
	ID     string
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

func (q *ApplicationUplinkQueue) uidCloseKey(uid, streamID string) string {
	return q.Redis.Key("uid", uid, "close", streamID)
}

const payloadKey = "payload"

func (q *ApplicationUplinkQueue) Add(ctx context.Context, ups ...*ttnpb.ApplicationUp) error {
	for _, up := range ups {
		s, err := ttnredis.MarshalProto(up)
		if err != nil {
			return err
		}
		if err := q.Redis.XAdd(&redis.XAddArgs{
			Stream:       q.uidUplinkKey(unique.ID(ctx, up.ApplicationIdentifiers)),
			MaxLenApprox: q.MaxLen,
			Values: map[string]interface{}{
				payloadKey: s,
			},
		}).Err(); err != nil {
			return ttnredis.ConvertError(err)
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
	streamULID, err := ulid.New(ulid.Now(), rand.Reader)
	if err != nil {
		return err
	}
	uid := unique.ID(ctx, appID)
	upStream := q.uidUplinkKey(uid)
	closeStream := q.uidCloseKey(uid, streamULID.String())

	_, err = q.Redis.XGroupCreateMkStream(closeStream, q.Group, "0").Result()
	if err != nil {
		return ttnredis.ConvertError(err)
	}

	dl, hasDL := ctx.Deadline()
	if hasDL {
		_, err = q.Redis.ExpireAt(closeStream, dl).Result()
		if err != nil {
			return ttnredis.ConvertError(err)
		}
	}

	doneCh := make(chan struct{})
	defer func() {
		close(doneCh)
	}()
	go func() {
		logger := log.FromContext(ctx).WithField("key", closeStream)
		select {
		case <-ctx.Done():
			_, err := q.Redis.XAdd(&redis.XAddArgs{
				Stream: closeStream,
				Values: map[string]interface{}{"": ""},
			}).Result()
			if err != nil {
				logger.WithError(err).Error("Failed to add message to Redis stream")
				return
			}
			<-doneCh

		case <-doneCh:
		}

		_, err := q.Redis.Del(closeStream).Result()
		if err != nil {
			logger.WithError(err).Error("Failed to delete Redis key")
		}
	}()

	_, err = q.Redis.XGroupCreateMkStream(upStream, q.Group, "0").Result()
	if err != nil && !ttnredis.IsConsumerGroupExistsErr(err) {
		return ttnredis.ConvertError(err)
	}

	processStreams := func(arg *redis.XReadGroupArgs) error {
		rets, err := q.Redis.XReadGroup(arg).Result()
		if err != nil {
			return ttnredis.ConvertError(err)
		}

		for _, ret := range rets {
			switch ret.Stream {
			case closeStream:
				return ctx.Err()

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
					_, err = q.Redis.XAck(upStream, q.Group, msg.ID).Result()
					if err != nil {
						return ttnredis.ConvertError(err)
					}
				}
			}
		}
		return nil
	}

	if err = processStreams(&redis.XReadGroupArgs{
		Group:    q.Group,
		Consumer: q.ID,
		Streams:  []string{upStream, "0"},
	}); err != nil {
		return err
	}
	for {
		var timeout time.Duration
		if hasDL {
			timeout = time.Until(dl)
			if timeout <= 0 {
				return context.DeadlineExceeded
			}
		}
		if err = processStreams(&redis.XReadGroupArgs{
			Group:    q.Group,
			Consumer: q.ID,
			Streams:  []string{closeStream, upStream, ">", ">"},
			Block:    timeout,
		}); err != nil {
			return err
		}
	}
}
