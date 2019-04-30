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
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/identityserver/blacklist"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

var (
	evtCreateGateway = events.Define("gateway.create", "create gateway")
	evtUpdateGateway = events.Define("gateway.update", "update gateway")
	evtDeleteGateway = events.Define("gateway.delete", "delete gateway")
)

func (is *IdentityServer) createGateway(ctx context.Context, req *ttnpb.CreateGatewayRequest) (gtw *ttnpb.Gateway, err error) {
	if err = blacklist.Check(ctx, req.GatewayID); err != nil {
		return nil, err
	}
	if usrIDs := req.Collaborator.GetUserIDs(); usrIDs != nil {
		if err = rights.RequireUser(ctx, *usrIDs, ttnpb.RIGHT_USER_GATEWAYS_CREATE); err != nil {
			return nil, err
		}
	} else if orgIDs := req.Collaborator.GetOrganizationIDs(); orgIDs != nil {
		if err = rights.RequireOrganization(ctx, *orgIDs, ttnpb.RIGHT_ORGANIZATION_GATEWAYS_CREATE); err != nil {
			return nil, err
		}
	}
	if err := validateContactInfo(req.Gateway.ContactInfo); err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		gtw, err = store.GetGatewayStore(db).CreateGateway(ctx, &req.Gateway)
		if err != nil {
			return err
		}
		if err = store.GetMembershipStore(db).SetMember(
			ctx,
			&req.Collaborator,
			gtw.GatewayIdentifiers.EntityIdentifiers(),
			ttnpb.RightsFrom(ttnpb.RIGHT_ALL),
		); err != nil {
			return err
		}
		if len(req.ContactInfo) > 0 {
			cleanContactInfo(req.ContactInfo)
			gtw.ContactInfo, err = store.GetContactInfoStore(db).SetContactInfo(ctx, gtw.EntityIdentifiers(), req.ContactInfo)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtCreateGateway(ctx, req.GatewayIdentifiers, nil))
	is.invalidateCachedMembershipsForAccount(ctx, &req.Collaborator)
	return gtw, nil
}

func (is *IdentityServer) getGateway(ctx context.Context, req *ttnpb.GetGatewayRequest) (gtw *ttnpb.Gateway, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	req.FieldMask.Paths = cleanFieldMaskPaths(ttnpb.GatewayFieldPathsNested, req.FieldMask.Paths, getPaths, nil)
	if err = rights.RequireGateway(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_INFO); err != nil {
		if ttnpb.HasOnlyAllowedFields(req.FieldMask.Paths, ttnpb.PublicGatewayFields...) {
			defer func() { gtw = gtw.PublicSafe() }()
		} else {
			return nil, err
		}
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		gtw, err = store.GetGatewayStore(db).GetGateway(ctx, &req.GatewayIdentifiers, &req.FieldMask)
		if err != nil {
			return err
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "contact_info") {
			gtw.ContactInfo, err = store.GetContactInfoStore(db).GetContactInfo(ctx, gtw.EntityIdentifiers())
			if err != nil {
				return err
			}
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return gtw, nil
}

func (is *IdentityServer) getGatewayIdentifiersForEUI(ctx context.Context, req *ttnpb.GetGatewayIdentifiersForEUIRequest) (ids *ttnpb.GatewayIdentifiers, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		gtw, err := store.GetGatewayStore(db).GetGateway(ctx, &ttnpb.GatewayIdentifiers{
			EUI: &req.EUI,
		}, &types.FieldMask{Paths: []string{"ids.gateway_id", "ids.eui"}})
		if err != nil {
			return err
		}
		ids = &gtw.GatewayIdentifiers
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (is *IdentityServer) listGateways(ctx context.Context, req *ttnpb.ListGatewaysRequest) (gtws *ttnpb.Gateways, err error) {
	req.FieldMask.Paths = cleanFieldMaskPaths(ttnpb.GatewayFieldPathsNested, req.FieldMask.Paths, getPaths, nil)
	var gtwRights map[string]*ttnpb.Rights
	if req.Collaborator == nil {
		callerRights, _, err := is.getRights(ctx)
		if err != nil {
			return nil, err
		}
		gtwRights = make(map[string]*ttnpb.Rights, len(callerRights))
		for ids, rights := range callerRights {
			if ids := ids.GetGatewayIDs(); ids != nil {
				gtwRights[unique.ID(ctx, ids)] = rights
			}
		}
		if len(gtwRights) == 0 {
			return &ttnpb.Gateways{}, nil
		}
	}
	if usrIDs := req.Collaborator.GetUserIDs(); usrIDs != nil {
		if err = rights.RequireUser(ctx, *usrIDs, ttnpb.RIGHT_USER_GATEWAYS_LIST); err != nil {
			return nil, err
		}
	} else if orgIDs := req.Collaborator.GetOrganizationIDs(); orgIDs != nil {
		if err = rights.RequireOrganization(ctx, *orgIDs, ttnpb.RIGHT_ORGANIZATION_GATEWAYS_LIST); err != nil {
			return nil, err
		}
	}
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	gtws = &ttnpb.Gateways{}
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		if gtwRights == nil {
			rights, err := store.GetMembershipStore(db).FindMemberRights(ctx, req.Collaborator, "gateway")
			if err != nil {
				return err
			}
			gtwRights = make(map[string]*ttnpb.Rights, len(rights))
			for ids, rights := range rights {
				gtwRights[unique.ID(ctx, ids)] = rights
			}
		}
		if len(gtwRights) == 0 {
			return nil
		}
		gtwIDs := make([]*ttnpb.GatewayIdentifiers, 0, len(gtwRights))
		for uid := range gtwRights {
			gtwID, err := unique.ToGatewayID(uid)
			if err != nil {
				continue
			}
			gtwIDs = append(gtwIDs, &gtwID)
		}
		gtws.Gateways, err = store.GetGatewayStore(db).FindGateways(ctx, gtwIDs, &req.FieldMask)
		if err != nil {
			return err
		}
		for _, gtw := range gtws.Gateways {
			if !gtwRights[unique.ID(ctx, gtw.GatewayIdentifiers)].IncludesAll(ttnpb.RIGHT_GATEWAY_INFO) {
				gtw = gtw.PublicSafe()
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return gtws, nil
}

func (is *IdentityServer) updateGateway(ctx context.Context, req *ttnpb.UpdateGatewayRequest) (gtw *ttnpb.Gateway, err error) {
	if err = rights.RequireGateway(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC); err != nil {
		return nil, err
	}
	req.FieldMask.Paths = cleanFieldMaskPaths(ttnpb.GatewayFieldPathsNested, req.FieldMask.Paths, nil, getPaths)
	if len(req.FieldMask.Paths) == 0 {
		req.FieldMask.Paths = updatePaths
	}
	if ttnpb.HasAnyField(req.FieldMask.Paths, "contact_info") {
		if err := validateContactInfo(req.Gateway.ContactInfo); err != nil {
			return nil, err
		}
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		gtw, err = store.GetGatewayStore(db).UpdateGateway(ctx, &req.Gateway, &req.FieldMask)
		if err != nil {
			return err
		}
		if ttnpb.HasAnyField(req.FieldMask.Paths, "contact_info") {
			cleanContactInfo(req.ContactInfo)
			gtw.ContactInfo, err = store.GetContactInfoStore(db).SetContactInfo(ctx, gtw.EntityIdentifiers(), req.ContactInfo)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtUpdateGateway(ctx, req.GatewayIdentifiers, req.FieldMask.Paths))
	return gtw, nil
}

func (is *IdentityServer) deleteGateway(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (*types.Empty, error) {
	if err := rights.RequireGateway(ctx, *ids, ttnpb.RIGHT_GATEWAY_DELETE); err != nil {
		return nil, err
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		return store.GetGatewayStore(db).DeleteGateway(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtDeleteGateway(ctx, ids, nil))
	return ttnpb.Empty, nil
}

type gatewayRegistry struct {
	*IdentityServer
}

func (gr *gatewayRegistry) Create(ctx context.Context, req *ttnpb.CreateGatewayRequest) (*ttnpb.Gateway, error) {
	return gr.createGateway(ctx, req)
}
func (gr *gatewayRegistry) Get(ctx context.Context, req *ttnpb.GetGatewayRequest) (*ttnpb.Gateway, error) {
	return gr.getGateway(ctx, req)
}
func (gr *gatewayRegistry) GetIdentifiersForEUI(ctx context.Context, req *ttnpb.GetGatewayIdentifiersForEUIRequest) (*ttnpb.GatewayIdentifiers, error) {
	return gr.getGatewayIdentifiersForEUI(ctx, req)
}
func (gr *gatewayRegistry) List(ctx context.Context, req *ttnpb.ListGatewaysRequest) (*ttnpb.Gateways, error) {
	return gr.listGateways(ctx, req)
}
func (gr *gatewayRegistry) Update(ctx context.Context, req *ttnpb.UpdateGatewayRequest) (*ttnpb.Gateway, error) {
	return gr.updateGateway(ctx, req)
}
func (gr *gatewayRegistry) Delete(ctx context.Context, req *ttnpb.GatewayIdentifiers) (*types.Empty, error) {
	return gr.deleteGateway(ctx, req)
}
