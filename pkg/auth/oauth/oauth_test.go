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
	client   = &types.DefaultClient{
		ID:     "foo",
		URI:    "http://example.com/oauth/callback",
		Secret: "secret",
		Grants: types.Grants{
			AuthorizationCode: true,
			RefreshToken:      true,
		},
	}
	authorizer = &TestAuthorizer{
		Body: "<html />",
	}
)

// cleanStore returns a new store instance attached to a newly created database
// where all migrations has been applied and also has been feed with some users.
func cleanStore(t testing.TB, database string) *sql.Store {
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

	return sql.FromDB(db)
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

	_, err = store.Clients.Register(client)
	a.So(err, should.BeNil)

	err = store.Clients.Approve(client.ID)
	a.So(err, should.BeNil)

	server := New(issuer, keys, store.OAuth, authorizer)

	mux := web.New(logger)
	server.Register(mux)

	return mux, keys
}

func TestAuthorizationFlow(t *testing.T) {
	a := assertions.New(t)
	server, keys := testServer(t)

	state := "state"

	uri := fmt.Sprintf("https://%s/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=%s&scope=%s", issuer, client.ID, client.URI, state, Scope([]ttnpb.Right{
		ttnpb.RIGHT_USER_PROFILE_READ,
	}))

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
		a.So(u.Query().Get("code"), should.NotBeEmpty)
		a.So(u.Query().Get("state"), should.Equal, state)
	}

	// authorize client (html form)
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
		// req := httptest.NewRequest("POST", uri, strings.NewReader(fmt.Sprintf("code=%s&grant_type=authorization_code&redirect_uri=%s", code, client.URI)))
		// req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		req := httptest.NewRequest("POST", uri, strings.NewReader(fmt.Sprintf(`{"code":"%s", "grant_type": "authorization_code", "redirect_uri": "%s"}`, code, client.URI)))
		req.Header.Set("Content-Type", "application/json")

		req.SetBasicAuth(client.ID, client.Secret)

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

		fmt.Println(claims)
	}
}

type TestAuthorizer struct {
	Body string
}

func (a *TestAuthorizer) CheckLogin(c echo.Context) (string, error) {
	return "john-doe", nil
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
