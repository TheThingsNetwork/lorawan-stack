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

package identityserver

import (
	"context"
	"strings"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	errUnauthenticated          = errors.DefineUnauthenticated("unauthenticated", "unauthenticated")
	errUnsupportedAuthorization = errors.DefineUnauthenticated("unsupported_authorization", "Unsupported authorization method")
	errInvalidAuthorization     = errors.DefinePermissionDenied("invalid_authorization", "invalid authorization")
	errTokenExpired             = errors.DefineUnauthenticated("token_expired", "access token expired")
)

func (is *IdentityServer) authInfo(ctx context.Context) (*ttnpb.AuthInfoResponse, error) {
	md := rpcmetadata.FromIncomingContext(ctx)
	if md.AuthType == "" {
		return &ttnpb.AuthInfoResponse{}, nil
	}
	if strings.ToLower(md.AuthType) != "bearer" {
		return nil, errUnsupportedAuthorization
	}
	token := md.AuthValue
	tokenType, tokenID, tokenKey, err := auth.SplitToken(token)
	if err != nil {
		return nil, err
	}

	var fetch func(db *gorm.DB) error
	res := new(ttnpb.AuthInfoResponse)
	userFieldMask := &types.FieldMask{Paths: []string{"admin"}}
	var user *ttnpb.User
	var userRights *ttnpb.Rights

	switch tokenType {
	case auth.APIKey:
		fetch = func(db *gorm.DB) error {
			ids, apiKey, err := store.GetAPIKeyStore(db).GetAPIKey(ctx, tokenID)
			if err != nil {
				return err
			}
			valid, err := auth.Password(apiKey.GetKey()).Validate(tokenKey)
			if err != nil {
				return err
			}
			if !valid {
				return errInvalidAuthorization
			}
			apiKey.Key = ""
			res.AccessMethod = &ttnpb.AuthInfoResponse_APIKey{
				APIKey: &ttnpb.AuthInfoResponse_APIKeyAccess{
					APIKey:    *apiKey,
					EntityIDs: *ids,
				},
			}
			if userIDs := ids.GetUserIDs(); userIDs != nil {
				user, err = store.GetUserStore(db).GetUser(ctx, userIDs, userFieldMask)
				if err != nil {
					return err
				}
				userRights = ttnpb.RightsFrom(apiKey.Rights...)
			}
			return nil
		}
	case auth.AccessToken:
		fetch = func(db *gorm.DB) error {
			accessToken, err := store.GetOAuthStore(db).GetAccessToken(ctx, tokenID)
			if err != nil {
				return err
			}
			valid, err := auth.Password(accessToken.GetAccessToken()).Validate(tokenKey)
			if err != nil {
				return err
			}
			if !valid {
				return errInvalidAuthorization
			}
			if accessToken.ExpiresAt.Before(time.Now()) {
				return errTokenExpired
			}
			accessToken.AccessToken, accessToken.RefreshToken = "", ""
			res.AccessMethod = &ttnpb.AuthInfoResponse_OAuthAccessToken{
				OAuthAccessToken: accessToken,
			}
			user, err = store.GetUserStore(db).GetUser(ctx, &accessToken.UserIDs, userFieldMask)
			if err != nil {
				return err
			}
			userRights = ttnpb.RightsFrom(accessToken.Rights...)
			return nil
		}
	default:
		return nil, errUnsupportedAuthorization
	}

	err = is.withDatabase(ctx, fetch)
	if err != nil {
		return nil, err
	}

	if user != nil {
		if user.Admin {
			res.UniversalRights = ttnpb.AllRights.Implied().Intersect(userRights.Implied()) // TODO: Use restricted Admin rights.
		}
	}

	return res, nil
}

func entityRights(authInfo *ttnpb.AuthInfoResponse) (*ttnpb.EntityIdentifiers, *ttnpb.Rights) {
	if apiKey := authInfo.GetAPIKey(); apiKey != nil {
		return &apiKey.EntityIDs, ttnpb.RightsFrom(apiKey.Rights...)
	} else if accessToken := authInfo.GetOAuthAccessToken(); accessToken != nil {
		return accessToken.UserIDs.EntityIdentifiers(), ttnpb.RightsFrom(accessToken.Rights...)
	}
	return nil, nil
}

func (is *IdentityServer) entityRights(ctx context.Context, authInfo *ttnpb.AuthInfoResponse) (map[*ttnpb.EntityIdentifiers]*ttnpb.Rights, error) {
	ids, rights := entityRights(authInfo)
	if ids == nil {
		return nil, nil
	}
	entityRights := make(map[*ttnpb.EntityIdentifiers]*ttnpb.Rights)
	entityRights[ids] = rights.Implied()
	memberRights, err := is.memberRights(ctx, ids)
	if err != nil {
		return nil, err
	}
	for ids, memberRights := range memberRights {
		entityRights[ids] = memberRights.Implied().Intersect(rights.Implied())
	}
	return entityRights, nil
}

func (is *IdentityServer) memberRights(ctx context.Context, ids *ttnpb.EntityIdentifiers) (entityRights map[*ttnpb.EntityIdentifiers]*ttnpb.Rights, err error) {
	var ouIDs *ttnpb.OrganizationOrUserIdentifiers
	switch ids := ids.Identifiers().(type) {
	case *ttnpb.OrganizationIdentifiers:
		ouIDs = ids.OrganizationOrUserIdentifiers()
	case *ttnpb.UserIdentifiers:
		ouIDs = ids.OrganizationOrUserIdentifiers()
	}
	if ouIDs == nil {
		return nil, nil
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		entityRights, err = store.GetMembershipStore(db).FindMemberRights(ctx, ouIDs, "")
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	subMemberRights := make(map[*ttnpb.EntityIdentifiers]*ttnpb.Rights)
	for ids, rights := range entityRights {
		memberRights, err := is.memberRights(ctx, ids)
		if err != nil {
			return nil, err
		}
		for ids, memberRights := range memberRights {
			subMemberRights[ids] = memberRights.Implied().Intersect(rights.Implied())
		}
	}
	for ids, rights := range subMemberRights {
		entityRights[ids] = rights
	}
	return entityRights, nil
}

type entityAccess struct {
	*IdentityServer
}

func (ea *entityAccess) AuthInfo(ctx context.Context, _ *types.Empty) (*ttnpb.AuthInfoResponse, error) {
	return ea.authInfo(ctx)
}
