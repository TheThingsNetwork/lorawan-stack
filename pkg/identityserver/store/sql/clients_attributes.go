// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/satori/go.uuid"
)

// LoadAttributes loads the extra attributes in app if it is a store.Attributer.
func (s *ClientStore) LoadAttributes(ids ttnpb.ClientIdentifiers, app store.Client) error {
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

func (s *ClientStore) loadAttributes(q db.QueryContext, clientID uuid.UUID, client store.Client) error {
	attr, ok := client.(store.Attributer)
	if !ok {
		return nil
	}

	return s.extraAttributesStore.loadAttributes(q, clientID, attr)
}

// StoreAttributes store the extra attributes of app if it is a store.Attributer
// and writes the resulting client in result.
func (s *ClientStore) StoreAttributes(ids ttnpb.ClientIdentifiers, client store.Client) (err error) {
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

func (s *ClientStore) storeAttributes(q db.QueryContext, clientID uuid.UUID, client store.Client) error {
	attr, ok := client.(store.Attributer)
	if !ok {
		return nil
	}

	return s.extraAttributesStore.storeAttributes(q, clientID, attr)
}
