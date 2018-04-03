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

// applicationWithFoo implements both store.Application and store.Attributer interfaces.
type applicationWithFoo struct {
	*ttnpb.Application
	Foo string
}

// GetApplication returns the DefaultApplication.
func (u *applicationWithFoo) GetApplication() *ttnpb.Application {
	return u.Application
}

// Namespaces returns the namespaces applicationWithFoo have extra attributes in.
func (u *applicationWithFoo) Namespaces() []string {
	return []string{
		"foo",
	}
}

// Attributes returns for a given namespace a map containing the type extra attributes.
func (u *applicationWithFoo) Attributes(namespace string) map[string]interface{} {
	if namespace != "foo" {
		return nil
	}

	return map[string]interface{}{
		"foo": u.Foo,
	}
}

// Fill fills an applicationWithFoo type with the extra attributes that were found in the store.
func (u *applicationWithFoo) Fill(namespace string, attributes map[string]interface{}) error {
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

func TestApplicationAttributer(t *testing.T) {
	a := assertions.New(t)

	schema := `
		CREATE TABLE IF NOT EXISTS foo_applications (
			id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			application_id   UUID NOT NULL REFERENCES applications(id),
			foo              STRING
		);
	`

	migrations.Registry.Register(migrations.Registry.Count()+1, "test_applications_attributer_schema", schema, "")
	s := testStore(t, attributesDatabase)
	s.MigrateAll()

	specializer := func(base ttnpb.Application) store.Application {
		return &applicationWithFoo{Application: &base}
	}

	application := &ttnpb.Application{
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "attributer"},
		Description:            ".",
	}

	withFoo := &applicationWithFoo{
		Application: application,
		Foo:         "bar",
	}

	err := s.Applications.Create(withFoo)
	a.So(err, should.BeNil)

	found, err := s.Applications.GetByID(withFoo.GetApplication().ApplicationIdentifiers, specializer)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeApplicationIgnoringAutoFields, withFoo)
	a.So(found.(*applicationWithFoo).Foo, should.Equal, withFoo.Foo)
}
