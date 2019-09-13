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
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc"
)

func init() {
	// remove clients assigned to the user by the populator
	userID := paginationUser.UserIdentifiers
	for _, client := range population.Clients {
		for id, collaborators := range population.Memberships {
			if client.IDString() == id.IDString() {
				for i, collaborator := range collaborators {
					if collaborator.IDString() == userID.GetUserID() {
						collaborators = collaborators[:i+copy(collaborators[i:], collaborators[i+1:])]
					}
				}
			}
		}
	}

	// add deterministic number of clients
	for i := 0; i < 3; i++ {
		clientID := population.Clients[i].EntityIdentifiers()
		population.Memberships[clientID] = append(population.Memberships[clientID], &ttnpb.Collaborator{
			OrganizationOrUserIdentifiers: *paginationUser.OrganizationOrUserIdentifiers(),
			Rights:                        []ttnpb.Right{ttnpb.RIGHT_CLIENT_ALL},
		})
	}
}

func TestClientsPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewClientRegistryClient(cc)

		_, err := reg.Create(ctx, &ttnpb.CreateClientRequest{
			Client: ttnpb.Client{
				ClientIdentifiers: ttnpb.ClientIdentifiers{ClientID: "foo-cli"},
			},
			Collaborator: *ttnpb.UserIdentifiers{UserID: "foo-usr"}.OrganizationOrUserIdentifiers(),
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Get(ctx, &ttnpb.GetClientRequest{
			ClientIdentifiers: ttnpb.ClientIdentifiers{ClientID: "foo-cli"},
			FieldMask:         types.FieldMask{Paths: []string{"name"}},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsUnauthenticated(err), should.BeTrue)
		}

		listRes, err := reg.List(ctx, &ttnpb.ListClientsRequest{
			FieldMask: types.FieldMask{Paths: []string{"name"}},
		})

		a.So(err, should.BeNil)
		if a.So(listRes, should.NotBeNil) {
			a.So(listRes.Clients, should.BeEmpty)
		}

		_, err = reg.List(ctx, &ttnpb.ListClientsRequest{
			Collaborator: ttnpb.UserIdentifiers{UserID: "foo-usr"}.OrganizationOrUserIdentifiers(),
			FieldMask:    types.FieldMask{Paths: []string{"name"}},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Update(ctx, &ttnpb.UpdateClientRequest{
			Client: ttnpb.Client{
				ClientIdentifiers: ttnpb.ClientIdentifiers{ClientID: "foo-cli"},
				Name:              "Updated Name",
			},
			FieldMask: types.FieldMask{Paths: []string{"name"}},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Delete(ctx, &ttnpb.ClientIdentifiers{ClientID: "foo-cli"})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
	})
}

func TestClientsCRUD(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewClientRegistryClient(cc)

		userID, creds := population.Users[defaultUserIdx].UserIdentifiers, userCreds(defaultUserIdx)
		credsWithoutRights := userCreds(defaultUserIdx, "key without rights")

		created, err := reg.Create(ctx, &ttnpb.CreateClientRequest{
			Client: ttnpb.Client{
				ClientIdentifiers: ttnpb.ClientIdentifiers{ClientID: "foo"},
				Name:              "Foo Client",
			},
			Collaborator: *userID.OrganizationOrUserIdentifiers(),
		}, creds)

		a.So(err, should.BeNil)
		if a.So(created, should.NotBeNil) {
			a.So(created.Name, should.Equal, "Foo Client")
		}

		got, err := reg.Get(ctx, &ttnpb.GetClientRequest{
			ClientIdentifiers: created.ClientIdentifiers,
			FieldMask:         types.FieldMask{Paths: []string{"name"}},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.Name, should.Equal, created.Name)
		}

		got, err = reg.Get(ctx, &ttnpb.GetClientRequest{
			ClientIdentifiers: created.ClientIdentifiers,
			FieldMask:         types.FieldMask{Paths: []string{"ids"}},
		}, credsWithoutRights)

		a.So(err, should.BeNil)

		got, err = reg.Get(ctx, &ttnpb.GetClientRequest{
			ClientIdentifiers: created.ClientIdentifiers,
			FieldMask:         types.FieldMask{Paths: []string{"attributes"}},
		}, credsWithoutRights)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		updated, err := reg.Update(ctx, &ttnpb.UpdateClientRequest{
			Client: ttnpb.Client{
				ClientIdentifiers: created.ClientIdentifiers,
				Name:              "Updated Name",
			},
			FieldMask: types.FieldMask{Paths: []string{"name"}},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(updated, should.NotBeNil) {
			a.So(updated.Name, should.Equal, "Updated Name")
		}

		for _, collaborator := range []*ttnpb.OrganizationOrUserIdentifiers{nil, userID.OrganizationOrUserIdentifiers()} {
			list, err := reg.List(ctx, &ttnpb.ListClientsRequest{
				FieldMask:    types.FieldMask{Paths: []string{"name"}},
				Collaborator: collaborator,
			}, creds)

			a.So(err, should.BeNil)
			if a.So(list, should.NotBeNil) && a.So(list.Clients, should.NotBeEmpty) {
				var found bool
				for _, item := range list.Clients {
					if item.ClientIdentifiers == created.ClientIdentifiers {
						found = true
						a.So(item.Name, should.Equal, updated.Name)
					}
				}
				a.So(found, should.BeTrue)
			}
		}

		_, err = reg.Delete(ctx, &created.ClientIdentifiers, creds)

		a.So(err, should.BeNil)
	})
}

func TestClientsPagination(t *testing.T) {
	a := assertions.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := paginationUser.UserIdentifiers
		creds := userCreds(paginationUserIdx)

		reg := ttnpb.NewClientRegistryClient(cc)

		list, err := reg.List(test.Context(), &ttnpb.ListClientsRequest{
			FieldMask:    types.FieldMask{Paths: []string{"name"}},
			Collaborator: userID.OrganizationOrUserIdentifiers(),
			Limit:        2,
			Page:         1,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(list, should.NotBeNil) {
			a.So(list.Clients, should.HaveLength, 2)
		}

		list, err = reg.List(test.Context(), &ttnpb.ListClientsRequest{
			FieldMask:    types.FieldMask{Paths: []string{"name"}},
			Collaborator: userID.OrganizationOrUserIdentifiers(),
			Limit:        2,
			Page:         2,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(list, should.NotBeNil) {
			a.So(list.Clients, should.HaveLength, 1)
		}

		list, err = reg.List(test.Context(), &ttnpb.ListClientsRequest{
			FieldMask:    types.FieldMask{Paths: []string{"name"}},
			Collaborator: userID.OrganizationOrUserIdentifiers(),
			Limit:        2,
			Page:         3,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(list, should.NotBeNil) {
			a.So(list.Clients, should.BeEmpty)
		}
	})
}
