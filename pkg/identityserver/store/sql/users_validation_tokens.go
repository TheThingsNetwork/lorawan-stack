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

// SaveValidationToken saves the validation token.
func (s *userStore) SaveValidationToken(ids ttnpb.UserIdentifiers, token store.ValidationToken) (err error) {
	err = s.transact(func(tx *db.Tx) error {
		userID, err := s.getUserID(tx, ids)
		if err != nil {
			return err
		}

		return s.saveValidationToken(tx, userID, token)
	})
	return
}

func (s *userStore) saveValidationToken(q db.QueryContext, userID uuid.UUID, token store.ValidationToken) (err error) {
	_, err = q.Exec(
		`INSERT
			INTO validation_tokens (
				validation_token,
				user_id,
				created_at,
				expires_in
			)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id)
			DO UPDATE SET
				validation_token = excluded.validation_token,
				created_at = excluded.created_at,
				expires_in = excluded.expires_in`,
		token.ValidationToken,
		userID,
		token.CreatedAt,
		token.ExpiresIn)
	return
}

// GetValidationToken retrieves the validation token.
func (s *userStore) GetValidationToken(token string) (identifier ttnpb.UserIdentifiers, data *store.ValidationToken, err error) {
	err = s.transact(func(tx *db.Tx) error {
		var userID uuid.UUID
		userID, data, err = s.getValidationToken(tx, token)
		if err != nil {
			return err
		}

		identifier, err = s.getUserIdentifiersFromID(tx, userID)
		return err
	})
	return
}

func (s *userStore) getValidationToken(q db.QueryContext, token string) (id uuid.UUID, data *store.ValidationToken, err error) {
	var res struct {
		UserID uuid.UUID
		*store.ValidationToken
	}
	err = q.SelectOne(
		&res,
		`SELECT
				user_id,
				validation_token,
				created_at,
				expires_in
			FROM validation_tokens
			WHERE validation_token = $1`,
		token)
	if db.IsNoRows(err) {
		err = store.ErrValidationTokenNotFound.New(nil)
	}
	id = res.UserID
	data = res.ValidationToken
	return
}

// DeleteValidationToken deletes the validation token.
func (s *userStore) DeleteValidationToken(token string) error {
	return s.deleteValidationToken(s.queryer(), token)
}

func (s *userStore) deleteValidationToken(q db.QueryContext, token string) (err error) {
	t := new(string)
	err = q.SelectOne(
		t,
		`DELETE
			FROM validation_tokens
			WHERE validation_token = $1
			RETURNING validation_token`,
		token)
	if db.IsNoRows(err) {
		err = store.ErrValidationTokenNotFound.New(nil)
	}
	return
}
