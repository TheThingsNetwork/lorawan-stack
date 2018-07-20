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

	"github.com/kr/pretty"
	"github.com/smartystreets/assertions"
	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc/metadata"
)

const (
	peerConnectionTimeout = 5 * time.Second
	nsReceptionTimeout    = 2 * time.Second
)

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
				ContextFunc: func() context.Context { return test.Context() },
				SendMsgFunc: func(interface{}) error { return nil },
				RecvMsgFunc: func(interface{}) error { return nil },
			},
			SetHeaderFunc:  func(metadata.MD) error { return nil },
			SendHeaderFunc: func(metadata.MD) error { return nil },
			SetTrailerFunc: func(metadata.MD) { return },
		},
	}
}

func TestLinkGateway(t *testing.T) {
	a := assertions.New(t)

	logger := test.GetLogger(t)
	ctx := log.NewContext(test.Context(), logger)
	ctx = clusterauth.NewContext(ctx, nil)
	ctx, cancel := context.WithCancel(ctx)

	gtwID, err := unique.ToGatewayID(registeredGatewayUID)
	a.So(err, should.BeNil)

	is, isAddr := StartMockIsGatewayServer(ctx, []ttnpb.Gateway{
		{
			GatewayIdentifiers: gtwID,
			FrequencyPlanID:    "EU_863_870",
		},
	})
	is.rights = []ttnpb.Right{ttnpb.RIGHT_GATEWAY_INFO, ttnpb.RIGHT_GATEWAY_LINK} // result of ListGatewayRights
	ns, nsAddr := StartMockGsNsServer(ctx)

	c := component.MustNew(logger, &component.Config{
		ServiceBase: config.ServiceBase{
			Cluster: config.Cluster{
				Name:           "test-gateway-server",
				IdentityServer: isAddr,
				NetworkServer:  nsAddr,
			},
			GRPC: config.GRPC{
				AllowInsecureForCredentials: true,
			},
		},
	})
	c.FrequencyPlans.Fetcher = test.FrequencyPlansFetcher
	gs, err := gatewayserver.New(c, gatewayserver.Config{})
	if !a.So(err, should.BeNil) {
		t.Fatal("Gateway Server could not be initialized:", err)
	}

	err = gs.Start()
	if !a.So(err, should.BeNil) {
		t.Fatal("Gateway Server could not start:", err)
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
		md := metadata.Pairs("id", gtwID.GatewayID)
		if ctxMd, ok := metadata.FromIncomingContext(ctx); ok {
			md = metadata.Join(ctxMd, md)
		}
		ctx = metadata.NewIncomingContext(ctx, md)

		ctx = rights.NewContextWithFetcher(
			ctx,
			rights.FetcherFunc(func(ctx context.Context, ids ttnpb.Identifiers) ([]ttnpb.Right, error) {
				return []ttnpb.Right{ttnpb.RIGHT_GATEWAY_LINK}, nil
			}),
		)
		return ctx
	}

	gsStart := time.Now()
	for gs.GetPeer(ttnpb.PeerInfo_IDENTITY_SERVER, []string{}, nil) == nil {
		if time.Since(gsStart) > peerConnectionTimeout {
			t.Fatal("Identity Server was not initialized in time by the Gateway Server - timeout")
		}
		time.Sleep(200 * time.Millisecond)
	}
	for gs.GetPeer(ttnpb.PeerInfo_NETWORK_SERVER, []string{}, nil) == nil {
		if time.Since(gsStart) > peerConnectionTimeout {
			t.Fatal("Gateway Server could not reach Network Server in time - timeout")
		}
		time.Sleep(200 * time.Millisecond)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := gs.LinkGateway(link)
		select {
		case <-ctx.Done():
		default:
			t.Fatal("Link unexpectedly quit:", err)
		}
		wg.Done()
	}()

	// TODO: monitor cluster claim on IDs https://github.com/TheThingsIndustries/lorawan-stack/issues/941

	t.Run("Join request", func(t *testing.T) {
		join := ttnpb.NewPopulatedUplinkMessageJoinRequest(test.Randy)
		select {
		case up <- &ttnpb.GatewayUp{UplinkMessages: []*ttnpb.UplinkMessage{join}}:
		case <-time.After(nsReceptionTimeout):
			t.Fatal("The Gateway Server never called Link.Recv() to receive the join request. This might be due to an unexpected error in the GatewayServer.LinkGateway() function.")
		}

		select {
		case msg := <-ns.messageReceived:
			if msg != "HandleUplink" {
				t.Fatal("Expected Gateway Server to call HandleUplink on the Network Server, instead received", msg)
			}
		case <-time.After(nsReceptionTimeout):
			t.Fatal("The Gateway Server never called the Network Server's HandleUplink to handle the join-request. This might be due to an unexpected error in the GatewayServer.LinkGateway() function.")
		}
	})

	t.Run("Uplink", func(t *testing.T) {
		genericAESKey := types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
		uplink := ttnpb.NewPopulatedUplinkMessageUplink(test.Randy,
			// Generic SNwkSIntKey and FNwkSIntKey
			genericAESKey, genericAESKey, false)
		select {
		case up <- &ttnpb.GatewayUp{UplinkMessages: []*ttnpb.UplinkMessage{uplink}}:
		case <-time.After(nsReceptionTimeout):
			t.Fatal("The Gateway Server never called Link.Recv() to receive the uplink. This might be due to an unexpected error in the GatewayServer.LinkGateway() function.")
		}

		select {
		case msg := <-ns.messageReceived:
			if msg != "HandleUplink" {
				t.Fatal("Expected Gateway Server to call HandleUplink on the Network Server, instead received", msg)
			}
		case <-time.After(nsReceptionTimeout):
			t.Fatal("The Gateway Server never called the Network Server's HandleUplink to handle the uplink. This might be due to an unexpected error in the GatewayServer.LinkGateway() function.")
		}
	})

	t.Run("Downlink", func(t *testing.T) {
		a := assertions.New(t)

		downlink := ttnpb.NewPopulatedDownlinkMessage(test.Randy, false)
		downlink.TxMetadata.GatewayIdentifiers = ttnpb.GatewayIdentifiers{
			GatewayID: gtwID.GatewayID,
		}
		downlink.Settings.Frequency = 863000000
		downlink.Settings.SpreadingFactor = 7
		downlink.Settings.Bandwidth = 125000
		downlink.Settings.CodingRate = "4/5"
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			receivedDown := <-down
			a.So(pretty.Diff(downlink, receivedDown.DownlinkMessage), should.BeEmpty)
			wg.Done()
		}()
		_, err := gs.ScheduleDownlink(ctx, downlink)

		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		wg.Wait()
	})

	cancel()
	wg.Wait()

	// TODO: monitor cluster claim on IDs https://github.com/TheThingsIndustries/lorawan-stack/issues/941

	gs.Close()
}

func TestGetFrequencyPlan(t *testing.T) {
	a := assertions.New(t)

	logger := test.GetLogger(t)
	c := component.MustNew(test.GetLogger(t), &component.Config{})
	c.FrequencyPlans.Fetcher = test.FrequencyPlansFetcher
	gs, err := gatewayserver.New(c, gatewayserver.Config{})
	if !a.So(err, should.BeNil) {
		logger.Fatal("Gateway Server could not start")
	}

	fp, err := gs.GetFrequencyPlan(test.Context(), &ttnpb.GetFrequencyPlanRequest{FrequencyPlanID: "EU_863_870"})
	a.So(err, should.BeNil)
	a.So(fp.BandID, should.Equal, "EU_863_870")
	a.So(len(fp.Channels), should.Equal, 8)

	defer gs.Close()
}
