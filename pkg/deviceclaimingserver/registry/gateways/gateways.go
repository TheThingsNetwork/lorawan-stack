// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

// Package gateways provide gateway registry functions.
package gateways

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Cluster provides cluster operations.
type Cluster interface {
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole, ids cluster.EntityIdentifiers) (*grpc.ClientConn, error)
	AllowInsecureForCredentials() bool
	WithClusterAuth() grpc.CallOption
}

// GatewayRegistry abstracts the entity registry's gateway functions.
type GatewayRegistry interface {
	// AssertGatewayRights checks whether the gateway authentication
	// (provided in the context) contains the required rights.
	AssertGatewayRights(ctx context.Context, ids *ttnpb.GatewayIdentifiers, required ...ttnpb.Right) error
	// GetIdentifiersForEUI returns the gateway identifiers for the EUI.
	GetIdentifiersForEUI(ctx context.Context, eui types.EUI64) (*ttnpb.GatewayIdentifiers, error)
	// Create creates a gateway.
	Create(ctx context.Context, in *ttnpb.CreateGatewayRequest) (*ttnpb.Gateway, error)
	// Delete the gateway. This may not release the gateway ID for reuse, but it does release the EUI.
	Delete(ctx context.Context, in *ttnpb.GatewayIdentifiers) (*emptypb.Empty, error)
	// Get the gateway. This may not release the gateway ID for reuse, but it does release the EUI.
	Get(ctx context.Context, req *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error)
}

// Registry implements GatewayRegistry.
type Registry struct {
	Cluster
}

func (reg Registry) newEntityRegistryClient(ctx context.Context) (ttnpb.GatewayRegistryClient, error) {
	cc, err := reg.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	if err != nil {
		return nil, err
	}
	return ttnpb.NewGatewayRegistryClient(cc), nil
}

// callOptFromContext returns a gRPC call option from the provided context.
func (reg Registry) callOptFromContext(ctx context.Context) (grpc.CallOption, error) {
	return rpcmetadata.WithForwardedAuth(ctx, reg.AllowInsecureForCredentials())
}

// GetIdentifiersForEUI implements GatewayRegistry.
func (reg Registry) GetIdentifiersForEUI(
	ctx context.Context,
	gatewayEUI types.EUI64,
) (*ttnpb.GatewayIdentifiers, error) {
	// Check if the gateway is registered.
	gatewayRegistry, err := reg.newEntityRegistryClient(ctx)
	if err != nil {
		return nil, err
	}
	return gatewayRegistry.GetIdentifiersForEUI(ctx, &ttnpb.GetGatewayIdentifiersForEUIRequest{
		Eui: gatewayEUI.Bytes(),
	}, reg.WithClusterAuth())
}

// AssertGatewayRights implements GatewayRegistry.
func (Registry) AssertGatewayRights(
	ctx context.Context,
	ids *ttnpb.GatewayIdentifiers,
	required ...ttnpb.Right,
) error {
	return rights.RequireGateway(ctx, ids, required...)
}

// Create implements GatewayRegistry.
func (reg Registry) Create(ctx context.Context, req *ttnpb.CreateGatewayRequest) (*ttnpb.Gateway, error) {
	callOpt, err := reg.callOptFromContext(ctx)
	if err != nil {
		return nil, err
	}
	gatewayRegistry, err := reg.newEntityRegistryClient(ctx)
	if err != nil {
		return nil, err
	}
	return gatewayRegistry.Create(ctx, req, callOpt)
}

// Delete implements GatewayRegistry.
func (reg Registry) Delete(ctx context.Context, req *ttnpb.GatewayIdentifiers) (*emptypb.Empty, error) {
	callOpt, err := reg.callOptFromContext(ctx)
	if err != nil {
		return nil, err
	}
	gatewayRegistry, err := reg.newEntityRegistryClient(ctx)
	if err != nil {
		return nil, err
	}
	return gatewayRegistry.Delete(ctx, req, callOpt)
}

// Get implements GatewayRegistry.
func (reg Registry) Get(ctx context.Context, req *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error) {
	callOpt, err := reg.callOptFromContext(ctx)
	if err != nil {
		return nil, err
	}
	gatewayRegistry, err := reg.newEntityRegistryClient(ctx)
	if err != nil {
		return nil, err
	}
	return gatewayRegistry.Get(ctx, req, callOpt)
}
