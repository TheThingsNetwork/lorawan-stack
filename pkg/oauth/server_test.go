// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package oauth_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/oauth"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/webui"
	"golang.org/x/net/publicsuffix"
)

type loginFormData struct {
	encoding string
	UserID   string `json:"user_id"`
	Password string `json:"password"`
}

var (
	mockSession = &ttnpb.UserSession{
		UserIdentifiers: ttnpb.UserIdentifiers{UserID: "user"},
		SessionID:       "session_id",
		CreatedAt:       time.Now().Truncate(time.Second),
	}
	mockUser = &ttnpb.User{
		UserIdentifiers: ttnpb.UserIdentifiers{UserID: "user"},
	}
	mockClient = &ttnpb.Client{
		ClientIdentifiers: ttnpb.ClientIdentifiers{ClientID: "client"},
		State:             ttnpb.STATE_APPROVED,
		Grants:            []ttnpb.GrantType{ttnpb.GRANT_AUTHORIZATION_CODE, ttnpb.GRANT_REFRESH_TOKEN},
		RedirectURIs:      []string{"http://callback"},
		Rights:            []ttnpb.Right{ttnpb.RIGHT_USER_INFO},
	}
)

func TestOAuthFlow(t *testing.T) {
	ctx := test.Context()
	store := &mockStore{}
	password, err := auth.Hash("pass")
	if err != nil {
		panic(err)
	}
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		panic(err)
	}
	c := component.MustNew(test.GetLogger(t), &component.Config{
		ServiceBase: config.ServiceBase{
			HTTP: config.HTTP{
				Cookie: config.Cookie{
					HashKey:  []byte("12345678123456781234567812345678"),
					BlockKey: []byte("12345678123456781234567812345678"),
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}
	s := oauth.NewServer(ctx, store, oauth.Config{
		Mount: "/oauth",
		UI: oauth.UIConfig{
			TemplateData: webui.TemplateData{
				SiteName: "The Things Network",
				Title:    "OAuth",
			},
		},
	})
	c.RegisterWeb(s)
	if err = c.Start(); err != nil {
		panic(err)
	}

	for _, tt := range []struct {
		StoreSetup   func(*mockStore)
		StoreCheck   func(*testing.T, *mockStore)
		Method       string
		Path         string
		Body         interface{}
		ExpectedCode int
		ExpectedBody string
	}{
		{
			Method:       "GET",
			Path:         "/oauth",
			ExpectedCode: http.StatusOK,
			ExpectedBody: "The Things Network OAuth",
		},
		{
			Method:       "GET",
			Path:         "/oauth",
			ExpectedCode: http.StatusOK,
			ExpectedBody: "The Things Network OAuth",
		},
		{
			Method:       "GET",
			Path:         "/oauth/login",
			ExpectedCode: http.StatusOK,
			ExpectedBody: "The Things Network OAuth",
		},
		{
			Method:       "GET",
			Path:         "/oauth/api/me",
			ExpectedCode: http.StatusUnauthorized,
		},
		{
			Method:       "POST",
			Path:         "/oauth/api/auth/logout",
			ExpectedCode: http.StatusUnauthorized,
		},
		{
			StoreSetup: func(s *mockStore) {
				s.err.getUser = mockErrUnauthenticated
			},
			Method:       "POST",
			Path:         "/oauth/api/auth/login",
			Body:         loginFormData{"json", "user", "pass"},
			ExpectedCode: http.StatusUnauthorized,
			StoreCheck: func(t *testing.T, s *mockStore) {
				a := assertions.New(t)
				a.So(s.calls, should.Contain, "GetUser")
				if a.So(s.req.userIDs, should.NotBeNil) {
					a.So(s.req.userIDs.UserID, should.Equal, "user")
				}
			},
		},
		{
			StoreSetup: func(s *mockStore) {
				s.err.getUser = mockErrUnauthenticated
			},
			Method:       "POST",
			Path:         "/oauth/api/auth/login",
			Body:         loginFormData{"form", "user", "pass"},
			ExpectedCode: http.StatusUnauthorized,
			StoreCheck: func(t *testing.T, s *mockStore) {
				a := assertions.New(t)
				a.So(s.calls, should.Contain, "GetUser")
				if a.So(s.req.userIDs, should.NotBeNil) {
					a.So(s.req.userIDs.UserID, should.Equal, "user")
				}
			},
		},
		{
			StoreSetup: func(s *mockStore) {
				s.res.user = &ttnpb.User{Password: string(password)}
			},
			Method:       "POST",
			Path:         "/oauth/api/auth/login",
			Body:         loginFormData{"json", "user", "wrong_pass"},
			ExpectedCode: http.StatusUnauthorized,
		},
		{
			StoreSetup: func(s *mockStore) {
				s.res.user = &ttnpb.User{Password: string(password)}
				s.res.session = mockSession
			},
			Method:       "POST",
			Path:         "/oauth/api/auth/login",
			Body:         loginFormData{"json", "user", "pass"},
			ExpectedCode: http.StatusOK,
		},
		{
			StoreSetup: func(s *mockStore) {
				s.res.session = mockSession
				s.res.user = mockUser
			},
			Method:       "GET",
			Path:         "/oauth/api/me",
			ExpectedCode: http.StatusOK,
			ExpectedBody: `"user_id":"user"`,
			StoreCheck: func(t *testing.T, s *mockStore) {
				a := assertions.New(t)
				a.So(s.calls, should.Contain, "GetUser")
				a.So(s.req.userIDs.GetUserID(), should.Equal, "user")
				a.So(s.req.sessionID, should.Equal, "session_id") // actually the before-last call.
			},
		},
		{
			StoreSetup: func(s *mockStore) {
				s.res.session = mockSession
				s.res.user = mockUser
				s.res.client = mockClient
				s.err.getAuthorization = mockErrNotFound
			},
			Method:       "GET",
			Path:         "/oauth/authorize?client_id=client&redirect_uri=http://callback&response_type=code&state=foo",
			ExpectedCode: http.StatusOK,
			ExpectedBody: `"client":{"ids":{"client_id":"client"}`,
			StoreCheck: func(t *testing.T, s *mockStore) {
				a := assertions.New(t)
				a.So(s.req.clientIDs.GetClientID(), should.Equal, "client")
				a.So(s.calls, should.Contain, "GetClient")
				a.So(s.calls, should.Contain, "GetAuthorization")
			},
		},
		{
			StoreSetup: func(s *mockStore) {
				s.res.session = mockSession
				s.res.user = mockUser
			},
			Method:       "POST",
			Path:         "/oauth/api/auth/logout",
			ExpectedCode: http.StatusOK,
			StoreCheck: func(t *testing.T, s *mockStore) {
				a := assertions.New(t)
				a.So(s.calls, should.Contain, "DeleteSession")
				a.So(s.req.userIDs.GetUserID(), should.Equal, "user")
				a.So(s.req.sessionID, should.Equal, "session_id")
			},
		},
	} {
		t.Run(fmt.Sprintf("%s %s", tt.Method, tt.Path), func(t *testing.T) {
			store.reset()
			if tt.StoreSetup != nil {
				tt.StoreSetup(store)
			}

			var body io.Reader
			var contentType string
			switch b := tt.Body.(type) {
			case loginFormData:
				if b.encoding == "json" {
					json, _ := json.Marshal(b)
					body = bytes.NewBuffer(json)
					contentType = "application/json"
				}
				if b.encoding == "form" {
					body = bytes.NewBuffer([]byte(url.Values{
						"user_id":  []string{b.UserID},
						"password": []string{b.Password},
					}.Encode()))
					contentType = "application/x-www-form-urlencoded"
				}
			case io.Reader:
				body = b
			}

			req := httptest.NewRequest(tt.Method, tt.Path, body)
			req.URL.Scheme, req.URL.Host = "http", req.Host
			if contentType != "" {
				req.Header.Set("Content-Type", contentType)
			}
			for _, c := range jar.Cookies(req.URL) {
				req.AddCookie(c)
				if c.Name == "_csrf" {
					req.Header.Set("X-CSRF-Token", c.Value)
				}
			}
			res := httptest.NewRecorder()

			c.ServeHTTP(res, req)

			a := assertions.New(t)
			a.So(res.Code, should.Equal, tt.ExpectedCode)
			if tt.ExpectedBody != "" {
				if a.So(res.Body, should.NotBeNil) {
					a.So(res.Body.String(), should.ContainSubstring, tt.ExpectedBody)
				}
			}
			if cookies := res.Result().Cookies(); len(cookies) > 0 {
				jar.SetCookies(req.URL, cookies)
			}

			if tt.StoreCheck != nil {
				tt.StoreCheck(t, store)
			}
		})
	}

}
