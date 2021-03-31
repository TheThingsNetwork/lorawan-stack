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
	"github.com/labstack/echo/v4"
)

// EchoMiddleware is an Echo middleware that rate limits HTTP requests by remote IP and request URL.
func EchoMiddleware(limiter Interface, class string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			resource := echoRequestResource(c, class)
			limit, result := limiter.RateLimit(resource)
			result.SetHTTPHeaders(c.Response().Header())
			if limit {
				return errRateLimitExceeded.WithAttributes("key", resource.Key(), "rate", result.Limit)
			}
			return next(c)
		}
	}
}
