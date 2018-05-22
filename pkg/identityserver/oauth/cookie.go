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

package oauth

import (
	"encoding/gob"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/web/cookie"
)

func init() {
	gob.Register(authCookie{})
}

const authCookieName = "_is_auth"

// AuthCookie returns the cookie parameters for the auth cookie.
func (s *Server) AuthCookie() *cookie.Cookie {
	return &cookie.Cookie{
		Name:     authCookieName,
		Path:     s.config.mount,
		HTTPOnly: true,
	}
}

// authCookie is the format of the auth cookie.
type authCookie struct {
	// UserID is the ID of the logged in user.
	UserID string `json:"user_id"`

	// LoggedIn is the time the user logged in.
	LoggedIn time.Time `json:"logged_in"`
}

// getCookie reads the auth cookie on the request.
func (s *Server) getCookie(c echo.Context) (authCookie, error) {
	value := authCookie{}
	ok, err := s.AuthCookie().Get(c, &value)
	if err != nil {
		return value, errors.NewWithCause(err, "Failed to get auth cookie")
	}

	if !ok {
		return value, echo.NewHTTPError(http.StatusForbidden, "Not authorized")
	}

	return value, nil
}

// setCookie updates the auth cookie on the request.
func (s *Server) setCookie(c echo.Context, value authCookie) error {
	return s.AuthCookie().Set(c, value)
}

// updateCookie updates the auth cookie using the update function. It returns
// errors from the update fn.
func (s *Server) updateCookie(c echo.Context, fn func(value *authCookie) error) error {
	d := s.AuthCookie()

	value := new(authCookie)
	ok, err := d.Get(c, value)
	if err != nil {
		return errors.NewWithCause(err, "Failed to get auth cookie")
	}

	if !ok {
		value = &authCookie{}
	}

	err = fn(value)
	if err != nil {
		return err
	}

	return d.Set(c, *value)
}

// removeCookie removes the auth cookie.
func (s *Server) removeCookie(c echo.Context) {
	s.AuthCookie().Remove(c)
}
