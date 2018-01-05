// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
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

func (s *apiKeysStore) SaveAPIKey(entityID string, key *ttnpb.APIKey) error {
	err := s.transact(func(tx *db.Tx) error {
		err := s.saveAPIKey(tx, entityID, key)
		if err != nil {
			return err
		}

		return s.saveAPIKeyRights(tx, entityID, key.Name, key.Rights)
	})
	return err
}

func (s *apiKeysStore) saveAPIKey(q db.QueryContext, entityID string, key *ttnpb.APIKey) error {
	_, err := q.Exec(
		fmt.Sprintf(`
			INSERT
				INTO %ss_api_keys (%s, key, key_name)
				VALUES ($1, $2, $3)`, s.entity, s.foreignKey),
		entityID,
		key.Key,
		key.Name)
	if _, yes := db.IsDuplicate(err); yes {
		return ErrAPIKeyNameConflict.New(errors.Attributes{
			"name": key.Name,
		})
	}
	return err
}

func (s *apiKeysStore) saveAPIKeyRights(q db.QueryContext, entityID, keyName string, rights []ttnpb.Right) error {
	query, args := s.saveAPIKeyRightsQuery(entityID, keyName, rights)
	_, err := q.Exec(query, args...)
	return err
}

func (s *apiKeysStore) saveAPIKeyRightsQuery(entityID, keyName string, rights []ttnpb.Right) (string, []interface{}) {
	args := make([]interface{}, 2+len(rights))
	args[0] = entityID
	args[1] = keyName

	boundValues := make([]string, len(rights))

	for i, right := range rights {
		args[i+2] = right
		boundValues[i] = fmt.Sprintf("($1, $2, $%d)", i+3)
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

func (s *apiKeysStore) GetAPIKey(key string) (string, *ttnpb.APIKey, error) {
	var entityID string
	var apiKey *ttnpb.APIKey
	var err error
	err = s.transact(func(tx *db.Tx) error {
		entityID, apiKey, err = s.getAPIKey(tx, key)
		if err != nil {
			return err
		}

		apiKey.Rights, err = s.getAPIKeyRights(tx, entityID, apiKey.Name)
		return err
	})
	if err != nil {
		return "", nil, err
	}
	return entityID, apiKey, nil
}

func (s *apiKeysStore) getAPIKey(q db.QueryContext, key string) (string, *ttnpb.APIKey, error) {
	var res struct {
		EntityID string
		*ttnpb.APIKey
	}
	err := q.SelectOne(
		res,
		fmt.Sprintf(`
			SELECT
				key,
				key_name AS name,
				%s AS entity_id
			FROM %ss_api_keys
			WHERE key = $1`, s.foreignKey, s.entity),
		key)
	if db.IsNoRows(err) {
		return "", nil, ErrAPIKeyNotFound.New(nil)
	}
	if err != nil {
		return "", nil, err
	}
	return res.EntityID, res.APIKey, nil
}

func (s *apiKeysStore) GetAPIKeyByName(entityID, keyName string) (*ttnpb.APIKey, error) {
	var key *ttnpb.APIKey
	var err error
	err = s.transact(func(tx *db.Tx) error {
		key, err = s.getAPIKeyByName(tx, entityID, keyName)
		if err != nil {
			return err
		}

		key.Rights, err = s.getAPIKeyRights(tx, entityID, keyName)
		return err
	})
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (s *apiKeysStore) getAPIKeyByName(q db.QueryContext, entityID, keyName string) (*ttnpb.APIKey, error) {
	res := new(ttnpb.APIKey)
	err := q.SelectOne(
		res,
		fmt.Sprintf(`
			SELECT
				key,
				key_name AS name
			FROM %ss_api_keys
			WHERE %s = $1 AND key_name = $2`, s.entity, s.foreignKey),
		entityID,
		keyName)
	if db.IsNoRows(err) {
		return nil, ErrAPIKeyNotFound.New(errors.Attributes{
			"name": keyName,
		})
	}
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *apiKeysStore) getAPIKeyRights(q db.QueryContext, entityID, keyName string) ([]ttnpb.Right, error) {
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

func (s *apiKeysStore) ListAPIKeys(entityID string) ([]*ttnpb.APIKey, error) {
	var res []*ttnpb.APIKey
	var err error
	err = s.transact(func(tx *db.Tx) error {
		res, err = s.listAPIKeys(tx, entityID)
		if err != nil {
			return err
		}

		for i, key := range res {
			res[i].Rights, err = s.getAPIKeyRights(tx, entityID, key.Name)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *apiKeysStore) listAPIKeys(q db.QueryContext, entityID string) ([]*ttnpb.APIKey, error) {
	var res []*ttnpb.APIKey
	err := q.Select(
		&res,
		fmt.Sprintf(`
			SELECT
				key,
				key_name AS name
			FROM %ss_api_keys
			WHERE %s = $1`, s.entity, s.foreignKey),
		entityID)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *apiKeysStore) UpdateAPIKeyRights(entityID, keyName string, rights []ttnpb.Right) error {
	err := s.transact(func(tx *db.Tx) error {
		err := s.deleteAPIKeyRights(tx, entityID, keyName)
		if err != nil {
			return err
		}

		return s.saveAPIKeyRights(tx, entityID, keyName, rights)
	})
	return err
}

func (s *apiKeysStore) DeleteAPIKey(entityID, keyName string) error {
	err := s.transact(func(tx *db.Tx) error {
		err := s.deleteAPIKeyRights(tx, entityID, keyName)
		if err != nil {
			return err
		}

		return s.deleteAPIKey(tx, entityID, keyName)
	})
	return err
}

func (s *apiKeysStore) deleteAPIKey(q db.QueryContext, entityID, keyName string) error {
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
		return ErrAPIKeyNotFound.New(errors.Attributes{
			"name": keyName,
		})
	}
	return err
}

func (s *apiKeysStore) deleteAPIKeyRights(q db.QueryContext, entityID, keyName string) error {
	_, err := q.Exec(
		fmt.Sprintf(`
			DELETE
				FROM %ss_api_keys_rights
				WHERE %s = $1 AND key_name = $2`, s.entity, s.foreignKey),
		entityID,
		keyName)
	return err
}

func (s *apiKeysStore) deleteAPIKeys(q db.QueryContext, entityID string) error {
	_, err := q.Exec(
		fmt.Sprintf(`
			DELETE
				FROM %ss_api_keys_rights
				WHERE %s = $1`, s.entity, s.foreignKey),
		entityID)
	if err != nil {
		return err
	}

	_, err = q.Exec(
		fmt.Sprintf(`
			DELETE
				FROM %ss_api_keys
				WHERE %s = $1`, s.entity, s.foreignKey),
		entityID)

	return err
}
