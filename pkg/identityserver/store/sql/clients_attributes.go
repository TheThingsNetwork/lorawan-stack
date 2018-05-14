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
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// LoadAttributes loads the extra attributes in app if it is a store.Attributer.
func (s *clientStore) LoadAttributes(ids ttnpb.ClientIdentifiers, app store.Client) error {
	attr, ok := app.(store.Attributer)
	if !ok {
		return nil
	}

	err := s.transact(func(tx *db.Tx) error {
		clientID, err := s.getClientID(tx, ids)
		if err != nil {
			return err
		}

		return s.extraAttributesStore.loadAttributes(tx, clientID, attr)
	})

	return err
}

func (s *clientStore) loadAttributes(q db.QueryContext, clientID uuid.UUID, client store.Client) error {
	attr, ok := client.(store.Attributer)
	if !ok {
		return nil
	}

	return s.extraAttributesStore.loadAttributes(q, clientID, attr)
}

// StoreAttributes store the extra attributes of app if it is a store.Attributer
// and writes the resulting client in result.
func (s *clientStore) StoreAttributes(ids ttnpb.ClientIdentifiers, client store.Client) (err error) {
	_, ok := client.(store.Attributer)
	if !ok {
		return nil
	}

	err = s.transact(func(tx *db.Tx) error {
		clientID, err := s.getClientID(tx, ids)
		if err != nil {
			return err
		}

		return s.storeAttributes(tx, clientID, client)
	})

	return
}

func (s *clientStore) storeAttributes(q db.QueryContext, clientID uuid.UUID, client store.Client) error {
	attr, ok := client.(store.Attributer)
	if !ok {
		return nil
	}

	return s.extraAttributesStore.storeAttributes(q, clientID, attr)
}
