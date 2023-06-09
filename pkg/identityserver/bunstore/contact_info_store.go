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
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/tracing/tracer"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	storeutil "go.thethings.network/lorawan-stack/v3/pkg/util/store"
	"google.golang.org/protobuf/types/known/timestamppb"
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
			return nil, storeutil.WrapDriverError(err)
		}
	}

	if len(toUpdate) > 0 {
		_, err := s.DB.NewUpdate().
			Model(&toUpdate).
			Column("public", "validated_at").
			Bulk().
			Exec(ctx)
		if err != nil {
			return nil, storeutil.WrapDriverError(err)
		}
	}

	if len(toCreate) > 0 {
		_, err := s.DB.NewInsert().
			Model(&toCreate).
			Exec(ctx)
		if err != nil {
			return nil, storeutil.WrapDriverError(err)
		}
	}

	sort.Sort(result)

	return result, nil
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
		return nil, storeutil.WrapDriverError(err)
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
	ctx, span := tracer.StartFromContext(ctx, "GetContactInfo", trace.WithAttributes(
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
	ctx, span := tracer.StartFromContext(ctx, "SetContactInfo", trace.WithAttributes(
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

func (s *contactInfoStore) ValidateContactInfo(ctx context.Context, pb *ttnpb.ContactInfoValidation) error {
	ctx, span := tracer.StartFromContext(ctx, "ValidateContactInfo", trace.WithAttributes(
		attribute.String("entity_type", pb.GetEntity().EntityType()),
		attribute.String("entity_id", pb.GetEntity().IDString()),
	))
	defer span.End()
	if len(pb.GetContactInfo()) != 1 {
		return store.ErrValidationWithoutContactInfo.New()
	}
	contInfo := pb.GetContactInfo()[0]

	entityType, entityUUID, err := s.getEntity(ctx, pb.GetEntity())
	if err != nil {
		return err
	}

	models, err := s.getContactInfoModelsBy(ctx, func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.
			Where("entity_type = ? AND entity_id = ?", entityType, entityUUID).
			Where("contact_method = ? AND LOWER(value) = LOWER(?)", contInfo.GetContactMethod(), contInfo.GetValue())
	})
	if err != nil {
		return err
	}
	if len(models) != 1 {
		return store.ErrContactInfoNotFound.New()
	}
	model := models[0]

	_, err = s.DB.NewUpdate().
		Model(model).
		WherePK().
		Set("validated_at = ?", pb.ExpiresAt.AsTime()).
		Exec(ctx)
	if err != nil {
		return storeutil.WrapDriverError(err)
	}

	if pb.Entity.EntityType() == "user" && contInfo.ContactMethod == ttnpb.ContactMethod_CONTACT_METHOD_EMAIL {
		usrModel, err := s.getUserModelBy(ctx, func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where("lower(primary_email_address) = lower(?)", contInfo.Value)
		}, nil)
		if err != nil {
			return err
		}
		_, err = s.DB.NewUpdate().
			Model(usrModel).
			WherePK().
			Set("primary_email_address_validated_at = ?", now()).
			Exec(ctx)
		if err != nil {
			return storeutil.WrapDriverError(err)
		}
	}
	return nil
}

func (s *contactInfoStore) DeleteEntityContactInfo(ctx context.Context, entityID ttnpb.IDStringer) error {
	ctx, span := tracer.StartFromContext(ctx, "DeleteEntityContactInfo", trace.WithAttributes(
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
		return storeutil.WrapDriverError(err)
	}

	return nil
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
	return m.Model.BeforeAppendModel(ctx, query)
}

func validationToPB(m *ContactInfoValidation) *ttnpb.ContactInfoValidation {
	val := &ttnpb.ContactInfoValidation{
		Id:        m.Reference,
		Token:     m.Token,
		CreatedAt: ttnpb.ProtoTime(&m.CreatedAt),
		ExpiresAt: ttnpb.ProtoTime(m.ExpiresAt),
		ContactInfo: []*ttnpb.ContactInfo{{
			ContactMethod: ttnpb.ContactMethod(m.ContactMethod),
			Value:         m.Value,
		}},
	}

	switch m.EntityType {
	case store.EntityApplication:
		val.Entity = (&ttnpb.ApplicationIdentifiers{ApplicationId: m.EntityID}).GetEntityIdentifiers()
	case store.EntityClient:
		val.Entity = (&ttnpb.ClientIdentifiers{ClientId: m.EntityID}).GetEntityIdentifiers()
	case store.EntityGateway:
		val.Entity = (&ttnpb.GatewayIdentifiers{GatewayId: m.EntityID}).GetEntityIdentifiers()
	case store.EntityOrganization:
		val.Entity = (&ttnpb.OrganizationIdentifiers{OrganizationId: m.EntityID}).GetEntityIdentifiers()
	case store.EntityUser:
		val.Entity = (&ttnpb.UserIdentifiers{UserId: m.EntityID}).GetEntityIdentifiers()
	}
	return val
}

func (s *contactInfoStore) getContactInfoValidationModelBy(
	ctx context.Context,
	by func(*bun.SelectQuery) *bun.SelectQuery,
) (*ContactInfoValidation, error) {
	model := &ContactInfoValidation{}
	selectQuery := newSelectModel(ctx, s.DB, model).Apply(by)
	err := selectQuery.Scan(ctx)
	if err != nil {
		return nil, storeutil.WrapDriverError(err)
	}
	return model, nil
}

func (s *contactInfoStore) CreateValidation(
	ctx context.Context, pb *ttnpb.ContactInfoValidation,
) (*ttnpb.ContactInfoValidation, error) {
	ctx, span := tracer.StartFromContext(ctx, "CreateValidation", trace.WithAttributes(
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
		return nil, storeutil.WrapDriverError(err)
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
		return nil, storeutil.WrapDriverError(err)
	}

	return &ttnpb.ContactInfoValidation{
		Id:          model.Reference,
		Token:       model.Token,
		Entity:      pb.Entity,
		ContactInfo: pb.ContactInfo,
		CreatedAt:   timestamppb.New(model.CreatedAt),
		ExpiresAt:   ttnpb.ProtoTime(model.ExpiresAt),
	}, nil
}

func (s *contactInfoStore) GetValidation(
	ctx context.Context, pb *ttnpb.ContactInfoValidation,
) (*ttnpb.ContactInfoValidation, error) {
	ctx, span := tracer.StartFromContext(ctx, "GetValidation", trace.WithAttributes(
		attribute.String("entity_type", pb.GetEntity().EntityType()),
		attribute.String("entity_id", pb.GetEntity().IDString()),
	))
	defer span.End()

	model, err := s.getContactInfoValidationModelBy(ctx, func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("reference = ? AND token = ?", pb.Id, pb.Token)
	})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, store.ErrValidationTokenNotFound.WithAttributes(
				"validation_id", pb.Id,
			)
		}
		return nil, err
	}

	if model.Used {
		return nil, store.ErrValidationTokenAlreadyUsed.WithAttributes(
			"validation_id", pb.Id,
		)
	}

	if model.ExpiresAt != nil && model.ExpiresAt.Before(s.now()) {
		return nil, store.ErrValidationTokenExpired.WithAttributes(
			"validation_id", pb.Id,
		)
	}

	friendlyID, err := s.getEntityID(ctx, model.EntityType, model.EntityID)
	if err != nil {
		return nil, err
	}

	model.EntityID = friendlyID
	return validationToPB(model), nil
}

func (s *contactInfoStore) ExpireValidation(ctx context.Context, pb *ttnpb.ContactInfoValidation) error {
	ctx, span := tracer.StartFromContext(ctx, "ExpireValidation", trace.WithAttributes(
		attribute.String("entity_type", pb.GetEntity().EntityType()),
		attribute.String("entity_id", pb.GetEntity().IDString()),
	))
	defer span.End()
	model, err := s.getContactInfoValidationModelBy(ctx, func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("reference = ? AND token = ?", pb.Id, pb.Token)
	})
	if err != nil {
		return err
	}

	_, err = s.DB.NewUpdate().
		Model(model).
		WherePK().
		Set("expires_at = ?, used = true", now()).
		Exec(ctx)
	if err != nil {
		return storeutil.WrapDriverError(err)
	}
	return nil
}
