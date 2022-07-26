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

package store

import (
	"context"
	"runtime/trace"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// GetOAuthStore returns an OAuthStore on the given db (or transaction).
func GetOAuthStore(db *gorm.DB) store.OAuthStore {
	return &oauthStore{baseStore: newStore(db)}
}

type oauthStore struct {
	*baseStore
}

func (s *oauthStore) ListAuthorizations(
	ctx context.Context, userIDs *ttnpb.UserIdentifiers,
) ([]*ttnpb.OAuthClientAuthorization, error) {
	defer trace.StartRegion(ctx, "list authorizations").End()
	user, err := s.findEntity(ctx, userIDs, "id")
	if err != nil {
		return nil, err
	}
	var authModels []ClientAuthorization
	query := s.query(ctx, ClientAuthorization{}).Where(ClientAuthorization{
		UserID: user.PrimaryKey(),
	})
	query = query.Order(store.OrderFromContext(ctx, "client_authorizations", createdAt, "DESC"))
	if limit, offset := store.LimitAndOffsetFromContext(ctx); limit != 0 {
		var total uint64
		query.Count(&total)
		store.SetTotal(ctx, total)
		query = query.Limit(limit).Offset(offset)
	}
	err = query.Preload("Client").Find(&authModels).Error
	if err != nil {
		return nil, err
	}
	store.SetTotal(ctx, uint64(len(authModels)))
	authProtos := make([]*ttnpb.OAuthClientAuthorization, len(authModels))
	for i, authModel := range authModels {
		authProto := authModel.toPB()
		authProto.UserIds = userIDs
		authProtos[i] = authProto
	}
	return authProtos, nil
}

func (s *oauthStore) GetAuthorization(
	ctx context.Context, userIDs *ttnpb.UserIdentifiers, clientIDs *ttnpb.ClientIdentifiers,
) (*ttnpb.OAuthClientAuthorization, error) {
	defer trace.StartRegion(ctx, "get authorization").End()
	client, err := s.findEntity(ctx, clientIDs, "id")
	if err != nil {
		return nil, err
	}
	user, err := s.findEntity(ctx, userIDs, "id")
	if err != nil {
		return nil, err
	}
	var authModel ClientAuthorization
	err = s.query(ctx, ClientAuthorization{}).Where(ClientAuthorization{
		ClientID: client.PrimaryKey(),
		UserID:   user.PrimaryKey(),
	}).First(&authModel).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, store.ErrAuthorizationNotFound.WithAttributes(
				"user_id", userIDs.GetUserId(),
				"client_id", clientIDs.ClientId,
			)
		}
		return nil, err
	}
	authProto := authModel.toPB()
	authProto.ClientIds = clientIDs
	authProto.UserIds = userIDs
	return authProto, nil
}

func (s *oauthStore) Authorize(
	ctx context.Context, authorization *ttnpb.OAuthClientAuthorization,
) (*ttnpb.OAuthClientAuthorization, error) {
	defer trace.StartRegion(ctx, "create or update authorization").End()
	client, err := s.findEntity(ctx, authorization.ClientIds, "id")
	if err != nil {
		return nil, err
	}
	user, err := s.findEntity(ctx, authorization.UserIds, "id")
	if err != nil {
		return nil, err
	}
	var authModel ClientAuthorization
	err = s.query(ctx, ClientAuthorization{}).Where(ClientAuthorization{
		ClientID: client.PrimaryKey(),
		UserID:   user.PrimaryKey(),
	}).First(&authModel).Error
	if err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			return nil, err
		}
		authModel = ClientAuthorization{
			ClientID: client.PrimaryKey(),
			UserID:   user.PrimaryKey(),
		}
		authModel.SetContext(ctx)
	}
	authModel.Rights = authorization.Rights
	query := s.query(ctx, ClientAuthorization{}).Save(&authModel)
	if query.Error != nil {
		return nil, query.Error
	}
	authProto := authModel.toPB()
	authProto.ClientIds = authorization.ClientIds
	authProto.UserIds = authorization.UserIds
	return authProto, nil
}

func (s *oauthStore) DeleteAuthorization(
	ctx context.Context, userIDs *ttnpb.UserIdentifiers, clientIDs *ttnpb.ClientIdentifiers,
) error {
	defer trace.StartRegion(ctx, "delete authorization").End()
	client, err := s.findEntity(ctx, clientIDs, "id")
	if err != nil {
		return err
	}
	user, err := s.findEntity(ctx, userIDs, "id")
	if err != nil {
		return err
	}
	err = s.query(ctx, ClientAuthorization{}).Where(ClientAuthorization{
		ClientID: client.PrimaryKey(),
		UserID:   user.PrimaryKey(),
	}).Delete(&ClientAuthorization{}).Error
	return err
}

func (s *oauthStore) DeleteUserAuthorizations(ctx context.Context, userIDs *ttnpb.UserIdentifiers) error {
	defer trace.StartRegion(ctx, "delete user authorizations").End()
	user, err := s.findEntity(store.WithSoftDeleted(ctx, false), userIDs, "id")
	if err != nil {
		return err
	}
	err = s.query(ctx, ClientAuthorization{}).Where(ClientAuthorization{
		UserID: user.PrimaryKey(),
	}).Delete(&ClientAuthorization{}).Error
	if err != nil {
		return err
	}
	err = s.query(ctx, AuthorizationCode{}).Where(AuthorizationCode{
		UserID: user.PrimaryKey(),
	}).Delete(&AuthorizationCode{}).Error
	if err != nil {
		return err
	}
	return s.query(ctx, AccessToken{}).Where(AccessToken{
		UserID: user.PrimaryKey(),
	}).Delete(&AccessToken{}).Error
}

func (s *oauthStore) DeleteClientAuthorizations(ctx context.Context, clientIDs *ttnpb.ClientIdentifiers) error {
	defer trace.StartRegion(ctx, "delete client authorizations").End()
	client, err := s.findEntity(store.WithSoftDeleted(ctx, false), clientIDs, "id")
	if err != nil {
		return err
	}
	err = s.query(ctx, ClientAuthorization{}).Where(ClientAuthorization{
		ClientID: client.PrimaryKey(),
	}).Delete(&ClientAuthorization{}).Error
	if err != nil {
		return err
	}
	err = s.query(ctx, AuthorizationCode{}).Where(AuthorizationCode{
		ClientID: client.PrimaryKey(),
	}).Delete(&AuthorizationCode{}).Error
	if err != nil {
		return err
	}
	return s.query(ctx, AccessToken{}).Where(AccessToken{
		ClientID: client.PrimaryKey(),
	}).Delete(&AccessToken{}).Error
}

func (s *oauthStore) CreateAuthorizationCode(
	ctx context.Context, code *ttnpb.OAuthAuthorizationCode,
) (*ttnpb.OAuthAuthorizationCode, error) {
	defer trace.StartRegion(ctx, "create authorization code").End()
	client, err := s.findEntity(ctx, code.ClientIds, "id")
	if err != nil {
		return nil, err
	}
	user, err := s.findEntity(ctx, code.UserIds, "id")
	if err != nil {
		return nil, err
	}
	codeModel := AuthorizationCode{
		ClientID:    client.PrimaryKey(),
		UserID:      user.PrimaryKey(),
		Rights:      code.Rights,
		Code:        code.Code,
		RedirectURI: code.RedirectUri,
		State:       code.State,
		ExpiresAt:   cleanTimePtr(ttnpb.StdTime(code.ExpiresAt)),
	}
	if createdAt := ttnpb.StdTime(code.CreatedAt); createdAt != nil {
		codeModel.CreatedAt = cleanTime(*createdAt)
	}
	if code.UserSessionId != "" {
		codeModel.UserSessionID = &code.UserSessionId
	}
	if err = s.createEntity(ctx, &codeModel); err != nil {
		return nil, err
	}
	codeProto := codeModel.toPB()
	codeProto.ClientIds = code.ClientIds
	codeProto.UserIds = code.UserIds
	return codeProto, nil
}

func (s *oauthStore) GetAuthorizationCode(ctx context.Context, code string) (*ttnpb.OAuthAuthorizationCode, error) {
	defer trace.StartRegion(ctx, "get authorization code").End()
	if code == "" {
		return nil, store.ErrAuthorizationCodeNotFound.New()
	}
	var codeModel AuthorizationCode
	err := s.query(ctx, AuthorizationCode{}).Where(AuthorizationCode{
		Code: code,
	}).Preload("Client").Preload("User.Account").First(&codeModel).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, store.ErrAuthorizationCodeNotFound.New()
		}
	}
	return codeModel.toPB(), nil
}

func (s *oauthStore) DeleteAuthorizationCode(ctx context.Context, code string) error {
	if code == "" {
		return store.ErrAuthorizationCodeNotFound.New()
	}
	defer trace.StartRegion(ctx, "delete authorization code").End()
	err := s.query(ctx, AuthorizationCode{}).Where(AuthorizationCode{
		Code: code,
	}).Delete(&AuthorizationCode{}).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return store.ErrAuthorizationCodeNotFound.New()
		}
		return err
	}
	return nil
}

func (s *oauthStore) CreateAccessToken(
	ctx context.Context, token *ttnpb.OAuthAccessToken, previousID string,
) (*ttnpb.OAuthAccessToken, error) {
	defer trace.StartRegion(ctx, "create access token").End()
	client, err := s.findEntity(ctx, token.ClientIds, "id")
	if err != nil {
		return nil, err
	}
	user, err := s.findEntity(ctx, token.UserIds, "id")
	if err != nil {
		return nil, err
	}
	tokenModel := AccessToken{
		ClientID:     client.PrimaryKey(),
		UserID:       user.PrimaryKey(),
		Rights:       token.Rights,
		TokenID:      token.Id,
		PreviousID:   previousID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    cleanTimePtr(ttnpb.StdTime(token.ExpiresAt)),
	}
	if createdAt := ttnpb.StdTime(token.CreatedAt); createdAt != nil {
		tokenModel.CreatedAt = cleanTime(*createdAt)
	}
	if token.UserSessionId != "" {
		tokenModel.UserSessionID = &token.UserSessionId
	}
	if err = s.createEntity(ctx, &tokenModel); err != nil {
		return nil, err
	}
	tokenProto := tokenModel.toPB()
	tokenProto.ClientIds = token.ClientIds
	tokenProto.UserIds = token.UserIds
	return tokenProto, nil
}

func (s *oauthStore) ListAccessTokens(
	ctx context.Context, userIDs *ttnpb.UserIdentifiers, clientIDs *ttnpb.ClientIdentifiers,
) ([]*ttnpb.OAuthAccessToken, error) {
	defer trace.StartRegion(ctx, "list access tokens").End()
	client, err := s.findEntity(ctx, clientIDs, "id")
	if err != nil {
		return nil, err
	}
	user, err := s.findEntity(ctx, userIDs, "id")
	if err != nil {
		return nil, err
	}
	var tokenModels []AccessToken
	query := s.query(ctx, AccessToken{}).Where(AccessToken{
		ClientID: client.PrimaryKey(),
		UserID:   user.PrimaryKey(),
	})
	query = query.Order(store.OrderFromContext(ctx, "access_tokens", createdAt, "DESC"))
	if limit, offset := store.LimitAndOffsetFromContext(ctx); limit != 0 {
		var total uint64
		query.Count(&total)
		store.SetTotal(ctx, total)
		query = query.Limit(limit).Offset(offset)
	}
	err = query.Find(&tokenModels).Error
	if err != nil {
		return nil, err
	}
	store.SetTotal(ctx, uint64(len(tokenModels)))
	tokenProtos := make([]*ttnpb.OAuthAccessToken, len(tokenModels))
	for i, tokenModel := range tokenModels {
		tokenProto := tokenModel.toPB()
		tokenProto.ClientIds = clientIDs
		tokenProto.UserIds = userIDs
		tokenProtos[i] = tokenProto
	}
	return tokenProtos, nil
}

func (s *oauthStore) GetAccessToken(ctx context.Context, id string) (*ttnpb.OAuthAccessToken, error) {
	if id == "" {
		return nil, store.ErrAccessTokenNotFound.WithAttributes("access_token_id", id)
	}
	defer trace.StartRegion(ctx, "get access token").End()
	var tokenModel struct {
		AccessToken
		FriendlyClientID string
		FriendlyUserID   string
	}
	err := s.query(ctx, AccessToken{}).
		Select(`"access_tokens".*, "accounts"."uid" AS "friendly_user_id", "clients"."client_id" AS "friendly_client_id"`).
		Joins(
			`JOIN "accounts" ON "accounts"."account_type" = 'user' AND "accounts"."account_id" = "access_tokens"."user_id"`,
		).
		Joins(`JOIN "clients" ON "clients"."id" = "access_tokens"."client_id"`).
		Where(AccessToken{TokenID: id}).Scan(&tokenModel).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, store.ErrAccessTokenNotFound.WithAttributes("access_token_id", id)
		}
	}
	tokenProto := tokenModel.AccessToken.toPB()
	tokenProto.ClientIds = &ttnpb.ClientIdentifiers{ClientId: tokenModel.FriendlyClientID}
	tokenProto.UserIds = &ttnpb.UserIdentifiers{UserId: tokenModel.FriendlyUserID}
	return tokenProto, nil
}

func (s *oauthStore) DeleteAccessToken(ctx context.Context, id string) error {
	if id == "" {
		return store.ErrAccessTokenNotFound.WithAttributes("access_token_id", id)
	}
	defer trace.StartRegion(ctx, "delete access token").End()
	err := s.query(ctx, AccessToken{}).Where(AccessToken{
		TokenID: id,
	}).Delete(&AccessToken{}).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return store.ErrAccessTokenNotFound.WithAttributes("access_token_id", id)
		}
		return err
	}
	return nil
}
