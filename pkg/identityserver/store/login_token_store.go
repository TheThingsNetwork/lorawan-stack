// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
	"time"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// GetLoginTokenStore returns an LoginTokenStore on the given db (or transaction).
func GetLoginTokenStore(db *gorm.DB) LoginTokenStore {
	return &loginTokenStore{store: newStore(db)}
}

type loginTokenStore struct {
	*store
}

func (s *loginTokenStore) FindActiveLoginTokens(ctx context.Context, userIDs *ttnpb.UserIdentifiers) ([]*ttnpb.LoginToken, error) {
	defer trace.StartRegion(ctx, "find active login tokens").End()
	user, err := s.findEntity(ctx, userIDs, "id")
	if err != nil {
		return nil, err
	}
	var loginTokenModels []LoginToken
	if err := s.query(ctx, LoginToken{}).Where(
		"user_id = ?", user.PrimaryKey(),
	).Where(
		"expires_at > now()",
	).Where(
		"used = false",
	).Find(&loginTokenModels).Error; err != nil {
		return nil, err
	}
	if len(loginTokenModels) == 0 {
		return nil, nil
	}
	loginTokenProtos := make([]*ttnpb.LoginToken, len(loginTokenModels))
	for i, loginTokenModel := range loginTokenModels {
		loginTokenProtos[i] = loginTokenModel.toPB()
	}
	return loginTokenProtos, nil
}

func (s *loginTokenStore) CreateLoginToken(ctx context.Context, loginToken *ttnpb.LoginToken) (*ttnpb.LoginToken, error) {
	defer trace.StartRegion(ctx, "create login token").End()
	user, err := s.findEntity(ctx, loginToken.UserIdentifiers, "id")
	if err != nil {
		return nil, err
	}
	model := LoginToken{
		UserID:    user.PrimaryKey(),
		Token:     loginToken.Token,
		ExpiresAt: loginToken.ExpiresAt,
	}
	if err := s.createEntity(ctx, &model); err != nil {
		return nil, convertError(err)
	}
	return model.toPB(), nil
}

var (
	errLoginTokenNotFound    = errors.DefineNotFound("login_token_not_found", "login token not found")
	errLoginTokenAlreadyUsed = errors.DefineAlreadyExists("login_token_already_used", "login token already used")
	errLoginTokenExpired     = errors.DefineInvalidArgument("login_token_expired", "login token expired")
)

func (s *loginTokenStore) ConsumeLoginToken(ctx context.Context, token string) (*ttnpb.LoginToken, error) {
	defer trace.StartRegion(ctx, "consume login token").End()
	var loginTokenModel LoginToken
	if err := s.query(ctx, LoginToken{}).Where(LoginToken{Token: token}).Preload("User.Account").First(&loginTokenModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errLoginTokenNotFound.New()
		}
		return nil, err
	}
	if loginTokenModel.Used {
		return nil, errLoginTokenAlreadyUsed
	}
	if loginTokenModel.ExpiresAt.Before(time.Now()) {
		return nil, errLoginTokenExpired
	}
	loginTokenModel.Used = true
	if err := s.updateEntity(ctx, &loginTokenModel, "used"); err != nil {
		return nil, err
	}
	return loginTokenModel.toPB(), nil
}
