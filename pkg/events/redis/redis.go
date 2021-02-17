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
	"strings"
	"sync"

	"github.com/go-redis/redis/v8"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/events/basic"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// NewPubSub creates a new PubSub that publishes and subscribes to Redis.
func NewPubSub(ctx context.Context, taskStarter component.TaskStarter, conf ttnredis.Config) *PubSub {
	ttnRedisClient := ttnredis.New(&conf)
	eventChannel := func(ctx context.Context, name string, ids *ttnpb.EntityIdentifiers) string {
		if name == "" {
			name = "*"
		}
		if ids == nil {
			return ttnRedisClient.Key("events", "*", "*", name)
		}
		return ttnRedisClient.Key("events", ids.EntityType(), unique.ID(ctx, ids), name)
	}
	ctx = log.NewContextWithFields(ctx, log.Fields(
		"namespace", "events/redis",
	))
	ctx, cancel := context.WithCancel(ctx)
	ps := &PubSub{
		PubSub:        basic.NewPubSub(),
		ctx:           ctx,
		cancel:        cancel,
		client:        ttnRedisClient.Client,
		eventChannel:  eventChannel,
		subscriptions: make(map[string]int),
	}
	ps.sub = ps.client.Subscribe(ctx)
	taskStarter.StartTask(&component.TaskConfig{
		Context: ps.ctx,
		ID:      "events_redis_subscribe",
		Func:    ps.subscribeTask,
		Restart: component.TaskRestartOnFailure,
		Backoff: component.DefaultTaskBackoffConfig,
	})
	return ps
}

// PubSub with Redis backend.
type PubSub struct {
	*basic.PubSub

	eventChannel func(ctx context.Context, name string, ids *ttnpb.EntityIdentifiers) string

	ctx    context.Context
	cancel context.CancelFunc
	client *redis.Client

	subOnce       sync.Once
	mu            sync.RWMutex
	sub           *redis.PubSub
	subscriptions map[string]int
}

var errChannelClosed = errors.DefineAborted("channel_closed", "channel closed")

func (ps *PubSub) subscribeTask(ctx context.Context) error {
	logger := log.FromContext(ctx)

	ps.mu.Lock()
	ps.sub = ps.client.Subscribe(ctx)
	patterns := make([]string, 0, len(ps.subscriptions))
	for pattern := range ps.subscriptions {
		patterns = append(patterns, pattern)
	}
	logger.WithField("patterns", patterns).Debug("Subscribe to Redis channels")
	ps.sub.PSubscribe(ctx, patterns...)
	ps.mu.Unlock()

	defer func() {
		if err := ps.sub.Close(); err != nil {
			logger.WithError(err).Warn("Failed to close Redis subscription")
		}
	}()

	ch := ps.sub.Channel()
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
			ps.PubSub.Publish(&patternEvent{Event: evt, pattern: msg.Pattern})
		}
	}
}

type patternEvent struct {
	events.Event
	pattern string
}

func (ps *PubSub) eventChannelPatterns(ctx context.Context, name string, ids []*ttnpb.EntityIdentifiers) []string {
	if name == "" {
		name = "*"
	} else {
		name = strings.Replace(name, "**", "*", -1)
	}

	if len(ids) == 0 {
		ids = []*ttnpb.EntityIdentifiers{nil}
	}

	var patterns []string
	for _, id := range ids {
		patterns = append(patterns, ps.eventChannel(ctx, name, id))
		if appID := id.GetApplicationIDs(); appID != nil {
			pattern := ps.eventChannel(ctx, name, ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: *appID,
				DeviceID:               "*",
			}.EntityIdentifiers())
			patterns = append(patterns, pattern)
		}
	}

	return patterns
}

// Subscribe implements the events.Subscriber interface.
func (ps *PubSub) Subscribe(ctx context.Context, name string, ids []*ttnpb.EntityIdentifiers, hdl events.Handler) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	sub := &subscription{
		name:     name,
		ids:      ids,
		patterns: ps.eventChannelPatterns(ctx, name, ids),
		hdl:      hdl,
	}

	ps.PubSub.AddSubscription(sub)

	for _, pattern := range sub.patterns {
		if ps.subscriptions[pattern] == 0 {
			err := ps.sub.PSubscribe(ps.ctx, pattern)
			if err != nil {
				return err
			}
		}
		ps.subscriptions[pattern]++
	}

	go func() {
		<-ctx.Done()

		ps.PubSub.RemoveSubscription(sub)

		ps.mu.Lock()
		defer ps.mu.Unlock()

		for _, pattern := range sub.patterns {
			ps.subscriptions[pattern]--
			if ps.subscriptions[pattern] == 0 {
				log.FromContext(ctx).WithField("pattern", pattern).Debug("Unsubscribe from Redis channels")
				err := ps.sub.PUnsubscribe(ps.ctx, pattern)
				if err != nil {
					log.FromContext(ps.ctx).WithField("pattern", pattern).WithError(err).Warn("Could not unsubscribe")
				}
				delete(ps.subscriptions, pattern)
			}
		}
	}()

	return nil
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
	if ids := evt.Identifiers(); len(ids) > 0 {
		for _, id := range ids {
			if err := ps.client.Publish(ps.ctx, ps.eventChannel(evt.Context(), evt.Name(), id), b).Err(); err != nil {
				logger.WithError(err).Warn("Failed to publish event")
			}
		}
	} else {
		if err := ps.client.Publish(ps.ctx, ps.eventChannel(evt.Context(), evt.Name(), nil), b).Err(); err != nil {
			logger.WithError(err).Warn("Failed to publish event")
		}
	}
}
