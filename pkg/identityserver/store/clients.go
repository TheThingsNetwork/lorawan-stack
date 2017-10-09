// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/factory"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// ClientStore is a store that holds authorized third party Clients.
type ClientStore interface {
	// Create creates a new Client.
	Create(client types.Client) error

	// GetByID finds a client by ID and retrieves it.
	GetByID(clientID string) (types.Client, error)

	// ListByUser finds all the clients to which an user is collaborator to.
	ListByUser(username string) ([]types.Client, error)

	// Update updates the client.
	Update(client types.Client) error

	// Archive sets the ArchivedAt field of the client to the current timestamp.
	Archive(clientID string) error

	// SetClientState allows to modify the reviewing state field of a client.
	SetClientState(clientID string, state ttnpb.ClientState) error

	// SetClientOfficial allows to set an unset a client as official labeled.
	SetClientOfficial(clientID string, official bool) error

	// SetCollaborator inserts or updates a collaborator within a client.
	// If the list of rights is empty the collaborator will be unset.
	SetCollaborator(clientID string, collaborator ttnpb.Collaborator) error

	// ListCollaborators retrieves the collaborators for a given client.
	ListCollaborators(clientID string) ([]ttnpb.Collaborator, error)

	// ListUserRights returns the rights the user has for a client.
	ListUserRights(clientID string, username string) ([]ttnpb.Right, error)

	// LoadAttributes loads extra attributes into the Client if it's an Attributer.
	LoadAttributes(client types.Client) error

	// WriteAttributes writes the extra attributes on the Client if it's an
	// Attributer to the store.
	WriteAttributes(client, result types.Client) error

	// SetFactory allows to replace the default ttnpb.Client factory.
	SetFactory(factory factory.ClientFactory)
}
