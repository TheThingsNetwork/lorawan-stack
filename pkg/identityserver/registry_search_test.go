// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package identityserver

import (
	"testing"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/storetest"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func TestRegistrySearch(t *testing.T) {
	t.Parallel()

	const (
		desc      = "some test description"
		descQuery = "test description"
	)

	p := &storetest.Population{}

	adminUsr := p.NewUser()
	adminUsr.Admin = true
	adminUsrKey, _ := p.NewAPIKey(adminUsr.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	adminUsrCreds := rpcCreds(adminUsrKey)

	for i := 0; i < 10; i++ {
		app := p.NewApplication(nil)
		if i < 5 {
			app.Description = desc
		}
	}
	for i := 0; i < 10; i++ {
		cli := p.NewClient(nil)
		if i < 5 {
			cli.Description = desc
		}
	}
	for i := 0; i < 10; i++ {
		gtw := p.NewGateway(nil)
		if i < 5 {
			gtw.Description = desc
		}
	}
	for i := 0; i < 10; i++ {
		org := p.NewOrganization(nil)
		if i < 5 {
			org.Description = desc
		}
	}
	for i := 0; i < 10; i++ {
		usr := p.NewUser()
		if i < 5 {
			usr.Description = desc
		}
	}

	a, ctx := test.New(t)

	testWithIdentityServer(t, func(_ *IdentityServer, cc *grpc.ClientConn) {
		cli := ttnpb.NewEntityRegistrySearchClient(cc)

		apps, err := cli.SearchApplications(ctx, &ttnpb.SearchApplicationsRequest{
			DescriptionContains: descQuery,
			FieldMask:           ttnpb.FieldMask("ids"),
		}, adminUsrCreds)

		a.So(err, should.BeNil)
		if a.So(apps, should.NotBeNil) {
			a.So(apps.Applications, should.HaveLength, 5)
		}

		clis, err := cli.SearchClients(ctx, &ttnpb.SearchClientsRequest{
			DescriptionContains: descQuery,
			FieldMask:           ttnpb.FieldMask("ids"),
		}, adminUsrCreds)

		a.So(err, should.BeNil)
		if a.So(clis, should.NotBeNil) {
			a.So(clis.Clients, should.HaveLength, 5)
		}

		gtws, err := cli.SearchGateways(ctx, &ttnpb.SearchGatewaysRequest{
			DescriptionContains: descQuery,
			FieldMask:           ttnpb.FieldMask("ids"),
		}, adminUsrCreds)

		a.So(err, should.BeNil)
		if a.So(gtws, should.NotBeNil) {
			a.So(gtws.Gateways, should.HaveLength, 5)
		}

		orgs, err := cli.SearchOrganizations(ctx, &ttnpb.SearchOrganizationsRequest{
			DescriptionContains: descQuery,
			FieldMask:           ttnpb.FieldMask("ids"),
		}, adminUsrCreds)

		a.So(err, should.BeNil)
		if a.So(orgs, should.NotBeNil) {
			a.So(orgs.Organizations, should.HaveLength, 5)
		}

		usrs, err := cli.SearchUsers(ctx, &ttnpb.SearchUsersRequest{
			DescriptionContains: descQuery,
			FieldMask:           ttnpb.FieldMask("ids"),
		}, adminUsrCreds)

		a.So(err, should.BeNil)
		if a.So(usrs, should.NotBeNil) {
			a.So(usrs.Users, should.HaveLength, 5)
		}
	}, withPrivateTestDatabase(p))
}

// TestRegistrySearchDeletedEntities validates that deleted entities are listed properly when providing the adequate
// values on the request body.
func TestRegistrySearchDeletedEntities(t *testing.T) { // nolint:gocyclo
	t.Parallel()

	p := &storetest.Population{}

	usr := p.NewUser()
	usrKey, _ := p.NewAPIKey(usr.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	usrCreds := rpcCreds(usrKey)

	adminUsr := p.NewUser()
	adminUsr.Admin = true
	adminUsrKey, _ := p.NewAPIKey(adminUsr.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	adminUsrCreds := rpcCreds(adminUsrKey)

	const (
		notDeletedAmount = 6
		deletedAmount    = 4
	)

	apps := make([]*ttnpb.Application, 10)
	clis := make([]*ttnpb.Client, 10)
	gtws := make([]*ttnpb.Gateway, 10)
	orgs := make([]*ttnpb.Organization, 10)
	usrs := make([]*ttnpb.User, 10)

	// Making an uneven number of deleted entities just to avoid having a flaky tests.
	for i := 0; i < 10; i++ {
		apps[i] = p.NewApplication(usr.GetOrganizationOrUserIdentifiers())
		clis[i] = p.NewClient(usr.GetOrganizationOrUserIdentifiers())
		gtws[i] = p.NewGateway(usr.GetOrganizationOrUserIdentifiers())
		orgs[i] = p.NewOrganization(usr.GetOrganizationOrUserIdentifiers())
		usrs[i] = p.NewUser()
	}

	noOwnerApp := p.NewApplication(nil)
	noOwnerClient := p.NewClient(nil)
	noOwnerGtw := p.NewGateway(nil)
	noOwnerOrg := p.NewOrganization(nil)

	testWithIdentityServer(t, func(_ *IdentityServer, cc *grpc.ClientConn) {
		cli := ttnpb.NewEntityRegistrySearchClient(cc)
		appReg := ttnpb.NewApplicationRegistryClient(cc)
		cliReg := ttnpb.NewClientRegistryClient(cc)
		gtwReg := ttnpb.NewGatewayRegistryClient(cc)
		orgReg := ttnpb.NewOrganizationRegistryClient(cc)

		a, ctx := test.New(t)

		// Delete some entities associated with the non admin user.
		for i := 0; i < deletedAmount; i++ {
			_, err := appReg.Delete(ctx, apps[i].Ids, usrCreds)
			a.So(err, should.BeNil)
			_, err = cliReg.Delete(ctx, clis[i].Ids, usrCreds)
			a.So(err, should.BeNil)
			_, err = gtwReg.Delete(ctx, gtws[i].Ids, usrCreds)
			a.So(err, should.BeNil)
			_, err = orgReg.Delete(ctx, orgs[i].Ids, usrCreds)
			a.So(err, should.BeNil)
		}
		// Delete some entities not associated with `usr`, this should not reflect on the results of the test below.
		_, err := appReg.Delete(ctx, noOwnerApp.Ids, adminUsrCreds)
		a.So(err, should.BeNil)
		_, err = cliReg.Delete(ctx, noOwnerClient.Ids, adminUsrCreds)
		a.So(err, should.BeNil)
		_, err = gtwReg.Delete(ctx, noOwnerGtw.Ids, adminUsrCreds)
		a.So(err, should.BeNil)
		_, err = orgReg.Delete(ctx, noOwnerOrg.Ids, adminUsrCreds)
		a.So(err, should.BeNil)

		t.Run("Applications", func(t *testing.T) { // nolint:paralleltest
			t.Run("Not Deleted", func(t *testing.T) { // nolint:paralleltest
				a, ctx := test.New(t)
				got, err := cli.SearchApplications(ctx, &ttnpb.SearchApplicationsRequest{
					FieldMask: ttnpb.FieldMask("ids"),
				}, usrCreds)
				if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
					a.So(got.Applications, should.HaveLength, notDeletedAmount)
				}
			})
			t.Run("Deleted", func(t *testing.T) { // nolint:paralleltest
				a, ctx := test.New(t)
				got, err := cli.SearchApplications(ctx, &ttnpb.SearchApplicationsRequest{
					Deleted:   true,
					FieldMask: ttnpb.FieldMask("ids"),
				}, usrCreds)
				if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
					a.So(got.Applications, should.HaveLength, deletedAmount)
				}
			})
		})
		t.Run("Clients", func(t *testing.T) { // nolint:paralleltest
			t.Run("Not Deleted", func(t *testing.T) { // nolint:paralleltest
				a, ctx := test.New(t)
				got, err := cli.SearchClients(ctx, &ttnpb.SearchClientsRequest{
					FieldMask: ttnpb.FieldMask("ids"),
				}, usrCreds)
				if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
					a.So(got.Clients, should.HaveLength, notDeletedAmount)
				}
			})
			t.Run("Deleted", func(t *testing.T) { // nolint:paralleltest
				a, ctx := test.New(t)
				got, err := cli.SearchClients(ctx, &ttnpb.SearchClientsRequest{
					Deleted:   true,
					FieldMask: ttnpb.FieldMask("ids"),
				}, usrCreds)
				if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
					a.So(got.Clients, should.HaveLength, deletedAmount)
				}
			})
		})
		t.Run("Gateways", func(t *testing.T) { // nolint:paralleltest
			t.Run("Not Deleted", func(t *testing.T) { // nolint:paralleltest
				a, ctx := test.New(t)
				got, err := cli.SearchGateways(ctx, &ttnpb.SearchGatewaysRequest{
					FieldMask: ttnpb.FieldMask("ids"),
				}, usrCreds)
				if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
					a.So(got.Gateways, should.HaveLength, notDeletedAmount)
				}
			})
			t.Run("Deleted", func(t *testing.T) { // nolint:paralleltest
				a, ctx := test.New(t)
				got, err := cli.SearchGateways(ctx, &ttnpb.SearchGatewaysRequest{
					Deleted:   true,
					FieldMask: ttnpb.FieldMask("ids"),
				}, usrCreds)
				if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
					a.So(got.Gateways, should.HaveLength, deletedAmount)
				}
			})
		})
		t.Run("Organizations", func(t *testing.T) { // nolint:paralleltest
			t.Run("Not Deleted", func(t *testing.T) { // nolint:paralleltest
				a, ctx := test.New(t)
				got, err := cli.SearchOrganizations(ctx, &ttnpb.SearchOrganizationsRequest{
					FieldMask: ttnpb.FieldMask("ids"),
				}, usrCreds)
				if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
					a.So(got.Organizations, should.HaveLength, notDeletedAmount)
				}
			})
			t.Run("Deleted", func(t *testing.T) { // nolint:paralleltest
				a, ctx := test.New(t)
				got, err := cli.SearchOrganizations(ctx, &ttnpb.SearchOrganizationsRequest{
					Deleted:   true,
					FieldMask: ttnpb.FieldMask("ids"),
				}, usrCreds)
				if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
					a.So(got.Organizations, should.HaveLength, deletedAmount)
				}
			})
		})

		// Admin only operations
		usrReg := ttnpb.NewUserRegistryClient(cc)
		// Delete some users.
		for i := 0; i < deletedAmount; i++ {
			_, err := usrReg.Delete(ctx, usrs[i].Ids, adminUsrCreds)
			a.So(err, should.BeNil)
		}
		t.Run("Users", func(t *testing.T) { // nolint: paralleltest
			t.Run("Not Deleted", func(t *testing.T) { // nolint:paralleltest
				a, ctx := test.New(t)
				got, err := cli.SearchUsers(ctx, &ttnpb.SearchUsersRequest{}, usrCreds)
				a.So(got, should.BeNil)
				a.So(errors.IsPermissionDenied(err), should.BeTrue)

				got, err = cli.SearchUsers(ctx, &ttnpb.SearchUsersRequest{
					FieldMask: ttnpb.FieldMask("ids"),
				}, adminUsrCreds)
				if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
					// The `+2` refers to the two non-deleted users created at the beginning of the test.
					a.So(got.Users, should.HaveLength, (notDeletedAmount + 2))
				}
			})
			t.Run("Deleted", func(t *testing.T) { // nolint:paralleltest
				a, ctx := test.New(t)
				got, err := cli.SearchUsers(ctx, &ttnpb.SearchUsersRequest{Deleted: true}, usrCreds)
				a.So(got, should.BeNil)
				a.So(errors.IsPermissionDenied(err), should.BeTrue)

				got, err = cli.SearchUsers(ctx, &ttnpb.SearchUsersRequest{
					Deleted:   true,
					FieldMask: ttnpb.FieldMask("ids"),
				}, adminUsrCreds)
				if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
					a.So(got.Users, should.HaveLength, deletedAmount)
				}
			})
			t.Run("Read ContactInfo", func(t *testing.T) { // nolint:paralleltest
				a, ctx := test.New(t)
				got, err := cli.SearchUsers(ctx, &ttnpb.SearchUsersRequest{
					FieldMask: ttnpb.FieldMask("contact_info"),
				}, adminUsrCreds)
				if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
					// The `+2` refers to the two non-deleted users created at the beginning of the test.
					a.So(got.Users, should.HaveLength, (notDeletedAmount + 2))
					for _, user := range got.Users {
						a.So(user.ContactInfo, should.HaveLength, 1)
						a.So(user.ContactInfo[0].Value, should.Equal, user.Ids.UserId+"@example.com")
					}
				}
			})

		})
	}, withPrivateTestDatabase(p))
}
