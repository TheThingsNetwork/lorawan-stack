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

// Package eventstest provides a blackbox test for events PubSub implementations.
package eventstest

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"golang.org/x/sync/errgroup"
)

// TestBackend runs the common test suite for all events backends.
func TestBackend(ctx context.Context, t *testing.T, a *assertions.Assertion, backend events.PubSub) {
	now := time.Now()
	correlationID := fmt.Sprintf("%s@%s", t.Name(), now)

	ctx = events.ContextWithCorrelationID(ctx, correlationID)

	timeout := test.Delay
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline) / 10
	}

	eui := types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}
	devAddr := types.DevAddr{1, 2, 3, 4}
	appID := ttnpb.ApplicationIdentifiers{
		ApplicationId: "test-app",
	}
	devID := ttnpb.EndDeviceIdentifiers{
		ApplicationIds: &appID,
		DeviceId:       "test-dev",
		DevEui:         &eui,
		JoinEui:        &eui,
		DevAddr:        &devAddr,
	}
	gtwID := ttnpb.GatewayIdentifiers{
		GatewayId: "test-gtw",
		Eui:       &eui,
	}

	ch0 := make(events.Channel, 10)
	ch1 := make(events.Channel, 10)
	ch2 := make(events.Channel, 10)

	checkEvent := func(e events.Event) {
		a.So(e.Time().IsZero(), should.BeFalse)
		a.So(e.Context(), should.NotBeNil)
	}

	subCtx, unsubscribe := context.WithCancel(ctx)
	defer unsubscribe()

	backend.Publish(events.New(ctx, "test.some.evt1", "test event 1", events.WithIdentifiers(&appID)))

	runtime.Gosched()
	time.Sleep(timeout)

	if store, ok := backend.(events.Store); ok {
		chx := make(events.Channel, 10)
		histSubCtx, cancel := context.WithCancel(subCtx)
		var g errgroup.Group
		g.Go(func() error {
			after := now.Add(-1 * time.Second)
			return store.SubscribeWithHistory(
				histSubCtx,
				[]string{"test.some.evt1"},
				[]*ttnpb.EntityIdentifiers{appID.GetEntityIdentifiers()},
				&after, 1, chx,
			)
		})
		defer func() {
			cancel()
			if err := g.Wait(); err != nil && !errors.IsCanceled(err) {
				t.Error(err)
			}
			a.So(chx, should.HaveLength, 2)
		}()
	}

	a.So(backend.Subscribe(
		subCtx,
		[]string{"test.some.evt0", "test.some.evt1"},
		nil,
		ch0,
	), should.BeNil)

	a.So(backend.Subscribe(
		subCtx,
		[]string{"test.some.evt0", "test.some.evt1"},
		[]*ttnpb.EntityIdentifiers{appID.GetEntityIdentifiers()},
		ch1,
	), should.BeNil)

	a.So(backend.Subscribe(
		subCtx,
		[]string{"test.other.evt2"},
		[]*ttnpb.EntityIdentifiers{gtwID.GetEntityIdentifiers()},
		ch2,
	), should.BeNil)

	runtime.Gosched()
	time.Sleep(timeout)

	backend.Publish(events.New(ctx, "test.some.evt0", "test event 0"))
	checkEvt0 := func(e events.Event) {
		checkEvent(e)
		a.So(e.Name(), should.Equal, "test.some.evt0")
		a.So(e.Identifiers(), should.BeNil)
	}

	backend.Publish(events.New(ctx, "test.some.evt1", "test event 1", events.WithIdentifiers(&appID)))
	checkEvt1 := func(e events.Event) {
		checkEvent(e)
		a.So(e.Name(), should.Equal, "test.some.evt1")
		if a.So(e.Identifiers(), should.NotBeNil) && a.So(e.Identifiers(), should.HaveLength, 1) {
			a.So(e.Identifiers()[0].GetApplicationIds(), should.Resemble, &appID)
		}
	}

	backend.Publish(events.New(ctx, "test.other.evt2", "test event 2", events.WithIdentifiers(&devID, &gtwID)))
	checkEvt2 := func(e events.Event) {
		checkEvent(e)
		a.So(e.Name(), should.Equal, "test.other.evt2")
		if a.So(e.Identifiers(), should.NotBeNil) && a.So(e.Identifiers(), should.HaveLength, 2) {
			a.So(e.Identifiers()[0].GetDeviceIds(), should.Resemble, &devID)
			a.So(e.Identifiers()[1].GetGatewayIds(), should.Resemble, &gtwID)
		}
	}

	runtime.Gosched()
	time.Sleep(timeout)

	if a.So(ch0, should.HaveLength, 2) {
		evt := <-ch0
		if evt.Name() == "test.some.evt0" { // Events may arrive out-of-order.
			checkEvt0(evt)
			checkEvt1(<-ch0)
		} else {
			checkEvt1(evt)
			checkEvt0(<-ch0)
		}
	}

	if a.So(ch1, should.HaveLength, 1) {
		checkEvt1(<-ch1)
	}

	if a.So(ch2, should.HaveLength, 1) {
		checkEvt2(<-ch2)
	}

	if store, ok := backend.(events.Store); ok {
		after := now.Add(-1 * time.Second)

		evts, err := store.FetchHistory(ctx, []string{
			"test.some.evt1",
		}, []*ttnpb.EntityIdentifiers{
			appID.GetEntityIdentifiers(),
			devID.GetEntityIdentifiers(),
			gtwID.GetEntityIdentifiers(),
		}, &after, 0)
		a.So(err, should.BeNil)
		a.So(evts, should.HaveLength, 2)

		evts, err = store.FetchHistory(ctx, nil, []*ttnpb.EntityIdentifiers{
			appID.GetEntityIdentifiers(),
		}, &after, 1)
		a.So(err, should.BeNil)
		a.So(evts, should.HaveLength, 1)

		evts, err = store.FindRelated(ctx, correlationID)
		a.So(err, should.BeNil)
		a.So(evts, should.HaveLength, 4)
	}
}
