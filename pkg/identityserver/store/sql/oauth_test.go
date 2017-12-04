// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestAuthorizationCode(t *testing.T) {
	a := assertions.New(t)
	s := cleanStore(t, database)

	userID := testUsers()["john-doe"].UserID

	client := testClients()["test-client"]
	err := s.Clients.Create(client)
	a.So(err, should.BeNil)

	data := &types.AuthorizationData{
		AuthorizationCode: "123456",
		ClientID:          client.ClientIdentifier.ClientID,
		CreatedAt:         time.Now(),
		ExpiresIn:         5 * time.Second,
		Scope:             "scope",
		RedirectURI:       "https://example.com/oauth/callback",
		State:             "state",
		UserID:            userID,
	}

	err = s.OAuth.SaveAuthorizationCode(data)
	a.So(err, should.BeNil)

	found, err := s.OAuth.GetAuthorizationCode(data.AuthorizationCode)
	a.So(err, should.BeNil)

	a.So(found.AuthorizationCode, should.Equal, data.AuthorizationCode)
	a.So(found.ClientID, should.Equal, data.ClientID)
	a.So(found.ExpiresIn, should.Equal, data.ExpiresIn)
	a.So(found.Scope, should.Equal, data.Scope)
	a.So(found.RedirectURI, should.Equal, data.RedirectURI)
	a.So(found.State, should.Equal, data.State)

	c, err := s.Clients.GetByID(found.ClientID, clientFactory)
	a.So(err, should.BeNil)
	a.So(c, test.ShouldBeClientIgnoringAutoFields, client)

	err = s.OAuth.DeleteAuthorizationCode(data.AuthorizationCode)
	a.So(err, should.BeNil)

	_, err = s.OAuth.GetAuthorizationCode(data.AuthorizationCode)
	a.So(err, should.NotBeNil)
}

func TestRefreshToken(t *testing.T) {
	a := assertions.New(t)

	s := cleanStore(t, database)

	client := testClients()["test-client"]
	err := s.Clients.Create(client)
	a.So(err, should.BeNil)

	data := &types.RefreshData{
		RefreshToken: "123456",
		ClientID:     client.ClientIdentifier.ClientID,
		CreatedAt:    time.Now(),
		Scope:        "scope",
		RedirectURI:  "https://example.com/oauth/callback",
	}

	err = s.OAuth.SaveRefreshToken(data)
	a.So(err, should.BeNil)

	found, err := s.OAuth.GetRefreshToken(data.RefreshToken)
	a.So(err, should.BeNil)

	a.So(found.RefreshToken, should.Equal, data.RefreshToken)
	a.So(found.ClientID, should.Equal, data.ClientID)
	a.So(found.Scope, should.Equal, data.Scope)
	a.So(found.RedirectURI, should.Equal, data.RedirectURI)

	c, err := s.Clients.GetByID(found.ClientID, clientFactory)
	a.So(err, should.BeNil)
	a.So(c, test.ShouldBeClientIgnoringAutoFields, client)

	err = s.OAuth.DeleteRefreshToken(data.RefreshToken)
	a.So(err, should.BeNil)

	_, err = s.OAuth.GetRefreshToken(data.RefreshToken)
	a.So(err, should.NotBeNil)
}
