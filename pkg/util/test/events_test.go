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

package test_test

import (
	"context"
	"testing"
	"time"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	. "go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestSetDefaultEventPubSub(t *testing.T) {
	a := assertions.New(t)

	ctx := ContextWithTB(Context(), t)
	ctx, cancel := context.WithTimeout(ctx, (1<<5)*Delay)
	defer cancel()

	testEvent1 := events.New(ctx, "test-set-default-event-pub-sub-1", "test-event-1",
		events.WithIdentifiers(&ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"}),
	)
	testEvent2 := events.New(ctx, "test-set-default-event-pub-sub-2", "test-event-2",
		events.WithIdentifiers(&ttnpb.GatewayIdentifiers{GatewayId: "test-gtw"}),
	)

	events.Publish(testEvent1)
	events.Publish(testEvent2)

	publishCh1 := make(chan EventPubSubPublishRequest)
	undo1 := SetDefaultEventsPubSub(&MockEventPubSub{
		PublishFunc: MakeEventPubSubPublishChFunc(publishCh1),
	})
	go events.Publish(testEvent1)
	if !AssertEventPubSubPublishRequest(ctx, publishCh1, func(ev events.Event) bool {
		return a.So(ev, should.Equal, testEvent1)
	}) {
		t.FailNow()
	}
	close(publishCh1)

	publishCh2 := make(chan EventPubSubPublishRequest)
	time.AfterFunc(Delay, undo1)
	undo2 := SetDefaultEventsPubSub(&MockEventPubSub{
		PublishFunc: MakeEventPubSubPublishChFunc(publishCh2),
	})
	go events.Publish(testEvent2)
	if !AssertEventPubSubPublishRequest(ctx, publishCh2, func(ev events.Event) bool {
		return a.So(ev, should.Equal, testEvent2)
	}) {
		t.FailNow()
	}
	close(publishCh2)
	undo2()

	events.Publish(testEvent1)
	events.Publish(testEvent2)
}

func TestCollectEvents(t *testing.T) {
	ctx := Context()
	testEvent1 := events.New(ctx, "test-collect-events-1", "test-event-1",
		events.WithIdentifiers(&ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"}),
	)
	testEvent2 := events.New(ctx, "test-collect-events-2", "test-event-2",
		events.WithIdentifiers(&ttnpb.GatewayIdentifiers{GatewayId: "test-gtw"}),
	)
	assertions.New(t).So(CollectEvents(func() {
		events.Publish(testEvent1)
		events.Publish(testEvent1)
		events.Publish(testEvent2)
	}), should.Resemble, []events.Event{
		testEvent1,
		testEvent1,
		testEvent2,
	})
}

func TestRedirectEvents(t *testing.T) {
	ctx := Context()
	testEvent1 := events.New(ctx, "test-redirect-events-1", "test-event-1",
		events.WithIdentifiers(&ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"}),
	)
	testEvent2 := events.New(ctx, "test-redirect-events-2", "test-event-2",
		events.WithIdentifiers(&ttnpb.GatewayIdentifiers{GatewayId: "test-gtw"}),
	)
	ch := make(chan events.Event, 2)
	defer RedirectEvents(ch)()
	events.Publish(testEvent1)
	events.Publish(testEvent2)
	close(ch)
	var evs []events.Event
	for ev := range ch {
		evs = append(evs, ev)
	}
	assertions.New(t).So(
		evs, should.Resemble, []events.Event{
			testEvent1,
			testEvent2,
		},
	)
}

func TestWaitEvent(t *testing.T) {
	ctx := Context()
	ch := make(chan events.Event, 1)
	ch <- events.New(ctx, "test-wait-event-1", "test-event-1",
		events.WithIdentifiers(&ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"}),
	)
	defer close(ch)
	assertions.New(t).So(
		WaitEvent(ctx, ch, "test-wait-event-1"),
		should.BeTrue)
}
