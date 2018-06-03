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
	errshould "go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

// clientWithFoo implements both store.Client and store.Attributer interfaces.
type clientWithFoo struct {
	*ttnpb.Client
	Foo string
}

// GetClient returns the base Client.
func (u *clientWithFoo) GetClient() *ttnpb.Client {
	return u.Client
}

// Namespaces returns the namespaces clientWithFoo have extra attributes in.
func (u *clientWithFoo) Namespaces() []string {
	return []string{
		"foo",
	}
}

// Attributes returns for a given namespace a map containing the type extra attributes.
func (u *clientWithFoo) Attributes(namespace string) map[string]interface{} {
	if namespace != "foo" {
		return nil
	}

	return map[string]interface{}{
		"foo": u.Foo,
	}
}

// Fill fills an clientWithFoo type with the extra attributes that were found in the store.
func (u *clientWithFoo) Fill(namespace string, attributes map[string]interface{}) error {
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

func TestClientAttributer(t *testing.T) {
	a := assertions.New(t)

	schema := `
		CREATE TABLE IF NOT EXISTS foo_clients (
			id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			client_id   UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
			foo         STRING
		);
	`

	migrations.Registry.Register(migrations.Registry.Count()+1, "test_clients_attributer_schema", schema, "")
	s := testStore(t, attributesDatabase)
	s.Init() // to apply the migration

	specializer := func(base ttnpb.Client) store.Client {
		return &clientWithFoo{Client: &base}
	}

	base := *client
	base.ClientIdentifiers.ClientID = "attributer"

	withFoo := &clientWithFoo{
		Client: &base,
		Foo:    "bar",
	}

	err := s.Clients.Create(withFoo)
	a.So(err, should.BeNil)

	found, err := s.Clients.GetByID(withFoo.GetClient().ClientIdentifiers, specializer)
	a.So(err, should.BeNil)
	a.So(found, should.EqualFieldsWithIgnores(identityserver.ClientGeneratedFields...), withFoo)
	a.So(found.(*clientWithFoo).Foo, should.Equal, withFoo.Foo)

	err = s.Clients.Delete(withFoo.GetClient().ClientIdentifiers)
	a.So(err, should.BeNil)

	found, err = s.Clients.GetByID(withFoo.GetClient().ClientIdentifiers, specializer)
	a.So(err, should.NotBeNil)
	a.So(err, errshould.Describe, store.ErrClientNotFound)
	a.So(found, should.BeNil)
}
