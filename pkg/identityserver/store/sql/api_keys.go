// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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

		return s.saveAPIKeyRights(tx, key.Key, key.Rights)
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

func (s *apiKeysStore) saveAPIKeyRights(q db.QueryContext, key string, rights []ttnpb.Right) error {
	query, args := s.saveAPIKeyRightsQuery(key, rights)
	_, err := q.Exec(query, args...)
	return err
}

func (s *apiKeysStore) saveAPIKeyRightsQuery(key string, rights []ttnpb.Right) (string, []interface{}) {
	args := make([]interface{}, 1+len(rights))
	args[0] = key

	boundValues := make([]string, len(rights))

	for i, right := range rights {
		args[i+1] = right
		boundValues[i] = fmt.Sprintf("($1, $%d)", i+2)
	}

	query := fmt.Sprintf(`
		INSERT
			INTO %ss_api_keys_rights (key, "right")
			VALUES %s
			ON CONFLICT (key, "right")
			DO NOTHING`,
		s.entity,
		strings.Join(boundValues, ", "))

	return query, args
}

func (s *apiKeysStore) GetAPIKey(entityID, keyName string) (*ttnpb.APIKey, error) {
	var key *ttnpb.APIKey
	var err error
	err = s.transact(func(tx *db.Tx) error {
		key, err = s.getAPIKey(tx, entityID, keyName)
		if err != nil {
			return err
		}

		key.Rights, err = s.getAPIKeyRights(tx, key.Key)
		return err
	})
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (s *apiKeysStore) getAPIKey(q db.QueryContext, entityID, keyName string) (*ttnpb.APIKey, error) {
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

func (s *apiKeysStore) getAPIKeyRights(q db.QueryContext, key string) ([]ttnpb.Right, error) {
	var res []ttnpb.Right
	err := q.Select(
		&res,
		fmt.Sprintf(`
			SELECT
				"right"
			FROM %ss_api_keys_rights
			WHERE key = $1`, s.entity),
		key)
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
			res[i].Rights, err = s.getAPIKeyRights(tx, key.Key)
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
		key, err := s.getAPIKey(tx, entityID, keyName)
		if err != nil {
			return err
		}

		err = s.deleteAPIKeyRights(tx, key.Key)
		if err != nil {
			return err
		}

		return s.saveAPIKeyRights(tx, key.Key, rights)
	})
	return err
}

func (s *apiKeysStore) DeleteAPIKey(entityID, keyName string) error {
	err := s.transact(func(tx *db.Tx) error {
		key, err := s.getAPIKey(tx, entityID, keyName)
		if err != nil {
			return err
		}

		err = s.deleteAPIKeyRights(tx, key.Key)
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

func (s *apiKeysStore) deleteAPIKeyRights(q db.QueryContext, key string) error {
	_, err := q.Exec(
		fmt.Sprintf(`
			DELETE
				FROM %ss_api_keys_rights
				WHERE key = $1`, s.entity),
		key)
	return err
}
