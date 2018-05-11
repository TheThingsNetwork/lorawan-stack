// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

	"github.com/labstack/echo"
	"go.thethings.network/lorawan-stack/pkg/log"
)

// Log is middleware that logs the request.
func Log(logger log.Interface) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			stop := time.Now()

			req := c.Request()

			version := req.Header.Get("X-Version")
			if version == "" {
				version = "unknown"
				if v := req.URL.Query().Get("version"); v != "" {
					version = v
				}
			}

			status := c.Response().Status
			w := logger.WithFields(log.Fields(
				"duration", stop.Sub(start),
				"method", req.Method,
				"url", req.URL.String(),
				"remote_addr", req.RemoteAddr,
				"request_id", c.Response().Header().Get("X-Request-ID"),
				"response_size", c.Response().Size,
				"status", status,
				"version", version,
			))

			if loc := c.Response().Header().Get("Location"); status >= 300 && status < 400 && loc != "" {
				w = w.WithField("location", loc)
			}

			if fwd := req.Header.Get("X-Forwarded-For"); fwd != "" {
				w = w.WithField("forwarded_for", fwd)
			}

			if err != nil {
				w = w.WithError(err)
			}

			if status < 500 {
				w.Info("Request handled")
			} else {
				w.Error("Request error")
			}

			return err
		}
	}
}
