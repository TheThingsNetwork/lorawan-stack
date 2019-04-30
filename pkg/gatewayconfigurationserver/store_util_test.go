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

package gatewayconfigurationserver_test

import (
	"context"
	"net"

	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type mockGatewayClientData struct {
	ctx struct {
		GetIdentifiersForEUI context.Context
		Create               context.Context
		Get                  context.Context
		Update               context.Context
		List                 context.Context
		Delete               context.Context
	}
	req struct {
		GetIdentifiersForEUI *ttnpb.GetGatewayIdentifiersForEUIRequest
		Create               *ttnpb.CreateGatewayRequest
		Get                  *ttnpb.GetGatewayRequest
		Update               *ttnpb.UpdateGatewayRequest
		List                 *ttnpb.ListGatewaysRequest
		Delete               *ttnpb.GatewayIdentifiers
	}
	res struct {
		GetIdentifiersForEUI *ttnpb.GatewayIdentifiers
		Create               *ttnpb.Gateway
		Get                  *ttnpb.Gateway
		Update               *ttnpb.Gateway
		List                 *ttnpb.Gateways
		Delete               *types.Empty
	}
	err struct {
		GetIdentifiersForEUI error
		Create               error
		Get                  error
		Update               error
		List                 error
		Delete               error
	}
}

type mockGatewayClient struct {
	ttnpb.GatewayRegistryServer
	ttnpb.GatewayAccessServer
	mockGatewayClientData
}

func startMockIS(ctx context.Context) (*mockGatewayClient, string) {
	is := &mockGatewayClient{}
	srv := rpcserver.New(ctx)
	ttnpb.RegisterGatewayRegistryServer(srv.Server, is)
	ttnpb.RegisterGatewayAccessServer(srv.Server, is)
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	go srv.Serve(lis)
	return is, lis.Addr().String()
}

func (m *mockGatewayClient) reset() {
	m.mockGatewayClientData = mockGatewayClientData{}
}

func (m *mockGatewayClient) GetIdentifiersForEUI(ctx context.Context, in *ttnpb.GetGatewayIdentifiersForEUIRequest) (*ttnpb.GatewayIdentifiers, error) {
	m.ctx.GetIdentifiersForEUI, m.req.GetIdentifiersForEUI = ctx, in
	return m.res.GetIdentifiersForEUI, m.err.GetIdentifiersForEUI
}

func (m *mockGatewayClient) Create(ctx context.Context, in *ttnpb.CreateGatewayRequest) (*ttnpb.Gateway, error) {
	m.ctx.Create, m.req.Create = ctx, in
	return m.res.Create, m.err.Create
}

func (m *mockGatewayClient) Get(ctx context.Context, in *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error) {
	m.ctx.Get, m.req.Get = ctx, in
	return m.res.Get, m.err.Get
}

func (m *mockGatewayClient) Update(ctx context.Context, in *ttnpb.UpdateGatewayRequest) (*ttnpb.Gateway, error) {
	m.ctx.Update, m.req.Update = ctx, in
	return m.res.Update, m.err.Update
}

func (m *mockGatewayClient) Delete(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (*types.Empty, error) {
	m.ctx.Delete, m.req.Delete = ctx, ids
	return m.res.Delete, m.err.Delete
}

func (m *mockGatewayClient) List(ctx context.Context, in *ttnpb.ListGatewaysRequest) (*ttnpb.Gateways, error) {
	m.ctx.List, m.req.List = ctx, in
	return m.res.List, m.err.List
}
