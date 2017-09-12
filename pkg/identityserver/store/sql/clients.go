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

// ClientStore implements store.ClientStore.
type ClientStore struct {
	*Store
	factory factory.ClientFactory
}

// ErrClientNotFound is returned when trying to fetch a client that does not exists.
var ErrClientNotFound = errors.New("client not found")

// ErrClientIDTaken is returned when trying to create a new client with an ID.
// that already exists
var ErrClientIDTaken = errors.New("client ID already taken")

// ErrClientCollaboratorNotFound is returned when trying to remove a collaborator.
// that does not exist
var ErrClientCollaboratorNotFound = errors.New("client collaborator not found")

// ErrClientCollaboratorRightNotFound is returned when trying to revoke a right.
// from a collaborator that is not granted
var ErrClientCollaboratorRightNotFound = errors.New("client collaborator right not found")

// Register creates a new Client and returns the new created Client.
func (s *ClientStore) Register(client types.Client) (types.Client, error) {
	result := s.factory.Client()
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.register(tx, client, result)
	})
	return result, err
}

func (s *ClientStore) register(q db.QueryContext, client, result types.Client) error {
	cli := client.GetClient()
	err := q.NamedSelectOne(
		result,
		`INSERT
			INTO clients (id, description, secret, uri, grants, scope)
			VALUES (:id, :description, :secret, :uri, :grants, :scope)
			RETURNING *`,
		cli)

	if _, yes := db.IsDuplicate(err); yes {
		return ErrClientIDTaken
	}

	if err != nil {
		return err
	}

	return s.writeAttributes(q, client, result)
}

// FindByID finds a Client by ID and retrieves it.
func (s *ClientStore) FindByID(clientID string) (types.Client, error) {
	result := s.factory.Client()
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.client(tx, clientID, result)
	})
	return result, err
}

func (s *ClientStore) client(q db.QueryContext, clientID string, result types.Client) error {
	err := q.SelectOne(result, "SELECT * FROM clients WHERE id = $1", clientID)
	if db.IsNoRows(err) {
		return ErrClientNotFound
	}
	if err != nil {
		return err
	}
	return s.loadAttributes(q, result)
}

// FindByUser finds all the Clients an user is collaborator to.
func (s *ClientStore) FindByUser(username string) ([]types.Client, error) {
	var result []types.Client
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.userClients(tx, username, &result)
	})
	return result, err
}

func (s *ClientStore) userClients(q db.QueryContext, username string, result *[]types.Client) error {
	var clientIDs []string
	err := q.Select(
		&clientIDs,
		`SELECT DISTINCT client_id
			FROM clients_collaborators
			WHERE username = $1`,
		username)
	if !db.IsNoRows(err) && err != nil {
		return err
	}

	for _, clientID := range clientIDs {
		client := s.factory.Client()
		err := s.client(q, clientID, client)
		if err != nil {
			return err
		}
		*result = append(*result, client)
	}

	return nil
}

// Edit updates the Client and returns the updated Client.
func (s *ClientStore) Edit(client types.Client) (types.Client, error) {
	result := s.factory.Client()
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.edit(tx, client, result)
	})
	return result, err
}

func (s *ClientStore) edit(q db.QueryContext, client, result types.Client) error {
	cli := client.GetClient()
	err := q.NamedSelectOne(
		result,
		`UPDATE clients
			SET description = :description, secret = :secret, uri = :uri,
			grants = :grants, scope = :scope
			WHERE id = :id
			RETURNING *`,
		cli)

	if db.IsNoRows(err) {
		return ErrClientNotFound
	}

	if err != nil {
		return err
	}

	return s.writeAttributes(q, client, result)
}

// Delete deletes a Client.
func (s *ClientStore) Delete(clientID string) error {
	// Note: ON DELETE CASCADE is not supported yet but will be soon
	// https://github.com/cockroachdb/cockroach/issues/14848
	err := s.db.Transact(func(tx *db.Tx) error {
		collaborators, err := s.listCollaborators(tx, clientID)
		if err != nil {
			return err
		}

		for _, collaborator := range collaborators {
			err := s.removeCollaborator(tx, clientID, collaborator.Username)
			if err != nil {
				return err
			}
		}

		err = s.delete(tx, clientID)
		if err != nil {
			return err
		}

		return nil
	})
	return err
}

func (s *ClientStore) delete(q db.QueryContext, clientID string) error {
	var i string
	err := q.SelectOne(&i, "DELETE FROM clients WHERE id = $1 RETURNING id", clientID)
	if db.IsNoRows(err) {
		return ErrClientNotFound
	}
	return err
}

// Archive disables a Client.
func (s *ClientStore) Archive(clientID string) error {
	return s.archive(s.db, clientID)
}

func (s *ClientStore) archive(q db.QueryContext, clientID string) error {
	var i string
	err := q.SelectOne(
		&i,
		`UPDATE clients
			SET archived = $1
			WHERE id = $2
			RETURNING id`,
		time.Now(),
		clientID)
	if db.IsNoRows(err) {
		return ErrClientNotFound
	}
	return err
}

// Approve marks a Client approved by the tenant admins, so it can be used.
func (s *ClientStore) Approve(clientID string) error {
	return s.approve(s.db, clientID)
}

func (s *ClientStore) approve(q db.QueryContext, clientID string) error {
	var i string
	err := q.SelectOne(
		&i,
		`UPDATE clients
			SET state = $2
			WHERE id = $1
			RETURNING id`,
		clientID,
		types.ApprovedClient)
	if db.IsNoRows(err) {
		return ErrClientNotFound
	}
	return err
}

// Reject marks a Client as rejected by the tenant admins, so it cannot be used anymore.
func (s *ClientStore) Reject(clientID string) error {
	return s.reject(s.db, clientID)
}

func (s *ClientStore) reject(q db.QueryContext, clientID string) error {
	var i string
	err := q.SelectOne(
		&i,
		`UPDATE clients
			SET state = $2
			WHERE id = $1
			RETURNING id`,
		clientID,
		types.RejectedClient)
	if db.IsNoRows(err) {
		return ErrClientNotFound
	}
	return err
}

// AddCollaborator adds a collaborator to a given Client.
func (s *ClientStore) AddCollaborator(clientID string, collaborator types.Collaborator) error {
	err := s.db.Transact(func(tx *db.Tx) error {
		return s.addCollaborator(tx, clientID, collaborator)
	})
	return err
}

func (s *ClientStore) addCollaborator(q db.QueryContext, clientID string, collaborator types.Collaborator) error {
	for _, right := range collaborator.Rights {
		err := s.addRight(q, clientID, collaborator.Username, right)
		if err != nil {
			return err
		}
	}
	return nil
}

// ListCollaborators retrieves the collaborators for a given Client.
func (s *ClientStore) ListCollaborators(clientID string) ([]types.Collaborator, error) {
	return s.listCollaborators(s.db, clientID)
}

func (s *ClientStore) listCollaborators(q db.QueryContext, clientID string) ([]types.Collaborator, error) {
	var collaborators []struct {
		types.Collaborator
		Right string `db:"right"`
	}
	err := q.Select(
		&collaborators,
		`SELECT username, "right"
			FROM clients_collaborators
			WHERE client_id = $1`,
		clientID)
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

// RemoveCollaborator removes a collaborator from a given Client.
func (s *ClientStore) RemoveCollaborator(clientID string, username string) error {
	return s.removeCollaborator(s.db, clientID, username)
}

func (s *ClientStore) removeCollaborator(q db.QueryContext, clientID string, username string) error {
	var u string
	err := q.SelectOne(
		&u,
		`DELETE
			FROM clients_collaborators
			WHERE client_id = $1 AND username = $2
			RETURNING username`,
		clientID,
		username)
	if db.IsNoRows(err) {
		return ErrClientCollaboratorNotFound
	}
	return err
}

// AddRight grants a given right to a given User for a Client.
func (s *ClientStore) AddRight(clientID string, username string, right types.Right) error {
	return s.addRight(s.db, clientID, username, right)
}

func (s *ClientStore) addRight(q db.QueryContext, clientID string, username string, right types.Right) error {
	_, err := q.Exec(
		`INSERT
			INTO clients_collaborators (client_id, username, "right")
			VALUES ($1, $2, $3)
			ON CONFLICT (client_id, username, "right")
			DO NOTHING`,
		clientID,
		username,
		right)
	return err
}

// ListUserRights returns the rights the user has for a Client.
func (s *ClientStore) ListUserRights(clientID string, username string) ([]types.Right, error) {
	return s.listUserRights(s.db, clientID, username)
}

func (s *ClientStore) listUserRights(q db.QueryContext, clientID string, username string) ([]types.Right, error) {
	var rights []types.Right
	err := q.Select(
		&rights,
		`SELECT "right"
			FROM clients_collaborators
			WHERE client_id = $1 AND username = $2`,
		clientID,
		username)
	return rights, err
}

// RemoveRight revokes a given right from a given Client collaborator.
func (s *ClientStore) RemoveRight(clientID string, username string, right types.Right) error {
	return s.removeRight(s.db, clientID, username, right)
}

func (s *ClientStore) removeRight(q db.QueryContext, clientID string, username string, right types.Right) error {
	var u string
	err := q.SelectOne(
		&u,
		`DELETE
			FROM clients_collaborators
			WHERE client_id = $1 AND username = $2 AND "right" = $3
			RETURNING username`,
		clientID,
		username,
		right)
	if db.IsNoRows(err) {
		return ErrClientCollaboratorRightNotFound
	}
	return err
}

// LoadAttributes loads the client attributes into result if it is an Attributer.
func (s *ClientStore) LoadAttributes(client types.Client) error {
	return s.db.Transact(func(tx *db.Tx) error {
		return s.loadAttributes(tx, client)
	})
}

func (s *ClientStore) loadAttributes(q db.QueryContext, client types.Client) error {
	attr, ok := client.(store.Attributer)
	if !ok {
		return nil
	}

	for _, namespace := range attr.Namespaces() {
		m := make(map[string]interface{})
		err := q.Select(
			&m,
			fmt.Sprintf("SELECT * FROM %s_clients WHERE client_id = $1", namespace),
			client.GetClient().ID)
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

// WriteAttributes writes the client attributes into result if it is an Attributer.
func (s *ClientStore) WriteAttributes(client, result types.Client) error {
	return s.db.Transact(func(tx *db.Tx) error {
		return s.writeAttributes(tx, client, result)
	})
}

func (s *ClientStore) writeAttributes(q db.QueryContext, client, result types.Client) error {
	attr, ok := client.(store.Attributer)
	if !ok {
		return nil
	}

	for _, namespace := range attr.Namespaces() {
		query, values := helpers.WriteAttributes(attr, namespace, "clients", "client_id", client.GetClient().ID)

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

// SetFactory allows to replace the DefaultClient factory.
func (s *ClientStore) SetFactory(factory factory.ClientFactory) {
	s.factory = factory
}
