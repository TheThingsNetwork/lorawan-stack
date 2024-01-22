// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/tracing/tracer"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	storeutil "go.thethings.network/lorawan-stack/v3/pkg/util/store"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// EmailValidation is the contact info validation model in the database.
type EmailValidation struct {
	bun.BaseModel `bun:"table:email_validations,alias:ev"`
	Model

	ExpiresAt *time.Time `bun:"expires_at"`

	Reference string `bun:"reference,nullzero,notnull"`
	Token     string `bun:"token,nullzero,notnull"`
	Used      bool   `bun:"used"`

	UserUUID     string `bun:"user_uuid,nullzero,notnull"`
	EmailAddress string `bun:"email_address,nullzero,notnull"`
}

// BeforeAppendModel is a hook that modifies the model on SELECT and UPDATE queries.
func (m *EmailValidation) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if err := m.Model.BeforeAppendModel(ctx, query); err != nil {
		return err
	}
	return nil
}

func (m *EmailValidation) toPB() *ttnpb.EmailValidation {
	val := &ttnpb.EmailValidation{
		Id:        m.Reference,
		Token:     m.Token,
		CreatedAt: ttnpb.ProtoTime(&m.CreatedAt),
		ExpiresAt: ttnpb.ProtoTime(m.ExpiresAt),
		UpdatedAt: ttnpb.ProtoTime(&m.UpdatedAt),
		Address:   m.EmailAddress,
	}
	return val
}

type emailValidationStore struct{ *entityStore }

func newEmailValidationStore(baseStore *baseStore) *emailValidationStore {
	return &emailValidationStore{entityStore: newEntityStore(baseStore)}
}

func (s *emailValidationStore) getEmailValidationModelBy(
	ctx context.Context,
	by func(*bun.SelectQuery) *bun.SelectQuery,
) (*EmailValidation, error) {
	model := &EmailValidation{}
	selectQuery := newSelectModel(ctx, s.DB, model).Apply(by)
	err := selectQuery.Scan(ctx)
	if err != nil {
		return nil, storeutil.WrapDriverError(err)
	}
	return model, nil
}

func (s *emailValidationStore) CreateEmailValidation(
	ctx context.Context, pb *ttnpb.EmailValidation,
) (*ttnpb.EmailValidation, error) {
	ctx, span := tracer.StartFromContext(ctx, "CreateEmailValidation")
	defer span.End()

	usrModel, err := s.getUserModelBy(ctx, s.userStore.selectWithPrimaryEmailAddress(ctx, pb.Address), nil)
	if err != nil {
		return nil, err
	}

	n, err := s.newSelectModel(ctx, &EmailValidation{}).
		Where("?TableAlias.user_uuid = ?", usrModel.ID).
		Where("LOWER(?TableAlias.email_address) = LOWER(?)", pb.Address).
		Where("?TableAlias.used = false").
		Where("?TableAlias.expires_at IS NULL OR ?TableAlias.expires_at > NOW()").
		Count(ctx)
	if err != nil {
		return nil, storeutil.WrapDriverError(err)
	}
	if n > 0 {
		return nil, store.ErrValidationAlreadySent.New()
	}

	model := &EmailValidation{
		Reference:    pb.Id,
		Token:        pb.Token,
		UserUUID:     usrModel.ID,
		EmailAddress: pb.Address,
		ExpiresAt:    cleanTimePtr(ttnpb.StdTime(pb.ExpiresAt)),
	}

	_, err = s.DB.NewInsert().Model(model).Exec(ctx)
	if err != nil {
		return nil, storeutil.WrapDriverError(err)
	}

	return &ttnpb.EmailValidation{
		Id:        model.Reference,
		Token:     model.Token,
		Address:   pb.Address,
		CreatedAt: timestamppb.New(model.CreatedAt),
		ExpiresAt: ttnpb.ProtoTime(model.ExpiresAt),
	}, nil
}

func (s *emailValidationStore) GetEmailValidation(
	ctx context.Context, pb *ttnpb.EmailValidation,
) (*ttnpb.EmailValidation, error) {
	ctx, span := tracer.StartFromContext(ctx, "GetEmailValidation")
	defer span.End()

	model, err := s.getEmailValidationModelBy(ctx, func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("?TableAlias.reference = ? AND ?TableAlias.token = ?", pb.Id, pb.Token)
	})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, store.ErrValidationTokenNotFound.WithAttributes("validation_id", pb.Id)
		}
		return nil, err
	}

	if model.Used {
		return nil, store.ErrValidationTokenAlreadyUsed.WithAttributes("validation_id", pb.Id)
	}

	if model.ExpiresAt != nil && model.ExpiresAt.Before(s.now()) {
		return nil, store.ErrValidationTokenExpired.WithAttributes("validation_id", pb.Id)
	}

	return model.toPB(), nil
}

func (s *emailValidationStore) ExpireEmailValidation(ctx context.Context, pb *ttnpb.EmailValidation) error {
	ctx, span := tracer.StartFromContext(ctx, "ExpireEmailValidation")
	defer span.End()

	model, err := s.getEmailValidationModelBy(ctx, func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("?TableAlias.reference = ? AND ?TableAlias.token = ?", pb.Id, pb.Token)
	})
	if err != nil {
		return err
	}

	usrModel, err := s.getUserModelBy(ctx, s.userStore.selectWithPrimaryEmailAddress(ctx, model.EmailAddress), nil)
	if err != nil {
		return err
	}

	if time.Now().After(*model.ExpiresAt) {
		return store.ErrValidationTokenExpired
	}
	if model.Used {
		return store.ErrValidationTokenAlreadyUsed
	}

	validatedAt := now()
	err = s.transact(ctx, func(ctx context.Context, tx bun.IDB) error {
		_, err = tx.NewUpdate().
			Model(usrModel).
			WherePK().
			Set("primary_email_address_validated_at = ?", validatedAt).
			Exec(ctx)
		if err != nil {
			return err
		}

		_, err = tx.NewUpdate().
			Model(model).
			WherePK().
			Set("expires_at = ?, used = true", validatedAt).
			Exec(ctx)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return storeutil.WrapDriverError(err)
	}

	return nil
}

// GetRefreshableEmailValidation returns a not used validation for a given user.
func (s *emailValidationStore) GetRefreshableEmailValidation(
	ctx context.Context, id *ttnpb.UserIdentifiers, refreshInterval time.Duration,
) (*ttnpb.EmailValidation, error) {
	ctx, span := tracer.StartFromContext(ctx, "GetRefreshableEmailValidation", trace.WithAttributes(
		attribute.String("refresh_interval", refreshInterval.String()),
	))
	defer span.End()

	usrModel, err := s.getUserModelBy(
		ctx, s.userStore.selectWithID(ctx, id.GetUserId()), []string{"primary_email_address"},
	)
	if err != nil {
		return nil, err
	}

	model, err := s.getEmailValidationModelBy(ctx, func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.
			Where("?TableAlias.user_uuid = ?", usrModel.ID).
			Where("LOWER(?TableAlias.email_address) = LOWER(?)", usrModel.PrimaryEmailAddress).
			Where("?TableAlias.used = false").
			Where("?TableAlias.expires_at IS NULL OR ?TableAlias.expires_at > NOW()").
			Where("?TableAlias.updated_at <= (NOW() - interval ?)", refreshInterval.String())
	})
	if err != nil {
		return nil, err
	}

	return model.toPB(), nil
}

// RefreshEmailValidation refreshes a email validation for an user.
func (s *emailValidationStore) RefreshEmailValidation(ctx context.Context, pb *ttnpb.EmailValidation) error {
	ctx, span := tracer.StartFromContext(ctx, "RefreshEmailValidation")
	defer span.End()

	model, err := s.getEmailValidationModelBy(ctx, func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.
			Where("reference = ? AND token = ?", pb.Id, pb.Token).
			Where("?TableAlias.used = false").
			// Done in order to avoid concurrent updates from happening in the same validation.
			Where("updated_at <= ?", pb.UpdatedAt.AsTime())
	})
	if err != nil {
		if errors.IsNotFound(err) {
			return store.ErrValidationTokenNotFound.WithAttributes("validation_id", pb.Id)
		}
		return err
	}

	_, err = s.DB.NewUpdate().Model(model).WherePK().Exec(ctx)
	if err != nil {
		return storeutil.WrapDriverError(err)
	}

	return nil
}
