// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package pbkdf2

import (
	"crypto/sha256"
	"crypto/sha512"
	"hash"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
)

// Algorithm is the type of the accepted algorithms.
type Algorithm string

const (
	// Sha256 is the sha256 algorithm.
	Sha256 Algorithm = "sha256"

	// Sha512 is the sha512 algorithm.
	Sha512 Algorithm = "sha512"
)

// Algorithm implements fmt.Stringer.
func (a Algorithm) String() string {
	return string(a)
}

// MarshalText implements encoding.TextMarshaler.
func (a Algorithm) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

// parseAlgorithm parses a string into an Algorithm and checks if it is supported.
func parseAlgorithm(str string) (Algorithm, error) {
	alg := Algorithm(str)
	switch alg {
	case Sha256, Sha512:
		return alg, nil
	default:
		return alg, errors.Errorf("Unsupported algorithm: %s", str)
	}
}

// Hash returns the hash.Hash that calculates the hash.
func (a *Algorithm) Hash() hash.Hash {
	if a == nil {
		return nil
	}

	switch *a {
	case Sha256:
		return sha256.New()
	case Sha512:
		return sha512.New()
	default:
		return nil
	}
}
