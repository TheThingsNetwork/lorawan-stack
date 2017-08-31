// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package auth

import (
	"crypto"
	"errors"
	"sync"
)

var (
	// ErrUnknownIdentityServer occurs when trying to get a tokenkey for an identity server that is unknown.
	ErrUnknownIdentityServer = errors.New("Unknown identity server")

	// ErrUnknownKID occurs when trying to get a tokenkey but the KID does not exist.
	ErrUnknownKID = errors.New("Unknown kid")
)

// TokenKeyProvider can get the public key of tokens, used for JWT validation.
type TokenKeyProvider interface {
	// Get returns the public key used by the given identity server. If kid is non-empty it will get the key
	// with that key id.
	Get(server string, kid string) (crypto.PublicKey, error)
}

// ConstProvider is a TokenKeyProvider that can return the token keys it was initialized with.
type ConstProvider struct {
	sync.Mutex
	Tokens map[string]map[string]crypto.PublicKey
}

// Get implements TokenKeyProvider.
func (p *ConstProvider) Get(server string, kid string) (crypto.PublicKey, error) {
	p.Lock()
	defer p.Unlock()

	keys := p.Tokens[server]
	if keys == nil {
		return nil, ErrUnknownIdentityServer
	}

	key := keys[kid]
	if key == nil {
		return nil, ErrUnknownKID
	}

	return key, nil
}
