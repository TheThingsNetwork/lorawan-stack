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

	routingpb "go.packetbroker.org/api/routing"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// PBControlPlane is a mock Packet Broker Control Plane.
type PBControlPlane struct {
	routingpb.UnimplementedPolicyManagerServer

	*grpc.Server
	ListDefaultPoliciesHandler     func(ctx context.Context, req *routingpb.ListDefaultPoliciesRequest) (*routingpb.ListDefaultPoliciesResponse, error)
	GetDefaultPolicyHandler        func(ctx context.Context, req *routingpb.GetDefaultPolicyRequest) (*routingpb.GetPolicyResponse, error)
	SetDefaultPolicyHandler        func(ctx context.Context, req *routingpb.SetPolicyRequest) (*emptypb.Empty, error)
	ListHomeNetworkPoliciesHandler func(ctx context.Context, req *routingpb.ListHomeNetworkPoliciesRequest) (*routingpb.ListHomeNetworkPoliciesResponse, error)
	GetHomeNetworkPolicyHandler    func(ctx context.Context, req *routingpb.GetHomeNetworkPolicyRequest) (*routingpb.GetPolicyResponse, error)
	SetHomeNetworkPolicyHandler    func(ctx context.Context, req *routingpb.SetPolicyRequest) (*emptypb.Empty, error)
	ListEffectivePoliciesHandler   func(ctx context.Context, req *routingpb.ListEffectivePoliciesRequest) (*routingpb.ListEffectivePoliciesResponse, error)
	ListNetworksWithPolicyHandler  func(ctx context.Context, req *routingpb.ListNetworksWithPolicyRequest) (*routingpb.ListNetworksResponse, error)
}

// NewPBControlPlane instantiates a new mock Packet Broker Control Plane.
func NewPBControlPlane(tb testing.TB) *PBControlPlane {
	cp := &PBControlPlane{
		Server: grpc.NewServer(
			grpc.UnaryInterceptor(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
				ctx = test.ContextWithTB(ctx, tb)
				return handler(ctx, req)
			}),
		),
	}
	routingpb.RegisterPolicyManagerServer(cp.Server, cp)
	return cp
}

func (s *PBControlPlane) ListDefaultPolicies(ctx context.Context, req *routingpb.ListDefaultPoliciesRequest) (*routingpb.ListDefaultPoliciesResponse, error) {
	if s.ListDefaultPoliciesHandler == nil {
		panic("ListDefaultPolicies called but not set")
	}
	return s.ListDefaultPoliciesHandler(ctx, req)
}

func (s *PBControlPlane) GetDefaultPolicy(ctx context.Context, req *routingpb.GetDefaultPolicyRequest) (*routingpb.GetPolicyResponse, error) {
	if s.GetDefaultPolicyHandler == nil {
		panic("GetDefaultPolicy called but not set")
	}
	return s.GetDefaultPolicyHandler(ctx, req)
}

func (s *PBControlPlane) SetDefaultPolicy(ctx context.Context, req *routingpb.SetPolicyRequest) (*emptypb.Empty, error) {
	if s.SetDefaultPolicyHandler == nil {
		panic("SetDefaultPolicy called but not set")
	}
	return s.SetDefaultPolicyHandler(ctx, req)
}

func (s *PBControlPlane) ListHomeNetworkPolicies(ctx context.Context, req *routingpb.ListHomeNetworkPoliciesRequest) (*routingpb.ListHomeNetworkPoliciesResponse, error) {
	if s.ListHomeNetworkPoliciesHandler == nil {
		panic("ListHomeNetworkPolicies called but not set")
	}
	return s.ListHomeNetworkPoliciesHandler(ctx, req)
}

func (s *PBControlPlane) GetHomeNetworkPolicy(ctx context.Context, req *routingpb.GetHomeNetworkPolicyRequest) (*routingpb.GetPolicyResponse, error) {
	if s.GetHomeNetworkPolicyHandler == nil {
		panic("GetHomeNetworkPolicy called but not set")
	}
	return s.GetHomeNetworkPolicyHandler(ctx, req)
}

func (s *PBControlPlane) SetHomeNetworkPolicy(ctx context.Context, req *routingpb.SetPolicyRequest) (*emptypb.Empty, error) {
	if s.SetHomeNetworkPolicyHandler == nil {
		panic("SetHomeNetworkPolicy called but not set")
	}
	return s.SetHomeNetworkPolicyHandler(ctx, req)
}

func (s *PBControlPlane) ListEffectivePolicies(ctx context.Context, req *routingpb.ListEffectivePoliciesRequest) (*routingpb.ListEffectivePoliciesResponse, error) {
	if s.ListEffectivePoliciesHandler == nil {
		panic("ListEffectivePolicies called but not set")
	}
	return s.ListEffectivePoliciesHandler(ctx, req)
}

func (s *PBControlPlane) ListNetworksWithPolicy(ctx context.Context, req *routingpb.ListNetworksWithPolicyRequest) (*routingpb.ListNetworksResponse, error) {
	if s.ListNetworksWithPolicyHandler == nil {
		panic("ListNetworksWithPolicy called but not set")
	}
	return s.ListNetworksWithPolicyHandler(ctx, req)
}
