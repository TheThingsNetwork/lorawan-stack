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

// application represents the schema of the stored application.
type application struct {
	ID uuid.UUID
	ttnpb.Application
}

type applicationStore struct {
	storer
	*accountStore
	*extraAttributesStore
	*apiKeysStore
}

func newApplicationStore(store storer) *applicationStore {
	return &applicationStore{
		storer:               store,
		accountStore:         newAccountStore(store),
		extraAttributesStore: newExtraAttributesStore(store, "application"),
		apiKeysStore:         newAPIKeysStore(store, "application"),
	}
}

func (s *applicationStore) getApplicationIdentifiersFromID(q db.QueryContext, id uuid.UUID) (res ttnpb.ApplicationIdentifiers, err error) {
	err = q.SelectOne(
		&res,
		`SELECT
				application_id
			FROM applications
			WHERE id = $1`,
		id)
	return
}

// getApplicationID returns the UUID of the application that matches the identifier.
func (s *applicationStore) getApplicationID(q db.QueryContext, ids ttnpb.ApplicationIdentifiers) (id uuid.UUID, err error) {
	err = q.SelectOne(
		&id,
		`SELECT
				id
			FROM applications
			WHERE application_id = $1`,
		ids.ApplicationID)
	if db.IsNoRows(err) {
		err = store.ErrApplicationNotFound.New(nil)
	}
	return
}

// Create creates a new application.
func (s *applicationStore) Create(application store.Application) error {
	err := s.transact(func(tx *db.Tx) error {
		appID, err := s.create(tx, application.GetApplication())
		if err != nil {
			return err
		}

		return s.storeAttributes(tx, appID, application)
	})
	return err
}

func (s *applicationStore) create(q db.QueryContext, application *ttnpb.Application) (id uuid.UUID, err error) {
	err = q.NamedSelectOne(
		&id,
		`INSERT
			INTO applications (
				application_id,
				description,
				created_at,
				updated_at)
			VALUES (
				:application_id,
				:description,
				:created_at,
				:updated_at)
			RETURNING id`,
		application)

	if _, yes := db.IsDuplicate(err); yes {
		err = store.ErrApplicationIDTaken.New(nil)
	}

	return
}

// GetByID finds the application that matches the identifier and retrieves it.
func (s *applicationStore) GetByID(id ttnpb.ApplicationIdentifiers, specializer store.ApplicationSpecializer) (result store.Application, err error) {
	err = s.transact(func(tx *db.Tx) error {
		application, err := s.getByID(tx, id)
		if err != nil {
			return err
		}

		result = specializer(application.Application)

		return s.loadAttributes(tx, application.ID, result)
	}, db.ReadOnly(true))

	return
}

func (s *applicationStore) getByID(q db.QueryContext, id ttnpb.ApplicationIdentifiers) (result application, err error) {
	err = q.NamedSelectOne(
		&result,
		`SELECT
				id,
				application_id,
				description
			FROM applications
			WHERE application_id = :application_id`,
		id)
	if db.IsNoRows(err) {
		err = store.ErrApplicationNotFound.New(nil)
	}
	return
}

// Update updates the Application.
func (s *applicationStore) Update(application store.Application) error {
	app := application.GetApplication()

	err := s.transact(func(tx *db.Tx) error {
		appID, err := s.getApplicationID(tx, app.ApplicationIdentifiers)
		if err != nil {
			return err
		}

		err = s.update(tx, appID, app)
		if err != nil {
			return err
		}

		return s.storeAttributes(tx, appID, application)
	})
	return err
}

func (s *applicationStore) update(q db.QueryContext, appID uuid.UUID, app *ttnpb.Application) (err error) {
	var id string
	err = q.NamedSelectOne(
		&id,
		`UPDATE applications
			SET
				description = :description,
				updated_at = :updated_at
			WHERE id = :id
			RETURNING application_id`,
		application{
			ID:          appID,
			Application: *app,
		})
	if db.IsNoRows(err) {
		err = store.ErrApplicationNotFound.New(nil)
	}
	return
}

// Delete deletes the application that matches the identifier.
func (s *applicationStore) Delete(id ttnpb.ApplicationIdentifiers) (err error) {
	err = s.transact(func(tx *db.Tx) error {
		appID, err := s.getApplicationID(tx, id)
		if err != nil {
			return err
		}

		return s.delete(tx, appID)
	})
	return
}

// delete deletes the application itself. All rows in other tables that references
// this entity must be deleted before this one gets deleted.
func (s *applicationStore) delete(q db.QueryContext, appID uuid.UUID) (err error) {
	var res string
	err = q.SelectOne(
		&res,
		`DELETE
			FROM applications
			WHERE id = $1
			RETURNING application_id`,
		appID)
	if db.IsNoRows(err) {
		err = store.ErrApplicationNotFound.New(nil)
	}
	return
}
