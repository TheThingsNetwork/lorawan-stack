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

package interop

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func handler(c echo.Context) error {
	return nil
}

func TestGroup(t *testing.T) {
	a := assertions.New(t)
	s, err := New(test.Context(), config.Interop{})
	if !a.So(err, should.BeNil) {
		t.Fatal("Could not create an interop instance")
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

func TestServeHTTP(t *testing.T) {
	a := assertions.New(t)
	s, err := New(test.Context(), config.Interop{})
	if !a.So(err, should.BeNil) {
		t.Fatal("Could not create an interop instance")
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
	s, err := New(test.Context(), config.Interop{})
	if !a.So(err, should.BeNil) {
		t.Fatal("Could not create an interop instance")
	}

	s.RootGroup("/sub")
	a.So(s.server, should.NotHaveRoute, "GET", "/")
	a.So(s.server, should.NotHaveRoute, "GET", "/sub/another")
	a.So(s.server, should.HaveRoute, "GET", "/sub")
}
