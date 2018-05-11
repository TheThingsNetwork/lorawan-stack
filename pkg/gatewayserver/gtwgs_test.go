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

package gatewayserver_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc/metadata"
)

const timeout = 5 * time.Second

type mockLink struct {
	*test.MockServerStream

	SendFunc func(*ttnpb.GatewayDown) error
	RecvFunc func() (*ttnpb.GatewayUp, error)
}

func (m *mockLink) Send(d *ttnpb.GatewayDown) error {
	return m.SendFunc(d)
}

func (m *mockLink) Recv() (*ttnpb.GatewayUp, error) {
	return m.RecvFunc()
}

func newMockLink() *mockLink {
	return &mockLink{
		MockServerStream: &test.MockServerStream{
			MockStream: &test.MockStream{
				ContextFunc: func() context.Context { return context.Background() },
				SendMsgFunc: func(interface{}) error { return nil },
				RecvMsgFunc: func(interface{}) error { return nil },
			},
			SetHeaderFunc:  func(metadata.MD) error { return nil },
			SendHeaderFunc: func(metadata.MD) error { return nil },
			SetTrailerFunc: func(metadata.MD) { return },
		},
	}
}

func TestLink(t *testing.T) {
	a := assertions.New(t)

	logger := test.GetLogger(t)
	ctx := log.NewContext(context.Background(), logger)
	ctx, cancel := context.WithCancel(ctx)

	store, err := test.NewFrequencyPlansStore()
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	defer store.Destroy()

	registeredGatewayID := "registered-gateway"
	_, isAddr := StartMockIsGatewayServer(ctx, map[ttnpb.GatewayIdentifiers]ttnpb.Gateway{
		{GatewayID: registeredGatewayID}: {FrequencyPlanID: "EU_863_870"},
	})
	ns, nsAddr := StartMockGsNsServer(ctx)

	c := component.MustNew(logger, &component.Config{
		ServiceBase: config.ServiceBase{
			Cluster: config.Cluster{
				Name:           "test-gateway-server",
				IdentityServer: isAddr,
				NetworkServer:  nsAddr,
			},
			FrequencyPlans: config.FrequencyPlans{
				StoreDirectory: store.Directory(),
			},
		},
	})
	gs, err := gatewayserver.New(c, gatewayserver.Config{
		DisableAuth: true,
	})
	if !a.So(err, should.BeNil) {
		t.Fatal("Gateway server could not be initialized:", err)
	}

	err = gs.Start()
	if !a.So(err, should.BeNil) {
		t.Fatal("Gateway server could not start:", err)
	}

	up := make(chan *ttnpb.GatewayUp)
	down := make(chan *ttnpb.GatewayDown)

	link := newMockLink()
	link.SendFunc = func(downToSend *ttnpb.GatewayDown) error {
		select {
		case down <- downToSend:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	link.RecvFunc = func() (*ttnpb.GatewayUp, error) {
		select {
		case upReceived := <-up:
			return upReceived, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	link.ContextFunc = func() context.Context {
		return metadata.NewIncomingContext(ctx, metadata.MD{
			"id": []string{registeredGatewayID},
		})
	}

	gsStart := time.Now()
	for gs.GetPeer(ttnpb.PeerInfo_IDENTITY_SERVER, []string{}, nil) == nil {
		if time.Since(gsStart) > timeout {
			t.Fatal("Identity server was not initialized in time by the gateway server - timeout")
		}
		time.Sleep(2 * time.Millisecond)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := gs.Link(link)
		select {
		case <-ctx.Done():
		default:
			t.Fatal("Link unexpectedly quit:", err)
		}
		wg.Done()
	}()

	// StartServingGateway
	{
		select {
		case msg := <-ns.messageReceived:
			if msg != "StartServingGateway" {
				t.Fatal("Expected GS to call HandleUplink on the NS, instead received", msg)
			}
		case <-time.After(timeout):
			t.Fatal("The gateway server never called the network server's StartServingGateway method. This might be due to an unexpected error in the GatewayServer.Link() function.")
		}

		select {
		case up <- &ttnpb.GatewayUp{GatewayStatus: ttnpb.NewPopulatedGatewayStatus(test.Randy, false)}:
		case <-time.After(timeout):
			t.Fatal("The gateway server never called Link.Recv() to receive the status message. This might be due to an unexpected error in the GatewayServer.Link() function.")
		}
	}

	// Join request
	{
		join := ttnpb.NewPopulatedUplinkMessageJoinRequest(test.Randy)
		select {
		case up <- &ttnpb.GatewayUp{UplinkMessages: []*ttnpb.UplinkMessage{join}}:
		case <-time.After(timeout):
			t.Fatal("The gateway server never called Link.Recv() to receive the join request. This might be due to an unexpected error in the GatewayServer.Link() function.")
		}

		select {
		case msg := <-ns.messageReceived:
			if msg != "HandleUplink" {
				t.Fatal("Expected GS to call HandleUplink on the NS, instead received", msg)
			}
		case <-time.After(timeout):
			t.Fatal("The gateway server never called the network server's HandleUplink to handle the join request. This might be due to an unexpected error in the GatewayServer.Link() function.")
		}
	}

	// Uplink
	{
		genericAESKey := types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
		uplink := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy,
			// Generic SNwkSIntKey and FNwkSIntKey
			genericAESKey, genericAESKey, false)
		select {
		case up <- &ttnpb.GatewayUp{UplinkMessages: []*ttnpb.UplinkMessage{uplink}}:
		case <-time.After(timeout):
			t.Fatal("The gateway server never called Link.Recv() to receive the uplink. This might be due to an unexpected error in the GatewayServer.Link() function.")
		}

		select {
		case msg := <-ns.messageReceived:
			if msg != "HandleUplink" {
				t.Fatal("Expected GS to call HandleUplink on the NS, instead received", msg)
			}
		case <-time.After(timeout):
			t.Fatal("The gateway server never called the network server's HandleUplink to handle the uplink. This might be due to an unexpected error in the GatewayServer.Link() function.")
		}
	}

	cancel()
	wg.Wait()

	select {
	case msg := <-ns.messageReceived:
		if msg != "StopServingGateway" {
			t.Fatal("Expected GS to call StopServingGateway on the NS, instead received", msg)
		}
	case <-time.After(timeout):
		t.Fatal("The gateway server never called the network server's StopServingGateway method. This might be due to an unexpected error in the GatewayServer.Link() function.")
	}

	gs.Close()
}

func TestGetFrequencyPlan(t *testing.T) {
	a := assertions.New(t)

	store, err := test.NewFrequencyPlansStore()
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	defer store.Destroy()

	logger := test.GetLogger(t)
	c := component.MustNew(test.GetLogger(t), &component.Config{ServiceBase: config.ServiceBase{
		FrequencyPlans: config.FrequencyPlans{StoreDirectory: store.Directory()},
	}})
	gs, err := gatewayserver.New(c, gatewayserver.Config{})
	if !a.So(err, should.BeNil) {
		logger.Fatal("Gateway server could not start")
	}

	fp, err := gs.GetFrequencyPlan(context.Background(), &ttnpb.GetFrequencyPlanRequest{FrequencyPlanID: "EU_863_870"})
	a.So(err, should.BeNil)
	a.So(fp.BandID, should.Equal, "EU_863_870")
	a.So(len(fp.Channels), should.Equal, 8)

	defer gs.Close()
}
