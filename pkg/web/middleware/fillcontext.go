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
	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/fillcontext"
)

// FillContext fills the request context by executing the given fillers.
func FillContext(fillers ...fillcontext.Filler) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		if fillers == nil {
			return next
		}
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			for _, fill := range fillers {
				ctx = fill(ctx)
			}
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}
