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
	entropy = 24

	// idLen is the max length of the id
	idLen = 36

	// minLen is the minimum lenght of a decoded key, consisting of:
	// entropy + typ + id_len + id +  tenant
	minLen = entropy + 1 + idLen + 1 + 1
)

// Type is the type of the API key.
type Type byte

const (
	// Invalid is the invalid type.
	Invalid Type = 0x00

	// Application is the application API key type.
	Application Type = 0x01

	// Gateway is the gateway API key type.
	Gateway Type = 0x02
)

// GenerateApplicationAPIKey generates an application API key for the gateway with the specified
// application ID in the specified tenant.
// If the application ID does not have the correct length, it will return the empty string.
func GenerateApplicationAPIKey(tenant, appID string) string {
	return generateKey(Application, tenant, appID)
}

// GenerateGatewayAPIKey generates a gateway API key for the gateway with the specified
// gateway ID in the specified tenant.
// If the gateway ID does not have the correct length, it will return the empty string.
func GenerateGatewayAPIKey(tenant, gwID string) string {
	return generateKey(Gateway, tenant, gwID)
}

// generateKey generates an API key that has the following byte layout, base64 encoded.
//
//     | entropy  | type   | id length | id + random padding | tenant    |
//     +----------+--------+-----------+---------------------+-----------+
//     | 24 bytes | 1 byte | 1 byte    | 36 bytes            | var bytes |
//
// The id is padded with random bytes to make sure all api keys within the same tenant have the same length,
// it is prefixed with the length of the id so we can decode the id from the API key.
// The entropy of an API key varies between 24 and 36 - len(id).
func generateKey(typ Type, tenant string, id string) string {
	if len(id) > idLen {
		return ""
	}

	if typ == Invalid {
		return ""
	}

	raw := append(random.Bytes(entropy), byte(typ))
	raw = append(raw, pad([]byte(id), 36)...)
	raw = append(raw, []byte(tenant)...)

	return enc.EncodeToString(raw)
}

// String implements fmt.Stringer.
func (k Type) String() string {
	switch k {
	case Invalid:
		return "invalid"
	case Application:
		return "application"
	case Gateway:
		return "gateway"
	}
	return "invalid"
}

// KeyInfo holds a description for the API key.
type KeyInfo struct {
	// Type is the type of the api key.
	Type Type

	// ID is the id of the entity the key is for.
	ID string

	// Tenant is the hostname of the tenant the key is for.
	Tenant string

	// Entropy is the amount of entropy bits in the key.
	Entropy int
}

// DecodeKey gets the key info from the base64 encoded key.
func DecodeKey(key string) (*KeyInfo, error) {
	dec, err := enc.DecodeString(key)
	if err != nil {
		return nil, err
	}

	if len(dec) < minLen {
		return nil, fmt.Errorf("Invalid number of segments in key")
	}

	typ := Type(dec[entropy])
	if typ == Invalid {
		return nil, fmt.Errorf("Invalid key type")
	}

	id := unpad(dec[entropy+1 : entropy+1+idLen+1])
	if id == nil {
		return nil, fmt.Errorf("Invalid id in API Key")
	}

	tenant := dec[entropy+1+idLen+1:]

	return &KeyInfo{
		Type:    typ,
		ID:      string(id),
		Tenant:  string(tenant),
		Entropy: 8 * (entropy + idLen - len(id)),
	}, nil
}
