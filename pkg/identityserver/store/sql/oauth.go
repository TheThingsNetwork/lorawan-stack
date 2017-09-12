// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
)

// OAuthStore implements store.ApplicationStore.
type OAuthStore struct {
	*Store
	*ClientStore
}

func init() {
}

// SaveAuthorizationCode saves the authorization code.
func (o *OAuthStore) SaveAuthorizationCode(authorization store.AuthorizationData) error {
	return saveAuthorizationCode(o.db, authorization)
}

func saveAuthorizationCode(q db.QueryContext, authorization store.AuthorizationData) error {
	result := new(string)
	return q.NamedSelectOne(
		result,
		`INSERT
			INTO authorization_codes (
				authorization_code,
				client_id,
				created_at,
				expires_in,
				scope,
				redirect_uri,
				state,
			)
			VALUES (
				:authorization_code,
				:client_id,
				:created_at,
				:expires_in,
				:scope,
				:redirect_uri,
				:state,
			)
			RETURNING authorization_code`,
		authorization,
	)
}

// FindAuthorizationCode finds the authorization code.
func (o *OAuthStore) FindAuthorizationCode(authorizationCode string) (*store.AuthorizationData, error) {
	return findAuthorizationCode(o.db, authorizationCode)
}

func findAuthorizationCode(q db.QueryContext, authorizationCode string) (*store.AuthorizationData, error) {
	result := new(store.AuthorizationData)
	err := q.SelectOne(
		result,
		`SELECT * FROM authorization_codes
			WHERE authorization_code = $1`,
		authorizationCode,
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteAuthorizationCode deletes the authorization code.
func (o *OAuthStore) DeleteAuthorizationCode(authorizationCode string) error {
	return deleteAuthorizationCode(o.db, authorizationCode)
}

func deleteAuthorizationCode(q db.QueryContext, authorizationCode string) error {
	code := new(string)
	return q.SelectOne(
		code,
		`DELETE
			FROM authorization_codes
			WHERE authorization_code = $1
			RETURNING authorization_code`,
		authorizationCode,
	)
}

// SaveRefreshToken saves the refresh token.
func (o *OAuthStore) SaveRefreshToken(access store.RefreshData) error {
	return saveRefreshToken(o.db, access)
}

func saveRefreshToken(q db.QueryContext, refresh store.RefreshData) error {
	result := new(string)
	return q.NamedSelectOne(
		result,
		`INSERT
			INTO refresh_tokens (
				refresh_token,
				client_id,
				created_at,
				scope,
				redirect_uri,
			)
			VALUES (
				:refresh_token,
				:client_id,
				:created_at,
				:scope,
				:redirect_uri,
			)
			RETURNING refresh_token`,
		refresh,
	)
}

// FindRefreshToken finds the refresh token.
func (o *OAuthStore) FindRefreshToken(refreshToken string) (*store.RefreshData, error) {
	return findRefreshToken(o.db, refreshToken)
}

func findRefreshToken(q db.QueryContext, refreshToken string) (*store.RefreshData, error) {
	result := new(store.RefreshData)
	err := q.SelectOne(
		result,
		`SELECT * FROM refresh_tokens
			WHERE refresh_token = $1`,
		refreshToken,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteRefreshToken deletes the refresh token from the database.
func (o *OAuthStore) DeleteRefreshToken(refreshToken string) error {
	return deleteRefreshToken(o.db, refreshToken)
}

func deleteRefreshToken(q db.QueryContext, refreshToken string) error {
	token := new(string)
	return q.SelectOne(
		token,
		`DELETE
			FROM refresh_tokens
			WHERE refresh_token = $1
			RETURNING refresh_token`,
		refreshToken,
	)
}
