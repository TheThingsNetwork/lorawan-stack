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

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

var (
	errOAuthError   = errors.DefineUnauthenticated("oauth_authorization", "OAuth authorization failed with `{error}`", "error_description")
	errMissingState = errors.DefineInvalidArgument("missing_state", "missing state parameter in request")
	errMissingCode  = errors.DefineInvalidArgument("missing_code", "missing code parameter in request")
	errInvalidState = errors.DefineInvalidArgument("invalid_state", "invalid state parameter in request")
)

// Callback is the OAuth callback that accepts the authorization code
// and exchanges it for an access token (+ refresh token). The access token is
// stored in the authorization cookie for later reference.
func (console *Console) Callback(c echo.Context) error {
	if e := c.QueryParam("error"); e != "" {
		return errOAuthError.WithAttributes("error", e, "error_description", c.QueryParam("error_description"))
	}

	state := c.QueryParam("state")
	if state == "" {
		return errMissingState
	}

	code := c.QueryParam("code")
	if code == "" {
		return errMissingCode
	}

	stateCookie, err := console.getStateCookie(c)
	if err != nil {
		return err
	}

	if stateCookie.Secret != state {
		return errInvalidState
	}

	// Exchange token.
	token, err := console.oauth(c).Exchange(c.Request().Context(), code)
	if err != nil {
		return err
	}

	// Remove state cookie.
	console.removeStateCookie(c)

	err = console.setAuthCookie(c, authCookie{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	})
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, console.configFromContext(c.Request().Context()).UI.CanonicalURL+stateCookie.Next)
}
