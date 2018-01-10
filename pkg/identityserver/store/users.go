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

// UserFactory is a function that returns a User used to
// construct the results in read operations.
type UserFactory func() User

// UserStore is a store that holds Users.
type UserStore interface {
	// Create creates an user.
	Create(user User) error

	// GetByID finds the user by ID and retrieves it.
	GetByID(userID string, factory UserFactory) (User, error)

	// GetByEmail finds the user by email address and retrieves it.
	GetByEmail(email string, factory UserFactory) (User, error)

	// Update updates an user.
	Update(user User) error

	// TODO(gomezjdaniel#274): use sql 'ON DELETE CASCADE' when CockroachDB implements it.
	// Delete deletes an user.
	Delete(userID string) error

	// SaveValidationToken saves the validation token.
	SaveValidationToken(userID string, token *ValidationToken) error

	// GetValidationToken retrieves the validation token.
	GetValidationToken(token string) (string, *ValidationToken, error)

	// DeleteValidationToken deletes the validation token.
	DeleteValidationToken(token string) error

	// SaveAPIKey stores an API Key attached to an user.
	SaveAPIKey(userID string, key *ttnpb.APIKey) error

	// GetAPIKey retrieves an API key by value and the user ID.
	GetAPIKey(key string) (string, *ttnpb.APIKey, error)

	// GetAPIKeyByName retrieves an API key from an user.
	GetAPIKeyByName(userID, keyName string) (*ttnpb.APIKey, error)

	// UpdateAPIKeyRights updates the right of an API key.
	UpdateAPIKeyRights(userID, keyName string, rights []ttnpb.Right) error

	// ListAPIKey list all the API keys that an user has.
	ListAPIKeys(userID string) ([]*ttnpb.APIKey, error)

	// DeleteAPIKey deletes a given API key from an user.
	DeleteAPIKey(userID, keyName string) error

	// LoadAttributes loads all user attributes if the User is an Attributer.
	LoadAttributes(userID string, user User) error

	// StoreAttributes writes all of the user attributes if the User is an
	// Attributer and returns the written User in result.
	StoreAttributes(userID string, user, result User) error
}
