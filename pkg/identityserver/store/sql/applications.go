// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"fmt"
	"strings"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/factory"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/helpers"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// ApplicationStore implements store.ApplicationStore.
type ApplicationStore struct {
	*Store
	*collaboratorStore
	factory factory.ApplicationFactory
}

func init() {
	ErrApplicationNotFound.Register()
	ErrApplicationIDTaken.Register()
	ErrApplicationAPIKeyNotFound.Register()
}

// ErrApplicationNotFound is returned when trying to fetch an application that
// does not exist.
var ErrApplicationNotFound = &errors.ErrDescriptor{
	MessageFormat: "Application `{application_id}` does not exist",
	Code:          1,
	Type:          errors.NotFound,
}

// ErrApplicationIDTaken is returned when trying to create a new application
// with an ID that already exists.
var ErrApplicationIDTaken = &errors.ErrDescriptor{
	MessageFormat: "Application id `{application_id}` is already taken",
	Code:          2,
	Type:          errors.AlreadyExists,
}

// ErrApplicationAPIKeyNotFound is returned when trying to access or delete
// an application API key that does not exist.
var ErrApplicationAPIKeyNotFound = &errors.ErrDescriptor{
	MessageFormat: "API key `{key_name}` does not exist for application `{application_id}`",
	Code:          3,
	Type:          errors.NotFound,
}

func NewApplicationStore(store *Store, factory factory.ApplicationFactory) *ApplicationStore {
	return &ApplicationStore{
		Store:             store,
		factory:           factory,
		collaboratorStore: newCollaboratorStore(store, "application"),
	}
}

// Create creates a new application.
func (s *ApplicationStore) Create(application types.Application) error {
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.create(tx, application)
	})
	return err
}

func (s *ApplicationStore) create(q db.QueryContext, application types.Application) error {
	app := application.GetApplication()
	_, err := q.NamedExec(
		`INSERT
			INTO applications (application_id, description, updated_at, archived_at)
			VALUES (:application_id, :description, :updated_at, :archived_at)`,
		app)

	if _, yes := db.IsDuplicate(err); yes {
		return ErrApplicationIDTaken.New(errors.Attributes{
			"application_id": app.ApplicationID,
		})
	}

	if err != nil {
		return err
	}

	return s.writeAttributes(q, application, nil)
}

// GetByID finds the application by ID and retrieves it.
func (s *ApplicationStore) GetByID(appID string) (types.Application, error) {
	result := s.factory.BuildApplication()
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.getByID(tx, appID, result)
	}, db.ReadOnly(true))
	return result, err
}

func (s *ApplicationStore) getByID(q db.QueryContext, appID string, result types.Application) error {
	err := s.application(q, appID, result)
	if err != nil {
		return err
	}

	apiKeys, err := s.listAPIKeys(q, appID)
	if err != nil {
		return err
	}
	result.SetAPIKeys(apiKeys)

	return s.loadAttributes(q, appID, result)
}

func (s *ApplicationStore) application(q db.QueryContext, appID string, result types.Application) error {
	err := q.SelectOne(result, "SELECT * FROM applications WHERE application_id = $1", appID)
	if db.IsNoRows(err) {
		return ErrApplicationNotFound.New(errors.Attributes{
			"application_id": appID,
		})
	}

	if err != nil {
		return err
	}

	return s.loadAttributes(q, appID, result)
}

// FindByUser returns the Applications to which an User is a collaborator.
func (s *ApplicationStore) ListByUser(userID string) ([]types.Application, error) {
	var result []types.Application
	err := s.db.Transact(func(tx *db.Tx) error {
		appIDs, err := s.userApplications(tx, userID)
		if err != nil {
			return err
		}

		for _, appID := range appIDs {
			app := s.factory.BuildApplication()

			err := s.getByID(tx, appID, app)
			if err != nil {
				return err
			}

			result = append(result, app)
		}

		return nil
	})
	return result, err
}

func (s *ApplicationStore) userApplications(q db.QueryContext, userID string) ([]string, error) {
	var appIDs []string
	err := q.Select(
		&appIDs,
		`SELECT DISTINCT application_id
			FROM applications_collaborators
			WHERE user_id = $1`,
		userID)

	if !db.IsNoRows(err) && err != nil {
		return nil, err
	}

	return appIDs, nil
}

// Edit updates the Application and returns the updated Application.
func (s *ApplicationStore) Update(application types.Application) error {
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.update(tx, application)
	})
	return err
}

func (s *ApplicationStore) update(q db.QueryContext, application types.Application) error {
	app := application.GetApplication()
	app.UpdatedAt = time.Now()

	var id string
	err := q.NamedSelectOne(
		&id,
		`UPDATE applications
			SET description = :description, updated_at = :updated_at
			WHERE application_id = :application_id
			RETURNING application_id`,
		app)

	if db.IsNoRows(err) {
		return ErrApplicationNotFound.New(errors.Attributes{
			"application_id": app.ApplicationID,
		})
	}

	if err != nil {
		return err
	}

	return s.writeAttributes(q, application, nil)
}

// syncUpdatedAt modifies the application UpdatedAt field to the current timestamp.
func (s *ApplicationStore) syncUpdatedAt(q db.QueryContext, appID string) error {
	var id string
	err := q.SelectOne(
		&id,
		`UPDATE applications
			SET updated_at = $1
			WHERE application_id = $2
			RETURNING application_id`,
		time.Now(),
		appID)
	if db.IsNoRows(err) {
		return ErrApplicationNotFound.New(errors.Attributes{
			"application_id": appID,
		})
	}
	return err
}

// Archive disables the Application.
func (s *ApplicationStore) Archive(appID string) error {
	return s.archive(s.db, appID)
}

func (s *ApplicationStore) archive(q db.QueryContext, appID string) error {
	var id string
	err := q.SelectOne(
		&id,
		`UPDATE applications
			SET archived_at = $1
			WHERE application_id = $2
			RETURNING application_id`,
		time.Now(),
		appID)
	if db.IsNoRows(err) {
		return ErrApplicationNotFound.New(errors.Attributes{
			"application_id": appID,
		})
	}
	return err
}

// AddAPIKey adds a new Application API key to a given Application.
func (s *ApplicationStore) AddAPIKey(appID string, key ttnpb.APIKey) error {
	err := s.db.Transact(func(tx *db.Tx) error {
		err := s.addAPIKey(tx, appID, key)
		if err != nil {
			return err
		}

		return s.syncUpdatedAt(tx, appID)
	})
	return err
}

func (s *ApplicationStore) addAPIKey(q db.QueryContext, appID string, key ttnpb.APIKey) error {
	query, args := s.addAPIKeyQuery(appID, key)
	_, err := q.Exec(query, args...)
	return err
}

func (s *ApplicationStore) addAPIKeyQuery(appID string, key ttnpb.APIKey) (string, []interface{}) {
	args := make([]interface{}, 3+len(key.Rights))
	args[0] = appID
	args[1] = key.Name
	args[2] = key.Key

	boundValues := make([]string, len(key.Rights))

	for i, right := range key.Rights {
		args[i+3] = right
		boundValues[i] = fmt.Sprintf("($1, $2, $3, $%d)", i+4)
	}

	query := fmt.Sprintf(
		`INSERT
			INTO applications_api_keys (application_id, name, key, "right")
			VALUES %s
			ON CONFLICT (application_id, name, "right")
			DO NOTHING`,
		strings.Join(boundValues, ", "))

	return query, args
}

func (s *ApplicationStore) listAPIKeys(q db.QueryContext, appID string) ([]ttnpb.APIKey, error) {
	var keys []struct {
		ttnpb.APIKey
		Right ttnpb.Right
	}
	err := q.Select(
		&keys,
		`SELECT name, key, "right"
			FROM applications_api_keys
			WHERE application_id = $1`,
		appID)

	if len(keys) == 0 {
		return nil, nil
	}

	if !db.IsNoRows(err) && err != nil {
		return nil, err
	}

	byName := make(map[string]*ttnpb.APIKey)
	for _, key := range keys {
		if k, ok := byName[key.Name]; ok {
			k.Rights = append(k.Rights, key.Right)
			continue
		}

		byName[key.Name] = &ttnpb.APIKey{
			Name:   key.Name,
			Key:    key.Key,
			Rights: []ttnpb.Right{key.Right},
		}
	}

	apiKeys := make([]ttnpb.APIKey, 0, len(byName))
	for _, key := range byName {
		apiKeys = append(apiKeys, *key)
	}

	return apiKeys, nil
}

// RemoveAPIKey removes an Application API key from a given Application.
func (s *ApplicationStore) RemoveAPIKey(appID string, keyName string) error {
	err := s.db.Transact(func(tx *db.Tx) error {
		err := s.removeAPIKey(tx, appID, keyName)
		if err != nil {
			return err
		}

		return s.syncUpdatedAt(tx, appID)
	})
	return err
}

func (s *ApplicationStore) removeAPIKey(q db.QueryContext, appID string, keyName string) error {
	var i string
	err := q.SelectOne(
		&i,
		`DELETE
			FROM applications_api_keys
			WHERE application_id = $1 AND name = $2
			RETURNING application_id`,
		appID,
		keyName)
	if db.IsNoRows(err) {
		return ErrApplicationAPIKeyNotFound.New(errors.Attributes{
			"application_id": appID,
			"key_name":       keyName,
		})
	}
	return err
}

// LoadAttributes loads extra attributes into the Application.
func (s *ApplicationStore) LoadAttributes(application types.Application) error {
	return s.db.Transact(func(tx *db.Tx) error {
		return s.loadAttributes(tx, application.GetApplication().ApplicationID, application)
	})
}

func (s *ApplicationStore) loadAttributes(q db.QueryContext, appID string, application types.Application) error {
	attr, ok := application.(store.Attributer)
	if !ok {
		return nil
	}

	// fill the application from all specified namespaces
	for _, namespace := range attr.Namespaces() {
		m := make(map[string]interface{})
		err := q.SelectOne(
			&m,
			fmt.Sprintf("SELECT * FROM %s_applications WHERE application_id = $1", namespace),
			appID)
		if err != nil {
			return err
		}

		err = attr.Fill(namespace, m)
		if err != nil {
			return err
		}
	}

	return nil
}

// WriteAttributes writes the extra attributes on the Application if it is an
// Attributer to the store.
func (s *ApplicationStore) WriteAttributes(application types.Application, result types.Application) error {
	return s.db.Transact(func(tx *db.Tx) error {
		return s.writeAttributes(tx, application, result)
	})
}

func (s *ApplicationStore) writeAttributes(q db.QueryContext, application, result types.Application) error {
	attr, ok := application.(store.Attributer)
	if !ok {
		return nil
	}

	for _, namespace := range attr.Namespaces() {
		query, values := helpers.WriteAttributes(attr, namespace, "applications", "application_id", application.GetApplication().ApplicationID)

		r := make(map[string]interface{})
		err := q.SelectOne(r, query, values...)
		if err != nil {
			return err
		}

		if rattr, ok := result.(store.Attributer); ok {
			err = rattr.Fill(namespace, r)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// SetFactory allows to replace the default ttnpb.Application factory.
func (s *ApplicationStore) SetFactory(factory factory.ApplicationFactory) {
	s.factory = factory
}
