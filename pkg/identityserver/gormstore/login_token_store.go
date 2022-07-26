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
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// GetLoginTokenStore returns an LoginTokenStore on the given db (or transaction).
func GetLoginTokenStore(db *gorm.DB) store.LoginTokenStore {
	return &loginTokenStore{baseStore: newStore(db)}
}

type loginTokenStore struct {
	*baseStore
}

func (s *loginTokenStore) FindActiveLoginTokens(
	ctx context.Context, userIDs *ttnpb.UserIdentifiers,
) ([]*ttnpb.LoginToken, error) {
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
	loginTokenProtos := make([]*ttnpb.LoginToken, len(loginTokenModels))
	for i, loginTokenModel := range loginTokenModels {
		loginTokenProtos[i] = loginTokenModel.toPB()
		loginTokenProtos[i].UserIds = userIDs
	}
	return loginTokenProtos, nil
}

func (s *loginTokenStore) CreateLoginToken(
	ctx context.Context, loginToken *ttnpb.LoginToken,
) (*ttnpb.LoginToken, error) {
	defer trace.StartRegion(ctx, "create login token").End()
	user, err := s.findEntity(ctx, loginToken.GetUserIds(), "id")
	if err != nil {
		return nil, err
	}
	model := LoginToken{
		UserID:    user.PrimaryKey(),
		Token:     loginToken.Token,
		ExpiresAt: ttnpb.StdTime(loginToken.ExpiresAt),
	}
	if err := s.createEntity(ctx, &model); err != nil {
		return nil, convertError(err)
	}
	pb := model.toPB()
	pb.UserIds = loginToken.UserIds
	return pb, nil
}

func (s *loginTokenStore) ConsumeLoginToken(ctx context.Context, token string) (*ttnpb.LoginToken, error) {
	defer trace.StartRegion(ctx, "consume login token").End()
	var loginTokenModel LoginToken
	if err := s.query(ctx, LoginToken{}).
		Where(LoginToken{Token: token}).
		Preload("User.Account").
		First(&loginTokenModel).
		Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, store.ErrLoginTokenNotFound.New()
		}
		return nil, err
	}
	if loginTokenModel.Used {
		return nil, store.ErrLoginTokenAlreadyUsed.New()
	}
	if loginTokenModel.ExpiresAt.Before(time.Now()) {
		return nil, store.ErrLoginTokenExpired.New()
	}
	loginTokenModel.Used = true
	if err := s.updateEntity(ctx, &loginTokenModel, "used"); err != nil {
		return nil, err
	}
	return loginTokenModel.toPB(), nil
}
