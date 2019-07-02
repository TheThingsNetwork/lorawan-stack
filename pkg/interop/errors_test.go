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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestErrorHandler(t *testing.T) {
	errTest := errors.DefineInvalidArgument("test_error_handler", "test error")

	for i, tc := range []struct {
		Error             error
		ResponseAssertion func(*testing.T, int, []byte) bool
	}{
		{
			Error: fmt.Errorf("unknown"),
			ResponseAssertion: func(t *testing.T, statusCode int, data []byte) bool {
				a := assertions.New(t)
				return a.So(statusCode, should.Equal, http.StatusInternalServerError) &&
					a.So(data, should.BeEmpty)
			},
		},
		{
			Error: errTest,
			ResponseAssertion: func(t *testing.T, statusCode int, data []byte) bool {
				a := assertions.New(t)
				return a.So(statusCode, should.Equal, http.StatusBadRequest) &&
					a.So(data, should.BeEmpty)
			},
		},
		{
			Error: ErrJoinReq.WithCause(errTest),
			ResponseAssertion: func(t *testing.T, statusCode int, data []byte) bool {
				a := assertions.New(t)
				return a.So(statusCode, should.Equal, http.StatusBadRequest) &&
					a.So(string(data), should.Equal, `{"message":"error:pkg/interop:test_error_handler (test error)"}`+"\n")
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			server := echo.New()
			server.HTTPErrorHandler = ErrorHandler
			server.POST("/", func(c echo.Context) error { return tc.Error })

			req := httptest.NewRequest(echo.POST, "/", nil)
			rec := httptest.NewRecorder()
			server.ServeHTTP(rec, req)

			res := rec.Result()
			defer res.Body.Close()
			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("Failed to read body: %v", err)
			}
			if !tc.ResponseAssertion(t, res.StatusCode, data) {
				t.Fatal("Response assertion failed")
			}
		})
	}
}
