// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package middleware

import "github.com/labstack/echo"

// Noop is middleware that does nothing.
func Noop(next echo.HandlerFunc) echo.HandlerFunc {
	return next
}
