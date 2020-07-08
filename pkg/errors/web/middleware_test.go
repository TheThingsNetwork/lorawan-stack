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
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/errors/web"
)

func TestErrorHandling(t *testing.T) {
	e := echo.New()

	ttnNotFound := errors.DefineNotFound("test_not_found", "test not found")

	for _, tc := range []struct {
		Name  string
		Error error
		JSON  string
	}{
		{
			Name:  "Echo404",
			Error: echo.ErrNotFound,
			JSON: `{
  "code": 5,
  "details": [
    {
      "@type": "type.googleapis.com/ttn.lorawan.v3.ErrorDetails",
      "attributes": {
        "message": "Not Found"
      },
      "code": 5,
      "message_format": "Not Found",
      "namespace": "pkg/errors/web"
    }
  ],
  "message": "error:pkg/errors/web:unknown (Not Found)"
}`,
		},
		{
			Name:  "TTN404",
			Error: ttnNotFound,
			JSON: `{
  "code": 5,
  "details": [
    {
      "@type": "type.googleapis.com/ttn.lorawan.v3.ErrorDetails",
      "code": 5,
      "message_format": "test not found",
      "name": "test_not_found",
      "namespace": "pkg/errors/web_test"
    }
  ],
  "message": "error:pkg/errors/web_test:test_not_found (test not found)"
}`,
		},
		{
			Name:  "TTN404InsideEcho",
			Error: &echo.HTTPError{Internal: ttnNotFound},
			JSON: `{
  "code": 5,
  "details": [
    {
      "@type": "type.googleapis.com/ttn.lorawan.v3.ErrorDetails",
      "code": 5,
      "message_format": "test not found",
      "name": "test_not_found",
      "namespace": "pkg/errors/web_test"
    }
  ],
  "message": "error:pkg/errors/web_test:test_not_found (test not found)"
}`,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			req, rec := httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder()
			if !a.So(web.ErrorMiddleware(nil)(func(echo.Context) error {
				return tc.Error
			})(e.NewContext(req, rec)), should.BeNil) {
				t.FailNow()
			}

			res := rec.Result()
			b, err := ioutil.ReadAll(res.Body)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(res.StatusCode, should.Equal, http.StatusNotFound)
			a.So(string(b), should.EqualJSON, tc.JSON)
		})
	}

}
