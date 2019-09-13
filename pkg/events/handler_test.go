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

package events_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func ExampleHandlerFunc() {
	handler := events.HandlerFunc(func(e events.Event) {
		fmt.Printf("Received event %v\n", e)
	})

	events.Subscribe("example", handler)

	// From this moment on, "example" events will be delivered to the handler func.

	events.Unsubscribe("example", handler)

	// Note that in-transit events may still be delivered after Unsubscribe returns.
}

func ExampleChannel() {
	eventChan := make(events.Channel, 2)

	events.Subscribe("example", eventChan)

	// From this moment on, "example" events will be delivered to the channel.
	// As soon as the channel is full, events will be dropped, so it's probably a
	// good idea to start handling the channel before subscribing.

	go func() {
		for e := range eventChan {
			fmt.Printf("Received event %v\n", e)
		}
	}()

	// Later:
	events.Unsubscribe("example", eventChan)

	// Note that in-transit events may still be delivered after Unsubscribe returns.
	// This means that you can't immediately close the channel after unsubscribing.
}

func TestChannelReceive(t *testing.T) {
	a := assertions.New(t)

	eventChan := make(events.Channel, 2)
	eventChan.Notify(events.New(test.Context(), "evt", nil, nil))
	eventChan.Notify(events.New(test.Context(), "evt", nil, nil))
	eventChan.Notify(events.New(test.Context(), "overflow", nil, nil))

	ctx, cancel := context.WithCancel(test.Context())

	a.So(eventChan.ReceiveTimeout(test.Delay), should.NotBeNil)
	a.So(eventChan.ReceiveContext(ctx), should.NotBeNil)

	cancel()

	a.So(eventChan.ReceiveTimeout(test.Delay), should.BeNil)
	a.So(eventChan.ReceiveContext(ctx), should.BeNil)
}

func ExampleContextHandler() {
	// Usually the context comes from somewhere else (e.g. a streaming RPC):
	ctx, cancel := context.WithCancel(test.Context())
	defer cancel()

	eventChan := make(events.Channel, 2)
	handler := events.ContextHandler(ctx, eventChan)

	events.Subscribe("example", handler)

	// From this moment on, "example" events will be delivered to the channel.
	// As soon as the channel is full, events will be dropped, so it's probably a
	// good idea to start handling the channel before subscribing.

	go func() {
		for {
			select {
			case <-ctx.Done():
				// Don't forget to unsubscribe:
				events.Unsubscribe("example", handler)

				// The ContextHandler will make sure that no events are delivered after
				// the context is canceled, so it is now safe to close the channel:
				close(eventChan)
				return
			case e := <-eventChan:
				fmt.Printf("Received event %v\n", e)
			}
		}
	}()
}
