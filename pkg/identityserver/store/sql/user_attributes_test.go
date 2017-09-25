// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"errors"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/factory"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

// userWithFooFactory implements factory.UserFactory.
type userWithFooFactory struct{}

// User returns an userWithFoo type.
func (f userWithFooFactory) User() types.User {
	return &userWithFoo{}
}

// userWithFoo implements both types.User and store.Attributer interfaces.
type userWithFoo struct {
	*types.DefaultUser
	Foo string
}

// GetUser returns the DefaultUser.
func (u *userWithFoo) GetUser() *types.DefaultUser {
	return u.DefaultUser
}

// Namespaces returns the namespaces userWithFoo have extra attributes in.
func (u *userWithFoo) Namespaces() []string {
	return []string{
		"foo",
	}
}

// Attributes returns for a given namespace a map containing the type extra attributes.
func (u *userWithFoo) Attributes(namespace string) map[string]interface{} {
	if namespace != "foo" {
		return nil
	}

	return map[string]interface{}{
		"foo": u.Foo,
	}
}

// Fill fills an userWithFoo type with the extra attributes that were found in the store.
func (u *userWithFoo) Fill(namespace string, attributes map[string]interface{}) error {
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
	a := assertions.New(t)
	s := testStore(t)

	// Set userWithFooFactory as the User Factory
	s.Users.SetFactory(userWithFooFactory{})

	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS foo_users (
			username STRING(36) REFERENCES users(username) NOT NULL,
			foo		  STRING
		);
	`)
	a.So(err, should.BeNil)

	user := &types.DefaultUser{
		Username: "john-doe",
		Password: "secret",
		Email:    "john@example.net",
	}

	withFoo := &userWithFoo{
		DefaultUser: user,
		Foo:         "bar",
	}

	created, err := s.Users.Register(withFoo)
	a.So(err, should.BeNil)

	a.So(created, test.ShouldBeUserIgnoringAutoFields, withFoo)
	a.So(created.GetUser().Username, should.Equal, withFoo.GetUser().Username)
	a.So((created.(*userWithFoo)).Foo, should.Equal, withFoo.Foo)

	found, err := s.Users.FindByUsername(created.GetUser().Username)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeUser, created)
	a.So(found, test.ShouldBeUserIgnoringAutoFields, withFoo)
	a.So((found.(*userWithFoo)).Foo, should.Equal, (created.(*userWithFoo)).Foo)

	// Set back the DefaultUser factory
	s.Users.SetFactory(factory.DefaultUser{})
}
