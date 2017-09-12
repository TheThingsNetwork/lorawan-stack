// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
)

// AuthorizationData is the data stored for an authorization code.
type AuthorizationData struct {
	// Code is the actual opaque authorization code.
	Code string `db:"code"`

	// ClientID is the id of the client this authorization code is for.
	ClientID string `db:"client_id"`

	// CreatedAt is the time when the authorization code was created.
	CreatedAt time.Time `db:"created_at"`

	// ExpiresIn is the time the authorization code should be valid for.
	ExpiresIn time.Duration `db:"expires_in"`

	// Scope is the scope of the authorization code.
	Scope []types.Scope `db:"scope"`

	// RedirectURI is the redirect URI from the request.
	RedirectURI string `db:"redirect_uri"`

	// State is the state the client passed when authorizing.
	State string `db:"state"`
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
	Scope []types.Scope `db:"scope"`

	// RedirectURI is the redirect URI from the request.
	RedirectURI string `db:"redirect_uri"`
}

// OAuthStore is a store that manages OAuth refresh tokens and authorization codes.
type OAuthStore interface {
	ClientStore

	// SaveAuthorizationCode saves the authorization code.
	SaveAuthorizationCode(authorization AuthorizationData) error

	// FindAuthorizationCode finds the authorization code.
	FindAuthorizationCode(authorizationCode string) (*AuthorizationData, error)

	// DeleteAuthorizationCode deletes the authorization code.
	DeleteAuthorizationCode(authorizationCode string) error

	// SaveRefreshToken saves the refresh token.
	SaveRefreshToken(refresh RefreshData) error

	// FindRefreshToken finds the refresh token.
	FindRefreshToken(refreshToken string) (*RefreshData, error)

	// FindRefreshToken deletes the refresh token from the database.
	DeleteRefreshToken(refreshToken string) error
}
