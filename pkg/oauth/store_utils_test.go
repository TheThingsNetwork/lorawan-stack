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

package oauth_test

import (
	"context"

	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type mockStoreContents struct {
	calls []string
	req   struct {
		ctx               context.Context
		fieldMask         *types.FieldMask
		session           *ttnpb.UserSession
		sessionID         string
		userIDs           *ttnpb.UserIdentifiers
		clientIDs         *ttnpb.ClientIdentifiers
		authorization     *ttnpb.OAuthClientAuthorization
		authorizationCode *ttnpb.OAuthAuthorizationCode
	}
	res struct {
		session       *ttnpb.UserSession
		user          *ttnpb.User
		client        *ttnpb.Client
		authorization *ttnpb.OAuthClientAuthorization
	}
	err struct {
		getUser                 error
		createSession           error
		getSession              error
		deleteSession           error
		getClient               error
		getAuthorization        error
		authorize               error
		createAuthorizationCode error
	}
}

type mockStore struct {
	store.UserStore
	store.UserSessionStore
	store.ClientStore
	store.OAuthStore

	mockStoreContents
}

func (s *mockStore) reset() {
	s.mockStoreContents = mockStoreContents{}
}

var (
	mockErrUnauthenticated = grpc.Errorf(codes.Unauthenticated, "Unauthenticated")
	mockErrNotFound        = grpc.Errorf(codes.NotFound, "NotFound")
)

func (s *mockStore) GetUser(ctx context.Context, id *ttnpb.UserIdentifiers, fieldMask *types.FieldMask) (*ttnpb.User, error) {
	s.req.ctx, s.req.userIDs, s.req.fieldMask = ctx, id, fieldMask
	s.calls = append(s.calls, "GetUser")
	return s.res.user, s.err.getUser
}

func (s *mockStore) CreateSession(ctx context.Context, sess *ttnpb.UserSession) (*ttnpb.UserSession, error) {
	s.req.ctx, s.req.session = ctx, sess
	s.calls = append(s.calls, "CreateSession")
	return s.res.session, s.err.createSession
}

func (s *mockStore) GetSession(ctx context.Context, userIDs *ttnpb.UserIdentifiers, sessionID string) (*ttnpb.UserSession, error) {
	s.req.ctx, s.req.userIDs, s.req.sessionID = ctx, userIDs, sessionID
	s.calls = append(s.calls, "GetSession")
	return s.res.session, s.err.getSession
}

func (s *mockStore) DeleteSession(ctx context.Context, userIDs *ttnpb.UserIdentifiers, sessionID string) error {
	s.req.ctx, s.req.userIDs, s.req.sessionID = ctx, userIDs, sessionID
	s.calls = append(s.calls, "DeleteSession")
	return s.err.deleteSession
}

func (s *mockStore) GetClient(ctx context.Context, id *ttnpb.ClientIdentifiers, fieldMask *types.FieldMask) (*ttnpb.Client, error) {
	s.req.ctx, s.req.clientIDs, s.req.fieldMask = ctx, id, fieldMask
	s.calls = append(s.calls, "GetClient")
	return s.res.client, s.err.getClient
}

func (s *mockStore) GetAuthorization(ctx context.Context, userIDs *ttnpb.UserIdentifiers, clientIDs *ttnpb.ClientIdentifiers) (*ttnpb.OAuthClientAuthorization, error) {
	s.req.ctx, s.req.userIDs, s.req.clientIDs = ctx, userIDs, clientIDs
	s.calls = append(s.calls, "GetAuthorization")
	return s.res.authorization, s.err.getAuthorization
}

func (s *mockStore) Authorize(ctx context.Context, req *ttnpb.OAuthClientAuthorization) (authorization *ttnpb.OAuthClientAuthorization, err error) {
	s.req.ctx, s.req.authorization = ctx, req
	s.calls = append(s.calls, "Authorize")
	return s.res.authorization, s.err.authorize
}

func (s *mockStore) CreateAuthorizationCode(ctx context.Context, code *ttnpb.OAuthAuthorizationCode) error {
	s.req.ctx, s.req.authorizationCode = ctx, code
	s.calls = append(s.calls, "CreateAuthorizationCode")
	return s.err.createAuthorizationCode
}
