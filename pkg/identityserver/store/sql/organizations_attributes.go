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
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/satori/go.uuid"
)

// LoadAttributes loads the extra attributes in app if it is a store.Attributer.
func (s *OrganizationStore) LoadAttributes(ids ttnpb.OrganizationIdentifiers, app store.Organization) error {
	attr, ok := app.(store.Attributer)
	if !ok {
		return nil
	}

	err := s.transact(func(tx *db.Tx) error {
		orgID, err := s.getOrganizationID(tx, ids)
		if err != nil {
			return err
		}

		return s.extraAttributesStore.loadAttributes(tx, orgID, attr)
	})

	return err
}

func (s *OrganizationStore) loadAttributes(q db.QueryContext, orgID uuid.UUID, organization store.Organization) error {
	attr, ok := organization.(store.Attributer)
	if !ok {
		return nil
	}

	return s.extraAttributesStore.loadAttributes(q, orgID, attr)
}

// StoreAttributes store the extra attributes of app if it is a store.Attributer
// and writes the resulting organization in result.
func (s *OrganizationStore) StoreAttributes(ids ttnpb.OrganizationIdentifiers, organization store.Organization) (err error) {
	_, ok := organization.(store.Attributer)
	if !ok {
		return nil
	}

	err = s.transact(func(tx *db.Tx) error {
		orgID, err := s.getOrganizationID(tx, ids)
		if err != nil {
			return err
		}

		return s.storeAttributes(tx, orgID, organization)
	})

	return
}

func (s *OrganizationStore) storeAttributes(q db.QueryContext, orgID uuid.UUID, organization store.Organization) error {
	attr, ok := organization.(store.Attributer)
	if !ok {
		return nil
	}

	return s.extraAttributesStore.storeAttributes(q, orgID, attr)
}
