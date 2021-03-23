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

	"github.com/go-redis/redis/v8"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/time"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// DownlinkTaskQueue is an implementation of networkserver.DownlinkTaskQueue.
type DownlinkTaskQueue struct {
	queue *ttnredis.TaskQueue
}

const (
	downlinkKey = "downlink"
)

// NewDownlinkTaskQueue returns new downlink task queue.
func NewDownlinkTaskQueue(cl *ttnredis.Client, maxLen int64, group, id string) *DownlinkTaskQueue {
	return &DownlinkTaskQueue{
		queue: &ttnredis.TaskQueue{
			Redis:  cl,
			MaxLen: maxLen,
			Group:  group,
			ID:     id,
			Key:    cl.Key(downlinkKey),
		},
	}
}

// Init initializes the DownlinkTaskQueue.
func (q *DownlinkTaskQueue) Init(ctx context.Context) error {
	return q.queue.Init(ctx)
}

// Close closes the DownlinkTaskQueue.
func (q *DownlinkTaskQueue) Close(ctx context.Context) error {
	return q.queue.Close(ctx)
}

// Add adds downlink task for device identified by devID at time startAt.
func (q *DownlinkTaskQueue) Add(ctx context.Context, devID ttnpb.EndDeviceIdentifiers, startAt time.Time, replace bool) error {
	return q.queue.Add(ctx, nil, unique.ID(ctx, devID), startAt, replace)
}

// Pop calls f on the earliest downlink task in the schedule, for which timestamp is in range [0, time.Now()],
// if such is available, otherwise it blocks until it is.
func (q *DownlinkTaskQueue) Pop(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) (time.Time, error)) error {
	return q.queue.Pop(ctx, nil, func(p redis.Pipeliner, uid string, startAt time.Time) error {
		ids, err := unique.ToDeviceID(uid)
		if err != nil {
			return err
		}
		ctx, err := unique.WithContext(ctx, uid)
		if err != nil {
			return err
		}
		t, err := f(ctx, ids, startAt)
		if err != nil || t.IsZero() {
			return err
		}
		return q.queue.Add(ctx, p, uid, t, true)
	})
}
