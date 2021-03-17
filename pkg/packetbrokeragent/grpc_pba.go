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

	pbtypes "github.com/gogo/protobuf/types"
	iampb "go.packetbroker.org/api/iam"
	routingpb "go.packetbroker.org/api/routing"
	packetbroker "go.packetbroker.org/api/v3"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type pbaServer struct {
	*Agent
	iamConn *grpc.ClientConn
}

func (s *pbaServer) GetInfo(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.PacketBrokerInfo, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}

	var (
		tenantID     = s.tenantIDExtractor(ctx)
		registration interface {
			GetName() string
			GetDevAddrBlocks() []*packetbroker.DevAddrBlock
			GetAdministrativeContact() *packetbroker.ContactInfo
			GetTechnicalContact() *packetbroker.ContactInfo
		}
		err error
	)
	if tenantID == "" {
		var res *iampb.GetNetworkResponse
		res, err = iampb.NewNetworkRegistryClient(s.iamConn).GetNetwork(ctx, &iampb.NetworkRequest{
			NetId: s.netID.MarshalNumber(),
		})
		registration = res.GetNetwork()
	} else {
		var res *iampb.GetTenantResponse
		res, err = iampb.NewTenantRegistryClient(s.iamConn).GetTenant(ctx, &iampb.TenantRequest{
			NetId:    s.netID.MarshalNumber(),
			TenantId: tenantID,
		})
		registration = res.GetTenant()
	}
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}
		registration = nil
	}

	res := &ttnpb.PacketBrokerInfo{
		ForwarderEnabled:   s.forwarderConfig.Enable,
		HomeNetworkEnabled: s.homeNetworkConfig.Enable,
	}
	if registration != nil {
		res.Registration = &ttnpb.PacketBrokerNetwork{
			Id: &ttnpb.PacketBrokerNetworkIdentifier{
				NetID:    s.netID.MarshalNumber(),
				TenantId: tenantID,
			},
			Name:          registration.GetName(),
			DevAddrBlocks: s.asDevAddrBlocks(registration.GetDevAddrBlocks()),
			ContactInfo:   s.asContactInfo(registration.GetAdministrativeContact(), registration.GetTechnicalContact()),
		}
	}

	return res, nil
}

var (
	errNetwork      = errors.DefineFailedPrecondition("network", "not supported for network")
	errRegistration = errors.Define("registration", "get registration information")
)

func (s *pbaServer) Register(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.PacketBrokerNetwork, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	tenantID := s.tenantIDExtractor(ctx)
	if tenantID == "" {
		return nil, errNetwork.New()
	}

	_, err := iampb.NewTenantRegistryClient(s.iamConn).GetTenant(ctx, &iampb.TenantRequest{
		NetId:    s.netID.MarshalNumber(),
		TenantId: tenantID,
	})
	var create bool
	if err != nil {
		if errors.IsNotFound(err) {
			create = true
		} else {
			return nil, err
		}
	}

	registration, err := s.registrationInfoExtractor(ctx)
	if err != nil {
		return nil, errRegistration.WithCause(err)
	}
	devAddrBlocks := s.toDevAddrBlocks(registration.DevAddrBlocks)
	adminContact, technicalContact := s.toContactInfo(registration.ContactInfo)

	if create {
		_, err = iampb.NewTenantRegistryClient(s.iamConn).CreateTenant(ctx, &iampb.CreateTenantRequest{
			Tenant: &packetbroker.Tenant{
				NetId:                 s.netID.MarshalNumber(),
				TenantId:              tenantID,
				Name:                  registration.Name,
				DevAddrBlocks:         devAddrBlocks,
				AdministrativeContact: adminContact,
				TechnicalContact:      technicalContact,
			},
		})
	} else {
		_, err = iampb.NewTenantRegistryClient(s.iamConn).UpdateTenant(ctx, &iampb.UpdateTenantRequest{
			NetId:    s.netID.MarshalNumber(),
			TenantId: tenantID,
			Name: &pbtypes.StringValue{
				Value: registration.Name,
			},
			DevAddrBlocks: &iampb.DevAddrBlocksValue{
				Value: devAddrBlocks,
			},
			AdministrativeContact: &iampb.ContactInfoValue{
				Value: adminContact,
			},
			TechnicalContact: &iampb.ContactInfoValue{
				Value: technicalContact,
			},
		})
	}

	if err != nil {
		return nil, err
	}

	return &ttnpb.PacketBrokerNetwork{
		Id: &ttnpb.PacketBrokerNetworkIdentifier{
			NetID:    s.netID.MarshalNumber(),
			TenantId: tenantID,
		},
		Name:          registration.Name,
		DevAddrBlocks: s.asDevAddrBlocks(devAddrBlocks),
		ContactInfo:   s.asContactInfo(adminContact, technicalContact),
	}, nil
}

func (s *pbaServer) Deregister(ctx context.Context, _ *pbtypes.Empty) (*pbtypes.Empty, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	tenantID := s.tenantIDExtractor(ctx)
	if tenantID == "" {
		return nil, errNetwork.New()
	}

	_, err := iampb.NewTenantRegistryClient(s.iamConn).DeleteTenant(ctx, &iampb.TenantRequest{
		NetId:    s.netID.MarshalNumber(),
		TenantId: tenantID,
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

func (s *pbaServer) GetForwarderDefaultRoutingPolicy(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.PacketBrokerDefaultRoutingPolicy, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}

	res, err := routingpb.NewPolicyManagerClient(s.cpConn).GetDefaultPolicy(ctx, &routingpb.GetDefaultPolicyRequest{
		ForwarderNetId:    s.netID.MarshalNumber(),
		ForwarderTenantId: s.tenantIDExtractor(ctx),
	})
	if err != nil {
		return nil, err
	}

	_ = res

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
