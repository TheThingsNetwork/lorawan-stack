// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
	"strconv"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func (is *IdentityServer) createApplication(ctx context.Context, req *ttnpb.CreateApplicationRequest) (app *ttnpb.Application, err error) {
	if usrIDs := req.Collaborator.GetUserIDs(); usrIDs != nil {
		if err = rights.RequireUser(ctx, *usrIDs, ttnpb.RIGHT_USER_APPLICATIONS_CREATE); err != nil {
			return nil, err
		}
	} else if orgIDs := req.Collaborator.GetOrganizationIDs(); orgIDs != nil {
		if err = rights.RequireOrganization(ctx, *orgIDs, ttnpb.RIGHT_ORGANIZATION_APPLICATIONS_CREATE); err != nil {
			return nil, err
		}
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		appStore := store.GetApplicationStore(db)
		app, err = appStore.CreateApplication(ctx, &req.Application)
		if err != nil {
			return err
		}
		memberStore := store.GetMembershipStore(db)
		err = memberStore.SetMember(ctx, &req.Collaborator, app.ApplicationIdentifiers.EntityIdentifiers(), ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_ALL))
		if err != nil {
			return err
		}
		// TODO: Create initial Application API key with "link" rights
		return nil
	})
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (is *IdentityServer) getApplication(ctx context.Context, req *ttnpb.GetApplicationRequest) (app *ttnpb.Application, err error) {
	err = rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_INFO)
	if err != nil {
		return nil, err
	}
	// TODO: Filter FieldMask by Rights
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		appStore := store.GetApplicationStore(db)
		app, err = appStore.GetApplication(ctx, &req.ApplicationIdentifiers, &req.FieldMask)
		return err
	})
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (is *IdentityServer) listApplications(ctx context.Context, req *ttnpb.ListApplicationsRequest) (apps *ttnpb.Applications, err error) {
	var appRights map[string]*ttnpb.Rights
	if req.Collaborator == nil {
		rights, ok := rights.FromContext(ctx)
		if !ok {
			return &ttnpb.Applications{}, nil
		}
		appRights = rights.ApplicationRights
		if len(appRights) == 0 {
			return &ttnpb.Applications{}, nil
		}
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
	ctx = store.SetTotalCount(ctx, &total)
	defer func() {
		if err == nil {
			grpc.SetHeader(ctx, metadata.Pairs("x-total-count", strconv.FormatUint(total, 10)))
		}
	}()
	apps = new(ttnpb.Applications)
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		if appRights == nil {
			memberStore := store.GetMembershipStore(db)
			rights, err := memberStore.FindMemberRights(ctx, req.Collaborator, "application")
			if err != nil {
				return err
			}
			appRights = make(map[string]*ttnpb.Rights, len(rights))
			for ids, rights := range rights {
				appRights[unique.ID(ctx, ids)] = rights
			}
		}
		if len(appRights) == 0 {
			return nil
		}
		appIDs := make([]*ttnpb.ApplicationIdentifiers, 0, len(appRights))
		for uid := range appRights {
			appID, err := unique.ToApplicationID(uid)
			if err != nil {
				continue
			}
			appIDs = append(appIDs, &appID)
		}
		appStore := store.GetApplicationStore(db)
		apps.Applications, err = appStore.FindApplications(ctx, appIDs, &req.FieldMask)
		if err != nil {
			return err
		}
		for _, app := range apps.Applications {
			// TODO: Filter FieldMask by Rights
			_ = appRights[unique.ID(ctx, app.ApplicationIdentifiers)]
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return apps, nil
}

func (is *IdentityServer) updateApplication(ctx context.Context, req *ttnpb.UpdateApplicationRequest) (app *ttnpb.Application, err error) {
	err = rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC)
	if err != nil {
		return nil, err
	}
	// TODO: Filter FieldMask by Rights
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		appStore := store.GetApplicationStore(db)
		app, err = appStore.UpdateApplication(ctx, &req.Application, &req.FieldMask)
		return err
	})
	if err != nil {
		return nil, err
	}
	return app, nil
}

func (is *IdentityServer) deleteApplication(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*types.Empty, error) {
	err := rights.RequireApplication(ctx, *ids, ttnpb.RIGHT_APPLICATION_DELETE)
	if err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		appStore := store.GetApplicationStore(db)
		err = appStore.DeleteApplication(ctx, ids)
		return err
	})
	if err != nil {
		return nil, err
	}
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
