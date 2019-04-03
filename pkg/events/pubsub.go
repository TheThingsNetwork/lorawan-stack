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

package events

import (
	"runtime/trace"
	"sync"

	"github.com/gobwas/glob"
)

// PubSub interface combines the Publisher and Subscriber interfaces.
type PubSub interface {
	Publisher
	Subscriber
}

// Publisher interface lets you publish events.
type Publisher interface {
	// Publish emits an event on the default event pubsub.
	Publish(evt Event)
}

// Subscriber interface lets you subscribe to events.
type Subscriber interface {
	// Subscribe adds an event handler to the default event pubsub.
	// The name can be a glob in order to catch multiple event types.
	// The handler must be non-blocking.
	Subscribe(name string, hdl Handler) error
	// Unsubscribe removes an event handler from the default event pubsub.
	// Queued or in-transit events may still be delivered to the handler
	// even after Unsubscribe returns.
	Unsubscribe(name string, hdl Handler)
}

type handler struct {
	eventName string
	glob.Glob
	Handler
}

type tracedEvent struct {
	event Event
	trace *trace.Region
}

type pubsub struct {
	mu       sync.RWMutex
	handlers []handler
	events   chan tracedEvent
}

// DefaultBufferSize is the default number of events that can be buffered before Publish starts to block.
const DefaultBufferSize = 64

// NewPubSub returns a new event pubsub and starts a goroutine for handling.
func NewPubSub(bufSize uint) PubSub {
	e := &pubsub{
		events: make(chan tracedEvent, bufSize),
	}
	go e.Run()
	return e
}

func (e *pubsub) Run() {
	for evt := range e.events {
		e.mu.RLock()
		handlers := e.handlers
		e.mu.RUnlock()
		for _, l := range handlers {
			if l.Match(evt.event.Name()) {
				l.Notify(evt.event)
			}
		}
		evt.trace.End()
	}
}

func (e *pubsub) Subscribe(name string, hdl Handler) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	for _, l := range e.handlers {
		if l.eventName == name && l.Handler == hdl {
			return nil
		}
	}
	glob, err := glob.Compile(name, '.')
	if err != nil {
		return err
	}
	e.handlers = append(e.handlers, handler{
		eventName: name,
		Glob:      glob,
		Handler:   hdl,
	})
	subscriptions.WithLabelValues(name).Inc()
	return nil
}

func (e *pubsub) Unsubscribe(name string, hdl Handler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for i, l := range e.handlers {
		if l.eventName == name && l.Handler == hdl {
			e.handlers = append(e.handlers[:i], e.handlers[i+1:]...)
			subscriptions.WithLabelValues(name).Dec()
			return
		}
	}
}

func (e *pubsub) Publish(evt Event) {
	localEvent := local(evt)
	publishes.WithLabelValues(evt.Context(), evt.Name()).Inc()
	e.events <- tracedEvent{
		event: localEvent.withCaller(),
		trace: trace.StartRegion(evt.Context(), "publish event"),
	}
}
