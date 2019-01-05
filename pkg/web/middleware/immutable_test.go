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
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestImmutable(t *testing.T) {
	a := assertions.New(t)
	e := echo.New()

	e.GET("/", handler)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)
	err := Immutable(handler)(c)

	a.So(err, should.BeNil)
	a.So(rec.Header().Get("Cache-Control"), should.Equal, "public; max-age=365000000; immutable")
}
