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

package store

import (
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

func TestUserSessionStore(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		prepareTest(db, &Account{}, &User{}, &UserSession{})

		user := &User{
			Account: Account{
				UID: "test",
			},
			Name: "Test User",
		}

		userIDs := ttnpb.UserIdentifiers{UserID: "test"}
		doesNotExistIDs := ttnpb.UserIdentifiers{UserID: "does_not_exist"}

		if err := newStore(db).createEntity(ctx, user); err != nil {
			panic(err)
		}

		store := GetUserSessionStore(db)

		_, err := store.CreateSession(ctx, &ttnpb.UserSession{UserIdentifiers: doesNotExistIDs})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		created, err := store.CreateSession(ctx, &ttnpb.UserSession{
			UserIdentifiers: userIDs,
		})

		a.So(err, should.BeNil)
		if a.So(created, should.NotBeNil) {
			a.So(created.SessionID, should.NotBeEmpty)
			a.So(created.CreatedAt, should.NotBeZeroValue)
			a.So(created.UpdatedAt, should.NotBeZeroValue)
			a.So(created.ExpiresAt, should.BeNil)
		}

		_, err = store.GetSession(ctx, &doesNotExistIDs, created.SessionID)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		got, err := store.GetSession(ctx, &userIDs, created.SessionID)

		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.CreatedAt, should.Equal, created.CreatedAt)
			a.So(got.UpdatedAt, should.Equal, created.UpdatedAt)
			a.So(got.ExpiresAt, should.BeNil)
		}

		later := time.Now().Add(time.Hour)
		updated, err := store.UpdateSession(ctx, &ttnpb.UserSession{
			UserIdentifiers: userIDs,
			SessionID:       created.SessionID,
			ExpiresAt:       &later,
		})

		a.So(err, should.BeNil)
		if a.So(updated, should.NotBeNil) {
			a.So(updated.CreatedAt, should.Equal, created.CreatedAt)
			a.So(updated.UpdatedAt, should.NotEqual, created.UpdatedAt)
			a.So(updated.ExpiresAt, should.NotBeNil)
		}

		_, err = store.UpdateSession(ctx, &ttnpb.UserSession{
			UserIdentifiers: ttnpb.UserIdentifiers{UserID: "does_not_exist"},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		_, err = store.UpdateSession(ctx, &ttnpb.UserSession{UserIdentifiers: userIDs, SessionID: "00000000-0000-0000-0000-000000000000"})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		_, err = store.FindSessions(ctx, &doesNotExistIDs)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		list, err := store.FindSessions(ctx, &userIDs)

		a.So(err, should.BeNil)
		if a.So(list, should.HaveLength, 1) {
			a.So(list[0].CreatedAt, should.Equal, created.CreatedAt)
			a.So(list[0].UpdatedAt, should.Equal, updated.UpdatedAt)
			a.So(list[0].ExpiresAt, should.Resemble, updated.ExpiresAt)
		}

		err = store.DeleteSession(ctx, &doesNotExistIDs, created.SessionID)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		err = store.DeleteSession(ctx, &userIDs, created.SessionID)

		a.So(err, should.BeNil)

		_, err = store.GetSession(ctx, &userIDs, created.SessionID)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		list, err = store.FindSessions(ctx, &userIDs)

		a.So(err, should.BeNil)
		a.So(list, should.BeEmpty)
	})
}
