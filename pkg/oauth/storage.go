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

package oauth

import (
	"context"
	"strings"
	"time"

	"github.com/openshift/osin"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

const redirectURISeparator = ";"

// osinClient type is the same as ttnpb.Client, while implementing the osin.Client interface.
type osinClient ttnpb.Client

func (cli osinClient) GetId() string {
	return cli.ClientIdentifiers.ClientID
}

func (cli osinClient) GetSecret() string {
	panic("GetSecret should never be called by osin")
}

func (cli osinClient) ClientSecretMatches(secret string) bool {
	if cli.Secret == "" {
		return secret == ""
	}
	valid, _ := auth.Validate(cli.Secret, secret)
	return valid
}

func (cli osinClient) GetRedirectUri() string {
	return strings.Join(cli.RedirectURIs, redirectURISeparator)
}

func (cli osinClient) GetUserData() interface{} { return nil }

// userData is used as the UserData interface in osin structs.
type userData struct {
	ttnpb.UserIdentifiers
	ID string
}

// storage wraps IS stores, while implementing the osin.Storage interface.
type storage struct {
	ctx     context.Context
	clients store.ClientStore
	oauth   store.OAuthStore
}

func (s *storage) Clone() osin.Storage { return s }

func (s *storage) Close() {}

func (s *storage) GetClient(id string) (osin.Client, error) {
	cli, err := s.clients.GetClient(
		s.ctx,
		&ttnpb.ClientIdentifiers{ClientID: id},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return osinClient(*cli), nil
}

func (s *storage) SaveAuthorize(data *osin.AuthorizeData) error {
	userIDs := data.UserData.(userData).UserIdentifiers
	client := ttnpb.Client(data.Client.(osinClient))
	rights := rightsFromScope(data.Scope)
	_, err := s.oauth.Authorize(s.ctx, &ttnpb.OAuthClientAuthorization{
		ClientIDs: client.ClientIdentifiers,
		UserIDs:   userIDs,
		Rights:    rights,
	})
	if err != nil {
		return err
	}
	if data.CreatedAt.IsZero() {
		data.CreatedAt = time.Now()
	}
	err = s.oauth.CreateAuthorizationCode(s.ctx, &ttnpb.OAuthAuthorizationCode{
		ClientIDs:   client.ClientIdentifiers,
		UserIDs:     userIDs,
		Rights:      rights,
		Code:        data.Code,
		RedirectURI: data.RedirectUri,
		State:       data.State,
		CreatedAt:   data.CreatedAt,
		ExpiresAt:   data.CreatedAt.Add(time.Duration(data.ExpiresIn) * time.Second),
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) LoadAuthorize(code string) (data *osin.AuthorizeData, err error) {
	authorizationCode, err := s.oauth.GetAuthorizationCode(s.ctx, code)
	if err != nil {
		return nil, err
	}
	client, err := s.GetClient(authorizationCode.ClientIDs.ClientID)
	if err != nil {
		return nil, err
	}
	return &osin.AuthorizeData{
		Client:      client,
		Code:        code,
		ExpiresIn:   int32(authorizationCode.ExpiresAt.Sub(authorizationCode.CreatedAt).Seconds()),
		Scope:       rightsToScope(authorizationCode.Rights...),
		RedirectUri: authorizationCode.RedirectURI,
		State:       authorizationCode.State,
		CreatedAt:   authorizationCode.CreatedAt,
		UserData:    userData{UserIdentifiers: authorizationCode.UserIDs},
	}, nil
}

func (s *storage) RemoveAuthorize(code string) error {
	return s.oauth.DeleteAuthorizationCode(s.ctx, code)
}

var errTokenMismatch = errors.DefineInternal(
	"token_mismatch",
	"refresh token ID `{refresh_token_id}` does not match access token ID `{access_token_id}`",
)

var errNoAccessToken = errors.DefineInvalidArgument(
	"no_access_token",
	"the provided token is not an access token`",
)

var errNoRefreshToken = errors.DefineInvalidArgument(
	"no_refresh_token",
	"the provided token is not a refresh token`",
)

var errInvalidToken = errors.DefineInvalidArgument("token", "invalid token")

func (s *storage) SaveAccess(data *osin.AccessData) error {
	var accessHash, refreshHash string
	tokenType, accessID, accessKey, err := auth.SplitToken(data.AccessToken)
	if err != nil {
		return err
	}
	if tokenType != auth.AccessToken {
		return errNoAccessToken
	}
	accessHash, err = auth.Hash(s.ctx, accessKey)
	if err != nil {
		return err
	}
	if data.RefreshToken != "" {
		tokenType, refreshID, refreshKey, err := auth.SplitToken(data.RefreshToken)
		if err != nil {
			return err
		}
		if tokenType != auth.RefreshToken {
			return errNoRefreshToken
		}
		if refreshID != accessID {
			return errTokenMismatch.WithAttributes("refresh_token_id", refreshID, "access_token_id", accessID)
		}
		refreshHash, err = auth.Hash(s.ctx, refreshKey)
		if err != nil {
			return err
		}
	}
	userIDs := data.UserData.(userData).UserIdentifiers
	client := ttnpb.Client(data.Client.(osinClient))
	rights := rightsFromScope(data.Scope)
	if data.CreatedAt.IsZero() {
		data.CreatedAt = time.Now()
	}
	var previousID string
	if data.AccessData != nil {
		previousID = data.AccessData.UserData.(userData).ID
		if data.AccessData.AccessToken != "" {
			data.AccessData.AccessToken = previousID // Used for deleting the old access token
		}
		if data.AccessData.RefreshToken != "" {
			data.AccessData.RefreshToken = previousID // Used for deleting the old access token
		}
	}
	return s.oauth.CreateAccessToken(s.ctx, &ttnpb.OAuthAccessToken{
		ClientIDs:    client.ClientIdentifiers,
		UserIDs:      userIDs,
		Rights:       rights,
		ID:           accessID,
		AccessToken:  accessHash,
		RefreshToken: refreshHash,
		CreatedAt:    data.CreatedAt,
		ExpiresAt:    data.CreatedAt.Add(time.Duration(data.ExpiresIn) * time.Second),
	}, previousID)
}

func (s *storage) loadAccess(id string) (*osin.AccessData, error) {
	accessToken, err := s.oauth.GetAccessToken(s.ctx, id)
	if err != nil {
		return nil, err
	}
	client, err := s.GetClient(accessToken.ClientIDs.ClientID)
	if err != nil {
		return nil, err
	}
	return &osin.AccessData{
		Client:       client,
		AccessToken:  accessToken.AccessToken,
		RefreshToken: accessToken.RefreshToken,
		ExpiresIn:    int32(accessToken.ExpiresAt.Sub(accessToken.CreatedAt).Seconds()),
		Scope:        rightsToScope(accessToken.Rights...),
		CreatedAt:    accessToken.CreatedAt,
		UserData:     userData{UserIdentifiers: accessToken.UserIDs, ID: id},
	}, nil
}

func (s *storage) LoadAccess(token string) (*osin.AccessData, error) {
	panic("LoadAccess should never be called by osin")
}

func (s *storage) RemoveAccess(token string) error {
	if tokenType, id, _, err := auth.SplitToken(token); err == nil {
		if tokenType != auth.AccessToken {
			return errNoAccessToken
		}
		return s.oauth.DeleteAccessToken(s.ctx, id)
	}
	return s.oauth.DeleteAccessToken(s.ctx, token)
}

func (s *storage) LoadRefresh(token string) (*osin.AccessData, error) {
	tokenType, id, tokenKey, err := auth.SplitToken(token)
	if err != nil {
		return nil, err
	}
	if tokenType != auth.RefreshToken {
		return nil, errNoRefreshToken
	}
	data, err := s.loadAccess(id)
	if err != nil {
		return nil, err
	}
	valid, err := auth.Validate(data.RefreshToken, tokenKey)
	if !valid || err != nil {
		return nil, errInvalidToken
	}
	return data, nil
}

func (s *storage) RemoveRefresh(token string) error {
	if tokenType, id, _, err := auth.SplitToken(token); err == nil {
		if tokenType != auth.RefreshToken {
			return errNoRefreshToken
		}
		return s.oauth.DeleteAccessToken(s.ctx, id)
	}
	return s.oauth.DeleteAccessToken(s.ctx, token)
}
