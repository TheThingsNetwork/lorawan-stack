// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import "github.com/TheThingsNetwork/ttn/pkg/ttnpb"

// Application is the interface of all things that can be an application. This can
// be used to build richer user types that can still be read and written to a database.
type Application interface {
	// GetApplication returns the ttnpb.Application that represents this
	// application.
	GetApplication() *ttnpb.Application
}

// ApplicationFactory is a function that returns a Application used to construct
// the results in read operations.
type ApplicationFactory func() Application

// ApplicationStore is a store that holds Applications.
type ApplicationStore interface {
	// Create creates a new application.
	Create(app Application) error

	// GetByID finds the application by ID and retrieves it.
	GetByID(appID string, factory ApplicationFactory) (Application, error)

	// ListByUser returns the applications to which an user is a collaborator.
	ListByUser(appID string, factory ApplicationFactory) ([]Application, error)

	// Update updates the application.
	Update(app Application) error

	// TODO(gomezjdaniel#274): use sql 'ON DELETE CASCADE' when CockroachDB implements it.
	// Delete deletes an application.
	Delete(appID string) error

	// SaveAPIKey stores an API Key attached to an application.
	SaveAPIKey(appID string, key *ttnpb.APIKey) error

	// GetAPIKey retrieves an API key by value and the appplication ID.
	GetAPIKey(key string) (string, *ttnpb.APIKey, error)

	// GetAPIKeyByName retrieves an API key from an application.
	GetAPIKeyByName(appID, keyName string) (*ttnpb.APIKey, error)

	// UpdateAPIKeyRights updates the right of an API key.
	UpdateAPIKeyRights(appID, keyName string, rights []ttnpb.Right) error

	// ListAPIKey list all the API keys that an application has.
	ListAPIKeys(appID string) ([]*ttnpb.APIKey, error)

	// DeleteAPIKey deletes a given API key from an application.
	DeleteAPIKey(appID, keyName string) error

	// SetCollaborator inserts or updates a collaborator within an application.
	// If the list of rights is empty the collaborator will be unset.
	SetCollaborator(collaborator *ttnpb.ApplicationCollaborator) error

	// HasUserRights checks whether an user has a set of given rights to an application.
	HasUserRights(appID, userID string, rights ...ttnpb.Right) (bool, error)

	// ListCollaborators retrieves all the collaborators from an application.
	// Optionally a list of rights can be passed to filter them.
	ListCollaborators(appID string, rights ...ttnpb.Right) ([]*ttnpb.ApplicationCollaborator, error)

	// ListUserRights returns the rights a given User has for an Application.
	ListUserRights(appID string, userID string) ([]ttnpb.Right, error)

	// LoadAttributes loads extra attributes into the Application.
	LoadAttributes(appID string, app Application) error

	// StoreAttributes writes the extra attributes on the Application if it is an
	// Attributer to the store.
	StoreAttributes(appID string, app, result Application) error
}
