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
	"sync"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/errors/common"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// GatewayGeneratedFields are the fields that are automatically generated.
var GatewayGeneratedFields = []string{
	"CreatedAt",
	"UpdatedAt",
	"Gateway.CreatedAt",
	"Gateway.UpdatedAt",
}

type gatewayService struct {
	*IdentityServer

	*gtwConfigPusher
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

		// Check for blacklisted IDs.
		if !settings.IsIDAllowed(req.Gateway.GatewayID) {
			return ErrBlacklistedID.New(errors.Attributes{
				"id": req.Gateway.GatewayID,
			})
		}

		now := time.Now().UTC()

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
			CreatedAt:          now,
			UpdatedAt:          now,
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

	if err != nil {
		return nil, err
	}

	events.Publish(evtCreateGateway(ctx, req.GetGateway().GatewayIdentifiers, nil))

	return ttnpb.Empty, nil
}

// GetGateway returns a gateway information.
func (s *gatewayService) GetGateway(ctx context.Context, req *ttnpb.GatewayIdentifiers) (*ttnpb.Gateway, error) {
	ids := *req

	err := s.enforceGatewayRights(ctx, ids, ttnpb.RIGHT_GATEWAY_INFO)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Gateways.GetByID(ids, s.specializers.Gateway)
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

	found, err := s.store.Gateways.ListByOrganizationOrUser(ids, s.specializers.Gateway)
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

func copyGatewayFields(origin, recipient *ttnpb.Gateway, paths []string) error {
	for _, path := range paths {
		switch {
		case ttnpb.FieldPathGatewayDescription.MatchString(path):
			recipient.Description = origin.Description
		case ttnpb.FieldPathGatewayFrequencyPlanID.MatchString(path):
			recipient.FrequencyPlanID = origin.FrequencyPlanID
		case ttnpb.FieldPathGatewayPrivacySettingsStatusPublic.MatchString(path):
			recipient.PrivacySettings.StatusPublic = origin.PrivacySettings.StatusPublic
		case ttnpb.FieldPathGatewayPrivacySettingsLocationPublic.MatchString(path):
			recipient.PrivacySettings.LocationPublic = origin.PrivacySettings.LocationPublic
		case ttnpb.FieldPathGatewayPrivacySettingsContactable.MatchString(path):
			recipient.PrivacySettings.Contactable = origin.PrivacySettings.Contactable
		case ttnpb.FieldPathGatewayAutoUpdate.MatchString(path):
			recipient.AutoUpdate = origin.AutoUpdate
		case ttnpb.FieldPathGatewayPlatform.MatchString(path):
			recipient.Platform = origin.Platform
		case ttnpb.FieldPathGatewayAntennas.MatchString(path):
			if origin.Antennas == nil {
				origin.Antennas = []ttnpb.GatewayAntenna{}
			}
			recipient.Antennas = origin.Antennas
		case ttnpb.FieldPathGatewayAttributes.MatchString(path):
			attr := ttnpb.FieldPathGatewayAttributes.FindStringSubmatch(path)[1]

			if value, ok := origin.Attributes[attr]; ok && len(value) > 0 {
				recipient.Attributes[attr] = value
			} else {
				delete(recipient.Attributes, attr)
			}
		case ttnpb.FieldPathGatewayClusterAddress.MatchString(path):
			recipient.ClusterAddress = origin.ClusterAddress
		case ttnpb.FieldPathGatewayContactAccountIDs.MatchString(path):
			recipient.ContactAccountIDs = &ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_UserID{UserID: origin.ContactAccountIDs.GetUserID()}}
		case ttnpb.FieldPathGatewayDisableTxDelay.MatchString(path):
			recipient.DisableTxDelay = origin.DisableTxDelay
		default:
			return ttnpb.ErrInvalidPathUpdateMask.New(errors.Attributes{
				"path": path,
			})
		}
	}

	return nil
}

// UpdateGateway updates a gateway.
func (s *gatewayService) UpdateGateway(ctx context.Context, req *ttnpb.UpdateGatewayRequest) (*pbtypes.Empty, error) {
	err := s.enforceGatewayRights(ctx, req.Gateway.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_SETTINGS_BASIC)
	if err != nil {
		return nil, err
	}

	var gtw *ttnpb.Gateway
	err = s.store.Transact(func(tx *store.Store) error {
		found, err := tx.Gateways.GetByID(req.Gateway.GatewayIdentifiers, s.specializers.Gateway)
		if err != nil {
			return err
		}
		gtw = found.GetGateway()

		fieldsMask := req.GetUpdateMask()
		paths := fieldsMask.GetPaths()
		if err := copyGatewayFields(&req.Gateway, gtw, paths); err != nil {
			return err
		}

		gtw.UpdatedAt = time.Now().UTC()

		return tx.Gateways.Update(gtw)
	})

	if err != nil {
		return nil, err
	}

	events.Publish(evtUpdateGateway(ctx, req.GetGateway().GatewayIdentifiers, req.UpdateMask.Paths))

	return ttnpb.Empty, nil
}

// DeleteGateway deletes a gateway.
func (s *gatewayService) DeleteGateway(ctx context.Context, req *ttnpb.GatewayIdentifiers) (*pbtypes.Empty, error) {
	ids := *req

	err := s.enforceGatewayRights(ctx, ids, ttnpb.RIGHT_GATEWAY_DELETE)
	if err != nil {
		return nil, err
	}

	if err = s.store.Gateways.Delete(ids); err != nil {
		return nil, err
	}

	events.Publish(evtDeleteGateway(ctx, req, nil))

	return ttnpb.Empty, nil
}

// GenerateGatewayAPIKey generates a gateway API key and returns it.
func (s *gatewayService) GenerateGatewayAPIKey(ctx context.Context, req *ttnpb.GenerateGatewayAPIKeyRequest) (*ttnpb.APIKey, error) {
	err := s.enforceGatewayRights(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_SETTINGS_API_KEYS)
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

	events.Publish(evtGenerateGatewayAPIKey(ctx, req.GatewayIdentifiers, ttnpb.APIKey{Name: key.Name, Rights: key.Rights}))

	return &key, nil
}

// ListGatewayAPIKeys list all the API keys from a gateway.
func (s *gatewayService) ListGatewayAPIKeys(ctx context.Context, req *ttnpb.GatewayIdentifiers) (*ttnpb.ListGatewayAPIKeysResponse, error) {
	ids := *req

	err := s.enforceGatewayRights(ctx, ids, ttnpb.RIGHT_GATEWAY_SETTINGS_API_KEYS)
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
	err := s.enforceGatewayRights(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_SETTINGS_API_KEYS)
	if err != nil {
		return nil, err
	}

	s.store.Gateways.UpdateAPIKeyRights(req.GatewayIdentifiers, req.Name, req.Rights)
	if err != nil {
		return nil, err
	}

	events.Publish(evtUpdateGatewayAPIKey(ctx, req.GatewayIdentifiers, ttnpb.APIKey{Name: req.Name, Rights: req.Rights}))

	return ttnpb.Empty, nil
}

// RemoveGatewayAPIKey removes a gateway API key.
func (s *gatewayService) RemoveGatewayAPIKey(ctx context.Context, req *ttnpb.RemoveGatewayAPIKeyRequest) (*pbtypes.Empty, error) {
	err := s.enforceGatewayRights(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_SETTINGS_API_KEYS)
	if err != nil {
		return nil, err
	}

	err = s.store.Gateways.DeleteAPIKey(req.GatewayIdentifiers, req.Name)
	if err != nil {
		return nil, err
	}

	events.Publish(evtDeleteGatewayAPIKey(ctx, req.GatewayIdentifiers, ttnpb.APIKey{Name: req.Name}))

	return ttnpb.Empty, nil
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
	// 1. Get rights of target user/organization within the gateway.
	// 2. The set of rights the caller can modify are the rights the caller has
	//		within the gateway. We refer to them as modifiable rights.
	// 3. The rights the caller cannot modify are the difference of the target
	//		user/organization rights minus the modifiable rights.
	// 4. The modifiable rights the target user/organization will have are the
	//		intersection between `2.` and the rights of the request.
	// 5. The final set of rights is given by the sum of `2.` plus `4.`.

	rights, err := s.store.Gateways.ListCollaboratorRights(req.GatewayIdentifiers, req.OrganizationOrUserIdentifiers)
	if err != nil {
		return nil, err
	}

	ad := authorizationDataFromContext(ctx)

	// `modifiable` is the set of rights the caller can modify.
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

		rights, err := missingGatewayRights(tx, req.GatewayIdentifiers)
		if err != nil {
			return err
		}

		if len(rights) != 0 {
			return ErrUnmanageableGateway.New(errors.Attributes{
				"gateway_id":     req.GatewayIdentifiers.GatewayID,
				"missing_rights": rights,
			})
		}

		return nil
	})

	return ttnpb.Empty, err
}

// Checks if the sum of rights that collaborators with `SETTINGS_COLLABORATOR`
// right is equal to the entire set of defined gateway rights. Otherwise returns
// the list of missing rights.
func missingGatewayRights(tx *store.Store, ids ttnpb.GatewayIdentifiers) ([]ttnpb.Right, error) {
	collaborators, err := tx.Gateways.ListCollaborators(ids, ttnpb.RIGHT_GATEWAY_SETTINGS_COLLABORATORS)
	if err != nil {
		return nil, err
	}

	rights := ttnpb.AllGatewayRights()
	for _, collaborator := range collaborators {
		rights = ttnpb.DifferenceRights(rights, collaborator.Rights)

		if len(rights) == 0 {
			return nil, nil
		}
	}

	return rights, nil
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

		// Result rights are the intersection between the scope of the Client
		// and the rights that the user has to the gateway.
		resp.Rights = ttnpb.IntersectRights(ad.Rights, rights)
	case auth.Key:
		if !ad.GatewayIdentifiers().Contains(*req) {
			return nil, common.ErrPermissionDenied.New(nil)
		}

		resp.Rights = ad.Rights
	}

	return resp, nil
}

func (s *gatewayService) getGatewayWithFields(ids ttnpb.GatewayIdentifiers, fieldMask *pbtypes.FieldMask) (*ttnpb.Gateway, error) {
	found, err := s.store.Gateways.GetByID(ids, s.specializers.Gateway)
	if err != nil {
		return nil, err
	}
	gtw := found.GetGateway()
	if fieldMask == nil || len(fieldMask.GetPaths()) == 0 {
		return gtw, nil
	}

	toSend := &ttnpb.Gateway{}
	if err = copyGatewayFields(gtw, toSend, fieldMask.GetPaths()); err != nil {
		return nil, err
	}
	return toSend, nil
}

func (s *gatewayService) PullConfiguration(req *ttnpb.PullConfigurationRequest, stream ttnpb.GtwGr_PullConfigurationServer) error {
	ctx := stream.Context()
	ad, err := s.buildAuthorizationData(ctx)
	if err != nil {
		return err
	}

	if !ad.HasRights(ttnpb.RIGHT_GATEWAY_INFO, ttnpb.RIGHT_GATEWAY_LINK) {
		return common.ErrPermissionDenied.New(nil)
	}

	uid := ad.GatewayIdentifiers().UniqueID(ctx)

	gtw, err := s.getGatewayWithFields(ad.GatewayIdentifiers(), req.GetProjectionMask())
	if err != nil {
		return err
	}

	if err := stream.Send(gtw); err != nil {
		return err
	}

	gtwConfigs := s.gtwConfigPusher.subscribe(ctx, uid)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case _, ok := <-gtwConfigs:
			if !ok {
				return ErrOtherPullConfigurationStreamOpened.New(nil)
			}
			newConfig, err := s.getGatewayWithFields(ad.GatewayIdentifiers(), req.GetProjectionMask())
			if err != nil {
				continue
			}
			if err := stream.Send(newConfig); err != nil {
				return err
			}
		}
	}
}

type gtwConfigPusher struct {
	*gatewayService

	subscriptionsMu sync.RWMutex
	subscriptions   map[string]chan struct{}
}

func (c *gtwConfigPusher) subscribe(ctx context.Context, gatewayUID string) chan struct{} {
	newSubscription := make(chan struct{})
	c.subscriptionsMu.Lock()
	if oldSubscription, ok := c.subscriptions[gatewayUID]; ok {
		close(oldSubscription)
	}
	c.subscriptions[gatewayUID] = newSubscription
	c.subscriptionsMu.Unlock()

	return newSubscription
}

func (c *gtwConfigPusher) Notify(evt events.Event) {
	ctx := evt.Context()
	gtwIDs, ok := evt.Identifiers().(ttnpb.GatewayIdentifiers)
	if !ok {
		log.FromContext(ctx).Error("No gateway identifiers found with gateway update event")
		return
	}

	uid := gtwIDs.UniqueID(ctx)
	c.subscriptionsMu.RLock()
	subscription := c.subscriptions[uid]
	c.subscriptionsMu.RUnlock()
	select {
	case subscription <- struct{}{}:
	default:
		log.FromContext(evt.Context()).WithField("gateway_uid", uid).Error("Gateway update signal was not received by PullConfiguration subscription")
	}
}
