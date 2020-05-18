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

package middleware

import (
	"github.com/gorilla/csrf"
	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/v3/pkg/webmiddleware"
	"go.thethings.network/lorawan-stack/v3/pkg/webui"
)

// CSRF adds cross-site request forgery protection to the handler. The sync token
// is passed to the template to be picked up by JavaScript.
func CSRF(lookupName, path string, authKey []byte) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		pass := func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				token := csrf.Token(c.Request())
				templateData := c.Get("template_data").(webui.TemplateData)
				templateData.CSRFToken = token
				c.Set("template_data", templateData)
				c.Response().Header().Set("X-CSRF-Token", token)
				return next(c)
			}
		}
		return func(c echo.Context) error {
			return echo.WrapMiddleware(
				webmiddleware.CSRF(
					authKey,
					csrf.CookieName(lookupName),
					csrf.FieldName(lookupName),
					csrf.Secure(c.Request().URL.Scheme == "https"),
					csrf.Path(path),
					csrf.SameSite(csrf.SameSiteStrictMode),
				))(
				pass(next),
			)(c)
		}
	}
}
