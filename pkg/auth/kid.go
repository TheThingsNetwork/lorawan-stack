// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package auth

import "crypto"

// PrivateKeyWithKID is a private key with a specific kid which can be used to disambiguate between token keys when rotating them.
type PrivateKeyWithKID struct {
	crypto.PrivateKey

	// KID is the key id.
	KID string
}

// WithKID returns a crypto.PrivateKeyWithKID that explicitly has the KID set.
func WithKID(key crypto.PrivateKey, kid string) *PrivateKeyWithKID {
	return &PrivateKeyWithKID{
		PrivateKey: key,
		KID:        kid,
	}
}

// GetKID returns the KID for a given private key, or the empty string if it does not have one.
func GetKID(key crypto.PrivateKey) string {
	if w, ok := key.(*PrivateKeyWithKID); ok {
		return w.KID
	}
	return ""
}
