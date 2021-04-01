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

package ratelimit

import (
	"net"
	"net/http"

	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
)

const httpXRealIPHeader = "X-Real-IP"

func httpRemoteIP(r *http.Request) string {
	if xRealIP := r.Header.Get(httpXRealIPHeader); xRealIP != "" {
		return xRealIP
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

// HTTPMiddleware is an HTTP middleware that rate limits by remote IP and the request URL.
// The remote IP is retrieved by the X-Real-IP header. Use this middleware after webmiddleware.ProxyHeaders()
func HTTPMiddleware(limiter Interface, class string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resource := httpRequestResource(r, class)
			limit, result := limiter.RateLimit(resource)
			result.SetHTTPHeaders(w.Header())
			if limit {
				webhandlers.Error(w, r, errRateLimitExceeded.WithAttributes("key", resource.Key(), "rate", result.Limit))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
