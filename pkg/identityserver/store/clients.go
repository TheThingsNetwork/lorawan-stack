// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

// ClientSpecializer returns a new Client with the given base ttnpb.Client.
type ClientSpecializer func(ttnpb.Client) Client

// ClientStore is a store that holds authorized third party Clients.
type ClientStore interface {
	// Create creates a new Client.
	Create(Client) error

	// GetByID finds a client by ID and retrieves it.
	GetByID(ttnpb.ClientIdentifiers, ClientSpecializer) (Client, error)

	// List list all the clients.
	List(ClientSpecializer) ([]Client, error)

	// ListByUser returns all the clients created by the user.
	ListByUser(ttnpb.UserIdentifiers, ClientSpecializer) ([]Client, error)

	// Update updates the client.
	Update(Client) error

	// Delete deletes a client.
	Delete(ttnpb.ClientIdentifiers) error

	// LoadAttributes loads extra attributes into the Client if it's an Attributer.
	LoadAttributes(ttnpb.ClientIdentifiers, Client) error

	// StoreAttributes writes the extra attributes on the Client if it's an
	// Attributer to the store.
	StoreAttributes(ttnpb.ClientIdentifiers, Client) error
}
