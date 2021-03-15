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

package packetbrokeragent

import (
	"context"

	"github.com/gogo/protobuf/types"
	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type pbaServer struct {
}

func (s *pbaServer) GetRegistration(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.PacketBrokerRegistration, error) {
	if err := rights.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (s *pbaServer) Register(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.PacketBrokerRegistration, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (s *pbaServer) Deregister(ctx context.Context, _ *pbtypes.Empty) (*types.Empty, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (s *pbaServer) GetForwarderDefaultRoutingPolicy(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.PacketBrokerDefaultRoutingPolicy, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (s *pbaServer) SetForwarderDefaultRoutingPolicy(ctx context.Context, req *ttnpb.SetPacketBrokerDefaultRoutingPolicyRequest) (*pbtypes.Empty, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (s *pbaServer) DeleteForwarderDefaultRoutingPolicy(ctx context.Context, _ *pbtypes.Empty) (*pbtypes.Empty, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (s *pbaServer) ListForwarderRoutingPolicies(ctx context.Context, req *ttnpb.ListForwarderRoutingPoliciesRequest) (*ttnpb.PacketBrokerRoutingPolicies, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (s *pbaServer) GetForwarderRoutingPolicy(ctx context.Context, req *ttnpb.PacketBrokerNetworkIdentifier) (*ttnpb.PacketBrokerRoutingPolicy, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (s *pbaServer) SetForwarderRoutingPolicy(ctx context.Context, req *ttnpb.SetPacketBrokerRoutingPolicyRequest) (*ttnpb.PacketBrokerRoutingPolicy, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (s *pbaServer) DeleteForwarderRoutingPolicy(ctx context.Context, req *ttnpb.PacketBrokerNetworkIdentifier) (*pbtypes.Empty, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (s *pbaServer) ListHomeNetworks(ctx context.Context, req *ttnpb.ListHomeNetworksRequest) (*ttnpb.PacketBrokerNetworks, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

func (s *pbaServer) ListHomeNetworkRoutingPolicies(ctx context.Context, req *ttnpb.ListHomeNetworksRoutingPoliciesRequest) (*ttnpb.PacketBrokerRoutingPolicies, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}
