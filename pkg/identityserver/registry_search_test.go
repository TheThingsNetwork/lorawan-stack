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
