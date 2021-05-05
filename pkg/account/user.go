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

package account

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/pbkdf2"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/oauth"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var tokenHashSettings auth.HashValidator = pbkdf2.PBKDF2{
	Iterations: 1000,
	KeyLength:  32,
	Algorithm:  pbkdf2.Sha256,
	SaltLength: 16,
}

func (s *server) CurrentUser(c echo.Context) error {
	session, err := s.session.Get(c)
	if err != nil {
		return err
	}
	user, err := s.session.GetUser(c)
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

var (
	errMissingUserID             = errors.DefineInvalidArgument("missing_user_id", "missing user_id")
	errMissingPassword           = errors.DefineInvalidArgument("missing_password", "missing password")
	errIncorrectPasswordOrUserID = errors.DefineInvalidArgument("no_user_id_password_match", "incorrect password or user ID")
)

type loginRequest struct {
	UserID   string `json:"user_id" form:"user_id"`
	Password string `json:"password" form:"password"`
}

// ValidateContext validates the login request.
func (req *loginRequest) ValidateContext(ctx context.Context) error {
	if strings.TrimSpace(req.UserID) == "" {
		return errMissingUserID.New()
	}
	if strings.TrimSpace(req.Password) == "" {
		return errMissingPassword.New()
	}
	if err := (&ttnpb.UserIdentifiers{
		UserId: req.UserID,
	}).ValidateFields("user_id"); err != nil {
		return err
	}
	return nil
}

func (s *server) Login(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(loginRequest)
	if err := c.Bind(req); err != nil {
		return err
	}
	if err := req.ValidateContext(c.Request().Context()); err != nil {
		return err
	}
	if err := s.session.DoLogin(ctx, req.UserID, req.Password); err != nil {
		return err
	}
	if err := s.CreateUserSession(c, ttnpb.UserIdentifiers{UserId: req.UserID}); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

type tokenLoginRequest struct {
	Token string `json:"token" form:"token"`
}

var errMissingToken = errors.DefineInvalidArgument("missing_token", "missing token")

// ValidateContext validates the token login request.
func (req *tokenLoginRequest) ValidateContext(ctx context.Context) error {
	if strings.TrimSpace(req.Token) == "" {
		return errMissingToken.New()
	}
	return nil
}

func (s *server) TokenLogin(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(tokenLoginRequest)
	if err := c.Bind(req); err != nil {
		return err
	}
	if err := req.ValidateContext(c.Request().Context()); err != nil {
		return err
	}
	loginToken, err := s.store.ConsumeLoginToken(ctx, req.Token)
	if err != nil {
		return err
	}
	if err := s.CreateUserSession(c, loginToken.UserIdentifiers); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (s *server) CreateUserSession(c echo.Context, userIDs ttnpb.UserIdentifiers) error {
	ctx := c.Request().Context()
	tokenSecret, err := auth.GenerateKey(ctx)
	if err != nil {
		return err
	}
	hashedSecret, err := auth.Hash(auth.NewContextWithHashValidator(ctx, tokenHashSettings), tokenSecret)
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
	events.Publish(oauth.EvtUserLogin.NewWithIdentifiersAndData(ctx, &userIDs, nil))
	return s.session.UpdateAuthCookie(c, func(cookie *auth.CookieShape) error {
		cookie.UserID = session.UserIdentifiers.UserId
		cookie.SessionID = session.SessionID
		cookie.SessionSecret = tokenSecret
		return nil
	})
}

func (s *server) Logout(c echo.Context) error {
	ctx := c.Request().Context()
	session, err := s.session.Get(c)
	if err != nil {
		return err
	}
	events.Publish(oauth.EvtUserLogout.NewWithIdentifiersAndData(ctx, &session.UserIdentifiers, nil))
	if err = s.store.DeleteSession(ctx, &session.UserIdentifiers, session.SessionID); err != nil {
		return err
	}
	s.session.RemoveAuthCookie(c)
	return c.NoContent(http.StatusNoContent)
}
