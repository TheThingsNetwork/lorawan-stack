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

package identityserver

import (
	"context"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func (is *IdentityServer) listOAuthClientAuthorizations(ctx context.Context, req *ttnpb.ListOAuthClientAuthorizationsRequest) (authorizations *ttnpb.OAuthClientAuthorizations, err error) {
	if err := rights.RequireUser(ctx, req.UserIdentifiers, ttnpb.RIGHT_USER_AUTHORIZED_CLIENTS); err != nil {
		return nil, err
	}
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	authorizations = &ttnpb.OAuthClientAuthorizations{}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		authorizations.Authorizations, err = store.GetOAuthStore(db).ListAuthorizations(ctx, &req.UserIdentifiers)
		return err
	})
	if err != nil {
		return nil, err
	}
	return authorizations, nil
}

func (is *IdentityServer) listOAuthAccessTokens(ctx context.Context, req *ttnpb.ListOAuthAccessTokensRequest) (tokens *ttnpb.OAuthAccessTokens, err error) {
	authInfo, err := is.authInfo(ctx)
	if err != nil {
		return nil, err
	}
	accessToken := authInfo.GetOAuthAccessToken()
	if accessToken == nil || accessToken.UserIDs.UserID != req.UserIDs.UserID || accessToken.ClientIDs.ClientID != req.ClientIDs.ClientID {
		if err := rights.RequireUser(ctx, req.UserIDs, ttnpb.RIGHT_USER_AUTHORIZED_CLIENTS); err != nil {
			return nil, err
		}
	}
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	tokens = &ttnpb.OAuthAccessTokens{}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		tokens.Tokens, err = store.GetOAuthStore(db).ListAccessTokens(ctx, &req.UserIDs, &req.ClientIDs)
		return err
	})
	for _, token := range tokens.Tokens {
		token.AccessToken, token.RefreshToken = "", ""
	}
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (is *IdentityServer) deleteOAuthAuthorization(ctx context.Context, req *ttnpb.OAuthClientAuthorizationIdentifiers) (*types.Empty, error) {
	if err := rights.RequireUser(ctx, req.UserIDs, ttnpb.RIGHT_USER_AUTHORIZED_CLIENTS); err != nil {
		return nil, err
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		return store.GetOAuthStore(db).DeleteAuthorization(ctx, &req.UserIDs, &req.ClientIDs)
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

var errAccessTokenMismatch = errors.DefineInvalidArgument("access_token_mismatch", "access token ID did not match user or client identifiers")

func (is *IdentityServer) deleteOAuthAccessToken(ctx context.Context, req *ttnpb.OAuthAccessTokenIdentifiers) (*types.Empty, error) {
	authInfo, err := is.authInfo(ctx)
	if err != nil {
		return nil, err
	}
	accessToken := authInfo.GetOAuthAccessToken()
	if accessToken == nil || accessToken.UserIDs.UserID != req.UserIDs.UserID || accessToken.ClientIDs.ClientID != req.ClientIDs.ClientID {
		if err := rights.RequireUser(ctx, req.UserIDs, ttnpb.RIGHT_USER_AUTHORIZED_CLIENTS); err != nil {
			return nil, err
		}
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		oauthStore := store.GetOAuthStore(db)
		if accessToken != nil && accessToken.ID != req.ID {
			accessToken, err := oauthStore.GetAccessToken(ctx, req.ID)
			if err != nil {
				return err
			}
			if accessToken.UserIDs.UserID != req.UserIDs.UserID || accessToken.ClientIDs.ClientID != req.ClientIDs.ClientID {
				return errAccessTokenMismatch
			}
		}
		return oauthStore.DeleteAccessToken(ctx, req.ID)
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

type oauthRegistry struct {
	*IdentityServer
}

func (or *oauthRegistry) List(ctx context.Context, req *ttnpb.ListOAuthClientAuthorizationsRequest) (*ttnpb.OAuthClientAuthorizations, error) {
	return or.listOAuthClientAuthorizations(ctx, req)
}

func (or *oauthRegistry) ListTokens(ctx context.Context, req *ttnpb.ListOAuthAccessTokensRequest) (*ttnpb.OAuthAccessTokens, error) {
	return or.listOAuthAccessTokens(ctx, req)
}

func (or *oauthRegistry) Delete(ctx context.Context, req *ttnpb.OAuthClientAuthorizationIdentifiers) (*types.Empty, error) {
	return or.deleteOAuthAuthorization(ctx, req)
}

func (or *oauthRegistry) DeleteToken(ctx context.Context, req *ttnpb.OAuthAccessTokenIdentifiers) (*types.Empty, error) {
	return or.deleteOAuthAccessToken(ctx, req)
}
