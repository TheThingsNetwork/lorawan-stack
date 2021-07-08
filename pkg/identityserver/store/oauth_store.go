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
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// GetOAuthStore returns an OAuthStore on the given db (or transaction).
func GetOAuthStore(db *gorm.DB) OAuthStore {
	return &oauthStore{store: newStore(db)}
}

type oauthStore struct {
	*store
}

func (s *oauthStore) ListAuthorizations(ctx context.Context, userIDs *ttnpb.UserIdentifiers) ([]*ttnpb.OAuthClientAuthorization, error) {
	defer trace.StartRegion(ctx, "list authorizations").End()
	user, err := s.findEntity(ctx, userIDs, "id")
	if err != nil {
		return nil, err
	}
	var authModels []ClientAuthorization
	query := s.query(ctx, ClientAuthorization{}).Where(ClientAuthorization{
		UserID: user.PrimaryKey(),
	})
	query = query.Order(orderFromContext(ctx, "client_authorizations", "created_at", "DESC"))
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 {
		countTotal(ctx, query.Model(ClientAuthorization{}))
		query = query.Limit(limit).Offset(offset)
	}
	err = query.Preload("Client").Find(&authModels).Error
	if err != nil {
		return nil, err
	}
	setTotal(ctx, uint64(len(authModels)))
	authProtos := make([]*ttnpb.OAuthClientAuthorization, len(authModels))
	for i, authModel := range authModels {
		authProto := authModel.toPB()
		authProto.UserIds.UserId = userIDs.UserId
		authProtos[i] = authProto
	}
	return authProtos, nil
}

func (s *oauthStore) GetAuthorization(ctx context.Context, userIDs *ttnpb.UserIdentifiers, clientIDs *ttnpb.ClientIdentifiers) (*ttnpb.OAuthClientAuthorization, error) {
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
			return nil, errAuthorizationNotFound.WithAttributes("user_id", userIDs.UserId, "client_id", clientIDs.ClientId)
		}
		return nil, err
	}
	authProto := authModel.toPB()
	authProto.ClientIds.ClientId = clientIDs.ClientId
	authProto.UserIds.UserId = userIDs.UserId
	return authProto, nil
}

func (s *oauthStore) Authorize(ctx context.Context, authorization *ttnpb.OAuthClientAuthorization) (*ttnpb.OAuthClientAuthorization, error) {
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
	authModel.Rights = Rights{Rights: authorization.Rights}
	query := s.query(ctx, ClientAuthorization{}).Save(&authModel)
	if query.Error != nil {
		return nil, query.Error
	}
	authProto := authModel.toPB()
	authProto.ClientIds = authorization.ClientIds
	authProto.UserIds = authorization.UserIds
	return authProto, nil
}

func (s *oauthStore) DeleteAuthorization(ctx context.Context, userIDs *ttnpb.UserIdentifiers, clientIDs *ttnpb.ClientIdentifiers) error {
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
	user, err := s.findDeletedEntity(ctx, userIDs, "id")
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
	client, err := s.findDeletedEntity(ctx, clientIDs, "id")
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

func (s *oauthStore) CreateAuthorizationCode(ctx context.Context, code *ttnpb.OAuthAuthorizationCode) error {
	defer trace.StartRegion(ctx, "create authorization code").End()
	client, err := s.findEntity(ctx, code.ClientIds, "id")
	if err != nil {
		return err
	}
	user, err := s.findEntity(ctx, code.UserIds, "id")
	if err != nil {
		return err
	}
	codeModel := AuthorizationCode{
		ClientID:    client.PrimaryKey(),
		UserID:      user.PrimaryKey(),
		Rights:      Rights{Rights: code.Rights},
		Code:        code.Code,
		RedirectURI: code.RedirectURI,
		State:       code.State,
		ExpiresAt:   code.ExpiresAt,
	}
	if code.UserSessionID != "" {
		codeModel.UserSessionID = &code.UserSessionID
	}
	codeModel.CreatedAt = cleanTime(code.CreatedAt)
	return s.createEntity(ctx, &codeModel)
}

func (s *oauthStore) GetAuthorizationCode(ctx context.Context, code string) (*ttnpb.OAuthAuthorizationCode, error) {
	defer trace.StartRegion(ctx, "get authorization code").End()
	if code == "" {
		return nil, errAuthorizationCodeNotFound.WithAttributes("authorization_code", code)
	}
	var codeModel AuthorizationCode
	err := s.query(ctx, AuthorizationCode{}).Where(AuthorizationCode{
		Code: code,
	}).Preload("Client").Preload("User.Account").First(&codeModel).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errAuthorizationCodeNotFound.WithAttributes("authorization_code", code)
		}
	}
	return codeModel.toPB(), nil
}

func (s *oauthStore) DeleteAuthorizationCode(ctx context.Context, code string) error {
	if code == "" {
		return errAuthorizationCodeNotFound.WithAttributes("authorization_code", code)
	}
	defer trace.StartRegion(ctx, "delete authorization code").End()
	err := s.query(ctx, AuthorizationCode{}).Where(AuthorizationCode{
		Code: code,
	}).Delete(&AuthorizationCode{}).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errAuthorizationCodeNotFound.WithAttributes("authorization_code", code)
		}
		return err
	}
	return nil
}

func (s *oauthStore) CreateAccessToken(ctx context.Context, token *ttnpb.OAuthAccessToken, previousID string) error {
	defer trace.StartRegion(ctx, "create access token").End()
	client, err := s.findEntity(ctx, token.ClientIds, "id")
	if err != nil {
		return err
	}
	user, err := s.findEntity(ctx, token.UserIds, "id")
	if err != nil {
		return err
	}
	tokenModel := AccessToken{
		ClientID:     client.PrimaryKey(),
		UserID:       user.PrimaryKey(),
		Rights:       Rights{Rights: token.Rights},
		TokenID:      token.ID,
		PreviousID:   previousID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.ExpiresAt,
	}
	if token.UserSessionID != "" {
		tokenModel.UserSessionID = &token.UserSessionID
	}
	tokenModel.CreatedAt = cleanTime(token.CreatedAt)
	return s.createEntity(ctx, &tokenModel)
}

func (s *oauthStore) ListAccessTokens(ctx context.Context, userIDs *ttnpb.UserIdentifiers, clientIDs *ttnpb.ClientIdentifiers) ([]*ttnpb.OAuthAccessToken, error) {
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
	query = query.Order(orderFromContext(ctx, "access_tokens", "created_at", "DESC"))
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 {
		countTotal(ctx, query.Model(AccessToken{}))
		query = query.Limit(limit).Offset(offset)
	}
	err = query.Find(&tokenModels).Error
	if err != nil {
		return nil, err
	}
	setTotal(ctx, uint64(len(tokenModels)))
	tokenProtos := make([]*ttnpb.OAuthAccessToken, len(tokenModels))
	for i, tokenModel := range tokenModels {
		tokenProto := tokenModel.toPB()
		tokenProto.ClientIds.ClientId = clientIDs.ClientId
		tokenProto.UserIds.UserId = userIDs.UserId
		tokenProtos[i] = tokenProto
	}
	return tokenProtos, nil
}

func (s *oauthStore) GetAccessToken(ctx context.Context, id string) (*ttnpb.OAuthAccessToken, error) {
	if id == "" {
		return nil, errAccessTokenNotFound.WithAttributes("access_token_id", id)
	}
	defer trace.StartRegion(ctx, "get access token").End()
	var tokenModel struct {
		AccessToken
		FriendlyClientID string
		FriendlyUserID   string
	}
	err := s.query(ctx, AccessToken{}).
		Select(`"access_tokens".*, "accounts"."uid" AS "friendly_user_id", "clients"."client_id" AS "friendly_client_id"`).
		Joins(`JOIN "accounts" ON "accounts"."account_type" = 'user' AND "accounts"."account_id" = "access_tokens"."user_id"`).
		Joins(`JOIN "clients" ON "clients"."id" = "access_tokens"."client_id"`).
		Where(AccessToken{TokenID: id}).Scan(&tokenModel).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errAccessTokenNotFound.WithAttributes("access_token_id", id)
		}
	}
	tokenProto := tokenModel.AccessToken.toPB()
	tokenProto.ClientIds.ClientId = tokenModel.FriendlyClientID
	tokenProto.UserIds.UserId = tokenModel.FriendlyUserID
	return tokenProto, nil
}

func (s *oauthStore) DeleteAccessToken(ctx context.Context, id string) error {
	if id == "" {
		return errAccessTokenNotFound.WithAttributes("access_token_id", id)
	}
	defer trace.StartRegion(ctx, "delete access token").End()
	err := s.query(ctx, AccessToken{}).Where(AccessToken{
		TokenID: id,
	}).Delete(&AccessToken{}).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errAccessTokenNotFound.WithAttributes("access_token_id", id)
		}
		return err
	}
	return nil
}
