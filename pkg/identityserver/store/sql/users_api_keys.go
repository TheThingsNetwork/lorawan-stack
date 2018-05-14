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

// SaveAPIKey stores an API Key attached to an user.
func (s *userStore) SaveAPIKey(ids ttnpb.UserIdentifiers, key ttnpb.APIKey) error {
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

// GetAPIKey retrieves an API key by value and the user identifiers.
func (s *userStore) GetAPIKey(value string) (ids ttnpb.UserIdentifiers, key ttnpb.APIKey, err error) {
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

// GetAPIKeyByName retrieves an API key from an user.
func (s *userStore) GetAPIKeyByName(ids ttnpb.UserIdentifiers, keyName string) (key ttnpb.APIKey, err error) {
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

// UpdateAPIKeyRights updates the right of an API key.
func (s *userStore) UpdateAPIKeyRights(ids ttnpb.UserIdentifiers, keyName string, rights []ttnpb.Right) error {
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

// ListAPIKeys list all the API keys that an user has.
func (s *userStore) ListAPIKeys(ids ttnpb.UserIdentifiers) (keys []ttnpb.APIKey, err error) {
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

// DeleteAPIKey deletes a given API key from an user.
func (s *userStore) DeleteAPIKey(ids ttnpb.UserIdentifiers, keyName string) error {
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
