// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package sql_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	. "go.thethings.network/lorawan-stack/pkg/identityserver/store/sql"
	"go.thethings.network/lorawan-stack/pkg/identityserver/test"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var client = &ttnpb.Client{
	ClientIdentifiers: ttnpb.ClientIdentifiers{ClientID: "test-client"},
	Secret:            "123456",
	RedirectURI:       "/oauth/callback",
	Grants:            []ttnpb.GrantType{ttnpb.GRANT_AUTHORIZATION_CODE, ttnpb.GRANT_PASSWORD},
	Rights:            []ttnpb.Right{ttnpb.RIGHT_APPLICATION_INFO},
	CreatorIDs:        bob.UserIdentifiers,
}

var clientSpecializer = func(base ttnpb.Client) store.Client {
	return &base
}

func TestClients(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

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
