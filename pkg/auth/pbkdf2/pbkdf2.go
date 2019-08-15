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

// Package pbkdf2 implements the PBKDF2 algorithm method used to hash passwords.
package pbkdf2

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/random"
	"golang.org/x/crypto/pbkdf2"
)

var defaultInstance = PBKDF2{
	Iterations: 20000,
	KeyLength:  32,
	Algorithm:  Sha256,
	SaltLength: 16,
}

// Default returns a the default PBKDF2 instance.
func Default() PBKDF2 { return defaultInstance }

// PBKDF2 is a password derivation method.
type PBKDF2 struct {
	// Iterations is the number of iterations to use in the PBKDF2 algorithm.
	Iterations int

	// Algorithm is the hashing algorithm used.
	Algorithm Algorithm

	// SaltLength is the length of the salt used.
	SaltLength int

	// KeyLength is the length of the desired key.
	KeyLength int
}

// Name returns the name of the PBKDF2 hashing method.
func (PBKDF2) Name() string {
	return "PBKDF2"
}

var errZeroLengthSalt = errors.DefineInternal(
	"zero_length_salt",
	"password salt can not have zero length",
)

// Hash hashes a plain text password.
func (p PBKDF2) Hash(plain string) (string, error) {
	if p.SaltLength == 0 {
		return "", errZeroLengthSalt
	}

	salt := random.String(p.SaltLength)
	hash := hash64([]byte(plain), []byte(salt), p.Iterations, p.KeyLength, p.Algorithm)
	pass := fmt.Sprintf("%s$%s$%v$%s$%s", p.Name(), p.Algorithm, p.Iterations, salt, string(hash))

	return pass, nil
}

var errInvalidPBKDF2 = errors.DefineInternal( // internal because hash is in DB.
	"invalid_pbkdf2_format",
	"password hash has invalid PBKDF2 format",
)

var errIterations = errors.DefineInternal(
	"pbkdf2_iterations",
	"could not determine number of iterations from `{iterations}`",
)

var errKeyLength = errors.DefineInternal(
	"pbkdf2_key_length",
	"could not determine key length",
)

// Validate validates a plaintext password against a hashed one.
// The format of the hashed password should be:
//
//     PBKDF2$<algorithm>$<iterations>$<salt>$<key in base64>
//
func (PBKDF2) Validate(hashed, plain string) (bool, error) {
	parts := strings.Split(hashed, "$")
	if len(parts) != 5 {
		return false, errInvalidPBKDF2
	}

	alg := parts[1]
	algorithm, err := parseAlgorithm(alg)
	if err != nil {
		return false, err
	}

	iter, err := strconv.ParseInt(parts[2], 10, 32)
	if err != nil {
		return false, errIterations.WithAttributes("iterations", parts[2])
	}
	salt := parts[3]
	key := parts[4]

	// Get the key length.
	keylen, err := keyLen(key)
	if err != nil {
		return false, errKeyLength.WithCause(err)
	}

	// Hash the plaintext.
	hash := hash64([]byte(plain), []byte(salt), int(iter), keylen, algorithm)

	// Compare the hashed plaintext and the stored hash.
	return subtle.ConstantTimeCompare(hash, []byte(key)) == 1, nil
}

// hash64 hashes a plain password and encodes it to base64.
func hash64(plain, salt []byte, iterations int, keyLength int, algorithm Algorithm) []byte {
	key := pbkdf2.Key(plain, salt, iterations, keyLength, algorithm.Hash)
	res := make([]byte, base64.RawURLEncoding.EncodedLen(len(key)))
	base64.RawURLEncoding.Encode(res, key)
	return res
}

// get the key length from the key.
func keyLen(key string) (int, error) {
	buf, err := base64.RawURLEncoding.DecodeString(key)
	if err != nil {
		return 0, err
	}
	return len(buf), nil
}
