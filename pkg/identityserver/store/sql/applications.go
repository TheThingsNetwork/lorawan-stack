// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"errors"
	"fmt"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/factory"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/helpers"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
)

// ApplicationStore implements store.ApplicationStore.
type ApplicationStore struct {
	*Store
	factory factory.ApplicationFactory
}

// ErrApplicationNotFound is returned when trying to fetch an application that
// does not exist.
var ErrApplicationNotFound = errors.New("application not found")

// ErrApplicationIDTaken is returned when trying to create a new application
// with an ID that already exists.
var ErrApplicationIDTaken = errors.New("application ID already taken")

// ErrAppEUINotFound is returned when trying to remove an AppEUI that does not exist.
var ErrAppEUINotFound = errors.New("application EUI not found")

// ErrApplicationAPIKeyNotFound is returned when trying to access or delete
// an application API key that does not exist.
var ErrApplicationAPIKeyNotFound = errors.New("application API key not found")

// ErrApplicationCollaboratorNotFound is returned when trying to remove a
// collaborator that does not exist.
var ErrApplicationCollaboratorNotFound = errors.New("application collaborator not found")

// ErrApplicationCollaboratorRightNotFound is returned when trying to revoke a
// right from a collaborator that is not granted.
var ErrApplicationCollaboratorRightNotFound = errors.New("application collaborator right not found")

// SetFactory replaces the factory.
func (s *ApplicationStore) SetFactory(factory factory.ApplicationFactory) {
	s.factory = factory
}

// LoadAttributes loads the applications attributes into result if it is an Attributer.
func (s *ApplicationStore) LoadAttributes(application types.Application) error {
	return s.db.Transact(func(tx *db.Tx) error {
		return s.loadAttributes(tx, application.GetApplication().ID, application)
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

// WriteAttributes writes the applications attributes into result if it is an Attributer.
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
		query, values := helpers.WriteAttributes(attr, namespace, "applications", "application_id", application.GetApplication().ID)

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

// FindByID finds the application by ID.
func (s *ApplicationStore) FindByID(appID string) (types.Application, error) {
	result := s.factory.Application()
	err := s.db.Transact(func(tx *db.Tx) error {
		// fetch app
		err := s.application(tx, appID, result)
		if err != nil {
			return err
		}

		// fetch euis
		euis, err := s.appEUIs(tx, appID)
		if err != nil {
			return err
		}
		result.SetEUIs(euis)

		// fetch api keys
		apiKeys, err := s.apiKeys(tx, appID)
		if err != nil {
			return err
		}
		result.SetAPIKeys(apiKeys)

		return s.loadAttributes(tx, appID, result)
	}, db.ReadOnly(true))

	return result, err
}

func (s *ApplicationStore) application(q db.QueryContext, appID string, result types.Application) error {
	err := q.SelectOne(result, "SELECT * FROM applications WHERE id = $1", appID)
	if db.IsNoRows(err) {
		return ErrApplicationNotFound
	}

	return err
}

// AppEUIs fetches all application EUIs that belong to the app with the
// specified appID.
func (s *ApplicationStore) AppEUIs(appID string) ([]types.AppEUI, error) {
	return s.appEUIs(s.db, appID)
}

func (s *ApplicationStore) appEUIs(q db.QueryContext, appID string) ([]types.AppEUI, error) {
	var appEUIs []types.AppEUI
	err := q.Select(
		&appEUIs,
		`SELECT eui
			FROM applications_euis
			WHERE app_id = $1`,
		appID)
	if err != nil && !db.IsNoRows(err) {
		return nil, err
	}

	return appEUIs, nil
}

// APIKeys gets all api keys that belong to the appID.
func (s *ApplicationStore) APIKeys(appID string) ([]types.ApplicationAPIKey, error) {
	return s.apiKeys(s.db, appID)
}

func (s *ApplicationStore) apiKeys(q db.QueryContext, appID string) ([]types.ApplicationAPIKey, error) {
	var keys []struct {
		types.ApplicationAPIKey
		Right string `db:"right"`
	}
	err := q.Select(
		&keys,
		`SELECT name, key, "right"
			FROM applications_api_keys
			WHERE app_id = $1`,
		appID)
	if !db.IsNoRows(err) && err != nil {
		return nil, err
	}

	byName := make(map[string]*types.ApplicationAPIKey)
	for _, key := range keys {
		if k, ok := byName[key.Name]; ok {
			k.Rights = append(k.Rights, types.Right(key.Right))
			continue
		}

		byName[key.Name] = &types.ApplicationAPIKey{
			Name: key.Name,
			Key:  key.Key,
			Rights: []types.Right{
				types.Right(key.Right),
			},
		}
	}

	apiKeys := make([]types.ApplicationAPIKey, 0, len(byName))
	for _, key := range byName {
		apiKeys = append(apiKeys, *key)
	}

	return apiKeys, nil
}

// FindByUser returns the applications to which an user is a collaborator.
func (s *ApplicationStore) FindByUser(username string) ([]types.Application, error) {
	var applications []types.Application
	err := s.db.Transact(func(tx *db.Tx) error {
		// get applications ids
		appIDs, err := s.userApplications(tx, username)
		if err != nil {
			return err
		}

		// fetch applications
		for _, appID := range appIDs {
			app := s.factory.Application()
			err := s.application(tx, appID, app)
			if err != nil {
				return err
			}

			// fetch euis
			euis, err := s.appEUIs(tx, appID)
			if err != nil {
				return err
			}
			app.SetEUIs(euis)

			// fetch api keys
			apiKeys, err := s.apiKeys(tx, appID)
			if err != nil {
				return err
			}
			app.SetAPIKeys(apiKeys)

			applications = append(applications, app)
		}
		return nil
	})
	return applications, err
}

// userApplications fetches all applications that a given user is collaborator.
func (s *ApplicationStore) userApplications(q db.QueryContext, username string) ([]string, error) {
	var appIDs []string
	err := q.Select(
		&appIDs,
		`SELECT DISTINCT app_id
			FROM applications_collaborators
			WHERE username = $1`,
		username)
	if !db.IsNoRows(err) && err != nil {
		return nil, err
	}

	return appIDs, nil
}

// Create creates a new application and returns the resulting application.
func (s *ApplicationStore) Create(application types.Application) (types.Application, error) {
	result := s.factory.Application()
	err := s.db.Transact(func(tx *db.Tx) error {
		err := s.create(tx, application, result)
		if err != nil {
			return err
		}

		app := application.GetApplication()

		// add euis
		for _, eui := range app.EUIs {
			err := s.addAppEUI(tx, app.ID, eui)
			if err != nil {
				return err
			}
		}
		result.SetEUIs(app.EUIs)

		// add api keys
		for _, apiKey := range app.APIKeys {
			err := s.addApplicationAPIKey(tx, app.ID, apiKey)
			if err != nil {
				return err
			}
		}
		result.SetAPIKeys(app.APIKeys)

		return nil
	})
	return result, err
}

func (s *ApplicationStore) create(q db.QueryContext, application, result types.Application) error {
	app := application.GetApplication()
	err := q.NamedSelectOne(
		result,
		`INSERT
			INTO applications (id, description)
			VALUES (:id, :description)
			RETURNING *`,
		app)

	if _, yes := db.IsDuplicate(err); yes {
		return ErrApplicationIDTaken
	}

	if err != nil {
		return err
	}

	return s.writeAttributes(q, application, result)
}

// AddAppEUI adds an eui to the app.
func (s *ApplicationStore) AddAppEUI(appID string, eui types.AppEUI) error {
	return s.addAppEUI(s.db, appID, eui)
}

func (s *ApplicationStore) addAppEUI(q db.QueryContext, appID string, eui types.AppEUI) error {
	_, err := q.Exec(
		`INSERT
			INTO applications_euis (app_id, eui)
			VALUES ($1, $2)
			ON CONFLICT (app_id, eui)
			DO NOTHING`,
		appID,
		eui)
	return err
}

// DeleteAppEUI delete an eui to the app.
func (s *ApplicationStore) DeleteAppEUI(appID string, eui types.AppEUI) error {
	return s.deleteAppEUI(s.db, appID, eui)
}

func (s *ApplicationStore) deleteAppEUI(q db.QueryContext, appID string, eui types.AppEUI) error {
	var e string
	err := q.SelectOne(
		&e,
		`DELETE
			FROM applications_euis
			WHERE app_id = $1 AND eui = $2
			RETURNING eui`,
		appID,
		eui)
	if db.IsNoRows(err) {
		return ErrAppEUINotFound
	}
	return err
}

// AddApplicationAPIKey adds an api key to the app.
func (s *ApplicationStore) AddApplicationAPIKey(appID string, key types.ApplicationAPIKey) error {
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.addApplicationAPIKey(tx, appID, key)
	})
	return err
}

func (s *ApplicationStore) addApplicationAPIKey(q db.QueryContext, appID string, key types.ApplicationAPIKey) error {
	for _, right := range key.Rights {
		_, err := q.Exec(
			`INSERT
				INTO applications_api_keys (app_id, name, key, "right")
				VALUES ($1, $2, $3, $4)
				ON CONFLICT(app_id, name, "right")
				DO NOTHING`,
			appID,
			key.Name,
			key.Key,
			right,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteApplicationAPIKey deletes an api key to the app.
func (s *ApplicationStore) DeleteApplicationAPIKey(appID string, keyName string) error {
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.deleteApplicationAPIKey(tx, appID, keyName)
	})
	return err
}

func (s *ApplicationStore) deleteApplicationAPIKey(q db.QueryContext, appID string, keyName string) error {
	var i string
	err := q.SelectOne(
		&i,
		`DELETE
			FROM applications_api_keys
			WHERE app_id = $1 AND name = $2
			RETURNING app_id`,
		appID,
		keyName)
	if db.IsNoRows(err) {
		return ErrApplicationAPIKeyNotFound
	}
	return err
}

// Update updates an application and returns the resulting application.
func (s *ApplicationStore) Update(application types.Application) (types.Application, error) {
	result := s.factory.Application()
	err := s.db.Transact(func(tx *db.Tx) error {
		err := s.update(tx, application, result)
		if err != nil {
			return err
		}

		app := application.GetApplication()

		// update euis
		for _, eui := range app.EUIs {
			err := s.addAppEUI(tx, app.ID, eui)
			if err != nil {
				return err
			}
		}
		result.SetEUIs(app.EUIs)

		// update api keys
		for _, key := range app.APIKeys {
			err := s.addApplicationAPIKey(tx, app.ID, key)
			if err != nil {
				return err
			}
		}
		result.SetAPIKeys(app.APIKeys)

		return nil
	})
	return result, err
}

func (s *ApplicationStore) update(q db.QueryContext, application, result types.Application) error {
	app := application.GetApplication()
	err := q.NamedSelectOne(
		result,
		`UPDATE applications
			SET description = :description
			WHERE id = :id
			RETURNING *`,
		app)

	if db.IsNoRows(err) {
		return ErrApplicationNotFound
	}

	if err != nil {
		return err
	}

	return s.writeAttributes(q, application, result)
}

// Archive archives an application.
func (s *ApplicationStore) Archive(appID string) error {
	return s.archive(s.db, appID)
}

func (s *ApplicationStore) archive(q db.QueryContext, appID string) error {
	var id string
	err := q.SelectOne(
		&id,
		`UPDATE applications
			SET archived = $1
			WHERE id = $2
			RETURNING id`,
		time.Now(),
		appID)
	if db.IsNoRows(err) {
		return ErrApplicationNotFound
	}
	return err
}

// Collaborators returns the list of collaborators to a given application.
func (s *ApplicationStore) Collaborators(appID string) ([]types.Collaborator, error) {
	return s.collaborators(s.db, appID)
}

func (s *ApplicationStore) collaborators(q db.QueryContext, appID string) ([]types.Collaborator, error) {
	var collaborators []struct {
		types.Collaborator
		Right string `db:"right"`
	}
	err := q.Select(
		&collaborators,
		`SELECT username, "right"
			FROM applications_collaborators
			WHERE app_id = $1`,
		appID)
	if !db.IsNoRows(err) && err != nil {
		return nil, err
	}

	byUser := make(map[string]*types.Collaborator)
	for _, collaborator := range collaborators {
		if _, exists := byUser[collaborator.Username]; !exists {
			byUser[collaborator.Username] = &types.Collaborator{
				Username: collaborator.Username,
				Rights:   []types.Right{types.Right(collaborator.Right)},
			}
			continue
		}

		byUser[collaborator.Username].Rights = append(byUser[collaborator.Username].Rights, types.Right(collaborator.Right))
	}

	result := make([]types.Collaborator, 0, len(byUser))
	for _, collaborator := range byUser {
		result = append(result, *collaborator)
	}

	return result, nil
}

// AddCollaborator adds a new collaborator to an application.
func (s *ApplicationStore) AddCollaborator(appID string, collaborator types.Collaborator) error {
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.addCollaborator(tx, appID, collaborator)
	})
	return err
}

func (s *ApplicationStore) addCollaborator(q db.QueryContext, appID string, collaborator types.Collaborator) error {
	for _, right := range collaborator.Rights {
		err := s.grantRight(q, appID, collaborator.Username, right)
		if err != nil {
			return err
		}
	}
	return nil
}

// GrantRight grants a right to a specific user in a given application.
func (s *ApplicationStore) GrantRight(appID string, username string, right types.Right) error {
	return s.grantRight(s.db, appID, username, right)
}

func (s *ApplicationStore) grantRight(q db.QueryContext, appID string, username string, right types.Right) error {
	_, err := q.Exec(
		`INSERT
			INTO applications_collaborators (app_id, username, "right")
			VALUES ($1, $2, $3)
			ON CONFLICT (app_id, username, "right")
			DO NOTHING`,
		appID,
		username,
		right)
	return err
}

// RevokeRight revokes a specific right to a specific user in a given application.
func (s *ApplicationStore) RevokeRight(appID string, username string, right types.Right) error {
	return s.revokeRight(s.db, appID, username, right)
}

func (s *ApplicationStore) revokeRight(q db.QueryContext, appID string, username string, right types.Right) error {
	var u string
	err := q.SelectOne(
		&u,
		`DELETE
			FROM applications_collaborators
			WHERE app_id = $1 AND username = $2 AND "right" = $3
			RETURNING username`,
		appID,
		username,
		right)
	if db.IsNoRows(err) {
		return ErrApplicationCollaboratorRightNotFound
	}
	return err
}

// RemoveCollaborator removes a collaborator of a given application.
func (s *ApplicationStore) RemoveCollaborator(appID string, username string) error {
	return s.removeCollaborator(s.db, appID, username)
}

func (s *ApplicationStore) removeCollaborator(q db.QueryContext, appID string, username string) error {
	var u string
	err := q.SelectOne(
		&u,
		`DELETE
			FROM applications_collaborators
			WHERE app_id = $1 AND username = $2
			RETURNING username`,
		appID,
		username)
	if db.IsNoRows(err) {
		return ErrApplicationCollaboratorNotFound
	}
	return err
}

// UserRights returns the list of rights that an user has to a given application.
func (s *ApplicationStore) UserRights(appID string, username string) ([]types.Right, error) {
	return s.userRights(s.db, appID, username)
}

func (s *ApplicationStore) userRights(q db.QueryContext, appID string, username string) ([]types.Right, error) {
	var rights []types.Right
	err := q.Select(
		&rights,
		`SELECT "right"
			FROM applications_collaborators
			WHERE app_id = $1 AND username = $2`,
		appID,
		username)
	return rights, err
}
