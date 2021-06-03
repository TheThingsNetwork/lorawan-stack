// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

// Package is abstracts the Identity Server Gateway functions.
package is

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
)

// Cluster provides cluster operations.
type Cluster interface {
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole, ids cluster.EntityIdentifiers) (*grpc.ClientConn, error)
}

// IS exposes Identity Server functions.
type IS struct {
	Cluster
	gatewayRegistry ttnpb.GatewayRegistryClient
}

// New returns a new IS.
func New(c Cluster) *IS {
	return &IS{
		Cluster: c,
	}
}

// SetGatewayRegistry overrides the gateway registry of the IS.
func (is IS) SetGatewayRegistry(_ context.Context, registry ttnpb.GatewayRegistryClient) {
	is.gatewayRegistry = registry
}

// AssertGatewayRights implements EntityRegistry.
func (is IS) AssertGatewayRights(ctx context.Context, ids ttnpb.GatewayIdentifiers, required ...ttnpb.Right) error {
	return rights.RequireGateway(ctx, ids, required...)
}

// GetIdentifiersForEUI implements EntityRegistry.
func (is IS) GetIdentifiersForEUI(ctx context.Context, req *ttnpb.GetGatewayIdentifiersForEUIRequest, opts ...grpc.CallOption) (*ttnpb.GatewayIdentifiers, error) {
	registry, err := is.getRegistry(ctx, &ttnpb.GatewayIdentifiers{Eui: &req.Eui})
	if err != nil {
		return nil, err
	}

	return registry.GetIdentifiersForEUI(ctx, req, opts...)
}

// Get implements EntityRegistry.
func (is IS) Get(ctx context.Context, req *ttnpb.GetGatewayRequest, opts ...grpc.CallOption) (*ttnpb.Gateway, error) {
	registry, err := is.getRegistry(ctx, &req.GatewayIdentifiers)
	if err != nil {
		return nil, err
	}
	return registry.Get(ctx, req, opts...)
}

// Update the gateway, changing the fields specified by the field mask to the provided values.
func (is IS) Update(ctx context.Context, req *ttnpb.UpdateGatewayRequest, opts ...grpc.CallOption) (*ttnpb.Gateway, error) {
	registry, err := is.getRegistry(ctx, &req.GatewayIdentifiers)
	if err != nil {
		return nil, err
	}
	return registry.Update(ctx, req, opts...)
}

func (is IS) getRegistry(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (ttnpb.GatewayRegistryClient, error) {
	if is.gatewayRegistry != nil {
		return is.gatewayRegistry, nil
	}
	var (
		cc  *grpc.ClientConn
		err error
	)
	if ids != nil {
		cc, err = is.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, ids)
	} else {
		// Don't pass a (*ttnpb.GatewayIdentifiers)(nil) to GetPeerConn.
		cc, err = is.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	}
	if err != nil {
		return nil, err
	}
	return ttnpb.NewGatewayRegistryClient(cc), nil
}
