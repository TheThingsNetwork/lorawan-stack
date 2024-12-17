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

func TestClientCollaboratorsPermissions(t *testing.T) { // nolint:gocyclo
	p := &storetest.Population{}

	usr1 := p.NewUser()

	readAdmin := p.NewUser()
	readAdmin.Admin = true
	readAdminKey, _ := p.NewAPIKey(readAdmin.GetEntityIdentifiers(), ttnpb.AllReadAdminRights.GetRights()...)
	readAdminCreds := rpcCreds(readAdminKey)

	cli1 := p.NewClient(usr1.GetOrganizationOrUserIdentifiers())
	cliKey, _ := p.NewAPIKey(cli1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_GATEWAY_INFO,
		ttnpb.Right_RIGHT_GATEWAY_LINK,
	)
	cliCreds := rpcCreds(cliKey)

	p.NewMembership(
		usr1.GetOrganizationOrUserIdentifiers(),
		cli1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_CLIENT_SETTINGS_COLLABORATORS,
	)

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		is.config.AdminRights.All = true

		reg := ttnpb.NewClientAccessClient(cc)

		t.Run("Invalid credentials", func(t *testing.T) { // nolint:paralleltest
			for _, opts := range [][]grpc.CallOption{nil, {cliCreds}} {
				_, err := reg.GetCollaborator(ctx, &ttnpb.GetClientCollaboratorRequest{
					ClientIds:    cli1.GetIds(),
					Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
				})
				a.So(errors.IsPermissionDenied(err), should.BeTrue)

				_, err = reg.ListCollaborators(ctx, &ttnpb.ListClientCollaboratorsRequest{
					ClientIds: cli1.GetIds(),
				})
				a.So(errors.IsUnauthenticated(err), should.BeTrue)

				_, err = reg.SetCollaborator(ctx, &ttnpb.SetClientCollaboratorRequest{
					ClientIds: cli1.GetIds(),
					Collaborator: &ttnpb.Collaborator{
						Ids:    usr1.GetOrganizationOrUserIdentifiers(),
						Rights: []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_INFO},
					},
				}, opts...)
				a.So(errors.IsPermissionDenied(err), should.BeTrue)
			}
		})

		t.Run("Admin read-only", func(t *testing.T) { // nolint:paralleltest
			_, err := reg.GetCollaborator(ctx, &ttnpb.GetClientCollaboratorRequest{
				ClientIds:    cli1.GetIds(),
				Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
			}, readAdminCreds)
			a.So(errors.IsPermissionDenied(err), should.BeTrue)

			// ListCollaborators without credentials.
			_, err = reg.ListCollaborators(ctx, &ttnpb.ListClientCollaboratorsRequest{
				ClientIds: cli1.GetIds(),
			}, readAdminCreds)
			a.So(errors.IsPermissionDenied(err), should.BeFalse)

			_, err = reg.SetCollaborator(ctx, &ttnpb.SetClientCollaboratorRequest{
				ClientIds: cli1.GetIds(),
				Collaborator: &ttnpb.Collaborator{
					Ids:    usr1.GetOrganizationOrUserIdentifiers(),
					Rights: []ttnpb.Right{ttnpb.Right_RIGHT_GATEWAY_INFO},
				},
			}, readAdminCreds)
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsPermissionDenied(err), should.BeTrue)
			}
		})

		t.Run("cluster auth", func(t *testing.T) { // nolint:paralleltest
			rights, err := reg.ListRights(ctx, cli1.GetIds(), is.WithClusterAuth())
			if a.So(err, should.BeNil) && a.So(rights, should.NotBeNil) {
				a.So(ttnpb.AllClusterRights.Intersect(ttnpb.AllClientRights).Sub(rights).Rights, should.BeEmpty)
			}
		})
	}, withPrivateTestDatabase(p))
}

func TestClientCollaborators(t *testing.T) { // nolint:gocyclo
	p := &storetest.Population{}

	admin := p.NewUser()
	admin.Admin = true
	adminKey, _ := p.NewAPIKey(admin.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	adminCreds := rpcCreds(adminKey)

	usr1 := p.NewUser()
	usr1Key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_CLIENT_SETTINGS_COLLABORATORS)
	usr1Creds := rpcCreds(usr1Key)

	cli1 := p.NewClient(usr1.GetOrganizationOrUserIdentifiers())

	usr2 := p.NewUser()
	p.NewMembership(
		usr2.GetOrganizationOrUserIdentifiers(),
		cli1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_CLIENT_SETTINGS_COLLABORATORS,
	)

	usr3 := p.NewUser()

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		is.config.AdminRights.All = true

		reg := ttnpb.NewClientAccessClient(cc)

		// GetCollaborator that doesn't exist.
		got, err := reg.GetCollaborator(ctx, &ttnpb.GetClientCollaboratorRequest{
			ClientIds:    cli1.GetIds(),
			Collaborator: usr3.GetOrganizationOrUserIdentifiers(),
		}, adminCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		a.So(got, should.BeNil)

		// Collaborator CRUD with different valid credentials.
		for _, opts := range [][]grpc.CallOption{{adminCreds}, {usr1Creds}} {
			// Adds `usr3` as a collaborator of `cli1` client.
			_, err = reg.SetCollaborator(ctx, &ttnpb.SetClientCollaboratorRequest{
				ClientIds: cli1.GetIds(),
				Collaborator: &ttnpb.Collaborator{
					Ids:    usr3.GetOrganizationOrUserIdentifiers(),
					Rights: []ttnpb.Right{ttnpb.Right_RIGHT_CLIENT_SETTINGS_COLLABORATORS},
				},
			}, opts...)
			a.So(err, should.BeNil)

			// Lists collaborators of the `cli1` client.
			list, err := reg.ListCollaborators(ctx, &ttnpb.ListClientCollaboratorsRequest{
				ClientIds: cli1.GetIds(),
			}, opts...)
			if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) && a.So(list.Collaborators, should.HaveLength, 3) {
				for _, k := range list.Collaborators {
					if unique.ID(ctx, k.GetIds()) == unique.ID(ctx, usr3.GetIds()) {
						a.So(k.Rights, should.Resemble, []ttnpb.Right{ttnpb.Right_RIGHT_CLIENT_SETTINGS_COLLABORATORS})
					}
				}
			}

			// Gets `usr3` collaborator of the `cli1` client.
			got, err = reg.GetCollaborator(ctx, &ttnpb.GetClientCollaboratorRequest{
				ClientIds:    cli1.GetIds(),
				Collaborator: usr3.GetOrganizationOrUserIdentifiers(),
			}, opts...)
			if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
				a.So(got.Rights, should.Resemble, []ttnpb.Right{
					ttnpb.Right_RIGHT_CLIENT_SETTINGS_COLLABORATORS,
				})
			}

			// TODO: Remove SetCollaborator test case (https://github.com/TheThingsNetwork/lorawan-stack/issues/6488).
			t.Run("Delete via set method", func(*testing.T) { // nolint:paralleltest
				_, err = reg.SetCollaborator(ctx, &ttnpb.SetClientCollaboratorRequest{
					ClientIds: cli1.GetIds(),
					Collaborator: &ttnpb.Collaborator{
						Ids:    usr3.GetOrganizationOrUserIdentifiers(),
						Rights: []ttnpb.Right{},
					},
				}, opts...)
				a.So(err, should.BeNil)

				// Verifies that it has been deleted.
				got, err = reg.GetCollaborator(ctx, &ttnpb.GetClientCollaboratorRequest{
					ClientIds:    cli1.GetIds(),
					Collaborator: usr3.GetOrganizationOrUserIdentifiers(),
				}, opts...)
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsNotFound(err), should.BeTrue)
				}
				a.So(got, should.BeNil)
			})

			// Recreates `usr3` collaborator of the `cli1` client.
			_, err = reg.SetCollaborator(ctx, &ttnpb.SetClientCollaboratorRequest{
				ClientIds: cli1.GetIds(),
				Collaborator: &ttnpb.Collaborator{
					Ids:    usr3.GetOrganizationOrUserIdentifiers(),
					Rights: []ttnpb.Right{ttnpb.Right_RIGHT_CLIENT_SETTINGS_COLLABORATORS},
				},
			}, opts...)
			a.So(err, should.BeNil)

			t.Run("Delete via delete method", func(*testing.T) { // nolint:paralleltest
				empty, err := reg.DeleteCollaborator(ctx, &ttnpb.DeleteClientCollaboratorRequest{
					ClientIds:       cli1.GetIds(),
					CollaboratorIds: usr3.GetOrganizationOrUserIdentifiers(),
				}, opts...)
				a.So(err, should.BeNil)
				a.So(empty, should.Resemble, ttnpb.Empty)

				// Verifies that it has been deleted.
				got, err = reg.GetCollaborator(ctx, &ttnpb.GetClientCollaboratorRequest{
					ClientIds:    cli1.GetIds(),
					Collaborator: usr3.GetOrganizationOrUserIdentifiers(),
				}, opts...)
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsNotFound(err), should.BeTrue)
				}
				a.So(got, should.BeNil)
			})
		}

		// Remove the other collaborator with _ALL rights.
		_, err = reg.DeleteCollaborator(ctx, &ttnpb.DeleteClientCollaboratorRequest{
			ClientIds:       cli1.GetIds(),
			CollaboratorIds: usr2.GetOrganizationOrUserIdentifiers(),
		}, usr1Creds)
		a.So(err, should.BeNil)

		// Try removing the only collaborator with _ALL rights.
		_, err = reg.DeleteCollaborator(ctx, &ttnpb.DeleteClientCollaboratorRequest{
			ClientIds:       cli1.GetIds(),
			CollaboratorIds: usr1.GetOrganizationOrUserIdentifiers(),
		}, usr1Creds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsFailedPrecondition(err), should.BeTrue)
		}
	}, withPrivateTestDatabase(p))
}

func TestClientContactRestrictions(t *testing.T) {
	p := &storetest.Population{}

	usr1 := p.NewUser()
	usr1Key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	usr1Creds := rpcCreds(usr1Key)

	gtw1 := p.NewClient(usr1.GetOrganizationOrUserIdentifiers())

	usr2 := p.NewUser()
	p.NewMembership(
		usr2.GetOrganizationOrUserIdentifiers(),
		gtw1.GetEntityIdentifiers(),
		ttnpb.Right_RIGHT_CLIENT_INFO,
	)
	// Set the user as administrative contact for the client.
	gtw1.AdministrativeContact = usr2.GetOrganizationOrUserIdentifiers()

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		regClt := ttnpb.NewClientRegistryClient(cc)
		accessClt := ttnpb.NewClientAccessClient(cc)

		// Attempt to delete a collaborator that is an administrative contact.
		_, err := accessClt.DeleteCollaborator(ctx, &ttnpb.DeleteClientCollaboratorRequest{
			ClientIds:       gtw1.Ids,
			CollaboratorIds: usr2.GetOrganizationOrUserIdentifiers(),
		}, usr1Creds)
		a.So(errors.IsFailedPrecondition(err), should.BeTrue)

		// Change the administrative contact.
		_, err = regClt.Update(ctx, &ttnpb.UpdateClientRequest{
			Client: &ttnpb.Client{
				Ids:                   gtw1.Ids,
				AdministrativeContact: usr1.GetOrganizationOrUserIdentifiers(),
			},
			FieldMask: ttnpb.FieldMask("administrative_contact"),
		}, usr1Creds)
		a.So(err, should.BeNil)

		// Attempt to delete a collaborator that is an administrative contact.
		_, err = accessClt.DeleteCollaborator(ctx, &ttnpb.DeleteClientCollaboratorRequest{
			ClientIds:       gtw1.Ids,
			CollaboratorIds: usr2.GetOrganizationOrUserIdentifiers(),
		}, usr1Creds)
		a.So(err, should.BeNil)
	}, withPrivateTestDatabase(p))
}
