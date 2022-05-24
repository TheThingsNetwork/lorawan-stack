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

	var trace []string
	middleware := []MiddlewareFunc{
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				trace = append(trace, "outer begin")
				next.ServeHTTP(w, r)
				trace = append(trace, "outer end")
			})
		},
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				trace = append(trace, "inner begin")
				next.ServeHTTP(w, r)
				trace = append(trace, "inner end")
			})
		},
	}

	chain := Chain(middleware, http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		trace = append(trace, "handler")
	}))
	chain.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	a.So(trace, should.Resemble, []string{"outer begin", "inner begin", "handler", "inner end", "outer end"})
}

func TestConditional(t *testing.T) {
	a := assertions.New(t)

	flag := false
	middleware := Conditional(
		func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				flag = true
				next.ServeHTTP(w, r)
			})
		},
		func(r *http.Request) bool {
			return r.Header.Get("X-Condition") == "foo"
		},
	)
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Condition", "bar")
	rec := httptest.NewRecorder()
	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.So(flag, should.Equal, false)
	})).ServeHTTP(rec, r)

	r = httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Condition", "foo")
	rec = httptest.NewRecorder()
	middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.So(flag, should.Equal, true)
	})).ServeHTTP(rec, r)
}
