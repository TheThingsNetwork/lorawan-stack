// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
)

// OAuthStore implements store.OAuthStore.
type OAuthStore struct {
	storer
	*ClientStore
}

// NewOAuthStore creates a new OAuth store.
func NewOAuthStore(store storer, clients *ClientStore) *OAuthStore {
	return &OAuthStore{
		storer:      store,
		ClientStore: clients,
	}
}

// SaveAuthorizationCode saves the authorization code.
func (s *OAuthStore) SaveAuthorizationCode(authorization *types.AuthorizationData) error {
	return saveAuthorizationCode(s.queryer(), authorization)
}

func saveAuthorizationCode(q db.QueryContext, data *types.AuthorizationData) error {
	_, err := q.NamedExec(
		`INSERT
			INTO authorization_codes (
				authorization_code,
				client_id,
				created_at,
				expires_in,
				scope,
				redirect_uri,
				state,
				user_id
			)
			VALUES (
				:authorization_code,
				:client_id,
				:created_at,
				:expires_in,
				:scope,
				:redirect_uri,
				:state,
				:user_id
			)
		`,
		data,
	)

	return err
}

// FindAuthorizationCode finds the authorization code.
func (s *OAuthStore) FindAuthorizationCode(authorizationCode string) (*types.AuthorizationData, types.Client, error) {
	var data *types.AuthorizationData
	var client types.Client
	err := s.transact(func(tx *db.Tx) error {
		var err error
		data, client, err = s.findAuthorizationCode(tx, authorizationCode)
		return err
	})

	return data, client, err
}

func (s *OAuthStore) findAuthorizationCode(q db.QueryContext, authorizationCode string) (*types.AuthorizationData, types.Client, error) {
	result := new(types.AuthorizationData)
	err := q.SelectOne(
		result,
		`SELECT * FROM authorization_codes
			WHERE authorization_code = $1`,
		authorizationCode,
	)

	if err != nil {
		return nil, nil, err
	}

	client := s.ClientStore.factory.BuildClient()
	err = s.ClientStore.client(q, result.ClientID, client)
	if err != nil {
		return nil, nil, err
	}

	return result, client, nil
}

// DeleteAuthorizationCode deletes the authorization code.
func (s *OAuthStore) DeleteAuthorizationCode(authorizationCode string) error {
	return deleteAuthorizationCode(s.queryer(), authorizationCode)
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
func (s *OAuthStore) SaveRefreshToken(access *types.RefreshData) error {
	return saveRefreshToken(s.queryer(), access)
}

func saveRefreshToken(q db.QueryContext, refresh *types.RefreshData) error {
	result := new(string)
	return q.NamedSelectOne(
		result,
		`INSERT
			INTO refresh_tokens (
				refresh_token,
				client_id,
				created_at,
				scope,
				redirect_uri
			)
			VALUES (
				:refresh_token,
				:client_id,
				:created_at,
				:scope,
				:redirect_uri
			)
			RETURNING refresh_token`,
		refresh,
	)
}

// FindRefreshToken finds the refresh token.
func (s *OAuthStore) FindRefreshToken(refreshToken string) (*types.RefreshData, types.Client, error) {
	var data *types.RefreshData
	var client types.Client
	err := s.transact(func(tx *db.Tx) error {
		var err error
		data, client, err = s.findRefreshToken(tx, refreshToken)
		return err
	})

	return data, client, err
}

func (s *OAuthStore) findRefreshToken(q db.QueryContext, refreshToken string) (*types.RefreshData, types.Client, error) {
	result := new(types.RefreshData)
	err := q.SelectOne(
		result,
		`SELECT * FROM refresh_tokens
			WHERE refresh_token = $1`,
		refreshToken,
	)
	if err != nil {
		return nil, nil, err
	}

	client := s.ClientStore.factory.BuildClient()
	err = s.ClientStore.client(q, result.ClientID, client)
	if err != nil {
		return nil, nil, err
	}

	return result, client, nil
}

// DeleteRefreshToken deletes the refresh token from the database.
func (s *OAuthStore) DeleteRefreshToken(refreshToken string) error {
	return deleteRefreshToken(s.queryer(), refreshToken)
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
