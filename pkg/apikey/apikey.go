// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package apikey

import (
	"encoding/base64"
	"fmt"

	"github.com/TheThingsNetwork/ttn/pkg/random"
)

var (
	// enc is the encoder we use
	enc = base64.RawURLEncoding

	// entropy is the amount of entropy we use
	entropy = 32

	// tenantLen is the minimum length to pad tenants to.
	// Longer tenant id's will not be padded.
	tenantLen = 32

	// minLen is the minimum lenght of a decoded key, consisting of:
	// entropy
	minLen = entropy
)

// GenerateAPIKey generates an API key for the specified tenant.
// The key has the following byte layout, base64 encoded:
//
//     | entropy  |  tenant             |
//     +----------+---------------------+
//     | 32 bytes |  ? (variable) bytes |
//
// The id is padded with random bytes to make sure all api keys have the same length,
// it is prefixed with the length of the id so we can decode the id from the API key.
// The entropy of an API key varies between 20 and 35 - len(tenant).
func GenerateAPIKey(tenant string) string {
	raw := append(random.Bytes(entropy), pad([]byte(tenant), tenantLen)...)

	return enc.EncodeToString(raw)
}

// KeyTenant gets the tenant from the base64 encoded key.
func KeyTenant(key string) (string, error) {
	dec, err := enc.DecodeString(key)
	if err != nil {
		return "", err
	}

	if len(dec) < minLen {
		return "", fmt.Errorf("Invalid number of segments in key")
	}

	tenant := unpad(dec[entropy:])
	if tenant == nil {
		return "", fmt.Errorf("Invalid format of key")
	}

	return string(tenant), nil
}
