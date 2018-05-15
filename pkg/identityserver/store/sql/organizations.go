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

type organization struct {
	ID uuid.UUID
	ttnpb.Organization
}

// OrganizationStore implements store.OrganizationStore.
type OrganizationStore struct {
	storer
	*extraAttributesStore
	*apiKeysStore
	*accountStore
	*UserStore
}

// NewOrganizationStore returns an organization store.
func NewOrganizationStore(store storer) *OrganizationStore {
	return &OrganizationStore{
		storer:               store,
		extraAttributesStore: newExtraAttributesStore(store, "organization"),
		apiKeysStore:         newAPIKeysStore(store, "organization"),
		accountStore:         newAccountStore(store),
		UserStore:            store.store().Users.(*UserStore),
	}
}

func (s *OrganizationStore) getOrganizationIdentifiersFromID(q db.QueryContext, id uuid.UUID) (res ttnpb.OrganizationIdentifiers, err error) {
	err = q.SelectOne(
		&res,
		`SELECT
				organization_id
			FROM organizations
			WHERE id = $1`,
		id)
	return
}

func (s *OrganizationStore) getOrganizationID(q db.QueryContext, ids ttnpb.OrganizationIdentifiers) (id uuid.UUID, err error) {
	err = q.SelectOne(
		&id,
		`SELECT
				id
			FROM organizations
			WHERE organization_id = $1`,
		ids.OrganizationID)
	if db.IsNoRows(err) {
		err = store.ErrOrganizationNotFound.New(nil)
	}
	return
}

// Create creates an organization.
func (s *OrganizationStore) Create(org store.Organization) error {
	err := s.transact(func(tx *db.Tx) error {
		o := org.GetOrganization()

		id, err := s.accountStore.registerOrganizationID(tx, o.OrganizationID)
		if err != nil {
			return err
		}

		err = s.create(tx, organization{
			ID:           id,
			Organization: *o,
		})
		if err != nil {
			return err
		}

		return s.storeAttributes(tx, id, org)
	})
	return err
}

func (s *OrganizationStore) create(q db.QueryContext, data organization) (err error) {
	_, err = q.NamedExec(
		`INSERT
			INTO organizations (
				id,
				organization_id,
				name,
				description,
				url,
				location,
				email,
				created_at,
				updated_at
			)
			VALUES (
				:id,
				lower(:organization_id),
				:name,
				:description,
				:url,
				:location,
				lower(:email),
				:created_at,
				:updated_at)
			RETURNING id`,
		data)
	return
}

// GetByID finds the organization by ID and retrieves it.
func (s *OrganizationStore) GetByID(ids ttnpb.OrganizationIdentifiers, specializer store.OrganizationSpecializer) (result store.Organization, err error) {
	err = s.transact(func(tx *db.Tx) error {
		orgID, err := s.getOrganizationID(tx, ids)
		if err != nil {
			return err
		}

		organization, err := s.getByID(tx, orgID)
		if err != nil {
			return err
		}

		result = specializer(organization)

		return s.loadAttributes(tx, orgID, result)
	})
	return
}

func (s *OrganizationStore) getByID(q db.QueryContext, orgID uuid.UUID) (result ttnpb.Organization, err error) {
	err = q.SelectOne(
		&result,
		`SELECT
				organization_id,
				name,
				description,
				url,
				location,
				email
			FROM organizations
			WHERE id = $1`,
		orgID)
	if db.IsNoRows(err) {
		err = store.ErrOrganizationNotFound.New(nil)
	}
	return
}

// Update updates an organization.
func (s *OrganizationStore) Update(organization store.Organization) error {
	err := s.transact(func(tx *db.Tx) error {
		org := organization.GetOrganization()

		orgID, err := s.getOrganizationID(tx, org.OrganizationIdentifiers)
		if err != nil {
			return err
		}

		err = s.update(tx, orgID, org)
		if err != nil {
			return err
		}

		return s.storeAttributes(tx, orgID, organization)
	})

	return err
}

func (s *OrganizationStore) update(q db.QueryContext, orgID uuid.UUID, data *ttnpb.Organization) error {
	var id string
	err := q.NamedSelectOne(
		&id,
		`UPDATE organizations
			SET
				name = :name,
				description = :description,
				url = :url,
				location = :location,
				email = lower(:email),
				updated_at = :updated_at
			WHERE id = :id
			RETURNING organization_id`,
		organization{
			ID:           orgID,
			Organization: *data,
		})

	if db.IsNoRows(err) {
		return store.ErrOrganizationNotFound.New(nil)
	}

	return err
}

// Delete deletes an organization.
func (s *OrganizationStore) Delete(ids ttnpb.OrganizationIdentifiers) error {
	err := s.transact(func(tx *db.Tx) error {
		orgID, err := s.getOrganizationID(tx, ids)
		if err != nil {
			return err
		}

		err = s.delete(tx, orgID)
		if err != nil {
			return err
		}

		return s.accountStore.deleteOrganizationID(tx, orgID)
	})

	return err
}

func (s *OrganizationStore) delete(q db.QueryContext, orgID uuid.UUID) (err error) {
	var id string
	err = q.SelectOne(
		&id,
		`DELETE
			FROM organizations
			WHERE id = $1
			RETURNING organization_id`,
		orgID)
	if db.IsNoRows(err) {
		err = store.ErrOrganizationNotFound.New(nil)
	}
	return
}
