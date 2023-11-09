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
	"fmt"
	"strconv"

	iampb "go.packetbroker.org/api/iam"
	iampbv2 "go.packetbroker.org/api/iam/v2"
	mappingpb "go.packetbroker.org/api/mapping/v2"
	routingpb "go.packetbroker.org/api/routing"
	packetbroker "go.packetbroker.org/api/v3"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/warning"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

const listPageSize = 100

type pbaServer struct {
	ttnpb.UnimplementedPbaServer

	*Agent
	iamConn,
	cpConn *grpc.ClientConn
}

func (s *pbaServer) GetInfo(ctx context.Context, _ *emptypb.Empty) (*ttnpb.PacketBrokerInfo, error) {
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
			GetListed() bool
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

	// Register and deregister is only available if Packet Broker Agent is configured with NetID level authorization, and
	// if the registration is a tenant within that NetID.
	id, err := s.authenticator.AuthInfo(ctx)
	registerEnabled := err == nil && id.TenantId == "" && tenantID != ""

	res := &ttnpb.PacketBrokerInfo{
		ForwarderEnabled:   s.forwarderConfig.Enable,
		HomeNetworkEnabled: s.homeNetworkConfig.Enable,
		RegisterEnabled:    registerEnabled,
	}
	if registration != nil {
		res.Registration = &ttnpb.PacketBrokerNetwork{
			Id: &ttnpb.PacketBrokerNetworkIdentifier{
				NetId:    s.netID.MarshalNumber(),
				TenantId: tenantID,
			},
			Name:          registration.GetName(),
			DevAddrBlocks: fromPBDevAddrBlocks(registration.GetDevAddrBlocks()),
			ContactInfo:   fromPBContactInfo(registration.GetAdministrativeContact(), registration.GetTechnicalContact()),
			Listed:        registration.GetListed(),
		}
	}

	return res, nil
}

func addAuthorityWarning(ctx context.Context, authority string) {
	var name string
	switch authority {
	case "ttnr": // The Things Network Registry is the authority used by The Things Industries.
		name = "The Things Industries"
	default:
		name = "an external authority"
	}
	warning.Add(ctx, fmt.Sprintf(
		"The Packet Broker tenant is managed by %s: not all settings can be updated through The Things Stack.", name,
	))
}

var (
	errNetwork      = errors.DefineFailedPrecondition("network", "not supported for network")
	errRegistration = errors.Define("registration", "get registration information")
)

func (s *pbaServer) Register(ctx context.Context, req *ttnpb.PacketBrokerRegisterRequest) (*ttnpb.PacketBrokerNetwork, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	tenantID := s.tenantIDExtractor(ctx)
	if tenantID == "" {
		return nil, errNetwork.New()
	}

	res, err := iampb.NewTenantRegistryClient(s.iamConn).GetTenant(ctx, &iampb.TenantRequest{
		NetId:    s.netID.MarshalNumber(),
		TenantId: tenantID,
	})
	var create,
		managedByAuthority bool
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}
		create = true
	} else if managedByAuthority = res.Tenant.Authority != ""; managedByAuthority {
		addAuthorityWarning(ctx, res.Tenant.Authority)
	}

	registration, err := s.registrationInfoExtractor(ctx, s.homeNetworkClusterID, s.clusterIDBuilder)
	if err != nil {
		return nil, errRegistration.WithCause(err)
	}
	listed := req.Listed != nil && req.Listed.Value || req.Listed == nil && registration.Listed
	devAddrBlocks := toPBDevAddrBlocks(registration.DevAddrBlocks)
	adminContact, technicalContact := toPBContactInfo(registration.ContactInfo)

	if create {
		var res *iampb.CreateTenantResponse
		res, err = iampb.NewTenantRegistryClient(s.iamConn).CreateTenant(ctx, &iampb.CreateTenantRequest{
			Tenant: &packetbroker.Tenant{
				NetId:                 s.netID.MarshalNumber(),
				TenantId:              tenantID,
				Name:                  registration.Name,
				DevAddrBlocks:         devAddrBlocks,
				AdministrativeContact: adminContact,
				TechnicalContact:      technicalContact,
				Listed:                listed,
			},
		})
		if err == nil && res.Tenant.Authority != "" {
			addAuthorityWarning(ctx, res.Tenant.Authority)
		}
	} else {
		req := &iampb.UpdateTenantRequest{
			NetId:    s.netID.MarshalNumber(),
			TenantId: tenantID,
			Listed:   wrapperspb.Bool(listed),
		}
		// If the Packet Broker tenant is not managed by an authority, other fields can be updated.
		if !managedByAuthority {
			req.Name = wrapperspb.String(registration.Name)
			req.AdministrativeContact = &packetbroker.ContactInfoValue{
				Value: adminContact,
			}
			req.TechnicalContact = &packetbroker.ContactInfoValue{
				Value: technicalContact,
			}
			// Managing DevAddr blocks is only available if Packet Broker Agent is configured with NetID level authorization,
			// and if the registration is a tenant within that NetID.
			if id, err := s.authenticator.AuthInfo(ctx); err == nil && id.TenantId == "" {
				req.DevAddrBlocks = &iampb.DevAddrBlocksValue{
					Value: devAddrBlocks,
				}
			}
		}
		_, err = iampb.NewTenantRegistryClient(s.iamConn).UpdateTenant(ctx, req)
	}

	if err != nil {
		return nil, err
	}

	return &ttnpb.PacketBrokerNetwork{
		Id: &ttnpb.PacketBrokerNetworkIdentifier{
			NetId:    s.netID.MarshalNumber(),
			TenantId: tenantID,
		},
		Name:          registration.Name,
		DevAddrBlocks: fromPBDevAddrBlocks(devAddrBlocks),
		ContactInfo:   fromPBContactInfo(adminContact, technicalContact),
		Listed:        listed,
	}, nil
}

func (s *pbaServer) Deregister(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
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

func (s *pbaServer) GetHomeNetworkDefaultRoutingPolicy(ctx context.Context, _ *emptypb.Empty) (*ttnpb.PacketBrokerDefaultRoutingPolicy, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	if !s.forwarderConfig.Enable {
		return nil, errNotEnabled.New()
	}

	res, err := routingpb.NewPolicyManagerClient(s.cpConn).GetDefaultPolicy(ctx, &routingpb.GetDefaultPolicyRequest{
		ForwarderNetId:    s.netID.MarshalNumber(),
		ForwarderTenantId: s.tenantIDExtractor(ctx),
	})
	if err != nil {
		return nil, err
	}
	return fromPBDefaultRoutingPolicy(res.GetPolicy()), nil
}

func (s *pbaServer) SetHomeNetworkDefaultRoutingPolicy(ctx context.Context, req *ttnpb.SetPacketBrokerDefaultRoutingPolicyRequest) (*emptypb.Empty, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	if !s.forwarderConfig.Enable {
		return nil, errNotEnabled.New()
	}

	_, err := routingpb.NewPolicyManagerClient(s.cpConn).SetDefaultPolicy(ctx, &routingpb.SetPolicyRequest{
		Policy: &packetbroker.RoutingPolicy{
			ForwarderNetId:    s.netID.MarshalNumber(),
			ForwarderTenantId: s.tenantIDExtractor(ctx),
			Uplink:            toPBUplinkRoutingPolicy(req.GetUplink()),
			Downlink:          toPBDownlinkRoutingPolicy(req.GetDownlink()),
		},
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

func (s *pbaServer) DeleteHomeNetworkDefaultRoutingPolicy(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	if !s.forwarderConfig.Enable {
		return nil, errNotEnabled.New()
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
	if !s.forwarderConfig.Enable {
		return nil, errNotEnabled.New()
	}

	var (
		limit        = int(req.GetLimit())
		page         = int(req.GetPage())
		client       = routingpb.NewPolicyManagerClient(s.cpConn)
		netID        = s.netID.MarshalNumber()
		tenantID     = s.tenantIDExtractor(ctx)
		updatedSince *timestamppb.Timestamp
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
			req.UpdatedSince = updatedSince
		}
		res, err := client.ListHomeNetworkPolicies(ctx, req)
		if err != nil {
			return nil, err
		}
		if len(res.Policies) == 0 {
			break
		}
		policies = append(policies, res.GetPolicies()...)
		updatedSince = res.Policies[len(res.Policies)-1].GetUpdatedAt()
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
		res.Policies[i] = fromPBRoutingPolicy(p)
	}
	grpc.SetHeader(ctx, metadata.Pairs("x-total-count", strconv.FormatInt(total, 10)))
	return res, nil
}

func (s *pbaServer) GetHomeNetworkRoutingPolicy(ctx context.Context, req *ttnpb.PacketBrokerNetworkIdentifier) (*ttnpb.PacketBrokerRoutingPolicy, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	if !s.forwarderConfig.Enable {
		return nil, errNotEnabled.New()
	}

	res, err := routingpb.NewPolicyManagerClient(s.cpConn).GetHomeNetworkPolicy(ctx, &routingpb.GetHomeNetworkPolicyRequest{
		ForwarderNetId:      s.netID.MarshalNumber(),
		ForwarderTenantId:   s.tenantIDExtractor(ctx),
		HomeNetworkNetId:    req.GetNetId(),
		HomeNetworkTenantId: req.GetTenantId(),
	})
	if err != nil {
		return nil, err
	}
	return fromPBRoutingPolicy(res.GetPolicy()), nil
}

func (s *pbaServer) SetHomeNetworkRoutingPolicy(ctx context.Context, req *ttnpb.SetPacketBrokerRoutingPolicyRequest) (*emptypb.Empty, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	if !s.forwarderConfig.Enable {
		return nil, errNotEnabled.New()
	}

	_, err := routingpb.NewPolicyManagerClient(s.cpConn).SetHomeNetworkPolicy(ctx, &routingpb.SetPolicyRequest{
		Policy: &packetbroker.RoutingPolicy{
			ForwarderNetId:      s.netID.MarshalNumber(),
			ForwarderTenantId:   s.tenantIDExtractor(ctx),
			HomeNetworkNetId:    req.GetHomeNetworkId().GetNetId(),
			HomeNetworkTenantId: req.GetHomeNetworkId().GetTenantId(),
			Uplink:              toPBUplinkRoutingPolicy(req.GetUplink()),
			Downlink:            toPBDownlinkRoutingPolicy(req.GetDownlink()),
		},
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

func (s *pbaServer) DeleteHomeNetworkRoutingPolicy(ctx context.Context, req *ttnpb.PacketBrokerNetworkIdentifier) (*emptypb.Empty, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	if !s.forwarderConfig.Enable {
		return nil, errNotEnabled.New()
	}

	_, err := routingpb.NewPolicyManagerClient(s.cpConn).SetHomeNetworkPolicy(ctx, &routingpb.SetPolicyRequest{
		Policy: &packetbroker.RoutingPolicy{
			ForwarderNetId:      s.netID.MarshalNumber(),
			ForwarderTenantId:   s.tenantIDExtractor(ctx),
			HomeNetworkNetId:    req.GetNetId(),
			HomeNetworkTenantId: req.GetTenantId(),
		},
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

func (s *pbaServer) GetHomeNetworkDefaultGatewayVisibility(ctx context.Context, req *emptypb.Empty) (*ttnpb.PacketBrokerDefaultGatewayVisibility, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	if !s.forwarderConfig.Enable {
		return nil, errNotEnabled.New()
	}

	res, err := mappingpb.NewGatewayVisibilityManagerClient(s.cpConn).GetDefaultVisibility(ctx, &mappingpb.GetDefaultGatewayVisibilityRequest{
		ForwarderNetId:    s.netID.MarshalNumber(),
		ForwarderTenantId: s.tenantIDExtractor(ctx),
	})
	if err != nil {
		return nil, err
	}
	return fromPBDefaultGatewayVisibility(res.GetVisibility()), nil
}

func (s *pbaServer) SetHomeNetworkDefaultGatewayVisibility(ctx context.Context, req *ttnpb.SetPacketBrokerDefaultGatewayVisibilityRequest) (*emptypb.Empty, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	if !s.forwarderConfig.Enable {
		return nil, errNotEnabled.New()
	}

	_, err := mappingpb.NewGatewayVisibilityManagerClient(s.cpConn).SetDefaultVisibility(ctx, &mappingpb.SetGatewayVisibilityRequest{
		Visibility: &packetbroker.GatewayVisibility{
			ForwarderNetId:    s.netID.MarshalNumber(),
			ForwarderTenantId: s.tenantIDExtractor(ctx),
			Location:          req.Visibility.Location,
			AntennaPlacement:  req.Visibility.AntennaPlacement,
			AntennaCount:      req.Visibility.AntennaCount,
			FineTimestamps:    req.Visibility.FineTimestamps,
			ContactInfo:       req.Visibility.ContactInfo,
			Status:            req.Visibility.Status,
			FrequencyPlan:     req.Visibility.FrequencyPlan,
			PacketRates:       req.Visibility.PacketRates,
		},
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

func (s *pbaServer) DeleteHomeNetworkDefaultGatewayVisibility(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	if !s.forwarderConfig.Enable {
		return nil, errNotEnabled.New()
	}

	_, err := mappingpb.NewGatewayVisibilityManagerClient(s.cpConn).SetDefaultVisibility(ctx, &mappingpb.SetGatewayVisibilityRequest{
		Visibility: &packetbroker.GatewayVisibility{
			ForwarderNetId:    s.netID.MarshalNumber(),
			ForwarderTenantId: s.tenantIDExtractor(ctx),
		},
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

func (s *pbaServer) listNetworks(ctx context.Context, req func() ([]*packetbroker.NetworkOrTenant, uint32, error)) (*ttnpb.PacketBrokerNetworks, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}

	networks, total, err := req()
	if err != nil {
		return nil, err
	}
	res := &ttnpb.PacketBrokerNetworks{
		Networks: make([]*ttnpb.PacketBrokerNetwork, 0, len(networks)),
	}
	for _, n := range networks {
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
				NetId: member.Network.GetNetId(),
			}
			network = member.Network
		case *packetbroker.NetworkOrTenant_Tenant:
			id = &ttnpb.PacketBrokerNetworkIdentifier{
				NetId:    member.Tenant.GetNetId(),
				TenantId: member.Tenant.GetTenantId(),
			}
			network = member.Tenant
		}
		res.Networks = append(res.Networks, &ttnpb.PacketBrokerNetwork{
			Id:            id,
			Name:          network.GetName(),
			DevAddrBlocks: fromPBDevAddrBlocks(network.GetDevAddrBlocks()),
			ContactInfo:   fromPBContactInfo(network.GetAdministrativeContact(), network.GetTechnicalContact()),
		})
	}
	grpc.SetHeader(ctx, metadata.Pairs("x-total-count", strconv.FormatInt(int64(total), 10)))
	return res, nil
}

func (s *pbaServer) ListNetworks(ctx context.Context, req *ttnpb.ListPacketBrokerNetworksRequest) (*ttnpb.PacketBrokerNetworks, error) {
	page := req.Page
	if page == 0 {
		page = 1
	}
	return s.listNetworks(ctx, func() ([]*packetbroker.NetworkOrTenant, uint32, error) {
		if req.WithRoutingPolicy {
			res, err := routingpb.NewPolicyManagerClient(s.cpConn).ListNetworksWithPolicy(ctx, &routingpb.ListNetworksWithPolicyRequest{
				NetId:            s.netID.MarshalNumber(),
				TenantId:         s.tenantIDExtractor(ctx),
				Offset:           (page - 1) * req.Limit,
				Limit:            req.Limit,
				TenantIdContains: req.TenantIdContains,
				NameContains:     req.NameContains,
			})
			return res.GetNetworks(), res.GetTotal(), err
		}
		res, err := iampbv2.NewCatalogClient(s.iamConn).ListNetworks(ctx, &iampbv2.ListNetworksRequest{
			NetId:            s.netID.MarshalNumber(),
			TenantId:         s.tenantIDExtractor(ctx),
			Offset:           (page - 1) * req.Limit,
			Limit:            req.Limit,
			TenantIdContains: req.TenantIdContains,
			NameContains:     req.NameContains,
			PolicyReference: &iampbv2.ListNetworksRequest_PolicyReference{
				NetId:    s.netID.MarshalNumber(),
				TenantId: s.tenantIDExtractor(ctx),
			},
		})
		return res.GetNetworks(), res.GetTotal(), err
	})
}

func (s *pbaServer) ListHomeNetworks(ctx context.Context, req *ttnpb.ListPacketBrokerHomeNetworksRequest) (*ttnpb.PacketBrokerNetworks, error) {
	page := req.Page
	if page == 0 {
		page = 1
	}
	return s.listNetworks(ctx, func() ([]*packetbroker.NetworkOrTenant, uint32, error) {
		res, err := iampbv2.NewCatalogClient(s.iamConn).ListHomeNetworks(ctx, &iampbv2.ListNetworksRequest{
			NetId:            s.netID.MarshalNumber(),
			TenantId:         s.tenantIDExtractor(ctx),
			Offset:           (page - 1) * req.Limit,
			Limit:            req.Limit,
			TenantIdContains: req.TenantIdContains,
			NameContains:     req.NameContains,
			PolicyReference: &iampbv2.ListNetworksRequest_PolicyReference{
				NetId:    s.netID.MarshalNumber(),
				TenantId: s.tenantIDExtractor(ctx),
			},
		})
		return res.GetNetworks(), res.GetTotal(), err
	})
}

func (s *pbaServer) ListForwarderRoutingPolicies(ctx context.Context, req *ttnpb.ListForwarderRoutingPoliciesRequest) (*ttnpb.PacketBrokerRoutingPolicies, error) {
	if err := rights.RequireIsAdmin(ctx); err != nil {
		return nil, err
	}
	if !s.homeNetworkConfig.Enable {
		return nil, errNotEnabled.New()
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
		res.Policies[i] = fromPBRoutingPolicy(p)
	}
	grpc.SetHeader(ctx, metadata.Pairs("x-total-count", strconv.FormatInt(int64(policies.GetTotal()), 10)))
	return res, nil
}
