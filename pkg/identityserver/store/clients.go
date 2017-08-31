// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/factory"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
)

// ClientStore is a store that holds authorized third party clients
type ClientStore interface {
	// FindByID retrieves a client by its ID
	FindByID(clientID string) (types.Client, error)

	// FindByUser finds all the clients an user is creator to
	FindByUser(username string) ([]types.Client, error)

	// Create creates a new client and returns the created one
	Create(client types.Client) (types.Client, error)

	// Update updates a client and returns its updated version
	Update(client types.Client) (types.Client, error)

	// Delete deletes a client
	Delete(clientID string) error

	// Archive disables a client
	Archive(clientID string) error

	// Approve approves a client so it can be used
	Approve(clientID string) error

	// Reject rejects the client, meaning that it will can not be used anymore by users
	Reject(clientID string) error

	// Collaborators retrieves the collaborators for a client
	Collaborators(clientID string) ([]types.Collaborator, error)

	// AddCollaborator adds a collaborator to a client
	AddCollaborator(clientID string, collaborator types.Collaborator) error

	// GrantRight grants a given right to a given collaborator
	GrantRight(clientID string, username string, right types.Right) error

	// RevokeRight revokes a given right to a given collaborator
	RevokeRight(clientID string, username string, right types.Right) error

	// RemoveCollaborator removes a collaborator from a client
	RemoveCollaborator(clientID string, username string) error

	// UserRights returns the rights the user has to the client
	UserRights(clientID string, username string) ([]types.Right, error)

	// LoadAttributes loads extra attributes into the client if it's an Attributer
	LoadAttributes(client types.Client) error

	// WriteAttributes writes the extra attributes on the client if it's an
	// Attributer to the store
	WriteAttributes(client, result types.Client) error

	// SetFactory allows to replace the DefaultClient factory
	SetFactory(factory factory.ClientFactory)
}
