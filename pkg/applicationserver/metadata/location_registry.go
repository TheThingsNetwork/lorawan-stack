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

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/random"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/workerpool"
	"google.golang.org/grpc"
)

// EndDeviceLocationRegistry is a registry for end device locations.
type EndDeviceLocationRegistry interface {
	// Get retrieves the end device locations.
	Get(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (map[string]*ttnpb.Location, error)
	// Merge merges the end device locations.
	Merge(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, update map[string]*ttnpb.Location) (map[string]*ttnpb.Location, error)
}

type noopEndDeviceLocationRegistry struct{}

// Get implements EndDeviceLocationRegistry.
func (noopEndDeviceLocationRegistry) Get(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (map[string]*ttnpb.Location, error) {
	return nil, nil
}

// Merge implements EndDeviceLocationRegistry.
func (noopEndDeviceLocationRegistry) Merge(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, update map[string]*ttnpb.Location) (map[string]*ttnpb.Location, error) {
	return update, nil
}

// NewNoopEndDeviceLocationRegistry returns a noop EndDeviceLocationRegistry.
func NewNoopEndDeviceLocationRegistry() EndDeviceLocationRegistry {
	return noopEndDeviceLocationRegistry{}
}

type metricsEndDeviceLocationRegistry struct {
	inner EndDeviceLocationRegistry
}

// Get implements EndDeviceLocationRegistry.
func (m *metricsEndDeviceLocationRegistry) Get(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (map[string]*ttnpb.Location, error) {
	registerMetadataRegistryRetrieval(ctx, locationLabel)
	return m.inner.Get(ctx, ids)
}

// Merge implements EndDeviceLocationRegistry.
func (m *metricsEndDeviceLocationRegistry) Merge(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, update map[string]*ttnpb.Location) (map[string]*ttnpb.Location, error) {
	registerMetadataRegistryUpdate(ctx, locationLabel)
	return m.inner.Merge(ctx, ids, update)
}

// NewMetricsEndDeviceLocationRegistry returns an EndDeviceLocationRegistry that collects metrics.
func NewMetricsEndDeviceLocationRegistry(inner EndDeviceLocationRegistry) EndDeviceLocationRegistry {
	return &metricsEndDeviceLocationRegistry{
		inner: inner,
	}
}

// ClusterPeerAccess provides access to cluster peers.
type ClusterPeerAccess interface {
	GetPeerConn(ctx context.Context, role ttnpb.ClusterRole, ids cluster.EntityIdentifiers) (*grpc.ClientConn, error)
	WithClusterAuth() grpc.CallOption
}

var (
	endDeviceLocationPath      = []string{"locations"}
	endDeviceLocationFieldMask = &pbtypes.FieldMask{Paths: endDeviceLocationPath}
)

type clusterEndDeviceLocationRegistry struct {
	ClusterPeerAccess
	timeout time.Duration
}

// Get implements EndDeviceLocationRegistry.
func (c clusterEndDeviceLocationRegistry) Get(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (map[string]*ttnpb.Location, error) {
	cc, err := c.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	if err != nil {
		return nil, err
	}
	cl := ttnpb.NewEndDeviceRegistryClient(cc)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	dev, err := cl.Get(ctx, &ttnpb.GetEndDeviceRequest{
		EndDeviceIds: ids,
		FieldMask:    endDeviceLocationFieldMask,
	}, c.WithClusterAuth())
	if err != nil {
		return nil, err
	}
	return dev.Locations, nil
}

// Merge implements EndDeviceLocationRegistry.
func (c clusterEndDeviceLocationRegistry) Merge(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, update map[string]*ttnpb.Location) (map[string]*ttnpb.Location, error) {
	cc, err := c.GetPeerConn(ctx, ttnpb.ClusterRole_ENTITY_REGISTRY, nil)
	if err != nil {
		return nil, err
	}
	cl := ttnpb.NewEndDeviceRegistryClient(cc)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	dev, err := cl.Get(ctx, &ttnpb.GetEndDeviceRequest{
		EndDeviceIds: ids,
		FieldMask:    endDeviceLocationFieldMask,
	}, c.WithClusterAuth())
	if err != nil {
		return nil, err
	}
	if len(update) == 0 {
		return dev.Locations, nil
	}
	if dev.Locations == nil {
		dev.Locations = make(map[string]*ttnpb.Location, len(update))
	}
	for k, l := range update {
		dev.Locations[k] = l
	}
	_, err = cl.Update(ctx, &ttnpb.UpdateEndDeviceRequest{
		EndDevice: &ttnpb.EndDevice{
			Ids:       ids,
			Locations: dev.Locations,
		},
		FieldMask: endDeviceLocationFieldMask,
	}, c.WithClusterAuth())
	if err != nil {
		return nil, err
	}
	return dev.Locations, nil
}

// NewClusterEndDeviceLocationRegistry returns an EndDeviceLocationRegistry connected to the Entity Registry.
func NewClusterEndDeviceLocationRegistry(cluster ClusterPeerAccess, timeout time.Duration) EndDeviceLocationRegistry {
	return &clusterEndDeviceLocationRegistry{
		ClusterPeerAccess: cluster,
		timeout:           timeout,
	}
}

type cachedEndDeviceLocationRegistry struct {
	registry EndDeviceLocationRegistry
	cache    EndDeviceLocationCache

	minRefreshInterval time.Duration
	maxRefreshInterval time.Duration
	ttl                time.Duration

	replicationPool workerpool.WorkerPool
}

// Get implements EndDeviceLocationRegistry.
func (c *cachedEndDeviceLocationRegistry) Get(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers) (map[string]*ttnpb.Location, error) {
	locations, storedAt, err := c.cache.Get(ctx, ids)
	switch {
	case err != nil && !errors.IsNotFound(err):
		return nil, err
	case err != nil && errors.IsNotFound(err):
		locations = nil
	case err == nil:
		age := time.Since(*storedAt)
		if age <= c.minRefreshInterval {
			// If the object is younger than the minimum refresh interval, just return the cached value.
			return locations, nil
		}
		if remaining := c.maxRefreshInterval - age; remaining > 0 {
			// If the objects age is between the minimum and maximum refresh interval, check if we should asynchronously
			// refresh the cache.
			window := c.maxRefreshInterval - c.minRefreshInterval
			threshold := time.Duration(random.Int63n(int64(window)))
			// remaining is the remaining window of the refresh interval in the (0, window) interval.
			// threshold is a uniformly distributed duration in the [0, window) interval.
			if remaining >= threshold {
				return locations, nil
			}
		}
	}
	if err := c.replicationPool.Publish(ctx, ids); err != nil {
		log.FromContext(ctx).WithError(err).Warn("Failed to publish end device locations replication request")
	}
	return locations, nil
}

// Merge implements EndDeviceLocationRegistry.
func (c *cachedEndDeviceLocationRegistry) Merge(ctx context.Context, ids *ttnpb.EndDeviceIdentifiers, update map[string]*ttnpb.Location) (map[string]*ttnpb.Location, error) {
	locations, err := c.registry.Merge(ctx, ids, update)
	if err != nil {
		return nil, err
	}
	if err := c.cache.Set(ctx, ids, locations, c.ttl); err != nil {
		return nil, err
	}
	return locations, nil
}

// NewCachedEndDeviceLocationRegistry returns an EndDeviceLocationRegistry that caches the responses of the provided EndDeviceLocationRegistry in the provided
// EndDeviceLocationCache. On cache miss, the registry will retrieve and cache the locations asynchronously.
// Items whose TTL is within the soft TTL window have a chance to trigger an asynchronous cache synchronization event on location retrieval.
// The probability of a synchronization event increases linearly between the soft TTL (0%) and the hard TTL (100%).
func NewCachedEndDeviceLocationRegistry(ctx context.Context, c workerpool.Component, registry EndDeviceLocationRegistry, cache EndDeviceLocationCache, minRefreshInterval, maxRefreshInterval, ttl time.Duration) EndDeviceLocationRegistry {
	st := &cachedEndDeviceLocationRegistry{
		registry: registry,
		cache:    cache,

		minRefreshInterval: minRefreshInterval,
		maxRefreshInterval: maxRefreshInterval,
		ttl:                ttl,

		replicationPool: workerpool.NewWorkerPool(workerpool.Config{
			Component: c,
			Context:   ctx,
			Name:      "replicate_end_device_locations",
			Handler: func(ctx context.Context, item interface{}) {
				ids := item.(*ttnpb.EndDeviceIdentifiers)
				locations, err := registry.Get(ctx, ids)
				if err != nil {
					log.FromContext(ctx).WithError(err).Warn("Failed to retrieve end device locations")
					return
				}
				if err := cache.Set(ctx, ids, locations, ttl); err != nil {
					log.FromContext(ctx).WithError(err).Warn("Failed to cache end device locations")
					return
				}
			},
		}),
	}
	return st
}
