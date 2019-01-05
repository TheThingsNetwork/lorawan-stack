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

package console

import (
	"net/http"
	"time"

	"github.com/labstack/echo"
	"golang.org/x/oauth2"
)

// RefreshToken is an echo request handler that refreshes the token currently
// stored in the authCookie and sets in the response the new authCookie.
func (console *Console) RefreshToken(c echo.Context) error {
	value, err := console.getAuthCookie(c)
	if err != nil {
		return err
	}

	// Just return if current access token if still valid.
	if time.Now().UTC().Before(value.Expiry) {
		return c.JSON(http.StatusOK, value)
	}

	token, err := console.oauth.TokenSource(c.Request().Context(), &oauth2.Token{
		AccessToken:  value.AccessToken,
		RefreshToken: value.RefreshToken,
		Expiry:       value.Expiry,
	}).Token()
	if err != nil {
		return err
	}

	fresh := authCookie{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}

	err = console.setAuthCookie(c, fresh)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, fresh)
}
