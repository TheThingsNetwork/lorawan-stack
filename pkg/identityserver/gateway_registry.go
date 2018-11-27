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

func (is *IdentityServer) createGateway(ctx context.Context, req *ttnpb.CreateGatewayRequest) (gtw *ttnpb.Gateway, err error) {
	if usrIDs := req.Collaborator.GetUserIDs(); usrIDs != nil {
		if err = rights.RequireUser(ctx, *usrIDs, ttnpb.RIGHT_USER_GATEWAYS_CREATE); err != nil {
			return nil, err
		}
	} else if orgIDs := req.Collaborator.GetOrganizationIDs(); orgIDs != nil {
		if err = rights.RequireOrganization(ctx, *orgIDs, ttnpb.RIGHT_ORGANIZATION_GATEWAYS_CREATE); err != nil {
			return nil, err
		}
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		gtwStore := store.GetGatewayStore(db)
		gtw, err = gtwStore.CreateGateway(ctx, &req.Gateway)
		if err != nil {
			return err
		}
		memberStore := store.GetMembershipStore(db)
		err = memberStore.SetMember(ctx, &req.Collaborator, gtw.GatewayIdentifiers.EntityIdentifiers(), ttnpb.RightsFrom(ttnpb.RIGHT_GATEWAY_ALL))
		if err != nil {
			return err
		}
		// TODO: Create initial Gateway API key with "link" rights
		return nil
	})
	if err != nil {
		return nil, err
	}
	return gtw, nil
}

func (is *IdentityServer) getGateway(ctx context.Context, req *ttnpb.GetGatewayRequest) (gtw *ttnpb.Gateway, err error) {
	err = rights.RequireGateway(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_INFO)
	if err != nil {
		return nil, err
	}
	// TODO: Filter FieldMask by Rights
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		gtwStore := store.GetGatewayStore(db)
		gtw, err = gtwStore.GetGateway(ctx, &req.GatewayIdentifiers, &req.FieldMask)
		return err
	})
	if err != nil {
		return nil, err
	}
	return gtw, nil
}

func (is *IdentityServer) listGateways(ctx context.Context, req *ttnpb.ListGatewaysRequest) (gtws *ttnpb.Gateways, err error) {
	var gtwRights map[string]*ttnpb.Rights
	if req.Collaborator == nil {
		callerRights, err := is.getRights(ctx)
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
	ctx = store.SetTotalCount(ctx, &total)
	defer func() {
		if err == nil {
			grpc.SetHeader(ctx, metadata.Pairs("x-total-count", strconv.FormatUint(total, 10)))
		}
	}()
	gtws = new(ttnpb.Gateways)
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		if gtwRights == nil {
			memberStore := store.GetMembershipStore(db)
			rights, err := memberStore.FindMemberRights(ctx, req.Collaborator, "gateway")
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
		gtwStore := store.GetGatewayStore(db)
		gtws.Gateways, err = gtwStore.FindGateways(ctx, gtwIDs, &req.FieldMask)
		if err != nil {
			return err
		}
		for _, gtw := range gtws.Gateways {
			// TODO: Filter FieldMask by Rights
			_ = gtwRights[unique.ID(ctx, gtw.GatewayIdentifiers)]
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return gtws, nil
}

func (is *IdentityServer) updateGateway(ctx context.Context, req *ttnpb.UpdateGatewayRequest) (gtw *ttnpb.Gateway, err error) {
	err = rights.RequireGateway(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC)
	if err != nil {
		return nil, err
	}
	// TODO: Filter FieldMask by Rights
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		gtwStore := store.GetGatewayStore(db)
		gtw, err = gtwStore.UpdateGateway(ctx, &req.Gateway, &req.FieldMask)
		return err
	})
	if err != nil {
		return nil, err
	}
	return gtw, nil
}

func (is *IdentityServer) deleteGateway(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (*types.Empty, error) {
	err := rights.RequireGateway(ctx, *ids, ttnpb.RIGHT_GATEWAY_DELETE)
	if err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		gtwStore := store.GetGatewayStore(db)
		err = gtwStore.DeleteGateway(ctx, ids)
		return err
	})
	if err != nil {
		return nil, err
	}
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
func (gr *gatewayRegistry) List(ctx context.Context, req *ttnpb.ListGatewaysRequest) (*ttnpb.Gateways, error) {
	return gr.listGateways(ctx, req)
}
func (gr *gatewayRegistry) Update(ctx context.Context, req *ttnpb.UpdateGatewayRequest) (*ttnpb.Gateway, error) {
	return gr.updateGateway(ctx, req)
}
func (gr *gatewayRegistry) Delete(ctx context.Context, req *ttnpb.GatewayIdentifiers) (*types.Empty, error) {
	return gr.deleteGateway(ctx, req)
}
