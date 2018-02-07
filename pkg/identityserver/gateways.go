// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/util"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
)

type gatewayService struct {
	*IdentityServer
}

// CreateGateway creates a gateway in the network, sets the user as collaborator
// with all rights and creates an API key
func (s *gatewayService) CreateGateway(ctx context.Context, req *ttnpb.CreateGatewayRequest) (*pbtypes.Empty, error) {
	userID, err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_GATEWAYS_CREATE)
	if err != nil {
		return nil, err
	}

	err = s.store.Transact(func(tx *store.Store) error {
		settings, err := tx.Settings.Get()
		if err != nil {
			return err
		}

		// check for blacklisted ids
		if !util.IsIDAllowed(req.Gateway.GatewayID, settings.BlacklistedIDs) {
			return ErrBlacklistedID.New(errors.Attributes{
				"id": req.Gateway.GatewayID,
			})
		}

		err = tx.Gateways.Create(&ttnpb.Gateway{
			GatewayIdentifier: req.Gateway.GatewayIdentifier,
			Description:       req.Gateway.Description,
			FrequencyPlanID:   req.Gateway.FrequencyPlanID,
			PrivacySettings:   req.Gateway.PrivacySettings,
			AutoUpdate:        req.Gateway.AutoUpdate,
			Platform:          req.Gateway.Platform,
			Antennas:          req.Gateway.Antennas,
			Attributes:        req.Gateway.Attributes,
			ClusterAddress:    req.Gateway.ClusterAddress,
			ContactAccount:    req.Gateway.ContactAccount,
		})
		if err != nil {
			return err
		}

		return tx.Gateways.SetCollaborator(&ttnpb.GatewayCollaborator{
			GatewayIdentifier: req.Gateway.GatewayIdentifier,
			UserIdentifier:    ttnpb.UserIdentifier{userID},
			Rights:            ttnpb.AllGatewayRights,
		})
	})

	return nil, err
}

// GetGateway returns a gateway information.
func (s *gatewayService) GetGateway(ctx context.Context, req *ttnpb.GatewayIdentifier) (*ttnpb.Gateway, error) {
	err := s.enforceGatewayRights(ctx, req.GatewayID, ttnpb.RIGHT_GATEWAY_INFO)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Gateways.GetByID(req.GatewayID, s.config.Factories.Gateway)
	if err != nil {
		return nil, err
	}

	return found.GetGateway(), nil
}

// ListGateways returns all the gateways the current user is collaborator of.
func (s *gatewayService) ListGateways(ctx context.Context, req *ttnpb.ListGatewaysRequest) (*ttnpb.ListGatewaysResponse, error) {
	userID, err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_GATEWAYS_LIST)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Gateways.ListByUser(userID, s.config.Factories.Gateway)
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
	err := s.enforceGatewayRights(ctx, req.Gateway.GatewayID, ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC)
	if err != nil {
		return nil, err
	}

	err = s.store.Transact(func(tx *store.Store) error {
		found, err := tx.Gateways.GetByID(req.Gateway.GatewayID, s.config.Factories.Gateway)
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
			case ttnpb.FieldPathGatewayContactAccountUserID.MatchString(path):
				gtw.ContactAccount.UserID = req.Gateway.ContactAccount.UserID
			default:
				return ttnpb.ErrInvalidPathUpdateMask.New(errors.Attributes{
					"path": path,
				})
			}
		}

		return tx.Gateways.Update(gtw)
	})

	return nil, err
}

// DeleteGateway deletes a gateway.
func (s *gatewayService) DeleteGateway(ctx context.Context, req *ttnpb.GatewayIdentifier) (*pbtypes.Empty, error) {
	err := s.enforceGatewayRights(ctx, req.GatewayID, ttnpb.RIGHT_GATEWAY_DELETE)
	if err != nil {
		return nil, err
	}

	return nil, s.store.Gateways.Delete(req.GatewayID)
}

// GenerateGatewayAPIKey generates a gateway API key and returns it.
func (s *gatewayService) GenerateGatewayAPIKey(ctx context.Context, req *ttnpb.GenerateGatewayAPIKeyRequest) (*ttnpb.APIKey, error) {
	err := s.enforceGatewayRights(ctx, req.GatewayID, ttnpb.RIGHT_GATEWAY_SETTINGS_KEYS)
	if err != nil {
		return nil, err
	}

	k, err := auth.GenerateGatewayAPIKey(s.config.Hostname)
	if err != nil {
		return nil, err
	}

	key := &ttnpb.APIKey{
		Key:    k,
		Name:   req.Name,
		Rights: req.Rights,
	}

	err = s.store.Gateways.SaveAPIKey(req.GatewayID, key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// ListGatewayAPIKeys list all the API keys from a gateway.
func (s *gatewayService) ListGatewayAPIKeys(ctx context.Context, req *ttnpb.GatewayIdentifier) (*ttnpb.ListGatewayAPIKeysResponse, error) {
	err := s.enforceGatewayRights(ctx, req.GatewayID, ttnpb.RIGHT_GATEWAY_SETTINGS_KEYS)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Gateways.ListAPIKeys(req.GatewayID)
	if err != nil {
		return nil, err
	}

	return &ttnpb.ListGatewayAPIKeysResponse{
		APIKeys: found,
	}, nil
}

// UpdateGatewayAPIKey updates an API key rights.
func (s *gatewayService) UpdateGatewayAPIKey(ctx context.Context, req *ttnpb.UpdateGatewayAPIKeyRequest) (*pbtypes.Empty, error) {
	err := s.enforceGatewayRights(ctx, req.GatewayID, ttnpb.RIGHT_GATEWAY_SETTINGS_KEYS)
	if err != nil {
		return nil, err
	}

	return nil, s.store.Gateways.UpdateAPIKeyRights(req.GatewayID, req.Name, req.Rights)
}

// RemoveGatewayAPIKey removes a gateway API key.
func (s *gatewayService) RemoveGatewayAPIKey(ctx context.Context, req *ttnpb.RemoveGatewayAPIKeyRequest) (*pbtypes.Empty, error) {
	err := s.enforceGatewayRights(ctx, req.GatewayID, ttnpb.RIGHT_GATEWAY_SETTINGS_KEYS)
	if err != nil {
		return nil, err
	}

	return nil, s.store.Gateways.DeleteAPIKey(req.GatewayID, req.Name)
}

// SetGatewayCollaborator sets or unsets a gateway collaborator. It returns error
// if after unset a collaborators there is no at least one collaborator with
// `RIGHT_GATEWAY_SETTINGS_COLLABORATORS` right.
func (s *gatewayService) SetGatewayCollaborator(ctx context.Context, req *ttnpb.GatewayCollaborator) (*pbtypes.Empty, error) {
	err := s.enforceGatewayRights(ctx, req.GatewayID, ttnpb.RIGHT_GATEWAY_SETTINGS_COLLABORATORS)
	if err != nil {
		return nil, err
	}

	err = s.store.Transact(func(tx *store.Store) error {
		err := tx.Gateways.SetCollaborator(req)
		if err != nil {
			return err
		}

		// check that there is at least one collaborator in with SETTINGS_COLLABORATOR right
		collaborators, err := tx.Gateways.ListCollaborators(req.GatewayID, ttnpb.RIGHT_GATEWAY_SETTINGS_COLLABORATORS)
		if err != nil {
			return err
		}

		if len(collaborators) == 0 {
			return ErrSetGatewayCollaboratorFailed.New(errors.Attributes{
				"gateway_id": req.GatewayID,
			})
		}

		return nil
	})

	return nil, err
}

// ListGatewayCollaborators returns all the collaborators that a gateway has.
func (s *gatewayService) ListGatewayCollaborators(ctx context.Context, req *ttnpb.GatewayIdentifier) (*ttnpb.ListGatewayCollaboratorsResponse, error) {
	err := s.enforceGatewayRights(ctx, req.GatewayID, ttnpb.RIGHT_GATEWAY_SETTINGS_COLLABORATORS)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Gateways.ListCollaborators(req.GatewayID)
	if err != nil {
		return nil, err
	}

	return &ttnpb.ListGatewayCollaboratorsResponse{
		Collaborators: found,
	}, nil
}

// ListGatewayRights returns the rights the caller user has to a gateway.
func (s *gatewayService) ListGatewayRights(ctx context.Context, req *ttnpb.GatewayIdentifier) (*ttnpb.ListGatewayRightsResponse, error) {
	claims, err := s.claimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	resp := new(ttnpb.ListGatewayRightsResponse)

	switch claims.Source {
	case auth.Token:
		userID := claims.UserID()

		rights, err := s.store.Gateways.ListUserRights(req.GatewayID, userID)
		if err != nil {
			return nil, err
		}

		// result rights are the intersection between the scope of the Client
		// and the rights that the user has to the application.
		resp.Rights = util.RightsIntersection(claims.Rights, rights)
	case auth.Key:
		if claims.GatewayID() != req.GatewayID {
			return nil, ErrNotAuthorized.New(nil)
		}

		resp.Rights = claims.Rights
	}

	return resp, nil
}
