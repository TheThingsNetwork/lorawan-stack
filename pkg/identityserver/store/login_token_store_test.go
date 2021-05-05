// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

func TestLoginTokenStore(t *testing.T) {
	a, ctx := test.New(t)
	now := cleanTime(time.Now())

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		s := newStore(db)
		store := GetLoginTokenStore(db)

		prepareTest(db,
			&Account{}, &User{},
			&LoginToken{},
		)

		usr := &User{Account: Account{UID: "login-token-test-user"}}
		s.createEntity(ctx, usr)

		loginToken, err := store.CreateLoginToken(ctx, &ttnpb.LoginToken{
			UserIdentifiers: ttnpb.UserIdentifiers{UserId: usr.Account.UID},
			ExpiresAt:       cleanTime(now.Add(-1 * time.Minute)),
			Token:           "test-expired-login-token",
		})
		a.So(err, should.BeNil)

		_, err = store.ConsumeLoginToken(ctx, loginToken.Token)
		if a.So(err, should.NotBeNil) {
			a.So(errors.Resemble(err, errLoginTokenExpired), should.BeTrue)
		}

		loginToken, err = store.CreateLoginToken(ctx, &ttnpb.LoginToken{
			UserIdentifiers: ttnpb.UserIdentifiers{UserId: usr.Account.UID},
			ExpiresAt:       cleanTime(now.Add(time.Hour)),
			Token:           "test-login-token",
		})
		a.So(err, should.BeNil)

		tokens, err := store.FindActiveLoginTokens(ctx, &ttnpb.UserIdentifiers{UserId: usr.Account.UID})
		if a.So(err, should.BeNil) && a.So(tokens, should.HaveLength, 1) {
			a.So(tokens[0].Token, should.Equal, loginToken.Token)
		}

		consumedToken, err := store.ConsumeLoginToken(ctx, loginToken.Token)
		if a.So(err, should.BeNil) {
			a.So(consumedToken.UserIdentifiers, should.Resemble, ttnpb.UserIdentifiers{UserId: usr.Account.UID})
		}

		tokens, err = store.FindActiveLoginTokens(ctx, &ttnpb.UserIdentifiers{UserId: usr.Account.UID})
		if a.So(err, should.BeNil) {
			a.So(tokens, should.BeEmpty)
		}

		_, err = store.ConsumeLoginToken(ctx, loginToken.Token)
		if a.So(err, should.NotBeNil) {
			a.So(errors.Resemble(err, errLoginTokenAlreadyUsed), should.BeTrue)
		}
	})
}
