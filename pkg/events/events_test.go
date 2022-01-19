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
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/events/basic"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type wrappedEvent struct {
	events.Event
}

type testData struct{}

func (testData) GetCorrelationIDs() []string {
	return []string{"TestNew"}
}

func TestNew(t *testing.T) {
	a := assertions.New(t)
	ctx := events.ContextWithCorrelationID(test.Context(), t.Name())
	evt := events.New(ctx, "test.evt", "test event", events.WithAuthFromContext())
	a.So(evt.CorrelationIDs(), should.Resemble, []string{"TestNew"})
	a.So(evt.Visibility().GetRights(), should.Contain, ttnpb.Right_RIGHT_ALL)
	a.So(evt.AuthType(), should.Equal, "")
	a.So(evt.AuthTokenType(), should.Equal, "")
	a.So(evt.AuthTokenID(), should.Equal, "")
	a.So(evt.RemoteIP(), should.Equal, "")

	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(
		"authorization", "bearer MFRWG.token_id.token_key"))
	ctx = peer.NewContext(ctx, &peer.Peer{
		Addr: &net.TCPAddr{IP: net.IP{10, 10, 10, 10}, Port: 10000},
	})
	evt = events.New(ctx, "test.evt", "test event", events.WithAuthFromContext(), events.WithClientInfoFromContext())
	a.So(evt.AuthType(), should.Equal, "bearer")
	a.So(evt.AuthTokenType(), should.Equal, "AccessToken")
	a.So(evt.AuthTokenID(), should.Equal, "token_id")
	a.So(evt.RemoteIP(), should.Equal, "10.10.10.10")

	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("x-forwarded-for", "20.20.20.20"))
	evt = events.New(ctx, "test.evt", "test event", events.WithAuthFromContext(), events.WithClientInfoFromContext())
	a.So(evt.RemoteIP(), should.Equal, "20.20.20.20")

	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("x-forwarded-for", "30.30.30.30, 20.20.20.20"))
	evt = events.New(ctx, "test.evt", "test event", events.WithAuthFromContext(), events.WithClientInfoFromContext())
	a.So(evt.RemoteIP(), should.Equal, "30.30.30.30")

	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("user-agent", "agent/0.1"))
	evt = events.New(ctx, "test.evt", "test event", events.WithAuthFromContext(), events.WithClientInfoFromContext())
	a.So(evt.UserAgent(), should.Equal, "agent/0.1")

	ctx = metadata.NewIncomingContext(ctx, metadata.Pairs("grpcgateway-user-agent", "agent-from-grpcgateway-header/0.2"))
	evt = events.New(ctx, "test.evt", "test event", events.WithAuthFromContext(), events.WithClientInfoFromContext())
	a.So(evt.UserAgent(), should.Equal, "agent-from-grpcgateway-header/0.2")
}

func TestUnmarshalJSON(t *testing.T) {
	a := assertions.New(t)
	{
		evt := events.New(
			context.Background(), "name", "description",
			events.WithIdentifiers(&ttnpb.ApplicationIdentifiers{ApplicationId: "application_id"}),
			events.WithData("data"),
			events.WithVisibility(ttnpb.Right_RIGHT_ALL),
		)
		json, err := json.Marshal(evt)
		a.So(err, should.BeNil)
		evt2, err := events.UnmarshalJSON(json)
		a.So(err, should.BeNil)
		a.So(evt2, should.Resemble, evt)
	}

	{
		var fieldmask []string
		evt := events.New(
			context.Background(), "name", "description",
			events.WithIdentifiers(&ttnpb.ApplicationIdentifiers{ApplicationId: "application_id"}),
			events.WithData(fieldmask),
			events.WithVisibility(ttnpb.Right_RIGHT_ALL),
		)
		json, err := json.Marshal(evt)
		a.So(err, should.BeNil)
		evt2, err := events.UnmarshalJSON(json)
		a.So(err, should.BeNil)
		a.So(evt2.Data(), should.BeNil)
	}
}

func Example() {
	// The context typically comes from the request or something.
	ctx := test.Context()

	// This is required for unit test to pass.
	defer test.SetDefaultEventsPubSub(basic.NewPubSub())()

	// The WaitGroup is only for synchronizing the unit test
	var wg sync.WaitGroup
	wg.Add(1)

	events.Subscribe(ctx, []string{"ns.mac.adr.send_req"}, nil, events.HandlerFunc(func(e events.Event) {
		fmt.Printf("Received event %s\n", e.Name())

		wg.Done() // only for synchronizing the unit test
	}))

	// You can send any arbitrary event; you don't have to pass any identifiers or data.
	events.Publish(events.New(test.Context(), "test.hello_world", "the events system says hello, world"))

	// Defining the event is not mandatory, but will be needed in order to translate the descriptions.
	// Event names are lowercase snake_case and can be dot-separated as component.subsystem.subsystem.event
	// Event descriptions are short descriptions of what the event means.
	// Visibility rights are optional. If no rights are supplied, then the _ALL right is assumed.
	adrSendEvent := events.Define("ns.mac.adr.send_req", "send ADR request", events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ))

	// These variables come from the request or you got them from the db or something.
	var (
		dev      ttnpb.EndDevice
		requests []ttnpb.MACCommand_LinkADRReq
	)

	// It's nice to be able to correlate events; we use a Correlation ID for that.
	// In most cases, there will already be a correlation ID in the context; this function will append a new one to the ones already in the context.
	ctx = events.ContextWithCorrelationID(ctx, events.NewCorrelationID())

	// Publishing an event to the events package will dispatch it on the "global" event pubsub.
	events.Publish(adrSendEvent.NewWithIdentifiersAndData(ctx, dev.Ids, requests))

	wg.Wait() // only for synchronizing the unit test

	// Output:
	// Received event ns.mac.adr.send_req
}
