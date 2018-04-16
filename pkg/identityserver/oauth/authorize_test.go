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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/labstack/echo"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"golang.org/x/oauth2"
)

func TestAuthorizationFlow(t *testing.T) {
	a := assertions.New(t)
	server := testServer(t)

	state := "state"
	rights := []ttnpb.Right{
		ttnpb.RIGHT_USER_INFO,
		ttnpb.RIGHT_USER_APPLICATIONS_LIST,
	}

	// Log in.
	body := strings.NewReader(`{"user_id":"` + userID + `","password":"` + password + `"}`)
	req := httptest.NewRequest(echo.POST, "/api/auth/login?n=/foo/bar", body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	resp := rec.Result()
	a.So(resp.StatusCode, should.Equal, http.StatusOK)
	a.So(resp.Cookies(), should.NotBeEmpty)

	var code string

	// Authorize client.
	{
		uri := fmt.Sprintf("/oauth/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=%s&scope=%s", client.ClientID, client.RedirectURI, state, url.QueryEscape(Scope(rights)))

		req := httptest.NewRequest("POST", uri, strings.NewReader(`authorize=true`))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.AddCookie(resp.Cookies()[0])

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

	// Exchange authorization code per an access token.
	{
		req := httptest.NewRequest("POST", "/oauth/token", strings.NewReader(fmt.Sprintf(`{"code":"%s", "grant_type": "authorization_code", "redirect_uri": "%s"}`, code, client.RedirectURI)))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(resp.Cookies()[0])

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
	}
}
