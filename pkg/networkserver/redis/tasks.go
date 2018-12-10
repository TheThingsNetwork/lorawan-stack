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

	ttnredis "go.thethings.network/lorawan-stack/pkg/redis"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

type DownlinkTaskQueue struct {
	*ttnredis.TaskQueue
}

const (
	downlinkKey = "downlink"
)

func NewDownlinkTaskQueue(cl *ttnredis.Client, maxLen int64, group, id string) *DownlinkTaskQueue {
	return &DownlinkTaskQueue{TaskQueue: &ttnredis.TaskQueue{
		Redis:  cl,
		MaxLen: maxLen,
		Group:  group,
		ID:     id,
		Key:    cl.Key(downlinkKey),
	}}
}

func (q *DownlinkTaskQueue) Add(ctx context.Context, devID ttnpb.EndDeviceIdentifiers, startAt time.Time) error {
	return q.TaskQueue.Add(unique.ID(ctx, devID), startAt)
}

func (q *DownlinkTaskQueue) Pop(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error {
	return q.TaskQueue.Pop(ctx, func(uid string, startAt time.Time) error {
		ids, err := unique.ToDeviceID(uid)
		if err != nil {
			return err
		}
		ctx, err := unique.WithContext(ctx, uid)
		if err != nil {
			return err
		}
		return f(ctx, ids, startAt)
	})
}
