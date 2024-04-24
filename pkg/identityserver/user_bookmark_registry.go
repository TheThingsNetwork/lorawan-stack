// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

func (is *IdentityServer) createUserBookmark(
	ctx context.Context, req *ttnpb.CreateUserBookmarkRequest,
) (*ttnpb.UserBookmark, error) {
	if err := rights.RequireUser(ctx, req.UserIds, ttnpb.Right_RIGHT_USER_SETTINGS_BASIC); err != nil {
		return nil, err
	}

	bookmark, err := is.store.CreateBookmark(ctx, &ttnpb.UserBookmark{
		UserIds:   req.UserIds,
		EntityIds: req.EntityIds,
	})
	if err != nil {
		return nil, err
	}

	return bookmark, nil
}

func (is *IdentityServer) listUserBookmarks(
	ctx context.Context, req *ttnpb.ListUserBookmarksRequest,
) (*ttnpb.UserBookmarks, error) {
	err := rights.RequireUser(ctx, req.UserIds, ttnpb.Right_RIGHT_USER_INFO)
	if err != nil {
		return nil, err
	}

	if req.Deleted {
		ctx = store.WithSoftDeleted(ctx, true)
	}

	var total uint64
	ctx = store.WithOrder(ctx, req.Order)
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()

	bookmarks, err := is.store.FindBookmarks(ctx, req.UserIds, req.EntityTypes...)
	if err != nil {
		return nil, err
	}

	return &ttnpb.UserBookmarks{Bookmarks: bookmarks}, nil
}

func (is *IdentityServer) deleteUserBookmark(
	ctx context.Context, req *ttnpb.DeleteUserBookmarkRequest,
) (*emptypb.Empty, error) {
	if err := rights.RequireUser(ctx, req.UserIds, ttnpb.Right_RIGHT_USER_SETTINGS_BASIC); err != nil {
		return nil, err
	}

	err := is.store.PurgeBookmark(ctx, &ttnpb.UserBookmark{
		UserIds:   req.UserIds,
		EntityIds: req.EntityIds,
	})
	if err != nil {
		return nil, err
	}

	return ttnpb.Empty, nil
}

func (is *IdentityServer) batchDeleteUserBookmarks(
	ctx context.Context, req *ttnpb.BatchDeleteUserBookmarksRequest,
) (*emptypb.Empty, error) {
	if err := rights.RequireUser(ctx, req.UserIds, ttnpb.Right_RIGHT_USER_SETTINGS_BASIC); err != nil {
		return nil, err
	}

	_, err := is.store.BatchPurgeBookmarks(ctx, req.UserIds, req.EntityIds)
	if err != nil {
		return nil, err
	}

	return ttnpb.Empty, nil
}

type userBookmarkRegistry struct {
	ttnpb.UnimplementedUserBookmarkRegistryServer

	*IdentityServer
}

func (ubr *userBookmarkRegistry) Create(
	ctx context.Context, req *ttnpb.CreateUserBookmarkRequest,
) (*ttnpb.UserBookmark, error) {
	return ubr.createUserBookmark(ctx, req)
}

func (ubr *userBookmarkRegistry) List(
	ctx context.Context, req *ttnpb.ListUserBookmarksRequest,
) (*ttnpb.UserBookmarks, error) {
	return ubr.listUserBookmarks(ctx, req)
}

func (ubr *userBookmarkRegistry) Delete(
	ctx context.Context, req *ttnpb.DeleteUserBookmarkRequest,
) (*emptypb.Empty, error) {
	return ubr.deleteUserBookmark(ctx, req)
}

func (ubr *userBookmarkRegistry) BatchDelete(
	ctx context.Context, req *ttnpb.BatchDeleteUserBookmarksRequest,
) (*emptypb.Empty, error) {
	return ubr.batchDeleteUserBookmarks(ctx, req)
}
