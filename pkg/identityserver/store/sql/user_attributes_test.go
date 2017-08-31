// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"errors"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/factory"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	. "github.com/smartystreets/assertions"
)

// userWithFooFactory is the type to implement the custom factory for UserWithFoo
type userWithFooFactory struct{}

// User implements UserFactory
func (f userWithFooFactory) User() types.User {
	return &UserWithFoo{}
}

type UserWithFoo struct {
	*types.DefaultUser
	Foo string
}

// GetUser implements User
func (u *UserWithFoo) GetUser() *types.DefaultUser {
	return u.DefaultUser
}

func (u *UserWithFoo) Namespaces() []string {
	return []string{
		"foo",
	}
}

func (u *UserWithFoo) Attributes(namespace string) map[string]interface{} {
	if namespace != "foo" {
		return nil
	}

	return map[string]interface{}{
		"foo": u.Foo,
	}
}

func (u *UserWithFoo) Fill(namespace string, attributes map[string]interface{}) error {
	if namespace != "foo" {
		return nil
	}

	foo, ok := attributes["foo"]
	if !ok {
		return nil
	}

	str, ok := foo.(string)
	if !ok {
		return errors.New("Foo should be a string")
	}

	u.Foo = str
	return nil
}

func TestUserAttributer(t *testing.T) {
	a := New(t)
	s := testStore()

	// Set the UserWithFoo factory in the User Store
	s.Users.SetFactory(userWithFooFactory{})

	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS foo_users (
			username STRING(36) REFERENCES users(username) NOT NULL,
			foo		  STRING
		);
	`)
	a.So(err, ShouldBeNil)

	user := &types.DefaultUser{
		Username: "john-doe",
		Password: "secret",
		Email:    "john@example.net",
	}

	withFoo := &UserWithFoo{
		DefaultUser: user,
		Foo:         "bar",
	}

	created, err := s.Users.Create(withFoo)
	a.So(err, ShouldBeNil)

	a.So(created, test.ShouldBeUserIgnoringAutoFields, withFoo)
	a.So(created.GetUser().Username, ShouldEqual, withFoo.GetUser().Username)
	a.So((created.(*UserWithFoo)).Foo, ShouldEqual, withFoo.Foo)

	found, err := s.Users.FindByUsername(created.GetUser().Username)
	a.So(err, ShouldBeNil)
	a.So(found, test.ShouldBeUser, created)
	a.So(found, test.ShouldBeUserIgnoringAutoFields, withFoo)
	a.So((found.(*UserWithFoo)).Foo, ShouldEqual, (created.(*UserWithFoo)).Foo)

	// Set back the DefaultUser factory
	s.Users.SetFactory(factory.DefaultUser{})
}
