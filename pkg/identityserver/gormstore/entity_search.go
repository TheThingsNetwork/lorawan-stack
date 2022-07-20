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
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// GetEntitySearch returns an EntitySearch on the given db (or transaction).
func GetEntitySearch(db *gorm.DB) store.EntitySearch {
	return &entitySearch{baseStore: newStore(db)}
}

type entitySearch struct {
	*baseStore
}

type metaFields interface {
	GetIdContains() string
	GetNameContains() string
	GetDescriptionContains() string
	GetAttributesContain() map[string]string
}

func likePattern(v string) string { return "%" + v + "%" }

func ftsQuery(query *gorm.DB, field string) string {
	if dbKind, ok := query.Get("db:kind"); ok && dbKind == "PostgreSQL" {
		if dbMajor, ok := query.Get("db:version:major"); ok && dbMajor.(int) >= 11 {
			return fmt.Sprintf("to_tsvector('english', %s) @@ websearch_to_tsquery('english', ?)", field)
		}
		return fmt.Sprintf("to_tsvector('english', %s) @@ to_tsquery('english', ?)", field)
	}
	return fmt.Sprintf("%s ILIKE '%%' || ? || '%%'", field)
}

func (s *entitySearch) queryMetaFields(
	ctx context.Context, query *gorm.DB, entityType string, req metaFields,
) *gorm.DB {
	if v := req.GetIdContains(); v != "" {
		switch entityType {
		case organization, user:
			query = query.Where(`"accounts"."uid" ILIKE ?`, likePattern(v))
		case endDevice:
			query = query.Where(`"end_devices"."device_id" ILIKE ?`, likePattern(v))
		default:
			query = query.Where(fmt.Sprintf(`"%[1]ss"."%[1]s_id" ILIKE ?`, entityType), likePattern(v))
		}
	}
	if v := req.GetNameContains(); v != "" {
		query = query.Where("name ILIKE ?", likePattern(v))
	}
	if v := req.GetDescriptionContains(); v != "" {
		query = query.Where(ftsQuery(query, "description"), v)
	}
	if kv := req.GetAttributesContain(); len(kv) > 0 {
		sub := s.query(ctx, &Attribute{}).Select("entity_id")
		switch entityType {
		case endDevice:
			sub = sub.Where("entity_type = ?", "device")
		default:
			sub = sub.Where("entity_type = ?", entityType)
		}
		for k, v := range kv {
			sub = sub.Where("key = ? AND value ILIKE ?", k, likePattern(v))
		}
		query = query.Where(fmt.Sprintf(`"%ss"."id" IN (?)`, entityType), sub.QueryExpr())
	}
	return query
}

func (s *entitySearch) queryMembership(
	ctx context.Context, query *gorm.DB, entityType string, member *ttnpb.OrganizationOrUserIdentifiers,
) *gorm.DB {
	if member == nil {
		return query
	}
	membershipsQuery := (&membershipStore{baseStore: s.baseStore}).
		queryMemberships(ctx, member, entityType, nil, true).
		Select(`"direct_memberships"."entity_id"`).
		QueryExpr()
	if entityType == organization {
		query = query.Where(
			`"accounts"."account_type" = ? AND "accounts"."account_id" IN (?)`,
			entityType, membershipsQuery,
		)
	} else {
		query = query.Where(fmt.Sprintf(`"%[1]ss"."id" IN (?)`, entityType), membershipsQuery)
	}
	return query
}

type searchResult struct {
	FriendlyID string
}

func (*entitySearch) runPaginatedQuery(
	ctx context.Context, query *gorm.DB, entityType string,
) ([]searchResult, error) {
	query = query.Order(store.OrderFromContext(ctx, fmt.Sprintf("%ss", entityType), "friendly_id", "ASC"))
	page := query
	if limit, offset := store.LimitAndOffsetFromContext(ctx); limit != 0 {
		page = query.Limit(limit).Offset(offset)
	}
	var results []searchResult
	if err := page.Scan(&results).Error; err != nil {
		return nil, err
	}
	if limit, offset := store.LimitAndOffsetFromContext(ctx); limit != 0 && (offset > 0 || len(results) == int(limit)) {
		var total uint64
		query.Count(&total)
		store.SetTotal(ctx, total)
	} else {
		store.SetTotal(ctx, uint64(len(results)))
	}
	return results, nil
}

func (s *entitySearch) SearchApplications(
	ctx context.Context, member *ttnpb.OrganizationOrUserIdentifiers, req *ttnpb.SearchApplicationsRequest,
) ([]*ttnpb.ApplicationIdentifiers, error) {
	defer trace.StartRegion(ctx, "find applications").End()
	if req.Deleted {
		ctx = store.WithSoftDeleted(ctx, true)
	}
	query := s.query(ctx, &Application{})
	query = query.Select(`"application_id" AS "friendly_id"`)
	query = s.queryMembership(ctx, query, application, member)
	if q := req.GetQuery(); q != "" {
		query = query.Where(
			`"application_id" ILIKE ? OR "name" ILIKE ? OR `+ftsQuery(query, "description"),
			likePattern(q), likePattern(q), q,
		)
	}
	query = s.queryMetaFields(ctx, query, application, req)
	results, err := s.runPaginatedQuery(ctx, query, application)
	if err != nil {
		return nil, err
	}
	identifiers := make([]*ttnpb.ApplicationIdentifiers, len(results))
	for i, result := range results {
		identifiers[i] = &ttnpb.ApplicationIdentifiers{ApplicationId: result.FriendlyID}
	}
	return identifiers, nil
}

func (s *entitySearch) SearchClients(
	ctx context.Context, member *ttnpb.OrganizationOrUserIdentifiers, req *ttnpb.SearchClientsRequest,
) ([]*ttnpb.ClientIdentifiers, error) {
	defer trace.StartRegion(ctx, "find clients").End()
	if req.Deleted {
		ctx = store.WithSoftDeleted(ctx, true)
	}
	query := s.query(ctx, &Client{})
	query = query.Select(`"client_id" AS "friendly_id"`)
	query = s.queryMembership(ctx, query, client, member)
	if q := req.GetQuery(); q != "" {
		query = query.Where(
			`"client_id" ILIKE ? OR "name" ILIKE ? OR `+ftsQuery(query, "description"),
			likePattern(q), likePattern(q), q,
		)
	}
	query = s.queryMetaFields(ctx, query, client, req)
	if len(req.State) > 0 {
		stateNumbers := make([]int, len(req.State))
		for i, state := range req.State {
			stateNumbers[i] = int(state)
		}
		query = query.Where(`"state" IN (?)`, stateNumbers)
	}
	results, err := s.runPaginatedQuery(ctx, query, client)
	if err != nil {
		return nil, err
	}
	identifiers := make([]*ttnpb.ClientIdentifiers, len(results))
	for i, result := range results {
		identifiers[i] = &ttnpb.ClientIdentifiers{ClientId: result.FriendlyID}
	}
	return identifiers, nil
}

func (s *entitySearch) SearchGateways(
	ctx context.Context, member *ttnpb.OrganizationOrUserIdentifiers, req *ttnpb.SearchGatewaysRequest,
) ([]*ttnpb.GatewayIdentifiers, error) {
	defer trace.StartRegion(ctx, "find gateways").End()
	if req.Deleted {
		ctx = store.WithSoftDeleted(ctx, true)
	}
	query := s.query(ctx, &Gateway{})
	query = query.Select(`"gateway_id" AS "friendly_id"`)
	query = s.queryMembership(ctx, query, gateway, member)
	if q := req.GetQuery(); q != "" {
		query = query.Where(
			`"gateway_id" ILIKE ? OR "gateway_eui" ILIKE ? OR "name" ILIKE ? OR `+ftsQuery(query, "description"),
			likePattern(q), likePattern(q), likePattern(q), q,
		)
	}
	query = s.queryMetaFields(ctx, query, gateway, req)
	if v := req.EuiContains; v != "" {
		query = query.Where(`"gateway_eui" ILIKE ?`, likePattern(v))
	}
	results, err := s.runPaginatedQuery(ctx, query, gateway)
	if err != nil {
		return nil, err
	}
	identifiers := make([]*ttnpb.GatewayIdentifiers, len(results))
	for i, result := range results {
		identifiers[i] = &ttnpb.GatewayIdentifiers{GatewayId: result.FriendlyID}
	}
	return identifiers, nil
}

func (s *entitySearch) SearchOrganizations(
	ctx context.Context, member *ttnpb.OrganizationOrUserIdentifiers, req *ttnpb.SearchOrganizationsRequest,
) ([]*ttnpb.OrganizationIdentifiers, error) {
	defer trace.StartRegion(ctx, "find organizations").End()
	if req.Deleted {
		ctx = store.WithSoftDeleted(ctx, true)
	}
	query := s.query(ctx, &Organization{})
	query = query.
		Joins(`JOIN "accounts" ON "accounts"."account_type" = 'organization' AND "accounts"."account_id" = "organizations"."id"`). //nolint:lll
		Select(`"accounts"."uid" AS "friendly_id"`)
	query = s.queryMembership(ctx, query, organization, member)
	if q := req.GetQuery(); q != "" {
		query = query.Where(
			`"accounts"."uid" ILIKE ? OR "name" ILIKE ? OR `+ftsQuery(query, "description"),
			likePattern(q), likePattern(q), q,
		)
	}
	query = s.queryMetaFields(ctx, query, organization, req)
	results, err := s.runPaginatedQuery(ctx, query, organization)
	if err != nil {
		return nil, err
	}
	identifiers := make([]*ttnpb.OrganizationIdentifiers, len(results))
	for i, result := range results {
		identifiers[i] = &ttnpb.OrganizationIdentifiers{OrganizationId: result.FriendlyID}
	}
	return identifiers, nil
}

func (s *entitySearch) SearchUsers(
	ctx context.Context, req *ttnpb.SearchUsersRequest,
) ([]*ttnpb.UserIdentifiers, error) {
	defer trace.StartRegion(ctx, "find users").End()
	if req.Deleted {
		ctx = store.WithSoftDeleted(ctx, true)
	}
	query := s.query(ctx, &User{})
	query = query.
		Joins(`JOIN "accounts" ON "accounts"."account_type" = 'user' AND "accounts"."account_id" = "users"."id"`).
		Select(`"accounts"."uid" AS "friendly_id"`)
	if q := req.GetQuery(); q != "" {
		q = likePattern(q)
		query = query.Where(
			`"accounts"."uid" ILIKE ? OR "name" ILIKE ? OR `+ftsQuery(query, "description"),
			likePattern(q), likePattern(q), q,
		)
	}
	query = s.queryMetaFields(ctx, query, user, req)
	if len(req.State) > 0 {
		stateNumbers := make([]int, len(req.State))
		for i, state := range req.State {
			stateNumbers[i] = int(state)
		}
		query = query.Where(`"state" IN (?)`, stateNumbers)
	}
	results, err := s.runPaginatedQuery(ctx, query, user)
	if err != nil {
		return nil, err
	}
	identifiers := make([]*ttnpb.UserIdentifiers, len(results))
	for i, result := range results {
		identifiers[i] = &ttnpb.UserIdentifiers{UserId: result.FriendlyID}
	}
	return identifiers, nil
}

func (s *entitySearch) SearchAccounts(
	ctx context.Context, req *ttnpb.SearchAccountsRequest,
) ([]*ttnpb.OrganizationOrUserIdentifiers, error) {
	defer trace.StartRegion(ctx, "find accounts").End()

	var query *gorm.DB

	if entityID := req.GetEntityIdentifiers(); entityID != nil {
		query = (&membershipStore{baseStore: s.baseStore}).queryWithDirectMemberships(
			ctx, entityID.EntityType(), entityID.IDString(),
		)
	} else {
		query = s.query(ctx, &Account{}).
			Table(`accounts AS direct_accounts`).
			Select([]string{
				`"direct_accounts"."account_type" "direct_account_type"`,
				`"direct_accounts"."uid" "direct_account_friendly_id"`,
			})
	}

	if req.OnlyUsers {
		query = query.Where(`"direct_accounts"."account_type" = 'user'`)
	}

	query = query.Order("direct_account_friendly_id")

	if q := req.GetQuery(); q != "" {
		query = query.Where(`"direct_accounts"."uid" ILIKE ?`, likePattern(q))
	}

	page := query
	if limit, offset := store.LimitAndOffsetFromContext(ctx); limit != 0 {
		page = query.Limit(limit).Offset(offset)
	}

	var results []struct {
		DirectAccountType       string
		DirectAccountFriendlyID string
	}
	if err := page.Scan(&results).Error; err != nil {
		return nil, err
	}

	if limit, offset := store.LimitAndOffsetFromContext(ctx); limit != 0 && (offset > 0 || len(results) == int(limit)) {
		var total uint64
		query.Count(&total)
		store.SetTotal(ctx, total)
	} else {
		store.SetTotal(ctx, uint64(len(results)))
	}

	identifiers := make([]*ttnpb.OrganizationOrUserIdentifiers, len(results))

	for i, m := range results {
		switch m.DirectAccountType {
		case user:
			identifiers[i] = (&ttnpb.UserIdentifiers{
				UserId: m.DirectAccountFriendlyID,
			}).GetOrganizationOrUserIdentifiers()
		case organization:
			identifiers[i] = (&ttnpb.OrganizationIdentifiers{
				OrganizationId: m.DirectAccountFriendlyID,
			}).GetOrganizationOrUserIdentifiers()
		}
	}

	return identifiers, nil
}

func (s *entitySearch) SearchEndDevices(
	ctx context.Context, req *ttnpb.SearchEndDevicesRequest,
) ([]*ttnpb.EndDeviceIdentifiers, error) {
	defer trace.StartRegion(ctx, "find end devices").End()

	query := s.query(ctx, &EndDevice{}).
		Where(&EndDevice{ApplicationID: req.GetApplicationIds().GetApplicationId()}).
		Select(`"device_id" AS "friendly_id"`)

	if q := req.GetQuery(); q != "" {
		query = query.Where(
			`"device_id" ILIKE ? OR "dev_eui" ILIKE ? OR "join_eui" ILIKE ? OR "name" ILIKE ? OR `+ftsQuery(query, "description"), //nolint:lll
			likePattern(q), likePattern(q), likePattern(q), likePattern(q), q,
		)
	}

	query = s.queryMetaFields(ctx, query, endDevice, req)

	if v := req.DevEuiContains; v != "" {
		query = query.Where("dev_eui ILIKE ?", likePattern(v))
	}
	if v := req.JoinEuiContains; v != "" {
		query = query.Where("join_eui ILIKE ?", likePattern(v))
	}
	// DevAddrContains

	query = query.Order(store.OrderFromContext(ctx, "end_devices", "device_id", "ASC"))
	page := query
	if limit, offset := store.LimitAndOffsetFromContext(ctx); limit != 0 {
		page = query.Limit(limit).Offset(offset)
	}
	var results []struct {
		FriendlyID string
	}
	if err := page.Scan(&results).Error; err != nil {
		return nil, err
	}
	limit, offset := store.LimitAndOffsetFromContext(ctx)
	if limit != 0 && (offset > 0 || len(results) == int(limit)) {
		var total uint64
		query.Count(&total)
		store.SetTotal(ctx, total)
	} else {
		store.SetTotal(ctx, uint64(len(results)))
	}
	identifiers := make([]*ttnpb.EndDeviceIdentifiers, len(results))
	for i, result := range results {
		identifiers[i] = &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: &ttnpb.ApplicationIdentifiers{
				ApplicationId: req.GetApplicationIds().GetApplicationId(),
			},
			DeviceId: result.FriendlyID,
		}
	}
	return identifiers, nil
}
