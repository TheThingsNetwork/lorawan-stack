// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package frequencyplans

import (
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

const DefaultCacheExpiry = time.Hour

type cacheEntry struct {
	fp      ttnpb.FrequencyPlan
	err     error
	lastHit time.Time
}

type cache struct {
	s Store

	fps    map[string]cacheEntry
	expiry time.Duration

	idsCache   []string
	idsLastHit time.Time

	mu sync.Mutex
}

// Cache wraps the given store with a cache, so that all returned frequency plans are cached for `expiry`.
func Cache(s Store, expiry time.Duration) Store {
	return &cache{
		s:      s,
		mu:     sync.Mutex{},
		fps:    make(map[string]cacheEntry),
		expiry: expiry,
	}
}

func (c *cache) GetAllIDs() []string {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.idsLastHit.After(time.Now().Add(-1 * c.expiry)) {
		idsCache := c.idsCache
		return idsCache
	}

	ids := c.s.GetAllIDs()
	c.idsCache = ids
	c.idsLastHit = time.Now()
	return ids
}

func (c *cache) GetByID(id string) (ttnpb.FrequencyPlan, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, hit := c.fps[id]
	if hit && entry.lastHit.After(time.Now().Add(-1*c.expiry)) {
		fp := entry.fp
		err := entry.err
		return fp, err
	}

	fp, err := c.s.GetByID(id)
	c.fps[id] = cacheEntry{
		fp:      fp,
		err:     err,
		lastHit: time.Now(),
	}
	return fp, err
}
