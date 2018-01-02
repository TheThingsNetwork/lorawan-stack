// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var clientFactory = func() types.Client {
	return &ttnpb.Client{}

}

func testClients() map[string]*ttnpb.Client {
	return map[string]*ttnpb.Client{
		"test-client": {
			ClientIdentifier: ttnpb.ClientIdentifier{"test-client"},
			Secret:           "123456",
			RedirectURI:      "/oauth/callback",
			Grants:           []ttnpb.GrantType{ttnpb.GRANT_AUTHORIZATION_CODE, ttnpb.GRANT_PASSWORD},
			Rights:           []ttnpb.Right{ttnpb.RIGHT_APPLICATION_INFO},
		},
		"foo-client": {
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
		a.So(ErrClientIDTaken.Describes(err), should.BeTrue)
	}
}

func TestClientUpdate(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	client := testClients()["test-client"]
	client.Description = "Fancy Description"

	err := s.Clients.Update(client)
	a.So(err, should.BeNil)

	found, err := s.Clients.GetByID(client.ClientID, clientFactory)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeClientIgnoringAutoFields, client)
}
