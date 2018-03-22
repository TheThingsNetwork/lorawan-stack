// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/satori/go.uuid"
)

// LoadAttributes loads the extra attributes in app if it is a store.Attributer.
func (s *ApplicationStore) LoadAttributes(id ttnpb.ApplicationIdentifiers, app store.Application) error {
	attr, ok := app.(store.Attributer)
	if !ok {
		return nil
	}

	err := s.transact(func(tx *db.Tx) error {
		appID, err := s.getApplicationID(tx, id)
		if err != nil {
			return err
		}

		return s.extraAttributesStore.loadAttributes(tx, appID, attr)
	})

	return err
}

func (s *ApplicationStore) loadAttributes(q db.QueryContext, appID uuid.UUID, application store.Application) error {
	attr, ok := application.(store.Attributer)
	if !ok {
		return nil
	}

	return s.extraAttributesStore.loadAttributes(q, appID, attr)
}

// StoreAttributes store the extra attributes of app if it is a store.Attributer
// and writes the resulting application in result.
func (s *ApplicationStore) StoreAttributes(id ttnpb.ApplicationIdentifiers, application store.Application) (err error) {
	_, ok := application.(store.Attributer)
	if !ok {
		return nil
	}

	err = s.transact(func(tx *db.Tx) error {
		appID, err := s.getApplicationID(tx, id)
		if err != nil {
			return err
		}

		return s.storeAttributes(tx, appID, application)
	})

	return
}

func (s *ApplicationStore) storeAttributes(q db.QueryContext, appID uuid.UUID, application store.Application) error {
	attr, ok := application.(store.Attributer)
	if !ok {
		return nil
	}

	return s.extraAttributesStore.storeAttributes(q, appID, attr)
}
