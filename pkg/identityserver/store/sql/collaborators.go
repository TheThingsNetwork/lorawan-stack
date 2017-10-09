// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"fmt"
	"strings"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// collaboratorStore is a store that can be reused amongs other stores and holds
// the logic to manage collaborators of an entity.
type collaboratorStore struct {
	*Store
	table      string
	foreignKey string
}

// newCollaboratorStore creates a newCollaboratorStore.
func newCollaboratorStore(store *Store, table, foreignKey string) *collaboratorStore {
	return &collaboratorStore{
		Store:      store,
		table:      table,
		foreignKey: foreignKey,
	}
}

// SetCollaborator inserts or modifies a collaborator within an entity.
// If the provided list of rights is empty the collaborator will be unset.
func (s *collaboratorStore) SetCollaborator(entityID string, collaborator ttnpb.Collaborator) error {
	if len(collaborator.Rights) == 0 {
		return s.unsetCollaborator(s.db, entityID, collaborator.UserID)
	}

	err := s.db.Transact(func(tx *db.Tx) error {
		return s.setCollaborator(tx, entityID, collaborator)
	})
	return err
}

func (s *collaboratorStore) unsetCollaborator(q db.QueryContext, entityID, userID string) error {
	query := fmt.Sprintf(`
		DELETE
			FROM %s
			WHERE %s = $1 AND user_id = $2`,
		s.table,
		s.foreignKey)

	_, err := q.Exec(query, entityID, userID)
	return err
}

func (s *collaboratorStore) setCollaborator(q db.QueryContext, entityID string, collaborator ttnpb.Collaborator) error {
	query, args := s.removeRightsDiffQuery(entityID, collaborator)
	_, err := q.Exec(query, args...)
	if err != nil {
		return err
	}

	query, args = s.addRightsQuery(entityID, collaborator.UserID, collaborator.Rights)
	_, err = q.Exec(query, args...)

	return err
}

func (s *collaboratorStore) removeRightsDiffQuery(entityID string, collaborator ttnpb.Collaborator) (string, []interface{}) {
	args := make([]interface{}, 2+len(collaborator.Rights))
	args[0] = entityID
	args[1] = collaborator.UserID

	boundVariables := make([]string, len(collaborator.Rights))

	for i, right := range collaborator.Rights {
		args[i+2] = right
		boundVariables[i] = fmt.Sprintf("$%d", i+3)
	}

	query := fmt.Sprintf(
		`DELETE
			FROM %s
			WHERE %s = $1 AND user_id = $2 AND "right" NOT IN (%s)`,
		s.table,
		s.foreignKey,
		strings.Join(boundVariables, ", "))

	return query, args
}

func (s *collaboratorStore) addRightsQuery(entityID, userID string, rights []ttnpb.Right) (string, []interface{}) {
	args := make([]interface{}, 2+len(rights))
	args[0] = entityID
	args[1] = userID

	boundValues := make([]string, len(rights))

	for i, right := range rights {
		args[i+2] = right
		boundValues[i] = fmt.Sprintf("($1, $2, $%d)", i+3)
	}

	query := fmt.Sprintf(`
			INSERT
				INTO %s (%s, user_id, "right")
				VALUES %s
				ON CONFLICT (%s, user_id, "right")
				DO NOTHING`,
		s.table,
		s.foreignKey,
		strings.Join(boundValues, " ,"),
		s.foreignKey)

	return query, args
}

// ListCollaborators retrieves all the collaborators from an entity.
func (s *collaboratorStore) ListCollaborators(entityID string) ([]ttnpb.Collaborator, error) {
	return s.listCollaborators(s.db, entityID)
}

func (s *collaboratorStore) listCollaborators(q db.QueryContext, entityID string) ([]ttnpb.Collaborator, error) {
	query := fmt.Sprintf(`
		SELECT user_id, "right"
			FROM %s
			WHERE %s = $1`,
		s.table,
		s.foreignKey)

	var collaborators []struct {
		ttnpb.Collaborator
		Right ttnpb.Right
	}
	err := q.Select(&collaborators, query, entityID)
	if !db.IsNoRows(err) && err != nil {
		return nil, err
	}

	byUser := make(map[string]*ttnpb.Collaborator)
	for _, collaborator := range collaborators {
		if _, exists := byUser[collaborator.UserID]; !exists {
			byUser[collaborator.UserID] = &ttnpb.Collaborator{
				UserIdentifier: ttnpb.UserIdentifier{collaborator.UserID},
				Rights:         []ttnpb.Right{collaborator.Right},
			}
			continue
		}

		byUser[collaborator.UserID].Rights = append(byUser[collaborator.UserID].Rights, collaborator.Right)
	}

	result := make([]ttnpb.Collaborator, 0, len(byUser))
	for _, collaborator := range byUser {
		result = append(result, *collaborator)
	}

	return result, nil
}

// ListUserRights returns the rights a given user has for an entity.
func (s *collaboratorStore) ListUserRights(entityID string, userID string) ([]ttnpb.Right, error) {
	return s.listUserRights(s.db, entityID, userID)
}

func (s *collaboratorStore) listUserRights(q db.QueryContext, entityID string, userID string) ([]ttnpb.Right, error) {
	query := fmt.Sprintf(`
		SELECT "right"
			FROM %s
			WHERE %s = $1 AND user_id = $2`,
		s.table,
		s.foreignKey)

	var rights []ttnpb.Right
	err := q.Select(&rights, query, entityID, userID)

	return rights, err
}
