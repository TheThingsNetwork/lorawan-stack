// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/claims"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
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

		err = tx.Organizations.Create(&ttnpb.Organization{
			OrganizationIdentifier: req.Organization.OrganizationIdentifier,
			Name:        req.Organization.Name,
			Description: req.Organization.Description,
			URL:         req.Organization.URL,
			Location:    req.Organization.Location,
			Email:       req.Organization.Email,
		})
		if err != nil {
			return err
		}

		return tx.Organizations.SetMember(&ttnpb.OrganizationMember{
			OrganizationIdentifier: req.Organization.OrganizationIdentifier,
			UserIdentifier:         ttnpb.UserIdentifier{UserID: claims.FromContext(ctx).UserID()},
			Rights:                 ttnpb.AllOrganizationRights(),
		})
	})

	return nil, err
}

// GetOrganization returns the organization that matches the identifier.
func (s *organizationService) GetOrganization(ctx context.Context, req *ttnpb.OrganizationIdentifier) (*ttnpb.Organization, error) {
	err := s.enforceOrganizationRights(ctx, req.OrganizationID, ttnpb.RIGHT_ORGANIZATION_INFO)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Organizations.GetByID(req.OrganizationID, s.config.Specializers.Organization)
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

	found, err := s.store.Organizations.ListByUser(claims.FromContext(ctx).UserID(), s.config.Specializers.Organization)
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
	err := s.enforceOrganizationRights(ctx, req.Organization.OrganizationID, ttnpb.RIGHT_ORGANIZATION_SETTINGS_BASIC)
	if err != nil {
		return nil, err
	}

	err = s.store.Transact(func(tx *store.Store) error {
		found, err := tx.Organizations.GetByID(req.Organization.OrganizationID, s.config.Specializers.Organization)
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

		return tx.Organizations.Update(organization)
	})

	return nil, err
}

// DeleteOrganization deletes an organization.
func (s *organizationService) DeleteOrganization(ctx context.Context, req *ttnpb.OrganizationIdentifier) (*pbtypes.Empty, error) {
	err := s.enforceOrganizationRights(ctx, req.OrganizationID, ttnpb.RIGHT_ORGANIZATION_DELETE)
	if err != nil {
		return nil, err
	}

	return nil, s.store.Organizations.Delete(req.OrganizationID)
}

// GenerateOrganizationAPIKey generates an organization API key and returns it.
func (s *organizationService) GenerateOrganizationAPIKey(ctx context.Context, req *ttnpb.GenerateOrganizationAPIKeyRequest) (*ttnpb.APIKey, error) {
	err := s.enforceOrganizationRights(ctx, req.OrganizationID, ttnpb.RIGHT_ORGANIZATION_SETTINGS_KEYS)
	if err != nil {
		return nil, err
	}

	k, err := auth.GenerateOrganizationAPIKey(s.config.Hostname)
	if err != nil {
		return nil, err
	}

	key := &ttnpb.APIKey{
		Key:    k,
		Name:   req.Name,
		Rights: req.Rights,
	}

	err = s.store.Organizations.SaveAPIKey(req.OrganizationID, key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// ListOrganizationAPIKeys list all the API keys of an organization.
func (s *organizationService) ListOrganizationAPIKeys(ctx context.Context, req *ttnpb.OrganizationIdentifier) (*ttnpb.ListOrganizationAPIKeysResponse, error) {
	err := s.enforceOrganizationRights(ctx, req.OrganizationID, ttnpb.RIGHT_ORGANIZATION_SETTINGS_KEYS)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Organizations.ListAPIKeys(req.OrganizationID)
	if err != nil {
		return nil, err
	}

	return &ttnpb.ListOrganizationAPIKeysResponse{
		APIKeys: found,
	}, nil
}

// UpdateOrganizationAPIKey updates the rights of an organization API key.
func (s *organizationService) UpdateOrganizationAPIKey(ctx context.Context, req *ttnpb.UpdateOrganizationAPIKeyRequest) (*pbtypes.Empty, error) {
	err := s.enforceOrganizationRights(ctx, req.OrganizationID, ttnpb.RIGHT_ORGANIZATION_SETTINGS_KEYS)
	if err != nil {
		return nil, err
	}

	return nil, s.store.Organizations.UpdateAPIKeyRights(req.OrganizationID, req.Name, req.Rights)
}

// RemoveOrganizationAPIKey removes an organization API key.
func (s *organizationService) RemoveOrganizationAPIKey(ctx context.Context, req *ttnpb.RemoveOrganizationAPIKeyRequest) (*pbtypes.Empty, error) {
	err := s.enforceOrganizationRights(ctx, req.OrganizationID, ttnpb.RIGHT_ORGANIZATION_SETTINGS_KEYS)
	if err != nil {
		return nil, err
	}

	return nil, s.store.Organizations.DeleteAPIKey(req.OrganizationID, req.Name)
}

// SetOrganizationMember sets a membership between an user and an organization
// upon a given set of rights.
//
// The call will return error if after perform the operation the sum of rights
// that all members with `RIGHT_ORGANIZATION_SETTINGS_COLLABORATORS` right
// is not equal to entire set of available `RIGHT_ORGANIZATION_XXXXXX` rights.
func (s *organizationService) SetOrganizationMember(ctx context.Context, req *ttnpb.OrganizationMember) (*pbtypes.Empty, error) {
	err := s.enforceOrganizationRights(ctx, req.OrganizationID, ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS)
	if err != nil {
		return nil, err
	}

	// The resulting set of rights is determined by the following steps:
	// 1. Get rights of target user/organization to the organization
	// 2. The set of rights the caller can modify are the rights the caller has
	//		to the organization. We refer them as modifiable rights
	// 3. The rights the caller cannot modify is the difference of the target
	//		user/organization rights minus the modifiable rights.
	// 4. The modifiable rights the target user/organization will have is the
	//		intersection between `2.` and the rights of the request
	// 5. The final set of rights is given by the sum of `2.` plus `4.`

	rights, err := s.store.Organizations.ListUserRights(req.OrganizationID, req.UserID)
	if err != nil {
		return nil, err
	}

	claims := claims.FromContext(ctx)

	// modifiable is the set of rights the caller can modify
	var modifiable []ttnpb.Right
	switch claims.Source() {
	case auth.Key:
		modifiable = claims.Rights()
	case auth.Token:
		modifiable, err = s.store.Organizations.ListUserRights(req.OrganizationID, claims.UserID())
		if err != nil {
			return nil, err
		}
	}

	req.Rights = append(ttnpb.DifferenceRights(rights, modifiable), ttnpb.IntersectRights(req.Rights, modifiable)...)

	err = s.store.Transact(func(tx *store.Store) error {
		err := tx.Organizations.SetMember(req)
		if err != nil {
			return err
		}

		// Check if the sum of rights that members with `SETTINGS_MEMBER` right is
		// equal to the entire set of defined organization rights.
		members, err := tx.Organizations.ListMembers(req.OrganizationID, ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS)
		if err != nil {
			return err
		}

		rights = ttnpb.AllOrganizationRights()
		for _, member := range members {
			rights = ttnpb.DifferenceRights(rights, member.Rights)

			if len(rights) == 0 {
				return nil
			}
		}

		return ErrSetOrganizationMemberFailed.New(errors.Attributes{
			"missing_rights": rights,
		})
	})

	return nil, err
}

// ListOrganizationMembers returns all members from the organization that
// matches the identifier.
func (s *organizationService) ListOrganizationMembers(ctx context.Context, req *ttnpb.OrganizationIdentifier) (*ttnpb.ListOrganizationMembersResponse, error) {
	err := s.enforceOrganizationRights(ctx, req.OrganizationID, ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Organizations.ListMembers(req.OrganizationID)
	if err != nil {
		return nil, err
	}

	return &ttnpb.ListOrganizationMembersResponse{
		Members: found,
	}, nil
}

// ListOrganizationRights returns the rights the caller user has to an organization.
func (s *organizationService) ListOrganizationRights(ctx context.Context, req *ttnpb.OrganizationIdentifier) (*ttnpb.ListOrganizationRightsResponse, error) {
	claims := claims.FromContext(ctx)

	resp := new(ttnpb.ListOrganizationRightsResponse)

	switch claims.Source() {
	case auth.Token:
		userID := claims.UserID()

		rights, err := s.store.Organizations.ListUserRights(req.OrganizationID, userID)
		if err != nil {
			return nil, err
		}

		// Result rights are the intersection between the scope of the Client
		// and the rights that the user has to the organization.
		resp.Rights = ttnpb.IntersectRights(claims.Rights(), rights)
	case auth.Key:
		if claims.OrganizationID() != req.OrganizationID {
			return nil, ErrNotAuthorized.New(nil)
		}

		resp.Rights = claims.Rights()
	}

	return resp, nil
}
