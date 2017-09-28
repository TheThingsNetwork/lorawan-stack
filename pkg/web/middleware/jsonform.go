// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package middleware

import (
	"fmt"
	"net/url"

	"github.com/labstack/echo"
)

// JSONForm is a best-effort to bind json bodies into form values.
// This is useful for libraries that expect the form values to be set.
func JSONForm(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		typ := c.Request().Header.Get("Content-Type")
		if typ == echo.MIMEApplicationJSON || typ == echo.MIMEApplicationJSONCharsetUTF8 {
			res := make(map[string]interface{})
			err := c.Bind(&res)
			if err == nil {
				if c.Request().Form == nil {
					c.Request().Form = make(url.Values)
				}
				bind(res, c.Request().Form)
			}
		}

		return next(c)
	}
}

func bind(m map[string]interface{}, vals url.Values) {
	for k, v := range m {
		vals.Add(k, fmt.Sprintf("%v", v))
	}
}
