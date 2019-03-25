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

	echo "github.com/labstack/echo/v4"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/random"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"golang.org/x/net/publicsuffix"
)

var testCookieSettings = &Cookie{
	Name:     "test_cookie",
	HTTPOnly: true,
}

type testCookieValue struct {
	Value string
}

func testSetCookie(t *testing.T, value string) echo.HandlerFunc {
	a := assertions.New(t)
	return func(c echo.Context) error {
		err := testCookieSettings.Set(c, &testCookieValue{
			Value: value,
		})
		a.So(err, should.BeNil)

		return err
	}
}

func testGetCookie(t *testing.T, value string, exists bool) echo.HandlerFunc {
	a := assertions.New(t)
	return func(c echo.Context) error {
		var v testCookieValue
		ok, err := testCookieSettings.Get(c, &v)
		a.So(err, should.BeNil)
		a.So(ok, should.Equal, exists)
		a.So(v.Value, should.Equal, value)

		present := testCookieSettings.Exists(c)
		a.So(present, should.Equal, exists)

		return err
	}
}

func testDeleteCookie() echo.HandlerFunc {
	return func(c echo.Context) error {
		testCookieSettings.Remove(c)

		return nil
	}
}

func TestCookie(t *testing.T) {
	e := echo.New()
	a := assertions.New(t)
	blockKey := random.Bytes(32)
	hashKey := random.Bytes(64)

	e.Use(Cookies(blockKey, hashKey))

	e.GET("/set", testSetCookie(t, "test_value"))
	e.GET("/get", testGetCookie(t, "test_value", true))
	e.GET("/del", testDeleteCookie())
	e.GET("/no_cookie", testGetCookie(t, "", false))

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

		e.ServeHTTP(rec, req)

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
