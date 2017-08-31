// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"fmt"
	"testing"

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

	alice := testUsers()["alice"]
	bob := testUsers()["bob"]

	// inserting a user with a username that already exists should. result in an error
	{
		_, err := s.Users.Create(alice)
		a.So(err, should.NotBeNil)
		a.So(err.Error(), should.Equal, ErrUsernameTaken.Error())
	}

	// find by email
	{
		found, err := s.Users.FindByEmail(bob.Email)
		a.So(err, should.BeNil)
		a.So(found, test.ShouldBeUserIgnoringAutoFields, bob)
	}

	// find by username
	{
		found, err := s.Users.FindByUsername(bob.Username)
		a.So(err, should.BeNil)
		a.So(found, test.ShouldBeUserIgnoringAutoFields, bob)
	}
}

func TestUserUpdate(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	alice := testUsers()["alice"]
	alice.Password = "qwerty"

	updated, err := s.Users.Update(alice)
	a.So(err, should.BeNil)
	a.So(updated, test.ShouldBeUserIgnoringAutoFields, alice)
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
		s.Users.Create(&types.DefaultUser{
			Username: string(n),
			Email:    fmt.Sprintf("%v@gmail.com", n),
			Password: "secret",
		})
	}
}
