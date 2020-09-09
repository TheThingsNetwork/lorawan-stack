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

// Package redis implements an events.PubSub implementation that uses Redis PubSub.
package redis

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/go-redis/redis/v7"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
)

// WrapPubSub wraps an existing PubSub and publishes all events received from Redis to that PubSub.
func WrapPubSub(ctx context.Context, wrapped events.PubSub, taskStarter component.TaskStarter, conf ttnredis.Config) *PubSub {
	ttnRedisClient := ttnredis.New(&conf)
	eventChannel := ttnRedisClient.Key("events")
	ctx = log.NewContextWithFields(ctx, log.Fields(
		"namespace", "events/redis",
		"channel", eventChannel,
	))
	ctx, cancel := context.WithCancel(ctx)
	return &PubSub{
		PubSub:       wrapped,
		taskStarter:  taskStarter,
		ctx:          ctx,
		cancel:       cancel,
		client:       ttnRedisClient.Client,
		eventChannel: eventChannel,
	}
}

// NewPubSub creates a new PubSub that publishes and subscribes to Redis.
func NewPubSub(ctx context.Context, taskStarter component.TaskStarter, conf ttnredis.Config) *PubSub {
	return WrapPubSub(ctx, events.NewPubSub(events.DefaultBufferSize), taskStarter, conf)
}

// PubSub with Redis backend.
type PubSub struct {
	events.PubSub

	taskStarter  component.TaskStarter
	ctx          context.Context
	cancel       context.CancelFunc
	eventChannel string
	client       *redis.Client
	subOnce      sync.Once
}

var errChannelClosed = errors.DefineAborted("channel_closed", "channel closed")

func (ps *PubSub) subscribeTask(ctx context.Context) error {
	logger := log.FromContext(ctx)
	sub := ps.client.Subscribe(ps.eventChannel)
	logger.Info("Subscribed")
	defer func() {
		if err := sub.Close(); err != nil {
			logger.WithError(err).Warn("Failed to close Redis subscription")
		} else {
			logger.Info("Unsubscribed")
		}
	}()
	ch := sub.Channel()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-ch:
			if !ok {
				return errChannelClosed.New()
			}
			evt, err := events.UnmarshalJSON([]byte(msg.Payload))
			if err != nil {
				logger.WithError(err).Warn("Failed to unmarshal event from JSON")
				continue
			}
			ps.PubSub.Publish(evt)
		}
	}
}

// Subscribe implements the events.Subscriber interface.
func (ps *PubSub) Subscribe(name string, hdl events.Handler) error {
	ps.subOnce.Do(func() {
		ps.taskStarter.StartTask(&component.TaskConfig{
			Context: ps.ctx,
			ID:      "events_redis_subscribe",
			Func:    ps.subscribeTask,
			Restart: component.TaskRestartOnFailure,
			Backoff: component.DefaultTaskBackoffConfig,
		})
	})
	return ps.PubSub.Subscribe(name, hdl)
}

// Close the Redis publisher.
func (ps *PubSub) Close(ctx context.Context) error {
	ps.cancel()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ps.ctx.Done():
		if err := ps.client.Close(); err != nil {
			return err
		}
		return ps.ctx.Err()
	}
}

// Publish an event to Redis.
func (ps *PubSub) Publish(evt events.Event) {
	logger := log.FromContext(ps.ctx)
	b, err := json.Marshal(evt)
	if err != nil {
		logger.WithError(err).Warn("Failed to marshal event to JSON")
		return
	}
	if err := ps.client.Publish(ps.eventChannel, b).Err(); err != nil {
		logger.WithError(err).Warn("Failed to publish event")
	}
}
