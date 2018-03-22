// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/satori/go.uuid"
)

// LoadAttributes loads the extra attributes in app if it is a store.Attributer.
func (s *UserStore) LoadAttributes(ids ttnpb.UserIdentifiers, app store.User) error {
	attr, ok := app.(store.Attributer)
	if !ok {
		return nil
	}

	err := s.transact(func(tx *db.Tx) error {
		userID, err := s.getUserID(tx, ids)
		if err != nil {
			return err
		}

		return s.extraAttributesStore.loadAttributes(tx, userID, attr)
	})

	return err
}

func (s *UserStore) loadAttributes(q db.QueryContext, userID uuid.UUID, user store.User) error {
	attr, ok := user.(store.Attributer)
	if !ok {
		return nil
	}

	return s.extraAttributesStore.loadAttributes(q, userID, attr)
}

// StoreAttributes store the extra attributes of app if it is a store.Attributer
// and writes the resulting user in result.
func (s *UserStore) StoreAttributes(ids ttnpb.UserIdentifiers, user store.User) (err error) {
	_, ok := user.(store.Attributer)
	if !ok {
		return nil
	}

	err = s.transact(func(tx *db.Tx) error {
		userID, err := s.getUserID(tx, ids)
		if err != nil {
			return err
		}

		return s.storeAttributes(tx, userID, user)
	})

	return
}

func (s *UserStore) storeAttributes(q db.QueryContext, userID uuid.UUID, user store.User) error {
	attr, ok := user.(store.Attributer)
	if !ok {
		return nil
	}

	return s.extraAttributesStore.storeAttributes(q, userID, attr)
}
