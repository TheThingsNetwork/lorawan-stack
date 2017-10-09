// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

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
