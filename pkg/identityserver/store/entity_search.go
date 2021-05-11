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
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// GetEntitySearch returns an EntitySearch on the given db (or transaction).
func GetEntitySearch(db *gorm.DB) EntitySearch {
	return &entitySearch{store: newStore(db)}
}

type entitySearch struct {
	*store
}

type metaFields interface {
	GetIDContains() string
	GetNameContains() string
	GetDescriptionContains() string
	GetAttributesContain() map[string]string
}

func (s *entitySearch) queryMetaFields(ctx context.Context, query *gorm.DB, entityType string, req metaFields) *gorm.DB {
	if v := req.GetIDContains(); v != "" {
		switch entityType {
		case "organization", "user":
			query = query.Where(`"accounts"."uid" LIKE ?`, "%"+v+"%")
		case "end_device":
			query = query.Where(`"end_devices"."device_id" LIKE ?`, "%"+v+"%")
		default:
			query = query.Where(fmt.Sprintf(`"%[1]ss"."%[1]s_id" LIKE ?`, entityType), "%"+v+"%")
		}
	}
	if dbKind, ok := query.Get("db:kind"); ok && dbKind == "PostgreSQL" {
		language := "english"
		if v := req.GetNameContains(); v != "" {
			query = query.Where(fmt.Sprintf("to_tsvector('%[1]s', name) @@ to_tsquery('%[1]s', ?)", language), v)
		}
		if v := req.GetDescriptionContains(); v != "" {
			query = query.Where(fmt.Sprintf("to_tsvector('%[1]s', description) @@ to_tsquery('%[1]s', ?)", language), v)
		}
	} else {
		if v := req.GetNameContains(); v != "" {
			query = query.Where("name ILIKE ?", fmt.Sprintf("%%%s%%", v))
		}
		if v := req.GetDescriptionContains(); v != "" {
			query = query.Where("description ILIKE ?", fmt.Sprintf("%%%s%%", v))
		}
	}
	if kv := req.GetAttributesContain(); len(kv) > 0 {
		sub := s.query(ctx, &Attribute{}).Select("entity_id")
		switch entityType {
		case "end_device":
			sub = sub.Where("entity_type = ?", "device")
		default:
			sub = sub.Where("entity_type = ?", entityType)
		}
		for k, v := range kv {
			sub = sub.Where("key = ? AND value ILIKE ?", k, fmt.Sprintf("%%%s%%", v))
		}
		query = query.Where(fmt.Sprintf(`"%ss"."id" IN (?)`, entityType), sub.QueryExpr())
	}
	return query
}

func (s *entitySearch) queryMembership(ctx context.Context, query *gorm.DB, entityType string, member *ttnpb.OrganizationOrUserIdentifiers) *gorm.DB {
	if member == nil {
		return query
	}
	membershipsQuery := (&membershipStore{store: s.store}).queryMemberships(ctx, member, entityType, true).Select("entity_id").QueryExpr()
	if entityType == "organization" {
		query = query.Where(`"accounts"."account_type" = ? AND "accounts"."account_id" IN (?)`, entityType, membershipsQuery)
	} else {
		query = query.Where(fmt.Sprintf(`"%[1]ss"."id" IN (?)`, entityType), membershipsQuery)
	}
	return query
}

type searchResult struct {
	FriendlyID string
}

func (s *entitySearch) runPaginatedQuery(ctx context.Context, query *gorm.DB, entityType string) ([]searchResult, error) {
	query = query.Order(orderFromContext(ctx, fmt.Sprintf("%ss", entityType), "friendly_id", "ASC"))
	page := query
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 {
		page = query.Limit(limit).Offset(offset)
	}
	var results []searchResult
	if err := page.Scan(&results).Error; err != nil {
		return nil, err
	}
	if limit, offset := limitAndOffsetFromContext(ctx); limit != 0 && (offset > 0 || len(results) == int(limit)) {
		countTotal(ctx, query)
	} else {
		setTotal(ctx, uint64(len(results)))
	}
	return results, nil
}

const application = "application"

func (s *entitySearch) FindApplications(ctx context.Context, member *ttnpb.OrganizationOrUserIdentifiers, req *ttnpb.SearchApplicationsRequest) ([]*ttnpb.ApplicationIdentifiers, error) {
	defer trace.StartRegion(ctx, "find applications").End()
	if req.Deleted {
		ctx = WithSoftDeleted(ctx, true)
	}
	query := s.query(ctx, &Application{})
	query = query.Select(fmt.Sprintf(`"%[1]ss"."%[1]s_id" AS "friendly_id"`, application))
	query = s.queryMembership(ctx, query, application, member)
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

const client = "client"

func (s *entitySearch) FindClients(ctx context.Context, member *ttnpb.OrganizationOrUserIdentifiers, req *ttnpb.SearchClientsRequest) ([]*ttnpb.ClientIdentifiers, error) {
	defer trace.StartRegion(ctx, "find clients").End()
	if req.Deleted {
		ctx = WithSoftDeleted(ctx, true)
	}
	query := s.query(ctx, &Client{})
	query = query.Select(fmt.Sprintf(`"%[1]ss"."%[1]s_id" AS "friendly_id"`, client))
	query = s.queryMembership(ctx, query, client, member)
	query = s.queryMetaFields(ctx, query, client, req)
	if len(req.State) > 0 {
		stateNumbers := make([]int, len(req.State))
		for i, state := range req.State {
			stateNumbers[i] = int(state)
		}
		query = query.Where(fmt.Sprintf(`"%[1]ss"."state" IN (?)`, client), stateNumbers)
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

const gateway = "gateway"

func (s *entitySearch) FindGateways(ctx context.Context, member *ttnpb.OrganizationOrUserIdentifiers, req *ttnpb.SearchGatewaysRequest) ([]*ttnpb.GatewayIdentifiers, error) {
	defer trace.StartRegion(ctx, "find gateways").End()
	if req.Deleted {
		ctx = WithSoftDeleted(ctx, true)
	}
	query := s.query(ctx, &Gateway{})
	query = query.Select(fmt.Sprintf(`"%[1]ss"."%[1]s_id" AS "friendly_id"`, gateway))
	query = s.queryMembership(ctx, query, gateway, member)
	query = s.queryMetaFields(ctx, query, gateway, req)
	if v := req.EuiContains; v != "" {
		query = query.Where(fmt.Sprintf(`"%[1]ss"."gateway_eui" ILIKE ?`, gateway), fmt.Sprintf("%%%s%%", v))
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

const organization = "organization"

func (s *entitySearch) FindOrganizations(ctx context.Context, member *ttnpb.OrganizationOrUserIdentifiers, req *ttnpb.SearchOrganizationsRequest) ([]*ttnpb.OrganizationIdentifiers, error) {
	defer trace.StartRegion(ctx, "find organizations").End()
	if req.Deleted {
		ctx = WithSoftDeleted(ctx, true)
	}
	query := s.query(ctx, &Organization{})
	query = query.
		Joins(`JOIN "accounts" ON "accounts"."account_type" = 'organization' AND "accounts"."account_id" = "organizations"."id"`).
		Select(`"accounts"."uid" AS "friendly_id"`)
	query = s.queryMembership(ctx, query, organization, member)
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

const user = "user"

func (s *entitySearch) FindUsers(ctx context.Context, member *ttnpb.OrganizationOrUserIdentifiers, req *ttnpb.SearchUsersRequest) ([]*ttnpb.UserIdentifiers, error) {
	defer trace.StartRegion(ctx, "find users").End()
	if req.Deleted {
		ctx = WithSoftDeleted(ctx, true)
	}
	query := s.query(ctx, &User{})
	query = query.
		Joins(`JOIN "accounts" ON "accounts"."account_type" = 'user' AND "accounts"."account_id" = "users"."id"`).
		Select(`"accounts"."uid" AS "friendly_id"`)
	query = s.queryMembership(ctx, query, user, member)
	query = s.queryMetaFields(ctx, query, user, req)
	if len(req.State) > 0 {
		stateNumbers := make([]int, len(req.State))
		for i, state := range req.State {
			stateNumbers[i] = int(state)
		}
		query = query.Where(fmt.Sprintf(`"%[1]ss"."state" IN (?)`, user), stateNumbers)
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

func (s *entitySearch) FindEndDevices(ctx context.Context, req *ttnpb.SearchEndDevicesRequest) ([]*ttnpb.EndDeviceIdentifiers, error) {
	defer trace.StartRegion(ctx, "find end devices").End()

	query := s.query(ctx, &EndDevice{}).
		Where(&EndDevice{ApplicationID: req.ApplicationId}).
		Select(`"end_devices"."device_id" AS "friendly_id"`)
	query = s.queryMetaFields(ctx, query, "end_device", req)

	if v := req.DevEuiContains; v != "" {
		query = query.Where("dev_eui ILIKE ?", fmt.Sprintf("%%%s%%", v))
	}
	if v := req.JoinEuiContains; v != "" {
		query = query.Where("join_eui ILIKE ?", fmt.Sprintf("%%%s%%", v))
	}
	// DevAddrContains

	query = query.Order(orderFromContext(ctx, "end_devices", "device_id", "ASC"))
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
	identifiers := make([]*ttnpb.EndDeviceIdentifiers, len(results))
	for i, result := range results {
		identifiers[i] = &ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: req.ApplicationId},
			DeviceId:               result.FriendlyID,
		}
	}
	return identifiers, nil
}
