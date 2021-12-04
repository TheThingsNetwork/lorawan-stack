// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package metadata

import (
	"context"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// EndDeviceLocationCache is a cache for end device locations.
type EndDeviceLocationCache interface {
	// Get retrieves the end device locations and the remaining TTL for the entry.
	Get(ctx context.Context, ids ttnpb.EndDeviceIdentifiers) (map[string]*ttnpb.Location, time.Duration, error)
	// SetLocations sets the end device locations.
	SetLocations(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, update map[string]*ttnpb.Location, ttl time.Duration) error
	// SetErrorDetails sets the the end device locations error details.
	SetErrorDetails(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, details *ttnpb.ErrorDetails, ttl time.Duration) error
	// Delete removes the locations from the cache.
	Delete(ctx context.Context, ids ttnpb.EndDeviceIdentifiers) error
}

type metricsEndDeviceLocationCache struct {
	inner EndDeviceLocationCache
}

// Get implements EndDeviceLocationCache.
func (c *metricsEndDeviceLocationCache) Get(ctx context.Context, ids ttnpb.EndDeviceIdentifiers) (map[string]*ttnpb.Location, time.Duration, error) {
	m, ttl, err := c.inner.Get(ctx, ids)
	if ttl == 0 {
		registerMetadataCacheMiss(ctx, locationLabel)
	} else {
		registerMetadataCacheHit(ctx, locationLabel)
	}
	return m, ttl, err
}

// SetLocations implements EndDeviceLocationCache.
func (c *metricsEndDeviceLocationCache) SetLocations(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, update map[string]*ttnpb.Location, ttl time.Duration) error {
	return c.inner.SetLocations(ctx, ids, update, ttl)
}

// SetError implements EndDEviceLocationCache.
func (c *metricsEndDeviceLocationCache) SetErrorDetails(ctx context.Context, ids ttnpb.EndDeviceIdentifiers, details *ttnpb.ErrorDetails, ttl time.Duration) error {
	return c.inner.SetErrorDetails(ctx, ids, details, ttl)
}

// Delete implements EndDeviceLocationCache.
func (c *metricsEndDeviceLocationCache) Delete(ctx context.Context, ids ttnpb.EndDeviceIdentifiers) error {
	return c.inner.Delete(ctx, ids)
}

// NewMetricsEndDeviceLocationCache constructs an EndDeviceLocationCache that collects metrics.
func NewMetricsEndDeviceLocationCache(inner EndDeviceLocationCache) EndDeviceLocationCache {
	return &metricsEndDeviceLocationCache{
		inner: inner,
	}
}
