// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
)

// OAuthStore is a store that manages OAuth refresh tokens and authorization codes.
type OAuthStore interface {
	ClientStore

	// SaveAuthorizationCode saves the authorization code.
	SaveAuthorizationCode(authorization *types.AuthorizationData) error

	// FindAuthorizationCode finds the authorization code.
	FindAuthorizationCode(authorizationCode string) (*types.AuthorizationData, types.Client, error)

	// DeleteAuthorizationCode deletes the authorization code.
	DeleteAuthorizationCode(authorizationCode string) error

	// SaveRefreshToken saves the refresh token.
	SaveRefreshToken(refresh *types.RefreshData) error

	// FindRefreshToken finds the refresh token.
	FindRefreshToken(refreshToken string) (*types.RefreshData, types.Client, error)

	// FindRefreshToken deletes the refresh token from the database.
	DeleteRefreshToken(refreshToken string) error
}
