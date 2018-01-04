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
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/util"
	"github.com/TheThingsNetwork/ttn/pkg/random"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	ttntypes "github.com/TheThingsNetwork/ttn/pkg/types"
	pbtypes "github.com/gogo/protobuf/types"
)

// make sure IdentityServer implements ttnpb.IsUserServer.
var _ ttnpb.IsUserServer = new(IdentityServer)

// CreateUser creates an user in the network.
func (is *IdentityServer) CreateUser(ctx context.Context, req *ttnpb.CreateUserRequest) (*pbtypes.Empty, error) {
	settings, err := is.store.Settings.Get()
	if err != nil {
		return nil, err
	}

	// check for blacklisted ids
	if !util.IsIDAllowed(req.User.UserID, settings.BlacklistedIDs) {
		return nil, ErrBlacklistedID.New(errors.Attributes{
			"id": req.User.UserID,
		})
	}

	// check for allowed emails
	if !util.IsEmailAllowed(req.User.Email, settings.AllowedEmails) {
		return nil, ErrNotAllowedEmail.New(errors.Attributes{
			"email": req.User.Email,
		})
	}

	password, err := ttntypes.Hash(req.User.Password)
	if err != nil {
		return nil, err
	}

	user := &ttnpb.User{
		UserIdentifier: req.User.UserIdentifier,
		Name:           req.User.Name,
		Email:          req.User.Email,
		Password:       string(password),
		State:          ttnpb.STATE_PENDING,
	}

	if !settings.AdminApproval {
		user.State = ttnpb.STATE_APPROVED
	}

	if settings.SkipValidation {
		user.ValidatedAt = time.Now()

		return nil, is.store.Users.Create(user)
	}

	err = is.store.Transact(func(s *store.Store) error {
		err := s.Users.Create(user)
		if err != nil {
			return err
		}

		token := &types.ValidationToken{
			ValidationToken: random.String(64),
			CreatedAt:       time.Now(),
			ExpiresIn:       int32(settings.ValidationTokenTTL.Seconds()),
		}

		err = s.Users.SaveValidationToken(req.User.UserID, token)
		if err != nil {
			return err
		}

		return is.email.Send(
			user.Email,
			templates.EmailValidation(),
			map[string]interface{}{
				templates.EmailValidationHostname: "",
				templates.EmailValidationToken:    token.ValidationToken,
			})
	})

	return nil, err
}

// GetUser returns the account of the current user.
func (is *IdentityServer) GetUser(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.User, error) {
	userID, err := is.userCheck(ctx, ttnpb.RIGHT_USER_PROFILE_READ)
	if err != nil {
		return nil, err
	}

	found, err := is.store.Users.GetByID(userID, is.factories.user)
	if err != nil {
		return nil, err
	}

	found.GetUser().Password = ""

	return found.GetUser(), nil
}

// UpdateUser updates the account of the current user.
// If the email is modified a validation email will be sent.
func (is *IdentityServer) UpdateUser(ctx context.Context, req *ttnpb.UpdateUserRequest) (*pbtypes.Empty, error) {
	userID, err := is.userCheck(ctx, ttnpb.RIGHT_USER_PROFILE_WRITE)
	if err != nil {
		return nil, err
	}

	found, err := is.store.Users.GetByID(userID, is.factories.user)
	if err != nil {
		return nil, err
	}

	settings, err := is.store.Settings.Get()
	if err != nil {
		return nil, err
	}

	newEmail := false
	for _, path := range req.UpdateMask.Paths {
		switch true {
		case ttnpb.FieldPathUserName.MatchString(path):
			found.GetUser().Name = req.User.Name
		case ttnpb.FieldPathUserEmail.MatchString(path):
			if strings.ToLower(req.User.Email) != strings.ToLower(found.GetUser().Email) {
				newEmail = true
				found.GetUser().ValidatedAt = time.Time{}
			}

			if !util.IsEmailAllowed(req.User.Email, settings.AllowedEmails) {
				return nil, ErrNotAllowedEmail.New(errors.Attributes{
					"email": req.User.Email,
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
		return nil, is.store.Users.Update(found)
	}

	err = is.store.Transact(func(s *store.Store) error {
		err := is.store.Users.Update(found)
		if err != nil {
			return err
		}

		token := &types.ValidationToken{
			ValidationToken: random.String(64),
			CreatedAt:       time.Now(),
			ExpiresIn:       int32(settings.ValidationTokenTTL.Seconds()),
		}

		err = is.store.Users.SaveValidationToken(userID, token)
		if err != nil {
			return err
		}

		return is.email.Send(
			found.GetUser().Email,
			templates.EmailValidation(),
			map[string]interface{}{
				templates.EmailValidationHostname: "",
				templates.EmailValidationToken:    token.ValidationToken,
			})
	})

	return nil, err
}

// UpdateUserPassword updates the password of the current user.
func (is *IdentityServer) UpdateUserPassword(ctx context.Context, req *ttnpb.UpdateUserPasswordRequest) (*pbtypes.Empty, error) {
	userID, err := is.userCheck(ctx, ttnpb.RIGHT_USER_PROFILE_WRITE)
	if err != nil {
		return nil, err
	}

	found, err := is.store.Users.GetByID(userID, is.factories.user)
	if err != nil {
		return nil, err
	}

	matches, err := ttntypes.Password(found.GetUser().Password).Validate(req.Old)
	if err != nil {
		return nil, err
	}

	if !matches {
		return nil, ErrPasswordsDoNotMatch.New(nil)
	}

	hashed, err := ttntypes.Hash(req.New)
	if err != nil {
		return nil, err
	}

	found.GetUser().Password = string(hashed)

	return nil, is.store.Users.Update(found)
}

// DeleteUser deletes the account of the current user.
func (is *IdentityServer) DeleteUser(ctx context.Context, _ *pbtypes.Empty) (*pbtypes.Empty, error) {
	userID, err := is.userCheck(ctx, ttnpb.RIGHT_USER_DELETE)
	if err != nil {
		return nil, err
	}

	err = is.store.Transact(func(s *store.Store) error {
		err := s.Users.Delete(userID)
		if err != nil {
			return err
		}

		apps, err := s.Applications.ListByUser(userID, is.factories.application)
		if err != nil {
			return err
		}

		for _, app := range apps {
			appID := app.GetApplication().ApplicationID

			c, err := s.Applications.ListCollaborators(appID, ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS)
			if err != nil {
				return err
			}

			if len(c) == 0 {
				return errors.Errorf("Failed to delete user `%s`: cannot leave application `%s` without at least one collaborator with `application:settings:collaborators` right", userID, appID)
			}
		}

		gtws, err := s.Gateways.ListByUser(userID, is.factories.gateway)
		if err != nil {
			return err
		}

		for _, gtw := range gtws {
			gtwID := gtw.GetGateway().GatewayID

			c, err := s.Gateways.ListCollaborators(gtwID, ttnpb.RIGHT_GATEWAY_SETTINGS_COLLABORATORS)
			if err != nil {
				return err
			}

			if len(c) == 0 {
				return errors.Errorf("Failed to delete user `%s`: cannot leave gateway `%s` without at least one collaborator with `gateway:settings:collaborators` right", userID, gtwID)
			}
		}

		return nil
	})

	return nil, err
}

// GenerateUserAPIKey generates an user API key and returns it.
func (is *IdentityServer) GenerateUserAPIKey(ctx context.Context, req *ttnpb.GenerateUserAPIKeyRequest) (*ttnpb.APIKey, error) {
	userID, err := is.userCheck(ctx, ttnpb.RIGHT_USER_KEYS)
	if err != nil {
		return nil, err
	}

	//  TODO(gomezjdaniel): use the tenantID from the request metadata to generate
	//  to generate the user API key.
	k, err := auth.GenerateUserAPIKey("")
	if err != nil {
		return nil, err
	}

	key := &ttnpb.APIKey{
		Key:    k,
		Name:   req.Name,
		Rights: req.Rights,
	}

	err = is.store.Users.SaveAPIKey(userID, key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// ListUserAPIKeys returns all the API keys from the current user.
func (is *IdentityServer) ListUserAPIKeys(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.ListUserAPIKeysResponse, error) {
	userID, err := is.userCheck(ctx, ttnpb.RIGHT_USER_KEYS)
	if err != nil {
		return nil, err
	}

	found, err := is.store.Users.ListAPIKeys(userID)
	if err != nil {
		return nil, err
	}

	return &ttnpb.ListUserAPIKeysResponse{
		APIKeys: found,
	}, nil
}

// UpdateUserAPIKey updates an API key from the current user.
func (is *IdentityServer) UpdateUserAPIKey(ctx context.Context, req *ttnpb.UpdateUserAPIKeyRequest) (*pbtypes.Empty, error) {
	userID, err := is.userCheck(ctx, ttnpb.RIGHT_USER_PROFILE_WRITE)
	if err != nil {
		return nil, err
	}

	return nil, is.store.Users.UpdateAPIKeyRights(userID, req.Name, req.Rights)
}

// RemoveUserAPIKey removes an API key from the current user.
func (is *IdentityServer) RemoveUserAPIKey(ctx context.Context, req *ttnpb.RemoveUserAPIKeyRequest) (*pbtypes.Empty, error) {
	userID, err := is.userCheck(ctx, ttnpb.RIGHT_USER_KEYS)
	if err != nil {
		return nil, err
	}

	return nil, is.store.Users.DeleteAPIKey(userID, req.Name)
}

// ValidateUserEmail validates the user's email with the token sent to the email.
func (is *IdentityServer) ValidateUserEmail(ctx context.Context, req *ttnpb.ValidateUserEmailRequest) (*pbtypes.Empty, error) {
	err := is.store.Transact(func(store *store.Store) error {
		userID, token, err := store.Users.GetValidationToken(req.Token)
		if err != nil {
			return err
		}

		if token.IsExpired() {
			return errors.New("token expired")
		}

		user, err := store.Users.GetByID(userID, is.factories.user)
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

// RequestUserEmailValidation requests a new validation email if the user's emai
// isn't validated yet.
func (is *IdentityServer) RequestUserEmailValidation(ctx context.Context, _ *pbtypes.Empty) (*pbtypes.Empty, error) {
	userID, err := is.userCheck(ctx, ttnpb.RIGHT_USER_PROFILE_WRITE)
	if err != nil {
		return nil, err
	}

	user, err := is.store.Users.GetByID(userID, is.factories.user)
	if err != nil {
		return nil, err
	}

	if !user.GetUser().ValidatedAt.IsZero() {
		return nil, errors.New("email already validated")
	}

	settings, err := is.store.Settings.Get()
	if err != nil {
		return nil, err
	}

	token := &types.ValidationToken{
		ValidationToken: random.String(64),
		CreatedAt:       time.Now(),
		ExpiresIn:       int32(settings.ValidationTokenTTL.Seconds()),
	}

	err = is.store.Users.SaveValidationToken(userID, token)
	if err != nil {
		return nil, err
	}

	return nil, is.email.Send(
		user.GetUser().Email,
		templates.EmailValidation(),
		map[string]interface{}{
			templates.EmailValidationHostname: "",
			templates.EmailValidationToken:    token.ValidationToken,
		},
	)
}

// ListAuthorizedClients returns all the authorized third-party clients that
// the current user has.
func (is *IdentityServer) ListAuthorizedClients(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.ListAuthorizedClientsResponse, error) {
	userID, err := is.userCheck(ctx, ttnpb.RIGHT_USER_AUTHORIZEDCLIENTS)
	if err != nil {
		return nil, err
	}

	found, err := is.store.OAuth.ListAuthorizedClients(userID, is.factories.client)
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
func (is *IdentityServer) RevokeAuthorizedClient(ctx context.Context, req *ttnpb.ClientIdentifier) (*pbtypes.Empty, error) {
	userID, err := is.userCheck(ctx, ttnpb.RIGHT_USER_AUTHORIZEDCLIENTS)
	if err != nil {
		return nil, err
	}

	return nil, is.store.OAuth.RevokeAuthorizedClient(userID, req.ClientID)
}
