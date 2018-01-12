// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package grpc

import (
	"fmt"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// Cache is the interface that describes a cache to store gRPC responses
// of the ListApplicationRights and ListGatewayRights methods.
type Cache interface {
	// ApplicationKey returns the key to cache a `ListApplicationRightsResponse`.
	ApplicationKey(appID string) string

	// GatewayKey returns the key to cache a `ListGatewayRightsResponse`.
	GatewayKey(gtwID string) string

	// GetOrFetch gets the entry with the given key and reloads it if it's expired.
	// It is safe to be used concurrently.
	GetOrFetch(key string, fetch func() ([]ttnpb.Right, error)) ([]ttnpb.Right, error)
}

// noopCache implements Cache and does nothing.
type noopCache struct{}

func (c *noopCache) ApplicationKey(appID string) string { return "" }
func (c *noopCache) GatewayKey(gtwID string) string     { return "" }
func (c *noopCache) GetOrFetch(key string, fetch func() ([]ttnpb.Right, error)) []ttnpb.Right {
	return nil, nil
}

// TTLCache is a cache where all entries have a TTL.
type TTLCache struct {
	mu      sync.RWMutex
	ttl     time.Duration
	entries map[string]*entry
}

// NewTTLCache returns a TTLCache.
func NewTTLCache(ttl time.Duration) *TTLCache {
	return &TTLCache{
		ttl:     ttl,
		entries: make(map[string]*entry),
	}
}

func (c *TTLCache) ApplicationKey(appID string) string {
	return fmt.Sprintf("application:%s", appID)
}

func (c *TTLCache) GatewayKey(gtwID string) string {
	return fmt.Sprintf("gateway:%s", gtwID)
}

func (c *TTLCachee) GetOrFetch(key string, fetch func() ([]ttnpb.Right, error)) ([]ttnpb.Right, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok {
		goto Fetch
	}
	if e.IsExpired(c.ttl) {
		goto Fetch
	}
	return e.value, nil

Fetch:
	rights, err := fetch()
	if err != nil {
		return nil, err
	}
	c.entries[key] = &entry{
		createdAt: time.Now(),
		value:     rights,
	}
	return rights, nil
}

// entry is a TTLCache entry.
type entry struct {
	createdAt time.Time
	value     []ttnpb.Right
}

func (e *entry) IsExpired(ttl time.Duration) bool {
	return e.createdAt.Add(ttl).After(time.Now())
}
