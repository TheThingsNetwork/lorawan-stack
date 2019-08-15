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
	"encoding/gob"
	"net/http"
	"time"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/web/cookie"
)

func init() {
	gob.Register(authCookie{})
}

// AuthCookie returns a new authCookie.
func (oc *OAuthClient) AuthCookie() *cookie.Cookie {
	return &cookie.Cookie{
		Name:     oc.config.AuthCookieName,
		HTTPOnly: true,
		Path:     oc.getMountPath(),
	}
}

// authCookie is the shape of the authentication cookie.
type authCookie struct {
	AccessToken  string
	RefreshToken string
	Expiry       time.Time
}

func (oc *OAuthClient) getAuthCookie(c echo.Context) (authCookie, error) {
	value := authCookie{}
	ok, err := oc.AuthCookie().Get(c, &value)
	if err != nil {
		return authCookie{}, err
	}
	if !ok {
		return authCookie{}, echo.NewHTTPError(http.StatusUnauthorized, "No auth cookie")
	}
	return value, nil
}

func (oc *OAuthClient) setAuthCookie(c echo.Context, value authCookie) error {
	return oc.AuthCookie().Set(c, value)
}

func (oc *OAuthClient) removeAuthCookie(c echo.Context) {
	oc.AuthCookie().Remove(c)
}
