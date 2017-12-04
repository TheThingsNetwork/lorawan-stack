// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"fmt"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var userFactory = func() types.User {
	return &ttnpb.User{}
}

func testUsers() map[string]*ttnpb.User {
	return map[string]*ttnpb.User{
		"alice": {
			UserIdentifier: ttnpb.UserIdentifier{"alice"},
			Password:       "123456",
			Email:          "alice@alice.com",
		},
		"bob": {
			UserIdentifier: ttnpb.UserIdentifier{"bob"},
			Password:       "1234567",
			Email:          "bob@bob.com",
		},
		"john-doe": {
			UserIdentifier: ttnpb.UserIdentifier{"john-doe"},
			Password:       "123456",
			Email:          "john@doe.com",
		},
	}
}

func TestUserTx(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	john := testUsers()["alice"]
	john.UserID = "john"
	john.Email = "john@john.com"

	err := s.Transact(func(s store.Store) error {
		if err := s.Users.Create(john); err != nil {
			return err
		}

		john.Name = "PEPE"
		if err := s.Users.Update(john); err != nil {
			return err
		}

		return s.Users.Archive(john.UserID)
	})
	a.So(err, should.BeNil)

	found, err := s.Users.GetByID(john.UserID, userFactory)
	a.So(err, should.BeNil)
	john.ArchivedAt = found.GetUser().ArchivedAt
	a.So(found, test.ShouldBeUserIgnoringAutoFields, john)
}

func TestUserCreate(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	for _, user := range testUsers() {
		err := s.Users.Create(user)
		a.So(err, should.NotBeNil)
		a.So(err.(errors.Error).Code(), should.Equal, 402)
		a.So(err.(errors.Error).Type(), should.Equal, errors.AlreadyExists)
	}
}

func TestUserGet(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	alice := testUsers()["alice"]
	bob := testUsers()["bob"]

	// Find by email
	{
		found, err := s.Users.GetByEmail(alice.Email, userFactory)
		a.So(err, should.BeNil)
		a.So(found, test.ShouldBeUserIgnoringAutoFields, alice)
	}

	// Find by user ID
	{
		found, err := s.Users.GetByID(bob.UserID, userFactory)
		a.So(err, should.BeNil)
		a.So(found, test.ShouldBeUserIgnoringAutoFields, bob)
	}
}

func TestUserUpdate(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	bob := testUsers()["bob"]
	alice := testUsers()["alice"]

	// Update user
	{
		alice.Password = "qwerty"
		err := s.Users.Update(alice)
		a.So(err, should.BeNil)

		updated, err := s.Users.GetByID(alice.UserID, userFactory)
		a.So(err, should.BeNil)
		a.So(updated, test.ShouldBeUserIgnoringAutoFields, alice)
	}

	// Try to update email to a taken one should throw an error
	{
		alice.Email = bob.Email

		err := s.Users.Update(alice)
		a.So(err, should.NotBeNil)
		a.So(err.(errors.Error).Code(), should.Equal, 403)
		a.So(err.(errors.Error).Type(), should.Equal, errors.AlreadyExists)
	}
}

func TestUserArchive(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	bob := testUsers()["bob"]

	err := s.Users.Archive(bob.UserID)
	a.So(err, should.BeNil)

	found, err := s.Users.GetByID(bob.UserID, userFactory)
	a.So(err, should.BeNil)
	a.So(found.GetUser().ArchivedAt.IsZero(), should.BeFalse)
}

func BenchmarkUserCreate(b *testing.B) {
	s := testStore(b)

	for n := 0; n < b.N; n++ {
		s.Users.Create(&ttnpb.User{
			UserIdentifier: ttnpb.UserIdentifier{string(n)},
			Email:          fmt.Sprintf("%v@gmail.com", n),
			Password:       "secret",
		})
	}
}
