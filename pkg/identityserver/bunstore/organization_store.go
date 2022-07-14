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
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Organization is the organization model in the database.
type Organization struct {
	bun.BaseModel `bun:"table:organizations,alias:org,select:organization_accounts"`

	Model
	SoftDelete

	Account EmbeddedAccount `bun:"embed:account_"`

	Name        string `bun:"name,nullzero"`
	Description string `bun:"description,nullzero"`

	Attributes []*Attribute `bun:"rel:has-many,join:type=entity_type,join:id=entity_id,polymorphic"`

	ContactInfo []*ContactInfo `bun:"rel:has-many,join:type=entity_type,join:id=entity_id,polymorphic"`

	AdministrativeContactID *string  `bun:"administrative_contact_id,type:uuid"`
	AdministrativeContact   *Account `bun:"rel:belongs-to,join:administrative_contact_id=id"`

	TechnicalContactID *string  `bun:"technical_contact_id,type:uuid"`
	TechnicalContact   *Account `bun:"rel:belongs-to,join:technical_contact_id=id"`
}

// BeforeAppendModel is a hook that modifies the model on SELECT and UPDATE queries.
func (m *Organization) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if err := m.Model.BeforeAppendModel(ctx, query); err != nil {
		return err
	}
	return nil
}

func organizationToPB(m *Organization, fieldMask ...string) (*ttnpb.Organization, error) {
	pb := &ttnpb.Organization{
		Ids: &ttnpb.OrganizationIdentifiers{
			OrganizationId: m.Account.UID,
		},

		CreatedAt: ttnpb.ProtoTimePtr(m.CreatedAt),
		UpdatedAt: ttnpb.ProtoTimePtr(m.UpdatedAt),
		DeletedAt: ttnpb.ProtoTime(m.DeletedAt),

		Name:        m.Name,
		Description: m.Description,
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

	res := &ttnpb.Organization{}
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

type organizationStore struct {
	*baseStore
}

func newOrganizationStore(baseStore *baseStore) *organizationStore {
	return &organizationStore{
		baseStore: baseStore,
	}
}

func (s *organizationStore) CreateOrganization(
	ctx context.Context, pb *ttnpb.Organization,
) (*ttnpb.Organization, error) {
	ctx, span := tracer.Start(ctx, "CreateOrganization", trace.WithAttributes(
		attribute.String("organization_id", pb.GetIds().GetOrganizationId()),
	))
	defer span.End()

	organizationModel := &Organization{
		Account: EmbeddedAccount{
			UID: pb.GetIds().GetOrganizationId(),
		},
		Name:        pb.Name,
		Description: pb.Description,
	}

	if contact := pb.AdministrativeContact; contact != nil {
		account, err := s.getAccountModel(ctx, contact.EntityType(), contact.IDString())
		if err != nil {
			return nil, err
		}
		organizationModel.AdministrativeContact = account
		organizationModel.AdministrativeContactID = &account.ID
	}
	if contact := pb.TechnicalContact; contact != nil {
		account, err := s.getAccountModel(ctx, contact.EntityType(), contact.IDString())
		if err != nil {
			return nil, err
		}
		organizationModel.TechnicalContact = account
		organizationModel.TechnicalContactID = &account.ID
	}

	// Run organization+account creation in a transaction if we're not already in one.
	err := s.transact(ctx, func(ctx context.Context, tx bun.IDB) error {
		_, err := tx.NewInsert().
			Model(organizationModel).
			Exec(ctx)
		if err != nil {
			return err
		}

		accountModel := &Account{
			UID:         pb.GetIds().GetOrganizationId(),
			AccountType: "organization",
			AccountID:   organizationModel.ID,
		}
		_, err = tx.NewInsert().
			Model(accountModel).
			Exec(ctx)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, wrapDriverError(err)
	}

	if len(pb.Attributes) > 0 {
		organizationModel.Attributes, err = s.replaceAttributes(
			ctx, nil, pb.Attributes, "organization", organizationModel.ID,
		)
		if err != nil {
			return nil, err
		}
	}

	if len(pb.ContactInfo) > 0 {
		organizationModel.ContactInfo, err = s.replaceContactInfo(
			ctx, nil, pb.ContactInfo, "organization", organizationModel.ID,
		)
		if err != nil {
			return nil, err
		}
	}

	pb, err = organizationToPB(organizationModel)
	if err != nil {
		return nil, err
	}

	return pb, nil
}

func (*organizationStore) selectWithFields(q *bun.SelectQuery, fieldMask store.FieldMask) (*bun.SelectQuery, error) {
	if fieldMask == nil {
		q = q.ExcludeColumn()
	} else {
		columns := []string{
			"id",
			"created_at",
			"updated_at",
			"deleted_at",
			"account_uid",
		}
		for _, f := range fieldMask.TopLevel() {
			switch f {
			default:
				return nil, fmt.Errorf("unknown field %q", f)
			case "ids", "created_at", "updated_at", "deleted_at":
				// Always selected.
			case "name", "description":
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

func (s *organizationStore) listOrganizationsBy(
	ctx context.Context,
	by func(*bun.SelectQuery) *bun.SelectQuery,
	fieldMask store.FieldMask,
) ([]*ttnpb.Organization, error) {
	models := []*Organization{}
	selectQuery := s.DB.NewSelect().
		Model(&models).
		Apply(selectWithSoftDeletedFromContext(ctx)).
		Apply(by)

	// Count the total number of results.
	count, err := selectQuery.Count(ctx)
	if err != nil {
		return nil, wrapDriverError(err)
	}
	store.SetTotal(ctx, uint64(count))

	// Apply ordering, paging and field masking.
	selectQuery = selectQuery.
		Apply(selectWithOrderFromContext(ctx, "account_uid", map[string]string{
			"organization_id": "account_uid",
			"name":            "name",
			"created_at":      "created_at",
		})).
		Apply(selectWithLimitAndOffsetFromContext(ctx))

	selectQuery, err = s.selectWithFields(selectQuery, fieldMask)
	if err != nil {
		return nil, err
	}

	// Scan the results.
	err = selectQuery.Scan(ctx)
	if err != nil {
		return nil, wrapDriverError(err)
	}

	// Convert the results to protobuf.
	pbs := make([]*ttnpb.Organization, len(models))
	for i, model := range models {
		pb, err := organizationToPB(model, fieldMask...)
		if err != nil {
			return nil, err
		}
		pbs[i] = pb
	}

	return pbs, nil
}

func (s *organizationStore) selectWithID(ctx context.Context, ids ...string) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		q = q.Apply(selectWithContext(ctx))
		switch len(ids) {
		case 0:
			return q
		case 1:
			return q.Where("?TableAlias.account_uid = ?", ids[0])
		default:
			return q.Where("?TableAlias.account_uid IN (?)", bun.In(ids))
		}
	}
}

func (s *organizationStore) FindOrganizations(
	ctx context.Context, ids []*ttnpb.OrganizationIdentifiers, fieldMask store.FieldMask,
) ([]*ttnpb.Organization, error) {
	ctx, span := tracer.Start(ctx, "FindOrganizations", trace.WithAttributes(
		attribute.StringSlice("organization_ids", idStrings(ids...)),
	))
	defer span.End()

	return s.listOrganizationsBy(ctx, s.selectWithID(ctx, idStrings(ids...)...), fieldMask)
}

func (s *organizationStore) getOrganizationModelBy(
	ctx context.Context,
	by func(*bun.SelectQuery) *bun.SelectQuery,
	fieldMask store.FieldMask,
) (*Organization, error) {
	model := &Organization{}
	selectQuery := s.DB.NewSelect().
		Model(model).
		Apply(selectWithSoftDeletedFromContext(ctx)).
		Apply(by)

	selectQuery, err := s.selectWithFields(selectQuery, fieldMask)
	if err != nil {
		return nil, err
	}

	if err := selectQuery.Scan(ctx); err != nil {
		return nil, wrapDriverError(err)
	}

	return model, nil
}

func (s *organizationStore) GetOrganization(
	ctx context.Context, id *ttnpb.OrganizationIdentifiers, fieldMask store.FieldMask,
) (*ttnpb.Organization, error) {
	ctx, span := tracer.Start(ctx, "GetOrganization", trace.WithAttributes(
		attribute.String("organization_id", id.GetOrganizationId()),
	))
	defer span.End()

	model, err := s.getOrganizationModelBy(
		ctx, s.selectWithID(ctx, id.GetOrganizationId()), fieldMask,
	)
	if err != nil {
		return nil, err
	}
	pb, err := organizationToPB(model, fieldMask...)
	if err != nil {
		return nil, err
	}
	return pb, nil
}

func (s *organizationStore) updateOrganizationModel( //nolint:gocyclo
	ctx context.Context, model *Organization, pb *ttnpb.Organization, fieldMask store.FieldMask,
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
				ctx, model.Attributes, pb.Attributes, "organization", model.ID,
			)
			if err != nil {
				return err
			}

		case "contact_info":
			model.ContactInfo, err = s.replaceContactInfo(
				ctx, model.ContactInfo, pb.ContactInfo, "organization", model.ID,
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
		}
	}

	_, err = s.DB.NewUpdate().
		Model(model).
		WherePK().
		Column(columns...).
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	return nil
}

func (s *organizationStore) UpdateOrganization(
	ctx context.Context, pb *ttnpb.Organization, fieldMask store.FieldMask,
) (*ttnpb.Organization, error) {
	ctx, span := tracer.Start(ctx, "UpdateOrganization", trace.WithAttributes(
		attribute.String("organization_id", pb.GetIds().GetOrganizationId()),
	))
	defer span.End()

	model, err := s.getOrganizationModelBy(
		ctx, s.selectWithID(ctx, pb.GetIds().GetOrganizationId()), fieldMask,
	)
	if err != nil {
		return nil, err
	}

	if err = s.updateOrganizationModel(ctx, model, pb, fieldMask); err != nil {
		return nil, err
	}

	// Convert the result to protobuf.
	updatedPB, err := organizationToPB(model, fieldMask...)
	if err != nil {
		return nil, err
	}

	return updatedPB, nil
}

func (s *organizationStore) DeleteOrganization(ctx context.Context, id *ttnpb.OrganizationIdentifiers) error {
	ctx, span := tracer.Start(ctx, "DeleteOrganization", trace.WithAttributes(
		attribute.String("organization_id", id.GetOrganizationId()),
	))
	defer span.End()

	model, err := s.getOrganizationModelBy(ctx, s.selectWithID(ctx, id.GetOrganizationId()), store.FieldMask{"ids"})
	if err != nil {
		return err
	}

	_, err = s.DB.NewDelete().
		Model(model).
		WherePK().
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	return nil
}

func (s *organizationStore) RestoreOrganization(ctx context.Context, id *ttnpb.OrganizationIdentifiers) error {
	ctx, span := tracer.Start(ctx, "RestoreOrganization", trace.WithAttributes(
		attribute.String("organization_id", id.GetOrganizationId()),
	))
	defer span.End()

	model, err := s.getOrganizationModelBy(
		store.WithSoftDeleted(ctx, true),
		s.selectWithID(ctx, id.GetOrganizationId()),
		store.FieldMask{"ids"},
	)
	if err != nil {
		return err
	}

	_, err = s.DB.NewUpdate().
		Model(model).
		WherePK().
		WhereAllWithDeleted().
		Set("deleted_at = NULL").
		Exec(ctx)
	if err != nil {
		return wrapDriverError(err)
	}

	return nil
}

func (s *organizationStore) PurgeOrganization(ctx context.Context, id *ttnpb.OrganizationIdentifiers) error {
	ctx, span := tracer.Start(ctx, "PurgeOrganization", trace.WithAttributes(
		attribute.String("organization_id", id.GetOrganizationId()),
	))
	defer span.End()

	model, err := s.getOrganizationModelBy(
		store.WithSoftDeleted(ctx, false),
		s.selectWithID(ctx, id.GetOrganizationId()),
		store.FieldMask{"attributes", "contact_info"},
	)
	if err != nil {
		return err
	}

	if len(model.Attributes) > 0 {
		_, err = s.replaceAttributes(ctx, model.Attributes, nil, "organization", model.ID)
		if err != nil {
			return err
		}
	}

	if len(model.ContactInfo) > 0 {
		_, err = s.replaceContactInfo(ctx, model.ContactInfo, nil, "organization", model.ID)
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
		return wrapDriverError(err)
	}

	return nil
}
