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
	"strings"

	echo "github.com/labstack/echo/v4"
)

// HandleLogin is the handler for redirecting the user to the authorization
// endpoint.
func (oc *OAuthClient) HandleLogin(c echo.Context) error {
	next := c.QueryParam("next")

	// Only allow relative paths.
	if !strings.HasPrefix(next, "/") && !strings.HasPrefix(next, "#") && !strings.HasPrefix(next, "?") {
		next = ""
	}

	// Set state cookie.
	state := newState(next)
	if err := oc.setStateCookie(c, state); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, oc.oauth(c).AuthCodeURL(state.Secret))
}
