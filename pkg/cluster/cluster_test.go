// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package cluster

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/rpcmiddleware/rpclog"
	"github.com/TheThingsNetwork/ttn/pkg/rpcserver"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

var ctx context.Context

func init() {
	logger, _ := log.NewLogger(
		log.WithLevel(log.DebugLevel),
		log.WithHandler(log.NewCLI(os.Stdout)),
	)
	ctx = log.NewContext(context.Background(), logger.WithField("namespace", "cluster"))
	rpclog.ReplaceGrpcLogger(logger.WithField("namespace", "grpc"))
}

func TestCluster(t *testing.T) {
	a := assertions.New(t)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer lis.Close()

	go grpc.NewServer().Serve(lis)

	config := &config.ServiceBase{Cluster: config.Cluster{
		Address:           lis.Addr().String(),
		IdentityServer:    lis.Addr().String(),
		GatewayServer:     lis.Addr().String(),
		NetworkServer:     lis.Addr().String(),
		ApplicationServer: lis.Addr().String(),
		JoinServer:        lis.Addr().String(),
		Join:              []string{lis.Addr().String()},
	}}

	c, err := New(context.Background(), config, []rpcserver.Registerer{}...)
	a.So(err, should.BeNil)

	a.So(c.Join(), should.BeNil)

	grpc.Dial(lis.Addr().String(), grpc.WithInsecure(), grpc.WithBlock())

	// The IS should be there within reasonable time
	var is Peer
	for i := 0; i < 20; i++ {
		time.Sleep(20 * time.Millisecond) // Wait for peers to join cluster
		is = c.GetPeer(ttnpb.PeerInfo_IDENTITY_SERVER, nil, nil)
		if is != nil {
			break
		}
	}
	if !a.So(is, should.NotBeNil) {
		t.FailNow()
	}

	// The others should also be there
	gs := c.GetPeer(ttnpb.PeerInfo_GATEWAY_SERVER, nil, nil)
	a.So(gs, should.NotBeNil)
	ns := c.GetPeer(ttnpb.PeerInfo_NETWORK_SERVER, nil, nil)
	a.So(ns, should.NotBeNil)
	as := c.GetPeer(ttnpb.PeerInfo_APPLICATION_SERVER, nil, nil)
	a.So(as, should.NotBeNil)
	js := c.GetPeer(ttnpb.PeerInfo_JOIN_SERVER, nil, nil)
	a.So(js, should.NotBeNil)

	a.So(c.Leave(), should.BeNil)

	a.So(is.Conn().GetState(), should.Equal, connectivity.Shutdown)
	a.So(gs.Conn().GetState(), should.Equal, connectivity.Shutdown)
	a.So(ns.Conn().GetState(), should.Equal, connectivity.Shutdown)
	a.So(as.Conn().GetState(), should.Equal, connectivity.Shutdown)
	a.So(js.Conn().GetState(), should.Equal, connectivity.Shutdown)
}
