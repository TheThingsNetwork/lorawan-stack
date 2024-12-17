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

func TestApplicationAPIKeysPermissions(t *testing.T) {
	p := &storetest.Population{}

	readOnlyAdmin := p.NewUser()
	readOnlyAdmin.Admin = true
	readOnlyAdminKey, _ := p.NewAPIKey(readOnlyAdmin.GetEntityIdentifiers(), ttnpb.AllReadAdminRights.GetRights()...)
	readOnlyAdminKeyCreds := rpcCreds(readOnlyAdminKey)

	usr1 := p.NewUser()
	app1 := p.NewApplication(usr1.GetOrganizationOrUserIdentifiers())
	appKey, _ := p.NewAPIKey(app1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_APPLICATION_INFO,
		ttnpb.Right_RIGHT_APPLICATION_LINK,
	)
	appCreds := rpcCreds(appKey)

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		is.config.AdminRights.All = true
		reg := ttnpb.NewApplicationAccessClient(cc)

		t.Run("API Key CRUD with different invalid credentials", func(t *testing.T) { // nolint:paralleltest
			for _, opts := range [][]grpc.CallOption{nil, {appCreds}, {readOnlyAdminKeyCreds}} {
				created, err := reg.CreateAPIKey(ctx, &ttnpb.CreateApplicationAPIKeyRequest{
					ApplicationIds: app1.GetIds(),
					Name:           "api-key-name",
					Rights:         []ttnpb.Right{ttnpb.Right_RIGHT_APPLICATION_INFO},
				}, opts...)
				if a.So(err, should.NotBeNil) && a.So(errors.IsPermissionDenied(err), should.BeTrue) {
					a.So(created, should.BeNil)
				}

				list, err := reg.ListAPIKeys(ctx, &ttnpb.ListApplicationAPIKeysRequest{
					ApplicationIds: app1.GetIds(),
				}, opts...)
				if a.So(err, should.NotBeNil) && a.So(errors.IsPermissionDenied(err), should.BeTrue) {
					a.So(list, should.BeNil)
				}

				got, err := reg.GetAPIKey(ctx, &ttnpb.GetApplicationAPIKeyRequest{
					ApplicationIds: app1.GetIds(),
					KeyId:          appKey.GetId(),
				}, opts...)
				if a.So(err, should.NotBeNil) && a.So(errors.IsPermissionDenied(err), should.BeTrue) {
					a.So(got, should.BeNil)
				}

				updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
					ApplicationIds: app1.GetIds(),
					ApiKey: &ttnpb.APIKey{
						Id:   appKey.GetId(),
						Name: "api-key-name-updated",
					},
					FieldMask: ttnpb.FieldMask("name"),
				}, opts...)
				if a.So(err, should.NotBeNil) && a.So(errors.IsPermissionDenied(err), should.BeTrue) {
					a.So(updated, should.BeNil)
				}

				_, err = reg.DeleteAPIKey(ctx, &ttnpb.DeleteApplicationAPIKeyRequest{
					ApplicationIds: app1.GetIds(),
					KeyId:          created.GetId(),
				}, opts...)
				if !a.So(errors.IsPermissionDenied(err), should.BeTrue) {
					t.FailNow()
				}
			}
		})
	}, withPrivateTestDatabase(p))
}

func TestApplicationAPIKeys(t *testing.T) { // nolint:gocyclo
	p := &storetest.Population{}

	admin := p.NewUser()
	admin.Admin = true
	adminKey, _ := p.NewAPIKey(admin.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	adminCreds := rpcCreds(adminKey)

	usr1 := p.NewUser()
	usr1Key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	usr1Creds := rpcCreds(usr1Key)

	app1 := p.NewApplication(usr1.GetOrganizationOrUserIdentifiers())
	limitedKey, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_APPLICATION_INFO,
		ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC,
		ttnpb.Right_RIGHT_APPLICATION_SETTINGS_API_KEYS,
	)
	limitedCreds := rpcCreds(limitedKey)

	appKey, _ := p.NewAPIKey(app1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_APPLICATION_INFO,
		ttnpb.Right_RIGHT_APPLICATION_LINK,
	)

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		is.config.AdminRights.All = true

		reg := ttnpb.NewApplicationAccessClient(cc)

		// GetAPIKey that doesn't exist.
		got, err := reg.GetAPIKey(ctx, &ttnpb.GetApplicationAPIKeyRequest{
			ApplicationIds: app1.GetIds(),
			KeyId:          "does-not-exist",
		}, limitedCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		a.So(got, should.BeNil)

		// UpdateAPIKey that doesn't exist.
		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIds: app1.GetIds(),
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
		apiKey, err := reg.CreateAPIKey(ctx, &ttnpb.CreateApplicationAPIKeyRequest{
			ApplicationIds: app1.GetIds(),
			Name:           "api-key-name",
			Rights:         []ttnpb.Right{ttnpb.Right_RIGHT_APPLICATION_ALL},
		}, limitedCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(apiKey, should.BeNil)

		// UpdateAPIKey adding rights that caller doesn't have.
		updated, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIds: app1.GetIds(),
			ApiKey: &ttnpb.APIKey{
				Id: appKey.GetId(),
				Rights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_INFO,
					ttnpb.Right_RIGHT_APPLICATION_LINK,
					ttnpb.Right_RIGHT_APPLICATION_DELETE,
				},
			},
			FieldMask: ttnpb.FieldMask("rights"),
		}, limitedCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(updated, should.BeNil)

		// UpdateAPIKey removing rights that caller doesn't have.
		updated, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIds: app1.GetIds(),
			ApiKey: &ttnpb.APIKey{
				Id: appKey.GetId(),
				Rights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_INFO,
				},
			},
			FieldMask: ttnpb.FieldMask("rights"),
		}, limitedCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(updated, should.BeNil)

		// UpdateAPIKey removing rights that caller has and adding rights that caller has.
		updated, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIds: app1.GetIds(),
			ApiKey: &ttnpb.APIKey{
				Id: appKey.GetId(),
				Rights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC,
					ttnpb.Right_RIGHT_APPLICATION_LINK,
				},
			},
			FieldMask: ttnpb.FieldMask("rights"),
		}, limitedCreds)
		if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) {
			a.So(updated.Rights, should.Resemble, []ttnpb.Right{
				ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC,
				ttnpb.Right_RIGHT_APPLICATION_LINK,
			})
		}

		// API Key CRUD with different valid credentials.
		for _, opts := range [][]grpc.CallOption{{adminCreds}, {usr1Creds}, {limitedCreds}} {
			created, err := reg.CreateAPIKey(ctx, &ttnpb.CreateApplicationAPIKeyRequest{
				ApplicationIds: app1.GetIds(),
				Name:           "api-key-name",
				Rights:         []ttnpb.Right{ttnpb.Right_RIGHT_APPLICATION_INFO},
			}, opts...)
			if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) {
				a.So(created.Name, should.Equal, "api-key-name")
				a.So(created.Rights, should.Resemble, []ttnpb.Right{ttnpb.Right_RIGHT_APPLICATION_INFO})
			}

			list, err := reg.ListAPIKeys(ctx, &ttnpb.ListApplicationAPIKeysRequest{
				ApplicationIds: app1.GetIds(),
			}, opts...)
			if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) && a.So(list.ApiKeys, should.HaveLength, 2) {
				for _, k := range list.ApiKeys {
					if k.Id == created.Id {
						a.So(k.Name, should.Resemble, created.Name)
					}
				}
			}

			got, err := reg.GetAPIKey(ctx, &ttnpb.GetApplicationAPIKeyRequest{
				ApplicationIds: app1.GetIds(),
				KeyId:          created.GetId(),
			}, opts...)
			if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
				a.So(got.Name, should.Equal, created.Name)
			}

			updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
				ApplicationIds: app1.GetIds(),
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
				_, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
					ApplicationIds: app1.GetIds(),
					ApiKey:         &ttnpb.APIKey{Id: created.GetId()},
				}, opts...)
				a.So(err, should.BeNil)

				got, err = reg.GetAPIKey(ctx, &ttnpb.GetApplicationAPIKeyRequest{
					ApplicationIds: app1.GetIds(),
					KeyId:          created.GetId(),
				}, opts...)
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsNotFound(err), should.BeTrue)
				}
				a.So(got, should.BeNil)
			})

			// Recreates api-key of the `app1` Application.
			created, err = reg.CreateAPIKey(ctx, &ttnpb.CreateApplicationAPIKeyRequest{
				ApplicationIds: app1.GetIds(),
				Name:           "api-key-name",
				Rights:         []ttnpb.Right{ttnpb.Right_RIGHT_APPLICATION_INFO},
			}, opts...)
			if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) {
				a.So(created.Name, should.Equal, "api-key-name")
				a.So(created.Rights, should.Resemble, []ttnpb.Right{ttnpb.Right_RIGHT_APPLICATION_INFO})
			}

			t.Run("Delete via delete method", func(*testing.T) { // nolint:paralleltest
				empty, err := reg.DeleteAPIKey(ctx, &ttnpb.DeleteApplicationAPIKeyRequest{
					ApplicationIds: app1.GetIds(),
					KeyId:          created.GetId(),
				}, opts...)
				a.So(err, should.BeNil)
				a.So(empty, should.Resemble, ttnpb.Empty)

				got, err = reg.GetAPIKey(ctx, &ttnpb.GetApplicationAPIKeyRequest{
					ApplicationIds: app1.GetIds(),
					KeyId:          created.GetId(),
				}, opts...)
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsNotFound(err), should.BeTrue)
				}
				a.So(got, should.BeNil)
			})
		}
	}, withPrivateTestDatabase(p))
}

func TestApplicationCollaboratorsPermissions(t *testing.T) { // nolint:gocyclo
	p := &storetest.Population{}

	readOnlyAdmin := p.NewUser()
	readOnlyAdmin.Admin = true
	readOnlyAdminKey, _ := p.NewAPIKey(readOnlyAdmin.GetEntityIdentifiers(), ttnpb.AllReadAdminRights.GetRights()...)
	readOnlyAdminKeyCreds := rpcCreds(readOnlyAdminKey)

	usr1 := p.NewUser()
	app1 := p.NewApplication(usr1.GetOrganizationOrUserIdentifiers())
	appKey, _ := p.NewAPIKey(app1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_APPLICATION_INFO,
		ttnpb.Right_RIGHT_APPLICATION_LINK,
	)
	appCreds := rpcCreds(appKey)

	usr2 := p.NewUser()
	p.NewMembership(
		usr2.GetOrganizationOrUserIdentifiers(),
		app1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_APPLICATION_INFO,
		ttnpb.Right_RIGHT_APPLICATION_LINK,
	)

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		is.config.AdminRights.All = true

		reg := ttnpb.NewApplicationAccessClient(cc)

		// Collaborator CRUD with different invalid credentials.
		for _, opts := range [][]grpc.CallOption{nil, {appCreds}, {readOnlyAdminKeyCreds}} {
			_, err := reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
				ApplicationIds: app1.GetIds(),
				Collaborator: &ttnpb.Collaborator{
					Ids: usr2.GetOrganizationOrUserIdentifiers(),
					Rights: []ttnpb.Right{
						ttnpb.Right_RIGHT_APPLICATION_INFO,
					},
				},
			}, opts...)
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsPermissionDenied(err), should.BeTrue)
			}

			got, err := reg.GetCollaborator(ctx, &ttnpb.GetApplicationCollaboratorRequest{
				ApplicationIds: app1.GetIds(),
				Collaborator:   usr2.GetOrganizationOrUserIdentifiers(),
			}, opts...)
			if a.So(err, should.NotBeNil) && a.So(errors.IsPermissionDenied(err), should.BeTrue) {
				a.So(got, should.BeNil)
			}
		}

		// ListCollaborators without credentials.
		list, err := reg.ListCollaborators(ctx, &ttnpb.ListApplicationCollaboratorsRequest{
			ApplicationIds: app1.GetIds(),
		})
		if a.So(err, should.NotBeNil) && a.So(errors.IsUnauthenticated(err), should.BeTrue) {
			a.So(list, should.BeNil)
		}
	}, withPrivateTestDatabase(p))
}

func TestApplicationCollaborators(t *testing.T) { // nolint:gocyclo
	p := &storetest.Population{}

	admin := p.NewUser()
	admin.Admin = true
	adminKey, _ := p.NewAPIKey(admin.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	adminCreds := rpcCreds(adminKey)

	usr1 := p.NewUser()
	usr1Key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	usr1Creds := rpcCreds(usr1Key)

	app1 := p.NewApplication(usr1.GetOrganizationOrUserIdentifiers())

	limitedKey, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_APPLICATION_INFO,
		ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC,
		ttnpb.Right_RIGHT_APPLICATION_SETTINGS_COLLABORATORS,
	)
	limitedCreds := rpcCreds(limitedKey)

	usr2 := p.NewUser()
	p.NewMembership(
		usr2.GetOrganizationOrUserIdentifiers(),
		app1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_APPLICATION_INFO,
		ttnpb.Right_RIGHT_APPLICATION_LINK,
	)

	usr3 := p.NewUser()

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		is.config.AdminRights.All = true

		reg := ttnpb.NewApplicationAccessClient(cc)

		// GetCollaborator that doesn't exist.
		got, err := reg.GetCollaborator(ctx, &ttnpb.GetApplicationCollaboratorRequest{
			ApplicationIds: app1.GetIds(),
			Collaborator:   usr3.GetOrganizationOrUserIdentifiers(),
		}, adminCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		a.So(got, should.BeNil)

		// SetCollaborator adding rights that caller doesn't have.
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIds: app1.GetIds(),
			Collaborator: &ttnpb.Collaborator{
				Ids: usr2.GetOrganizationOrUserIdentifiers(),
				Rights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_INFO,
					ttnpb.Right_RIGHT_APPLICATION_LINK,
					ttnpb.Right_RIGHT_APPLICATION_DELETE,
				},
			},
		}, limitedCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		// SetCollaborator removing rights that caller doesn't have.
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIds: app1.GetIds(),
			Collaborator: &ttnpb.Collaborator{
				Ids: usr2.GetOrganizationOrUserIdentifiers(),
				Rights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_INFO,
				},
			},
		}, limitedCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		// SetCollaborator removing rights that caller has and adding rights that caller has.
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIds: app1.GetIds(),
			Collaborator: &ttnpb.Collaborator{
				Ids: usr2.GetOrganizationOrUserIdentifiers(),
				Rights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_SETTINGS_BASIC,
					ttnpb.Right_RIGHT_APPLICATION_LINK,
				},
			},
		}, limitedCreds)
		a.So(err, should.BeNil)

		// ListCollaborators without credentials.
		list, err := reg.ListCollaborators(ctx, &ttnpb.ListApplicationCollaboratorsRequest{
			ApplicationIds: app1.GetIds(),
		})
		if a.So(err, should.NotBeNil) && a.So(errors.IsUnauthenticated(err), should.BeTrue) {
			a.So(list, should.BeNil)
		}

		// Collaborator CRUD with different valid credentials.
		for _, opts := range [][]grpc.CallOption{{adminCreds}, {usr1Creds}, {limitedCreds}} {
			_, err := reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
				ApplicationIds: app1.GetIds(),
				Collaborator: &ttnpb.Collaborator{
					Ids: usr3.GetOrganizationOrUserIdentifiers(),
					Rights: []ttnpb.Right{
						ttnpb.Right_RIGHT_APPLICATION_INFO,
					},
				},
			}, opts...)
			a.So(err, should.BeNil)

			list, err := reg.ListCollaborators(ctx, &ttnpb.ListApplicationCollaboratorsRequest{
				ApplicationIds: app1.GetIds(),
			}, opts...)
			if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) && a.So(list.Collaborators, should.HaveLength, 3) {
				for _, k := range list.Collaborators {
					if unique.ID(ctx, k.GetIds()) == unique.ID(ctx, usr3.GetIds()) {
						a.So(k.Rights, should.Resemble, []ttnpb.Right{
							ttnpb.Right_RIGHT_APPLICATION_INFO,
						})
					}
				}
			}

			got, err := reg.GetCollaborator(ctx, &ttnpb.GetApplicationCollaboratorRequest{
				ApplicationIds: app1.GetIds(),
				Collaborator:   usr3.GetOrganizationOrUserIdentifiers(),
			}, opts...)
			if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
				a.So(got.Rights, should.Resemble, []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_INFO,
				})
			}

			// TODO: Remove SetCollaborator test case (https://github.com/TheThingsNetwork/lorawan-stack/issues/6488).
			t.Run("Delete via set method", func(*testing.T) { // nolint:paralleltest
				_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
					ApplicationIds: app1.GetIds(),
					Collaborator: &ttnpb.Collaborator{
						Ids:    usr3.GetOrganizationOrUserIdentifiers(),
						Rights: []ttnpb.Right{},
					},
				}, opts...)
				a.So(err, should.BeNil)

				// Verifies that it has been deleted.
				got, err = reg.GetCollaborator(ctx, &ttnpb.GetApplicationCollaboratorRequest{
					ApplicationIds: app1.GetIds(),
					Collaborator:   usr3.GetOrganizationOrUserIdentifiers(),
				}, opts...)
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsNotFound(err), should.BeTrue)
				}
				a.So(got, should.BeNil)
			})

			// Recreates `usr3` collaborator of the `app1` Application.
			_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
				ApplicationIds: app1.GetIds(),
				Collaborator: &ttnpb.Collaborator{
					Ids:    usr3.GetOrganizationOrUserIdentifiers(),
					Rights: []ttnpb.Right{ttnpb.Right_RIGHT_APPLICATION_INFO},
				},
			}, opts...)
			a.So(err, should.BeNil)

			t.Run("Delete via delete method", func(*testing.T) { // nolint:paralleltest
				empty, err := reg.DeleteCollaborator(ctx, &ttnpb.DeleteApplicationCollaboratorRequest{
					ApplicationIds:  app1.GetIds(),
					CollaboratorIds: usr3.GetOrganizationOrUserIdentifiers(),
				}, opts...)
				a.So(err, should.BeNil)
				a.So(empty, should.Resemble, ttnpb.Empty)

				// Verifies that it has been deleted.
				got, err = reg.GetCollaborator(ctx, &ttnpb.GetApplicationCollaboratorRequest{
					ApplicationIds: app1.GetIds(),
					Collaborator:   usr3.GetOrganizationOrUserIdentifiers(),
				}, opts...)
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsNotFound(err), should.BeTrue)
				}
				a.So(got, should.BeNil)
			})
		}

		// Try removing the only collaborator with _ALL rights.
		_, err = reg.DeleteCollaborator(ctx, &ttnpb.DeleteApplicationCollaboratorRequest{
			ApplicationIds:  app1.GetIds(),
			CollaboratorIds: usr1.GetOrganizationOrUserIdentifiers(),
		}, usr1Creds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsFailedPrecondition(err), should.BeTrue)
		}
	}, withPrivateTestDatabase(p))
}

func TestApplicationAccessClusterAuth(t *testing.T) {
	p := &storetest.Population{}
	app1 := p.NewApplication(nil)

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewApplicationAccessClient(cc)

		rights, err := reg.ListRights(ctx, app1.GetIds(), is.WithClusterAuth())
		if a.So(err, should.BeNil) && a.So(rights, should.NotBeNil) {
			a.So(ttnpb.AllClusterRights.Intersect(ttnpb.AllApplicationRights).Sub(rights).Rights, should.BeEmpty)
		}
	}, withPrivateTestDatabase(p))
}

func TestApplicationContactRestrictions(t *testing.T) {
	p := &storetest.Population{}

	usr1 := p.NewUser()
	usr1Key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	usr1Creds := rpcCreds(usr1Key)

	app1 := p.NewApplication(usr1.GetOrganizationOrUserIdentifiers())

	usr2 := p.NewUser()
	p.NewMembership(
		usr2.GetOrganizationOrUserIdentifiers(),
		app1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_APPLICATION_INFO,
	)
	// Set the user as administrative contact for the application.
	app1.AdministrativeContact = usr2.GetOrganizationOrUserIdentifiers()

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		regClt := ttnpb.NewApplicationRegistryClient(cc)
		accessClt := ttnpb.NewApplicationAccessClient(cc)

		// Attempt to delete a collaborator that is an administrative contact.
		_, err := accessClt.DeleteCollaborator(ctx, &ttnpb.DeleteApplicationCollaboratorRequest{
			ApplicationIds:  app1.Ids,
			CollaboratorIds: usr2.GetOrganizationOrUserIdentifiers(),
		}, usr1Creds)
		a.So(errors.IsFailedPrecondition(err), should.BeTrue)

		// Change the administrative contact.
		_, err = regClt.Update(ctx, &ttnpb.UpdateApplicationRequest{
			Application: &ttnpb.Application{
				Ids:                   app1.Ids,
				AdministrativeContact: usr1.GetOrganizationOrUserIdentifiers(),
			},
			FieldMask: ttnpb.FieldMask("administrative_contact"),
		}, usr1Creds)
		a.So(err, should.BeNil)

		// Attempt to delete a collaborator that is an administrative contact.
		_, err = accessClt.DeleteCollaborator(ctx, &ttnpb.DeleteApplicationCollaboratorRequest{
			ApplicationIds:  app1.Ids,
			CollaboratorIds: usr2.GetOrganizationOrUserIdentifiers(),
		}, usr1Creds)
		a.So(err, should.BeNil)
	}, withPrivateTestDatabase(p))
}
