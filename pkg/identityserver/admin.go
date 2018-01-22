// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type adminService struct {
	*IdentityServer
}

// GetSettings fetches the current dynamic settings of the Identity Server.
func (s *adminService) GetSettings(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.IdentityServerSettings, error) {
	err := s.enforceAdmin(ctx)
	if err != nil {
		return nil, err
	}

	settings, err := s.store.Settings.Get()
	if err != nil {
		return nil, err
	}

	return settings, nil
}

// UpdateSettings updates the dynamic settings.
func (s *adminService) UpdateSettings(ctx context.Context, req *ttnpb.UpdateSettingsRequest) (*pbtypes.Empty, error) {
	if err := s.enforceAdmin(ctx); err != nil {
		return nil, err
	}

	settings, err := s.store.Settings.Get()
	if err != nil {
		return nil, err
	}

	for _, path := range req.UpdateMask.Paths {
		switch {
		case ttnpb.FieldPathSettingsBlacklistedIDs.MatchString(path):
			if req.Settings.BlacklistedIDs == nil {
				req.Settings.BlacklistedIDs = []string{}
			}
			settings.BlacklistedIDs = req.Settings.BlacklistedIDs
		case ttnpb.FieldPathSettingsUserRegistrationSkipValidation.MatchString(path):
			settings.SkipValidation = req.Settings.SkipValidation
		case ttnpb.FieldPathSettingsUserRegistrationSelfRegistration.MatchString(path):
			settings.SelfRegistration = req.Settings.SelfRegistration
		case ttnpb.FieldPathSettingsUserRegistrationAdminApproval.MatchString(path):
			settings.AdminApproval = req.Settings.AdminApproval
		case ttnpb.FieldPathSettingsValidationTokenTTL.MatchString(path):
			settings.ValidationTokenTTL = req.Settings.ValidationTokenTTL
		case ttnpb.FieldPathSettingsAllowedEmails.MatchString(path):
			if req.Settings.AllowedEmails == nil {
				req.Settings.AllowedEmails = []string{}
			}
			settings.AllowedEmails = req.Settings.AllowedEmails
		default:
			return nil, ttnpb.ErrInvalidPathUpdateMask.New(errors.Attributes{
				"path": path,
			})
		}
	}

	return nil, s.store.Settings.Set(settings)
}

func (s *adminService) CreateUser(ctx context.Context, req *ttnpb.CreateUserRequest) (*pbtypes.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *adminService) GetUser(ctx context.Context, req *ttnpb.UserIdentifier) (*ttnpb.User, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *adminService) ListUsers(ctx context.Context, req *ttnpb.ListUsersRequest) (*ttnpb.ListUsersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *adminService) UpdateUser(ctx context.Context, req *ttnpb.UpdateUserRequest) (*pbtypes.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *adminService) ResetUserPassword(ctx context.Context, req *ttnpb.UserIdentifier) (*pbtypes.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *adminService) DeleteUser(ctx context.Context, req *ttnpb.UserIdentifier) (*pbtypes.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *adminService) SendInvitation(ctx context.Context, req *ttnpb.SendInvitationRequest) (*pbtypes.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *adminService) ListInvitations(ctx context.Context, req *ttnpb.ListInvitationsRequest) (*ttnpb.ListInvitationsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *adminService) RevokeInvitation(ctx context.Context, req *ttnpb.RevokeInvitationRequest) (*pbtypes.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *adminService) GetClient(ctx context.Context, req *ttnpb.ClientIdentifier) (*ttnpb.Client, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *adminService) ListClients(ctx context.Context, req *ttnpb.ListClientsRequest) (*ttnpb.ListClientsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *adminService) UpdateClient(ctx context.Context, req *ttnpb.UpdateClientRequest) (*pbtypes.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (s *adminService) DeleteClient(ctx context.Context, req *ttnpb.ClientIdentifier) (*pbtypes.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}
