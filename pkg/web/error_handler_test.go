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
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func errorHandler(c echo.Context) error {
	return errors.New("This handler throws an error")
}

func httpErrorHandler(c echo.Context) error {
	return echo.NewHTTPError(http.StatusNotImplemented, "Not implemented")
}

func TestErrorHandler(t *testing.T) {
	a := assertions.New(t)
	e := echo.New()

	e.HTTPErrorHandler = ErrorHandler

	e.GET("/error", errorHandler)
	e.GET("/httperror", httpErrorHandler)

	{
		req := httptest.NewRequest(echo.GET, "/error", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		resp := rec.Result()

		body, _ := ioutil.ReadAll(resp.Body)
		a.So(string(body), should.ContainSubstring, "This handler throws an error")
	}

	{
		req := httptest.NewRequest(echo.GET, "/httperror", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		resp := rec.Result()

		body, _ := ioutil.ReadAll(resp.Body)
		a.So(resp.StatusCode, should.Equal, http.StatusNotImplemented)
		a.So(string(body), should.ContainSubstring, "Not implemented")
	}
}
