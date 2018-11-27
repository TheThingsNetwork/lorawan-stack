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
	"context"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// GetContactInfoStore returns an ContactInfoStore on the given db (or transaction).
func GetContactInfoStore(db *gorm.DB) ContactInfoStore {
	return &contactInfoStore{db: db}
}

type contactInfoStore struct {
	db *gorm.DB
}

func (s *contactInfoStore) GetContactInfo(ctx context.Context, entityID *ttnpb.EntityIdentifiers) ([]*ttnpb.ContactInfo, error) {
	entity, err := findEntity(ctx, s.db, entityID, "id")
	if err != nil {
		return nil, err
	}
	var models []ContactInfo
	err = s.db.Where(&ContactInfo{
		EntityType: entityTypeForID(entityID),
		EntityID:   entity.PrimaryKey(),
	}).Find(&models).Error
	if err != nil {
		return nil, err
	}
	pb := make([]*ttnpb.ContactInfo, len(models))
	for i, model := range models {
		pb[i] = model.toPB()
	}
	return pb, nil
}

func (s *contactInfoStore) SetContactInfo(ctx context.Context, entityID *ttnpb.EntityIdentifiers, pb []*ttnpb.ContactInfo) ([]*ttnpb.ContactInfo, error) {
	entity, err := findEntity(ctx, s.db, entityID, "id")
	if err != nil {
		return nil, err
	}
	entityType, entityUUID := entityTypeForID(entityID), entity.PrimaryKey()

	var existing []ContactInfo
	err = s.db.Where(&ContactInfo{
		EntityType: entityType,
		EntityID:   entityUUID,
	}).Find(&existing).Error
	if err != nil {
		return nil, err
	}

	type contactInfoID struct {
		contactType   int
		contactMethod int
		value         string
	}
	type contactInfo struct {
		ContactInfo
		deleted bool
	}

	existingByUUID := make(map[string]ContactInfo, len(existing))
	existingByInfo := make(map[contactInfoID]*contactInfo, len(existing))

	for _, existing := range existing {
		existingByUUID[existing.ID] = existing
		existingByInfo[contactInfoID{existing.ContactType, existing.ContactMethod, existing.Value}] = &contactInfo{
			ContactInfo: existing,
			deleted:     true,
		}
	}

	var toCreate []*ContactInfo
	for _, pb := range pb {
		if existing, ok := existingByInfo[contactInfoID{int(pb.ContactType), int(pb.ContactMethod), pb.Value}]; ok {
			existing.deleted = false
			existing.fromPB(pb)
		} else {
			model := ContactInfo{}
			model.fromPB(pb)
			toCreate = append(toCreate, &model)
		}
	}

	var toUpdate []*ContactInfo
	var toDelete []string
	for _, existing := range existingByInfo {
		if existing.deleted {
			toDelete = append(toDelete, existing.ContactInfo.ID)
		} else {
			toUpdate = append(toUpdate, &existing.ContactInfo)
		}
	}

	for _, info := range toCreate {
		info.EntityType, info.EntityID = entityType, entityUUID
		err = s.db.Save(&info).Error
		if err != nil {
			return nil, err
		}
	}

	for _, info := range toUpdate {
		err = s.db.Save(&info).Error
		if err != nil {
			return nil, err
		}
	}

	if len(toDelete) > 0 {
		err = s.db.Where("id in (?)", toDelete).Delete(&ContactInfo{}).Error
		if err != nil {
			return nil, err
		}
	}

	pb = make([]*ttnpb.ContactInfo, 0, len(toUpdate)+len(toCreate))
	for _, model := range toUpdate {
		pb = append(pb, model.toPB())
	}
	for _, model := range toCreate {
		pb = append(pb, model.toPB())
	}

	return pb, nil
}
