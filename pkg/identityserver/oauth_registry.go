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

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (is *IdentityServer) listOAuthClientAuthorizations(ctx context.Context, req *ttnpb.ListOAuthClientAuthorizationsRequest) (authorizations *ttnpb.OAuthClientAuthorizations, err error) {
	if err := rights.RequireUser(ctx, req.UserIds, ttnpb.Right_RIGHT_USER_AUTHORIZED_CLIENTS); err != nil {
		return nil, err
	}
	ctx = store.WithOrder(ctx, req.Order)
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	authorizations = &ttnpb.OAuthClientAuthorizations{}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		authorizations.Authorizations, err = st.ListAuthorizations(ctx, req.UserIds)
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
	accessToken := authInfo.GetOauthAccessToken()
	if accessToken == nil || accessToken.UserIds.GetUserId() != req.UserIds.GetUserId() || accessToken.ClientIds.ClientId != req.ClientIds.ClientId {
		if err := rights.RequireUser(ctx, req.UserIds, ttnpb.Right_RIGHT_USER_AUTHORIZED_CLIENTS); err != nil {
			return nil, err
		}
	}
	ctx = store.WithOrder(ctx, req.Order)
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	tokens = &ttnpb.OAuthAccessTokens{}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		tokens.Tokens, err = st.ListAccessTokens(ctx, req.UserIds, req.ClientIds)
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

func (is *IdentityServer) deleteOAuthAuthorization(ctx context.Context, req *ttnpb.OAuthClientAuthorizationIdentifiers) (*emptypb.Empty, error) {
	if err := rights.RequireUser(ctx, req.UserIds, ttnpb.Right_RIGHT_USER_AUTHORIZED_CLIENTS); err != nil {
		return nil, err
	}
	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		return st.DeleteAuthorization(ctx, req.UserIds, req.ClientIds)
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

var errAccessTokenMismatch = errors.DefineInvalidArgument("access_token_mismatch", "access token ID did not match user or client identifiers")

func (is *IdentityServer) deleteOAuthAccessToken(ctx context.Context, req *ttnpb.OAuthAccessTokenIdentifiers) (*emptypb.Empty, error) {
	authInfo, err := is.authInfo(ctx)
	if err != nil {
		return nil, err
	}
	accessToken := authInfo.GetOauthAccessToken()
	if accessToken == nil || accessToken.UserIds.GetUserId() != req.UserIds.GetUserId() || accessToken.ClientIds.ClientId != req.ClientIds.ClientId {
		if err := rights.RequireUser(ctx, req.UserIds, ttnpb.Right_RIGHT_USER_AUTHORIZED_CLIENTS); err != nil {
			return nil, err
		}
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		if accessToken != nil && accessToken.Id != req.Id {
			accessToken, err := st.GetAccessToken(ctx, req.Id)
			if err != nil {
				return err
			}
			if accessToken.UserIds.GetUserId() != req.UserIds.GetUserId() || accessToken.ClientIds.ClientId != req.ClientIds.ClientId {
				return errAccessTokenMismatch.New()
			}
		}
		return st.DeleteAccessToken(ctx, req.Id)
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

type oauthRegistry struct {
	ttnpb.UnimplementedOAuthAuthorizationRegistryServer

	*IdentityServer
}

func (or *oauthRegistry) List(ctx context.Context, req *ttnpb.ListOAuthClientAuthorizationsRequest) (*ttnpb.OAuthClientAuthorizations, error) {
	return or.listOAuthClientAuthorizations(ctx, req)
}

func (or *oauthRegistry) ListTokens(ctx context.Context, req *ttnpb.ListOAuthAccessTokensRequest) (*ttnpb.OAuthAccessTokens, error) {
	return or.listOAuthAccessTokens(ctx, req)
}

func (or *oauthRegistry) Delete(ctx context.Context, req *ttnpb.OAuthClientAuthorizationIdentifiers) (*emptypb.Empty, error) {
	return or.deleteOAuthAuthorization(ctx, req)
}

func (or *oauthRegistry) DeleteToken(ctx context.Context, req *ttnpb.OAuthAccessTokenIdentifiers) (*emptypb.Empty, error) {
	return or.deleteOAuthAccessToken(ctx, req)
}
