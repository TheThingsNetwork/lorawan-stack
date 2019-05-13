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
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// GetAPIKeyStore returns an APIKeyStore on the given db (or transaction).
func GetAPIKeyStore(db *gorm.DB) APIKeyStore {
	return &apiKeyStore{store: newStore(db)}
}

type apiKeyStore struct {
	*store
}

func (s *apiKeyStore) CreateAPIKey(ctx context.Context, entityID *ttnpb.EntityIdentifiers, key *ttnpb.APIKey) error {
	defer trace.StartRegion(ctx, "create api key").End()
	entity, err := s.findEntity(ctx, entityID, "id")
	if err != nil {
		return err
	}
	model := &APIKey{
		APIKeyID:   key.ID,
		Key:        key.Key,
		Rights:     Rights{Rights: key.Rights},
		Name:       key.Name,
		EntityID:   entity.PrimaryKey(),
		EntityType: entityTypeForID(entityID),
	}
	return s.createEntity(ctx, model)
}

func (s *apiKeyStore) FindAPIKeys(ctx context.Context, entityID *ttnpb.EntityIdentifiers) ([]*ttnpb.APIKey, error) {
	defer trace.StartRegion(ctx, "find api keys").End()
	entity, err := s.findEntity(ctx, entityID, "id")
	if err != nil {
		return nil, err
	}
	query := s.query(ctx, APIKey{}).Where(&APIKey{
		EntityID:   entity.PrimaryKey(),
		EntityType: entityTypeForID(entityID),
	})
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 {
		countTotal(ctx, query)
		query = query.Limit(limit).Offset(offset)
	}
	var keyModels []APIKey
	if err = query.Find(&keyModels).Error; err != nil {
		return nil, err
	}
	setTotal(ctx, uint64(len(keyModels)))
	keyProtos := make([]*ttnpb.APIKey, len(keyModels))
	for i, apiKey := range keyModels {
		keyProtos[i] = apiKey.toPB()
	}
	return keyProtos, nil
}

var errAPIKeyEntity = errors.DefineCorruption("api_key_entity", "API key not linked to an entity")

func (s *apiKeyStore) GetAPIKey(ctx context.Context, id string) (*ttnpb.EntityIdentifiers, *ttnpb.APIKey, error) {
	defer trace.StartRegion(ctx, "get api key").End()
	query := s.query(ctx, APIKey{})
	var keyModel APIKey
	if err := query.Where(APIKey{APIKeyID: id}).First(&keyModel).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, nil, errAPIKeyNotFound
		}
		return nil, nil, err
	}
	if err := ctx.Err(); err != nil { // Early exit if context canceled
		return nil, nil, err
	}
	k := polymorphicEntity{EntityType: keyModel.EntityType, EntityUUID: keyModel.EntityID}
	identifiers, err := s.findIdentifiers(k)
	if err != nil {
		return nil, nil, err
	}
	ids, ok := identifiers[k]
	if !ok {
		return nil, nil, errAPIKeyEntity
	}
	return ids, keyModel.toPB(), nil
}

func (s *apiKeyStore) UpdateAPIKey(ctx context.Context, entityID *ttnpb.EntityIdentifiers, key *ttnpb.APIKey) (*ttnpb.APIKey, error) {
	defer trace.StartRegion(ctx, "update api key").End()
	entity, err := s.findEntity(ctx, entityID, "id")
	if err != nil {
		return nil, err
	}
	query := s.query(ctx, APIKey{})
	var keyModel APIKey
	err = query.Where(APIKey{
		APIKeyID:   key.ID,
		EntityID:   entity.PrimaryKey(),
		EntityType: entityTypeForID(entityID),
	}).First(&keyModel).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errAPIKeyNotFound
		}
		return nil, err
	}
	if len(key.Rights) == 0 {
		return nil, query.Delete(&keyModel).Error
	}
	keyModel.Name = key.Name
	keyModel.Rights = Rights{Rights: key.Rights}
	if err = query.Select("name", "rights", "updated_at").Save(&keyModel).Error; err != nil {
		return nil, err
	}
	return keyModel.toPB(), nil
}
