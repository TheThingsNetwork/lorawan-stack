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
	gormstore "go.thethings.network/lorawan-stack/v3/pkg/identityserver/gormstore"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	evtCreateGateway = events.Define(
		"gateway.create", "create gateway",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUpdateGateway = events.Define(
		"gateway.update", "update gateway",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_INFO),
		events.WithUpdatedFieldsDataType(),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteGateway = events.Define(
		"gateway.delete", "delete gateway",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtRestoreGateway = events.Define(
		"gateway.restore", "restore gateway",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_INFO),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtPurgeGateway = events.Define(
		"gateway.purge", "purge gateway",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_INFO),
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
	reqGtw := req.GetGateway()
	if err = blacklist.Check(ctx, reqGtw.GetIds().GetGatewayId()); err != nil {
		return nil, err
	}
	if usrIDs := req.Collaborator.GetUserIds(); usrIDs != nil {
		if !is.IsAdmin(ctx) && !is.configFromContext(ctx).UserRights.CreateGateways {
			return nil, errAdminsCreateGateways.New()
		}
		if err = rights.RequireUser(ctx, *usrIDs, ttnpb.Right_RIGHT_USER_GATEWAYS_CREATE); err != nil {
			return nil, err
		}
	} else if orgIDs := req.Collaborator.GetOrganizationIds(); orgIDs != nil {
		if err = rights.RequireOrganization(ctx, *orgIDs, ttnpb.Right_RIGHT_ORGANIZATION_GATEWAYS_CREATE); err != nil {
			return nil, err
		}
	}

	if req.Gateway.AdministrativeContact == nil {
		req.Gateway.AdministrativeContact = req.Collaborator
	} else if err := validateCollaboratorEqualsContact(req.Collaborator, req.Gateway.AdministrativeContact); err != nil {
		return nil, err
	}
	if req.Gateway.TechnicalContact == nil {
		req.Gateway.TechnicalContact = req.Collaborator
	} else if err := validateCollaboratorEqualsContact(req.Collaborator, req.Gateway.TechnicalContact); err != nil {
		return nil, err
	}
	if err := validateContactInfo(reqGtw.ContactInfo); err != nil {
		return nil, err
	}

	if len(reqGtw.FrequencyPlanIds) == 0 && reqGtw.FrequencyPlanId != "" {
		reqGtw.FrequencyPlanIds = []string{reqGtw.FrequencyPlanId}
	}

	if reqGtw.LbsLnsSecret != nil {
		value := reqGtw.LbsLnsSecret.Value
		if is.config.Gateways.EncryptionKeyID != "" {
			value, err = is.KeyVault.Encrypt(ctx, reqGtw.LbsLnsSecret.Value, is.config.Gateways.EncryptionKeyID)
			if err != nil {
				return nil, err
			}
		} else {
			log.FromContext(ctx).Warn("No encryption key defined, store LBS LNS Secret in plaintext")
		}
		reqGtw.LbsLnsSecret.Value = value
		reqGtw.LbsLnsSecret.KeyId = is.config.Gateways.EncryptionKeyID
	}

	if reqGtw.TargetCupsKey != nil {
		value := reqGtw.TargetCupsKey.Value
		if is.config.Gateways.EncryptionKeyID != "" {
			value, err = is.KeyVault.Encrypt(ctx, reqGtw.TargetCupsKey.Value, is.config.Gateways.EncryptionKeyID)
			if err != nil {
				return nil, err
			}
		} else {
			log.FromContext(ctx).Warn("No encryption key defined, store Target CUPS Key in plaintext")
		}
		reqGtw.TargetCupsKey.Value = value
		reqGtw.TargetCupsKey.KeyId = is.config.Gateways.EncryptionKeyID
	}

	if reqGtw.ClaimAuthenticationCode != nil {
		if err = validateClaimAuthenticationCode(*reqGtw.ClaimAuthenticationCode); err != nil {
			return nil, err
		}
		value := reqGtw.ClaimAuthenticationCode.Secret.Value
		if is.config.Gateways.EncryptionKeyID != "" {
			value, err = is.KeyVault.Encrypt(ctx, value, is.config.Gateways.EncryptionKeyID)
			if err != nil {
				return nil, err
			}
		} else {
			log.FromContext(ctx).Warn("No encryption key defined, store Claim Authentication Code in plaintext")
		}
		reqGtw.ClaimAuthenticationCode.Secret.Value = value
		reqGtw.ClaimAuthenticationCode.Secret.KeyId = is.config.Gateways.EncryptionKeyID
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		gtw, err = gormstore.GetGatewayStore(db).CreateGateway(ctx, reqGtw)
		if err != nil {
			return err
		}
		if err = is.getMembershipStore(ctx, db).SetMember(
			ctx,
			req.Collaborator,
			gtw.GetIds().GetEntityIdentifiers(),
			ttnpb.RightsFrom(ttnpb.Right_RIGHT_ALL),
		); err != nil {
			return err
		}
		if len(reqGtw.ContactInfo) > 0 {
			cleanContactInfo(reqGtw.ContactInfo)
			gtw.ContactInfo, err = gormstore.GetContactInfoStore(db).SetContactInfo(ctx, gtw.GetIds(), reqGtw.ContactInfo)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		if errors.IsAlreadyExists(err) && errors.Resemble(err, gormstore.ErrEUITaken) {
			if ids, err := is.getGatewayIdentifiersForEUI(ctx, &ttnpb.GetGatewayIdentifiersForEUIRequest{
				Eui: reqGtw.GetIds().GetEui(),
			}); err == nil {
				return nil, errGatewayEUITaken.WithAttributes(
					"gateway_eui", reqGtw.GetIds().GetEui().String(),
					"gateway_id", ids.GetGatewayId(),
				)
			}
		}
		return nil, err
	}
	events.Publish(evtCreateGateway.NewWithIdentifiersAndData(ctx, reqGtw.GetIds(), nil))

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

	if err = rights.RequireGateway(ctx, *req.GetGatewayIds(), ttnpb.Right_RIGHT_GATEWAY_INFO); err != nil {
		if !ttnpb.HasOnlyAllowedFields(req.FieldMask.GetPaths(), ttnpb.PublicGatewayFields...) {
			return nil, err
		}
		defer func() { gtw = gtw.PublicSafe() }()
	}

	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "lbs_lns_secret", "claim_authentication_code", "target_cups_key") {
		if err = rights.RequireGateway(ctx, *req.GetGatewayIds(), ttnpb.Right_RIGHT_GATEWAY_READ_SECRETS); err != nil {
			return nil, err
		}
	}

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		gtw, err = gormstore.GetGatewayStore(db).GetGateway(ctx, req.GetGatewayIds(), req.FieldMask)
		if err != nil {
			return err
		}
		if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info") {
			gtw.ContactInfo, err = gormstore.GetContactInfoStore(db).GetContactInfo(ctx, gtw.GetIds())
			if err != nil {
				return err
			}
		}
		return err
	})
	if err != nil {
		return nil, err
	}

	if gtw.LbsLnsSecret != nil {
		value := gtw.LbsLnsSecret.Value
		if gtw.LbsLnsSecret.KeyId != "" {
			value, err = is.KeyVault.Decrypt(ctx, gtw.LbsLnsSecret.Value, gtw.LbsLnsSecret.KeyId)
			if err != nil {
				return nil, err
			}
		} else {
			log.FromContext(ctx).Warn("No encryption key defined, return stored LBS LNS Secret value")
		}
		gtw.LbsLnsSecret.Value = value
		gtw.LbsLnsSecret.KeyId = is.config.Gateways.EncryptionKeyID
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
	if len(gtw.FrequencyPlanIds) > 0 {
		gtw.FrequencyPlanId = gtw.FrequencyPlanIds[0]
	}

	if gtw.TargetCupsKey != nil {
		value := gtw.TargetCupsKey.Value
		if gtw.TargetCupsKey.KeyId != "" {
			value, err = is.KeyVault.Decrypt(ctx, gtw.TargetCupsKey.Value, gtw.TargetCupsKey.KeyId)
			if err != nil {
				return nil, err
			}
		} else {
			log.FromContext(ctx).Warn("No encryption key defined, return stored Target CUPS Key value")
		}
		gtw.TargetCupsKey.Value = value
		gtw.TargetCupsKey.KeyId = is.config.Gateways.EncryptionKeyID
	}

	return gtw, nil
}

func (is *IdentityServer) getGatewayIdentifiersForEUI(ctx context.Context, req *ttnpb.GetGatewayIdentifiersForEUIRequest) (ids *ttnpb.GatewayIdentifiers, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		gtw, err := gormstore.GetGatewayStore(db).GetGateway(ctx, &ttnpb.GatewayIdentifiers{
			Eui: req.Eui,
		}, &pbtypes.FieldMask{Paths: []string{"ids.gateway_id", "ids.eui"}})
		if err != nil {
			return err
		}
		ids = gtw.GetIds()
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

	authInfo, err := is.authInfo(ctx)
	if err != nil {
		return nil, err
	}
	callerAccountID := authInfo.GetOrganizationOrUserIdentifiers()
	var includeIndirect bool
	if req.Collaborator == nil {
		req.Collaborator = callerAccountID
		includeIndirect = true
	}
	if req.Collaborator == nil {
		return &ttnpb.Gateways{}, nil
	}

	if usrIDs := req.Collaborator.GetUserIds(); usrIDs != nil {
		if err = rights.RequireUser(ctx, *usrIDs, ttnpb.Right_RIGHT_USER_GATEWAYS_LIST); err != nil {
			return nil, err
		}
	} else if orgIDs := req.Collaborator.GetOrganizationIds(); orgIDs != nil {
		if err = rights.RequireOrganization(ctx, *orgIDs, ttnpb.Right_RIGHT_ORGANIZATION_GATEWAYS_LIST); err != nil {
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
	var callerMemberships store.MembershipChains

	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		membershipStore := is.getMembershipStore(ctx, db)
		ids, err := membershipStore.FindMemberships(paginateCtx, req.Collaborator, "gateway", includeIndirect)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		callerMemberships, err = membershipStore.FindAccountMembershipChains(ctx, callerAccountID, "gateway", idStrings(ids...)...)
		if err != nil {
			return err
		}
		gtwIDs := make([]*ttnpb.GatewayIdentifiers, 0, len(ids))
		for _, id := range ids {
			if gtwID := id.GetEntityIdentifiers().GetGatewayIds(); gtwID != nil {
				gtwIDs = append(gtwIDs, gtwID)
			}
		}
		gtws.Gateways, err = gormstore.GetGatewayStore(db).FindGateways(ctx, gtwIDs, req.FieldMask)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	for i, gtw := range gtws.Gateways {
		entityRights := callerMemberships.GetRights(callerAccountID, gtw.GetIds()).Union(authInfo.GetUniversalRights())

		// Backwards compatibility for frequency_plan_id field.
		if len(gtw.FrequencyPlanIds) > 0 {
			gtw.FrequencyPlanId = gtw.FrequencyPlanIds[0]
		}

		if !entityRights.IncludesAll(ttnpb.Right_RIGHT_GATEWAY_INFO) {
			gtws.Gateways[i] = gtw.PublicSafe()
		}

		if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "lbs_lns_secret") {
			if !entityRights.IncludesAll(ttnpb.Right_RIGHT_GATEWAY_READ_SECRETS) {
				gtws.Gateways[i].LbsLnsSecret = nil
			} else if gtws.Gateways[i].LbsLnsSecret != nil {
				value := gtws.Gateways[i].LbsLnsSecret.Value
				if gtws.Gateways[i].LbsLnsSecret.KeyId != "" {
					value, err = is.KeyVault.Decrypt(ctx, gtws.Gateways[i].LbsLnsSecret.Value, gtws.Gateways[i].LbsLnsSecret.KeyId)
					if err != nil {
						return nil, err
					}
				} else {
					logger := log.FromContext(ctx)
					logger.Warn("No encryption key defined, return stored LBS LNS Secret value")
				}
				gtws.Gateways[i].LbsLnsSecret.Value = value
				gtws.Gateways[i].LbsLnsSecret.KeyId = is.config.Gateways.EncryptionKeyID
			}
		}

		if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "target_cups_key") {
			if !entityRights.IncludesAll(ttnpb.Right_RIGHT_GATEWAY_READ_SECRETS) {
				gtws.Gateways[i].TargetCupsKey = nil
			} else if gtws.Gateways[i].TargetCupsKey != nil {
				value := gtws.Gateways[i].TargetCupsKey.Value
				if gtws.Gateways[i].TargetCupsKey.KeyId != "" {
					value, err = is.KeyVault.Decrypt(ctx, gtws.Gateways[i].TargetCupsKey.Value, gtws.Gateways[i].TargetCupsKey.KeyId)
					if err != nil {
						return nil, err
					}
				} else {
					logger := log.FromContext(ctx)
					logger.Warn("No encryption key defined, return stored Target CUPS Key Secret value")
				}
				gtws.Gateways[i].TargetCupsKey.Value = value
				gtws.Gateways[i].TargetCupsKey.KeyId = is.config.Gateways.EncryptionKeyID
			}
		}

		if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "claim_authentication_code") {
			if !entityRights.IncludesAll(ttnpb.Right_RIGHT_GATEWAY_READ_SECRETS) {
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
	reqGtw := req.GetGateway()
	if err = rights.RequireGateway(ctx, *reqGtw.GetIds(), ttnpb.Right_RIGHT_GATEWAY_SETTINGS_BASIC); err != nil {
		// Allow setting the location field or the attributes field with the RIGHT_GATEWAY_LINK right.
		isLink := rights.RequireGateway(ctx, *reqGtw.GetIds(), ttnpb.Right_RIGHT_GATEWAY_LINK) == nil
		if !(isLink && ttnpb.HasOnlyAllowedFields(req.FieldMask.GetPaths(), "antennas", "attributes")) {
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
	if len(reqGtw.FrequencyPlanIds) == 0 && reqGtw.FrequencyPlanId != "" {
		reqGtw.FrequencyPlanIds = []string{reqGtw.FrequencyPlanId}
	}

	req.FieldMask = cleanFieldMaskPaths(ttnpb.GatewayFieldPathsNested, req.FieldMask, nil, append(getPaths, "frequency_plan_id"))
	if len(req.FieldMask.GetPaths()) == 0 {
		req.FieldMask = &pbtypes.FieldMask{Paths: updatePaths}
	}
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info") {
		if err := validateContactInfo(reqGtw.ContactInfo); err != nil {
			return nil, err
		}
	}
	req.FieldMask.Paths = ttnpb.FlattenPaths(req.FieldMask.Paths, []string{"administrative_contact", "technical_contact"})

	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "lbs_lns_secret") {
		if err := rights.RequireGateway(ctx, *reqGtw.GetIds(), ttnpb.Right_RIGHT_GATEWAY_WRITE_SECRETS); err != nil {
			return nil, err
		} else if reqGtw.LbsLnsSecret != nil {
			value := reqGtw.LbsLnsSecret.Value
			ptLBSLNSSecret = reqGtw.LbsLnsSecret.Value
			if is.config.Gateways.EncryptionKeyID != "" {
				value, err = is.KeyVault.Encrypt(ctx, reqGtw.LbsLnsSecret.Value, is.config.Gateways.EncryptionKeyID)
				if err != nil {
					return nil, err
				}
			} else {
				logger := log.FromContext(ctx)
				logger.Warn("No encryption key defined, store LBS LNS Secret in plaintext")
			}
			reqGtw.LbsLnsSecret.Value = value
			reqGtw.LbsLnsSecret.KeyId = is.config.Gateways.EncryptionKeyID
		}
	}

	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "target_cups_key") {
		if err := rights.RequireGateway(ctx, *reqGtw.GetIds(), ttnpb.Right_RIGHT_GATEWAY_WRITE_SECRETS); err != nil {
			return nil, err
		} else if reqGtw.TargetCupsKey != nil {
			value := reqGtw.TargetCupsKey.Value
			ptTargetCUPSKeySecret = reqGtw.TargetCupsKey.Value
			if is.config.Gateways.EncryptionKeyID != "" {
				value, err = is.KeyVault.Encrypt(ctx, reqGtw.TargetCupsKey.Value, is.config.Gateways.EncryptionKeyID)
				if err != nil {
					return nil, err
				}
			} else {
				logger := log.FromContext(ctx)
				logger.Warn("No encryption key defined, store Target CUPS Key in plaintext")
			}
			reqGtw.TargetCupsKey.Value = value
			reqGtw.TargetCupsKey.KeyId = is.config.Gateways.EncryptionKeyID
		}
	}

	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "claim_authentication_code") {
		if err := rights.RequireGateway(ctx, *reqGtw.GetIds(), ttnpb.Right_RIGHT_GATEWAY_WRITE_SECRETS); err != nil {
			return nil, err
		} else if reqGtw.ClaimAuthenticationCode != nil {
			if err := validateClaimAuthenticationCode(*reqGtw.ClaimAuthenticationCode); err != nil {
				return nil, err
			}
			value := reqGtw.ClaimAuthenticationCode.Secret.Value
			ptCACSecret = reqGtw.ClaimAuthenticationCode.Secret.Value
			if is.config.Gateways.EncryptionKeyID != "" {
				value, err = is.KeyVault.Encrypt(ctx, value, is.config.Gateways.EncryptionKeyID)
				if err != nil {
					return nil, err
				}
			} else {
				logger := log.FromContext(ctx)
				logger.Warn("No encryption key defined, store Claim Authentication Code in plaintext")
			}
			reqGtw.ClaimAuthenticationCode.Secret.Value = value
			reqGtw.ClaimAuthenticationCode.Secret.KeyId = is.config.Gateways.EncryptionKeyID
		}
	}

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		if err := validateContactIsCollaborator(ctx, db, req.Gateway.AdministrativeContact, req.Gateway.GetEntityIdentifiers()); err != nil {
			return err
		}
		if err := validateContactIsCollaborator(ctx, db, req.Gateway.TechnicalContact, req.Gateway.GetEntityIdentifiers()); err != nil {
			return err
		}
		gtw, err = gormstore.GetGatewayStore(db).UpdateGateway(ctx, reqGtw, req.FieldMask)
		if err != nil {
			return err
		}
		if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info") {
			cleanContactInfo(reqGtw.ContactInfo)
			gtw.ContactInfo, err = gormstore.GetContactInfoStore(db).SetContactInfo(ctx, gtw.GetIds(), reqGtw.ContactInfo)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtUpdateGateway.NewWithIdentifiersAndData(ctx, reqGtw.GetIds(), req.FieldMask.GetPaths()))

	if len(ptCACSecret) != 0 {
		gtw.ClaimAuthenticationCode.Secret.Value = ptCACSecret
	}
	if len(ptLBSLNSSecret) != 0 {
		gtw.LbsLnsSecret.Value = ptLBSLNSSecret
	}
	if len(ptTargetCUPSKeySecret) != 0 {
		gtw.TargetCupsKey.Value = ptTargetCUPSKeySecret
	}

	return gtw, nil
}

func (is *IdentityServer) deleteGateway(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireGateway(ctx, *ids, ttnpb.Right_RIGHT_GATEWAY_DELETE); err != nil {
		return nil, err
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		return gormstore.GetGatewayStore(db).DeleteGateway(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtDeleteGateway.NewWithIdentifiersAndData(ctx, ids, nil))
	return ttnpb.Empty, nil
}

func (is *IdentityServer) restoreGateway(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireGateway(store.WithSoftDeleted(ctx, false), *ids, ttnpb.Right_RIGHT_GATEWAY_DELETE); err != nil {
		return nil, err
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		gtwStore := gormstore.GetGatewayStore(db)
		gtw, err := gtwStore.GetGateway(store.WithSoftDeleted(ctx, true), ids, softDeleteFieldMask)
		if err != nil {
			return err
		}
		deletedAt := ttnpb.StdTime(gtw.DeletedAt)
		if deletedAt == nil {
			panic("store.WithSoftDeleted(ctx, true) returned result that is not deleted")
		}
		if time.Since(*deletedAt) > is.configFromContext(ctx).Delete.Restore {
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
		return nil, errAdminsPurgeGateways.New()
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		// delete related API keys before purging the gateway
		err := gormstore.GetAPIKeyStore(db).DeleteEntityAPIKeys(ctx, ids.GetEntityIdentifiers())
		if err != nil {
			return err
		}
		// delete related memberships before purging the gateway
		err = gormstore.GetMembershipStore(db).DeleteEntityMembers(ctx, ids.GetEntityIdentifiers())
		if err != nil {
			return err
		}
		// delete related contact info before purging the gateway
		err = gormstore.GetContactInfoStore(db).DeleteEntityContactInfo(ctx, ids)
		if err != nil {
			return err
		}
		return gormstore.GetGatewayStore(db).PurgeGateway(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtPurgeGateway.NewWithIdentifiersAndData(ctx, ids, nil))
	return ttnpb.Empty, nil
}

func validateClaimAuthenticationCode(authCode ttnpb.GatewayClaimAuthenticationCode) error {
	if authCode.Secret == nil {
		return errClaimAuthenticationCode.New()
	}
	if validFrom, validTo := ttnpb.StdTime(authCode.ValidFrom), ttnpb.StdTime(authCode.ValidTo); validFrom != nil && validTo != nil {
		if validTo.Before(*validFrom) {
			return errClaimAuthenticationCode.New()
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
