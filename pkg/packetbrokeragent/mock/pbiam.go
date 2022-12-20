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

package mock

import (
	"context"
	"testing"

	pbtypes "github.com/gogo/protobuf/types"
	iampb "go.packetbroker.org/api/iam"
	iampbv2 "go.packetbroker.org/api/iam/v2"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"google.golang.org/grpc"
)

// PBIAM is a mock Packet Broker IAM.
type PBIAM struct {
	*grpc.Server
	Registry struct {
		ListNetworksHandler  func(ctx context.Context, req *iampb.ListNetworksRequest) (*iampb.ListNetworksResponse, error)
		CreateNetworkHandler func(ctx context.Context, req *iampb.CreateNetworkRequest) (*iampb.CreateNetworkResponse, error)
		GetNetworkHandler    func(ctx context.Context, req *iampb.NetworkRequest) (*iampb.GetNetworkResponse, error)
		UpdateNetworkHandler func(ctx context.Context, req *iampb.UpdateNetworkRequest) (*pbtypes.Empty, error)
		DeleteNetworkHandler func(ctx context.Context, req *iampb.NetworkRequest) (*pbtypes.Empty, error)
		ListTenantsHandler   func(ctx context.Context, req *iampb.ListTenantsRequest) (*iampb.ListTenantsResponse, error)
		CreateTenantHandler  func(ctx context.Context, req *iampb.CreateTenantRequest) (*iampb.CreateTenantResponse, error)
		GetTenantHandler     func(ctx context.Context, req *iampb.TenantRequest) (*iampb.GetTenantResponse, error)
		UpdateTenantHandler  func(ctx context.Context, req *iampb.UpdateTenantRequest) (*pbtypes.Empty, error)
		DeleteTenantHandler  func(ctx context.Context, req *iampb.TenantRequest) (*pbtypes.Empty, error)
	}
	Catalog struct {
		ListNetworksHandler     func(ctx context.Context, req *iampbv2.ListNetworksRequest) (*iampbv2.ListNetworksResponse, error)
		ListHomeNetworksHandler func(ctx context.Context, req *iampbv2.ListNetworksRequest) (*iampbv2.ListNetworksResponse, error)
		ListJoinServersHandler  func(ctx context.Context, req *iampbv2.ListJoinServersRequest) (*iampbv2.ListJoinServersResponse, error)
	}
}

// NewPBIAM instantiates a new mock Packet Broker IAM.
func NewPBIAM(tb testing.TB) *PBIAM {
	iam := &PBIAM{
		Server: grpc.NewServer(
			grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
				ctx = test.ContextWithTB(ctx, tb)
				return handler(ctx, req)
			}),
		),
	}
	iampb.RegisterNetworkRegistryServer(iam.Server, &pbIAMRegistry{PBIAM: iam})
	iampb.RegisterTenantRegistryServer(iam.Server, &pbIAMRegistry{PBIAM: iam})
	iampbv2.RegisterCatalogServer(iam.Server, &pbIAMCatalog{PBIAM: iam})
	return iam
}

type pbIAMRegistry struct {
	iampb.UnimplementedNetworkRegistryServer
	iampb.UnimplementedTenantRegistryServer

	*PBIAM
}

func (s *pbIAMRegistry) ListNetworks(ctx context.Context, req *iampb.ListNetworksRequest) (*iampb.ListNetworksResponse, error) {
	if s.Registry.ListNetworksHandler == nil {
		panic("ListNetworks called but not set")
	}
	return s.Registry.ListNetworksHandler(ctx, req)
}

func (s *pbIAMRegistry) CreateNetwork(ctx context.Context, req *iampb.CreateNetworkRequest) (*iampb.CreateNetworkResponse, error) {
	if s.Registry.CreateNetworkHandler == nil {
		panic("CreateNetwork called but not set")
	}
	return s.Registry.CreateNetworkHandler(ctx, req)
}

func (s *pbIAMRegistry) GetNetwork(ctx context.Context, req *iampb.NetworkRequest) (*iampb.GetNetworkResponse, error) {
	if s.Registry.GetNetworkHandler == nil {
		panic("GetNetwork called but not set")
	}
	return s.Registry.GetNetworkHandler(ctx, req)
}

func (s *pbIAMRegistry) UpdateNetwork(ctx context.Context, req *iampb.UpdateNetworkRequest) (*pbtypes.Empty, error) {
	if s.Registry.UpdateNetworkHandler == nil {
		panic("UpdateNetwork called but not set")
	}
	return s.Registry.UpdateNetworkHandler(ctx, req)
}

func (s *pbIAMRegistry) DeleteNetwork(ctx context.Context, req *iampb.NetworkRequest) (*pbtypes.Empty, error) {
	if s.Registry.DeleteNetworkHandler == nil {
		panic("DeleteNetwork called but not set")
	}
	return s.Registry.DeleteNetworkHandler(ctx, req)
}

func (s *pbIAMRegistry) ListTenants(ctx context.Context, req *iampb.ListTenantsRequest) (*iampb.ListTenantsResponse, error) {
	if s.Registry.ListTenantsHandler == nil {
		panic("ListTenants called but not set")
	}
	return s.Registry.ListTenantsHandler(ctx, req)
}

func (s *pbIAMRegistry) CreateTenant(ctx context.Context, req *iampb.CreateTenantRequest) (*iampb.CreateTenantResponse, error) {
	if s.Registry.CreateTenantHandler == nil {
		panic("CreateTenant called but not set")
	}
	return s.Registry.CreateTenantHandler(ctx, req)
}

func (s *pbIAMRegistry) GetTenant(ctx context.Context, req *iampb.TenantRequest) (*iampb.GetTenantResponse, error) {
	if s.Registry.GetTenantHandler == nil {
		panic("GetTenant called but not set")
	}
	return s.Registry.GetTenantHandler(ctx, req)
}

func (s *pbIAMRegistry) UpdateTenant(ctx context.Context, req *iampb.UpdateTenantRequest) (*pbtypes.Empty, error) {
	if s.Registry.UpdateTenantHandler == nil {
		panic("UpdateTenant called but not set")
	}
	return s.Registry.UpdateTenantHandler(ctx, req)
}

func (s *pbIAMRegistry) DeleteTenant(ctx context.Context, req *iampb.TenantRequest) (*pbtypes.Empty, error) {
	if s.Registry.DeleteTenantHandler == nil {
		panic("DeleteTenant called but not set")
	}
	return s.Registry.DeleteTenantHandler(ctx, req)
}

type pbIAMCatalog struct {
	iampbv2.UnimplementedCatalogServer

	*PBIAM
}

func (s *pbIAMCatalog) ListNetworks(ctx context.Context, req *iampbv2.ListNetworksRequest) (*iampbv2.ListNetworksResponse, error) {
	if s.Catalog.ListNetworksHandler == nil {
		panic("ListHomeNetworks called but not set")
	}
	return s.Catalog.ListNetworksHandler(ctx, req)
}

func (s *pbIAMCatalog) ListHomeNetworks(ctx context.Context, req *iampbv2.ListNetworksRequest) (*iampbv2.ListNetworksResponse, error) {
	if s.Catalog.ListHomeNetworksHandler == nil {
		panic("ListHomeNetworks called but not set")
	}
	return s.Catalog.ListHomeNetworksHandler(ctx, req)
}

func (s *pbIAMCatalog) ListJoinServers(ctx context.Context, req *iampbv2.ListJoinServersRequest) (*iampbv2.ListJoinServersResponse, error) {
	if s.Catalog.ListJoinServersHandler == nil {
		panic("ListJoinServers called but not set")
	}
	return s.Catalog.ListJoinServersHandler(ctx, req)
}
