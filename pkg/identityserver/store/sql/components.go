// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"errors"
	"fmt"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/factory"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/helpers"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
)

// ComponentStore implements store.ComponentStore.
type ComponentStore struct {
	*Store
	factory factory.ComponentFactory
}

// ErrComponentNotFound is returned when trying to fetch a component that
// does not exist.
var ErrComponentNotFound = errors.New("component not found")

// ErrComponentIDTaken is returned when trying to create a new component with
// an ID that already exists.
var ErrComponentIDTaken = errors.New("component ID already taken")

// ErrComponentCollaboratorNotFound is returned when trying to remove a collaborator
// that does not exist.
var ErrComponentCollaboratorNotFound = errors.New("component collaborator not found")

// ErrComponentCollaboratorRightNotFound is returned when trying to revoke a right
// from a collaborator that is not granted.
var ErrComponentCollaboratorRightNotFound = errors.New("component collaborator right not found")

// Register creates a new Component and returns the new created Component.
func (s *ComponentStore) Register(component types.Component) (types.Component, error) {
	result := s.factory.Component()
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.register(tx, component, result)
	})
	return result, err
}

func (s *ComponentStore) register(q db.QueryContext, component, result types.Component) error {
	comp := component.GetComponent()
	err := q.NamedSelectOne(
		result,
		"INSERT INTO components (id, type) VALUES (:id, :type) RETURNING *",
		comp)

	if _, yes := db.IsDuplicate(err); yes {
		return ErrComponentIDTaken
	}

	if err != nil {
		return err
	}

	return s.writeAttributes(q, component, nil)
}

// FindByID finds a Component ID and returns it.
func (s *ComponentStore) FindByID(componentID string) (types.Component, error) {
	result := s.factory.Component()
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.component(tx, componentID, result)
	})
	return result, err
}

func (s *ComponentStore) component(q db.QueryContext, componentID string, result types.Component) error {
	err := q.SelectOne(result, "SELECT * FROM components WHERE id = $1", componentID)
	if db.IsNoRows(err) {
		return ErrComponentNotFound
	}
	if err != nil {
		return err
	}
	return s.loadAttributes(q, result)
}

// FindByUser retrieves all the networks Components that an User is collaborator to.
func (s *ComponentStore) FindByUser(username string) ([]types.Component, error) {
	var components []types.Component
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.userComponents(tx, username, &components)
	})
	return components, err
}

func (s *ComponentStore) userComponents(q db.QueryContext, username string, result *[]types.Component) error {
	var componentIDs []string
	err := q.Select(
		&componentIDs,
		"SELECT component_id FROM components_collaborators WHERE username = $1",
		username)

	if !db.IsNoRows(err) && err != nil {
		return err
	}

	for _, componentID := range componentIDs {
		component := s.factory.Component()
		err := s.component(q, componentID, component)
		if err != nil {
			return err
		}
		*result = append(*result, component)
	}
	return nil
}

// Edit updates the Component and returns the updated Component.
func (s *ComponentStore) Edit(component types.Component) (types.Component, error) {
	result := s.factory.Component()
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.edit(tx, component, result)
	})
	return result, err
}

func (s *ComponentStore) edit(q db.QueryContext, component, result types.Component) error {
	comp := component.GetComponent()
	err := q.NamedSelectOne(
		result,
		"UPDATE components SET type = :type WHERE id = :id RETURNING *",
		comp)

	if err != nil {
		return err
	}

	return s.writeAttributes(q, component, nil)
}

// Delete deletes a Component and all its collaborators.
func (s *ComponentStore) Delete(componentID string) error {
	// Note: ON DELETE CASCADE is not supported yet but will be soon
	// https://github.com/cockroachdb/cockroach/issues/14848
	err := s.db.Transact(func(tx *db.Tx) error {
		collaborators, err := s.listCollaborators(tx, componentID)
		if err != nil {
			return err
		}

		for _, collaborator := range collaborators {
			err := s.removeCollaborator(tx, componentID, collaborator.Username)
			if err != nil {
				return err
			}
		}

		err = s.delete(tx, componentID)
		if err != nil {
			return err
		}

		return nil
	})
	return err
}

func (s *ComponentStore) delete(q db.QueryContext, componentID string) error {
	var i string
	err := q.SelectOne(
		&i,
		"DELETE FROM components WHERE id = $1 RETURNING id",
		componentID)
	if db.IsNoRows(err) {
		return ErrComponentNotFound
	}
	return nil
}

// AddCollaborator adds a collaborator to a Component.
func (s *ComponentStore) AddCollaborator(componentID string, collaborator types.Collaborator) error {
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.addCollaborator(tx, componentID, collaborator)
	})
	return err
}

func (s *ComponentStore) addCollaborator(q db.QueryContext, componentID string, collaborator types.Collaborator) error {
	for _, right := range collaborator.Rights {
		err := s.addRight(q, componentID, collaborator.Username, right)
		if err != nil {
			return err
		}
	}
	return nil
}

// ListCollaborators returns the collaborators of a given Component.
func (s *ComponentStore) ListCollaborators(componentID string) ([]types.Collaborator, error) {
	return s.listCollaborators(s.db, componentID)
}

func (s *ComponentStore) listCollaborators(q db.QueryContext, componentID string) ([]types.Collaborator, error) {
	var collaborators []struct {
		types.Collaborator
		Right string `db:"right"`
	}
	err := q.Select(
		&collaborators,
		`SELECT username, "right"
			FROM components_collaborators
			WHERE component_id = $1`,
		componentID)
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

// RemoveCollaborator removes a collaborator from a Component.
func (s *ComponentStore) RemoveCollaborator(componentID string, username string) error {
	return s.removeCollaborator(s.db, componentID, username)
}

func (s *ComponentStore) removeCollaborator(q db.QueryContext, componentID string, username string) error {
	var u string
	err := q.SelectOne(
		&u,
		`DELETE
			FROM components_collaborators
			WHERE component_id = $1 AND username = $2
			RETURNING username`,
		componentID,
		username)
	if db.IsNoRows(err) {
		return ErrComponentCollaboratorNotFound
	}
	return err
}

// AddRight grants a given right to a given User.
func (s *ComponentStore) AddRight(componentID string, username string, right types.Right) error {
	return s.addRight(s.db, componentID, username, right)
}

func (s *ComponentStore) addRight(q db.QueryContext, componentID string, username string, right types.Right) error {
	_, err := q.Exec(
		`INSERT
			INTO components_collaborators (component_id, username, "right")
			VALUES ($1, $2, $3)
			ON CONFLICT (component_id, username, "right")
			DO NOTHING`,
		componentID,
		username,
		right)
	return err
}

// ListUserRights returns the rights the User has for a Component.
func (s *ComponentStore) ListUserRights(componentID string, username string) ([]types.Right, error) {
	return s.listUserRights(s.db, componentID, username)
}

func (s *ComponentStore) listUserRights(q db.QueryContext, componentID string, username string) ([]types.Right, error) {
	var rights []types.Right
	err := q.Select(
		&rights,
		`SELECT "right"
			FROM components_collaborators
			WHERE component_id = $1 AND username = $2`,
		componentID,
		username)
	return rights, err
}

// RemoveRight revokes a given right to a given collaborator.
func (s *ComponentStore) RemoveRight(componentID string, username string, right types.Right) error {
	return s.removeRight(s.db, componentID, username, right)
}

func (s *ComponentStore) removeRight(q db.QueryContext, componentID string, username string, right types.Right) error {
	var u string
	err := q.SelectOne(
		&u,
		`DELETE
			FROM components_collaborators
			WHERE component_id = $1 AND username = $2 AND "right" = $3
			RETURNING username`,
		componentID,
		username,
		right)
	if db.IsNoRows(err) {
		return ErrComponentCollaboratorRightNotFound
	}
	return err
}

// LoadAttributes loads extra attributes into the Component if it's an Attributer.
func (s *ComponentStore) LoadAttributes(component types.Component) error {
	return s.db.Transact(func(tx *db.Tx) error {
		return s.loadAttributes(tx, component)
	})
}

func (s *ComponentStore) loadAttributes(q db.QueryContext, component types.Component) error {
	attr, ok := component.(store.Attributer)
	if !ok {
		return nil
	}

	for _, namespace := range attr.Namespaces() {
		m := make(map[string]interface{})
		err := q.Select(
			&m,
			fmt.Sprintf("SELECT * FROM %s_components WHERE component_id = $1", namespace),
			component.GetComponent().ID)
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

// WriteAttributes writes the extra attributes on the Component if it's an
// Attributer to the store.
func (s *ComponentStore) WriteAttributes(component, result types.Component) error {
	return s.db.Transact(func(tx *db.Tx) error {
		return s.writeAttributes(tx, component, result)
	})
}

func (s *ComponentStore) writeAttributes(q db.QueryContext, component, result types.Component) error {
	attr, ok := component.(store.Attributer)
	if !ok {
		return nil
	}

	for _, namespace := range attr.Namespaces() {
		query, values := helpers.WriteAttributes(attr, namespace, "components", "component_id", component.GetComponent().ID)

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

// SetFactory allows to replace the DefaultComponent factory.
func (s *ComponentStore) SetFactory(factory factory.ComponentFactory) {
	s.factory = factory
}
