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

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/email"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/identityserver/emails"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtCreateUserAPIKey = events.Define(
		"user.api-key.create", "create user API key",
		ttnpb.RIGHT_USER_SETTINGS_API_KEYS,
	)
	evtUpdateUserAPIKey = events.Define(
		"user.api-key.update", "update user API key",
		ttnpb.RIGHT_USER_SETTINGS_API_KEYS,
	)
	evtDeleteUserAPIKey = events.Define(
		"user.api-key.delete", "delete user API key",
		ttnpb.RIGHT_USER_SETTINGS_API_KEYS,
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
	key, token, err := generateAPIKey(ctx, req.Name, req.Rights...)
	if err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		return store.GetAPIKeyStore(db).CreateAPIKey(ctx, req.UserIdentifiers, key)
	})
	if err != nil {
		return nil, err
	}
	key.Key = token
	events.Publish(evtCreateUserAPIKey(ctx, req.UserIdentifiers, nil))
	err = is.SendUserEmail(ctx, &req.UserIdentifiers, func(data emails.Data) email.MessageData {
		data.SetEntity(req.EntityIdentifiers())
		return &emails.APIKeyCreated{Data: data, Identifier: key.PrettyName(), Rights: key.Rights}
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
		keys.APIKeys, err = store.GetAPIKeyStore(db).FindAPIKeys(ctx, req.UserIdentifiers)
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
		_, key, err = store.GetAPIKeyStore(db).GetAPIKey(ctx, req.KeyID)
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
	// Require that caller has at least the rights of the API key.
	if err = rights.RequireUser(ctx, req.UserIdentifiers, req.Rights...); err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		key, err = store.GetAPIKeyStore(db).UpdateAPIKey(ctx, req.UserIdentifiers, &req.APIKey)
		return err
	})
	if err != nil {
		return nil, err
	}
	if key == nil {
		return &ttnpb.APIKey{}, nil
	}
	key.Key = ""
	if len(req.Rights) > 0 {
		events.Publish(evtUpdateUserAPIKey(ctx, req.UserIdentifiers, nil))
		err = is.SendUserEmail(ctx, &req.UserIdentifiers, func(data emails.Data) email.MessageData {
			data.SetEntity(req.EntityIdentifiers())
			return &emails.APIKeyChanged{Data: data, Identifier: key.PrettyName(), Rights: key.Rights}
		})
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Could not send API key update notification email")
		}
	} else {
		events.Publish(evtDeleteUserAPIKey(ctx, req.UserIdentifiers, nil))
	}
	return key, nil
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
