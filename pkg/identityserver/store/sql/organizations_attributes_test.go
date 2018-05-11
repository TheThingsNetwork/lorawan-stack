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
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	. "go.thethings.network/lorawan-stack/pkg/identityserver/store/sql"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store/sql/migrations"
	"go.thethings.network/lorawan-stack/pkg/identityserver/test"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
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
			id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			organization_id   UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
			foo               STRING
		);
	`

	migrations.Registry.Register(migrations.Registry.Count()+1, "test_organizations_attributer_schema", schema, "")
	s := testStore(t, attributesDatabase)
	s.Init() // to apply the migration

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

	err = s.Organizations.Delete(withFoo.GetOrganization().OrganizationIdentifiers)
	a.So(err, should.BeNil)

	found, err = s.Organizations.GetByID(withFoo.GetOrganization().OrganizationIdentifiers, specializer)
	a.So(err, should.NotBeNil)
	a.So(ErrOrganizationNotFound.Describes(err), should.BeTrue)
	a.So(found, should.BeNil)
}
