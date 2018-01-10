// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import "github.com/TheThingsNetwork/ttn/pkg/ttnpb"

// Gateway is the interface of all things that can be a gateway.
type Gateway interface {
	// GetGateway returns the ttnpb.Gateway that represents this gateway.
	GetGateway() *ttnpb.Gateway

	// SetAttributes sets the free-form attributes.
	SetAttributes(attributes map[string]string)

	// SetAntennas sets the antennas.
	SetAntennas(antennas []ttnpb.GatewayAntenna)
}

// GatewayFactory is a function that returns a Gateway used to construct the
// results in read operations.
type GatewayFactory func() Gateway

// GatewayStore is a store that holds Gateways.
type GatewayStore interface {
	// Create creates a new gateway.
	Create(gtw Gateway) error

	// GetByID finds a gateway by ID and retrieves it.
	GetByID(gtwID string, factory GatewayFactory) (Gateway, error)

	// ListByUser returns all the gateways to which an user is collaborator.
	ListByUser(userID string, factory GatewayFactory) ([]Gateway, error)

	// Update updates the gateway.
	Update(gtw Gateway) error

	// TODO(gomezjdaniel#274): use sql 'ON DELETE CASCADE' when CockroachDB implements it.
	// Delete deletes a gateway.
	Delete(gtwID string) error

	// SaveAPIKey stores an API Key attached to a gateway.
	SaveAPIKey(gtwID string, key *ttnpb.APIKey) error

	// GetAPIKey retrieves an API key by value and the gateway ID.
	GetAPIKey(key string) (string, *ttnpb.APIKey, error)

	// GetAPIKeyByName retrieves an API key from a gateway.
	GetAPIKeyByName(gtwID, keyName string) (*ttnpb.APIKey, error)

	// UpdateAPIKeyRights updates the right of an API key.
	UpdateAPIKeyRights(gtwID, keyName string, rights []ttnpb.Right) error

	// ListAPIKey list all the API keys that a gateway has.
	ListAPIKeys(gtwID string) ([]*ttnpb.APIKey, error)

	// DeleteAPIKey deletes a given API key from a gateway.
	DeleteAPIKey(gtwID, keyName string) error

	// SetCollaborator inserts or updates a collaborator within a gateway.
	// If the list of rights is empty the collaborator will be unset.
	SetCollaborator(collaborator *ttnpb.GatewayCollaborator) error

	// HasUserRights checks whether an user has a set of given rights to a gateway.
	HasUserRights(gtwID, userID string, rights ...ttnpb.Right) (bool, error)

	// ListCollaborators retrieves all the gateway collaborators.
	// Optionally a list of rights can be passed to filter them.
	ListCollaborators(gtwID string, rights ...ttnpb.Right) ([]*ttnpb.GatewayCollaborator, error)

	// ListUserRights returns the rights the user has for a gateway.
	ListUserRights(gtwID string, userID string) ([]ttnpb.Right, error)

	// LoadAttributes loads extra attributes into the gateway if it's an Attributer.
	LoadAttributes(gtwID string, gtw Gateway) error

	// StoreAttributes writes the extra attributes on the gatewat if it's an
	// Attributer to the store.
	StoreAttributes(gtwID string, gtw, res Gateway) error
}
