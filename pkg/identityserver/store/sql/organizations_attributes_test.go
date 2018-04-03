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

// organizationWithFoo implements both store.Organization and store.Attributer interfaces.
type organizationWithFoo struct {
	*ttnpb.Organization
	Foo string
}

// GetOrganization returns the DefaultOrganization.
func (u *organizationWithFoo) GetOrganization() *ttnpb.Organization {
	return u.Organization
}

// Namespaces returns the namespaces organizationWithFoo have extra attributes in.
func (u *organizationWithFoo) Namespaces() []string {
	return []string{
		"foo",
	}
}

// Attributes returns for a given namespace a map containing the type extra attributes.
func (u *organizationWithFoo) Attributes(namespace string) map[string]interface{} {
	if namespace != "foo" {
		return nil
	}

	return map[string]interface{}{
		"foo": u.Foo,
	}
}

// Fill fills an organizationWithFoo type with the extra attributes that were found in the store.
func (u *organizationWithFoo) Fill(namespace string, attributes map[string]interface{}) error {
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

func TestOrganizationAttributer(t *testing.T) {
	a := assertions.New(t)

	schema := `
		CREATE TABLE IF NOT EXISTS foo_organizations (
			id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			organization_id   UUID NOT NULL REFERENCES organizations(id),
			foo              STRING
		);
	`

	migrations.Registry.Register(migrations.Registry.Count()+1, "test_organizations_attributer_schema", schema, "")
	s := testStore(t, attributesDatabase)
	s.MigrateAll()

	specializer := func(base ttnpb.Organization) store.Organization {
		return &organizationWithFoo{Organization: &base}
	}

	organization := &ttnpb.Organization{
		OrganizationIdentifiers: ttnpb.OrganizationIdentifiers{OrganizationID: "attributer-organization"},
		Description:             ".",
	}

	withFoo := &organizationWithFoo{
		Organization: organization,
		Foo:          "bar",
	}

	err := s.Organizations.Create(withFoo)
	a.So(err, should.BeNil)

	found, err := s.Organizations.GetByID(withFoo.GetOrganization().OrganizationIdentifiers, specializer)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeOrganizationIgnoringAutoFields, withFoo)
	a.So(found.(*organizationWithFoo).Foo, should.Equal, withFoo.Foo)
}
