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

// ComponentStore implements store.ComponentStore
type ComponentStore struct {
	*Store
	factory factory.ComponentFactory
}

// ErrComponentNotFound is returned when trying to fetch a component that
// does not exist
var ErrComponentNotFound = errors.New("component not found")

// ErrComponentIDTaken is returned when trying to create a new component with
// an ID that already exists
var ErrComponentIDTaken = errors.New("component ID already taken")

// ErrComponentCollaboratorNotFound is returned when trying to remove a collaborator
// that does not exist
var ErrComponentCollaboratorNotFound = errors.New("component collaborator not found")

// ErrComponentCollaboratorRightNotFound is returned when trying to revoke a right
// from a collaborator that is not granted
var ErrComponentCollaboratorRightNotFound = errors.New("component collaborator right not found")

// SetFactory replaces the factory
func (s *ComponentStore) SetFactory(factory factory.ComponentFactory) {
	s.factory = factory
}

// LoadAttributes loads the component attributes into result if it is an
// ComponentAttributer
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

// WriteAttributes writes the component attributes into result if it is an ComponentAttributer
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

// FindByID finds the component by its ID
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

// FindByUser returns the components to which a user is a collaborator
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

// Create creates a new component and returns the resulting component
func (s *ComponentStore) Create(component types.Component) (types.Component, error) {
	result := s.factory.Component()
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.create(tx, component, result)
	})
	return result, err
}

func (s *ComponentStore) create(q db.QueryContext, component, result types.Component) error {
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

// Update udpates a component and returns the resulting component
func (s *ComponentStore) Update(component types.Component) (types.Component, error) {
	result := s.factory.Component()
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.update(tx, component, result)
	})
	return result, err
}

func (s *ComponentStore) update(q db.QueryContext, component, result types.Component) error {
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

// Delete deletes a component
func (s *ComponentStore) Delete(componentID string) error {
	// Note: ON DELETE CASCADE is not supported yet but will be soon
	// https://github.com/cockroachdb/cockroach/issues/14848
	err := s.db.Transact(func(tx *db.Tx) error {
		collaborators, err := s.collaborators(tx, componentID)
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

// Collaborators returns the list of collaborators to a given component
func (s *ComponentStore) Collaborators(componentID string) ([]types.Collaborator, error) {
	return s.collaborators(s.db, componentID)
}

func (s *ComponentStore) collaborators(q db.QueryContext, componentID string) ([]types.Collaborator, error) {
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

// AddCollaborator adds a new collaborator to a component
func (s *ComponentStore) AddCollaborator(componentID string, collaborator types.Collaborator) error {
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.addCollaborator(tx, componentID, collaborator)
	})
	return err
}

func (s *ComponentStore) addCollaborator(q db.QueryContext, componentID string, collaborator types.Collaborator) error {
	for _, right := range collaborator.Rights {
		err := s.grantRight(q, componentID, collaborator.Username, right)
		if err != nil {
			return err
		}
	}
	return nil
}

// GrantRight grants a right to a specific user in a given component
func (s *ComponentStore) GrantRight(componentID string, username string, right types.Right) error {
	return s.grantRight(s.db, componentID, username, right)
}

func (s *ComponentStore) grantRight(q db.QueryContext, componentID string, username string, right types.Right) error {
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

// RevokeRight revokes a specific right to a specific user in a given component
func (s *ComponentStore) RevokeRight(componentID string, username string, right types.Right) error {
	return s.revokeRight(s.db, componentID, username, right)
}

func (s *ComponentStore) revokeRight(q db.QueryContext, componentID string, username string, right types.Right) error {
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

// RemoveCollaborator removes a collaborator of a given component
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

// UserRights returns the list of rights that an user has to a given component
func (s *ComponentStore) UserRights(componentID string, username string) ([]types.Right, error) {
	return s.userRights(s.db, componentID, username)
}

func (s *ComponentStore) userRights(q db.QueryContext, componentID string, username string) ([]types.Right, error) {
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
