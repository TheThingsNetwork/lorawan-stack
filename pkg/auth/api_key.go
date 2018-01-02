// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package auth

// Key is the value used in the JOSE header to denote that it is an API key.
const Key = "key"

// APIKeyType denotes the API key type.
type APIKeyType int

const (
	// ApplicationKey denotes it is an application API key.
	ApplicationKey = iota

	// GatewayKey denotes it is a gateway API key.
	GatewayKey

	// UserKey denotes it is an user API key.
	UserKey
)

// String implements fmt.Stringer.
func (k APIKeyType) String() string {
	switch k {
	case ApplicationKey:
		return "application"
	case GatewayKey:
		return "gateway"
	case UserKey:
		return "user"
	default:
		return "invalid type"
	}
}

// GenerateApplicationAPIKey generates an application API Key using the JOSE header.
func GenerateApplicationAPIKey(issuer string) (string, error) {
	return generate(Key, &Payload{
		Issuer: issuer,
		Type:   ApplicationKey,
	})
}

// GenerateGatewayAPIKey generates a gateway API Key using the JOSE header.
func GenerateGatewayAPIKey(issuer string) (string, error) {
	return generate(Key, &Payload{
		Issuer: issuer,
		Type:   GatewayKey,
	})
}

// GenerateUserAPIKey generates an user API Key using the JOSE header.
func GenerateUserAPIKey(issuer string) (string, error) {
	return generate(Key, &Payload{
		Issuer: issuer,
		Type:   UserKey,
	})
}
