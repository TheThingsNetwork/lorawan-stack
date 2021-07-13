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
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	evtCreateGateway = events.Define(
		"gateway.create", "create gateway",
		events.WithVisibility(ttnpb.RIGHT_GATEWAY_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUpdateGateway = events.Define(
		"gateway.update", "update gateway",
		events.WithVisibility(ttnpb.RIGHT_GATEWAY_INFO),
		events.WithUpdatedFieldsDataType(),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteGateway = events.Define(
		"gateway.delete", "delete gateway",
		events.WithVisibility(ttnpb.RIGHT_GATEWAY_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtRestoreGateway = events.Define(
		"gateway.restore", "restore gateway",
		events.WithVisibility(ttnpb.RIGHT_GATEWAY_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtPurgeGateway = events.Define(
		"gateway.purge", "purge gateway",
		events.WithVisibility(ttnpb.RIGHT_GATEWAY_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
)

var (
	errAdminsCreateGateways    = errors.DefinePermissionDenied("admins_create_gateways", "gateways may only be created by admins, or in organizations")
	errGatewayEUITaken         = errors.DefineAlreadyExists("gateway_eui_taken", "a gateway with EUI `{gateway_eui}` is already registered as `{gateway_id}`")
	errAdminsPurgeGateways     = errors.DefinePermissionDenied("admins_purge_gateways", "gateways may only be purged by admins")
	errClaimAuthenticationCode = errors.DefineInvalidArgument("claim_authentication_code", "invalid claim authentication code")
)

func (is *IdentityServer) createGateway(ctx context.Context, req *ttnpb.CreateGatewayRequest) (gtw *ttnpb.Gateway, err error) {
	if err = blacklist.Check(ctx, req.GatewayId); err != nil {
		return nil, err
	}
	if usrIDs := req.Collaborator.GetUserIds(); usrIDs != nil {
		if !is.IsAdmin(ctx) && !is.configFromContext(ctx).UserRights.CreateGateways {
			return nil, errAdminsCreateGateways
		}
		if err = rights.RequireUser(ctx, *usrIDs, ttnpb.RIGHT_USER_GATEWAYS_CREATE); err != nil {
			return nil, err
		}
	} else if orgIDs := req.Collaborator.GetOrganizationIds(); orgIDs != nil {
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

	if req.LBSLNSSecret != nil {
		value := req.LBSLNSSecret.Value
		if is.config.Gateways.EncryptionKeyID != "" {
			value, err = is.KeyVault.Encrypt(ctx, req.LBSLNSSecret.Value, is.config.Gateways.EncryptionKeyID)
			if err != nil {
				return nil, err
			}
		} else {
			log.FromContext(ctx).Warn("No encryption key defined, store LBS LNS Secret in plaintext")
		}
		req.LBSLNSSecret.Value = value
		req.LBSLNSSecret.KeyId = is.config.Gateways.EncryptionKeyID
	}

	if req.TargetCUPSKey != nil {
		value := req.TargetCUPSKey.Value
		if is.config.Gateways.EncryptionKeyID != "" {
			value, err = is.KeyVault.Encrypt(ctx, req.TargetCUPSKey.Value, is.config.Gateways.EncryptionKeyID)
			if err != nil {
				return nil, err
			}
		} else {
			log.FromContext(ctx).Warn("No encryption key defined, store Target CUPS Key in plaintext")
		}
		req.TargetCUPSKey.Value = value
		req.TargetCUPSKey.KeyId = is.config.Gateways.EncryptionKeyID
	}

	if req.ClaimAuthenticationCode != nil {
		if err := validateClaimAuthenticationCode(*req.ClaimAuthenticationCode); err == nil {
			value := req.ClaimAuthenticationCode.Secret.Value
			if is.config.Gateways.EncryptionKeyID != "" {
				value, err = is.KeyVault.Encrypt(ctx, value, is.config.Gateways.EncryptionKeyID)
				if err != nil {
					return nil, err
				}
			} else {
				log.FromContext(ctx).Warn("No encryption key defined, store Claim Authentication Code in plaintext")
			}
			req.ClaimAuthenticationCode.Secret.Value = value
			req.ClaimAuthenticationCode.Secret.KeyId = is.config.Gateways.EncryptionKeyID

		} else {
			return nil, err
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
			gtw.GatewayIdentifiers.GetEntityIdentifiers(),
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
				Eui: *req.Eui,
			}); err == nil {
				return nil, errGatewayEUITaken.WithAttributes(
					"gateway_eui", req.Eui.String(),
					"gateway_id", ids.GetGatewayId(),
				)
			}
		}
		return nil, err
	}
	events.Publish(evtCreateGateway.NewWithIdentifiersAndData(ctx, &req.GatewayIdentifiers, nil))

	return gtw, nil
}

func (is *IdentityServer) getGateway(ctx context.Context, req *ttnpb.GetGatewayRequest) (gtw *ttnpb.Gateway, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}

	// Backwards compatibility for frequency_plan_id field.
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "frequency_plan_id") {
		if !ttnpb.HasAnyField(req.FieldMask.GetPaths(), "frequency_plan_ids") {
			req.FieldMask.Paths = append(req.FieldMask.GetPaths(), "frequency_plan_ids")
		}
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.GatewayFieldPathsNested, req.FieldMask, getPaths, []string{"frequency_plan_id"})

	if err = rights.RequireGateway(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_INFO); err != nil {
		if ttnpb.HasOnlyAllowedFields(req.FieldMask.GetPaths(), ttnpb.PublicGatewayFields...) {
			defer func() { gtw = gtw.PublicSafe() }()
		} else {
			return nil, err
		}
	}

	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "lbs_lns_secret", "claim_authentication_code", "target_cups_key") {
		if err = rights.RequireGateway(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_READ_SECRETS); err != nil {
			return nil, err
		}
	}

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		gtw, err = store.GetGatewayStore(db).GetGateway(ctx, &req.GatewayIdentifiers, req.FieldMask)
		if err != nil {
			return err
		}
		if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info") {
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

	if gtw.LBSLNSSecret != nil {
		value := gtw.LBSLNSSecret.Value
		if gtw.LBSLNSSecret.KeyId != "" {
			value, err = is.KeyVault.Decrypt(ctx, gtw.LBSLNSSecret.Value, gtw.LBSLNSSecret.KeyId)
			if err != nil {
				return nil, err
			}
		} else {
			log.FromContext(ctx).Warn("No encryption key defined, return stored LBS LNS Secret value")
		}
		gtw.LBSLNSSecret.Value = value
		gtw.LBSLNSSecret.KeyId = is.config.Gateways.EncryptionKeyID
	}

	if gtw.ClaimAuthenticationCode != nil && gtw.ClaimAuthenticationCode.Secret != nil {
		value := gtw.ClaimAuthenticationCode.Secret.Value
		if gtw.ClaimAuthenticationCode.Secret.KeyId != "" {
			value, err = is.KeyVault.Decrypt(ctx, gtw.ClaimAuthenticationCode.Secret.Value, gtw.ClaimAuthenticationCode.Secret.KeyId)
			if err != nil {
				return nil, err
			}
		} else {
			log.FromContext(ctx).Warn("No encryption key defined, return stored Claim Authentication Code value")
		}
		gtw.ClaimAuthenticationCode.Secret.Value = value
		gtw.ClaimAuthenticationCode.Secret.KeyId = is.config.Gateways.EncryptionKeyID
	}

	// Backwards compatibility for frequency_plan_id field.
	if len(gtw.FrequencyPlanIDs) > 0 {
		gtw.FrequencyPlanID = gtw.FrequencyPlanIDs[0]
	}

	if gtw.TargetCUPSKey != nil {
		value := gtw.TargetCUPSKey.Value
		if gtw.TargetCUPSKey.KeyId != "" {
			value, err = is.KeyVault.Decrypt(ctx, gtw.TargetCUPSKey.Value, gtw.TargetCUPSKey.KeyId)
			if err != nil {
				return nil, err
			}
		} else {
			log.FromContext(ctx).Warn("No encryption key defined, return stored Target CUPS Key value")
		}
		gtw.TargetCUPSKey.Value = value
		gtw.TargetCUPSKey.KeyId = is.config.Gateways.EncryptionKeyID
	}

	return gtw, nil
}

func (is *IdentityServer) getGatewayIdentifiersForEUI(ctx context.Context, req *ttnpb.GetGatewayIdentifiersForEUIRequest) (ids *ttnpb.GatewayIdentifiers, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		gtw, err := store.GetGatewayStore(db).GetGateway(ctx, &ttnpb.GatewayIdentifiers{
			Eui: &req.Eui,
		}, &pbtypes.FieldMask{Paths: []string{"ids.gateway_id", "ids.eui"}})
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
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "frequency_plan_id") {
		if !ttnpb.HasAnyField(req.FieldMask.GetPaths(), "frequency_plan_ids") {
			req.FieldMask.Paths = append(req.FieldMask.GetPaths(), "frequency_plan_ids")
		}
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.GatewayFieldPathsNested, req.FieldMask, getPaths, []string{"frequency_plan_id"})

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
	if usrIDs := req.Collaborator.GetUserIds(); usrIDs != nil {
		if err = rights.RequireUser(ctx, *usrIDs, ttnpb.RIGHT_USER_GATEWAYS_LIST); err != nil {
			return nil, err
		}
	} else if orgIDs := req.Collaborator.GetOrganizationIds(); orgIDs != nil {
		if err = rights.RequireOrganization(ctx, *orgIDs, ttnpb.RIGHT_ORGANIZATION_GATEWAYS_LIST); err != nil {
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
			if gtwID := id.GetEntityIdentifiers().GetGatewayIds(); gtwID != nil {
				gtwIDs = append(gtwIDs, gtwID)
			}
		}
		gtws.Gateways, err = store.GetGatewayStore(db).FindGateways(ctx, gtwIDs, req.FieldMask)
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

		if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "lbs_lns_secret") {
			if rights.RequireGateway(ctx, gtw.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_READ_SECRETS) != nil {
				gtws.Gateways[i].LBSLNSSecret = nil
			} else if gtws.Gateways[i].LBSLNSSecret != nil {
				value := gtws.Gateways[i].LBSLNSSecret.Value
				if gtws.Gateways[i].LBSLNSSecret.KeyId != "" {
					value, err = is.KeyVault.Decrypt(ctx, gtws.Gateways[i].LBSLNSSecret.Value, gtws.Gateways[i].LBSLNSSecret.KeyId)
					if err != nil {
						return nil, err
					}
				} else {
					logger := log.FromContext(ctx)
					logger.Warn("No encryption key defined, return stored LBS LNS Secret value")
				}
				gtws.Gateways[i].LBSLNSSecret.Value = value
				gtws.Gateways[i].LBSLNSSecret.KeyId = is.config.Gateways.EncryptionKeyID
			}
		}

		if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "target_cups_key") {
			if rights.RequireGateway(ctx, gtw.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_READ_SECRETS) != nil {
				gtws.Gateways[i].TargetCUPSKey = nil
			} else if gtws.Gateways[i].TargetCUPSKey != nil {
				value := gtws.Gateways[i].TargetCUPSKey.Value
				if gtws.Gateways[i].TargetCUPSKey.KeyId != "" {
					value, err = is.KeyVault.Decrypt(ctx, gtws.Gateways[i].TargetCUPSKey.Value, gtws.Gateways[i].TargetCUPSKey.KeyId)
					if err != nil {
						return nil, err
					}
				} else {
					logger := log.FromContext(ctx)
					logger.Warn("No encryption key defined, return stored Target CUPS Key Secret value")
				}
				gtws.Gateways[i].TargetCUPSKey.Value = value
				gtws.Gateways[i].TargetCUPSKey.KeyId = is.config.Gateways.EncryptionKeyID
			}
		}

		if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "claim_authentication_code") {
			if rights.RequireGateway(ctx, gtw.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_READ_SECRETS) != nil {
				gtws.Gateways[i].ClaimAuthenticationCode = nil
			} else if gtws.Gateways[i].ClaimAuthenticationCode != nil && gtws.Gateways[i].ClaimAuthenticationCode.Secret != nil {
				value := gtws.Gateways[i].ClaimAuthenticationCode.Secret.Value
				if keyID := gtws.Gateways[i].ClaimAuthenticationCode.Secret.KeyId; keyID != "" {
					value, err = is.KeyVault.Decrypt(ctx, value, keyID)
					if err != nil {
						return nil, err
					}
				} else {
					logger := log.FromContext(ctx)
					logger.Warn("No encryption key defined, return stored Claim Authentication Code value")
				}
				gtws.Gateways[i].ClaimAuthenticationCode.Secret.Value = value
			}
		}
	}

	return gtws, nil
}

func (is *IdentityServer) updateGateway(ctx context.Context, req *ttnpb.UpdateGatewayRequest) (gtw *ttnpb.Gateway, err error) {
	if err = rights.RequireGateway(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC); err != nil {
		// Allow setting only the location field with the RIGHT_GATEWAY_LINK right.
		isLink := rights.RequireGateway(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_LINK) == nil
		if topLevel := ttnpb.TopLevelFields(req.FieldMask.GetPaths()); !isLink || len(topLevel) != 1 || topLevel[0] != "antennas" {
			return nil, err
		}
	}

	// Store plaintext values to return in the response to clients.
	var ptLBSLNSSecret, ptCACSecret, ptTargetCUPSKeySecret []byte

	// Backwards compatibility for frequency_plan_id field.
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "frequency_plan_id") {
		if !ttnpb.HasAnyField(req.FieldMask.GetPaths(), "frequency_plan_ids") {
			req.FieldMask.Paths = append(req.FieldMask.GetPaths(), "frequency_plan_ids")
		}
	}
	if len(req.FrequencyPlanIDs) == 0 && req.FrequencyPlanID != "" {
		req.FrequencyPlanIDs = []string{req.FrequencyPlanID}
	}

	req.FieldMask = cleanFieldMaskPaths(ttnpb.GatewayFieldPathsNested, req.FieldMask, nil, append(getPaths, "frequency_plan_id"))
	if len(req.FieldMask.GetPaths()) == 0 {
		req.FieldMask = &pbtypes.FieldMask{Paths: updatePaths}
	}
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info") {
		if err := validateContactInfo(req.Gateway.ContactInfo); err != nil {
			return nil, err
		}
	}

	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "lbs_lns_secret") {
		if err := rights.RequireGateway(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_WRITE_SECRETS); err != nil {
			return nil, err
		} else if req.LBSLNSSecret != nil {
			value := req.LBSLNSSecret.Value
			ptLBSLNSSecret = req.LBSLNSSecret.Value
			if is.config.Gateways.EncryptionKeyID != "" {
				value, err = is.KeyVault.Encrypt(ctx, req.LBSLNSSecret.Value, is.config.Gateways.EncryptionKeyID)
				if err != nil {
					return nil, err
				}
			} else {
				logger := log.FromContext(ctx)
				logger.Warn("No encryption key defined, store LBS LNS Secret in plaintext")
			}
			req.LBSLNSSecret.Value = value
			req.LBSLNSSecret.KeyId = is.config.Gateways.EncryptionKeyID
		}
	}

	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "target_cups_key") {
		if err := rights.RequireGateway(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_WRITE_SECRETS); err != nil {
			return nil, err
		} else if req.TargetCUPSKey != nil {
			value := req.TargetCUPSKey.Value
			ptTargetCUPSKeySecret = req.TargetCUPSKey.Value
			if is.config.Gateways.EncryptionKeyID != "" {
				value, err = is.KeyVault.Encrypt(ctx, req.TargetCUPSKey.Value, is.config.Gateways.EncryptionKeyID)
				if err != nil {
					return nil, err
				}
			} else {
				logger := log.FromContext(ctx)
				logger.Warn("No encryption key defined, store Target CUPS Key in plaintext")
			}
			req.TargetCUPSKey.Value = value
			req.TargetCUPSKey.KeyId = is.config.Gateways.EncryptionKeyID
		}
	}

	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "claim_authentication_code") {
		if err := rights.RequireGateway(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_WRITE_SECRETS); err != nil {
			return nil, err
		} else if req.ClaimAuthenticationCode != nil {
			if err := validateClaimAuthenticationCode(*req.ClaimAuthenticationCode); err == nil {
				value := req.ClaimAuthenticationCode.Secret.Value
				ptCACSecret = req.ClaimAuthenticationCode.Secret.Value
				if is.config.Gateways.EncryptionKeyID != "" {
					value, err = is.KeyVault.Encrypt(ctx, value, is.config.Gateways.EncryptionKeyID)
					if err != nil {
						return nil, err
					}
				} else {
					logger := log.FromContext(ctx)
					logger.Warn("No encryption key defined, store Claim Authentication Code in plaintext")
				}
				req.ClaimAuthenticationCode.Secret.Value = value
				req.ClaimAuthenticationCode.Secret.KeyId = is.config.Gateways.EncryptionKeyID
			} else {
				return nil, err
			}
		}
	}

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		gtw, err = store.GetGatewayStore(db).UpdateGateway(ctx, &req.Gateway, req.FieldMask)
		if err != nil {
			return err
		}
		if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info") {
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
	events.Publish(evtUpdateGateway.NewWithIdentifiersAndData(ctx, &req.GatewayIdentifiers, req.FieldMask.GetPaths()))

	if len(ptCACSecret) != 0 {
		gtw.ClaimAuthenticationCode.Secret.Value = ptCACSecret
	}
	if len(ptLBSLNSSecret) != 0 {
		gtw.LBSLNSSecret.Value = ptLBSLNSSecret
	}
	if len(ptTargetCUPSKeySecret) != 0 {
		gtw.TargetCUPSKey.Value = ptTargetCUPSKeySecret
	}

	return gtw, nil
}

func (is *IdentityServer) deleteGateway(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireGateway(ctx, *ids, ttnpb.RIGHT_GATEWAY_DELETE); err != nil {
		return nil, err
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		return store.GetGatewayStore(db).DeleteGateway(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtDeleteGateway.NewWithIdentifiersAndData(ctx, ids, nil))
	return ttnpb.Empty, nil
}

func (is *IdentityServer) restoreGateway(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireGateway(store.WithSoftDeleted(ctx, false), *ids, ttnpb.RIGHT_GATEWAY_DELETE); err != nil {
		return nil, err
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		gtwStore := store.GetGatewayStore(db)
		gtw, err := gtwStore.GetGateway(store.WithSoftDeleted(ctx, true), ids, softDeleteFieldMask)
		if err != nil {
			return err
		}
		if gtw.DeletedAt == nil {
			panic("store.WithSoftDeleted(ctx, true) returned result that is not deleted")
		}
		if time.Since(*gtw.DeletedAt) > is.configFromContext(ctx).Delete.Restore {
			return errRestoreWindowExpired.New()
		}
		return gtwStore.RestoreGateway(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtRestoreGateway.NewWithIdentifiersAndData(ctx, ids, nil))
	return ttnpb.Empty, nil
}

func (is *IdentityServer) purgeGateway(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (*pbtypes.Empty, error) {
	if !is.IsAdmin(ctx) {
		return nil, errAdminsPurgeGateways
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		// delete related API keys before purging the gateway
		err := store.GetAPIKeyStore(db).DeleteEntityAPIKeys(ctx, ids.GetEntityIdentifiers())
		if err != nil {
			return err
		}
		// delete related memberships before purging the gateway
		err = store.GetMembershipStore(db).DeleteEntityMembers(ctx, ids.GetEntityIdentifiers())
		if err != nil {
			return err
		}
		// delete related contact info before purging the gateway
		err = store.GetContactInfoStore(db).DeleteEntityContactInfo(ctx, ids)
		if err != nil {
			return err
		}
		return store.GetGatewayStore(db).PurgeGateway(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtPurgeGateway.NewWithIdentifiersAndData(ctx, ids, nil))
	return ttnpb.Empty, nil
}

func validateClaimAuthenticationCode(authCode ttnpb.GatewayClaimAuthenticationCode) error {
	if authCode.Secret == nil {
		return errClaimAuthenticationCode
	}
	if authCode.ValidFrom != nil && authCode.ValidTo != nil {
		if authCode.ValidTo.Before(*authCode.ValidFrom) {
			return errClaimAuthenticationCode
		}
	}
	return nil
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

func (gr *gatewayRegistry) Delete(ctx context.Context, req *ttnpb.GatewayIdentifiers) (*pbtypes.Empty, error) {
	return gr.deleteGateway(ctx, req)
}

func (gr *gatewayRegistry) Restore(ctx context.Context, req *ttnpb.GatewayIdentifiers) (*pbtypes.Empty, error) {
	return gr.restoreGateway(ctx, req)
}

func (gr *gatewayRegistry) Purge(ctx context.Context, req *ttnpb.GatewayIdentifiers) (*pbtypes.Empty, error) {
	return gr.purgeGateway(ctx, req)
}
