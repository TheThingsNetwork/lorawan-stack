// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/factory"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
)

// UserStore is a store that holds Users.
type UserStore interface {
	// Create creates an user.
	Create(user types.User) error

	// GetByID finds the user by ID and retrieves it.
	GetByID(userID string) (types.User, error)

	// GetByEmail finds the user by email address and retrieves it.
	GetByEmail(email string) (types.User, error)

	// Update updates an user.
	Update(user types.User) error

	// Archive sets the ArchivedAt field of an user to the current timestamp.
	Archive(userID string) error

	// LoadAttributes loads all user attributes if the User is an Attributer.
	LoadAttributes(userID string, user types.User) error

	// WriteAttributes writes all of the user attributes if the User is an
	// Attributer and returns the written User in result.
	WriteAttributes(user, result types.User) error

	// SetFactory allows to replace the default ttnpb.User factory.
	SetFactory(factory factory.UserFactory)
}
