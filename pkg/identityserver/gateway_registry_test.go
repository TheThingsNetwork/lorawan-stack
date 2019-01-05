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
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc"
)

func init() {
	// remove gateways assigned to the user by the populator
	userID := paginationUser.UserIdentifiers
	for _, gw := range population.Gateways {
		for id, collaborators := range population.Memberships {
			if gw.EntityIdentifiers().IDString() == id.IDString() {
				for i, collaborator := range collaborators {
					if collaborator.EntityIdentifiers().IDString() == userID.GetUserID() {
						collaborators = collaborators[:i+copy(collaborators[i:], collaborators[i+1:])]
					}
				}
			}
		}
	}

	// add deterministic number of gateways
	for i := 0; i < 3; i++ {
		gatewayID := population.Gateways[i].EntityIdentifiers()
		population.Memberships[gatewayID] = append(population.Memberships[gatewayID], &ttnpb.Collaborator{
			OrganizationOrUserIdentifiers: *paginationUser.OrganizationOrUserIdentifiers(),
			Rights:                        []ttnpb.Right{ttnpb.RIGHT_GATEWAY_ALL},
		})
	}
}

func TestGatewaysPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewGatewayRegistryClient(cc)

		_, err := reg.Create(ctx, &ttnpb.CreateGatewayRequest{
			Gateway: ttnpb.Gateway{
				GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "foo-gtw"},
			},
			Collaborator: *ttnpb.UserIdentifiers{UserID: "foo-usr"}.OrganizationOrUserIdentifiers(),
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Get(ctx, &ttnpb.GetGatewayRequest{
			GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "foo-gtw"},
			FieldMask:          types.FieldMask{Paths: []string{"name"}},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsUnauthenticated(err), should.BeTrue)
		}

		listRes, err := reg.List(ctx, &ttnpb.ListGatewaysRequest{
			FieldMask: types.FieldMask{Paths: []string{"name"}},
		})

		a.So(err, should.BeNil)
		a.So(listRes.Gateways, should.BeEmpty)

		_, err = reg.List(ctx, &ttnpb.ListGatewaysRequest{
			Collaborator: ttnpb.UserIdentifiers{UserID: "foo-usr"}.OrganizationOrUserIdentifiers(),
			FieldMask:    types.FieldMask{Paths: []string{"name"}},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Update(ctx, &ttnpb.UpdateGatewayRequest{
			Gateway: ttnpb.Gateway{
				GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "foo-gtw"},
				Name:               "Updated Name",
			},
			FieldMask: types.FieldMask{Paths: []string{"name"}},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Delete(ctx, &ttnpb.GatewayIdentifiers{GatewayID: "foo-gtw"})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
	})
}

func TestGatewaysCRUD(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewGatewayRegistryClient(cc)

		userID, creds := population.Users[defaultUserIdx].UserIdentifiers, userCreds(defaultUserIdx)
		credsWithoutRights := userCreds(defaultUserIdx, "key without rights")

		created, err := reg.Create(ctx, &ttnpb.CreateGatewayRequest{
			Gateway: ttnpb.Gateway{
				GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "foo"},
				Name:               "Foo Gateway",
			},
			Collaborator: *userID.OrganizationOrUserIdentifiers(),
		}, creds)

		a.So(err, should.BeNil)
		a.So(created.Name, should.Equal, "Foo Gateway")

		got, err := reg.Get(ctx, &ttnpb.GetGatewayRequest{
			GatewayIdentifiers: created.GatewayIdentifiers,
			FieldMask:          types.FieldMask{Paths: []string{"name"}},
		}, creds)

		a.So(err, should.BeNil)
		a.So(got.Name, should.Equal, created.Name)

		got, err = reg.Get(ctx, &ttnpb.GetGatewayRequest{
			GatewayIdentifiers: created.GatewayIdentifiers,
			FieldMask:          types.FieldMask{Paths: []string{"ids"}},
		}, credsWithoutRights)

		a.So(err, should.BeNil)

		got, err = reg.Get(ctx, &ttnpb.GetGatewayRequest{
			GatewayIdentifiers: created.GatewayIdentifiers,
			FieldMask:          types.FieldMask{Paths: []string{"attributes"}},
		}, credsWithoutRights)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		updated, err := reg.Update(ctx, &ttnpb.UpdateGatewayRequest{
			Gateway: ttnpb.Gateway{
				GatewayIdentifiers: created.GatewayIdentifiers,
				Name:               "Updated Name",
			},
			FieldMask: types.FieldMask{Paths: []string{"name"}},
		}, creds)

		a.So(err, should.BeNil)
		a.So(updated.Name, should.Equal, "Updated Name")

		for _, collaborator := range []*ttnpb.OrganizationOrUserIdentifiers{nil, userID.OrganizationOrUserIdentifiers()} {
			list, err := reg.List(ctx, &ttnpb.ListGatewaysRequest{
				FieldMask:    types.FieldMask{Paths: []string{"name"}},
				Collaborator: collaborator,
			}, creds)
			a.So(err, should.BeNil)
			if a.So(list.Gateways, should.NotBeEmpty) {
				var found bool
				for _, item := range list.Gateways {
					if item.GatewayIdentifiers == created.GatewayIdentifiers {
						found = true
						a.So(item.Name, should.Equal, updated.Name)
					}
				}
				a.So(found, should.BeTrue)
			}
		}

		_, err = reg.Delete(ctx, &created.GatewayIdentifiers, creds)
		a.So(err, should.BeNil)

	})
}

func TestGatewaysPagination(t *testing.T) {
	a := assertions.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := paginationUser.UserIdentifiers
		creds := userCreds(paginationUserIdx)

		reg := ttnpb.NewGatewayRegistryClient(cc)

		md := rpcmetadata.MD{Limit: 2, Page: 1}
		ctx := md.ToOutgoingContext(test.Context())

		list, err := reg.List(ctx, &ttnpb.ListGatewaysRequest{
			FieldMask:    types.FieldMask{Paths: []string{"name"}},
			Collaborator: userID.OrganizationOrUserIdentifiers(),
		}, creds)

		a.So(list, should.NotBeNil)
		a.So(list.Gateways, should.HaveLength, 2)
		a.So(err, should.BeNil)

		md = rpcmetadata.MD{Limit: 2, Page: 2}
		ctx = md.ToOutgoingContext(test.Context())

		list, err = reg.List(ctx, &ttnpb.ListGatewaysRequest{
			FieldMask:    types.FieldMask{Paths: []string{"name"}},
			Collaborator: userID.OrganizationOrUserIdentifiers(),
		}, creds)

		a.So(list, should.NotBeNil)
		a.So(list.Gateways, should.HaveLength, 1)
		a.So(err, should.BeNil)

		md = rpcmetadata.MD{Limit: 2, Page: 3}
		ctx = md.ToOutgoingContext(test.Context())

		list, err = reg.List(ctx, &ttnpb.ListGatewaysRequest{
			FieldMask:    types.FieldMask{Paths: []string{"name"}},
			Collaborator: userID.OrganizationOrUserIdentifiers(),
		}, creds)

		a.So(list, should.NotBeNil)
		a.So(list.Gateways, should.BeEmpty)
		a.So(err, should.BeNil)
	})
}
