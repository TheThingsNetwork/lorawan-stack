// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"github.com/RangelReale/osin"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// Client is the interface of all things that can be a third-party client.
type Client interface {
	osin.Client

	osin.ClientSecretMatcher

	// GetClient returns the ttnpb.Client that represents this client.
	GetClient() *ttnpb.Client
}

// ClientFactory is a function that returns a Client used to construct the results
// in read operations.
type ClientFactory func() Client

// ClientStore is a store that holds authorized third party Clients.
type ClientStore interface {
	// Create creates a new Client.
	Create(client Client) error

	// GetByID finds a client by ID and retrieves it.
	GetByID(clientID string, factory ClientFactory) (Client, error)

	// List list all the clients.
	List(factory ClientFactory) ([]Client, error)

	// ListByUser returns all the clients created by the user.
	ListByUser(userID string, factory ClientFactory) ([]Client, error)

	// Update updates the client.
	Update(client Client) error

	// TODO(gomezjdaniel#274): use sql 'ON DELETE CASCADE' when CockroachDB implements it.
	// Delete deletes a client.
	Delete(clientID string) error

	// LoadAttributes loads extra attributes into the Client if it's an Attributer.
	LoadAttributes(clientID string, client Client) error

	// StoreAttributes writes the extra attributes on the Client if it's an
	// Attributer to the store.
	StoreAttributes(clientID string, client, result Client) error
}
