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
	"fmt"

	"github.com/uptrace/bun"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type entityStore struct {
	*baseStore
	*applicationStore
	*clientStore
	*gatewayStore
	*organizationStore
	*userStore
}

func newEntityStore(baseStore *baseStore) *entityStore {
	return &entityStore{
		baseStore:         baseStore,
		applicationStore:  &applicationStore{baseStore},
		clientStore:       &clientStore{baseStore},
		gatewayStore:      &gatewayStore{baseStore},
		organizationStore: &organizationStore{baseStore},
		userStore:         &userStore{baseStore},
	}
}

// entityFriendlyIDs is the model for the entity_friendly_ids view in the database.
type entityFriendlyIDs struct {
	bun.BaseModel `bun:"table:entity_friendly_ids,alias:ids"`

	EntityType string `bun:"entity_type,notnull"`
	EntityID   string `bun:"entity_id,notnull"`
	FriendlyID string `bun:"friendly_id,notnull"`
}

func (s *entityStore) getEntityUUID(ctx context.Context, entityType, entityID string) (string, error) {
	var uuid string
	err := s.DB.NewSelect().
		Model(&entityFriendlyIDs{}).
		Column("entity_id").
		Apply(selectWithContext(ctx)).
		Where("entity_type = ?", entityType).
		Where("friendly_id = ?", entityID).
		Scan(ctx, &uuid)
	if err != nil {
		return "", wrapDriverError(err)
	}
	return uuid, nil
}

func (s *entityStore) getEntityID(ctx context.Context, entityType, entityUUID string) (string, error) {
	var friendlyID string
	err := s.DB.NewSelect().
		Model(&entityFriendlyIDs{}).
		Column("friendly_id").
		Apply(selectWithContext(ctx)).
		Where("entity_type = ?", entityType).
		Where("entity_id = ?", entityUUID).
		Scan(ctx, &friendlyID)
	if err != nil {
		return "", wrapDriverError(err)
	}
	return friendlyID, nil
}

func (*entityStore) getEntityIdentifiers(entityType string, friendlyID string) *ttnpb.EntityIdentifiers {
	switch entityType {
	default:
		panic(fmt.Errorf("invalid entity type: %s", entityType))
	case "application":
		return (&ttnpb.ApplicationIdentifiers{ApplicationId: friendlyID}).GetEntityIdentifiers()
	case "client":
		return (&ttnpb.ClientIdentifiers{ClientId: friendlyID}).GetEntityIdentifiers()
	case "gateway":
		return (&ttnpb.GatewayIdentifiers{GatewayId: friendlyID}).GetEntityIdentifiers()
	case "organization":
		return (&ttnpb.OrganizationIdentifiers{OrganizationId: friendlyID}).GetEntityIdentifiers()
	case "user":
		return (&ttnpb.UserIdentifiers{UserId: friendlyID}).GetEntityIdentifiers()
	}
}
