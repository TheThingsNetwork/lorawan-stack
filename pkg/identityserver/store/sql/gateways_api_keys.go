// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/satori/go.uuid"
)

func (s *GatewayStore) SaveAPIKey(ids ttnpb.GatewayIdentifiers, key ttnpb.APIKey) error {
	err := s.transact(func(tx *db.Tx) error {
		gatewayID, err := s.getGatewayID(tx, ids)
		if err != nil {
			return err
		}

		err = s.saveAPIKey(tx, gatewayID, key)
		if err != nil {
			return err
		}

		return s.saveAPIKeyRights(tx, gatewayID, key.Name, key.Rights)
	})
	return err
}

func (s *GatewayStore) GetAPIKey(value string) (ids ttnpb.GatewayIdentifiers, key ttnpb.APIKey, err error) {
	err = s.transact(func(tx *db.Tx) error {
		var gtwID uuid.UUID
		gtwID, key, err = s.getAPIKey(tx, value)
		if err != nil {
			return err
		}

		key.Rights, err = s.getAPIKeyRights(tx, gtwID, key.Name)
		if err != nil {
			return err
		}

		ids, err = s.getGatewayIdentifiersFromID(tx, gtwID)
		if err != nil {
			return err
		}

		return nil
	})
	return
}

func (s *GatewayStore) GetAPIKeyByName(ids ttnpb.GatewayIdentifiers, keyName string) (key ttnpb.APIKey, err error) {
	err = s.transact(func(tx *db.Tx) error {
		gatewayID, err := s.getGatewayID(tx, ids)
		if err != nil {
			return err
		}

		key, err = s.getAPIKeyByName(tx, gatewayID, keyName)
		if err != nil {
			return err
		}

		key.Rights, err = s.getAPIKeyRights(tx, gatewayID, keyName)

		return err
	})
	return
}

func (s *GatewayStore) ListAPIKeys(ids ttnpb.GatewayIdentifiers) (keys []ttnpb.APIKey, err error) {
	err = s.transact(func(tx *db.Tx) error {
		gatewayID, err := s.getGatewayID(tx, ids)
		if err != nil {
			return err
		}

		keys, err = s.listAPIKeys(tx, gatewayID)
		if err != nil {
			return err
		}

		for i, key := range keys {
			keys[i].Rights, err = s.getAPIKeyRights(tx, gatewayID, key.Name)
			if err != nil {
				return err
			}
		}

		return nil
	})
	return
}

func (s *GatewayStore) UpdateAPIKeyRights(ids ttnpb.GatewayIdentifiers, keyName string, rights []ttnpb.Right) error {
	err := s.transact(func(tx *db.Tx) error {
		gatewayID, err := s.getGatewayID(tx, ids)
		if err != nil {
			return err
		}

		err = s.deleteAPIKeyRights(tx, gatewayID, keyName)
		if err != nil {
			return err
		}

		return s.saveAPIKeyRights(tx, gatewayID, keyName, rights)
	})
	return err
}

func (s *GatewayStore) DeleteAPIKey(ids ttnpb.GatewayIdentifiers, keyName string) error {
	err := s.transact(func(tx *db.Tx) error {
		gatewayID, err := s.getGatewayID(tx, ids)
		if err != nil {
			return err
		}

		err = s.deleteAPIKeyRights(tx, gatewayID, keyName)
		if err != nil {
			return err
		}

		return s.deleteAPIKey(tx, gatewayID, keyName)
	})
	return err
}
