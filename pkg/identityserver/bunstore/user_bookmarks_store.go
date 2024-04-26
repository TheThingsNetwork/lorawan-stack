// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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
	"strings"

	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/tracing/tracer"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	storeutil "go.thethings.network/lorawan-stack/v3/pkg/util/store"
)

// UserBookmarks is the user_bookmarks model in the database.
type UserBookmarks struct {
	bun.BaseModel `bun:"table:user_bookmarks"`
	Model
	SoftDelete

	UserID     string `bun:"user_id,notnull,nullzero"`
	EntityType string `bun:"entity_type,notnull,nullzero"`
	EntityID   string `bun:"entity_id,notnull,nullzero"`
}

// BeforeAppendModel is a hook that modifies the model on SELECT and UPDATE queries.
func (m *UserBookmarks) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if err := m.Model.BeforeAppendModel(ctx, query); err != nil {
		return err
	}
	return nil
}

func userBookmarkToPB(m *UserBookmarks, _ ...string) (*ttnpb.UserBookmark, error) {
	res := &ttnpb.UserBookmark{
		UserIds: &ttnpb.UserIdentifiers{UserId: m.UserID},
	}

	switch m.EntityType {
	case store.EntityApplication:
		res.EntityIds = (&ttnpb.ApplicationIdentifiers{ApplicationId: m.EntityID}).GetEntityIdentifiers()
	case store.EntityClient:
		res.EntityIds = (&ttnpb.ClientIdentifiers{ClientId: m.EntityID}).GetEntityIdentifiers()
	case store.EntityGateway:
		res.EntityIds = (&ttnpb.GatewayIdentifiers{GatewayId: m.EntityID}).GetEntityIdentifiers()
	case store.EntityOrganization:
		res.EntityIds = (&ttnpb.OrganizationIdentifiers{OrganizationId: m.EntityID}).GetEntityIdentifiers()
	case store.EntityUser:
		res.EntityIds = (&ttnpb.UserIdentifiers{UserId: m.EntityID}).GetEntityIdentifiers()
	case store.EntityEndDevice:
		devIDs := strings.Split(m.EntityID, ".")
		res.EntityIds = (&ttnpb.EndDeviceIdentifiers{
			DeviceId:       devIDs[1],
			ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: devIDs[0]},
		}).GetEntityIdentifiers()
	default:
		return nil, store.ErrInvalidEntityType.WithAttributes("entity_type", m.EntityType)
	}

	return res, nil
}

type userBookmarkStore struct{ *entityStore }

func newUserBookmarkStore(baseStore *baseStore) *userBookmarkStore {
	return &userBookmarkStore{entityStore: newEntityStore(baseStore)}
}

func (*userBookmarkStore) selectWithUserID(
	_ context.Context, ids ...string,
) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		switch len(ids) {
		case 0:
			return q
		case 1:
			return q.Where("?TableAlias.user_id = ?", ids[0])
		default:
			return q.Where("?TableAlias.user_id IN (?)", bun.In(ids))
		}
	}
}

func (*userBookmarkStore) selectWithEntityType(
	_ context.Context, ids ...string,
) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		switch len(ids) {
		case 0:
			return q
		case 1:
			return q.Where("?TableAlias.entity_type = ?", ids[0])
		default:
			return q.Where("?TableAlias.entity_type IN (?)", bun.In(ids))
		}
	}
}

func (*userBookmarkStore) selectWithEntityID(
	ctx context.Context, ids ...ttnpb.IDStringer,
) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		switch len(ids) {
		case 0:
			return q
		case 1:
			return q.Where(
				"?TableAlias.entity_type = ? AND ?TableAlias.entity_id = ?",
				ids[0].EntityType(),
				ids[0].IDString(),
			)
		default:
			entityByType := make(map[string][]string)
			for _, id := range ids {
				entityByType[id.EntityType()] = append(entityByType[id.EntityType()], id.IDString())
			}
			q = q.WhereGroup(" AND ", func(sq *bun.SelectQuery) *bun.SelectQuery {
				for entityType, ids := range entityByType {
					sq = sq.WhereOr(
						"?TableAlias.entity_type = ? AND ?TableAlias.entity_id IN (?)",
						entityType,
						bun.In(ids),
					)
				}
				return sq
			})
			return q
		}
	}
}

func (s *userBookmarkStore) getUserBookmarkBy(
	ctx context.Context,
	by func(*bun.SelectQuery) *bun.SelectQuery,
) (*UserBookmarks, error) {
	model := &UserBookmarks{}
	selectQuery := s.newSelectModel(ctx, model).Apply(by)

	if err := selectQuery.Scan(ctx); err != nil {
		return nil, storeutil.WrapDriverError(err)
	}

	return model, nil
}

func (s *userBookmarkStore) listUserBookmarksBy(
	ctx context.Context,
	by func(*bun.SelectQuery) *bun.SelectQuery,
) ([]*UserBookmarks, error) {
	models := []*UserBookmarks{}
	selectQuery := newSelectModels(ctx, s.DB, &models).Apply(by)

	// Count the total number of results.
	count, err := selectQuery.Count(ctx)
	if err != nil {
		return nil, storeutil.WrapDriverError(err)
	}
	store.SetTotal(ctx, uint64(count))

	// Apply ordering and paging.
	selectQuery = selectQuery.
		Apply(selectWithOrderFromContext(ctx, "created_at", map[string]string{
			"user_id":     "user_id",
			"entity_type": "entity_type",
			"entity_id":   "entity_id",
			"created_at":  "created_at",
		})).
		Apply(selectWithLimitAndOffsetFromContext(ctx))

	// Scan the results.
	err = selectQuery.Scan(ctx)
	if err != nil {
		return nil, storeutil.WrapDriverError(err)
	}
	return models, nil
}

func (s *userBookmarkStore) CreateBookmark(
	ctx context.Context, pb *ttnpb.UserBookmark,
) (*ttnpb.UserBookmark, error) {
	ctx, span := tracer.StartFromContext(ctx, "CreateBookmark", trace.WithAttributes(
		attribute.String("user_id", pb.GetUserIds().UserId),
		attribute.String("entity_type", pb.GetEntityIds().EntityType()),
		attribute.String("entity_id", pb.GetEntityIds().IDString()),
	))
	defer span.End()

	model := &UserBookmarks{
		UserID:     pb.UserIds.UserId,
		EntityID:   pb.GetEntityIds().IDString(),
		EntityType: pb.GetEntityIds().EntityType(),
	}

	err := s.transact(ctx, func(ctx context.Context, tx bun.IDB) error {
		_, _, err := s.getEntity(ctx, pb.GetEntityIds())
		if err != nil {
			return err
		}
		_, err = tx.NewInsert().
			Model(model).
			Exec(ctx)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, storeutil.WrapDriverError(err)
	}

	return userBookmarkToPB(model)
}

func (s *userBookmarkStore) FindBookmarks(
	ctx context.Context, id *ttnpb.UserIdentifiers, entityTypes ...string,
) ([]*ttnpb.UserBookmark, error) {
	ctx, span := tracer.StartFromContext(ctx, "FindBookmarks", trace.WithAttributes(
		attribute.String("user_id", id.GetUserId()),
	))
	defer span.End()

	models, err := s.listUserBookmarksBy(
		ctx,
		func(sq *bun.SelectQuery) *bun.SelectQuery {
			sq = s.selectWithUserID(ctx, id.GetUserId())(sq)
			sq = s.selectWithEntityType(ctx, entityTypes...)(sq)
			return sq
		},
	)
	if err != nil {
		return nil, err
	}

	pbs := make([]*ttnpb.UserBookmark, len(models))
	for i, model := range models {
		pb, err := userBookmarkToPB(model)
		if err != nil {
			return nil, err
		}
		pbs[i] = pb
	}

	return pbs, nil
}

func (s *userBookmarkStore) PurgeBookmark(
	ctx context.Context, pb *ttnpb.UserBookmark,
) error {
	ctx, span := tracer.StartFromContext(ctx, "PurgeBookmark", trace.WithAttributes(
		attribute.String("user_id", pb.GetUserIds().UserId),
		attribute.String("entity_type", pb.GetEntityIds().EntityType()),
		attribute.String("entity_id", pb.GetEntityIds().IDString()),
	))
	defer span.End()

	model, err := s.getUserBookmarkBy(
		ctx,
		func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("?TableAlias.user_id = ? AND ?TableAlias.entity_type = ? AND ?TableAlias.entity_id = ?",
				pb.GetUserIds().UserId,
				pb.GetEntityIds().EntityType(),
				pb.GetEntityIds().IDString(),
			)
		},
	)
	if err != nil {
		if errors.IsNotFound(err) {
			return store.ErrUserBookmarkNotFound.New()
		}
		return err
	}

	_, err = s.DB.NewDelete().
		Model(model).
		WherePK().
		ForceDelete().
		Exec(ctx)
	if err != nil {
		return storeutil.WrapDriverError(err)
	}

	return nil
}

func (s *userBookmarkStore) BatchPurgeBookmarks(
	ctx context.Context, usrID *ttnpb.UserIdentifiers, entityIDs []*ttnpb.EntityIdentifiers,
) ([]*ttnpb.UserBookmark, error) {
	ctx, span := tracer.StartFromContext(ctx, "BatchPurgeBookmarks", trace.WithAttributes(
		attribute.String("user_id", usrID.IDString()),
	))
	defer span.End()

	entities := make([]ttnpb.IDStringer, len(entityIDs))
	for i, entityID := range entityIDs {
		entities[i] = entityID
	}

	models, err := s.listUserBookmarksBy(ctx,
		func(sq *bun.SelectQuery) *bun.SelectQuery {
			sq = s.selectWithUserID(ctx, usrID.GetUserId())(sq)
			sq = s.selectWithEntityID(ctx, entities...)(sq)
			return sq
		})
	if err != nil {
		return nil, err
	}

	if len(models) > 0 {
		_, err = s.DB.NewDelete().
			Model(&models).
			WherePK().
			ForceDelete().
			Exec(ctx)
		if err != nil {
			return nil, storeutil.WrapDriverError(err)
		}
	}

	// Convert the results to protobuf.
	pbs := make([]*ttnpb.UserBookmark, len(models))
	for i, model := range models {
		pb, err := userBookmarkToPB(model)
		if err != nil {
			return nil, err
		}
		pbs[i] = pb
	}

	return pbs, nil
}

func (s *userBookmarkStore) DeleteEntityBookmarks(
	ctx context.Context, entityID *ttnpb.EntityIdentifiers,
) error {
	ctx, span := tracer.StartFromContext(ctx, "DeleteEntityBookmarks", trace.WithAttributes(
		attribute.String("entity_id", entityID.IDString()),
	))
	defer span.End()

	models, err := s.listUserBookmarksBy(ctx, s.selectWithEntityID(ctx, entityID))
	if err != nil {
		return err
	}

	if len(models) > 0 {
		_, err = s.DB.NewDelete().
			Model(&models).
			WherePK().
			Exec(ctx)
		if err != nil {
			return storeutil.WrapDriverError(err)
		}
	}

	return nil
}

func (s *userBookmarkStore) RestoreEntityBookmarks(
	ctx context.Context, entityID *ttnpb.EntityIdentifiers,
) error {
	ctx, span := tracer.StartFromContext(ctx, "RestoreEntityBookmarks", trace.WithAttributes(
		attribute.String("entity_id", entityID.IDString()),
	))
	defer span.End()

	models, err := s.listUserBookmarksBy(
		store.WithSoftDeleted(ctx, true),
		s.selectWithEntityID(ctx, entityID),
	)
	if err != nil {
		return err
	}

	if len(models) > 0 {
		_, err = s.DB.NewUpdate().
			Model(&models).
			WherePK().
			WhereAllWithDeleted().
			Set("deleted_at = NULL").
			Exec(ctx)
		if err != nil {
			return storeutil.WrapDriverError(err)
		}
	}

	return nil
}

func (s *userBookmarkStore) PurgeEntityBookmarks(
	ctx context.Context, entityID *ttnpb.EntityIdentifiers,
) error {
	ctx, span := tracer.StartFromContext(ctx, "PurgeEntityBookmarks", trace.WithAttributes(
		attribute.String("entity_id", entityID.IDString()),
	))
	defer span.End()

	models, err := s.listUserBookmarksBy(
		store.WithSoftDeleted(ctx, false),
		s.selectWithEntityID(ctx, entityID),
	)
	if err != nil {
		return err
	}

	if len(models) > 0 {
		_, err = s.DB.NewDelete().
			Model(&models).
			WherePK().
			ForceDelete().
			Exec(ctx)
		if err != nil {
			return storeutil.WrapDriverError(err)
		}
	}

	return nil
}

func (s *userBookmarkStore) DeleteUserBookmarks(
	ctx context.Context, usrID *ttnpb.UserIdentifiers,
) error {
	ctx, span := tracer.StartFromContext(ctx, "DeleteUserBookmarks", trace.WithAttributes(
		attribute.String("user_id", usrID.GetUserId()),
	))
	defer span.End()

	models, err := s.listUserBookmarksBy(ctx, s.selectWithUserID(ctx, usrID.GetUserId()))
	if err != nil {
		return err
	}

	if len(models) > 0 {
		_, err = s.DB.NewDelete().
			Model(&models).
			WherePK().
			Exec(ctx)
		if err != nil {
			return storeutil.WrapDriverError(err)
		}
	}

	return nil
}

func (s *userBookmarkStore) RestoreUserBookmarks(
	ctx context.Context, usrID *ttnpb.UserIdentifiers,
) error {
	ctx, span := tracer.StartFromContext(ctx, "RestoreUserBookmarks", trace.WithAttributes(
		attribute.String("user_id", usrID.GetUserId()),
	))
	defer span.End()

	models, err := s.listUserBookmarksBy(
		store.WithSoftDeleted(ctx, true),
		s.selectWithUserID(ctx, usrID.GetUserId()),
	)
	if err != nil {
		return err
	}

	if len(models) > 0 {
		_, err = s.DB.NewUpdate().
			Model(&models).
			WherePK().
			WhereAllWithDeleted().
			Set("deleted_at = NULL").
			Exec(ctx)
		if err != nil {
			return storeutil.WrapDriverError(err)
		}
	}

	return nil
}

func (s *userBookmarkStore) PurgeUserBookmarks(
	ctx context.Context, usrID *ttnpb.UserIdentifiers,
) error {
	ctx, span := tracer.StartFromContext(ctx, "PurgeUserBookmarks", trace.WithAttributes(
		attribute.String("user_id", usrID.GetUserId()),
	))
	defer span.End()

	models, err := s.listUserBookmarksBy(
		store.WithSoftDeleted(ctx, false),
		s.selectWithUserID(ctx, usrID.GetUserId()),
	)
	if err != nil {
		return err
	}

	if len(models) > 0 {
		_, err = s.DB.NewDelete().
			Model(&models).
			WherePK().
			ForceDelete().
			Exec(ctx)
		if err != nil {
			return storeutil.WrapDriverError(err)
		}
	}

	return nil
}
