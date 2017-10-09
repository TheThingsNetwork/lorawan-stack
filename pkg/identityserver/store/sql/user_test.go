// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"fmt"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func testUsers() map[string]*ttnpb.User {
	return map[string]*ttnpb.User{
		"alice": &ttnpb.User{
			UserIdentifier: ttnpb.UserIdentifier{"alice"},
			Password:       "123456",
			Email:          "alice@alice.com",
		},
		"bob": &ttnpb.User{
			UserIdentifier: ttnpb.UserIdentifier{"bob"},
			Password:       "1234567",
			Email:          "bob@bob.com",
		},
	}
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
		found, err := s.Users.GetByEmail(alice.Email)
		a.So(err, should.BeNil)
		a.So(found, test.ShouldBeUserIgnoringAutoFields, alice)
	}

	// Find by user ID
	{
		found, err := s.Users.GetByID(bob.UserID)
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

		updated, err := s.Users.GetByID(alice.UserID)
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

	found, err := s.Users.GetByID(bob.UserID)
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
