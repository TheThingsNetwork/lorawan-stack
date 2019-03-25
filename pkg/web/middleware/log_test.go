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
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func errorHandler(c echo.Context) error {
	return c.String(http.StatusInternalServerError, "500")
}

func redirectHandler(c echo.Context) error {
	return c.Redirect(http.StatusMovedPermanently, "/other")
}

func forwardMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Request().Header.Set("X-Forwarded-For", "/other")
		return next(c)
	}
}

func noopHandler(c echo.Context) error { return nil }

func invalidHandler(c echo.Context) error {
	return errors.New("This handler throws an error")
}

func TestLogging(t *testing.T) {

	logger, _ := log.NewLogger(log.WithHandler(log.NoopHandler))

	messages := []log.Entry{}

	// collect is a middleware that collects the messages
	collect := log.MiddlewareFunc(func(next log.Handler) log.Handler {
		return log.HandlerFunc(func(entry log.Entry) error {
			messages = append(messages, entry)
			return next.HandleLog(entry)
		})
	})

	logger.Use(collect)

	a := assertions.New(t)
	e := echo.New()

	// Test Logging middleware
	{
		handler := Log(logger)(handler)
		{
			req := httptest.NewRequest("GET", "/", nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(req, rec)
			err := handler(c)

			a.So(err, should.BeNil)
		}

		fields := messages[0].Fields().Fields()
		a.So(len(messages), should.Equal, 1)
		a.So(messages[0].Message(), should.Equal, "Request handled")
		a.So(messages[0].Level(), should.Equal, log.InfoLevel)
		a.So(fields["method"], should.Equal, "GET")
		a.So(fields["url"], should.Equal, "/")
		a.So(fields["response_size"], should.Equal, 3)
		a.So(fields["status"], should.Equal, 200)
		a.So(fields, should.ContainKey, "duration")
		a.So(fields, should.ContainKey, "remote_addr")
		a.So(fields, should.ContainKey, "request_id")
		a.So(fields, should.ContainKey, "response_size")
		a.So(fields, should.NotContainKey, "redirect")
	}

	// Reset messages
	messages = nil

	// Test Logging middleware on error
	{
		handler := Log(logger)(errorHandler)

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		err := handler(c)

		a.So(err, should.BeNil)

		fields := messages[0].Fields().Fields()
		a.So(len(messages), should.Equal, 1)
		a.So(messages[0].Message(), should.Equal, "Request error")
		a.So(messages[0].Level(), should.Equal, log.ErrorLevel)
		a.So(fields["status"], should.Equal, 500)
	}

	// Reset messages
	messages = nil

	// Test Logging middleware on redirect
	{
		handler := Log(logger)(redirectHandler)

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		err := handler(c)

		a.So(err, should.BeNil)

		fields := messages[0].Fields().Fields()
		a.So(len(messages), should.Equal, 1)
		a.So(messages[0].Message(), should.Equal, "Request handled")
		a.So(messages[0].Level(), should.Equal, log.InfoLevel)
		a.So(fields, should.ContainKey, "location")
	}

	// Reset messages
	messages = nil

	// Test Logging middleware on forward
	{
		handler := forwardMiddleware(Log(logger)(noopHandler))
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		err := handler(c)

		a.So(err, should.BeNil)

		fields := messages[0].Fields().Fields()
		a.So(len(messages), should.Equal, 1)
		a.So(fields, should.ContainKey, "forwarded_for")
	}

	// Reset messages
	messages = nil

	// Test Logging middleware with invalid handler
	{
		handler := Log(logger)(invalidHandler)
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		c := e.NewContext(req, rec)
		err := handler(c)

		a.So(err, should.NotBeNil)

		fields := messages[0].Fields().Fields()
		a.So(len(messages), should.Equal, 1)
		a.So(fields, should.ContainKey, "error")
	}
}
