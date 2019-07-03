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

package cloud_test

import (
	"context"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/events/cloud"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	_ "gocloud.dev/pubsub/mempubsub"
)

func Example() {
	// Import the desired cloud pub-sub drivers (see godoc.org/gocloud.dev).
	// In this example we use "gocloud.dev/pubsub/mempubsub".

	// This sends all events received from a Go Cloud pub sub to the default pubsub.
	cloudPubSub, err := cloud.WrapPubSub(context.TODO(), events.DefaultPubSub(), "mem://events", "mem://events")
	if err != nil {
		// Handle error.
	}

	// Replace the default pubsub so that we will now publish to a Go Cloud pub sub.
	events.SetDefaultPubSub(cloudPubSub)
}

func TestCloudPubSub(t *testing.T) {
	a := assertions.New(t)

	events.IncludeCaller = true

	var eventCh = make(chan events.Event)
	handler := events.HandlerFunc(func(e events.Event) {
		t.Logf("Received event %v", e)
		a.So(e.Time().IsZero(), should.BeFalse)
		a.So(e.Context(), should.NotBeNil)
		eventCh <- e
	})

	pubsub, err := cloud.NewPubSub(test.Context(), "mem://events_test", "mem://events_test")
	a.So(err, should.BeNil)

	defer pubsub.Close()

	pubsub.Subscribe("cloud.**", handler)

	ctx := events.ContextWithCorrelationID(test.Context(), t.Name())

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

	cloud.SetContentType(pubsub, "application/json")

	pubsub.Publish(events.New(ctx, "cloud.test.evt0", &appID, nil))
	select {
	case e := <-eventCh:
		a.So(e.Name(), should.Equal, "cloud.test.evt0")
		if a.So(e.Identifiers(), should.NotBeNil) && a.So(e.Identifiers(), should.HaveLength, 1) {
			a.So(e.Identifiers()[0].GetApplicationIDs(), should.Resemble, &appID)
		}
	case <-time.After(time.Second):
		t.Error("Did not receive expected event")
		t.FailNow()
	}

	cloud.SetContentType(pubsub, "application/protobuf")

	pubsub.Publish(events.New(ctx, "cloud.test.evt1", ttnpb.CombineIdentifiers(&devID, &gtwID), nil))
	select {
	case e := <-eventCh:
		a.So(e.Name(), should.Equal, "cloud.test.evt1")
		if a.So(e.Identifiers(), should.NotBeNil) && a.So(e.Identifiers(), should.HaveLength, 2) {
			a.So(e.Identifiers()[0].GetDeviceIDs(), should.Resemble, &devID)
			a.So(e.Identifiers()[1].GetGatewayIDs(), should.Resemble, &gtwID)
		}
	case <-time.After(time.Second):
		t.Error("Did not receive expected event")
		t.FailNow()
	}
}
