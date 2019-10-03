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

package mock

import (
	"context"
	"fmt"
	"net"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/rpcserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc/metadata"
)

// IdentityServer is the mock IS for GS tests.
type IdentityServer struct {
	ttnpb.GatewayRegistryServer
	ttnpb.GatewayAccessServer
	gateways     map[string]*ttnpb.Gateway
	gatewayAuths map[string][]string
}

// NewIS creates and starts an instance of the IS.
func NewIS(ctx context.Context) (*IdentityServer, string) {
	is := &IdentityServer{
		gateways:     make(map[string]*ttnpb.Gateway),
		gatewayAuths: make(map[string][]string),
	}
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

// Add adds a new gateway.
func (is *IdentityServer) Add(ctx context.Context, ids ttnpb.GatewayIdentifiers, key string) {
	uid := unique.ID(ctx, ids)
	is.gateways[uid] = &ttnpb.Gateway{
		GatewayIdentifiers: ids,
	}
	if key != "" {
		is.gatewayAuths[uid] = []string{fmt.Sprintf("Bearer %v", key)}
	}
}

var errNotFound = errors.DefineNotFound("not_found", "not found")

// Get retrives the Gateway.
func (is *IdentityServer) Get(ctx context.Context, req *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error) {
	uid := unique.ID(ctx, req.GatewayIdentifiers)
	app, ok := is.gateways[uid]
	if !ok {
		return nil, errNotFound
	}
	return app, nil
}

// ListRights lists the rights from context for a particular gateway.
func (is *IdentityServer) ListRights(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (res *ttnpb.Rights, err error) {
	res = &ttnpb.Rights{}
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return
	}
	authorization, ok := md["authorization"]
	if !ok || len(authorization) == 0 {
		return
	}
	auths, ok := is.gatewayAuths[unique.ID(ctx, *ids)]
	if !ok {
		return
	}
	for _, auth := range auths {
		if auth == authorization[0] {
			res.Rights = append(res.Rights,
				ttnpb.RIGHT_GATEWAY_INFO,
				ttnpb.RIGHT_GATEWAY_LINK,
			)
		}
	}
	return
}
