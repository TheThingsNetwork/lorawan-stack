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

	"github.com/labstack/echo"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestNormalize(t *testing.T) {
	a := assertions.New(t)
	e := echo.New()

	// Tests with permanent normalization mode
	{
		normalizeHandler := Normalize(RedirectPermanent)(handler)

		// Redirect properly with trailing slash
		{
			req := httptest.NewRequest("GET", "/test/", nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			err := normalizeHandler(c)

			a.So(err, should.BeNil)
			a.So(rec.Code, should.Equal, http.StatusMovedPermanently)
			a.So(rec.Header().Get("Location"), should.Equal, "/test")
		}
		// Do not redirect when URI is ok
		{
			req := httptest.NewRequest("GET", "/test", nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			err := normalizeHandler(c)

			a.So(err, should.BeNil)
			a.So(rec.Code, should.Equal, http.StatusOK)
			a.So("Location", should.NotBeIn, rec.Header())
		}
		// Do not normalize when uri is "/"
		{
			req := httptest.NewRequest("GET", "/", nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			err := normalizeHandler(c)

			a.So(err, should.BeNil)
			a.So(rec.Code, should.Equal, http.StatusOK)
			a.So("Location", should.NotBeIn, rec.Header())
		}
		// Set http.StatusPermanentRedirect when method != GET
		{
			req := httptest.NewRequest("POST", "/test/", nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			err := normalizeHandler(c)

			a.So(err, should.BeNil)
			a.So(rec.Code, should.Equal, http.StatusPermanentRedirect)
			a.So(rec.Header().Get("Location"), should.Equal, "/test")
		}
	}

	// Tests with temporary normalization mode
	{
		normalizeHandler := Normalize(RedirectTemporary)(handler)

		// Redirect properly with trailing slash
		{
			req := httptest.NewRequest("GET", "/test/", nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			err := normalizeHandler(c)

			a.So(err, should.BeNil)
			a.So(rec.Code, should.Equal, http.StatusFound)
			a.So(rec.Header().Get("Location"), should.Equal, "/test")
		}

		// Set http.StatusTemporaryRedirect when method != GET
		{
			req := httptest.NewRequest("POST", "/test/", nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			err := normalizeHandler(c)

			a.So(err, should.BeNil)
			a.So(rec.Code, should.Equal, http.StatusTemporaryRedirect)
			a.So(rec.Header().Get("Location"), should.Equal, "/test")
		}
	}

	// Test with continue normalization mode, setting http.StatusFound
	{
		normalizeHandler := Normalize(Continue)(handler)

		req := httptest.NewRequest("GET", "/test/", nil)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		err := normalizeHandler(c)

		a.So(err, should.BeNil)
		a.So(rec.Code, should.Equal, http.StatusFound)
		a.So(rec.Header().Get("Location"), should.Equal, "/test")
	}

	// Test with ignore normalization mode, resolving to Noop middleware
	{
		normalizeHandler := Normalize(Ignore)(handler)

		a.So(normalizeHandler, should.Equal, handler)
	}
}
