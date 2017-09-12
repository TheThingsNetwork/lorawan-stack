// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/factory"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
)

// ComponentStore is a store that holds network Components.
type ComponentStore interface {
	// Register creates a new Component and returns the new created Component.
	Register(component types.Component) (types.Component, error)

	// FindByID finds a Component ID and returns it.
	FindByID(componentID string) (types.Component, error)

	// FindByUser retrieves all the networks Components that an User is collaborator to.
	FindByUser(username string) ([]types.Component, error)

	// Edit updates the Component and returns the updated Component.
	Edit(component types.Component) (types.Component, error)

	// Delete deletes a Component and all its collaborators.
	Delete(componentID string) error

	// AddCollaborator adds a collaborator to a Component.
	AddCollaborator(componentID string, collaborator types.Collaborator) error

	// ListCollaborators returns the collaborators of a given Component.
	ListCollaborators(componentID string) ([]types.Collaborator, error)

	// RemoveCollaborator removes a collaborator from a Component.
	RemoveCollaborator(componentID string, username string) error

	// AddRight grants a given right to a given User.
	AddRight(componentID string, username string, right types.Right) error

	// ListUserRights returns the rights the User has for a Component.
	ListUserRights(componentID string, username string) ([]types.Right, error)

	// RemoveRight revokes a given right to a given collaborator.
	RemoveRight(componentID string, username string, right types.Right) error

	// LoadAttributes loads extra attributes into the Component if it's an Attributer.
	LoadAttributes(component types.Component) error

	// WriteAttributes writes the extra attributes on the Component if it's an
	// Attributer to the store.
	WriteAttributes(component, result types.Component) error

	// SetFactory allows to replace the DefaultComponent factory.
	SetFactory(factory factory.ComponentFactory)
}
