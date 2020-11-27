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

	"github.com/go-redis/redis/v8"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

type ApplicationUplinkQueue struct {
	applicationQueue *ttnredis.TaskQueue

	redis  *ttnredis.Client
	maxLen int64
	group  string
	id     string
	key    string
}

const (
	uidKey     = "uid"
	payloadKey = "payload"
)

// NewApplicationUplinkQueue returns new application uplink queue.
func NewApplicationUplinkQueue(cl *ttnredis.Client, maxLen int64, group, id string) *ApplicationUplinkQueue {
	return &ApplicationUplinkQueue{
		applicationQueue: &ttnredis.TaskQueue{
			Redis:  cl,
			MaxLen: maxLen,
			Group:  group,
			ID:     id,
			Key:    cl.Key("application"),
		},
		redis:  cl,
		maxLen: maxLen,
		group:  group,
		id:     id,
		key:    cl.Key("application-uplink"),
	}
}

func ApplicationUplinkQueueUIDGenericUplinkKey(r keyer, uid string) string {
	return r.Key(uidKey, uid, "uplinks")
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
			uid := unique.ID(ctx, up.ApplicationIdentifiers)

			s, err := ttnredis.MarshalProto(up)
			if err != nil {
				return err
			}

			var uidStreamID string
			switch pld := up.Up.(type) {
			case *ttnpb.ApplicationUp_JoinAccept:
				uidStreamID = q.uidJoinAcceptKey(uid)
			case *ttnpb.ApplicationUp_DownlinkQueueInvalidated:
				uidStreamID = q.uidInvalidationKey(uid)
				p.Set(ctx, deviceUIDLastInvalidationKey(q.redis, unique.ID(ctx, up.EndDeviceIdentifiers)), pld.DownlinkQueueInvalidated.LastFCntDown, 0)
			default:
				uidStreamID = q.uidGenericUplinkKey(uid)
			}
			p.XAdd(ctx, &redis.XAddArgs{
				Stream:       uidStreamID,
				MaxLenApprox: q.maxLen,
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

func (q *ApplicationUplinkQueue) Pop(ctx context.Context, f func(context.Context, ttnpb.ApplicationIdentifiers, networkserver.ApplicationUplinkQueueRangeFunc) (time.Time, error)) error {
	return q.applicationQueue.Pop(ctx, nil, func(p redis.Pipeliner, uid string, _ time.Time) error {
		appID, err := unique.ToApplicationID(uid)
		if err != nil {
			return err
		}
		ctx, err := unique.WithContext(ctx, uid)
		if err != nil {
			return err
		}
		var invalidationFCnts map[string]uint64
		var ups []*ttnpb.ApplicationUp
		msgIDs := make(map[string][]string, 3)
		joinAcceptUpStream := q.uidJoinAcceptKey(uid)
		invalidationUpStream := q.uidInvalidationKey(uid)
		genericUpStream := q.uidGenericUplinkKey(uid)

		streams := [...]string{
			joinAcceptUpStream,
			invalidationUpStream,
			genericUpStream,
		}
		cmds, err := q.redis.Pipelined(ctx, func(p redis.Pipeliner) error {
			for _, stream := range streams {
				p.XGroupCreateMkStream(ctx, stream, q.group, "0")
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
			p.XGroupDelConsumer(ctx, streams[i], q.group, q.id)
		}
		if initErr != nil {
			return ttnredis.ConvertError(initErr)
		}

		t, err := f(ctx, appID, func(limit int, g func(...*ttnpb.ApplicationUp) error) (bool, error) {
			rets, err := q.redis.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    q.group,
				Consumer: q.id,
				Streams:  []string{joinAcceptUpStream, joinAcceptUpStream, invalidationUpStream, invalidationUpStream, genericUpStream, genericUpStream, "0", ">", "0", ">", "0", ">"},
				Count:    int64(limit),
				Block:    -1, // do not block
			}).Result()
			if err != nil {
				return false, ttnredis.ConvertError(err)
			}

			// Clear and/or pre-allocate data structures for reuse.
			var msgCount int
			for _, ret := range rets {
				n := len(ret.Messages)
				if n == 0 {
					continue
				}
				ss, ok := msgIDs[ret.Stream]
				if !ok || cap(ss) < n {
					msgIDs[ret.Stream] = make([]string, 0, n)
				} else {
					msgIDs[ret.Stream] = ss[:0]
				}
				msgCount += n
			}
			if msgCount == 0 {
				return true, nil
			}
			if cap(ups) < msgCount {
				ups = make([]*ttnpb.ApplicationUp, 0, msgCount)
			} else {
				ups = ups[:0]
			}

			for _, ret := range rets {
				for _, msg := range ret.Messages {
					v, ok := msg.Values[payloadKey]
					if !ok {
						return false, errMissingPayload.New()
					}
					s, ok := v.(string)
					if !ok {
						return false, errInvalidPayload.New()
					}
					up := &ttnpb.ApplicationUp{}
					if err = ttnredis.UnmarshalProto(s, up); err != nil {
						return false, err
					}
					var skip bool
					if ret.Stream == invalidationUpStream {
						devUID := unique.ID(ctx, up.EndDeviceIdentifiers)
						lastFCnt, ok := invalidationFCnts[devUID]
						if !ok {
							lastFCnt, err = q.redis.Get(ctx, deviceUIDLastInvalidationKey(q.redis, devUID)).Uint64()
							if err != nil {
								return false, ttnredis.ConvertError(err)
							}
							if invalidationFCnts == nil {
								invalidationFCnts = make(map[string]uint64, len(rets))
							}
							invalidationFCnts[devUID] = lastFCnt
						}
						skip = uint64(up.GetDownlinkQueueInvalidated().GetLastFCntDown()) < lastFCnt
					}
					if !skip {
						ups = append(ups, up)
					}
					msgIDs[ret.Stream] = append(msgIDs[ret.Stream], msg.ID)
				}
			}
			if err = g(ups...); err != nil {
				return false, err
			}
			for streamID, ids := range msgIDs {
				// NOTE: Both calls below copy contents of ids internally.
				p.XAck(ctx, streamID, q.group, ids...)
				p.XDel(ctx, streamID, ids...)
			}
			return msgCount < limit, nil
		})
		if err != nil || t.IsZero() {
			return err
		}
		return q.applicationQueue.Add(ctx, p, uid, t, true)
	})
}
