// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/ratelimit"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func httpRequest(url string, remoteIP string) *http.Request {
	req := httptest.NewRequest("GET", url, nil)
	req.Header.Set("x-real-ip", remoteIP)
	return req
}

func TestHTTP(t *testing.T) {
	a := assertions.New(t)

	limiter := &mockLimiter{}

	const class = "http:test"
	middleware := ratelimit.HTTPMiddleware(limiter, class)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	t.Run("Pass", func(t *testing.T) {
		limiter.limit = false
		limiter.result = ratelimit.Result{Limit: 10}

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httpRequest("/path", "10.10.10.10"))

		a.So(rec.Header().Get("x-rate-limit-limit"), should.Equal, "10")
		a.So(rec.Result().StatusCode, should.Equal, http.StatusOK)

		a.So(limiter.calledWithResource.Key(), should.ContainSubstring, "/path")
		a.So(limiter.calledWithResource.Key(), should.ContainSubstring, "10.10.10.10")
		a.So(limiter.calledWithResource.Classes(), should.Resemble, []string{class, "http"})
	})

	t.Run("PathTemplate", func(t *testing.T) {
		limiter.limit = false
		limiter.result = ratelimit.Result{Limit: 10}

		restore := ratelimit.SetPathTemplate(func(r *http.Request) (string, bool) {
			return "/path/{id}", true
		})
		defer restore()

		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httpRequest("/path/123", "10.10.10.10"))

		a.So(rec.Header().Get("x-rate-limit-limit"), should.Equal, "10")
		a.So(rec.Result().StatusCode, should.Equal, http.StatusOK)

		a.So(limiter.calledWithResource.Key(), should.ContainSubstring, "/path/123")
		a.So(limiter.calledWithResource.Key(), should.ContainSubstring, "10.10.10.10")
		a.So(limiter.calledWithResource.Classes(), should.Resemble, []string{"http:test:/path/{id}", class, "http"})
	})

	t.Run("AuthToken", func(t *testing.T) {
		limiter.limit = false
		limiter.result = ratelimit.Result{Limit: 10}

		rec := httptest.NewRecorder()
		req := httpRequest("/path", "10.10.10.10").WithContext(tokenContext(authTokenID))
		handler.ServeHTTP(rec, req)

		a.So(rec.Header().Get("x-rate-limit-limit"), should.Equal, "10")
		a.So(rec.Result().StatusCode, should.Equal, http.StatusOK)

		a.So(limiter.calledWithResource.Key(), should.ContainSubstring, "/path")
		a.So(limiter.calledWithResource.Key(), should.ContainSubstring, authTokenID)
		a.So(limiter.calledWithResource.Classes(), should.Resemble, []string{class, "http"})
	})

	t.Run("Limit", func(t *testing.T) {
		limiter.limit = true
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httpRequest("/path", "10.10.10.10"))

		a.So(rec.Header().Get("x-rate-limit-limit"), should.NotBeEmpty)
		a.So(rec.Result().StatusCode, should.Equal, http.StatusTooManyRequests)
	})

	t.Run("Vary", func(t *testing.T) {
		for _, tc := range []struct {
			name   string
			req1   *http.Request
			req2   *http.Request
			assert func(any, ...any) string
		}{
			{
				name:   "ByIP",
				req1:   httpRequest("/path", "10.10.10.10"),
				req2:   httpRequest("/path", "10.10.10.11"),
				assert: should.NotEqual,
			},
			{
				name:   "ByPath",
				req1:   httpRequest("/path", "10.10.10.10"),
				req2:   httpRequest("/path/other", "10.10.10.10"),
				assert: should.NotEqual,
			},
			{
				name:   "IgnoreQueryString",
				req1:   httpRequest("/path?x=y", "10.10.10.10"),
				req2:   httpRequest("/path", "10.10.10.10"),
				assert: should.Equal,
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				getKey := func(r *http.Request) string {
					handler.ServeHTTP(httptest.NewRecorder(), r)
					return limiter.calledWithResource.Key()
				}

				assertions.New(t).So(getKey(tc.req1), tc.assert, getKey(tc.req2))
			})
		}
	})
}
