// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/storetest"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func TestUserAPIKeysPermissions(t *testing.T) { // nolint:gocyclo
	p := &storetest.Population{}

	readOnlyAdmin := p.NewUser()
	readOnlyAdmin.Admin = true
	readOnlyAdminKey, _ := p.NewAPIKey(readOnlyAdmin.GetEntityIdentifiers(), ttnpb.AllReadAdminRights.GetRights()...)
	readOnlyAdminKeyCreds := rpcCreds(readOnlyAdminKey)

	usr1 := p.NewUser()
	usr1Key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_USER_INFO,
		ttnpb.Right_RIGHT_USER_APPLICATIONS_LIST,
	)
	usr1Creds := rpcCreds(usr1Key)

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		is.config.AdminRights.All = true

		reg := ttnpb.NewUserAccessClient(cc)

		for _, opts := range [][]grpc.CallOption{nil, {usr1Creds}, {readOnlyAdminKeyCreds}} {
			created, err := reg.CreateAPIKey(ctx, &ttnpb.CreateUserAPIKeyRequest{
				UserIds: usr1.GetIds(),
				Name:    "api-key-name",
				Rights:  []ttnpb.Right{ttnpb.Right_RIGHT_USER_INFO},
			}, opts...)
			if a.So(err, should.NotBeNil) && a.So(errors.IsPermissionDenied(err), should.BeTrue) {
				a.So(created, should.BeNil)
			}

			list, err := reg.ListAPIKeys(ctx, &ttnpb.ListUserAPIKeysRequest{
				UserIds: usr1.GetIds(),
			}, opts...)
			if a.So(err, should.NotBeNil) && a.So(errors.IsPermissionDenied(err), should.BeTrue) {
				a.So(list, should.BeNil)
			}

			got, err := reg.GetAPIKey(ctx, &ttnpb.GetUserAPIKeyRequest{
				UserIds: usr1.GetIds(),
				KeyId:   usr1Key.GetId(),
			}, opts...)
			if a.So(err, should.NotBeNil) && a.So(errors.IsPermissionDenied(err), should.BeTrue) {
				a.So(got, should.BeNil)
			}

			updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateUserAPIKeyRequest{
				UserIds: usr1.GetIds(),
				ApiKey: &ttnpb.APIKey{
					Id:   usr1Key.GetId(),
					Name: "api-key-name-updated",
				},
				FieldMask: ttnpb.FieldMask("name"),
			}, opts...)
			if a.So(err, should.NotBeNil) && a.So(errors.IsPermissionDenied(err), should.BeTrue) {
				a.So(updated, should.BeNil)
			}

			_, err = reg.DeleteAPIKey(ctx, &ttnpb.DeleteUserAPIKeyRequest{
				UserIds: usr1.GetIds(),
				KeyId:   usr1Key.GetId(),
			}, opts...)
			if !a.So(errors.IsPermissionDenied(err), should.BeTrue) {
				t.FailNow()
			}
		}
	}, withPrivateTestDatabase(p))
}

func TestUserAPIKeys(t *testing.T) { // nolint:gocyclo
	p := &storetest.Population{}

	admin := p.NewUser()
	admin.Admin = true
	adminKey, _ := p.NewAPIKey(admin.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	adminCreds := rpcCreds(adminKey)

	usr1 := p.NewUser()
	usr1Key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_USER_INFO,
		ttnpb.Right_RIGHT_USER_APPLICATIONS_LIST,
	)
	limitedKey, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_USER_INFO,
		ttnpb.Right_RIGHT_USER_SETTINGS_BASIC,
		ttnpb.Right_RIGHT_USER_SETTINGS_API_KEYS,
	)
	limitedCreds := rpcCreds(limitedKey)

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		is.config.AdminRights.All = true

		reg := ttnpb.NewUserAccessClient(cc)

		// GetAPIKey that doesn't exist.
		got, err := reg.GetAPIKey(ctx, &ttnpb.GetUserAPIKeyRequest{
			UserIds: usr1.GetIds(),
			KeyId:   "does-not-exist",
		}, limitedCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		a.So(got, should.BeNil)

		// UpdateAPIKey that doesn't exist.
		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateUserAPIKeyRequest{
			UserIds: usr1.GetIds(),
			ApiKey: &ttnpb.APIKey{
				Id: "does-not-exist",
			},
			FieldMask: ttnpb.FieldMask("name"),
		}, limitedCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		a.So(updated, should.BeNil)

		// CreateAPIKey with rights that caller doesn't have.
		apiKey, err := reg.CreateAPIKey(ctx, &ttnpb.CreateUserAPIKeyRequest{
			UserIds: usr1.GetIds(),
			Name:    "api-key-name",
			Rights:  []ttnpb.Right{ttnpb.Right_RIGHT_USER_ALL},
		}, limitedCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(apiKey, should.BeNil)

		// UpdateAPIKey adding rights that caller doesn't have.
		updated, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateUserAPIKeyRequest{
			UserIds: usr1.GetIds(),
			ApiKey: &ttnpb.APIKey{
				Id: usr1Key.GetId(),
				Rights: []ttnpb.Right{
					ttnpb.Right_RIGHT_USER_INFO,
					ttnpb.Right_RIGHT_USER_APPLICATIONS_LIST,
					ttnpb.Right_RIGHT_USER_DELETE,
				},
			},
			FieldMask: ttnpb.FieldMask("rights"),
		}, limitedCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(updated, should.BeNil)

		// UpdateAPIKey removing rights that caller doesn't have.
		updated, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateUserAPIKeyRequest{
			UserIds: usr1.GetIds(),
			ApiKey: &ttnpb.APIKey{
				Id: usr1Key.GetId(),
				Rights: []ttnpb.Right{
					ttnpb.Right_RIGHT_USER_INFO,
				},
			},
			FieldMask: ttnpb.FieldMask("rights"),
		}, limitedCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(updated, should.BeNil)

		// UpdateAPIKey removing rights that caller has and adding rights that caller has.
		updated, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateUserAPIKeyRequest{
			UserIds: usr1.GetIds(),
			ApiKey: &ttnpb.APIKey{
				Id: usr1Key.GetId(),
				Rights: []ttnpb.Right{
					ttnpb.Right_RIGHT_USER_SETTINGS_BASIC,
					ttnpb.Right_RIGHT_USER_APPLICATIONS_LIST,
				},
			},
			FieldMask: ttnpb.FieldMask("rights"),
		}, limitedCreds)
		if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) {
			a.So(updated.Rights, should.Resemble, []ttnpb.Right{
				ttnpb.Right_RIGHT_USER_SETTINGS_BASIC,
				ttnpb.Right_RIGHT_USER_APPLICATIONS_LIST,
			})
		}

		// API Key CRUD with different valid credentials.
		for _, opts := range [][]grpc.CallOption{{adminCreds}, {limitedCreds}} {
			created, err := reg.CreateAPIKey(ctx, &ttnpb.CreateUserAPIKeyRequest{
				UserIds: usr1.GetIds(),
				Name:    "api-key-name",
				Rights:  []ttnpb.Right{ttnpb.Right_RIGHT_USER_INFO},
			}, opts...)
			if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) {
				a.So(created.Name, should.Equal, "api-key-name")
				a.So(created.Rights, should.Resemble, []ttnpb.Right{ttnpb.Right_RIGHT_USER_INFO})
			}

			list, err := reg.ListAPIKeys(ctx, &ttnpb.ListUserAPIKeysRequest{
				UserIds: usr1.GetIds(),
			}, opts...)
			if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) && a.So(list.ApiKeys, should.HaveLength, 3) {
				for _, k := range list.ApiKeys {
					if k.Id == created.Id {
						a.So(k.Name, should.Resemble, created.Name)
					}
				}
			}

			got, err := reg.GetAPIKey(ctx, &ttnpb.GetUserAPIKeyRequest{
				UserIds: usr1.GetIds(),
				KeyId:   created.GetId(),
			}, opts...)
			if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
				a.So(got.Name, should.Equal, created.Name)
			}

			updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateUserAPIKeyRequest{
				UserIds: usr1.GetIds(),
				ApiKey: &ttnpb.APIKey{
					Id:   created.GetId(),
					Name: "api-key-name-updated",
				},
				FieldMask: ttnpb.FieldMask("name"),
			}, opts...)
			if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) {
				a.So(updated.Name, should.Equal, "api-key-name-updated")
			}

			// TODO: Remove UpdateAPIKey test case (https://github.com/TheThingsNetwork/lorawan-stack/issues/6488).
			t.Run("Delete via update method", func(*testing.T) { // nolint:paralleltest
				_, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateUserAPIKeyRequest{
					UserIds: usr1.GetIds(),
					ApiKey:  &ttnpb.APIKey{Id: created.GetId()},
				}, opts...)
				a.So(err, should.BeNil)

				got, err = reg.GetAPIKey(ctx, &ttnpb.GetUserAPIKeyRequest{
					UserIds: usr1.GetIds(),
					KeyId:   created.GetId(),
				}, opts...)
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsNotFound(err), should.BeTrue)
				}
				a.So(got, should.BeNil)
			})

			// Recreates api-key of the `usr1` User.
			created, err = reg.CreateAPIKey(ctx, &ttnpb.CreateUserAPIKeyRequest{
				UserIds: usr1.GetIds(),
				Name:    "api-key-name",
				Rights:  []ttnpb.Right{ttnpb.Right_RIGHT_USER_INFO},
			}, opts...)
			if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) {
				a.So(created.Name, should.Equal, "api-key-name")
				a.So(created.Rights, should.Resemble, []ttnpb.Right{ttnpb.Right_RIGHT_USER_INFO})
			}

			t.Run("Delete via delete method", func(*testing.T) { // nolint:paralleltest
				empty, err := reg.DeleteAPIKey(ctx, &ttnpb.DeleteUserAPIKeyRequest{
					UserIds: usr1.GetIds(),
					KeyId:   created.GetId(),
				}, opts...)
				a.So(err, should.BeNil)
				a.So(empty, should.Resemble, ttnpb.Empty)

				got, err = reg.GetAPIKey(ctx, &ttnpb.GetUserAPIKeyRequest{
					UserIds: usr1.GetIds(),
					KeyId:   created.GetId(),
				}, opts...)
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsNotFound(err), should.BeTrue)
				}
				a.So(got, should.BeNil)
			})
		}
	}, withPrivateTestDatabase(p))
}

func TestUserAccessClusterAuth(t *testing.T) {
	p := &storetest.Population{}
	usr1 := p.NewUser()

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewUserAccessClient(cc)

		rights, err := reg.ListRights(ctx, usr1.GetIds(), is.WithClusterAuth())
		if a.So(err, should.BeNil) && a.So(rights, should.NotBeNil) {
			a.So(ttnpb.AllClusterRights.Intersect(ttnpb.AllUserRights).Sub(rights).Rights, should.BeEmpty)
		}
	}, withPrivateTestDatabase(p))
}

func TestUserAccesLoginTokens(t *testing.T) {
	p := &storetest.Population{}
	adminUsr := p.NewUser()
	adminUsr.Admin = true
	adminKey, _ := p.NewAPIKey(adminUsr.GetEntityIdentifiers(), ttnpb.Right_RIGHT_USER_ALL)
	adminCreds := rpcCreds(adminKey)
	usr1 := p.NewUser()

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewUserAccessClient(cc)

		is.config.LoginTokens.Enabled = false

		_, err := reg.CreateLoginToken(ctx, &ttnpb.CreateLoginTokenRequest{
			UserIds: usr1.GetIds(),
		})
		if a.So(err, should.NotBeNil) {
			a.So(errors.Resemble(err, errLoginTokensDisabled), should.BeTrue)
		}

		is.config.LoginTokens.Enabled = true
		is.config.LoginTokens.TokenTTL = 10 * time.Minute

		token, err := reg.CreateLoginToken(ctx, &ttnpb.CreateLoginTokenRequest{
			UserIds: usr1.GetIds(),
		})
		if a.So(err, should.BeNil) {
			a.So(token.Token, should.BeBlank)
		}

		token, err = reg.CreateLoginToken(ctx, &ttnpb.CreateLoginTokenRequest{
			UserIds: usr1.GetIds(),
		}, adminCreds)
		if a.So(err, should.BeNil) {
			a.So(token.Token, should.NotBeBlank)
		}

		token, err = reg.CreateLoginToken(ctx, &ttnpb.CreateLoginTokenRequest{
			UserIds: adminUsr.GetIds(),
		}, adminCreds)
		if a.So(err, should.BeNil) {
			a.So(token.Token, should.BeBlank)
		}
	}, withPrivateTestDatabase(p))
}
