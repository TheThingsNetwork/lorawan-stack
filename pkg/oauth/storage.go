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
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	oauth_store "go.thethings.network/lorawan-stack/v3/pkg/oauth/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const redirectURISeparator = ";"

// osinClient type is just a pointer to ttnpb.Client, while implementing the osin.Client interface.
type osinClient struct {
	*ttnpb.Client
}

func (cli osinClient) GetId() string {
	return cli.Ids.GetClientId()
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
	return strings.Join(cli.RedirectUris, redirectURISeparator)
}

func (cli osinClient) GetUserData() interface{} { return nil }

// userData is used as the UserData interface in osin structs.
type userData struct {
	*ttnpb.UserSessionIdentifiers
	ID string
}

// storage wraps IS stores, while implementing the osin.Storage interface.
type storage struct {
	ctx   context.Context
	store oauth_store.TransactionalInterface
}

func (s *storage) Clone() osin.Storage { return s }

func (s *storage) Close() {}

func (s *storage) GetClient(id string) (osin.Client, error) {
	var cli *ttnpb.Client
	err := s.store.Transact(s.ctx, func(ctx context.Context, st oauth_store.Interface) (err error) {
		cli, err = st.GetClient(ctx, &ttnpb.ClientIdentifiers{ClientId: id}, nil)
		return err
	})
	if err != nil {
		return nil, err
	}
	return osinClient{cli}, nil
}

func (s *storage) SaveAuthorize(data *osin.AuthorizeData) error {
	userSessionIDs := data.UserData.(userData).UserSessionIdentifiers
	client := data.Client.(osinClient).Client
	rights := rightsFromScope(data.Scope)
	err := s.store.Transact(s.ctx, func(ctx context.Context, st oauth_store.Interface) (err error) {
		_, err = st.Authorize(ctx, &ttnpb.OAuthClientAuthorization{
			ClientIds: client.GetIds(),
			UserIds:   userSessionIDs.GetUserIds(),
			Rights:    rights,
		})
		if err != nil {
			return err
		}
		if data.CreatedAt.IsZero() {
			data.CreatedAt = time.Now()
		}
		_, err = st.CreateAuthorizationCode(ctx, &ttnpb.OAuthAuthorizationCode{
			ClientIds:     client.GetIds(),
			UserIds:       userSessionIDs.GetUserIds(),
			UserSessionId: userSessionIDs.SessionId,
			Rights:        rights,
			Code:          data.Code,
			RedirectUri:   data.RedirectUri,
			State:         data.State,
			CreatedAt:     timestamppb.New(data.CreatedAt),
			ExpiresAt:     timestamppb.New(data.CreatedAt.Add(time.Duration(data.ExpiresIn) * time.Second)),
		})
		return err
	})
	return err
}

func (s *storage) LoadAuthorize(code string) (data *osin.AuthorizeData, err error) {
	var (
		authorizationCode *ttnpb.OAuthAuthorizationCode
		client            *ttnpb.Client
	)
	err = s.store.Transact(s.ctx, func(ctx context.Context, st oauth_store.Interface) (err error) {
		authorizationCode, err = st.GetAuthorizationCode(ctx, code)
		if err != nil {
			return err
		}
		client, err = st.GetClient(ctx, authorizationCode.ClientIds, nil)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	var expiresIn int32
	if expiresAt := ttnpb.StdTime(authorizationCode.ExpiresAt); expiresAt != nil {
		expiresIn = int32(expiresAt.Sub(*ttnpb.StdTime(authorizationCode.CreatedAt)).Seconds())
	}
	return &osin.AuthorizeData{
		Client:      osinClient{client},
		Code:        code,
		ExpiresIn:   expiresIn,
		Scope:       rightsToScope(authorizationCode.Rights...),
		RedirectUri: authorizationCode.RedirectUri,
		State:       authorizationCode.State,
		CreatedAt:   *ttnpb.StdTime(authorizationCode.CreatedAt),
		UserData: userData{
			UserSessionIdentifiers: &ttnpb.UserSessionIdentifiers{
				UserIds:   authorizationCode.UserIds,
				SessionId: authorizationCode.UserSessionId,
			},
		},
	}, nil
}

func (s *storage) RemoveAuthorize(code string) error {
	return s.store.Transact(s.ctx, func(ctx context.Context, st oauth_store.Interface) error {
		return st.DeleteAuthorizationCode(ctx, code)
	})
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
		return errNoAccessToken.New()
	}
	accessHash, err = auth.Hash(auth.NewContextWithHashValidator(s.ctx, tokenHashSettings), accessKey)
	if err != nil {
		return err
	}
	if data.RefreshToken != "" {
		tokenType, refreshID, refreshKey, err := auth.SplitToken(data.RefreshToken)
		if err != nil {
			return err
		}
		if tokenType != auth.RefreshToken {
			return errNoRefreshToken.New()
		}
		if refreshID != accessID {
			return errTokenMismatch.WithAttributes("refresh_token_id", refreshID, "access_token_id", accessID)
		}
		refreshHash, err = auth.Hash(auth.NewContextWithHashValidator(s.ctx, tokenHashSettings), refreshKey)
		if err != nil {
			return err
		}
	}
	userSessionIDs := data.UserData.(userData).UserSessionIdentifiers
	client := data.Client.(osinClient).Client
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
	err = s.store.Transact(s.ctx, func(ctx context.Context, st oauth_store.Interface) error {
		_, err := st.CreateAccessToken(ctx, &ttnpb.OAuthAccessToken{
			ClientIds:     client.GetIds(),
			UserIds:       userSessionIDs.GetUserIds(),
			UserSessionId: userSessionIDs.SessionId,
			Rights:        rights,
			Id:            accessID,
			AccessToken:   accessHash,
			RefreshToken:  refreshHash,
			CreatedAt:     timestamppb.New(data.CreatedAt),
			ExpiresAt:     timestamppb.New(data.CreatedAt.Add(time.Duration(data.ExpiresIn) * time.Second)),
		}, previousID)
		return err
	})
	return err
}

func (s *storage) loadAccess(id string) (*osin.AccessData, error) {
	var (
		accessToken *ttnpb.OAuthAccessToken
		client      *ttnpb.Client
	)
	err := s.store.Transact(s.ctx, func(ctx context.Context, st oauth_store.Interface) (err error) {
		accessToken, err = st.GetAccessToken(ctx, id)
		if err != nil {
			return err
		}
		client, err = st.GetClient(ctx, accessToken.ClientIds, nil)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	var expiresIn int32
	if expiresAt := ttnpb.StdTime(accessToken.ExpiresAt); expiresAt != nil {
		expiresIn = int32(expiresAt.Sub(*ttnpb.StdTime(accessToken.CreatedAt)).Seconds())
	}
	return &osin.AccessData{
		Client:       osinClient{client},
		AccessToken:  accessToken.AccessToken,
		RefreshToken: accessToken.RefreshToken,
		ExpiresIn:    expiresIn,
		Scope:        rightsToScope(accessToken.Rights...),
		CreatedAt:    *ttnpb.StdTime(accessToken.CreatedAt),
		UserData: userData{
			UserSessionIdentifiers: &ttnpb.UserSessionIdentifiers{
				UserIds:   accessToken.UserIds,
				SessionId: accessToken.UserSessionId,
			},
			ID: id,
		},
	}, nil
}

func (s *storage) LoadAccess(token string) (*osin.AccessData, error) {
	panic("LoadAccess should never be called by osin")
}

func (s *storage) RemoveAccess(token string) error {
	if tokenType, id, _, err := auth.SplitToken(token); err == nil {
		if tokenType != auth.AccessToken {
			return errNoAccessToken.New()
		}
		token = id
	}
	return s.store.Transact(s.ctx, func(ctx context.Context, st oauth_store.Interface) error {
		return st.DeleteAccessToken(ctx, token)
	})
}

func (s *storage) LoadRefresh(token string) (*osin.AccessData, error) {
	tokenType, id, tokenKey, err := auth.SplitToken(token)
	if err != nil {
		return nil, err
	}
	if tokenType != auth.RefreshToken {
		return nil, errNoRefreshToken.New()
	}
	data, err := s.loadAccess(id)
	if err != nil {
		return nil, err
	}
	valid, err := auth.Validate(data.RefreshToken, tokenKey)
	if !valid || err != nil {
		return nil, errInvalidToken.New()
	}
	return data, nil
}

func (s *storage) RemoveRefresh(token string) error {
	if tokenType, id, _, err := auth.SplitToken(token); err == nil {
		if tokenType != auth.RefreshToken {
			return errNoRefreshToken.New()
		}
		token = id
	}
	return s.store.Transact(s.ctx, func(ctx context.Context, st oauth_store.Interface) error {
		return st.DeleteAccessToken(ctx, token)
	})
}
