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

package cluster_test

import (
	"context"
	"net"
	"os"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/rpclog"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

var ctx context.Context

func init() {
	logger, _ := log.NewLogger(
		log.WithLevel(log.DebugLevel),
		log.WithHandler(log.NewCLI(os.Stdout)),
	)
	ctx = log.NewContext(test.Context(), logger.WithField("namespace", "cluster"))
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

	config := config.Cluster{
		Address:           lis.Addr().String(),
		IdentityServer:    lis.Addr().String(),
		GatewayServer:     lis.Addr().String(),
		NetworkServer:     lis.Addr().String(),
		ApplicationServer: lis.Addr().String(),
		JoinServer:        lis.Addr().String(),
		Join:              []string{lis.Addr().String()},
	}

	ctx := test.Context()

	c, err := New(ctx, &config)
	a.So(err, should.BeNil)

	a.So(c.Join(), should.BeNil)

	grpc.Dial(lis.Addr().String(), grpc.WithInsecure(), grpc.WithBlock())

	// The Identity Server playing the ACCESS role should be there within reasonable time.
	var ac Peer
	for i := 0; i < 20; i++ {
		time.Sleep(20 * time.Millisecond) // Wait for peers to join cluster.
		ac, err = c.GetPeer(ctx, ttnpb.ClusterRole_ACCESS, nil)
		if err == nil {
			break
		}
	}
	if !a.So(ac, should.NotBeNil) {
		t.FailNow()
	}

	er, err := c.GetPeer(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	a.So(er, should.NotBeNil)
	a.So(err, should.BeNil)
	gs, err := c.GetPeer(ctx, ttnpb.ClusterRole_GATEWAY_SERVER, nil)
	a.So(gs, should.NotBeNil)
	a.So(err, should.BeNil)
	ns, err := c.GetPeer(ctx, ttnpb.ClusterRole_NETWORK_SERVER, nil)
	a.So(ns, should.NotBeNil)
	a.So(err, should.BeNil)
	as, err := c.GetPeer(ctx, ttnpb.ClusterRole_APPLICATION_SERVER, nil)
	a.So(as, should.NotBeNil)
	a.So(err, should.BeNil)
	js, err := c.GetPeer(ctx, ttnpb.ClusterRole_JOIN_SERVER, nil)
	a.So(js, should.NotBeNil)
	a.So(err, should.BeNil)

	a.So(c.Leave(), should.BeNil)

	a.So(ac.Conn().GetState(), should.Equal, connectivity.Shutdown)
	a.So(er.Conn().GetState(), should.Equal, connectivity.Shutdown)
	a.So(gs.Conn().GetState(), should.Equal, connectivity.Shutdown)
	a.So(ns.Conn().GetState(), should.Equal, connectivity.Shutdown)
	a.So(as.Conn().GetState(), should.Equal, connectivity.Shutdown)
	a.So(js.Conn().GetState(), should.Equal, connectivity.Shutdown)
}
