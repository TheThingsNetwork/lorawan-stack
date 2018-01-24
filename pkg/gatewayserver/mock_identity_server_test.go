// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package gatewayserver_test

import (
	"context"
	"net"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/rpcserver"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	google_protobuf2 "github.com/gogo/protobuf/types"
	"google.golang.org/grpc"
)

// IsGatewayServer implements ttnpb.IsGatewayServer
type IsGatewayServer struct {
	gateways map[ttnpb.GatewayIdentifier]ttnpb.Gateway
}

func (m *IsGatewayServer) CreateGateway(context.Context, *ttnpb.CreateGatewayRequest) (*google_protobuf2.Empty, error) {
	panic("Not implemented")
}

func (m *IsGatewayServer) GetGateway(ctx context.Context, id *ttnpb.GatewayIdentifier) (*ttnpb.Gateway, error) {
	gateway, ok := m.gateways[*id]
	if !ok {
		return nil, errors.New("Gateway not found")
	}
	return &gateway, nil
}

func (m *IsGatewayServer) ListGateways(context.Context, *ttnpb.ListGatewaysRequest) (*ttnpb.ListGatewaysResponse, error) {
	panic("Not implemented")
}

func (m *IsGatewayServer) UpdateGateway(context.Context, *ttnpb.UpdateGatewayRequest) (*google_protobuf2.Empty, error) {
	panic("Not implemented")
}

func (m *IsGatewayServer) DeleteGateway(context.Context, *ttnpb.GatewayIdentifier) (*google_protobuf2.Empty, error) {
	panic("Not implemented")
}

func (m *IsGatewayServer) GenerateGatewayAPIKey(context.Context, *ttnpb.GenerateGatewayAPIKeyRequest) (*ttnpb.APIKey, error) {
	panic("Not implemented")
}

func (m *IsGatewayServer) ListGatewayAPIKeys(context.Context, *ttnpb.GatewayIdentifier) (*ttnpb.ListGatewayAPIKeysResponse, error) {
	panic("Not implemented")
}

func (m *IsGatewayServer) RemoveGatewayAPIKey(context.Context, *ttnpb.RemoveGatewayAPIKeyRequest) (*google_protobuf2.Empty, error) {
	panic("Not implemented")
}

func (m *IsGatewayServer) UpdateGatewayAPIKey(context.Context, *ttnpb.UpdateGatewayAPIKeyRequest) (*google_protobuf2.Empty, error) {
	panic("Not implemented")
}

func (m *IsGatewayServer) SetGatewayCollaborator(context.Context, *ttnpb.GatewayCollaborator) (*google_protobuf2.Empty, error) {
	panic("Not implemented")
}

func (m *IsGatewayServer) ListGatewayCollaborators(context.Context, *ttnpb.GatewayIdentifier) (*ttnpb.ListGatewayCollaboratorsResponse, error) {
	panic("Not implemented")
}

func (m *IsGatewayServer) ListGatewayRights(context.Context, *ttnpb.GatewayIdentifier) (*ttnpb.ListGatewayRightsResponse, error) {
	panic("Not implemented")
}

func StartMockIsGatewayServer(ctx context.Context, gateways map[ttnpb.GatewayIdentifier]ttnpb.Gateway) (*grpc.Server, string) {
	is := &IsGatewayServer{gateways: gateways}

	serve := func(addr string) (*grpc.Server, string) {
		srv := rpcserver.New(ctx)
		ttnpb.RegisterIsGatewayServer(srv.Server, is)

		for {
			lis, err := net.Listen("tcp", addr)
			if err == nil {
				go srv.Serve(lis)
				return srv.Server, lis.Addr().String()
			}
		}
	}

	srv, addr := serve("127.0.0.1:0")
	return srv, addr
}
