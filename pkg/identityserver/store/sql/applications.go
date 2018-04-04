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

// application represents the schema of the stored application.
type application struct {
	ID uuid.UUID
	ttnpb.Application
}

// ApplicationStore implements store.ApplicationStore.
type ApplicationStore struct {
	storer
	*accountStore
	*extraAttributesStore
	*apiKeysStore
}

// NewApplicationStore returns an ApplicationStore.
func NewApplicationStore(store storer) *ApplicationStore {
	return &ApplicationStore{
		storer:               store,
		accountStore:         newAccountStore(store),
		extraAttributesStore: newExtraAttributesStore(store, "application"),
		apiKeysStore:         newAPIKeysStore(store, "application"),
	}
}

func (s *ApplicationStore) getApplicationIdentifiersFromID(q db.QueryContext, id uuid.UUID) (res ttnpb.ApplicationIdentifiers, err error) {
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
func (s *ApplicationStore) getApplicationID(q db.QueryContext, ids ttnpb.ApplicationIdentifiers) (id uuid.UUID, err error) {
	err = q.SelectOne(
		&id,
		`SELECT
				id
			FROM applications
			WHERE application_id = $1`,
		ids.ApplicationID)
	if db.IsNoRows(err) {
		err = ErrApplicationNotFound.New(nil)
	}
	return
}

// Create creates a new application.
func (s *ApplicationStore) Create(application store.Application) error {
	err := s.transact(func(tx *db.Tx) error {
		appID, err := s.create(tx, application.GetApplication())
		if err != nil {
			return err
		}

		return s.storeAttributes(tx, appID, application)
	})
	return err
}

func (s *ApplicationStore) create(q db.QueryContext, application *ttnpb.Application) (id uuid.UUID, err error) {
	err = q.NamedSelectOne(
		&id,
		`INSERT
			INTO applications (
				application_id,
				description)
			VALUES (
				:application_id,
				:description)
			RETURNING id`,
		application)

	if _, yes := db.IsDuplicate(err); yes {
		err = ErrApplicationIDTaken.New(nil)
	}

	return
}

// GetByID finds the application that matches the identifier and retrieves it.
func (s *ApplicationStore) GetByID(id ttnpb.ApplicationIdentifiers, specializer store.ApplicationSpecializer) (result store.Application, err error) {
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

func (s *ApplicationStore) getByID(q db.QueryContext, id ttnpb.ApplicationIdentifiers) (result application, err error) {
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
		err = ErrApplicationNotFound.New(nil)
	}
	return
}

// Update updates the Application.
func (s *ApplicationStore) Update(application store.Application) error {
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

func (s *ApplicationStore) update(q db.QueryContext, appID uuid.UUID, app *ttnpb.Application) (err error) {
	var id string
	err = q.NamedSelectOne(
		&id,
		`UPDATE applications
			SET
				description = :description,
				updated_at = current_timestamp()
			WHERE id = :id
			RETURNING application_id`,
		application{
			ID:          appID,
			Application: *app,
		})
	if db.IsNoRows(err) {
		err = ErrApplicationNotFound.New(nil)
	}
	return
}

// Delete deletes the application that matches the identifier.
func (s *ApplicationStore) Delete(id ttnpb.ApplicationIdentifiers) (err error) {
	err = s.transact(func(tx *db.Tx) error {
		appID, err := s.getApplicationID(tx, id)
		if err != nil {
			return err
		}

		err = s.deleteCollaborators(tx, appID)
		if err != nil {
			return err
		}

		err = s.deleteAPIKeys(tx, appID)
		if err != nil {
			return err
		}

		return s.delete(tx, appID)
	})
	return
}

// delete deletes the application itself. All rows in other tables that references
// this entity must be deleted before this one gets deleted.
func (s *ApplicationStore) delete(q db.QueryContext, appID uuid.UUID) (err error) {
	var res string
	err = q.SelectOne(
		&res,
		`DELETE
			FROM applications
			WHERE id = $1
			RETURNING application_id`,
		appID)
	if db.IsNoRows(err) {
		err = ErrApplicationNotFound.New(nil)
	}
	return
}
