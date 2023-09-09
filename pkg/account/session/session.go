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
	"net/http"
	"runtime/trace"
	"time"

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
	Store TransactionalStore
}

// Store used by the account app server.
type Store interface {
	// UserStore and UserSessionStore are needed for user login/logout.
	store.UserStore
	store.UserSessionStore
}

// TransactionalStore is Store, but with a method that uses a transaction.
type TransactionalStore interface {
	Store

	// Transact runs a transaction using the store.
	Transact(context.Context, func(context.Context, Store) error) error
}

func (s *Session) authCookie() *cookie.Cookie {
	return &cookie.Cookie{
		Name:     authCookieName,
		Path:     "/",
		HTTPOnly: true,
	}
}

var errAuthCookie = errors.DefineUnauthenticated("auth_cookie", "get auth cookie")

func (s *Session) getAuthCookie(w http.ResponseWriter, r *http.Request) (cookie auth.CookieShape, err error) {
	ok, err := s.authCookie().Get(w, r, &cookie)
	if err != nil {
		return cookie, err
	}
	if !ok {
		return cookie, errAuthCookie.New()
	}
	return cookie, nil
}

// UpdateAuthCookie updates the current authentication cookie.
func (s *Session) UpdateAuthCookie(w http.ResponseWriter, r *http.Request, update func(value *auth.CookieShape) error) error {
	cookie := &auth.CookieShape{}
	_, err := s.authCookie().Get(w, r, cookie)
	if err != nil {
		return err
	}
	if err := update(cookie); err != nil {
		return err
	}
	return s.authCookie().Set(w, r, cookie)
}

// RemoveAuthCookie deletes the authentication cookie.
func (s *Session) RemoveAuthCookie(w http.ResponseWriter, r *http.Request) {
	s.authCookie().Remove(w, r)
}

type userSessionKeyType struct{}

var userSessionKey userSessionKeyType

var errSessionExpired = errors.DefineUnauthenticated("session_expired", "session expired")

// Get retrieves the current session.
func (s *Session) Get(w http.ResponseWriter, r *http.Request) (*http.Request, *ttnpb.UserSession, error) {
	ctx := r.Context()
	if session, ok := ctx.Value(userSessionKey).(*ttnpb.UserSession); ok {
		return r, session, nil
	}
	cookie, err := s.getAuthCookie(w, r)
	if err != nil {
		return r, nil, err
	}
	var session *ttnpb.UserSession
	err = s.Store.Transact(ctx, func(ctx context.Context, st Store) (err error) {
		session, err = st.GetSession(
			ctx,
			&ttnpb.UserIdentifiers{UserId: cookie.UserID},
			cookie.SessionID,
		)
		return err
	})
	if err != nil {
		if errors.IsNotFound(err) {
			s.RemoveAuthCookie(w, r)
		}
		return r, nil, err
	}
	if expiresAt := ttnpb.StdTime(session.ExpiresAt); expiresAt != nil && expiresAt.Before(time.Now()) {
		s.RemoveAuthCookie(w, r)
		return r, nil, errSessionExpired.New()
	}
	return r.WithContext(context.WithValue(ctx, userSessionKey, session)), session, nil
}

type userKeyType struct{}

var userKey userKeyType

// GetUser retrieves the user that is associated with the current session.
func (s *Session) GetUser(w http.ResponseWriter, r *http.Request) (*http.Request, *ttnpb.User, error) {
	if user, ok := r.Context().Value(userKey).(*ttnpb.User); ok {
		return r, user, nil
	}
	r, session, err := s.Get(w, r)
	if err != nil {
		return r, nil, err
	}
	ctx := r.Context()
	var user *ttnpb.User
	err = s.Store.Transact(ctx, func(ctx context.Context, st Store) (err error) {
		user, err = st.GetUser(
			ctx,
			&ttnpb.UserIdentifiers{UserId: session.GetUserIds().GetUserId()},
			nil,
		)
		return err
	})
	if err != nil {
		return r, nil, err
	}
	return r.WithContext(context.WithValue(ctx, userKey, user)), user, nil
}

// DoLogin performs the authentication using user id and password.
func (s *Session) DoLogin(ctx context.Context, userID, password string) error {
	ids := &ttnpb.UserIdentifiers{UserId: userID}
	if err := ids.ValidateContext(ctx); err != nil {
		return err
	}
	var user *ttnpb.User
	err := s.Store.Transact(ctx, func(ctx context.Context, st Store) (err error) {
		user, err = st.GetUser(
			ctx,
			ids,
			[]string{"password"},
		)
		return err
	})
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
		events.Publish(evtUserLoginFailed.NewWithIdentifiersAndData(ctx, user.GetIds(), nil))
		return errIncorrectPasswordOrUserID.New()
	}
	return nil
}
