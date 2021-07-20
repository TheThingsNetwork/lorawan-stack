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
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type registrySearch struct {
	*IdentityServer
}

var errSearchForbidden = errors.DefinePermissionDenied("search_forbidden", "search is forbidden")

func (rs *registrySearch) memberForSearch(ctx context.Context) (*ttnpb.OrganizationOrUserIdentifiers, error) {
	authInfo, err := rs.authInfo(ctx)
	if err != nil {
		return nil, err
	}
	if authInfo.IsAdmin {
		return nil, nil
	}
	member := authInfo.GetOrganizationOrUserIdentifiers()
	if member != nil {
		return member, nil
	}
	return nil, errSearchForbidden.New()
}

func (rs *registrySearch) SearchApplications(ctx context.Context, req *ttnpb.SearchApplicationsRequest) (*ttnpb.Applications, error) {
	member, err := rs.memberForSearch(ctx)
	if err != nil {
		return nil, err
	}
	var searchFields []string
	if req.IDContains != "" {
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
	err = rs.withDatabase(ctx, func(db *gorm.DB) error {
		entityIDs, err := store.GetEntitySearch(db).FindApplications(ctx, member, req)
		if err != nil {
			return err
		}
		var ids []*ttnpb.ApplicationIdentifiers
		for _, id := range entityIDs {
			if rights.RequireApplication(ctx, *id, ttnpb.RIGHT_APPLICATION_INFO) == nil {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			return nil
		}
		ctx = store.WithPagination(ctx, 0, 0, nil) // Reset pagination (already done in EntitySearch.FindApplications).
		res.Applications, err = store.GetApplicationStore(db).FindApplications(ctx, ids, req.FieldMask)
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

func (rs *registrySearch) SearchClients(ctx context.Context, req *ttnpb.SearchClientsRequest) (*ttnpb.Clients, error) {
	member, err := rs.memberForSearch(ctx)
	if err != nil {
		return nil, err
	}
	var searchFields []string
	if req.IDContains != "" {
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
	err = rs.withDatabase(ctx, func(db *gorm.DB) error {
		entityIDs, err := store.GetEntitySearch(db).FindClients(ctx, member, req)
		if err != nil {
			return err
		}
		var ids []*ttnpb.ClientIdentifiers
		for _, id := range entityIDs {
			if rights.RequireClient(ctx, *id, ttnpb.RIGHT_CLIENT_ALL) == nil {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			return nil
		}
		ctx = store.WithPagination(ctx, 0, 0, nil) // Reset pagination (already done in EntitySearch.FindClients).
		res.Clients, err = store.GetClientStore(db).FindClients(ctx, ids, req.FieldMask)
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

func (rs *registrySearch) SearchGateways(ctx context.Context, req *ttnpb.SearchGatewaysRequest) (*ttnpb.Gateways, error) {
	member, err := rs.memberForSearch(ctx)
	if err != nil {
		return nil, err
	}
	// Backwards compatibility for frequency_plan_id field.
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "frequency_plan_id") {
		if !ttnpb.HasAnyField(req.FieldMask.GetPaths(), "frequency_plan_ids") {
			req.FieldMask.Paths = append(req.FieldMask.GetPaths(), "frequency_plan_ids")
		}
	}
	var searchFields []string
	if req.IDContains != "" {
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
	req.FieldMask = cleanFieldMaskPaths(ttnpb.GatewayFieldPathsNested, req.FieldMask, append(getPaths, searchFields...), nil)
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
	err = rs.withDatabase(ctx, func(db *gorm.DB) error {
		entityIDs, err := store.GetEntitySearch(db).FindGateways(ctx, member, req)
		if err != nil {
			return err
		}
		var ids []*ttnpb.GatewayIdentifiers
		for _, id := range entityIDs {
			if rights.RequireGateway(ctx, *id, ttnpb.RIGHT_GATEWAY_INFO) == nil {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			return nil
		}
		ctx = store.WithPagination(ctx, 0, 0, nil) // Reset pagination (already done in EntitySearch.FindGateways).
		res.Gateways, err = store.GetGatewayStore(db).FindGateways(ctx, ids, req.FieldMask)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, gtw := range res.Gateways {
		// Backwards compatibility for frequency_plan_id field.
		if len(gtw.FrequencyPlanIDs) > 0 {
			gtw.FrequencyPlanID = gtw.FrequencyPlanIDs[0]
		}
	}
	return res, nil
}

func (rs *registrySearch) SearchOrganizations(ctx context.Context, req *ttnpb.SearchOrganizationsRequest) (*ttnpb.Organizations, error) {
	member, err := rs.memberForSearch(ctx)
	if err != nil {
		return nil, err
	}
	var searchFields []string
	if req.IDContains != "" {
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
	err = rs.withDatabase(ctx, func(db *gorm.DB) error {
		entityIDs, err := store.GetEntitySearch(db).FindOrganizations(ctx, member, req)
		if err != nil {
			return err
		}
		var ids []*ttnpb.OrganizationIdentifiers
		for _, id := range entityIDs {
			if rights.RequireOrganization(ctx, *id, ttnpb.RIGHT_ORGANIZATION_INFO) == nil {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			return nil
		}
		ctx = store.WithPagination(ctx, 0, 0, nil) // Reset pagination (already done in EntitySearch.FindOrganizations).
		res.Organizations, err = store.GetOrganizationStore(db).FindOrganizations(ctx, ids, req.FieldMask)
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

func (rs *registrySearch) SearchUsers(ctx context.Context, req *ttnpb.SearchUsersRequest) (*ttnpb.Users, error) {
	member, err := rs.memberForSearch(ctx)
	if err != nil {
		return nil, err
	}
	if member != nil {
		return nil, errSearchForbidden.New()
	}
	var searchFields []string
	if req.IDContains != "" {
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
		entityIDs, err := store.GetEntitySearch(db).FindUsers(ctx, nil, req)
		if err != nil {
			return err
		}
		var ids []*ttnpb.UserIdentifiers
		for _, id := range entityIDs {
			if rights.RequireUser(ctx, *id, ttnpb.RIGHT_USER_INFO) == nil {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			return nil
		}
		ctx = store.WithPagination(ctx, 0, 0, nil) // Reset pagination (already done in EntitySearch.FindUsers).
		res.Users, err = store.GetUserStore(db).FindUsers(ctx, ids, req.FieldMask)
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
	err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_DEVICES_READ)
	if err != nil {
		return nil, err
	}
	var searchFields []string
	if req.IDContains != "" {
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
		ids, err := store.GetEntitySearch(db).FindEndDevices(ctx, req)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		ctx = store.WithPagination(ctx, 0, 0, nil) // Reset pagination (already done in EntitySearch.FindEndDevices).
		res.EndDevices, err = store.GetEndDeviceStore(db).FindEndDevices(ctx, ids, req.FieldMask)
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
