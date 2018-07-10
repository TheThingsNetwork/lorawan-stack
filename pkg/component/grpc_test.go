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

package component_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
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
	up chan *ttnpb.ApplicationUp
}

// Subscribe implements ttnpb.AsServer
func (as *asImplementation) Subscribe(id *ttnpb.ApplicationIdentifiers, stream ttnpb.As_SubscribeServer) error {
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case up := <-as.up:
			stream.Send(up)
		}
	}
}

type gsImplementation struct{}

// GetGatewayObservations implements ttnpb.GsServer
func (gs *gsImplementation) GetGatewayObservations(_ context.Context, _ *ttnpb.GatewayIdentifiers) (*ttnpb.GatewayObservations, error) {
	return &ttnpb.GatewayObservations{}, nil
}

func TestUnaryHook(t *testing.T) {
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

	hooks.RegisterUnaryHook("/ttn.lorawan.v3.Gs", cluster.HookName, c.UnaryHook())
	hooks.RegisterStreamHook("/ttn.lorawan.v3.As", cluster.HookName, c.StreamHook())

	as := &asImplementation{up: make(chan *ttnpb.ApplicationUp)}
	ttnpb.RegisterAsServer(s, as)
	gs := &gsImplementation{}
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
	asClient := ttnpb.NewAsClient(grpcClient)
	gsClient := ttnpb.NewGsClient(grpcClient)

	ctx := context.Background()

	// Failing calls
	{
		_, err = gsClient.GetGatewayObservations(ctx, &ttnpb.GatewayIdentifiers{})
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
		_, err = gsClient.GetGatewayObservations(ctx, &ttnpb.GatewayIdentifiers{}, c.ClusterAuth())
		a.So(err, should.BeNil)
		sub, err := asClient.Subscribe(ctx, &ttnpb.ApplicationIdentifiers{}, c.ClusterAuth())
		a.So(err, should.BeNil)
		go func() {
			as.up <- &ttnpb.ApplicationUp{
				Up: &ttnpb.ApplicationUp_UplinkMessage{
					UplinkMessage: &ttnpb.ApplicationUplink{
						SessionKeyID: "key",
					},
				},
			}
		}()
		up, err := sub.Recv()
		a.So(err, should.BeNil)
		a.So(up.GetUplinkMessage().SessionKeyID, should.Equal, "key")
	}
}
