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
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	storeutil "go.thethings.network/lorawan-stack/v3/pkg/util/store"
)

type entitySearch struct {
	*membershipStore
}

func newEntitySearch(baseStore *baseStore) *entitySearch {
	return &entitySearch{
		membershipStore: newMembershipStore(baseStore),
	}
}

type metaFields interface {
	GetIdContains() string
	GetNameContains() string
	GetDescriptionContains() string
	GetAttributesContain() map[string]string
}

func ilike(field string) string {
	return `?TableAlias."` + field + `" ILIKE '%' || ? || '%'`
}

func (s *entitySearch) ftsQuery(field string) string {
	if s.server == "PostgreSQL" && s.major >= 11 {
		return fmt.Sprintf(`to_tsvector('english', ?TableAlias."%s") @@ websearch_to_tsquery('english', ?)`, field)
	}
	return ilike(field)
}

func (s *entitySearch) queryStringQuery(
	queryString string, fields ...string,
) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		return q.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			q = q.Where(ilike(fields[0]), queryString)
			for _, field := range fields[1:] {
				if field == "description" {
					q = q.WhereOr(s.ftsQuery(field), queryString)
				} else {
					q = q.WhereOr(ilike(field), queryString)
				}
			}
			return q
		})
	}
}

func (s *entitySearch) selectWithMetaFields(
	ctx context.Context, entityType string, req metaFields,
) func(*bun.SelectQuery) *bun.SelectQuery {
	return func(q *bun.SelectQuery) *bun.SelectQuery {
		if v := req.GetIdContains(); v != "" {
			switch entityType {
			case "organization", "user":
				q = q.Where(ilike("account_uid"), v)
			case "end_device":
				q = q.Where(ilike("device_id"), v)
			default:
				q = q.Where(ilike(fmt.Sprintf("%s_id", entityType)), v)
			}
		}
		if v := req.GetNameContains(); v != "" {
			q = q.Where(ilike("name"), v)
		}
		if v := req.GetDescriptionContains(); v != "" {
			q = q.Where(s.ftsQuery("description"), v)
		}
		if kv := req.GetAttributesContain(); len(kv) > 0 {
			attrQuery := s.newSelectModel(ctx, &Attribute{}).
				Column("entity_id")
			switch entityType {
			case "end_device":
				attrQuery = attrQuery.Where(`"entity_type" = ?`, "device")
			default:
				attrQuery = attrQuery.Where(`"entity_type" = ?`, entityType)
			}
			for k, v := range kv {
				attrQuery = attrQuery.Where(`"key" = ?`, k).Where(ilike("value"), v)
			}
			q = q.Where(`"id" IN (?)`, attrQuery)
		}
		return q
	}
}

func getIDs[A interface{ GetIds() B }, B ttnpb.IDStringer](in []A, mods ...func(B) B) []B {
	out := make([]B, len(in))
	for i, v := range in {
		out[i] = v.GetIds()
		for _, mod := range mods {
			out[i] = mod(out[i])
		}
	}
	return out
}

func (s *entitySearch) SearchApplications(
	ctx context.Context, accountID *ttnpb.OrganizationOrUserIdentifiers, req *ttnpb.SearchApplicationsRequest,
) ([]*ttnpb.ApplicationIdentifiers, error) {
	ctx, span := tracer.Start(ctx, "SearchApplications")
	defer span.End()

	var selectors []func(*bun.SelectQuery) *bun.SelectQuery

	if accountID != nil {
		span.SetAttributes(
			attribute.String("member_type", accountID.EntityType()),
			attribute.String("member_id", accountID.IDString()),
		)
		selectWithUUID, err := s.selectWithUUIDsInMemberships(
			ctx,
			accountID,
			"application",
			accountID.EntityType() == "user",
		)
		if err != nil {
			return nil, err
		}
		selectors = append(selectors, selectWithUUID)
	}

	if queryString := req.GetQuery(); queryString != "" {
		selectors = append(selectors, s.queryStringQuery(
			queryString, "application_id", "name", "description",
		))
	}

	selectors = append(selectors, s.selectWithMetaFields(ctx, "application", req))

	pbs, err := s.listApplicationsBy(ctx, combineApply(selectors...), store.FieldMask{"ids"})
	if err != nil {
		return nil, err
	}

	return getIDs[*ttnpb.Application, *ttnpb.ApplicationIdentifiers](pbs), nil
}

func (s *entitySearch) SearchClients(
	ctx context.Context, accountID *ttnpb.OrganizationOrUserIdentifiers, req *ttnpb.SearchClientsRequest,
) ([]*ttnpb.ClientIdentifiers, error) {
	ctx, span := tracer.Start(ctx, "SearchClients")
	defer span.End()

	var selectors []func(*bun.SelectQuery) *bun.SelectQuery

	if accountID != nil {
		span.SetAttributes(
			attribute.String("member_type", accountID.EntityType()),
			attribute.String("member_id", accountID.IDString()),
		)
		selectWithUUID, err := s.selectWithUUIDsInMemberships(ctx, accountID, "client", accountID.EntityType() == "user")
		if err != nil {
			return nil, err
		}
		selectors = append(selectors, selectWithUUID)
	}

	if queryString := req.GetQuery(); queryString != "" {
		selectors = append(selectors, s.queryStringQuery(
			queryString, "client_id", "name", "description",
		))
	}

	selectors = append(selectors, s.selectWithMetaFields(ctx, "client", req))

	if len(req.State) > 0 {
		selectors = append(selectors, func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where(`"state" IN (?)`, bun.In(convertIntSlice[ttnpb.State, int](req.State)))
		})
	}

	pbs, err := s.listClientsBy(ctx, combineApply(selectors...), store.FieldMask{"ids"})
	if err != nil {
		return nil, err
	}

	return getIDs[*ttnpb.Client, *ttnpb.ClientIdentifiers](pbs), nil
}

func (s *entitySearch) SearchEndDevices(
	ctx context.Context, req *ttnpb.SearchEndDevicesRequest,
) ([]*ttnpb.EndDeviceIdentifiers, error) {
	ctx, span := tracer.Start(ctx, "SearchEndDevices")
	defer span.End()

	var selectors []func(*bun.SelectQuery) *bun.SelectQuery

	if req.GetApplicationIds() != nil {
		span.SetAttributes(
			attribute.String("application_id", req.GetApplicationIds().GetApplicationId()),
		)
		selectors = append(selectors, s.endDeviceStore.selectWithID(ctx, req.GetApplicationIds().GetApplicationId()))
	}

	if queryString := req.GetQuery(); queryString != "" {
		selectors = append(selectors, s.queryStringQuery(
			queryString, "device_id", "dev_eui", "join_eui", "name", "description",
		))
	}

	selectors = append(selectors, s.selectWithMetaFields(ctx, "end_device", req))

	if v := req.DevEuiContains; v != "" {
		selectors = append(selectors, func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where(ilike("dev_eui"), v)
		})
	}
	if v := req.JoinEuiContains; v != "" {
		selectors = append(selectors, func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where(ilike("join_eui"), v)
		})
	}

	pbs, err := s.listEndDevicesBy(ctx, combineApply(selectors...), store.FieldMask{"ids"})
	if err != nil {
		return nil, err
	}

	return getIDs(pbs, func(ids *ttnpb.EndDeviceIdentifiers) *ttnpb.EndDeviceIdentifiers {
		return &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: ids.ApplicationIds,
			DeviceId:       ids.DeviceId,
		}
	}), nil
}

func (s *entitySearch) SearchGateways(
	ctx context.Context, accountID *ttnpb.OrganizationOrUserIdentifiers, req *ttnpb.SearchGatewaysRequest,
) ([]*ttnpb.GatewayIdentifiers, error) {
	ctx, span := tracer.Start(ctx, "SearchGateways")
	defer span.End()

	var selectors []func(*bun.SelectQuery) *bun.SelectQuery

	if accountID != nil {
		span.SetAttributes(
			attribute.String("member_type", accountID.EntityType()),
			attribute.String("member_id", accountID.IDString()),
		)
		selectWithUUID, err := s.selectWithUUIDsInMemberships(ctx, accountID, "gateway", accountID.EntityType() == "user")
		if err != nil {
			return nil, err
		}
		selectors = append(selectors, selectWithUUID)
	}

	if queryString := req.GetQuery(); queryString != "" {
		selectors = append(selectors, s.queryStringQuery(
			queryString, "gateway_id", "gateway_eui", "name", "description",
		))
	}

	selectors = append(selectors, s.selectWithMetaFields(ctx, "gateway", req))

	if v := req.EuiContains; v != "" {
		selectors = append(selectors, func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where(ilike("gateway_eui"), v)
		})
	}

	pbs, err := s.listGatewaysBy(ctx, combineApply(selectors...), store.FieldMask{"ids"})
	if err != nil {
		return nil, err
	}

	return getIDs(pbs, func(ids *ttnpb.GatewayIdentifiers) *ttnpb.GatewayIdentifiers {
		return &ttnpb.GatewayIdentifiers{
			GatewayId: ids.GatewayId,
		}
	}), nil
}

func (s *entitySearch) SearchOrganizations(
	ctx context.Context, accountID *ttnpb.OrganizationOrUserIdentifiers, req *ttnpb.SearchOrganizationsRequest,
) ([]*ttnpb.OrganizationIdentifiers, error) {
	ctx, span := tracer.Start(ctx, "SearchOrganizations")
	defer span.End()

	var selectors []func(*bun.SelectQuery) *bun.SelectQuery

	if accountID != nil {
		span.SetAttributes(
			attribute.String("member_type", accountID.EntityType()),
			attribute.String("member_id", accountID.IDString()),
		)
		selectWithUUID, err := s.selectWithUUIDsInMemberships(ctx, accountID, "organization", false)
		if err != nil {
			return nil, err
		}
		selectors = append(selectors, selectWithUUID)
	}

	if queryString := req.GetQuery(); queryString != "" {
		selectors = append(selectors, s.queryStringQuery(
			queryString, "account_uid", "name", "description",
		))
	}

	selectors = append(selectors, s.selectWithMetaFields(ctx, "organization", req))

	pbs, err := s.listOrganizationsBy(ctx, combineApply(selectors...), store.FieldMask{"ids"})
	if err != nil {
		return nil, err
	}

	return getIDs[*ttnpb.Organization, *ttnpb.OrganizationIdentifiers](pbs), nil
}

func (s *entitySearch) SearchUsers(
	ctx context.Context, req *ttnpb.SearchUsersRequest,
) ([]*ttnpb.UserIdentifiers, error) {
	ctx, span := tracer.Start(ctx, "SearchUsers")
	defer span.End()

	var selectors []func(*bun.SelectQuery) *bun.SelectQuery

	if queryString := req.GetQuery(); queryString != "" {
		selectors = append(selectors, s.queryStringQuery(
			queryString, "account_uid", "name", "description",
		))
	}

	selectors = append(selectors, s.selectWithMetaFields(ctx, "user", req))

	if len(req.State) > 0 {
		selectors = append(selectors, func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.Where(`"state" IN (?)`, bun.In(convertIntSlice[ttnpb.State, int](req.State)))
		})
	}

	pbs, err := s.listUsersBy(ctx, combineApply(selectors...), store.FieldMask{"ids"})
	if err != nil {
		return nil, err
	}

	return getIDs[*ttnpb.User, *ttnpb.UserIdentifiers](pbs), nil
}

func (s *entitySearch) SearchAccounts(
	ctx context.Context, req *ttnpb.SearchAccountsRequest,
) ([]*ttnpb.OrganizationOrUserIdentifiers, error) {
	ctx, span := tracer.Start(ctx, "SearchAccounts")
	defer span.End()

	var selectQuery *bun.SelectQuery

	if entityID := req.GetEntityIdentifiers(); entityID != nil {
		entityType, entityUUID, err := s.getEntity(ctx, entityID)
		if err != nil {
			return nil, err
		}

		selectQuery = s.newSelectModel(ctx, &directEntityMembership{}).
			ColumnExpr("account_type, account_friendly_id").
			Where("entity_type = ?", entityType).
			Where("entity_id = ?", entityUUID).
			Order("account_friendly_id")
		if req.OnlyUsers {
			selectQuery = selectQuery.Where(`account_type = 'user'`)
		}
		if q := req.GetQuery(); q != "" {
			selectQuery = selectQuery.Where(ilike("account_friendly_id"), q)
		}
	} else {
		selectQuery = s.newSelectModel(ctx, &Account{}).
			ColumnExpr("account_type, uid AS account_friendly_id").
			Order("uid")
		if req.OnlyUsers {
			selectQuery = selectQuery.Where(`account_type = 'user'`)
		}
		if q := req.GetQuery(); q != "" {
			selectQuery = selectQuery.Where(ilike("uid"), q)
		}
	}

	// Count the total number of results.
	count, err := selectQuery.Count(ctx)
	if err != nil {
		return nil, storeutil.WrapDriverError(err)
	}
	store.SetTotal(ctx, uint64(count))

	// Apply paging.
	selectQuery = selectQuery.
		Apply(selectWithLimitAndOffsetFromContext(ctx))

	var results []struct {
		AccountType       string
		AccountFriendlyID string
	}

	// Scan the results.
	err = selectQuery.Scan(ctx, &results)
	if err != nil {
		return nil, storeutil.WrapDriverError(err)
	}

	// Convert the results to protobuf.
	identifiers := make([]*ttnpb.OrganizationOrUserIdentifiers, len(results))

	for i, m := range results {
		switch m.AccountType {
		case "organization":
			identifiers[i] = (&ttnpb.OrganizationIdentifiers{
				OrganizationId: m.AccountFriendlyID,
			}).GetOrganizationOrUserIdentifiers()
		case "user":
			identifiers[i] = (&ttnpb.UserIdentifiers{
				UserId: m.AccountFriendlyID,
			}).GetOrganizationOrUserIdentifiers()
		}
	}

	return identifiers, nil
}
