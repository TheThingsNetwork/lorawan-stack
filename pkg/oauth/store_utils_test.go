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
	"context"

	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type mockStore struct {
	store.UserStore
	store.UserSessionStore
	store.ClientStore
	store.OAuthStore

	lastCall string
	req      struct {
		ctx       context.Context
		fieldMask *types.FieldMask
		session   *ttnpb.UserSession
		sessionID string
		userIDs   *ttnpb.UserIdentifiers
	}
	res struct {
		err     error
		session *ttnpb.UserSession
		user    *ttnpb.User
	}
}

var mockErrUnauthenticated = grpc.Errorf(codes.Unauthenticated, "Unauthenticated")

func (s *mockStore) GetUser(ctx context.Context, id *ttnpb.UserIdentifiers, fieldMask *types.FieldMask) (*ttnpb.User, error) {
	s.req.ctx, s.req.userIDs, s.req.fieldMask = ctx, id, fieldMask
	s.lastCall = "GetUser"
	return s.res.user, s.res.err
}

func (s *mockStore) CreateSession(ctx context.Context, sess *ttnpb.UserSession) (*ttnpb.UserSession, error) {
	s.req.ctx, s.req.session = ctx, sess
	s.lastCall = "CreateSession"
	return s.res.session, s.res.err
}

func (s *mockStore) GetSession(ctx context.Context, userIDs *ttnpb.UserIdentifiers, sessionID string) (*ttnpb.UserSession, error) {
	s.req.ctx, s.req.userIDs, s.req.sessionID = ctx, userIDs, sessionID
	s.lastCall = "GetSession"
	return s.res.session, s.res.err
}

func (s *mockStore) DeleteSession(ctx context.Context, userIDs *ttnpb.UserIdentifiers, sessionID string) error {
	s.req.ctx, s.req.userIDs, s.req.sessionID = ctx, userIDs, sessionID
	s.lastCall = "DeleteSession"
	return s.res.err
}
