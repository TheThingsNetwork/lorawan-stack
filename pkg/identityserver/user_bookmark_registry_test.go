// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

func TestUsersBookmarksPermissions(t *testing.T) {
	p := &storetest.Population{}

	usr1 := p.NewUser()
	usr1.Password = "OldPassword"
	usr1.PrimaryEmailAddress = "user-1@email.com"
	validatedAtTime := time.Now().Truncate(time.Millisecond)
	usr1.PrimaryEmailAddressValidatedAt = ttnpb.ProtoTime(&validatedAtTime)

	app1 := p.NewApplication(usr1.GetOrganizationOrUserIdentifiers())
	app2 := p.NewApplication(usr1.GetOrganizationOrUserIdentifiers())

	keyWithoutRights, _ := p.NewAPIKey(usr1.GetEntityIdentifiers())
	credsWithoutRights := rpcCreds(keyWithoutRights)

	readOnlyAdmin := p.NewUser()
	readOnlyAdmin.Admin = true
	readOnlyAdminKey, _ := p.NewAPIKey(readOnlyAdmin.GetEntityIdentifiers(), ttnpb.AllReadAdminRights.GetRights()...)
	readOnlyAdminKeyCreds := rpcCreds(readOnlyAdminKey)

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(_ *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewUserBookmarkRegistryClient(cc)

		t.Run("Invalid credentials", func(t *testing.T) { // nolint:paralleltest
			for _, opts := range [][]grpc.CallOption{nil, {credsWithoutRights}} {
				_, err := reg.Create(ctx, &ttnpb.CreateUserBookmarkRequest{
					UserIds:   usr1.Ids,
					EntityIds: app1.GetEntityIdentifiers(),
				}, opts...)
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsPermissionDenied(err), should.BeTrue)
				}

				_, err = reg.List(ctx, &ttnpb.ListUserBookmarksRequest{
					UserIds: usr1.Ids,
				}, opts...)
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsPermissionDenied(err), should.BeTrue)
				}

				_, err = reg.Delete(ctx, &ttnpb.DeleteUserBookmarkRequest{
					UserIds:   usr1.Ids,
					EntityIds: app1.GetEntityIdentifiers(),
				}, opts...)
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsPermissionDenied(err), should.BeTrue)
				}

				_, err = reg.BatchDelete(ctx, &ttnpb.BatchDeleteUserBookmarksRequest{
					UserIds: usr1.Ids,
					EntityIds: []*ttnpb.EntityIdentifiers{
						app1.GetEntityIdentifiers(),
						app2.GetEntityIdentifiers(),
					},
				}, opts...)
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsPermissionDenied(err), should.BeTrue)
				}
			}
		})

		t.Run("Admin read-only", func(t *testing.T) { // nolint:paralleltest
			_, err := reg.Create(ctx, &ttnpb.CreateUserBookmarkRequest{
				UserIds:   usr1.Ids,
				EntityIds: app1.GetEntityIdentifiers(),
			}, readOnlyAdminKeyCreds)
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsPermissionDenied(err), should.BeTrue)
			}

			_, err = reg.List(ctx, &ttnpb.ListUserBookmarksRequest{
				UserIds: usr1.Ids,
			}, readOnlyAdminKeyCreds)
			a.So(errors.IsPermissionDenied(err), should.BeFalse)

			_, err = reg.Delete(ctx, &ttnpb.DeleteUserBookmarkRequest{
				UserIds:   usr1.Ids,
				EntityIds: app1.GetEntityIdentifiers(),
			}, readOnlyAdminKeyCreds)
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsPermissionDenied(err), should.BeTrue)
			}

			_, err = reg.BatchDelete(ctx, &ttnpb.BatchDeleteUserBookmarksRequest{
				UserIds: usr1.Ids,
				EntityIds: []*ttnpb.EntityIdentifiers{
					app1.GetEntityIdentifiers(),
					app2.GetEntityIdentifiers(),
				},
			}, readOnlyAdminKeyCreds)
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsPermissionDenied(err), should.BeTrue)
			}
		})
	}, withPrivateTestDatabase(p))
}

func TestUsersBookmarksOperations(t *testing.T) {
	t.Parallel()

	p := &storetest.Population{}

	usr1 := p.NewUser()
	usr1.Password = "OldPassword"
	usr1.PrimaryEmailAddress = "user-1@email.com"
	validatedAtTime := time.Now().Truncate(time.Millisecond)
	usr1.PrimaryEmailAddressValidatedAt = ttnpb.ProtoTime(&validatedAtTime)

	app1 := p.NewApplication(usr1.GetOrganizationOrUserIdentifiers())
	app2 := p.NewApplication(usr1.GetOrganizationOrUserIdentifiers())

	key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	creds := rpcCreds(key)

	testWithIdentityServer(t, func(_ *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewUserBookmarkRegistryClient(cc)

		t.Run("Create", func(t *testing.T) { // nolint:paralleltest
			a, ctx := test.New(t)
			got, err := reg.Create(ctx, &ttnpb.CreateUserBookmarkRequest{
				UserIds:   usr1.Ids,
				EntityIds: app1.GetEntityIdentifiers(),
			}, creds)
			a.So(err, should.BeNil)
			a.So(got, should.Resemble, &ttnpb.UserBookmark{UserIds: usr1.Ids, EntityIds: app1.GetEntityIdentifiers()})
		})
		t.Run("Create/Duplicate", func(t *testing.T) { // nolint:paralleltest
			a, ctx := test.New(t)
			got, err := reg.Create(ctx, &ttnpb.CreateUserBookmarkRequest{
				UserIds:   usr1.Ids,
				EntityIds: app1.GetEntityIdentifiers(),
			}, creds)
			a.So(got, should.BeNil)
			a.So(errors.IsAlreadyExists(err), should.BeTrue)
		})
		t.Run("Create/UnkownEntity", func(t *testing.T) { // nolint:paralleltest
			a, ctx := test.New(t)

			// org1 is not present in the test's population.
			org1 := &ttnpb.OrganizationIdentifiers{OrganizationId: "org-1"}
			got, err := reg.Create(ctx, &ttnpb.CreateUserBookmarkRequest{
				UserIds:   usr1.Ids,
				EntityIds: org1.GetEntityIdentifiers(),
			}, creds)
			a.So(errors.IsNotFound(err), should.BeTrue)
			a.So(got, should.BeNil)
		})

		t.Run("Create/ExtraBookmark", func(t *testing.T) { // nolint:paralleltest
			a, ctx := test.New(t)
			got, err := reg.Create(ctx, &ttnpb.CreateUserBookmarkRequest{
				UserIds:   usr1.Ids,
				EntityIds: app2.GetEntityIdentifiers(),
			}, creds)
			a.So(err, should.BeNil)
			a.So(got, should.Resemble, &ttnpb.UserBookmark{UserIds: usr1.Ids, EntityIds: app2.GetEntityIdentifiers()})
		})

		t.Run("FindBookmarks/HasEntityType", func(t *testing.T) { // nolint:paralleltest
			a, ctx := test.New(t)

			// Invalid entity type.
			got, err := reg.List(ctx, &ttnpb.ListUserBookmarksRequest{
				UserIds:     usr1.Ids,
				EntityTypes: []string{"invalid"},
			}, creds)
			a.So(got, should.BeNil)
			a.So(errors.IsInvalidArgument(err), should.BeTrue)

			// Entity type without bookmarks.
			got, err = reg.List(ctx, &ttnpb.ListUserBookmarksRequest{
				UserIds:     usr1.Ids,
				EntityTypes: []string{"organization"},
			}, creds)
			a.So(got, should.NotBeNil)
			a.So(got.GetBookmarks(), should.HaveLength, 0)
			a.So(err, should.BeNil)

			// Entity type with bookmarks.
			got, err = reg.List(ctx, &ttnpb.ListUserBookmarksRequest{
				UserIds:     usr1.Ids,
				EntityTypes: []string{"application"},
			}, creds)
			a.So(got.GetBookmarks(), should.HaveLength, 2)
			a.So(err, should.BeNil)

			// Multiple entity types in request
			got, err = reg.List(ctx, &ttnpb.ListUserBookmarksRequest{
				UserIds:     usr1.Ids,
				EntityTypes: []string{"application", "organization", "gateway"},
			}, creds)
			a.So(got.GetBookmarks(), should.HaveLength, 2)
			a.So(err, should.BeNil)
		})
		t.Run("FindBookmarks", func(t *testing.T) { // nolint:paralleltest
			a, ctx := test.New(t)
			got, err := reg.List(ctx, &ttnpb.ListUserBookmarksRequest{
				UserIds: usr1.Ids,
			}, creds)
			if a.So(err, should.BeNil) && a.So(got.Bookmarks, should.HaveLength, 2) {
				a.So(
					got.Bookmarks[0],
					should.Resemble,
					&ttnpb.UserBookmark{UserIds: usr1.Ids, EntityIds: app1.GetEntityIdentifiers()},
				)
				a.So(
					got.Bookmarks[1],
					should.Resemble,
					&ttnpb.UserBookmark{UserIds: usr1.Ids, EntityIds: app2.GetEntityIdentifiers()},
				)
			}
		})

		t.Run("Delete", func(t *testing.T) { // nolint:paralleltest
			a, ctx := test.New(t)
			_, err := reg.Delete(ctx, &ttnpb.DeleteUserBookmarkRequest{
				UserIds:   usr1.Ids,
				EntityIds: app1.GetEntityIdentifiers(),
			}, creds)
			a.So(err, should.BeNil)
		})
		t.Run("FindBookmarks/AfterDelete", func(t *testing.T) { // nolint:paralleltest
			a, ctx := test.New(t)
			got, err := reg.List(ctx, &ttnpb.ListUserBookmarksRequest{
				UserIds: usr1.Ids,
			}, creds)
			if a.So(err, should.BeNil) && a.So(got.Bookmarks, should.HaveLength, 1) {
				a.So(
					got.Bookmarks[0],
					should.Resemble,
					&ttnpb.UserBookmark{UserIds: usr1.Ids, EntityIds: app2.GetEntityIdentifiers()},
				)
			}
		})

		t.Run("BatchDelete", func(t *testing.T) { // nolint:paralleltest
			a, ctx := test.New(t)
			_, err := reg.BatchDelete(ctx, &ttnpb.BatchDeleteUserBookmarksRequest{
				UserIds: usr1.Ids,
				EntityIds: []*ttnpb.EntityIdentifiers{
					app1.GetEntityIdentifiers(),
					app2.GetEntityIdentifiers(),
				},
			}, creds)
			a.So(err, should.BeNil)
		})
		t.Run("FindBookmarks/AfterBatchDelete", func(t *testing.T) { // nolint:paralleltest
			a, ctx := test.New(t)
			got, err := reg.List(ctx, &ttnpb.ListUserBookmarksRequest{
				UserIds: usr1.Ids,
			}, creds)
			a.So(err, should.BeNil)
			a.So(got.Bookmarks, should.HaveLength, 0)
		})
	}, withPrivateTestDatabase(p))
}
