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

	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/storetest"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestOrganizationsNestedError(t *testing.T) {
	t.Parallel()

	p := &storetest.Population{}

	usr := p.NewUser()
	usrKey, _ := p.NewAPIKey(usr.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	usrCreds := rpcCreds(usrKey)

	org := p.NewOrganization(usr.GetOrganizationOrUserIdentifiers())

	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewOrganizationRegistryClient(cc)

		_, err := reg.Create(ctx, &ttnpb.CreateOrganizationRequest{
			Organization: &ttnpb.Organization{
				Ids: &ttnpb.OrganizationIdentifiers{OrganizationId: "foo-org"},
			},
			Collaborator: org.GetOrganizationOrUserIdentifiers(),
		}, usrCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsInvalidArgument(err), should.BeTrue)
		}

		_, err = reg.List(ctx, &ttnpb.ListOrganizationsRequest{
			FieldMask:    ttnpb.FieldMask("name"),
			Collaborator: org.GetOrganizationOrUserIdentifiers(),
		}, usrCreds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsInvalidArgument(err), should.BeTrue)
		}
	}, withPrivateTestDatabase(p))
}

func TestOrganizationsPermissionDenied(t *testing.T) {
	p := &storetest.Population{}
	usr1 := p.NewUser()
	org1 := p.NewOrganization(usr1.GetOrganizationOrUserIdentifiers())

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(_ *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewOrganizationRegistryClient(cc)

		_, err := reg.Create(ctx, &ttnpb.CreateOrganizationRequest{
			Organization: &ttnpb.Organization{
				Ids: &ttnpb.OrganizationIdentifiers{OrganizationId: "foo-org"},
			},
			Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
		})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Get(ctx, &ttnpb.GetOrganizationRequest{
			OrganizationIds: org1.GetIds(),
			FieldMask:       ttnpb.FieldMask("name"),
		})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsUnauthenticated(err), should.BeTrue)
		}

		listRes, err := reg.List(ctx, &ttnpb.ListOrganizationsRequest{
			FieldMask: ttnpb.FieldMask("name"),
		})
		a.So(err, should.BeNil)
		if a.So(listRes, should.NotBeNil) {
			a.So(listRes.Organizations, should.BeEmpty)
		}

		_, err = reg.List(ctx, &ttnpb.ListOrganizationsRequest{
			Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
			FieldMask:    ttnpb.FieldMask("name"),
		})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Update(ctx, &ttnpb.UpdateOrganizationRequest{
			Organization: &ttnpb.Organization{
				Ids:  org1.GetIds(),
				Name: "Updated Name",
			},
			FieldMask: ttnpb.FieldMask("name"),
		})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Delete(ctx, org1.GetIds())
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
	})
}

func TestOrganizationsCRUD(t *testing.T) {
	p := &storetest.Population{}

	adminUsr := p.NewUser()
	adminUsr.Admin = true
	adminKey, _ := p.NewAPIKey(adminUsr.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	adminCreds := rpcCreds(adminKey)

	usr1 := p.NewUser()
	for i := 0; i < 5; i++ {
		p.NewOrganization(usr1.GetOrganizationOrUserIdentifiers())
	}

	usr2 := p.NewUser()
	for i := 0; i < 5; i++ {
		p.NewOrganization(usr2.GetOrganizationOrUserIdentifiers())
	}

	key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	creds := rpcCreds(key)
	keyWithoutRights, _ := p.NewAPIKey(usr1.GetEntityIdentifiers())
	credsWithoutRights := rpcCreds(keyWithoutRights)

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewOrganizationRegistryClient(cc)

		is.config.UserRights.CreateOrganizations = false

		_, err := reg.Create(ctx, &ttnpb.CreateOrganizationRequest{
			Organization: &ttnpb.Organization{
				Ids:  &ttnpb.OrganizationIdentifiers{OrganizationId: "foo"},
				Name: "Foo Organization",
			},
			Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
		}, creds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		is.config.UserRights.CreateOrganizations = true

		created, err := reg.Create(ctx, &ttnpb.CreateOrganizationRequest{
			Organization: &ttnpb.Organization{
				Ids:  &ttnpb.OrganizationIdentifiers{OrganizationId: "foo"},
				Name: "Foo Organization",
			},
			Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
		}, creds)
		if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) {
			a.So(created.Name, should.Equal, "Foo Organization")
		}

		got, err := reg.Get(ctx, &ttnpb.GetOrganizationRequest{
			OrganizationIds: created.GetIds(),
			FieldMask:       ttnpb.FieldMask("name"),
		}, creds)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got.Name, should.Equal, created.Name)
		}

		got, err = reg.Get(ctx, &ttnpb.GetOrganizationRequest{
			OrganizationIds: created.GetIds(),
			FieldMask:       ttnpb.FieldMask("ids"),
		}, credsWithoutRights)
		a.So(err, should.BeNil)

		got, err = reg.Get(ctx, &ttnpb.GetOrganizationRequest{
			OrganizationIds: created.GetIds(),
			FieldMask:       ttnpb.FieldMask("attributes"),
		}, credsWithoutRights)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		updated, err := reg.Update(ctx, &ttnpb.UpdateOrganizationRequest{
			Organization: &ttnpb.Organization{
				Ids:  created.GetIds(),
				Name: "Updated Name",
			},
			FieldMask: ttnpb.FieldMask("name"),
		}, creds)
		if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) {
			a.So(updated.Name, should.Equal, "Updated Name")
		}

		for _, collaborator := range []*ttnpb.OrganizationOrUserIdentifiers{nil, usr1.GetOrganizationOrUserIdentifiers()} {
			list, err := reg.List(ctx, &ttnpb.ListOrganizationsRequest{
				FieldMask:    ttnpb.FieldMask("name"),
				Collaborator: collaborator,
			}, creds)
			if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) && a.So(list.Organizations, should.HaveLength, 6) {
				var found bool
				for _, item := range list.Organizations {
					if item.GetIds().GetOrganizationId() == created.GetIds().GetOrganizationId() {
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

func TestOrganizationsPagination(t *testing.T) {
	p := &storetest.Population{}

	usr1 := p.NewUser()
	for i := 0; i < 3; i++ {
		p.NewOrganization(usr1.GetOrganizationOrUserIdentifiers())
	}

	key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	creds := rpcCreds(key)

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewOrganizationRegistryClient(cc)

		var md metadata.MD

		list, err := reg.List(ctx, &ttnpb.ListOrganizationsRequest{
			FieldMask:    ttnpb.FieldMask("name"),
			Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
			Limit:        2,
			Page:         1,
		}, creds, grpc.Header(&md))
		if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) {
			a.So(list.Organizations, should.HaveLength, 2)
			a.So(md.Get("x-total-count"), should.Resemble, []string{"3"})
		}

		list, err = reg.List(ctx, &ttnpb.ListOrganizationsRequest{
			FieldMask:    ttnpb.FieldMask("name"),
			Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
			Limit:        2,
			Page:         2,
		}, creds)
		if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) {
			a.So(list.Organizations, should.HaveLength, 1)
		}

		list, err = reg.List(ctx, &ttnpb.ListOrganizationsRequest{
			FieldMask:    ttnpb.FieldMask("name"),
			Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
			Limit:        2,
			Page:         3,
		}, creds)
		if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) {
			a.So(list.Organizations, should.BeEmpty)
		}
	}, withPrivateTestDatabase(p))
}
