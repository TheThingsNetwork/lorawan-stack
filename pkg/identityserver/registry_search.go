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

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type registrySearch struct {
	ttnpb.UnimplementedEntityRegistrySearchServer
	ttnpb.UnimplementedEndDeviceRegistrySearchServer

	*IdentityServer
}

var errSearchForbidden = errors.DefinePermissionDenied("search_forbidden", "search is forbidden")

func (rs *registrySearch) SearchApplications(
	ctx context.Context, req *ttnpb.SearchApplicationsRequest,
) (*ttnpb.Applications, error) {
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

	contactInfoInPath := ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info")
	if contactInfoInPath {
		req.FieldMask.Paths = ttnpb.ExcludeFields(req.FieldMask.Paths, "contact_info")
		req.FieldMask.Paths = append(req.FieldMask.Paths, "administrative_contact", "technical_contact")
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

	err = rs.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		entityIDs, err := st.SearchApplications(ctx, member, req)
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
			callerMemberships, err = st.FindAccountMembershipChains(ctx, member, "application", idStrings...)
			if err != nil {
				return err
			}
		}
		ctx = store.WithPagination(ctx, 0, 0, nil) // Reset pagination (already done in EntitySearch.FindApplications).
		res.Applications, err = st.FindApplications(ctx, entityIDs, req.FieldMask.GetPaths())
		if err != nil {
			return err
		}

		if contactInfoInPath {
			for _, app := range res.Applications {
				app.ContactInfo, err = getContactsFromEntity(ctx, app, st)
				if err != nil {
					return err
				}
			}
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

	err = rs.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		entityIDs, err := st.SearchClients(ctx, member, req)
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
			callerMemberships, err = st.FindAccountMembershipChains(ctx, member, "client", idStrings...)
			if err != nil {
				return err
			}
		}
		ctx = store.WithPagination(ctx, 0, 0, nil) // Reset pagination (already done in EntitySearch.FindClients).
		res.Clients, err = st.FindClients(ctx, entityIDs, req.FieldMask.GetPaths())
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
		if !entityRights.IncludesAll(ttnpb.Right_RIGHT_CLIENT_INFO) {
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

	err = rs.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		entityIDs, err := st.SearchGateways(ctx, member, req)
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
			callerMemberships, err = st.FindAccountMembershipChains(ctx, member, "gateway", idStrings...)
			if err != nil {
				return err
			}
		}
		ctx = store.WithPagination(ctx, 0, 0, nil) // Reset pagination (already done in EntitySearch.FindGateways).
		res.Gateways, err = st.FindGateways(ctx, entityIDs, req.FieldMask.GetPaths())
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

	err = rs.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		entityIDs, err := st.SearchOrganizations(ctx, member, req)
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
			callerMemberships, err = st.FindAccountMembershipChains(ctx, member, "organization", idStrings...)
			if err != nil {
				return err
			}
		}
		ctx = store.WithPagination(ctx, 0, 0, nil) // Reset pagination (already done in EntitySearch.FindOrganizations).
		res.Organizations, err = st.FindOrganizations(ctx, entityIDs, req.FieldMask.GetPaths())
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
			if !entityRights.IncludesAll(ttnpb.Right_RIGHT_CLIENT_INFO) {
				res.Organizations[i] = org.PublicSafe()
			}
		}
	}

	return res, nil
}

func (rs *registrySearch) SearchUsers(
	ctx context.Context, req *ttnpb.SearchUsersRequest,
) (*ttnpb.Users, error) {
	authInfo, err := rs.authInfo(ctx)
	if err != nil {
		return nil, err
	}
	if !authInfo.IsAdmin {
		return nil, errSearchForbidden.New()
	}

	contactInfoInPath := ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info")
	if contactInfoInPath {
		req.FieldMask.Paths = ttnpb.ExcludeFields(req.FieldMask.Paths, "contact_info")
		req.FieldMask.Paths = append(req.FieldMask.Paths, "primary_email_address", "primary_email_address_validated_at")
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

	err = rs.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		entityIDs, err := st.SearchUsers(ctx, req)
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
		res.Users, err = st.FindUsers(ctx, ids, req.FieldMask.GetPaths())
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Add ContactInfo to the response if its present in the field mask.
	if contactInfoInPath {
		for _, usr := range res.Users {
			usr.ContactInfo = append(usr.ContactInfo, &ttnpb.ContactInfo{
				ContactType:   ttnpb.ContactType_CONTACT_TYPE_OTHER,
				ContactMethod: ttnpb.ContactMethod_CONTACT_METHOD_EMAIL,
				Value:         usr.PrimaryEmailAddress,
				ValidatedAt:   usr.PrimaryEmailAddressValidatedAt,
				Public:        false,
			})
		}
	}

	return res, nil
}

func (rs *registrySearch) SearchAccounts(
	ctx context.Context, req *ttnpb.SearchAccountsRequest,
) (*ttnpb.SearchAccountsResponse, error) {
	authInfo, err := rs.authInfo(ctx)
	if err != nil {
		return nil, err
	}

	// Limit the amount of data we provide to non-admin users when searching across all accounts.
	if req.CollaboratorOf == nil && !authInfo.IsAdmin {
		// Require a search query of at least 2 characters.
		if len(req.Query) < 2 {
			return &ttnpb.SearchAccountsResponse{}, nil
		}

		// Limit the number of results to 10.
		var total uint64
		ctx = store.WithPagination(ctx, 10, 1, &total)
		defer func() {
			if err == nil {
				setTotalHeader(ctx, total)
			}
		}()
	}

	res := &ttnpb.SearchAccountsResponse{}
	err = rs.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		ids, err := st.SearchAccounts(ctx, req)
		if err != nil {
			return err
		}
		res.AccountIds = ids
		return nil
	})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (rs *registrySearch) SearchEndDevices(ctx context.Context, req *ttnpb.SearchEndDevicesRequest) (*ttnpb.EndDevices, error) {
	err := rights.RequireApplication(ctx, req.ApplicationIds, ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ)
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
	err = rs.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		ids, err := st.SearchEndDevices(ctx, req)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		ctx = store.WithPagination(ctx, 0, 0, nil) // Reset pagination (already done in EntitySearch.FindEndDevices).
		res.EndDevices, err = st.FindEndDevices(ctx, ids, req.FieldMask.GetPaths())
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
