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
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestApplicationsPermissions(t *testing.T) {
	p := &storetest.Population{}
	usr1 := p.NewUser()
	app1 := p.NewApplication(usr1.GetOrganizationOrUserIdentifiers())

	readAdminUsr := p.NewUser()
	readAdminUsr.Admin = true
	readAdminUsrKey, _ := p.NewAPIKey(readAdminUsr.GetEntityIdentifiers(), ttnpb.AllReadAdminRights.GetRights()...)
	readAdminUsrCreds := rpcCreds(readAdminUsrKey)

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(_ *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewApplicationRegistryClient(cc)

		t.Run("Permission denied", func(t *testing.T) { // nolint:paralleltest
			reg := ttnpb.NewApplicationRegistryClient(cc)

			_, err := reg.Create(ctx, &ttnpb.CreateApplicationRequest{
				Application: &ttnpb.Application{
					Ids: &ttnpb.ApplicationIdentifiers{ApplicationId: "foo-app"},
				},
				Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
			})
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsPermissionDenied(err), should.BeTrue)
			}

			_, err = reg.Get(ctx, &ttnpb.GetApplicationRequest{
				ApplicationIds: app1.GetIds(),
				FieldMask:      ttnpb.FieldMask("name"),
			})
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsUnauthenticated(err), should.BeTrue)
			}

			listRes, err := reg.List(ctx, &ttnpb.ListApplicationsRequest{
				FieldMask: ttnpb.FieldMask("name"),
			})
			a.So(err, should.BeNil)
			if a.So(listRes, should.NotBeNil) {
				a.So(listRes.Applications, should.BeEmpty)
			}

			_, err = reg.List(ctx, &ttnpb.ListApplicationsRequest{
				Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
				FieldMask:    ttnpb.FieldMask("name"),
			})
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsPermissionDenied(err), should.BeTrue)
			}

			_, err = reg.Update(ctx, &ttnpb.UpdateApplicationRequest{
				Application: &ttnpb.Application{
					Ids:  app1.GetIds(),
					Name: "Updated Name",
				},
				FieldMask: ttnpb.FieldMask("name"),
			})
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsPermissionDenied(err), should.BeTrue)
			}

			_, err = reg.Delete(ctx, app1.GetIds())
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsPermissionDenied(err), should.BeTrue)
			}
		})

		t.Run("Admin read-only", func(t *testing.T) { // nolint:paralleltest
			_, err := reg.Create(ctx, &ttnpb.CreateApplicationRequest{
				Application: &ttnpb.Application{
					Ids: &ttnpb.ApplicationIdentifiers{ApplicationId: "foo-app"},
				},
				Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
			}, readAdminUsrCreds)
			a.So(errors.IsPermissionDenied(err), should.BeTrue)

			_, err = reg.Get(ctx, &ttnpb.GetApplicationRequest{
				ApplicationIds: app1.GetIds(),
				FieldMask:      ttnpb.FieldMask("name"),
			}, readAdminUsrCreds)
			a.So(errors.IsPermissionDenied(err), should.BeFalse)

			_, err = reg.List(ctx, &ttnpb.ListApplicationsRequest{
				FieldMask: ttnpb.FieldMask("name"),
			}, readAdminUsrCreds)
			a.So(errors.IsPermissionDenied(err), should.BeFalse)

			_, err = reg.List(ctx, &ttnpb.ListApplicationsRequest{
				Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
				FieldMask:    ttnpb.FieldMask("name"),
			}, readAdminUsrCreds)
			a.So(errors.IsPermissionDenied(err), should.BeFalse)

			_, err = reg.Update(ctx, &ttnpb.UpdateApplicationRequest{
				Application: &ttnpb.Application{
					Ids:  app1.GetIds(),
					Name: "Updated Name",
				},
				FieldMask: ttnpb.FieldMask("name"),
			}, readAdminUsrCreds)
			a.So(errors.IsPermissionDenied(err), should.BeTrue)

			_, err = reg.Delete(ctx, app1.GetIds(), readAdminUsrCreds)
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		})
	}, withPrivateTestDatabase(p))
}

func TestApplicationsCRUD(t *testing.T) {
	p := &storetest.Population{}

	adminUsr := p.NewUser()
	adminUsr.Admin = true
	adminKey, _ := p.NewAPIKey(adminUsr.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	adminCreds := rpcCreds(adminKey)

	usr1 := p.NewUser()
	for i := 0; i < 5; i++ {
		p.NewApplication(usr1.GetOrganizationOrUserIdentifiers())
	}

	usr2 := p.NewUser()
	for i := 0; i < 5; i++ {
		p.NewApplication(usr2.GetOrganizationOrUserIdentifiers())
	}

	key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	creds := rpcCreds(key)
	keyWithoutRights, _ := p.NewAPIKey(usr1.GetEntityIdentifiers())
	credsWithoutRights := rpcCreds(keyWithoutRights)

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewApplicationRegistryClient(cc)

		// Test batch fetch with cluster authorization
		list, err := reg.List(ctx, &ttnpb.ListApplicationsRequest{
			FieldMask:    ttnpb.FieldMask("ids"),
			Collaborator: nil,
			Deleted:      true,
		}, is.WithClusterAuth())
		if a.So(err, should.BeNil) {
			a.So(list.Applications, should.HaveLength, 10)
		}

		is.config.UserRights.CreateApplications = false

		_, err = reg.Create(ctx, &ttnpb.CreateApplicationRequest{
			Application: &ttnpb.Application{
				Ids:  &ttnpb.ApplicationIdentifiers{ApplicationId: "foo"},
				Name: "Foo Application",
			},
			Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
		}, creds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		is.config.UserRights.CreateApplications = true

		created, err := reg.Create(ctx, &ttnpb.CreateApplicationRequest{
			Application: &ttnpb.Application{
				Ids:  &ttnpb.ApplicationIdentifiers{ApplicationId: "foo"},
				Name: "Foo Application",
			},
			Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
		}, creds)
		if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) {
			a.So(created.Name, should.Equal, "Foo Application")
		}

		got, err := reg.Get(ctx, &ttnpb.GetApplicationRequest{
			ApplicationIds: created.GetIds(),
			FieldMask:      ttnpb.FieldMask("name"),
		}, creds)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got.Name, should.Equal, created.Name)
		}

		got, err = reg.Get(ctx, &ttnpb.GetApplicationRequest{
			ApplicationIds: created.GetIds(),
			FieldMask:      ttnpb.FieldMask("ids"),
		}, credsWithoutRights)
		a.So(err, should.BeNil)

		got, err = reg.Get(ctx, &ttnpb.GetApplicationRequest{
			ApplicationIds: created.GetIds(),
			FieldMask:      ttnpb.FieldMask("attributes"),
		}, credsWithoutRights)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		// TODO: Remove (https://github.com/TheThingsNetwork/lorawan-stack/issues/6804)
		t.Run("Contact_info fieldmask", func(t *testing.T) { // nolint:paralleltest
			a, ctx := test.New(t)
			got, err := reg.Get(ctx, &ttnpb.GetApplicationRequest{
				ApplicationIds: created.GetIds(),
				FieldMask:      ttnpb.FieldMask("contact_info"),
			}, creds)
			if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
				a.So(got.ContactInfo, should.HaveLength, 2)
				a.So(got.ContactInfo[0].Value, should.Equal, usr1.PrimaryEmailAddress)
				a.So(got.ContactInfo[0].ContactType, should.Equal, ttnpb.ContactType_CONTACT_TYPE_OTHER)
				a.So(got.ContactInfo[1].Value, should.Equal, usr1.PrimaryEmailAddress)
				a.So(got.ContactInfo[1].ContactType, should.Equal, ttnpb.ContactType_CONTACT_TYPE_TECHNICAL)
			}

			// Testing the `PublicSafe` method, which should not return the contact_info's email address when the caller
			// does not have the appropriate rights.
			got, err = reg.Get(ctx, &ttnpb.GetApplicationRequest{
				ApplicationIds: created.GetIds(),
				FieldMask:      ttnpb.FieldMask("contact_info"),
			}, credsWithoutRights)
			if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
				a.So(got.ContactInfo, should.HaveLength, 0)
			}
		})

		updated, err := reg.Update(ctx, &ttnpb.UpdateApplicationRequest{
			Application: &ttnpb.Application{
				Ids:  created.GetIds(),
				Name: "Updated Name",
			},
			FieldMask: ttnpb.FieldMask("name"),
		}, creds)
		if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) {
			a.So(updated.Name, should.Equal, "Updated Name")
		}

		t.Run("Contact Info Restrictions", func(t *testing.T) { // nolint:paralleltest
			a, ctx := test.New(t)

			oldSetOtherAsContacts := is.config.CollaboratorRights.SetOthersAsContacts
			t.Cleanup(func() { is.config.CollaboratorRights.SetOthersAsContacts = oldSetOtherAsContacts })
			is.config.CollaboratorRights.SetOthersAsContacts = false

			// Set usr-2 as collaborator to application.
			aac := ttnpb.NewApplicationAccessClient(cc)
			_, err := aac.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
				ApplicationIds: created.GetIds(),
				Collaborator: &ttnpb.Collaborator{
					Ids:    usr2.GetOrganizationOrUserIdentifiers(),
					Rights: []ttnpb.Right{ttnpb.Right_RIGHT_ALL},
				},
			}, creds)
			a.So(err, should.BeNil)

			// Attempt to set another collaborator as administrative contact.
			_, err = reg.Update(ctx, &ttnpb.UpdateApplicationRequest{
				Application: &ttnpb.Application{
					Ids:                   created.GetIds(),
					AdministrativeContact: usr2.GetOrganizationOrUserIdentifiers(),
				},
				FieldMask: ttnpb.FieldMask("administrative_contact"),
			}, creds)
			a.So(errors.IsPermissionDenied(err), should.BeTrue)

			// Admin can bypass contact info restrictions.
			_, err = reg.Update(ctx, &ttnpb.UpdateApplicationRequest{
				Application: &ttnpb.Application{
					Ids:                   created.GetIds(),
					AdministrativeContact: usr1.GetOrganizationOrUserIdentifiers(),
				},
				FieldMask: ttnpb.FieldMask("administrative_contact"),
			}, adminCreds)
			a.So(err, should.BeNil)

			is.config.CollaboratorRights.SetOthersAsContacts = true

			// Now usr-1 can set usr-2 as technical contact.
			_, err = reg.Update(ctx, &ttnpb.UpdateApplicationRequest{
				Application: &ttnpb.Application{
					Ids:              created.GetIds(),
					Name:             "Updated Name",
					TechnicalContact: usr2.GetOrganizationOrUserIdentifiers(),
				},
				FieldMask: ttnpb.FieldMask("technical_contact"),
			}, creds)
			a.So(err, should.BeNil)
		})

		for _, collaborator := range []*ttnpb.OrganizationOrUserIdentifiers{
			nil, usr1.GetOrganizationOrUserIdentifiers(),
		} {
			list, err := reg.List(ctx, &ttnpb.ListApplicationsRequest{
				FieldMask:    ttnpb.FieldMask("name"),
				Collaborator: collaborator,
			}, creds)
			if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) && a.So(list.Applications, should.HaveLength, 6) {
				var found bool
				for _, item := range list.Applications {
					if item.GetIds().GetApplicationId() == created.GetIds().GetApplicationId() {
						found = true
						a.So(item.Name, should.Equal, updated.Name)
					}
				}
				a.So(found, should.BeTrue)
			}
		}

		_, err = reg.Delete(ctx, created.GetIds(), creds)
		a.So(err, should.BeNil)

		_, err = reg.Purge(ctx, created.GetIds(), creds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Purge(ctx, created.GetIds(), adminCreds)
		a.So(err, should.BeNil)
	}, withPrivateTestDatabase(p))
}

func TestApplicationsPagination(t *testing.T) {
	p := &storetest.Population{}

	usr1 := p.NewUser()
	for i := 0; i < 3; i++ {
		p.NewApplication(usr1.GetOrganizationOrUserIdentifiers())
	}

	key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	creds := rpcCreds(key)

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewApplicationRegistryClient(cc)

		var md metadata.MD

		list, err := reg.List(ctx, &ttnpb.ListApplicationsRequest{
			FieldMask:    ttnpb.FieldMask("name"),
			Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
			Limit:        2,
			Page:         1,
		}, creds, grpc.Header(&md))
		if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) {
			a.So(list.Applications, should.HaveLength, 2)
			a.So(md.Get("x-total-count"), should.Resemble, []string{"3"})
		}

		list, err = reg.List(ctx, &ttnpb.ListApplicationsRequest{
			FieldMask:    ttnpb.FieldMask("name"),
			Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
			Limit:        2,
			Page:         2,
		}, creds)
		if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) {
			a.So(list.Applications, should.HaveLength, 1)
		}

		list, err = reg.List(ctx, &ttnpb.ListApplicationsRequest{
			FieldMask:    ttnpb.FieldMask("name"),
			Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
			Limit:        2,
			Page:         3,
		}, creds)
		if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) {
			a.So(list.Applications, should.BeEmpty)
		}
	}, withPrivateTestDatabase(p))
}
