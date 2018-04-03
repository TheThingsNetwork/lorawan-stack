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
	"github.com/TheThingsNetwork/ttn/pkg/random"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/gobwas/glob"
	pbtypes "github.com/gogo/protobuf/types"
)

// isEmailAllowed checks whether an input email is allowed given the glob list
// of allowed emails in the settings.
//
// Note this method was not placed on ttnpb as part of the IdentityServerSettings
// type as it makes use of an external package.
func isEmailAllowed(email string, allowedEmails []string) bool {
	if len(allowedEmails) == 0 {
		return true
	}

	found := false
	for i := range allowedEmails {
		found = glob.MustCompile(strings.ToLower(allowedEmails[i])).Match(strings.ToLower(email))
		if found {
			break
		}
	}

	return found
}

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

		// If invitation-only mode is enabled check that an invitation token is provided.
		if settings.InvitationOnly && req.InvitationToken == "" {
			return ErrInvitationTokenMissing.New(nil)
		}

		// check for blacklisted ids
		if !settings.IsIDAllowed(req.User.UserID) {
			return ErrBlacklistedID.New(errors.Attributes{
				"id": req.User.UserID,
			})
		}

		password, err := auth.Hash(req.User.Password)
		if err != nil {
			return err
		}

		user := &ttnpb.User{
			UserIdentifiers: req.User.UserIdentifiers,
			Name:            req.User.Name,
			Email:           req.User.Email,
			Password:        string(password),
			State:           ttnpb.STATE_PENDING,
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
			if !isEmailAllowed(req.User.Email, settings.AllowedEmails) {
				return ErrEmailAddressNotAllowed.New(errors.Attributes{
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
			token := store.ValidationToken{
				ValidationToken: random.String(64),
				CreatedAt:       time.Now().UTC(),
				ExpiresIn:       int32(settings.ValidationTokenTTL.Seconds()),
			}

			err = tx.Users.SaveValidationToken(user.UserIdentifiers, token)
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
	err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_PROFILE_READ)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Users.GetByID(claimsFromContext(ctx).UserIdentifiers(), s.config.Specializers.User)
	if err != nil {
		return nil, err
	}

	user := found.GetUser()
	user.Password = ""

	return user, nil
}

// UpdateUser updates the account of the current user.
// If email address is updated it sends an email to validate it if and only if
// the `SkipValidation` setting is disabled.
func (s *userService) UpdateUser(ctx context.Context, req *ttnpb.UpdateUserRequest) (*pbtypes.Empty, error) {
	err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_PROFILE_WRITE)
	if err != nil {
		return nil, err
	}

	err = s.store.Transact(func(tx *store.Store) error {
		found, err := tx.Users.GetByID(claimsFromContext(ctx).UserIdentifiers(), s.config.Specializers.User)
		if err != nil {
			return err
		}
		user := found.GetUser()

		settings, err := tx.Settings.Get()
		if err != nil {
			return err
		}

		newEmail := false
		for _, path := range req.UpdateMask.Paths {
			switch {
			case ttnpb.FieldPathUserName.MatchString(path):
				user.Name = req.User.Name
			case ttnpb.FieldPathUserEmail.MatchString(path):
				if !isEmailAllowed(req.User.Email, settings.AllowedEmails) {
					return ErrEmailAddressNotAllowed.New(errors.Attributes{
						"allowed_emails": settings.AllowedEmails,
					})
				}

				newEmail = strings.ToLower(user.Email) != strings.ToLower(req.User.Email)
				if newEmail {
					if settings.SkipValidation {
						user.ValidatedAt = time.Now().UTC()
					} else {
						user.ValidatedAt = time.Time{}
					}
				}

				user.Email = req.User.Email
			default:
				return ttnpb.ErrInvalidPathUpdateMask.New(errors.Attributes{
					"path": path,
				})
			}
		}

		err = tx.Users.Update(user)
		if err != nil {
			return err
		}

		if !newEmail || (newEmail && settings.SkipValidation) {
			return nil
		}

		token := store.ValidationToken{
			ValidationToken: random.String(64),
			CreatedAt:       time.Now().UTC(),
			ExpiresIn:       int32(settings.ValidationTokenTTL.Seconds()),
		}

		err = tx.Users.SaveValidationToken(claimsFromContext(ctx).UserIdentifiers(), token)
		if err != nil {
			return err
		}

		return s.email.Send(user.Email, &templates.EmailValidation{
			OrganizationName: s.config.OrganizationName,
			PublicURL:        s.config.PublicURL,
			Token:            token.ValidationToken,
		})
	})

	return nil, err
}

// UpdateUserPassword updates the password of the current user.
func (s *userService) UpdateUserPassword(ctx context.Context, req *ttnpb.UpdateUserPasswordRequest) (*pbtypes.Empty, error) {
	err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_PROFILE_WRITE)
	if err != nil {
		return nil, err
	}

	err = s.store.Transact(func(tx *store.Store) error {
		found, err := tx.Users.GetByID(claimsFromContext(ctx).UserIdentifiers(), s.config.Specializers.User)
		if err != nil {
			return err
		}
		user := found.GetUser()

		matches, err := auth.Password(user.Password).Validate(req.Old)
		if err != nil {
			return err
		}

		if !matches {
			return ErrInvalidPassword.New(nil)
		}

		hashed, err := auth.Hash(req.New)
		if err != nil {
			return err
		}

		user.Password = string(hashed)

		return tx.Users.Update(user)
	})

	return nil, err
}

// DeleteUser deletes the account of the current user.
func (s *userService) DeleteUser(ctx context.Context, _ *pbtypes.Empty) (*pbtypes.Empty, error) {
	err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_DELETE)
	if err != nil {
		return nil, err
	}

	err = s.store.Transact(func(tx *store.Store) error {
		apps, err := tx.Applications.ListByOrganizationOrUser(organizationOrUserIDsUserIDs(claimsFromContext(ctx).UserIdentifiers()), s.config.Specializers.Application)
		if err != nil {
			return err
		}

		gtws, err := tx.Gateways.ListByOrganizationOrUser(organizationOrUserIDsUserIDs(claimsFromContext(ctx).UserIdentifiers()), s.config.Specializers.Gateway)
		if err != nil {
			return err
		}

		err = tx.Users.Delete(claimsFromContext(ctx).UserIdentifiers())
		if err != nil {
			return err
		}

		for _, app := range apps {
			appID := app.GetApplication().ApplicationIdentifiers

			c, err := tx.Applications.ListCollaborators(appID, ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS)
			if err != nil {
				return err
			}

			if len(c) == 0 {
				return errors.Errorf("Failed to delete user `%s`: cannot leave application `%s` without at least one collaborator with `RIGHT_APPLICATION_SETTINGS_COLLABORATORS` right", claimsFromContext(ctx).UserIdentifiers(), appID)
			}
		}

		for _, gtw := range gtws {
			gtwID := gtw.GetGateway().GatewayIdentifiers

			c, err := tx.Gateways.ListCollaborators(gtwID, ttnpb.RIGHT_GATEWAY_SETTINGS_COLLABORATORS)
			if err != nil {
				return err
			}

			if len(c) == 0 {
				return errors.Errorf("Failed to delete user `%s`: cannot leave gateway `%s` without at least one collaborator with `RIGHT_GATEWAY_SETTINGS_COLLABORATORS` right", claimsFromContext(ctx).UserIdentifiers(), gtwID)
			}
		}

		return nil
	})

	return nil, err
}

// GenerateUserAPIKey generates an user API key and returns it.
func (s *userService) GenerateUserAPIKey(ctx context.Context, req *ttnpb.GenerateUserAPIKeyRequest) (*ttnpb.APIKey, error) {
	err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_KEYS)
	if err != nil {
		return nil, err
	}

	k, err := auth.GenerateUserAPIKey(s.config.Hostname)
	if err != nil {
		return nil, err
	}

	key := ttnpb.APIKey{
		Key:    k,
		Name:   req.Name,
		Rights: req.Rights,
	}

	err = s.store.Users.SaveAPIKey(claimsFromContext(ctx).UserIdentifiers(), key)
	if err != nil {
		return nil, err
	}

	return &key, nil
}

// ListUserAPIKeys returns all the API keys from the current user.
func (s *userService) ListUserAPIKeys(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.ListUserAPIKeysResponse, error) {
	err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_KEYS)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Users.ListAPIKeys(claimsFromContext(ctx).UserIdentifiers())
	if err != nil {
		return nil, err
	}

	keys := make([]*ttnpb.APIKey, 0, len(found))
	for i := range found {
		keys = append(keys, &found[i])
	}

	return &ttnpb.ListUserAPIKeysResponse{
		APIKeys: keys,
	}, nil
}

// UpdateUserAPIKey updates an API key from the current user.
func (s *userService) UpdateUserAPIKey(ctx context.Context, req *ttnpb.UpdateUserAPIKeyRequest) (*pbtypes.Empty, error) {
	err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_PROFILE_WRITE)
	if err != nil {
		return nil, err
	}

	return nil, s.store.Users.UpdateAPIKeyRights(claimsFromContext(ctx).UserIdentifiers(), req.Name, req.Rights)
}

// RemoveUserAPIKey removes an API key from the current user.
func (s *userService) RemoveUserAPIKey(ctx context.Context, req *ttnpb.RemoveUserAPIKeyRequest) (*pbtypes.Empty, error) {
	err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_KEYS)
	if err != nil {
		return nil, err
	}

	return nil, s.store.Users.DeleteAPIKey(claimsFromContext(ctx).UserIdentifiers(), req.Name)
}

// ValidateUserEmail validates the user's email with the token sent to the email.
func (s *userService) ValidateUserEmail(ctx context.Context, req *ttnpb.ValidateUserEmailRequest) (*pbtypes.Empty, error) {
	err := s.store.Transact(func(tx *store.Store) error {
		userID, token, err := tx.Users.GetValidationToken(req.Token)
		if err != nil {
			return err
		}

		if token.IsExpired() {
			return ErrValidationTokenExpired.New(nil)
		}

		user, err := tx.Users.GetByID(userID, s.config.Specializers.User)
		if err != nil {
			return err
		}

		user.GetUser().ValidatedAt = time.Now().UTC()

		err = tx.Users.Update(user)
		if err != nil {
			return err
		}

		return tx.Users.DeleteValidationToken(req.Token)
	})

	return nil, err
}

// RequestUserEmailValidation requests a new validation email if the user's email
// isn't validated yet.
func (s *userService) RequestUserEmailValidation(ctx context.Context, _ *pbtypes.Empty) (*pbtypes.Empty, error) {
	err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_PROFILE_WRITE)
	if err != nil {
		return nil, err
	}

	err = s.store.Transact(func(tx *store.Store) error {
		found, err := tx.Users.GetByID(claimsFromContext(ctx).UserIdentifiers(), s.config.Specializers.User)
		if err != nil {
			return err
		}
		user := found.GetUser()

		if !user.ValidatedAt.IsZero() {
			return ErrEmailAlreadyValidated.New(nil)
		}

		settings, err := tx.Settings.Get()
		if err != nil {
			return err
		}

		token := store.ValidationToken{
			ValidationToken: random.String(64),
			CreatedAt:       time.Now().UTC(),
			ExpiresIn:       int32(settings.ValidationTokenTTL.Seconds()),
		}

		err = tx.Users.SaveValidationToken(claimsFromContext(ctx).UserIdentifiers(), token)
		if err != nil {
			return err
		}

		return s.email.Send(user.Email, &templates.EmailValidation{
			OrganizationName: s.config.OrganizationName,
			PublicURL:        s.config.PublicURL,
			Token:            token.ValidationToken,
		})
	})

	return nil, err
}

// ListAuthorizedClients returns all the authorized third-party clients that
// the current user has.
func (s *userService) ListAuthorizedClients(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.ListAuthorizedClientsResponse, error) {
	err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_AUTHORIZEDCLIENTS)
	if err != nil {
		return nil, err
	}

	found, err := s.store.OAuth.ListAuthorizedClients(claimsFromContext(ctx).UserIdentifiers(), s.config.Specializers.Client)
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
func (s *userService) RevokeAuthorizedClient(ctx context.Context, req *ttnpb.ClientIdentifiers) (*pbtypes.Empty, error) {
	err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_AUTHORIZEDCLIENTS)
	if err != nil {
		return nil, err
	}

	return nil, s.store.OAuth.RevokeAuthorizedClient(claimsFromContext(ctx).UserIdentifiers(), *req)
}
