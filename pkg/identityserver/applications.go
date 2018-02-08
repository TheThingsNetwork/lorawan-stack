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

type applicationService struct {
	*IdentityServer
}

// CreateApplication creates an application and sets the user as collaborator
// with all possible rights.
func (s *applicationService) CreateApplication(ctx context.Context, req *ttnpb.CreateApplicationRequest) (*pbtypes.Empty, error) {
	var id ttnpb.OrganizationOrUserIdentifier

	if req.OrganizationID != "" {
		err := s.enforceOrganizationRights(ctx, req.OrganizationID, ttnpb.RIGHT_ORGANIZATION_APPLICATIONS_CREATE)
		if err != nil {
			return nil, err
		}

		id = ttnpb.OrganizationOrUserIdentifier{ID: &ttnpb.OrganizationOrUserIdentifier_OrganizationID{req.OrganizationID}}
	} else {
		userID, err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_APPLICATIONS_CREATE)
		if err != nil {
			return nil, err
		}

		id = ttnpb.OrganizationOrUserIdentifier{ID: &ttnpb.OrganizationOrUserIdentifier_UserID{userID}}
	}

	err := s.store.Transact(func(tx *store.Store) error {
		settings, err := tx.Settings.Get()
		if err != nil {
			return err
		}

		// check for blacklisted ids
		if !util.IsIDAllowed(req.Application.ApplicationID, settings.BlacklistedIDs) {
			return ErrBlacklistedID.New(errors.Attributes{
				"id": req.Application.ApplicationID,
			})
		}

		err = tx.Applications.Create(&ttnpb.Application{
			ApplicationIdentifier: req.Application.ApplicationIdentifier,
			Description:           req.Application.Description,
		})
		if err != nil {
			return err
		}

		return tx.Applications.SetCollaborator(&ttnpb.ApplicationCollaborator{
			ApplicationIdentifier:        req.Application.ApplicationIdentifier,
			OrganizationOrUserIdentifier: id,
			Rights: ttnpb.AllApplicationRights(),
		})
	})

	return nil, err
}

// GetApplication returns an application.
func (s *applicationService) GetApplication(ctx context.Context, req *ttnpb.ApplicationIdentifier) (*ttnpb.Application, error) {
	err := s.enforceApplicationRights(ctx, req.ApplicationID, ttnpb.RIGHT_APPLICATION_INFO)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Applications.GetByID(req.ApplicationID, s.config.Specializers.Application)
	if err != nil {
		return nil, err
	}

	return found.GetApplication(), nil
}

// ListApplications returns all applications where the user is collaborator.
func (s *applicationService) ListApplications(ctx context.Context, req *ttnpb.ListApplicationsRequest) (*ttnpb.ListApplicationsResponse, error) {
	var id string
	var err error

	if id = req.OrganizationID; id != "" {
		err = s.enforceOrganizationRights(ctx, req.OrganizationID, ttnpb.RIGHT_ORGANIZATION_APPLICATIONS_LIST)
	} else {
		id, err = s.enforceUserRights(ctx, ttnpb.RIGHT_USER_APPLICATIONS_LIST)
	}

	if err != nil {
		return nil, err
	}

	found, err := s.store.Applications.ListByOrganizationOrUser(id, s.config.Specializers.Application)
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
	err := s.enforceApplicationRights(ctx, req.Application.ApplicationID, ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC)
	if err != nil {
		return nil, err
	}

	err = s.store.Transact(func(tx *store.Store) error {
		found, err := tx.Applications.GetByID(req.Application.ApplicationID, s.config.Specializers.Application)
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

		return tx.Applications.Update(application)
	})

	return nil, err
}

// DeleteApplication deletes an application.
func (s *applicationService) DeleteApplication(ctx context.Context, req *ttnpb.ApplicationIdentifier) (*pbtypes.Empty, error) {
	err := s.enforceApplicationRights(ctx, req.ApplicationID, ttnpb.RIGHT_APPLICATION_DELETE)
	if err != nil {
		return nil, err
	}

	return nil, s.store.Applications.Delete(req.ApplicationID)
}

// GenerateApplicationAPIKey generates an application API key and returns it.
func (s *applicationService) GenerateApplicationAPIKey(ctx context.Context, req *ttnpb.GenerateApplicationAPIKeyRequest) (*ttnpb.APIKey, error) {
	err := s.enforceApplicationRights(ctx, req.ApplicationID, ttnpb.RIGHT_APPLICATION_SETTINGS_KEYS)
	if err != nil {
		return nil, err
	}

	k, err := auth.GenerateApplicationAPIKey(s.config.Hostname)
	if err != nil {
		return nil, err
	}

	key := &ttnpb.APIKey{
		Key:    k,
		Name:   req.Name,
		Rights: req.Rights,
	}

	err = s.store.Applications.SaveAPIKey(req.ApplicationID, key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// ListApplicationAPIKeys list all the API keys of an application.
func (s *applicationService) ListApplicationAPIKeys(ctx context.Context, req *ttnpb.ApplicationIdentifier) (*ttnpb.ListApplicationAPIKeysResponse, error) {
	err := s.enforceApplicationRights(ctx, req.ApplicationID, ttnpb.RIGHT_APPLICATION_SETTINGS_KEYS)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Applications.ListAPIKeys(req.ApplicationID)
	if err != nil {
		return nil, err
	}

	return &ttnpb.ListApplicationAPIKeysResponse{
		APIKeys: found,
	}, nil
}

// UpdateApplicationAPIKey updates the rights of an application API key.
func (s *applicationService) UpdateApplicationAPIKey(ctx context.Context, req *ttnpb.UpdateApplicationAPIKeyRequest) (*pbtypes.Empty, error) {
	err := s.enforceApplicationRights(ctx, req.ApplicationID, ttnpb.RIGHT_APPLICATION_SETTINGS_KEYS)
	if err != nil {
		return nil, err
	}

	return nil, s.store.Applications.UpdateAPIKeyRights(req.ApplicationID, req.Name, req.Rights)
}

// RemoveApplicationAPIKey removes an application API key.
func (s *applicationService) RemoveApplicationAPIKey(ctx context.Context, req *ttnpb.RemoveApplicationAPIKeyRequest) (*pbtypes.Empty, error) {
	err := s.enforceApplicationRights(ctx, req.ApplicationID, ttnpb.RIGHT_APPLICATION_SETTINGS_KEYS)
	if err != nil {
		return nil, err
	}

	return nil, s.store.Applications.DeleteAPIKey(req.ApplicationID, req.Name)
}

// SetApplicationCollaborator allows to set and unset an application collaborator.
// It fails if after unset a collaborator there is no at least one collaborator
// with `RIGHT_APPLICATION_SETTINGS_COLLABORATORS` right.
func (s *applicationService) SetApplicationCollaborator(ctx context.Context, req *ttnpb.ApplicationCollaborator) (*pbtypes.Empty, error) {
	err := s.enforceApplicationRights(ctx, req.ApplicationID, ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS)
	if err != nil {
		return nil, err
	}

	err = s.store.Transact(func(tx *store.Store) error {
		err := tx.Applications.SetCollaborator(req)
		if err != nil {
			return err
		}

		// check that there is at least one collaborator in with SETTINGS_COLLABORATOR right
		collaborators, err := tx.Applications.ListCollaborators(req.ApplicationID, ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS)
		if err != nil {
			return err
		}

		if len(collaborators) == 0 {
			return ErrSetApplicationCollaboratorFailed.New(errors.Attributes{
				"application_id": req.ApplicationID,
			})
		}

		return nil
	})

	return nil, err
}

// ListApplicationCollaborators returns all the collaborators from an application.
func (s *applicationService) ListApplicationCollaborators(ctx context.Context, req *ttnpb.ApplicationIdentifier) (*ttnpb.ListApplicationCollaboratorsResponse, error) {
	err := s.enforceApplicationRights(ctx, req.ApplicationID, ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Applications.ListCollaborators(req.ApplicationID)
	if err != nil {
		return nil, err
	}

	return &ttnpb.ListApplicationCollaboratorsResponse{
		Collaborators: found,
	}, nil
}

// ListApplicationRights returns the rights the caller user has to an application.
func (s *applicationService) ListApplicationRights(ctx context.Context, req *ttnpb.ApplicationIdentifier) (*ttnpb.ListApplicationRightsResponse, error) {
	claims, err := s.claimsFromContext(ctx)
	if err != nil {
		return nil, err
	}

	resp := new(ttnpb.ListApplicationRightsResponse)

	switch claims.Source {
	case auth.Token:
		userID := claims.UserID()

		rights, err := s.store.Applications.ListCollaboratorRights(req.ApplicationID, userID)
		if err != nil {
			return nil, err
		}

		// result rights are the intersection between the scope of the Client
		// and the rights that the user has to the application.
		resp.Rights = util.RightsIntersection(claims.Rights, rights)
	case auth.Key:
		if claims.ApplicationID() != req.ApplicationID {
			return nil, ErrNotAuthorized.New(nil)
		}

		resp.Rights = claims.Rights
	}

	return resp, nil
}
