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
