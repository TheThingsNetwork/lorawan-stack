// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// ValidationToken is an expirable token.
type ValidationToken struct {
	// ValidationToken is the token itself.
	ValidationToken string

	// CreatedAt denotes when the token was created.
	CreatedAt time.Time

	// ExpiresIn denotes the TTL of the token in seconds.
	ExpiresIn int32
}

// IsExpired checks whether the token is expired or not.
func (v ValidationToken) IsExpired() bool {
	return v.CreatedAt.Add(time.Duration(v.ExpiresIn) * time.Second).Before(time.Now())
}

// User is the interface of all things that can be an User. This can be used to
// build richer user types that can still be read and written to a database.
type User interface {
	// GetUser returns the ttnpb.User that represents this user.
	GetUser() *ttnpb.User
}

// UserSpecializer returns a new User with the given base ttnpb.User.
type UserSpecializer func(ttnpb.User) User

// UserStore is a store that holds Users.
type UserStore interface {
	// Create creates an user.
	Create(User) error

	// GetByID finds the user by ID and retrieves it.
	GetByID(ttnpb.UserIdentifiers, UserSpecializer) (User, error)

	// GetByEmail finds the user by email address and retrieves it.
	// TODO(gomezjdaniel): include email field as part of `UserIdentifiers`.
	GetByEmail(string, UserSpecializer) (User, error)

	// List returns all the users.
	List(UserSpecializer) ([]User, error)

	// Update updates an user.
	Update(User) error

	// TODO(gomezjdaniel#274): use sql 'ON DELETE CASCADE' when CockroachDB implements it.
	// Delete deletes an user.
	Delete(ttnpb.UserIdentifiers) error

	// SaveValidationToken saves the validation token.
	SaveValidationToken(ttnpb.UserIdentifiers, ValidationToken) error

	// GetValidationToken retrieves the validation token.
	GetValidationToken(string) (ttnpb.UserIdentifiers, *ValidationToken, error)

	// DeleteValidationToken deletes the validation token.
	DeleteValidationToken(string) error

	// SaveAPIKey stores an API Key attached to an user.
	SaveAPIKey(ttnpb.UserIdentifiers, ttnpb.APIKey) error

	// GetAPIKey retrieves an API key by value and the user ID.
	GetAPIKey(string) (ttnpb.UserIdentifiers, ttnpb.APIKey, error)

	// GetAPIKeyByName retrieves an API key from an user.
	GetAPIKeyByName(ttnpb.UserIdentifiers, string) (ttnpb.APIKey, error)

	// UpdateAPIKeyRights updates the right of an API key.
	UpdateAPIKeyRights(ttnpb.UserIdentifiers, string, []ttnpb.Right) error

	// ListAPIKey list all the API keys that an user has.
	ListAPIKeys(ttnpb.UserIdentifiers) ([]ttnpb.APIKey, error)

	// DeleteAPIKey deletes a given API key from an user.
	DeleteAPIKey(ttnpb.UserIdentifiers, string) error

	// LoadAttributes loads all user attributes if the User is an Attributer.
	LoadAttributes(ttnpb.UserIdentifiers, User) error

	// StoreAttributes writes all of the user attributes if the User is an
	// Attributer and returns the written User in result.
	StoreAttributes(ttnpb.UserIdentifiers, User) error
}
