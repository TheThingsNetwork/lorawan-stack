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

package identityserver

import (
	"context"

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	gormstore "go.thethings.network/lorawan-stack/v3/pkg/identityserver/gormstore"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type registrySearch struct {
	*IdentityServer
}

var errSearchForbidden = errors.DefinePermissionDenied("search_forbidden", "search is forbidden")

func (rs *registrySearch) SearchApplications(ctx context.Context, req *ttnpb.SearchApplicationsRequest) (*ttnpb.Applications, error) {
	authInfo, err := rs.authInfo(ctx)
	if err != nil {
		return nil, err
	}
	member := authInfo.GetOrganizationOrUserIdentifiers()
	if member == nil {
		return nil, errSearchForbidden.New()
	}
	if authInfo.IsAdmin {
		member = nil
	}

	var searchFields []string
	if req.IdContains != "" {
		searchFields = append(searchFields, "ids")
	}
	if req.NameContains != "" {
		searchFields = append(searchFields, "name")
	}
	if req.DescriptionContains != "" {
		searchFields = append(searchFields, "description")
	}
	if len(req.AttributesContain) > 0 {
		searchFields = append(searchFields, "attributes")
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.ApplicationFieldPathsNested, req.FieldMask, append(getPaths, searchFields...), nil)
	if req.Deleted {
		ctx = store.WithSoftDeleted(ctx, true)
	}

	ctx = store.WithOrder(ctx, req.Order)
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()

	res := &ttnpb.Applications{}
	var callerMemberships store.MembershipChains

	err = rs.withDatabase(ctx, func(db *gorm.DB) error {
		entityIDs, err := gormstore.GetEntitySearch(db).SearchApplications(ctx, member, req)
		if err != nil {
			return err
		}
		if len(entityIDs) == 0 {
			return nil
		}
		if member != nil {
			idStrings := make([]string, len(entityIDs))
			for i, entityID := range entityIDs {
				idStrings[i] = entityID.IDString()
			}
			callerMemberships, err = rs.getMembershipStore(ctx, db).FindAccountMembershipChains(ctx, member, "application", idStrings...)
			if err != nil {
				return err
			}
		}
		ctx = store.WithPagination(ctx, 0, 0, nil) // Reset pagination (already done in EntitySearch.FindApplications).
		res.Applications, err = gormstore.GetApplicationStore(db).FindApplications(ctx, entityIDs, req.FieldMask)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	for i, app := range res.Applications {
		entityRights := callerMemberships.GetRights(member, app.GetIds()).Union(authInfo.GetUniversalRights())
		if !entityRights.IncludesAll(ttnpb.Right_RIGHT_APPLICATION_INFO) {
			res.Applications[i] = app.PublicSafe()
		}
	}

	return res, nil
}

func (rs *registrySearch) SearchClients(ctx context.Context, req *ttnpb.SearchClientsRequest) (*ttnpb.Clients, error) {
	authInfo, err := rs.authInfo(ctx)
	if err != nil {
		return nil, err
	}
	member := authInfo.GetOrganizationOrUserIdentifiers()
	if member == nil {
		return nil, errSearchForbidden.New()
	}
	if authInfo.IsAdmin {
		member = nil
	}

	var searchFields []string
	if req.IdContains != "" {
		searchFields = append(searchFields, "ids")
	}
	if req.NameContains != "" {
		searchFields = append(searchFields, "name")
	}
	if req.DescriptionContains != "" {
		searchFields = append(searchFields, "description")
	}
	if len(req.AttributesContain) > 0 {
		searchFields = append(searchFields, "attributes")
	}
	if len(req.State) > 0 {
		searchFields = append(searchFields, "state")
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.ClientFieldPathsNested, req.FieldMask, append(getPaths, searchFields...), nil)
	if req.Deleted {
		ctx = store.WithSoftDeleted(ctx, true)
	}

	ctx = store.WithOrder(ctx, req.Order)
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()

	res := &ttnpb.Clients{}
	var callerMemberships store.MembershipChains

	err = rs.withDatabase(ctx, func(db *gorm.DB) error {
		entityIDs, err := gormstore.GetEntitySearch(db).SearchClients(ctx, member, req)
		if err != nil {
			return err
		}
		if len(entityIDs) == 0 {
			return nil
		}
		if member != nil {
			idStrings := make([]string, len(entityIDs))
			for i, entityID := range entityIDs {
				idStrings[i] = entityID.IDString()
			}
			callerMemberships, err = rs.getMembershipStore(ctx, db).FindAccountMembershipChains(ctx, member, "client", idStrings...)
			if err != nil {
				return err
			}
		}
		ctx = store.WithPagination(ctx, 0, 0, nil) // Reset pagination (already done in EntitySearch.FindClients).
		res.Clients, err = gormstore.GetClientStore(db).FindClients(ctx, entityIDs, req.FieldMask)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	for i, cli := range res.Clients {
		entityRights := callerMemberships.GetRights(member, cli.GetIds()).Union(authInfo.GetUniversalRights())
		if !entityRights.IncludesAll(ttnpb.Right_RIGHT_CLIENT_ALL) {
			res.Clients[i] = cli.PublicSafe()
		}
	}

	return res, nil
}

func (rs *registrySearch) SearchGateways(ctx context.Context, req *ttnpb.SearchGatewaysRequest) (*ttnpb.Gateways, error) {
	authInfo, err := rs.authInfo(ctx)
	if err != nil {
		return nil, err
	}
	member := authInfo.GetOrganizationOrUserIdentifiers()
	if member == nil {
		return nil, errSearchForbidden.New()
	}
	if authInfo.IsAdmin {
		member = nil
	}

	// Backwards compatibility for frequency_plan_id field.
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "frequency_plan_id") {
		if !ttnpb.HasAnyField(req.FieldMask.GetPaths(), "frequency_plan_ids") {
			req.FieldMask.Paths = append(req.FieldMask.GetPaths(), "frequency_plan_ids")
		}
	}
	var searchFields []string
	if req.IdContains != "" {
		searchFields = append(searchFields, "ids")
	}
	if req.NameContains != "" {
		searchFields = append(searchFields, "name")
	}
	if req.DescriptionContains != "" {
		searchFields = append(searchFields, "description")
	}
	if len(req.AttributesContain) > 0 {
		searchFields = append(searchFields, "attributes")
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.GatewayFieldPathsNested, req.FieldMask, append(getPaths, searchFields...), []string{"frequency_plan_id"})
	if req.Deleted {
		ctx = store.WithSoftDeleted(ctx, true)
	}

	ctx = store.WithOrder(ctx, req.Order)
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()

	res := &ttnpb.Gateways{}
	var callerMemberships store.MembershipChains

	err = rs.withDatabase(ctx, func(db *gorm.DB) error {
		entityIDs, err := gormstore.GetEntitySearch(db).SearchGateways(ctx, member, req)
		if err != nil {
			return err
		}
		if len(entityIDs) == 0 {
			return nil
		}
		if member != nil {
			idStrings := make([]string, len(entityIDs))
			for i, entityID := range entityIDs {
				idStrings[i] = entityID.IDString()
			}
			callerMemberships, err = rs.getMembershipStore(ctx, db).FindAccountMembershipChains(ctx, member, "gateway", idStrings...)
			if err != nil {
				return err
			}
		}
		ctx = store.WithPagination(ctx, 0, 0, nil) // Reset pagination (already done in EntitySearch.FindGateways).
		res.Gateways, err = gormstore.GetGatewayStore(db).FindGateways(ctx, entityIDs, req.FieldMask)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if member != nil {
		for i, gtw := range res.Gateways {
			entityRights := callerMemberships.GetRights(member, gtw.GetIds()).Union(authInfo.GetUniversalRights())
			if !entityRights.IncludesAll(ttnpb.Right_RIGHT_GATEWAY_INFO) {
				res.Gateways[i] = gtw.PublicSafe()
			}
		}
	}

	for _, gtw := range res.Gateways {
		// Backwards compatibility for frequency_plan_id field.
		if len(gtw.FrequencyPlanIds) > 0 {
			gtw.FrequencyPlanId = gtw.FrequencyPlanIds[0]
		}
	}

	return res, nil
}

func (rs *registrySearch) SearchOrganizations(ctx context.Context, req *ttnpb.SearchOrganizationsRequest) (*ttnpb.Organizations, error) {
	authInfo, err := rs.authInfo(ctx)
	if err != nil {
		return nil, err
	}
	member := authInfo.GetOrganizationOrUserIdentifiers()
	if member == nil {
		return nil, errSearchForbidden.New()
	}
	if authInfo.IsAdmin {
		member = nil
	}

	var searchFields []string
	if req.IdContains != "" {
		searchFields = append(searchFields, "ids")
	}
	if req.NameContains != "" {
		searchFields = append(searchFields, "name")
	}
	if req.DescriptionContains != "" {
		searchFields = append(searchFields, "description")
	}
	if len(req.AttributesContain) > 0 {
		searchFields = append(searchFields, "attributes")
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.OrganizationFieldPathsNested, req.FieldMask, append(getPaths, searchFields...), nil)
	if req.Deleted {
		ctx = store.WithSoftDeleted(ctx, true)
	}

	ctx = store.WithOrder(ctx, req.Order)
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()

	res := &ttnpb.Organizations{}
	var callerMemberships store.MembershipChains

	err = rs.withDatabase(ctx, func(db *gorm.DB) error {
		entityIDs, err := gormstore.GetEntitySearch(db).SearchOrganizations(ctx, member, req)
		if err != nil {
			return err
		}
		if len(entityIDs) == 0 {
			return nil
		}
		if member != nil {
			idStrings := make([]string, len(entityIDs))
			for i, entityID := range entityIDs {
				idStrings[i] = entityID.IDString()
			}
			callerMemberships, err = rs.getMembershipStore(ctx, db).FindAccountMembershipChains(ctx, member, "organization", idStrings...)
			if err != nil {
				return err
			}
		}
		ctx = store.WithPagination(ctx, 0, 0, nil) // Reset pagination (already done in EntitySearch.FindOrganizations).
		res.Organizations, err = gormstore.GetOrganizationStore(db).FindOrganizations(ctx, entityIDs, req.FieldMask)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if member != nil {
		for i, org := range res.Organizations {
			entityRights := callerMemberships.GetRights(member, org.GetIds()).Union(authInfo.GetUniversalRights())
			if !entityRights.IncludesAll(ttnpb.Right_RIGHT_CLIENT_ALL) {
				res.Organizations[i] = org.PublicSafe()
			}
		}
	}

	return res, nil
}

func (rs *registrySearch) SearchUsers(ctx context.Context, req *ttnpb.SearchUsersRequest) (*ttnpb.Users, error) {
	authInfo, err := rs.authInfo(ctx)
	if err != nil {
		return nil, err
	}
	if !authInfo.IsAdmin {
		return nil, errSearchForbidden.New()
	}

	var searchFields []string
	if req.IdContains != "" {
		searchFields = append(searchFields, "ids")
	}
	if req.NameContains != "" {
		searchFields = append(searchFields, "name")
	}
	if req.DescriptionContains != "" {
		searchFields = append(searchFields, "description")
	}
	if len(req.AttributesContain) > 0 {
		searchFields = append(searchFields, "attributes")
	}
	if len(req.State) > 0 {
		searchFields = append(searchFields, "state")
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.UserFieldPathsNested, req.FieldMask, append(getPaths, searchFields...), nil)
	if req.Deleted {
		ctx = store.WithSoftDeleted(ctx, true)
	}

	ctx = store.WithOrder(ctx, req.Order)
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()

	res := &ttnpb.Users{}

	err = rs.withDatabase(ctx, func(db *gorm.DB) error {
		entityIDs, err := gormstore.GetEntitySearch(db).SearchUsers(ctx, req)
		if err != nil {
			return err
		}
		if len(entityIDs) == 0 {
			return nil
		}
		var ids []*ttnpb.UserIdentifiers
		for _, id := range entityIDs {
			ids = append(ids, id)
		}
		if len(ids) == 0 {
			return nil
		}
		ctx = store.WithPagination(ctx, 0, 0, nil) // Reset pagination (already done in EntitySearch.FindUsers).
		res.Users, err = gormstore.GetUserStore(db).FindUsers(ctx, ids, req.FieldMask)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (rs *registrySearch) SearchEndDevices(ctx context.Context, req *ttnpb.SearchEndDevicesRequest) (*ttnpb.EndDevices, error) {
	err := rights.RequireApplication(ctx, *req.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ)
	if err != nil {
		return nil, err
	}
	var searchFields []string
	if req.IdContains != "" {
		searchFields = append(searchFields, "ids")
	}
	if req.NameContains != "" {
		searchFields = append(searchFields, "name")
	}
	if req.DescriptionContains != "" {
		searchFields = append(searchFields, "description")
	}
	if len(req.AttributesContain) > 0 {
		searchFields = append(searchFields, "attributes")
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.EndDeviceFieldPathsNested, req.FieldMask, append(getPaths, searchFields...), nil)

	ctx = store.WithOrder(ctx, req.Order)
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()

	res := &ttnpb.EndDevices{}
	err = rs.withDatabase(ctx, func(db *gorm.DB) error {
		ids, err := gormstore.GetEntitySearch(db).SearchEndDevices(ctx, req)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		ctx = store.WithPagination(ctx, 0, 0, nil) // Reset pagination (already done in EntitySearch.FindEndDevices).
		res.EndDevices, err = gormstore.GetEndDeviceStore(db).FindEndDevices(ctx, ids, req.FieldMask)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
