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
	"net"

	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// IsGatewayServer implements ttnpb.IsGatewayServer
type IsGatewayServer struct {
	gateways []ttnpb.Gateway
}

func (m *IsGatewayServer) CreateGateway(context.Context, *ttnpb.CreateGatewayRequest) (*types.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (m *IsGatewayServer) GetGateway(ctx context.Context, id *ttnpb.GatewayIdentifiers) (*ttnpb.Gateway, error) {
	for _, storedGateway := range m.gateways {
		if storedGateway.GatewayIdentifiers.Contains(*id) {
			return &storedGateway, nil
		}
	}

	return nil, errors.New("Gateway not found")
}

func (m *IsGatewayServer) ListGateways(context.Context, *ttnpb.ListGatewaysRequest) (*ttnpb.ListGatewaysResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (m *IsGatewayServer) UpdateGateway(context.Context, *ttnpb.UpdateGatewayRequest) (*types.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (m *IsGatewayServer) DeleteGateway(context.Context, *ttnpb.GatewayIdentifiers) (*types.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (m *IsGatewayServer) GenerateGatewayAPIKey(context.Context, *ttnpb.GenerateGatewayAPIKeyRequest) (*ttnpb.APIKey, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (m *IsGatewayServer) ListGatewayAPIKeys(context.Context, *ttnpb.GatewayIdentifiers) (*ttnpb.ListGatewayAPIKeysResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (m *IsGatewayServer) RemoveGatewayAPIKey(context.Context, *ttnpb.RemoveGatewayAPIKeyRequest) (*types.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (m *IsGatewayServer) UpdateGatewayAPIKey(context.Context, *ttnpb.UpdateGatewayAPIKeyRequest) (*types.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (m *IsGatewayServer) SetGatewayCollaborator(context.Context, *ttnpb.GatewayCollaborator) (*types.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (m *IsGatewayServer) ListGatewayCollaborators(context.Context, *ttnpb.GatewayIdentifiers) (*ttnpb.ListGatewayCollaboratorsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func (m *IsGatewayServer) ListGatewayRights(context.Context, *ttnpb.GatewayIdentifiers) (*ttnpb.ListGatewayRightsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

func StartMockIsGatewayServer(ctx context.Context, gateways []ttnpb.Gateway) (*grpc.Server, string) {
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
