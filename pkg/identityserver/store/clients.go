// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// ClientFactory is a function that returns a types.Client used to
// construct the results in read operations.
type ClientFactory func() types.Client

// ClientStore is a store that holds authorized third party Clients.
type ClientStore interface {
	// Create creates a new Client.
	Create(client types.Client) error

	// GetByID finds a client by ID and retrieves it.
	GetByID(clientID string, factory ClientFactory) (types.Client, error)

	// Update updates the client.
	Update(client types.Client) error

	// Archive sets the ArchivedAt field of the client to the current timestamp.
	Archive(clientID string) error

	// SetClientState allows to modify the reviewing state field of a client.
	SetClientState(clientID string, state ttnpb.ReviewingState) error

	// SetClientOfficial allows to set an unset a client as official labeled.
	SetClientOfficial(clientID string, official bool) error

	// LoadAttributes loads extra attributes into the Client if it's an Attributer.
	LoadAttributes(client types.Client) error

	// WriteAttributes writes the extra attributes on the Client if it's an
	// Attributer to the store.
	WriteAttributes(client, result types.Client) error
}
