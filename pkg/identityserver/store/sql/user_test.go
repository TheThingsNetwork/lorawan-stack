// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"fmt"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func testUsers() map[string]*types.DefaultUser {
	return map[string]*types.DefaultUser{
		"alice": &types.DefaultUser{
			Username: "alice",
			Password: "123456",
			Email:    "alice@alice.com",
		},
		"bob": &types.DefaultUser{
			Username: "bob",
			Password: "1234567",
			Email:    "bob@bob.com",
		},
	}
}

func TestUserCreate(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	for _, user := range testUsers() {
		_, err := s.Users.Register(user)
		a.So(err, should.NotBeNil)
		a.So(err.(errors.Error).Code(), should.Equal, 402)
		a.So(err.(errors.Error).Type(), should.Equal, errors.AlreadyExists)
	}
}

func TestUserFind(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	alice := testUsers()["alice"]
	bob := testUsers()["bob"]

	// Find by email
	{
		found, err := s.Users.FindByEmail(alice.Email)
		a.So(err, should.BeNil)
		a.So(found, test.ShouldBeUserIgnoringAutoFields, alice)
	}

	// Find by username
	{
		found, err := s.Users.FindByUsername(bob.Username)
		a.So(err, should.BeNil)
		a.So(found, test.ShouldBeUserIgnoringAutoFields, bob)
	}
}

func TestUserEdit(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	bob := testUsers()["bob"]
	alice := testUsers()["alice"]
	alice.Password = "qwerty"

	// Update user
	{
		updated, err := s.Users.Edit(alice)
		a.So(err, should.BeNil)
		a.So(updated, test.ShouldBeUserIgnoringAutoFields, alice)
	}

	alice.Email = bob.Email
	// Try to update email to an existing one
	{
		_, err := s.Users.Edit(alice)
		a.So(err, should.NotBeNil)
		a.So(err.(errors.Error).Code(), should.Equal, 403)
		a.So(err.(errors.Error).Type(), should.Equal, errors.AlreadyExists)
	}
}

func TestUserArchive(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	bob := testUsers()["bob"]

	err := s.Users.Archive(bob.Username)
	a.So(err, should.BeNil)

	found, err := s.Users.FindByUsername(bob.Username)
	a.So(err, should.BeNil)
	a.So(found.GetUser().Archived, should.NotBeNil)
}

func BenchmarkUserCreate(b *testing.B) {
	s := testStore()

	for n := 0; n < b.N; n++ {
		s.Users.Register(&types.DefaultUser{
			Username: string(n),
			Email:    fmt.Sprintf("%v@gmail.com", n),
			Password: "secret",
		})
	}
}
