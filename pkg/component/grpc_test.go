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

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/smartystreets/assertions"
	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

var (
	clusterKey  = "16AB00AB8D11A78316AB00AB8D11A783"
	grpcTimeout = 5 * time.Second
)

type asImplementation struct {
	*component.Component
	ttnpb.UnimplementedAppAsServer

	up chan *ttnpb.ApplicationUp
}

func (as *asImplementation) Subscribe(id *ttnpb.ApplicationIdentifiers, stream ttnpb.AppAs_SubscribeServer) error {
	if err := clusterauth.Authorized(stream.Context()); err != nil {
		return err
	}
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case up := <-as.up:
			stream.Send(up)
		}
	}
}

type gsImplementation struct {
	*component.Component
}

func (gs *gsImplementation) GetGatewayConnectionStats(ctx context.Context, _ *ttnpb.GatewayIdentifiers) (*ttnpb.GatewayConnectionStats, error) {
	if err := clusterauth.Authorized(ctx); err != nil {
		return nil, err
	}
	return &ttnpb.GatewayConnectionStats{}, nil
}

func TestHooks(t *testing.T) {
	a := assertions.New(t)

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	defer lis.Close()

	config := &component.Config{
		ServiceBase: config.ServiceBase{Cluster: config.Cluster{
			Name:          "test-cluster",
			NetworkServer: lis.Addr().String(),
			Keys:          []string{clusterKey},
		}},
	}

	c, err := component.New(test.GetLogger(t), config)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	err = c.Start()
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(errors.UnaryServerInterceptor(), hooks.UnaryServerInterceptor())),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(errors.StreamServerInterceptor(), hooks.StreamServerInterceptor())),
	)

	hooks.RegisterUnaryHook("/ttn.lorawan.v3.Gs", cluster.HookName, c.ClusterAuthUnaryHook())
	hooks.RegisterStreamHook("/ttn.lorawan.v3.AppAs", cluster.HookName, c.ClusterAuthStreamHook())

	as := &asImplementation{
		Component: c,
		up:        make(chan *ttnpb.ApplicationUp),
	}
	ttnpb.RegisterAppAsServer(s, as)
	gs := &gsImplementation{Component: c}
	ttnpb.RegisterGsServer(s, gs)
	go s.Serve(lis)
	defer s.Stop()

	grpcClient, err := grpc.Dial(lis.Addr().String(),
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(errors.UnaryClientInterceptor()),
		grpc.WithStreamInterceptor(errors.StreamClientInterceptor()),
	)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	start := time.Now()
	for grpcClient.GetState() != connectivity.Ready {
		if time.Now().Sub(start) > grpcTimeout {
			t.Fatal("The gRPC client did not become ready")
		}
		time.Sleep(time.Millisecond)
	}
	asClient := ttnpb.NewAppAsClient(grpcClient)
	gsClient := ttnpb.NewGsClient(grpcClient)

	ctx := test.Context()

	// Failing calls
	{
		_, err = gsClient.GetGatewayConnectionStats(ctx, &ttnpb.GatewayIdentifiers{})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsUnauthenticated(err), should.BeTrue)
		}

		sub, err := asClient.Subscribe(ctx, &ttnpb.ApplicationIdentifiers{})
		a.So(err, should.BeNil)

		_, err = sub.Recv()
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsUnauthenticated(err), should.BeTrue)
		}
	}

	// Successful calls
	{
		_, err = gsClient.GetGatewayConnectionStats(ctx, &ttnpb.GatewayIdentifiers{}, c.WithClusterAuth())
		a.So(err, should.BeNil)
		sub, err := asClient.Subscribe(ctx, &ttnpb.ApplicationIdentifiers{}, c.WithClusterAuth())
		a.So(err, should.BeNil)
		go func() {
			as.up <- &ttnpb.ApplicationUp{
				Up: &ttnpb.ApplicationUp_UplinkMessage{
					UplinkMessage: &ttnpb.ApplicationUplink{
						SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
					},
				},
			}
		}()
		up, err := sub.Recv()
		a.So(err, should.BeNil)
		a.So(up.GetUplinkMessage().SessionKeyID, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
	}
}
