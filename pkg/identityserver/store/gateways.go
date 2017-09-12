// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/factory"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
)

// GatewayStore is a store that holds Gateways.
type GatewayStore interface {
	// Register creates a new Gateway and returns the new created Gateway.
	Register(gtw types.Gateway) (types.Gateway, error)

	// FindByID finds a Gateway by ID and retrieves it.
	FindByID(gtwID string) (types.Gateway, error)

	// FindByUser returns all the Gateways to which a given User is collaborator.
	FindByUser(username string) ([]types.Gateway, error)

	// Edit updates the Gateway and returns the updated Gateway.
	Edit(gtw types.Gateway) (types.Gateway, error)

	// Archive disables a Gateway.
	Archive(gtwID string) error

	// SetAttribute inserts or modifies an existing Gateway attribute.
	SetAttribute(gtwID, attribute, value string) error

	// ListAttributes returns all the Gateway attributes.
	ListAttributes(gtwID string) (map[string]string, error)

	// RemoveAttribute removes a specific Gateway attribute.
	RemoveAttribute(gtwID, attribute string) error

	// SetAntenna inserts or modifies an existing Gateway antenna.
	SetAntenna(gtwID string, antenna types.GatewayAntenna) error

	// ListAntennas returns all the registered antennas that belong to a certain Gateway.
	ListAntennas(gtwID string) ([]types.GatewayAntenna, error)

	// RemoveAntenna deletes an antenna from a gateway.
	RemoveAntenna(gtwID, antennaID string) error

	// AddCollaborator adds a collaborator to a gateway.
	AddCollaborator(gtwID string, collaborator types.Collaborator) error

	// ListCollaborators retrieves all the gateway collaborators.
	ListCollaborators(gtwID string) ([]types.Collaborator, error)

	// ListOwners retrieves all the owners of a gateway.
	ListOwners(gtwID string) ([]string, error)

	// RemoveCollaborator removes a collaborator from a gateway.
	RemoveCollaborator(gtwID string, username string) error

	// AddRight grants a given right to a given User.
	AddRight(gtwID string, username string, right types.Right) error

	// ListUserRights returns the rights the User has for a gateway.
	ListUserRights(gtwID string, username string) ([]types.Right, error)

	// RemoveRight revokes a given right from a given User.
	RemoveRight(gtwID string, username string, right types.Right) error

	// LoadAttributes loads extra attributes into the gateway if it's an Attributer.
	LoadAttributes(gtw types.Gateway) error

	// WriteAttributes writes the extra attributes on the gatewat if it's an
	// Attributer to the store.
	WriteAttributes(gtw, res types.Gateway) error

	// SetFactory allows to replace the DefaultGateway factory.
	SetFactory(factory factory.GatewayFactory)
}
