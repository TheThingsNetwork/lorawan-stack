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

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

func (is *IdentityServer) listUserRights(ctx context.Context, ids *ttnpb.UserIdentifiers) (*ttnpb.Rights, error) {
	rights, ok := rights.FromContext(ctx)
	if !ok {
		return &ttnpb.Rights{}, nil
	}
	usrRights, ok := rights.UserRights[unique.ID(ctx, ids)]
	if !ok || usrRights == nil {
		return &ttnpb.Rights{}, nil
	}
	return usrRights, nil
}

func (is *IdentityServer) createUserAPIKey(ctx context.Context, req *ttnpb.CreateUserAPIKeyRequest) (key *ttnpb.APIKey, err error) {
	err = rights.RequireUser(ctx, req.UserIdentifiers, ttnpb.RIGHT_USER_SETTINGS_API_KEYS)
	if err != nil {
		return nil, err
	}
	err = rights.RequireUser(ctx, req.UserIdentifiers, req.Rights...)
	if err != nil {
		return nil, err
	}
	id, err := auth.GenerateID(ctx)
	if err != nil {
		return nil, err
	}
	token, err := auth.APIKey.Generate(ctx, id)
	if err != nil {
		return nil, err
	}
	key = &ttnpb.APIKey{
		ID:     id,
		Key:    token,
		Name:   req.Name,
		Rights: req.Rights,
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		keyStore := store.GetAPIKeyStore(db)
		err = keyStore.CreateAPIKey(ctx, req.UserIdentifiers.EntityIdentifiers(), key)
		return err
	})
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (is *IdentityServer) listUserAPIKeys(ctx context.Context, ids *ttnpb.UserIdentifiers) (keys *ttnpb.APIKeys, err error) {
	err = rights.RequireUser(ctx, *ids, ttnpb.RIGHT_USER_SETTINGS_API_KEYS)
	if err != nil {
		return nil, err
	}
	keys = new(ttnpb.APIKeys)
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		keyStore := store.GetAPIKeyStore(db)
		keys.APIKeys, err = keyStore.FindAPIKeys(ctx, ids.EntityIdentifiers())
		return err
	})
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func (is *IdentityServer) updateUserAPIKey(ctx context.Context, req *ttnpb.UpdateUserAPIKeyRequest) (key *ttnpb.APIKey, err error) {
	err = rights.RequireUser(ctx, req.UserIdentifiers, ttnpb.RIGHT_USER_SETTINGS_API_KEYS)
	if err != nil {
		return nil, err
	}
	err = rights.RequireUser(ctx, req.UserIdentifiers, req.Rights...)
	if err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		keyStore := store.GetAPIKeyStore(db)
		key, err = keyStore.UpdateAPIKey(ctx, req.UserIdentifiers.EntityIdentifiers(), &req.APIKey)
		return err
	})
	if err != nil {
		return nil, err
	}
	if key == nil {
		return &ttnpb.APIKey{}, nil
	}
	return key, nil
}

type userAccess struct {
	*IdentityServer
}

func (ua *userAccess) ListRights(ctx context.Context, req *ttnpb.UserIdentifiers) (*ttnpb.Rights, error) {
	return ua.listUserRights(ctx, req)
}
func (ua *userAccess) CreateAPIKey(ctx context.Context, req *ttnpb.CreateUserAPIKeyRequest) (*ttnpb.APIKey, error) {
	return ua.createUserAPIKey(ctx, req)
}
func (ua *userAccess) ListAPIKeys(ctx context.Context, req *ttnpb.UserIdentifiers) (*ttnpb.APIKeys, error) {
	return ua.listUserAPIKeys(ctx, req)
}
func (ua *userAccess) UpdateAPIKey(ctx context.Context, req *ttnpb.UpdateUserAPIKeyRequest) (*ttnpb.APIKey, error) {
	return ua.updateUserAPIKey(ctx, req)
}
