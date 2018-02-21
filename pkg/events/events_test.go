// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/events"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

type wrappedEvent struct {
	events.Event
}

func TestEvents(t *testing.T) {
	a := assertions.New(t)

	events.IncludeCaller = true

	var totalEvents int
	newTotal := make(chan int)
	allEvents := events.HandlerFunc(func(e events.Event) {
		t.Logf("Received event %v", e)
		a.So(e.Time().IsZero(), should.BeFalse)
		a.So(e.Context(), should.NotBeNil)
		totalEvents++
		newTotal <- totalEvents
	})

	var eventCh = make(chan events.Event)
	handler := events.HandlerFunc(func(e events.Event) {
		eventCh <- e
	})

	pubsub := events.NewPubSub()

	pubsub.Subscribe("**", allEvents)

	ctx := events.ContextWithCorrelationID(context.Background(), t.Name())

	pubsub.Publish(events.New(ctx, "test.evt0", nil, nil))
	a.So(<-newTotal, should.Equal, 1)
	a.So(eventCh, should.HaveLength, 0)

	pubsub.Subscribe("test.*", handler)
	pubsub.Subscribe("test.*", handler) // second time should not matter

	evt := events.New(ctx, "test.evt1", "id", "hello")
	a.So(evt.CorrelationIDs(), should.Contain, t.Name())

	wrapped := wrappedEvent{Event: evt}
	pubsub.Publish(wrapped)
	a.So(<-newTotal, should.Equal, 2)

	received := <-eventCh
	a.So(received.Context(), should.Equal, evt.Context())
	a.So(received.Name(), should.Equal, evt.Name())
	a.So(received.Time(), should.Equal, evt.Time())
	a.So(received.Identifiers(), should.Equal, evt.Identifiers())
	a.So(received.Data(), should.Equal, evt.Data())
	a.So(received.CorrelationIDs(), should.Resemble, evt.CorrelationIDs())
	a.So(received.Origin(), should.Equal, evt.Origin())

	pubsub.Unsubscribe("test.*", handler)

	pubsub.Publish(events.New(ctx, "test.evt2", nil, nil))
	a.So(<-newTotal, should.Equal, 3)
	a.So(eventCh, should.HaveLength, 0)
}
