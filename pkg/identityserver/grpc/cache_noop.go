// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package grpc

import "github.com/TheThingsNetwork/ttn/pkg/ttnpb"

// noopCache implements Cache and caches nothing.
type noopCache struct{}

func (c *noopCache) GetOrFetch(auth, entityID string, fetch func() ([]ttnpb.Right, error)) ([]ttnpb.Right, error) {
	return fetch()
}
