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

	"github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc"
)

func TestApplicationsPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewApplicationRegistryClient(cc)

		_, err := reg.Create(ctx, &ttnpb.CreateApplicationRequest{
			Application: ttnpb.Application{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "foo-app"},
			},
			Collaborator: *ttnpb.UserIdentifiers{UserID: "foo-usr"}.OrganizationOrUserIdentifiers(),
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Get(ctx, &ttnpb.GetApplicationRequest{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "foo-app"},
			FieldMask:              types.FieldMask{Paths: []string{"name"}},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsUnauthenticated(err), should.BeTrue)
		}

		listRes, err := reg.List(ctx, &ttnpb.ListApplicationsRequest{
			FieldMask: types.FieldMask{Paths: []string{"name"}},
		})

		a.So(err, should.BeNil)
		a.So(listRes.Applications, should.BeEmpty)

		_, err = reg.List(ctx, &ttnpb.ListApplicationsRequest{
			Collaborator: ttnpb.UserIdentifiers{UserID: "foo-usr"}.OrganizationOrUserIdentifiers(),
			FieldMask:    types.FieldMask{Paths: []string{"name"}},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Update(ctx, &ttnpb.UpdateApplicationRequest{
			Application: ttnpb.Application{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "foo-app"},
				Name:                   "Updated Name",
			},
			FieldMask: types.FieldMask{Paths: []string{"name"}},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Delete(ctx, &ttnpb.ApplicationIdentifiers{ApplicationID: "foo-app"})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
	})
}

func TestApplicationsCRUD(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, creds := population.Users[0].UserIdentifiers, userCreds(0)

		reg := ttnpb.NewApplicationRegistryClient(cc)

		created, err := reg.Create(ctx, &ttnpb.CreateApplicationRequest{
			Application: ttnpb.Application{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "foo"},
				Name:                   "Foo Application",
			},
			Collaborator: *userID.OrganizationOrUserIdentifiers(),
		}, creds)

		a.So(err, should.BeNil)
		a.So(created.Name, should.Equal, "Foo Application")

		got, err := reg.Get(ctx, &ttnpb.GetApplicationRequest{
			ApplicationIdentifiers: created.ApplicationIdentifiers,
			FieldMask:              types.FieldMask{Paths: []string{"name"}},
		}, creds)

		a.So(err, should.BeNil)
		a.So(got.Name, should.Equal, created.Name)

		updated, err := reg.Update(ctx, &ttnpb.UpdateApplicationRequest{
			Application: ttnpb.Application{
				ApplicationIdentifiers: created.ApplicationIdentifiers,
				Name:                   "Updated Name",
			},
			FieldMask: types.FieldMask{Paths: []string{"name"}},
		}, creds)

		a.So(err, should.BeNil)
		a.So(updated.Name, should.Equal, "Updated Name")

		for _, collaborator := range []*ttnpb.OrganizationOrUserIdentifiers{nil, userID.OrganizationOrUserIdentifiers()} {
			list, err := reg.List(ctx, &ttnpb.ListApplicationsRequest{
				FieldMask:    types.FieldMask{Paths: []string{"name"}},
				Collaborator: collaborator,
			}, creds)
			a.So(err, should.BeNil)
			if a.So(list.Applications, should.NotBeEmpty) {
				var found bool
				for _, item := range list.Applications {
					if item.ApplicationIdentifiers == created.ApplicationIdentifiers {
						found = true
						a.So(item.Name, should.Equal, updated.Name)
					}
				}
				a.So(found, should.BeTrue)
			}
		}

		_, err = reg.Delete(ctx, &created.ApplicationIdentifiers, creds)
		a.So(err, should.BeNil)

	})
}
