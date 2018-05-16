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
	"fmt"
	"strings"

	"github.com/satori/go.uuid"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/identityserver/db"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type apiKeysStore struct {
	storer
	entity     string
	foreignKey string
}

func newAPIKeysStore(store storer, entity string) *apiKeysStore {
	return &apiKeysStore{
		storer:     store,
		entity:     entity,
		foreignKey: fmt.Sprintf("%s_id", entity),
	}
}

func (s *apiKeysStore) saveAPIKey(q db.QueryContext, entityID uuid.UUID, key ttnpb.APIKey) error {
	_, err := q.Exec(
		fmt.Sprintf(`
			INSERT
				INTO %ss_api_keys (%s, key, key_name)
				VALUES ($1, $2, $3)`, s.entity, s.foreignKey),
		entityID,
		key.Key,
		key.Name)
	if _, yes := db.IsDuplicate(err); yes {
		return store.ErrAPIKeyNameConflict.New(errors.Attributes{
			"name": key.Name,
		})
	}
	return err
}

func (s *apiKeysStore) saveAPIKeyRights(q db.QueryContext, entityID uuid.UUID, keyName string, rights []ttnpb.Right) error {
	query, args := s.saveAPIKeyRightsQuery(entityID, keyName, rights)
	_, err := q.Exec(query, args...)
	return err
}

func (s *apiKeysStore) saveAPIKeyRightsQuery(entityID uuid.UUID, keyName string, rights []ttnpb.Right) (string, []interface{}) {
	args := make([]interface{}, 0, 2+len(rights))
	args = append(args, entityID, keyName)

	boundValues := make([]string, 0, len(rights))
	for i, right := range rights {
		args = append(args, right)
		boundValues = append(boundValues, fmt.Sprintf("($1, $2, $%d)", i+3))
	}

	query := fmt.Sprintf(`
		INSERT
			INTO %ss_api_keys_rights (%s, key_name, "right")
			VALUES %s
			ON CONFLICT (%s, key_name, "right")
			DO NOTHING`,
		s.entity,
		s.foreignKey,
		strings.Join(boundValues, ", "),
		s.foreignKey)

	return query, args
}

func (s *apiKeysStore) getAPIKey(q db.QueryContext, value string) (id uuid.UUID, key ttnpb.APIKey, err error) {
	var res struct {
		EntityID uuid.UUID
		ttnpb.APIKey
	}
	err = q.SelectOne(
		&res,
		fmt.Sprintf(`
			SELECT
				key,
				key_name AS name,
				%s AS entity_id
			FROM %ss_api_keys
			WHERE key = $1`, s.foreignKey, s.entity),
		value)
	if db.IsNoRows(err) {
		err = store.ErrAPIKeyNotFound.New(nil)
	}
	id = res.EntityID
	key = res.APIKey
	return
}

func (s *apiKeysStore) getAPIKeyByName(q db.QueryContext, entityID uuid.UUID, keyName string) (key ttnpb.APIKey, err error) {
	err = q.SelectOne(
		&key,
		fmt.Sprintf(`
			SELECT
				key,
				key_name AS name
			FROM %ss_api_keys
			WHERE %s = $1 AND key_name = $2`, s.entity, s.foreignKey),
		entityID,
		keyName)
	if db.IsNoRows(err) {
		err = store.ErrAPIKeyNotFound.New(nil)
	}
	return
}

func (s *apiKeysStore) getAPIKeyRights(q db.QueryContext, entityID uuid.UUID, keyName string) ([]ttnpb.Right, error) {
	var res []ttnpb.Right
	err := q.Select(
		&res,
		fmt.Sprintf(`
			SELECT
				"right"
			FROM %ss_api_keys_rights
			WHERE %s = $1 AND key_name = $2`, s.entity, s.foreignKey),
		entityID,
		keyName)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *apiKeysStore) listAPIKeys(q db.QueryContext, entityID uuid.UUID) (res []ttnpb.APIKey, err error) {
	err = q.Select(
		&res,
		fmt.Sprintf(`
			SELECT
				key,
				key_name AS name
			FROM %ss_api_keys
			WHERE %s = $1`, s.entity, s.foreignKey),
		entityID)
	return
}

func (s *apiKeysStore) deleteAPIKey(q db.QueryContext, entityID uuid.UUID, keyName string) error {
	res := new(string)
	err := q.SelectOne(
		res,
		fmt.Sprintf(`
			DELETE
				FROM %ss_api_keys
				WHERE %s = $1 AND key_name = $2
				RETURNING key`, s.entity, s.foreignKey),
		entityID,
		keyName)
	if db.IsNoRows(err) {
		return store.ErrAPIKeyNotFound.New(nil)
	}
	return err
}

func (s *apiKeysStore) deleteAPIKeyRights(q db.QueryContext, entityID uuid.UUID, keyName string) error {
	var n string
	err := q.SelectOne(
		&n,
		fmt.Sprintf(`
			DELETE
				FROM %ss_api_keys_rights
				WHERE %s = $1 AND key_name = $2
				RETURNING key_name`, s.entity, s.foreignKey),
		entityID,
		keyName)
	if db.IsNoRows(err) {
		return store.ErrAPIKeyNotFound.New(nil)
	}
	return err
}
