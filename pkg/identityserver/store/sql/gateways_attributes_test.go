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

// gatewayWithFoo implements both store.Gateway and store.Attributer interfaces.
type gatewayWithFoo struct {
	*ttnpb.Gateway
	Foo string
}

// GetGateway returns the DefaultGateway.
func (u *gatewayWithFoo) GetGateway() *ttnpb.Gateway {
	return u.Gateway
}

// Namespaces returns the namespaces gatewayWithFoo have extra attributes in.
func (u *gatewayWithFoo) Namespaces() []string {
	return []string{
		"foo",
	}
}

// Attributes returns for a given namespace a map containing the type extra attributes.
func (u *gatewayWithFoo) Attributes(namespace string) map[string]interface{} {
	if namespace != "foo" {
		return nil
	}

	return map[string]interface{}{
		"foo": u.Foo,
	}
}

// Fill fills an gatewayWithFoo type with the extra attributes that were found in the store.
func (u *gatewayWithFoo) Fill(namespace string, attributes map[string]interface{}) error {
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

func TestGatewayAttributer(t *testing.T) {
	a := assertions.New(t)

	schema := `
		CREATE TABLE IF NOT EXISTS foo_gateways (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			gateway_id   UUID NOT NULL REFERENCES gateways(id),
			foo          STRING
		);
	`

	migrations.Registry.Register(migrations.Registry.Count()+1, "test_gateways_attributer_schema", schema, "")
	s := testStore(t, attributesDatabase)
	s.MigrateAll()

	specializer := func(base ttnpb.Gateway) store.Gateway {
		return &gatewayWithFoo{Gateway: &base}
	}

	gateway := &ttnpb.Gateway{
		GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "attributer"},
		Description:        ".",
		Antennas:           []ttnpb.GatewayAntenna{},
		Radios:             []ttnpb.GatewayRadio{},
		Attributes:         make(map[string]string),
	}

	withFoo := &gatewayWithFoo{
		Gateway: gateway,
		Foo:     "bar",
	}

	err := s.Gateways.Create(withFoo)
	a.So(err, should.BeNil)

	found, err := s.Gateways.GetByID(withFoo.GetGateway().GatewayIdentifiers, specializer)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeGatewayIgnoringAutoFields, withFoo)
	a.So(found.(*gatewayWithFoo).Foo, should.Equal, withFoo.Foo)
}
