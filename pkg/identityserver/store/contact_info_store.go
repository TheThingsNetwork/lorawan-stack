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
	"time"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/errors"
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
	defer trace.StartRegion(ctx, "get contact info").End()
	entity, err := findEntity(ctx, s.db, entityID, "id")
	if err != nil {
		return nil, err
	}
	var models []ContactInfo
	err = s.db.Where(ContactInfo{
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
	defer trace.StartRegion(ctx, "update contact info").End()
	entity, err := findEntity(ctx, s.db, entityID, "id")
	if err != nil {
		return nil, err
	}
	entityType, entityUUID := entityTypeForID(entityID), entity.PrimaryKey()

	var existing []ContactInfo
	err = s.db.Where(ContactInfo{
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
		if err = s.db.Save(&info).Error; err != nil {
			return nil, err
		}
	}

	for _, info := range toUpdate {
		if err = s.db.Save(&info).Error; err != nil {
			return nil, err
		}
	}

	if len(toDelete) > 0 {
		if err = s.db.Where("id in (?)", toDelete).Delete(&ContactInfo{}).Error; err != nil {
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

func (s *contactInfoStore) CreateValidation(ctx context.Context, validation *ttnpb.ContactInfoValidation) (*ttnpb.ContactInfoValidation, error) {
	defer trace.StartRegion(ctx, "create contact info validation").End()
	var (
		contactMethod ttnpb.ContactMethod
		value         string
	)
	for i, info := range validation.ContactInfo {
		if i == 0 {
			contactMethod = info.ContactMethod
			value = info.Value
			continue
		}
		if info.ContactMethod != contactMethod || info.Value != value {
			panic("inconsistent contact info in validation")
		}
	}
	entity, err := findEntity(ctx, s.db, validation.Entity, "id")
	if err != nil {
		return nil, err
	}

	var model ContactInfoValidation
	model.fromPB(validation)
	model.EntityType, model.EntityID = entityTypeForID(validation.Entity), entity.PrimaryKey()
	model.ContactMethod = int(contactMethod)
	model.Value = value

	model.SetContext(ctx)
	query := s.db.Create(&model)
	if query.Error != nil {
		return nil, query.Error
	}

	pb := model.toPB()
	pb.Entity, pb.ContactInfo = validation.Entity, validation.ContactInfo

	return pb, nil
}

var (
	errValidationTokenNotFound = errors.DefineNotFound("validation_token", "validation token not found")
	errValidationTokenExpired  = errors.DefineNotFound("validation_token_expired", "validation token expired")
)

func (s *contactInfoStore) Validate(ctx context.Context, validation *ttnpb.ContactInfoValidation) error {
	defer trace.StartRegion(ctx, "validate contact info").End()
	now := cleanTime(time.Now())

	var model ContactInfoValidation
	err := s.db.Scopes(withContext(ctx)).Where(ContactInfoValidation{
		Reference: validation.ID,
		Token:     validation.Token,
	}).Find(&model).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errValidationTokenNotFound
		}
		return err
	}

	if model.ExpiresAt.Before(time.Now()) {
		return errValidationTokenExpired
	}

	err = s.db.Model(ContactInfo{}).Scopes(withContext(ctx)).Where(ContactInfo{
		EntityID:      model.EntityID,
		EntityType:    model.EntityType,
		ContactMethod: model.ContactMethod,
		Value:         model.Value,
	}).Update(ContactInfo{
		ValidatedAt: &now,
	}).Error
	if err != nil {
		return err
	}

	if model.EntityType == "user" && model.ContactMethod == int(ttnpb.CONTACT_METHOD_EMAIL) {
		err = s.db.Model(User{}).Scopes(withContext(ctx)).Where(User{
			Model: Model{ID: model.EntityID},
		}).Where(User{
			PrimaryEmailAddress: model.Value,
		}).Update(User{
			PrimaryEmailAddressValidatedAt: &now,
		}).Error
		if err != nil {
			return err
		}
	}

	return s.db.Delete(&model).Error
}
