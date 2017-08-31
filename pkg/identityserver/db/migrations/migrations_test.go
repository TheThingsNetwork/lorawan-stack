// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package migrations

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestMigrationsRegistry(t *testing.T) {
	a := assertions.New(t)
	registry := NewRegistry()

	// check that registry is empty
	a.So(registry.Count(), should.BeZeroValue)

	// register one migration and retrieve it from registry after
	a.So(func() {
		registry.Register(1, "foo", "", "")
	}, should.NotPanic)

	migration, exists := registry.Get(1)
	a.So(migration, should.Resemble, &Migration{
		Order: 1,
		Name:  "foo",
	})
	a.So(exists, should.BeTrue)

	// check that indeed registry contains one migration only
	a.So(registry.Count(), should.Equal, 1)

	// register a migration with an order < 1 should panic
	a.So(func() {
		registry.Register(0, "foo", "", "")
	}, should.Panic)

	// register a migration with an order that is already registered should panic
	a.So(func() {
		registry.Register(1, "bar", "", "")
	}, should.Panic)

	// register a migration with an order that will result in a gap in the order
	// of the registy should panic
	a.So(func() {
		registry.Register(3, "foo", "", "")
	}, should.Panic)

	// retrieve a migration that does not exist
	migration, exists = registry.Get(3)
	a.So(migration, should.BeNil)
	a.So(exists, should.BeFalse)
}
