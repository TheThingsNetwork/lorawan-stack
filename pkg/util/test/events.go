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
	"bytes"
	"context"
	"sync"

	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/proto"
)

type MockEventPubSub struct {
	PublishFunc   func(...events.Event)
	SubscribeFunc func(context.Context, []string, []*ttnpb.EntityIdentifiers, events.Handler) error
}

// Publish calls PublishFunc if set and panics otherwise.
func (m MockEventPubSub) Publish(evs ...events.Event) {
	if m.PublishFunc == nil {
		panic("Publish called, but not set")
	}
	m.PublishFunc(evs...)
}

// Subscribe calls SubscribeFunc if set and panics otherwise.
func (m MockEventPubSub) Subscribe(ctx context.Context, names []string, ids []*ttnpb.EntityIdentifiers, hdl events.Handler) error {
	if m.SubscribeFunc == nil {
		panic("Subscribe called, but not set")
	}
	return m.SubscribeFunc(ctx, names, ids, hdl)
}

type EventPubSubPublishRequest struct {
	Event    events.Event
	Response chan<- struct{}
}

func MakeEventPubSubPublishChFunc(reqCh chan<- EventPubSubPublishRequest) func(...events.Event) {
	return func(evs ...events.Event) {
		for _, ev := range evs {
			respCh := make(chan struct{})
			reqCh <- EventPubSubPublishRequest{
				Event:    ev,
				Response: respCh,
			}
			<-respCh
		}
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

func AssertEventPubSubPublishRequests(ctx context.Context, reqCh <-chan EventPubSubPublishRequest, n int, assert func(evs ...events.Event) bool) bool {
	t := MustTFromContext(ctx)
	t.Helper()

	var evs []events.Event
	for i := 0; i < n; i++ {
		if !AssertEventPubSubPublishRequest(ctx, reqCh, func(ev events.Event) bool {
			t.Logf("Received event number %d out of %d expected: %v", i+1, n, ev)
			evs = append(evs, ev)
			return true
		}) {
			t.Errorf("Failed to receive event number %d out of %d expected", i+1, n)
			return false
		}
	}
	return assert(evs...)
}

type EventEqualConfig struct {
	UniqueID       bool
	Time           bool
	Identifiers    bool
	Data           bool
	CorrelationIDs bool
	Origin         bool
	Context        bool
	Visibility     bool
	Authentication bool
	RemoteIP       bool
	UserAgent      bool
}

func MakeEventEqual(conf EventEqualConfig) func(a, b events.Event) bool {
	return func(a, b events.Event) bool {
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

		if !conf.UniqueID {
			ap.UniqueId = ""
			bp.UniqueId = ""
		}
		if !conf.Time {
			ap.Time = nil
			bp.Time = nil
		}
		if !conf.Identifiers {
			ap.Identifiers = nil
			bp.Identifiers = nil
		}
		if !conf.Data {
			ap.Data = nil
			bp.Data = nil
		}
		if !conf.CorrelationIDs {
			ap.CorrelationIds = nil
			bp.CorrelationIds = nil
		}
		if !conf.Origin {
			ap.Origin = ""
			bp.Origin = ""
		}
		if !conf.Context {
			ap.Context = nil
			bp.Context = nil
		}
		if !conf.Visibility {
			ap.Visibility = nil
			bp.Visibility = nil
		}
		if !conf.Authentication {
			ap.Authentication = nil
			bp.Authentication = nil
		}
		if !conf.RemoteIP {
			ap.RemoteIp = ""
			bp.RemoteIp = ""
		}
		if !conf.UserAgent {
			ap.UserAgent = ""
			bp.UserAgent = ""
		}

		apb, err := proto.Marshal(ap)
		if err != nil {
			return false
		}

		bpb, err := proto.Marshal(bp)
		if err != nil {
			return false
		}

		return bytes.Equal(apb, bpb)
	}
}

var EventEqual = MakeEventEqual(EventEqualConfig{
	Identifiers:    true,
	Data:           true,
	CorrelationIDs: true,
	Origin:         true,
	Context:        true,
	Visibility:     true,
	RemoteIP:       true,
	UserAgent:      true,
})

func EventBuilderEqual(a, b events.Builder) bool {
	ctx := Context()
	return EventEqual(a.New(ctx), b.New(ctx))
}

var (
	eventsPubSubMu               = &sync.RWMutex{}
	eventsPubSub   events.PubSub = &MockEventPubSub{
		PublishFunc:   func(...events.Event) {},
		SubscribeFunc: func(context.Context, []string, []*ttnpb.EntityIdentifiers, events.Handler) error { return nil },
	}
)

func init() {
	events.SetDefaultPubSub(&MockEventPubSub{
		PublishFunc: func(evs ...events.Event) {
			eventsPubSub.Publish(evs...)
		},
		SubscribeFunc: func(ctx context.Context, names []string, ids []*ttnpb.EntityIdentifiers, hdl events.Handler) error {
			return eventsPubSub.Subscribe(ctx, names, ids, hdl)
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
		PublishFunc: func(evts ...events.Event) { evs = append(evs, evts...) },
	})()
	f()
	return evs
}

// RedirectEvents redirects the published events to the
// provided channel until the returned function is called.
func RedirectEvents(ch chan events.Event) func() {
	return SetDefaultEventsPubSub(&MockEventPubSub{
		PublishFunc: func(evs ...events.Event) {
			for _, ev := range evs {
				ch <- ev
			}
		},
		SubscribeFunc: func(ctx context.Context, names []string, ids []*ttnpb.EntityIdentifiers, hdl events.Handler) error {
			return nil
		},
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
