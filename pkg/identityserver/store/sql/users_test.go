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

package sql_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	. "github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var userSpecializer = func(base ttnpb.User) store.User {
	return &base
}

var alice = &ttnpb.User{
	UserIdentifiers: ttnpb.UserIdentifiers{
		UserID: "alice",
		Email:  "alice@alice.com",
	},
	Password:    "123456",
	ValidatedAt: timeValue(time.Now()),
}

var bob = &ttnpb.User{
	UserIdentifiers: ttnpb.UserIdentifiers{
		UserID: "bob",
		Email:  "bob@bob.com",
	},
	Password: "123456",
}

func TestUsers(t *testing.T) {
	uids := ttnpb.UserIdentifiers{
		UserID: "test-user",
		Email:  "test@email.com",
	}

	for _, tc := range []struct {
		tcName string
		uids   ttnpb.UserIdentifiers
		sids   ttnpb.UserIdentifiers
	}{
		{
			"SearchByUserID",
			uids,
			ttnpb.UserIdentifiers{
				UserID: uids.UserID,
			},
		},
		{
			"SearchByEmail",
			uids,
			ttnpb.UserIdentifiers{
				Email: uids.Email,
			},
		},
		{
			"SearchByAllIdentifiers",
			uids,
			uids,
		},
	} {
		t.Run(tc.tcName, func(t *testing.T) {
			testUserStore(t, tc.uids, tc.sids)
		})
	}
}

func testUserStore(t *testing.T, uids, sids ttnpb.UserIdentifiers) {
	a := assertions.New(t)
	s := testStore(t, database)

	user := &ttnpb.User{
		UserIdentifiers: uids,
		Name:            "Foo",
		Password:        "Bar",
		Admin:           true,
		ValidatedAt:     timeValue(time.Now()),
	}

	users, err := s.Users.List(userSpecializer)
	a.So(err, should.BeNil)
	a.So(users, should.HaveLength, 2)

	err = s.Users.Create(user)
	a.So(err, should.BeNil)

	users, err = s.Users.List(userSpecializer)
	a.So(err, should.BeNil)
	// It has length 3 because: alice + bob + this user's test.
	a.So(users, should.HaveLength, 3)

	found, err := s.Users.GetByID(sids, userSpecializer)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeUserIgnoringAutoFields, user)

	user.UserIdentifiers.Email = "new@email.com"
	user.Password = "new_password"
	err = s.Users.Update(sids, user)
	a.So(err, should.BeNil)
	if sids.Email != "" {
		sids.Email = user.UserIdentifiers.Email
	}

	updated, err := s.Users.GetByID(sids, userSpecializer)
	a.So(err, should.BeNil)
	a.So(updated, test.ShouldBeUserIgnoringAutoFields, user)

	key := ttnpb.APIKey{
		Key:    "abcabcabc",
		Name:   "foo",
		Rights: []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)},
	}

	list, err := s.Users.ListAPIKeys(sids)
	a.So(err, should.BeNil)
	a.So(list, should.HaveLength, 0)

	err = s.Users.SaveAPIKey(sids, key)
	a.So(err, should.BeNil)

	key2 := ttnpb.APIKey{
		Key:    "123abc",
		Name:   "foo",
		Rights: []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)},
	}
	err = s.Users.SaveAPIKey(sids, key2)
	a.So(err, should.NotBeNil)
	a.So(ErrAPIKeyNameConflict.Describes(err), should.BeTrue)

	ids, foundKey, err := s.Users.GetAPIKey(key.Key)
	a.So(err, should.BeNil)
	a.So(ids, should.Resemble, user.UserIdentifiers)
	a.So(foundKey, should.Resemble, key)

	foundKey, err = s.Users.GetAPIKeyByName(sids, key.Name)
	a.So(err, should.BeNil)
	a.So(foundKey, should.Resemble, key)

	key.Rights = append(key.Rights, ttnpb.Right(5))
	err = s.Users.UpdateAPIKeyRights(user.UserIdentifiers, key.Name, key.Rights)
	a.So(err, should.BeNil)

	list, err = s.Users.ListAPIKeys(sids)
	a.So(err, should.BeNil)
	if a.So(list, should.HaveLength, 1) {
		a.So(list[0], should.Resemble, key)
	}

	err = s.Users.DeleteAPIKey(sids, key.Name)
	a.So(err, should.BeNil)

	_, err = s.Users.GetAPIKeyByName(sids, key.Name)
	a.So(err, should.NotBeNil)
	a.So(ErrAPIKeyNotFound.Describes(err), should.BeTrue)

	err = s.Users.Delete(sids)
	a.So(err, should.BeNil)

	found, err = s.Users.GetByID(sids, userSpecializer)
	a.So(err, should.NotBeNil)
	a.So(ErrUserNotFound.Describes(err), should.BeTrue)
	a.So(found, should.BeNil)
}

func TestUserTx(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	user := &ttnpb.User{
		UserIdentifiers: ttnpb.UserIdentifiers{
			UserID: "tx-test",
			Email:  "tx@tx.com",
		},
		Password: "123456",
	}

	err := s.Transact(func(s *store.Store) error {
		if err := s.Users.Create(user); err != nil {
			return err
		}

		user.Name = "PEPE"
		return s.Users.Update(user.UserIdentifiers, user)
	})
	a.So(err, should.BeNil)

	found, err := s.Users.GetByID(user.UserIdentifiers, userSpecializer)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeUserIgnoringAutoFields, user)

	err = s.Users.Delete(user.UserIdentifiers)
	a.So(err, should.BeNil)

	found, err = s.Users.GetByID(user.UserIdentifiers, userSpecializer)
	a.So(err, should.NotBeNil)
	a.So(ErrUserNotFound.Describes(err), should.BeTrue)
	a.So(found, should.BeNil)
}

func TestUserValidationToken(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	userID := alice.UserIdentifiers
	token := store.ValidationToken{
		ValidationToken: "foo-token",
		CreatedAt:       time.Now(),
		ExpiresIn:       3600,
	}

	err := s.Users.SaveValidationToken(userID, token)
	a.So(err, should.BeNil)

	uID, found, err := s.Users.GetValidationToken(token.ValidationToken)
	a.So(err, should.BeNil)
	a.So(uID, should.Resemble, userID)
	a.So(found.ValidationToken, should.Equal, token.ValidationToken)
	a.So(found.CreatedAt, should.HappenWithin, time.Millisecond, token.CreatedAt)
	a.So(found.ExpiresIn, should.Equal, token.ExpiresIn)

	err = s.Users.DeleteValidationToken(token.ValidationToken)
	a.So(err, should.BeNil)

	_, _, err = s.Users.GetValidationToken(token.ValidationToken)
	a.So(ErrValidationTokenNotFound.Describes(err), should.BeTrue)

	err = s.Users.SaveValidationToken(userID, token)
	a.So(err, should.BeNil)

	newToken := store.ValidationToken{
		ValidationToken: "bar-token",
		CreatedAt:       time.Now(),
		ExpiresIn:       3600,
	}

	// Previous token will be erased.
	err = s.Users.SaveValidationToken(userID, newToken)
	a.So(err, should.BeNil)

	_, _, err = s.Users.GetValidationToken(token.ValidationToken)
	a.So(ErrValidationTokenNotFound.Describes(err), should.BeTrue)

	uID, found, err = s.Users.GetValidationToken(newToken.ValidationToken)
	a.So(err, should.BeNil)
	a.So(uID, should.Resemble, userID)
	a.So(found.ValidationToken, should.Equal, newToken.ValidationToken)
	a.So(found.CreatedAt, should.HappenWithin, time.Millisecond, newToken.CreatedAt)
	a.So(found.ExpiresIn, should.Equal, newToken.ExpiresIn)
}

func BenchmarkUserCreate(b *testing.B) {
	s := testStore(b, database)

	for n := 0; n < b.N; n++ {
		s.Users.Create(&ttnpb.User{
			UserIdentifiers: ttnpb.UserIdentifiers{
				UserID: string(n),
				Email:  fmt.Sprintf("%v@gmail.com", n),
			},
			Password: "secret",
		})
	}
}
