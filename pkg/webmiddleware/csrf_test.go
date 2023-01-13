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

package webmiddleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/csrf"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	. "go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
)

func TestCSRF(t *testing.T) {
	a := assertions.New(t)
	authKey := []byte("1234123412341234123412341234123412341234123412341234123412341234")
	m := CSRF(authKey)

	t.Run("Protects non-idempotent methods when using a Session Token", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", nil)
		r.Header.Set("Authorization", "Bearer "+auth.JoinToken(auth.SessionToken, "XXX", "YYY"))
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		})).ServeHTTP(rec, r)
		res := rec.Result()
		a.So(res.StatusCode, should.Equal, http.StatusForbidden)
	})

	t.Run("Allows non-idempotent methods when using an API Key", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", nil)
		r.Header.Set("Authorization", "Bearer "+auth.JoinToken(auth.APIKey, "XXX", "YYY"))
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		})).ServeHTTP(rec, r)
		res := rec.Result()
		a.So(res.StatusCode, should.Equal, http.StatusOK)
	})

	t.Run("Allows access with valid CSRF token", func(t *testing.T) {
		var csrfToken string
		var r *http.Request

		// Obtain CSRF token.
		r = httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			csrfToken = csrf.Token(r)
		})).ServeHTTP(rec, r)
		res := rec.Result()

		cookies := res.Cookies()
		a.So(cookies, should.HaveLength, 1)

		// Make request
		r = httptest.NewRequest(http.MethodPost, "/", nil)
		r.Header.Set("X-CSRF-Token", csrfToken)
		r.AddCookie(cookies[0])
		rec = httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(rec, r)
		res = rec.Result()
		a.So(res.StatusCode, should.Equal, http.StatusOK)
	})
}
