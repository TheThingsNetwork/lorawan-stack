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

func TestOauthStore(t *testing.T) {
	ctx := test.Context()

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		prepareTest(db,
			&ClientAuthorization{},
			&AuthorizationCode{},
			&AccessToken{},
			&User{},
			&Client{},
			&Account{},
		)
		store := GetOAuthStore(db)

		db.Create(&User{Account: Account{UID: "test-user"}})
		userIDs := &ttnpb.UserIdentifiers{UserID: "test-user"}

		db.Create(&Client{ClientID: "test-client"})
		clientIDs := &ttnpb.ClientIdentifiers{ClientID: "test-client"}

		rights := []ttnpb.Right{ttnpb.RIGHT_ALL}

		t.Run("Authorize", func(t *testing.T) {
			a := assertions.New(t)

			empty, err := store.GetAuthorization(ctx, userIDs, clientIDs)

			a.So(empty, should.BeNil)
			a.So(err, should.NotBeNil)
			a.So(errors.IsNotFound(err), should.BeTrue)

			start := time.Now()

			created, err := store.Authorize(ctx, &ttnpb.OAuthClientAuthorization{
				ClientIDs: *clientIDs,
				UserIDs:   *userIDs,
				Rights:    rights,
			})

			a.So(created, should.NotBeNil)
			a.So(err, should.BeNil)
			a.So(created.UserIDs.UserID, should.Equal, "test-user")
			a.So(created.ClientIDs.ClientID, should.Equal, "test-client")
			a.So(created.CreatedAt, should.HappenAfter, start)
			a.So(created.UpdatedAt, should.HappenAfter, start)

			got, err := store.GetAuthorization(ctx, userIDs, clientIDs)

			a.So(got, should.NotBeNil)
			a.So(err, should.BeNil)
			a.So(created.UserIDs.UserID, should.Equal, "test-user")
			a.So(created.ClientIDs.ClientID, should.Equal, "test-client")
			a.So(got.CreatedAt, should.HappenAfter, start)
			a.So(got.UpdatedAt, should.HappenAfter, start)

			err = store.DeleteAuthorization(ctx, userIDs, clientIDs)

			a.So(err, should.BeNil)

			deleted, err := store.GetAuthorization(ctx, userIDs, clientIDs)

			a.So(deleted, should.BeNil)
			a.So(err, should.NotBeNil)
			a.So(errors.IsNotFound(err), should.BeTrue)
		})
	})
}
