// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql_test

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/migrations"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
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
			client_id   UUID NOT NULL REFERENCES clients(id),
			foo         STRING
		);
	`

	migrations.Registry.Register(migrations.Registry.Count()+1, "test_clients_attributer_schema", schema, "")
	s := testStore(t, attributesDatabase)
	s.MigrateAll()

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
	a.So(found, test.ShouldBeClientIgnoringAutoFields, withFoo)
	a.So(found.(*clientWithFoo).Foo, should.Equal, withFoo.Foo)
}
