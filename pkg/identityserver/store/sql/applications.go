// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"fmt"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
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

func init() {
	ErrApplicationNotFound.Register()
	ErrApplicationIDTaken.Register()
	ErrAppEUINotFound.Register()
	ErrApplicationAPIKeyNotFound.Register()
	ErrApplicationCollaboratorNotFound.Register()
	ErrApplicationCollaboratorRightNotFound.Register()
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

// ErrAppEUINotFound is returned when trying to remove an AppEUI that does not exist.
var ErrAppEUINotFound = &errors.ErrDescriptor{
	MessageFormat: "AppEUI `{eui}` does not exist for application `{application_id}`",
	Code:          3,
	Type:          errors.NotFound,
}

// ErrApplicationAPIKeyNotFound is returned when trying to access or delete
// an application API key that does not exist.
var ErrApplicationAPIKeyNotFound = &errors.ErrDescriptor{
	MessageFormat: "API key `{key_name}` does not exist for application `{application_id}`",
	Code:          4,
	Type:          errors.NotFound,
}

// ErrApplicationCollaboratorNotFound is returned when trying to remove a
// collaborator that does not exist.
var ErrApplicationCollaboratorNotFound = &errors.ErrDescriptor{
	MessageFormat: "User `{username}` is not a collaborator for application `{application_id}`",
	Code:          5,
	Type:          errors.NotFound,
}

// ErrApplicationCollaboratorRightNotFound is returned when trying to revoke a
// right from a collaborator that is not granted.
var ErrApplicationCollaboratorRightNotFound = &errors.ErrDescriptor{
	MessageFormat: "User `{username}` does not have the right `{right}` for application `{application_id}`",
	Code:          6,
	Type:          errors.NotFound,
}

// Register creates a new Application and returns the new created Application.
func (s *ApplicationStore) Register(application types.Application) (types.Application, error) {
	result := s.factory.Application()
	err := s.db.Transact(func(tx *db.Tx) error {
		err := s.register(tx, application, result)
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
			err := s.addAPIKey(tx, app.ID, apiKey)
			if err != nil {
				return err
			}
		}
		result.SetAPIKeys(app.APIKeys)

		return nil
	})
	return result, err
}

func (s *ApplicationStore) register(q db.QueryContext, application, result types.Application) error {
	app := application.GetApplication()
	err := q.NamedSelectOne(
		result,
		`INSERT
			INTO applications (id, description)
			VALUES (:id, :description)
			RETURNING *`,
		app)

	if _, yes := db.IsDuplicate(err); yes {
		return ErrApplicationIDTaken.New(errors.Attributes{
			"application_id": app.ID,
		})
	}

	if err != nil {
		return err
	}

	return s.writeAttributes(q, application, result)
}

// FindByID finds the Application by ID and retrieves it.
func (s *ApplicationStore) FindByID(appID string) (types.Application, error) {
	result := s.factory.Application()
	err := s.db.Transact(func(tx *db.Tx) error {
		// fetch app
		err := s.application(tx, appID, result)
		if err != nil {
			return err
		}

		// fetch euis
		euis, err := s.listAppEUIs(tx, appID)
		if err != nil {
			return err
		}
		result.SetEUIs(euis)

		// fetch api keys
		apiKeys, err := s.listAPIKeys(tx, appID)
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
		return ErrApplicationNotFound.New(errors.Attributes{
			"application_id": appID,
		})
	}

	return err
}

// FindByUser returns the Applications to which an User is a collaborator.
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
			euis, err := s.listAppEUIs(tx, appID)
			if err != nil {
				return err
			}
			app.SetEUIs(euis)

			// fetch api keys
			apiKeys, err := s.listAPIKeys(tx, appID)
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

// Edit updates the Application and returns the updated Application.
func (s *ApplicationStore) Edit(application types.Application) (types.Application, error) {
	result := s.factory.Application()
	err := s.db.Transact(func(tx *db.Tx) error {
		err := s.edit(tx, application, result)
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
			err := s.addAPIKey(tx, app.ID, key)
			if err != nil {
				return err
			}
		}
		result.SetAPIKeys(app.APIKeys)

		return nil
	})
	return result, err
}

func (s *ApplicationStore) edit(q db.QueryContext, application, result types.Application) error {
	app := application.GetApplication()
	err := q.NamedSelectOne(
		result,
		`UPDATE applications
			SET description = :description
			WHERE id = :id
			RETURNING *`,
		app)

	if db.IsNoRows(err) {
		return ErrApplicationNotFound.New(errors.Attributes{
			"application_id": app.ID,
		})
	}

	if err != nil {
		return err
	}

	return s.writeAttributes(q, application, result)
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
			SET archived = $1
			WHERE id = $2
			RETURNING id`,
		time.Now(),
		appID)
	if db.IsNoRows(err) {
		return ErrApplicationNotFound.New(errors.Attributes{
			"application_id": appID,
		})
	}
	return err
}

// AddAppEUI adds a new AppEUI to a given Application.
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

// ListAppEUIs returns all the AppEUIs that belong to a given Application.
func (s *ApplicationStore) ListAppEUIs(appID string) ([]types.AppEUI, error) {
	return s.listAppEUIs(s.db, appID)
}

func (s *ApplicationStore) listAppEUIs(q db.QueryContext, appID string) ([]types.AppEUI, error) {
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

// RemoveAppEUI remove an AppEUI from a given Application.
func (s *ApplicationStore) RemoveAppEUI(appID string, eui types.AppEUI) error {
	return s.removeAppEUI(s.db, appID, eui)
}

func (s *ApplicationStore) removeAppEUI(q db.QueryContext, appID string, eui types.AppEUI) error {
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
		return ErrAppEUINotFound.New(errors.Attributes{
			"application_id": appID,
			"eui":            eui,
		})
	}
	return err
}

// AddAPIKey adds a new Application API key to a given Application.
func (s *ApplicationStore) AddAPIKey(appID string, key types.ApplicationAPIKey) error {
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.addAPIKey(tx, appID, key)
	})
	return err
}

func (s *ApplicationStore) addAPIKey(q db.QueryContext, appID string, key types.ApplicationAPIKey) error {
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

// ListAPIKeys returns all the registered application API keys that belong to a
// given Application.
func (s *ApplicationStore) ListAPIKeys(appID string) ([]types.ApplicationAPIKey, error) {
	return s.listAPIKeys(s.db, appID)
}

func (s *ApplicationStore) listAPIKeys(q db.QueryContext, appID string) ([]types.ApplicationAPIKey, error) {
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

// RemoveAPIKey removes an Application API key from a given Application.
func (s *ApplicationStore) RemoveAPIKey(appID string, keyName string) error {
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.removeAPIKey(tx, appID, keyName)
	})
	return err
}

func (s *ApplicationStore) removeAPIKey(q db.QueryContext, appID string, keyName string) error {
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
		return ErrApplicationAPIKeyNotFound.New(errors.Attributes{
			"application_id": appID,
			"key_name":       keyName,
		})
	}
	return err
}

// AddCollaborator adds an Application collaborator.
func (s *ApplicationStore) AddCollaborator(appID string, collaborator types.Collaborator) error {
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.addCollaborator(tx, appID, collaborator)
	})
	return err
}

func (s *ApplicationStore) addCollaborator(q db.QueryContext, appID string, collaborator types.Collaborator) error {
	for _, right := range collaborator.Rights {
		err := s.addRight(q, appID, collaborator.Username, right)
		if err != nil {
			return err
		}
	}
	return nil
}

// ListCollaborators retrieves all the collaborators from an Application.
func (s *ApplicationStore) ListCollaborators(appID string) ([]types.Collaborator, error) {
	return s.listCollaborators(s.db, appID)
}

func (s *ApplicationStore) listCollaborators(q db.QueryContext, appID string) ([]types.Collaborator, error) {
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

// RemoveCollaborator removes a collaborator from an Application.
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
		return ErrApplicationCollaboratorNotFound.New(errors.Attributes{
			"application_id": appID,
			"username":       username,
		})
	}
	return err
}

// AddRight grants a given right to a given User.
func (s *ApplicationStore) AddRight(appID string, username string, right types.Right) error {
	return s.addRight(s.db, appID, username, right)
}

func (s *ApplicationStore) addRight(q db.QueryContext, appID string, username string, right types.Right) error {
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

// ListUserRights returns the rights a given User has for an Application.
func (s *ApplicationStore) ListUserRights(appID string, username string) ([]types.Right, error) {
	return s.listUserRights(s.db, appID, username)
}

func (s *ApplicationStore) listUserRights(q db.QueryContext, appID string, username string) ([]types.Right, error) {
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

// RemoveRight revokes a given right to a given collaborator.
func (s *ApplicationStore) RemoveRight(appID string, username string, right types.Right) error {
	return s.removeRight(s.db, appID, username, right)
}

func (s *ApplicationStore) removeRight(q db.QueryContext, appID string, username string, right types.Right) error {
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
		return ErrApplicationCollaboratorRightNotFound.New(errors.Attributes{
			"application_id": appID,
			"username":       username,
			"right":          right,
		})
	}
	return err
}

// LoadAttributes loads extra attributes into the Application.
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

// SetFactory allows to replace the DefaultApplication factory.
func (s *ApplicationStore) SetFactory(factory factory.ApplicationFactory) {
	s.factory = factory
}
