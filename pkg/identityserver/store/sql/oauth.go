// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
)

func init() {
	ErrRefreshTokenNotFound.Register()
	ErrRefreshTokenConflict.Register()
}

// ErrAuthorizationCodeNotFound is returned when trying to fetch or delete an authorization code that
// does not exist.
var ErrAuthorizationCodeNotFound = &errors.ErrDescriptor{
	MessageFormat: "authorization Code token does not exist",
	Code:          501,
	Type:          errors.NotFound,
}

// ErrDuplicateAuthorizationCode is returned when trying to create an authorization code that already exists.
var ErrAuthorizationCodeConflict = &errors.ErrDescriptor{
	MessageFormat: "Authorization code already exists",
	Code:          502,
	Type:          errors.Conflict,
}

// ErrRefreshTokenNotFound is returned when trying to fetch or delete a refresh token that
// does not exist.
var ErrRefreshTokenNotFound = &errors.ErrDescriptor{
	MessageFormat: "Refresh token does not exist",
	Code:          503,
	Type:          errors.NotFound,
}

// ErrDuplicateRefreshToken is returned when trying to create a refresh token that already exists.
var ErrRefreshTokenConflict = &errors.ErrDescriptor{
	MessageFormat: "Refresh token already exists",
	Code:          504,
	Type:          errors.Conflict,
}

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

	if _, dup := db.IsDuplicate(err); dup {
		return ErrAuthorizationCodeConflict.New(nil)
	}

	return err
}

// GetAuthorizationCode finds the authorization code.
func (s *OAuthStore) GetAuthorizationCode(authorizationCode string) (*types.AuthorizationData, types.Client, error) {
	var data *types.AuthorizationData
	var client types.Client
	err := s.transact(func(tx *db.Tx) error {
		var err error
		data, client, err = s.getAuthorizationCode(tx, authorizationCode)
		return err
	})

	return data, client, err
}

func (s *OAuthStore) getAuthorizationCode(q db.QueryContext, authorizationCode string) (*types.AuthorizationData, types.Client, error) {
	result := new(types.AuthorizationData)
	err := q.SelectOne(
		result,
		`SELECT * FROM authorization_codes
			WHERE authorization_code = $1`,
		authorizationCode,
	)

	if db.IsNoRows(err) {
		return nil, nil, ErrAuthorizationCodeNotFound.New(nil)
	}

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
	err := q.SelectOne(
		code,
		`DELETE
			FROM authorization_codes
			WHERE authorization_code = $1
			RETURNING authorization_code`,
		authorizationCode,
	)

	if db.IsNoRows(err) {
		return ErrAuthorizationCodeNotFound.New(nil)
	}

	return err
}

// SaveRefreshToken saves the refresh token.
func (s *OAuthStore) SaveRefreshToken(access *types.RefreshData) error {
	return saveRefreshToken(s.queryer(), access)
}

func saveRefreshToken(q db.QueryContext, refresh *types.RefreshData) error {
	result := new(string)
	err := q.NamedSelectOne(
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

	if _, dup := db.IsDuplicate(err); dup {
		return ErrRefreshTokenConflict.New(nil)
	}

	return err
}

// GetRefreshToken finds the refresh token.
func (s *OAuthStore) GetRefreshToken(refreshToken string) (*types.RefreshData, types.Client, error) {
	var data *types.RefreshData
	var client types.Client
	err := s.transact(func(tx *db.Tx) error {
		var err error
		data, client, err = s.getRefreshToken(tx, refreshToken)
		return err
	})

	return data, client, err
}

func (s *OAuthStore) getRefreshToken(q db.QueryContext, refreshToken string) (*types.RefreshData, types.Client, error) {
	result := new(types.RefreshData)
	err := q.SelectOne(
		result,
		`SELECT * FROM refresh_tokens
			WHERE refresh_token = $1`,
		refreshToken,
	)

	if db.IsNoRows(err) {
		return nil, nil, ErrRefreshTokenNotFound.New(nil)
	}

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
	err := q.SelectOne(
		token,
		`DELETE
			FROM refresh_tokens
			WHERE refresh_token = $1
			RETURNING refresh_token`,
		refreshToken,
	)

	if db.IsNoRows(err) {
		return ErrRefreshTokenNotFound.New(nil)
	}

	return err
}
