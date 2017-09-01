// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package client

import (
	"crypto"
	"sync"

	"github.com/TheThingsNetwork/ttn/pkg/auth"
)

func cacheKey(iss, kid string) string {
	return iss + "^" + kid
}

// ProviderClient is a auth.TokenKeyProvider that fetches the keys over gRPC, like Client,
// but also caches the result.
type ProviderClient struct {
	sync.RWMutex
	client  *Client
	issuers []string
	cache   map[string]crypto.PublicKey
}

// NewClient returns a new TokenKeyProvider that uses gRPC to fetch token keys from a server.
func NewClient(issuers []string) *ProviderClient {
	return &ProviderClient{
		client:  NewSimpleClient(issuers),
		issuers: issuers,
		cache:   make(map[string]crypto.PublicKey, len(issuers)),
	}
}

// Get fetches the specified key from the specified issuer.
func (p *ProviderClient) Get(iss string, kid string) (crypto.PublicKey, error) {
	// try to get key from cache
	key := p.get(iss, kid)
	if key != nil {
		return key, nil
	}

	// getting key from cache failed, try and fetch it
	p.Lock()
	defer p.Unlock()

	key, err := p.client.Get(iss, kid)
	if err != nil {
		return nil, err
	}

	// store the key in cache
	p.set(iss, key)

	return key, nil
}

// update updates the keys from all issuers.
func (p *ProviderClient) update() error {
	p.Lock()
	defer p.Unlock()

	for _, iss := range p.issuers {
		_, err := p.Get(iss, "")
		if err != nil {
			return err
		}
	}

	return nil
}

// set stores the provided key under the issuer.
// Not safe for concurrent use (writes to cache).
func (p *ProviderClient) set(iss string, key crypto.PublicKey) {
	kid := auth.GetKID(key)
	p.cache[cacheKey(iss, kid)] = key
}

// get tries to fetch the key with the specified kid from the issuers cache, or returns nil if it does not exist.
func (p *ProviderClient) get(iss, kid string) crypto.PublicKey {
	p.RLock()
	defer p.RUnlock()
	return p.cache[cacheKey(iss, kid)]
}
