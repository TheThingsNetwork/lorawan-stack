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
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/component"
	"github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmetadata"
	"github.com/TheThingsNetwork/ttn/pkg/rpcserver"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"google.golang.org/grpc"
)

func Example() {
	logger, err := log.NewLogger()
	if err != nil {
		panic(err)
	}

	c := component.MustNew(logger, &component.Config{ServiceBase: config.ServiceBase{}})

	gs := gatewayserver.New(c, &gatewayserver.Config{})
	gs.Run()
}

func TestGatewayServer(t *testing.T) {
	a := assertions.New(t)

	dir := createFPStore(a)
	defer removeFPStore(a, dir)

	c := component.MustNew(test.GetLogger(t), &component.Config{})
	gs := gatewayserver.New(c, &gatewayserver.Config{
		FileFrequencyPlansStore: dir,
	})

	roles := gs.Roles()
	a.So(len(roles), should.Equal, 1)
	a.So(roles[0], should.Equal, ttnpb.PeerInfo_GATEWAY_SERVER)

	defer gs.Close()
}

func TestLink(t *testing.T) {
	a := assertions.New(t)

	ctx := log.NewContext(context.Background(), test.GetLogger(t))

	dir := createFPStore(a)
	defer removeFPStore(a, dir)

	registeredGatewayID, registeredGatewayUnknownFPID := "registered-gateway", "registered-gw-unknown-fp"
	registeredGatewayFP, registeredGatewayUnknownFP := "EU_863_870", "UNKNOWN_FP"
	registeredGateways := map[ttnpb.GatewayIdentifiers]ttnpb.Gateway{
		{GatewayID: registeredGatewayID}:          {FrequencyPlanID: registeredGatewayFP},
		{GatewayID: registeredGatewayUnknownFPID}: {FrequencyPlanID: registeredGatewayUnknownFP},
	}
	_, isAddr := StartMockIsGatewayServer(ctx, registeredGateways)
	ns, nsAddr := StartMockGsNsServer(ctx)

	c := component.MustNew(test.GetLogger(t), &component.Config{
		ServiceBase: config.ServiceBase{
			Cluster: config.Cluster{
				Name:           "test-gateway-server",
				IdentityServer: isAddr,
				NetworkServer:  nsAddr,
			},
			GRPC: config.GRPC{Listen: ":8088"},
			HTTP: config.HTTP{Listen: ":8080", PProf: true},
		},
	})

	var client ttnpb.GtwGsClient
	srv := grpc.NewServer()

	gs := gatewayserver.New(c, &gatewayserver.Config{FileFrequencyPlansStore: dir})

	// Initializing server and client
	{
		gs.RegisterServices(srv)
		err := gs.Start()
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		defer gs.Close()

		md := rpcmetadata.MD{ID: registeredGatewayID}
		ctx = md.ToOutgoingContext(ctx)
		conn, err := rpcserver.StartLoopback(ctx, srv, grpc.WithPerRPCCredentials(md))
		a.So(err, should.BeNil)
		client = ttnpb.NewGtwGsClient(conn)
	}

	// Frequency plan
	{
		fp, err := client.GetFrequencyPlan(log.NewContext(ctx, log.FromContext(ctx).WithField("situation", "client_request")), &ttnpb.FrequencyPlanRequest{FrequencyPlanID: "EU_863_870"})
		a.So(err, should.BeNil)
		a.So(fp.BandID, should.Equal, registeredGatewayFP)
	}

	// Failing link
	{
		failedLinkMd := rpcmetadata.MD{ID: registeredGatewayUnknownFPID}
		failedLinkCtx := failedLinkMd.ToOutgoingContext(ctx)
		failedLinkConn, err := rpcserver.StartLoopback(failedLinkCtx, srv)
		a.So(err, should.BeNil)
		failedLinkClient := ttnpb.NewGtwGsClient(failedLinkConn)

		failedLink, err := failedLinkClient.Link(failedLinkCtx)
		a.So(err, should.BeNil)
		_, err = failedLink.Recv()
		a.So(err, should.NotBeNil)
	}

	linkCtx, linkCancel := context.WithCancel(ctx)
	defer linkCancel()
	ns.Add(1)
	link, err := client.Link(linkCtx, grpc.FailFast(true))
	a.So(err, should.BeNil)
	ns.Wait()

	// Sending empty uplink
	{
		err = link.Send(&ttnpb.GatewayUp{
			UplinkMessages: []*ttnpb.UplinkMessage{{RawPayload: []byte{}}},
		})
		a.So(err, should.BeNil)
	}

	// Sending uplink with content
	{
		ns.Add(1)
		err = link.Send(&ttnpb.GatewayUp{
			UplinkMessages: []*ttnpb.UplinkMessage{
				{RawPayload: []byte{}, EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{DevAddr: &types.DevAddr{}}},
			},
		})
		a.So(err, should.BeNil)
		ns.Wait()
	}

	downlinkContent := []byte{1, 2, 3}

	// Scheduling a downlink
	{
		_, err = gs.ScheduleDownlink(ctx, &ttnpb.DownlinkMessage{
			Settings: ttnpb.TxSettings{
				Bandwidth:       125000,
				CodingRate:      "4/5",
				Frequency:       866000000,
				Modulation:      ttnpb.Modulation_LORA,
				SpreadingFactor: 7,
			},
			RawPayload: downlinkContent,
			TxMetadata: ttnpb.TxMetadata{
				GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: registeredGatewayID},
			},
		})
		a.So(err, should.BeNil)
	}

	// Verifying the downlink has been received
	{
		down, err := link.Recv()
		a.So(err, should.BeNil)
		a.So(down.DownlinkMessage.Settings.Bandwidth, should.Equal, 125000)
		a.So(len(down.DownlinkMessage.RawPayload), should.Equal, len(downlinkContent))
	}

	// Gateway information
	{
		obs, err := gs.GetGatewayObservations(ctx, &ttnpb.GatewayIdentifiers{GatewayID: registeredGatewayID})
		a.So(err, should.BeNil)
		a.So(obs, should.NotBeNil)
	}

	ns.Add(1)
	linkCancel()
	ns.Done()

	gs.Close()
}
