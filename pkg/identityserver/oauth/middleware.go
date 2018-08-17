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
	"fmt"
	"net/http"
	"net/url"

	"github.com/labstack/echo"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// userKey is the key where the user will be stored on the request.
const userKey = "user"

// nextKey is contains the name of the query parameter that denotes where to
// redirect after log in.
const nextKey = "n"

// getUser gets the user from the cookie and looks it up in the store,
// attaching the result to the echo context.
func (s *Server) getUser(c echo.Context) (store.User, error) {
	cookie, err := s.getCookie(c)
	if err != nil {
		return nil, err
	}

	user, err := s.config.Store.Users.GetByID(ttnpb.UserIdentifiers{UserID: cookie.UserID}, s.config.Specializers.User)
	if err != nil {
		return nil, err
	}

	c.Set(userKey, user)

	return user, nil
}

// RequireLogin is an echo middleware that requires a user to be logged in
// by checking the cookie on the request. If no user is logged in, it returns
// a 403 Forbidden error.
func (s *Server) RequireLogin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := s.getUser(c)
		if err != nil {
			return ErrNotAuthenticated.New(nil)
		}

		return next(c)
	}
}

// RedirectToLogin is an echo middleware that requires a user to be logged in
// by checking the cookie on the request. If no user is logged in, it redirects
// to the login page.
func (s *Server) RedirectToLogin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := s.getUser(c)
		if err != nil {
			values := make(url.Values)
			values.Add(nextKey, c.Request().URL.String())
			return c.Redirect(http.StatusFound, fmt.Sprintf("%s/login?%s", s.config.PublicURL, values.Encode()))
		}

		return next(c)
	}
}

// RedirectToAccount is an echo middleware that requires no user to be logged in
// by checking the cookie on the request. If a user is logged in, it redirects
// to the account page.
func (s *Server) RedirectToAccount(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := s.getUser(c)
		if err == nil {
			return c.Redirect(http.StatusFound, fmt.Sprintf("%s/account", s.config.PublicURL))
		}

		return next(c)
	}
}

// RedirectToNext is an echo middleware that requires no user to be logged in
// by checking the cookie on the request. If a user is logged in, it redirects
// to the url in the next parameter.
func (s *Server) RedirectToNext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		_, err := s.getUser(c)
		if err == nil {
			next := c.QueryParam(nextKey)
			if next == "" {
				next = "/account"
			}
			return c.Redirect(http.StatusFound, fmt.Sprintf("%s%s", s.config.PublicURL, next))
		}

		return next(c)
	}
}

// getUser gets the user stored in the echo context by one of the middlewares.
func getUser(c echo.Context) (store.User, error) {
	user, ok := c.Get(userKey).(store.User)
	if !ok {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Invalid user on request")
	}

	if user == nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "No user on request")
	}

	return user, nil
}
