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

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
)

type applicationService struct {
	*IdentityServer
}

// CreateApplication creates a new application on the network and adds the
// authenticated user as collaborator with all the possible rights. If an
// organization identifier is provided the application will be created under
// the organization whose will be added as collaborator with all the possible
// rights if and only if the authenticated user is member of the organization
// with enough rights.
func (s *applicationService) CreateApplication(ctx context.Context, req *ttnpb.CreateApplicationRequest) (*pbtypes.Empty, error) {
	var id ttnpb.OrganizationOrUserIdentifiers

	if req.OrganizationID != "" {
		err := s.enforceOrganizationRights(ctx, req.OrganizationIdentifiers, ttnpb.RIGHT_ORGANIZATION_APPLICATIONS_CREATE)
		if err != nil {
			return nil, err
		}

		id = organizationOrUserIDsOrganizationIDs(req.OrganizationIdentifiers)
	} else {
		err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_APPLICATIONS_CREATE)
		if err != nil {
			return nil, err
		}

		id = organizationOrUserIDsUserIDs(authorizationDataFromContext(ctx).UserIdentifiers())
	}

	err := s.store.Transact(func(tx *store.Store) error {
		settings, err := tx.Settings.Get()
		if err != nil {
			return err
		}

		// check for blacklisted ids
		if !settings.IsIDAllowed(req.Application.ApplicationID) {
			return ErrBlacklistedID.New(errors.Attributes{
				"id": req.Application.ApplicationID,
			})
		}

		now := time.Now().UTC()

		err = tx.Applications.Create(&ttnpb.Application{
			ApplicationIdentifiers: req.Application.ApplicationIdentifiers,
			Description:            req.Application.Description,
			CreatedAt:              now,
			UpdatedAt:              now,
		})
		if err != nil {
			return err
		}

		return tx.Applications.SetCollaborator(ttnpb.ApplicationCollaborator{
			ApplicationIdentifiers:        req.Application.ApplicationIdentifiers,
			OrganizationOrUserIdentifiers: id,
			Rights: ttnpb.AllApplicationRights(),
		})
	})

	return ttnpb.Empty, err
}

// GetApplication returns an application.
func (s *applicationService) GetApplication(ctx context.Context, req *ttnpb.ApplicationIdentifiers) (*ttnpb.Application, error) {
	ids := *req

	err := s.enforceApplicationRights(ctx, ids, ttnpb.RIGHT_APPLICATION_INFO)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Applications.GetByID(ids, s.specializers.Application)
	if err != nil {
		return nil, err
	}

	return found.GetApplication(), nil
}

// ListApplications returns all applications where the user is collaborator.
func (s *applicationService) ListApplications(ctx context.Context, req *ttnpb.ListApplicationsRequest) (*ttnpb.ListApplicationsResponse, error) {
	var ids ttnpb.OrganizationOrUserIdentifiers
	var err error

	if oids := req.OrganizationIdentifiers; !oids.IsZero() {
		ids = organizationOrUserIDsOrganizationIDs(oids)
		err = s.enforceOrganizationRights(ctx, req.OrganizationIdentifiers, ttnpb.RIGHT_ORGANIZATION_APPLICATIONS_LIST)
	} else {
		ids = organizationOrUserIDsUserIDs(authorizationDataFromContext(ctx).UserIdentifiers())
		err = s.enforceUserRights(ctx, ttnpb.RIGHT_USER_APPLICATIONS_LIST)
	}

	if err != nil {
		return nil, err
	}

	found, err := s.store.Applications.ListByOrganizationOrUser(ids, s.specializers.Application)
	if err != nil {
		return nil, err
	}

	resp := &ttnpb.ListApplicationsResponse{
		Applications: make([]*ttnpb.Application, 0, len(found)),
	}

	for _, app := range found {
		resp.Applications = append(resp.Applications, app.GetApplication())
	}

	return resp, nil
}

// UpdateApplication updates an application.
func (s *applicationService) UpdateApplication(ctx context.Context, req *ttnpb.UpdateApplicationRequest) (*pbtypes.Empty, error) {
	err := s.enforceApplicationRights(ctx, req.Application.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC)
	if err != nil {
		return nil, err
	}

	err = s.store.Transact(func(tx *store.Store) error {
		found, err := tx.Applications.GetByID(req.Application.ApplicationIdentifiers, s.specializers.Application)
		if err != nil {
			return err
		}
		application := found.GetApplication()

		for _, path := range req.UpdateMask.Paths {
			switch {
			case ttnpb.FieldPathApplicationDescription.MatchString(path):
				application.Description = req.Application.Description
			default:
				return ttnpb.ErrInvalidPathUpdateMask.New(errors.Attributes{
					"path": path,
				})
			}
		}

		application.UpdatedAt = time.Now().UTC()

		return tx.Applications.Update(application)
	})

	return ttnpb.Empty, err
}

// DeleteApplication deletes an application.
func (s *applicationService) DeleteApplication(ctx context.Context, req *ttnpb.ApplicationIdentifiers) (*pbtypes.Empty, error) {
	ids := *req

	err := s.enforceApplicationRights(ctx, ids, ttnpb.RIGHT_APPLICATION_DELETE)
	if err != nil {
		return nil, err
	}

	return ttnpb.Empty, s.store.Applications.Delete(ids)
}

// GenerateApplicationAPIKey generates an application API key and returns it.
func (s *applicationService) GenerateApplicationAPIKey(ctx context.Context, req *ttnpb.GenerateApplicationAPIKeyRequest) (*ttnpb.APIKey, error) {
	err := s.enforceApplicationRights(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS)
	if err != nil {
		return nil, err
	}

	k, err := auth.GenerateApplicationAPIKey(s.config.Hostname)
	if err != nil {
		return nil, err
	}

	key := ttnpb.APIKey{
		Key:    k,
		Name:   req.Name,
		Rights: req.Rights,
	}

	err = s.store.Applications.SaveAPIKey(req.ApplicationIdentifiers, key)
	if err != nil {
		return nil, err
	}

	return &key, nil
}

// ListApplicationAPIKeys list all the API keys of an application.
func (s *applicationService) ListApplicationAPIKeys(ctx context.Context, req *ttnpb.ApplicationIdentifiers) (*ttnpb.ListApplicationAPIKeysResponse, error) {
	ids := *req

	err := s.enforceApplicationRights(ctx, ids, ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Applications.ListAPIKeys(ids)
	if err != nil {
		return nil, err
	}

	keys := make([]*ttnpb.APIKey, 0, len(found))
	for i := range found {
		keys = append(keys, &found[i])
	}

	return &ttnpb.ListApplicationAPIKeysResponse{
		APIKeys: keys,
	}, nil
}

// UpdateApplicationAPIKey updates the rights of an application API key.
func (s *applicationService) UpdateApplicationAPIKey(ctx context.Context, req *ttnpb.UpdateApplicationAPIKeyRequest) (*pbtypes.Empty, error) {
	err := s.enforceApplicationRights(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS)
	if err != nil {
		return nil, err
	}

	return ttnpb.Empty, s.store.Applications.UpdateAPIKeyRights(req.ApplicationIdentifiers, req.Name, req.Rights)
}

// RemoveApplicationAPIKey removes an application API key.
func (s *applicationService) RemoveApplicationAPIKey(ctx context.Context, req *ttnpb.RemoveApplicationAPIKeyRequest) (*pbtypes.Empty, error) {
	err := s.enforceApplicationRights(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS)
	if err != nil {
		return nil, err
	}

	return ttnpb.Empty, s.store.Applications.DeleteAPIKey(req.ApplicationIdentifiers, req.Name)
}

// SetApplicationCollaborator sets a collaborationship between an user and an
// application upon a given set of rights.
//
// The call will return error if after perform the operation the sum of rights
// that all collaborators with `RIGHT_APPLICATION_SETTINGS_COLLABORATORS` right
// is not equal to entire set of available `RIGHT_APPLICATION_XXXXXX` rights.
func (s *applicationService) SetApplicationCollaborator(ctx context.Context, req *ttnpb.ApplicationCollaborator) (*pbtypes.Empty, error) {
	err := s.enforceApplicationRights(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS)
	if err != nil {
		return nil, err
	}

	// The resulting set of rights is determined by the following steps:
	// 1. Get rights of target user/organization to the application
	// 2. The set of rights the caller can modify are the rights the caller has
	//		to the application. We refer them as modifiable rights
	// 3. The rights the caller cannot modify is the difference of the targer
	//		user/organization rights minus the modifiable rights.
	// 4. The modifiable rights the target user/organization will have is the
	//		intersection between `2.` and the rights of the request
	// 5. The final set of rights is given by the sum of `2.` plus `4.`

	rights, err := s.store.Applications.ListCollaboratorRights(req.ApplicationIdentifiers, req.OrganizationOrUserIdentifiers)
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
		modifiable, err = s.store.Applications.ListCollaboratorRights(req.ApplicationIdentifiers, organizationOrUserIDsUserIDs(ad.UserIdentifiers()))
		if err != nil {
			return nil, err
		}
	}

	req.Rights = append(ttnpb.DifferenceRights(rights, modifiable), ttnpb.IntersectRights(req.Rights, modifiable)...)

	err = s.store.Transact(func(tx *store.Store) error {
		err := tx.Applications.SetCollaborator(*req)
		if err != nil {
			return err
		}

		// Check if the sum of rights that collaborators with `SETTINGS_COLLABORATOR`
		// right is equal to the entire set of defined application rights.
		collaborators, err := tx.Applications.ListCollaborators(req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS)
		if err != nil {
			return err
		}

		rights = ttnpb.AllApplicationRights()
		for _, collaborator := range collaborators {
			rights = ttnpb.DifferenceRights(rights, collaborator.Rights)

			if len(rights) == 0 {
				return nil
			}
		}

		return ErrSetApplicationCollaboratorFailed.New(errors.Attributes{
			"missing_rights": rights,
		})
	})

	return ttnpb.Empty, err
}

// ListApplicationCollaborators returns all the collaborators from an application.
func (s *applicationService) ListApplicationCollaborators(ctx context.Context, req *ttnpb.ApplicationIdentifiers) (*ttnpb.ListApplicationCollaboratorsResponse, error) {
	ids := *req

	err := s.enforceApplicationRights(ctx, ids, ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Applications.ListCollaborators(ids)
	if err != nil {
		return nil, err
	}

	collaborators := make([]*ttnpb.ApplicationCollaborator, 0, len(found))
	for i := range found {
		collaborators = append(collaborators, &found[i])
	}

	return &ttnpb.ListApplicationCollaboratorsResponse{
		Collaborators: collaborators,
	}, nil
}

// ListApplicationRights returns the rights the caller user has to an application.
func (s *applicationService) ListApplicationRights(ctx context.Context, req *ttnpb.ApplicationIdentifiers) (*ttnpb.ListApplicationRightsResponse, error) {
	ad := authorizationDataFromContext(ctx)

	resp := new(ttnpb.ListApplicationRightsResponse)

	switch ad.Source {
	case auth.Token:
		rights, err := s.store.Applications.ListCollaboratorRights(*req, organizationOrUserIDsUserIDs(ad.UserIdentifiers()))
		if err != nil {
			return nil, err
		}

		// result rights are the intersection between the scope of the Client
		// and the rights that the user has to the application.
		resp.Rights = ttnpb.IntersectRights(ad.Rights, rights)
	case auth.Key:
		if !ad.ApplicationIdentifiers().Contains(*req) {
			return nil, ErrNotAuthorized.New(nil)
		}

		resp.Rights = ad.Rights
	}

	return resp, nil
}
