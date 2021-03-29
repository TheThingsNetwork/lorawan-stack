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
	ListNetworksHandler     func(ctx context.Context, req *iampb.ListNetworksRequest) (*iampb.ListNetworksResponse, error)
	CreateNetworkHandler    func(ctx context.Context, req *iampb.CreateNetworkRequest) (*iampb.CreateNetworkResponse, error)
	GetNetworkHandler       func(ctx context.Context, req *iampb.NetworkRequest) (*iampb.GetNetworkResponse, error)
	UpdateNetworkHandler    func(ctx context.Context, req *iampb.UpdateNetworkRequest) (*pbtypes.Empty, error)
	DeleteNetworkHandler    func(ctx context.Context, req *iampb.NetworkRequest) (*pbtypes.Empty, error)
	ListTenantsHandler      func(ctx context.Context, req *iampb.ListTenantsRequest) (*iampb.ListTenantsResponse, error)
	CreateTenantHandler     func(ctx context.Context, req *iampb.CreateTenantRequest) (*iampb.CreateTenantResponse, error)
	GetTenantHandler        func(ctx context.Context, req *iampb.TenantRequest) (*iampb.GetTenantResponse, error)
	UpdateTenantHandler     func(ctx context.Context, req *iampb.UpdateTenantRequest) (*pbtypes.Empty, error)
	DeleteTenantHandler     func(ctx context.Context, req *iampb.TenantRequest) (*pbtypes.Empty, error)
	ListHomeNetworksHandler func(ctx context.Context, req *iampbv2.ListHomeNetworksRequest) (*iampbv2.ListHomeNetworksResponse, error)
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
	iampb.RegisterNetworkRegistryServer(iam.Server, iam)
	iampb.RegisterTenantRegistryServer(iam.Server, iam)
	iampbv2.RegisterCatalogServer(iam.Server, iam)
	return iam
}

func (s *PBIAM) ListNetworks(ctx context.Context, req *iampb.ListNetworksRequest) (*iampb.ListNetworksResponse, error) {
	if s.ListNetworksHandler == nil {
		panic("ListNetworks called but not set")
	}
	return s.ListNetworksHandler(ctx, req)
}

func (s *PBIAM) CreateNetwork(ctx context.Context, req *iampb.CreateNetworkRequest) (*iampb.CreateNetworkResponse, error) {
	if s.CreateNetworkHandler == nil {
		panic("CreateNetwork called but not set")
	}
	return s.CreateNetworkHandler(ctx, req)
}

func (s *PBIAM) GetNetwork(ctx context.Context, req *iampb.NetworkRequest) (*iampb.GetNetworkResponse, error) {
	if s.GetNetworkHandler == nil {
		panic("GetNetwork called but not set")
	}
	return s.GetNetworkHandler(ctx, req)
}

func (s *PBIAM) UpdateNetwork(ctx context.Context, req *iampb.UpdateNetworkRequest) (*pbtypes.Empty, error) {
	if s.UpdateNetworkHandler == nil {
		panic("UpdateNetwork called but not set")
	}
	return s.UpdateNetworkHandler(ctx, req)
}

func (s *PBIAM) DeleteNetwork(ctx context.Context, req *iampb.NetworkRequest) (*pbtypes.Empty, error) {
	if s.DeleteNetworkHandler == nil {
		panic("DeleteNetwork called but not set")
	}
	return s.DeleteNetworkHandler(ctx, req)
}

func (s *PBIAM) ListTenants(ctx context.Context, req *iampb.ListTenantsRequest) (*iampb.ListTenantsResponse, error) {
	if s.ListTenantsHandler == nil {
		panic("ListTenants called but not set")
	}
	return s.ListTenantsHandler(ctx, req)
}

func (s *PBIAM) CreateTenant(ctx context.Context, req *iampb.CreateTenantRequest) (*iampb.CreateTenantResponse, error) {
	if s.CreateTenantHandler == nil {
		panic("CreateTenant called but not set")
	}
	return s.CreateTenantHandler(ctx, req)
}

func (s *PBIAM) GetTenant(ctx context.Context, req *iampb.TenantRequest) (*iampb.GetTenantResponse, error) {
	if s.GetTenantHandler == nil {
		panic("GetTenant called but not set")
	}
	return s.GetTenantHandler(ctx, req)
}

func (s *PBIAM) UpdateTenant(ctx context.Context, req *iampb.UpdateTenantRequest) (*pbtypes.Empty, error) {
	if s.UpdateTenantHandler == nil {
		panic("UpdateTenant called but not set")
	}
	return s.UpdateTenantHandler(ctx, req)
}

func (s *PBIAM) DeleteTenant(ctx context.Context, req *iampb.TenantRequest) (*pbtypes.Empty, error) {
	if s.DeleteTenantHandler == nil {
		panic("DeleteTenant called but not set")
	}
	return s.DeleteTenantHandler(ctx, req)
}

func (s *PBIAM) ListHomeNetworks(ctx context.Context, req *iampbv2.ListHomeNetworksRequest) (*iampbv2.ListHomeNetworksResponse, error) {
	if s.ListHomeNetworksHandler == nil {
		panic("ListHomeNetworks called but not set")
	}
	return s.ListHomeNetworksHandler(ctx, req)
}
