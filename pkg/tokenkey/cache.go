// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package tokenkey

import "sync"

// Cache is the interface of a cache for tokenkeys.
type Cache interface {
	// Set sets the fetched public key in the cache (pem-encoded).
	Set(iss, kid, alg, pem string) error

	// Get returns the public key as a pem-encoded string.
	Get(iss, kid string) (pem string, alg string, err error)
}

// NilCache is the cache that never caches anything.
var NilCache = &nilCache{}

// nilCache does not do anything.
type nilCache struct{}

// Set implements Cache.
func (n *nilCache) Set(iss, kid, alg, pem string) error {
	return nil
}

// Get implements Cache.
func (n *nilCache) Get(iss, kid string) (string, string, error) {
	return "", "", nil
}

type entry struct {
	Alg string
	Key string
}

type MemoryCache struct {
	sync.RWMutex
	keys map[string]entry
}

func (m *MemoryCache) key(iss, kid string) string {
	return iss + "?kid=" + kid
}

// Set implements Cache.
func (m *MemoryCache) Set(iss, kid, alg, pem string) error {
	m.Lock()
	defer m.Unlock()

	if m.keys == nil {
		m.keys = make(map[string]entry, 1)
	}

	m.keys[m.key(iss, kid)] = entry{
		Alg: alg,
		Key: pem,
	}

	return nil
}

// Get implements Cache.
func (m *MemoryCache) Get(iss, kid string) (string, string, error) {
	m.RLock()
	defer m.RUnlock()

	entry, ok := m.keys[m.key(iss, kid)]
	if !ok {
		return "", "", nil
	}

	return entry.Key, entry.Alg, nil
}
