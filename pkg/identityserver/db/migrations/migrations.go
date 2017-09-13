// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package migrations

import "github.com/TheThingsNetwork/ttn/pkg/errors"

// Direction represents whether a migration is forwards or backwards.
type Direction string

const (
	// DirectionForwards represents a forwards type migration.
	DirectionForwards Direction = "forwards"

	// DirectionBackwards represents a backwards type migration.
	DirectionBackwards Direction = "backwards"
)

// Migration represents a database migration.
type Migration struct {
	Order     int
	Name      string
	Forwards  string
	Backwards string
}

// Registry is the type that holds all migrations indexed by its order.
type Registry map[int]*Migration

// NewRegistry builds a new registry.
func NewRegistry() Registry {
	return make(Registry)
}

// Register registers a new migration into the registry.
func (r Registry) Register(order int, name, forwards, backwards string) {
	if order < 1 {
		panic(errors.Errorf("Invalid migration order `%d` for migration `%s`. Order must be > 0", order, name))
	}

	if _, exists := r.Get(order); exists {
		panic(errors.Errorf("A migration with order `%d` already exists", order))
	}

	if _, exists := r.Get(order - 1); !exists && order != 1 {
		panic(errors.Errorf("Trying to register a migration with order `%v` but migration with order `%v` does not exist", order, order-1))
	}

	r[order] = &Migration{
		Order:     order,
		Name:      name,
		Forwards:  forwards,
		Backwards: backwards,
	}
}

// Get returns by order a migration of the registry and a bool indicating if it is exists.
func (r Registry) Get(order int) (*Migration, bool) {
	m, exists := r[order]
	return m, exists
}

// Count returns how many migrations are registered.
func (r Registry) Count() int {
	return len(r)
}
