// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"database/sql/driver"
	"strings"
)

// Grant represents an OAuth2 flow a client can use to get access to a token
type Grant string

const (
	// AuthorizationCodeGrant is the grant type used to exchange an authorization
	// code for an access token
	AuthorizationCodeGrant Grant = "authorization_code"

	// PasswordGrant is the grant type used to exchange an username and password
	// for an access token
	PasswordGrant Grant = "password"

	// RefreshTokenGrant is the grant type used to exchange a refresh token for
	// an access token
	RefreshTokenGrant Grant = "refresh_token"
)

// String implements fmt.Stringer interface
func (g Grant) String() string {
	return string(g)
}

// Grants is the type that represents which grants a client can use
type Grants struct {
	// AuthorizationCode denotes whether the client allows Authorization Code grant type
	AuthorizationCode bool

	// Password denotes whether the client allows Password grant type
	Password bool

	// RefreshToken denotes whether the client allows Refresh Token grant type
	RefreshToken bool
}

// Value implements sql.Valuer interface
func (g Grants) Value() (driver.Value, error) {
	grants := make([]string, 0)

	if g.AuthorizationCode {
		grants = append(grants, AuthorizationCodeGrant.String())
	}

	if g.Password {
		grants = append(grants, PasswordGrant.String())
	}

	if g.RefreshToken {
		grants = append(grants, RefreshTokenGrant.String())
	}

	return strings.Join(grants, ","), nil
}

// Scan implements sql.Scanner interface
func (g *Grants) Scan(src interface{}) error {
	grants := strings.Split(src.(string), ",")

	for _, grant := range grants {
		switch Grant(grant) {
		case AuthorizationCodeGrant:
			g.AuthorizationCode = true
		case PasswordGrant:
			g.Password = true
		case RefreshTokenGrant:
			g.RefreshToken = true
		}
	}

	return nil
}
