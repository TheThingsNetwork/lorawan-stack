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
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/events/basic"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/task"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// NewPubSub creates a new PubSub that publishes and subscribes to Redis.
func NewPubSub(ctx context.Context, taskStarter task.Starter, conf config.RedisEvents) events.PubSub {
	ttnRedisClient := ttnredis.New(&conf.Config)
	ctx = log.NewContextWithFields(ctx, log.Fields(
		"namespace", "events/redis",
	))
	ctx, cancel := context.WithCancel(ctx)
	ps := &PubSub{
		PubSub: basic.NewPubSub(),
		ctx:    ctx,
		cancel: cancel,
		client: ttnRedisClient,

		subscriptions: make(map[string]int),
	}
	ps.sub = ps.client.Subscribe(ctx)

	workers := conf.Workers
	if workers == 0 {
		workers = 1
	}

	for i := 0; i < workers; i++ {
		taskStarter.StartTask(&task.Config{
			Context: ps.ctx,
			ID:      fmt.Sprintf("events_redis_subscribe_%02d", i),
			Func:    ps.subscribeTask,
			Restart: task.RestartOnFailure,
			Backoff: task.DefaultBackoffConfig,
		})
	}

	if !conf.Store.Enable {
		return ps
	}

	pss := &PubSubStore{
		PubSub:                    ps,
		historyTTL:                conf.Store.TTL,
		entityHistoryCount:        conf.Store.EntityCount,
		entityHistoryTTL:          conf.Store.EntityTTL,
		correlationIDHistoryCount: conf.Store.CorrelationIDCount,
	}
	if pss.historyTTL == 0 {
		pss.historyTTL = 10 * time.Minute
	}
	if pss.entityHistoryCount == 0 {
		pss.entityHistoryCount = 100
	}
	if pss.entityHistoryTTL == 0 {
		pss.entityHistoryTTL = time.Hour
	}
	if pss.correlationIDHistoryCount == 0 {
		pss.correlationIDHistoryCount = 100
	}

	return pss
}

// PubSub with Redis backend.
type PubSub struct {
	*basic.PubSub
	ctx           context.Context
	cancel        context.CancelFunc
	client        *ttnredis.Client
	mu            sync.RWMutex
	sub           *redis.PubSub
	subscriptions map[string]int
}

func (ps *PubSub) eventChannel(ctx context.Context, name string, ids *ttnpb.EntityIdentifiers) string {
	if name == "" {
		name = "*"
	}
	if ids == nil {
		return ps.client.Key("events", "*", "*", name)
	}
	return ps.client.Key("events", ids.EntityType(), unique.ID(ctx, ids), name)
}

var errChannelClosed = errors.DefineAborted("channel_closed", "channel closed")

const (
	protoEncodingPrefix = "v3-event-proto:"
	metaEncodingPrefix  = "v3-event-meta:"
)

func (ps *PubSub) subscribeTask(ctx context.Context) error {
	logger := log.FromContext(ctx)
	ch := ps.sub.Channel()
	store := &PubSubStore{
		PubSub: ps,
		// NOTE: only for reading; no additional settings needed.
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-ch:
			if !ok {
				return errChannelClosed.New()
			}
			var evtPB *ttnpb.Event
			switch {
			case strings.HasPrefix(msg.Payload, protoEncodingPrefix):
				evtPB = &ttnpb.Event{}
				err := decodeEventData(msg.Payload, evtPB)
				if err != nil {
					logger.WithError(err).Warn("Failed to decode event payload")
					continue
				}
			case strings.HasPrefix(msg.Payload, metaEncodingPrefix):
				m := strings.Split(strings.TrimPrefix(msg.Payload, metaEncodingPrefix), " ")
				var err error
				evtPB, err = store.LoadEvent(ctx, m[0])
				if err != nil {
					logger.WithError(err).Warn("Failed to load event payload")
					continue
				}
			default:
				logger.Warn("Skip decoding event with unexpected encoding")
				continue
			}
			evt, err := events.FromProto(evtPB)
			if err != nil {
				logger.WithError(err).Warn("Failed to convert event from protobuf")
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

func (ps *PubSub) eventChannelPatterns(ctx context.Context, names []string, ids []*ttnpb.EntityIdentifiers) []string {
	if len(names) == 0 {
		names = []string{"*"}
	}
	if len(ids) == 0 {
		ids = []*ttnpb.EntityIdentifiers{nil}
	}

	var patterns []string
	for _, name := range names {
		for _, id := range ids {
			patterns = append(patterns, ps.eventChannel(ctx, name, id))
			if appID := id.GetApplicationIds(); appID != nil {
				pattern := ps.eventChannel(ctx, name, (&ttnpb.EndDeviceIdentifiers{
					ApplicationIds: appID,
					DeviceId:       "*",
				}).GetEntityIdentifiers())
				patterns = append(patterns, pattern)
			}
		}
	}

	return patterns
}

// Subscribe implements the events.Subscriber interface.
func (ps *PubSub) Subscribe(ctx context.Context, names []string, ids []*ttnpb.EntityIdentifiers, hdl events.Handler) error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	basicSub, err := basic.NewSubscription(ctx, names, ids, hdl)
	if err != nil {
		return err
	}

	sub := &subscription{
		basicSub: basicSub,
		patterns: ps.eventChannelPatterns(ctx, names, ids),
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

	b, err := encodeEventData(evt)
	if err != nil {
		logger.WithError(err).Warn("Failed to encode event")
		return
	}

	_, err = ps.client.Pipelined(ps.ctx, func(tx redis.Pipeliner) error {
		ids := evt.Identifiers()
		if len(ids) == 0 {
			tx.Publish(ps.ctx, ps.eventChannel(evt.Context(), evt.Name(), nil), b)
		}
		for _, id := range ids {
			tx.Publish(ps.ctx, ps.eventChannel(evt.Context(), evt.Name(), id), b)
		}
		return nil
	})
	if err != nil {
		logger.WithError(err).Warn("Failed to publish event")
	}
}
