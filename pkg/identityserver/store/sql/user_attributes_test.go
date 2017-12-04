// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var userWithFooFactory = func() types.User {
	return &userWithFoo{}
}

// userWithFoo implements both types.User and store.Attributer interfaces.
type userWithFoo struct {
	*ttnpb.User
	Foo string
}

// GetUser returns the DefaultUser.
func (u *userWithFoo) GetUser() *ttnpb.User {
	return u.User
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

	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS foo_users (
			user_id  STRING(36) REFERENCES users(user_id) NOT NULL,
			foo		 STRING
		);
	`)
	a.So(err, should.BeNil)

	user := &ttnpb.User{
		UserIdentifier: ttnpb.UserIdentifier{"attributer"},
		Password:       "secret",
		Email:          "john@example.net",
	}

	withFoo := &userWithFoo{
		User: user,
		Foo:  "bar",
	}

	err = s.Users.Create(withFoo)
	a.So(err, should.BeNil)

	found, err := s.Users.GetByID(withFoo.GetUser().UserID, userWithFooFactory)
	a.So(err, should.BeNil)
	//a.So(found, test.ShouldBeUser, created)
	a.So(found, test.ShouldBeUserIgnoringAutoFields, withFoo)
	//a.So((found.(*userWithFoo)).Foo, should.Equal, (cre.Foo)
}
