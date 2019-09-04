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

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestRedirectToHost(t *testing.T) {
	a := assertions.New(t)
	e := echo.New()

	redirectHandler := RedirectToHost("example.com")(handler)

	{
		req := httptest.NewRequest("GET", "https://example.com/", nil)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		err := redirectHandler(c)

		a.So(err, should.BeNil)
		a.So(rec.Code, should.Equal, http.StatusOK)
	}

	{
		req := httptest.NewRequest("GET", "http://internal.cluster/", nil)
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("X-Forwarded-Host", "example.com")
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		err := redirectHandler(c)

		a.So(err, should.BeNil)
		a.So(rec.Code, should.Equal, http.StatusOK)
	}

	{
		req := httptest.NewRequest("GET", "http://otherexample.com/", nil)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		err := redirectHandler(c)

		a.So(err, should.BeNil)
		a.So(rec.Code, should.Equal, http.StatusFound)
		a.So(rec.Header().Get("Location"), should.Equal, "http://example.com/")
	}

	{
		req := httptest.NewRequest("GET", "http://example.com/", nil)
		req.Header.Set("X-Forwarded-Host", "otherexample.com")
		req.Header.Set("X-Forwarded-Proto", "https")
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		err := redirectHandler(c)

		a.So(err, should.BeNil)
		a.So(rec.Code, should.Equal, http.StatusFound)
		a.So(rec.Header().Get("Location"), should.Equal, "https://example.com/")
	}
}

func TestRedirectToHTTPS(t *testing.T) {
	a := assertions.New(t)
	e := echo.New()

	redirectHandler := RedirectToHTTPS(map[int]int{
		80:   443,
		1885: 8885,
	})(handler)

	{
		req := httptest.NewRequest("GET", "https://example.com/", nil)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		err := redirectHandler(c)

		a.So(err, should.BeNil)
		a.So(rec.Code, should.Equal, http.StatusOK)
	}

	{
		req := httptest.NewRequest("GET", "http://internal.cluster/", nil)
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Header.Set("X-Forwarded-Host", "example.com")
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		err := redirectHandler(c)

		a.So(err, should.BeNil)
		a.So(rec.Code, should.Equal, http.StatusOK)
	}

	{
		req := httptest.NewRequest("GET", "http://example.com/", nil)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		err := redirectHandler(c)

		a.So(err, should.BeNil)
		a.So(rec.Code, should.Equal, http.StatusPermanentRedirect)
		a.So(rec.Header().Get("Location"), should.Equal, "https://example.com/")
	}

	{
		req := httptest.NewRequest("GET", "http://internal.cluster/", nil)
		req.Header.Set("X-Forwarded-Proto", "http")
		req.Header.Set("X-Forwarded-Host", "example.com")
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		err := redirectHandler(c)

		a.So(err, should.BeNil)
		a.So(rec.Code, should.Equal, http.StatusPermanentRedirect)
		a.So(rec.Header().Get("Location"), should.Equal, "https://example.com/")
	}

	{
		req := httptest.NewRequest("GET", "http://example.com:1885/", nil)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		err := redirectHandler(c)

		a.So(err, should.BeNil)
		a.So(rec.Code, should.Equal, http.StatusPermanentRedirect)
		a.So(rec.Header().Get("Location"), should.Equal, "https://example.com:8885/")
	}
}
