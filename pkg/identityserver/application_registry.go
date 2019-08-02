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
)

var (
	evtCreateApplication = events.Define(
		"application.create", "create application",
		ttnpb.RIGHT_APPLICATION_INFO,
	)
	evtUpdateApplication = events.Define(
		"application.update", "update application",
		ttnpb.RIGHT_APPLICATION_INFO,
	)
	evtDeleteApplication = events.Define(
		"application.delete", "delete application",
		ttnpb.RIGHT_APPLICATION_INFO,
	)
)

func (is *IdentityServer) createApplication(ctx context.Context, req *ttnpb.CreateApplicationRequest) (app *ttnpb.Application, err error) {
	if err = blacklist.Check(ctx, req.ApplicationID); err != nil {
		return nil, err
	}
	if usrIDs := req.Collaborator.GetUserIDs(); usrIDs != nil {
		if err = rights.RequireUser(ctx, *usrIDs, ttnpb.RIGHT_USER_APPLICATIONS_CREATE); err != nil {
			return nil, err
		}
	} else if orgIDs := req.Collaborator.GetOrganizationIDs(); orgIDs != nil {
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
		if err = store.GetMembershipStore(db).SetMember(
			ctx,
			&req.Collaborator,
			app.ApplicationIdentifiers,
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
	events.Publish(evtCreateApplication(ctx, req.ApplicationIdentifiers, nil))
	is.invalidateCachedMembershipsForAccount(ctx, &req.Collaborator)
	return app, nil
}

func (is *IdentityServer) getApplication(ctx context.Context, req *ttnpb.GetApplicationRequest) (app *ttnpb.Application, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	req.FieldMask.Paths = cleanFieldMaskPaths(ttnpb.ApplicationFieldPathsNested, req.FieldMask.Paths, getPaths, nil)
	if err = rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_INFO); err != nil {
		if ttnpb.HasOnlyAllowedFields(req.FieldMask.Paths, ttnpb.PublicApplicationFields...) {
			defer func() { app = app.PublicSafe() }()
		} else {
			return nil, err
		}
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		app, err = store.GetApplicationStore(db).GetApplication(ctx, &req.ApplicationIdentifiers, &req.FieldMask)
		if err != nil {
			return err
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "contact_info") {
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
	req.FieldMask.Paths = cleanFieldMaskPaths(ttnpb.ApplicationFieldPathsNested, req.FieldMask.Paths, getPaths, nil)
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
	if usrIDs := req.Collaborator.GetUserIDs(); usrIDs != nil {
		if err = rights.RequireUser(ctx, *usrIDs, ttnpb.RIGHT_USER_APPLICATIONS_LIST); err != nil {
			return nil, err
		}
	} else if orgIDs := req.Collaborator.GetOrganizationIDs(); orgIDs != nil {
		if err = rights.RequireOrganization(ctx, *orgIDs, ttnpb.RIGHT_ORGANIZATION_APPLICATIONS_LIST); err != nil {
			return nil, err
		}
	}
	var total uint64
	paginateCtx := store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	apps = &ttnpb.Applications{}
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		ids, err := store.GetMembershipStore(db).FindMemberships(paginateCtx, req.Collaborator, "application", includeIndirect)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		appIDs := make([]*ttnpb.ApplicationIdentifiers, 0, len(ids))
		for _, id := range ids {
			if appID := id.EntityIdentifiers().GetApplicationIDs(); appID != nil {
				appIDs = append(appIDs, appID)
			}
		}
		apps.Applications, err = store.GetApplicationStore(db).FindApplications(ctx, appIDs, &req.FieldMask)
		if err != nil {
			return err
		}
		for i, app := range apps.Applications {
			if rights.RequireApplication(ctx, app.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_INFO) != nil {
				apps.Applications[i] = app.PublicSafe()
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return apps, nil
}

func (is *IdentityServer) updateApplication(ctx context.Context, req *ttnpb.UpdateApplicationRequest) (app *ttnpb.Application, err error) {
	if err = rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC); err != nil {
		return nil, err
	}
	req.FieldMask.Paths = cleanFieldMaskPaths(ttnpb.ApplicationFieldPathsNested, req.FieldMask.Paths, nil, getPaths)
	if len(req.FieldMask.Paths) == 0 {
		req.FieldMask.Paths = updatePaths
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "contact_info") {
		if err := validateContactInfo(req.Application.ContactInfo); err != nil {
			return nil, err
		}
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		app, err = store.GetApplicationStore(db).UpdateApplication(ctx, &req.Application, &req.FieldMask)
		if err != nil {
			return err
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "contact_info") {
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
	events.Publish(evtUpdateApplication(ctx, req.ApplicationIdentifiers, req.FieldMask.Paths))
	return app, nil
}

var errApplicationHasDevices = errors.DefineFailedPrecondition("application_has_devices", "application still has `{count}` devices")

func (is *IdentityServer) deleteApplication(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*types.Empty, error) {
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
	events.Publish(evtDeleteApplication(ctx, ids, nil))
	return ttnpb.Empty, nil
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

func (ar *applicationRegistry) Delete(ctx context.Context, req *ttnpb.ApplicationIdentifiers) (*types.Empty, error) {
	return ar.deleteApplication(ctx, req)
}
