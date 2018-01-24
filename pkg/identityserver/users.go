// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"context"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/email/templates"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/util"
	"github.com/TheThingsNetwork/ttn/pkg/random"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/types"
	pbtypes "github.com/gogo/protobuf/types"
)

type userService struct {
	*IdentityServer
}

// CreateUser creates an user in the network.
func (s *userService) CreateUser(ctx context.Context, req *ttnpb.CreateUserRequest) (*pbtypes.Empty, error) {
	err := s.store.Transact(func(tx *store.Store) error {
		settings, err := tx.Settings.Get()
		if err != nil {
			return err
		}

		// if self-registration is disabled check that an invitation token is provided
		if !settings.SelfRegistration && req.InvitationToken == "" {
			return ErrInvitationTokenMissing.New(nil)
		}

		// check for blacklisted ids
		if !util.IsIDAllowed(req.User.UserID, settings.BlacklistedIDs) {
			return ErrBlacklistedID.New(errors.Attributes{
				"id": req.User.UserID,
			})
		}

		password, err := types.Hash(req.User.Password)
		if err != nil {
			return err
		}

		user := &ttnpb.User{
			UserIdentifier: req.User.UserIdentifier,
			Name:           req.User.Name,
			Email:          req.User.Email,
			Password:       string(password),
			State:          ttnpb.STATE_PENDING,
		}

		if settings.SkipValidation {
			user.ValidatedAt = time.Now().UTC()
		}

		if !settings.AdminApproval {
			user.State = ttnpb.STATE_APPROVED
		}

		err = tx.Users.Create(user)
		if err != nil {
			return err
		}

		// check whether the provided email is allowed or not when an invitation token
		// wasn't provided
		if req.InvitationToken == "" {
			if !util.IsEmailAllowed(req.User.Email, settings.AllowedEmails) {
				return ErrEmailAddressNotAllowed.New(errors.Attributes{
					"email":          req.User.Email,
					"allowed_emails": settings.AllowedEmails,
				})
			}
		} else {
			err = tx.Invitations.Use(req.User.Email, req.InvitationToken)
			if err != nil {
				return err
			}
		}

		if !settings.SkipValidation {
			token := &store.ValidationToken{
				ValidationToken: random.String(64),
				CreatedAt:       time.Now(),
				ExpiresIn:       int32(settings.ValidationTokenTTL.Seconds()),
			}

			err = tx.Users.SaveValidationToken(user.UserID, token)
			if err != nil {
				return err
			}

			return s.email.Send(user.Email, &templates.EmailValidation{
				OrganizationName: s.config.OrganizationName,
				PublicURL:        s.config.PublicURL,
				Token:            token.ValidationToken,
			})
		}

		return nil
	})

	return nil, err
}

// GetUser returns the account of the current user.
func (s *userService) GetUser(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.User, error) {
	userID, err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_PROFILE_READ)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Users.GetByID(userID, s.config.Factories.User)
	if err != nil {
		return nil, err
	}

	found.GetUser().Password = ""

	return found.GetUser(), nil
}

// UpdateUser updates the account of the current user.
// If email address is updated it sends an email to validate it if and only if
// the `SkipValidation` setting is disabled.
func (s *userService) UpdateUser(ctx context.Context, req *ttnpb.UpdateUserRequest) (*pbtypes.Empty, error) {
	userID, err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_PROFILE_WRITE)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Users.GetByID(userID, s.config.Factories.User)
	if err != nil {
		return nil, err
	}

	settings, err := s.store.Settings.Get()
	if err != nil {
		return nil, err
	}

	newEmail := false
	for _, path := range req.UpdateMask.Paths {
		switch {
		case ttnpb.FieldPathUserName.MatchString(path):
			found.GetUser().Name = req.User.Name
		case ttnpb.FieldPathUserEmail.MatchString(path):
			if strings.ToLower(req.User.Email) != strings.ToLower(found.GetUser().Email) {
				newEmail = true
				found.GetUser().ValidatedAt = time.Time{}
			}

			if !util.IsEmailAllowed(req.User.Email, settings.AllowedEmails) {
				return nil, ErrEmailAddressNotAllowed.New(errors.Attributes{
					"email":          req.User.Email,
					"allowed_emails": settings.AllowedEmails,
				})
			}

			found.GetUser().Email = req.User.Email
		default:
			return nil, ttnpb.ErrInvalidPathUpdateMask.New(errors.Attributes{
				"path": path,
			})
		}
	}

	if !newEmail {
		return nil, s.store.Users.Update(found)
	}

	err = s.store.Transact(func(st *store.Store) error {
		err := st.Users.Update(found)
		if err != nil {
			return err
		}

		token := &store.ValidationToken{
			ValidationToken: random.String(64),
			CreatedAt:       time.Now(),
			ExpiresIn:       int32(settings.ValidationTokenTTL.Seconds()),
		}

		err = st.Users.SaveValidationToken(userID, token)
		if err != nil {
			return err
		}

		return s.email.Send(found.GetUser().Email, &templates.EmailValidation{
			OrganizationName: s.config.OrganizationName,
			PublicURL:        s.config.PublicURL,
			Token:            token.ValidationToken,
		})
	})

	return nil, err
}

// UpdateUserPassword updates the password of the current user.
func (s *userService) UpdateUserPassword(ctx context.Context, req *ttnpb.UpdateUserPasswordRequest) (*pbtypes.Empty, error) {
	userID, err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_PROFILE_WRITE)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Users.GetByID(userID, s.config.Factories.User)
	if err != nil {
		return nil, err
	}

	matches, err := types.Password(found.GetUser().Password).Validate(req.Old)
	if err != nil {
		return nil, err
	}

	if !matches {
		return nil, ErrInvalidPassword.New(nil)
	}

	hashed, err := types.Hash(req.New)
	if err != nil {
		return nil, err
	}

	found.GetUser().Password = string(hashed)

	return nil, s.store.Users.Update(found)
}

// DeleteUser deletes the account of the current user.
func (s *userService) DeleteUser(ctx context.Context, _ *pbtypes.Empty) (*pbtypes.Empty, error) {
	userID, err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_DELETE)
	if err != nil {
		return nil, err
	}

	err = s.store.Transact(func(st *store.Store) error {
		err := st.Users.Delete(userID)
		if err != nil {
			return err
		}

		apps, err := st.Applications.ListByUser(userID, s.config.Factories.Application)
		if err != nil {
			return err
		}

		for _, app := range apps {
			appID := app.GetApplication().ApplicationID

			c, err := st.Applications.ListCollaborators(appID, ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS)
			if err != nil {
				return err
			}

			if len(c) == 0 {
				return errors.Errorf("Failed to delete user `%s`: cannot leave application `%s` without at least one collaborator with `RIGHT_APPLICATION_SETTINGS_COLLABORATORS` right", userID, appID)
			}
		}

		gtws, err := st.Gateways.ListByUser(userID, s.config.Factories.Gateway)
		if err != nil {
			return err
		}

		for _, gtw := range gtws {
			gtwID := gtw.GetGateway().GatewayID

			c, err := st.Gateways.ListCollaborators(gtwID, ttnpb.RIGHT_GATEWAY_SETTINGS_COLLABORATORS)
			if err != nil {
				return err
			}

			if len(c) == 0 {
				return errors.Errorf("Failed to delete user `%s`: cannot leave gateway `%s` without at least one collaborator with `RIGHT_GATEWAY_SETTINGS_COLLABORATORS` right", userID, gtwID)
			}
		}

		return nil
	})

	return nil, err
}

// GenerateUserAPIKey generates an user API key and returns it.
func (s *userService) GenerateUserAPIKey(ctx context.Context, req *ttnpb.GenerateUserAPIKeyRequest) (*ttnpb.APIKey, error) {
	userID, err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_KEYS)
	if err != nil {
		return nil, err
	}

	k, err := auth.GenerateUserAPIKey(s.config.Hostname)
	if err != nil {
		return nil, err
	}

	key := &ttnpb.APIKey{
		Key:    k,
		Name:   req.Name,
		Rights: req.Rights,
	}

	err = s.store.Users.SaveAPIKey(userID, key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// ListUserAPIKeys returns all the API keys from the current user.
func (s *userService) ListUserAPIKeys(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.ListUserAPIKeysResponse, error) {
	userID, err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_KEYS)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Users.ListAPIKeys(userID)
	if err != nil {
		return nil, err
	}

	return &ttnpb.ListUserAPIKeysResponse{
		APIKeys: found,
	}, nil
}

// UpdateUserAPIKey updates an API key from the current user.
func (s *userService) UpdateUserAPIKey(ctx context.Context, req *ttnpb.UpdateUserAPIKeyRequest) (*pbtypes.Empty, error) {
	userID, err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_PROFILE_WRITE)
	if err != nil {
		return nil, err
	}

	return nil, s.store.Users.UpdateAPIKeyRights(userID, req.Name, req.Rights)
}

// RemoveUserAPIKey removes an API key from the current user.
func (s *userService) RemoveUserAPIKey(ctx context.Context, req *ttnpb.RemoveUserAPIKeyRequest) (*pbtypes.Empty, error) {
	userID, err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_KEYS)
	if err != nil {
		return nil, err
	}

	return nil, s.store.Users.DeleteAPIKey(userID, req.Name)
}

// ValidateUserEmail validates the user's email with the token sent to the email.
func (s *userService) ValidateUserEmail(ctx context.Context, req *ttnpb.ValidateUserEmailRequest) (*pbtypes.Empty, error) {
	err := s.store.Transact(func(store *store.Store) error {
		userID, token, err := store.Users.GetValidationToken(req.Token)
		if err != nil {
			return err
		}

		if token.IsExpired() {
			return ErrValidationTokenExpired.New(nil)
		}

		user, err := store.Users.GetByID(userID, s.config.Factories.User)
		if err != nil {
			return err
		}

		user.GetUser().ValidatedAt = time.Now()

		err = store.Users.Update(user)
		if err != nil {
			return err
		}

		return store.Users.DeleteValidationToken(req.Token)
	})

	return nil, err
}

// RequestUserEmailValidation requests a new validation email if the user's email
// isn't validated yet.
func (s *userService) RequestUserEmailValidation(ctx context.Context, _ *pbtypes.Empty) (*pbtypes.Empty, error) {
	userID, err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_PROFILE_WRITE)
	if err != nil {
		return nil, err
	}

	user, err := s.store.Users.GetByID(userID, s.config.Factories.User)
	if err != nil {
		return nil, err
	}

	if !user.GetUser().ValidatedAt.IsZero() {
		return nil, ErrEmailAlreadyValidated.New(errors.Attributes{
			"email": user.GetUser().Email,
		})
	}

	settings, err := s.store.Settings.Get()
	if err != nil {
		return nil, err
	}

	token := &store.ValidationToken{
		ValidationToken: random.String(64),
		CreatedAt:       time.Now(),
		ExpiresIn:       int32(settings.ValidationTokenTTL.Seconds()),
	}

	err = s.store.Users.SaveValidationToken(userID, token)
	if err != nil {
		return nil, err
	}

	return nil, s.email.Send(user.GetUser().Email, &templates.EmailValidation{
		OrganizationName: s.config.OrganizationName,
		PublicURL:        s.config.PublicURL,
		Token:            token.ValidationToken,
	})
}

// ListAuthorizedClients returns all the authorized third-party clients that
// the current user has.
func (s *userService) ListAuthorizedClients(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.ListAuthorizedClientsResponse, error) {
	userID, err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_AUTHORIZEDCLIENTS)
	if err != nil {
		return nil, err
	}

	found, err := s.store.OAuth.ListAuthorizedClients(userID, s.config.Factories.Client)
	if err != nil {
		return nil, err
	}

	resp := &ttnpb.ListAuthorizedClientsResponse{
		Clients: make([]*ttnpb.Client, 0, len(found)),
	}

	for _, client := range found {
		cli := client.GetClient()
		cli.Secret = ""
		cli.RedirectURI = ""
		cli.Grants = nil
		resp.Clients = append(resp.Clients, cli)
	}

	return resp, nil
}

// RevokeAuthorizedClient revokes an authorized third-party client.
func (s *userService) RevokeAuthorizedClient(ctx context.Context, req *ttnpb.ClientIdentifier) (*pbtypes.Empty, error) {
	userID, err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_AUTHORIZEDCLIENTS)
	if err != nil {
		return nil, err
	}

	return nil, s.store.OAuth.RevokeAuthorizedClient(userID, req.ClientID)
}
