// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// ApplicationFactory is a function that returns a types.Application used to
// construct the results in read operations.
type ApplicationFactory func() types.Application

// ApplicationStore is a store that holds Applications.
type ApplicationStore interface {
	// Create creates a new application.
	Create(app types.Application) error

	// GetByID finds the application by ID and retrieves it.
	GetByID(appID string, factory ApplicationFactory) (types.Application, error)

	// ListByUser returns the applications to which an user is a collaborator.
	ListByUser(userID string, factory ApplicationFactory) ([]types.Application, error)

	// Update updates the application.
	Update(app types.Application) error

	// TODO(gomezjdaniel): wait for CockroachDB to introduce 'ON DELETE CASCADE'
	// 		-> https://github.com/cockroachdb/cockroach/issues/14848
	// Delete deletes an application.
	//Delete(appID string) error

	// AddAPIKey adds a new application API key to a given application.
	AddAPIKey(appID string, key ttnpb.APIKey) error

	// RemoveAPIKey removes an application API key from a given application.
	RemoveAPIKey(appID string, keyName string) error

	// SetCollaborator inserts or updates a collaborator within an application.
	// If the list of rights is empty the collaborator will be unset.
	SetCollaborator(collaborator ttnpb.ApplicationCollaborator) error

	// ListCollaborators retrieves all the collaborators from an application.
	ListCollaborators(appID string) ([]ttnpb.ApplicationCollaborator, error)

	// ListUserRights returns the rights a given User has for an Application.
	ListUserRights(appID string, userID string) ([]ttnpb.Right, error)

	// LoadAttributes loads extra attributes into the Application.
	LoadAttributes(appID string, app types.Application) error

	// StoreAttributes writes the extra attributes on the Application if it is an
	// Attributer to the store.
	StoreAttributes(appID string, app, result types.Application) error
}
