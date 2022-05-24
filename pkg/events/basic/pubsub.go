// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

// Package basic provides a basic events PubSub implementation.
package basic

import (
	"context"
	"runtime/trace"
	"sync"

	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// PubSub is a basic event PubSub implementation.
type PubSub struct {
	mu            sync.RWMutex
	subscriptions []events.Subscription
}

// NewPubSub returns a new basic event PubSub.
func NewPubSub() *PubSub {
	return &PubSub{}
}

// Subscribe starts a subscription for events that match name and identifiers.
// The subscription lasts until the context is canceled.
// The handler will be notified of matching events and must not block.
func (e *PubSub) Subscribe(
	ctx context.Context, names []string, identifiers []*ttnpb.EntityIdentifiers, hdl events.Handler,
) error {
	s, err := NewSubscription(ctx, names, identifiers, hdl)
	if err != nil {
		return err
	}

	e.AddSubscription(s)
	go func() {
		<-ctx.Done()
		e.RemoveSubscription(s)
	}()

	return nil
}

// AddSubscription adds an event subscription to the PubSub. The exact same
// subscription must be used in RemoveSubscription.
func (e *PubSub) AddSubscription(s events.Subscription) {
	e.mu.Lock()
	defer e.mu.Unlock()
	subscriptions := make([]events.Subscription, 0, len(e.subscriptions)+1)
	subscriptions = append(subscriptions, e.subscriptions...)
	e.subscriptions = append(subscriptions, s)
}

// RemoveSubscription removes an event subscription to the PubSub. This method
// expects the exact same subscription that was previously used in AddSubscription.
func (e *PubSub) RemoveSubscription(s events.Subscription) {
	e.mu.Lock()
	defer e.mu.Unlock()
	subscriptions := make([]events.Subscription, 0, len(e.subscriptions))
	for _, sub := range e.subscriptions {
		if sub != s {
			subscriptions = append(subscriptions, sub)
		}
	}
	e.subscriptions = subscriptions
}

// Publish publishes an event, which will notify all matching subscriptions.
func (e *PubSub) Publish(evs ...events.Event) {
	e.mu.RLock()
	subscriptions := e.subscriptions
	e.mu.RUnlock()
	for _, evt := range evs {
		trace.WithRegion(evt.Context(), "publish event", func() {
			for _, sub := range subscriptions {
				if sub.Match(evt) {
					sub.Notify(evt)
				}
			}
		})
	}
}
