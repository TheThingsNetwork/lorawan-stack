// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/factory"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
)

// UserStore is a store that holds Users.
type UserStore interface {
	// Register creates an User and returns the new created User.
	Register(user types.User) (types.User, error)

	// FindByUsername finds the User by username and returns it.
	FindByUsername(username string) (types.User, error)

	// FindByEmail finds the User by email address and returns it.
	FindByEmail(email string) (types.User, error)

	// Edit updates an User and returns the updated User.
	Edit(user types.User) (types.User, error)

	// Archive disables an User.
	Archive(username string) error

	// LoadAttributes loads all user attributes if the User is an Attributer.
	LoadAttributes(username string, user types.User) error

	// WriteAttributes writes all of the user attributes if the User is an
	// Attributer and returns the written User in result.
	WriteAttributes(user, result types.User) error

	// SetFactory allows to replace the DefaultUser factory.
	SetFactory(factory factory.UserFactory)
}
