// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// AuthorizationData is the data stored for an authorization code.
type AuthorizationData struct {
	// AuthorizationCode is the actual opaque authorization code.
	AuthorizationCode string

	// ClientID is the ID of the client this authorization code is for.
	ClientID string

	// CreatedAt is the time when the authorization code was created.
	CreatedAt time.Time

	// ExpiresIn is the time the authorization code should be valid for since it was issued.
	ExpiresIn time.Duration

	// Scope is the scope of the authorization code.
	Scope string

	// RedirectURI is the redirect URI from the request.
	RedirectURI string

	// State is the state the client passed when authorizing.
	State string

	// UserID is the user ID the authorization is for.
	UserID string
}

// IsExpired returns error if the receiver is expired.
func (a *AuthorizationData) IsExpired() error {
	if exp := a.CreatedAt.Add(a.ExpiresIn); exp.Before(time.Now()) {
		return errors.Errorf("Authorization code is expired by %v", time.Now().Sub(exp))
	}
	return nil
}

// AccessData is the data stored for access tokens.
type AccessData struct {
	// AccessToken is the actual opaque access token.
	AccessToken string

	// ClientID is the ID of the client this access token is for.
	ClientID string

	// UserID is the user ID the access token is for.
	UserID string

	// CreatedAt is the time when the access token was created.
	CreatedAt time.Time

	// ExpiresIn is the time the access token should be valid for since it was issued.
	ExpiresIn time.Duration

	// Scope is the scope of the access token.
	Scope string

	// RedirectURI is the redirect URI from the request.
	RedirectURI string
}

// IsExpired returns error if the receiver is expired.
func (a *AccessData) IsExpired() error {
	if exp := a.CreatedAt.Add(a.ExpiresIn); exp.Before(time.Now()) {
		return errors.Errorf("Access token is expired by %v", time.Now().Sub(exp))
	}
	return nil
}

// RefreshData is the data stored for refresh tokens.
type RefreshData struct {
	// RefreshToken is the actaul opaque refresh token
	RefreshToken string

	// ClientID is the id of the client this refresh token is for.
	ClientID string

	// UserID is the user ID the refresh token is for.
	UserID string

	// CreatedAt is the time when the refresh token was created.
	CreatedAt time.Time

	// Scope is the scope of the authorization code.
	Scope string

	// RedirectURI is the redirect URI from the request.
	RedirectURI string
}

// OAuthStore is a store that manages OAuth authorization codes, access tokens
// and refresh tokens.
type OAuthStore interface {
	// SaveAuthorizationCode saves the authorization code.
	SaveAuthorizationCode(AuthorizationData) error

	// GetAuthorizationCode finds the authorization code.
	GetAuthorizationCode(string) (AuthorizationData, error)

	// DeleteAuthorizationCode deletes the authorization code.
	DeleteAuthorizationCode(string) error

	// SaveAccessToken saves the access token.
	SaveAccessToken(AccessData) error

	// GetAccessToken finds the access token.
	GetAccessToken(string) (AccessData, error)

	// DeleteAccessToken deletes the access token.
	DeleteAccessToken(string) error

	// SaveRefreshToken saves the refresh token.
	SaveRefreshToken(RefreshData) error

	// GetRefreshToken finds the refresh token.
	GetRefreshToken(string) (RefreshData, error)

	// DeleteRefreshToken deletes the refresh token from the database.
	DeleteRefreshToken(string) error

	// ListAuthorizedClients returns a list of clients authorized by a given user.
	ListAuthorizedClients(ttnpb.UserIdentifiers, ClientSpecializer) ([]Client, error)

	// RevokeAuthorizedClient deletes the access tokens and refresh token
	// granted to a client by a given user.
	RevokeAuthorizedClient(ttnpb.UserIdentifiers, ttnpb.ClientIdentifiers) error
}
