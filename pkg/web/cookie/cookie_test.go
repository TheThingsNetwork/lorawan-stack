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

package cookie

import (
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
	"golang.org/x/net/publicsuffix"
)

var testCookieSettings = &Cookie{
	Name:     "test_cookie",
	HTTPOnly: true,
}

type testCookieValue struct {
	Value string
}

func testSetCookie(t *testing.T, value string) http.HandlerFunc {
	a := assertions.New(t)
	return func(w http.ResponseWriter, r *http.Request) {
		err := testCookieSettings.Set(w, r, &testCookieValue{
			Value: value,
		})
		a.So(err, should.BeNil)
	}
}

func testGetCookie(t *testing.T, value string, exists bool) http.HandlerFunc {
	a := assertions.New(t)
	return func(w http.ResponseWriter, r *http.Request) {
		var v testCookieValue
		ok, err := testCookieSettings.Get(w, r, &v)
		a.So(err, should.BeNil)
		a.So(ok, should.Equal, exists)
		a.So(v.Value, should.Equal, value)

		present := testCookieSettings.Exists(r)
		a.So(present, should.Equal, exists)
	}
}

func testDeleteCookie() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		testCookieSettings.Remove(w, r)
	}
}

func TestCookie(t *testing.T) {
	root := mux.NewRouter()

	a := assertions.New(t)
	blockKey := random.Bytes(32)
	hashKey := random.Bytes(64)

	root.Use(mux.MiddlewareFunc(webmiddleware.Cookies(hashKey, blockKey)))

	root.Path("/set").HandlerFunc(testSetCookie(t, "test_value")).Methods(http.MethodGet)
	root.Path("/get").HandlerFunc(testGetCookie(t, "test_value", true)).Methods(http.MethodGet)
	root.Path("/del").HandlerFunc(testDeleteCookie()).Methods(http.MethodGet)
	root.Path("/no_cookie").HandlerFunc(testGetCookie(t, "", false)).Methods(http.MethodGet)

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		panic(err)
	}

	doGET := func(path string) *http.Response {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.URL.Scheme, req.URL.Host = "http", req.Host
		for _, c := range jar.Cookies(req.URL) {
			req.AddCookie(c)
		}

		rec := httptest.NewRecorder()

		root.ServeHTTP(rec, req)

		resp := rec.Result()
		resp.Request = req

		if cookies := resp.Cookies(); len(cookies) > 0 {
			jar.SetCookies(req.URL, cookies)
		}

		return resp
	}

	resp := doGET("/no_cookie")

	cookies := jar.Cookies(resp.Request.URL)
	a.So(cookies, should.BeEmpty)

	resp = doGET("/del")

	cookies = jar.Cookies(resp.Request.URL)
	a.So(cookies, should.BeEmpty)

	resp = doGET("/set")

	cookies = jar.Cookies(resp.Request.URL)
	if a.So(cookies, should.HaveLength, 1) {
		cookie := cookies[0]
		a.So(cookie.Name, should.Equal, "test_cookie")
	}

	resp = doGET("/get")

	cookies = jar.Cookies(resp.Request.URL)
	if a.So(cookies, should.HaveLength, 1) {
		cookie := cookies[0]
		a.So(cookie.Name, should.Equal, "test_cookie")
	}

	resp = doGET("/del")

	cookies = jar.Cookies(resp.Request.URL)
	a.So(cookies, should.BeEmpty)

	resp = doGET("/no_cookie")

	cookies = jar.Cookies(resp.Request.URL)
	a.So(cookies, should.BeEmpty)
}
