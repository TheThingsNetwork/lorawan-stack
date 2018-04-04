// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

	// OrganizationKey denotes it is an organization API key.
	OrganizationKey
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
	case OrganizationKey:
		return "organization"
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

// GenerateOrganizationAPIKey generates an organization API Key using the JOSE header.
func GenerateOrganizationAPIKey(issuer string) (string, error) {
	return generate(Key, &Payload{
		Issuer: issuer,
		Type:   OrganizationKey,
	})
}
