// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package session

import (
	"context"
	"runtime/trace"
	"time"

	"github.com/gogo/protobuf/types"
	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/web/cookie"
)

const authCookieName = "_session"

var errIncorrectPasswordOrUserID = errors.DefineInvalidArgument("no_user_id_password_match", "incorrect password or user ID")

// Session is the session helper.
type Session struct {
	Store Store
}

// Store used by the account app server.
type Store interface {
	// UserStore and UserSessionStore are needed for user login/logout.
	store.UserStore
	store.UserSessionStore
}

func (s *Session) authCookie() *cookie.Cookie {
	return &cookie.Cookie{
		Name:     authCookieName,
		Path:     "/",
		HTTPOnly: true,
	}
}

var errAuthCookie = errors.DefineUnauthenticated("auth_cookie", "could not get auth cookie")

func (s *Session) getAuthCookie(c echo.Context) (cookie auth.CookieShape, err error) {
	ok, err := s.authCookie().Get(c, &cookie)
	if err != nil {
		return cookie, err
	}
	if !ok {
		return cookie, errAuthCookie.New()
	}
	return cookie, nil
}

// UpdateAuthCookie updates the current authentication cookie.
func (s *Session) UpdateAuthCookie(c echo.Context, update func(value *auth.CookieShape) error) error {
	cookie := &auth.CookieShape{}
	_, err := s.authCookie().Get(c, cookie)
	if err != nil {
		return err
	}
	if err = update(cookie); err != nil {
		return err
	}
	return s.authCookie().Set(c, cookie)
}

// RemoveAuthCookie deletes the authentication cookie.
func (s *Session) RemoveAuthCookie(c echo.Context) {
	s.authCookie().Remove(c)
}

const userSessionKey = "user_session"

var errSessionExpired = errors.DefineUnauthenticated("session_expired", "session expired")

// Get retrieves the current session.
func (s *Session) Get(c echo.Context) (*ttnpb.UserSession, error) {
	existing := c.Get(userSessionKey)
	if session, ok := existing.(*ttnpb.UserSession); ok {
		return session, nil
	}
	cookie, err := s.getAuthCookie(c)
	if err != nil {
		return nil, err
	}
	session, err := s.Store.GetSession(
		c.Request().Context(),
		&ttnpb.UserIdentifiers{UserID: cookie.UserID},
		cookie.SessionID,
	)
	if err != nil {
		return nil, err
	}
	if session.ExpiresAt != nil && session.ExpiresAt.Before(time.Now()) {
		return nil, errSessionExpired.New()
	}
	c.Set(userSessionKey, session)
	return session, nil
}

const userKey = "user"

// GetUser retrieves the user that is associated with the current session.
func (s *Session) GetUser(c echo.Context) (*ttnpb.User, error) {
	existing := c.Get(userKey)
	if user, ok := existing.(*ttnpb.User); ok {
		return user, nil
	}
	session, err := s.Get(c)
	if err != nil {
		return nil, err
	}
	user, err := s.Store.GetUser(
		c.Request().Context(),
		&ttnpb.UserIdentifiers{UserID: session.UserIdentifiers.UserID},
		nil,
	)
	if err != nil {
		return nil, err
	}
	c.Set(userKey, user)
	return user, nil
}

// DoLogin performs the authentication using user id and password.
func (s *Session) DoLogin(ctx context.Context, userID, password string) error {
	ids := &ttnpb.UserIdentifiers{UserID: userID}
	if err := ids.ValidateContext(ctx); err != nil {
		return err
	}
	user, err := s.Store.GetUser(
		ctx,
		ids,
		&types.FieldMask{Paths: []string{"password"}},
	)
	if err != nil {
		if errors.IsNotFound(err) {
			return errIncorrectPasswordOrUserID.New()
		}
		return err
	}
	region := trace.StartRegion(ctx, "validate password")
	ok, err := auth.Validate(user.Password, password)
	region.End()
	if err != nil || !ok {
		events.Publish(evtUserLoginFailed.NewWithIdentifiersAndData(ctx, user.UserIdentifiers, nil))
		return errIncorrectPasswordOrUserID.New()
	}
	return nil
}
