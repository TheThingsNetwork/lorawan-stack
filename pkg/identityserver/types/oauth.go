// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import "time"

// AuthorizationData is the data stored for an authorization code.
type AuthorizationData struct {
	// AuthorizationCode is the actual opaque authorization code.
	AuthorizationCode string

	// ClientID is the id of the client this authorization code is for.
	ClientID string

	// CreatedAt is the time when the authorization code was created.
	CreatedAt time.Time

	// ExpiresIn is the time the authorization code should be valid for.
	ExpiresIn time.Duration

	// Scope is the scope of the authorization code.
	Scope string

	// RedirectURI is the redirect URI from the request.
	RedirectURI string

	// State is the state the client passed when authorizing.
	State string

	// Username is the username the authorization is for.
	Username string
}

// RefreshData is the data stored for refresh tokens.
type RefreshData struct {
	// RefreshToken is the actaul opaque refresh token
	RefreshToken string

	// ClientID is the id of the client this refresh token is for.
	ClientID string

	// CreatedAt is the time when the refresh token was created.
	CreatedAt time.Time

	// Scope is the scope of the authorization code.
	Scope string

	// RedirectURI is the redirect URI from the request.
	RedirectURI string
}
