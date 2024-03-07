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

package storetest

import (
	"testing"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	is "go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

// TestBasicBookmarkOperations tests the UserBookmarkStore interface.
func (st *StoreTest) TestBasicBookmarkOperations(t *testing.T) { // nolint:paralleltest
	usr1 := st.population.NewUser()
	usr2 := st.population.NewUser()
	org1 := st.population.NewOrganization(usr1.GetOrganizationOrUserIdentifiers())
	gtw1 := st.population.NewGateway(usr1.GetOrganizationOrUserIdentifiers())
	app1 := st.population.NewApplication(usr1.GetOrganizationOrUserIdentifiers())
	clt1 := st.population.NewClient(usr1.GetOrganizationOrUserIdentifiers())
	dev1 := st.population.NewEndDevice(app1.GetIds())

	s, ok := st.PrepareDB(t).(interface {
		Store
		is.UserBookmarkStore
	})
	defer st.DestroyDB(t, true)
	if !ok {
		t.Skip("Store does not implement UserBookmarkStore")
	}
	defer s.Close()

	entityIDs := []*ttnpb.EntityIdentifiers{
		usr1.GetEntityIdentifiers(),
		org1.GetEntityIdentifiers(),
		gtw1.GetEntityIdentifiers(),
		app1.GetEntityIdentifiers(),
		clt1.GetEntityIdentifiers(),
		dev1.GetEntityIdentifiers(),
	}

	usrIDS := []*ttnpb.UserIdentifiers{
		usr1.GetIds(),
		usr2.GetIds(),
	}

	t.Run("CreateBookmark", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		for _, entityID := range entityIDs {
			for _, usrID := range usrIDS {
				got, err := s.CreateBookmark(ctx, &ttnpb.UserBookmark{
					UserIds:   usrID,
					EntityIds: entityID,
				})
				a.So(err, should.BeNil)
				a.So(got.UserIds, should.Resemble, usrID)
				a.So(got.EntityIds, should.Resemble, entityID)
			}
		}
	})

	t.Run("CreateBookmark_Duplicate", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		_, err := s.CreateBookmark(ctx, &ttnpb.UserBookmark{
			UserIds:   usr1.GetIds(),
			EntityIds: org1.GetEntityIdentifiers(),
		})
		a.So(err, should.NotBeNil)
	})
	t.Run("FindBookmarks_AfterCreateBookmark", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		bookmarks, err := s.FindBookmarks(ctx, usr1.GetIds())
		a.So(err, should.BeNil)
		a.So(bookmarks, should.HaveLength, len(entityIDs))
	})

	t.Run("PurgeBookmark", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		err := s.PurgeBookmark(ctx, &ttnpb.UserBookmark{
			UserIds:   usr1.GetIds(),
			EntityIds: org1.GetEntityIdentifiers(),
		})
		a.So(err, should.BeNil)
	})
	t.Run("PurgeBookmark_Not_Found", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		err := s.PurgeBookmark(ctx, &ttnpb.UserBookmark{
			UserIds:   usr1.GetIds(),
			EntityIds: org1.GetEntityIdentifiers(),
		})
		a.So(errors.IsNotFound(err), should.BeTrue)
	})
	t.Run("FindBookmarks_AfterPurgeBookmark", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		bookmarks, err := s.FindBookmarks(ctx, usr1.GetIds())
		a.So(err, should.BeNil)
		a.So(bookmarks, should.HaveLength, len(entityIDs)-1)
	})

	t.Run("BatchPurgeBookmarks", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		got, err := s.BatchPurgeBookmarks(ctx, usr1.GetIds(), entityIDs)
		a.So(got, should.HaveLength, len(entityIDs)-1) // Minus the app bookmark removed in previous tests.
		a.So(err, should.BeNil)

		got, err = s.BatchPurgeBookmarks(ctx, usr2.GetIds(), entityIDs)
		a.So(got, should.HaveLength, len(entityIDs))
		a.So(err, should.BeNil)
	})
	t.Run("FindBookmarks_AfterBatchPurgeBookmarks", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		bookmarks, err := s.FindBookmarks(ctx, usr1.GetIds())
		a.So(err, should.BeNil)
		a.So(bookmarks, should.HaveLength, 0)

		bookmarks, err = s.FindBookmarks(ctx, usr2.GetIds())
		a.So(err, should.BeNil)
		a.So(bookmarks, should.HaveLength, 0)
	})
}

// TestBookmarksEntityAndUserOperations tests the UserBookmarkStore interface's entity and user operations.
func (st *StoreTest) TestBookmarksEntityAndUserOperations(t *testing.T) { // nolint:paralleltest
	usr1 := st.population.NewUser()
	usr2 := st.population.NewUser()
	org1 := st.population.NewOrganization(usr1.GetOrganizationOrUserIdentifiers())
	gtw1 := st.population.NewGateway(usr1.GetOrganizationOrUserIdentifiers())
	app1 := st.population.NewApplication(usr1.GetOrganizationOrUserIdentifiers())
	clt1 := st.population.NewClient(usr1.GetOrganizationOrUserIdentifiers())
	dev1 := st.population.NewEndDevice(app1.GetIds())

	s, ok := st.PrepareDB(t).(interface {
		Store
		is.UserBookmarkStore
	})
	defer st.DestroyDB(t, true)
	if !ok {
		t.Skip("Store does not implement UserBookmarkStore")
	}
	defer s.Close()

	entityIDs := []*ttnpb.EntityIdentifiers{
		usr1.GetEntityIdentifiers(),
		org1.GetEntityIdentifiers(),
		gtw1.GetEntityIdentifiers(),
		app1.GetEntityIdentifiers(),
		clt1.GetEntityIdentifiers(),
		dev1.GetEntityIdentifiers(),
	}

	usrIDS := []*ttnpb.UserIdentifiers{
		usr1.GetIds(),
		usr2.GetIds(),
	}

	t.Run("CreateBookmark", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		for _, entityID := range entityIDs {
			for _, usrID := range usrIDS {
				got, err := s.CreateBookmark(ctx, &ttnpb.UserBookmark{
					UserIds:   usrID,
					EntityIds: entityID,
				})
				a.So(err, should.BeNil)
				a.So(got.UserIds, should.Resemble, usrID)
				a.So(got.EntityIds, should.Resemble, entityID)
			}
		}
	})

	t.Run("DeleteEntityBookmarks", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		err := s.DeleteEntityBookmarks(ctx, app1.GetEntityIdentifiers())
		a.So(err, should.BeNil)
	})
	t.Run("DeleteEntityBookmarks_Second_time", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		err := s.DeleteEntityBookmarks(ctx, app1.GetEntityIdentifiers())
		a.So(err, should.BeNil)
	})
	t.Run("FindBookmarks_AfterDeleteEntityBookmarks", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		bookmarks, err := s.FindBookmarks(ctx, usr1.GetIds())
		a.So(err, should.BeNil)
		a.So(bookmarks, should.HaveLength, len(entityIDs)-1)

		bookmarks, err = s.FindBookmarks(ctx, usr2.GetIds())
		a.So(err, should.BeNil)
		a.So(bookmarks, should.HaveLength, len(entityIDs)-1)
	})

	t.Run("RestoreEntityBookmarks", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		err := s.RestoreEntityBookmarks(ctx, app1.GetEntityIdentifiers())
		a.So(err, should.BeNil)
	})
	t.Run("RestoreEntityBookmarks_Second_time", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		err := s.RestoreEntityBookmarks(ctx, app1.GetEntityIdentifiers())
		a.So(err, should.BeNil)
	})
	t.Run("FindBookmarks_AfterRestoreEntityBookmarks", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		bookmarks, err := s.FindBookmarks(ctx, usr1.GetIds())
		a.So(err, should.BeNil)
		a.So(bookmarks, should.HaveLength, len(entityIDs))

		bookmarks, err = s.FindBookmarks(ctx, usr2.GetIds())
		a.So(err, should.BeNil)
		a.So(bookmarks, should.HaveLength, len(entityIDs))
	})

	t.Run("PurgeEntityBookmarks", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		err := s.PurgeEntityBookmarks(ctx, app1.GetEntityIdentifiers())
		a.So(err, should.BeNil)
	})
	t.Run("PurgeEntityBookmarks_Second_time", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		err := s.PurgeEntityBookmarks(ctx, app1.GetEntityIdentifiers())
		a.So(err, should.BeNil)
	})
	t.Run("FindBookmarks_AfterPurgeEntityBookmarks", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		bookmarks, err := s.FindBookmarks(ctx, usr1.GetIds())
		a.So(err, should.BeNil)
		a.So(bookmarks, should.HaveLength, len(entityIDs)-1)

		bookmarks, err = s.FindBookmarks(ctx, usr2.GetIds())
		a.So(err, should.BeNil)
		a.So(bookmarks, should.HaveLength, len(entityIDs)-1)
	})

	t.Run("DeleteUserBookmarks", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		err := s.DeleteUserBookmarks(ctx, usr1.GetIds())
		a.So(err, should.BeNil)
	})
	t.Run("DeleteUserBookmarks_Second_time", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		err := s.DeleteUserBookmarks(ctx, usr1.GetIds())
		a.So(err, should.BeNil)
	})
	t.Run("FindBookmarks_AfterDeleteUserBookmarks", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		bookmarks, err := s.FindBookmarks(ctx, usr1.GetIds())
		a.So(err, should.BeNil)
		a.So(bookmarks, should.HaveLength, 0)

		bookmarks, err = s.FindBookmarks(ctx, usr2.GetIds())
		a.So(err, should.BeNil)
		a.So(bookmarks, should.HaveLength, len(entityIDs)-1) // Minus the app bookmark removed in previous tests.
	})

	t.Run("RestoreUserBookmarks", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		err := s.RestoreUserBookmarks(ctx, usr1.GetIds())
		a.So(err, should.BeNil)
	})
	t.Run("RestoreUserBookmarks_Second_time", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		err := s.RestoreUserBookmarks(ctx, usr1.GetIds())
		a.So(err, should.BeNil)
	})
	t.Run("FindBookmarks_AfterRestoreUserBookmarks", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		bookmarks, err := s.FindBookmarks(ctx, usr1.GetIds())
		a.So(err, should.BeNil)
		a.So(bookmarks, should.HaveLength, len(entityIDs)-1) // Minus the app bookmark removed in previous tests.

		bookmarks, err = s.FindBookmarks(ctx, usr2.GetIds())
		a.So(err, should.BeNil)
		a.So(bookmarks, should.HaveLength, len(entityIDs)-1) // Minus the app bookmark removed in previous tests.
	})

	t.Run("PurgeUserBookmarks", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		err := s.PurgeUserBookmarks(ctx, usr1.GetIds())
		a.So(err, should.BeNil)
	})
	t.Run("PurgeUserBookmarks_Second_time", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		err := s.PurgeUserBookmarks(ctx, usr1.GetIds())
		a.So(err, should.BeNil)
	})
	t.Run("FindBookmarks_AfterPurgeUserBookmarks", func(t *testing.T) { // nolint:paralleltest
		a, ctx := test.New(t)
		bookmarks, err := s.FindBookmarks(ctx, usr1.GetIds())
		a.So(err, should.BeNil)
		a.So(bookmarks, should.HaveLength, 0)

		bookmarks, err = s.FindBookmarks(ctx, usr2.GetIds())
		a.So(err, should.BeNil)
		a.So(bookmarks, should.HaveLength, len(entityIDs)-1) // Minus the app bookmark removed in previous tests.
	})
}
