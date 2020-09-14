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

func ApplicationUplinkQueueUIDGenericUplinkKey(cl *ttnredis.Client, uid string) string {
	return cl.Key("uid", uid, "uplinks")
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

const payloadKey = "payload"

func (q *ApplicationUplinkQueue) Add(ctx context.Context, ups ...*ttnpb.ApplicationUp) error {
	for _, up := range ups {
		uid := unique.ID(ctx, up.ApplicationIdentifiers)

		s, err := ttnredis.MarshalProto(up)
		if err != nil {
			return err
		}

		var streamID string
		var pipelined func(p redis.Pipeliner)
		switch pld := up.Up.(type) {
		case *ttnpb.ApplicationUp_JoinAccept:
			streamID = q.uidJoinAcceptKey(uid)
		case *ttnpb.ApplicationUp_DownlinkQueueInvalidated:
			streamID = q.uidInvalidationKey(uid)
			pipelined = func(p redis.Pipeliner) {
				p.Set(deviceUIDLastInvalidationKey(q.redis, unique.ID(ctx, up.EndDeviceIdentifiers)), pld.DownlinkQueueInvalidated.LastFCntDown, 0)
			}
		default:
			streamID = q.uidGenericUplinkKey(uid)
		}
		_, err = q.redis.Pipelined(func(p redis.Pipeliner) error {
			p.XAdd(&redis.XAddArgs{
				Stream:       streamID,
				MaxLenApprox: q.maxLen,
				Values: map[string]interface{}{
					payloadKey: s,
				},
			})
			if pipelined != nil {
				pipelined(p)
			}
			return nil
		})
		if err != nil {
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

func (q *ApplicationUplinkQueue) initConsumer(ctx context.Context, streamID string) (func(), error) {
	_, err := q.redis.XGroupCreateMkStream(streamID, q.group, "0").Result()
	if err != nil && !ttnredis.IsConsumerGroupExistsErr(err) {
		return nil, ttnredis.ConvertError(err)
	}
	return func() {
		if err := q.redis.XGroupDelConsumer(streamID, q.group, q.id).Err(); err != nil {
			log.FromContext(ctx).WithError(err).WithFields(log.Fields(
				"consumer", q.id,
				"group", q.group,
				"stream", streamID,
			)).Error("Failed to delete application uplink queue redis consumer")
		}
	}, nil
}

// Subscribe ranges over uplink keys using f until ctx is done.
// Subscribe assumes that there's at most 1 active consumer in q.group per stream at all times.
func (q *ApplicationUplinkQueue) Subscribe(ctx context.Context, appID ttnpb.ApplicationIdentifiers, f func(context.Context, *ttnpb.ApplicationUp) error) error {
	uid := unique.ID(ctx, appID)

	genericUpStream := q.uidGenericUplinkKey(uid)
	invalidationUpStream := q.uidInvalidationKey(uid)
	joinAcceptUpStream := q.uidJoinAcceptKey(uid)

	for _, streamID := range [...]string{
		genericUpStream,
		invalidationUpStream,
		joinAcceptUpStream,
	} {
		delConsumer, err := q.initConsumer(ctx, streamID)
		if err != nil {
			return err
		}
		defer delConsumer()
	}

	upCh := make(chan *ttnpb.ApplicationUp, 1)
	_, ok := q.subscriptions.LoadOrStore(uid, upCh)
	if ok {
		panic(fmt.Sprintf("duplicate subscription for application %s", uid))
	}
	defer q.subscriptions.Delete(uid)
	for {
		rets, err := q.redis.XReadGroup(&redis.XReadGroupArgs{
			Group:    q.group,
			Consumer: q.id,
			Streams:  []string{joinAcceptUpStream, joinAcceptUpStream, invalidationUpStream, invalidationUpStream, genericUpStream, genericUpStream, "0", ">", "0", ">", "0", ">"},
			Count:    1,
		}).Result()
		if err != nil && err != redis.Nil {
			return ttnredis.ConvertError(err)
		}
		var invalidationFCnts map[string]uint64
		for _, ret := range rets {
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
				var skipF bool
				if ret.Stream == invalidationUpStream {
					devUID := unique.ID(ctx, up.EndDeviceIdentifiers)
					lastFCnt, ok := invalidationFCnts[devUID]
					if !ok {
						lastFCnt, err = q.redis.Get(deviceUIDLastInvalidationKey(q.redis, devUID)).Uint64()
						if err != nil {
							return ttnredis.ConvertError(err)
						}
						if invalidationFCnts == nil {
							invalidationFCnts = make(map[string]uint64, len(rets))
						}
						invalidationFCnts[devUID] = lastFCnt
					}
					skipF = uint64(up.GetDownlinkQueueInvalidated().GetLastFCntDown()) < lastFCnt
				}
				if !skipF {
					if err = f(ctx, up); err != nil {
						return err
					}
				}
				_, err := q.redis.Pipelined(func(p redis.Pipeliner) error {
					p.XAck(ret.Stream, q.group, msg.ID)
					p.XDel(ret.Stream, msg.ID)
					return nil
				})
				if err != nil {
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
