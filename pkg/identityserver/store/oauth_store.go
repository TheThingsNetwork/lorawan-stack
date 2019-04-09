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
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// GetOAuthStore returns an OAuthStore on the given db (or transaction).
func GetOAuthStore(db *gorm.DB) OAuthStore {
	return &oauthStore{db: db}
}

type oauthStore struct {
	db *gorm.DB
}

func (s *oauthStore) ListAuthorizations(ctx context.Context, userIDs *ttnpb.UserIdentifiers) ([]*ttnpb.OAuthClientAuthorization, error) {
	defer trace.StartRegion(ctx, "list authorizations").End()
	user, err := findEntity(ctx, s.db, userIDs.EntityIdentifiers(), "id")
	if err != nil {
		return nil, err
	}
	var authModels []ClientAuthorization
	err = s.db.Where(ClientAuthorization{
		UserID: user.PrimaryKey(),
	}).Preload("Client").Find(&authModels).Error
	if err != nil {
		return nil, err
	}
	authProtos := make([]*ttnpb.OAuthClientAuthorization, len(authModels))
	for i, authModel := range authModels {
		authProto := authModel.toPB()
		authProto.UserIDs.UserID = userIDs.UserID
		authProtos[i] = authProto
	}
	return authProtos, nil
}

func (s *oauthStore) GetAuthorization(ctx context.Context, userIDs *ttnpb.UserIdentifiers, clientIDs *ttnpb.ClientIdentifiers) (*ttnpb.OAuthClientAuthorization, error) {
	defer trace.StartRegion(ctx, "get authorization").End()
	client, err := findEntity(ctx, s.db, clientIDs.EntityIdentifiers(), "id")
	if err != nil {
		return nil, err
	}
	user, err := findEntity(ctx, s.db, userIDs.EntityIdentifiers(), "id")
	if err != nil {
		return nil, err
	}
	var authModel ClientAuthorization
	err = s.db.Where(ClientAuthorization{
		ClientID: client.PrimaryKey(),
		UserID:   user.PrimaryKey(),
	}).First(&authModel).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errAuthorizationNotFound.WithAttributes("user_id", userIDs.UserID, "client_id", clientIDs.ClientID)
		}
		return nil, err
	}
	authProto := authModel.toPB()
	authProto.ClientIDs.ClientID = clientIDs.ClientID
	authProto.UserIDs.UserID = userIDs.UserID
	return authProto, nil
}

func (s *oauthStore) Authorize(ctx context.Context, authorization *ttnpb.OAuthClientAuthorization) (*ttnpb.OAuthClientAuthorization, error) {
	defer trace.StartRegion(ctx, "create or update authorization").End()
	client, err := findEntity(ctx, s.db, authorization.ClientIDs.EntityIdentifiers(), "id")
	if err != nil {
		return nil, err
	}
	user, err := findEntity(ctx, s.db, authorization.UserIDs.EntityIdentifiers(), "id")
	if err != nil {
		return nil, err
	}
	var authModel ClientAuthorization
	err = s.db.Where(ClientAuthorization{
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
	query := s.db.Save(&authModel)
	if query.Error != nil {
		return nil, query.Error
	}
	authProto := authModel.toPB()
	authProto.ClientIDs = authorization.ClientIDs
	authProto.UserIDs = authorization.UserIDs
	return authProto, nil
}

func (s *oauthStore) DeleteAuthorization(ctx context.Context, userIDs *ttnpb.UserIdentifiers, clientIDs *ttnpb.ClientIdentifiers) error {
	defer trace.StartRegion(ctx, "delete authorization").End()
	client, err := findEntity(ctx, s.db, clientIDs.EntityIdentifiers(), "id")
	if err != nil {
		return err
	}
	user, err := findEntity(ctx, s.db, userIDs.EntityIdentifiers(), "id")
	if err != nil {
		return err
	}
	err = s.db.Where(ClientAuthorization{
		ClientID: client.PrimaryKey(),
		UserID:   user.PrimaryKey(),
	}).Delete(&ClientAuthorization{}).Error
	return err
}

func (s *oauthStore) CreateAuthorizationCode(ctx context.Context, code *ttnpb.OAuthAuthorizationCode) error {
	defer trace.StartRegion(ctx, "create authorization code").End()
	client, err := findEntity(ctx, s.db, code.ClientIDs.EntityIdentifiers(), "id")
	if err != nil {
		return err
	}
	user, err := findEntity(ctx, s.db, code.UserIDs.EntityIdentifiers(), "id")
	if err != nil {
		return err
	}
	codeModel := AuthorizationCode{
		ClientAuthorization: ClientAuthorization{
			ClientID: client.PrimaryKey(),
			UserID:   user.PrimaryKey(),
			Rights:   Rights{Rights: code.Rights},
		},
		Code:        code.Code,
		RedirectURI: code.RedirectURI,
		State:       code.State,
		ExpiresAt:   code.ExpiresAt,
	}
	codeModel.SetContext(ctx)
	codeModel.CreatedAt = cleanTime(code.CreatedAt)
	query := s.db.Save(&codeModel)
	if query.Error != nil {
		return query.Error
	}
	return nil
}

func (s *oauthStore) GetAuthorizationCode(ctx context.Context, code string) (*ttnpb.OAuthAuthorizationCode, error) {
	defer trace.StartRegion(ctx, "get authorization code").End()
	if code == "" {
		return nil, errAuthorizationCodeNotFound
	}
	var codeModel AuthorizationCode
	err := s.db.Scopes(withContext(ctx)).Where(AuthorizationCode{
		Code: code,
	}).Preload("Client").Preload("User.Account").First(&codeModel).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errAuthorizationCodeNotFound
		}
	}
	return codeModel.toPB(), nil
}

func (s *oauthStore) DeleteAuthorizationCode(ctx context.Context, code string) error {
	if code == "" {
		return errAuthorizationCodeNotFound
	}
	defer trace.StartRegion(ctx, "delete authorization code").End()
	err := s.db.Scopes(withContext(ctx)).Where(AuthorizationCode{
		Code: code,
	}).Delete(&AuthorizationCode{}).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errAuthorizationCodeNotFound
		}
		return err
	}
	return nil
}

func (s *oauthStore) CreateAccessToken(ctx context.Context, token *ttnpb.OAuthAccessToken, previousID string) error {
	defer trace.StartRegion(ctx, "create access token").End()
	client, err := findEntity(ctx, s.db, token.ClientIDs.EntityIdentifiers(), "id")
	if err != nil {
		return err
	}
	user, err := findEntity(ctx, s.db, token.UserIDs.EntityIdentifiers(), "id")
	if err != nil {
		return err
	}
	tokenModel := AccessToken{
		ClientAuthorization: ClientAuthorization{
			ClientID: client.PrimaryKey(),
			UserID:   user.PrimaryKey(),
			Rights:   Rights{Rights: token.Rights},
		},
		TokenID:      token.ID,
		PreviousID:   previousID,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.ExpiresAt,
	}
	tokenModel.SetContext(ctx)
	tokenModel.CreatedAt = cleanTime(token.CreatedAt)
	query := s.db.Save(&tokenModel)
	if query.Error != nil {
		return query.Error
	}
	return nil
}

func (s *oauthStore) ListAccessTokens(ctx context.Context, userIDs *ttnpb.UserIdentifiers, clientIDs *ttnpb.ClientIdentifiers) ([]*ttnpb.OAuthAccessToken, error) {
	defer trace.StartRegion(ctx, "list access tokens").End()
	client, err := findEntity(ctx, s.db, clientIDs.EntityIdentifiers(), "id")
	if err != nil {
		return nil, err
	}
	user, err := findEntity(ctx, s.db, userIDs.EntityIdentifiers(), "id")
	if err != nil {
		return nil, err
	}
	var tokenModels []AccessToken
	err = s.db.Scopes(withContext(ctx)).Where(AccessToken{
		ClientAuthorization: ClientAuthorization{
			ClientID: client.PrimaryKey(),
			UserID:   user.PrimaryKey(),
		},
	}).Find(&tokenModels).Error
	if err != nil {
		return nil, err
	}
	tokenProtos := make([]*ttnpb.OAuthAccessToken, len(tokenModels))
	for i, tokenModel := range tokenModels {
		tokenProto := tokenModel.toPB()
		tokenProto.ClientIDs.ClientID = clientIDs.ClientID
		tokenProto.UserIDs.UserID = userIDs.UserID
		tokenProtos[i] = tokenProto
	}
	return tokenProtos, nil
}

func (s *oauthStore) GetAccessToken(ctx context.Context, id string) (*ttnpb.OAuthAccessToken, error) {
	if id == "" {
		return nil, errAccessTokenNotFound
	}
	defer trace.StartRegion(ctx, "get access token").End()
	var tokenModel AccessToken
	err := s.db.Scopes(withContext(ctx)).Where(AccessToken{
		TokenID: id,
	}).Preload("Client").Preload("User.Account").First(&tokenModel).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errAccessTokenNotFound
		}
	}
	return tokenModel.toPB(), nil
}

func (s *oauthStore) DeleteAccessToken(ctx context.Context, id string) error {
	if id == "" {
		return errAccessTokenNotFound
	}
	defer trace.StartRegion(ctx, "delete access token").End()
	err := s.db.Scopes(withContext(ctx)).Where(AccessToken{
		TokenID: id,
	}).Delete(&AccessToken{}).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errAccessTokenNotFound
		}
		return err
	}
	return nil
}
