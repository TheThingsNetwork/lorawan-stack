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

package packetbrokeragent_test

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	iampb "go.packetbroker.org/api/iam"
	iampbv2 "go.packetbroker.org/api/iam/v2"
	routingpb "go.packetbroker.org/api/routing"
	packetbroker "go.packetbroker.org/api/v3"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	. "go.thethings.network/lorawan-stack/v3/pkg/packetbrokeragent"
	"go.thethings.network/lorawan-stack/v3/pkg/packetbrokeragent/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestPba(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, tc := range []struct {
		name                     string
		withIAMHandlers          func(*mock.PBIAM)
		withControlPlaneHandlers func(*mock.PBControlPlane)
		do                       func(context.Context, ttnpb.PbaClient)
	}{
		{
			name: "GetInfo",
			withIAMHandlers: func(p *mock.PBIAM) {
				p.Registry.GetTenantHandler = func(ctx context.Context, req *iampb.TenantRequest) (*iampb.GetTenantResponse, error) {
					return &iampb.GetTenantResponse{
						Tenant: &packetbroker.Tenant{
							NetId:    0x13,
							TenantId: "foo-tenant",
							Name:     "Test Network",
							DevAddrBlocks: []*packetbroker.DevAddrBlock{
								{
									Prefix: &packetbroker.DevAddrPrefix{
										Value:  0x26000000,
										Length: 24,
									},
									HomeNetworkClusterId: "test-cluster",
								},
							},
							AdministrativeContact: &packetbroker.ContactInfo{
								Email: "admin@example.com",
							},
							TechnicalContact: &packetbroker.ContactInfo{
								Email: "tech@example.com",
							},
							Listed: true,
						},
					}, nil
				}
			},
			do: func(ctx context.Context, client ttnpb.PbaClient) {
				t, a := test.MustNewTFromContext(ctx)
				res, err := client.GetInfo(ctx, ttnpb.Empty)
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				a.So(res, should.Resemble, &ttnpb.PacketBrokerInfo{
					Registration: &ttnpb.PacketBrokerNetwork{
						Id: &ttnpb.PacketBrokerNetworkIdentifier{
							NetId:    0x13,
							TenantId: "foo-tenant",
						},
						Name: "Test Network",
						DevAddrBlocks: []*ttnpb.PacketBrokerDevAddrBlock{
							{
								DevAddrPrefix: &ttnpb.DevAddrPrefix{
									DevAddr: devAddrPtr(types.DevAddr{0x26, 0x0, 0x0, 0x0}),
									Length:  24,
								},
								HomeNetworkClusterId: "test-cluster",
							},
						},
						ContactInfo: []*ttnpb.ContactInfo{
							{
								ContactType:   ttnpb.ContactType_CONTACT_TYPE_OTHER,
								ContactMethod: ttnpb.CONTACT_METHOD_EMAIL,
								Value:         "admin@example.com",
							},
							{
								ContactType:   ttnpb.ContactType_CONTACT_TYPE_TECHNICAL,
								ContactMethod: ttnpb.CONTACT_METHOD_EMAIL,
								Value:         "tech@example.com",
							},
						},
						Listed: true,
					},
					ForwarderEnabled:   true,
					HomeNetworkEnabled: true,
				})
			},
		},
		{
			name: "Register/Create",
			withIAMHandlers: func(p *mock.PBIAM) {
				p.Registry.GetTenantHandler = func(ctx context.Context, req *iampb.TenantRequest) (*iampb.GetTenantResponse, error) {
					return nil, status.Error(codes.NotFound, "not found")
				}
				p.Registry.CreateTenantHandler = func(ctx context.Context, req *iampb.CreateTenantRequest) (*iampb.CreateTenantResponse, error) {
					_, a := test.MustNewTFromContext(ctx)
					a.So(req.Tenant, should.Resemble, &packetbroker.Tenant{
						NetId:    0x13,
						TenantId: "foo-tenant",
						Name:     "Test Network",
						DevAddrBlocks: []*packetbroker.DevAddrBlock{
							{
								Prefix: &packetbroker.DevAddrPrefix{
									Value:  0x26000000,
									Length: 24,
								},
								HomeNetworkClusterId: "test-cluster",
							},
						},
						AdministrativeContact: &packetbroker.ContactInfo{
							Email: "admin@example.com",
						},
						TechnicalContact: &packetbroker.ContactInfo{
							Email: "tech@example.com",
						},
						Listed: true,
					})
					return &iampb.CreateTenantResponse{
						Tenant: req.Tenant,
					}, nil
				}
			},
			do: func(ctx context.Context, client ttnpb.PbaClient) {
				t, a := test.MustNewTFromContext(ctx)
				res, err := client.Register(ctx, &ttnpb.PacketBrokerRegisterRequest{})
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				a.So(res, should.Resemble, &ttnpb.PacketBrokerNetwork{
					Id: &ttnpb.PacketBrokerNetworkIdentifier{
						NetId:    0x13,
						TenantId: "foo-tenant",
					},
					Name: "Test Network",
					DevAddrBlocks: []*ttnpb.PacketBrokerDevAddrBlock{
						{
							DevAddrPrefix: &ttnpb.DevAddrPrefix{
								DevAddr: devAddrPtr(types.DevAddr{0x26, 0x0, 0x0, 0x0}),
								Length:  24,
							},
							HomeNetworkClusterId: "test-cluster",
						},
					},
					ContactInfo: []*ttnpb.ContactInfo{
						{
							ContactType:   ttnpb.ContactType_CONTACT_TYPE_OTHER,
							ContactMethod: ttnpb.CONTACT_METHOD_EMAIL,
							Value:         "admin@example.com",
						},
						{
							ContactType:   ttnpb.ContactType_CONTACT_TYPE_TECHNICAL,
							ContactMethod: ttnpb.CONTACT_METHOD_EMAIL,
							Value:         "tech@example.com",
						},
					},
					Listed: true,
				})
			},
		},
		{
			name: "Register/Update",
			withIAMHandlers: func(p *mock.PBIAM) {
				p.Registry.GetTenantHandler = func(ctx context.Context, req *iampb.TenantRequest) (*iampb.GetTenantResponse, error) {
					return &iampb.GetTenantResponse{
						Tenant: &packetbroker.Tenant{
							NetId:    0x13,
							TenantId: "foo-tenant",
							Name:     "Test Network",
							DevAddrBlocks: []*packetbroker.DevAddrBlock{
								{
									Prefix: &packetbroker.DevAddrPrefix{
										Value:  0x26000000,
										Length: 24,
									},
									HomeNetworkClusterId: "test-cluster",
								},
							},
							AdministrativeContact: &packetbroker.ContactInfo{
								Email: "admin@example.com",
							},
							TechnicalContact: &packetbroker.ContactInfo{
								Email: "tech@example.com",
							},
							Listed: false,
						},
					}, nil
				}
				p.Registry.UpdateTenantHandler = func(ctx context.Context, req *iampb.UpdateTenantRequest) (*pbtypes.Empty, error) {
					_, a := test.MustNewTFromContext(ctx)
					a.So(req, should.Resemble, &iampb.UpdateTenantRequest{
						NetId:    0x13,
						TenantId: "foo-tenant",
						Name: &pbtypes.StringValue{
							Value: "Test Network",
						},
						// NOTE: DevAddrBlocks are not updated here, as the tenant cannot change their own DevAddr blocks.
						AdministrativeContact: &packetbroker.ContactInfoValue{
							Value: &packetbroker.ContactInfo{
								Email: "admin@example.com",
							},
						},
						TechnicalContact: &packetbroker.ContactInfoValue{
							Value: &packetbroker.ContactInfo{
								Email: "tech@example.com",
							},
						},
						Listed: &pbtypes.BoolValue{
							Value: true,
						},
					})
					return ttnpb.Empty, nil
				}
			},
			do: func(ctx context.Context, client ttnpb.PbaClient) {
				t, a := test.MustNewTFromContext(ctx)
				res, err := client.Register(ctx, &ttnpb.PacketBrokerRegisterRequest{})
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				a.So(res, should.Resemble, &ttnpb.PacketBrokerNetwork{
					Id: &ttnpb.PacketBrokerNetworkIdentifier{
						NetId:    0x13,
						TenantId: "foo-tenant",
					},
					Name: "Test Network",
					DevAddrBlocks: []*ttnpb.PacketBrokerDevAddrBlock{
						{
							DevAddrPrefix: &ttnpb.DevAddrPrefix{
								DevAddr: devAddrPtr(types.DevAddr{0x26, 0x0, 0x0, 0x0}),
								Length:  24,
							},
							HomeNetworkClusterId: "test-cluster",
						},
					},
					ContactInfo: []*ttnpb.ContactInfo{
						{
							ContactType:   ttnpb.ContactType_CONTACT_TYPE_OTHER,
							ContactMethod: ttnpb.CONTACT_METHOD_EMAIL,
							Value:         "admin@example.com",
						},
						{
							ContactType:   ttnpb.ContactType_CONTACT_TYPE_TECHNICAL,
							ContactMethod: ttnpb.CONTACT_METHOD_EMAIL,
							Value:         "tech@example.com",
						},
					},
					Listed: true,
				})
			},
		},
		{
			name: "Deregister",
			withIAMHandlers: func(p *mock.PBIAM) {
				p.Registry.DeleteTenantHandler = func(ctx context.Context, req *iampb.TenantRequest) (*pbtypes.Empty, error) {
					_, a := test.MustNewTFromContext(ctx)
					a.So(req.NetId, should.Equal, 0x13)
					a.So(req.TenantId, should.Equal, "foo-tenant")
					return ttnpb.Empty, nil
				}
			},
			do: func(ctx context.Context, client ttnpb.PbaClient) {
				_, a := test.MustNewTFromContext(ctx)
				_, err := client.Deregister(ctx, ttnpb.Empty)
				a.So(err, should.BeNil)
			},
		},
		{
			name: "RoutingPolicy/Default/Get",
			withControlPlaneHandlers: func(p *mock.PBControlPlane) {
				p.GetDefaultPolicyHandler = func(ctx context.Context, req *routingpb.GetDefaultPolicyRequest) (*routingpb.GetPolicyResponse, error) {
					_, a := test.MustNewTFromContext(ctx)
					a.So(req.ForwarderNetId, should.Equal, 0x13)
					a.So(req.ForwarderTenantId, should.Equal, "foo-tenant")
					return &routingpb.GetPolicyResponse{
						Policy: &packetbroker.RoutingPolicy{
							ForwarderNetId:    0x13,
							ForwarderTenantId: "foo-tenant",
							UpdatedAt:         pbtypes.TimestampNow(),
							Uplink: &packetbroker.RoutingPolicy_Uplink{
								JoinRequest:     true,
								MacData:         true,
								ApplicationData: true,
							},
							Downlink: &packetbroker.RoutingPolicy_Downlink{
								JoinAccept: true,
								MacData:    true,
							},
						},
					}, nil
				}
			},
			do: func(ctx context.Context, client ttnpb.PbaClient) {
				_, a := test.MustNewTFromContext(ctx)
				res, err := client.GetHomeNetworkDefaultRoutingPolicy(ctx, ttnpb.Empty)
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				a.So(res, should.Resemble, &ttnpb.PacketBrokerDefaultRoutingPolicy{
					UpdatedAt: res.UpdatedAt,
					Uplink: &ttnpb.PacketBrokerRoutingPolicyUplink{
						JoinRequest:     true,
						MacData:         true,
						ApplicationData: true,
					},
					Downlink: &ttnpb.PacketBrokerRoutingPolicyDownlink{
						JoinAccept: true,
						MacData:    true,
					},
				})
				a.So(test.Must(pbtypes.TimestampFromProto(res.UpdatedAt)).(time.Time), should.HappenBetween, time.Now().Add(-1*time.Second), time.Now())
			},
		},
		{
			name: "RoutingPolicy/Default/Set",
			withControlPlaneHandlers: func(p *mock.PBControlPlane) {
				p.SetDefaultPolicyHandler = func(ctx context.Context, req *routingpb.SetPolicyRequest) (*pbtypes.Empty, error) {
					_, a := test.MustNewTFromContext(ctx)
					a.So(req.Policy, should.Resemble, &packetbroker.RoutingPolicy{
						ForwarderNetId:    0x13,
						ForwarderTenantId: "foo-tenant",
						Uplink: &packetbroker.RoutingPolicy_Uplink{
							JoinRequest: true,
						},
						Downlink: &packetbroker.RoutingPolicy_Downlink{
							JoinAccept: true,
						},
					})
					return ttnpb.Empty, nil
				}
			},
			do: func(ctx context.Context, client ttnpb.PbaClient) {
				_, a := test.MustNewTFromContext(ctx)
				_, err := client.SetHomeNetworkDefaultRoutingPolicy(ctx, &ttnpb.SetPacketBrokerDefaultRoutingPolicyRequest{
					Uplink: &ttnpb.PacketBrokerRoutingPolicyUplink{
						JoinRequest: true,
					},
					Downlink: &ttnpb.PacketBrokerRoutingPolicyDownlink{
						JoinAccept: true,
					},
				})
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
			},
		},
		{
			name: "RoutingPolicy/Default/Delete",
			withControlPlaneHandlers: func(p *mock.PBControlPlane) {
				p.SetDefaultPolicyHandler = func(ctx context.Context, req *routingpb.SetPolicyRequest) (*pbtypes.Empty, error) {
					_, a := test.MustNewTFromContext(ctx)
					a.So(req.Policy, should.Resemble, &packetbroker.RoutingPolicy{
						ForwarderNetId:    0x13,
						ForwarderTenantId: "foo-tenant",
					})
					return ttnpb.Empty, nil
				}
			},
			do: func(ctx context.Context, client ttnpb.PbaClient) {
				_, a := test.MustNewTFromContext(ctx)
				_, err := client.DeleteHomeNetworkDefaultRoutingPolicy(ctx, ttnpb.Empty)
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
			},
		},
		{
			name: "RoutingPolicy/HomeNetwork/List",
			withControlPlaneHandlers: func(p *mock.PBControlPlane) {
				policies := make([]*packetbroker.RoutingPolicy, 42)
				for i := 0; i < len(policies); i++ {
					policies[i] = &packetbroker.RoutingPolicy{
						ForwarderNetId:    0x13,
						ForwarderTenantId: "foo-tenant",
						HomeNetworkNetId:  uint32(i) + 1,
						UpdatedAt:         test.Must(pbtypes.TimestampProto(time.Unix(int64(i), 0))).(*pbtypes.Timestamp),
						Uplink: &packetbroker.RoutingPolicy_Uplink{
							JoinRequest:     i%2 == 0,
							MacData:         i%3 == 0,
							ApplicationData: i%4 == 0,
							SignalQuality:   i%5 == 0,
							Localization:    i%6 == 0,
						},
						Downlink: &packetbroker.RoutingPolicy_Downlink{
							JoinAccept:      i%7 == 0,
							MacData:         i%8 == 0,
							ApplicationData: i%9 == 0,
						},
					}
				}
				p.ListHomeNetworkPoliciesHandler = func(ctx context.Context, req *routingpb.ListHomeNetworkPoliciesRequest) (*routingpb.ListHomeNetworkPoliciesResponse, error) {
					_, a := test.MustNewTFromContext(ctx)
					a.So(req.ForwarderNetId, should.Equal, 0x13)
					a.So(req.ForwarderTenantId, should.Equal, "foo-tenant")
					limit := req.Limit
					if limit == 0 {
						limit = 1
					}
					res := &routingpb.ListHomeNetworkPoliciesResponse{
						Total: uint32(len(policies)),
					}
					for _, p := range policies {
						if p.UpdatedAt.Compare(req.UpdatedSince) > 0 {
							res.Policies = append(res.Policies, p)
						}
						if len(res.Policies) >= int(limit) {
							break
						}
					}
					return res, nil
				}
			},
			do: func(ctx context.Context, client ttnpb.PbaClient) {
				t := test.MustTFromContext(ctx)
				for _, limit := range []int{1, 2, 1000} {
					t.Run(strconv.FormatInt(int64(limit), 10), func(t *testing.T) {
						a := assertions.New(t)
						var policies []*ttnpb.PacketBrokerRoutingPolicy
						for page := 1; ; page++ {
							md := &metadata.MD{}
							res, err := client.ListHomeNetworkRoutingPolicies(ctx, &ttnpb.ListHomeNetworkRoutingPoliciesRequest{
								Limit: uint32(limit),
								Page:  uint32(page),
							}, grpc.Header(md))
							if !a.So(err, should.BeNil) {
								t.FailNow()
							}
							if !a.So(test.Must(strconv.ParseInt(md.Get("x-total-count")[0], 10, 32)).(int64), should.Equal, 42) {
								t.FailNow()
							}
							if !a.So(len(res.Policies), should.BeLessThanOrEqualTo, limit) {
								t.FailNow()
							}
							if len(res.Policies) == 0 {
								break
							}
							policies = append(policies, res.Policies...)
						}
						a.So(policies, should.HaveLength, 42)
						for i := 0; i < len(policies); i++ {
							a.So(policies[i], should.Resemble, &ttnpb.PacketBrokerRoutingPolicy{
								ForwarderId: &ttnpb.PacketBrokerNetworkIdentifier{
									NetId:    0x13,
									TenantId: "foo-tenant",
								},
								HomeNetworkId: &ttnpb.PacketBrokerNetworkIdentifier{
									NetId: uint32(i) + 1,
								},
								UpdatedAt: test.Must(pbtypes.TimestampProto(time.Unix(int64(i), 0))).(*pbtypes.Timestamp),
								Uplink: &ttnpb.PacketBrokerRoutingPolicyUplink{
									JoinRequest:     i%2 == 0,
									MacData:         i%3 == 0,
									ApplicationData: i%4 == 0,
									SignalQuality:   i%5 == 0,
									Localization:    i%6 == 0,
								},
								Downlink: &ttnpb.PacketBrokerRoutingPolicyDownlink{
									JoinAccept:      i%7 == 0,
									MacData:         i%8 == 0,
									ApplicationData: i%9 == 0,
								},
							})
						}
					})
				}
			},
		},
		{
			name: "Network/List",
			withIAMHandlers: func(p *mock.PBIAM) {
				networks := generateNetworks(42)
				p.Catalog.ListNetworksHandler = func(ctx context.Context, req *iampbv2.ListNetworksRequest) (*iampbv2.ListNetworksResponse, error) {
					offset := int(req.Offset)
					limit := int(req.Limit)
					if limit == 0 {
						limit = 1
					}
					var slice []*packetbroker.NetworkOrTenant
					if len(networks) > offset {
						slice = networks[offset:]
						if len(slice) > limit {
							slice = slice[:limit]
						}
					}
					return &iampbv2.ListNetworksResponse{
						Networks: slice,
						Total:    uint32(len(networks)),
					}, nil
				}
			},
			withControlPlaneHandlers: func(p *mock.PBControlPlane) {
				networks := generateNetworks(21)
				p.ListNetworksWithPolicyHandler = func(ctx context.Context, req *routingpb.ListNetworksWithPolicyRequest) (*routingpb.ListNetworksResponse, error) {
					offset := int(req.Offset)
					limit := int(req.Limit)
					if limit == 0 {
						limit = 1
					}
					var slice []*packetbroker.NetworkOrTenant
					if len(networks) > offset {
						slice = networks[offset:]
						if len(slice) > limit {
							slice = slice[:limit]
						}
					}
					return &routingpb.ListNetworksResponse{
						Networks: slice,
						Total:    uint32(len(networks)),
					}, nil
				}
			},
			do: func(ctx context.Context, client ttnpb.PbaClient) {
				t := test.MustTFromContext(ctx)
				for i, req := range []struct {
					withPolicy    bool
					limit         int
					expectedTotal int
				}{
					{
						withPolicy:    false,
						limit:         1,
						expectedTotal: 42,
					},
					{
						withPolicy:    true,
						limit:         100,
						expectedTotal: 21,
					},
				} {
					t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) {
						a := assertions.New(t)
						var networks []*ttnpb.PacketBrokerNetwork
						for page := 1; ; page++ {
							md := &metadata.MD{}
							res, err := client.ListNetworks(ctx, &ttnpb.ListPacketBrokerNetworksRequest{
								Limit:             uint32(req.limit),
								Page:              uint32(page),
								WithRoutingPolicy: req.withPolicy,
							}, grpc.Header(md))
							if !a.So(err, should.BeNil) {
								t.FailNow()
							}
							if !a.So(test.Must(strconv.ParseInt(md.Get("x-total-count")[0], 10, 32)).(int64), should.Equal, req.expectedTotal) {
								t.FailNow()
							}
							if !a.So(len(res.Networks), should.BeLessThanOrEqualTo, req.limit) {
								t.FailNow()
							}
							if len(res.Networks) == 0 {
								break
							}
							networks = append(networks, res.Networks...)
						}
						a.So(networks, should.HaveLength, req.expectedTotal)
						for i, n := range networks {
							if i%2 == 0 {
								a.So(n.Id, should.Resemble, &ttnpb.PacketBrokerNetworkIdentifier{
									NetId: uint32(i),
								})
							} else {
								a.So(n.Id, should.Resemble, &ttnpb.PacketBrokerNetworkIdentifier{
									NetId:    uint32(i),
									TenantId: fmt.Sprintf("tenant-%d", i),
								})
							}
							a.So(n.Name, should.Equal, fmt.Sprintf("Network %06X", i))
						}
					})
				}
			},
		},
		{
			name: "HomeNetwork/List",
			withIAMHandlers: func(p *mock.PBIAM) {
				networks := generateNetworks(42)
				p.Catalog.ListHomeNetworksHandler = func(ctx context.Context, req *iampbv2.ListNetworksRequest) (*iampbv2.ListNetworksResponse, error) {
					offset := int(req.Offset)
					limit := int(req.Limit)
					if limit == 0 {
						limit = 1
					}
					var slice []*packetbroker.NetworkOrTenant
					if len(networks) > offset {
						slice = networks[offset:]
						if len(slice) > limit {
							slice = slice[:limit]
						}
					}
					return &iampbv2.ListNetworksResponse{
						Networks: slice,
						Total:    uint32(len(networks)),
					}, nil
				}
			},
			do: func(ctx context.Context, client ttnpb.PbaClient) {
				t := test.MustTFromContext(ctx)
				for _, limit := range []int{1, 2, 1000} {
					t.Run(strconv.FormatInt(int64(limit), 10), func(t *testing.T) {
						a := assertions.New(t)
						var networks []*ttnpb.PacketBrokerNetwork
						for page := 1; ; page++ {
							md := &metadata.MD{}
							res, err := client.ListHomeNetworks(ctx, &ttnpb.ListPacketBrokerHomeNetworksRequest{
								Limit: uint32(limit),
								Page:  uint32(page),
							}, grpc.Header(md))
							if !a.So(err, should.BeNil) {
								t.FailNow()
							}
							if !a.So(test.Must(strconv.ParseInt(md.Get("x-total-count")[0], 10, 32)).(int64), should.Equal, 42) {
								t.FailNow()
							}
							if !a.So(len(res.Networks), should.BeLessThanOrEqualTo, limit) {
								t.FailNow()
							}
							if len(res.Networks) == 0 {
								break
							}
							networks = append(networks, res.Networks...)
						}
						a.So(networks, should.HaveLength, 42)
						for i, n := range networks {
							if i%2 == 0 {
								a.So(n.Id, should.Resemble, &ttnpb.PacketBrokerNetworkIdentifier{
									NetId: uint32(i),
								})
							} else {
								a.So(n.Id, should.Resemble, &ttnpb.PacketBrokerNetworkIdentifier{
									NetId:    uint32(i),
									TenantId: fmt.Sprintf("tenant-%d", i),
								})
							}
							a.So(n.Name, should.Equal, fmt.Sprintf("Network %06X", i))
						}
					})
				}
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a, ctx := test.NewWithContext(ctx, t)

			iamLis, err := net.Listen("tcp", ":0")
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			iam := mock.NewPBIAM(t)
			if tc.withIAMHandlers != nil {
				tc.withIAMHandlers(iam)
			}
			go iam.Serve(iamLis)
			defer iam.GracefulStop()

			cpLis, err := net.Listen("tcp", ":0")
			if err != nil {
				t.Fatalf("Listen Control Plane: %v", err)
			}
			controlplane := mock.NewPBControlPlane(t)
			if tc.withControlPlaneHandlers != nil {
				tc.withControlPlaneHandlers(controlplane)
			}
			go controlplane.Serve(cpLis)
			defer controlplane.GracefulStop()

			c := componenttest.NewComponent(t, &component.Config{})
			c.AddContextFiller(func(ctx context.Context) context.Context {
				return rights.NewContextWithAuthInfo(ctx, &ttnpb.AuthInfoResponse{
					IsAdmin: true,
				})
			})
			c.AddContextFiller(func(ctx context.Context) context.Context {
				return test.ContextWithTB(ctx, t)
			})
			_, err = New(c, &Config{
				IAMAddress:          iamLis.Addr().String(),
				ControlPlaneAddress: cpLis.Addr().String(),
				NetID:               types.NetID{0x0, 0x0, 0x13},
				TenantID:            "foo-tenant",
				ClusterID:           "test-cluster",
				Registration: RegistrationConfig{
					Name: "Test Network",
					AdministrativeContact: ContactInfoConfig{
						Email: "admin@example.com",
					},
					TechnicalContact: ContactInfoConfig{
						Email: "tech@example.com",
					},
					Listed: true,
				},
				Forwarder: ForwarderConfig{
					Enable: true,
				},
				HomeNetwork: HomeNetworkConfig{
					Enable: true,
					DevAddrPrefixes: []types.DevAddrPrefix{
						{
							DevAddr: types.DevAddr{0x26, 0x0, 0x0, 0x0},
							Length:  24,
						},
					},
				},
			}, testOptions...)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			if !a.So(c.Start(), should.BeNil) {
				t.FailNow()
			}
			defer c.Close()

			tc.do(ctx, ttnpb.NewPbaClient(c.LoopbackConn()))
		})
	}
}

func generateNetworks(n int) []*packetbroker.NetworkOrTenant {
	networks := make([]*packetbroker.NetworkOrTenant, n)
	for i := 0; i < len(networks); i++ {
		networks[i] = &packetbroker.NetworkOrTenant{}
		if i%2 == 0 {
			networks[i].Value = &packetbroker.NetworkOrTenant_Network{
				Network: &packetbroker.Network{
					NetId: uint32(i),
					Name:  fmt.Sprintf("Network %06X", i),
				},
			}
		} else {
			networks[i].Value = &packetbroker.NetworkOrTenant_Tenant{
				Tenant: &packetbroker.Tenant{
					NetId:    uint32(i),
					TenantId: fmt.Sprintf("tenant-%d", i),
					Name:     fmt.Sprintf("Network %06X", i),
					DevAddrBlocks: []*packetbroker.DevAddrBlock{
						{
							Prefix: &packetbroker.DevAddrPrefix{
								Value:  uint32(i) << 16,
								Length: 16,
							},
							HomeNetworkClusterId: fmt.Sprintf("cluster-%d", i),
						},
					},
				},
			}
		}
	}
	return networks
}
