// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package rights

import (
	"context"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// ttlCache is a cache where all entries have a TTL and are garbage collected.
type ttlCache struct {
	ctx     context.Context
	logger  log.Interface
	ttl     time.Duration
	mu      sync.RWMutex
	entries map[string]*entry
}

// newTTLCache creates a new instance of ttlCache and starts a goroutine for
// the garbage collector.
func newTTLCache(ctx context.Context, ttl time.Duration) *ttlCache {
	c := &ttlCache{
		ctx:     ctx,
		logger:  log.FromContext(ctx),
		ttl:     ttl,
		entries: make(map[string]*entry),
	}
	go c.garbageCollector()
	return c
}

// GetOrFetch returns the cached rights of the given key or calls `fetch` to load
// them if they cache entry is expired or not found.
func (c *ttlCache) GetOrFetch(key string, fetch func() ([]ttnpb.Right, error)) ([]ttnpb.Right, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.entries[key]
	if ok && !e.IsExpired(c.ttl) {
		return e.value, nil
	}

	// Otherwise fetch.
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
		select {
		case <-c.ctx.Done():
			c.logger.WithError(c.ctx.Err()).Info("TTL cache garbage collector has been stopped")
			return
		case <-ticker.C:
			c.mu.Lock()
			for key, e := range c.entries {
				if e.IsExpired(c.ttl) {
					delete(c.entries, key)
				}
			}
			c.mu.Unlock()
		}
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
