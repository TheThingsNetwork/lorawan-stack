// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import "github.com/TheThingsNetwork/ttn/pkg/identityserver/types"

// OAuthStore is a store that manages OAuth refresh tokens and authorization codes.
type OAuthStore interface {
	// SaveAuthorizationCode saves the authorization code.
	SaveAuthorizationCode(authorization *types.AuthorizationData) error

	// GetAuthorizationCode finds the authorization code.
	GetAuthorizationCode(authorizationCode string) (*types.AuthorizationData, error)

	// DeleteAuthorizationCode deletes the authorization code.
	DeleteAuthorizationCode(authorizationCode string) error

	// SaveRefreshToken saves the refresh token.
	SaveRefreshToken(refresh *types.RefreshData) error

	// GetRefreshToken finds the refresh token.
	GetRefreshToken(refreshToken string) (*types.RefreshData, error)

	// DeleteRefreshToken deletes the refresh token from the database.
	DeleteRefreshToken(refreshToken string) error
}
