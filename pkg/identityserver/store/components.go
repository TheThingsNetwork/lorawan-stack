// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/factory"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
)

// ComponentStore is a store that holds network components.
type ComponentStore interface {
	// FindByID finds a component by its ID.
	FindByID(componentID string) (types.Component, error)

	// FindByUser retrieves all the networks components that an user is collaborator to.
	FindByUser(username string) ([]types.Component, error)

	// Create creates a new component.
	Create(component types.Component) (types.Component, error)

	// Update updates a component.
	Update(component types.Component) (types.Component, error)

	// Delete deletes a component and all its collaborators.
	Delete(componentID string) error

	// Collaborators returns the collaborators of a given component.
	Collaborators(componentID string) ([]types.Collaborator, error)

	// AddCollaborator adds a component collaborator.
	AddCollaborator(componentID string, collaborator types.Collaborator) error

	// GrantRight grants a given right to a given collaborator.
	GrantRight(componentID string, username string, right types.Right) error

	// RevokeRight revokes a given right to a given collaborator.
	RevokeRight(componentID string, username string, right types.Right) error

	// RemoveCollaborator removes a collaborator from a component.
	RemoveCollaborator(componentID string, username string) error

	// UserRights returns the rights the user has to a component.
	UserRights(componentID string, username string) ([]types.Right, error)

	// LoadAttributes loads extra attributes into the component if it's an Attributer.
	LoadAttributes(component types.Component) error

	// WriteAttributes writes the extra attributes on the component if it's an
	// Attributer to the store.
	WriteAttributes(component, result types.Component) error

	// SetFactory allows to replace the DefaultComponent factory.
	SetFactory(factory factory.ComponentFactory)
}
