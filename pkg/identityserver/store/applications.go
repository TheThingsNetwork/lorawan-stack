// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/factory"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
)

// ApplicationStore is a store that holds Applications.
type ApplicationStore interface {
	// Register creates a new Application and returns the new created Application.
	Register(app types.Application) (types.Application, error)

	// FindByID finds the Application by ID and retrieves it.
	FindByID(appID string) (types.Application, error)

	// FindByUser returns the Applications to which an User is a collaborator.
	FindByUser(username string) ([]types.Application, error)

	// Edit updates the Application and returns the updated Application.
	Edit(app types.Application) (types.Application, error)

	// Archive disables the Application.
	Archive(appID string) error

	// AddAppEUI adds a new AppEUI to a given Application.
	AddAppEUI(appID string, eui types.AppEUI) error

	// ListAppEUIs returns all the AppEUIs that belong to a given Application.
	ListAppEUIs(appID string) ([]types.AppEUI, error)

	// RemoveAppEUI remove an AppEUI from a given Application.
	RemoveAppEUI(appID string, eui types.AppEUI) error

	// AddAPIKey adds a new Application API key to a given Application.
	AddAPIKey(appID string, key types.ApplicationAPIKey) error

	// ListAPIKeys returns all the registered application API keys that
	// belong to a given Application.
	ListAPIKeys(appID string) ([]types.ApplicationAPIKey, error)

	// RemoveAPIKey removes an Application API key from a given Application.
	RemoveAPIKey(appID string, keyName string) error

	// AddCollaborator adds an Application collaborator.
	AddCollaborator(appID string, collaborator types.Collaborator) error

	// ListCollaborators retrieves all the collaborators from an Application.
	ListCollaborators(appID string) ([]types.Collaborator, error)

	// RemoveCollaborator removes a collaborator from an Application.
	RemoveCollaborator(appID string, username string) error

	// AddRight grants a given right to a given User.
	AddRight(appID string, username string, right types.Right) error

	// ListUserRights returns the rights a given User has for an Application.
	ListUserRights(appID string, username string) ([]types.Right, error)

	// RemoveRight revokes a given right to a given collaborator.
	RemoveRight(appID string, username string, right types.Right) error

	// LoadAttributes loads extra attributes into the Application.
	LoadAttributes(app types.Application) error

	// WriteAttributes writes the extra attributes on the Application if it is an
	// Attributer to the store.
	WriteAttributes(app types.Application, result types.Application) error

	// SetFactory allows to replace the DefaultApplication factory.
	SetFactory(factory factory.ApplicationFactory)
}
