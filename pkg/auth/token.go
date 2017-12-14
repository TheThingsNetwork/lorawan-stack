// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package auth

// Token is the value used in the JOSE header to denote that it is an access token.
const Token = "token"

// GenerateAccessToken generates an Access Token using the JOSE header.
func GenerateAccessToken(issuer string) (string, error) {
	return generate(Token, &Payload{
		Issuer: issuer,
	})
}
