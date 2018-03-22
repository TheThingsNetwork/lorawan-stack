// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql_test

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	. "github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var client = &ttnpb.Client{
	ClientIdentifiers: ttnpb.ClientIdentifiers{ClientID: "test-client"},
	Secret:            "123456",
	RedirectURI:       "/oauth/callback",
	Grants:            []ttnpb.GrantType{ttnpb.GRANT_AUTHORIZATION_CODE, ttnpb.GRANT_PASSWORD},
	Rights:            []ttnpb.Right{ttnpb.RIGHT_APPLICATION_INFO},
	Creator:           bob.UserIdentifiers,
}

var clientSpecializer = func(base ttnpb.Client) store.Client {
	return &base
}

func TestClients(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	err := s.Clients.Create(client)
	a.So(err, should.BeNil)

	err = s.Clients.Create(client)
	a.So(err, should.NotBeNil)
	a.So(ErrClientIDTaken.Describes(err), should.BeTrue)

	found, err := s.Clients.GetByID(client.ClientIdentifiers, clientSpecializer)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeClientIgnoringAutoFields, client)

	clients, err := s.Clients.List(clientSpecializer)
	a.So(err, should.BeNil)
	if a.So(clients, should.HaveLength, 1) {
		a.So(clients[0], test.ShouldBeClientIgnoringAutoFields, client)
	}

	clients, err = s.Clients.ListByUser(bob.UserIdentifiers, clientSpecializer)
	a.So(err, should.BeNil)
	if a.So(clients, should.HaveLength, 1) {
		a.So(clients[0], test.ShouldBeClientIgnoringAutoFields, client)
	}

	client.Description = "Fancy Description"
	err = s.Clients.Update(client)
	a.So(err, should.BeNil)

	found, err = s.Clients.GetByID(client.ClientIdentifiers, clientSpecializer)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeClientIgnoringAutoFields, client)

	err = s.Clients.Delete(client.ClientIdentifiers)
	a.So(err, should.BeNil)

	found, err = s.Clients.GetByID(client.ClientIdentifiers, clientSpecializer)
	a.So(err, should.NotBeNil)
	a.So(ErrClientNotFound.Describes(err), should.BeTrue)
	a.So(found, should.BeNil)
}
