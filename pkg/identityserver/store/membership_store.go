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
	"fmt"
	"runtime/trace"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// GetMembershipStore returns an MembershipStore on the given db (or transaction).
func GetMembershipStore(db *gorm.DB) MembershipStore {
	return &membershipStore{store: newStore(db)}
}

type membershipStore struct {
	*store
}

func (s *membershipStore) FindMemberships(ctx context.Context, id *ttnpb.OrganizationOrUserIdentifiers, entityType string, includeIndirect bool) ([]ttnpb.Identifiers, error) {
	defer trace.StartRegion(ctx, fmt.Sprintf("find %s memberships of %s", entityType, id.IDString())).End()
	accountQuery := s.query(ctx, Account{}).
		Select(`"accounts"."id"`).
		Where(fmt.Sprintf(`"accounts"."account_type" = '%s' AND "accounts"."uid" = ?`, id.EntityType()), id.IDString()).
		QueryExpr()
	query := s.query(ctx, modelForEntityType(entityType))
	if entityType == "organization" {
		query = query.Table("accounts").
			Select(`DISTINCT "accounts"."uid" AS "friendly_id"`).
			Joins(fmt.Sprintf(`JOIN "memberships" ON "memberships"."entity_type" = '%s' AND "memberships"."entity_id" = "accounts"."account_id"`, entityType))
	} else {
		query = query.
			Select(fmt.Sprintf(`DISTINCT "%[1]ss"."%[1]s_id" AS "friendly_id"`, entityType)).
			Joins(fmt.Sprintf(`JOIN "memberships" ON "memberships"."entity_type" = '%[1]s' AND "memberships"."entity_id" = "%[1]ss"."id"`, entityType))
	}
	query = query.Order(`"friendly_id"`).
		Where(fmt.Sprintf(`"memberships"."entity_type" = '%s' AND "memberships"."account_id" = (?)`, entityType), accountQuery)
	if includeIndirect && id.EntityType() == "user" {
		organizationQuery := s.query(ctx, Account{}).
			Select(`"accounts"."id"`).
			Joins(`JOIN "memberships" ON "memberships"."entity_type" = "accounts"."account_type" AND "memberships"."entity_id" = "accounts"."account_id"`).
			Where(`"memberships"."account_id" IN (?)`, accountQuery).
			QueryExpr()
		query = query.Or(`"memberships"."account_id" IN (?)`, organizationQuery)
	}
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
	identifiers := make([]ttnpb.Identifiers, len(results))
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

func (s *membershipStore) FindIndirectMemberships(ctx context.Context, userID *ttnpb.UserIdentifiers, entityID ttnpb.Identifiers) ([]IndirectMembership, error) {
	defer trace.StartRegion(ctx, fmt.Sprintf("find indirect memberships of user on %s", entityID.EntityType())).End()
	userQuery := s.query(ctx, Account{}).
		Select(`"accounts"."id"`).
		Where(`"accounts"."account_type" = 'user' AND "accounts"."uid" = ?`, userID.IDString()).
		QueryExpr()
	entityQuery := s.query(ctx, modelForID(entityID), withID(entityID)).
		Select(fmt.Sprintf(`"%ss"."id"`, entityID.EntityType())).
		QueryExpr()
	query := s.query(ctx, Account{}).
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
			OrganizationIdentifiers: &ttnpb.OrganizationIdentifiers{OrganizationID: res.OrganizationID},
			OrganizationRights:      &entityRights,
		}
	}
	return commonOrganizations, nil
}

func (s *membershipStore) FindMembers(ctx context.Context, entityID ttnpb.Identifiers) (map[*ttnpb.OrganizationOrUserIdentifiers]*ttnpb.Rights, error) {
	defer trace.StartRegion(ctx, fmt.Sprintf("find members of %s", entityID.EntityType())).End()
	entity, err := s.findEntity(ctx, entityID, "id")
	if err != nil {
		return nil, err
	}
	query := s.query(ctx, Membership{}).Where(&Membership{
		EntityID:   entity.PrimaryKey(),
		EntityType: entityTypeForID(entityID),
	}).Preload("Account")
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 {
		countTotal(ctx, query.Model(&Membership{}))
		query = query.Limit(limit).Offset(offset)
	}
	var memberships []Membership
	if err = query.Find(&memberships).Error; err != nil {
		return nil, err
	}
	setTotal(ctx, uint64(len(memberships)))
	membershipRights := make(map[*ttnpb.OrganizationOrUserIdentifiers]*ttnpb.Rights, len(memberships))
	for _, membership := range memberships {
		ids := membership.Account.OrganizationOrUserIdentifiers()
		rights := ttnpb.Rights(membership.Rights)
		membershipRights[ids] = &rights
	}
	return membershipRights, nil
}

var errMembershipNotFound = errors.DefineNotFound(
	"membership_not_found",
	"account `{account_id}` is not a member of `{entity_type}` `{entity_id}`",
)

func (s *membershipStore) GetMember(ctx context.Context, id *ttnpb.OrganizationOrUserIdentifiers, entityID ttnpb.Identifiers) (*ttnpb.Rights, error) {
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

func (s *membershipStore) SetMember(ctx context.Context, id *ttnpb.OrganizationOrUserIdentifiers, entityID ttnpb.Identifiers, rights *ttnpb.Rights) (err error) {
	defer trace.StartRegion(ctx, "update membership").End()
	account, err := s.findAccount(ctx, id)
	if err != nil {
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
