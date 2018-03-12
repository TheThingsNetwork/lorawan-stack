// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package auth

import (
	"crypto/subtle"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/auth/pbkdf2"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
)

// hashingMethod is a method to hash a password.
type hashingMethod interface {
	// Name returns the hashing method name that is used to identify which method
	// was used to hash a given password.
	Name() string

	// Hash hashes the given plain text password.
	Hash(plain string) (string, error)

	// Validate checks whether the given plain text password is equal or not to
	// the given hashed password.
	Validate(hashed, plain string) (bool, error)
}

// hashingMethods contains all the supported hashing methods.
// Be sure to add your hashing method to this list if you implement a new one.
var hashingMethods = []hashingMethod{
	pbkdf2.Default(),
}

// defaultAlgorithm is the default algorithm used for hashing passwords.
var defaultAlgorithm = pbkdf2.Default()

// Password represents a hashed password.
type Password string

// Hash hashes a plaintext password into a Password.
func Hash(plain string) (Password, error) {
	str, err := defaultAlgorithm.Hash(plain)
	if err != nil {
		return "", err
	}

	return Password(str), nil
}

// Validate checks if the password matches the plaintext password.
// While using this over a secure channel is probably fine, consider using a
// scheme where the hashing happens on the client side, to prevent the server
// from having the password at all. You can use p.Equals to accomplish that.
func (p Password) Validate(plain string) (bool, error) {
	str := string(p)
	parts := strings.SplitN(str, "$", 2)

	if len(parts) < 2 {
		return false, errors.Errorf("Could not derive type from password hash: %s", str)
	}

	typ := parts[0]

	for _, method := range hashingMethods {
		if strings.ToLower(typ) == strings.ToLower(method.Name()) {
			return method.Validate(str, plain)
		}
	}

	return false, errors.Errorf("Got unexpected hash type: %s", typ)
}

// Equals safely checks whether or not the other hashed password and this one
// are the same. This can be used in schemes where the password is hashed at the
// client side and the hash is sent over instead of the plaintext password.
func (p Password) Equals(other Password) bool {
	return subtle.ConstantTimeEq(int32(len(other)), int32(len(p))) == 1 && subtle.ConstantTimeCompare([]byte(p), []byte(other)) == 1
}
