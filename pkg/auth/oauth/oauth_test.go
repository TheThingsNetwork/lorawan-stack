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

package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/migrations"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/TheThingsNetwork/ttn/pkg/web"
	"github.com/labstack/echo"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"golang.org/x/oauth2"
)

const (
	address  = "postgres://root@localhost:26257/%s?sslmode=disable"
	database = "is_oauth_tests"
	issuer   = "issuer.test.local"
	userID   = "john-doe"
)

var (
	client = &ttnpb.Client{
		ClientIdentifiers: ttnpb.ClientIdentifiers{ClientID: "foo"},
		RedirectURI:       "http://example.com/oauth/callback",
		Secret:            "secret",
		Grants: []ttnpb.GrantType{
			ttnpb.GRANT_AUTHORIZATION_CODE,
			ttnpb.GRANT_REFRESH_TOKEN,
		},
		State: ttnpb.STATE_APPROVED,
		Rights: []ttnpb.Right{
			ttnpb.RIGHT_USER_PROFILE_READ,
		},
		CreatorIDs: ttnpb.UserIdentifiers{UserID: userID},
	}
	authorizer = &TestAuthorizer{
		Body: "<html />",
	}
	s *sql.Store
)

// cleanStore returns a new store instance attached to a newly created database
// where all migrations has been applied and also has been feed with some users.
func cleanStore(logger log.Interface, database string) *sql.Store {
	// open database connection
	db, err := db.Open(context.Background(), fmt.Sprintf(address, database), migrations.Registry)
	if err != nil {
		logger.WithError(err).Fatal("Failed to establish a connection with the CockroachDB instance")
	}

	// drop database
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s CASCADE", database))
	if err != nil {
		logger.WithError(err).Fatalf("Failed to delete database `%s`", database)
	}

	// create it again
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", database))
	if err != nil {
		logger.WithError(err).Fatalf("Failed to create database `%s`", database)
	}

	// apply all migrations
	err = db.MigrateAll()
	if err != nil {
		logger.WithError(err).Fatal("Failed to apply the migrations from the registry")
	}

	return sql.FromDB(db)
}

func testServer(t *testing.T) *web.Server {
	logger := test.GetLogger(t).WithField("tag", "OAuth")

	a := assertions.New(t)

	if s == nil {
		store := cleanStore(logger, database)

		err := store.Users.Create(&ttnpb.User{
			UserIdentifiers: ttnpb.UserIdentifiers{
				UserID: userID,
			},
		})
		a.So(err, should.BeNil)

		err = store.Clients.Create(client)
		a.So(err, should.BeNil)

		s = store
	}

	server := New(logger, issuer, s, authorizer)

	mux := web.New(logger)
	server.Register(mux)

	return mux
}

func TestAuthorizationFlowJSON(t *testing.T) {
	a := assertions.New(t)
	server := testServer(t)

	state := "state"
	rights := []ttnpb.Right{
		ttnpb.RIGHT_USER_PROFILE_READ,
	}

	uri := fmt.Sprintf("https://%s/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=%s&scope=%s", issuer, client.ClientID, client.RedirectURI, state, Scope(rights))

	var code string

	// get authorization page
	{
		req := httptest.NewRequest("GET", uri, nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)
		resp := w.Result()

		body, err := ioutil.ReadAll(resp.Body)
		a.So(err, should.BeNil)

		a.So(resp.StatusCode, should.Equal, http.StatusOK)
		a.So(string(body), should.Resemble, authorizer.Body)
		a.So(resp.Header.Get("Location"), should.BeEmpty)
	}

	// authorize client (JSON)
	{
		req := httptest.NewRequest("POST", uri, strings.NewReader(`{"auth": true}`))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)
		resp := w.Result()

		loc := resp.Header.Get("Location")
		a.So(resp.StatusCode, should.Equal, http.StatusFound)

		u, err := url.Parse(loc)
		a.So(err, should.BeNil)
		code = u.Query().Get("code")
		a.So(code, should.NotBeEmpty)
		a.So(u.Query().Get("state"), should.Equal, state)

	}

	var accessToken string

	// exchange code
	{
		uri = fmt.Sprintf("https://%s/oauth/token", issuer)
		req := httptest.NewRequest("POST", uri, strings.NewReader(fmt.Sprintf(`{"code":"%s", "grant_type": "authorization_code", "redirect_uri": "%s"}`, code, client.RedirectURI)))
		req.Header.Set("Content-Type", "application/json")

		req.SetBasicAuth(client.ClientID, client.Secret)

		_ = code
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)
		resp := w.Result()

		a.So(resp.StatusCode, should.Equal, http.StatusOK)

		body, err := ioutil.ReadAll(resp.Body)
		a.So(err, should.BeNil)

		tok := &oauth2.Token{}
		err = json.Unmarshal(body, tok)
		a.So(err, should.BeNil)

		a.So(tok.AccessToken, should.NotBeEmpty)
		a.So(tok.TokenType, should.Equal, "bearer")

		{
			found, err := s.OAuth.GetAccessToken(tok.AccessToken)
			a.So(err, should.BeNil)
			a.So(found.ClientID, should.Equal, client.ClientID)
			a.So(found.UserID, should.Equal, userID)
			a.So(found.Scope, should.Equal, Scope(rights))
		}

		{
			found, err := s.OAuth.GetRefreshToken(tok.RefreshToken)
			a.So(err, should.BeNil)
			a.So(found.ClientID, should.Equal, client.ClientID)
			a.So(found.UserID, should.Equal, userID)
			a.So(found.Scope, should.Equal, Scope(rights))
		}

		accessToken = tok.AccessToken
	}

	// introspect the token using the info endpoint
	{
		uri = fmt.Sprintf("https://%s/oauth/info", issuer)
		req := httptest.NewRequest("POST", uri, nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)
		resp := w.Result()

		a.So(resp.StatusCode, should.Equal, http.StatusOK)

		body, err := ioutil.ReadAll(resp.Body)
		a.So(err, should.BeNil)

		tok := make(map[string]interface{})
		err = json.Unmarshal(body, &tok)
		a.So(err, should.BeNil)

		a.So(tok["access_token"], should.Equal, accessToken)
		a.So(tok["token_type"], should.Equal, "bearer")
		a.So(tok["client_id"], should.Equal, client.ClientID)
		a.So(tok["scope"], should.Equal, Scope(rights))
		a.So(tok["expires_in"], should.Equal, AccessExpiration)
		a.So(tok["user_id"], should.Equal, userID)
	}
}

func TestAuthorizationFlowForm(t *testing.T) {
	a := assertions.New(t)
	server := testServer(t)

	state := "state"
	rights := []ttnpb.Right{
		ttnpb.RIGHT_USER_PROFILE_READ,
	}

	uri := fmt.Sprintf("https://%s/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=%s&scope=%s", issuer, client.ClientID, client.RedirectURI, state, Scope(rights))

	var code string

	// get authorization page
	{
		req := httptest.NewRequest("GET", uri, nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)
		resp := w.Result()

		body, err := ioutil.ReadAll(resp.Body)
		a.So(err, should.BeNil)

		a.So(resp.StatusCode, should.Equal, http.StatusOK)
		a.So(string(body), should.Resemble, authorizer.Body)
		a.So(resp.Header.Get("Location"), should.BeEmpty)
	}

	// authorize client
	{
		req := httptest.NewRequest("POST", uri, strings.NewReader(`authorize=true`))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)
		resp := w.Result()

		loc := resp.Header.Get("Location")
		a.So(resp.StatusCode, should.Equal, http.StatusFound)

		u, err := url.Parse(loc)
		a.So(err, should.BeNil)
		a.So(u.Query().Get("code"), should.NotBeEmpty)
		a.So(u.Query().Get("state"), should.Equal, state)

		code = u.Query().Get("code")
	}

	var accessToken string

	// exchange code
	{
		uri = fmt.Sprintf("https://%s/oauth/token", issuer)
		req := httptest.NewRequest("POST", uri, strings.NewReader(fmt.Sprintf("code=%s&grant_type=authorization_code&redirect_uri=%s", code, client.RedirectURI)))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		req.SetBasicAuth(client.ClientID, client.Secret)

		_ = code
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)
		resp := w.Result()

		a.So(resp.StatusCode, should.Equal, http.StatusOK)

		body, err := ioutil.ReadAll(resp.Body)
		a.So(err, should.BeNil)

		tok := &oauth2.Token{}
		err = json.Unmarshal(body, tok)
		a.So(err, should.BeNil)

		a.So(tok.AccessToken, should.NotBeEmpty)
		a.So(tok.TokenType, should.Equal, "bearer")

		{
			found, err := s.OAuth.GetAccessToken(tok.AccessToken)
			a.So(err, should.BeNil)
			a.So(found.ClientID, should.Equal, client.ClientID)
			a.So(found.UserID, should.Equal, userID)
			a.So(found.Scope, should.Equal, Scope(rights))
		}

		{
			found, err := s.OAuth.GetRefreshToken(tok.RefreshToken)
			a.So(err, should.BeNil)
			a.So(found.ClientID, should.Equal, client.ClientID)
			a.So(found.UserID, should.Equal, userID)
			a.So(found.Scope, should.Equal, Scope(rights))
		}

		accessToken = tok.AccessToken
	}

	// introspect the token using the info endpoint
	{
		uri = fmt.Sprintf("https://%s/oauth/info", issuer)
		req := httptest.NewRequest("POST", uri, nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)
		resp := w.Result()

		a.So(resp.StatusCode, should.Equal, http.StatusOK)

		body, err := ioutil.ReadAll(resp.Body)
		a.So(err, should.BeNil)

		tok := make(map[string]interface{})
		err = json.Unmarshal(body, &tok)
		a.So(err, should.BeNil)

		a.So(tok["access_token"], should.Equal, accessToken)
		a.So(tok["token_type"], should.Equal, "bearer")
		a.So(tok["client_id"], should.Equal, client.ClientID)
		a.So(tok["scope"], should.Equal, Scope(rights))
		a.So(tok["expires_in"], should.Equal, AccessExpiration)
		a.So(tok["user_id"], should.Equal, userID)
	}
}

type TestAuthorizer struct {
	Body string
}

func (a *TestAuthorizer) CheckLogin(c echo.Context) (string, error) {
	return userID, nil
}

type authBody struct {
	Authorize bool `json:"auth" form:"authorize"`
}

func (a *TestAuthorizer) Authorize(c echo.Context, client store.Client) (bool, error) {
	if c.Request().Method != "POST" {
		c.HTML(http.StatusOK, a.Body)
	}

	body := &authBody{}
	err := c.Bind(body)
	if err != nil {
		return false, err
	}

	return body.Authorize, nil
}
