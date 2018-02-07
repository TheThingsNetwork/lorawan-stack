// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var clientSpecializer = func(base ttnpb.Client) store.Client {
	return &base
}

func testClients() map[string]*ttnpb.Client {
	return map[string]*ttnpb.Client{
		"test-client": {
			ClientIdentifier: ttnpb.ClientIdentifier{"test-client"},
			Secret:           "123456",
			RedirectURI:      "/oauth/callback",
			Grants:           []ttnpb.GrantType{ttnpb.GRANT_AUTHORIZATION_CODE, ttnpb.GRANT_PASSWORD},
			Rights:           []ttnpb.Right{ttnpb.RIGHT_APPLICATION_INFO},
			Creator:          ttnpb.UserIdentifier{"bob"},
		},
		"foo-client": {
			ClientIdentifier: ttnpb.ClientIdentifier{"foo-client"},
			Secret:           "foofoofoo",
			RedirectURI:      "https://foo.bar/oauth/callback",
			Grants:           []ttnpb.GrantType{ttnpb.GRANT_AUTHORIZATION_CODE},
			Rights:           []ttnpb.Right{ttnpb.RIGHT_USER_ADMIN, ttnpb.RIGHT_GATEWAY_INFO},
			Creator:          ttnpb.UserIdentifier{"bob"},
		},
	}
}

func testClientCreate(t testing.TB, s *Store) {
	a := assertions.New(t)

	clients := testClients()

	for _, client := range clients {
		err := s.Clients.Create(client)
		a.So(err, should.BeNil)
	}

	// Attempt to recreate them should throw an error
	for _, client := range clients {
		err := s.Clients.Create(client)
		a.So(err, should.NotBeNil)
		a.So(ErrClientIDTaken.Describes(err), should.BeTrue)
	}
}

func TestClientList(t *testing.T) {
	s := testStore(t, database)

	clients, err := s.Clients.List(clientSpecializer)
	testClientList(t, clients, err)

	clients, err = s.Clients.ListByUser(testUsers()["bob"].UserID, clientSpecializer)
	testClientList(t, clients, err)
}

func testClientList(t *testing.T, clients []store.Client, err error) {
	a := assertions.New(t)

	client1 := testClients()["test-client"]
	client2 := testClients()["foo-client"]

	a.So(err, should.BeNil)
	if a.So(clients, should.HaveLength, 2) {
		for _, client := range clients {
			switch client.GetClient().ClientID {
			case client1.ClientID:
				a.So(client, test.ShouldBeClientIgnoringAutoFields, client1)
			case client2.ClientID:
				a.So(client, test.ShouldBeClientIgnoringAutoFields, client2)
			}
		}
	}
}

func TestClientUpdate(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	client := testClients()["test-client"]
	client.Description = "Fancy Description"

	err := s.Clients.Update(client)
	a.So(err, should.BeNil)

	found, err := s.Clients.GetByID(client.ClientID, clientSpecializer)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeClientIgnoringAutoFields, client)
}

func TestClientDelete(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	userID := testUsers()["bob"].UserID
	clientID := "delete-test"

	testClientDeleteFeedDatabase(t, userID, clientID)

	err := s.Clients.Delete(clientID)
	a.So(err, should.BeNil)

	found, err := s.Clients.GetByID(clientID, clientSpecializer)
	a.So(err, should.NotBeNil)
	a.So(ErrClientNotFound.Describes(err), should.BeTrue)
	a.So(found, should.BeNil)
}

func testClientDeleteFeedDatabase(t *testing.T, userID, clientID string) {
	a := assertions.New(t)
	s := testStore(t, database)

	client := &ttnpb.Client{
		ClientIdentifier: ttnpb.ClientIdentifier{clientID},
		Creator:          ttnpb.UserIdentifier{userID},
	}
	err := s.Clients.Create(client)
	a.So(err, should.BeNil)

	oauth, ok := s.store().OAuth.(*OAuthStore)
	if a.So(ok, should.BeTrue) {
		err := oauth.saveAuthorizationCode(s.queryer(), &store.AuthorizationData{
			AuthorizationCode: "123",
			ClientID:          clientID,
			UserID:            userID,
		})
		a.So(err, should.BeNil)

		err = oauth.saveAccessToken(s.queryer(), &store.AccessData{
			AccessToken: "123",
			ClientID:    clientID,
			UserID:      userID,
		})
		a.So(err, should.BeNil)

		err = oauth.saveRefreshToken(s.queryer(), &store.RefreshData{
			RefreshToken: "123",
			ClientID:     clientID,
			UserID:       userID,
		})
		a.So(err, should.BeNil)
	}
}
