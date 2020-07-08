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
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/blacklist"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
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
	evtStoreSecret = events.Define(
		"gateway.secret.store", "store gateway secret",
		ttnpb.RIGHT_GATEWAY_INFO,
	)
	evtRetrieveSecret = events.Define(
		"gateway.secret.retrieve", "retrieve gateway secret",
		ttnpb.RIGHT_GATEWAY_INFO,
	)
)

var (
	errAdminsCreateGateways = errors.DefinePermissionDenied("admins_create_gateways", "gateways may only be created by admins, or in organizations")
	errGatewayEUITaken      = errors.DefineAlreadyExists("gateway_eui_taken", "a gateway with EUI `{gateway_eui}` is already registered as `{gateway_id}`")
)

func (is *IdentityServer) createGateway(ctx context.Context, req *ttnpb.CreateGatewayRequest) (gtw *ttnpb.Gateway, err error) {
	if err = blacklist.Check(ctx, req.GatewayID); err != nil {
		return nil, err
	}
	if usrIDs := req.Collaborator.GetUserIDs(); usrIDs != nil {
		if !is.IsAdmin(ctx) && !is.configFromContext(ctx).UserRights.CreateGateways {
			return nil, errAdminsCreateGateways
		}
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

	// Don't allow setting the secret field while creating the gateway.
	req.Gateway.Secret = nil

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
		if errors.IsAlreadyExists(err) && errors.Resemble(err, store.ErrEUITaken) {
			if ids, err := is.getGatewayIdentifiersForEUI(ctx, &ttnpb.GetGatewayIdentifiersForEUIRequest{
				EUI: *req.EUI,
			}); err == nil {
				return nil, errGatewayEUITaken.WithAttributes(
					"gateway_eui", req.EUI.String(),
					"gateway_id", ids.GetGatewayID(),
				)
			}
		}
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

	// Don't allow getting the secret field.
	req.FieldMask.Paths = ttnpb.ExcludeFields(req.FieldMask.Paths, "secret")

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

	// Don't allow listing the secret field.
	req.FieldMask.Paths = ttnpb.ExcludeFields(req.FieldMask.Paths, "secret")

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

	// Don't allow updating the secret field.
	req.FieldMask.Paths = ttnpb.ExcludeFields(req.FieldMask.Paths, "secret")

	// Backwards compatibility for frequency_plan_id field.
	if ttnpb.HasAnyField(req.FieldMask.Paths, "frequency_plan_id") {
		if !ttnpb.HasAnyField(req.FieldMask.Paths, "frequency_plan_ids") {
			req.FieldMask.Paths = append(req.FieldMask.Paths, "frequency_plan_ids")
		}
	}
	if len(req.FrequencyPlanIDs) == 0 && req.FrequencyPlanID != "" {
		req.FrequencyPlanIDs = []string{req.FrequencyPlanID}
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

func (is *IdentityServer) storeGatewaySecret(ctx context.Context, req *ttnpb.StoreGatewaySecretRequest) (*types.Empty, error) {
	// Require that caller has rights to encrypt and store the Secret.
	if err := rights.RequireGateway(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_WRITE_SECRET); err != nil {
		return nil, err
	}
	encrypted, err := is.KeyVault.Encrypt(ctx, []byte(req.PlainText.Value), is.config.GatewaySecretKeyID)
	if err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		_, err = store.GetGatewayStore(db).UpdateGateway(ctx, &ttnpb.Gateway{
			GatewayIdentifiers: req.GatewayIdentifiers,
			Secret:             encrypted,
		}, &types.FieldMask{Paths: []string{"secret"}})
		return err
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtStoreSecret(ctx, req.GatewayIdentifiers, nil))
	return &types.Empty{}, nil
}

func (is *IdentityServer) retrieveGatewaySecret(ctx context.Context, req *ttnpb.RetrieveGatewaySecretRequest) (*ttnpb.GatewaySecretPlainText, error) {
	// Require that caller has rights to retrive and decrypt the Secret.
	if err := rights.RequireGateway(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_READ_SECRET); err != nil {
		return nil, err
	}
	var gtw *ttnpb.Gateway
	err := is.withReadDatabase(ctx, func(db *gorm.DB) (err error) {
		gtw, err = store.GetGatewayStore(db).GetGateway(ctx, &req.GatewayIdentifiers, &types.FieldMask{Paths: []string{"secret"}})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	var decrypted []byte
	if len(gtw.Secret) != 0 {
		var err error
		decrypted, err = is.KeyVault.Decrypt(ctx, gtw.Secret, is.config.GatewaySecretKeyID)
		if err != nil {
			return nil, err
		}
	}
	return &ttnpb.GatewaySecretPlainText{
		Value: string(decrypted),
	}, nil
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

func (gr *gatewayRegistry) StoreGatewaySecret(ctx context.Context, req *ttnpb.StoreGatewaySecretRequest) (*types.Empty, error) {
	return gr.storeGatewaySecret(ctx, req)
}

func (gr *gatewayRegistry) RetrieveGatewaySecret(ctx context.Context, req *ttnpb.RetrieveGatewaySecretRequest) (*ttnpb.GatewaySecretPlainText, error) {
	return gr.retrieveGatewaySecret(ctx, req)
}
