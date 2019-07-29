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

package oauth

import (
	"context"
	"encoding/json"
	"net/http"
	"runtime/trace"
	"time"

	"github.com/gogo/protobuf/types"
	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/web/cookie"
)

const authCookieName = "_session"

func (s *server) authCookie() *cookie.Cookie {
	return &cookie.Cookie{
		Name:     authCookieName,
		Path:     s.config.UI.MountPath(),
		HTTPOnly: true,
	}
}

type authCookie struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
}

var errAuthCookie = errors.DefineUnauthenticated("auth_cookie", "could not get auth cookie")

func (s *server) getAuthCookie(c echo.Context) (cookie authCookie, err error) {
	ok, err := s.authCookie().Get(c, &cookie)
	if err != nil {
		return cookie, err
	}
	if !ok {
		return cookie, errAuthCookie
	}
	return cookie, nil
}

func (s *server) updateAuthCookie(c echo.Context, update func(value *authCookie) error) error {
	cookie := &authCookie{}
	_, err := s.authCookie().Get(c, cookie)
	if err != nil {
		return err
	}
	if err = update(cookie); err != nil {
		return err
	}
	return s.authCookie().Set(c, cookie)
}

func (s *server) removeAuthCookie(c echo.Context) {
	s.authCookie().Remove(c)
}

const userSessionKey = "user_session"

var errSessionExpired = errors.DefineUnauthenticated("session_expired", "session expired")

func (s *server) getSession(c echo.Context) (*ttnpb.UserSession, error) {
	existing := c.Get(userSessionKey)
	if session, ok := existing.(*ttnpb.UserSession); ok {
		return session, nil
	}
	cookie, err := s.getAuthCookie(c)
	if err != nil {
		return nil, err
	}
	session, err := s.store.GetSession(
		c.Request().Context(),
		&ttnpb.UserIdentifiers{UserID: cookie.UserID},
		cookie.SessionID,
	)
	if err != nil {
		return nil, err
	}
	if session.ExpiresAt != nil && session.ExpiresAt.Before(time.Now()) {
		return nil, errSessionExpired
	}
	c.Set(userSessionKey, session)
	return session, nil
}

const userKey = "user"

func (s *server) getUser(c echo.Context) (*ttnpb.User, error) {
	existing := c.Get(userKey)
	if user, ok := existing.(*ttnpb.User); ok {
		return user, nil
	}
	session, err := s.getSession(c)
	if err != nil {
		return nil, err
	}
	user, err := s.store.GetUser(
		c.Request().Context(),
		&ttnpb.UserIdentifiers{UserID: session.UserID},
		nil,
	)
	if err != nil {
		return nil, err
	}
	c.Set(userKey, user)
	return user, nil
}

func (s *server) CurrentUser(c echo.Context) error {
	session, err := s.getSession(c)
	if err != nil {
		return err
	}
	user, err := s.getUser(c)
	if err != nil {
		return err
	}
	safeUser := user.PublicSafe()
	userJSON, _ := jsonpb.TTN().Marshal(safeUser)
	return c.JSON(http.StatusOK, struct {
		User       json.RawMessage `json:"user"`
		LoggedInAt time.Time       `json:"logged_in_at"`
	}{
		User:       userJSON,
		LoggedInAt: session.CreatedAt,
	})
}

type loginRequest struct {
	UserID   string `json:"user_id" form:"user_id"`
	Password string `json:"password" form:"password"`
}

var errIncorrectPasswordOrUserID = errors.DefineUnauthenticated("no_user_id_password_match", "incorrect password or user ID")

func (s *server) doLogin(ctx context.Context, userID, password string) error {
	ids := &ttnpb.UserIdentifiers{UserID: userID}
	if err := ids.ValidateContext(ctx); err != nil {
		return err
	}
	user, err := s.store.GetUser(
		ctx,
		ids,
		&types.FieldMask{Paths: []string{"password"}},
	)
	if err != nil {
		if errors.IsNotFound(err) {
			return errIncorrectPasswordOrUserID
		}
		return err
	}
	region := trace.StartRegion(ctx, "validate password")
	ok, err := auth.Password(user.Password).Validate(password)
	region.End()
	if err != nil || !ok {
		events.Publish(evtUserLoginFailed(ctx, user.UserIdentifiers, nil))
		return errIncorrectPasswordOrUserID
	}
	return nil
}

func (s *server) Login(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(loginRequest)
	if err := c.Bind(req); err != nil {
		return err
	}
	if err := s.doLogin(ctx, req.UserID, req.Password); err != nil {
		return err
	}
	userIDs := ttnpb.UserIdentifiers{UserID: req.UserID}
	session, err := s.store.CreateSession(ctx, &ttnpb.UserSession{
		UserIdentifiers: userIDs,
	})
	if err != nil {
		return err
	}
	events.Publish(evtUserLogin(ctx, userIDs, nil))
	err = s.updateAuthCookie(c, func(cookie *authCookie) error {
		cookie.UserID = session.UserID
		cookie.SessionID = session.SessionID
		return nil
	})
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (s *server) Logout(c echo.Context) error {
	ctx := c.Request().Context()
	session, err := s.getSession(c)
	if err != nil {
		return err
	}
	events.Publish(evtUserLogout(ctx, session.UserIdentifiers, nil))
	if err = s.store.DeleteSession(ctx, &session.UserIdentifiers, session.SessionID); err != nil {
		return err
	}
	s.removeAuthCookie(c)
	return c.NoContent(http.StatusNoContent)
}
