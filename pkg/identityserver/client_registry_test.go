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

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"google.golang.org/grpc"
)

func init() {
	// remove clients assigned to the user by the populator
	userID := paginationUser.GetIds()
	for _, client := range population.Clients {
		for id, collaborators := range population.Memberships {
			if client.IDString() == id.IDString() {
				for i, collaborator := range collaborators {
					if collaborator.IDString() == userID.GetUserId() {
						collaborators = collaborators[:i+copy(collaborators[i:], collaborators[i+1:])]
					}
				}
			}
		}
	}

	// add deterministic number of clients
	for i := 0; i < 3; i++ {
		clientID := population.Clients[i].GetEntityIdentifiers()
		population.Memberships[clientID] = append(population.Memberships[clientID], &ttnpb.Collaborator{
			Ids:    paginationUser.OrganizationOrUserIdentifiers(),
			Rights: []ttnpb.Right{ttnpb.Right_RIGHT_CLIENT_ALL},
		})
	}
}

func TestClientsPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewClientRegistryClient(cc)

		_, err := reg.Create(ctx, &ttnpb.CreateClientRequest{
			Client: &ttnpb.Client{
				Ids: &ttnpb.ClientIdentifiers{ClientId: "foo-cli"},
			},
			Collaborator: ttnpb.UserIdentifiers{UserId: "foo-usr"}.OrganizationOrUserIdentifiers(),
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Get(ctx, &ttnpb.GetClientRequest{
			ClientIds: &ttnpb.ClientIdentifiers{ClientId: "foo-cli"},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"name"}},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsUnauthenticated(err), should.BeTrue)
		}

		listRes, err := reg.List(ctx, &ttnpb.ListClientsRequest{
			FieldMask: &pbtypes.FieldMask{Paths: []string{"name"}},
		})

		a.So(err, should.BeNil)
		if a.So(listRes, should.NotBeNil) {
			a.So(listRes.Clients, should.BeEmpty)
		}

		_, err = reg.List(ctx, &ttnpb.ListClientsRequest{
			Collaborator: ttnpb.UserIdentifiers{UserId: "foo-usr"}.OrganizationOrUserIdentifiers(),
			FieldMask:    &pbtypes.FieldMask{Paths: []string{"name"}},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Update(ctx, &ttnpb.UpdateClientRequest{
			Client: &ttnpb.Client{
				Ids:  &ttnpb.ClientIdentifiers{ClientId: "foo-cli"},
				Name: "Updated Name",
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"name"}},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Delete(ctx, &ttnpb.ClientIdentifiers{ClientId: "foo-cli"})

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

		userID, creds := population.Users[defaultUserIdx].GetIds(), userCreds(defaultUserIdx)
		credsWithoutRights := userCreds(defaultUserIdx, "key without rights")

		is.config.UserRights.CreateClients = false

		_, err := reg.Create(ctx, &ttnpb.CreateClientRequest{
			Client: &ttnpb.Client{
				Ids:  &ttnpb.ClientIdentifiers{ClientId: "foo"},
				Name: "Foo Client",
			},
			Collaborator: userID.GetOrganizationOrUserIdentifiers(),
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		is.config.UserRights.CreateClients = true

		created, err := reg.Create(ctx, &ttnpb.CreateClientRequest{
			Client: &ttnpb.Client{
				Ids:  &ttnpb.ClientIdentifiers{ClientId: "foo"},
				Name: "Foo Client",
			},
			Collaborator: userID.OrganizationOrUserIdentifiers(),
		}, creds)

		a.So(err, should.BeNil)
		if a.So(created, should.NotBeNil) {
			a.So(created.Name, should.Equal, "Foo Client")
		}

		got, err := reg.Get(ctx, &ttnpb.GetClientRequest{
			ClientIds: created.GetIds(),
			FieldMask: &pbtypes.FieldMask{Paths: []string{"name"}},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.Name, should.Equal, created.Name)
		}

		got, err = reg.Get(ctx, &ttnpb.GetClientRequest{
			ClientIds: created.GetIds(),
			FieldMask: &pbtypes.FieldMask{Paths: []string{"ids"}},
		}, credsWithoutRights)

		a.So(err, should.BeNil)

		got, err = reg.Get(ctx, &ttnpb.GetClientRequest{
			ClientIds: created.GetIds(),
			FieldMask: &pbtypes.FieldMask{Paths: []string{"attributes"}},
		}, credsWithoutRights)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		updated, err := reg.Update(ctx, &ttnpb.UpdateClientRequest{
			Client: &ttnpb.Client{
				Ids:  created.GetIds(),
				Name: "Updated Name",
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"name"}},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(updated, should.NotBeNil) {
			a.So(updated.Name, should.Equal, "Updated Name")
		}

		updated, err = reg.Update(ctx, &ttnpb.UpdateClientRequest{
			Client: &ttnpb.Client{
				Ids:              created.GetIds(),
				State:            ttnpb.State_STATE_FLAGGED,
				StateDescription: "something is wrong",
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"state", "state_description"}},
		}, userCreds(adminUserIdx))

		a.So(err, should.BeNil)
		if a.So(updated, should.NotBeNil) {
			a.So(updated.State, should.Equal, ttnpb.State_STATE_FLAGGED)
			a.So(updated.StateDescription, should.Equal, "something is wrong")
		}

		updated, err = reg.Update(ctx, &ttnpb.UpdateClientRequest{
			Client: &ttnpb.Client{
				Ids:   created.GetIds(),
				State: ttnpb.State_STATE_APPROVED,
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"state"}},
		}, userCreds(adminUserIdx))

		a.So(err, should.BeNil)
		if a.So(updated, should.NotBeNil) {
			a.So(updated.State, should.Equal, ttnpb.State_STATE_APPROVED)
		}

		got, err = reg.Get(ctx, &ttnpb.GetClientRequest{
			ClientIds: created.GetIds(),
			FieldMask: &pbtypes.FieldMask{Paths: []string{"state", "state_description"}},
		}, creds)

		if a.So(err, should.BeNil) {
			a.So(got.State, should.Equal, ttnpb.State_STATE_APPROVED)
			a.So(got.StateDescription, should.Equal, "")
		}

		for _, collaborator := range []*ttnpb.OrganizationOrUserIdentifiers{nil, userID.OrganizationOrUserIdentifiers()} {
			list, err := reg.List(ctx, &ttnpb.ListClientsRequest{
				FieldMask:    &pbtypes.FieldMask{Paths: []string{"name"}},
				Collaborator: collaborator,
			}, creds)

			a.So(err, should.BeNil)
			if a.So(list, should.NotBeNil) && a.So(list.Clients, should.NotBeEmpty) {
				var found bool
				for _, item := range list.Clients {
					if item.GetIds().GetClientId() == created.GetIds().GetClientId() {
						found = true
						a.So(item.Name, should.Equal, "Updated Name")
					}
				}
				a.So(found, should.BeTrue)
			}
		}

		_, err = reg.Delete(ctx, created.GetIds(), creds)
		a.So(err, should.BeNil)

		_, err = reg.Purge(ctx, created.GetIds(), creds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Purge(ctx, created.GetIds(), userCreds(adminUserIdx))
		a.So(err, should.BeNil)
	})
}

func TestClientsPagination(t *testing.T) {
	a := assertions.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := paginationUser.GetIds()
		creds := userCreds(paginationUserIdx)

		reg := ttnpb.NewClientRegistryClient(cc)

		list, err := reg.List(test.Context(), &ttnpb.ListClientsRequest{
			FieldMask:    &pbtypes.FieldMask{Paths: []string{"name"}},
			Collaborator: userID.OrganizationOrUserIdentifiers(),
			Limit:        2,
			Page:         1,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(list, should.NotBeNil) {
			a.So(list.Clients, should.HaveLength, 2)
		}

		list, err = reg.List(test.Context(), &ttnpb.ListClientsRequest{
			FieldMask:    &pbtypes.FieldMask{Paths: []string{"name"}},
			Collaborator: userID.OrganizationOrUserIdentifiers(),
			Limit:        2,
			Page:         2,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(list, should.NotBeNil) {
			a.So(list.Clients, should.HaveLength, 1)
		}

		list, err = reg.List(test.Context(), &ttnpb.ListClientsRequest{
			FieldMask:    &pbtypes.FieldMask{Paths: []string{"name"}},
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
