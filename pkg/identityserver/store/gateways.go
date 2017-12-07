// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// GatewayFactory is a function that returns a types.Gateway used to
// construct the results in read operations.
type GatewayFactory func() types.Gateway

// GatewayStore is a store that holds Gateways.
type GatewayStore interface {
	// Create creates a new gateway.
	Create(gtw types.Gateway) error

	// GetByID finds a gateway by ID and retrieves it.
	GetByID(gtwID string, factory GatewayFactory) (types.Gateway, error)

	// ListByUser returns all the gateways to which an user is collaborator.
	ListByUser(userID string, factory GatewayFactory) ([]types.Gateway, error)

	// Update updates the gateway.
	Update(gtw types.Gateway) error

	// TODO(gomezjdaniel): wait for CockroachDB to introduce 'ON DELETE CASCADE'
	// 		-> https://github.com/cockroachdb/cockroach/issues/14848
	// Delete deletes a gateway.
	//Delete(gtwID string) error

	// Archive sets the ArchivedAt field of the gateway to the current timestamp.
	Archive(gtwID string) error

	// SetCollaborator inserts or updates a collaborator within a gateway.
	// If the list of rights is empty the collaborator will be unset.
	SetCollaborator(collaborator ttnpb.GatewayCollaborator) error

	// ListCollaborators retrieves all the gateway collaborators.
	ListCollaborators(gtwID string) ([]ttnpb.GatewayCollaborator, error)

	// ListUserRights returns the rights the user has for a gateway.
	ListUserRights(gtwID string, userID string) ([]ttnpb.Right, error)

	// LoadAttributes loads extra attributes into the gateway if it's an Attributer.
	LoadAttributes(gtwID string, gtw types.Gateway) error

	// StoreAttributes writes the extra attributes on the gatewat if it's an
	// Attributer to the store.
	StoreAttributes(gtwID string, gtw, res types.Gateway) error
}
