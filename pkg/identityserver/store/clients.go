// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import "github.com/TheThingsNetwork/ttn/pkg/identityserver/types"

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

	// TODO(gomezjdaniel): wait for CockroachDB to introduce 'ON DELETE CASCADE'
	// 		-> https://github.com/cockroachdb/cockroach/issues/14848
	// Delete deletes a client.
	//Delete(clientID string) error

	// Archive sets the ArchivedAt field of the client to the current timestamp.
	Archive(clientID string) error

	// LoadAttributes loads extra attributes into the Client if it's an Attributer.
	LoadAttributes(clientID string, client types.Client) error

	// WriteAttributes writes the extra attributes on the Client if it's an
	// Attributer to the store.
	WriteAttributes(clientID string, client, result types.Client) error
}
