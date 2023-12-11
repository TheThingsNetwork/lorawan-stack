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

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/storetest"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func TestGatewayAPIKeys(t *testing.T) { // nolint:gocyclo
	p := &storetest.Population{}

	admin := p.NewUser()
	admin.Admin = true
	adminKey, _ := p.NewAPIKey(admin.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	adminCreds := rpcCreds(adminKey)

	usr1 := p.NewUser()
	usr1Key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	usr1Creds := rpcCreds(usr1Key)

	gtw1 := p.NewGateway(usr1.GetOrganizationOrUserIdentifiers())
	limitedKey, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_GATEWAY_INFO,
		ttnpb.Right_RIGHT_GATEWAY_SETTINGS_BASIC,
		ttnpb.Right_RIGHT_GATEWAY_SETTINGS_API_KEYS,
	)
	limitedCreds := rpcCreds(limitedKey)

	gtwKey, _ := p.NewAPIKey(gtw1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_GATEWAY_INFO,
		ttnpb.Right_RIGHT_GATEWAY_LINK,
	)
	gtwCreds := rpcCreds(gtwKey)

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		is.config.AdminRights.All = true

		reg := ttnpb.NewGatewayAccessClient(cc)

		// GetAPIKey that doesn't exist.
		got, err := reg.GetAPIKey(ctx, &ttnpb.GetGatewayAPIKeyRequest{
			GatewayIds: gtw1.GetIds(),
			KeyId:      "does-not-exist",
		}, limitedCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		a.So(got, should.BeNil)

		// UpdateAPIKey that doesn't exist.
		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateGatewayAPIKeyRequest{
			GatewayIds: gtw1.GetIds(),
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
		apiKey, err := reg.CreateAPIKey(ctx, &ttnpb.CreateGatewayAPIKeyRequest{
			GatewayIds: gtw1.GetIds(),
			Name:       "api-key-name",
			Rights:     []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_ALL},
		}, limitedCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(apiKey, should.BeNil)

		// UpdateAPIKey adding rights that caller doesn't have.
		updated, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateGatewayAPIKeyRequest{
			GatewayIds: gtw1.GetIds(),
			ApiKey: &ttnpb.APIKey{
				Id: gtwKey.GetId(),
				Rights: []ttnpb.Right{
					ttnpb.Right_RIGHT_GATEWAY_INFO,
					ttnpb.Right_RIGHT_GATEWAY_LINK,
					ttnpb.Right_RIGHT_GATEWAY_DELETE,
				},
			},
			FieldMask: ttnpb.FieldMask("rights"),
		}, limitedCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(updated, should.BeNil)

		// UpdateAPIKey removing rights that caller doesn't have.
		updated, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateGatewayAPIKeyRequest{
			GatewayIds: gtw1.GetIds(),
			ApiKey: &ttnpb.APIKey{
				Id: gtwKey.GetId(),
				Rights: []ttnpb.Right{
					ttnpb.Right_RIGHT_GATEWAY_INFO,
				},
			},
			FieldMask: ttnpb.FieldMask("rights"),
		}, limitedCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(updated, should.BeNil)

		// UpdateAPIKey removing rights that caller has and adding rights that caller has.
		updated, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateGatewayAPIKeyRequest{
			GatewayIds: gtw1.GetIds(),
			ApiKey: &ttnpb.APIKey{
				Id: gtwKey.GetId(),
				Rights: []ttnpb.Right{
					ttnpb.Right_RIGHT_GATEWAY_SETTINGS_BASIC,
					ttnpb.Right_RIGHT_GATEWAY_LINK,
				},
			},
			FieldMask: ttnpb.FieldMask("rights"),
		}, limitedCreds)
		if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) {
			a.So(updated.Rights, should.Resemble, []ttnpb.Right{
				ttnpb.Right_RIGHT_GATEWAY_SETTINGS_BASIC,
				ttnpb.Right_RIGHT_GATEWAY_LINK,
			})
		}

		// API Key CRUD with different invalid credentials.
		for _, opts := range [][]grpc.CallOption{nil, {gtwCreds}} {
			created, err := reg.CreateAPIKey(ctx, &ttnpb.CreateGatewayAPIKeyRequest{
				GatewayIds: gtw1.GetIds(),
				Name:       "api-key-name",
				Rights:     []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_INFO},
			}, opts...)
			if a.So(err, should.NotBeNil) && a.So(errors.IsPermissionDenied(err), should.BeTrue) {
				a.So(created, should.BeNil)
			}

			list, err := reg.ListAPIKeys(ctx, &ttnpb.ListGatewayAPIKeysRequest{
				GatewayIds: gtw1.GetIds(),
			}, opts...)
			if a.So(err, should.NotBeNil) && a.So(errors.IsPermissionDenied(err), should.BeTrue) {
				a.So(list, should.BeNil)
			}

			got, err := reg.GetAPIKey(ctx, &ttnpb.GetGatewayAPIKeyRequest{
				GatewayIds: gtw1.GetIds(),
				KeyId:      gtwKey.GetId(),
			}, opts...)
			if a.So(err, should.NotBeNil) && a.So(errors.IsPermissionDenied(err), should.BeTrue) {
				a.So(got, should.BeNil)
			}

			updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateGatewayAPIKeyRequest{
				GatewayIds: gtw1.GetIds(),
				ApiKey: &ttnpb.APIKey{
					Id:   gtwKey.GetId(),
					Name: "api-key-name-updated",
				},
				FieldMask: ttnpb.FieldMask("name"),
			}, opts...)
			if a.So(err, should.NotBeNil) && a.So(errors.IsPermissionDenied(err), should.BeTrue) {
				a.So(updated, should.BeNil)
			}

			_, err = reg.DeleteAPIKey(ctx, &ttnpb.DeleteGatewayAPIKeyRequest{
				GatewayIds: gtw1.GetIds(),
				KeyId:      created.GetId(),
			}, opts...)
			if !a.So(errors.IsPermissionDenied(err), should.BeTrue) {
				t.FailNow()
			}
		}

		// API Key CRUD with different valid credentials.
		for _, opts := range [][]grpc.CallOption{{adminCreds}, {usr1Creds}, {limitedCreds}} {
			created, err := reg.CreateAPIKey(ctx, &ttnpb.CreateGatewayAPIKeyRequest{
				GatewayIds: gtw1.GetIds(),
				Name:       "api-key-name",
				Rights:     []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_INFO},
			}, opts...)
			if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) {
				a.So(created.Name, should.Equal, "api-key-name")
				a.So(created.Rights, should.Resemble, []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_INFO})
			}

			list, err := reg.ListAPIKeys(ctx, &ttnpb.ListGatewayAPIKeysRequest{
				GatewayIds: gtw1.GetIds(),
			}, opts...)
			if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) && a.So(list.ApiKeys, should.HaveLength, 2) {
				for _, k := range list.ApiKeys {
					if k.Id == created.Id {
						a.So(k.Name, should.Resemble, created.Name)
					}
				}
			}

			got, err := reg.GetAPIKey(ctx, &ttnpb.GetGatewayAPIKeyRequest{
				GatewayIds: gtw1.GetIds(),
				KeyId:      created.GetId(),
			}, opts...)
			if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
				a.So(got.Name, should.Equal, created.Name)
			}

			updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateGatewayAPIKeyRequest{
				GatewayIds: gtw1.GetIds(),
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
				_, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateGatewayAPIKeyRequest{
					GatewayIds: gtw1.GetIds(),
					ApiKey:     &ttnpb.APIKey{Id: created.GetId()},
				}, opts...)
				a.So(err, should.BeNil)

				got, err = reg.GetAPIKey(ctx, &ttnpb.GetGatewayAPIKeyRequest{
					GatewayIds: gtw1.GetIds(),
					KeyId:      created.GetId(),
				}, opts...)
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsNotFound(err), should.BeTrue)
				}
				a.So(got, should.BeNil)
			})

			// Recreates api-key of the `gtw1` Gateway.
			created, err = reg.CreateAPIKey(ctx, &ttnpb.CreateGatewayAPIKeyRequest{
				GatewayIds: gtw1.GetIds(),
				Name:       "api-key-name",
				Rights:     []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_INFO},
			}, opts...)
			if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) {
				a.So(created.Name, should.Equal, "api-key-name")
				a.So(created.Rights, should.Resemble, []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_INFO})
			}

			t.Run("Delete via delete method", func(*testing.T) { // nolint:paralleltest
				empty, err := reg.DeleteAPIKey(ctx, &ttnpb.DeleteGatewayAPIKeyRequest{
					GatewayIds: gtw1.GetIds(),
					KeyId:      created.GetId(),
				}, opts...)
				a.So(err, should.BeNil)
				a.So(empty, should.Resemble, ttnpb.Empty)

				got, err = reg.GetAPIKey(ctx, &ttnpb.GetGatewayAPIKeyRequest{
					GatewayIds: gtw1.GetIds(),
					KeyId:      created.GetId(),
				}, opts...)
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsNotFound(err), should.BeTrue)
				}
				a.So(got, should.BeNil)
			})
		}
	}, withPrivateTestDatabase(p))
}

func TestGatewayCollaborators(t *testing.T) { // nolint:gocyclo
	p := &storetest.Population{}

	admin := p.NewUser()
	admin.Admin = true
	adminKey, _ := p.NewAPIKey(admin.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	adminCreds := rpcCreds(adminKey)

	usr1 := p.NewUser()
	usr1Key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	usr1Creds := rpcCreds(usr1Key)

	gtw1 := p.NewGateway(usr1.GetOrganizationOrUserIdentifiers())

	limitedKey, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_GATEWAY_INFO,
		ttnpb.Right_RIGHT_GATEWAY_SETTINGS_BASIC,
		ttnpb.Right_RIGHT_GATEWAY_SETTINGS_COLLABORATORS,
	)
	limitedCreds := rpcCreds(limitedKey)

	gtwKey, _ := p.NewAPIKey(gtw1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_GATEWAY_INFO,
		ttnpb.Right_RIGHT_GATEWAY_LINK,
	)
	gtwCreds := rpcCreds(gtwKey)

	usr2 := p.NewUser()
	p.NewMembership(
		usr2.GetOrganizationOrUserIdentifiers(),
		gtw1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_GATEWAY_INFO,
		ttnpb.Right_RIGHT_GATEWAY_LINK,
	)

	usr3 := p.NewUser()

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		is.config.AdminRights.All = true

		reg := ttnpb.NewGatewayAccessClient(cc)

		// GetCollaborator that doesn't exist.
		got, err := reg.GetCollaborator(ctx, &ttnpb.GetGatewayCollaboratorRequest{
			GatewayIds:   gtw1.GetIds(),
			Collaborator: usr3.GetOrganizationOrUserIdentifiers(),
		}, adminCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		a.So(got, should.BeNil)

		// SetCollaborator adding rights that caller doesn't have.
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetGatewayCollaboratorRequest{
			GatewayIds: gtw1.GetIds(),
			Collaborator: &ttnpb.Collaborator{
				Ids: usr2.GetOrganizationOrUserIdentifiers(),
				Rights: []ttnpb.Right{
					ttnpb.Right_RIGHT_GATEWAY_INFO,
					ttnpb.Right_RIGHT_GATEWAY_LINK,
					ttnpb.Right_RIGHT_GATEWAY_DELETE,
				},
			},
		}, limitedCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		// SetCollaborator removing rights that caller doesn't have.
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetGatewayCollaboratorRequest{
			GatewayIds: gtw1.GetIds(),
			Collaborator: &ttnpb.Collaborator{
				Ids:    usr2.GetOrganizationOrUserIdentifiers(),
				Rights: []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_INFO},
			},
		}, limitedCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		// SetCollaborator removing rights that caller has and adding rights that caller has.
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetGatewayCollaboratorRequest{
			GatewayIds: gtw1.GetIds(),
			Collaborator: &ttnpb.Collaborator{
				Ids:    usr2.GetOrganizationOrUserIdentifiers(),
				Rights: []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_SETTINGS_BASIC, ttnpb.Right_RIGHT_GATEWAY_LINK},
			},
		}, limitedCreds)
		a.So(err, should.BeNil)

		// Collaborator CRUD with different invalid credentials.
		for _, opts := range [][]grpc.CallOption{nil, {gtwCreds}} {
			_, err := reg.SetCollaborator(ctx, &ttnpb.SetGatewayCollaboratorRequest{
				GatewayIds: gtw1.GetIds(),
				Collaborator: &ttnpb.Collaborator{
					Ids:    usr2.GetOrganizationOrUserIdentifiers(),
					Rights: []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_INFO},
				},
			}, opts...)
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsPermissionDenied(err), should.BeTrue)
			}

			got, err := reg.GetCollaborator(ctx, &ttnpb.GetGatewayCollaboratorRequest{
				GatewayIds:   gtw1.GetIds(),
				Collaborator: usr2.GetOrganizationOrUserIdentifiers(),
			}, opts...)
			if a.So(err, should.NotBeNil) && a.So(errors.IsPermissionDenied(err), should.BeTrue) {
				a.So(got, should.BeNil)
			}
		}

		// ListCollaborators without credentials.
		list, err := reg.ListCollaborators(ctx, &ttnpb.ListGatewayCollaboratorsRequest{
			GatewayIds: gtw1.GetIds(),
		})
		if a.So(err, should.NotBeNil) && a.So(errors.IsUnauthenticated(err), should.BeTrue) {
			a.So(list, should.BeNil)
		}

		// Collaborator CRUD with different valid credentials.
		for _, opts := range [][]grpc.CallOption{{adminCreds}, {usr1Creds}, {limitedCreds}} {
			_, err := reg.SetCollaborator(ctx, &ttnpb.SetGatewayCollaboratorRequest{
				GatewayIds: gtw1.GetIds(),
				Collaborator: &ttnpb.Collaborator{
					Ids:    usr3.GetOrganizationOrUserIdentifiers(),
					Rights: []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_INFO},
				},
			}, opts...)
			a.So(err, should.BeNil)

			list, err := reg.ListCollaborators(ctx, &ttnpb.ListGatewayCollaboratorsRequest{
				GatewayIds: gtw1.GetIds(),
			}, opts...)
			if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) && a.So(list.Collaborators, should.HaveLength, 3) {
				for _, k := range list.Collaborators {
					if unique.ID(ctx, k.GetIds()) == unique.ID(ctx, usr3.GetIds()) {
						a.So(k.Rights, should.Resemble, []ttnpb.Right{
							ttnpb.Right_RIGHT_GATEWAY_INFO,
						})
					}
				}
			}

			got, err := reg.GetCollaborator(ctx, &ttnpb.GetGatewayCollaboratorRequest{
				GatewayIds:   gtw1.GetIds(),
				Collaborator: usr3.GetOrganizationOrUserIdentifiers(),
			}, opts...)
			if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
				a.So(got.Rights, should.Resemble, []ttnpb.Right{
					ttnpb.Right_RIGHT_GATEWAY_INFO,
				})
			}

			// TODO: Remove SetCollaborator test case (https://github.com/TheThingsNetwork/lorawan-stack/issues/6488).
			t.Run("Delete via set method", func(*testing.T) { // nolint:paralleltest
				_, err = reg.SetCollaborator(ctx, &ttnpb.SetGatewayCollaboratorRequest{
					GatewayIds: gtw1.GetIds(),
					Collaborator: &ttnpb.Collaborator{
						Ids:    usr3.GetOrganizationOrUserIdentifiers(),
						Rights: []ttnpb.Right{},
					},
				}, opts...)
				a.So(err, should.BeNil)

				// Verifies that it has been deleted.
				got, err := reg.GetCollaborator(ctx, &ttnpb.GetGatewayCollaboratorRequest{
					GatewayIds:   gtw1.GetIds(),
					Collaborator: usr3.GetOrganizationOrUserIdentifiers(),
				}, opts...)
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsNotFound(err), should.BeTrue)
				}
				a.So(got, should.BeNil)
			})

			// Recreates `usr3` collaborator of the `gtw1` gateway.
			_, err = reg.SetCollaborator(ctx, &ttnpb.SetGatewayCollaboratorRequest{
				GatewayIds: gtw1.GetIds(),
				Collaborator: &ttnpb.Collaborator{
					Ids:    usr3.GetOrganizationOrUserIdentifiers(),
					Rights: []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_INFO},
				},
			}, opts...)
			a.So(err, should.BeNil)

			t.Run("Delete via delete method", func(*testing.T) { // nolint:paralleltest
				empty, err := reg.DeleteCollaborator(ctx, &ttnpb.DeleteGatewayCollaboratorRequest{
					GatewayIds:      gtw1.GetIds(),
					CollaboratorIds: usr3.GetOrganizationOrUserIdentifiers(),
				}, opts...)
				a.So(err, should.BeNil)
				a.So(empty, should.Resemble, ttnpb.Empty)

				// Verifies that it has been deleted.
				got, err := reg.GetCollaborator(ctx, &ttnpb.GetGatewayCollaboratorRequest{
					GatewayIds:   gtw1.GetIds(),
					Collaborator: usr3.GetOrganizationOrUserIdentifiers(),
				}, opts...)
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsNotFound(err), should.BeTrue)
				}
				a.So(got, should.BeNil)
			})
		}

		// Try removing the only collaborator with _ALL rights.
		_, err = reg.DeleteCollaborator(ctx, &ttnpb.DeleteGatewayCollaboratorRequest{
			GatewayIds:      gtw1.GetIds(),
			CollaboratorIds: usr1.GetOrganizationOrUserIdentifiers(),
		}, usr1Creds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsFailedPrecondition(err), should.BeTrue)
		}
	}, withPrivateTestDatabase(p))
}

func TestGatewayAccessClusterAuth(t *testing.T) {
	p := &storetest.Population{}
	gtw1 := p.NewGateway(nil)

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewGatewayAccessClient(cc)

		rights, err := reg.ListRights(ctx, gtw1.GetIds(), is.WithClusterAuth())
		if a.So(err, should.BeNil) && a.So(rights, should.NotBeNil) {
			a.So(ttnpb.AllClusterRights.Intersect(ttnpb.AllGatewayRights).Sub(rights).Rights, should.BeEmpty)
		}
	}, withPrivateTestDatabase(p))
}

func TestGatewayBatchAccess(t *testing.T) {
	p := &storetest.Population{}

	t.Parallel()
	a, ctx := test.New(t)

	admin := p.NewUser()
	admin.Admin = true
	adminKey, _ := p.NewAPIKey(admin.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	adminCreds := rpcCreds(adminKey)

	usr1 := p.NewUser()
	usr1Key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	usr1Creds := rpcCreds(usr1Key)

	gtw1 := p.NewGateway(usr1.GetOrganizationOrUserIdentifiers())
	gtw1.StatusPublic = true
	gtw1.LocationPublic = true
	gtw1Key, _ := p.NewAPIKey(gtw1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	gtw1Creds := rpcCreds(gtw1Key)

	gtw2 := p.NewGateway(usr1.GetOrganizationOrUserIdentifiers())

	limitedKey, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_GATEWAY_INFO,
		ttnpb.Right_RIGHT_GATEWAY_SETTINGS_BASIC,
		ttnpb.Right_RIGHT_GATEWAY_SETTINGS_COLLABORATORS,
	)
	limitedCreds := rpcCreds(limitedKey)

	collaboratorWithLimitedRights := p.NewUser()
	p.NewMembership(
		collaboratorWithLimitedRights.GetOrganizationOrUserIdentifiers(),
		gtw2.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_GATEWAY_TRAFFIC_READ,
		ttnpb.Right_RIGHT_GATEWAY_STATUS_READ,
	)
	collaboratorKey, _ := p.NewAPIKey(
		collaboratorWithLimitedRights.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_ALL,
	)
	collaboratorCreds := rpcCreds(collaboratorKey)

	usr2 := p.NewUser()
	p.NewMembership(
		usr2.GetOrganizationOrUserIdentifiers(),
		gtw1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_GATEWAY_INFO,
		ttnpb.Right_RIGHT_GATEWAY_TRAFFIC_READ,
	)

	gtw3 := p.NewGateway(usr2.GetOrganizationOrUserIdentifiers())

	randomUser := p.NewUser()
	randomUserKey, _ := p.NewAPIKey(randomUser.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	randomUserCreds := rpcCreds(randomUserKey)

	statsUser := p.NewUser()
	statsUserKey, _ := p.NewAPIKey(statsUser.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	statsUserCreds := rpcCreds(statsUserKey)

	locUser := p.NewUser()
	locUserKey, _ := p.NewAPIKey(locUser.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	locUserCreds := rpcCreds(locUserKey)

	dualMembershipUser := p.NewUser()
	dualMembershipUserKey, _ := p.NewAPIKey(dualMembershipUser.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	dualMembershipUserCreds := rpcCreds(dualMembershipUserKey)
	p.NewMembership(
		dualMembershipUser.GetOrganizationOrUserIdentifiers(),
		gtw3.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_GATEWAY_ALL,
	)
	org1 := p.NewOrganization(dualMembershipUser.GetOrganizationOrUserIdentifiers())
	p.NewMembership(
		org1.GetOrganizationOrUserIdentifiers(),
		gtw3.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_GATEWAY_ALL,
	)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewGatewayBatchAccessClient(cc)

		for _, tc := range []struct {
			Name               string
			Credentials        grpc.CallOption
			Request            *ttnpb.AssertGatewayRightsRequest
			ErrorAssertion     func(error) bool
			LimitedAdminRights bool
		}{
			{
				Name:           "Empty request",
				Credentials:    usr1Creds,
				Request:        &ttnpb.AssertGatewayRightsRequest{},
				ErrorAssertion: errors.IsInvalidArgument,
			},
			{
				Name:        "No rights requested",
				Credentials: usr1Creds,
				Request: &ttnpb.AssertGatewayRightsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtw1.GetIds(),
						gtw2.GetIds(),
					},
					Required: &ttnpb.Rights{
						Rights: []ttnpb.Right{},
					},
				},
				ErrorAssertion: errors.IsInvalidArgument,
			},
			{
				Name:        "No gateway rights",
				Credentials: usr1Creds,
				Request: &ttnpb.AssertGatewayRightsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtw1.GetIds(),
						gtw2.GetIds(),
					},
					Required: &ttnpb.Rights{
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_APPLICATION_ALL,
						},
					},
				},
				ErrorAssertion: errors.IsInvalidArgument,
			},
			{
				Name:        "Extra rights requested",
				Credentials: usr1Creds,
				Request: &ttnpb.AssertGatewayRightsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtw1.GetIds(),
						gtw2.GetIds(),
					},
					Required: &ttnpb.Rights{
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_APPLICATION_ALL,
							ttnpb.Right_RIGHT_GATEWAY_ALL,
						},
					},
				},
				ErrorAssertion: errors.IsInvalidArgument,
			},
			{
				Name:        "Not user or organization",
				Credentials: gtw1Creds,
				Request: &ttnpb.AssertGatewayRightsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtw1.GetIds(),
						gtw2.GetIds(),
					},
					Required: &ttnpb.Rights{
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_GATEWAY_ALL,
						},
					},
				},
				ErrorAssertion: errors.IsPermissionDenied,
			},
			{
				Name: "Limited rights",
				Request: &ttnpb.AssertGatewayRightsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtw1.GetIds(),
						gtw2.GetIds(),
					},
					Required: &ttnpb.Rights{
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_GATEWAY_ALL,
						},
					},
				},
				Credentials:    limitedCreds,
				ErrorAssertion: errors.IsPermissionDenied,
			},
			{
				Name: "Random user",
				Request: &ttnpb.AssertGatewayRightsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtw1.GetIds(),
						gtw2.GetIds(),
					},
					Required: &ttnpb.Rights{
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_GATEWAY_ALL,
						},
					},
				},
				Credentials:    randomUserCreds,
				ErrorAssertion: errors.IsPermissionDenied,
			},
			{
				Name: "One gateway not found",
				Request: &ttnpb.AssertGatewayRightsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtw1.GetIds(),
						gtw2.GetIds(),
						{
							GatewayId: "unknown",
						},
					},
					Required: &ttnpb.Rights{
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_GATEWAY_ALL,
						},
					},
				},
				Credentials:    usr1Creds,
				ErrorAssertion: errors.IsPermissionDenied,
			},
			{
				Name: "One gateway not found for admin",
				Request: &ttnpb.AssertGatewayRightsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtw1.GetIds(),
						gtw2.GetIds(),
						{
							GatewayId: "unknown",
						},
					},
					Required: &ttnpb.Rights{
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_GATEWAY_ALL,
						},
					},
				},
				Credentials:    adminCreds,
				ErrorAssertion: errors.IsNotFound,
			},
			{
				Name: "No rights for one gateway",
				Request: &ttnpb.AssertGatewayRightsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtw1.GetIds(),
						gtw2.GetIds(),
						gtw3.GetIds(),
					},
					Required: &ttnpb.Rights{
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_GATEWAY_ALL,
						},
					},
				},
				Credentials:    usr1Creds,
				ErrorAssertion: errors.IsPermissionDenied,
			},
			{
				Name: "Collaborator with limited rights",
				Request: &ttnpb.AssertGatewayRightsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtw1.GetIds(),
					},
					Required: &ttnpb.Rights{
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_GATEWAY_SETTINGS_API_KEYS,
						},
					},
				},
				Credentials:    collaboratorCreds,
				ErrorAssertion: errors.IsPermissionDenied,
			},
			{
				Name: "Universal rights for admin",
				Request: &ttnpb.AssertGatewayRightsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtw3.GetIds(),
					},
					Required: &ttnpb.Rights{
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_GATEWAY_SETTINGS_API_KEYS,
						},
					},
				},
				Credentials: adminCreds,
			},
			{
				Name: "Limited rights for admin",
				Request: &ttnpb.AssertGatewayRightsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtw1.GetIds(),
					},
					Required: &ttnpb.Rights{
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_GATEWAY_SETTINGS_API_KEYS,
						},
					},
				},
				Credentials:        adminCreds,
				ErrorAssertion:     errors.IsPermissionDenied,
				LimitedAdminRights: true,
			},
			{
				Name: "Read Stats Failure",
				Request: &ttnpb.AssertGatewayRightsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtw1.GetIds(),
						gtw2.GetIds(),
					},
					Required: &ttnpb.Rights{
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_GATEWAY_STATUS_READ,
						},
					},
				},
				Credentials:    statsUserCreds,
				ErrorAssertion: errors.IsPermissionDenied,
			},
			{
				Name: "Read Location Failure",
				Request: &ttnpb.AssertGatewayRightsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtw1.GetIds(),
						gtw3.GetIds(),
					},
					Required: &ttnpb.Rights{
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_GATEWAY_LOCATION_READ,
						},
					},
				},
				Credentials:    locUserCreds,
				ErrorAssertion: errors.IsPermissionDenied,
			},
			{
				Name: "Collaborator with insufficient rights for one gateway",
				Request: &ttnpb.AssertGatewayRightsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtw1.GetIds(),
						gtw2.GetIds(),
					},
					Required: &ttnpb.Rights{
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_GATEWAY_STATUS_READ,
							ttnpb.Right_RIGHT_GATEWAY_TRAFFIC_READ,
						},
					},
				},
				Credentials:    collaboratorCreds,
				ErrorAssertion: errors.IsPermissionDenied,
			},
			{
				Name: "Collaborator with excessive rights requested",
				Request: &ttnpb.AssertGatewayRightsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtw2.GetIds(),
					},
					Required: &ttnpb.Rights{
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_GATEWAY_STATUS_READ,
							ttnpb.Right_RIGHT_GATEWAY_TRAFFIC_READ,
							ttnpb.Right_RIGHT_GATEWAY_SETTINGS_API_KEYS,
						},
					},
				},
				Credentials:    collaboratorCreds,
				ErrorAssertion: errors.IsPermissionDenied,
			},
			{
				Name: "Not allowed to read private stats",
				Request: &ttnpb.AssertGatewayRightsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtw2.GetIds(),
					},
					Required: &ttnpb.Rights{
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_GATEWAY_STATUS_READ,
						},
					},
				},
				Credentials:    statsUserCreds,
				ErrorAssertion: errors.IsPermissionDenied,
			},
			{
				Name: "Not allowed to read private location",
				Request: &ttnpb.AssertGatewayRightsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtw2.GetIds(),
					},
					Required: &ttnpb.Rights{
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_GATEWAY_LOCATION_READ,
						},
					},
				},
				Credentials:    locUserCreds,
				ErrorAssertion: errors.IsPermissionDenied,
			},
			{
				Name: "Read Public Stats",
				Request: &ttnpb.AssertGatewayRightsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtw1.GetIds(),
					},
					Required: &ttnpb.Rights{
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_GATEWAY_STATUS_READ,
						},
					},
				},
				Credentials: statsUserCreds,
			},
			{
				Name: "Read Public Location",
				Request: &ttnpb.AssertGatewayRightsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtw1.GetIds(),
					},
					Required: &ttnpb.Rights{
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_GATEWAY_LOCATION_READ,
						},
					},
				},
				Credentials: locUserCreds,
			},
			{
				Name: "Valid collaborator",
				Request: &ttnpb.AssertGatewayRightsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtw2.GetIds(),
					},
					Required: &ttnpb.Rights{
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_GATEWAY_STATUS_READ,
							ttnpb.Right_RIGHT_GATEWAY_TRAFFIC_READ,
						},
					},
				},
				Credentials: collaboratorCreds,
			},
			{
				Name: "Valid request",
				Request: &ttnpb.AssertGatewayRightsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtw1.GetIds(),
						gtw2.GetIds(),
					},
					Required: &ttnpb.Rights{
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_GATEWAY_SETTINGS_API_KEYS,
							ttnpb.Right_RIGHT_GATEWAY_TRAFFIC_DOWN_WRITE,
							ttnpb.Right_RIGHT_GATEWAY_WRITE_SECRETS,
						},
					},
				},
				Credentials: usr1Creds,
			},
			{
				Name: "Dual membership user",
				Request: &ttnpb.AssertGatewayRightsRequest{
					GatewayIds: []*ttnpb.GatewayIdentifiers{
						gtw3.GetIds(),
					},
					Required: &ttnpb.Rights{
						Rights: []ttnpb.Right{
							ttnpb.Right_RIGHT_GATEWAY_ALL,
						},
					},
				},
				Credentials: dualMembershipUserCreds,
			},
		} {
			tc := tc
			t.Run(tc.Name, func(t *testing.T) {
				is.config.AdminRights.All = !tc.LimitedAdminRights

				_, err := reg.AssertRights(ctx, tc.Request, tc.Credentials)
				if err != nil {
					if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue) {
						t.Fatalf("Unexpected error: %v", err)
					}
				} else if tc.ErrorAssertion != nil {
					t.Fatalf("Expected error")
				}
			})
		}
	}, withPrivateTestDatabase(p))
}
