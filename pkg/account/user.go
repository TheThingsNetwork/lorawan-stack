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

	"go.thethings.network/lorawan-stack/v3/pkg/account/store"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/pbkdf2"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/oauth"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/webhandlers"
)

var tokenHashSettings auth.HashValidator = pbkdf2.PBKDF2{
	Iterations: 1000,
	KeyLength:  32,
	Algorithm:  pbkdf2.Sha256,
	SaltLength: 16,
}

func (s *server) CurrentUser(w http.ResponseWriter, r *http.Request) {
	r, session, err := s.session.Get(w, r)
	if err != nil {
		webhandlers.Error(w, r, err)
		return
	}
	r, user, err := s.session.GetUser(w, r)
	if err != nil {
		webhandlers.Error(w, r, err)
		return
	}
	safeUser := user.PublicSafe()
	userJSON, err := jsonpb.TTN().Marshal(safeUser)
	if err != nil {
		webhandlers.Error(w, r, err)
		return
	}
	webhandlers.JSON(w, r, struct {
		User       json.RawMessage `json:"user"`
		LoggedInAt *time.Time      `json:"logged_in_at"`
		SessionId  string          `json:"session_id"`
	}{
		User:       userJSON,
		LoggedInAt: ttnpb.StdTime(session.CreatedAt),
		SessionId:  session.SessionId,
	})
}

var (
	errMissingUserID             = errors.DefineInvalidArgument("missing_user_id", "missing user_id")
	errMissingPassword           = errors.DefineInvalidArgument("missing_password", "missing password")
	errIncorrectPasswordOrUserID = errors.DefineInvalidArgument("no_user_id_password_match", "incorrect password or user ID")
)

type loginRequest struct {
	UserID   string `json:"user_id" schema:"user_id"`
	Password string `json:"password" schema:"password"`
}

// ValidateContext validates the login request.
func (req *loginRequest) ValidateContext(ctx context.Context) error {
	if strings.TrimSpace(req.UserID) == "" {
		return errMissingUserID.New()
	}
	if strings.TrimSpace(req.Password) == "" {
		return errMissingPassword.New()
	}
	return (&ttnpb.UserIdentifiers{
		UserId: req.UserID,
	}).ValidateFields("user_id")
}

var errParse = errors.DefineAborted("parse", "request body parsing")

func (s *server) Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest loginRequest
	switch r.Header.Get("Content-Type") {
	case "application/json":
		if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
			webhandlers.Error(w, r, errParse.WithCause(err))
			return
		}
	default:
		if err := r.ParseForm(); err != nil {
			webhandlers.Error(w, r, errParse.WithCause(err))
			return
		}
		if err := s.schemaDecoder.Decode(&loginRequest, r.Form); err != nil {
			webhandlers.Error(w, r, errParse.WithCause(err))
			return
		}
	}
	ctx := r.Context()
	if err := loginRequest.ValidateContext(ctx); err != nil {
		webhandlers.Error(w, r, err)
		return
	}
	if err := s.session.DoLogin(ctx, loginRequest.UserID, loginRequest.Password); err != nil {
		webhandlers.Error(w, r, err)
		return
	}
	if err := s.CreateUserSession(w, r, &ttnpb.UserIdentifiers{UserId: loginRequest.UserID}); err != nil {
		webhandlers.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type tokenLoginRequest struct {
	Token string `json:"token" schema:"token"`
}

var errMissingToken = errors.DefineInvalidArgument("missing_token", "missing token")

// ValidateContext validates the token login request.
func (req *tokenLoginRequest) ValidateContext(ctx context.Context) error {
	if strings.TrimSpace(req.Token) == "" {
		return errMissingToken.New()
	}
	return nil
}

func (s *server) TokenLogin(w http.ResponseWriter, r *http.Request) {
	var tokenLoginRequest tokenLoginRequest
	switch r.Header.Get("Content-Type") {
	case "application/json":
		if err := json.NewDecoder(r.Body).Decode(&tokenLoginRequest); err != nil {
			webhandlers.Error(w, r, errParse.WithCause(err))
			return
		}
	default:
		if err := r.ParseForm(); err != nil {
			webhandlers.Error(w, r, errParse.WithCause(err))
			return
		}
		if err := s.schemaDecoder.Decode(&tokenLoginRequest, r.Form); err != nil {
			webhandlers.Error(w, r, errParse.WithCause(err))
			return
		}
	}
	ctx := r.Context()
	if err := tokenLoginRequest.ValidateContext(ctx); err != nil {
		webhandlers.Error(w, r, err)
		return
	}
	var loginToken *ttnpb.LoginToken
	err := s.store.Transact(ctx, func(ctx context.Context, st store.Interface) (err error) {
		loginToken, err = st.ConsumeLoginToken(ctx, tokenLoginRequest.Token)
		return err
	})
	if err != nil {
		webhandlers.Error(w, r, err)
		return
	}
	if err := s.CreateUserSession(w, r, loginToken.GetUserIds()); err != nil {
		webhandlers.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *server) CreateUserSession(w http.ResponseWriter, r *http.Request, userIDs *ttnpb.UserIdentifiers) error {
	ctx := r.Context()
	tokenSecret, err := auth.GenerateKey(ctx)
	if err != nil {
		return err
	}
	hashedSecret, err := auth.Hash(auth.NewContextWithHashValidator(ctx, tokenHashSettings), tokenSecret)
	if err != nil {
		return err
	}
	var session *ttnpb.UserSession
	err = s.store.Transact(ctx, func(ctx context.Context, st store.Interface) error {
		session, err = st.CreateSession(ctx, &ttnpb.UserSession{
			UserIds:       userIDs,
			SessionSecret: hashedSecret,
		})
		return err
	})
	if err != nil {
		return err
	}
	events.Publish(oauth.EvtUserLogin.NewWithIdentifiersAndData(ctx, userIDs, nil))
	return s.session.UpdateAuthCookie(w, r, func(cookie *auth.CookieShape) error {
		cookie.UserID = session.GetUserIds().GetUserId()
		cookie.SessionID = session.SessionId
		cookie.SessionSecret = tokenSecret
		return nil
	})
}

func (s *server) Logout(w http.ResponseWriter, r *http.Request) {
	r, session, err := s.session.Get(w, r)
	if err != nil {
		webhandlers.Error(w, r, err)
		return
	}
	ctx := r.Context()
	events.Publish(oauth.EvtUserLogout.NewWithIdentifiersAndData(ctx, session.GetUserIds(), nil))
	err = s.store.Transact(ctx, func(ctx context.Context, st store.Interface) error {
		return st.DeleteSession(ctx, session.GetUserIds(), session.SessionId)
	})
	if err != nil {
		webhandlers.Error(w, r, err)
		return
	}
	s.session.RemoveAuthCookie(w, r)
	w.WriteHeader(http.StatusNoContent)
}
