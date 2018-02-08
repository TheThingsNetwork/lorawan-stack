// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"fmt"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var userSpecializer = func(base ttnpb.User) store.User {
	return &base
}

func testUsers() map[string]*ttnpb.User {
	return map[string]*ttnpb.User{
		"alice": {
			UserIdentifier: ttnpb.UserIdentifier{UserID: "alice"},
			Password:       "123456",
			Email:          "alice@alice.com",
		},
		"bob": {
			UserIdentifier: ttnpb.UserIdentifier{UserID: "bob"},
			Password:       "1234567",
			Email:          "bob@bob.com",
		},
		"john-doe": {
			UserIdentifier: ttnpb.UserIdentifier{UserID: "john-doe"},
			Password:       "123456",
			Email:          "john@doe.com",
		},
	}
}

func TestUserTx(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	john := testUsers()["alice"]
	john.UserID = "john"
	john.Email = "john@john.com"

	err := s.Transact(func(s *store.Store) error {
		if err := s.Users.Create(john); err != nil {
			return err
		}

		john.Name = "PEPE"
		return s.Users.Update(john)
	})
	a.So(err, should.BeNil)

	found, err := s.Users.GetByID(john.UserID, userSpecializer)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeUserIgnoringAutoFields, john)

	// delete user
	err = s.Users.Delete(john.UserID)
	a.So(err, should.BeNil)
}

func TestUserCreate(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	for _, user := range testUsers() {
		err := s.Users.Create(user)
		a.So(err, should.NotBeNil)
		a.So(ErrUserIDTaken.Describes(err), should.BeTrue)
	}
}

func TestUserList(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	// TODO(gomezjdaniel): correct result is 3 instead of 4 as the example user
	// attributer wasn't deleted it.
	found, err := s.Users.List(userSpecializer)
	a.So(err, should.BeNil)
	a.So(found, should.HaveLength, 4)
}

func TestUserGet(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	alice := testUsers()["alice"]
	bob := testUsers()["bob"]

	// Find by email
	{
		found, err := s.Users.GetByEmail(alice.Email, userSpecializer)
		a.So(err, should.BeNil)
		a.So(found, test.ShouldBeUserIgnoringAutoFields, alice)
	}

	// Find by user ID
	{
		found, err := s.Users.GetByID(bob.UserID, userSpecializer)
		a.So(err, should.BeNil)
		a.So(found, test.ShouldBeUserIgnoringAutoFields, bob)
	}
}

func TestUserUpdate(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	bob := testUsers()["bob"]
	alice := testUsers()["alice"]

	// Update user
	{
		alice.Password = "qwerty"
		err := s.Users.Update(alice)
		a.So(err, should.BeNil)

		updated, err := s.Users.GetByID(alice.UserID, userSpecializer)
		a.So(err, should.BeNil)
		a.So(updated, test.ShouldBeUserIgnoringAutoFields, alice)
	}

	// Try to update email to a taken one should throw an error
	{
		alice.Email = bob.Email

		err := s.Users.Update(alice)
		a.So(err, should.NotBeNil)
		a.So(ErrUserEmailTaken.Describes(err), should.BeTrue)
	}
}

func TestUserValidationToken(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	userID := testUsers()["bob"].UserID
	token := &store.ValidationToken{
		ValidationToken: "foo-token",
		CreatedAt:       time.Now(),
		ExpiresIn:       3600,
	}

	err := s.Users.SaveValidationToken(userID, token)
	a.So(err, should.BeNil)

	uID, found, err := s.Users.GetValidationToken(token.ValidationToken)
	a.So(err, should.BeNil)
	a.So(uID, should.Equal, userID)
	a.So(found.ValidationToken, should.Equal, token.ValidationToken)
	a.So(found.CreatedAt, should.HappenWithin, time.Millisecond, token.CreatedAt)
	a.So(found.ExpiresIn, should.Equal, token.ExpiresIn)

	err = s.Users.DeleteValidationToken(token.ValidationToken)
	a.So(err, should.BeNil)

	_, _, err = s.Users.GetValidationToken(token.ValidationToken)
	a.So(ErrValidationTokenNotFound.Describes(err), should.BeTrue)

	err = s.Users.SaveValidationToken(userID, token)
	a.So(err, should.BeNil)

	newToken := &store.ValidationToken{
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
	a.So(uID, should.Equal, userID)
	a.So(found.ValidationToken, should.Equal, newToken.ValidationToken)
	a.So(found.CreatedAt, should.HappenWithin, time.Millisecond, newToken.CreatedAt)
	a.So(found.ExpiresIn, should.Equal, newToken.ExpiresIn)
}

func TestUserAPIKeys(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	userID := testUsers()["bob"].UserID
	key := &ttnpb.APIKey{
		Key:    "abcabcabc",
		Name:   "foo",
		Rights: []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)},
	}

	list, err := s.Users.ListAPIKeys(userID)
	a.So(err, should.BeNil)
	a.So(list, should.HaveLength, 0)

	err = s.Users.SaveAPIKey(userID, key)
	a.So(err, should.BeNil)

	key2 := &ttnpb.APIKey{
		Key:    "123abc",
		Name:   "foo",
		Rights: []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)},
	}
	err = s.Users.SaveAPIKey(userID, key2)
	a.So(err, should.NotBeNil)
	a.So(ErrAPIKeyNameConflict.Describes(err), should.BeTrue)

	found, err := s.Users.GetAPIKeyByName(userID, key.Name)
	a.So(err, should.BeNil)
	a.So(found, should.Resemble, key)

	key.Rights = append(key.Rights, ttnpb.Right(5))
	err = s.Users.UpdateAPIKeyRights(userID, key.Name, key.Rights)
	a.So(err, should.BeNil)

	list, err = s.Users.ListAPIKeys(userID)
	a.So(err, should.BeNil)
	if a.So(list, should.HaveLength, 1) {
		a.So(list[0], should.Resemble, key)
	}

	err = s.Users.DeleteAPIKey(userID, key.Name)
	a.So(err, should.BeNil)

	found, err = s.Users.GetAPIKeyByName(userID, key.Name)
	a.So(err, should.NotBeNil)
	a.So(ErrAPIKeyNotFound.Describes(err), should.BeTrue)
	a.So(found, should.BeNil)
}

func TestUserDelete(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	id := "test-delete"

	err := s.Users.Create(&ttnpb.User{
		UserIdentifier: ttnpb.UserIdentifier{UserID: id},
		Email:          "foo",
		Password:       "123",
		Name:           "bar",
	})
	a.So(err, should.BeNil)

	key := &ttnpb.APIKey{
		Name:   "foo",
		Key:    "123",
		Rights: []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)},
	}
	err = s.Users.SaveAPIKey(id, key)
	a.So(err, should.BeNil)

	testApplicationDeleteFeedDatabase(t, id, id)
	testGatewayDeleteFeedDatabase(t, id, id)
	testClientDeleteFeedDatabase(t, id, id)

	err = s.Users.Delete(id)
	a.So(err, should.BeNil)

	found, err := s.Users.GetByID(id, userSpecializer)
	a.So(err, should.NotBeNil)
	a.So(ErrUserNotFound.Describes(err), should.BeTrue)
	a.So(found, should.BeNil)
}

func BenchmarkUserCreate(b *testing.B) {
	s := testStore(b, database)

	for n := 0; n < b.N; n++ {
		s.Users.Create(&ttnpb.User{
			UserIdentifier: ttnpb.UserIdentifier{UserID: string(n)},
			Email:          fmt.Sprintf("%v@gmail.com", n),
			Password:       "secret",
		})
	}
}
