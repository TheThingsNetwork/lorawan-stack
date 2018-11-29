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
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc"
)

func init() {
	userAccessUser.Admin = false
	userAccessUser.State = ttnpb.STATE_APPROVED
	APIKeys := userAPIKeys(&userAccessUser.UserIdentifiers)
	for _, APIKey := range APIKeys.APIKeys {
		APIKey.Rights = []ttnpb.Right{ttnpb.RIGHT_USER_SETTINGS_API_KEYS}
	}
}

func TestUserAccessNotFound(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, creds := defaultUser.UserIdentifiers, userCreds(defaultUserIdx)

		reg := ttnpb.NewUserAccessClient(cc)

		APIKey := ttnpb.APIKey{
			ID:   "does-not-exist-id",
			Name: "test-user-api-key-name",
		}

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateUserAPIKeyRequest{
			UserIdentifiers: userID,
			APIKey:          APIKey,
		}, creds)

		a.So(updated, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsNotFound(err), should.BeTrue)
	})
}

func TestUserAccessRightsPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, creds := userAccessUser.UserIdentifiers, userCreds(userAccessUserIdx)

		reg := ttnpb.NewUserAccessClient(cc)

		APIKey, err := reg.CreateAPIKey(ctx, &ttnpb.CreateUserAPIKeyRequest{
			UserIdentifiers: userID,
			Name:            "test-api-key-name",
			Rights:          []ttnpb.Right{ttnpb.RIGHT_USER_ALL},
		}, creds)

		a.So(APIKey, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		APIKey = userAPIKeys(&userID).APIKeys[0]
		APIKey.Rights = []ttnpb.Right{ttnpb.RIGHT_USER_ALL}

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateUserAPIKeyRequest{
			UserIdentifiers: userID,
			APIKey:          *APIKey,
		}, creds)

		a.So(updated, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)
	})
}

func TestUserAccessPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := population.Users[0].UserIdentifiers

		reg := ttnpb.NewUserAccessClient(cc)

		rights, err := reg.ListRights(ctx, &userID)

		a.So(rights, should.NotBeNil)
		a.So(rights.Rights, should.BeEmpty)
		a.So(err, should.BeNil)

		APIKeys, err := reg.ListAPIKeys(ctx, &userID)

		a.So(APIKeys, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		APIKey, err := reg.CreateAPIKey(ctx, &ttnpb.CreateUserAPIKeyRequest{
			UserIdentifiers: userID,
			Name:            "test-api-key-name",
			Rights:          []ttnpb.Right{ttnpb.RIGHT_ALL},
		})

		a.So(APIKey, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		APIKey = userAPIKeys(&userID).APIKeys[0]

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateUserAPIKeyRequest{
			UserIdentifiers: userID,
			APIKey:          *APIKey,
		})

		a.So(updated, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)
	})
}

func TestUserAccessCRUD(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		user, creds := population.Users[0], userCreds(0)

		reg := ttnpb.NewUserAccessClient(cc)

		rights, err := reg.ListRights(ctx, &user.UserIdentifiers, creds)

		a.So(rights, should.NotBeNil)
		a.So(rights.Rights, should.NotBeEmpty)
		a.So(err, should.BeNil)

		userAPIKeys := userAPIKeys(&user.UserIdentifiers)
		APIKeys, err := reg.ListAPIKeys(ctx, &user.UserIdentifiers, creds)

		a.So(APIKeys, should.NotBeNil)
		a.So(err, should.BeNil)
		a.So(len(APIKeys.APIKeys), should.Equal, len(userAPIKeys.APIKeys))
		for i, APIkey := range APIKeys.APIKeys {
			a.So(APIkey.Name, should.Equal, userAPIKeys.APIKeys[i].Name)
			a.So(APIkey.ID, should.Equal, userAPIKeys.APIKeys[i].ID)
		}

		createdAPIKeyName := "test-created-api-key"
		created, err := reg.CreateAPIKey(ctx, &ttnpb.CreateUserAPIKeyRequest{
			UserIdentifiers: user.UserIdentifiers,
			Name:            createdAPIKeyName,
			Rights:          []ttnpb.Right{ttnpb.RIGHT_ALL},
		}, creds)

		a.So(created, should.NotBeNil)
		a.So(created.Name, should.Equal, createdAPIKeyName)
		a.So(err, should.BeNil)

		newAPIKeyName := "test-new-api-key"
		created.Name = newAPIKeyName
		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateUserAPIKeyRequest{
			UserIdentifiers: user.UserIdentifiers,
			APIKey:          *created,
		}, creds)

		a.So(updated, should.NotBeNil)
		a.So(updated.Name, should.Equal, newAPIKeyName)
		a.So(err, should.BeNil)
	})
}
