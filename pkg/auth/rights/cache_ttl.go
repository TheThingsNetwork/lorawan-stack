// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package rights

import (
	"context"
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
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
		return e.value, e.err
	}

	// Otherwise fetch.
	rights, err := fetch()
	c.entries[key] = &entry{
		createdAt: time.Now(),
		value:     rights,
		err:       err,
	}

	return rights, err
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
	err       error
}

func (e *entry) IsExpired(ttl time.Duration) bool {
	return e.createdAt.Add(ttl).Before(time.Now())
}
