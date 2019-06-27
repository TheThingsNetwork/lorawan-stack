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

	echo "github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

func (console *Console) freshToken(c echo.Context) (*oauth2.Token, error) {
	value, err := console.getAuthCookie(c)
	if err != nil {
		return nil, err
	}

	token := &oauth2.Token{
		AccessToken:  value.AccessToken,
		RefreshToken: value.RefreshToken,
		Expiry:       value.Expiry,
	}

	freshToken, err := console.oauth.TokenSource(c.Request().Context(), token).Token()
	if err != nil {
		return nil, err
	}

	if freshToken.AccessToken != token.AccessToken {
		err = console.setAuthCookie(c, authCookie{
			AccessToken:  freshToken.AccessToken,
			RefreshToken: freshToken.RefreshToken,
			Expiry:       freshToken.Expiry,
		})
		if err != nil {
			return nil, err
		}
	}

	return freshToken, nil
}

// Token is a handler that returns a valid OAuth token.
// It reads the token from the authorization cookie and refreshes it if needed.
// If the cookie is not there, it returns a 401 Unauthorized error.
func (console *Console) Token(c echo.Context) error {
	token, err := console.freshToken(c)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, struct {
		AccessToken string    `json:"access_token"`
		Expiry      time.Time `json:"expiry"`
	}{
		AccessToken: token.AccessToken,
		Expiry:      token.Expiry,
	})
}
