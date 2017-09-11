// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/factory"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
)

// UserStore is a store that holds users.
type UserStore interface {
	// FindByUsername finds the user by his username.
	FindByUsername(username string) (types.User, error)

	// FindByEmail finds the user by email address.
	FindByEmail(email string) (types.User, error)

	// Create creates an user.
	Create(user types.User) (types.User, error)

	// Update updates the user's profile.
	Update(user types.User) (types.User, error)

	// Archive disables an user.
	Archive(username string) error

	// LoadAttributes loads all user attributes if the user is an Attributer.
	LoadAttributes(username string, user types.User) error

	// WriteAttributes writes all of the user attributes if the user is an
	// Attributer and returns the written user in result.
	WriteAttributes(user, result types.User) error

	// SetFactory allows to replace the DefaultUser factory.
	SetFactory(factory factory.UserFactory)
}
