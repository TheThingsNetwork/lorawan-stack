// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package auth

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/TheThingsNetwork/ttn/pkg/tokenkey"
)

// Keys the key pairs used to sign and validate JWT tokens.
// It holds the current key pair as well as previous ones (to allow clients to validate older tokens).
type Keys struct {
	issuer  string
	keys    sync.Map
	current string
}

// NewKeys creates a new keys instance for the specified issuer.
func NewKeys(iss string) *Keys {
	return &Keys{
		issuer: iss,
	}
}

// Sign signs the claims with the current private key.
func (k *Keys) Sign(claims *Claims) (string, error) {
	_, key, err := k.GetCurrentPrivateKey()
	if err != nil {
		return "", err
	}

	return claims.Sign(key)
}

// GetPublicKey gets the public key for with the specified kid.
func (k *Keys) GetPublicKey(kid string) (crypto.PublicKey, error) {
	v, ok := k.keys.Load(kid)
	if !ok {
		return nil, fmt.Errorf("Could not find token key with kid `%s`", kid)
	}

	if err := checkPrivateKey(v); err != nil {
		panic(err)
	}

	switch key := v.(type) {
	case *rsa.PrivateKey:
		return &key.PublicKey, nil
	case *ecdsa.PrivateKey:
		return &key.PublicKey, nil
	}

	return nil, nil
}

// GetCurrentPublicKey returns the kid and public key of the currently active keypair.
func (k *Keys) GetCurrentPublicKey() (string, crypto.PublicKey, error) {
	key, err := k.GetPublicKey(k.current)
	return k.current, key, err
}

// GetPrivateKey gets the privatekey for the specified kid.
func (k *Keys) GetPrivateKey(kid string) (crypto.PrivateKey, error) {
	v, ok := k.keys.Load(kid)
	if !ok {
		return nil, fmt.Errorf("Could not find token key with kid `%s`", kid)
	}

	if err := checkPrivateKey(v); err != nil {
		panic(err)
	}

	return v, nil
}

// GetCurrentPrivateKey returns the kid and private key that is currently active.
func (k *Keys) GetCurrentPrivateKey() (string, crypto.PrivateKey, error) {
	key, err := k.GetPrivateKey(k.current)
	return k.current, key, err
}

// Rotate adds a new token private-public key pair and makes it the current keypair.
// The old keypair will be kept in memory to allow clients to validate older tokens with it.
func (k *Keys) Rotate(kid string, key crypto.PrivateKey) error {
	_, ok := k.keys.Load(kid)
	if ok {
		return fmt.Errorf("Token key with kid `%s` already exists", kid)
	}

	switch key.(type) {
	case *rsa.PrivateKey, *ecdsa.PrivateKey:
	default:
		return ErrUnsupportedSigningMethod
	}

	k.keys.Store(kid, key)
	k.current = kid

	return nil
}

// RotateFromPEM rotates in the new key from the content of a PEM-encoded private key.
func (k *Keys) RotateFromPEM(kid string, content []byte) error {
	block, _ := pem.Decode(content)
	if block == nil {
		return fmt.Errorf("Could not parse PEM")
	}

	var key crypto.PrivateKey
	var err error
	switch block.Type {
	case "RSA PRIVATE KEY":
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	case "EC PRIVATE KEY":
		key, err = x509.ParseECPrivateKey(block.Bytes)
	}
	if err != nil {
		return err
	}

	return k.Rotate(kid, key)
}

// RotateFromFile rotates in the new key from the content of a PEM-encoded private key file.
func (k *Keys) RotateFromFile(kid string, filename string) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return k.RotateFromPEM(kid, content)
}

// TokenKey implements tokenkey.Provider.
func (k *Keys) TokenKey(iss string, kid string) (crypto.PublicKey, error) {
	if iss != k.issuer {
		return nil, fmt.Errorf("Could not get public key for issuer %s", iss)
	}

	return k.GetPublicKey(kid)
}

// checkPrivateKey makes sure the argument is a private key and that it is supported by this package.
func checkPrivateKey(key crypto.PrivateKey) error {
	switch key.(type) {
	case *rsa.PrivateKey, *ecdsa.PrivateKey, *tokenkey.PrivateKeyWithKID:
		return nil
	}
	return fmt.Errorf("Expected type crypto.PrivateKey when loading the private key, but got %T", key)
}

// checkPublicKey  makes sure the argument is a public key and that it is supported by the keys.
func checkPublicKey(key crypto.PublicKey) error {
	switch key.(type) {
	case *rsa.PublicKey, *ecdsa.PublicKey:
		return nil
	}
	return fmt.Errorf("Expected type crypto.PublicKey when loading the public key, but got %T", key)
}
