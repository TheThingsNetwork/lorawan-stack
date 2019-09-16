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

func TestInvitationStore(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	now := time.Now()

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		prepareTest(db, &Account{}, &User{}, &Invitation{})

		store := GetInvitationStore(db)

		invitation, err := store.CreateInvitation(ctx, &ttnpb.Invitation{
			Email:     "john.doe@example.com",
			Token:     "invitation-token",
			ExpiresAt: cleanTime(now.Add(time.Hour * 24 * 7)),
		})

		a.So(err, should.BeNil)
		if a.So(invitation, should.NotBeNil) {
			a.So(invitation.Email, should.Equal, "john.doe@example.com")
		}

		_, err = store.GetInvitation(ctx, "wrong-invitation-token")

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		got, err := store.GetInvitation(ctx, "invitation-token")

		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.Email, should.Equal, invitation.Email)
		}

		invitations, err := store.FindInvitations(ctx)

		a.So(err, should.BeNil)
		if a.So(invitations, should.HaveLength, 1) {
			a.So(invitations[0].Email, should.Equal, invitation.Email)
		}

		newUser, err := GetUserStore(db).CreateUser(ctx, &ttnpb.User{UserIdentifiers: ttnpb.UserIdentifiers{UserID: "new-user"}, PrimaryEmailAddress: "new-user@example.com"})
		if err != nil {
			panic(err)
		}

		err = store.SetInvitationAcceptedBy(ctx, "wrong-invitation-token", &newUser.UserIdentifiers)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		err = store.SetInvitationAcceptedBy(ctx, "invitation-token", &newUser.UserIdentifiers)

		a.So(err, should.BeNil)

		err = store.SetInvitationAcceptedBy(ctx, "invitation-token", &newUser.UserIdentifiers)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsAlreadyExists(err), should.BeTrue)
		}

		err = store.DeleteInvitation(ctx, "john.doe@example.com")

		a.So(err, should.NotBeNil)

		_, err = store.CreateInvitation(ctx, &ttnpb.Invitation{
			Email:     "jane.doe@example.com",
			Token:     "other-invitation-token",
			ExpiresAt: cleanTime(now.Add(time.Hour * 24 * 7)),
		})

		a.So(err, should.BeNil)

		err = store.DeleteInvitation(ctx, "jane.doe@example.com")

		a.So(err, should.BeNil)
	})
}
