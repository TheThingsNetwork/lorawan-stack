// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package api

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
)

var _ ttnpb.IsApplicationServer = new(GRPC)

// CreateApplication creates an application and sets the user as collaborator
// with all possible rights.
func (g *GRPC) CreateApplication(ctx context.Context, req *ttnpb.CreateApplicationRequest) (*pbtypes.Empty, error) {
	userID, err := g.userCheck(ctx, ttnpb.RIGHT_USER_APPLICATIONS_CREATE)
	if err != nil {
		return nil, err
	}

	settings, err := g.store.Settings.Get()
	if err != nil {
		return nil, err
	}

	// check for blacklisted ids
	if !settings.IsIDAllowed(req.Application.ApplicationID) {
		return nil, ErrBlacklistedID.New(errors.Attributes{
			"id": req.Application.ApplicationID,
		})
	}

	err = g.store.Transact(func(s *store.Store) error {
		err := s.Applications.Create(&ttnpb.Application{
			ApplicationIdentifier: req.Application.ApplicationIdentifier,
			Description:           req.Application.Description,
		})
		if err != nil {
			return err
		}

		return s.Applications.SetCollaborator(&ttnpb.ApplicationCollaborator{
			ApplicationIdentifier: req.Application.ApplicationIdentifier,
			UserIdentifier:        ttnpb.UserIdentifier{userID},
			Rights:                ttnpb.AllApplicationRights,
		})
	})

	return nil, err
}

// GetApplication returns an application.
func (g *GRPC) GetApplication(ctx context.Context, req *ttnpb.ApplicationIdentifier) (*ttnpb.Application, error) {
	err := g.applicationCheck(ctx, req.ApplicationID, ttnpb.RIGHT_APPLICATION_INFO)
	if err != nil {
		return nil, err
	}

	found, err := g.store.Applications.GetByID(req.ApplicationID, g.factories.application)
	if err != nil {
		return nil, err
	}

	return found.GetApplication(), err
}

// ListApplications returns all applications where the user is collaborator.
func (g *GRPC) ListApplications(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.ListApplicationsResponse, error) {
	userID, err := g.userCheck(ctx, ttnpb.RIGHT_USER_APPLICATIONS_LIST)
	if err != nil {
		return nil, err
	}

	found, err := g.store.Applications.ListByUser(userID, g.factories.application)
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
func (g *GRPC) UpdateApplication(ctx context.Context, req *ttnpb.UpdateApplicationRequest) (*pbtypes.Empty, error) {
	err := g.applicationCheck(ctx, req.Application.ApplicationID, ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC)
	if err != nil {
		return nil, err
	}

	found, err := g.store.Applications.GetByID(req.Application.ApplicationID, g.factories.application)
	if err != nil {
		return nil, err
	}

	for _, path := range req.UpdateMask.Paths {
		switch true {
		case ttnpb.FieldPathApplicationDescription.MatchString(path):
			found.GetApplication().Description = req.Application.Description
		default:
			return nil, ttnpb.ErrInvalidPathUpdateMask.New(errors.Attributes{
				"path": path,
			})
		}
	}

	return nil, g.store.Applications.Update(found)
}

// DeleteApplication deletes an application.
func (g *GRPC) DeleteApplication(ctx context.Context, req *ttnpb.ApplicationIdentifier) (*pbtypes.Empty, error) {
	err := g.applicationCheck(ctx, req.ApplicationID, ttnpb.RIGHT_APPLICATION_DELETE)
	if err != nil {
		return nil, err
	}

	return nil, g.store.Applications.Delete(req.ApplicationID)
}

// GenerateApplicationKey generates an application API key and returns it.
func (g *GRPC) GenerateApplicationAPIKey(ctx context.Context, req *ttnpb.GenerateApplicationAPIKeyRequest) (*ttnpb.APIKey, error) {
	err := g.applicationCheck(ctx, req.ApplicationID, ttnpb.RIGHT_APPLICATION_SETTINGS_KEYS)
	if err != nil {
		return nil, err
	}

	// TODO(gomezjdaniel): add issuer
	k, err := auth.GenerateApplicationAPIKey("")
	if err != nil {
		return nil, err
	}

	key := &ttnpb.APIKey{
		Key:    k,
		Name:   req.Name,
		Rights: req.Rights,
	}

	err = g.store.Applications.SaveAPIKey(req.ApplicationID, key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// ListApplicationAPIKeys list all the API keys of an application.
func (g *GRPC) ListApplicationAPIKeys(ctx context.Context, req *ttnpb.ApplicationIdentifier) (*ttnpb.ListApplicationAPIKeysResponse, error) {
	err := g.applicationCheck(ctx, req.ApplicationID, ttnpb.RIGHT_APPLICATION_SETTINGS_KEYS)
	if err != nil {
		return nil, err
	}

	found, err := g.store.Applications.ListAPIKeys(req.ApplicationID)
	if err != nil {
		return nil, err
	}

	return &ttnpb.ListApplicationAPIKeysResponse{
		APIKeys: found,
	}, nil
}

// UpdateApplicationAPIKey updates the rights of an application API key.
func (g *GRPC) UpdateApplicationAPIKey(ctx context.Context, req *ttnpb.UpdateApplicationAPIKeyRequest) (*pbtypes.Empty, error) {
	err := g.applicationCheck(ctx, req.ApplicationID, ttnpb.RIGHT_APPLICATION_SETTINGS_KEYS)
	if err != nil {
		return nil, err
	}

	return nil, g.store.Applications.UpdateAPIKeyRights(req.ApplicationID, req.Name, req.Rights)
}

// RemoveApplicationAPIKey removes an application API key.
func (g *GRPC) RemoveApplicationAPIKey(ctx context.Context, req *ttnpb.RemoveApplicationAPIKeyRequest) (*pbtypes.Empty, error) {
	err := g.applicationCheck(ctx, req.ApplicationID, ttnpb.RIGHT_APPLICATION_SETTINGS_KEYS)
	if err != nil {
		return nil, err
	}

	return nil, g.store.Applications.DeleteAPIKey(req.ApplicationID, req.Name)
}

// SetApplicationCollaborators allows to set and unset an application collaborator.
// It fails if after unset a collaborator there are no at least one collaborator
// with `application:settings:collaborators` right.
func (g *GRPC) SetApplicationCollaborator(ctx context.Context, req *ttnpb.ApplicationCollaborator) (*pbtypes.Empty, error) {
	err := g.applicationCheck(ctx, req.ApplicationID, ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS)
	if err != nil {
		return nil, err
	}

	err = g.store.Transact(func(s *store.Store) error {
		err := s.Applications.SetCollaborator(req)
		if err != nil {
			return err
		}

		// check that there is at least one collaborator in with SETTINGS_COLLABORATOR right
		collaborators, err := s.Applications.ListCollaborators(req.ApplicationID, ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS)
		if err != nil {
			return err
		}

		if len(collaborators) == 0 {
			return errors.Errorf("failed to unset collaborator: `%s` must have at least one collaborator with `application:settings:collaborators right", req.ApplicationID)
		}

		return nil
	})

	return nil, err
}

// ListApplicationCollaborators returns all the collaborators from an application.
func (g *GRPC) ListApplicationCollaborators(ctx context.Context, req *ttnpb.ApplicationIdentifier) (*ttnpb.ListApplicationCollaboratorsResponse, error) {
	err := g.applicationCheck(ctx, req.ApplicationID, ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS)
	if err != nil {
		return nil, err
	}

	found, err := g.store.Applications.ListCollaborators(req.ApplicationID)
	if err != nil {
		return nil, err
	}

	return &ttnpb.ListApplicationCollaboratorsResponse{
		Collaborators: found,
	}, nil
}

// ListApplicationRights returns the rights the caller user has to an application.
func (g *GRPC) ListApplicationRights(ctx context.Context, req *ttnpb.ApplicationIdentifier) (*ttnpb.ListApplicationRightsResponse, error) {
	userID, err := g.userCheck(ctx)
	if err != nil {
		return nil, err
	}

	rights, err := g.store.Applications.ListUserRights(req.ApplicationID, userID)
	if err != nil {
		return nil, err
	}

	return &ttnpb.ListApplicationRightsResponse{
		Rights: rights,
	}, nil
}
