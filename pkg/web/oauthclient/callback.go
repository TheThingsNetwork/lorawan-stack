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

package oauthclient

import (
	"net/http"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

var errCallback = errors.DefineUnauthenticated("oauth_callback_error", "an error occurred: {error}")
var errNoStateParam = errors.DefinePermissionDenied("oauth_callback_no_state", "no state parameter present in request")
var errNoCode = errors.DefinePermissionDenied("oauth_callback_no_code", "no code parameter present in request")
var errInvalidState = errors.DefinePermissionDenied("oauth_callback_invalid_state", "invalid state parameter")

// HandleCallback is a handler that takes the auth code and exchanges it for the
// access token.
func (oc *OAuthClient) HandleCallback(c echo.Context) error {
	if e := c.QueryParam("error"); e != "" {
		return errCallback.WithAttributes("error", c.QueryParam("error_description"))
	}

	state := c.QueryParam("state")
	if state == "" {
		return errNoStateParam
	}

	code := c.QueryParam("code")
	if code == "" {
		return errNoCode
	}

	stateCookie, err := oc.getStateCookie(c)
	if err != nil {
		return err
	}

	if stateCookie.Secret != state {
		return errInvalidState
	}

	// Exchange token.
	token, err := oc.oauth(c).Exchange(c.Request().Context(), code)
	if err != nil {
		return err
	}

	oc.removeStateCookie(c)

	err = oc.setAuthCookie(c, authCookie{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	})
	if err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, oc.config.RootURL+stateCookie.Next)
}
