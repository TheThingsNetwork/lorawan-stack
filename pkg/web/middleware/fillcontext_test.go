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
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	echo "github.com/labstack/echo/v4"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestFillContext(t *testing.T) {
	a := assertions.New(t)

	type ctxKeyType struct{}
	var ctxKey ctxKeyType

	middleware := FillContext(func(ctx context.Context) context.Context {
		return context.WithValue(ctx, ctxKey, "value")
	})

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	err := middleware(func(c echo.Context) error {
		a.So(c.Request().Context().Value(ctxKey), should.Equal, "value")
		return nil
	})(c)
	a.So(err, should.BeNil)
	a.So(rec.Code, should.Equal, http.StatusOK)
}
