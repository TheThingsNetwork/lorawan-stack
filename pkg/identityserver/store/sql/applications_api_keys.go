// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/satori/go.uuid"
)

func (s *ApplicationStore) SaveAPIKey(ids ttnpb.ApplicationIdentifiers, key ttnpb.APIKey) error {
	err := s.transact(func(tx *db.Tx) error {
		applicationID, err := s.getApplicationID(tx, ids)
		if err != nil {
			return err
		}

		err = s.saveAPIKey(tx, applicationID, key)
		if err != nil {
			return err
		}

		return s.saveAPIKeyRights(tx, applicationID, key.Name, key.Rights)
	})
	return err
}

func (s *ApplicationStore) GetAPIKey(value string) (ids ttnpb.ApplicationIdentifiers, key ttnpb.APIKey, err error) {
	err = s.transact(func(tx *db.Tx) error {
		var appID uuid.UUID
		appID, key, err = s.getAPIKey(tx, value)
		if err != nil {
			return err
		}

		key.Rights, err = s.getAPIKeyRights(tx, appID, key.Name)
		if err != nil {
			return err
		}

		ids, err = s.getApplicationIdentifiersFromID(tx, appID)
		if err != nil {
			return err
		}

		return nil
	})
	return
}

func (s *ApplicationStore) GetAPIKeyByName(ids ttnpb.ApplicationIdentifiers, keyName string) (key ttnpb.APIKey, err error) {
	err = s.transact(func(tx *db.Tx) error {
		applicationID, err := s.getApplicationID(tx, ids)
		if err != nil {
			return err
		}

		key, err = s.getAPIKeyByName(tx, applicationID, keyName)
		if err != nil {
			return err
		}

		key.Rights, err = s.getAPIKeyRights(tx, applicationID, keyName)

		return err
	})
	return
}

func (s *ApplicationStore) ListAPIKeys(ids ttnpb.ApplicationIdentifiers) (keys []ttnpb.APIKey, err error) {
	err = s.transact(func(tx *db.Tx) error {
		applicationID, err := s.getApplicationID(tx, ids)
		if err != nil {
			return err
		}

		keys, err = s.listAPIKeys(tx, applicationID)
		if err != nil {
			return err
		}

		for i, key := range keys {
			keys[i].Rights, err = s.getAPIKeyRights(tx, applicationID, key.Name)
			if err != nil {
				return err
			}
		}

		return nil
	})
	return
}

func (s *ApplicationStore) UpdateAPIKeyRights(ids ttnpb.ApplicationIdentifiers, keyName string, rights []ttnpb.Right) error {
	err := s.transact(func(tx *db.Tx) error {
		applicationID, err := s.getApplicationID(tx, ids)
		if err != nil {
			return err
		}

		err = s.deleteAPIKeyRights(tx, applicationID, keyName)
		if err != nil {
			return err
		}

		return s.saveAPIKeyRights(tx, applicationID, keyName, rights)
	})
	return err
}

func (s *ApplicationStore) DeleteAPIKey(ids ttnpb.ApplicationIdentifiers, keyName string) error {
	err := s.transact(func(tx *db.Tx) error {
		applicationID, err := s.getApplicationID(tx, ids)
		if err != nil {
			return err
		}

		err = s.deleteAPIKeyRights(tx, applicationID, keyName)
		if err != nil {
			return err
		}

		return s.deleteAPIKey(tx, applicationID, keyName)
	})
	return err
}
