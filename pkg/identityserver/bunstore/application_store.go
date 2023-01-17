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

	"github.com/uptrace/bun"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Application is the application model in the database.
type Application struct {
	bun.BaseModel `bun:"table:applications,alias:app"`

	Model
	SoftDelete

	ApplicationID string `bun:"application_id,notnull"`

	Name        string `bun:"name,nullzero"`
	Description string `bun:"description,nullzero"`

	Attributes []*Attribute `bun:"rel:has-many,join:type=entity_type,join:id=entity_id,polymorphic"`

	ContactInfo []*ContactInfo `bun:"rel:has-many,join:type=entity_type,join:id=entity_id,polymorphic"`

	AdministrativeContactID *string  `bun:"administrative_contact_id,type:uuid"`
	AdministrativeContact   *Account `bun:"rel:belongs-to,join:administrative_contact_id=id"`

	TechnicalContactID *string  `bun:"technical_contact_id,type:uuid"`
	TechnicalContact   *Account `bun:"rel:belongs-to,join:technical_contact_id=id"`

	NetworkServerAddress     string `bun:"network_server_address,nullzero"`
	ApplicationServerAddress string `bun:"application_server_address,nullzero"`
	JoinServerAddress        string `bun:"join_server_address,nullzero"`

	DevEUICounter int `bun:"dev_eui_counter,nullzero"`
}

// BeforeAppendModel is a hook that modifies the model on SELECT and UPDATE queries.
func (m *Application) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if err := m.Model.BeforeAppendModel(ctx, query); err != nil {
		return err
	}
	return nil
}

func applicationToPB(m *Application, fieldMask ...string) (*ttnpb.Application, error) {
	pb := &ttnpb.Application{
		Ids: &ttnpb.ApplicationIdentifiers{
			ApplicationId: m.ApplicationID,
		},

		CreatedAt: ttnpb.ProtoTimePtr(m.CreatedAt),
		UpdatedAt: ttnpb.ProtoTimePtr(m.UpdatedAt),
		DeletedAt: ttnpb.ProtoTime(m.DeletedAt),

		Name:        m.Name,
		Description: m.Description,

		NetworkServerAddress:     m.NetworkServerAddress,
		ApplicationServerAddress: m.ApplicationServerAddress,
		JoinServerAddress:        m.JoinServerAddress,

		DevEuiCounter: uint32(m.DevEUICounter),
	}

	if len(m.Attributes) > 0 {
		pb.Attributes = make(map[string]string, len(m.Attributes))
		for _, a := range m.Attributes {
			pb.Attributes[a.Key] = a.Value
		}
	}

	if len(m.ContactInfo) > 0 {
		pb.ContactInfo = make([]*ttnpb.ContactInfo, len(m.ContactInfo))
		for i, contactInfo := range m.ContactInfo {
			pb.ContactInfo[i] = contactInfoToPB(contactInfo)
		}
		sort.Sort(contactInfoProtoSlice(pb.ContactInfo))
	}

	if m.AdministrativeContact != nil {
		pb.AdministrativeContact = m.AdministrativeContact.GetOrganizationOrUserIdentifiers()
	}
	if m.TechnicalContact != nil {
		pb.TechnicalContact = m.TechnicalContact.GetOrganizationOrUserIdentifiers()
	}

	if len(fieldMask) == 0 {
		return pb, nil
	}

	res := &ttnpb.Application{}
	if err := res.SetFields(pb, fieldMask...); err != nil {
		return nil, err
	}

	// Set fields that are always present.
	res.Ids = pb.Ids
	res.CreatedAt = pb.CreatedAt
	res.UpdatedAt = pb.UpdatedAt
	res.DeletedAt = pb.DeletedAt

	return res, nil
}

type applicationStore struct {
	*baseStore
}

func newApplicationStore(baseStore *baseStore) *applicationStore {
	return &applicationStore{
		baseStore: baseStore,
	}
}

func (s *applicationStore) CreateApplication(
	ctx context.Context, pb *ttnpb.Application,
) (*ttnpb.Application, error) {
	ctx, span := tracer.Start(ctx, "CreateApplication", trace.WithAttributes(
		attribute.String("application_id", pb.GetIds().GetApplicationId()),
	))
	defer span.End()

	applicationModel := &Application{
		ApplicationID: pb.GetIds().GetApplicationId(),
		Name:          pb.Name,
		Description:   pb.Description,

		NetworkServerAddress:     pb.NetworkServerAddress,
		ApplicationServerAddress: pb.ApplicationServerAddress,
		JoinServerAddress:        pb.JoinServerAddress,

		// NOTE: The DevEUI counter is managed by the EUIStore so should not be set here.
	}

	if contact := pb.AdministrativeContact; contact != nil {
		account, err := s.getAccountModel(ctx, contact.EntityType(), contact.IDString())
		if err != nil {
			return nil, err
		}
		applicationModel.AdministrativeContact = account
		applicationModel.AdministrativeContactID = &account.ID
	}
	if contact := pb.TechnicalContact; contact != nil {
		account, err := s.getAccountModel(ctx, contact.EntityType(), contact.IDString())
		if err != nil {
			return nil, err
		}
		applicationModel.TechnicalContact = account
		applicationModel.TechnicalContactID = &account.ID
	}

	_, err := s.DB.NewInsert().
		Model(applicationModel).
		Exec(ctx)
	if err != nil {
		return nil, errors.WrapDriverError(err)
	}

	if len(pb.Attributes) > 0 {
		applicationModel.Attributes, err = s.replaceAttributes(
			ctx, nil, pb.Attributes, "application", applicationModel.ID,
		)
		if err != nil {
			return nil, err
		}
	}

	if len(pb.ContactInfo) > 0 {
		applicationModel.ContactInfo, err = s.replaceContactInfo(
			ctx, nil, pb.ContactInfo, "application", applicationModel.ID,
		)
		if err != nil {
			return nil, err
		}
	}

	pb, err = applicationToPB(applicationModel)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func (*applicationStore) selectWithFields(q *bun.SelectQuery, fieldMask store.FieldMask) (*bun.SelectQuery, error) {
	if fieldMask == nil {
		q = q.ExcludeColumn()
	} else {
		columns := []string{
			"id",
			"created_at",
			"updated_at",
			"deleted_at",
			"application_id",
		}
		for _, f := range fieldMask.TopLevel() {
			switch f {
			default:
				return nil, fmt.Errorf("unknown field %q", f)
			case "ids", "created_at", "updated_at", "deleted_at":
				// Always selected.
			case "name", "description",
				"network_server_address", "application_server_address", "join_server_address",
				"dev_eui_counter":
				// Proto name equals model name.
				columns = append(columns, f)
			case "attributes":
				q = q.Relation("Attributes")
			case "contact_info":
				q = q.Relation("ContactInfo")
			case "administrative_contact":
				q = q.Relation("AdministrativeContact", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Column("uid", "account_type")
				})
			case "technical_contact":
				q = q.Relation("TechnicalContact", func(q *bun.SelectQuery) *bun.SelectQuery {
					return q.Column("uid", "account_type")
				})
			}
		}
		q = q.Column(columns...)
	}
	return q, nil
}

func (s *applicationStore) CountApplications(ctx context.Context) (uint64, error) {
	selectQuery := s.newSelectModel(ctx, &Application{})

	// Count the total number of results.
	count, err := selectQuery.Count(ctx)
	if err != nil {
		return 0, errors.WrapDriverError(err)
	}

	return uint64(count), nil
}

func (s *applicationStore) listApplicationsBy(
	ctx context.Context,
	by func(*bun.SelectQuery) *bun.SelectQuery,
	fieldMask store.FieldMask,
) ([]*ttnpb.Application, error) {
	models := []*Application{}
	selectQuery := newSelectModels(ctx, s.DB, &models).Apply(by)

	// Count the total number of results.
	count, err := selectQuery.Count(ctx)
	if err != nil {
		return nil, errors.WrapDriverError(err)
	}
	store.SetTotal(ctx, uint64(count))

	// Apply ordering, paging and field masking.
	selectQuery = selectQuery.
		Apply(selectWithOrderFromContext(ctx, "application_id", map[string]string{
			"application_id": "application_id",
			"name":           "name",
			"created_at":     "created_at",
		})).
		Apply(selectWithLimitAndOffsetFromContext(ctx))

	selectQuery, err = s.selectWithFields(selectQuery, fieldMask)
	if err != nil {
		return nil, err
	}

	// Scan the results.
	err = selectQuery.Scan(ctx)
	if err != nil {
		return nil, errors.WrapDriverError(err)
	}

	// Convert the results to protobuf.
	pbs := make([]*ttnpb.Application, len(models))
	for i, model := range models {
		pb, err := applicationToPB(model, fieldMask...)
		if err != nil {
			return nil, err
		}
		pbs[i] = pb
	}

	return pbs, nil
}

func (*applicationStore) selectWithID(
	_ context.Context, ids ...string,
) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		switch len(ids) {
		case 0:
			return q
		case 1:
			return q.Where("?TableAlias.application_id = ?", ids[0])
		default:
			return q.Where("?TableAlias.application_id IN (?)", bun.In(ids))
		}
	}
}

func (s *applicationStore) FindApplications(
	ctx context.Context, ids []*ttnpb.ApplicationIdentifiers, fieldMask store.FieldMask,
) ([]*ttnpb.Application, error) {
	ctx, span := tracer.Start(ctx, "FindApplications", trace.WithAttributes(
		attribute.StringSlice("application_ids", idStrings(ids...)),
	))
	defer span.End()

	return s.listApplicationsBy(ctx, s.selectWithID(ctx, idStrings(ids...)...), fieldMask)
}

func (s *applicationStore) getApplicationModelBy(
	ctx context.Context,
	by func(*bun.SelectQuery) *bun.SelectQuery,
	fieldMask store.FieldMask,
) (*Application, error) {
	model := &Application{}
	selectQuery := s.newSelectModel(ctx, model).Apply(by)

	selectQuery, err := s.selectWithFields(selectQuery, fieldMask)
	if err != nil {
		return nil, err
	}

	if err := selectQuery.Scan(ctx); err != nil {
		return nil, errors.WrapDriverError(err)
	}

	return model, nil
}

func (s *applicationStore) GetApplication(
	ctx context.Context, id *ttnpb.ApplicationIdentifiers, fieldMask store.FieldMask,
) (*ttnpb.Application, error) {
	ctx, span := tracer.Start(ctx, "GetApplication", trace.WithAttributes(
		attribute.String("application_id", id.GetApplicationId()),
	))
	defer span.End()

	model, err := s.getApplicationModelBy(
		ctx, s.selectWithID(ctx, id.GetApplicationId()), fieldMask,
	)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, store.ErrApplicationNotFound.WithAttributes(
				"application_id", id.GetApplicationId(),
			)
		}
		return nil, err
	}
	pb, err := applicationToPB(model, fieldMask...)
	if err != nil {
		return nil, err
	}
	return pb, nil
}

func (s *applicationStore) updateApplicationModel( //nolint:gocyclo
	ctx context.Context, model *Application, pb *ttnpb.Application, fieldMask store.FieldMask,
) (err error) {
	columns := store.FieldMask{"updated_at"}

	for _, field := range fieldMask {
		switch field {
		case "name":
			model.Name = pb.Name
			columns = append(columns, "name")

		case "description":
			model.Description = pb.Description
			columns = append(columns, "description")

		case "attributes":
			model.Attributes, err = s.replaceAttributes(
				ctx, model.Attributes, pb.Attributes, "application", model.ID,
			)
			if err != nil {
				return err
			}

		case "contact_info":
			model.ContactInfo, err = s.replaceContactInfo(
				ctx, model.ContactInfo, pb.ContactInfo, "application", model.ID,
			)
			if err != nil {
				return err
			}

		case "administrative_contact":
			if contact := pb.AdministrativeContact; contact != nil {
				account, err := s.getAccountModel(ctx, contact.EntityType(), contact.IDString())
				if err != nil {
					return err
				}
				model.AdministrativeContact = account
				model.AdministrativeContactID = &account.ID
			} else {
				model.AdministrativeContact = nil
				model.AdministrativeContactID = nil
			}
			columns = append(columns, "administrative_contact_id")

		case "technical_contact":
			if contact := pb.TechnicalContact; contact != nil {
				account, err := s.getAccountModel(ctx, contact.EntityType(), contact.IDString())
				if err != nil {
					return err
				}
				model.TechnicalContact = account
				model.TechnicalContactID = &account.ID
			} else {
				model.TechnicalContact = nil
				model.TechnicalContactID = nil
			}
			columns = append(columns, "technical_contact_id")

		case "network_server_address":
			model.NetworkServerAddress = pb.NetworkServerAddress
			columns = append(columns, "network_server_address")

		case "application_server_address":
			model.ApplicationServerAddress = pb.ApplicationServerAddress
			columns = append(columns, "application_server_address")

		case "join_server_address":
			model.JoinServerAddress = pb.JoinServerAddress
			columns = append(columns, "join_server_address")

		case "dev_eui_counter":
			// NOTE: The DevEUI counter is managed by the EUIStore so should not be updated here.
		}
	}

	_, err = s.DB.NewUpdate().
		Model(model).
		WherePK().
		Column(columns...).
		Exec(ctx)
	if err != nil {
		return errors.WrapDriverError(err)
	}

	return nil
}

func (s *applicationStore) UpdateApplication(
	ctx context.Context, pb *ttnpb.Application, fieldMask store.FieldMask,
) (*ttnpb.Application, error) {
	ctx, span := tracer.Start(ctx, "UpdateApplication", trace.WithAttributes(
		attribute.String("application_id", pb.GetIds().GetApplicationId()),
	))
	defer span.End()

	model, err := s.getApplicationModelBy(
		ctx, s.selectWithID(ctx, pb.GetIds().GetApplicationId()), fieldMask,
	)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, store.ErrApplicationNotFound.WithAttributes(
				"application_id", pb.GetIds().GetApplicationId(),
			)
		}
		return nil, err
	}

	if err = s.updateApplicationModel(ctx, model, pb, fieldMask); err != nil {
		return nil, err
	}

	// Convert the result to protobuf.
	updatedPB, err := applicationToPB(model, fieldMask...)
	if err != nil {
		return nil, err
	}

	return updatedPB, nil
}

func (s *applicationStore) DeleteApplication(ctx context.Context, id *ttnpb.ApplicationIdentifiers) error {
	ctx, span := tracer.Start(ctx, "DeleteApplication", trace.WithAttributes(
		attribute.String("application_id", id.GetApplicationId()),
	))
	defer span.End()

	model, err := s.getApplicationModelBy(ctx, s.selectWithID(ctx, id.GetApplicationId()), store.FieldMask{"ids"})
	if err != nil {
		if errors.IsNotFound(err) {
			return store.ErrApplicationNotFound.WithAttributes(
				"application_id", id.GetApplicationId(),
			)
		}
		return err
	}

	_, err = s.DB.NewDelete().
		Model(model).
		WherePK().
		Exec(ctx)
	if err != nil {
		return errors.WrapDriverError(err)
	}

	return nil
}

func (s *applicationStore) RestoreApplication(ctx context.Context, id *ttnpb.ApplicationIdentifiers) error {
	ctx, span := tracer.Start(ctx, "RestoreApplication", trace.WithAttributes(
		attribute.String("application_id", id.GetApplicationId()),
	))
	defer span.End()

	model, err := s.getApplicationModelBy(
		store.WithSoftDeleted(ctx, true),
		s.selectWithID(ctx, id.GetApplicationId()),
		store.FieldMask{"ids"},
	)
	if err != nil {
		if errors.IsNotFound(err) {
			return store.ErrApplicationNotFound.WithAttributes(
				"application_id", id.GetApplicationId(),
			)
		}
		return err
	}

	_, err = s.DB.NewUpdate().
		Model(model).
		WherePK().
		WhereAllWithDeleted().
		Set("deleted_at = NULL").
		Exec(ctx)
	if err != nil {
		return errors.WrapDriverError(err)
	}

	return nil
}

func (s *applicationStore) PurgeApplication(ctx context.Context, id *ttnpb.ApplicationIdentifiers) error {
	ctx, span := tracer.Start(ctx, "PurgeApplication", trace.WithAttributes(
		attribute.String("application_id", id.GetApplicationId()),
	))
	defer span.End()

	model, err := s.getApplicationModelBy(
		store.WithSoftDeleted(ctx, false),
		s.selectWithID(ctx, id.GetApplicationId()),
		store.FieldMask{"attributes", "contact_info"},
	)
	if err != nil {
		if errors.IsNotFound(err) {
			return store.ErrApplicationNotFound.WithAttributes(
				"application_id", id.GetApplicationId(),
			)
		}
		return err
	}

	if len(model.Attributes) > 0 {
		_, err = s.replaceAttributes(ctx, model.Attributes, nil, "application", model.ID)
		if err != nil {
			return err
		}
	}

	if len(model.ContactInfo) > 0 {
		_, err = s.replaceContactInfo(ctx, model.ContactInfo, nil, "application", model.ID)
		if err != nil {
			return err
		}
	}

	_, err = s.DB.NewDelete().
		Model(model).
		WherePK().
		ForceDelete().
		Exec(ctx)
	if err != nil {
		return errors.WrapDriverError(err)
	}

	return nil
}
