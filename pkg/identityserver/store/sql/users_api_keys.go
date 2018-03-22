// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/satori/go.uuid"
)

func (s *UserStore) SaveAPIKey(ids ttnpb.UserIdentifiers, key ttnpb.APIKey) error {
	err := s.transact(func(tx *db.Tx) error {
		userID, err := s.getUserID(tx, ids)
		if err != nil {
			return err
		}

		err = s.saveAPIKey(tx, userID, key)
		if err != nil {
			return err
		}

		return s.saveAPIKeyRights(tx, userID, key.Name, key.Rights)
	})
	return err
}

func (s *UserStore) GetAPIKey(value string) (ids ttnpb.UserIdentifiers, key ttnpb.APIKey, err error) {
	err = s.transact(func(tx *db.Tx) error {
		var userID uuid.UUID
		userID, key, err = s.getAPIKey(tx, value)
		if err != nil {
			return err
		}

		key.Rights, err = s.getAPIKeyRights(tx, userID, key.Name)
		if err != nil {
			return err
		}

		ids, err = s.getUserIdentifiersFromID(tx, userID)
		if err != nil {
			return err
		}

		return nil
	})
	return
}

func (s *UserStore) GetAPIKeyByName(ids ttnpb.UserIdentifiers, keyName string) (key ttnpb.APIKey, err error) {
	err = s.transact(func(tx *db.Tx) error {
		userID, err := s.getUserID(tx, ids)
		if err != nil {
			return err
		}

		key, err = s.getAPIKeyByName(tx, userID, keyName)
		if err != nil {
			return err
		}

		key.Rights, err = s.getAPIKeyRights(tx, userID, keyName)

		return err
	})
	return
}

func (s *UserStore) ListAPIKeys(ids ttnpb.UserIdentifiers) (keys []ttnpb.APIKey, err error) {
	err = s.transact(func(tx *db.Tx) error {
		userID, err := s.getUserID(tx, ids)
		if err != nil {
			return err
		}

		keys, err = s.listAPIKeys(tx, userID)
		if err != nil {
			return err
		}

		for i, key := range keys {
			keys[i].Rights, err = s.getAPIKeyRights(tx, userID, key.Name)
			if err != nil {
				return err
			}
		}

		return nil
	})
	return
}

func (s *UserStore) UpdateAPIKeyRights(ids ttnpb.UserIdentifiers, keyName string, rights []ttnpb.Right) error {
	err := s.transact(func(tx *db.Tx) error {
		userID, err := s.getUserID(tx, ids)
		if err != nil {
			return err
		}

		err = s.deleteAPIKeyRights(tx, userID, keyName)
		if err != nil {
			return err
		}

		return s.saveAPIKeyRights(tx, userID, keyName, rights)
	})
	return err
}

func (s *UserStore) DeleteAPIKey(ids ttnpb.UserIdentifiers, keyName string) error {
	err := s.transact(func(tx *db.Tx) error {
		userID, err := s.getUserID(tx, ids)
		if err != nil {
			return err
		}

		err = s.deleteAPIKeyRights(tx, userID, keyName)
		if err != nil {
			return err
		}

		return s.deleteAPIKey(tx, userID, keyName)
	})
	return err
}
