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

package eventstest

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

// TestBackend runs the common test suite for all events backends.
func TestBackend(ctx context.Context, t *testing.T, a *assertions.Assertion, backend events.PubSub) {
	ctx = events.ContextWithCorrelationID(ctx, t.Name())

	timeout := test.Delay
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline) / 10
	}

	eui := types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}
	devAddr := types.DevAddr{1, 2, 3, 4}
	appID := ttnpb.ApplicationIdentifiers{
		ApplicationID: "test-app",
	}
	devID := ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: appID,
		DeviceID:               "test-dev",
		DevEUI:                 &eui,
		JoinEUI:                &eui,
		DevAddr:                &devAddr,
	}
	gtwID := ttnpb.GatewayIdentifiers{
		GatewayID: "test-gtw",
		EUI:       &eui,
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

	a.So(backend.Subscribe(
		subCtx,
		"test.some.**",
		nil,
		ch0,
	), should.BeNil)

	a.So(backend.Subscribe(
		subCtx,
		"test.some.**",
		[]*ttnpb.EntityIdentifiers{appID.EntityIdentifiers()},
		ch1,
	), should.BeNil)

	a.So(backend.Subscribe(
		subCtx,
		"test.other.**",
		[]*ttnpb.EntityIdentifiers{gtwID.EntityIdentifiers()},
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

	backend.Publish(events.New(ctx, "test.some.evt1", "test event 1", events.WithIdentifiers(appID)))
	checkEvt1 := func(e events.Event) {
		checkEvent(e)
		a.So(e.Name(), should.Equal, "test.some.evt1")
		if a.So(e.Identifiers(), should.NotBeNil) && a.So(e.Identifiers(), should.HaveLength, 1) {
			a.So(e.Identifiers()[0].GetApplicationIDs(), should.Resemble, &appID)
		}
	}

	backend.Publish(events.New(ctx, "test.other.evt2", "test event 2", events.WithIdentifiers(&devID, &gtwID)))
	checkEvt2 := func(e events.Event) {
		checkEvent(e)
		a.So(e.Name(), should.Equal, "test.other.evt2")
		if a.So(e.Identifiers(), should.NotBeNil) && a.So(e.Identifiers(), should.HaveLength, 2) {
			a.So(e.Identifiers()[0].GetDeviceIDs(), should.Resemble, &devID)
			a.So(e.Identifiers()[1].GetGatewayIDs(), should.Resemble, &gtwID)
		}
	}

	runtime.Gosched()
	time.Sleep(timeout)

	if a.So(ch0, should.HaveLength, 2) {
		checkEvt0(<-ch0)
		checkEvt1(<-ch0)
	}

	if a.So(ch1, should.HaveLength, 1) {
		checkEvt1(<-ch1)
	} else {
		log.Errorf("ch1: %#v", <-ch1)
		log.Errorf("ch1: %#v", <-ch1)
	}

	if a.So(ch2, should.HaveLength, 1) {
		checkEvt2(<-ch2)
	}
}
