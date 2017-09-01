// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package auth

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sync"
)

// Manager is a tokenkey Manager that manages the keys used to sign the jwt tokens.
// It manages the current keypair as well as previous ones (to allow clients to validate older tokens).
type Manager struct {
	keys    sync.Map
	current string
}

// NewManager returns a new Manager and rotates in the specified key pair.
func NewManager(kid string, key crypto.PrivateKey) (*Manager, error) {
	m := &Manager{}
	err := m.Rotate(kid, key)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// Sign signs the claims with the current private key.
func (m *Manager) Sign(claims *Claims) (string, error) {
	key, err := m.GetCurrentPrivateKey()
	if err != nil {
		return "", err
	}

	return claims.Sign(key)
}

// GetTokenKey gets the token public key for with the specified kid. If the kid is the empty string, it will return the currently active key.
func (m *Manager) GetTokenKey(kid string) (crypto.PublicKey, error) {
	v, ok := m.keys.Load(kid)
	if !ok {
		return nil, fmt.Errorf("Could not find token key with kid `%s`", kid)
	}

	key, ok := v.(crypto.PrivateKey)
	if !ok {
		panic(fmt.Errorf("Expected type crypto.PrivateKey when loading the public key, but got %T", v))
	}

	return getPublic(key), nil
}

func getPublic(key crypto.PrivateKey) crypto.PublicKey {
	switch v := key.(type) {
	case *rsa.PrivateKey:
		return &v.PublicKey
	case *ecdsa.PrivateKey:
		return &v.PublicKey
	}
	return nil
}

// GetCurrentTokenKey returns the public key of the currently active keypair.
func (m *Manager) GetCurrentTokenKey() (string, crypto.PublicKey, error) {
	kid := m.current
	key, err := m.GetTokenKey(kid)
	return kid, key, err
}

// GetPrivateKey gets the privatekey for the specified kid.
func (m *Manager) GetPrivateKey(kid string) (crypto.PrivateKey, error) {
	v, ok := m.keys.Load(kid)
	if !ok {
		return nil, fmt.Errorf("Could not find token key with kid `%s`", kid)
	}

	key, ok := v.(crypto.PrivateKey)
	if !ok {
		panic(fmt.Errorf("Expected type crypto.PrivateKey when loading the private key, but got %T", v))
	}

	return key, nil
}

// GetCurrentPrivateKey returns the privatekey that is currently active.
func (m *Manager) GetCurrentPrivateKey() (crypto.PrivateKey, error) {
	return m.GetPrivateKey(m.current)
}

// Rotate adds a new token private-public keypair to the Manager and makes it the current keypair.
// The old keypair will be kept in memory to allow clients to validate older tokens with it.
func (m *Manager) Rotate(kid string, key crypto.PrivateKey) error {
	_, ok := m.keys.Load(kid)
	if ok {
		return fmt.Errorf("Token key with kid `%s` already exists", kid)
	}

	switch key.(type) {
	case *rsa.PrivateKey, *ecdsa.PrivateKey:
	default:
		return ErrUnsupportedSigningMethod
	}

	m.keys.Store(kid, key)
	m.current = kid

	return nil
}

// RotateFromPEM rotates in the new key from the content of a PEM-encoded private key file.
func (m *Manager) RotateFromPEM(kid string, content []byte) error {
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

	return m.Rotate(kid, key)
}
