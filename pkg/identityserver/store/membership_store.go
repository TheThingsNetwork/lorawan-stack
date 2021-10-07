// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
	"runtime/trace"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// GetMembershipStore returns an MembershipStore on the given db (or transaction).
func GetMembershipStore(db *gorm.DB) MembershipStore {
	return &membershipStore{store: newStore(db)}
}

type membershipStore struct {
	*store
}

func (s *membershipStore) queryWithIndirectMemberships(ctx context.Context, entityType string, entityIDs ...string) *gorm.DB {
	idColumnName := fmt.Sprintf(`"%[1]ss"."%[1]s_id"`, entityType)
	if entityType == "organization" {
		idColumnName = `"organization_accounts"."uid"`
	}
	query := s.query(ctx, modelForEntityType(entityType)).Select([]string{
		`"indirect_accounts"."account_type" "indirect_account_type"`,
		`"indirect_accounts"."uid" "indirect_account_friendly_id"`,
		`"indirect_memberships"."rights" "indirect_account_rights"`,
		`"direct_accounts"."account_type" "direct_account_type"`,
		`"direct_accounts"."uid" "direct_account_friendly_id"`,
		`"direct_memberships"."rights" "direct_account_rights"`,
		`"direct_memberships"."entity_type" "entity_type"`,
		idColumnName + ` "entity_friendly_id"`,
	})
	if len(entityIDs) > 0 {
		query = query.Where(idColumnName+" IN (?)", entityIDs)
	}
	if entityType == "organization" {
		query = query.Joins(`JOIN "accounts" "organization_accounts" ON "organization_accounts"."account_id" = "organizations"."id" AND "organization_accounts"."account_type" = 'organization'`)
	}
	query = query.Joins(fmt.Sprintf(`JOIN "memberships" "direct_memberships" ON "direct_memberships"."entity_id" = "%[1]ss"."id" AND "direct_memberships"."entity_type" = '%[1]s'`, entityType)).
		Joins(`JOIN "accounts" "direct_accounts" ON "direct_accounts"."id" = "direct_memberships"."account_id"`).
		Joins(`LEFT JOIN "memberships" "indirect_memberships" ON "indirect_memberships"."entity_id" = "direct_accounts"."account_id" AND "indirect_memberships"."entity_type" = "direct_accounts"."account_type"`).
		Joins(`LEFT JOIN "accounts" "indirect_accounts" ON "indirect_accounts"."id" = "indirect_memberships"."account_id"`)
	query = query.Where(`"direct_accounts"."deleted_at" IS NULL AND "indirect_accounts"."deleted_at" IS NULL`)
	return query
}

func (s *membershipStore) queryWithDirectMemberships(ctx context.Context, entityType string, entityIDs ...string) *gorm.DB {
	idColumnName := fmt.Sprintf(`"%[1]ss"."%[1]s_id"`, entityType)
	if entityType == "organization" {
		idColumnName = `"organization_accounts"."uid"`
	}
	query := s.query(ctx, modelForEntityType(entityType)).Select([]string{
		`"direct_accounts"."account_type" "direct_account_type"`,
		`"direct_accounts"."uid" "direct_account_friendly_id"`,
		`"direct_memberships"."rights" "direct_account_rights"`,
		`"direct_memberships"."entity_type" "entity_type"`,
		idColumnName + ` "entity_friendly_id"`,
	})
	if len(entityIDs) > 0 {
		query = query.Where(idColumnName+" IN (?)", entityIDs)
	}
	if entityType == "organization" {
		query = query.Joins(`JOIN "accounts" "organization_accounts" ON "organization_accounts"."account_id" = "organizations"."id" AND "organization_accounts"."account_type" = 'organization'`)
	}
	query = query.Joins(fmt.Sprintf(`JOIN "memberships" "direct_memberships" ON "direct_memberships"."entity_id" = "%[1]ss"."id" AND "direct_memberships"."entity_type" = '%[1]s'`, entityType)).
		Joins(`JOIN "accounts" "direct_accounts" ON "direct_accounts"."id" = "direct_memberships"."account_id"`)
	query = query.Where(`"direct_accounts"."deleted_at" IS NULL`)
	return query
}

func (s *membershipStore) queryMemberships(ctx context.Context, accountID *ttnpb.OrganizationOrUserIdentifiers, entityType string, entityIDs []string, includeIndirect bool) *gorm.DB {
	if accountID.EntityType() == "user" && includeIndirect {
		return s.queryWithIndirectMemberships(ctx, entityType, entityIDs...).Where(
			`("direct_accounts"."account_type" = 'user' AND "direct_accounts"."uid" = ?) OR ("indirect_accounts"."account_type" = 'user' AND "indirect_accounts"."uid" = ?)`,
			accountID.IDString(), accountID.IDString(),
		)
	}
	return s.queryWithDirectMemberships(ctx, entityType, entityIDs...).Where(
		fmt.Sprintf(`"direct_accounts"."account_type" = '%s' AND "direct_accounts"."uid" = ?`, accountID.EntityType()),
		accountID.IDString(),
	)
}

func (s *membershipStore) FindMemberships(ctx context.Context, accountID *ttnpb.OrganizationOrUserIdentifiers, entityType string, includeIndirect bool) ([]*ttnpb.EntityIdentifiers, error) {
	defer trace.StartRegion(ctx, fmt.Sprintf("find %s memberships of %s", entityType, accountID.IDString())).End()

	membershipsQuery := s.queryMemberships(ctx, accountID, entityType, nil, includeIndirect).Select(`"direct_memberships"."entity_id"`).QueryExpr()
	query := s.query(ctx, modelForEntityType(entityType)).Where(fmt.Sprintf(`"%[1]ss"."id" IN (?)`, entityType), membershipsQuery)
	switch entityType {
	case "organization":
		query = query.
			Joins(`JOIN "accounts" ON "accounts"."account_type" = 'organization' AND "accounts"."account_id" = "organizations"."id"`).
			Select(`"accounts"."uid" AS "friendly_id"`)
	default:
		query = query.
			Select(fmt.Sprintf(`"%[1]ss"."%[1]s_id" AS "friendly_id"`, entityType))
	}

	query = query.Order(orderFromContext(ctx, fmt.Sprintf("%[1]ss", entityType), "friendly_id", "ASC"))
	page := query
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 {
		page = query.Limit(limit).Offset(offset)
	}
	var results []struct {
		FriendlyID string
	}
	if err := page.Scan(&results).Error; err != nil {
		return nil, err
	}
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 && (offset > 0 || len(results) == int(limit)) {
		countTotal(ctx, query)
	} else {
		setTotal(ctx, uint64(len(results)))
	}
	identifiers := make([]*ttnpb.EntityIdentifiers, len(results))
	for i, result := range results {
		identifiers[i] = buildIdentifiers(entityType, result.FriendlyID)
	}
	return identifiers, nil
}

// IndirectMembership returns an indirect membership through an organization.
type IndirectMembership struct {
	RightsOnOrganization *ttnpb.Rights
	*ttnpb.OrganizationIdentifiers
	OrganizationRights *ttnpb.Rights
}

func (s *membershipStore) FindIndirectMemberships(ctx context.Context, userID *ttnpb.UserIdentifiers, entityID *ttnpb.EntityIdentifiers) ([]IndirectMembership, error) {
	defer trace.StartRegion(ctx, fmt.Sprintf("find indirect memberships of user on %s", entityID.EntityType())).End()
	userQuery := s.query(WithoutSoftDeleted(ctx), Account{}).
		Select(`"accounts"."id"`).
		Where(`"accounts"."account_type" = 'user' AND "accounts"."uid" = ?`, userID.IDString()).
		QueryExpr()
	entityQuery := s.query(ctx, modelForID(entityID), withID(entityID)).
		Select(fmt.Sprintf(`"%ss"."id"`, entityID.EntityType())).
		QueryExpr()
	query := s.query(WithoutSoftDeleted(ctx), Account{}).
		Select(`"usr_memberships"."rights" AS "usr_rights", "accounts"."uid" AS "organization_id", "entity_memberships"."rights" AS "entity_rights"`).
		Joins(`JOIN "memberships" "usr_memberships" ON "usr_memberships"."entity_type" = 'organization' AND "usr_memberships"."entity_id" = "accounts"."account_id"`).
		Joins(`JOIN "memberships" "entity_memberships" ON "entity_memberships"."account_id" = "accounts"."id"`).
		Where(`"usr_memberships"."account_id" = (?)`, userQuery).
		Where(fmt.Sprintf(`"entity_memberships"."entity_type" = '%s' AND "entity_memberships"."entity_id" = (?)`, entityID.EntityType()), entityQuery)
	var res []struct {
		UsrRights      Rights
		OrganizationID string
		EntityRights   Rights
	}
	if err := query.Scan(&res).Error; err != nil {
		return nil, err
	}
	commonOrganizations := make([]IndirectMembership, len(res))
	for i, res := range res {
		usrRights, entityRights := ttnpb.Rights(res.UsrRights), ttnpb.Rights(res.EntityRights)
		commonOrganizations[i] = IndirectMembership{
			RightsOnOrganization:    &usrRights,
			OrganizationIdentifiers: &ttnpb.OrganizationIdentifiers{OrganizationId: res.OrganizationID},
			OrganizationRights:      &entityRights,
		}
	}
	return commonOrganizations, nil
}

func (s *membershipStore) FindMembers(ctx context.Context, entityID *ttnpb.EntityIdentifiers) (map[*ttnpb.OrganizationOrUserIdentifiers]*ttnpb.Rights, error) {
	defer trace.StartRegion(ctx, fmt.Sprintf("find members of %s", entityID.EntityType())).End()
	entityQuery := s.query(ctx, modelForID(entityID), withID(entityID)).
		Select(fmt.Sprintf(`"%ss"."id"`, entityID.EntityType())).
		QueryExpr()
	query := s.query(ctx, Account{}).
		Select(`"accounts"."uid" AS "uid", "accounts"."account_type" AS "account_type", "memberships"."rights" AS "rights"`).
		Joins(`JOIN "memberships" ON "memberships"."account_id" = "accounts"."id"`).
		Where(fmt.Sprintf(`"memberships"."entity_type" = '%s' AND "memberships"."entity_id" = (?)`, entityID.EntityType()), entityQuery).
		Order(`"uid"`)
	page := query
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 {
		page = query.Limit(limit).Offset(offset)
	}
	var results []struct {
		UID         string
		AccountType string
		Rights      Rights
	}
	if err := page.Scan(&results).Error; err != nil {
		return nil, err
	}
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 && (offset > 0 || len(results) == int(limit)) {
		countTotal(ctx, query)
	} else {
		setTotal(ctx, uint64(len(results)))
	}
	membershipRights := make(map[*ttnpb.OrganizationOrUserIdentifiers]*ttnpb.Rights, len(results))
	for _, result := range results {
		ids := Account{AccountType: result.AccountType, UID: result.UID}.OrganizationOrUserIdentifiers()
		rights := ttnpb.Rights(result.Rights)
		membershipRights[ids] = &rights
	}
	return membershipRights, nil
}

var errMembershipNotFound = errors.DefineNotFound(
	"membership_not_found",
	"account `{account_id}` is not a member of `{entity_type}` `{entity_id}`",
)

func (s *membershipStore) GetMember(ctx context.Context, id *ttnpb.OrganizationOrUserIdentifiers, entityID *ttnpb.EntityIdentifiers) (*ttnpb.Rights, error) {
	defer trace.StartRegion(ctx, "get membership").End()
	accountQuery := s.query(ctx, Account{}).
		Select(`"accounts"."id"`).
		Where(fmt.Sprintf(`"accounts"."account_type" = '%s' AND "accounts"."uid" = ?`, id.EntityType()), id.IDString()).
		QueryExpr()
	entityQuery := s.query(ctx, modelForID(entityID), withID(entityID)).
		Select(fmt.Sprintf(`"%ss"."id"`, entityID.EntityType())).
		QueryExpr()
	query := s.query(ctx, &Membership{}).
		Select(`"memberships"."rights"`).
		Where(`"memberships"."account_id" = (?)`, accountQuery).
		Where(fmt.Sprintf(`"memberships"."entity_type" = '%s' AND "memberships"."entity_id" = (?)`, entityID.EntityType()), entityQuery)
	var membership Membership
	err := query.First(&membership).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errMembershipNotFound.WithAttributes(
				"account_id", id.IDString(),
				"entity_type", entityID.EntityType(),
				"entity_id", entityID.IDString(),
			)
		}
		return nil, err
	}
	rights := ttnpb.Rights(membership.Rights)
	return &rights, nil
}

var errAccountType = errors.DefineInvalidArgument(
	"account_type",
	"account of type `{account_type}` can not collaborate on `{entity_type}`",
)

func (s *membershipStore) SetMember(ctx context.Context, id *ttnpb.OrganizationOrUserIdentifiers, entityID *ttnpb.EntityIdentifiers, rights *ttnpb.Rights) error {
	defer trace.StartRegion(ctx, "update membership").End()
	var account Account
	err := s.query(ctx, Account{}).Where(Account{
		UID:         id.IDString(),
		AccountType: id.EntityType(),
	}).Find(&account).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return errNotFoundForID(id)
		}
		return err
	}
	entity, err := s.findEntity(ctx, entityID, "id")
	if err != nil {
		return err
	}
	if _, ok := entity.(*Organization); ok && account.AccountType != "user" {
		return errAccountType.WithAttributes("account_type", account.AccountType, "entity_type", "organization")
	}
	if err := ctx.Err(); err != nil { // Early exit if context canceled
		return err
	}

	query := s.query(ctx, Membership{})
	var membership Membership
	err = query.Where(&Membership{
		AccountID:  account.PrimaryKey(),
		EntityID:   entity.PrimaryKey(),
		EntityType: entityTypeForID(entityID),
	}).First(&membership).Error
	if err == nil {
		if len(rights.Rights) == 0 {
			return query.Delete(&membership).Error
		}
		query = query.Select("rights", "updated_at")
	} else if gorm.IsRecordNotFoundError(err) {
		if len(rights.Rights) == 0 {
			return err
		}
		membership = Membership{
			AccountID:  account.PrimaryKey(),
			EntityID:   entity.PrimaryKey(),
			EntityType: entityTypeForID(entityID),
		}
		membership.SetContext(ctx)
	} else if gorm.IsRecordNotFoundError(err) {
		return errMembershipNotFound.WithAttributes(
			"account_id", id.IDString(),
			"entity_type", entityID.EntityType(),
			"entity_id", entityID.IDString(),
		)
	} else {
		return err
	}
	membership.Rights = Rights(*rights)
	return query.Save(&membership).Error
}

func (s *membershipStore) DeleteEntityMembers(ctx context.Context, entityID *ttnpb.EntityIdentifiers) error {
	defer trace.StartRegion(ctx, "delete entity memberships").End()
	entity, err := s.findDeletedEntity(ctx, entityID, "id")
	if err != nil {
		return err
	}
	return s.query(ctx, Membership{}).Where(&Membership{
		EntityID:   entity.PrimaryKey(),
		EntityType: entityTypeForID(entityID),
	}).Delete(&Membership{}).Error
}

func (s *membershipStore) DeleteAccountMembers(ctx context.Context, id *ttnpb.OrganizationOrUserIdentifiers) error {
	defer trace.StartRegion(ctx, "delete account memberships").End()
	var account Account
	err := s.query(ctx, Account{}, withSoftDeleted()).Where(Account{
		UID:         id.IDString(),
		AccountType: id.EntityType(),
	}).Find(&account).Error
	if err != nil {
		return err
	}
	return s.query(ctx, Membership{}).Where(&Membership{
		AccountID: account.PrimaryKey(),
	}).Delete(&Membership{}).Error
}
