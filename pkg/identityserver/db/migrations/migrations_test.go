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

package migrations

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestMigrationsRegistry(t *testing.T) {
	a := assertions.New(t)
	registry := NewRegistry()

	// Check that registry is empty.
	a.So(registry.Count(), should.BeZeroValue)

	// Register one migration and retrieve it from registry after.
	a.So(func() {
		registry.Register(1, "foo", "", "")
	}, should.NotPanic)

	migration, exists := registry.Get(1)
	a.So(migration, should.Resemble, &Migration{
		Order: 1,
		Name:  "foo",
	})
	a.So(exists, should.BeTrue)

	// Check that indeed registry contains one migration only.
	a.So(registry.Count(), should.Equal, 1)

	// Register a migration with an order < 1 should panic.
	a.So(func() {
		registry.Register(0, "foo", "", "")
	}, should.Panic)

	// Register a migration with an order that is already registered should panic.
	a.So(func() {
		registry.Register(1, "bar", "", "")
	}, should.Panic)

	// Register a migration with an order that will result in a gap in the order.
	// of the registry should panic
	a.So(func() {
		registry.Register(3, "foo", "", "")
	}, should.Panic)

	// Retrieve a migration that does not exist.
	migration, exists = registry.Get(3)
	a.So(migration, should.BeNil)
	a.So(exists, should.BeFalse)
}
