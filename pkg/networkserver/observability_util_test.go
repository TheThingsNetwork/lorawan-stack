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

package networkserver

import (
	"go.thethings.network/lorawan-stack/pkg/events"
)

func init() {
	events.DefaultPubSub = events.NewPubSub(0)
}

var _ events.PubSub = &MockEventPubSub{}

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

// TODO(#1008) Move collectEvents to the test package
func collectEvents(f func()) (evs []events.Event) {
	oldPS := events.DefaultPubSub
	events.DefaultPubSub = &MockEventPubSub{
		PublishFunc: func(ev events.Event) { evs = append(evs, ev) },
	}
	defer func() { events.DefaultPubSub = oldPS }()

	f()
	return evs
}
