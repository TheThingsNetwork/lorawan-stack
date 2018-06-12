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
			gateway_id   UUID NOT NULL REFERENCES gateways(id) ON DELETE CASCADE,
			foo          VARCHAR
		);
	`

	migrations.Registry.Register(migrations.Registry.Count()+1, "test_gateways_attributer_schema", schema, "")
	s := testStore(t, attributesDatabase)
	s.Init() // to apply the migration

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
	a.So(found, should.EqualFieldsWithIgnores(identityserver.GatewayGeneratedFields...), withFoo)
	a.So(found.(*gatewayWithFoo).Foo, should.Equal, withFoo.Foo)

	err = s.Gateways.Delete(withFoo.GetGateway().GatewayIdentifiers)
	a.So(err, should.BeNil)

	found, err = s.Gateways.GetByID(withFoo.GetGateway().GatewayIdentifiers, specializer)
	a.So(err, should.NotBeNil)
	a.So(err, should.DescribeError, store.ErrGatewayNotFound)
	a.So(found, should.BeNil)
}
