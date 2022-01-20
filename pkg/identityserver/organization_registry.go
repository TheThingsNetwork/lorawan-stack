// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/blacklist"
	gormstore "go.thethings.network/lorawan-stack/v3/pkg/identityserver/gormstore"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	evtCreateOrganization = events.Define(
		"organization.create", "create organization",
		events.WithVisibility(ttnpb.Right_RIGHT_ORGANIZATION_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUpdateOrganization = events.Define(
		"organization.update", "update organization",
		events.WithVisibility(ttnpb.Right_RIGHT_ORGANIZATION_INFO),
		events.WithUpdatedFieldsDataType(),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteOrganization = events.Define(
		"organization.delete", "delete organization",
		events.WithVisibility(ttnpb.Right_RIGHT_ORGANIZATION_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtRestoreOrganization = events.Define(
		"organization.restore", "restore organization",
		events.WithVisibility(ttnpb.Right_RIGHT_ORGANIZATION_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtPurgeOrganization = events.Define(
		"organization.purge", "purge organization",
		events.WithVisibility(ttnpb.Right_RIGHT_ORGANIZATION_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
)

var (
	errNestedOrganizations       = errors.DefineInvalidArgument("nested_organizations", "organizations can not be nested")
	errAdminsCreateOrganizations = errors.DefinePermissionDenied("admins_create_organizations", "organizations may only be created by admins")
	errAdminsPurgeOrganizations  = errors.DefinePermissionDenied("admins_purge_organizations", "organizations may only be purged by admins")
)

func (is *IdentityServer) createOrganization(ctx context.Context, req *ttnpb.CreateOrganizationRequest) (org *ttnpb.Organization, err error) {
	if err = blacklist.Check(ctx, req.Organization.GetIds().GetOrganizationId()); err != nil {
		return nil, err
	}
	if usrIDs := req.GetCollaborator().GetUserIds(); usrIDs != nil {
		if !is.IsAdmin(ctx) && !is.configFromContext(ctx).UserRights.CreateOrganizations {
			return nil, errAdminsCreateOrganizations.New()
		}
		if err = rights.RequireUser(ctx, *usrIDs, ttnpb.Right_RIGHT_USER_ORGANIZATIONS_CREATE); err != nil {
			return nil, err
		}
	} else if orgIDs := req.GetCollaborator().GetOrganizationIds(); orgIDs != nil {
		return nil, errNestedOrganizations.New()
	}

	if req.Organization.AdministrativeContact == nil {
		req.Organization.AdministrativeContact = req.Collaborator
	} else if err := validateCollaboratorEqualsContact(req.Collaborator, req.Organization.AdministrativeContact); err != nil {
		return nil, err
	}
	if req.Organization.TechnicalContact == nil {
		req.Organization.TechnicalContact = req.Collaborator
	} else if err := validateCollaboratorEqualsContact(req.Collaborator, req.Organization.TechnicalContact); err != nil {
		return nil, err
	}
	if err := validateContactInfo(req.Organization.ContactInfo); err != nil {
		return nil, err
	}

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		org, err = gormstore.GetOrganizationStore(db).CreateOrganization(ctx, req.Organization)
		if err != nil {
			return err
		}
		if err = is.getMembershipStore(ctx, db).SetMember(
			ctx,
			req.GetCollaborator(),
			org.GetIds().GetEntityIdentifiers(),
			ttnpb.RightsFrom(ttnpb.Right_RIGHT_ALL),
		); err != nil {
			return err
		}
		if len(req.Organization.ContactInfo) > 0 {
			cleanContactInfo(req.Organization.ContactInfo)
			org.ContactInfo, err = gormstore.GetContactInfoStore(db).SetContactInfo(ctx, org.GetIds(), req.Organization.ContactInfo)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtCreateOrganization.NewWithIdentifiersAndData(ctx, req.Organization.GetIds(), nil))
	return org, nil
}

func (is *IdentityServer) getOrganization(ctx context.Context, req *ttnpb.GetOrganizationRequest) (org *ttnpb.Organization, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.OrganizationFieldPathsNested, req.FieldMask, getPaths, nil)
	if err = rights.RequireOrganization(ctx, *req.GetOrganizationIds(), ttnpb.Right_RIGHT_ORGANIZATION_INFO); err != nil {
		if !ttnpb.HasOnlyAllowedFields(req.FieldMask.GetPaths(), ttnpb.PublicOrganizationFields...) {
			return nil, err
		}
		defer func() { org = org.PublicSafe() }()
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		org, err = gormstore.GetOrganizationStore(db).GetOrganization(ctx, req.GetOrganizationIds(), req.FieldMask)
		if err != nil {
			return err
		}
		if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info") {
			org.ContactInfo, err = gormstore.GetContactInfoStore(db).GetContactInfo(ctx, org.GetIds())
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
	req.FieldMask = cleanFieldMaskPaths(ttnpb.OrganizationFieldPathsNested, req.FieldMask, getPaths, nil)

	authInfo, err := is.authInfo(ctx)
	if err != nil {
		return nil, err
	}
	callerAccountID := authInfo.GetOrganizationOrUserIdentifiers()
	if req.Collaborator == nil {
		req.Collaborator = callerAccountID
	}
	if req.Collaborator == nil {
		return &ttnpb.Organizations{}, nil
	}

	if usrIDs := req.Collaborator.GetUserIds(); usrIDs != nil {
		if err = rights.RequireUser(ctx, *usrIDs, ttnpb.Right_RIGHT_USER_ORGANIZATIONS_LIST); err != nil {
			return nil, err
		}
	} else if orgIDs := req.Collaborator.GetOrganizationIds(); orgIDs != nil {
		return nil, errNestedOrganizations.New()
	}

	if req.Deleted {
		ctx = store.WithSoftDeleted(ctx, true)
	}

	ctx = store.WithOrder(ctx, req.Order)
	var total uint64
	paginateCtx := store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()

	orgs = &ttnpb.Organizations{}
	var callerMemberships store.MembershipChains

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		membershipStore := is.getMembershipStore(ctx, db)
		ids, err := membershipStore.FindMemberships(paginateCtx, req.Collaborator, "organization", false)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		callerMemberships, err = membershipStore.FindAccountMembershipChains(ctx, callerAccountID, "organization", idStrings(ids...)...)
		if err != nil {
			return err
		}
		orgIDs := make([]*ttnpb.OrganizationIdentifiers, 0, len(ids))
		for _, id := range ids {
			if orgID := id.GetEntityIdentifiers().GetOrganizationIds(); orgID != nil {
				orgIDs = append(orgIDs, orgID)
			}
		}
		orgs.Organizations, err = gormstore.GetOrganizationStore(db).FindOrganizations(ctx, orgIDs, req.FieldMask)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	for i, org := range orgs.Organizations {
		entityRights := callerMemberships.GetRights(callerAccountID, org.GetIds()).Union(authInfo.GetUniversalRights())
		if !entityRights.IncludesAll(ttnpb.Right_RIGHT_ORGANIZATION_INFO) {
			orgs.Organizations[i] = org.PublicSafe()
		}
	}

	return orgs, nil
}

func (is *IdentityServer) updateOrganization(ctx context.Context, req *ttnpb.UpdateOrganizationRequest) (org *ttnpb.Organization, err error) {
	if err = rights.RequireOrganization(ctx, *req.Organization.GetIds(), ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_BASIC); err != nil {
		return nil, err
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.OrganizationFieldPathsNested, req.FieldMask, nil, getPaths)
	if len(req.FieldMask.GetPaths()) == 0 {
		req.FieldMask = &pbtypes.FieldMask{Paths: updatePaths}
	}
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info") {
		if err := validateContactInfo(req.Organization.ContactInfo); err != nil {
			return nil, err
		}
	}
	req.FieldMask.Paths = ttnpb.FlattenPaths(req.FieldMask.Paths, []string{"administrative_contact", "technical_contact"})
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		if err := validateContactIsCollaborator(ctx, db, req.Organization.AdministrativeContact, req.Organization.GetEntityIdentifiers()); err != nil {
			return err
		}
		if err := validateContactIsCollaborator(ctx, db, req.Organization.TechnicalContact, req.Organization.GetEntityIdentifiers()); err != nil {
			return err
		}
		org, err = gormstore.GetOrganizationStore(db).UpdateOrganization(ctx, req.Organization, req.FieldMask)
		if err != nil {
			return err
		}
		if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info") {
			cleanContactInfo(req.Organization.ContactInfo)
			org.ContactInfo, err = gormstore.GetContactInfoStore(db).SetContactInfo(ctx, org.Ids, req.Organization.ContactInfo)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtUpdateOrganization.NewWithIdentifiersAndData(ctx, req.Organization.GetIds(), req.FieldMask.GetPaths()))
	return org, nil
}

func (is *IdentityServer) deleteOrganization(ctx context.Context, ids *ttnpb.OrganizationIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireOrganization(ctx, *ids, ttnpb.Right_RIGHT_ORGANIZATION_DELETE); err != nil {
		return nil, err
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		return gormstore.GetOrganizationStore(db).DeleteOrganization(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtDeleteOrganization.NewWithIdentifiersAndData(ctx, ids, nil))
	return ttnpb.Empty, nil
}

func (is *IdentityServer) restoreOrganization(ctx context.Context, ids *ttnpb.OrganizationIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireOrganization(store.WithSoftDeleted(ctx, false), *ids, ttnpb.Right_RIGHT_ORGANIZATION_DELETE); err != nil {
		return nil, err
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		orgStore := gormstore.GetOrganizationStore(db)
		org, err := orgStore.GetOrganization(store.WithSoftDeleted(ctx, true), ids, softDeleteFieldMask)
		if err != nil {
			return err
		}
		deletedAt := ttnpb.StdTime(org.DeletedAt)
		if deletedAt == nil {
			panic("store.WithSoftDeleted(ctx, true) returned result that is not deleted")
		}
		if time.Since(*deletedAt) > is.configFromContext(ctx).Delete.Restore {
			return errRestoreWindowExpired.New()
		}
		return orgStore.RestoreOrganization(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtRestoreOrganization.NewWithIdentifiersAndData(ctx, ids, nil))
	return ttnpb.Empty, nil
}

func (is *IdentityServer) purgeOrganization(ctx context.Context, ids *ttnpb.OrganizationIdentifiers) (*pbtypes.Empty, error) {
	if !is.IsAdmin(ctx) {
		return nil, errAdminsPurgeOrganizations.New()
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		err := gormstore.GetContactInfoStore(db).DeleteEntityContactInfo(ctx, ids)
		if err != nil {
			return err
		}
		// Delete related API keys before purging the organization.
		err = gormstore.GetAPIKeyStore(db).DeleteEntityAPIKeys(ctx, ids.GetEntityIdentifiers())
		if err != nil {
			return err
		}
		err = gormstore.GetMembershipStore(db).DeleteAccountMembers(ctx, ids.GetOrganizationOrUserIdentifiers())
		if err != nil {
			return err
		}
		return gormstore.GetOrganizationStore(db).PurgeOrganization(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtPurgeOrganization.NewWithIdentifiersAndData(ctx, ids, nil))
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

func (or *organizationRegistry) Delete(ctx context.Context, req *ttnpb.OrganizationIdentifiers) (*pbtypes.Empty, error) {
	return or.deleteOrganization(ctx, req)
}

func (or *organizationRegistry) Restore(ctx context.Context, req *ttnpb.OrganizationIdentifiers) (*pbtypes.Empty, error) {
	return or.restoreOrganization(ctx, req)
}

func (or *organizationRegistry) Purge(ctx context.Context, req *ttnpb.OrganizationIdentifiers) (*pbtypes.Empty, error) {
	return or.purgeOrganization(ctx, req)
}
