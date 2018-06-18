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
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/identityserver"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store/sql/migrations"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

// userWithFoo implements both store.User and store.Attributer interfaces.
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

	schema := `
		CREATE TABLE IF NOT EXISTS foo_users (
			id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			foo		  VARCHAR
		);
	`

	migrations.Registry.Register(migrations.Registry.Count()+1, "test_users_attributer_schema", schema, "")
	s := testStore(t, attributesDatabase)
	s.Init() // to apply the migration

	specializer := func(base ttnpb.User) store.User {
		return &userWithFoo{User: &base}
	}

	user := &ttnpb.User{
		UserIdentifiers: ttnpb.UserIdentifiers{
			UserID: "attributer-user",
			Email:  "john@example.net",
		},
		Password: "secret",
	}

	withFoo := &userWithFoo{
		User: user,
		Foo:  "bar",
	}

	err := s.Users.Create(withFoo)
	a.So(err, should.BeNil)

	found, err := s.Users.GetByID(withFoo.GetUser().UserIdentifiers, specializer)
	a.So(err, should.BeNil)
	a.So(found, should.EqualFieldsWithIgnores(identityserver.UserGeneratedFields...), withFoo)
	a.So(found.(*userWithFoo).Foo, should.Equal, withFoo.Foo)

	err = s.Users.Delete(withFoo.GetUser().UserIdentifiers)
	a.So(err, should.BeNil)

	found, err = s.Users.GetByID(withFoo.GetUser().UserIdentifiers, specializer)
	a.So(err, should.NotBeNil)
	a.So(err, should.DescribeError, store.ErrUserNotFound)
	a.So(found, should.BeNil)
}
