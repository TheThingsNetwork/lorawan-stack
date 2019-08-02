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
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 {
		countTotal(ctx, query)
		query = query.Limit(limit).Offset(offset)
	}
	var results []struct {
		FriendlyID string
	}
	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}
	identifiers := make([]ttnpb.Identifiers, len(results))
	for i, result := range results {
		identifiers[i] = buildIdentifiers(entityType, result.FriendlyID)
	}
	return identifiers, nil
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

func (s *membershipStore) FindMemberRights(ctx context.Context, id *ttnpb.OrganizationOrUserIdentifiers, entityType string) (map[ttnpb.Identifiers]*ttnpb.Rights, error) {
	entityTypeForTrace := entityType
	if entityTypeForTrace == "" {
		entityTypeForTrace = "all"
	}
	defer trace.StartRegion(ctx, fmt.Sprintf("find %s memberships for %s", entityTypeForTrace, id.EntityType())).End()
	account, err := s.findAccount(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil { // Early exit if context canceled
		return nil, err
	}
	memberships, err := s.findAccountMemberships(account, entityType)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil { // Early exit if context canceled
		return nil, err
	}
	return s.identifierRights(entityRightsForMemberships(memberships))
}

var errMembershipNotFound = errors.DefineNotFound(
	"membership_not_found",
	"account `{account_id}` is not a member of `{entity_type}` `{entity_id}`",
)

func (s *membershipStore) GetMember(ctx context.Context, id *ttnpb.OrganizationOrUserIdentifiers, entityID ttnpb.Identifiers) (*ttnpb.Rights, error) {
	defer trace.StartRegion(ctx, "get membership").End()
	account, err := s.findAccount(ctx, id)
	if err != nil {
		return nil, err
	}
	entity, err := s.findEntity(ctx, entityID, "id")
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil { // Early exit if context canceled
		return nil, err
	}
	query := s.query(ctx, Membership{})
	var membership Membership
	err = query.Where(&Membership{
		AccountID:  account.PrimaryKey(),
		EntityID:   entity.PrimaryKey(),
		EntityType: entityTypeForID(entityID),
	}).First(&membership).Error
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
