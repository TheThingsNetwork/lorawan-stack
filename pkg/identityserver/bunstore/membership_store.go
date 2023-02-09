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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/telemetry/tracing/tracer"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	storeutil "go.thethings.network/lorawan-stack/v3/pkg/util/store"
)

// Membership is the membership model in the database.
type Membership struct {
	bun.BaseModel `bun:"table:memberships,alias:mem"`

	Model

	AccountID string   `bun:"account_id,notnull"`
	Account   *Account `bun:"rel:belongs-to,join:account_id=id"`

	Rights []int `bun:"rights,array,nullzero"`

	EntityID   string `bun:"entity_id,notnull"`
	EntityType string `bun:"entity_type,notnull"`
}

// BeforeAppendModel is a hook that modifies the model on SELECT and UPDATE queries.
func (m *Membership) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	if err := m.Model.BeforeAppendModel(ctx, query); err != nil {
		return err
	}
	return nil
}

// directEntityMembership is the model for the direct_entity_memberships view in the database.
type directEntityMembership struct {
	bun.BaseModel `bun:"table:direct_entity_memberships,alias:mem"`

	AccountType       string `bun:"account_type,notnull"`
	AccountID         string `bun:"account_id,notnull"`
	AccountFriendlyID string `bun:"account_friendly_id,notnull"`
	Rights            []int  `bun:"rights,array,nullzero"`
	EntityType        string `bun:"entity_type,notnull"`
	EntityID          string `bun:"entity_id,notnull"`
	EntityFriendlyID  string `bun:"entity_friendly_id,notnull"`
}

func (directEntityMembership) _isModel() {} // Just a view in the database, but we can treat it as a model.

// indirectEntityMembership is the model for the indirect_entity_memberships view in the database.
type indirectEntityMembership struct {
	bun.BaseModel `bun:"table:indirect_entity_memberships,alias:mem"`

	UserAccountID                 string `bun:"user_account_id,notnull"`
	UserAccountFriendlyID         string `bun:"user_account_friendly_id,notnull"`
	UserRights                    []int  `bun:"user_rights,array,nullzero"`
	OrganizationAccountID         string `bun:"organization_account_id,notnull"`
	OrganizationAccountFriendlyID string `bun:"organization_account_friendly_id,notnull"`
	EntityRights                  []int  `bun:"entity_rights,array,nullzero"`
	EntityType                    string `bun:"entity_type,notnull"`
	EntityID                      string `bun:"entity_id,notnull"`
	EntityFriendlyID              string `bun:"entity_friendly_id,notnull"`
}

func (indirectEntityMembership) _isModel() {} // Just a view in the database, but we can treat it as a model.

type membershipStore struct {
	*entityStore
}

func newMembershipStore(baseStore *baseStore) *membershipStore {
	return &membershipStore{
		entityStore: newEntityStore(baseStore),
	}
}

func (s *membershipStore) selectWithUUIDsInMemberships(
	ctx context.Context, accountID *ttnpb.OrganizationOrUserIdentifiers, entityType string, includeIndirect bool,
) (func(*bun.SelectQuery) *bun.SelectQuery, error) {
	if includeIndirect {
		if accountID.EntityType() != "user" {
			panic(fmt.Errorf("invalid account type %q for indirect membership", accountID.EntityType()))
		}
		if entityType == "organization" {
			panic(fmt.Errorf("invalid entity type %q for indirect membership", entityType))
		}
	}

	account, err := s.getAccountModel(ctx, accountID.EntityType(), accountID.IDString())
	if err != nil {
		return nil, err
	}

	directMembershipSelectQuery := s.newSelectModel(ctx, &directEntityMembership{}).
		Column("entity_id").
		Where("account_id = ?", account.ID).
		Where("entity_type = ?", entityType)

	indirectMembershipSelectQuery := s.newSelectModel(ctx, &indirectEntityMembership{}).
		Column("entity_id").
		Where("user_account_id = ?", account.ID).
		Where("entity_type = ?", entityType)

	if includeIndirect {
		return func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where(
				"id IN (? UNION ?)",
				directMembershipSelectQuery,
				indirectMembershipSelectQuery,
			)
		}, nil
	}

	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.Where("id IN (?)", directMembershipSelectQuery)
	}, nil
}

func (s *membershipStore) CountMemberships(
	ctx context.Context, accountID *ttnpb.OrganizationOrUserIdentifiers, entityType string,
) (uint64, error) {
	account, err := s.getAccountModel(ctx, accountID.EntityType(), accountID.IDString())
	if err != nil {
		return 0, err
	}

	selectQuery := s.newSelectModel(ctx, &directEntityMembership{}).
		Column("entity_id").
		Where("account_id = ?", account.ID).
		Where("entity_type = ?", entityType)

	// Count the total number of results.
	count, err := selectQuery.Count(ctx)
	if err != nil {
		return 0, storeutil.WrapDriverError(err)
	}

	return uint64(count), nil
}

func (s *membershipStore) FindMemberships(
	ctx context.Context, accountID *ttnpb.OrganizationOrUserIdentifiers, entityType string, includeIndirect bool,
) ([]*ttnpb.EntityIdentifiers, error) {
	ctx, span := tracer.StartFromContext(ctx, "FindMemberships", trace.WithAttributes(
		attribute.String("account_type", accountID.EntityType()),
		attribute.String("account_id", accountID.IDString()),
		attribute.String("entity_type", entityType),
	))
	defer span.End()

	selectWithUUID, err := s.selectWithUUIDsInMemberships(
		ctx, accountID, entityType, includeIndirect && accountID.EntityType() == "user",
	)
	if err != nil {
		return nil, err
	}

	switch entityType {
	default:
		return nil, fmt.Errorf("invalid entity type %q", entityType)
	case "application":
		res, err := s.listApplicationsBy(ctx, selectWithUUID, store.FieldMask{"ids"})
		if err != nil {
			return nil, err
		}
		return mapSlice(res, (*ttnpb.Application).GetEntityIdentifiers), nil
	case "client":
		res, err := s.listClientsBy(ctx, selectWithUUID, store.FieldMask{"ids"})
		if err != nil {
			return nil, err
		}
		return mapSlice(res, (*ttnpb.Client).GetEntityIdentifiers), nil
	case "gateway":
		res, err := s.listGatewaysBy(ctx, selectWithUUID, store.FieldMask{"ids"})
		if err != nil {
			return nil, err
		}
		return mapSlice(res, (*ttnpb.Gateway).GetEntityIdentifiers), nil
	case "organization":
		res, err := s.listOrganizationsBy(ctx, selectWithUUID, store.FieldMask{"ids"})
		if err != nil {
			return nil, err
		}
		return mapSlice(res, (*ttnpb.Organization).GetEntityIdentifiers), nil
	}
}

func (*membershipStore) getOrganizationOrUserIdentifiers(
	accountType string, friendlyID string,
) *ttnpb.OrganizationOrUserIdentifiers {
	switch accountType {
	default:
		panic(fmt.Errorf("invalid account type: %s", accountType))
	case "organization":
		return (&ttnpb.OrganizationIdentifiers{OrganizationId: friendlyID}).GetOrganizationOrUserIdentifiers()
	case "user":
		return (&ttnpb.UserIdentifiers{UserId: friendlyID}).GetOrganizationOrUserIdentifiers()
	}
}

func (s *membershipStore) FindAccountMembershipChains(
	ctx context.Context, accountID *ttnpb.OrganizationOrUserIdentifiers, entityType string, entityIDs ...string,
) ([]*store.MembershipChain, error) {
	ctx, span := tracer.StartFromContext(ctx, "FindAccountMembershipChains", trace.WithAttributes(
		attribute.String("account_type", accountID.EntityType()),
		attribute.String("account_id", accountID.IDString()),
		attribute.String("entity_type", entityType),
	))
	defer span.End()

	account, err := s.getAccountModel(ctx, accountID.EntityType(), accountID.IDString())
	if err != nil {
		return nil, err
	}

	entityUUIDs, err := s.getEntityUUIDs(ctx, entityType, entityIDs...)
	if err != nil {
		return nil, err
	}

	var directMemberships []*directEntityMembership
	directSelectQuery := newSelectModels(ctx, s.DB, &directMemberships).
		Where("account_id = ?", account.ID).
		Where("entity_type = ?", entityType)

	if len(entityUUIDs) > 0 {
		directSelectQuery = directSelectQuery.Where("entity_id IN (?)", bun.In(entityUUIDs))
	}

	if err = directSelectQuery.Scan(ctx); err != nil {
		return nil, storeutil.WrapDriverError(err)
	}

	var indirectMemberships []*indirectEntityMembership
	indirectSelectQuery := newSelectModels(ctx, s.DB, &indirectMemberships).
		Where("user_account_id = ?", account.ID).
		Where("entity_type = ?", entityType)

	if len(entityUUIDs) > 0 {
		indirectSelectQuery = indirectSelectQuery.Where("entity_id IN (?)", bun.In(entityUUIDs))
	}

	if err = indirectSelectQuery.Scan(ctx); err != nil {
		return nil, storeutil.WrapDriverError(err)
	}

	membershipChains := make([]*store.MembershipChain, 0, len(directMemberships)+len(indirectMemberships))

	for _, directMembership := range directMemberships {
		mc := &store.MembershipChain{
			RightsOnEntity: &ttnpb.Rights{
				Rights: convertIntSlice[int, ttnpb.Right](directMembership.Rights),
			},
			EntityIdentifiers: getEntityIdentifiers(
				directMembership.EntityType, directMembership.EntityFriendlyID,
			),
		}
		switch directMembership.AccountType {
		case "organization":
			mc.OrganizationIdentifiers = &ttnpb.OrganizationIdentifiers{
				OrganizationId: directMembership.AccountFriendlyID,
			}
		default:
			mc.UserIdentifiers = &ttnpb.UserIdentifiers{
				UserId: directMembership.AccountFriendlyID,
			}
		}
		membershipChains = append(membershipChains, mc)
	}

	for _, indirectMembership := range indirectMemberships {
		membershipChains = append(membershipChains, &store.MembershipChain{
			UserIdentifiers: &ttnpb.UserIdentifiers{
				UserId: indirectMembership.UserAccountFriendlyID,
			},
			RightsOnOrganization: &ttnpb.Rights{
				Rights: convertIntSlice[int, ttnpb.Right](indirectMembership.UserRights),
			},
			OrganizationIdentifiers: &ttnpb.OrganizationIdentifiers{
				OrganizationId: indirectMembership.OrganizationAccountFriendlyID,
			},
			RightsOnEntity: &ttnpb.Rights{
				Rights: convertIntSlice[int, ttnpb.Right](indirectMembership.EntityRights),
			},
			EntityIdentifiers: getEntityIdentifiers(
				indirectMembership.EntityType, indirectMembership.EntityFriendlyID,
			),
		})
	}

	return membershipChains, nil
}

func (s *membershipStore) FindMembers(
	ctx context.Context, entityID *ttnpb.EntityIdentifiers,
) ([]*store.MemberByID, error) {
	ctx, span := tracer.StartFromContext(ctx, "FindMembers", trace.WithAttributes(
		attribute.String("entity_type", entityID.EntityType()),
		attribute.String("entity_id", entityID.IDString()),
	))
	defer span.End()

	entityType, entityUUID, err := s.getEntity(ctx, entityID)
	if err != nil {
		return nil, err
	}

	var models []*directEntityMembership
	selectQuery := newSelectModels(ctx, s.DB, &models).
		Where("entity_type = ?", entityType).
		Where("entity_id = ?", entityUUID)

	count, err := selectQuery.Count(ctx)
	if err != nil {
		return nil, storeutil.WrapDriverError(err)
	}
	store.SetTotal(ctx, uint64(count))

	// Apply ordering and paging.
	selectQuery = selectQuery.
		Apply(selectWithOrderFromContext(ctx, "account_friendly_id", map[string]string{
			"id":     "account_friendly_id",
			"rights": "rights",
		})).
		Apply(selectWithLimitAndOffsetFromContext(ctx))

	if err := selectQuery.Scan(ctx); err != nil {
		return nil, storeutil.WrapDriverError(err)
	}

	res := make([]*store.MemberByID, len(models))
	for i, model := range models {
		res[i] = &store.MemberByID{
			Ids: s.getOrganizationOrUserIdentifiers(model.AccountType, model.AccountFriendlyID),
			Rights: &ttnpb.Rights{
				Rights: convertIntSlice[int, ttnpb.Right](model.Rights),
			},
		}
	}

	return res, nil
}

func (s *membershipStore) GetMember(
	ctx context.Context, accountID *ttnpb.OrganizationOrUserIdentifiers, entityID *ttnpb.EntityIdentifiers,
) (*ttnpb.Rights, error) {
	ctx, span := tracer.StartFromContext(ctx, "GetMember", trace.WithAttributes(
		attribute.String("account_type", accountID.EntityType()),
		attribute.String("account_id", accountID.IDString()),
		attribute.String("entity_type", entityID.EntityType()),
		attribute.String("entity_id", entityID.IDString()),
	))
	defer span.End()

	account, err := s.getAccountModel(ctx, accountID.EntityType(), accountID.IDString())
	if err != nil {
		return nil, err
	}
	entityType, entityUUID, err := s.getEntity(ctx, entityID)
	if err != nil {
		return nil, err
	}

	var model directEntityMembership
	err = s.newSelectModel(ctx, &model).
		Where("account_type = ?", getEntityType(accountID)).
		Where("account_id = ?", account.ID).
		Where("entity_type = ?", entityType).
		Where("entity_id = ?", entityUUID).
		Scan(ctx)
	if err != nil {
		err = storeutil.WrapDriverError(err)
		if errors.IsNotFound(err) {
			return nil, store.ErrMembershipNotFound.WithAttributes(
				"account_type", accountID.EntityType(),
				"account_id", accountID.IDString(),
				"entity_type", entityID.EntityType(),
				"entity_id", entityID.IDString(),
			)
		}
		return nil, err
	}

	return &ttnpb.Rights{
		Rights: convertIntSlice[int, ttnpb.Right](model.Rights),
	}, nil
}

func (s *membershipStore) SetMember(
	ctx context.Context,
	accountID *ttnpb.OrganizationOrUserIdentifiers,
	entityID *ttnpb.EntityIdentifiers,
	rights *ttnpb.Rights,
) error {
	ctx, span := tracer.StartFromContext(ctx, "SetMember", trace.WithAttributes(
		attribute.String("account_type", accountID.EntityType()),
		attribute.String("account_id", accountID.IDString()),
		attribute.String("entity_type", entityID.EntityType()),
		attribute.String("entity_id", entityID.IDString()),
	))
	defer span.End()

	account, err := s.getAccountModel(ctx, accountID.EntityType(), accountID.IDString())
	if err != nil {
		return err
	}
	entityType, entityUUID, err := s.getEntity(ctx, entityID)
	if err != nil {
		return err
	}

	model := &Membership{}
	err = s.newSelectModel(ctx, model).
		Where("account_id = ?", account.ID).
		Where("entity_type = ?", entityType).
		Where("entity_id = ?", entityUUID).
		Scan(ctx)
	if err != nil {
		err = storeutil.WrapDriverError(err)
		if errors.IsNotFound(err) {
			_, err = s.DB.NewInsert().
				Model(&Membership{
					AccountID:  account.ID,
					Account:    account,
					Rights:     convertIntSlice[ttnpb.Right, int](rights.GetRights()),
					EntityID:   entityUUID,
					EntityType: entityType,
				}).
				Exec(ctx)
			if err != nil {
				return storeutil.WrapDriverError(err)
			}
			return nil
		}
		return err
	}

	model.Rights = convertIntSlice[ttnpb.Right, int](rights.GetRights())

	_, err = s.DB.NewUpdate().
		Model(model).
		WherePK().
		Column("rights", "updated_at").
		Exec(ctx)
	if err != nil {
		return storeutil.WrapDriverError(err)
	}

	return nil
}

// DeleteMember elminates the direct member rights attached to an entity.
func (s *membershipStore) DeleteMember(
	ctx context.Context, ids *ttnpb.OrganizationOrUserIdentifiers, entityID *ttnpb.EntityIdentifiers,
) error {
	ctx, span := tracer.StartFromContext(ctx, "DeleteMember", trace.WithAttributes(
		attribute.String("account_type", ids.EntityType()),
		attribute.String("account_id", ids.IDString()),
		attribute.String("entity_type", entityID.EntityType()),
		attribute.String("entity_id", entityID.IDString()),
	))
	defer span.End()

	account, err := s.getAccountModel(ctx, ids.EntityType(), ids.IDString())
	if err != nil {
		return err
	}
	entityType, entityUUID, err := s.getEntity(ctx, entityID)
	if err != nil {
		return err
	}

	_, err = s.DB.NewDelete().Model(&Membership{}).
		Where("account_id = ?", account.ID).
		Where("entity_type = ?", entityType).
		Where("entity_id = ?", entityUUID).
		Exec(ctx)
	if err != nil {
		return storeutil.WrapDriverError(err)
	}
	return nil
}

func (s *membershipStore) DeleteEntityMembers(
	ctx context.Context, entityID *ttnpb.EntityIdentifiers,
) error {
	ctx, span := tracer.StartFromContext(ctx, "DeleteEntityMembers", trace.WithAttributes(
		attribute.String("entity_type", entityID.EntityType()),
		attribute.String("entity_id", entityID.IDString()),
	))
	defer span.End()

	entityType, entityUUID, err := s.getEntity(store.WithSoftDeleted(ctx, false), entityID)
	if err != nil {
		return err
	}

	model := &Membership{}
	deleteQuery := s.DB.NewDelete().
		Model(model).
		Where("entity_type = ?", entityType).
		Where("entity_id = ?", entityUUID)
	if _, err = deleteQuery.Exec(ctx); err != nil {
		return storeutil.WrapDriverError(err)
	}

	return nil
}

func (s *membershipStore) DeleteAccountMembers(
	ctx context.Context, accountID *ttnpb.OrganizationOrUserIdentifiers,
) error {
	ctx, span := tracer.StartFromContext(ctx, "DeleteAccountMembers", trace.WithAttributes(
		attribute.String("account_type", accountID.EntityType()),
		attribute.String("account_id", accountID.IDString()),
	))
	defer span.End()

	account, err := s.getAccountModel(store.WithSoftDeleted(ctx, false), accountID.EntityType(), accountID.IDString())
	if err != nil {
		return err
	}

	model := &Membership{}
	deleteQuery := s.DB.NewDelete().
		Model(model).
		Where("account_id = ?", account.ID)
	if _, err = deleteQuery.Exec(ctx); err != nil {
		return storeutil.WrapDriverError(err)
	}

	return nil
}
