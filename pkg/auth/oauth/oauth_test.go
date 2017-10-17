// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package oauth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/migrations"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/TheThingsNetwork/ttn/pkg/web"
	"github.com/labstack/echo"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"golang.org/x/oauth2"
)

var (
	address  = "postgres://root@localhost:26257/%s?sslmode=disable"
	database = "is_tests"
	issuer   = "issuer.test.local"
	username = "john-doe"
	client   = &ttnpb.Client{
		ClientIdentifier: ttnpb.ClientIdentifier{ClientID: "foo"},
		RedirectURI:      "http://example.com/oauth/callback",
		Secret:           "secret",
		Grants: []ttnpb.GrantType{
			ttnpb.GRANT_AUTHORIZATION_CODE,
			ttnpb.GRANT_REFRESH_TOKEN,
		},
		Rights: []ttnpb.Right{
			ttnpb.RIGHT_USER_PROFILE_READ,
		},
	}
	authorizer = &TestAuthorizer{
		Body: "<html />",
	}
	s *sql.Store
)

// cleanStore returns a new store instance attached to a newly created database
// where all migrations has been applied and also has been feed with some users.
func cleanStore(t testing.TB, database string) *sql.Store {
	if s != nil {
		return s
	}

	logger := test.GetLogger(t, "OAuth")

	// open database connection
	db, err := db.Open(context.Background(), fmt.Sprintf(address, database), migrations.Registry)
	if err != nil {
		logger.WithError(err).Fatal("Failed to establish a connection with the CockroachDB instance")
		return nil
	}

	// drop database
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", database))
	if err != nil {
		logger.WithError(err).Fatalf("Failed to delete database `%s`", database)
		return nil
	}

	// create it again
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", database))
	if err != nil {
		logger.WithError(err).Fatalf("Failed to create database `%s`", database)
		return nil
	}

	// apply all migrations
	err = db.MigrateAll()
	if err != nil {
		logger.WithError(err).Fatal("Failed to apply the migrations from the registry")
		return nil
	}

	s = sql.FromDB(db)

	return s
}

func testServer(t *testing.T) (*web.Server, *auth.Keys) {
	logger := test.GetLogger(t, "OAuth")

	a := assertions.New(t)

	key, err := ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	a.So(err, should.BeNil)

	keys := auth.NewKeys(issuer)

	err = keys.Rotate("", key)
	a.So(err, should.BeNil)

	store := cleanStore(t, database)

	_ = store.Clients.Create(client)

	err = store.Clients.SetClientState(client.ClientID, ttnpb.STATE_APPROVED)
	a.So(err, should.BeNil)

	server := New(issuer, keys, store.OAuth, authorizer)

	mux := web.New(logger)
	server.Register(mux)

	return mux, keys
}

func TestAuthorizationFlowJSON(t *testing.T) {
	a := assertions.New(t)
	server, keys := testServer(t)

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

	// exchange code
	{
		uri = "https://" + issuer + "/oauth/token"
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

		claims, err := auth.FromToken(keys, tok.AccessToken)
		a.So(err, should.BeNil)

		a.So(claims.Rights, should.Resemble, rights)
		a.So(claims.Client, should.Resemble, client.ClientID)
		a.So(claims.Issuer, should.Resemble, issuer)
		a.So(claims.Subject, should.Resemble, "user:"+username)
		a.So(claims.Username(), should.Resemble, username)
	}
}

func TestAuthorizationFlowForm(t *testing.T) {
	a := assertions.New(t)
	server, keys := testServer(t)

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

	// exchange code
	{
		uri = "https://" + issuer + "/oauth/token"
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

		claims, err := auth.FromToken(keys, tok.AccessToken)
		a.So(err, should.BeNil)

		a.So(claims.Rights, should.Resemble, rights)
		a.So(claims.Client, should.Resemble, client.ClientID)
		a.So(claims.Issuer, should.Resemble, issuer)
		a.So(claims.Subject, should.Resemble, "user:"+username)
		a.So(claims.Username(), should.Resemble, username)
		a.So(claims.Valid(), should.BeNil)
		a.So([]time.Time{time.Now().Add(-6 * time.Second), time.Unix(claims.IssuedAt, 0)}, should.BeChronological)
		a.So([]time.Time{time.Unix(claims.IssuedAt, 0), time.Now().Add(1 * time.Second)}, should.BeChronological)
		a.So([]time.Time{time.Now().Add(time.Hour - 5*time.Second), time.Unix(claims.ExpiresAt, 0), time.Now().Add(time.Hour + 5*time.Second)}, should.BeChronological)
	}
}

type TestAuthorizer struct {
	Body string
}

func (a *TestAuthorizer) CheckLogin(c echo.Context) (string, error) {
	return username, nil
}

type authBody struct {
	Authorize bool `json:"auth" form:"authorize"`
}

func (a *TestAuthorizer) Authorize(c echo.Context, client types.Client) (bool, error) {
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
