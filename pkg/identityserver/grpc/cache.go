// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package grpc

import "github.com/TheThingsNetwork/ttn/pkg/ttnpb"

// cache is the interface that describes a cache to store gRPC responses of the
// Identity Server `ListApplicationRights` and `ListGatewayRights` calls.
//
// The cache must be safe to be used concurrently.
type cache interface {
	// GetOrFetch gets the entry with the given key and loads it in the cache using
	// the given `fetch` method if it is not found.
	GetOrFetch(auth, entityID string, fetch func() ([]ttnpb.Right, error)) ([]ttnpb.Right, error)
}
