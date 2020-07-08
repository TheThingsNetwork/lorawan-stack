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

	"go.thethings.network/lorawan-stack/v3/pkg/auth"
)

// CookieAuth extracts the auth cookie and forwards it
// to the Authorization header.
func CookieAuth(cookieName string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "" {
				next.ServeHTTP(w, r)
				return
			}
			cookieValue, err := r.Cookie(cookieName)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			sc, err := GetSecureCookie(r.Context())
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			authCookie := &auth.CookieShape{}
			err = sc.Decode(cookieName, cookieValue.Value, authCookie)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			if authCookie.SessionSecret != "" {
				key := auth.JoinToken(auth.SessionToken, authCookie.SessionID, authCookie.SessionSecret)
				r.Header.Set("Authorization", "Bearer "+key)
			}
			next.ServeHTTP(w, r)
		})
	}
}
