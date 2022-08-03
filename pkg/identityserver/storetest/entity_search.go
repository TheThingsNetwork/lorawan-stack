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

package storetest

import (
	. "testing"

	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	is "go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func (st *StoreTest) TestEntitySearch(t *T) {
	usr1 := st.population.NewUser()
	usr1.Description = "This is the description of " + usr1.Name
	usr1.Attributes = attributes
	usr1.State = ttnpb.State_STATE_FLAGGED
	st.population.NewUser()

	app1 := st.population.NewApplication(usr1.GetOrganizationOrUserIdentifiers())
	app1.Description = "This is the description of " + app1.Name
	app1.Attributes = attributes
	app2 := st.population.NewApplication(nil)

	dev1 := st.population.NewEndDevice(app1.GetIds())
	dev1.Ids.JoinEui = types.EUI64{1, 1, 1, 1, 1, 1, 1, 1}.Bytes()
	dev1.Ids.DevEui = types.EUI64{2, 2, 2, 2, 2, 2, 2, 2}.Bytes()
	dev1.Description = "This is the description of " + dev1.Name
	dev1.Attributes = attributes
	st.population.NewEndDevice(app2.GetIds())

	cli1 := st.population.NewClient(usr1.GetOrganizationOrUserIdentifiers())
	cli1.Description = "This is the description of " + cli1.Name
	cli1.Attributes = attributes
	cli1.State = ttnpb.State_STATE_FLAGGED
	st.population.NewClient(nil)

	gtw1 := st.population.NewGateway(usr1.GetOrganizationOrUserIdentifiers())
	gtw1.Ids.Eui = types.EUI64{3, 3, 3, 3, 3, 3, 3, 3}.Bytes()
	gtw1.Description = "This is the description of " + gtw1.Name
	gtw1.Attributes = attributes
	st.population.NewGateway(nil)

	org1 := st.population.NewOrganization(usr1.GetOrganizationOrUserIdentifiers())
	org1.Description = "This is the description of " + org1.Name
	org1.Attributes = attributes
	st.population.NewOrganization(nil)

	s, ok := st.PrepareDB(t).(interface {
		Store

		is.EntitySearch
	})
	defer st.DestroyDB(t, false)
	if !ok {
		t.Skip("Store does not implement EntitySearch")
	}
	defer s.Close()

	t.Run("Applications", func(t *T) {
		t.Run("WithoutUser", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchApplications(ctx, nil, &ttnpb.SearchApplicationsRequest{})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) {
				a.So(ids, should.HaveLength, 2)
			}
		})

		t.Run("WithUser", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchApplications(ctx, usr1.GetOrganizationOrUserIdentifiers(), &ttnpb.SearchApplicationsRequest{})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) {
				a.So(ids, should.HaveLength, 1)
			}
		})

		t.Run("Query", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchApplications(ctx, nil, &ttnpb.SearchApplicationsRequest{
				Query: "-01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, app1.GetIds())
			}
		})
		t.Run("ID", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchApplications(ctx, nil, &ttnpb.SearchApplicationsRequest{
				IdContains: "01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, app1.GetIds())
			}
		})
		t.Run("Name", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchApplications(ctx, nil, &ttnpb.SearchApplicationsRequest{
				NameContains: "01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, app1.GetIds())
			}
		})
		t.Run("Description", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchApplications(ctx, nil, &ttnpb.SearchApplicationsRequest{
				DescriptionContains: "01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, app1.GetIds())
			}
		})
		t.Run("Attributes", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchApplications(ctx, nil, &ttnpb.SearchApplicationsRequest{
				AttributesContain: map[string]string{"foo": "ba"},
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, app1.GetIds())
			}
		})
	})

	t.Run("Clients", func(t *T) {
		t.Run("WithoutUser", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchClients(ctx, nil, &ttnpb.SearchClientsRequest{})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) {
				a.So(ids, should.HaveLength, 2)
			}
		})

		t.Run("WithUser", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchClients(ctx, usr1.GetOrganizationOrUserIdentifiers(), &ttnpb.SearchClientsRequest{})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) {
				a.So(ids, should.HaveLength, 1)
			}
		})

		t.Run("Query", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchClients(ctx, nil, &ttnpb.SearchClientsRequest{
				Query: "-01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, cli1.GetIds())
			}
		})
		t.Run("ID", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchClients(ctx, nil, &ttnpb.SearchClientsRequest{
				IdContains: "01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, cli1.GetIds())
			}
		})
		t.Run("Name", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchClients(ctx, nil, &ttnpb.SearchClientsRequest{
				NameContains: "01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, cli1.GetIds())
			}
		})
		t.Run("Description", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchClients(ctx, nil, &ttnpb.SearchClientsRequest{
				DescriptionContains: "01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, cli1.GetIds())
			}
		})
		t.Run("Attributes", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchClients(ctx, nil, &ttnpb.SearchClientsRequest{
				AttributesContain: map[string]string{"foo": "ba"},
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, cli1.GetIds())
			}
		})
		t.Run("State", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchClients(ctx, nil, &ttnpb.SearchClientsRequest{
				State: []ttnpb.State{ttnpb.State_STATE_FLAGGED},
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, cli1.GetIds())
			}
		})
	})

	t.Run("EndDevices", func(t *T) {
		dev1ID := &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: app1.GetIds(),
			DeviceId:       dev1.GetIds().GetDeviceId(),
		} // Ignore the EUIs.

		t.Run("WithoutApplication", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchEndDevices(ctx, &ttnpb.SearchEndDevicesRequest{})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) {
				a.So(ids, should.HaveLength, 2)
			}
		})

		t.Run("WithApplication", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchEndDevices(ctx, &ttnpb.SearchEndDevicesRequest{
				ApplicationIds: app1.GetIds(),
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) {
				a.So(ids, should.HaveLength, 1)
			}
		})

		t.Run("Query", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchEndDevices(ctx, &ttnpb.SearchEndDevicesRequest{
				ApplicationIds: app1.GetIds(),
				Query:          "01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, dev1ID)
			}
		})
		t.Run("ID", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchEndDevices(ctx, &ttnpb.SearchEndDevicesRequest{
				ApplicationIds: app1.GetIds(),
				IdContains:     "01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, dev1ID)
			}
		})
		t.Run("JoinEUI", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchEndDevices(ctx, &ttnpb.SearchEndDevicesRequest{
				ApplicationIds:  app1.GetIds(),
				JoinEuiContains: "0101",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, dev1ID)
			}
		})
		t.Run("DevEUI", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchEndDevices(ctx, &ttnpb.SearchEndDevicesRequest{
				ApplicationIds: app1.GetIds(),
				DevEuiContains: "0202",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, dev1ID)
			}
		})
		t.Run("Name", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchEndDevices(ctx, &ttnpb.SearchEndDevicesRequest{
				ApplicationIds: app1.GetIds(),
				NameContains:   "01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, dev1ID)
			}
		})
		t.Run("Description", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchEndDevices(ctx, &ttnpb.SearchEndDevicesRequest{
				ApplicationIds:      app1.GetIds(),
				DescriptionContains: "01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, dev1ID)
			}
		})
		t.Run("Attributes", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchEndDevices(ctx, &ttnpb.SearchEndDevicesRequest{
				ApplicationIds:    app1.GetIds(),
				AttributesContain: map[string]string{"foo": "ba"},
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, dev1ID)
			}
		})
	})

	t.Run("Gateways", func(t *T) {
		gtw1ID := &ttnpb.GatewayIdentifiers{GatewayId: gtw1.GetIds().GetGatewayId()} // Ignore the EUI.

		t.Run("WithoutUser", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchGateways(ctx, nil, &ttnpb.SearchGatewaysRequest{})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) {
				a.So(ids, should.HaveLength, 2)
			}
		})

		t.Run("WithUser", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchGateways(ctx, usr1.GetOrganizationOrUserIdentifiers(), &ttnpb.SearchGatewaysRequest{})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) {
				a.So(ids, should.HaveLength, 1)
			}
		})

		t.Run("Query", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchGateways(ctx, nil, &ttnpb.SearchGatewaysRequest{
				Query: "-01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, gtw1ID)
			}
		})
		t.Run("ID", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchGateways(ctx, nil, &ttnpb.SearchGatewaysRequest{
				IdContains: "01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, gtw1ID)
			}
		})
		t.Run("EUI", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchGateways(ctx, nil, &ttnpb.SearchGatewaysRequest{
				EuiContains: "0303",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, gtw1ID)
			}
		})
		t.Run("Name", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchGateways(ctx, nil, &ttnpb.SearchGatewaysRequest{
				NameContains: "01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, gtw1ID)
			}
		})
		t.Run("Description", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchGateways(ctx, nil, &ttnpb.SearchGatewaysRequest{
				DescriptionContains: "01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, gtw1ID)
			}
		})
		t.Run("Attributes", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchGateways(ctx, nil, &ttnpb.SearchGatewaysRequest{
				AttributesContain: map[string]string{"foo": "ba"},
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, gtw1ID)
			}
		})
	})

	t.Run("Organizations", func(t *T) {
		t.Run("WithoutUser", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchOrganizations(ctx, nil, &ttnpb.SearchOrganizationsRequest{})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) {
				a.So(ids, should.HaveLength, 2)
			}
		})

		t.Run("WithUser", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchOrganizations(ctx, usr1.GetOrganizationOrUserIdentifiers(), &ttnpb.SearchOrganizationsRequest{})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) {
				a.So(ids, should.HaveLength, 1)
			}
		})

		t.Run("Query", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchOrganizations(ctx, nil, &ttnpb.SearchOrganizationsRequest{
				Query: "-01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, org1.GetIds())
			}
		})
		t.Run("ID", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchOrganizations(ctx, nil, &ttnpb.SearchOrganizationsRequest{
				IdContains: "01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, org1.GetIds())
			}
		})
		t.Run("Name", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchOrganizations(ctx, nil, &ttnpb.SearchOrganizationsRequest{
				NameContains: "01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, org1.GetIds())
			}
		})
		t.Run("Description", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchOrganizations(ctx, nil, &ttnpb.SearchOrganizationsRequest{
				DescriptionContains: "01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, org1.GetIds())
			}
		})
		t.Run("Attributes", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchOrganizations(ctx, nil, &ttnpb.SearchOrganizationsRequest{
				AttributesContain: map[string]string{"foo": "ba"},
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, org1.GetIds())
			}
		})
	})

	t.Run("Users", func(t *T) {
		t.Run("Query", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchUsers(ctx, &ttnpb.SearchUsersRequest{
				Query: "-01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, usr1.GetIds())
			}
		})
		t.Run("ID", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchUsers(ctx, &ttnpb.SearchUsersRequest{
				IdContains: "01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, usr1.GetIds())
			}
		})
		t.Run("Name", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchUsers(ctx, &ttnpb.SearchUsersRequest{
				NameContains: "01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, usr1.GetIds())
			}
		})
		t.Run("Description", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchUsers(ctx, &ttnpb.SearchUsersRequest{
				DescriptionContains: "01",
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, usr1.GetIds())
			}
		})
		t.Run("Attributes", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchUsers(ctx, &ttnpb.SearchUsersRequest{
				AttributesContain: map[string]string{"foo": "ba"},
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, usr1.GetIds())
			}
		})
		t.Run("State", func(t *T) {
			a, ctx := test.New(t)
			ids, err := s.SearchUsers(ctx, &ttnpb.SearchUsersRequest{
				State: []ttnpb.State{ttnpb.State_STATE_FLAGGED},
			})
			if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) && a.So(ids, should.HaveLength, 1) {
				a.So(ids[0], should.Resemble, usr1.GetIds())
			}
		})
	})
}

func (st *StoreTest) TestEntitySearchPagination(t *T) {
	var users []*ttnpb.User
	for i := 0; i < 7; i++ {
		users = append(users, st.population.NewUser())
	}

	var applications []*ttnpb.Application
	for i := 0; i < 7; i++ {
		applications = append(applications, st.population.NewApplication(users[0].GetOrganizationOrUserIdentifiers()))
	}
	var clients []*ttnpb.Client
	for i := 0; i < 7; i++ {
		clients = append(clients, st.population.NewClient(users[0].GetOrganizationOrUserIdentifiers()))
	}
	var gateways []*ttnpb.Gateway
	for i := 0; i < 7; i++ {
		gateways = append(gateways, st.population.NewGateway(users[0].GetOrganizationOrUserIdentifiers()))
	}
	var organizations []*ttnpb.Organization
	for i := 0; i < 7; i++ {
		organizations = append(organizations, st.population.NewOrganization(users[0].GetOrganizationOrUserIdentifiers()))
	}

	var endDevices []*ttnpb.EndDevice
	for i := 0; i < 7; i++ {
		endDevices = append(endDevices, st.population.NewEndDevice(applications[0].GetIds()))
	}

	s, ok := st.PrepareDB(t).(interface {
		Store
		is.EntitySearch
	})
	defer st.DestroyDB(t, false)
	if !ok {
		t.Skip("Store does not implement EntitySearch")
	}
	defer s.Close()

	t.Run("Applications_Paginated", func(t *T) {
		a, ctx := test.New(t)

		var total uint64
		for _, page := range []uint32{1, 2, 3, 4} {
			paginateCtx := store.WithPagination(ctx, 2, page, &total)

			got, err := s.SearchApplications(paginateCtx, nil, &ttnpb.SearchApplicationsRequest{})
			if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
				if page == 4 {
					a.So(got, should.HaveLength, 1)
				} else {
					a.So(got, should.HaveLength, 2)
				}
				for i, e := range got {
					a.So(e, should.Resemble, applications[i+2*int(page-1)].Ids)
				}
			}

			a.So(total, should.Equal, 7)
		}
	})

	t.Run("Clients_Paginated", func(t *T) {
		a, ctx := test.New(t)

		var total uint64
		for _, page := range []uint32{1, 2, 3, 4} {
			paginateCtx := store.WithPagination(ctx, 2, page, &total)

			got, err := s.SearchClients(paginateCtx, nil, &ttnpb.SearchClientsRequest{})
			if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
				if page == 4 {
					a.So(got, should.HaveLength, 1)
				} else {
					a.So(got, should.HaveLength, 2)
				}
				for i, e := range got {
					a.So(e, should.Resemble, clients[i+2*int(page-1)].Ids)
				}
			}

			a.So(total, should.Equal, 7)
		}
	})

	t.Run("EndDevices_Paginated", func(t *T) {
		a, ctx := test.New(t)

		var total uint64
		for _, page := range []uint32{1, 2, 3, 4} {
			paginateCtx := store.WithPagination(ctx, 2, page, &total)

			got, err := s.SearchEndDevices(paginateCtx, &ttnpb.SearchEndDevicesRequest{
				ApplicationIds: applications[0].GetIds(),
			})
			if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
				if page == 4 {
					a.So(got, should.HaveLength, 1)
				} else {
					a.So(got, should.HaveLength, 2)
				}
				for i, e := range got {
					a.So(e, should.Resemble, endDevices[i+2*int(page-1)].Ids)
				}
			}

			a.So(total, should.Equal, 7)
		}
	})

	t.Run("Gateways_Paginated", func(t *T) {
		a, ctx := test.New(t)

		var total uint64
		for _, page := range []uint32{1, 2, 3, 4} {
			paginateCtx := store.WithPagination(ctx, 2, page, &total)

			got, err := s.SearchGateways(paginateCtx, nil, &ttnpb.SearchGatewaysRequest{})
			if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
				if page == 4 {
					a.So(got, should.HaveLength, 1)
				} else {
					a.So(got, should.HaveLength, 2)
				}
				for i, e := range got {
					a.So(e, should.Resemble, gateways[i+2*int(page-1)].Ids)
				}
			}

			a.So(total, should.Equal, 7)
		}
	})

	t.Run("Organizations_Paginated", func(t *T) {
		a, ctx := test.New(t)

		var total uint64
		for _, page := range []uint32{1, 2, 3, 4} {
			paginateCtx := store.WithPagination(ctx, 2, page, &total)

			got, err := s.SearchOrganizations(paginateCtx, nil, &ttnpb.SearchOrganizationsRequest{})
			if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
				if page == 4 {
					a.So(got, should.HaveLength, 1)
				} else {
					a.So(got, should.HaveLength, 2)
				}
				for i, e := range got {
					a.So(e, should.Resemble, organizations[i+2*int(page-1)].Ids)
				}
			}

			a.So(total, should.Equal, 7)
		}
	})

	t.Run("Users_Paginated", func(t *T) {
		a, ctx := test.New(t)

		var total uint64
		for _, page := range []uint32{1, 2, 3, 4} {
			paginateCtx := store.WithPagination(ctx, 2, page, &total)

			got, err := s.SearchUsers(paginateCtx, &ttnpb.SearchUsersRequest{})
			if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
				if page == 4 {
					a.So(got, should.HaveLength, 1)
				} else {
					a.So(got, should.HaveLength, 2)
				}
				for i, e := range got {
					a.So(e, should.Resemble, users[i+2*int(page-1)].Ids)
				}
			}

			a.So(total, should.Equal, 7)
		}
	})
}
