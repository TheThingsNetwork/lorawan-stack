// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package tokenkey

import (
	"crypto"
	"errors"
)

var (
	// ErrUnknownIdentityServer occurs when trying to get a tokenkey for an identity server that is unknown.
	ErrUnknownIdentityServer = errors.New("Unknown identity server")

	// ErrUnknownKID occurs when trying to get a tokenkey but the KID does not exist.
	ErrUnknownKID = errors.New("Unknown kid")
)

// Provider is the interface of things that can provide a token public key.
type Provider interface {
	// TokenKey returns the public key used by the given issuer. If kid is non-empty it will get the key
	// with that key id.
	TokenKey(issuer string, kid string) (crypto.PublicKey, error)
}
