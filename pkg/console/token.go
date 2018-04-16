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

package console

import (
	"net/http"

	"github.com/labstack/echo"
)

// Token is the handler that allows the user to get their OAuth token.
// It reads the token from the authorization cookie. If the cookie is not there,
// it returns a 401 Unauthorized error.
func (console *Console) Token(c echo.Context) error {
	value, err := console.getAuthCookie(c)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, value)
}
