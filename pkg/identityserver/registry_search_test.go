// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

	"github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc"
)

func TestRegistrySearch(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		creds := userCreds(adminUserIdx)

		cli := ttnpb.NewEntityRegistrySearchClient(cc)

		apps, err := cli.SearchApplications(ctx, &ttnpb.SearchEntitiesRequest{
			NameContains: "%",
			FieldMask:    types.FieldMask{Paths: []string{"ids"}},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(apps, should.NotBeNil) {
			a.So(apps.Applications, should.NotBeEmpty)
		}

		clis, err := cli.SearchClients(ctx, &ttnpb.SearchEntitiesRequest{
			NameContains: "%",
			FieldMask:    types.FieldMask{Paths: []string{"ids"}},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(clis, should.NotBeNil) {
			a.So(clis.Clients, should.NotBeEmpty)
		}

		gtws, err := cli.SearchGateways(ctx, &ttnpb.SearchEntitiesRequest{
			NameContains: "%",
			FieldMask:    types.FieldMask{Paths: []string{"ids"}},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(gtws, should.NotBeNil) {
			a.So(gtws.Gateways, should.NotBeEmpty)
		}

		orgs, err := cli.SearchOrganizations(ctx, &ttnpb.SearchEntitiesRequest{
			NameContains: "%",
			FieldMask:    types.FieldMask{Paths: []string{"ids"}},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(orgs, should.NotBeNil) {
			a.So(orgs.Organizations, should.NotBeEmpty)
		}

		usrs, err := cli.SearchUsers(ctx, &ttnpb.SearchEntitiesRequest{
			NameContains: "%",
			FieldMask:    types.FieldMask{Paths: []string{"ids"}},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(usrs, should.NotBeNil) {
			a.So(usrs.Users, should.NotBeEmpty)
		}
	})
}
