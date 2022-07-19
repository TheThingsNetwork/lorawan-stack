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

// APIKey is the API key model in the database.
type APIKey struct {
	bun.BaseModel `bun:"table:api_keys,alias:key"`

	Model

	APIKeyID string `bun:"api_key_id,nullzero"`

	Key    string `bun:"key,nullzero"`
	Rights []int  `bun:"rights,array,nullzero"`
	Name   string `bun:"name,nullzero"`

	// EntityType is "application", "client", "end_device", "gateway", "organization" or "user".
	EntityType string `bun:"entity_type,notnull"`
	// EntityID is Application.ID, Client.ID, EndDevice.ID, Gateway.ID, Organization.ID or User.ID.
	EntityID string `bun:"entity_id,notnull"`

	ExpiresAt *time.Time `bun:"expires_at"`
}

// BeforeAppendModel is a hook that modifies the model on SELECT and UPDATE queries.
func (m *APIKey) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if err := m.Model.BeforeAppendModel(ctx, query); err != nil {
		return err
	}
	return nil
}

func apiKeyToPB(m *APIKey) (*ttnpb.APIKey, error) {
	pb := &ttnpb.APIKey{
		Id:        m.APIKeyID,
		Key:       m.Key,
		Name:      m.Name,
		Rights:    convertIntSlice[int, ttnpb.Right](m.Rights),
		CreatedAt: ttnpb.ProtoTimePtr(m.CreatedAt),
		UpdatedAt: ttnpb.ProtoTimePtr(m.UpdatedAt),
		ExpiresAt: ttnpb.ProtoTime(m.ExpiresAt),
	}
	return pb, nil
}

type apiKeyStore struct {
	*entityStore
}

func newAPIKeyStore(baseStore *baseStore) *apiKeyStore {
	return &apiKeyStore{
		entityStore: newEntityStore(baseStore),
	}
}

func (s *apiKeyStore) CreateAPIKey(
	ctx context.Context, entityID *ttnpb.EntityIdentifiers, pb *ttnpb.APIKey,
) (*ttnpb.APIKey, error) {
	ctx, span := tracer.Start(ctx, "CreateAPIKey", trace.WithAttributes(
		attribute.String("entity_type", entityID.EntityType()),
		attribute.String("entity_id", entityID.IDString()),
	))
	defer span.End()

	entityType, entityUUID, err := s.getEntity(ctx, entityID)
	if err != nil {
		return nil, err
	}

	model := &APIKey{
		APIKeyID:   pb.Id,
		Key:        pb.Key,
		Rights:     convertIntSlice[ttnpb.Right, int](pb.Rights),
		Name:       pb.Name,
		EntityType: entityType,
		EntityID:   entityUUID,
		ExpiresAt:  ttnpb.StdTime(pb.ExpiresAt),
	}

	_, err = s.DB.NewInsert().
		Model(model).
		Exec(ctx)
	if err != nil {
		return nil, wrapDriverError(err)
	}

	pb, err = apiKeyToPB(model)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func (s *apiKeyStore) listAPIKeysBy(
	ctx context.Context,
	by func(*bun.SelectQuery) *bun.SelectQuery,
) ([]*ttnpb.APIKey, error) {
	models := []*APIKey{}
	selectQuery := s.DB.NewSelect().
		Model(&models).
		Apply(by)

	// Count the total number of results.
	count, err := selectQuery.Count(ctx)
	if err != nil {
		return nil, wrapDriverError(err)
	}
	store.SetTotal(ctx, uint64(count))

	// Apply ordering, paging and field masking.
	selectQuery = selectQuery.
		Apply(selectWithOrderFromContext(ctx, "api_key_id", map[string]string{
			"api_key_id": "api_key_id",
			"name":       "name",
			"created_at": "created_at",
			"expires_at": "expires_at",
		})).
		Apply(selectWithLimitAndOffsetFromContext(ctx))

	// Scan the results.
	err = selectQuery.Scan(ctx)
	if err != nil {
		return nil, wrapDriverError(err)
	}

	// Convert the results to protobuf.
	pbs := make([]*ttnpb.APIKey, len(models))
	for i, model := range models {
		pb, err := apiKeyToPB(model)
		if err != nil {
			return nil, err
		}
		pbs[i] = pb
	}

	return pbs, nil
}

func (*apiKeyStore) selectWithEntityIDs(
	_ context.Context, entityType, entityUUID string,
) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("entity_type = ? AND entity_id = ?", entityType, entityUUID)
	}
}

func (*apiKeyStore) selectWithAPIKeyID(
	_ context.Context, apiKeyID string,
) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("api_key_id = ?", apiKeyID)
	}
}

func (s *apiKeyStore) FindAPIKeys(
	ctx context.Context, entityID *ttnpb.EntityIdentifiers,
) ([]*ttnpb.APIKey, error) {
	ctx, span := tracer.Start(ctx, "FindAPIKeys", trace.WithAttributes(
		attribute.String("entity_type", entityID.EntityType()),
		attribute.String("entity_id", entityID.IDString()),
	))
	defer span.End()

	entityType, entityUUID, err := s.getEntity(ctx, entityID)
	if err != nil {
		return nil, err
	}

	return s.listAPIKeysBy(ctx, s.selectWithEntityIDs(ctx, entityType, entityUUID))
}

func (s *apiKeyStore) getAPIKeyModelBy(
	ctx context.Context,
	by func(*bun.SelectQuery) *bun.SelectQuery,
) (*APIKey, error) {
	model := &APIKey{}
	selectQuery := s.DB.NewSelect().
		Model(model).
		Apply(by)

	if err := selectQuery.Scan(ctx); err != nil {
		return nil, wrapDriverError(err)
	}

	return model, nil
}

func (s *apiKeyStore) GetAPIKey(
	ctx context.Context, entityID *ttnpb.EntityIdentifiers, id string,
) (*ttnpb.APIKey, error) {
	ctx, span := tracer.Start(ctx, "GetAPIKey", trace.WithAttributes(
		attribute.String("entity_type", entityID.EntityType()),
		attribute.String("entity_id", entityID.IDString()),
		attribute.String("api_key_id", id),
	))
	defer span.End()

	entityType, entityUUID, err := s.getEntity(ctx, entityID)
	if err != nil {
		return nil, err
	}

	model, err := s.getAPIKeyModelBy(ctx, combineApply(
		s.selectWithEntityIDs(ctx, entityType, entityUUID),
		s.selectWithAPIKeyID(ctx, id),
	))
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, store.ErrAPIKeyNotFound.WithAttributes(
				"entity_type", entityType,
				"entity_id", entityID.IDString(),
				"api_key_id", id,
			)
		}
		return nil, err
	}
	pb, err := apiKeyToPB(model)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func (s *apiKeyStore) GetAPIKeyByID(
	ctx context.Context, id string,
) (*ttnpb.EntityIdentifiers, *ttnpb.APIKey, error) {
	ctx, span := tracer.Start(ctx, "GetAPIKeyByID", trace.WithAttributes(
		attribute.String("api_key_id", id),
	))
	defer span.End()

	model, err := s.getAPIKeyModelBy(ctx, s.selectWithAPIKeyID(ctx, id))
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil, store.ErrAPIKeyNotFound.WithAttributes(
				"api_key_id", id,
			)
		}
		return nil, nil, err
	}
	pb, err := apiKeyToPB(model)
	if err != nil {
		return nil, nil, err
	}
	friendlyID, err := s.getEntityID(ctx, model.EntityType, model.EntityID)
	if err != nil {
		return nil, nil, err
	}
	ids := getEntityIdentifiers(model.EntityType, friendlyID)

	return ids, pb, nil
}

func (s *apiKeyStore) UpdateAPIKey(
	ctx context.Context, entityID *ttnpb.EntityIdentifiers, pb *ttnpb.APIKey, fieldMask store.FieldMask,
) (*ttnpb.APIKey, error) {
	ctx, span := tracer.Start(ctx, "UpdateAPIKey", trace.WithAttributes(
		attribute.String("entity_type", entityID.EntityType()),
		attribute.String("entity_id", entityID.IDString()),
		attribute.String("api_key_id", pb.Id),
	))
	defer span.End()

	entityType, entityUUID, err := s.getEntity(ctx, entityID)
	if err != nil {
		return nil, err
	}

	model, err := s.getAPIKeyModelBy(ctx, combineApply(
		s.selectWithEntityIDs(ctx, entityType, entityUUID),
		s.selectWithAPIKeyID(ctx, pb.Id),
	))
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, store.ErrAPIKeyNotFound.WithAttributes(
				"entity_type", entityType,
				"entity_id", entityID.IDString(),
				"api_key_id", pb.Id,
			)
		}
		return nil, err
	}

	// If empty rights are passed and rights are in the fieldmask, delete the key.
	// TODO: Refactor store interface to move this to a DeleteAPIKey method.
	// (https://github.com/TheThingsNetwork/lorawan-stack/issues/5587)
	if len(pb.Rights) == 0 && ttnpb.HasAnyField(fieldMask, "rights") {
		_, err = s.DB.NewDelete().
			Model(model).
			WherePK().
			Exec(ctx)
		if err != nil {
			return nil, wrapDriverError(err)
		}
		return nil, nil //nolint:nilnil // This will be fixed by the refactor mentioned above.
	}

	columns := store.FieldMask{"updated_at"}

	for _, field := range fieldMask {
		switch field {
		case "rights":
			model.Rights = convertIntSlice[ttnpb.Right, int](pb.Rights)
			columns = append(columns, "rights")

		case "name":
			model.Name = pb.Name
			columns = append(columns, "name")

		case "expires_at":
			model.ExpiresAt = ttnpb.StdTime(pb.ExpiresAt)
			columns = append(columns, "expires_at")
		}
	}

	_, err = s.DB.NewUpdate().
		Model(model).
		WherePK().
		Column(columns...).
		Exec(ctx)
	if err != nil {
		return nil, wrapDriverError(err)
	}

	updatedPB, err := apiKeyToPB(model)
	if err != nil {
		return nil, err
	}

	return updatedPB, nil
}

func (s *apiKeyStore) DeleteEntityAPIKeys(ctx context.Context, entityID *ttnpb.EntityIdentifiers) error {
	ctx, span := tracer.Start(ctx, "DeleteEntityAPIKeys", trace.WithAttributes(
		attribute.String("entity_type", entityID.EntityType()),
		attribute.String("entity_id", entityID.IDString()),
	))
	defer span.End()

	entityType, entityUUID, err := s.getEntity(ctx, entityID)
	if err != nil {
		return err
	}

	_, err = s.DB.NewDelete().
		Model(&APIKey{}).
		Where("entity_type = ? AND entity_id = ?", entityType, entityUUID).
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	return nil
}
