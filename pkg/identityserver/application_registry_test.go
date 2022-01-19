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
	"testing"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"google.golang.org/grpc"
)

func init() {
	// remove applications assigned to the user by the populator
	userID := paginationUser.GetIds()
	for _, app := range population.Applications {
		for id, collaborators := range population.Memberships {
			if app.IDString() == id.IDString() {
				for i, collaborator := range collaborators {
					if collaborator.IDString() == userID.GetUserId() {
						collaborators = collaborators[:i+copy(collaborators[i:], collaborators[i+1:])]
					}
				}
			}
		}
	}

	// add deterministic number of applications
	for i := 0; i < 3; i++ {
		applicationID := population.Applications[i].GetEntityIdentifiers()
		population.Memberships[applicationID] = append(population.Memberships[applicationID], &ttnpb.Collaborator{
			Ids:    paginationUser.OrganizationOrUserIdentifiers(),
			Rights: []ttnpb.Right{ttnpb.Right_RIGHT_APPLICATION_ALL},
		})
	}
}

func TestApplicationsPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewApplicationRegistryClient(cc)

		_, err := reg.Create(ctx, &ttnpb.CreateApplicationRequest{
			Application: &ttnpb.Application{
				Ids: &ttnpb.ApplicationIdentifiers{ApplicationId: "foo-app"},
			},
			Collaborator: ttnpb.UserIdentifiers{UserId: "foo-usr"}.OrganizationOrUserIdentifiers(),
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Get(ctx, &ttnpb.GetApplicationRequest{
			ApplicationIds: &ttnpb.ApplicationIdentifiers{ApplicationId: "foo-app"},
			FieldMask:      &pbtypes.FieldMask{Paths: []string{"name"}},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsUnauthenticated(err), should.BeTrue)
		}

		listRes, err := reg.List(ctx, &ttnpb.ListApplicationsRequest{
			FieldMask: &pbtypes.FieldMask{Paths: []string{"name"}},
		})

		a.So(err, should.BeNil)
		if a.So(listRes, should.NotBeNil) {
			a.So(listRes.Applications, should.BeEmpty)
		}

		_, err = reg.List(ctx, &ttnpb.ListApplicationsRequest{
			Collaborator: ttnpb.UserIdentifiers{UserId: "foo-usr"}.OrganizationOrUserIdentifiers(),
			FieldMask:    &pbtypes.FieldMask{Paths: []string{"name"}},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Update(ctx, &ttnpb.UpdateApplicationRequest{
			Application: &ttnpb.Application{
				Ids:  &ttnpb.ApplicationIdentifiers{ApplicationId: "foo-app"},
				Name: "Updated Name",
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"name"}},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Delete(ctx, &ttnpb.ApplicationIdentifiers{ApplicationId: "foo-app"})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
	})
}

func TestApplicationsCRUD(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewApplicationRegistryClient(cc)

		userID, creds := population.Users[defaultUserIdx].GetIds(), userCreds(defaultUserIdx)
		credsWithoutRights := userCreds(defaultUserIdx, "key without rights")

		is.config.UserRights.CreateApplications = false
		// Test batch fetch with cluster authorization
		list, err := reg.List(ctx, &ttnpb.ListApplicationsRequest{
			FieldMask:    &pbtypes.FieldMask{Paths: []string{"ids"}},
			Collaborator: nil,
			Deleted:      true,
		}, is.WithClusterAuth())

		a.So(err, should.BeNil)
		a.So(list.Applications, should.HaveLength, 16)

		_, err = reg.Create(ctx, &ttnpb.CreateApplicationRequest{
			Application: &ttnpb.Application{
				Ids:  &ttnpb.ApplicationIdentifiers{ApplicationId: "foo"},
				Name: "Foo Application",
			},
			Collaborator: userID.GetOrganizationOrUserIdentifiers(),
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
			Collaborator: userID.OrganizationOrUserIdentifiers(),
		}, creds)

		a.So(err, should.BeNil)
		if a.So(created, should.NotBeNil) {
			a.So(created.Name, should.Equal, "Foo Application")
		}

		got, err := reg.Get(ctx, &ttnpb.GetApplicationRequest{
			ApplicationIds: created.GetIds(),
			FieldMask:      &pbtypes.FieldMask{Paths: []string{"name"}},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.Name, should.Equal, created.Name)
		}

		got, err = reg.Get(ctx, &ttnpb.GetApplicationRequest{
			ApplicationIds: created.GetIds(),
			FieldMask:      &pbtypes.FieldMask{Paths: []string{"ids"}},
		}, credsWithoutRights)

		a.So(err, should.BeNil)

		got, err = reg.Get(ctx, &ttnpb.GetApplicationRequest{
			ApplicationIds: created.GetIds(),
			FieldMask:      &pbtypes.FieldMask{Paths: []string{"attributes"}},
		}, credsWithoutRights)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		updated, err := reg.Update(ctx, &ttnpb.UpdateApplicationRequest{
			Application: &ttnpb.Application{
				Ids:  created.GetIds(),
				Name: "Updated Name",
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"name"}},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(updated, should.NotBeNil) {
			a.So(updated.Name, should.Equal, "Updated Name")
		}

		for _, collaborator := range []*ttnpb.OrganizationOrUserIdentifiers{nil, userID.OrganizationOrUserIdentifiers()} {
			list, err := reg.List(ctx, &ttnpb.ListApplicationsRequest{
				FieldMask:    &pbtypes.FieldMask{Paths: []string{"name"}},
				Collaborator: collaborator,
			}, creds)

			a.So(err, should.BeNil)
			if a.So(list, should.NotBeNil) && a.So(list.Applications, should.NotBeEmpty) {
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

		// Check that returned value is not nil
		devEUIResponse, err := reg.IssueDevEUI(ctx, created.GetIds(), userCreds(adminUserIdx))
		a.So(err, should.BeNil)
		a.So(devEUIResponse, should.NotBeNil)
		a.So(devEUIResponse.DevEui, should.NotBeZeroValue)

		_, err = reg.Delete(ctx, created.GetIds(), creds)
		a.So(err, should.BeNil)

		_, err = reg.Purge(ctx, created.GetIds(), creds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Purge(ctx, created.GetIds(), userCreds(adminUserIdx))
		a.So(err, should.BeNil)
	})
}

func TestApplicationsPagination(t *testing.T) {
	a := assertions.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := paginationUser.GetIds()
		creds := userCreds(paginationUserIdx)

		reg := ttnpb.NewApplicationRegistryClient(cc)

		list, err := reg.List(test.Context(), &ttnpb.ListApplicationsRequest{
			FieldMask:    &pbtypes.FieldMask{Paths: []string{"name"}},
			Collaborator: userID.GetOrganizationOrUserIdentifiers(),
			Limit:        2,
			Page:         1,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(list, should.NotBeNil) {
			a.So(list.Applications, should.HaveLength, 2)
		}

		list, err = reg.List(test.Context(), &ttnpb.ListApplicationsRequest{
			FieldMask:    &pbtypes.FieldMask{Paths: []string{"name"}},
			Collaborator: userID.OrganizationOrUserIdentifiers(),
			Limit:        2,
			Page:         2,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(list, should.NotBeNil) {
			a.So(list.Applications, should.HaveLength, 1)
		}

		list, err = reg.List(test.Context(), &ttnpb.ListApplicationsRequest{
			FieldMask:    &pbtypes.FieldMask{Paths: []string{"name"}},
			Collaborator: userID.OrganizationOrUserIdentifiers(),
			Limit:        2,
			Page:         3,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(list, should.NotBeNil) {
			a.So(list.Applications, should.BeEmpty)
		}
	})
}
