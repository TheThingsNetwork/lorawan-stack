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

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var errUnsupportedAuthorization = errors.DefineUnauthenticated("unsupported_authorization", "Unsupported authorization method")

var errTokenExpired = errors.DefineUnauthenticated("token_expired", "access token expired")

var errInvalidAuthorization = errors.DefinePermissionDenied("invalid_authorization", "invalid authorization")

func (is *IdentityServer) fetchAPIKey(ctx context.Context, db *gorm.DB, id string) (string, *ttnpb.EntityIdentifiers, *ttnpb.Rights, error) {
	ids, apiKey, err := store.GetAPIKeyStore(db).GetAPIKey(ctx, id)
	if err != nil {
		return "", nil, nil, err
	}
	return apiKey.Key, ids, ttnpb.RightsFrom(apiKey.Rights...), nil
}

func (is *IdentityServer) fetchAccessToken(ctx context.Context, db *gorm.DB, id string) (string, *ttnpb.EntityIdentifiers, *ttnpb.Rights, error) {
	accessToken, err := store.GetOAuthStore(db).GetAccessToken(ctx, id)
	if err != nil {
		return "", nil, nil, err
	}
	if accessToken.ExpiresAt.Before(time.Now()) {
		return "", nil, nil, errTokenExpired
	}
	return accessToken.AccessToken, accessToken.UserIDs.EntityIdentifiers(), ttnpb.RightsFrom(accessToken.Rights...), nil
}

func (is *IdentityServer) fetchMemberRights(ctx context.Context, db *gorm.DB, ids *ttnpb.EntityIdentifiers) (map[*ttnpb.EntityIdentifiers]*ttnpb.Rights, error) {
	var ouIDs *ttnpb.OrganizationOrUserIdentifiers
	switch ids := ids.Identifiers().(type) {
	case *ttnpb.OrganizationIdentifiers:
		ouIDs = ids.OrganizationOrUserIdentifiers()
	case *ttnpb.UserIdentifiers:
		ouIDs = ids.OrganizationOrUserIdentifiers()
	default:
		return nil, nil
	}
	memberRights, err := store.GetMembershipStore(db).FindAllMemberRights(ctx, ouIDs, "")
	if err != nil {
		return nil, err
	}
	return memberRights, nil
}

var (
	impossibleAPIKeyRights      = ttnpb.RightsFrom()
	impossibleAccessTokenRights = ttnpb.RightsFrom(
		ttnpb.RIGHT_GATEWAY_LINK,
		ttnpb.RIGHT_APPLICATION_LINK,
	)
)

func (is *IdentityServer) getRights(ctx context.Context) (map[*ttnpb.EntityIdentifiers]*ttnpb.Rights, error) {
	md := rpcmetadata.FromIncomingContext(ctx)
	if md.AuthType == "" {
		return nil, nil
	}
	if strings.ToLower(md.AuthType) != "bearer" {
		return nil, errUnsupportedAuthorization
	}
	token := md.AuthValue
	tokenType, tokenID, tokenKey, err := auth.SplitToken(token)
	if err != nil {
		return nil, err
	}

	var fetch func(ctx context.Context, db *gorm.DB, id string) (string, *ttnpb.EntityIdentifiers, *ttnpb.Rights, error)
	var impossibleRights *ttnpb.Rights

	switch tokenType {
	case auth.AccessToken:
		fetch = is.fetchAccessToken
		impossibleRights = impossibleAccessTokenRights
	case auth.APIKey:
		fetch = is.fetchAPIKey
		impossibleRights = impossibleAPIKeyRights
	default:
		return nil, errUnsupportedAuthorization
	}

	identifierRights := make(map[*ttnpb.EntityIdentifiers]*ttnpb.Rights)

	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		secret, ids, rights, err := fetch(ctx, db, tokenID)
		if err != nil {
			return err
		}
		valid, err := auth.Password(secret).Validate(tokenKey)
		if err != nil {
			return err
		}
		if !valid {
			return errInvalidAuthorization
		}
		rights = rights.Implied().Sub(impossibleRights)
		identifierRights[ids] = rights
		memberRights, err := is.fetchMemberRights(ctx, db, ids)
		if err != nil {
			return err
		}
		for ids, memberRights := range memberRights {
			identifierRights[ids] = memberRights.Implied().Intersect(rights)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return identifierRights, nil
}

func (is *IdentityServer) ApplicationRights(ctx context.Context, appIDs ttnpb.ApplicationIdentifiers) (*ttnpb.Rights, error) {
	rights, err := is.getRights(ctx)
	if err != nil {
		return nil, err
	}
	for ids, rights := range rights {
		if ids := ids.GetApplicationIDs(); ids != nil && ids.ApplicationID == appIDs.ApplicationID {
			return rights, nil
		}
	}
	return &ttnpb.Rights{}, nil
}

func (is *IdentityServer) ClientRights(ctx context.Context, cliIDs ttnpb.ClientIdentifiers) (*ttnpb.Rights, error) {
	rights, err := is.getRights(ctx)
	if err != nil {
		return nil, err
	}
	for ids, rights := range rights {
		if ids := ids.GetClientIDs(); ids != nil && ids.ClientID == cliIDs.ClientID {
			return rights, nil
		}
	}
	return &ttnpb.Rights{}, nil
}

func (is *IdentityServer) GatewayRights(ctx context.Context, gtwIDs ttnpb.GatewayIdentifiers) (*ttnpb.Rights, error) {
	rights, err := is.getRights(ctx)
	if err != nil {
		return nil, err
	}
	for ids, rights := range rights {
		if ids := ids.GetGatewayIDs(); ids != nil && ids.GatewayID == gtwIDs.GatewayID {
			return rights, nil
		}
	}
	return &ttnpb.Rights{}, nil
}

func (is *IdentityServer) OrganizationRights(ctx context.Context, orgIDs ttnpb.OrganizationIdentifiers) (*ttnpb.Rights, error) {
	rights, err := is.getRights(ctx)
	if err != nil {
		return nil, err
	}
	for ids, rights := range rights {
		if ids := ids.GetOrganizationIDs(); ids != nil && ids.OrganizationID == orgIDs.OrganizationID {
			return rights, nil
		}
	}
	return &ttnpb.Rights{}, nil
}

func (is *IdentityServer) UserRights(ctx context.Context, userIDs ttnpb.UserIdentifiers) (*ttnpb.Rights, error) {
	rights, err := is.getRights(ctx)
	if err != nil {
		return nil, err
	}
	for ids, rights := range rights {
		if ids := ids.GetUserIDs(); ids != nil && ids.UserID == userIDs.UserID {
			return rights, nil
		}
	}
	return &ttnpb.Rights{}, nil
}
