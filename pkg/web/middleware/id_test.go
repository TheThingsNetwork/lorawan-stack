// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func handler(c echo.Context) error {
	return c.String(http.StatusOK, "OK!")
}

func TestLog(t *testing.T) {
	a := assertions.New(t)

	e := echo.New()

	prefix := "prefix"
	handler := ID(prefix)(handler)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	err := handler(c)
	a.So(err, should.BeNil)
	a.So(rec.Code, should.Equal, http.StatusOK)

	id := rec.Header().Get("X-Request-Id")

	a.So(id, should.NotBeEmpty)
	a.So(id, should.StartWith, prefix+".")
}

func TestLogEmptyPrefix(t *testing.T) {
	a := assertions.New(t)

	e := echo.New()

	prefix := ""
	handler := ID(prefix)(handler)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	err := handler(c)
	a.So(err, should.BeNil)
	a.So(rec.Code, should.Equal, http.StatusOK)

	id := rec.Header().Get("X-Request-Id")

	a.So(id, should.NotBeEmpty)
	a.So(id, should.NotStartWith, ".")
}
