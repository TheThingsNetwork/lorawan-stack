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

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/identityserver/blacklist"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

var (
	evtCreateOrganization = events.Define("organization.create", "create organization")
	evtUpdateOrganization = events.Define("organization.update", "update organization")
	evtDeleteOrganization = events.Define("organization.delete", "delete organization")
)

var errNestedOrganizations = errors.DefineInvalidArgument("nested_organizations", "organizations can not be nested")

func (is *IdentityServer) createOrganization(ctx context.Context, req *ttnpb.CreateOrganizationRequest) (org *ttnpb.Organization, err error) {
	if err = blacklist.Check(ctx, req.OrganizationID); err != nil {
		return nil, err
	}
	if usrIDs := req.Collaborator.GetUserIDs(); usrIDs != nil {
		if err = rights.RequireUser(ctx, *usrIDs, ttnpb.RIGHT_USER_ORGANIZATIONS_CREATE); err != nil {
			return nil, err
		}
	} else if orgIDs := req.Collaborator.GetOrganizationIDs(); orgIDs != nil {
		return nil, errNestedOrganizations
	}
	if err := validateContactInfo(req.Organization.ContactInfo); err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		org, err = store.GetOrganizationStore(db).CreateOrganization(ctx, &req.Organization)
		if err != nil {
			return err
		}
		if err = store.GetMembershipStore(db).SetMember(
			ctx,
			&req.Collaborator,
			org.OrganizationIdentifiers,
			ttnpb.RightsFrom(ttnpb.RIGHT_ALL),
		); err != nil {
			return err
		}
		if len(req.ContactInfo) > 0 {
			cleanContactInfo(req.ContactInfo)
			org.ContactInfo, err = store.GetContactInfoStore(db).SetContactInfo(ctx, org.OrganizationIdentifiers, req.ContactInfo)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtCreateOrganization(ctx, req.OrganizationIdentifiers, nil))
	is.invalidateCachedMembershipsForAccount(ctx, &req.Collaborator)
	return org, nil
}

func (is *IdentityServer) getOrganization(ctx context.Context, req *ttnpb.GetOrganizationRequest) (org *ttnpb.Organization, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	req.FieldMask.Paths = cleanFieldMaskPaths(ttnpb.OrganizationFieldPathsNested, req.FieldMask.Paths, getPaths, nil)
	if err = rights.RequireOrganization(ctx, req.OrganizationIdentifiers, ttnpb.RIGHT_ORGANIZATION_INFO); err != nil {
		if ttnpb.HasOnlyAllowedFields(req.FieldMask.Paths, ttnpb.PublicOrganizationFields...) {
			defer func() { org = org.PublicSafe() }()
		} else {
			return nil, err
		}
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		org, err = store.GetOrganizationStore(db).GetOrganization(ctx, &req.OrganizationIdentifiers, &req.FieldMask)
		if err != nil {
			return err
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "contact_info") {
			org.ContactInfo, err = store.GetContactInfoStore(db).GetContactInfo(ctx, org.OrganizationIdentifiers)
			if err != nil {
				return err
			}
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return org, nil
}

func (is *IdentityServer) listOrganizations(ctx context.Context, req *ttnpb.ListOrganizationsRequest) (orgs *ttnpb.Organizations, err error) {
	req.FieldMask.Paths = cleanFieldMaskPaths(ttnpb.OrganizationFieldPathsNested, req.FieldMask.Paths, getPaths, nil)
	var orgRights map[string]*ttnpb.Rights
	if req.Collaborator == nil {
		callerRights, _, err := is.getRights(ctx)
		if err != nil {
			return nil, err
		}
		orgRights = make(map[string]*ttnpb.Rights, len(callerRights))
		for ids, rights := range callerRights {
			if ids.EntityType() == "organization" {
				orgRights[unique.ID(ctx, ids)] = rights
			}
		}
		if len(orgRights) == 0 {
			return &ttnpb.Organizations{}, nil
		}
	}
	if usrIDs := req.Collaborator.GetUserIDs(); usrIDs != nil {
		if err = rights.RequireUser(ctx, *usrIDs, ttnpb.RIGHT_USER_ORGANIZATIONS_LIST); err != nil {
			return nil, err
		}
	} else if orgIDs := req.Collaborator.GetOrganizationIDs(); orgIDs != nil {
		return nil, errNestedOrganizations
	}
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	orgs = &ttnpb.Organizations{}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		if orgRights == nil {
			rights, err := store.GetMembershipStore(db).FindMemberRights(ctx, req.Collaborator, "organization")
			if err != nil {
				return err
			}
			orgRights = make(map[string]*ttnpb.Rights, len(rights))
			for ids, rights := range rights {
				orgRights[unique.ID(ctx, ids)] = rights
			}
		}
		if len(orgRights) == 0 {
			return nil
		}
		orgIDs := make([]*ttnpb.OrganizationIdentifiers, 0, len(orgRights))
		for uid := range orgRights {
			orgID, err := unique.ToOrganizationID(uid)
			if err != nil {
				continue
			}
			orgIDs = append(orgIDs, &orgID)
		}
		orgs.Organizations, err = store.GetOrganizationStore(db).FindOrganizations(ctx, orgIDs, &req.FieldMask)
		if err != nil {
			return err
		}
		for _, org := range orgs.Organizations {
			if !orgRights[unique.ID(ctx, org.OrganizationIdentifiers)].IncludesAll(ttnpb.RIGHT_ORGANIZATION_INFO) {
				org = org.PublicSafe()
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return orgs, nil
}

func (is *IdentityServer) updateOrganization(ctx context.Context, req *ttnpb.UpdateOrganizationRequest) (org *ttnpb.Organization, err error) {
	if err = rights.RequireOrganization(ctx, req.OrganizationIdentifiers, ttnpb.RIGHT_ORGANIZATION_SETTINGS_BASIC); err != nil {
		return nil, err
	}
	req.FieldMask.Paths = cleanFieldMaskPaths(ttnpb.OrganizationFieldPathsNested, req.FieldMask.Paths, nil, getPaths)
	if len(req.FieldMask.Paths) == 0 {
		req.FieldMask.Paths = updatePaths
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "contact_info") {
		if err := validateContactInfo(req.Organization.ContactInfo); err != nil {
			return nil, err
		}
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		org, err = store.GetOrganizationStore(db).UpdateOrganization(ctx, &req.Organization, &req.FieldMask)
		if err != nil {
			return err
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "contact_info") {
			cleanContactInfo(req.ContactInfo)
			org.ContactInfo, err = store.GetContactInfoStore(db).SetContactInfo(ctx, org.OrganizationIdentifiers, req.ContactInfo)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtUpdateOrganization(ctx, req.OrganizationIdentifiers, req.FieldMask.Paths))
	return org, nil
}

func (is *IdentityServer) deleteOrganization(ctx context.Context, ids *ttnpb.OrganizationIdentifiers) (*types.Empty, error) {
	if err := rights.RequireOrganization(ctx, *ids, ttnpb.RIGHT_ORGANIZATION_DELETE); err != nil {
		return nil, err
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		return store.GetOrganizationStore(db).DeleteOrganization(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtDeleteOrganization(ctx, ids, nil))
	return ttnpb.Empty, nil
}

type organizationRegistry struct {
	*IdentityServer
}

func (or *organizationRegistry) Create(ctx context.Context, req *ttnpb.CreateOrganizationRequest) (*ttnpb.Organization, error) {
	return or.createOrganization(ctx, req)
}
func (or *organizationRegistry) Get(ctx context.Context, req *ttnpb.GetOrganizationRequest) (*ttnpb.Organization, error) {
	return or.getOrganization(ctx, req)
}
func (or *organizationRegistry) List(ctx context.Context, req *ttnpb.ListOrganizationsRequest) (*ttnpb.Organizations, error) {
	return or.listOrganizations(ctx, req)
}
func (or *organizationRegistry) Update(ctx context.Context, req *ttnpb.UpdateOrganizationRequest) (*ttnpb.Organization, error) {
	return or.updateOrganization(ctx, req)
}
func (or *organizationRegistry) Delete(ctx context.Context, req *ttnpb.OrganizationIdentifiers) (*types.Empty, error) {
	return or.deleteOrganization(ctx, req)
}
