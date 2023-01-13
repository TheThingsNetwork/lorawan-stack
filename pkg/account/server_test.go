// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package account_test

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
	"go.thethings.network/lorawan-stack/v3/pkg/account"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/pbkdf2"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver"
	"go.thethings.network/lorawan-stack/v3/pkg/oauth"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/webui"
	"golang.org/x/net/publicsuffix"
)

type loginFormData struct {
	encoding string
	UserID   string `json:"user_id"`
	Password string `json:"password"`
}

type tokenFormData struct {
	encoding string
	Token    string `json:"token"`
}

type authorizeFormData struct {
	encoding  string
	Authorize bool `json:"authorize"`
}

var (
	now         = time.Now().Truncate(time.Second)
	mockSession = &ttnpb.UserSession{
		UserIds:   &ttnpb.UserIdentifiers{UserId: "user"},
		SessionId: "session_id",
		CreatedAt: ttnpb.ProtoTimePtr(now),
	}
	mockUser = &ttnpb.User{
		Ids: &ttnpb.UserIdentifiers{UserId: "user"},
	}
)

func init() {
	ctx := test.Context()

	hashValidator := pbkdf2.Default()
	hashValidator.Iterations = 10
	ctx = auth.NewContextWithHashValidator(ctx, hashValidator)

	password, err := auth.Hash(ctx, "pass")
	if err != nil {
		panic(err)
	}
	mockUser.Password = password
}

func TestAuthentication(t *testing.T) {
	store := &mockStore{}
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		panic(err)
	}
	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			HTTP: config.HTTP{
				Cookie: config.Cookie{
					HashKey:  []byte("12345678123456781234567812345678"),
					BlockKey: []byte("12345678123456781234567812345678"),
				},
			},
		},
	})
	s, err := account.NewServer(c, store, oauth.Config{
		Mount:       "/oauth",
		CSRFAuthKey: []byte("12345678123456781234567812345678"),
		UI: oauth.UIConfig{
			TemplateData: webui.TemplateData{
				SiteName:     "The Things Network",
				Title:        "Account",
				CanonicalURL: "https://example.com/oauth",
			},
		},
	}, identityserver.GenerateCSPString)
	if err != nil {
		panic(err)
	}
	c.RegisterWeb(s)
	componenttest.StartComponent(t, c)

	var csrfToken string
	var r *http.Request

	// Obtain CSRF token.
	r = httptest.NewRequest("GET", "/oauth/login", nil)
	r.URL.Scheme, r.URL.Host = "http", r.Host
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	c.ServeHTTP(rr, r)
	csrfToken = rr.Header().Get("X-CSRF-Token")

	if cookies := rr.Result().Cookies(); len(cookies) > 0 {
		jar.SetCookies(r.URL, cookies)
	}

	for _, tt := range []struct {
		Name             string
		StoreSetup       func(*mockStore)
		StoreCheck       func(*testing.T, *mockStore)
		Method           string
		Path             string
		Body             interface{}
		ExpectedCode     int
		ExpectedRedirect string
		ExpectedBody     string
	}{
		{
			Method:       "GET",
			Path:         "/oauth/",
			ExpectedCode: http.StatusOK,
			ExpectedBody: "The Things Network Account",
		},
		{
			Method:       "GET",
			Path:         "/oauth/login",
			ExpectedCode: http.StatusOK,
			ExpectedBody: "The Things Network Account",
		},
		{
			Name:         "GET me without auth",
			Method:       "GET",
			Path:         "/oauth/api/me",
			ExpectedCode: http.StatusUnauthorized,
		},
		{
			Name:         "logout without auth",
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
					a.So(s.req.userIDs.UserId, should.Equal, "user")
				}
			},
		},
		{
			Name: "login error",
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
					a.So(s.req.userIDs.UserId, should.Equal, "user")
				}
			},
		},
		{
			Name:         "login no user_id",
			Method:       "POST",
			Path:         "/oauth/api/auth/login",
			Body:         loginFormData{"json", "", "pass"},
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name:         "login no password",
			Method:       "POST",
			Path:         "/oauth/api/auth/login",
			Body:         loginFormData{"json", "user", ""},
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name: "login wrong password",
			StoreSetup: func(s *mockStore) {
				s.res.user = mockUser
			},
			Method:       "POST",
			Path:         "/oauth/api/auth/login",
			Body:         loginFormData{"json", "user", "wrong_pass"},
			ExpectedCode: http.StatusBadRequest,
		},
		{
			Name: "login",
			StoreSetup: func(s *mockStore) {
				s.res.user = mockUser
				s.res.session = mockSession
			},
			Method:       "POST",
			Path:         "/oauth/api/auth/login",
			Body:         loginFormData{"json", "user", "pass"},
			ExpectedCode: http.StatusNoContent,
		},
		{
			Name: "GET me with auth",
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
				a.So(s.req.userIDs.GetUserId(), should.Equal, "user")
				a.So(s.req.sessionID, should.Equal, "session_id") // actually the before-last call.
			},
		},
		{
			Name: "redirect to root",
			StoreSetup: func(s *mockStore) {
				s.res.session = mockSession
				s.res.user = mockUser
			},
			Method:           "GET",
			Path:             "/oauth/login",
			ExpectedCode:     http.StatusFound,
			ExpectedRedirect: "/oauth",
		},
		{
			Name: "logout",
			StoreSetup: func(s *mockStore) {
				s.res.session = mockSession
				s.res.user = mockUser
			},
			Method:       "POST",
			Path:         "/oauth/api/auth/logout",
			ExpectedCode: http.StatusNoContent,
			StoreCheck: func(t *testing.T, s *mockStore) {
				a := assertions.New(t)
				a.So(s.calls, should.Contain, "DeleteSession")
				a.So(s.req.userIDs.GetUserId(), should.Equal, "user")
				a.So(s.req.sessionID, should.Equal, "session_id")
			},
		},
		{
			Name:         "invalid token login",
			Method:       "POST",
			Path:         "/oauth/api/auth/token-login",
			Body:         tokenFormData{"form", ""},
			ExpectedCode: http.StatusBadRequest,
			StoreCheck: func(t *testing.T, s *mockStore) {
				a := assertions.New(t)
				a.So(s.calls, should.NotContain, "ConsumeLoginToken")
				a.So(s.calls, should.NotContain, "CreateSession")
			},
		},
		{
			Name: "token login",
			StoreSetup: func(s *mockStore) {
				s.res.loginToken = &ttnpb.LoginToken{
					UserIds: mockUser.GetIds(),
				}
				s.res.session = mockSession
			},
			Method:       "POST",
			Path:         "/oauth/api/auth/token-login",
			Body:         tokenFormData{"form", "this-is-the-token"},
			ExpectedCode: http.StatusNoContent,
			StoreCheck: func(t *testing.T, s *mockStore) {
				a := assertions.New(t)
				a.So(s.calls, should.Contain, "ConsumeLoginToken")
				a.So(s.req.token, should.Equal, "this-is-the-token")

				a.So(s.calls, should.Contain, "CreateSession")
				a.So(s.req.session.GetUserIds(), should.Resemble, mockUser.GetIds())
			},
		},
	} {
		name := tt.Name
		if name == "" {
			name = fmt.Sprintf("%s %s", tt.Method, tt.Path)
		}
		t.Run(name, func(t *testing.T) {
			store.reset()
			if tt.StoreSetup != nil {
				tt.StoreSetup(store)
			}

			req := httptest.NewRequest(tt.Method, tt.Path, nil)
			req.URL.Scheme, req.URL.Host = "http", req.Host

			var body *bytes.Buffer

			req.Header.Set("X-CSRF-Token", csrfToken)

			for _, c := range jar.Cookies(req.URL) {
				req.AddCookie(c)
			}

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
			case tokenFormData:
				if b.encoding == "json" {
					json, _ := json.Marshal(b)
					body = bytes.NewBuffer(json)
					contentType = "application/json"
				}
				if b.encoding == "form" {
					body = bytes.NewBuffer([]byte(url.Values{
						"token": []string{b.Token},
					}.Encode()))
					contentType = "application/x-www-form-urlencoded"
				}
			}

			if contentType != "" {
				req.Header.Set("Content-Type", contentType)
			}
			if body != nil {
				req.Body = io.NopCloser(body)
				req.ContentLength = int64(body.Len())
			}

			res := httptest.NewRecorder()

			c.ServeHTTP(res, req)

			a := assertions.New(t)
			a.So(res.Code, should.Equal, tt.ExpectedCode)
			a.So(res.Header().Get("location"), should.ContainSubstring, tt.ExpectedRedirect)
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
