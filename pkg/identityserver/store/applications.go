// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/factory"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
)

// ApplicationStore is a store that holds applications
type ApplicationStore interface {
	// FindByID finds the application by its id
	FindByID(appID string) (types.Application, error)

	// FindByUser returns the applications to which an user is a collaborator
	FindByUser(username string) ([]types.Application, error)

	// Create creates a new application and returns it
	Create(app types.Application) (types.Application, error)

	// AddAppEUI adds a new application EUI to the given application
	AddAppEUI(appID string, eui types.AppEUI) error

	// DeleteAppEUI deletes an application EUI from a given application
	DeleteAppEUI(appID string, eui types.AppEUI) error

	// AddApplicationAPIKey adds a new application api key to the given application
	AddApplicationAPIKey(appID string, key types.ApplicationAPIKey) error

	// DeleteApplicationAPIKey deletes an application api key
	DeleteApplicationAPIKey(appID string, keyName string) error

	// Update updates the application and returns it updated
	Update(app types.Application) (types.Application, error)

	// Archive disables the application
	Archive(appID string) error

	// Collaborators retrieves the collaborators for an app
	Collaborators(appID string) ([]types.Collaborator, error)

	// AddCollaborator adds an application collaborator
	AddCollaborator(appID string, collaborator types.Collaborator) error

	// GrantRight grants a given right to a given collaborator
	GrantRight(appID string, username string, right types.Right) error

	// RevokeRight revokes a given right to a given collaborator
	RevokeRight(appID string, username string, right types.Right) error

	// RemoveCollaborator removes a collaborator from an app
	RemoveCollaborator(appID string, username string) error

	// UserRights returns the rights the user has to an application
	UserRights(appID string, username string) ([]types.Right, error)

	// LoadAttributes loads extra attributes into the application
	LoadAttributes(app types.Application) error

	// WriteAttributes writes the extra attributes on application if it is an
	// Attributer to the store
	WriteAttributes(app types.Application, result types.Application) error

	// SetFactory allows to replace the DefaultApplication factory
	SetFactory(factory factory.ApplicationFactory)
}
