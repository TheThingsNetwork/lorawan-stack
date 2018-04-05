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
	UserIdentifiers: ttnpb.UserIdentifiers{UserID: "alice"},
	Password:        "123456",
	Email:           "alice@alice.com",
	ValidatedAt:     timeValue(time.Now()),
}

var bob = &ttnpb.User{
	UserIdentifiers: ttnpb.UserIdentifiers{UserID: "bob"},
	Password:        "123456",
	Email:           "bob@bob.com",
}

func TestUsers(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	// Users are already created on cleanStore so creation is skipped here.

	for _, user := range []*ttnpb.User{alice, bob} {
		err := s.Users.Create(user)
		a.So(err, should.NotBeNil)
		a.So(ErrUserIDTaken.Describes(err), should.BeTrue)
	}

	users, err := s.Users.List(userSpecializer)
	a.So(err, should.BeNil)
	a.So(users, should.HaveLength, 2)

	found, err := s.Users.GetByEmail(alice.Email, userSpecializer)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeUserIgnoringAutoFields, alice)

	found, err = s.Users.GetByID(bob.UserIdentifiers, userSpecializer)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeUserIgnoringAutoFields, bob)

	alice.Password = "qwerty"
	err = s.Users.Update(alice)
	a.So(err, should.BeNil)

	updated, err := s.Users.GetByID(alice.UserIdentifiers, userSpecializer)
	a.So(err, should.BeNil)
	a.So(updated, test.ShouldBeUserIgnoringAutoFields, alice)

	alice.Email = bob.Email

	err = s.Users.Update(alice)
	a.So(err, should.NotBeNil)
	a.So(ErrUserEmailTaken.Describes(err), should.BeTrue)

	key := ttnpb.APIKey{
		Key:    "abcabcabc",
		Name:   "foo",
		Rights: []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)},
	}

	list, err := s.Users.ListAPIKeys(alice.UserIdentifiers)
	a.So(err, should.BeNil)
	a.So(list, should.HaveLength, 0)

	err = s.Users.SaveAPIKey(alice.UserIdentifiers, key)
	a.So(err, should.BeNil)

	key2 := ttnpb.APIKey{
		Key:    "123abc",
		Name:   "foo",
		Rights: []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)},
	}
	err = s.Users.SaveAPIKey(alice.UserIdentifiers, key2)
	a.So(err, should.NotBeNil)
	a.So(ErrAPIKeyNameConflict.Describes(err), should.BeTrue)

	ids, foundKey, err := s.Users.GetAPIKey(key.Key)
	a.So(err, should.BeNil)
	a.So(ids, should.Resemble, alice.UserIdentifiers)
	a.So(foundKey, should.Resemble, key)

	foundKey, err = s.Users.GetAPIKeyByName(alice.UserIdentifiers, key.Name)
	a.So(err, should.BeNil)
	a.So(foundKey, should.Resemble, key)

	key.Rights = append(key.Rights, ttnpb.Right(5))
	err = s.Users.UpdateAPIKeyRights(alice.UserIdentifiers, key.Name, key.Rights)
	a.So(err, should.BeNil)

	list, err = s.Users.ListAPIKeys(alice.UserIdentifiers)
	a.So(err, should.BeNil)
	if a.So(list, should.HaveLength, 1) {
		a.So(list[0], should.Resemble, key)
	}

	err = s.Users.DeleteAPIKey(alice.UserIdentifiers, key.Name)
	a.So(err, should.BeNil)

	_, err = s.Users.GetAPIKeyByName(alice.UserIdentifiers, key.Name)
	a.So(err, should.NotBeNil)
	a.So(ErrAPIKeyNotFound.Describes(err), should.BeTrue)

	err = s.Users.Delete(bob.UserIdentifiers)
	a.So(err, should.BeNil)

	found, err = s.Users.GetByID(bob.UserIdentifiers, userSpecializer)
	a.So(err, should.NotBeNil)
	a.So(ErrUserNotFound.Describes(err), should.BeTrue)
	a.So(found, should.BeNil)
}

func TestUserTx(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	user := &ttnpb.User{
		UserIdentifiers: ttnpb.UserIdentifiers{UserID: "tx-test"},
		Password:        "123456",
		Email:           "tx@tx.com",
	}

	err := s.Transact(func(s *store.Store) error {
		if err := s.Users.Create(user); err != nil {
			return err
		}

		user.Name = "PEPE"
		return s.Users.Update(user)
	})
	a.So(err, should.BeNil)

	found, err := s.Users.GetByID(user.UserIdentifiers, userSpecializer)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeUserIgnoringAutoFields, user)

	// delete user
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

	// previous token will be erased
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
			UserIdentifiers: ttnpb.UserIdentifiers{UserID: string(n)},
			Email:           fmt.Sprintf("%v@gmail.com", n),
			Password:        "secret",
		})
	}
}
