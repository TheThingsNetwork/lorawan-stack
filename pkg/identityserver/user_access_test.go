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
	"sort"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"google.golang.org/grpc"
)

func init() {
	userAccessUser.Admin = false
	userAccessUser.State = ttnpb.State_STATE_APPROVED
	for _, apiKey := range userAPIKeys(userAccessUser.GetIds()).ApiKeys {
		apiKey.Rights = []ttnpb.Right{ttnpb.RIGHT_USER_SETTINGS_API_KEYS}
	}
}

func TestUserAccessNotFound(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, creds := population.Users[defaultUserIdx].GetIds(), userCreds(defaultUserIdx)

		reg := ttnpb.NewUserAccessClient(cc)

		apiKey := ttnpb.APIKey{
			Id:   "does-not-exist-id",
			Name: "test-user-api-key-name",
		}

		got, err := reg.GetAPIKey(ctx, &ttnpb.GetUserAPIKeyRequest{
			UserIds: userID,
			KeyId:   apiKey.Id,
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		a.So(got, should.BeNil)

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateUserAPIKeyRequest{
			UserIds:   userID,
			ApiKey:    &apiKey,
			FieldMask: &pbtypes.FieldMask{Paths: []string{"name"}},
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		a.So(updated, should.BeNil)
	})
}

func TestUserAccessRightsPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, creds := userAccessUser.GetIds(), userCreds(userAccessUserIdx)

		reg := ttnpb.NewUserAccessClient(cc)

		apiKey, err := reg.CreateAPIKey(ctx, &ttnpb.CreateUserAPIKeyRequest{
			UserIds: userID,
			Name:    "test-api-key-name",
			Rights:  []ttnpb.Right{ttnpb.RIGHT_USER_ALL},
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(apiKey, should.BeNil)

		apiKey = userAPIKeys(userID).ApiKeys[0]
		apiKey.Rights = []ttnpb.Right{ttnpb.RIGHT_USER_ALL}

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateUserAPIKeyRequest{
			UserIds:   userID,
			ApiKey:    apiKey,
			FieldMask: &pbtypes.FieldMask{Paths: []string{"rights", "name"}},
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(updated, should.BeNil)
	})
}

func TestUserAccessPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := population.Users[defaultUserIdx].GetIds()
		apiKeyID := userAPIKeys(userID).ApiKeys[0].Id

		reg := ttnpb.NewUserAccessClient(cc)

		rights, err := reg.ListRights(ctx, userID)

		a.So(err, should.BeNil)
		if a.So(rights, should.NotBeNil) {
			a.So(rights.Rights, should.BeEmpty)
		}

		apiKey, err := reg.GetAPIKey(ctx, &ttnpb.GetUserAPIKeyRequest{
			UserIds: userID,
			KeyId:   apiKeyID,
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(apiKey, should.BeNil)

		apiKeys, err := reg.ListAPIKeys(ctx, &ttnpb.ListUserAPIKeysRequest{
			UserIds: userID,
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(apiKeys, should.BeNil)

		apiKey, err = reg.CreateAPIKey(ctx, &ttnpb.CreateUserAPIKeyRequest{
			UserIds: userID,
			Name:    "test-api-key-name",
			Rights:  []ttnpb.Right{ttnpb.RIGHT_ALL},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(apiKey, should.BeNil)

		apiKey = userAPIKeys(userID).ApiKeys[0]

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateUserAPIKeyRequest{
			UserIds:   userID,
			ApiKey:    apiKey,
			FieldMask: &pbtypes.FieldMask{Paths: []string{"rights", "name"}},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(updated, should.BeNil)
	})
}

func TestUserAccessClusterAuth(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := population.Users[defaultUserIdx].GetIds()

		reg := ttnpb.NewUserAccessClient(cc)

		rights, err := reg.ListRights(ctx, userID, is.WithClusterAuth())

		a.So(err, should.BeNil)
		a.So(rights, should.NotBeNil)
		a.So(ttnpb.AllClusterRights.Intersect(ttnpb.AllUserRights).Sub(rights).Rights, should.BeEmpty)
	})
}

func TestUserAccessCRUD(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		user, creds := population.Users[defaultUserIdx], userCreds(defaultUserIdx)

		reg := ttnpb.NewUserAccessClient(cc)

		rights, err := reg.ListRights(ctx, user.GetIds(), creds)

		a.So(err, should.BeNil)
		if a.So(rights, should.NotBeNil) {
			a.So(rights.Rights, should.NotBeEmpty)
		}

		modifiedUserID := &ttnpb.UserIdentifiers{UserId: reverse(user.GetIds().GetUserId())}

		rights, err = reg.ListRights(ctx, modifiedUserID, creds)

		a.So(err, should.BeNil)
		if a.So(rights, should.NotBeNil) {
			a.So(rights.Rights, should.BeEmpty)
		}

		userAPIKeys := userAPIKeys(user.GetIds())
		userKey := userAPIKeys.ApiKeys[0]

		apiKey, err := reg.GetAPIKey(ctx, &ttnpb.GetUserAPIKeyRequest{
			UserIds: user.GetIds(),
			KeyId:   userKey.Id,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(apiKey, should.NotBeNil) {
			a.So(apiKey.Id, should.Equal, userKey.Id)
			a.So(apiKey.Key, should.BeEmpty)
		}

		sort.Slice(userAPIKeys.ApiKeys, func(i int, j int) bool { return userAPIKeys.ApiKeys[i].Name < userAPIKeys.ApiKeys[j].Name })
		apiKeys, err := reg.ListAPIKeys(ctx, &ttnpb.ListUserAPIKeysRequest{
			UserIds: user.GetIds(),
		}, creds)
		sort.Slice(apiKeys.ApiKeys, func(i int, j int) bool { return apiKeys.ApiKeys[i].Name < apiKeys.ApiKeys[j].Name })

		a.So(err, should.BeNil)
		a.So(apiKeys, should.NotBeNil)
		a.So(len(apiKeys.ApiKeys), should.Equal, len(userAPIKeys.ApiKeys))
		for i, APIkey := range apiKeys.ApiKeys {
			a.So(APIkey.Name, should.Equal, userAPIKeys.ApiKeys[i].Name)
			a.So(APIkey.Id, should.Equal, userAPIKeys.ApiKeys[i].Id)
		}

		createdAPIKeyName := "test-created-api-key"
		created, err := reg.CreateAPIKey(ctx, &ttnpb.CreateUserAPIKeyRequest{
			UserIds: user.GetIds(),
			Name:    createdAPIKeyName,
			Rights:  []ttnpb.Right{ttnpb.RIGHT_ALL},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(created, should.NotBeNil) {
			a.So(created.Name, should.Equal, createdAPIKeyName)
		}

		newAPIKeyName := "test-new-api-key"
		created.Name = newAPIKeyName
		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateUserAPIKeyRequest{
			UserIds:   user.GetIds(),
			ApiKey:    created,
			FieldMask: &pbtypes.FieldMask{Paths: []string{"name"}},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(updated, should.NotBeNil) {
			a.So(updated.Name, should.Equal, newAPIKeyName)
		}
	})
}

func TestUserAccesLoginTokens(t *testing.T) {
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		is.config.LoginTokens.Enabled = false
		user, _ := population.Users[defaultUserIdx], userCreds(defaultUserIdx)
		reg := ttnpb.NewUserAccessClient(cc)
		_, err := reg.CreateLoginToken(ctx, &ttnpb.CreateLoginTokenRequest{
			UserIds: user.GetIds(),
		})
		if a.So(err, should.NotBeNil) {
			a.So(errors.Resemble(err, errLoginTokensDisabled), should.BeTrue)
		}
	})

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		is.config.LoginTokens.Enabled = true
		is.config.LoginTokens.TokenTTL = 10 * time.Minute

		user, _ := population.Users[defaultUserIdx], userCreds(defaultUserIdx)
		adminUser, adminCreds := population.Users[adminUserIdx], userCreds(adminUserIdx)

		reg := ttnpb.NewUserAccessClient(cc)

		token, err := reg.CreateLoginToken(ctx, &ttnpb.CreateLoginTokenRequest{
			UserIds: user.GetIds(),
		})
		if a.So(err, should.BeNil) {
			a.So(token.Token, should.BeBlank)
		}

		token, err = reg.CreateLoginToken(ctx, &ttnpb.CreateLoginTokenRequest{
			UserIds: user.GetIds(),
		}, adminCreds)
		if a.So(err, should.BeNil) {
			a.So(token.Token, should.NotBeBlank)
		}

		token, err = reg.CreateLoginToken(ctx, &ttnpb.CreateLoginTokenRequest{
			UserIds: adminUser.GetIds(),
		}, adminCreds)
		if a.So(err, should.BeNil) {
			a.So(token.Token, should.BeBlank)
		}

		token, err = reg.CreateLoginToken(ctx, &ttnpb.CreateLoginTokenRequest{
			UserIds: adminUser.GetIds(),
		}, adminCreds)
		if a.So(err, should.BeNil) {
			a.So(token.Token, should.BeBlank)
		}
	})
}
