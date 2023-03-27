// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package telemetry

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
)

const (
	telemetryKey = "telemetry"
)

var errCallbackNotRegistered = errors.DefineNotFound("callback_not_registered", "callback not registered")

// TaskQueue represents a queue of telemetry tasks.
type TaskQueue interface {
	// Add telemetry task identified by `id` at time t.
	// Implementations must ensure that Add returns fast.
	Add(ctx context.Context, id string, t time.Time, replace bool) error

	// RegisterCallback registers a callback that is called when a task is popped.
	// All callbacks should be registered before the dispatch of the tasks in the queue, otherwise they are proned to
	// fail within the first pop call.
	RegisterCallback(callbackID string, callback TaskCallback)

	// Dispatch the tasks in the queue.
	Dispatch(ctx context.Context, consumerID string) error

	// Pop pops the most recent task in the schedule, for which timestamp is in range [0, time.Now()], the value of
	// the queue goes through the registered callback and determines which one should be called.
	Pop(ctx context.Context, consumerID string) error
}

// TaskQueueCloser is a function that closes the task queue.
type TaskQueueCloser func(context.Context) error

// RedisTaskQueue is an implementation of telemetry.TaskQueue.
type RedisTaskQueue struct {
	queue     *ttnredis.TaskQueue
	callbacks sync.Map
}

// TaskCallback is a callback that is called when a telemetry task is popped.
type TaskCallback func(context.Context) (time.Time, error)

// CallbackWithInterval is a wrapper that takes a task.Func and a time.Duration and returns a TaskCallback.
func CallbackWithInterval(_ context.Context, interval time.Duration, callback task.Func) TaskCallback {
	return func(ctx context.Context) (time.Time, error) {
		t := time.Now().Add(interval)
		if err := callback(ctx); err != nil {
			return t, err
		}
		return t, nil
	}
}

// NewRedisTaskQueue returns new telemetry task queue.
func NewRedisTaskQueue(
	ctx context.Context, cl *ttnredis.Client, maxLen int64, group string, streamBlockLimit time.Duration,
) (TaskQueue, TaskQueueCloser, error) {
	tq := &RedisTaskQueue{
		queue: &ttnredis.TaskQueue{
			Redis:            cl,
			MaxLen:           maxLen,
			Group:            group,
			Key:              cl.Key(telemetryKey),
			StreamBlockLimit: streamBlockLimit,
		},
	}
	if err := tq.Init(ctx); err != nil {
		return nil, nil, err
	}

	return tq, tq.Close, nil
}

// Init initializes the TelemetryTaskQueue.
func (q *RedisTaskQueue) Init(ctx context.Context) error {
	return q.queue.Init(ctx)
}

// Close closes the TelemetryTaskQueue.
func (q *RedisTaskQueue) Close(ctx context.Context) error {
	return q.queue.Close(ctx)
}

// Add telemetry task's identifier at time startAt.
func (q *RedisTaskQueue) Add(ctx context.Context, id string, startAt time.Time, replace bool) error {
	return q.queue.Add(ctx, nil, id, startAt, replace)
}

// RegisterCallback registers a callback that is called when a task is popped.
// All callbacks should be registered before the dispatch of the tasks in the queue, otherwise they are proned to fail
// within the first pop call.
func (q *RedisTaskQueue) RegisterCallback(callbackID string, callback TaskCallback) {
	q.callbacks.Store(callbackID, callback)
}

// Dispatch the tasks in the queue.
func (q *RedisTaskQueue) Dispatch(ctx context.Context, consumerID string) error {
	return q.queue.Dispatch(ctx, consumerID, nil)
}

// Pop pops the most recent task in the schedule, for which timestamp is in range [0, time.Now()], the value of
// the queue goes through the registered callback and determines which one should be called.
func (q *RedisTaskQueue) Pop(ctx context.Context, consumerID string) error {
	return q.queue.Pop(
		ctx, consumerID, nil, func(p redis.Pipeliner, id string, t time.Time) error {
			v, ok := q.callbacks.Load(id)
			if !ok {
				return errCallbackNotRegistered.WithAttributes("id", id)
			}

			tt, err := v.(TaskCallback)(ctx)
			if err != nil {
				log.FromContext(ctx).WithError(err).WithField("id", id).Warn("Failed to execute task from queue")
				return q.Add(ctx, id, tt, false)
			}

			return q.Add(ctx, id, tt, false)
		},
	)
}
