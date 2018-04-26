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

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
)

type gatewayService struct {
	*IdentityServer
}

// CreateGateway creates a gateway in the network, sets the user as collaborator
// with all rights and creates an API key
func (s *gatewayService) CreateGateway(ctx context.Context, req *ttnpb.CreateGatewayRequest) (*pbtypes.Empty, error) {
	var ids ttnpb.OrganizationOrUserIdentifiers

	if !req.OrganizationIdentifiers.IsZero() {
		err := s.enforceOrganizationRights(ctx, req.OrganizationIdentifiers, ttnpb.RIGHT_ORGANIZATION_GATEWAYS_CREATE)
		if err != nil {
			return nil, err
		}

		ids = organizationOrUserIDsOrganizationIDs(req.OrganizationIdentifiers)
	} else {
		err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_GATEWAYS_CREATE)
		if err != nil {
			return nil, err
		}

		ids = organizationOrUserIDsUserIDs(authorizationDataFromContext(ctx).UserIdentifiers())
	}

	err := s.store.Transact(func(tx *store.Store) error {
		settings, err := tx.Settings.Get()
		if err != nil {
			return err
		}

		// check for blacklisted ids
		if !settings.IsIDAllowed(req.Gateway.GatewayID) {
			return ErrBlacklistedID.New(errors.Attributes{
				"id": req.Gateway.GatewayID,
			})
		}

		err = tx.Gateways.Create(&ttnpb.Gateway{
			GatewayIdentifiers: req.Gateway.GatewayIdentifiers,
			Description:        req.Gateway.Description,
			FrequencyPlanID:    req.Gateway.FrequencyPlanID,
			PrivacySettings:    req.Gateway.PrivacySettings,
			AutoUpdate:         req.Gateway.AutoUpdate,
			Platform:           req.Gateway.Platform,
			Antennas:           req.Gateway.Antennas,
			Attributes:         req.Gateway.Attributes,
			ClusterAddress:     req.Gateway.ClusterAddress,
			ContactAccountIDs:  req.Gateway.ContactAccountIDs,
			DisableTxDelay:     req.Gateway.DisableTxDelay,
		})
		if err != nil {
			return err
		}

		return tx.Gateways.SetCollaborator(ttnpb.GatewayCollaborator{
			GatewayIdentifiers:            req.Gateway.GatewayIdentifiers,
			OrganizationOrUserIdentifiers: ids,
			Rights: ttnpb.AllGatewayRights(),
		})
	})

	return ttnpb.Empty, err
}

// GetGateway returns a gateway information.
func (s *gatewayService) GetGateway(ctx context.Context, req *ttnpb.GatewayIdentifiers) (*ttnpb.Gateway, error) {
	ids := *req

	err := s.enforceGatewayRights(ctx, ids, ttnpb.RIGHT_GATEWAY_INFO)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Gateways.GetByID(ids, s.config.Specializers.Gateway)
	if err != nil {
		return nil, err
	}

	return found.GetGateway(), nil
}

// ListGateways returns all the gateways the current user is collaborator of.
func (s *gatewayService) ListGateways(ctx context.Context, req *ttnpb.ListGatewaysRequest) (*ttnpb.ListGatewaysResponse, error) {
	var ids ttnpb.OrganizationOrUserIdentifiers
	var err error

	if !req.OrganizationIdentifiers.IsZero() {
		err = s.enforceOrganizationRights(ctx, req.OrganizationIdentifiers, ttnpb.RIGHT_ORGANIZATION_GATEWAYS_LIST)
		ids = organizationOrUserIDsOrganizationIDs(req.OrganizationIdentifiers)
	} else {
		err = s.enforceUserRights(ctx, ttnpb.RIGHT_USER_GATEWAYS_LIST)
		ids = organizationOrUserIDsUserIDs(authorizationDataFromContext(ctx).UserIdentifiers())
	}

	if err != nil {
		return nil, err
	}

	found, err := s.store.Gateways.ListByOrganizationOrUser(ids, s.config.Specializers.Gateway)
	if err != nil {
		return nil, err
	}

	resp := &ttnpb.ListGatewaysResponse{
		Gateways: make([]*ttnpb.Gateway, 0, len(found)),
	}

	for _, gtw := range found {
		resp.Gateways = append(resp.Gateways, gtw.GetGateway())
	}

	return resp, nil
}

// UpdateGateway updates a gateway.
func (s *gatewayService) UpdateGateway(ctx context.Context, req *ttnpb.UpdateGatewayRequest) (*pbtypes.Empty, error) {
	err := s.enforceGatewayRights(ctx, req.Gateway.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC)
	if err != nil {
		return nil, err
	}

	err = s.store.Transact(func(tx *store.Store) error {
		found, err := tx.Gateways.GetByID(req.Gateway.GatewayIdentifiers, s.config.Specializers.Gateway)
		if err != nil {
			return err
		}
		gtw := found.GetGateway()

		for _, path := range req.UpdateMask.Paths {
			switch {
			case ttnpb.FieldPathGatewayDescription.MatchString(path):
				gtw.Description = req.Gateway.Description
			case ttnpb.FieldPathGatewayFrequencyPlanID.MatchString(path):
				gtw.FrequencyPlanID = req.Gateway.FrequencyPlanID
			case ttnpb.FieldPathGatewayPrivacySettingsStatusPublic.MatchString(path):
				gtw.PrivacySettings.StatusPublic = req.Gateway.PrivacySettings.StatusPublic
			case ttnpb.FieldPathGatewayPrivacySettingsLocationPublic.MatchString(path):
				gtw.PrivacySettings.LocationPublic = req.Gateway.PrivacySettings.LocationPublic
			case ttnpb.FieldPathGatewayPrivacySettingsContactable.MatchString(path):
				gtw.PrivacySettings.Contactable = req.Gateway.PrivacySettings.Contactable
			case ttnpb.FieldPathGatewayAutoUpdate.MatchString(path):
				gtw.AutoUpdate = req.Gateway.AutoUpdate
			case ttnpb.FieldPathGatewayPlatform.MatchString(path):
				gtw.Platform = req.Gateway.Platform
			case ttnpb.FieldPathGatewayAntennas.MatchString(path):
				if req.Gateway.Antennas == nil {
					req.Gateway.Antennas = []ttnpb.GatewayAntenna{}
				}
				gtw.Antennas = req.Gateway.Antennas
			case ttnpb.FieldPathGatewayAttributes.MatchString(path):
				attr := ttnpb.FieldPathGatewayAttributes.FindStringSubmatch(path)[1]

				if value, ok := req.Gateway.Attributes[attr]; ok && len(value) > 0 {
					gtw.Attributes[attr] = value
				} else {
					delete(gtw.Attributes, attr)
				}
			case ttnpb.FieldPathGatewayClusterAddress.MatchString(path):
				gtw.ClusterAddress = req.Gateway.ClusterAddress
			case ttnpb.FieldPathGatewayContactAccountIDs.MatchString(path):
				gtw.ContactAccountIDs = &ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_UserID{UserID: req.Gateway.ContactAccountIDs.GetUserID()}}
			case ttnpb.FieldPathGatewayDisableTxDelay.MatchString(path):
				gtw.DisableTxDelay = req.Gateway.DisableTxDelay
			default:
				return ttnpb.ErrInvalidPathUpdateMask.New(errors.Attributes{
					"path": path,
				})
			}
		}

		return tx.Gateways.Update(gtw)
	})

	return ttnpb.Empty, err
}

// DeleteGateway deletes a gateway.
func (s *gatewayService) DeleteGateway(ctx context.Context, req *ttnpb.GatewayIdentifiers) (*pbtypes.Empty, error) {
	ids := *req

	err := s.enforceGatewayRights(ctx, ids, ttnpb.RIGHT_GATEWAY_DELETE)
	if err != nil {
		return nil, err
	}

	return ttnpb.Empty, s.store.Gateways.Delete(ids)
}

// GenerateGatewayAPIKey generates a gateway API key and returns it.
func (s *gatewayService) GenerateGatewayAPIKey(ctx context.Context, req *ttnpb.GenerateGatewayAPIKeyRequest) (*ttnpb.APIKey, error) {
	err := s.enforceGatewayRights(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_SETTINGS_KEYS)
	if err != nil {
		return nil, err
	}

	k, err := auth.GenerateGatewayAPIKey(s.config.Hostname)
	if err != nil {
		return nil, err
	}

	key := ttnpb.APIKey{
		Key:    k,
		Name:   req.Name,
		Rights: req.Rights,
	}

	err = s.store.Gateways.SaveAPIKey(req.GatewayIdentifiers, key)
	if err != nil {
		return nil, err
	}

	return &key, nil
}

// ListGatewayAPIKeys list all the API keys from a gateway.
func (s *gatewayService) ListGatewayAPIKeys(ctx context.Context, req *ttnpb.GatewayIdentifiers) (*ttnpb.ListGatewayAPIKeysResponse, error) {
	ids := *req

	err := s.enforceGatewayRights(ctx, ids, ttnpb.RIGHT_GATEWAY_SETTINGS_KEYS)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Gateways.ListAPIKeys(ids)
	if err != nil {
		return nil, err
	}

	keys := make([]*ttnpb.APIKey, 0, len(found))
	for i := range found {
		keys = append(keys, &found[i])
	}

	return &ttnpb.ListGatewayAPIKeysResponse{
		APIKeys: keys,
	}, nil
}

// UpdateGatewayAPIKey updates an API key rights.
func (s *gatewayService) UpdateGatewayAPIKey(ctx context.Context, req *ttnpb.UpdateGatewayAPIKeyRequest) (*pbtypes.Empty, error) {
	err := s.enforceGatewayRights(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_SETTINGS_KEYS)
	if err != nil {
		return nil, err
	}

	return ttnpb.Empty, s.store.Gateways.UpdateAPIKeyRights(req.GatewayIdentifiers, req.Name, req.Rights)
}

// RemoveGatewayAPIKey removes a gateway API key.
func (s *gatewayService) RemoveGatewayAPIKey(ctx context.Context, req *ttnpb.RemoveGatewayAPIKeyRequest) (*pbtypes.Empty, error) {
	err := s.enforceGatewayRights(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_SETTINGS_KEYS)
	if err != nil {
		return nil, err
	}

	return ttnpb.Empty, s.store.Gateways.DeleteAPIKey(req.GatewayIdentifiers, req.Name)
}

// SetGatewayCollaborator sets a collaborationship between an user and an
// gateway upon a given set of rights.
//
// The call will return error if after perform the operation the sum of rights
// that all collaborators with `RIGHT_GATEWAY_SETTINGS_COLLABORATORS` right
// is not equal to entire set of available `RIGHT_GATEWAY_XXXXXX` rights.
func (s *gatewayService) SetGatewayCollaborator(ctx context.Context, req *ttnpb.GatewayCollaborator) (*pbtypes.Empty, error) {
	err := s.enforceGatewayRights(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_SETTINGS_COLLABORATORS)
	if err != nil {
		return nil, err
	}

	// The resulting set of rights is determined by the following steps:
	// 1. Get rights of target user/organization to the gateway
	// 2. The set of rights the caller can modify are the rights the caller has
	//		to the gateway. We refer them as modifiable rights
	// 3. The rights the caller cannot modify is the difference of the target
	//		user/organization rights minus the modifiable rights.
	// 4. The modifiable rights the target user/organization will have is the
	//		intersection between `2.` and the rights of the request
	// 5. The final set of rights is given by the sum of `2.` plus `4.`

	rights, err := s.store.Gateways.ListCollaboratorRights(req.GatewayIdentifiers, req.OrganizationOrUserIdentifiers)
	if err != nil {
		return nil, err
	}

	ad := authorizationDataFromContext(ctx)

	// modifiable is the set of rights the caller can modify
	var modifiable []ttnpb.Right
	switch ad.Source {
	case auth.Key:
		modifiable = ad.Rights
	case auth.Token:
		modifiable, err = s.store.Gateways.ListCollaboratorRights(req.GatewayIdentifiers, organizationOrUserIDsUserIDs(ad.UserIdentifiers()))
		if err != nil {
			return nil, err
		}
	}

	req.Rights = append(ttnpb.DifferenceRights(rights, modifiable), ttnpb.IntersectRights(req.Rights, modifiable)...)

	err = s.store.Transact(func(tx *store.Store) error {
		err := tx.Gateways.SetCollaborator(*req)
		if err != nil {
			return err
		}

		// Check if the sum of rights that collaborators with `SETTINGS_COLLABORATOR`
		// right is equal to the entire set of defined gateway rights.
		collaborators, err := tx.Gateways.ListCollaborators(req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_SETTINGS_COLLABORATORS)
		if err != nil {
			return err
		}

		rights = ttnpb.AllGatewayRights()
		for _, collaborator := range collaborators {
			rights = ttnpb.DifferenceRights(rights, collaborator.Rights)

			if len(rights) == 0 {
				return nil
			}
		}

		return ErrSetGatewayCollaboratorFailed.New(errors.Attributes{
			"missing_rights": rights,
		})
	})

	return ttnpb.Empty, err
}

// ListGatewayCollaborators returns all the collaborators that a gateway has.
func (s *gatewayService) ListGatewayCollaborators(ctx context.Context, req *ttnpb.GatewayIdentifiers) (*ttnpb.ListGatewayCollaboratorsResponse, error) {
	ids := *req

	err := s.enforceGatewayRights(ctx, ids, ttnpb.RIGHT_GATEWAY_SETTINGS_COLLABORATORS)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Gateways.ListCollaborators(ids)
	if err != nil {
		return nil, err
	}

	collaborators := make([]*ttnpb.GatewayCollaborator, 0, len(found))
	for i := range found {
		collaborators = append(collaborators, &found[i])
	}

	return &ttnpb.ListGatewayCollaboratorsResponse{
		Collaborators: collaborators,
	}, nil
}

// ListGatewayRights returns the rights the caller user has to a gateway.
func (s *gatewayService) ListGatewayRights(ctx context.Context, req *ttnpb.GatewayIdentifiers) (*ttnpb.ListGatewayRightsResponse, error) {
	ad := authorizationDataFromContext(ctx)

	resp := new(ttnpb.ListGatewayRightsResponse)

	switch ad.Source {
	case auth.Token:
		rights, err := s.store.Gateways.ListCollaboratorRights(*req, organizationOrUserIDsUserIDs(ad.UserIdentifiers()))
		if err != nil {
			return nil, err
		}

		// result rights are the intersection between the scope of the Client
		// and the rights that the user has to the gateway.
		resp.Rights = ttnpb.IntersectRights(ad.Rights, rights)
	case auth.Key:
		if !ad.GatewayIdentifiers().Contains(*req) {
			return nil, ErrNotAuthorized.New(nil)
		}

		resp.Rights = ad.Rights
	}

	return resp, nil
}
