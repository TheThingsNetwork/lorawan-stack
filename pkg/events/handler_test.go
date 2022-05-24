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
	"runtime"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func ExampleHandlerFunc() {
	// The context typically comes from the request or something.
	ctx := test.Context()

	handler := events.HandlerFunc(func(e events.Event) {
		fmt.Printf("Received event %v\n", e)
	})

	subCtx, unsubscribe := context.WithCancel(ctx)
	if err := events.Subscribe(subCtx, []string{"example"}, nil, handler); err != nil {
		panic(err)
	}

	// From this moment on, "example" events will be delivered to the handler func.

	// We want to unsubscribe when this function returns.
	defer unsubscribe()

	// Note that in-transit events may still be delivered after unsubscribe returns.
}

func ExampleChannel() {
	// The context typically comes from the request or something.
	ctx := test.Context()

	eventChan := make(events.Channel, 2)

	subCtx, unsubscribe := context.WithCancel(ctx)
	if err := events.Subscribe(subCtx, []string{"example"}, nil, eventChan); err != nil {
		panic(err)
	}

	// From this moment on, "example" events will be delivered to the channel.
	// As soon as the channel is full, events will be dropped, so it's probably a
	// good idea to start handling the channel before subscribing.

	go func() {
		for e := range eventChan {
			fmt.Printf("Received event %v\n", e)
		}
	}()

	// We want to unsubscribe when this function returns.
	defer unsubscribe()

	// Note that in-transit events may still be delivered after Unsubscribe returns.
	// This means that you can't immediately close the channel after unsubscribing.
}

func TestChannelReceive(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	eventChan := make(events.Channel, 2)
	eventChan.Notify(events.New(test.Context(), "evt", "event"))
	eventChan.Notify(events.New(test.Context(), "evt", "event"))
	eventChan.Notify(events.New(test.Context(), "overflow", "this overflows the channel"))

	runtime.Gosched()

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

	if err := events.Subscribe(ctx, []string{"example"}, nil, handler); err != nil {
		panic(err)
	}

	// We automatically unsubscribe when he context gets canceled.

	// From this moment on, "example" events will be delivered to the channel.
	// As soon as the channel is full, events will be dropped, so it's probably a
	// good idea to start handling the channel before subscribing.

	go func() {
		for {
			select {
			case <-ctx.Done():
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
