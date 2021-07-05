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

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/blacklist"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	evtCreateApplication = events.Define(
		"application.create", "create application",
		events.WithVisibility(ttnpb.RIGHT_APPLICATION_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUpdateApplication = events.Define(
		"application.update", "update application",
		events.WithVisibility(ttnpb.RIGHT_APPLICATION_INFO),
		events.WithUpdatedFieldsDataType(),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteApplication = events.Define(
		"application.delete", "delete application",
		events.WithVisibility(ttnpb.RIGHT_APPLICATION_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtRestoreApplication = events.Define(
		"application.restore", "restore application",
		events.WithVisibility(ttnpb.RIGHT_APPLICATION_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtPurgeApplication = events.Define(
		"application.purge", "purge application",
		events.WithVisibility(ttnpb.RIGHT_APPLICATION_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtIssueDevEUIForApplication = events.Define(
		"application.issue_dev_eui", "issue DevEUI for application",
		events.WithVisibility(ttnpb.RIGHT_APPLICATION_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
)

var (
	errAdminsCreateApplications = errors.DefinePermissionDenied("admins_create_applications", "applications may only be created by admins, or in organizations")
	errAdminsPurgeApplications  = errors.DefinePermissionDenied("admins_purge_applications", "applications may only be purged by admins")
	errDevEUIIssuingNotEnabled  = errors.DefineInvalidArgument("dev_eui_issuing_not_enabled", "DevEUI issuing not configured")
)

func (is *IdentityServer) createApplication(ctx context.Context, req *ttnpb.CreateApplicationRequest) (app *ttnpb.Application, err error) {
	if err = blacklist.Check(ctx, req.ApplicationId); err != nil {
		return nil, err
	}
	if usrIDs := req.Collaborator.GetUserIds(); usrIDs != nil {
		if !is.IsAdmin(ctx) && !is.configFromContext(ctx).UserRights.CreateApplications {
			return nil, errAdminsCreateApplications
		}
		if err = rights.RequireUser(ctx, *usrIDs, ttnpb.RIGHT_USER_APPLICATIONS_CREATE); err != nil {
			return nil, err
		}
	} else if orgIDs := req.Collaborator.GetOrganizationIds(); orgIDs != nil {
		if err = rights.RequireOrganization(ctx, *orgIDs, ttnpb.RIGHT_ORGANIZATION_APPLICATIONS_CREATE); err != nil {
			return nil, err
		}
	}
	if err := validateContactInfo(req.Application.ContactInfo); err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		app, err = store.GetApplicationStore(db).CreateApplication(ctx, &req.Application)
		if err != nil {
			return err
		}
		if err = is.getMembershipStore(ctx, db).SetMember(
			ctx,
			&req.Collaborator,
			app.ApplicationIdentifiers.GetEntityIdentifiers(),
			ttnpb.RightsFrom(ttnpb.RIGHT_ALL),
		); err != nil {
			return err
		}
		if len(req.ContactInfo) > 0 {
			cleanContactInfo(req.ContactInfo)
			app.ContactInfo, err = store.GetContactInfoStore(db).SetContactInfo(ctx, app.ApplicationIdentifiers, req.ContactInfo)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtCreateApplication.NewWithIdentifiersAndData(ctx, &req.ApplicationIdentifiers, nil))
	return app, nil
}

func (is *IdentityServer) getApplication(ctx context.Context, req *ttnpb.GetApplicationRequest) (app *ttnpb.Application, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.ApplicationFieldPathsNested, req.FieldMask, getPaths, nil)
	if err = rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_INFO); err != nil {
		if ttnpb.HasOnlyAllowedFields(req.FieldMask.GetPaths(), ttnpb.PublicApplicationFields...) {
			defer func() { app = app.PublicSafe() }()
		} else {
			return nil, err
		}
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		app, err = store.GetApplicationStore(db).GetApplication(ctx, &req.ApplicationIdentifiers, req.FieldMask)
		if err != nil {
			return err
		}
		if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info") {
			app.ContactInfo, err = store.GetContactInfoStore(db).GetContactInfo(ctx, app.ApplicationIdentifiers)
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

func (is *IdentityServer) listApplications(ctx context.Context, req *ttnpb.ListApplicationsRequest) (apps *ttnpb.Applications, err error) {
	req.FieldMask = cleanFieldMaskPaths(ttnpb.ApplicationFieldPathsNested, req.FieldMask, getPaths, nil)
	var includeIndirect bool
	if req.Collaborator == nil {
		authInfo, err := is.authInfo(ctx)
		if err != nil {
			return nil, err
		}
		collaborator := authInfo.GetOrganizationOrUserIdentifiers()
		if collaborator == nil {
			return &ttnpb.Applications{}, nil
		}
		req.Collaborator = collaborator
		includeIndirect = true
	}
	if usrIDs := req.Collaborator.GetUserIds(); usrIDs != nil {
		if err = rights.RequireUser(ctx, *usrIDs, ttnpb.RIGHT_USER_APPLICATIONS_LIST); err != nil {
			return nil, err
		}
	} else if orgIDs := req.Collaborator.GetOrganizationIds(); orgIDs != nil {
		if err = rights.RequireOrganization(ctx, *orgIDs, ttnpb.RIGHT_ORGANIZATION_APPLICATIONS_LIST); err != nil {
			return nil, err
		}
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
	apps = &ttnpb.Applications{}
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		ids, err := is.getMembershipStore(ctx, db).FindMemberships(paginateCtx, req.Collaborator, "application", includeIndirect)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		appIDs := make([]*ttnpb.ApplicationIdentifiers, 0, len(ids))
		for _, id := range ids {
			if appID := id.GetEntityIdentifiers().GetApplicationIds(); appID != nil {
				appIDs = append(appIDs, appID)
			}
		}
		apps.Applications, err = store.GetApplicationStore(db).FindApplications(ctx, appIDs, req.FieldMask)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	for i, app := range apps.Applications {
		if rights.RequireApplication(ctx, app.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_INFO) != nil {
			apps.Applications[i] = app.PublicSafe()
		}
	}

	return apps, nil
}

func (is *IdentityServer) updateApplication(ctx context.Context, req *ttnpb.UpdateApplicationRequest) (app *ttnpb.Application, err error) {
	if err = rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC); err != nil {
		return nil, err
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.ApplicationFieldPathsNested, req.FieldMask, nil, getPaths)
	if len(req.FieldMask.GetPaths()) == 0 {
		req.FieldMask = &pbtypes.FieldMask{Paths: updatePaths}
	}
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info") {
		if err := validateContactInfo(req.Application.ContactInfo); err != nil {
			return nil, err
		}
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		app, err = store.GetApplicationStore(db).UpdateApplication(ctx, &req.Application, req.FieldMask)
		if err != nil {
			return err
		}
		if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info") {
			cleanContactInfo(req.ContactInfo)
			app.ContactInfo, err = store.GetContactInfoStore(db).SetContactInfo(ctx, app.ApplicationIdentifiers, req.ContactInfo)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtUpdateApplication.NewWithIdentifiersAndData(ctx, &req.ApplicationIdentifiers, req.FieldMask.GetPaths()))
	return app, nil
}

var errApplicationHasDevices = errors.DefineFailedPrecondition("application_has_devices", "application still has `{count}` devices")

func (is *IdentityServer) deleteApplication(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(ctx, *ids, ttnpb.RIGHT_APPLICATION_DELETE); err != nil {
		return nil, err
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		total, err := store.GetEndDeviceStore(db).CountEndDevices(ctx, ids)
		if err != nil {
			return err
		}
		if total > 0 {
			return errApplicationHasDevices.WithAttributes("count", int(total))
		}
		return store.GetApplicationStore(db).DeleteApplication(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtDeleteApplication.NewWithIdentifiersAndData(ctx, ids, nil))
	return ttnpb.Empty, nil
}

func (is *IdentityServer) restoreApplication(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireApplication(store.WithSoftDeleted(ctx, false), *ids, ttnpb.RIGHT_APPLICATION_DELETE); err != nil {
		return nil, err
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		appStore := store.GetApplicationStore(db)
		app, err := appStore.GetApplication(store.WithSoftDeleted(ctx, true), ids, softDeleteFieldMask)
		if err != nil {
			return err
		}
		if app.DeletedAt == nil {
			panic("store.WithSoftDeleted(ctx, true) returned result that is not deleted")
		}
		if time.Since(*app.DeletedAt) > is.configFromContext(ctx).Delete.Restore {
			return errRestoreWindowExpired.New()
		}
		return appStore.RestoreApplication(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtRestoreApplication.NewWithIdentifiersAndData(ctx, ids, nil))
	return ttnpb.Empty, nil
}

func (is *IdentityServer) purgeApplication(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*pbtypes.Empty, error) {
	if !is.IsAdmin(ctx) {
		return nil, errAdminsPurgeApplications
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		total, err := store.GetEndDeviceStore(db).CountEndDevices(ctx, ids)
		if err != nil {
			return err
		}
		if total > 0 {
			return errApplicationHasDevices.WithAttributes("count", int(total))
		}
		// delete related API keys before purging the application
		err = store.GetAPIKeyStore(db).DeleteEntityAPIKeys(ctx, ids.GetEntityIdentifiers())
		if err != nil {
			return err
		}
		// delete related memberships before purging the application
		err = store.GetMembershipStore(db).DeleteEntityMembers(ctx, ids.GetEntityIdentifiers())
		if err != nil {
			return err
		}
		// delete related contact info before purging the application
		err = store.GetContactInfoStore(db).DeleteEntityContactInfo(ctx, ids)
		if err != nil {
			return err
		}
		return store.GetApplicationStore(db).PurgeApplication(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtPurgeApplication.NewWithIdentifiersAndData(ctx, ids, nil))
	return ttnpb.Empty, nil
}

func (is *IdentityServer) issueDevEUI(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*ttnpb.IssueDevEUIResponse, error) {
	if err := rights.RequireApplication(store.WithSoftDeleted(ctx, false), *ids, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE); err != nil {
		return nil, err
	}
	if !is.config.DevEUIBlock.Enabled {
		return nil, errDevEUIIssuingNotEnabled.New()
	}
	res := &ttnpb.IssueDevEUIResponse{}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		devEUI, err := store.GetEUIStore(db).IssueDevEUIForApplication(ctx, ids, is.config.DevEUIBlock.ApplicationLimit)
		if err != nil {
			return err
		}
		res.DevEui = *devEUI
		return nil
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtIssueDevEUIForApplication.NewWithIdentifiersAndData(ctx, ids, nil))
	return res, nil
}

type applicationRegistry struct {
	*IdentityServer
}

func (ar *applicationRegistry) Create(ctx context.Context, req *ttnpb.CreateApplicationRequest) (*ttnpb.Application, error) {
	return ar.createApplication(ctx, req)
}

func (ar *applicationRegistry) Get(ctx context.Context, req *ttnpb.GetApplicationRequest) (*ttnpb.Application, error) {
	return ar.getApplication(ctx, req)
}

func (ar *applicationRegistry) List(ctx context.Context, req *ttnpb.ListApplicationsRequest) (*ttnpb.Applications, error) {
	return ar.listApplications(ctx, req)
}

func (ar *applicationRegistry) Update(ctx context.Context, req *ttnpb.UpdateApplicationRequest) (*ttnpb.Application, error) {
	return ar.updateApplication(ctx, req)
}

func (ar *applicationRegistry) Delete(ctx context.Context, req *ttnpb.ApplicationIdentifiers) (*pbtypes.Empty, error) {
	return ar.deleteApplication(ctx, req)
}

func (ar *applicationRegistry) Purge(ctx context.Context, req *ttnpb.ApplicationIdentifiers) (*pbtypes.Empty, error) {
	return ar.purgeApplication(ctx, req)
}

func (ar *applicationRegistry) Restore(ctx context.Context, req *ttnpb.ApplicationIdentifiers) (*pbtypes.Empty, error) {
	return ar.restoreApplication(ctx, req)
}

func (ar *applicationRegistry) IssueDevEUI(ctx context.Context, req *ttnpb.ApplicationIdentifiers) (*ttnpb.IssueDevEUIResponse, error) {
	return ar.issueDevEUI(ctx, req)
}
