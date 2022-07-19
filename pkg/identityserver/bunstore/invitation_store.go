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
)

// Invitation is the invitation model in the database.
type Invitation struct {
	bun.BaseModel `bun:"table:invitations,alias:inv"`

	Model

	Email string `bun:"email,notnull"`
	Token string `bun:"token,notnull"`

	ExpiresAt *time.Time `bun:"expires_at"`

	AcceptedBy   *User      `bun:"rel:belongs-to,join:accepted_by_id=id"`
	AcceptedByID *string    `bun:"accepted_by_id"`
	AcceptedAt   *time.Time `bun:"accepted_at"`
}

// BeforeAppendModel is a hook that modifies the model on SELECT and UPDATE queries.
func (m *Invitation) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if err := m.Model.BeforeAppendModel(ctx, query); err != nil {
		return err
	}
	return nil
}

func invitationToPB(m *Invitation) (*ttnpb.Invitation, error) {
	pb := &ttnpb.Invitation{
		Email:      m.Email,
		Token:      m.Token,
		ExpiresAt:  ttnpb.ProtoTime(m.ExpiresAt),
		CreatedAt:  ttnpb.ProtoTimePtr(m.CreatedAt),
		UpdatedAt:  ttnpb.ProtoTimePtr(m.UpdatedAt),
		AcceptedAt: ttnpb.ProtoTime(m.AcceptedAt),
	}
	if m.AcceptedBy != nil {
		pb.AcceptedBy = &ttnpb.UserIdentifiers{
			UserId: m.AcceptedBy.Account.UID,
		}
	}
	return pb, nil
}

type invitationStore struct {
	*entityStore
}

func newInvitationStore(baseStore *baseStore) *invitationStore {
	return &invitationStore{
		entityStore: newEntityStore(baseStore),
	}
}

func (s *invitationStore) CreateInvitation(
	ctx context.Context, pb *ttnpb.Invitation,
) (*ttnpb.Invitation, error) {
	ctx, span := tracer.Start(ctx, "CreateInvitation")
	defer span.End()

	model := &Invitation{
		Email:     pb.Email,
		Token:     pb.Token,
		ExpiresAt: ttnpb.StdTime(pb.ExpiresAt),
	}

	_, err := s.DB.NewInsert().
		Model(model).
		Exec(ctx)
	if err != nil {
		err = wrapDriverError(err)
		if errors.IsAlreadyExists(err) {
			return nil, store.ErrInvitationAlreadySent.New()
		}
		return nil, err
	}

	pb, err = invitationToPB(model)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func (s *invitationStore) FindInvitations(ctx context.Context) ([]*ttnpb.Invitation, error) {
	ctx, span := tracer.Start(ctx, "FindInvitations")
	defer span.End()

	models := []*Invitation{}
	selectQuery := s.DB.NewSelect().
		Model(&models).
		Apply(selectWithContext(ctx))

	// Count the total number of results.
	count, err := selectQuery.Count(ctx)
	if err != nil {
		return nil, wrapDriverError(err)
	}
	store.SetTotal(ctx, uint64(count))

	// Apply ordering and paging.
	selectQuery = selectQuery.
		Apply(selectWithOrderFromContext(ctx, "id", map[string]string{
			"invitation_id": "id",
			"email":         "email",
			"created_at":    "created_at",
			"expires_at":    "expires_at",
		})).
		Apply(selectWithLimitAndOffsetFromContext(ctx))

	// Include the user that accepted the invitation.
	selectQuery = selectQuery.
		Relation("AcceptedBy", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Column("account_uid")
		})

	// Scan the results.
	err = selectQuery.Scan(ctx)
	if err != nil {
		return nil, wrapDriverError(err)
	}

	// Convert the results to protobuf.
	pbs := make([]*ttnpb.Invitation, len(models))
	for i, model := range models {
		pb, err := invitationToPB(model)
		if err != nil {
			return nil, err
		}
		pbs[i] = pb
	}

	return pbs, nil
}

func (s *invitationStore) getInvitationModelBy(
	ctx context.Context,
	by func(*bun.SelectQuery) *bun.SelectQuery,
) (*Invitation, error) {
	model := &Invitation{}
	selectQuery := s.DB.NewSelect().
		Model(model).
		Apply(by)

	if err := selectQuery.Scan(ctx); err != nil {
		return nil, wrapDriverError(err)
	}

	return model, nil
}

func (s *invitationStore) GetInvitation(ctx context.Context, token string) (*ttnpb.Invitation, error) {
	ctx, span := tracer.Start(ctx, "GetInvitation", trace.WithAttributes(
		attribute.String("invitation_token", token),
	))
	defer span.End()

	model, err := s.getInvitationModelBy(ctx, func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.
			Where("token = ?", token).
			Relation("AcceptedBy", func(q *bun.SelectQuery) *bun.SelectQuery {
				return q.Column("account_uid")
			})
	})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, store.ErrInvitationNotFound.WithAttributes(
				"invitation_token", token,
			)
		}
		return nil, err
	}

	pb, err := invitationToPB(model)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func (s *invitationStore) SetInvitationAcceptedBy(
	ctx context.Context, token string, accepter *ttnpb.UserIdentifiers,
) error {
	ctx, span := tracer.Start(ctx, "SetInvitationAcceptedBy", trace.WithAttributes(
		attribute.String("invitation_token", token),
		attribute.String("user_id", accepter.GetUserId()),
	))
	defer span.End()

	model, err := s.getInvitationModelBy(ctx, func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("token = ?", token)
	})
	if err != nil {
		if errors.IsNotFound(err) {
			return store.ErrInvitationNotFound.WithAttributes(
				"invitation_token", token,
			)
		}
		return err
	}

	if model.ExpiresAt != nil && model.ExpiresAt.Before(time.Now()) {
		return store.ErrInvitationExpired.WithAttributes("invitation_token", token)
	}

	if model.AcceptedByID != nil {
		return store.ErrInvitationAlreadyUsed.WithAttributes("invitation_token", token)
	}

	_, userUUID, err := s.getEntity(ctx, accepter)
	if err != nil {
		return err
	}

	_, err = s.DB.NewUpdate().
		Model(model).
		WherePK().
		Set("accepted_by_id = ?, accepted_at = NOW()", userUUID).
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	return nil
}

func (s *invitationStore) DeleteInvitation(ctx context.Context, email string) error {
	ctx, span := tracer.Start(ctx, "DeleteInvitation")
	defer span.End()

	model, err := s.getInvitationModelBy(ctx, func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("LOWER(email) = lower(?)", email)
	})
	if err != nil {
		if errors.IsNotFound(err) {
			return store.ErrInvitationNotFound.New()
		}
		return err
	}

	_, err = s.DB.NewDelete().
		Model(model).
		WherePK().
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	return nil
}
