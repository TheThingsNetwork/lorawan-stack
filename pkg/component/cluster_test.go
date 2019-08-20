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

package component_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestPeers(t *testing.T) {
	a := assertions.New(t)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer lis.Close()

	ctx, cancel := context.WithCancel(test.Context())
	defer cancel()

	srv := rpcserver.New(ctx)
	go srv.Serve(lis)
	defer srv.Stop()

	var c *component.Component

	config := &component.Config{
		ServiceBase: config.ServiceBase{Cluster: config.Cluster{
			Name:          "test-cluster",
			NetworkServer: lis.Addr().String(),
			TLS:           false,
		}},
	}

	c, err = component.New(test.GetLogger(t), config)
	a.So(err, should.BeNil)
	err = c.Start()
	a.So(err, should.BeNil)

	unusedRoles := []ttnpb.ClusterRole{
		ttnpb.ClusterRole_APPLICATION_SERVER,
		ttnpb.ClusterRole_GATEWAY_SERVER,
		ttnpb.ClusterRole_JOIN_SERVER,
		ttnpb.ClusterRole_ACCESS,
		ttnpb.ClusterRole_ENTITY_REGISTRY,
	}

	var peer cluster.Peer
	for i := 0; i < 20; i++ {
		time.Sleep(20 * time.Millisecond) // Wait for peers to join cluster.
		peer = c.GetPeer(context.Background(), ttnpb.ClusterRole_NETWORK_SERVER, nil)
		if peer != nil {
			break
		}
	}

	if !a.So(peer, should.NotBeNil) {
		t.FailNow()
	}

	conn := peer.Conn()
	a.So(conn, should.NotBeNil)

	for _, role := range unusedRoles {
		peer = c.GetPeer(context.Background(), role, nil)
		a.So(peer, should.BeNil)
	}

	peers := c.GetPeers(context.Background(), ttnpb.ClusterRole_NETWORK_SERVER)
	a.So(peers, should.HaveLength, 1)

	for _, role := range unusedRoles {
		peers = c.GetPeers(context.Background(), role)
		a.So(peers, should.HaveLength, 0)
	}
}
