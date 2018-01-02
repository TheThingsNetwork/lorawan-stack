// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package middleware

import (
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/labstack/echo"
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
				"ip", req.RemoteAddr,
				"id", c.Response().Header().Get("X-Request-ID"),
				"size", c.Response().Size,
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
