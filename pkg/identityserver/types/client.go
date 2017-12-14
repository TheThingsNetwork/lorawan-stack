// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

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
