// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// Membership model.
type Membership struct {
	Model

	Account    *Account
	AccountID  string `gorm:"type:UUID;index;not null"`
	Rights     Rights `gorm:"type:INT ARRAY"`
	EntityID   string `gorm:"type:UUID;index;not null"`
	EntityType string `gorm:"index;not null"`
}

func init() {
	registerModel(&Membership{})
}

func findAccountMemberships(db *gorm.DB, account *Account, entityType string) ([]*Membership, error) {
	query := db.Where(&Account{AccountID: account.ID})
	if entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}
	var memberships []*Membership
	err := query.Find(&memberships).Error
	if err != nil {
		return nil, err
	}
	return memberships, nil
}

func reduceAccountMemberships(db *gorm.DB, account *Account, entityType, entityUUID string) ([]*Membership, error) {
	query := db.Where(&Account{AccountID: account.ID})

	switch {
	case entityType != "" && entityUUID != "":
		query = query.Where("entity_type = ? OR (entity_type = ? AND entity_id = ?)", "organization", entityType, entityUUID)
	case entityType != "":
		query = query.Where("entity_type IN (?)", []string{"organization", entityType})
	}

	var directMemberships []Membership
	err := query.Find(&directMemberships).Error
	if err != nil {
		return nil, err
	}

	type eid struct {
		entityType string
		entityUUID string
	}
	entityRights := make(map[eid]*ttnpb.Rights)
	organizationRights := make(map[string]*ttnpb.Rights, len(directMemberships))

	for _, membership := range directMemberships {
		membershipRights := ttnpb.Rights(membership.Rights)
		membershipRights = *membershipRights.Implied()
		entityRights[eid{membership.EntityType, membership.EntityID}] = &membershipRights
		if membership.EntityType == "organization" {
			organizationRights[membership.EntityID] = &membershipRights
		}
	}

	organizationUUIDs := make([]string, 0, len(organizationRights))
	for uuid := range organizationRights {
		organizationUUIDs = append(organizationUUIDs, uuid)
	}

	query = db.Table("memberships").
		Select("memberships.*, accounts.account_id AS organization_id").
		Joins("LEFT JOIN accounts ON accounts.id = memberships.account_id").
		Where("accounts.account_type = ? AND accounts.account_id IN (?)", "organization", organizationUUIDs)

	if entityType != "" {
		query = query.Where("memberships.entity_type = ?", entityType)
	}
	if entityUUID != "" {
		query = query.Where("memberships.entity_id = ?", entityUUID)
	}

	type organizationMembership struct {
		Membership
		OrganizationID string
	}
	var organizationMemberships []organizationMembership
	err = query.Scan(&organizationMemberships).Error
	if err != nil {
		return nil, err
	}

	for _, membership := range organizationMemberships {
		membershipRights := ttnpb.Rights(membership.Rights)
		membershipRights = *membershipRights.Implied()
		indirectRights := membershipRights.Intersect(organizationRights[membership.OrganizationID])
		existingRights := entityRights[eid{membership.EntityType, membership.EntityID}]
		entityRights[eid{membership.EntityType, membership.EntityID}] = existingRights.Union(indirectRights).Sorted()
	}

	memberships := make([]*Membership, 0, len(entityRights))
	for e, rights := range entityRights {
		if entityType != "" && e.entityType != entityType {
			continue
		}
		if entityUUID != "" && e.entityUUID != entityUUID {
			continue
		}
		memberships = append(memberships, &Membership{
			EntityID:   e.entityUUID,
			EntityType: e.entityType,
			Rights:     Rights(*rights),
		})
	}

	return memberships, nil
}

func entityRightsForMemberships(memberships []*Membership) map[polymorphicEntity]Rights {
	res := make(map[polymorphicEntity]Rights, len(memberships))
	for _, membership := range memberships {
		k := polymorphicEntity{EntityType: membership.EntityType, EntityUUID: membership.EntityID}
		res[k] = membership.Rights
	}
	return res
}

type polymorphicEntity struct {
	EntityUUID string
	EntityType string
}

func identifiers(db *gorm.DB, entities ...polymorphicEntity) (map[polymorphicEntity]*ttnpb.EntityIdentifiers, error) {
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
			switch entityType {
			case "application":
				identifiers[entity] = ttnpb.ApplicationIdentifiers{ApplicationID: result.FriendlyID}.EntityIdentifiers()
			case "client":
				identifiers[entity] = ttnpb.ClientIdentifiers{ClientID: result.FriendlyID}.EntityIdentifiers()
			case "gateway":
				identifiers[entity] = ttnpb.GatewayIdentifiers{GatewayID: result.FriendlyID}.EntityIdentifiers()
			case "organization":
				identifiers[entity] = ttnpb.OrganizationIdentifiers{OrganizationID: result.FriendlyID}.EntityIdentifiers()
			case "user":
				identifiers[entity] = ttnpb.UserIdentifiers{UserID: result.FriendlyID}.EntityIdentifiers()
			}
		}
	}
	return identifiers, nil
}

func identifierRights(db *gorm.DB, entityRights map[polymorphicEntity]Rights) (map[*ttnpb.EntityIdentifiers]*ttnpb.Rights, error) {
	entities := make([]polymorphicEntity, 0, len(entityRights))
	for entity := range entityRights {
		entities = append(entities, entity)
	}
	identifiers, err := identifiers(db, entities...)
	if err != nil {
		return nil, err
	}
	identifierRights := make(map[*ttnpb.EntityIdentifiers]*ttnpb.Rights, len(entityRights))
	for entity, rights := range entityRights {
		ids, ok := identifiers[entity]
		if !ok {
			continue
		}
		rights := ttnpb.Rights(rights)
		identifierRights[ids] = &rights
	}
	return identifierRights, nil
}
