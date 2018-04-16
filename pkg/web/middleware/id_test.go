// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

func TestIDFromRequest(t *testing.T) {
	a := assertions.New(t)

	e := echo.New()

	prefix := "prefix"
	handler := ID(prefix)(handler)

	custom := "custom-id"

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Request-ID", custom)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	err := handler(c)
	a.So(err, should.BeNil)
	a.So(rec.Code, should.Equal, http.StatusOK)

	id := rec.Header().Get("X-Request-Id")

	a.So(id, should.NotBeEmpty)
	a.So(id, should.Equal, custom)
}
