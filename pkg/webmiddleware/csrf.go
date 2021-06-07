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

package webmiddleware

import (
	"net/http"
	"strings"

	"github.com/gorilla/csrf"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
)

var errInvalidCSRFToken = errors.DefinePermissionDenied("invalid_csrf_token", "invalid csrf token")

// CSRF returns a middleware that enables CSRF protection via a sync token. The
// skipCheck parameter can be used to skip CSRF protection based on the request
// interface.
func CSRF(authKey []byte, opts ...csrf.Option) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if canSkipCSRFMiddleware(r) {
				next.ServeHTTP(w, r)
				return
			}
			defaultOptions := []csrf.Option{
				csrf.SameSite(csrf.SameSiteLaxMode),
				csrf.Secure(r.URL.Scheme == "https"),
				csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// In some cases we do want to execute the CSRF middleware, so that a CSRF token is set
					// but we don't want to enforce CSRF token validation.
					if canSkipCSRFCheck(r) {
						next.ServeHTTP(w, r)
						return
					}
					webhandlers.Error(w, r, errInvalidCSRFToken.New())
				})),
			}
			handler := csrf.Protect(authKey, append(defaultOptions, opts...)...)(next)
			handler.ServeHTTP(w, r)
		})
	}
}

func canSkipCSRFMiddleware(r *http.Request) bool {
	authVal := r.Header.Get("Authorization")
	if !strings.HasPrefix(authVal, "Bearer ") {
		return false // When using an empty Authorization header, we may still want to set the CSRF token cookie.
	}
	tokenType, _, _, err := auth.SplitToken(strings.TrimPrefix(authVal, "Bearer "))
	if err != nil {
		return false // When using an unsupported Bearer token, we may still want to set the CSRF token cookie.
	}
	switch tokenType {
	case auth.APIKey, auth.AccessToken:
		return true // When the caller uses a Bearer token of type API key or Access Token, we can skip CSRF middleware.
	default:
		return false
	}
}

func canSkipCSRFCheck(r *http.Request) bool {
	authVal := r.Header.Get("Authorization")
	if !strings.HasPrefix(authVal, "Bearer ") {
		return true // Unauthenticated requests don't need CSRF protection.
	}
	_, _, _, err := auth.SplitToken(strings.TrimPrefix(authVal, "Bearer "))
	if err != nil {
		return true // Unsupported Bearer tokens don't need CSRF protection.
	}
	return false
}
