// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/factory"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
)

// ClientStore is a store that holds authorized third party Clients.
type ClientStore interface {
	// Register creates a new Client and returns the new created Client.
	Register(client types.Client) (types.Client, error)

	// FindByID finds a Client by ID and retrieves it.
	FindByID(clientID string) (types.Client, error)

	// FindByUser finds all the Clients an user is collaborator to.
	FindByUser(username string) ([]types.Client, error)

	// Edit updates the Client and returns the updated Client.
	Edit(client types.Client) (types.Client, error)

	// Delete deletes a Client.
	Delete(clientID string) error

	// Archive disables a Client.
	Archive(clientID string) error

	// Approve marks a Client approved by the tenant admins, so it can be used.
	Approve(clientID string) error

	// Reject marks a Client as rejected by the tenant admins, so it cannot be used anymore.
	Reject(clientID string) error

	// AddCollaborator adds a collaborator to a given Client.
	AddCollaborator(clientID string, collaborator types.Collaborator) error

	// ListCollaborators retrieves the collaborators for a given Client.
	ListCollaborators(clientID string) ([]types.Collaborator, error)

	// RemoveCollaborator removes a collaborator from a given Client.
	RemoveCollaborator(clientID string, username string) error

	// AddRight grants a given right to a given User for a Client.
	AddRight(clientID string, username string, right types.Right) error

	// ListUserRights returns the rights the user has for a Client.
	ListUserRights(clientID string, username string) ([]types.Right, error)

	// RemoveRight revokes a given right from a given Client collaborator.
	RemoveRight(clientID string, username string, right types.Right) error

	// LoadAttributes loads extra attributes into the Client if it's an Attributer.
	LoadAttributes(client types.Client) error

	// WriteAttributes writes the extra attributes on the Client if it's an
	// Attributer to the store.
	WriteAttributes(client, result types.Client) error

	// SetFactory allows to replace the DefaultClient factory.
	SetFactory(factory factory.ClientFactory)
}
