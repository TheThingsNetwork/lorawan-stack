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
	"net"
	"net/http"
	"strconv"

	echo "github.com/labstack/echo/v4"
)

// RedirectToHTTPS redirects requests from HTTP to HTTPS.
func RedirectToHTTPS(fromToPorts map[int]int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if !c.IsTLS() && c.Request().Header.Get("X-Forwarded-Proto") != "https" {
				requestHost := c.Request().Host
				if forwardedHost := c.Request().Header.Get("X-Forwarded-Host"); forwardedHost != "" {
					requestHost = forwardedHost
				}
				host, port, err := net.SplitHostPort(requestHost)
				if err != nil {
					host = requestHost
					port = "80"
				}
				if port, err := strconv.Atoi(port); err == nil {
					if to, ok := fromToPorts[port]; ok {
						url := *c.Request().URL
						url.Scheme = "https"
						url.Host = host
						if to != 443 {
							url.Host = net.JoinHostPort(host, strconv.Itoa(to))
						}
						return c.Redirect(http.StatusPermanentRedirect, url.String())
					}
				}
			}
			return next(c)
		}
	}
}
