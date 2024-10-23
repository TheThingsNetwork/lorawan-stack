// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package identityserver

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (is *IdentityServer) listUserSessions(ctx context.Context, req *ttnpb.ListUserSessionsRequest) (sessions *ttnpb.UserSessions, err error) {
	if err := rights.RequireUser(ctx, req.GetUserIds(), ttnpb.Right_RIGHT_USER_ALL); err != nil {
		return nil, err
	}
	ctx = store.WithOrder(ctx, req.Order)
	var total uint64
	paginateCtx := store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	sessions = &ttnpb.UserSessions{}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		sessions.Sessions, err = st.FindSessions(paginateCtx, req.GetUserIds())
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, session := range sessions.Sessions {
		// ListUserSessionsRequest doesn't have a FieldMask, so we need to manually remove the secret.
		session.SessionSecret = ""
	}
	return sessions, nil
}

func (is *IdentityServer) deleteUserSession(ctx context.Context, req *ttnpb.UserSessionIdentifiers) (*emptypb.Empty, error) {
	if err := rights.RequireUser(ctx, req.GetUserIds(), ttnpb.Right_RIGHT_USER_ALL); err != nil {
		return nil, err
	}
	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		return st.DeleteSession(ctx, req.GetUserIds(), req.GetSessionId())
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

type userSessionRegistry struct {
	ttnpb.UnimplementedUserSessionRegistryServer

	*IdentityServer
}

func (ur *userSessionRegistry) List(ctx context.Context, req *ttnpb.ListUserSessionsRequest) (*ttnpb.UserSessions, error) {
	return ur.listUserSessions(ctx, req)
}

func (ur *userSessionRegistry) Delete(ctx context.Context, req *ttnpb.UserSessionIdentifiers) (*emptypb.Empty, error) {
	return ur.deleteUserSession(ctx, req)
}
