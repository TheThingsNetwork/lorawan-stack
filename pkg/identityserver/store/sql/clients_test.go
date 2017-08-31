// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/utils"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func testClients() map[string]*types.DefaultClient {
	return map[string]*types.DefaultClient{
		"test-client": &types.DefaultClient{
			ID:     "test-client",
			Secret: "123456",
			URI:    "/oauth/callback",
			Grants: types.Grants{Password: true, RefreshToken: true},
			Scope:  types.Scopes{Application: true},
		},
		"foo-client": &types.DefaultClient{
			ID:     "foo-client",
			Secret: "foofoofoo",
			URI:    "https://foo.bar/oauth/callback",
			Grants: types.Grants{Password: true, RefreshToken: true},
			Scope:  types.Scopes{Application: true, Profile: true},
		},
	}
}

func TestClientCreate(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	clients := testClients()

	for _, client := range clients {
		created, err := s.Clients.Create(client)
		a.So(err, should.BeNil)
		a.So(created, test.ShouldBeClientIgnoringAutoFields, client)
	}

	// recreating them should result in error
	for _, client := range clients {
		_, err := s.Clients.Create(client)
		a.So(err, should.NotBeNil)
		a.So(err.Error(), should.Equal, ErrClientIDTaken.Error())
	}
}

func TestClientCollaborators(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	user := testUsers()["alice"]
	client := testClients()["test-client"]
	rights := []types.Right{types.ClientDeleteRight}

	collaborator := utils.Collaborator(user.Username, rights)

	// add collaborator
	{
		err := s.Clients.AddCollaborator(client.ID, collaborator)
		a.So(err, should.BeNil)
	}

	// fetch client collaborators
	{
		collaborators, err := s.Clients.Collaborators(client.ID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 1)
		if len(collaborators) > 0 {
			a.So(collaborators[0], should.Resemble, collaborator)
		}
	}

	// find which components alice is collaborator
	{
		clients, err := s.Clients.FindByUser(user.Username)
		a.So(err, should.BeNil)
		a.So(clients, should.HaveLength, 1)
		if len(clients) > 0 {
			a.So(clients[0], test.ShouldBeClientIgnoringAutoFields, client)
		}
	}

	// right to be granted and revoked
	right := types.ClientSettingsRight

	// grant a right
	{
		err := s.Clients.GrantRight(client.ID, user.Username, right)
		a.So(err, should.BeNil)

		rights, err := s.Clients.UserRights(client.ID, user.Username)
		a.So(err, should.BeNil)
		a.So(rights, should.HaveLength, 2)
		if len(rights) > 0 {
			a.So(rights, should.Contain, right)
		}
	}

	// revoke a right
	{
		err := s.Clients.RevokeRight(client.ID, user.Username, right)
		a.So(err, should.BeNil)

		rights, err := s.Clients.UserRights(client.ID, user.Username)
		a.So(err, should.BeNil)
		a.So(rights, should.HaveLength, 1)
		if len(rights) > 0 {
			a.So(rights, should.NotContain, right)
		}
	}

	// delete collaborator
	{
		err := s.Clients.RemoveCollaborator(client.ID, user.Username)
		a.So(err, should.BeNil)

		collaborators, err := s.Clients.Collaborators(client.ID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 0)
	}
}

func TestClientFind(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	client := testClients()["test-client"]

	// find by id
	{
		found, err := s.Clients.FindByID(client.ID)
		a.So(err, should.BeNil)
		a.So(client, test.ShouldBeClientIgnoringAutoFields, found)
	}
}

func TestClientManagement(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	client := testClients()["foo-client"]

	// archive
	{
		err := s.Clients.Archive(client.ID)
		a.So(err, should.BeNil)

		found, err := s.Clients.FindByID(client.ID)
		a.So(err, should.BeNil)
		a.So(found.GetClient().Archived, should.NotBeNil)
	}

	// approve client
	{
		err := s.Clients.Approve(client.ID)
		a.So(err, should.BeNil)

		found, err := s.Clients.FindByID(client.ID)
		a.So(err, should.BeNil)
		a.So(found.GetClient().State, should.Resemble, types.ApprovedClient)
	}

	// check that 3 previous operations were reflected in the database
	{
		err := s.Clients.Reject(client.ID)
		a.So(err, should.BeNil)

		found, err := s.Clients.FindByID(client.ID)
		a.So(err, should.BeNil)
		a.So(found.GetClient().State, should.Resemble, types.RejectedClient)
	}

	// delete
	{
		err := s.Clients.Delete(client.ID)
		a.So(err, should.BeNil)

		_, err = s.Clients.FindByID(client.ID)
		a.So(err, should.NotBeNil)
		a.So(err.Error(), should.Equal, ErrClientNotFound.Error())
	}
}
