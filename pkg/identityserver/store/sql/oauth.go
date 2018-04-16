// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sql

import (
	"github.com/satori/go.uuid"
	"go.thethings.network/lorawan-stack/pkg/identityserver/db"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type oauthStore struct {
	storer
}

func newOAuthStore(store storer) *oauthStore {
	return &oauthStore{
		storer: store,
	}
}

// SaveAuthorizationCode saves the authorization code.
func (s *oauthStore) SaveAuthorizationCode(data store.AuthorizationData) error {
	err := s.transact(func(tx *db.Tx) error {
		st := s.store()

		userID, err := st.Users.getUserID(tx, ttnpb.UserIdentifiers{UserID: data.UserID})
		if err != nil {
			return err
		}

		clientID, err := st.Clients.getClientID(tx, ttnpb.ClientIdentifiers{ClientID: data.ClientID})
		if err != nil {
			return err
		}

		return s.saveAuthorizationCode(tx, userID, clientID, data)
	})
	return err
}

func (s *oauthStore) saveAuthorizationCode(q db.QueryContext, userID, clientID uuid.UUID, data store.AuthorizationData) error {
	_, err := q.Exec(
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
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`,
		data.AuthorizationCode,
		clientID,
		data.CreatedAt,
		data.ExpiresIn,
		data.Scope,
		data.RedirectURI,
		data.State,
		userID,
	)

	if _, dup := db.IsDuplicate(err); dup {
		return store.ErrAuthorizationCodeConflict.New(nil)
	}

	return err
}

type authorizationData struct {
	ClientUUID uuid.UUID
	UserUUID   uuid.UUID
	store.AuthorizationData
}

// GetAuthorizationCode finds the authorization code.
func (s *oauthStore) GetAuthorizationCode(authorizationCode string) (result store.AuthorizationData, err error) {
	err = s.transact(func(tx *db.Tx) error {
		data, err := s.getAuthorizationCode(tx, authorizationCode)
		if err != nil {
			return err
		}

		st := s.store()

		user, err := st.Users.getUserIdentifiersFromID(tx, data.UserUUID)
		if err != nil {
			return err
		}
		data.AuthorizationData.UserID = user.UserID

		client, err := st.Clients.getClientIdentifiersFromID(tx, data.ClientUUID)
		if err != nil {
			return err
		}
		data.AuthorizationData.ClientID = client.ClientID

		result = data.AuthorizationData

		return nil
	})
	return
}

func (s *oauthStore) getAuthorizationCode(q db.QueryContext, authorizationCode string) (data authorizationData, err error) {
	err = q.SelectOne(
		&data,
		`SELECT
				authorization_code,
				client_id AS client_uuid,
				created_at,
				expires_in,
				scope,
				redirect_uri,
				state,
				user_id AS user_uuid
			FROM authorization_codes
			WHERE authorization_code = $1`,
		authorizationCode)

	if db.IsNoRows(err) {
		err = store.ErrAuthorizationCodeNotFound.New(nil)
	}

	return
}

// DeleteAuthorizationCode deletes the authorization code.
func (s *oauthStore) DeleteAuthorizationCode(authorizationCode string) error {
	return s.deleteAuthorizationCode(s.queryer(), authorizationCode)
}

func (s *oauthStore) deleteAuthorizationCode(q db.QueryContext, authorizationCode string) error {
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
		return store.ErrAuthorizationCodeNotFound.New(nil)
	}

	return err
}

func (s *oauthStore) deleteAuthorizationCodesByUser(q db.QueryContext, userID uuid.UUID) error {
	_, err := q.Exec(`DELETE FROM authorization_codes WHERE user_id = $1`, userID)
	return err
}

type accessData struct {
	ClientUUID uuid.UUID
	UserUUID   uuid.UUID
	store.AccessData
}

// SaveAccessToken saves the access data.
func (s *oauthStore) SaveAccessToken(data store.AccessData) error {
	err := s.transact(func(tx *db.Tx) error {
		st := s.store()

		userID, err := st.Users.getUserID(tx, ttnpb.UserIdentifiers{UserID: data.UserID})
		if err != nil {
			return err
		}

		clientID, err := st.Clients.getClientID(tx, ttnpb.ClientIdentifiers{ClientID: data.ClientID})
		if err != nil {
			return err
		}

		return s.saveAccessToken(tx, userID, clientID, data)
	})
	return err
}

func (s *oauthStore) saveAccessToken(q db.QueryContext, userID, clientID uuid.UUID, access store.AccessData) error {
	result := new(string)
	err := q.SelectOne(
		result,
		`INSERT
			INTO access_tokens (
				access_token,
				client_id,
				user_id,
				created_at,
				expires_in,
				scope,
				redirect_uri
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING access_token`,
		access.AccessToken,
		clientID,
		userID,
		access.CreatedAt,
		access.ExpiresIn,
		access.Scope,
		access.RedirectURI,
	)

	if _, dup := db.IsDuplicate(err); dup {
		return store.ErrAccessTokenConflict.New(nil)
	}

	return err
}

// GetAccessToken finds the access token.
func (s *oauthStore) GetAccessToken(accessToken string) (result store.AccessData, err error) {
	err = s.transact(func(tx *db.Tx) error {
		data, err := s.getAccessToken(tx, accessToken)
		if err != nil {
			return err
		}

		st := s.store()

		user, err := st.Users.getUserIdentifiersFromID(tx, data.UserUUID)
		if err != nil {
			return err
		}
		data.AccessData.UserID = user.UserID

		client, err := st.Clients.getClientIdentifiersFromID(tx, data.ClientUUID)
		if err != nil {
			return err
		}
		data.AccessData.ClientID = client.ClientID

		result = data.AccessData

		return nil
	})
	return
}

func (s *oauthStore) getAccessToken(q db.QueryContext, accessToken string) (result accessData, err error) {
	err = q.SelectOne(
		&result,
		`SELECT
				access_token,
				client_id AS client_uuid,
				user_id AS user_uuid,
				created_at,
				expires_in,
				scope,
				redirect_uri
			FROM access_tokens
			WHERE access_token = $1`,
		accessToken,
	)

	if db.IsNoRows(err) {
		err = store.ErrAccessTokenNotFound.New(nil)
	}

	return
}

// DeleteAccessToken deletes the access token from the database.
func (s *oauthStore) DeleteAccessToken(accessToken string) error {
	return s.deleteAccessToken(s.queryer(), accessToken)
}

func (s *oauthStore) deleteAccessToken(q db.QueryContext, accessToken string) error {
	token := new(string)
	err := q.SelectOne(
		token,
		`DELETE
			FROM access_tokens
			WHERE access_token = $1
			RETURNING access_token`,
		accessToken,
	)

	if db.IsNoRows(err) {
		return store.ErrAccessTokenNotFound.New(nil)
	}

	return err
}

type refreshData struct {
	ClientUUID uuid.UUID
	UserUUID   uuid.UUID
	store.RefreshData
}

// SaveRefreshToken saves the refresh token.
func (s *oauthStore) SaveRefreshToken(data store.RefreshData) error {
	err := s.transact(func(tx *db.Tx) error {
		st := s.store()

		userID, err := st.Users.getUserID(tx, ttnpb.UserIdentifiers{UserID: data.UserID})
		if err != nil {
			return err
		}

		clientID, err := st.Clients.getClientID(tx, ttnpb.ClientIdentifiers{ClientID: data.ClientID})
		if err != nil {
			return err
		}

		return s.saveRefreshToken(tx, userID, clientID, data)
	})
	return err
}

func (s *oauthStore) saveRefreshToken(q db.QueryContext, userID, clientID uuid.UUID, data store.RefreshData) error {
	result := new(string)
	err := q.SelectOne(
		result,
		`INSERT
			INTO refresh_tokens (
				refresh_token,
				client_id,
				user_id,
				created_at,
				scope,
				redirect_uri
			)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING refresh_token`,
		data.RefreshToken,
		clientID,
		userID,
		data.CreatedAt,
		data.Scope,
		data.RedirectURI,
	)

	if _, dup := db.IsDuplicate(err); dup {
		return store.ErrRefreshTokenConflict.New(nil)
	}

	return err
}

// GetRefreshToken finds the refresh token.
func (s *oauthStore) GetRefreshToken(refreshToken string) (result store.RefreshData, err error) {
	err = s.transact(func(tx *db.Tx) error {
		data, err := s.getRefreshToken(tx, refreshToken)
		if err != nil {
			return err
		}

		st := s.store()

		user, err := st.Users.getUserIdentifiersFromID(tx, data.UserUUID)
		if err != nil {
			return err
		}
		data.RefreshData.UserID = user.UserID

		client, err := st.Clients.getClientIdentifiersFromID(tx, data.ClientUUID)
		if err != nil {
			return err
		}
		data.RefreshData.ClientID = client.ClientID

		result = data.RefreshData

		return nil
	})
	return
}

func (s *oauthStore) getRefreshToken(q db.QueryContext, refreshToken string) (result refreshData, err error) {
	err = q.SelectOne(
		&result,
		`SELECT
				refresh_token,
				client_id AS client_uuid,
				user_id AS user_uuid,
				created_at,
				scope,
				redirect_uri
			FROM refresh_tokens
			WHERE refresh_token = $1`,
		refreshToken,
	)

	if db.IsNoRows(err) {
		err = store.ErrRefreshTokenNotFound.New(nil)
	}

	return
}

// DeleteRefreshToken deletes the refresh token from the database.
func (s *oauthStore) DeleteRefreshToken(refreshToken string) error {
	return s.deleteRefreshToken(s.queryer(), refreshToken)
}

func (s *oauthStore) deleteRefreshToken(q db.QueryContext, refreshToken string) error {
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
		return store.ErrRefreshTokenNotFound.New(nil)
	}

	return err
}

// ListAuthorizedClients returns a list of clients authorized by a given user.
func (s *oauthStore) ListAuthorizedClients(ids ttnpb.UserIdentifiers, specializer store.ClientSpecializer) (result []store.Client, err error) {
	err = s.transact(func(tx *db.Tx) error {
		st := s.store()

		userID, err := st.Users.getUserID(tx, ids)
		if err != nil {
			return err
		}

		clientIDs, err := s.listAuthorizedClients(tx, userID)
		if err != nil {
			return err
		}

		for _, clientID := range clientIDs {
			client, err := st.Clients.getByID(tx, clientID)
			if err != nil {
				return err
			}

			specialized := specializer(client)

			err = st.Clients.loadAttributes(tx, clientID, specialized)
			if err != nil {
				return err
			}

			result = append(result, specialized)
		}

		return nil
	})
	return
}

func (s *oauthStore) listAuthorizedClients(q db.QueryContext, userID uuid.UUID) (ids []uuid.UUID, err error) {
	err = q.Select(
		&ids,
		`SELECT DISTINCT clients.id
			FROM clients
			JOIN refresh_tokens
			ON (
				clients.id = refresh_tokens.client_id AND refresh_tokens.user_id = $1
			)
			JOIN access_tokens
			ON (
				clients.id = access_tokens.client_id AND access_tokens.user_id = $1
			)`,
		userID)
	return
}

// RevokeAuthorizedClient deletes the access tokens and refresh token
// granted to a client by a given user.
func (s *oauthStore) RevokeAuthorizedClient(userIDs ttnpb.UserIdentifiers, clientIDs ttnpb.ClientIdentifiers) error {
	rows := 0
	err := s.transact(func(tx *db.Tx) error {
		st := s.store()

		userID, err := st.Users.getUserID(tx, userIDs)
		if err != nil {
			return err
		}

		clientID, err := st.Clients.getClientID(tx, clientIDs)
		if err != nil {
			return err
		}

		rowsa, err := s.deleteAccessTokensByUserAndClient(tx, userID, clientID)
		if err != nil {
			return err
		}

		rowsr, err := s.deleteRefreshTokenByUserAndClient(tx, userID, clientID)
		if err != nil {
			return err
		}

		rows = rowsa + rowsr

		return nil
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return store.ErrAuthorizedClientNotFound.New(nil)
	}
	return nil
}

func (s *oauthStore) deleteAccessTokensByUserAndClient(q db.QueryContext, userID, clientID uuid.UUID) (int, error) {
	res, err := q.Exec(
		`DELETE
			FROM access_tokens
			WHERE user_id = $1 AND client_id = $2`,
		userID,
		clientID)
	if err != nil {
		return 0, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(rows), nil
}

func (s *oauthStore) deleteRefreshTokenByUserAndClient(q db.QueryContext, userID, clientID uuid.UUID) (int, error) {
	res, err := q.Exec(
		`DELETE
			FROM refresh_tokens
			WHERE user_id = $1 AND client_id = $2`,
		userID,
		clientID)
	if err != nil {
		return 0, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(rows), nil
}

// IsClientAuthorized checks whether a client is currently authorized by an user.
func (s *oauthStore) IsClientAuthorized(user ttnpb.UserIdentifiers, client ttnpb.ClientIdentifiers) (res bool, err error) {
	err = s.transact(func(tx *db.Tx) error {
		st := s.store()

		userID, err := st.Users.getUserID(tx, user)
		if err != nil {
			return err
		}

		clientID, err := st.Clients.getClientID(tx, client)
		if err != nil {
			return err
		}

		res, err = s.isClientAuthorized(tx, userID, clientID)
		return err
	})
	return
}

func (s *oauthStore) isClientAuthorized(q db.QueryContext, userID, clientID uuid.UUID) (bool, error) {
	t := new(string)
	err := q.SelectOne(
		t,
		`SELECT
				refresh_token
			FROM refresh_tokens
			WHERE user_id = $1 AND client_id = $2
			LIMIT 1`,
		userID,
		clientID)
	if db.IsNoRows(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}
