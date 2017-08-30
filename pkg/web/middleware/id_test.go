// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

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

func TestID(t *testing.T) {
	a := assertions.New(t)
	id := newID("")

	// mock time to a fixed timestamp to emulate two simultaneous requests
	n := time.Now()
	now = func() time.Time { return n }

	var id1 string
	var id2 string
	var err error

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		id1, err = id.generate()
		a.So(err, should.BeNil)
		wg.Done()
	}()

	go func() {
		id2, err = id.generate()
		a.So(err, should.BeNil)
		wg.Done()
	}()

	wg.Wait()

	a.So(id1, should.NotEqual, id2)
}
