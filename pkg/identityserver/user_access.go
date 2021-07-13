// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/email"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/emails"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	evtCreateUserAPIKey = events.Define(
		"user.api-key.create", "create user API key",
		events.WithVisibility(ttnpb.RIGHT_USER_SETTINGS_API_KEYS),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUpdateUserAPIKey = events.Define(
		"user.api-key.update", "update user API key",
		events.WithVisibility(ttnpb.RIGHT_USER_SETTINGS_API_KEYS),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteUserAPIKey = events.Define(
		"user.api-key.delete", "delete user API key",
		events.WithVisibility(ttnpb.RIGHT_USER_SETTINGS_API_KEYS),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
)

func (is *IdentityServer) listUserRights(ctx context.Context, ids *ttnpb.UserIdentifiers) (*ttnpb.Rights, error) {
	usrRights, err := rights.ListUser(ctx, *ids)
	if err != nil {
		return nil, err
	}
	return usrRights.Intersect(ttnpb.AllEntityRights.Union(ttnpb.AllOrganizationRights, ttnpb.AllUserRights)), nil
}

func (is *IdentityServer) createUserAPIKey(ctx context.Context, req *ttnpb.CreateUserAPIKeyRequest) (key *ttnpb.APIKey, err error) {
	// Require that caller has rights to manage API keys.
	if err = rights.RequireUser(ctx, req.UserIdentifiers, ttnpb.RIGHT_USER_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}
	// Require that caller has at least the rights of the API key.
	if err = rights.RequireUser(ctx, req.UserIdentifiers, req.Rights...); err != nil {
		return nil, err
	}
	key, token, err := GenerateAPIKey(ctx, req.Name, req.ExpiresAt, req.Rights...)
	if err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		key, err = store.GetAPIKeyStore(db).CreateAPIKey(ctx, req.UserIdentifiers.GetEntityIdentifiers(), key)
		return err
	})
	if err != nil {
		return nil, err
	}
	key.Key = token
	events.Publish(evtCreateUserAPIKey.NewWithIdentifiersAndData(ctx, &req.UserIdentifiers, nil))
	err = is.SendUserEmail(ctx, &req.UserIdentifiers, func(data emails.Data) email.MessageData {
		data.SetEntity(req)
		return &emails.APIKeyCreated{Data: data, Key: key, Rights: key.Rights}
	})
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("Could not send API key created notification email")
	}
	return key, nil
}

func (is *IdentityServer) listUserAPIKeys(ctx context.Context, req *ttnpb.ListUserAPIKeysRequest) (keys *ttnpb.APIKeys, err error) {
	if err = rights.RequireUser(ctx, req.UserIdentifiers, ttnpb.RIGHT_USER_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	keys = &ttnpb.APIKeys{}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		keys.APIKeys, err = store.GetAPIKeyStore(db).FindAPIKeys(ctx, req.UserIdentifiers.GetEntityIdentifiers())
		return err
	})
	if err != nil {
		return nil, err
	}
	for _, key := range keys.APIKeys {
		key.Key = ""
	}
	return keys, nil
}

func (is *IdentityServer) getUserAPIKey(ctx context.Context, req *ttnpb.GetUserAPIKeyRequest) (key *ttnpb.APIKey, err error) {
	if err = rights.RequireUser(ctx, req.UserIdentifiers, ttnpb.RIGHT_USER_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		_, key, err = store.GetAPIKeyStore(db).GetAPIKey(ctx, req.KeyId)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	key.Key = ""
	return key, nil
}

func (is *IdentityServer) updateUserAPIKey(ctx context.Context, req *ttnpb.UpdateUserAPIKeyRequest) (key *ttnpb.APIKey, err error) {
	// Require that caller has rights to manage API keys.
	if err = rights.RequireUser(ctx, req.UserIdentifiers, ttnpb.RIGHT_USER_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		if len(req.APIKey.Rights) > 0 {
			_, key, err := store.GetAPIKeyStore(db).GetAPIKey(ctx, req.APIKey.ID)
			if err != nil {
				return err
			}

			newRights := ttnpb.RightsFrom(req.APIKey.Rights...)
			existingRights := ttnpb.RightsFrom(key.Rights...)

			// Require the caller to have all added rights.
			if err := rights.RequireUser(ctx, req.UserIdentifiers, newRights.Sub(existingRights).GetRights()...); err != nil {
				return err
			}
			// Require the caller to have all removed rights.
			if err := rights.RequireUser(ctx, req.UserIdentifiers, existingRights.Sub(newRights).GetRights()...); err != nil {
				return err
			}
		}

		key, err = store.GetAPIKeyStore(db).UpdateAPIKey(ctx, req.UserIdentifiers.GetEntityIdentifiers(), &req.APIKey, req.FieldMask)
		return err
	})
	if err != nil {
		return nil, err
	}
	if key == nil { // API key was deleted.
		events.Publish(evtDeleteUserAPIKey.NewWithIdentifiersAndData(ctx, &req.UserIdentifiers, nil))
		return &ttnpb.APIKey{}, nil
	}
	key.Key = ""
	events.Publish(evtUpdateUserAPIKey.NewWithIdentifiersAndData(ctx, &req.UserIdentifiers, nil))
	err = is.SendUserEmail(ctx, &req.UserIdentifiers, func(data emails.Data) email.MessageData {
		data.SetEntity(req)
		return &emails.APIKeyChanged{Data: data, Key: key, Rights: key.Rights}
	})
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("Could not send API key update notification email")
	}

	return key, nil
}

const maxActiveLoginTokens = 5

var (
	errLoginTokensDisabled   = errors.DefineFailedPrecondition("login_tokens_disabled", "login tokens are disabled")
	errLoginTokensStillValid = errors.DefineAlreadyExists("login_tokens_still_valid", "previously created login token still valid")
)

func (is *IdentityServer) createLoginToken(ctx context.Context, req *ttnpb.CreateLoginTokenRequest) (*ttnpb.CreateLoginTokenResponse, error) {
	loginTokenConfig := is.configFromContext(ctx).LoginTokens
	if !loginTokenConfig.Enabled {
		return nil, errLoginTokensDisabled.New()
	}

	var canCreateMoreTokens bool
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		activeTokens, err := store.GetLoginTokenStore(db).FindActiveLoginTokens(ctx, &req.UserIdentifiers)
		if err != nil {
			return err
		}
		canCreateMoreTokens = len(activeTokens) < maxActiveLoginTokens
		return nil
	})
	if err != nil {
		return nil, err
	}
	if !canCreateMoreTokens {
		return nil, errLoginTokensStillValid.New()
	}

	var canSkipEmail, canReturnToken bool
	if is.IsAdmin(ctx) {
		canSkipEmail = true // Admin callers can skip sending emails.
		err := is.withDatabase(ctx, func(db *gorm.DB) error {
			usr, err := store.GetUserStore(db).GetUser(ctx, &req.UserIdentifiers, &pbtypes.FieldMask{Paths: []string{"admin"}})
			if !usr.Admin {
				canReturnToken = true // Admin callers can get login tokens for non-admin users.
			}
			return err
		})
		if err != nil {
			return nil, err
		}
	}

	token, err := auth.GenerateKey(ctx)
	if err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		_, err := store.GetLoginTokenStore(db).CreateLoginToken(ctx, &ttnpb.LoginToken{
			UserIdentifiers: req.UserIdentifiers,
			ExpiresAt:       time.Now().Add(loginTokenConfig.TokenTTL),
			Token:           token,
		})
		return err
	})

	if !(canSkipEmail && req.SkipEmail) {
		err = is.SendUserEmail(ctx, &req.UserIdentifiers, func(data emails.Data) email.MessageData {
			return &emails.LoginToken{Data: data, LoginToken: token, TTL: loginTokenConfig.TokenTTL}
		})
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Could not send API key created notification email")
		}
	}
	if !canReturnToken {
		token = ""
	}
	return &ttnpb.CreateLoginTokenResponse{Token: token}, nil
}

type userAccess struct {
	*IdentityServer
}

func (ua *userAccess) ListRights(ctx context.Context, req *ttnpb.UserIdentifiers) (*ttnpb.Rights, error) {
	return ua.listUserRights(ctx, req)
}

func (ua *userAccess) GetAPIKey(ctx context.Context, req *ttnpb.GetUserAPIKeyRequest) (*ttnpb.APIKey, error) {
	return ua.getUserAPIKey(ctx, req)
}

func (ua *userAccess) CreateAPIKey(ctx context.Context, req *ttnpb.CreateUserAPIKeyRequest) (*ttnpb.APIKey, error) {
	return ua.createUserAPIKey(ctx, req)
}

func (ua *userAccess) ListAPIKeys(ctx context.Context, req *ttnpb.ListUserAPIKeysRequest) (*ttnpb.APIKeys, error) {
	return ua.listUserAPIKeys(ctx, req)
}

func (ua *userAccess) UpdateAPIKey(ctx context.Context, req *ttnpb.UpdateUserAPIKeyRequest) (*ttnpb.APIKey, error) {
	return ua.updateUserAPIKey(ctx, req)
}

func (ua *userAccess) CreateLoginToken(ctx context.Context, req *ttnpb.CreateLoginTokenRequest) (*ttnpb.CreateLoginTokenResponse, error) {
	return ua.createLoginToken(ctx, req)
}
