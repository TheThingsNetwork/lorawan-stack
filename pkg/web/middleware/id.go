// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package middleware

import (
	"github.com/labstack/echo"
	uuid "github.com/satori/go.uuid"
)

// ID adds a request id to the request.
func ID(prefix string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("X-Request-ID", genID(prefix))
			return next(c)
		}
	}
}

func genID(prefix string) string {
	uid := uuid.NewV4().String()
	if prefix == "" {
		return uid
	}

	return prefix + "." + uid
}
