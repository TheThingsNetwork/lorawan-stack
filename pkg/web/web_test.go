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

package web

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/random"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/web/middleware"
)

func handler(c echo.Context) error {
	return nil
}

func TestGroup(t *testing.T) {
	a := assertions.New(t)
	s, err := New(test.Context(), config.HTTP{})
	if !a.So(err, should.BeNil) {
		t.Fatal("Could not create a web instance")
	}

	a.So(s.server, should.NotHaveRoute, "GET", "/")
	s.GET("/", handler)
	a.So(s.server, should.HaveRoute, "GET", "/")

	a.So(s.server, should.NotHaveRoute, "POST", "/bar")
	s.POST("/bar", handler)
	a.So(s.server, should.NotHaveRoute, "GET", "/bar")
	a.So(s.server, should.HaveRoute, "POST", "/bar")

	{
		grp := s.Group("/")
		grp.GET("/baz", handler)
		a.So(s.server, should.HaveRoute, "GET", "/baz")
	}

	{
		grp := s.Group("/group")
		grp.GET("/g", handler)
		a.So(s.server, should.HaveRoute, "GET", "/group/g")

		ggrp := grp.Group("/quu")
		ggrp.GET("/q", handler)
		a.So(s.server, should.HaveRoute, "GET", "/group/quu/q")
	}
}

func TestIsZeros(t *testing.T) {
	a := assertions.New(t)
	{
		res := isZeros([]byte{0, 0, 0, 0, 0})
		a.So(res, should.BeTrue)
	}
	{
		res := isZeros([]byte{0, 0, 0, 1, 0})
		a.So(res, should.BeFalse)
	}
}

func TestServeHTTP(t *testing.T) {
	a := assertions.New(t)
	s, err := New(test.Context(), config.HTTP{})
	if !a.So(err, should.BeNil) {
		t.Fatal("Could not create a web instance")
	}

	// HTTP server returns 200 on valid route
	{
		req := httptest.NewRequest(echo.GET, "/", nil)
		rec := httptest.NewRecorder()

		s.GET("/", handler)

		s.ServeHTTP(rec, req)

		resp := rec.Result()
		a.So(resp.StatusCode, should.Equal, http.StatusOK)
	}
}

func TestRootGroup(t *testing.T) {
	a := assertions.New(t)
	s, err := New(test.Context(), config.HTTP{})
	if !a.So(err, should.BeNil) {
		t.Fatal("Could not create a web instance")
	}

	s.RootGroup("/sub").GET("/some", handler)
	a.So(s.server, should.NotHaveRoute, "GET", "/")
	a.So(s.server, should.NotHaveRoute, "GET", "/sub")
	a.So(s.server, should.HaveRoute, "GET", "/sub/some")
	a.So(s.server, should.NotHaveRoute, "GET", "/sub/another")
}

func TestStatic(t *testing.T) {
	a := assertions.New(t)
	s, err := New(test.Context(), config.HTTP{})
	if !a.So(err, should.BeNil) {
		t.Fatal("Could not create a web instance")
	}

	dir, err := os.Getwd()

	if !a.So(err, should.BeNil) {
		t.Fatal("Could not create resolve testing directory")
	}

	s.Static("/assets", http.Dir(dir), middleware.Immutable)

	// HTTP server returns 200 on valid file request
	{
		req := httptest.NewRequest(echo.GET, "/assets/web_test.go", nil)
		rec := httptest.NewRecorder()

		s.ServeHTTP(rec, req)

		resp := rec.Result()
		body, _ := ioutil.ReadAll(resp.Body)

		a.So(resp.StatusCode, should.Equal, http.StatusOK)
		a.So(strings.HasPrefix(string(body), "//"), should.BeTrue)
	}

	// HTTP server returns 404 on invalid file request
	{
		req := httptest.NewRequest(echo.GET, "/assets/null.txt", nil)
		rec := httptest.NewRecorder()

		s.Static("/assets", http.Dir(dir+"/teststatic"), middleware.Immutable)

		s.ServeHTTP(rec, req)

		resp := rec.Result()
		a.So(resp.StatusCode, should.Equal, http.StatusNotFound)
	}
}

func TestCookies(t *testing.T) {
	a := assertions.New(t)
	// Errors on illegal hash key byte size
	{
		c := config.HTTP{}
		c.Cookie.HashKey = random.Bytes(2)

		_, err := New(test.Context(), c)

		a.So(err, should.NotBeNil)
	}
	// Errors on non 32bit block key
	{
		c := config.HTTP{}
		c.Cookie.BlockKey = random.Bytes(31)

		_, err := New(test.Context(), c)

		a.So(err, should.NotBeNil)
	}
}
