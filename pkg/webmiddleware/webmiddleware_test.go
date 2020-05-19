// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package webmiddleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	. "go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
)

func TestChain(t *testing.T) {
	a := assertions.New(t)

	var layers []string
	middleware := []MiddlewareFunc{
		func(next http.Handler) http.Handler {
			layers = append(layers, "outer")
			return next
		},
		func(next http.Handler) http.Handler {
			layers = append(layers, "inner")
			return next
		},
	}

	var handlerCalled bool
	chain := Chain(middleware, http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		handlerCalled = true
	}))
	chain.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	a.So(layers, should.Resemble, []string{"outer", "inner"})
	a.So(handlerCalled, should.BeTrue)
}
