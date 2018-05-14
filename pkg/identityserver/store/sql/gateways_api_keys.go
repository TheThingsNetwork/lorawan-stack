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
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// SaveAPIKey stores an API Key attached to a gateway.
func (s *gatewayStore) SaveAPIKey(ids ttnpb.GatewayIdentifiers, key ttnpb.APIKey) error {
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

// GetAPIKey retrieves an API key by value and the gateway identifiers.
func (s *gatewayStore) GetAPIKey(value string) (ids ttnpb.GatewayIdentifiers, key ttnpb.APIKey, err error) {
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

// GetAPIKeyByName retrieves an API key from a gateway.
func (s *gatewayStore) GetAPIKeyByName(ids ttnpb.GatewayIdentifiers, keyName string) (key ttnpb.APIKey, err error) {
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

// ListAPIKeys list all the API keys that a gateway has.
func (s *gatewayStore) ListAPIKeys(ids ttnpb.GatewayIdentifiers) (keys []ttnpb.APIKey, err error) {
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

// UpdateAPIKeyRights updates the right of an API key.
func (s *gatewayStore) UpdateAPIKeyRights(ids ttnpb.GatewayIdentifiers, keyName string, rights []ttnpb.Right) error {
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

// DeleteAPIKey deletes a given API key from a gateway.
func (s *gatewayStore) DeleteAPIKey(ids ttnpb.GatewayIdentifiers, keyName string) error {
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
