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
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/errors/common"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type organizationService struct {
	*IdentityServer
}

// CreateOrganization creates an organization and sets the user as member with
// with all possible rights.
func (s *organizationService) CreateOrganization(ctx context.Context, req *ttnpb.CreateOrganizationRequest) (*pbtypes.Empty, error) {
	err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_ORGANIZATIONS_CREATE)
	if err != nil {
		return nil, err
	}

	err = s.store.Transact(func(tx *store.Store) error {
		settings, err := tx.Settings.Get()
		if err != nil {
			return err
		}

		// Check for blacklisted IDs.
		if !settings.IsIDAllowed(req.Organization.OrganizationID) {
			return ErrBlacklistedID.New(errors.Attributes{
				"id": req.Organization.OrganizationID,
			})
		}

		now := time.Now().UTC()

		err = tx.Organizations.Create(&ttnpb.Organization{
			OrganizationIdentifiers: req.Organization.OrganizationIdentifiers,
			Name:        req.Organization.Name,
			Description: req.Organization.Description,
			URL:         req.Organization.URL,
			Location:    req.Organization.Location,
			Email:       req.Organization.Email,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
		if err != nil {
			return err
		}

		return tx.Organizations.SetMember(ttnpb.OrganizationMember{
			OrganizationIdentifiers: req.Organization.OrganizationIdentifiers,
			UserIdentifiers:         authorizationDataFromContext(ctx).UserIdentifiers(),
			Rights:                  ttnpb.AllOrganizationRights(),
		})
	})

	return ttnpb.Empty, err
}

// GetOrganization returns the organization that matches the identifier.
func (s *organizationService) GetOrganization(ctx context.Context, req *ttnpb.OrganizationIdentifiers) (*ttnpb.Organization, error) {
	ids := *req

	err := s.enforceOrganizationRights(ctx, ids, ttnpb.RIGHT_ORGANIZATION_INFO)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Organizations.GetByID(ids, s.specializers.Organization)
	if err != nil {
		return nil, err
	}

	return found.GetOrganization(), nil
}

// ListOrganizations returns all organizations where the user is member.
func (s *organizationService) ListOrganizations(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.ListOrganizationsResponse, error) {
	err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_ORGANIZATIONS_LIST)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Organizations.ListByUser(authorizationDataFromContext(ctx).UserIdentifiers(), s.specializers.Organization)
	if err != nil {
		return nil, err
	}

	resp := &ttnpb.ListOrganizationsResponse{
		Organizations: make([]*ttnpb.Organization, 0, len(found)),
	}

	for _, org := range found {
		resp.Organizations = append(resp.Organizations, org.GetOrganization())
	}

	return resp, nil
}

// UpdateOrganization updates an organization.
func (s *organizationService) UpdateOrganization(ctx context.Context, req *ttnpb.UpdateOrganizationRequest) (*pbtypes.Empty, error) {
	err := s.enforceOrganizationRights(ctx, req.Organization.OrganizationIdentifiers, ttnpb.RIGHT_ORGANIZATION_SETTINGS_BASIC)
	if err != nil {
		return nil, err
	}

	err = s.store.Transact(func(tx *store.Store) error {
		found, err := tx.Organizations.GetByID(req.Organization.OrganizationIdentifiers, s.specializers.Organization)
		if err != nil {
			return err
		}
		organization := found.GetOrganization()

		for _, path := range req.UpdateMask.Paths {
			switch {
			case ttnpb.FieldPathOrganizationName.MatchString(path):
				organization.Name = req.Organization.Name
			case ttnpb.FieldPathOrganizationDescription.MatchString(path):
				organization.Description = req.Organization.Description
			case ttnpb.FieldPathOrganizationURL.MatchString(path):
				organization.URL = req.Organization.URL
			case ttnpb.FieldPathOrganizationLocation.MatchString(path):
				organization.Location = req.Organization.Location
			case ttnpb.FieldPathOrganizationEmail.MatchString(path):
				organization.Email = req.Organization.Email
			default:
				return ttnpb.ErrInvalidPathUpdateMask.New(errors.Attributes{
					"path": path,
				})
			}
		}

		organization.UpdatedAt = time.Now().UTC()

		return tx.Organizations.Update(organization)
	})

	return ttnpb.Empty, err
}

// DeleteOrganization deletes an organization.
func (s *organizationService) DeleteOrganization(ctx context.Context, req *ttnpb.OrganizationIdentifiers) (*pbtypes.Empty, error) {
	ids := *req

	err := s.enforceOrganizationRights(ctx, ids, ttnpb.RIGHT_ORGANIZATION_DELETE)
	if err != nil {
		return nil, err
	}

	err = s.store.Transact(func(tx *store.Store) error {
		// We need to make sure that after deleting the organization nor an application
		// nor a gateway reaches an unmanageable state. So first in the transaction
		// before actually deleting the organization we gather the applications and
		// gateways is collaborator of so afterwards we can check for invalid states.
		apps, err := tx.Applications.ListByOrganizationOrUser(organizationOrUserIDsOrganizationIDs(ids), s.specializers.Application)
		if err != nil {
			return err
		}

		gtws, err := tx.Gateways.ListByOrganizationOrUser(organizationOrUserIDsOrganizationIDs(ids), s.specializers.Gateway)
		if err != nil {
			return err
		}

		err = tx.Organizations.Delete(ids)
		if err != nil {
			return err
		}

		for _, app := range apps {
			appIDs := app.GetApplication().ApplicationIdentifiers

			rights, err := missingApplicationRights(tx, appIDs)
			if err != nil {
				return err
			}

			if len(rights) != 0 {
				return ErrUnmanageableApplication.New(errors.Attributes{
					"application_id": appIDs.ApplicationID,
					"missing_rights": rights,
				})
			}
		}

		for _, gtw := range gtws {
			gtwIDs := gtw.GetGateway().GatewayIdentifiers

			rights, err := missingGatewayRights(tx, gtwIDs)
			if err != nil {
				return err
			}

			if len(rights) != 0 {
				return ErrUnmanageableGateway.New(errors.Attributes{
					"gateway_id":     gtwIDs.GatewayID,
					"missing_rights": rights,
				})
			}
		}

		return nil
	})

	return ttnpb.Empty, err
}

// GenerateOrganizationAPIKey generates an organization API key and returns it.
func (s *organizationService) GenerateOrganizationAPIKey(ctx context.Context, req *ttnpb.GenerateOrganizationAPIKeyRequest) (*ttnpb.APIKey, error) {
	err := s.enforceOrganizationRights(ctx, req.OrganizationIdentifiers, ttnpb.RIGHT_ORGANIZATION_SETTINGS_API_KEYS)
	if err != nil {
		return nil, err
	}

	k, err := auth.GenerateOrganizationAPIKey(s.config.Hostname)
	if err != nil {
		return nil, err
	}

	key := ttnpb.APIKey{
		Key:    k,
		Name:   req.Name,
		Rights: req.Rights,
	}

	err = s.store.Organizations.SaveAPIKey(req.OrganizationIdentifiers, key)
	if err != nil {
		return nil, err
	}

	return &key, nil
}

// ListOrganizationAPIKeys list all the API keys of an organization.
func (s *organizationService) ListOrganizationAPIKeys(ctx context.Context, req *ttnpb.OrganizationIdentifiers) (*ttnpb.ListOrganizationAPIKeysResponse, error) {
	ids := *req

	err := s.enforceOrganizationRights(ctx, ids, ttnpb.RIGHT_ORGANIZATION_SETTINGS_API_KEYS)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Organizations.ListAPIKeys(ids)
	if err != nil {
		return nil, err
	}

	keys := make([]*ttnpb.APIKey, 0, len(found))
	for i := range found {
		keys = append(keys, &found[i])
	}

	return &ttnpb.ListOrganizationAPIKeysResponse{
		APIKeys: keys,
	}, nil
}

// UpdateOrganizationAPIKey updates the rights of an organization API key.
func (s *organizationService) UpdateOrganizationAPIKey(ctx context.Context, req *ttnpb.UpdateOrganizationAPIKeyRequest) (*pbtypes.Empty, error) {
	err := s.enforceOrganizationRights(ctx, req.OrganizationIdentifiers, ttnpb.RIGHT_ORGANIZATION_SETTINGS_API_KEYS)
	if err != nil {
		return nil, err
	}

	return ttnpb.Empty, s.store.Organizations.UpdateAPIKeyRights(req.OrganizationIdentifiers, req.Name, req.Rights)
}

// RemoveOrganizationAPIKey removes an organization API key.
func (s *organizationService) RemoveOrganizationAPIKey(ctx context.Context, req *ttnpb.RemoveOrganizationAPIKeyRequest) (*pbtypes.Empty, error) {
	err := s.enforceOrganizationRights(ctx, req.OrganizationIdentifiers, ttnpb.RIGHT_ORGANIZATION_SETTINGS_API_KEYS)
	if err != nil {
		return nil, err
	}

	return ttnpb.Empty, s.store.Organizations.DeleteAPIKey(req.OrganizationIdentifiers, req.Name)
}

// SetOrganizationMember sets a membership between an user and an organization
// upon a given set of rights.
//
// The call will return error if after perform the operation the sum of rights
// that all members with `RIGHT_ORGANIZATION_SETTINGS_COLLABORATORS` right
// is not equal to entire set of available `RIGHT_ORGANIZATION_XXXXXX` rights.
func (s *organizationService) SetOrganizationMember(ctx context.Context, req *ttnpb.OrganizationMember) (*pbtypes.Empty, error) {
	err := s.enforceOrganizationRights(ctx, req.OrganizationIdentifiers, ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS)
	if err != nil {
		return nil, err
	}

	// The resulting set of rights is determined by the following steps:
	// 1. Get rights of target user within the organization.
	// 2. The set of rights the caller can modify are the rights the caller has
	//		within the organization. We refer to them as modifiable rights.
	// 3. The rights the caller cannot modify are the difference of the target
	//		user rights minus the modifiable rights.
	// 4. The modifiable rights the target user/organization will have are the
	//		intersection between `2.` and the rights of the request.
	// 5. The final set of rights is given by the sum of `2.` plus `4.`.

	rights, err := s.store.Organizations.ListMemberRights(req.OrganizationIdentifiers, req.UserIdentifiers)
	if err != nil {
		return nil, err
	}

	ad := authorizationDataFromContext(ctx)

	// Modifiable is the set of rights the caller can modify.
	var modifiable []ttnpb.Right
	switch ad.Source {
	case auth.Key:
		modifiable = ad.Rights
	case auth.Token:
		modifiable, err = s.store.Organizations.ListMemberRights(req.OrganizationIdentifiers, ad.UserIdentifiers())
		if err != nil {
			return nil, err
		}
	}

	req.Rights = append(ttnpb.DifferenceRights(rights, modifiable), ttnpb.IntersectRights(req.Rights, modifiable)...)

	err = s.store.Transact(func(tx *store.Store) error {
		err := tx.Organizations.SetMember(*req)
		if err != nil {
			return err
		}

		rights, err := missingOrganizationRights(tx, req.OrganizationIdentifiers)
		if err != nil {
			return err
		}

		if len(rights) != 0 {
			return ErrUnmanageableOrganization.New(errors.Attributes{
				"organization_id": req.OrganizationIdentifiers.OrganizationID,
				"missing_rights":  rights,
			})
		}

		return nil
	})

	return ttnpb.Empty, err
}

// Checks if the sum of rights that members with `SETTINGS_MEMBER` right is
// equal to the entire set of defined organization rights. Otherwise returns
// the list of missing rights.
func missingOrganizationRights(tx *store.Store, ids ttnpb.OrganizationIdentifiers) ([]ttnpb.Right, error) {
	members, err := tx.Organizations.ListMembers(ids, ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS)
	if err != nil {
		return nil, err
	}

	rights := ttnpb.AllOrganizationRights()
	for _, member := range members {
		rights = ttnpb.DifferenceRights(rights, member.Rights)

		if len(rights) == 0 {
			return nil, nil
		}
	}

	return rights, nil
}

// ListOrganizationMembers returns all members from the organization that
// matches the identifier.
func (s *organizationService) ListOrganizationMembers(ctx context.Context, req *ttnpb.OrganizationIdentifiers) (*ttnpb.ListOrganizationMembersResponse, error) {
	ids := *req

	err := s.enforceOrganizationRights(ctx, ids, ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Organizations.ListMembers(ids)
	if err != nil {
		return nil, err
	}

	members := make([]*ttnpb.OrganizationMember, 0, len(found))
	for i := range found {
		members = append(members, &found[i])
	}

	return &ttnpb.ListOrganizationMembersResponse{
		Members: members,
	}, nil
}

// ListOrganizationRights returns the rights the caller user has to an organization.
func (s *organizationService) ListOrganizationRights(ctx context.Context, req *ttnpb.OrganizationIdentifiers) (*ttnpb.ListOrganizationRightsResponse, error) {
	ad := authorizationDataFromContext(ctx)

	resp := new(ttnpb.ListOrganizationRightsResponse)

	switch ad.Source {
	case auth.Token:
		rights, err := s.store.Organizations.ListMemberRights(*req, ad.UserIdentifiers())
		if err != nil {
			return nil, err
		}

		// Result rights are the intersection between the scope of the Client
		// and the rights that the user has to the organization.
		resp.Rights = ttnpb.IntersectRights(ad.Rights, rights)
	case auth.Key:
		if !ad.OrganizationIdentifiers().Contains(*req) {
			return nil, common.ErrPermissionDenied.New(nil)
		}

		resp.Rights = ad.Rights
	}

	return resp, nil
}
