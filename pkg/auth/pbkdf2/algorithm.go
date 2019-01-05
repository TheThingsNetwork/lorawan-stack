// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package pbkdf2

import (
	"crypto/sha256"
	"crypto/sha512"
	"hash"

	"go.thethings.network/lorawan-stack/pkg/errors"
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

var errUnknownAlgorithm = errors.DefineInternal(
	"unknown_algorithm",
	"algorithm `{alg}` unknown",
)

// parseAlgorithm parses a string into an Algorithm and checks if it is supported.
func parseAlgorithm(str string) (Algorithm, error) {
	alg := Algorithm(str)
	switch alg {
	case Sha256, Sha512:
		return alg, nil
	default:
		return alg, errUnknownAlgorithm.WithAttributes("alg", alg.String())
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
