// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/factory"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
)

// GatewayStore is a store that holds gateways.
type GatewayStore interface {
	// FindByID retreives the Gateway by id.
	FindByID(gtwID string) (types.Gateway, error)

	// FindByUser returns the Gateways to which a user is a collaborator.
	FindByUser(username string) ([]types.Gateway, error)

	// CreateGateway creates a new Gateway.
	Create(gtw types.Gateway) (types.Gateway, error)

	// Updateupdates the Gateway.
	Update(gtw types.Gateway) (types.Gateway, error)

	// Archive disables the Gateway.
	Archive(gtwID string) error

	// Owners retrieves the owners for a gateway.
	Owners(gtwID string) ([]string, error)

	// Antennas fetches all the registered antennas that belongs to a certain gateway.
	Antennas(gtwID string) ([]types.GatewayAntenna, error)

	// UpsertAntenna inserts or modifies an antenna to a certain gateway.
	UpsertAntenna(gtwID string, antenna types.GatewayAntenna) error

	// DeleteAntenna deletes an antenna from a gateway.
	DeleteAntenna(gtwID, antennaID string) error

	// Collaborators retrieves the collaborators for a gateway.
	Collaborators(gtwID string) ([]types.Collaborator, error)

	// AddCollaborator adds a collaborator to a gateway.
	AddCollaborator(gtwID string, collaborator types.Collaborator) error

	// GrantRight grants a given right to a given collaborator.
	GrantRight(gtwID string, username string, right types.Right) error

	// RevokeRight revokes a given right to a given collaborator.
	RevokeRight(gtwID string, username string, right types.Right) error

	// RemoveCollaborator removes a collaborator from a gateway.
	RemoveCollaborator(gtwID string, username string) error

	// UserRights returns the rights the user has to the Gateway.
	UserRights(gtwID string, username string) ([]types.Right, error)

	// LoadAttributes loads extra attributes into the gateway if it's an Attributer.
	LoadAttributes(gtw types.Gateway) error

	// WriteAttributes writes the extra attributes on the gatewat if it's an
	// Attributer to the store.
	WriteAttributes(gtw, res types.Gateway) error

	// SetFactory allows to replace the DefaultGateway factory.
	SetFactory(factory factory.GatewayFactory)
}
