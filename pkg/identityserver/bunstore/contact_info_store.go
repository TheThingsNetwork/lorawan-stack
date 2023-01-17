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
	"sort"
	"time"

	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// ContactInfo is the contact info model in the database.
type ContactInfo struct {
	bun.BaseModel `bun:"table:contact_infos,alias:ci"`

	UUID

	// EntityType is "application", "client", "end_device", "gateway", "organization" or "user".
	EntityType string `bun:"entity_type,notnull"`
	// EntityID is Application.ID, Client.ID, EndDevice.ID, Gateway.ID, Organization.ID or User.ID.
	EntityID string `bun:"entity_id,notnull"`

	ContactType   int    `bun:"contact_type,notnull"`
	ContactMethod int    `bun:"contact_method,notnull"`
	Value         string `bun:"value,notnull"`

	Public bool `bun:"public"`

	ValidatedAt *time.Time `bun:"validated_at"`
}

func (ContactInfo) _isModel() {} // It doesn't embed Model, but it's still a model.

func contactInfoFromPB(pb *ttnpb.ContactInfo, entityType, entityID string) *ContactInfo {
	return &ContactInfo{
		EntityType:    entityType,
		EntityID:      entityID,
		ContactType:   int(pb.ContactType),
		ContactMethod: int(pb.ContactMethod),
		Value:         pb.Value,
		Public:        pb.Public,
		ValidatedAt:   cleanTimePtr(ttnpb.StdTime(pb.ValidatedAt)),
	}
}

func contactInfoToPB(m *ContactInfo) *ttnpb.ContactInfo {
	return &ttnpb.ContactInfo{
		ContactType:   ttnpb.ContactType(m.ContactType),
		ContactMethod: ttnpb.ContactMethod(m.ContactMethod),
		Value:         m.Value,
		Public:        m.Public,
		ValidatedAt:   ttnpb.ProtoTime(m.ValidatedAt),
	}
}

// ContactInfoSlice is a slice of ContactInfo.
type ContactInfoSlice []*ContactInfo

func (a ContactInfoSlice) Len() int      { return len(a) }
func (a ContactInfoSlice) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ContactInfoSlice) Less(i, j int) bool {
	if a[i].Value < a[j].Value {
		return true
	}
	return a[i].ContactType < a[j].ContactType
}

type contactInfoProtoSlice []*ttnpb.ContactInfo

func (a contactInfoProtoSlice) Len() int      { return len(a) }
func (a contactInfoProtoSlice) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a contactInfoProtoSlice) Less(i, j int) bool {
	if a[i].Value < a[j].Value {
		return true
	}
	return a[i].ContactType < a[j].ContactType
}

func contactInfoSliceFromPB(pbs []*ttnpb.ContactInfo, entityType, entityID string) []*ContactInfo {
	out := make([]*ContactInfo, len(pbs))
	for i, pb := range pbs {
		out[i] = contactInfoFromPB(pb, entityType, entityID)
	}
	return out
}

func contactInfoMap(models []*ContactInfo) map[string]*ContactInfo {
	m := make(map[string]*ContactInfo, len(models))
	for _, model := range models {
		key := fmt.Sprintf("%d-%d-%s", model.ContactMethod, model.ContactType, model.Value)
		m[key] = model
	}
	return m
}

func (s *baseStore) replaceContactInfo(
	ctx context.Context, current []*ContactInfo, desired []*ttnpb.ContactInfo, entityType, entityID string,
) ([]*ContactInfo, error) {
	var (
		oldMap   = contactInfoMap(current)
		newMap   = contactInfoMap(contactInfoSliceFromPB(desired, entityType, entityID))
		toCreate = make([]*ContactInfo, 0, len(newMap))
		toUpdate = make([]*ContactInfo, 0, len(newMap))
		toDelete = make([]*ContactInfo, 0, len(oldMap))
		result   = make(ContactInfoSlice, 0, len(newMap))
	)

	for k, v := range newMap {
		// Ignore contact info that has not been updated.
		if current, ok := oldMap[k]; ok {
			delete(oldMap, k) // Don't need to delete this one.
			delete(newMap, k) // Don't need to create this one.
			if current.Public == v.Public && equalTime(current.ValidatedAt, v.ValidatedAt) {
				result = append(result, v)
				continue // Don't need to update this one.
			}
			v.ID = current.ID
			toUpdate = append(toUpdate, v)
			result = append(result, v)
			continue
		}
		toCreate = append(toCreate, v)
		result = append(result, v)
	}
	for _, v := range oldMap {
		toDelete = append(toDelete, v)
	}

	if len(toDelete) > 0 {
		_, err := s.DB.NewDelete().
			Model(&toDelete).
			WherePK().
			Exec(ctx)
		if err != nil {
			return nil, errors.WrapDriverError(err)
		}
	}

	if len(toUpdate) > 0 {
		_, err := s.DB.NewUpdate().
			Model(&toUpdate).
			Column("public", "validated_at").
			Bulk().
			Exec(ctx)
		if err != nil {
			return nil, errors.WrapDriverError(err)
		}
	}

	if len(toCreate) > 0 {
		_, err := s.DB.NewInsert().
			Model(&toCreate).
			Exec(ctx)
		if err != nil {
			return nil, errors.WrapDriverError(err)
		}
	}

	sort.Sort(result)

	return result, nil
}

// ContactInfoValidation is the contact info validation model in the database.
type ContactInfoValidation struct {
	bun.BaseModel `bun:"table:contact_info_validations,alias:civ"`

	Model

	Reference string `bun:"reference,nullzero"`
	Token     string `bun:"token,nullzero"`

	// EntityType is "application", "client", "gateway", "organization" or "user".
	EntityType string `bun:"entity_type,notnull"`
	// EntityID is Application.ID, Client.ID, Gateway.ID, Organization.ID or User.ID.
	EntityID string `bun:"entity_id,notnull"`

	ContactMethod int    `bun:"contact_method,notnull"`
	Value         string `bun:"value,nullzero"`

	Used bool `bun:"used,nullzero"`

	ExpiresAt *time.Time `bun:"expires_at"`
}

// BeforeAppendModel is a hook that modifies the model on SELECT and UPDATE queries.
func (m *ContactInfoValidation) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if err := m.Model.BeforeAppendModel(ctx, query); err != nil {
		return err
	}
	return nil
}

type contactInfoStore struct {
	*entityStore
}

func newContactInfoStore(baseStore *baseStore) *contactInfoStore {
	return &contactInfoStore{
		entityStore: newEntityStore(baseStore),
	}
}

func (s *contactInfoStore) getContactInfoModelsBy(
	ctx context.Context,
	by func(*bun.SelectQuery) *bun.SelectQuery,
) ([]*ContactInfo, error) {
	models := []*ContactInfo{}
	selectQuery := newSelectModels(ctx, s.DB, &models).
		Apply(by)

	err := selectQuery.Scan(ctx)
	if err != nil {
		return nil, errors.WrapDriverError(err)
	}
	return models, nil
}

func (*contactInfoStore) selectWithEntityIDs(
	_ context.Context, entityType, entityUUID string,
) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("entity_type = ? AND entity_id = ?", entityType, entityUUID)
	}
}

func (s *contactInfoStore) GetContactInfo(
	ctx context.Context, entityID ttnpb.IDStringer,
) ([]*ttnpb.ContactInfo, error) {
	ctx, span := tracer.Start(ctx, "GetContactInfo", trace.WithAttributes(
		attribute.String("entity_type", entityID.EntityType()),
		attribute.String("entity_id", entityID.IDString()),
	))
	defer span.End()

	entityType, entityUUID, err := s.getEntity(ctx, entityID)
	if err != nil {
		return nil, err
	}

	models, err := s.getContactInfoModelsBy(ctx, s.selectWithEntityIDs(ctx, entityType, entityUUID))
	if err != nil {
		return nil, err
	}

	pbs := make([]*ttnpb.ContactInfo, len(models))
	for i, contactInfo := range models {
		pbs[i] = contactInfoToPB(contactInfo)
	}
	sort.Sort(contactInfoProtoSlice(pbs))

	return pbs, nil
}

func (s *contactInfoStore) SetContactInfo(
	ctx context.Context, entityID ttnpb.IDStringer, pbs []*ttnpb.ContactInfo,
) ([]*ttnpb.ContactInfo, error) {
	ctx, span := tracer.Start(ctx, "SetContactInfo", trace.WithAttributes(
		attribute.String("entity_type", entityID.EntityType()),
		attribute.String("entity_id", entityID.IDString()),
	))
	defer span.End()

	entityType, entityUUID, err := s.getEntity(ctx, entityID)
	if err != nil {
		return nil, err
	}

	models, err := s.getContactInfoModelsBy(ctx, s.selectWithEntityIDs(ctx, entityType, entityUUID))
	if err != nil {
		return nil, err
	}

	models, err = s.replaceContactInfo(ctx, models, pbs, entityType, entityUUID)
	if err != nil {
		return nil, err
	}

	pbs = make([]*ttnpb.ContactInfo, len(models))
	for i, contactInfo := range models {
		pbs[i] = contactInfoToPB(contactInfo)
	}
	sort.Sort(contactInfoProtoSlice(pbs))

	return pbs, nil
}

func (s *contactInfoStore) CreateValidation(
	ctx context.Context, pb *ttnpb.ContactInfoValidation,
) (*ttnpb.ContactInfoValidation, error) {
	ctx, span := tracer.Start(ctx, "CreateValidation", trace.WithAttributes(
		attribute.String("entity_type", pb.GetEntity().EntityType()),
		attribute.String("entity_id", pb.GetEntity().IDString()),
	))
	defer span.End()

	var (
		contactMethod ttnpb.ContactMethod
		value         string
	)
	for i, info := range pb.ContactInfo {
		if i == 0 {
			contactMethod = info.ContactMethod
			value = info.Value
			continue
		}
		if info.ContactMethod != contactMethod || info.Value != value {
			panic("inconsistent contact info in validation")
		}
	}

	entityType, entityUUID, err := s.getEntity(ctx, pb.GetEntity())
	if err != nil {
		return nil, err
	}

	n, err := s.newSelectModel(ctx, &ContactInfoValidation{}).
		Where("entity_type = ? AND entity_id = ?", entityType, entityUUID).
		Where("contact_method = ? AND LOWER(value) = LOWER(?)", contactMethod, value).
		Where("expires_at IS NULL OR expires_at > NOW()").
		Count(ctx)
	if err != nil {
		return nil, errors.WrapDriverError(err)
	}
	if n > 0 {
		return nil, store.ErrValidationAlreadySent.New()
	}

	model := &ContactInfoValidation{
		Reference:     pb.Id,
		Token:         pb.Token,
		EntityType:    entityType,
		EntityID:      entityUUID,
		ContactMethod: int(contactMethod),
		Value:         value,
		ExpiresAt:     cleanTimePtr(ttnpb.StdTime(pb.ExpiresAt)),
	}

	_, err = s.DB.NewInsert().
		Model(model).
		Exec(ctx)
	if err != nil {
		return nil, errors.WrapDriverError(err)
	}

	return &ttnpb.ContactInfoValidation{
		Id:          model.Reference,
		Token:       model.Token,
		Entity:      pb.Entity,
		ContactInfo: pb.ContactInfo,
		CreatedAt:   ttnpb.ProtoTimePtr(model.CreatedAt),
		ExpiresAt:   ttnpb.ProtoTime(model.ExpiresAt),
	}, nil
}

func (s *contactInfoStore) Validate(ctx context.Context, validation *ttnpb.ContactInfoValidation) error {
	ctx, span := tracer.Start(ctx, "Validate")
	defer span.End()

	// TODO: Refactor store interface to split this ito a separate methods.
	// (https://github.com/TheThingsNetwork/lorawan-stack/issues/5587)

	model := &ContactInfoValidation{}

	err := s.newSelectModel(ctx, model).
		Where("reference = ? AND token = ?", validation.Id, validation.Token).
		Scan(ctx)
	if err != nil {
		err = errors.WrapDriverError(err)
		if errors.IsNotFound(err) {
			return store.ErrValidationTokenNotFound.WithAttributes(
				"validation_id", validation.Id,
			)
		}
		return err
	}

	if model.Used {
		return store.ErrValidationTokenAlreadyUsed.WithAttributes(
			"validation_id", validation.Id,
		)
	}

	if model.ExpiresAt != nil && model.ExpiresAt.Before(s.now()) {
		return store.ErrValidationTokenExpired.WithAttributes(
			"validation_id", validation.Id,
		)
	}

	now := s.now()

	_, err = s.DB.NewUpdate().
		Model(&ContactInfo{}).
		Where("entity_type = ? AND entity_id = ?", model.EntityType, model.EntityID).
		Where("contact_method = ? AND LOWER(value) = LOWER(?)", model.ContactMethod, model.Value).
		Set("validated_at = ?", now).
		Exec(ctx)
	if err != nil {
		return errors.WrapDriverError(err)
	}

	if model.EntityType == "user" &&
		ttnpb.ContactMethod(model.ContactMethod) == ttnpb.ContactMethod_CONTACT_METHOD_EMAIL {
		_, err = s.DB.NewUpdate().
			Model(&User{}).
			Where("lower(primary_email_address) = lower(?)", model.Value).
			Set("primary_email_address_validated_at = ?", now).
			Exec(ctx)
		if err != nil {
			return errors.WrapDriverError(err)
		}
	}

	_, err = s.DB.NewUpdate().
		Model(model).
		WherePK().
		Set("expires_at = ?, used = true", now).
		Exec(ctx)
	if err != nil {
		return errors.WrapDriverError(err)
	}

	return nil
}

func (s *contactInfoStore) DeleteEntityContactInfo(ctx context.Context, entityID ttnpb.IDStringer) error {
	ctx, span := tracer.Start(ctx, "DeleteEntityContactInfo", trace.WithAttributes(
		attribute.String("entity_type", entityID.EntityType()),
		attribute.String("entity_id", entityID.IDString()),
	))
	defer span.End()

	entityType, entityUUID, err := s.getEntity(store.WithSoftDeleted(ctx, false), entityID)
	if err != nil {
		return err
	}

	_, err = s.DB.NewDelete().
		Model(&ContactInfo{}).
		Where("entity_type = ? AND entity_id = ?", entityType, entityUUID).
		Exec(ctx)
	if err != nil {
		return errors.WrapDriverError(err)
	}

	return nil
}
