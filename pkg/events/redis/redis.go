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
	"encoding/json"
	"strings"

	"github.com/go-redis/redis"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/events"
	ttnredis "go.thethings.network/lorawan-stack/pkg/redis"
)

// WrapPubSub wraps an existing PubSub and publishes all events received from Redis to that PubSub.
func WrapPubSub(wrapped events.PubSub, conf config.Redis) (ps *PubSub) {
	ps = &PubSub{
		PubSub:       wrapped,
		client:       ttnredis.New(&ttnredis.Config{Redis: conf}).Client,
		eventChannel: strings.Join(append(conf.Namespace, "events"), ":"),
		closeWait:    make(chan struct{}),
	}
	ps.sub = ps.client.Subscribe(ps.eventChannel)
	go func() {
		defer close(ps.closeWait)
		for {
			msg, err := ps.sub.ReceiveMessage()
			if err != nil {
				return
			}
			if evt, err := events.UnmarshalJSON([]byte(msg.Payload)); err == nil {
				ps.PubSub.Publish(evt)
			}
		}
	}()
	return
}

// NewPubSub creates a new PubSub that publishes and subscribes to Redis.
func NewPubSub(conf config.Redis) *PubSub {
	return WrapPubSub(events.NewPubSub(events.DefaultBufferSize), conf)
}

// PubSub with Redis backend.
type PubSub struct {
	events.PubSub

	eventChannel string
	client       *redis.Client
	sub          *redis.PubSub
	closeWait    chan struct{}
}

// Close the Redis publisher.
func (ps *PubSub) Close() error {
	unsubErr := ps.sub.Unsubscribe(ps.eventChannel)
	closeErr := ps.client.Close()
	<-ps.closeWait
	if unsubErr != nil {
		return unsubErr
	}
	return closeErr
}

// Publish an event to Redis.
func (ps *PubSub) Publish(evt events.Event) {
	json, err := json.Marshal(evt)
	if err == nil {
		ps.client.Publish(ps.eventChannel, string(json))
	}
}
