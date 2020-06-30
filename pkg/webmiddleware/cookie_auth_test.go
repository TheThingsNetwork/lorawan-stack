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

package webmiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/securecookie"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
)

func TestCookieAuth(t *testing.T) {
	a := assertions.New(t)
	sc := securecookie.New(
		[]byte("1234123412341234123412341234123412341234123412341234123412341234"),
		[]byte("12341234123412341234123412341234"),
	)
	m := CookieAuth("_session")
	cookieMiddleware := Cookies(
		[]byte("1234123412341234123412341234123412341234123412341234123412341234"),
		[]byte("12341234123412341234123412341234"),
	)
	authCookie := &auth.CookieShape{
		UserID:        "test-user",
		SessionID:     "id-1234",
		SessionSecret: "secret-1234",
	}
	encoded, _ := sc.Encode("_session", authCookie)
	cookie := &http.Cookie{
		Name:     "_session",
		Value:    encoded,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
	}

	t.Run("Forwards cookie value to auth header", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, "/", nil)
		rec := httptest.NewRecorder()
		r.AddCookie(cookie)

		cookieMiddleware(
			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				a.So(r.Header.Get("Authorization"), should.Equal, "Bearer "+auth.JoinToken(auth.SessionToken, "id-1234", "secret-1234"))
			})),
		).ServeHTTP(rec, r)
		res := rec.Result()

		a.So(res.StatusCode, should.Equal, http.StatusOK)
	})

	t.Run("Does not overwrite existing auth header", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPut, "/", nil)
		rec := httptest.NewRecorder()
		r.AddCookie(cookie)
		r.Header.Set("Authorization", "Bearer 1234")

		cookieMiddleware(
			m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				a.So(r.Header.Get("Authorization"), should.Equal, "Bearer 1234")
			})),
		).ServeHTTP(rec, r)

		res := rec.Result()
		cookies := r.Cookies()

		a.So(cookies, should.HaveLength, 1)
		a.So(cookies[0].Name, should.Equal, "_session")
		a.So(cookies[0].Value, should.NotBeEmpty)
		a.So(res.StatusCode, should.Equal, http.StatusOK)
	})
}
