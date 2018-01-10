// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// ApplicationStore implements store.ApplicationStore.
type ApplicationStore struct {
	storer
	*extraAttributesStore
	*apiKeysStore
}

func NewApplicationStore(store storer) *ApplicationStore {
	return &ApplicationStore{
		storer:               store,
		extraAttributesStore: newExtraAttributesStore(store, "application"),
		apiKeysStore:         newAPIKeysStore(store, "application"),
	}
}

// Create creates a new application.
func (s *ApplicationStore) Create(application store.Application) error {
	err := s.transact(func(tx *db.Tx) error {
		err := s.create(tx, application)
		if err != nil {
			return err
		}

		return s.storeAttributes(tx, application.GetApplication().ApplicationID, application, nil)
	})
	return err
}

func (s *ApplicationStore) create(q db.QueryContext, application store.Application) error {
	app := application.GetApplication()
	_, err := q.NamedExec(
		`INSERT
			INTO applications (
				application_id,
				description)
			VALUES (
				:application_id,
				:description)`,
		app)

	if _, yes := db.IsDuplicate(err); yes {
		return ErrApplicationIDTaken.New(errors.Attributes{
			"application_id": app.ApplicationID,
		})
	}

	return err
}

// GetByID finds the application by ID and retrieves it.
func (s *ApplicationStore) GetByID(appID string, factory store.ApplicationFactory) (store.Application, error) {
	result := factory()

	err := s.transact(func(tx *db.Tx) error {
		err := s.getByID(tx, appID, result)
		if err != nil {
			return err
		}

		return s.loadAttributes(tx, appID, result)
	}, db.ReadOnly(true))

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *ApplicationStore) getByID(q db.QueryContext, appID string, result store.Application) error {
	err := q.SelectOne(
		result,
		`SELECT *
			FROM applications
			WHERE application_id = $1`,
		appID)
	if db.IsNoRows(err) {
		return ErrApplicationNotFound.New(errors.Attributes{
			"application_id": appID,
		})
	}

	return err
}

// FindByUser returns the Applications to which an User is a collaborator.
func (s *ApplicationStore) ListByUser(userID string, factory store.ApplicationFactory) ([]store.Application, error) {
	var result []store.Application

	err := s.transact(func(tx *db.Tx) error {
		applications, err := s.userApplications(tx, userID)
		if err != nil {
			return err
		}

		for _, application := range applications {
			app := factory()
			*(app.GetApplication()) = application

			err := s.loadAttributes(tx, app.GetApplication().ApplicationID, app)
			if err != nil {
				return err
			}

			result = append(result, app)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *ApplicationStore) userApplications(q db.QueryContext, userID string) ([]ttnpb.Application, error) {
	var applications []ttnpb.Application
	err := q.Select(
		&applications,
		`SELECT DISTINCT applications.*
			FROM applications
			JOIN applications_collaborators
			ON (
				applications.application_id = applications_collaborators.application_id
				AND
				user_id = $1
			)`,
		userID)

	if err != nil {
		return nil, err
	}

	if len(applications) == 0 {
		return make([]ttnpb.Application, 0), nil
	}

	return applications, nil
}

// Edit updates the Application and returns the updated Application.
func (s *ApplicationStore) Update(application store.Application) error {
	err := s.transact(func(tx *db.Tx) error {
		err := s.update(tx, application)
		if err != nil {
			return err
		}

		return s.storeAttributes(tx, application.GetApplication().ApplicationID, application, nil)
	})
	return err
}

func (s *ApplicationStore) update(q db.QueryContext, application store.Application) error {
	app := application.GetApplication()

	var id string
	err := q.NamedSelectOne(
		&id,
		`UPDATE applications
			SET description = :description, updated_at = current_timestamp()
			WHERE application_id = :application_id
			RETURNING application_id`,
		app)

	if db.IsNoRows(err) {
		return ErrApplicationNotFound.New(errors.Attributes{
			"application_id": app.ApplicationID,
		})
	}

	return err
}

// Delete deletes an application.
func (s *ApplicationStore) Delete(appID string) error {
	err := s.transact(func(tx *db.Tx) error {
		err := s.deleteCollaborators(tx, appID)
		if err != nil {
			return err
		}

		err = s.deleteAPIKeys(tx, appID)
		if err != nil {
			return err
		}

		return s.delete(tx, appID)
	})

	return err
}

// delete deletes the application itself. All rows in other tables that references
// this entity must be delete before this one gets deleted.
func (s *ApplicationStore) delete(q db.QueryContext, appID string) error {
	id := new(string)
	err := q.SelectOne(
		id,
		`DELETE
			FROM applications
			WHERE application_id = $1
			RETURNING application_id`,
		appID)
	if db.IsNoRows(err) {
		return ErrApplicationNotFound.New(errors.Attributes{
			"application_id": appID,
		})
	}
	return err
}

// deleteCollaborators deletes all the collaborators from one application.
func (s *ApplicationStore) deleteCollaborators(q db.QueryContext, appID string) error {
	_, err := q.Exec(
		`DELETE
			FROM applications_collaborators
			WHERE application_id = $1`,
		appID)
	return err
}

// SetCollaborator inserts or modifies a collaborator within an entity.
// If the provided list of rights is empty the collaborator will be unset.
func (s *ApplicationStore) SetCollaborator(collaborator *ttnpb.ApplicationCollaborator) error {
	if len(collaborator.Rights) == 0 {
		return s.unsetCollaborator(s.queryer(), collaborator.ApplicationID, collaborator.UserID)
	}

	err := s.transact(func(tx *db.Tx) error {
		return s.setCollaborator(tx, collaborator)
	})
	return err
}

func (s *ApplicationStore) unsetCollaborator(q db.QueryContext, appID, userID string) error {
	_, err := q.Exec(
		`DELETE
			FROM applications_collaborators
			WHERE application_id = $1 AND user_id = $2`, appID, userID)
	return err
}

func (s *ApplicationStore) setCollaborator(q db.QueryContext, collaborator *ttnpb.ApplicationCollaborator) error {
	query, args := s.removeRightsDiffQuery(collaborator)
	_, err := q.Exec(query, args...)
	if err != nil {
		return err
	}

	query, args = s.addRightsQuery(collaborator.ApplicationID, collaborator.UserID, collaborator.Rights)
	_, err = q.Exec(query, args...)

	return err
}

func (s *ApplicationStore) removeRightsDiffQuery(collaborator *ttnpb.ApplicationCollaborator) (string, []interface{}) {
	args := make([]interface{}, 2+len(collaborator.Rights))
	args[0] = collaborator.ApplicationID
	args[1] = collaborator.UserID

	boundVariables := make([]string, len(collaborator.Rights))

	for i, right := range collaborator.Rights {
		args[i+2] = right
		boundVariables[i] = fmt.Sprintf("$%d", i+3)
	}

	query := fmt.Sprintf(
		`DELETE
			FROM applications_collaborators
			WHERE application_id = $1 AND user_id = $2 AND "right" NOT IN (%s)`,
		strings.Join(boundVariables, ", "))

	return query, args
}

func (s *ApplicationStore) addRightsQuery(appID, userID string, rights []ttnpb.Right) (string, []interface{}) {
	args := make([]interface{}, 2+len(rights))
	args[0] = appID
	args[1] = userID

	boundValues := make([]string, len(rights))

	for i, right := range rights {
		args[i+2] = right
		boundValues[i] = fmt.Sprintf("($1, $2, $%d)", i+3)
	}

	query := fmt.Sprintf(
		`INSERT
			INTO applications_collaborators (application_id, user_id, "right")
			VALUES %s
			ON CONFLICT (application_id, user_id, "right")
			DO NOTHING`,
		strings.Join(boundValues, " ,"))

	return query, args
}

// HasUserRights checks whether an user has a set of given rights to an application.
func (s *ApplicationStore) HasUserRights(appID, userID string, rights ...ttnpb.Right) (bool, error) {
	return s.hasUserRights(s.queryer(), appID, userID, rights...)
}

func (s *ApplicationStore) hasUserRights(q db.QueryContext, appID, userID string, rights ...ttnpb.Right) (bool, error) {
	clauses := make([]string, 0, len(rights))
	args := make([]interface{}, 0, len(rights)+1)
	args = append(args, userID)

	for i, right := range rights {
		args = append(args, right)
		clauses = append(clauses, fmt.Sprintf(`"right" = $%d`, i+2))
	}

	res := new(string)
	err := q.SelectOne(
		res,
		fmt.Sprintf(
			`SELECT
				DISTINCT user_id
				FROM applications_collaborators
				WHERE user_id = $1 AND (%s)`, strings.Join(clauses, " OR ")),
		args...)
	if db.IsNoRows(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// ListCollaborators retrieves all the collaborators from an entity.
func (s *ApplicationStore) ListCollaborators(appID string, rights ...ttnpb.Right) ([]*ttnpb.ApplicationCollaborator, error) {
	return s.listCollaborators(s.queryer(), appID, rights...)
}

func (s *ApplicationStore) listCollaborators(q db.QueryContext, appID string, rights ...ttnpb.Right) ([]*ttnpb.ApplicationCollaborator, error) {
	query := ""
	args := make([]interface{}, 1)
	args[0] = appID

	if len(rights) == 0 {
		query = `
		SELECT user_id, "right"
			FROM applications_collaborators
			WHERE application_id = $1`
	} else {
		rightsClause := make([]string, 0, len(rights))
		for _, right := range rights {
			rightsClause = append(rightsClause, fmt.Sprintf(`"right" = '%d'`, right))
		}

		query = fmt.Sprintf(`
			SELECT user_id, "right"
	    	FROM applications_collaborators
	    	WHERE application_id = $1 AND user_id IN
	    	(
	      	SELECT user_id
	      		FROM
	      			(
	          		SELECT user_id, count(user_id) as count
	          	  	FROM applications_collaborators
	          			WHERE application_id = $1 AND (%s)
	          			GROUP BY user_id
	      			)
	      		WHERE count = $2
	  		)`,
			strings.Join(rightsClause, " OR "))

		args = append(args, len(rights))
	}

	var collaborators []struct {
		*ttnpb.ApplicationCollaborator
		Right ttnpb.Right
	}
	err := q.Select(&collaborators, query, args...)
	if !db.IsNoRows(err) && err != nil {
		return nil, err
	}

	byUser := make(map[string]*ttnpb.ApplicationCollaborator)
	for _, collaborator := range collaborators {
		if _, exists := byUser[collaborator.UserID]; !exists {
			byUser[collaborator.UserID] = &ttnpb.ApplicationCollaborator{
				ApplicationIdentifier: ttnpb.ApplicationIdentifier{appID},
				UserIdentifier:        ttnpb.UserIdentifier{collaborator.UserID},
				Rights:                []ttnpb.Right{collaborator.Right},
			}
			continue
		}

		byUser[collaborator.UserID].Rights = append(byUser[collaborator.UserID].Rights, collaborator.Right)
	}

	result := make([]*ttnpb.ApplicationCollaborator, 0, len(byUser))
	for _, collaborator := range byUser {
		result = append(result, collaborator)
	}

	return result, nil
}

// ListUserRights returns the rights a given user has for an entity.
func (s *ApplicationStore) ListUserRights(appID string, userID string) ([]ttnpb.Right, error) {
	return s.listUserRights(s.queryer(), appID, userID)
}

func (s *ApplicationStore) listUserRights(q db.QueryContext, appID string, userID string) ([]ttnpb.Right, error) {
	var rights []ttnpb.Right
	err := q.Select(
		&rights,
		`SELECT "right"
			FROM applications_collaborators
			WHERE application_id = $1 AND user_id = $2`,
		appID,
		userID)

	return rights, err
}

// LoadAttributes loads the extra attributes in app if it is a store.Attributer.
func (s *ApplicationStore) LoadAttributes(appID string, app store.Application) error {
	return s.loadAttributes(s.queryer(), appID, app)
}

func (s *ApplicationStore) loadAttributes(q db.QueryContext, appID string, app store.Application) error {
	attr, ok := app.(store.Attributer)
	if ok {
		return s.extraAttributesStore.loadAttributes(q, appID, attr)
	}

	return nil
}

// StoreAttributes store the extra attributes of app if it is a store.Attributer
// and writes the resulting application in result.
func (s *ApplicationStore) StoreAttributes(appID string, app, result store.Application) error {
	return s.storeAttributes(s.queryer(), appID, app, result)
}

func (s *ApplicationStore) storeAttributes(q db.QueryContext, appID string, app, result store.Application) error {
	attr, ok := app.(store.Attributer)
	if ok {
		res, ok := result.(store.Attributer)
		if result == nil || !ok {
			return s.extraAttributesStore.storeAttributes(q, appID, attr, nil)
		}

		return s.extraAttributesStore.storeAttributes(q, appID, attr, res)
	}

	return nil
}
