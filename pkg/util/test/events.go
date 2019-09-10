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

package test

import (
	"context"
	"reflect"
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/pkg/events"
)

type MockEventPubSub struct {
	PublishFunc     func(events.Event)
	SubscribeFunc   func(string, events.Handler) error
	UnsubscribeFunc func(string, events.Handler)
}

// Publish calls PublishFunc if set and panics otherwise.
func (m MockEventPubSub) Publish(ev events.Event) {
	if m.PublishFunc == nil {
		panic("Publish called, but not set")
	}
	m.PublishFunc(ev)
}

// Subscribe calls SubscribeFunc if set and panics otherwise.
func (m MockEventPubSub) Subscribe(name string, hdl events.Handler) error {
	if m.SubscribeFunc == nil {
		panic("Subscribe called, but not set")
	}
	return m.SubscribeFunc(name, hdl)
}

// Unsubscribe calls UnsubscribeFunc if set and panics otherwise.
func (m MockEventPubSub) Unsubscribe(name string, hdl events.Handler) {
	if m.UnsubscribeFunc == nil {
		panic("Unsubscribe called, but not set")
	}
	m.UnsubscribeFunc(name, hdl)
}

type EventPubSubPublishRequest struct {
	Event    events.Event
	Response chan<- struct{}
}

func MakeEventPubSubPublishChFunc(reqCh chan<- EventPubSubPublishRequest) func(events.Event) {
	return func(ev events.Event) {
		respCh := make(chan struct{})
		reqCh <- EventPubSubPublishRequest{
			Event:    ev,
			Response: respCh,
		}
		<-respCh
	}
}

func AssertEventPubSubPublishRequest(ctx context.Context, reqCh <-chan EventPubSubPublishRequest, assert func(ev events.Event) bool) bool {
	t := MustTFromContext(ctx)
	t.Helper()
	select {
	case <-ctx.Done():
		t.Error("Timed out while waiting for events.Publish to be called")
		return false

	case req := <-reqCh:
		t.Log("events.Publish called")
		if !assert(req.Event) {
			return false
		}
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for events.Publish response to be processed")
			return false

		case req.Response <- struct{}{}:
			return true
		}
	}
}

func EventEqual(a, b events.Event) bool {
	if a == nil {
		return b == nil
	}

	ap, err := events.Proto(a)
	if err != nil {
		return false
	}
	bp, err := events.Proto(b)
	if err != nil {
		return false
	}
	ap.Time = time.Time{}
	bp.Time = time.Time{}
	return reflect.DeepEqual(ap, bp)
}

var (
	eventsPubSubMu               = &sync.RWMutex{}
	eventsPubSub   events.PubSub = &MockEventPubSub{
		PublishFunc:     func(events.Event) {},
		SubscribeFunc:   func(string, events.Handler) error { return nil },
		UnsubscribeFunc: func(string, events.Handler) {},
	}
)

func init() {
	events.SetDefaultPubSub(&MockEventPubSub{
		PublishFunc: func(ev events.Event) {
			eventsPubSub.Publish(ev)
		},
		SubscribeFunc: func(name string, hdl events.Handler) error {
			return eventsPubSub.Subscribe(name, hdl)
		},
		UnsubscribeFunc: func(name string, hdl events.Handler) {
			eventsPubSub.Unsubscribe(name, hdl)
		},
	})
}

// SetDefaultEventsPubSub calls events.SetDefaultPubSub and
// returns a function that can be used to undo the action.
// Following calls to SetDefaultEventsPubSub will block until undo function is called.
func SetDefaultEventsPubSub(ps events.PubSub) func() {
	eventsPubSubMu.Lock()
	oldPS := eventsPubSub
	eventsPubSub = ps
	return func() {
		eventsPubSub = oldPS
		eventsPubSubMu.Unlock()
	}
}

// CollectEvents collects events published by f.
func CollectEvents(f func()) []events.Event {
	var evs []events.Event
	defer SetDefaultEventsPubSub(&MockEventPubSub{
		PublishFunc: func(ev events.Event) { evs = append(evs, ev) },
	})()
	f()
	return evs
}

// RedirectEvents redirects the published events to the
// provided channel until the returned function is called.
func RedirectEvents(ch chan events.Event) func() {
	return SetDefaultEventsPubSub(&MockEventPubSub{
		PublishFunc:     func(ev events.Event) { ch <- ev },
		SubscribeFunc:   func(name string, hdl events.Handler) error { return nil },
		UnsubscribeFunc: func(name string, hdl events.Handler) {},
	})
}

// WaitEvent waits for a specific event to be sent to the channel.
func WaitEvent(ctx context.Context, ch chan events.Event, name string) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		case ev, ok := <-ch:
			if !ok {
				panic("channel is closed")
			}
			if ev.Name() == name {
				return true
			}
		}
	}
}
