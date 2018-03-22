// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package gatewayserver_test

import (
	"context"
	"net"
	"sync"

	"github.com/TheThingsNetwork/ttn/pkg/rpcserver"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/gogo/protobuf/types"
)

// GsNsServer implements ttnpb.GsNsServer
type GsNsServer struct {
	*sync.WaitGroup
}

func (s GsNsServer) StartServingGateway(_ context.Context, id *ttnpb.GatewayIdentifier) (*types.Empty, error) {
	s.Done()
	return &types.Empty{}, nil
}

func (s GsNsServer) StopServingGateway(_ context.Context, id *ttnpb.GatewayIdentifier) (*types.Empty, error) {
	s.Done()
	return &types.Empty{}, nil
}

func (s GsNsServer) HandleUplink(_ context.Context, up *ttnpb.UplinkMessage) (*types.Empty, error) {
	s.Done()
	return &types.Empty{}, nil
}

func StartMockGsNsServer(ctx context.Context) (GsNsServer, string) {
	ns := GsNsServer{WaitGroup: &sync.WaitGroup{}}

	serve := func(ctx context.Context, addr string) string {
		srv := rpcserver.New(ctx)
		ttnpb.RegisterGsNsServer(srv.Server, ns)

		for {
			lis, err := net.Listen("tcp", addr)
			if err == nil {
				go srv.Serve(lis)
				return lis.Addr().String()
			}
		}
	}

	addr := serve(ctx, "127.0.0.1:0")
	return ns, addr
}
