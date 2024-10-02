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

	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/email"
	"go.thethings.network/lorawan-stack/v3/pkg/email/templates"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	evtCreateUserAPIKey = events.Define(
		"user.api-key.create", "create user API key",
		events.WithVisibility(ttnpb.Right_RIGHT_USER_SETTINGS_API_KEYS),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUpdateUserAPIKey = events.Define(
		"user.api-key.update", "update user API key",
		events.WithVisibility(ttnpb.Right_RIGHT_USER_SETTINGS_API_KEYS),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteUserAPIKey = events.Define(
		"user.api-key.delete", "delete user API key",
		events.WithVisibility(ttnpb.Right_RIGHT_USER_SETTINGS_API_KEYS),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
)

func (*IdentityServer) listUserRights(ctx context.Context, ids *ttnpb.UserIdentifiers) (*ttnpb.Rights, error) {
	usrRights, err := rights.ListUser(ctx, ids)
	if err != nil {
		return nil, err
	}
	return usrRights.Intersect(ttnpb.AllEntityRights.Union(ttnpb.AllOrganizationRights, ttnpb.AllUserRights)), nil
}

func (is *IdentityServer) createUserAPIKey(
	ctx context.Context, req *ttnpb.CreateUserAPIKeyRequest,
) (key *ttnpb.APIKey, err error) {
	// Require that caller has rights to manage API keys.
	if err = rights.RequireUser(ctx, req.GetUserIds(), ttnpb.Right_RIGHT_USER_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}
	// Require that caller has at least the rights of the API key.
	if err = rights.RequireUser(ctx, req.GetUserIds(), req.Rights...); err != nil {
		return nil, err
	}
	key, token, err := GenerateAPIKey(ctx, req.Name, ttnpb.StdTime(req.ExpiresAt), req.Rights...)
	if err != nil {
		return nil, err
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		key, err = st.CreateAPIKey(ctx, req.GetUserIds().GetEntityIdentifiers(), key)
		return err
	})
	if err != nil {
		return nil, err
	}
	key.Key = ""

	events.Publish(evtCreateUserAPIKey.NewWithIdentifiersAndData(ctx, req.GetUserIds(), key))
	go is.notifyInternal(ctx, &ttnpb.CreateNotificationRequest{
		EntityIds:        req.GetUserIds().GetEntityIdentifiers(),
		NotificationType: ttnpb.NotificationType_API_KEY_CREATED,
		Data:             ttnpb.MustMarshalAny(key),
		Receivers: []ttnpb.NotificationReceiver{
			ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_ADMINISTRATIVE_CONTACT,
		},
	})

	key.Key = token
	return key, nil
}

func (is *IdentityServer) listUserAPIKeys(
	ctx context.Context, req *ttnpb.ListUserAPIKeysRequest,
) (keys *ttnpb.APIKeys, err error) {
	if err = rights.RequireUser(ctx, req.GetUserIds(), ttnpb.Right_RIGHT_USER_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}
	ctx = store.WithOrder(ctx, req.Order)
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	keys = &ttnpb.APIKeys{}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		keys.ApiKeys, err = st.FindAPIKeys(ctx, req.GetUserIds().GetEntityIdentifiers())
		return err
	})
	if err != nil {
		return nil, err
	}
	for _, key := range keys.ApiKeys {
		key.Key = ""
	}
	return keys, nil
}

func (is *IdentityServer) getUserAPIKey(
	ctx context.Context, req *ttnpb.GetUserAPIKeyRequest,
) (key *ttnpb.APIKey, err error) {
	if err = rights.RequireUser(ctx, req.GetUserIds(), ttnpb.Right_RIGHT_USER_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}

	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		key, err = st.GetAPIKey(ctx, req.GetUserIds().GetEntityIdentifiers(), req.KeyId)
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

func (is *IdentityServer) updateUserAPIKey(
	ctx context.Context, req *ttnpb.UpdateUserAPIKeyRequest,
) (key *ttnpb.APIKey, err error) {
	// Require that caller has rights to manage API keys.
	if err = rights.RequireUser(ctx, req.GetUserIds(), ttnpb.Right_RIGHT_USER_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}

	// Backwards compatibility for older clients.
	if len(req.FieldMask.GetPaths()) == 0 {
		req.FieldMask = ttnpb.FieldMask("rights", "name")
	}

	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		if len(req.ApiKey.Rights) > 0 {
			key, err := st.GetAPIKey(ctx, req.GetUserIds().GetEntityIdentifiers(), req.ApiKey.Id)
			if err != nil {
				return err
			}

			newRights := ttnpb.RightsFrom(req.ApiKey.Rights...)
			existingRights := ttnpb.RightsFrom(key.Rights...)

			// Require the caller to have all added rights.
			err = rights.RequireUser(ctx, req.GetUserIds(), newRights.Sub(existingRights).GetRights()...)
			if err != nil {
				return err
			}
			// Require the caller to have all removed rights.
			err = rights.RequireUser(ctx, req.GetUserIds(), existingRights.Sub(newRights).GetRights()...)
			if err != nil {
				return err
			}
		}

		if len(req.ApiKey.Rights) == 0 && ttnpb.HasAnyField(req.GetFieldMask().GetPaths(), "rights") {
			// TODO: Remove delete capability (https://github.com/TheThingsNetwork/lorawan-stack/issues/6488).
			return st.DeleteAPIKey(ctx, req.GetUserIds().GetEntityIdentifiers(), req.ApiKey)
		}

		key, err = st.UpdateAPIKey(ctx, req.UserIds.GetEntityIdentifiers(), req.ApiKey, req.FieldMask.GetPaths())
		return err
	})
	if err != nil {
		return nil, err
	}
	if key == nil { // API key was deleted.
		events.Publish(evtDeleteUserAPIKey.NewWithIdentifiersAndData(ctx, req.GetUserIds(), nil))
		return &ttnpb.APIKey{}, nil
	}
	key.Key = ""

	events.Publish(evtUpdateUserAPIKey.NewWithIdentifiersAndData(ctx, req.GetUserIds(), key))
	go is.notifyInternal(ctx, &ttnpb.CreateNotificationRequest{
		EntityIds:        req.GetUserIds().GetEntityIdentifiers(),
		NotificationType: ttnpb.NotificationType_API_KEY_CHANGED,
		Data:             ttnpb.MustMarshalAny(key),
		Receivers: []ttnpb.NotificationReceiver{
			ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_ADMINISTRATIVE_CONTACT,
		},
	})

	return key, nil
}

func (is *IdentityServer) deleteUserAPIKey(
	ctx context.Context, req *ttnpb.DeleteUserAPIKeyRequest,
) (*emptypb.Empty, error) {
	// Require that caller has rights to manage API keys.
	err := rights.RequireUser(ctx, req.GetUserIds(), ttnpb.Right_RIGHT_USER_SETTINGS_API_KEYS)
	if err != nil {
		return ttnpb.Empty, err
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		return st.DeleteAPIKey(ctx, req.GetUserIds().GetEntityIdentifiers(), &ttnpb.APIKey{Id: req.KeyId})
	})
	if err != nil {
		return ttnpb.Empty, err
	}
	events.Publish(evtDeleteUserAPIKey.New(ctx, events.WithIdentifiers(req.GetUserIds())))
	return ttnpb.Empty, nil
}

const maxActiveLoginTokens = 5

var (
	errLoginTokensDisabled   = errors.DefineFailedPrecondition("login_tokens_disabled", "login tokens are disabled")
	errLoginTokensStillValid = errors.DefineAlreadyExists(
		"login_tokens_still_valid", "previously created login token still valid",
	)
)

func (is *IdentityServer) createLoginToken(
	ctx context.Context, req *ttnpb.CreateLoginTokenRequest,
) (*ttnpb.CreateLoginTokenResponse, error) {
	loginTokenConfig := is.configFromContext(ctx).LoginTokens
	if !loginTokenConfig.Enabled {
		return nil, errLoginTokensDisabled.New()
	}

	var canCreateMoreTokens bool
	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		activeTokens, err := st.FindActiveLoginTokens(ctx, req.GetUserIds())
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
		err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
			usr, err := st.GetUser(ctx, req.GetUserIds(), []string{"admin"})
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
	expiresAt := time.Now().Add(loginTokenConfig.TokenTTL)
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		_, err := st.CreateLoginToken(ctx, &ttnpb.LoginToken{
			UserIds:   req.GetUserIds(),
			ExpiresAt: timestamppb.New(expiresAt),
			Token:     token,
		})
		return err
	})
	if err != nil {
		return nil, err
	}

	if !canSkipEmail || !req.SkipEmail {
		go is.SendTemplateEmailToUserIDs(
			is.FromRequestContext(ctx),
			ttnpb.NotificationType_LOGIN_TOKEN,
			func(ctx context.Context, data email.TemplateData) (email.TemplateData, error) {
				return &templates.LoginTokenData{
					TemplateData: data,
					LoginToken:   token,
					TTL:          loginTokenConfig.TokenTTL,
				}, nil
			}, req.GetUserIds())
	}
	if !canReturnToken {
		token = ""
	}
	return &ttnpb.CreateLoginTokenResponse{Token: token}, nil
}

type userAccess struct {
	ttnpb.UnimplementedUserAccessServer

	*IdentityServer
}

func (ua *userAccess) ListRights(
	ctx context.Context, req *ttnpb.UserIdentifiers,
) (*ttnpb.Rights, error) {
	return ua.listUserRights(ctx, req)
}

func (ua *userAccess) GetAPIKey(
	ctx context.Context, req *ttnpb.GetUserAPIKeyRequest,
) (*ttnpb.APIKey, error) {
	return ua.getUserAPIKey(ctx, req)
}

func (ua *userAccess) CreateAPIKey(
	ctx context.Context, req *ttnpb.CreateUserAPIKeyRequest,
) (*ttnpb.APIKey, error) {
	return ua.createUserAPIKey(ctx, req)
}

func (ua *userAccess) ListAPIKeys(
	ctx context.Context, req *ttnpb.ListUserAPIKeysRequest,
) (*ttnpb.APIKeys, error) {
	return ua.listUserAPIKeys(ctx, req)
}

func (ua *userAccess) UpdateAPIKey(
	ctx context.Context, req *ttnpb.UpdateUserAPIKeyRequest,
) (*ttnpb.APIKey, error) {
	return ua.updateUserAPIKey(ctx, req)
}

func (ua *userAccess) DeleteAPIKey(
	ctx context.Context, req *ttnpb.DeleteUserAPIKeyRequest,
) (*emptypb.Empty, error) {
	return ua.deleteUserAPIKey(ctx, req)
}

func (ua *userAccess) CreateLoginToken(
	ctx context.Context, req *ttnpb.CreateLoginTokenRequest,
) (*ttnpb.CreateLoginTokenResponse, error) {
	return ua.createLoginToken(ctx, req)
}
