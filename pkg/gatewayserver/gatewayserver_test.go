// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package gatewayserver_test

import (
	"context"
	"os"
	"testing"
	"time"

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
	gs, err := gatewayserver.New(c, &gatewayserver.Config{})
	if err != nil {
		panic(err)
	}

	gs.Run()
}

func TestUnloadableLocalStore(t *testing.T) {
	a := assertions.New(t)

	c := component.MustNew(test.GetLogger(t), &component.Config{})
	_, err := gatewayserver.New(c, &gatewayserver.Config{
		LocalFrequencyPlansStore: os.TempDir(),
	})
	a.So(err, should.NotBeNil)
}

func TestUnloadableHTTPStore(t *testing.T) {
	a := assertions.New(t)

	c := component.MustNew(test.GetLogger(t), &component.Config{})
	_, err := gatewayserver.New(c, &gatewayserver.Config{
		HTTPFrequencyPlansStoreRoot: "http://fake-address-on-fake-port:3204834",
	})
	a.So(err, should.NotBeNil)
}

func TestGatewayServer(t *testing.T) {
	a := assertions.New(t)

	dir := createFPStore(a)
	defer removeFPStore(a, dir)

	c := component.MustNew(test.GetLogger(t), &component.Config{})
	gs, err := gatewayserver.New(c, &gatewayserver.Config{
		LocalFrequencyPlansStore: dir,
	})
	a.So(err, should.BeNil)

	roles := gs.Roles()
	a.So(len(roles), should.Equal, 1)
	a.So(roles[0], should.Equal, ttnpb.PeerInfo_GATEWAY_SERVER)

	defer gs.Close()
}

func TestLink(t *testing.T) {
	a := assertions.New(t)

	ctx := log.WithLogger(context.Background(), test.GetLogger(t))

	dir := createFPStore(a)
	defer removeFPStore(a, dir)

	registeredGatewayID, registeredGatewayUnknownFPID := "registered-gateway", "registered-gw-unknown-fp"
	registeredGatewayFP, registeredGatewayUnknownFP := "EU_863_870", "UNKNOWN_FP"
	registeredGateways := map[ttnpb.GatewayIdentifier]ttnpb.Gateway{
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

	srv := grpc.NewServer()
	gs, err := gatewayserver.New(c, &gatewayserver.Config{LocalFrequencyPlansStore: dir})
	a.So(err, should.BeNil)
	gs.RegisterServices(srv)
	err = gs.Start()
	a.So(err, should.BeNil)
	defer gs.Close()

	md := rpcmetadata.MD{ID: registeredGatewayID}
	ctx = md.ToOutgoingContext(ctx)
	conn, err := rpcserver.StartLoopback(ctx, srv, grpc.WithPerRPCCredentials(md))
	a.So(err, should.BeNil)
	client := ttnpb.NewGtwGsClient(conn)

	fp, err := client.GetFrequencyPlan(log.WithLogger(ctx, log.FromContext(ctx).WithField("situation", "client_request")), &ttnpb.FrequencyPlanRequest{FrequencyPlanID: "EU_863_870"})
	a.So(err, should.BeNil)
	a.So(fp.BandID, should.Equal, registeredGatewayFP)

	failedLinkMd := rpcmetadata.MD{ID: registeredGatewayUnknownFPID}
	failedLinkCtx := failedLinkMd.ToOutgoingContext(ctx)
	failedLinkConn, err := rpcserver.StartLoopback(failedLinkCtx, srv)
	a.So(err, should.BeNil)
	failedLinkClient := ttnpb.NewGtwGsClient(failedLinkConn)

	failedLink, err := failedLinkClient.Link(failedLinkCtx)
	a.So(err, should.BeNil)
	_, err = failedLink.Recv()
	a.So(err, should.NotBeNil)

	var link ttnpb.GtwGs_LinkClient
	linkCtx, linkCancel := context.WithCancel(ctx)
	link, err = client.Link(linkCtx, grpc.FailFast(true))
	a.So(err, should.BeNil)
	select {
	case <-ns.startServingGatewayChan:
	case <-time.After(10 * time.Second):
		a.So("timeout", should.BeNil)
	}

	err = link.Send(&ttnpb.GatewayUp{
		UplinkMessages: []*ttnpb.UplinkMessage{{RawPayload: []byte{}}},
	})
	a.So(err, should.BeNil)

	err = link.Send(&ttnpb.GatewayUp{
		UplinkMessages: []*ttnpb.UplinkMessage{
			{RawPayload: []byte{}, EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{DevAddr: &types.DevAddr{}}},
		},
	})
	a.So(err, should.BeNil)
	select {
	case <-ns.handleUplinkChan:
	case <-time.After(10 * time.Second):
		a.So("timeout", should.BeNil)
	}

	downlinkContent := []byte{1, 2, 3}
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
			GatewayIdentifier: ttnpb.GatewayIdentifier{GatewayID: registeredGatewayID},
		},
	})
	a.So(err, should.BeNil)

	down, err := link.Recv()
	a.So(err, should.BeNil)
	a.So(down.DownlinkMessage.Settings.Bandwidth, should.Equal, 125000)
	a.So(len(down.DownlinkMessage.RawPayload), should.Equal, len(downlinkContent))

	obs, err := gs.GetGatewayObservations(ctx, &ttnpb.GatewayIdentifier{GatewayID: registeredGatewayID})
	a.So(err, should.BeNil)
	a.So(obs, should.NotBeNil)

	linkCancel()
	select {
	case <-ns.stopServingGatewayChan:
	case <-time.After(10 * time.Second):
		a.So("timeout", should.BeNil)
	}

	_, err = gs.GetGatewayObservations(ctx, &ttnpb.GatewayIdentifier{GatewayID: registeredGatewayID})
	a.So(err, should.NotBeNil) // gateway disconnected

	defer gs.Close()
}
