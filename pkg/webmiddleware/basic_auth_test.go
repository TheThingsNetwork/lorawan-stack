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
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	. "go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
)

func TestBasicAuth(t *testing.T) {
	m := BasicAuth("Password Protected", AuthUser("username", "password"))

	t.Run("No Auth", func(t *testing.T) {
		a := assertions.New(t)
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Handler was called when it shouldn't have")
		})).ServeHTTP(rec, r)
		res := rec.Result()
		a.So(res.StatusCode, should.Equal, http.StatusUnauthorized)
		a.So(res.Header.Get("WWW-Authenticate"), should.ContainSubstring, "Password Protected")
	})

	t.Run("No Username or Password", func(t *testing.T) {
		a := assertions.New(t)
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.SetBasicAuth("", "")
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Handler was called when it shouldn't have")
		})).ServeHTTP(rec, r)
		res := rec.Result()
		a.So(res.StatusCode, should.Equal, http.StatusUnauthorized)
		a.So(res.Header.Get("WWW-Authenticate"), should.ContainSubstring, "Password Protected")
	})

	t.Run("No Username", func(t *testing.T) {
		a := assertions.New(t)
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.SetBasicAuth("", "password")
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Handler was called when it shouldn't have")
		})).ServeHTTP(rec, r)
		res := rec.Result()
		a.So(res.StatusCode, should.Equal, http.StatusUnauthorized)
		a.So(res.Header.Get("WWW-Authenticate"), should.ContainSubstring, "Password Protected")
	})

	t.Run("No Password", func(t *testing.T) {
		a := assertions.New(t)
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.SetBasicAuth("username", "")
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Handler was called when it shouldn't have")
		})).ServeHTTP(rec, r)
		res := rec.Result()
		a.So(res.StatusCode, should.Equal, http.StatusUnauthorized)
		a.So(res.Header.Get("WWW-Authenticate"), should.ContainSubstring, "Password Protected")
	})

	t.Run("Wrong Auth", func(t *testing.T) {
		a := assertions.New(t)
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.SetBasicAuth("username", "wrong-password")
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("Handler was called when it shouldn't have")
		})).ServeHTTP(rec, r)
		res := rec.Result()
		a.So(res.StatusCode, should.Equal, http.StatusUnauthorized)
		a.So(res.Header.Get("WWW-Authenticate"), should.ContainSubstring, "Password Protected")
	})

	t.Run("Successful Auth", func(t *testing.T) {
		a := assertions.New(t)
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.SetBasicAuth("username", "password")
		rec := httptest.NewRecorder()
		m(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Secret"))
		})).ServeHTTP(rec, r)
		res := rec.Result()
		a.So(res.StatusCode, should.Equal, http.StatusOK)
		body, _ := io.ReadAll(res.Body)
		a.So(string(body), should.Equal, "Secret")
	})
}
