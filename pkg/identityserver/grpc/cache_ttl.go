// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package grpc

import (
	"fmt"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// ttlCache is a cache where all entries have a TTL and are garbage collected.
type ttlCache struct {
	mu      sync.RWMutex
	ttl     time.Duration
	entries map[string]*entry
}

// newTTLCache creates a new instance of ttlCache.
func newTTLCache(ttl time.Duration) *ttlCache {
	c := &ttlCache{
		ttl:     ttl,
		entries: make(map[string]*entry),
	}
	go c.garbageCollector()
	return c
}

// GetOrFetch returns the rights of the given key or uses the given `fetch` function
// to loads them in the cache and returns it if the key is expired or not found.
func (c *ttlCache) GetOrFetch(auth, entityID string, fetch func() ([]ttnpb.Right, error)) ([]ttnpb.Right, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := fmt.Sprintf("%s:%s", auth, entityID)

	e, ok := c.entries[key]
	if ok && !e.IsExpired(c.ttl) {
		return e.value, nil
	}

	// otherwise fetch
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

// garbageCollector is the method executed in a goroutine in charge of delete
// those expired entries in the cache.
func (c *ttlCache) garbageCollector() {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()
	for {
		<-ticker.C
		c.mu.Lock()
		for key, e := range c.entries {
			if e.IsExpired(c.ttl) {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}

// entry is a ttlCache entry.
type entry struct {
	createdAt time.Time
	value     []ttnpb.Right
}

func (e *entry) IsExpired(ttl time.Duration) bool {
	return e.createdAt.Add(ttl).Before(time.Now())
}
