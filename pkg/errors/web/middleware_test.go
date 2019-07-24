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

package web_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/errors/web"
)

func TestErrorHandling(t *testing.T) {
	e := echo.New()

	ttnNotFound := errors.DefineNotFound("test_not_found", "test not found")

	for _, tt := range []struct {
		Name    string
		Handler echo.HandlerFunc
		Assert  func(a *assertions.Assertion, res *http.Response)
	}{
		{
			Name:    "Echo404",
			Handler: func(c echo.Context) error { return echo.ErrNotFound },
			Assert: func(a *assertions.Assertion, res *http.Response) {
				a.So(res.StatusCode, should.Equal, http.StatusNotFound)
				b, _ := ioutil.ReadAll(res.Body)
				a.So(string(b), should.ContainSubstring, `"namespace":"pkg/errors/web"`)
			},
		},
		{
			Name:    "TTN404",
			Handler: func(c echo.Context) error { return ttnNotFound },
			Assert: func(a *assertions.Assertion, res *http.Response) {
				a.So(res.StatusCode, should.Equal, http.StatusNotFound)
				b, _ := ioutil.ReadAll(res.Body)
				a.So(string(b), should.ContainSubstring, `"name":"test_not_found"`)
			},
		},
		{
			Name: "TTN404InsideEcho",
			Handler: func(c echo.Context) error {
				return &echo.HTTPError{Internal: ttnNotFound}
			},
			Assert: func(a *assertions.Assertion, res *http.Response) {
				a.So(res.StatusCode, should.Equal, http.StatusNotFound)
				b, _ := ioutil.ReadAll(res.Body)
				a.So(string(b), should.ContainSubstring, `"name":"test_not_found"`)
			},
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			a := assertions.New(t)
			req, rec := httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder()
			c := e.NewContext(req, rec)
			h := web.ErrorMiddleware(nil)(tt.Handler)
			a.So(h(c), should.Equal, nil)
			tt.Assert(a, rec.Result())
		})
	}

}
