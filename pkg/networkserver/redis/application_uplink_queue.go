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
	redis     *ttnredis.Client
	maxLen    int64
	groupID   string
	streamID  string
	minIdle   time.Duration
	consumers sync.Map
}

// NewApplicationUplinkQueue returns new application uplink queue.
func NewApplicationUplinkQueue(
	cl *ttnredis.Client,
	maxLen int64,
	groupID string,
	minIdle time.Duration,
) *ApplicationUplinkQueue {
	return &ApplicationUplinkQueue{
		redis:     cl,
		maxLen:    maxLen,
		groupID:   groupID,
		streamID:  cl.Key("uplinks"),
		minIdle:   minIdle,
		consumers: sync.Map{},
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

func addToBatch(
	ctx context.Context,
	m map[string]*contextualUplinkBatch,
	confirmID string,
	uid string,
	up *ttnpb.ApplicationUp,
) error {
	ctx, err := unique.WithContext(ctx, uid)
	if err != nil {
		return errInvalidUID.WithCause(err)
	}
	ids, err := unique.ToDeviceID(uid)
	if err != nil {
		return errInvalidUID.WithCause(err)
	}
	key := unique.ID(ctx, ids.ApplicationIds)
	batch, ok := m[key]
	if !ok {
		batch = &contextualUplinkBatch{
			ctx:        ctx,
			confirmIDs: make([]string, 0),
			uplinks:    make([]*ttnpb.ApplicationUp, 0),
		}
		m[key] = batch
	}
	batch.uplinks = append(batch.uplinks, up)
	batch.confirmIDs = append(batch.confirmIDs, confirmID)
	return nil
}

func (*ApplicationUplinkQueue) processMessages(
	ctx context.Context,
	msgs []redis.XMessage,
	ack func(...string) error,
	f func(context.Context, []*ttnpb.ApplicationUp) error,
) error {
	batches := map[string]*contextualUplinkBatch{}
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
	processedIDs := make([]string, 0, len(msgs))
	for _, batch := range batches {
		if err := f(batch.ctx, batch.uplinks); err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to process uplink batch")
			continue // Do not confirm messages that failed to process.
		}
		processedIDs = append(processedIDs, batch.confirmIDs...)
	}
	return ack(processedIDs...)
}

// Pop implements ApplicationUplinkQueue interface.
func (q *ApplicationUplinkQueue) Pop(
	ctx context.Context, consumerID string, limit int,
	f func(context.Context, []*ttnpb.ApplicationUp) error,
) error {
	q.consumers.Store(consumerID, struct{}{})
	return ttnredis.RangeStreams(
		ctx,
		q.redis,
		q.groupID,
		consumerID,
		int64(limit),
		q.minIdle,
		func(_ string, ack func(...string) error, msgs ...redis.XMessage) error {
			return q.processMessages(ctx, msgs, ack, f)
		},
		q.streamID,
	)
}
