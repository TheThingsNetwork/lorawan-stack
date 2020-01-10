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
	evtCreateGateway = events.Define(
		"gateway.create", "create gateway",
		ttnpb.RIGHT_GATEWAY_INFO,
	)
	evtUpdateGateway = events.Define(
		"gateway.update", "update gateway",
		ttnpb.RIGHT_GATEWAY_INFO,
	)
	evtDeleteGateway = events.Define(
		"gateway.delete", "delete gateway",
		ttnpb.RIGHT_GATEWAY_INFO,
	)
)

var errFrequencyPlans = errors.DefineInvalidArgument("frequency_plans", "frequency plans must be from the same band")

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
	if len(req.FrequencyPlanIDs) == 0 && req.FrequencyPlanID != "" {
		req.FrequencyPlanIDs = []string{req.FrequencyPlanID}
	}

	// Ensure that all Frequency Plan IDs are from the same band.
	lenFPIDs := len(req.FrequencyPlanIDs)
	if lenFPIDs > 1 {
		for i := 0; i < (lenFPIDs - 1); i++ {
			fp1, err := is.FrequencyPlans.GetByID(req.FrequencyPlanIDs[i])
			if err != nil {
				return nil, err
			}
			fp2, err := is.FrequencyPlans.GetByID(req.FrequencyPlanIDs[i+1])
			if err != nil {
				return nil, err
			}
			if fp1.BandID != fp2.BandID {
				return nil, errFrequencyPlans
			}
		}
	}

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		gtw, err = store.GetGatewayStore(db).CreateGateway(ctx, &req.Gateway)
		if err != nil {
			return err
		}
		if err = is.getMembershipStore(ctx, db).SetMember(
			ctx,
			&req.Collaborator,
			gtw.GatewayIdentifiers,
			ttnpb.RightsFrom(ttnpb.RIGHT_ALL),
		); err != nil {
			return err
		}
		if len(req.ContactInfo) > 0 {
			cleanContactInfo(req.ContactInfo)
			gtw.ContactInfo, err = store.GetContactInfoStore(db).SetContactInfo(ctx, gtw.GatewayIdentifiers, req.ContactInfo)
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
	return gtw, nil
}

func (is *IdentityServer) getGateway(ctx context.Context, req *ttnpb.GetGatewayRequest) (gtw *ttnpb.Gateway, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}

	// Backwards compatibility for frequency_plan_id field.
	if ttnpb.HasAnyField(req.FieldMask.Paths, "frequency_plan_id") {
		if !ttnpb.HasAnyField(req.FieldMask.Paths, "frequency_plan_ids") {
			req.FieldMask.Paths = append(req.FieldMask.Paths, "frequency_plan_ids")
		}
	}
	req.FieldMask.Paths = cleanFieldMaskPaths(ttnpb.GatewayFieldPathsNested, req.FieldMask.Paths, getPaths, []string{"frequency_plan_id"})

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
			gtw.ContactInfo, err = store.GetContactInfoStore(db).GetContactInfo(ctx, gtw.GatewayIdentifiers)
			if err != nil {
				return err
			}
		}
		return err
	})
	if err != nil {
		return nil, err
	}

	// Backwards compatibility for frequency_plan_id field.
	if len(gtw.FrequencyPlanIDs) > 0 {
		gtw.FrequencyPlanID = gtw.FrequencyPlanIDs[0]
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
	// Backwards compatibility for frequency_plan_id field.
	if ttnpb.HasAnyField(req.FieldMask.Paths, "frequency_plan_id") {
		if !ttnpb.HasAnyField(req.FieldMask.Paths, "frequency_plan_ids") {
			req.FieldMask.Paths = append(req.FieldMask.Paths, "frequency_plan_ids")
		}
	}
	req.FieldMask.Paths = cleanFieldMaskPaths(ttnpb.GatewayFieldPathsNested, req.FieldMask.Paths, getPaths, []string{"frequency_plan_id"})

	var includeIndirect bool
	if req.Collaborator == nil {
		authInfo, err := is.authInfo(ctx)
		if err != nil {
			return nil, err
		}
		collaborator := authInfo.GetOrganizationOrUserIdentifiers()
		if collaborator == nil {
			return &ttnpb.Gateways{}, nil
		}
		req.Collaborator = collaborator
		includeIndirect = true
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
	ctx = store.WithOrder(ctx, req.Order)
	var total uint64
	paginateCtx := store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	gtws = &ttnpb.Gateways{}
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		ids, err := is.getMembershipStore(ctx, db).FindMemberships(paginateCtx, req.Collaborator, "gateway", includeIndirect)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		gtwIDs := make([]*ttnpb.GatewayIdentifiers, 0, len(ids))
		for _, id := range ids {
			if gtwID := id.EntityIdentifiers().GetGatewayIDs(); gtwID != nil {
				gtwIDs = append(gtwIDs, gtwID)
			}
		}
		gtws.Gateways, err = store.GetGatewayStore(db).FindGateways(ctx, gtwIDs, &req.FieldMask)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	for i, gtw := range gtws.Gateways {
		// Backwards compatibility for frequency_plan_id field.
		if len(gtw.FrequencyPlanIDs) > 0 {
			gtw.FrequencyPlanID = gtw.FrequencyPlanIDs[0]
		}

		if rights.RequireGateway(ctx, gtw.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_INFO) != nil {
			gtws.Gateways[i] = gtw.PublicSafe()
		}
	}

	return gtws, nil
}

func (is *IdentityServer) updateGateway(ctx context.Context, req *ttnpb.UpdateGatewayRequest) (gtw *ttnpb.Gateway, err error) {
	if err = rights.RequireGateway(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC); err != nil {
		return nil, err
	}

	// Backwards compatibility for frequency_plan_id field.
	if ttnpb.HasAnyField(req.FieldMask.Paths, "frequency_plan_id") {
		if !ttnpb.HasAnyField(req.FieldMask.Paths, "frequency_plan_ids") {
			req.FieldMask.Paths = append(req.FieldMask.Paths, "frequency_plan_ids")
		}
	}
	if len(req.FrequencyPlanIDs) == 0 && req.FrequencyPlanID != "" {
		req.FrequencyPlanIDs = []string{req.FrequencyPlanID}
	}

	// Ensure that all Frequency Plan IDs are from the same band.
	lenFPIDs := len(req.FrequencyPlanIDs)
	if lenFPIDs > 1 {
		for i := 0; i < (lenFPIDs - 1); i++ {
			fp1, err := is.FrequencyPlans.GetByID(req.FrequencyPlanIDs[i])
			if err != nil {
				return nil, err
			}
			fp2, err := is.FrequencyPlans.GetByID(req.FrequencyPlanIDs[i+1])
			if err != nil {
				return nil, err
			}
			if fp1.BandID != fp2.BandID {
				return nil, errFrequencyPlans
			}
		}
	}

	req.FieldMask.Paths = cleanFieldMaskPaths(ttnpb.GatewayFieldPathsNested, req.FieldMask.Paths, nil, append(getPaths, "frequency_plan_id"))
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
			gtw.ContactInfo, err = store.GetContactInfoStore(db).SetContactInfo(ctx, gtw.GatewayIdentifiers, req.ContactInfo)
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
