// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package tokenkey

import (
	"crypto"
)

// ConstProvider is a Provider that can only return the token keys it was initialized with.
type ConstProvider struct {
	// Tokens are the tokens this provider has access to.
	Tokens map[string]map[string]crypto.PublicKey
}

// TokenKey implements TokenKeyProvider.
func (p *ConstProvider) TokenKey(server string, kid string) (crypto.PublicKey, error) {
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
