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

// SaveAPIKey stores an API Key attached to an organization.
func (s *organizationStore) SaveAPIKey(ids ttnpb.OrganizationIdentifiers, key ttnpb.APIKey) error {
	err := s.transact(func(tx *db.Tx) error {
		orgID, err := s.getOrganizationID(tx, ids)
		if err != nil {
			return err
		}

		err = s.saveAPIKey(tx, orgID, key)
		if err != nil {
			return err
		}

		return s.saveAPIKeyRights(tx, orgID, key.Name, key.Rights)
	})
	return err
}

// GetAPIKey retrieves an API key by value and the organization identifiers.
func (s *organizationStore) GetAPIKey(value string) (ids ttnpb.OrganizationIdentifiers, key ttnpb.APIKey, err error) {
	err = s.transact(func(tx *db.Tx) error {
		var orgID uuid.UUID
		orgID, key, err = s.getAPIKey(tx, value)
		if err != nil {
			return err
		}

		key.Rights, err = s.getAPIKeyRights(tx, orgID, key.Name)
		if err != nil {
			return err
		}

		ids, err = s.getOrganizationIdentifiersFromID(tx, orgID)
		if err != nil {
			return err
		}

		return nil
	})
	return
}

// GetAPIKeyByName retrieves an API key from an organization.
func (s *organizationStore) GetAPIKeyByName(ids ttnpb.OrganizationIdentifiers, keyName string) (key ttnpb.APIKey, err error) {
	err = s.transact(func(tx *db.Tx) error {
		orgID, err := s.getOrganizationID(tx, ids)
		if err != nil {
			return err
		}

		key, err = s.getAPIKeyByName(tx, orgID, keyName)
		if err != nil {
			return err
		}

		key.Rights, err = s.getAPIKeyRights(tx, orgID, keyName)

		return err
	})
	return
}

// ListAPIKeys list all the API keys that an organization has.
func (s *organizationStore) ListAPIKeys(ids ttnpb.OrganizationIdentifiers) (keys []ttnpb.APIKey, err error) {
	err = s.transact(func(tx *db.Tx) error {
		orgID, err := s.getOrganizationID(tx, ids)
		if err != nil {
			return err
		}

		keys, err = s.listAPIKeys(tx, orgID)
		if err != nil {
			return err
		}

		for i, key := range keys {
			keys[i].Rights, err = s.getAPIKeyRights(tx, orgID, key.Name)
			if err != nil {
				return err
			}
		}

		return nil
	})
	return
}

// UpdateAPIKeyRights updates the right of an API key.
func (s *organizationStore) UpdateAPIKeyRights(ids ttnpb.OrganizationIdentifiers, keyName string, rights []ttnpb.Right) error {
	err := s.transact(func(tx *db.Tx) error {
		orgID, err := s.getOrganizationID(tx, ids)
		if err != nil {
			return err
		}

		err = s.deleteAPIKeyRights(tx, orgID, keyName)
		if err != nil {
			return err
		}

		return s.saveAPIKeyRights(tx, orgID, keyName, rights)
	})
	return err
}

// DeleteAPIKey deletes a given API key from an organization.
func (s *organizationStore) DeleteAPIKey(ids ttnpb.OrganizationIdentifiers, keyName string) error {
	err := s.transact(func(tx *db.Tx) error {
		orgID, err := s.getOrganizationID(tx, ids)
		if err != nil {
			return err
		}

		err = s.deleteAPIKeyRights(tx, orgID, keyName)
		if err != nil {
			return err
		}

		return s.deleteAPIKey(tx, orgID, keyName)
	})
	return err
}
