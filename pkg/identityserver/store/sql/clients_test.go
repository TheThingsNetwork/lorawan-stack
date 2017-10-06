// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func testClients() map[string]*ttnpb.Client {
	return map[string]*ttnpb.Client{
		"test-client": &ttnpb.Client{
			ClientIdentifier: ttnpb.ClientIdentifier{"test-client"},
			Secret:           "123456",
			RedirectURI:      "/oauth/callback",
			Grants:           []ttnpb.GrantType{ttnpb.GRANT_AUTHORIZATION_CODE, ttnpb.GRANT_PASSWORD},
			Rights:           []ttnpb.Right{ttnpb.RIGHT_APPLICATION_INFO},
		},
		"foo-client": &ttnpb.Client{
			ClientIdentifier: ttnpb.ClientIdentifier{"foo-client"},
			Secret:           "foofoofoo",
			RedirectURI:      "https://foo.bar/oauth/callback",
			Grants:           []ttnpb.GrantType{ttnpb.GRANT_AUTHORIZATION_CODE},
		},
	}
}

func TestClientCreate(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	clients := testClients()

	for _, client := range clients {
		err := s.Clients.Create(client)
		a.So(err, should.BeNil)
	}

	// Attempt to recreate them should throw an error
	for _, client := range clients {
		err := s.Clients.Create(client)
		a.So(err, should.NotBeNil)
		a.So(err.(errors.Error).Code(), should.Equal, 21)
		a.So(err.(errors.Error).Type(), should.Equal, errors.AlreadyExists)
	}
}

func TestClientCollaborators(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	user := testUsers()["alice"]
	client := testClients()["foo-client"]

	// check indeed that application has no collaborator
	{
		collaborators, err := s.Clients.ListCollaborators(client.ClientID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 0)
	}

	collaborator := ttnpb.Collaborator{
		UserIdentifier: ttnpb.UserIdentifier{user.UserID},
		Rights: []ttnpb.Right{
			ttnpb.Right(1),
			ttnpb.Right(2),
		},
	}

	// add one
	{
		err := s.Clients.SetCollaborator(client.ClientID, collaborator)
		a.So(err, should.BeNil)
	}

	// check that was added
	{
		collaborators, err := s.Clients.ListCollaborators(client.ClientID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 1)
		a.So(collaborators, should.Contain, collaborator)
	}

	// fetch applications where Alice is collaborator
	{
		clients, err := s.Clients.ListByUser(user.UserID)
		a.So(err, should.BeNil)
		if a.So(clients, should.HaveLength, 1) {
			a.So(clients[0].GetClient().ClientID, should.Equal, client.ClientID)
		}
	}

	// modify rights
	{
		collaborator.Rights = append(collaborator.Rights, ttnpb.Right(3))
		err := s.Clients.SetCollaborator(client.ClientID, collaborator)
		a.So(err, should.BeNil)

		collaborators, err := s.Clients.ListCollaborators(client.ClientID)
		a.So(err, should.BeNil)
		if a.So(collaborators, should.HaveLength, 1) {
			a.So(collaborators[0].Rights, should.Resemble, collaborator.Rights)
		}
	}

	// fetch user rights
	{
		rights, err := s.Clients.ListUserRights(client.ClientID, user.UserID)
		a.So(err, should.BeNil)
		if a.So(rights, should.HaveLength, 3) {
			a.So(rights, should.Resemble, collaborator.Rights)
		}
	}

	// remove collaborator
	{
		collaborator.Rights = []ttnpb.Right{}
		err := s.Clients.SetCollaborator(client.ClientID, collaborator)
		a.So(err, should.BeNil)

		collaborators, err := s.Clients.ListCollaborators(client.ClientID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 0)
	}
}

func TestClientUpdate(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	client := testClients()["test-client"]
	client.Description = "Fancy Description"

	err := s.Clients.Update(client)
	a.So(err, should.BeNil)

	found, err := s.Clients.GetByID(client.ClientID)
	a.So(err, should.BeNil)
	a.So(client, test.ShouldBeClientIgnoringAutoFields, found)
}

func TestClientManagement(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	client := testClients()["foo-client"]

	// label as official
	{
		err := s.Clients.SetClientOfficial(client.ClientID, true)
		a.So(err, should.BeNil)

		found, err := s.Clients.GetByID(client.ClientID)
		a.So(err, should.BeNil)
		a.So(found.GetClient().OfficialLabeled, should.BeTrue)
	}

	// mark as approved
	{
		err := s.Clients.SetClientState(client.ClientID, ttnpb.STATE_APPROVED)
		a.So(err, should.BeNil)

		found, err := s.Clients.GetByID(client.ClientID)
		a.So(err, should.BeNil)
		a.So(found.GetClient().State, should.Resemble, ttnpb.STATE_APPROVED)
	}
}

func TestClientArchive(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	client := testClients()["test-client"]

	err := s.Clients.Archive(client.ClientID)
	a.So(err, should.BeNil)

	found, err := s.Clients.GetByID(client.ClientID)
	a.So(err, should.BeNil)

	a.So(found.GetClient().ArchivedAt.IsZero(), should.BeFalse)
}
