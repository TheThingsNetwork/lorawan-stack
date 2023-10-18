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
	"sync"

	"github.com/redis/go-redis/v9"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

type contextualUplinkBatch struct {
	ctx        context.Context
	confirmIDs []string
	uplinks    []*ttnpb.ApplicationUp
}

const (
	payloadKey    = "payload"
	payloadUIDKey = "uid"
)

var (
	errMissingPayload = errors.DefineDataLoss("missing_payload", "missing payload")
	errInvalidPayload = errors.DefineCorruption("invalid_payload", "invalid payload")
	errMissingUID     = errors.DefineDataLoss("missing_uid", "missing UID")
	errInvalidUID     = errors.DefineCorruption("invalid_uid", "invalid UID")
)

// ApplicationUplinkQueue is an implementation of ApplicationUplinkQueue.
type ApplicationUplinkQueue struct {
	redis            *ttnredis.Client
	maxLen           int64
	groupID          string
	streamID         string
	minIdle          time.Duration
	streamBlockLimit time.Duration
	consumers        sync.Map
}

// NewApplicationUplinkQueue returns new application uplink queue.
func NewApplicationUplinkQueue(
	cl *ttnredis.Client,
	maxLen int64,
	groupID string,
	minIdle time.Duration,
	streamBlockLimit time.Duration,
) *ApplicationUplinkQueue {
	return &ApplicationUplinkQueue{
		redis:            cl,
		maxLen:           maxLen,
		groupID:          groupID,
		streamID:         cl.Key("uplinks"),
		minIdle:          minIdle,
		streamBlockLimit: streamBlockLimit,
		consumers:        sync.Map{},
	}
}

func ApplicationUplinkQueueUIDGenericUplinkKey(r keyer, uid string) string {
	return ttnredis.Key(UIDKey(r, uid), "uplinks")
}

// Init initializes the ApplicationUplinkQueue.
func (q *ApplicationUplinkQueue) Init(ctx context.Context) error {
	cmd := q.redis.XGroupCreateMkStream(ctx, q.streamID, q.groupID, "0")
	if err := cmd.Err(); err != nil && !ttnredis.IsConsumerGroupExistsErr(err) {
		return ttnredis.ConvertError(err)
	}
	return nil
}

// Close removes all consumers from the consumer group.
func (q *ApplicationUplinkQueue) Close(ctx context.Context) error {
	pipeline := q.redis.Pipeline()
	q.consumers.Range(func(key, value any) bool {
		pipeline.XGroupDelConsumer(ctx, q.streamID, q.groupID, key.(string))
		return true
	})
	if _, err := pipeline.Exec(ctx); err != nil {
		return ttnredis.ConvertError(err)
	}
	return nil
}

// Add implements ApplicationUplinkQueue interface.
func (q *ApplicationUplinkQueue) Add(ctx context.Context, ups ...*ttnpb.ApplicationUp) error {
	if len(ups) == 0 {
		return nil
	}
	_, err := q.redis.Pipelined(ctx, func(p redis.Pipeliner) error {
		for _, up := range ups {
			s, err := ttnredis.MarshalProto(up)
			if err != nil {
				return err
			}
			p.XAdd(ctx, &redis.XAddArgs{
				Stream: q.streamID,
				MaxLen: q.maxLen,
				Approx: true,
				Values: map[string]any{
					payloadUIDKey: unique.ID(ctx, up.EndDeviceIds),
					payloadKey:    s,
				},
			})
		}
		return nil
	})
	if err != nil {
		return ttnredis.ConvertError(err)
	}
	return nil
}

func uidStrFrom(values map[string]any) (string, error) {
	uidValue, ok := values[payloadUIDKey]
	if !ok {
		return "", errMissingUID.New()
	}
	uid, ok := uidValue.(string)
	if !ok {
		return "", errInvalidUID.New()
	}
	return uid, nil
}

func applicationUpFrom(values map[string]any) (*ttnpb.ApplicationUp, error) {
	payloadValue, ok := values[payloadKey]
	if !ok {
		return nil, errMissingPayload.New()
	}
	payload, ok := payloadValue.(string)
	if !ok {
		return nil, errInvalidPayload.New()
	}
	up := &ttnpb.ApplicationUp{}
	if err := ttnredis.UnmarshalProto(payload, up); err != nil {
		return nil, errInvalidPayload.WithCause(err)
	}
	return up, nil
}

// Pop implements ApplicationUplinkQueue interface.
func (q *ApplicationUplinkQueue) Pop(
	ctx context.Context, consumerID string, limit int,
	f func(context.Context, []*ttnpb.ApplicationUp) error,
) error {
	q.consumers.Store(consumerID, struct{}{})

	msgs, _, err := q.redis.XAutoClaim(ctx, &redis.XAutoClaimArgs{
		Group:    q.groupID,
		Consumer: consumerID,
		Stream:   q.streamID,
		Start:    "-",
		MinIdle:  q.minIdle,
		Count:    int64(limit),
	}).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return ttnredis.ConvertError(err)
	}

	remainingCount := limit - len(msgs)
	if remainingCount <= 0 {
		remainingCount = 0
	}

	streams, err := q.redis.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    q.groupID,
		Consumer: consumerID,
		Streams:  []string{q.streamID, ">"},
		Count:    int64(remainingCount),
		Block:    q.streamBlockLimit,
	}).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return ttnredis.ConvertError(err)
	}
	batches := map[string]*contextualUplinkBatch{}
	if len(streams) > 0 {
		stream := streams[0]
		msgs = append(msgs, stream.Messages...)
	}

	for _, msg := range msgs {
		uid, err := uidStrFrom(msg.Values)
		if err != nil {
			return err
		}
		up, err := applicationUpFrom(msg.Values)
		if err != nil {
			return err
		}
		if err := addToBatch(ctx, batches, msg.ID, uid, up); err != nil {
			return err
		}
	}
	pipeliner := q.redis.Pipeline()
	for _, batch := range batches {
		if err := f(batch.ctx, batch.uplinks); err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to process uplink batch")
			continue // Do not confirm messages that failed to process.
		}

		pipeliner.XAck(ctx, q.streamID, q.groupID, batch.confirmIDs...)
		pipeliner.XDel(ctx, q.streamID, batch.confirmIDs...)
	}
	if _, err := pipeliner.Exec(ctx); err != nil {
		return ttnredis.ConvertError(err)
	}
	return nil
}
