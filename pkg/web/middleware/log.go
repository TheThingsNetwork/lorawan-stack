// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package middleware

import (
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/labstack/echo"
)

func Log(logger log.Interface) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			stop := time.Now()

			req := c.Request()

			remote := req.Header.Get("X-Forwarded-For")
			if remote == "" {
				remote = req.RemoteAddr
			}

			version := req.Header.Get("X-Version")
			if version == "" {
				version = "unknown"
				if v := req.URL.Query().Get("version"); v != "" {
					version = v
				}
			}

			status := c.Response().Status
			w := logger.WithFields(log.Fields(
				"Duration", stop.Sub(start),
				"Method", req.Method,
				"URL", req.URL.String(),
				"IP", remote,
				"ID", c.Response().Header().Get("X-Request-ID"),
				"Size", c.Response().Size,
				"Status", status,
				"Version", version,
			))

			if loc := c.Response().Header().Get("Location"); status >= 300 && status < 400 && loc != "" {
				w = w.WithField("Location", loc)
			}

			if err != nil {
				w = w.WithError(err)
			}

			if status < 500 {
				w.Info("Handled request")
			} else {
				w.Error("Server error")
			}

			return err
		}
	}
}
