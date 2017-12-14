// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestOAuthAuthorizationCode(t *testing.T) {
	a := assertions.New(t)
	s := cleanStore(t, database)

	userID := testUsers()["john-doe"].UserID
	client := testClients()["test-client"]

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

	err := s.OAuth.SaveAuthorizationCode(data)
	a.So(err, should.BeNil)

	found, err := s.OAuth.GetAuthorizationCode(data.AuthorizationCode)
	a.So(err, should.BeNil)

	a.So(found.AuthorizationCode, should.Equal, data.AuthorizationCode)
	a.So(found.ClientID, should.Equal, data.ClientID)
	a.So(found.UserID, should.Equal, data.UserID)
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

func TestOAuthAccessToken(t *testing.T) {
	a := assertions.New(t)
	s := cleanStore(t, database)

	userID := testUsers()["john-doe"].UserID
	client := testClients()["test-client"]

	data := &types.AccessData{
		AccessToken: "123456",
		ClientID:    client.ClientIdentifier.ClientID,
		UserID:      userID,
		CreatedAt:   time.Now(),
		ExpiresIn:   time.Hour,
		Scope:       "scope",
		RedirectURI: "https://example.com/oauth/callback",
	}

	err := s.OAuth.SaveAccessToken(data)
	a.So(err, should.BeNil)

	found, err := s.OAuth.GetAccessToken(data.AccessToken)
	a.So(err, should.BeNil)

	a.So(found.AccessToken, should.Equal, data.AccessToken)
	a.So(found.ClientID, should.Equal, data.ClientID)
	a.So(found.UserID, should.Equal, data.UserID)
	a.So(found.Scope, should.Equal, data.Scope)
	a.So(found.RedirectURI, should.Equal, data.RedirectURI)
	a.So(found.ExpiresIn, should.Equal, data.ExpiresIn)

	c, err := s.Clients.GetByID(found.ClientID, clientFactory)
	a.So(err, should.BeNil)
	a.So(c, test.ShouldBeClientIgnoringAutoFields, client)

	err = s.OAuth.DeleteAccessToken(data.AccessToken)
	a.So(err, should.BeNil)

	_, err = s.OAuth.GetAccessToken(data.AccessToken)
	a.So(err, should.NotBeNil)
}

func TestOAuthRefreshToken(t *testing.T) {
	a := assertions.New(t)
	s := cleanStore(t, database)

	userID := testUsers()["john-doe"].UserID
	client := testClients()["test-client"]

	data := &types.RefreshData{
		RefreshToken: "123456",
		ClientID:     client.ClientIdentifier.ClientID,
		UserID:       userID,
		CreatedAt:    time.Now(),
		Scope:        "scope",
		RedirectURI:  "https://example.com/oauth/callback",
	}

	err := s.OAuth.SaveRefreshToken(data)
	a.So(err, should.BeNil)

	found, err := s.OAuth.GetRefreshToken(data.RefreshToken)
	a.So(err, should.BeNil)

	a.So(found.RefreshToken, should.Equal, data.RefreshToken)
	a.So(found.ClientID, should.Equal, data.ClientID)
	a.So(found.UserID, should.Equal, data.UserID)
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

func TestOAuthRevokeClient(t *testing.T) {
	a := assertions.New(t)
	s := cleanStore(t, database)

	userID := testUsers()["john-doe"].UserID
	client := testClients()["test-client"]

	accessData := &types.AccessData{
		AccessToken: "123456",
		ClientID:    client.ClientIdentifier.ClientID,
		UserID:      userID,
		CreatedAt:   time.Now(),
		ExpiresIn:   time.Hour,
		Scope:       "scope",
		RedirectURI: "https://example.com/oauth/callback",
	}

	err := s.OAuth.SaveAccessToken(accessData)
	a.So(err, should.BeNil)

	refreshData := &types.RefreshData{
		RefreshToken: "123456",
		ClientID:     client.ClientIdentifier.ClientID,
		UserID:       userID,
		CreatedAt:    time.Now(),
		Scope:        "scope",
		RedirectURI:  "https://example.com/oauth/callback",
	}

	err = s.OAuth.SaveRefreshToken(refreshData)
	a.So(err, should.BeNil)

	err = s.OAuth.RevokeAuthorizedClient(userID, client.ClientID)
	a.So(err, should.BeNil)

	err = s.OAuth.RevokeAuthorizedClient(userID, client.ClientID)
	a.So(err, should.NotBeNil)
	a.So(ErrRefreshTokenNotFound.Describes(err), should.BeTrue)
}
