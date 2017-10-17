// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import "time"

// AuthorizationData is the data stored for an authorization code.
type AuthorizationData struct {
	// Code is the actual opaque authorization code.
	Code string `db:"authorization_code"`

	// ClientID is the id of the client this authorization code is for.
	ClientID string `db:"client_id"`

	// CreatedAt is the time when the authorization code was created.
	CreatedAt time.Time `db:"created_at"`

	// ExpiresIn is the time the authorization code should be valid for.
	ExpiresIn time.Duration `db:"expires_in"`

	// Scope is the scope of the authorization code.
	Scope string `db:"scope"`

	// RedirectURI is the redirect URI from the request.
	RedirectURI string `db:"redirect_uri"`

	// State is the state the client passed when authorizing.
	State string `db:"state"`

	// Username is the username the authorization is for.
	Username string `db:"username"`
}

// RefreshData is the data stored for refresh tokens.
type RefreshData struct {
	// RefreshToken is the actaul opaque refresh token
	RefreshToken string `db:"refresh_token"`

	// ClientID is the id of the client this refresh token is for.
	ClientID string `db:"client_id"`

	// CreatedAt is the time when the refresh token was created.
	CreatedAt time.Time `db:"created_at"`

	// Scope is the scope of the authorization code.
	Scope string `db:"scope"`

	// RedirectURI is the redirect URI from the request.
	RedirectURI string `db:"redirect_uri"`
}
