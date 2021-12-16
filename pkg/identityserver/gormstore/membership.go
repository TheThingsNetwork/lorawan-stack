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
	"fmt"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Membership model.
type Membership struct {
	Model

	Account    *Account
	AccountID  string `gorm:"type:UUID;index:membership_account_index;not null"`
	Rights     Rights `gorm:"type:INT ARRAY"`
	EntityID   string `gorm:"type:UUID;index:membership_entity_index;not null"`
	EntityType string `gorm:"type:VARCHAR(32);index:membership_entity_index;not null"`
}

func init() {
	registerModel(&Membership{})
}

type polymorphicEntity struct {
	EntityUUID string
	EntityType string
}

func (s *store) findIdentifiers(entities ...polymorphicEntity) (map[polymorphicEntity]*ttnpb.EntityIdentifiers, error) {
	return findIdentifiers(s.DB, entities...)
}

func findIdentifiers(db *gorm.DB, entities ...polymorphicEntity) (map[polymorphicEntity]*ttnpb.EntityIdentifiers, error) {
	var err error
	identifiers := make(map[polymorphicEntity]*ttnpb.EntityIdentifiers, len(entities))
	for _, entityType := range []string{"application", "client", "gateway", "organization", "user"} {
		uuids := make([]string, 0, len(entities))
		for _, entity := range entities {
			if entity.EntityType != entityType {
				continue
			}
			uuids = append(uuids, entity.EntityUUID)
		}
		if len(uuids) == 0 {
			continue
		}
		var results []struct {
			UUID       string
			FriendlyID string
		}
		if entityType == "organization" || entityType == "user" {
			err = db.Table("accounts").Select("account_id AS uuid, uid AS friendly_id").
				Where("account_type = ?", entityType).
				Where("account_id in (?)", uuids).
				Scan(&results).Error
		} else {
			err = db.Table(fmt.Sprintf("%ss", entityType)).Select(fmt.Sprintf("id as uuid, %s_id as friendly_id", entityType)).
				Where("id in (?)", uuids).Scan(&results).Error
		}
		if err != nil {
			return nil, err
		}
		for _, result := range results {
			entity := polymorphicEntity{EntityType: entityType, EntityUUID: result.UUID}
			identifiers[entity] = buildIdentifiers(entityType, result.FriendlyID)
		}
	}
	return identifiers, nil
}

func buildIdentifiers(entityType, id string) *ttnpb.EntityIdentifiers {
	switch entityType {
	case "application":
		return (&ttnpb.ApplicationIdentifiers{ApplicationId: id}).GetEntityIdentifiers()
	case "client":
		return (&ttnpb.ClientIdentifiers{ClientId: id}).GetEntityIdentifiers()
	case "gateway":
		return (&ttnpb.GatewayIdentifiers{GatewayId: id}).GetEntityIdentifiers()
	case "organization":
		return (&ttnpb.OrganizationIdentifiers{OrganizationId: id}).GetEntityIdentifiers()
	case "user":
		return (&ttnpb.UserIdentifiers{UserId: id}).GetEntityIdentifiers()
	default:
		panic(fmt.Sprintf("can't build identifiers for entity type %q", entityType))
	}
}
