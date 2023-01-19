// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
	"time"

	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	storeutil "go.thethings.network/lorawan-stack/v3/pkg/util/store"
)

// LoginToken is the login token model in the database.
type LoginToken struct {
	bun.BaseModel `bun:"table:login_tokens,alias:lt"`

	Model

	User   *User  `bun:"rel:belongs-to,join:user_id=id"`
	UserID string `bun:"user_id"`

	Token string `bun:"token,notnull"`

	ExpiresAt *time.Time `bun:"expires_at"`
	Used      bool       `bun:"used,nullzero"`
}

// BeforeAppendModel is a hook that modifies the model on SELECT and UPDATE queries.
func (m *LoginToken) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if err := m.Model.BeforeAppendModel(ctx, query); err != nil {
		return err
	}
	return nil
}

func loginTokenToPB(m *LoginToken, userIDs *ttnpb.UserIdentifiers) (*ttnpb.LoginToken, error) {
	pb := &ttnpb.LoginToken{
		UserIds:   userIDs,
		Token:     m.Token,
		ExpiresAt: ttnpb.ProtoTime(m.ExpiresAt),
		Used:      m.Used,
		CreatedAt: ttnpb.ProtoTimePtr(m.CreatedAt),
		UpdatedAt: ttnpb.ProtoTimePtr(m.UpdatedAt),
	}
	if userIDs == nil && m.User != nil {
		pb.UserIds = &ttnpb.UserIdentifiers{
			UserId: m.User.Account.UID,
		}
	}
	return pb, nil
}

type loginTokenStore struct {
	*entityStore
}

func newLoginTokenStore(baseStore *baseStore) *loginTokenStore {
	return &loginTokenStore{
		entityStore: newEntityStore(baseStore),
	}
}

func (s *loginTokenStore) FindActiveLoginTokens(
	ctx context.Context, userIDs *ttnpb.UserIdentifiers,
) ([]*ttnpb.LoginToken, error) {
	ctx, span := tracer.Start(ctx, "FindActiveLoginTokens", trace.WithAttributes(
		attribute.String("user_id", userIDs.GetUserId()),
	))
	defer span.End()

	_, userUUID, err := s.getEntity(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	models := []*LoginToken{}
	err = newSelectModels(ctx, s.DB, &models).
		Where("user_id = ?", userUUID).
		Where("expires_at > NOW()").
		Where("used = FALSE OR used IS NULL"). // TODO: Make "used" column NOT NULL (https://github.com/TheThingsNetwork/lorawan-stack/issues/5613).
		Scan(ctx)
	if err != nil {
		return nil, storeutil.WrapDriverError(err)
	}

	pbs := make([]*ttnpb.LoginToken, len(models))
	for i, model := range models {
		pb, err := loginTokenToPB(model, userIDs)
		if err != nil {
			return nil, err
		}
		pbs[i] = pb
	}

	return pbs, nil
}

func (s *loginTokenStore) CreateLoginToken(ctx context.Context, pb *ttnpb.LoginToken) (*ttnpb.LoginToken, error) {
	ctx, span := tracer.Start(ctx, "CreateLoginToken", trace.WithAttributes(
		attribute.String("user_id", pb.GetUserIds().GetUserId()),
	))
	defer span.End()

	_, userUUID, err := s.getEntity(ctx, pb.GetUserIds())
	if err != nil {
		return nil, err
	}

	model := &LoginToken{
		UserID:    userUUID,
		Token:     pb.Token,
		ExpiresAt: cleanTimePtr(ttnpb.StdTime(pb.ExpiresAt)),
	}

	_, err = s.DB.NewInsert().
		Model(model).
		Exec(ctx)
	if err != nil {
		return nil, storeutil.WrapDriverError(err)
	}

	pb, err = loginTokenToPB(model, pb.GetUserIds())
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func (s *loginTokenStore) ConsumeLoginToken(ctx context.Context, token string) (*ttnpb.LoginToken, error) {
	ctx, span := tracer.Start(ctx, "ConsumeLoginToken")
	defer span.End()

	model := &LoginToken{}
	err := s.newSelectModel(ctx, model).
		Where("token = ?", token).
		Relation("User", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Column("account_uid")
		}).
		Scan(ctx)
	if err != nil {
		err = storeutil.WrapDriverError(err)
		if errors.IsNotFound(err) {
			return nil, store.ErrLoginTokenNotFound.New()
		}
		return nil, err
	}

	if model.ExpiresAt != nil && model.ExpiresAt.Before(s.now()) {
		return nil, store.ErrLoginTokenExpired.New()
	}

	if model.Used {
		return nil, store.ErrLoginTokenAlreadyUsed.New()
	}

	_, err = s.DB.NewUpdate().
		Model(model).
		WherePK().
		Set("used = true").
		Exec(ctx)
	if err != nil {
		return nil, storeutil.WrapDriverError(err)
	}

	pb, err := loginTokenToPB(model, nil)
	if err != nil {
		return nil, err
	}

	return pb, nil
}
