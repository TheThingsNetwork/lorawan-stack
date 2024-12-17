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
	"time"

	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/blocklist"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/warning"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	evtCreateApplication = events.Define(
		"application.create", "create application",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUpdateApplication = events.Define(
		"application.update", "update application",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_INFO),
		events.WithUpdatedFieldsDataType(),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteApplication = events.Define(
		"application.delete", "delete application",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtRestoreApplication = events.Define(
		"application.restore", "restore application",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtPurgeApplication = events.Define(
		"application.purge", "purge application",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtIssueDevEUIForApplication = events.Define(
		"application.issue_dev_eui", "issue DevEUI for application",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
)

var (
	errAdminsCreateApplications = errors.DefinePermissionDenied(
		"admins_create_applications",
		"applications may only be created by admins, or in organizations",
	)
	errAdminsPurgeApplications = errors.DefinePermissionDenied(
		"admins_purge_applications",
		"applications may only be purged by admins",
	)
	errDevEUIIssuingNotEnabled = errors.DefineInvalidArgument(
		"dev_eui_issuing_not_enabled",
		"DevEUI issuing not configured",
	)
)

func (is *IdentityServer) createApplication( //nolint:gocyclo
	ctx context.Context, req *ttnpb.CreateApplicationRequest,
) (app *ttnpb.Application, err error) {
	if err = blocklist.Check(ctx, req.Application.GetIds().GetApplicationId()); err != nil {
		return nil, err
	}
	if usrIDs := req.Collaborator.GetUserIds(); usrIDs != nil {
		if !is.IsAdmin(ctx) && !is.configFromContext(ctx).UserRights.CreateApplications {
			return nil, errAdminsCreateApplications.New()
		}
		if err = rights.RequireUser(ctx, usrIDs, ttnpb.Right_RIGHT_USER_APPLICATIONS_CREATE); err != nil {
			return nil, err
		}
	} else if orgIDs := req.Collaborator.GetOrganizationIds(); orgIDs != nil {
		if err = rights.RequireOrganization(
			ctx, orgIDs, ttnpb.Right_RIGHT_ORGANIZATION_APPLICATIONS_CREATE,
		); err != nil {
			return nil, err
		}
	}
	if req.Application.AdministrativeContact == nil {
		req.Application.AdministrativeContact = req.Collaborator
	} else if err := validateCollaboratorEqualsContact(
		req.Collaborator, req.Application.AdministrativeContact,
	); err != nil {
		return nil, err
	}
	if req.Application.TechnicalContact == nil {
		req.Application.TechnicalContact = req.Collaborator
	} else if err := validateCollaboratorEqualsContact(
		req.Collaborator, req.Application.TechnicalContact,
	); err != nil {
		return nil, err
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		app, err = st.CreateApplication(ctx, req.Application)
		if err != nil {
			return err
		}
		if err = st.SetMember(
			ctx,
			req.Collaborator,
			app.GetIds().GetEntityIdentifiers(),
			ttnpb.RightsFrom(ttnpb.Right_RIGHT_ALL),
		); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtCreateApplication.NewWithIdentifiersAndData(ctx, req.Application.GetIds(), nil))
	return app, nil
}

func (is *IdentityServer) getApplication(
	ctx context.Context,
	req *ttnpb.GetApplicationRequest,
) (app *ttnpb.Application, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	contactInfoInPath := ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info")
	if contactInfoInPath {
		req.FieldMask.Paths = ttnpb.ExcludeFields(req.FieldMask.Paths, "contact_info")
		req.FieldMask.Paths = ttnpb.AddFields(req.FieldMask.Paths, "administrative_contact", "technical_contact")
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.ApplicationFieldPathsNested, req.FieldMask, getPaths, nil)
	if err = rights.RequireApplication(ctx, req.GetApplicationIds(), ttnpb.Right_RIGHT_APPLICATION_INFO); err != nil {
		if !ttnpb.HasOnlyAllowedFields(req.FieldMask.GetPaths(), ttnpb.PublicApplicationFields...) {
			return nil, err
		}
		defer func() { app = app.PublicSafe() }()
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		app, err = st.GetApplication(ctx, req.GetApplicationIds(), req.FieldMask.GetPaths())
		if err != nil {
			return err
		}
		if contactInfoInPath {
			app.ContactInfo, err = getContactsFromEntity(ctx, app, st)
			if err != nil {
				return err
			}
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (is *IdentityServer) listApplications( // nolint:gocyclo
	ctx context.Context,
	req *ttnpb.ListApplicationsRequest,
) (apps *ttnpb.Applications, err error) {
	contactInfoInPath := ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info")
	if contactInfoInPath {
		req.FieldMask.Paths = ttnpb.ExcludeFields(req.FieldMask.Paths, "contact_info")
		req.FieldMask.Paths = ttnpb.AddFields(req.FieldMask.Paths, "administrative_contact", "technical_contact")
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.ApplicationFieldPathsNested, req.FieldMask, getPaths, nil)

	authInfo, err := is.authInfo(ctx)
	if err != nil {
		return nil, err
	}
	callerAccountID := authInfo.GetOrganizationOrUserIdentifiers()
	var includeIndirect bool
	var clusterAuth bool
	// If request comes from cluster (list all applications), skip caller rights check.
	if clusterauth.Authorized(ctx) == nil && req.Collaborator == nil {
		clusterAuth = true
		req.FieldMask = cleanFieldMaskPaths([]string{"ids"}, req.FieldMask, nil, []string{"created_at", "updated_at"})
		if req.Deleted {
			ctx = store.WithSoftDeleted(ctx, false)
		}
	} else {
		if req.Collaborator == nil {
			req.Collaborator = callerAccountID
			includeIndirect = true
		}
		if req.Collaborator == nil {
			return &ttnpb.Applications{}, nil
		}

		if usrIDs := req.Collaborator.GetUserIds(); usrIDs != nil {
			if err = rights.RequireUser(ctx, usrIDs, ttnpb.Right_RIGHT_USER_APPLICATIONS_LIST); err != nil {
				return nil, err
			}
		} else if orgIDs := req.Collaborator.GetOrganizationIds(); orgIDs != nil {
			if err = rights.RequireOrganization(
				ctx, orgIDs, ttnpb.Right_RIGHT_ORGANIZATION_APPLICATIONS_LIST,
			); err != nil {
				return nil, err
			}
		}

		if req.Deleted {
			ctx = store.WithSoftDeleted(ctx, true)
		}
	}
	ctx = store.WithOrder(ctx, req.Order)
	var total uint64
	paginateCtx := store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()

	apps = &ttnpb.Applications{}
	var callerMemberships store.MembershipChains

	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		var ids []*ttnpb.EntityIdentifiers
		// If request comes from cluster, skip membership checks.
		if !clusterAuth {
			ids, err = st.FindMemberships(paginateCtx, req.Collaborator, "application", includeIndirect)
			if err != nil {
				return err
			}
			if len(ids) == 0 {
				return nil
			}
			callerMemberships, err = st.FindAccountMembershipChains(
				ctx,
				callerAccountID,
				"application",
				idStrings(ids...)...)
			if err != nil {
				return err
			}
		} else {
			ctx = paginateCtx
		}
		appIDs := make([]*ttnpb.ApplicationIdentifiers, 0, len(ids))
		for _, id := range ids {
			if appID := id.GetEntityIdentifiers().GetApplicationIds(); appID != nil {
				appIDs = append(appIDs, appID)
			}
		}
		apps.Applications, err = st.FindApplications(ctx, appIDs, req.FieldMask.GetPaths())
		if err != nil {
			return err
		}
		if contactInfoInPath {
			for _, app := range apps.Applications {
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

	for i, app := range apps.Applications {
		entityRights := callerMemberships.GetRights(callerAccountID, app.GetIds()).Union(authInfo.GetUniversalRights())
		if !entityRights.IncludesAll(ttnpb.Right_RIGHT_APPLICATION_INFO) {
			apps.Applications[i] = app.PublicSafe()
		}
	}

	return apps, nil
}

func (is *IdentityServer) updateApplication(
	ctx context.Context,
	req *ttnpb.UpdateApplicationRequest,
) (app *ttnpb.Application, err error) {
	if err = rights.RequireApplication(
		ctx, req.Application.GetIds(), ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC,
	); err != nil {
		return nil, err
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.ApplicationFieldPathsNested, req.FieldMask, nil, getPaths)
	if len(req.FieldMask.GetPaths()) == 0 {
		req.FieldMask = ttnpb.FieldMask(updatePaths...)
	}
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info") {
		warning.Add(ctx, "Contact info is deprecated and will be removed in the next major release")
		req.FieldMask.Paths = ttnpb.ExcludeFields(req.FieldMask.Paths, "contact_info")
	}

	if err := is.validateContactInfoRestrictions(
		ctx, req.Application.GetAdministrativeContact(), req.Application.GetTechnicalContact(),
	); err != nil {
		return nil, err
	}

	req.FieldMask.Paths = ttnpb.FlattenPaths(
		req.FieldMask.Paths,
		[]string{"administrative_contact", "technical_contact"},
	)
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		if err := validateContactIsCollaborator(
			ctx, st, req.Application.AdministrativeContact, req.Application.GetEntityIdentifiers(),
		); err != nil {
			return err
		}
		if err := validateContactIsCollaborator(
			ctx, st, req.Application.TechnicalContact, req.Application.GetEntityIdentifiers(),
		); err != nil {
			return err
		}
		app, err = st.UpdateApplication(ctx, req.Application, req.FieldMask.GetPaths())
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	events.Publish(
		evtUpdateApplication.NewWithIdentifiersAndData(ctx, req.Application.GetIds(), req.FieldMask.GetPaths()),
	)
	return app, nil
}

var errApplicationHasDevices = errors.DefineFailedPrecondition(
	"application_has_devices",
	"application still has `{count}` devices",
)

func (is *IdentityServer) deleteApplication(
	ctx context.Context,
	ids *ttnpb.ApplicationIdentifiers,
) (*emptypb.Empty, error) {
	if err := rights.RequireApplication(ctx, ids, ttnpb.Right_RIGHT_APPLICATION_DELETE); err != nil {
		return nil, err
	}
	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		total, err := st.CountEndDevices(ctx, ids)
		if err != nil {
			return err
		}
		if total > 0 {
			return errApplicationHasDevices.WithAttributes("count", int(total))
		}
		if err := st.DeleteEntityBookmarks(ctx, ids.GetEntityIdentifiers()); err != nil {
			return err
		}
		return st.DeleteApplication(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtDeleteApplication.NewWithIdentifiersAndData(ctx, ids, nil))
	return ttnpb.Empty, nil
}

func (is *IdentityServer) restoreApplication(
	ctx context.Context,
	ids *ttnpb.ApplicationIdentifiers,
) (*emptypb.Empty, error) {
	if err := rights.RequireApplication(
		store.WithSoftDeleted(ctx, false), ids, ttnpb.Right_RIGHT_APPLICATION_DELETE,
	); err != nil {
		return nil, err
	}
	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		app, err := st.GetApplication(store.WithSoftDeleted(ctx, true), ids, softDeleteFieldMask)
		if err != nil {
			return err
		}
		deletedAt := ttnpb.StdTime(app.DeletedAt)
		if deletedAt == nil {
			panic("store.WithSoftDeleted(ctx, true) returned result that is not deleted")
		}
		if time.Since(*deletedAt) > is.configFromContext(ctx).Delete.Restore {
			return errRestoreWindowExpired.New()
		}
		if err := st.RestoreEntityBookmarks(ctx, ids.GetEntityIdentifiers()); err != nil {
			return err
		}
		return st.RestoreApplication(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtRestoreApplication.NewWithIdentifiersAndData(ctx, ids, nil))
	return ttnpb.Empty, nil
}

func (is *IdentityServer) purgeApplication(
	ctx context.Context,
	ids *ttnpb.ApplicationIdentifiers,
) (*emptypb.Empty, error) {
	if !is.IsAdmin(ctx) {
		return nil, errAdminsPurgeApplications.New()
	}
	if err := rights.RequireApplication(
		store.WithSoftDeleted(ctx, false), ids, ttnpb.Right_RIGHT_APPLICATION_PURGE,
	); err != nil {
		return nil, err
	}
	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		total, err := st.CountEndDevices(ctx, ids)
		if err != nil {
			return err
		}
		if total > 0 {
			return errApplicationHasDevices.WithAttributes("count", int(total))
		}
		// delete related API keys before purging the application
		err = st.DeleteEntityAPIKeys(ctx, ids.GetEntityIdentifiers())
		if err != nil {
			return err
		}
		// delete related memberships before purging the application
		err = st.DeleteEntityMembers(ctx, ids.GetEntityIdentifiers())
		if err != nil {
			return err
		}
		if err := st.PurgeEntityBookmarks(ctx, ids.GetEntityIdentifiers()); err != nil {
			return err
		}
		return st.PurgeApplication(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtPurgeApplication.NewWithIdentifiersAndData(ctx, ids, nil))
	return ttnpb.Empty, nil
}

func (is *IdentityServer) issueDevEUI(
	ctx context.Context,
	ids *ttnpb.ApplicationIdentifiers,
) (*ttnpb.IssueDevEUIResponse, error) {
	if err := rights.RequireApplication(
		store.WithSoftDeleted(ctx, false), ids, ttnpb.Right_RIGHT_APPLICATION_DEVICES_WRITE,
	); err != nil {
		return nil, err
	}
	if !is.config.DevEUIBlock.Enabled {
		return nil, errDevEUIIssuingNotEnabled.New()
	}
	res := &ttnpb.IssueDevEUIResponse{}
	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		devEUI, err := st.IssueDevEUIForApplication(ctx, ids, is.config.DevEUIBlock.ApplicationLimit)
		if err != nil {
			return err
		}
		res.DevEui = devEUI.Bytes()
		return nil
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtIssueDevEUIForApplication.NewWithIdentifiersAndData(ctx, ids, nil))
	return res, nil
}

type applicationRegistry struct {
	ttnpb.UnimplementedApplicationRegistryServer

	*IdentityServer
}

func (ar *applicationRegistry) Create(
	ctx context.Context,
	req *ttnpb.CreateApplicationRequest,
) (*ttnpb.Application, error) {
	return ar.createApplication(ctx, req)
}

func (ar *applicationRegistry) Get(ctx context.Context, req *ttnpb.GetApplicationRequest) (*ttnpb.Application, error) {
	return ar.getApplication(ctx, req)
}

func (ar *applicationRegistry) List(
	ctx context.Context,
	req *ttnpb.ListApplicationsRequest,
) (*ttnpb.Applications, error) {
	return ar.listApplications(ctx, req)
}

func (ar *applicationRegistry) Update(
	ctx context.Context,
	req *ttnpb.UpdateApplicationRequest,
) (*ttnpb.Application, error) {
	return ar.updateApplication(ctx, req)
}

func (ar *applicationRegistry) Delete(ctx context.Context, req *ttnpb.ApplicationIdentifiers) (*emptypb.Empty, error) {
	return ar.deleteApplication(ctx, req)
}

func (ar *applicationRegistry) Purge(ctx context.Context, req *ttnpb.ApplicationIdentifiers) (*emptypb.Empty, error) {
	return ar.purgeApplication(ctx, req)
}

func (ar *applicationRegistry) Restore(ctx context.Context, req *ttnpb.ApplicationIdentifiers) (*emptypb.Empty, error) {
	return ar.restoreApplication(ctx, req)
}

func (ar *applicationRegistry) IssueDevEUI(
	ctx context.Context,
	req *ttnpb.ApplicationIdentifiers,
) (*ttnpb.IssueDevEUIResponse, error) {
	return ar.issueDevEUI(ctx, req)
}
