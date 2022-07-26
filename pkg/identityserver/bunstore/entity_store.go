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
	"strings"

	"github.com/uptrace/bun"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type entityStore struct {
	*baseStore
	*applicationStore
	*clientStore
	*endDeviceStore
	*gatewayStore
	*organizationStore
	*userStore
}

func newEntityStore(baseStore *baseStore) *entityStore {
	return &entityStore{
		baseStore:         baseStore,
		applicationStore:  &applicationStore{baseStore},
		clientStore:       &clientStore{baseStore},
		endDeviceStore:    &endDeviceStore{baseStore},
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

type identifiers interface {
	GetEntityIdentifiers() *ttnpb.EntityIdentifiers
}

var entityTypeReplacer = strings.NewReplacer(" ", "_")

func getEntityType(ids ttnpb.IDStringer) string {
	return entityTypeReplacer.Replace(ids.EntityType())
}

func (s *entityStore) getEntity(ctx context.Context, ids ttnpb.IDStringer) (entityType, entityUUID string, err error) {
	entityType = getEntityType(ids)

	if entityType == "end_device" {
		entityIDs, ok := ids.(identifiers)
		if !ok {
			entityIDs = getEntityIdentifiers(entityType, ids.IDString())
		}
		devIDs := entityIDs.GetEntityIdentifiers().GetDeviceIds()
		model, err := s.getEndDeviceModelBy(
			ctx,
			s.endDeviceStore.selectWithID(ctx, devIDs.GetApplicationIds().GetApplicationId(), devIDs.GetDeviceId()),
			store.FieldMask{"ids"},
		)
		if err != nil {
			return "", "", err
		}
		return entityType, model.ID, nil
	}

	var uuid string
	err = s.DB.NewSelect().
		Model(&entityFriendlyIDs{}).
		Column("entity_id").
		Apply(selectWithContext(ctx)).
		Where("entity_type = ?", strings.ReplaceAll(ids.EntityType(), " ", "_")).
		Where("friendly_id = ?", ids.IDString()).
		Scan(ctx, &uuid)
	if err != nil {
		err = wrapDriverError(err)
		if errors.IsNotFound(err) {
			return "", "", store.ErrEntityNotFound.WithAttributes(
				"entity_type", ids.EntityType(),
				"entity_id", ids.IDString(),
			)
		}
		return "", "", err
	}
	return entityType, uuid, nil
}

func (s *entityStore) getEntityUUIDs(ctx context.Context, entityType string, entityIDs ...string) ([]string, error) {
	var uuids []string
	err := s.DB.NewSelect().
		Model(&entityFriendlyIDs{}).
		Column("entity_id").
		Apply(selectWithContext(ctx)).
		Where("entity_type = ?", entityType).
		Where("friendly_id IN (?)", bun.In(entityIDs)).
		Scan(ctx, &uuids)
	if err != nil {
		return nil, wrapDriverError(err)
	}
	return uuids, nil
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

func getEntityIdentifiers(entityType string, friendlyID string) *ttnpb.EntityIdentifiers {
	switch entityType {
	default:
		panic(fmt.Errorf("invalid entity type: %s", entityType))
	case "application":
		return (&ttnpb.ApplicationIdentifiers{ApplicationId: friendlyID}).GetEntityIdentifiers()
	case "client":
		return (&ttnpb.ClientIdentifiers{ClientId: friendlyID}).GetEntityIdentifiers()
	case "end_device", "end device":
		parts := strings.SplitN(friendlyID, ".", 2)
		return (&ttnpb.EndDeviceIdentifiers{
			ApplicationIds: &ttnpb.ApplicationIdentifiers{
				ApplicationId: parts[0],
			},
			DeviceId: parts[1],
		}).GetEntityIdentifiers()
	case "gateway":
		return (&ttnpb.GatewayIdentifiers{GatewayId: friendlyID}).GetEntityIdentifiers()
	case "organization":
		return (&ttnpb.OrganizationIdentifiers{OrganizationId: friendlyID}).GetEntityIdentifiers()
	case "user":
		return (&ttnpb.UserIdentifiers{UserId: friendlyID}).GetEntityIdentifiers()
	}
}
