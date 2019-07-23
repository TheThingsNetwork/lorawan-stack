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
	"encoding/gob"
	"net/http"
	"time"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/web/cookie"
)

func init() {
	gob.Register(authCookie{})
}

const authCookieName = "_console_auth"

// AuthCookie returns a new authCookie.
func (console *Console) AuthCookie() *cookie.Cookie {
	return &cookie.Cookie{
		Name:     authCookieName,
		HTTPOnly: true,
		Path:     console.config.UI.MountPath(),
	}
}

// authCookie is the shape of the authentication cookie.
type authCookie struct {
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

func (console *Console) getAuthCookie(c echo.Context) (authCookie, error) {
	value := authCookie{}
	ok, err := console.AuthCookie().Get(c, &value)
	if err != nil {
		return authCookie{}, err
	}
	if !ok {
		return authCookie{}, echo.NewHTTPError(http.StatusUnauthorized, "No auth cookie")
	}
	return value, nil
}

func (console *Console) setAuthCookie(c echo.Context, value authCookie) error {
	return console.AuthCookie().Set(c, value)
}

func (console *Console) removeAuthCookie(c echo.Context) {
	console.AuthCookie().Remove(c)
}
