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
	"fmt"
	"net/http"
	"net/url"
	"runtime/trace"
	"time"

	"github.com/gogo/protobuf/types"
	echo "github.com/labstack/echo/v4"
	osin "github.com/openshift/osin"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/web/cookie"
)

const authCookieName = "_session"

func (s *server) authCookie() *cookie.Cookie {
	return &cookie.Cookie{
		Name:     authCookieName,
		Path:     "/",
		HTTPOnly: true,
	}
}

var errAuthCookie = errors.DefineUnauthenticated("auth_cookie", "could not get auth cookie")

func (s *server) getAuthCookie(c echo.Context) (cookie auth.CookieShape, err error) {
	ok, err := s.authCookie().Get(c, &cookie)
	if err != nil {
		return cookie, err
	}
	if !ok {
		return cookie, errAuthCookie.New()
	}
	return cookie, nil
}

func (s *server) updateAuthCookie(c echo.Context, update func(value *auth.CookieShape) error) error {
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
		return nil, errSessionExpired.New()
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
		&ttnpb.UserIdentifiers{UserID: session.UserIdentifiers.UserID},
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

var errIncorrectPasswordOrUserID = errors.DefineInvalidArgument("no_user_id_password_match", "incorrect password or user ID")

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
			return errIncorrectPasswordOrUserID.New()
		}
		return err
	}
	region := trace.StartRegion(ctx, "validate password")
	ok, err := auth.Validate(user.Password, password)
	region.End()
	if err != nil || !ok {
		events.Publish(evtUserLoginFailed(ctx, user.UserIdentifiers, nil))
		return errIncorrectPasswordOrUserID.New()
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
	tokenSecret, err := auth.GenerateKey(ctx)
	if err != nil {
		return err
	}
	hashedSecret, err := auth.Hash(ctx, tokenSecret)
	if err != nil {
		return err
	}
	session, err := s.store.CreateSession(ctx, &ttnpb.UserSession{
		UserIdentifiers: userIDs,
		SessionSecret:   hashedSecret,
	})
	if err != nil {
		return err
	}
	events.Publish(evtUserLogin(ctx, userIDs, nil))
	err = s.updateAuthCookie(c, func(cookie *auth.CookieShape) error {
		cookie.UserID = session.UserIdentifiers.UserID
		cookie.SessionID = session.SessionID
		cookie.SessionSecret = tokenSecret
		return nil
	})
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

var (
	errInvalidLogoutRedirectURI = errors.DefineInvalidArgument(
		"invalid_logout_redirect_uri",
		"the redirect URI did not match the one(s) defined in the client",
	)
	errMissingAccessTokenIDParam = errors.DefinePermissionDenied(
		"missing_param_access_token_id",
		"access token ID was not provided",
	)
)

func (s *server) ClientLogout(c echo.Context) error {
	ctx := c.Request().Context()
	accessTokenID := c.QueryParam("access_token_id")
	var redirectURI string
	if accessTokenID == "" {
		return errMissingAccessTokenIDParam
	}
	at, err := s.store.GetAccessToken(ctx, accessTokenID)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if at != nil {
		client, err := s.store.GetClient(ctx, &at.ClientIDs, &types.FieldMask{Paths: []string{"logout_redirect_uris"}})
		if err != nil {
			return err
		}
		if err = s.store.DeleteAccessToken(ctx, accessTokenID); err != nil {
			return err
		}
		events.Publish(evtUserLogout(ctx, at.UserIDs, nil))
		err = s.store.DeleteSession(ctx, &at.UserIDs, at.UserSessionID)
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
		redirectParam := c.QueryParam("post_logout_redirect_uri")
		if redirectParam == "" {
			if len(client.LogoutRedirectURIs) != 0 {
				redirectURI = client.LogoutRedirectURIs[0]
			} else {
				redirectURI = s.config.UI.MountPath()
			}
		} else {
			for _, uri := range client.LogoutRedirectURIs {
				redirectURI, err = osin.ValidateUri(uri, redirectParam)
			}
			if err != nil {
				return errInvalidLogoutRedirectURI.WithCause(err)
			}
		}
	}
	session, err := s.getSession(c)
	if err != nil && !errors.IsUnauthenticated(err) && !errors.IsNotFound(err) {
		return err
	}
	if session != nil {
		events.Publish(evtUserLogout(ctx, session.UserIdentifiers, nil))
		if err = s.store.DeleteSession(ctx, &session.UserIdentifiers, session.SessionID); err != nil {
			return err
		}
	}
	s.removeAuthCookie(c)
	url, err := url.Parse(redirectURI)
	if err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, fmt.Sprintf("%s?%s", url.Path, url.RawQuery))
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
