// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/satori/go.uuid"
)

// LoadAttributes loads the extra attributes in app if it is a store.Attributer.
func (s *GatewayStore) LoadAttributes(ids ttnpb.GatewayIdentifiers, app store.Gateway) error {
	attr, ok := app.(store.Attributer)
	if !ok {
		return nil
	}

	err := s.transact(func(tx *db.Tx) error {
		gtwID, err := s.getGatewayID(tx, ids)
		if err != nil {
			return err
		}

		return s.extraAttributesStore.loadAttributes(tx, gtwID, attr)
	})

	return err
}

func (s *GatewayStore) loadAttributes(q db.QueryContext, gtwID uuid.UUID, gateway store.Gateway) error {
	attr, ok := gateway.(store.Attributer)
	if !ok {
		return nil
	}

	return s.extraAttributesStore.loadAttributes(q, gtwID, attr)
}

// StoreAttributes store the extra attributes of app if it is a store.Attributer
// and writes the resulting gateway in result.
func (s *GatewayStore) StoreAttributes(ids ttnpb.GatewayIdentifiers, gateway store.Gateway) (err error) {
	_, ok := gateway.(store.Attributer)
	if !ok {
		return nil
	}

	err = s.transact(func(tx *db.Tx) error {
		gtwID, err := s.getGatewayID(tx, ids)
		if err != nil {
			return err
		}

		return s.storeAttributes(tx, gtwID, gateway)
	})

	return
}

func (s *GatewayStore) storeAttributes(q db.QueryContext, gtwID uuid.UUID, gateway store.Gateway) error {
	attr, ok := gateway.(store.Attributer)
	if !ok {
		return nil
	}

	return s.extraAttributesStore.storeAttributes(q, gtwID, attr)
}
