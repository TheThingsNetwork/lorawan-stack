// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package rights

import "github.com/TheThingsNetwork/ttn/pkg/ttnpb"

// noopCache implements cache and caches nothing.
type noopCache struct{}

func (c *noopCache) GetOrFetch(key string, fetch func() ([]ttnpb.Right, error)) ([]ttnpb.Right, error) {
	return fetch()
}
