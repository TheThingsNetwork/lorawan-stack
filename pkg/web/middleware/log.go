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
	"time"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/errors/web"
	"go.thethings.network/lorawan-stack/pkg/log"
)

// Log is middleware that logs the request.
func Log(logger log.Interface) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			logger := logger.WithFields(log.Fields(
				"method", req.Method,
				"url", req.URL.String(),
				"remote_addr", req.RemoteAddr,
				"request_id", c.Response().Header().Get("X-Request-ID"),
			))
			if fwd := req.Header.Get("X-Forwarded-For"); fwd != "" {
				logger = logger.WithField("forwarded_for", fwd)
			}

			ctx := log.NewContext(req.Context(), logger)
			c.SetRequest(req.WithContext(ctx))

			start := time.Now()
			err := next(c)
			stop := time.Now()

			logger = logger.WithFields(log.Fields(
				"duration", stop.Sub(start),
				"request_id", c.Response().Header().Get("X-Request-ID"),
			))

			var statusCode int
			if err != nil {
				var processedErr error
				statusCode, processedErr = web.ProcessError(err)
				logger = logger.WithError(processedErr)
			} else {
				res := c.Response()
				statusCode = res.Status
				logger = logger.WithField("response_size", res.Size)
				if loc := res.Header().Get("Location"); res.Status >= 300 && res.Status < 400 && loc != "" {
					logger = logger.WithField("location", loc)
				}
			}
			logger = logger.WithField("status", statusCode)

			if statusCode >= 500 {
				logger.Error("Request error")
			} else {
				logger.Info("Request handled")
			}

			return err
		}
	}
}
