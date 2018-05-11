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

// SaveAPIKey stores an API Key attached to an application.
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

// GetAPIKey retrieves an API key by value and the appplication identifiers.
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

// GetAPIKeyByName retrieves an API key from an application.
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

// ListAPIKeys list all the API keys that an application has.
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

// UpdateAPIKeyRights updates the right of an API key.
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

// DeleteAPIKey deletes a given API key from an application.
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
