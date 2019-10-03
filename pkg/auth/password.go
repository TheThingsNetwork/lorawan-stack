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

package auth

import (
	"context"
	"strings"

	"go.thethings.network/lorawan-stack/pkg/auth/pbkdf2"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

// HashValidator is a method to hash and validate a secret.
type HashValidator interface {
	// Name returns the hashing method name that is used to identify which method
	// was used to hash a given secret.
	Name() string

	// Hash hashes the given plain text secret.
	Hash(plain string) (string, error)

	// Validate checks whether the given plain text secret is equal or not to
	// the given hashed secret.
	Validate(hashed, plain string) (bool, error)
}

// defaultHashValidator is the default method for hashing and validating secrets.
var defaultHashValidator = pbkdf2.Default()

type hashValidatorContextKeyType struct{}

var hashValidatorContextKey hashValidatorContextKeyType

// NewContextWithHashValidator returns a context derived from parent that contains hashValidator.
func NewContextWithHashValidator(parent context.Context, hashValidator HashValidator) context.Context {
	return context.WithValue(parent, hashValidatorContextKey, hashValidator)
}

// HashValidatorFromContext returns the HashValidator from the context if present. Otherwise it returns default HashValidator.
func HashValidatorFromContext(ctx context.Context) HashValidator {
	if hashValidator, ok := ctx.Value(hashValidatorContextKey).(HashValidator); ok {
		return hashValidator
	}
	return defaultHashValidator
}

// hashValidators contains all the supported hashing methods.
// Be sure to add your hashing method to this list if you implement a new one.
var hashValidators = []HashValidator{
	defaultHashValidator,
}

// Hash hashes a plaintext secret.
func Hash(ctx context.Context, plain string) (string, error) {
	str, err := HashValidatorFromContext(ctx).Hash(plain)
	if err != nil {
		return "", err
	}
	return str, nil
}

var errInvalidHash = errors.DefineInternal(
	"invalid_hash",
	"invalid hash",
)

var errUnknownHashingMethod = errors.DefineInternal(
	"unknown_hashing_method",
	"unknown hashing method `{method}`",
)

// Validate checks if the hash matches the plaintext.
func Validate(hashed, plain string) (bool, error) {
	parts := strings.SplitN(hashed, "$", 2)

	if len(parts) < 2 {
		return false, errInvalidHash
	}

	typ := parts[0]

	for _, method := range hashValidators {
		if strings.ToLower(typ) == strings.ToLower(method.Name()) {
			return method.Validate(hashed, plain)
		}
	}

	return false, errUnknownHashingMethod.WithAttributes("method", typ)
}
