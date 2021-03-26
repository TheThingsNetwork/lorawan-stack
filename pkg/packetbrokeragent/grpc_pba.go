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
	"strconv"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	iampb "go.packetbroker.org/api/iam"
	iampbv2 "go.packetbroker.org/api/iam/v2"
	routingpb "go.packetbroker.org/api/routing"
	packetbroker "go.packetbroker.org/api/v3"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const listPageSize = 100

type pbaServer struct {
	*Agent
	iamConn,
	cpConn *grpc.ClientConn
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
				TenantID: tenantID,
			},
			Name:          registration.GetName(),
			DevAddrBlocks: asDevAddrBlocks(registration.GetDevAddrBlocks()),
			ContactInfo:   asContactInfo(registration.GetAdministrativeContact(), registration.GetTechnicalContact()),
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
	devAddrBlocks := toDevAddrBlocks(registration.DevAddrBlocks)
	adminContact, technicalContact := toContactInfo(registration.ContactInfo)

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
			TenantID: tenantID,
		},
		Name:          registration.Name,
		DevAddrBlocks: asDevAddrBlocks(devAddrBlocks),
		ContactInfo:   asContactInfo(adminContact, technicalContact),
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

func (s *pbaServer) GetHomeNetworkDefaultRoutingPolicy(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.PacketBrokerDefaultRoutingPolicy, error) {
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
	return asDefaultRoutingPolicy(res.GetPolicy()), nil
}

func (s *pbaServer) SetHomeNetworkDefaultRoutingPolicy(ctx context.Context, req *ttnpb.SetPacketBrokerDefaultRoutingPolicyRequest) (*pbtypes.Empty, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}

	_, err := routingpb.NewPolicyManagerClient(s.cpConn).SetDefaultPolicy(ctx, &routingpb.SetPolicyRequest{
		Policy: &packetbroker.RoutingPolicy{
			ForwarderNetId:    s.netID.MarshalNumber(),
			ForwarderTenantId: s.tenantIDExtractor(ctx),
			Uplink:            toUplinkRoutingPolicy(req.GetUplink()),
			Downlink:          toDownlinkRoutingPolicy(req.GetDownlink()),
		},
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

func (s *pbaServer) DeleteHomeNetworkDefaultRoutingPolicy(ctx context.Context, _ *pbtypes.Empty) (*pbtypes.Empty, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}

	_, err := routingpb.NewPolicyManagerClient(s.cpConn).SetDefaultPolicy(ctx, &routingpb.SetPolicyRequest{
		Policy: &packetbroker.RoutingPolicy{
			ForwarderNetId:    s.netID.MarshalNumber(),
			ForwarderTenantId: s.tenantIDExtractor(ctx),
		},
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

func (s *pbaServer) ListHomeNetworkRoutingPolicies(ctx context.Context, req *ttnpb.ListHomeNetworkRoutingPoliciesRequest) (*ttnpb.PacketBrokerRoutingPolicies, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}

	var (
		limit        = int(req.GetLimit())
		page         = int(req.GetPage())
		client       = routingpb.NewPolicyManagerClient(s.cpConn)
		netID        = s.netID.MarshalNumber()
		tenantID     = s.tenantIDExtractor(ctx)
		updatedSince *time.Time
		policies     []*packetbroker.RoutingPolicy
		total        int64
	)
	if limit == 0 || limit > listPageSize {
		limit = listPageSize
	}
	if page == 0 {
		page = 1
	}
	end := page * limit
	for len(policies) < end {
		req := &routingpb.ListHomeNetworkPoliciesRequest{
			ForwarderNetId:    netID,
			ForwarderTenantId: tenantID,
		}
		if updatedSince != nil {
			req.UpdatedSince, _ = pbtypes.TimestampProto(*updatedSince)
		}
		res, err := client.ListHomeNetworkPolicies(ctx, req)
		if err != nil {
			return nil, err
		}
		if len(res.Policies) == 0 {
			break
		}
		policies = append(policies, res.GetPolicies()...)
		if t, err := pbtypes.TimestampFromProto(res.Policies[len(res.Policies)-1].GetUpdatedAt()); err == nil {
			updatedSince = &t
		} else {
			return nil, err
		}
		total = int64(res.Total)
	}

	var (
		offset = (page - 1) * limit
		slice  []*packetbroker.RoutingPolicy
	)
	if len(policies) > offset {
		slice = policies[offset:]
		if len(policies) > end {
			slice = slice[:end]
		}
	}
	res := &ttnpb.PacketBrokerRoutingPolicies{
		Policies: make([]*ttnpb.PacketBrokerRoutingPolicy, len(slice)),
	}
	for i, p := range slice {
		res.Policies[i] = asRoutingPolicy(p)
	}
	grpc.SetHeader(ctx, metadata.Pairs("x-total-count", strconv.FormatInt(total, 10)))
	return res, nil
}

func (s *pbaServer) GetHomeNetworkRoutingPolicy(ctx context.Context, req *ttnpb.PacketBrokerNetworkIdentifier) (*ttnpb.PacketBrokerRoutingPolicy, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}

	res, err := routingpb.NewPolicyManagerClient(s.cpConn).GetHomeNetworkPolicy(ctx, &routingpb.GetHomeNetworkPolicyRequest{
		ForwarderNetId:      s.netID.MarshalNumber(),
		ForwarderTenantId:   s.tenantIDExtractor(ctx),
		HomeNetworkNetId:    req.GetNetID(),
		HomeNetworkTenantId: req.GetTenantID(),
	})
	if err != nil {
		return nil, err
	}
	return asRoutingPolicy(res.GetPolicy()), nil
}

func (s *pbaServer) SetHomeNetworkRoutingPolicy(ctx context.Context, req *ttnpb.SetPacketBrokerRoutingPolicyRequest) (*pbtypes.Empty, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}

	_, err := routingpb.NewPolicyManagerClient(s.cpConn).SetHomeNetworkPolicy(ctx, &routingpb.SetPolicyRequest{
		Policy: &packetbroker.RoutingPolicy{
			ForwarderNetId:      s.netID.MarshalNumber(),
			ForwarderTenantId:   s.tenantIDExtractor(ctx),
			HomeNetworkNetId:    req.GetHomeNetworkId().GetNetID(),
			HomeNetworkTenantId: req.GetHomeNetworkId().GetTenantID(),
			Uplink:              toUplinkRoutingPolicy(req.GetUplink()),
			Downlink:            toDownlinkRoutingPolicy(req.GetDownlink()),
		},
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

func (s *pbaServer) DeleteHomeNetworkRoutingPolicy(ctx context.Context, req *ttnpb.PacketBrokerNetworkIdentifier) (*pbtypes.Empty, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}

	_, err := routingpb.NewPolicyManagerClient(s.cpConn).SetHomeNetworkPolicy(ctx, &routingpb.SetPolicyRequest{
		Policy: &packetbroker.RoutingPolicy{
			ForwarderNetId:      s.netID.MarshalNumber(),
			ForwarderTenantId:   s.tenantIDExtractor(ctx),
			HomeNetworkNetId:    req.GetNetID(),
			HomeNetworkTenantId: req.GetTenantID(),
		},
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

func (s *pbaServer) ListHomeNetworks(ctx context.Context, req *ttnpb.ListHomeNetworksRequest) (*ttnpb.PacketBrokerNetworks, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}

	page := req.Page
	if page == 0 {
		page = 1
	}
	networks, err := iampbv2.NewCatalogClient(s.iamConn).ListHomeNetworks(ctx, &iampbv2.ListHomeNetworksRequest{
		Offset: (page - 1) * req.Limit,
		Limit:  req.Limit,
	})
	if err != nil {
		return nil, err
	}
	res := &ttnpb.PacketBrokerNetworks{
		Networks: make([]*ttnpb.PacketBrokerNetwork, 0, len(networks.GetHomeNetworks())),
	}
	for _, n := range networks.GetHomeNetworks() {
		var (
			id      *ttnpb.PacketBrokerNetworkIdentifier
			network interface {
				GetName() string
				GetDevAddrBlocks() []*packetbroker.DevAddrBlock
				GetAdministrativeContact() *packetbroker.ContactInfo
				GetTechnicalContact() *packetbroker.ContactInfo
			}
		)
		switch member := n.GetValue().(type) {
		case *packetbroker.NetworkOrTenant_Network:
			id = &ttnpb.PacketBrokerNetworkIdentifier{
				NetID: member.Network.GetNetId(),
			}
			network = member.Network
		case *packetbroker.NetworkOrTenant_Tenant:
			id = &ttnpb.PacketBrokerNetworkIdentifier{
				NetID:    member.Tenant.GetNetId(),
				TenantID: member.Tenant.GetTenantId(),
			}
			network = member.Tenant
		}
		res.Networks = append(res.Networks, &ttnpb.PacketBrokerNetwork{
			Id:            id,
			Name:          network.GetName(),
			DevAddrBlocks: asDevAddrBlocks(network.GetDevAddrBlocks()),
			ContactInfo:   asContactInfo(network.GetAdministrativeContact(), network.GetTechnicalContact()),
		})
	}
	grpc.SetHeader(ctx, metadata.Pairs("x-total-count", strconv.FormatInt(int64(networks.GetTotal()), 10)))
	return res, nil
}

func (s *pbaServer) ListForwarderRoutingPolicies(ctx context.Context, req *ttnpb.ListForwarderRoutingPoliciesRequest) (*ttnpb.PacketBrokerRoutingPolicies, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}

	page := req.Page
	if page == 0 {
		page = 1
	}
	policies, err := routingpb.NewPolicyManagerClient(s.cpConn).ListEffectivePolicies(ctx, &routingpb.ListEffectivePoliciesRequest{
		HomeNetworkNetId:    s.netID.MarshalNumber(),
		HomeNetworkTenantId: s.tenantIDExtractor(ctx),
		Offset:              (page - 1) * req.Limit,
		Limit:               req.Limit,
	})
	if err != nil {
		return nil, err
	}
	res := &ttnpb.PacketBrokerRoutingPolicies{
		Policies: make([]*ttnpb.PacketBrokerRoutingPolicy, len(policies.GetPolicies())),
	}
	for i, p := range policies.GetPolicies() {
		res.Policies[i] = asRoutingPolicy(p)
	}
	grpc.SetHeader(ctx, metadata.Pairs("x-total-count", strconv.FormatInt(int64(policies.GetTotal()), 10)))
	return res, nil
}
