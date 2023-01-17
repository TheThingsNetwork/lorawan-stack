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

// UserSession is the user session model in the database.
type UserSession struct {
	bun.BaseModel `bun:"table:user_sessions,alias:sess"`

	Model

	User   *User  `bun:"rel:belongs-to,join:user_id=id"`
	UserID string `bun:"user_id,notnull"`

	SessionSecret string `bun:"session_secret,nullzero"`

	ExpiresAt *time.Time `bun:"expires_at"`
}

// BeforeAppendModel is a hook that modifies the model on SELECT and UPDATE queries.
func (m *UserSession) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if err := m.Model.BeforeAppendModel(ctx, query); err != nil {
		return err
	}
	return nil
}

func userSessionToPB(m *UserSession, userIDs *ttnpb.UserIdentifiers) (*ttnpb.UserSession, error) {
	pb := &ttnpb.UserSession{
		UserIds:       userIDs,
		SessionId:     m.ID,
		CreatedAt:     ttnpb.ProtoTimePtr(m.CreatedAt),
		UpdatedAt:     ttnpb.ProtoTimePtr(m.UpdatedAt),
		ExpiresAt:     ttnpb.ProtoTime(m.ExpiresAt),
		SessionSecret: m.SessionSecret,
	}
	if userIDs == nil && m.User != nil {
		pb.UserIds = &ttnpb.UserIdentifiers{
			UserId: m.User.Account.UID,
		}
	}
	return pb, nil
}

type userSessionStore struct {
	*entityStore
}

func newUserSessionStore(baseStore *baseStore) *userSessionStore {
	return &userSessionStore{
		entityStore: newEntityStore(baseStore),
	}
}

func (s *userSessionStore) CreateSession(
	ctx context.Context, pb *ttnpb.UserSession,
) (*ttnpb.UserSession, error) {
	ctx, span := tracer.Start(ctx, "CreateSession", trace.WithAttributes(
		attribute.String("user_id", pb.GetUserIds().GetUserId()),
	))
	defer span.End()

	_, userUUID, err := s.getEntity(ctx, pb.GetUserIds())
	if err != nil {
		return nil, err
	}

	model := &UserSession{
		UserID:        userUUID,
		SessionSecret: pb.SessionSecret,
		ExpiresAt:     cleanTimePtr(ttnpb.StdTime(pb.ExpiresAt)),
	}

	_, err = s.DB.NewInsert().
		Model(model).
		Exec(ctx)
	if err != nil {
		return nil, errors.WrapDriverError(err)
	}

	pb, err = userSessionToPB(model, pb.GetUserIds())
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func (s *userSessionStore) listUserSessionsBy(
	ctx context.Context,
	by func(*bun.SelectQuery) *bun.SelectQuery,
) ([]*ttnpb.UserSession, error) {
	models := []*UserSession{}
	selectQuery := newSelectModels(ctx, s.DB, &models).Apply(by)

	// Count the total number of results.
	count, err := selectQuery.Count(ctx)
	if err != nil {
		return nil, errors.WrapDriverError(err)
	}
	store.SetTotal(ctx, uint64(count))

	// Apply ordering, paging and field masking.
	selectQuery = selectQuery.
		Apply(selectWithOrderFromContext(ctx, "id", map[string]string{
			"session_id": "id",
			"created_at": "created_at",
			"expires_at": "expires_at",
		})).
		Apply(selectWithLimitAndOffsetFromContext(ctx))

	// Scan the results.
	err = selectQuery.Scan(ctx)
	if err != nil {
		return nil, errors.WrapDriverError(err)
	}

	// Convert the results to protobuf.
	pbs := make([]*ttnpb.UserSession, len(models))
	for i, model := range models {
		pb, err := userSessionToPB(model, nil)
		if err != nil {
			return nil, err
		}
		pbs[i] = pb
	}

	return pbs, nil
}

func (*userSessionStore) selectWithUserIDs(
	_ context.Context, uuid string,
) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("user_id = ?", uuid)
	}
}

func (*userSessionStore) selectWithSessionID(
	_ context.Context, sessionID string,
) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("id = ?", sessionID)
	}
}

func (s *userSessionStore) FindSessions(
	ctx context.Context, userIDs *ttnpb.UserIdentifiers,
) ([]*ttnpb.UserSession, error) {
	ctx, span := tracer.Start(ctx, "FindSessions", trace.WithAttributes(
		attribute.String("user_id", userIDs.GetUserId()),
	))
	defer span.End()

	_, userUUID, err := s.getEntity(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	pbs, err := s.listUserSessionsBy(ctx, s.selectWithUserIDs(ctx, userUUID))
	if err != nil {
		return nil, err
	}

	for _, pb := range pbs {
		pb.UserIds = userIDs
	}

	return pbs, nil
}

func (s *userSessionStore) getUserSessionModelBy(
	ctx context.Context,
	by func(*bun.SelectQuery) *bun.SelectQuery,
) (*UserSession, error) {
	model := &UserSession{}
	selectQuery := s.newSelectModel(ctx, model).Apply(by)

	if err := selectQuery.Scan(ctx); err != nil {
		return nil, errors.WrapDriverError(err)
	}

	return model, nil
}

func (s *userSessionStore) GetSession(
	ctx context.Context, userIDs *ttnpb.UserIdentifiers, sessionID string,
) (*ttnpb.UserSession, error) {
	ctx, span := tracer.Start(ctx, "GetSession", trace.WithAttributes(
		attribute.String("user_id", userIDs.GetUserId()),
		attribute.String("session_id", sessionID),
	))
	defer span.End()

	_, userUUID, err := s.getEntity(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	model, err := s.getUserSessionModelBy(ctx, combineApply(
		s.selectWithUserIDs(ctx, userUUID),
		s.selectWithSessionID(ctx, sessionID),
	))
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, store.ErrUserSessionNotFound.WithAttributes(
				"user_id", userIDs.GetUserId(),
				"session_id", sessionID,
			)
		}
		return nil, err
	}

	pb, err := userSessionToPB(model, userIDs)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func (s *userSessionStore) GetSessionByID(ctx context.Context, sessionID string) (*ttnpb.UserSession, error) {
	ctx, span := tracer.Start(ctx, "GetSessionByID", trace.WithAttributes(
		attribute.String("session_id", sessionID),
	))
	defer span.End()

	model, err := s.getUserSessionModelBy(ctx, s.selectWithSessionID(ctx, sessionID))
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, store.ErrUserSessionNotFound.WithAttributes(
				"session_id", sessionID,
			)
		}
		return nil, err
	}
	pb, err := userSessionToPB(model, nil)
	if err != nil {
		return nil, err
	}

	friendlyUserID, err := s.getEntityID(ctx, "user", model.UserID)
	if err != nil {
		return nil, err
	}
	pb.UserIds = &ttnpb.UserIdentifiers{UserId: friendlyUserID}

	return pb, nil
}

func (s *userSessionStore) DeleteSession(ctx context.Context, userIDs *ttnpb.UserIdentifiers, sessionID string) error {
	ctx, span := tracer.Start(ctx, "DeleteSession", trace.WithAttributes(
		attribute.String("user_id", userIDs.GetUserId()),
		attribute.String("session_id", sessionID),
	))
	defer span.End()

	_, userUUID, err := s.getEntity(ctx, userIDs)
	if err != nil {
		return err
	}

	model, err := s.getUserSessionModelBy(ctx, combineApply(
		s.selectWithUserIDs(ctx, userUUID),
		s.selectWithSessionID(ctx, sessionID),
	))
	if err != nil {
		if errors.IsNotFound(err) {
			return store.ErrUserSessionNotFound.WithAttributes(
				"user_id", userIDs.GetUserId(),
				"session_id", sessionID,
			)
		}
		return err
	}

	_, err = s.DB.NewDelete().
		Model(model).
		WherePK().
		Exec(ctx)
	if err != nil {
		return errors.WrapDriverError(err)
	}
	return nil
}

func (s *userSessionStore) DeleteAllUserSessions(ctx context.Context, userIDs *ttnpb.UserIdentifiers) error {
	ctx, span := tracer.Start(ctx, "DeleteAllUserSessions", trace.WithAttributes(
		attribute.String("user_id", userIDs.GetUserId()),
	))
	defer span.End()

	_, userUUID, err := s.getEntity(store.WithSoftDeleted(ctx, false), userIDs)
	if err != nil {
		return err
	}

	_, err = s.DB.NewDelete().
		Model(&UserSession{}).
		Where("user_id = ?", userUUID).
		Exec(ctx)
	if err != nil {
		return errors.WrapDriverError(err)
	}

	return nil
}
