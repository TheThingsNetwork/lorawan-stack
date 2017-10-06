// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"fmt"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/factory"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/helpers"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// ClientStore implements store.ClientStore.
type ClientStore struct {
	storer
	*collaboratorStore
	factory factory.ClientFactory
}

func init() {
	ErrClientNotFound.Register()
	ErrClientIDTaken.Register()
}

// ErrClientNotFound is returned when trying to fetch a client that does not exists.
var ErrClientNotFound = &errors.ErrDescriptor{
	MessageFormat: "Client `{client_id}` does not exist",
	Code:          20,
	Type:          errors.NotFound,
}

// ErrClientIDTaken is returned when trying to create a new client with an ID.
// that already exists
var ErrClientIDTaken = &errors.ErrDescriptor{
	MessageFormat: "Client id `{client_id}` is already taken",
	Code:          21,
	Type:          errors.AlreadyExists,
}

func NewClientStore(store storer, factory factory.ClientFactory) *ClientStore {
	return &ClientStore{
		storer:            store,
		factory:           factory,
		collaboratorStore: newCollaboratorStore(store, "client"),
	}
}

// Create creates a client.
func (s *ClientStore) Create(client types.Client) error {
	err := s.transact(func(tx *db.Tx) error {
		return s.create(tx, client)
	})
	return err
}

func (s *ClientStore) create(q db.QueryContext, client types.Client) error {
	cli := client.GetClient()
	_, err := q.NamedExec(
		`INSERT
			INTO clients (client_id, description, secret, redirect_uri, grants,
					rights, updated_at, archived_at)
			VALUES (:client_id, :description, :secret, :redirect_uri, :grants,
					:rights, :updated_at, :archived_at)`,
		cli)

	if _, yes := db.IsDuplicate(err); yes {
		return ErrClientIDTaken.New(errors.Attributes{
			"client_id": cli.ClientID,
		})
	}

	if err != nil {
		return err
	}

	return s.writeAttributes(q, client, nil)
}

// GetByID finds a client by ID and retrieves it.
func (s *ClientStore) GetByID(clientID string) (types.Client, error) {
	result := s.factory.BuildClient()
	err := s.transact(func(tx *db.Tx) error {
		return s.client(tx, clientID, result)
	})
	return result, err
}

func (s *ClientStore) client(q db.QueryContext, clientID string, result types.Client) error {
	err := q.SelectOne(result, "SELECT * FROM clients WHERE client_id = $1", clientID)
	if db.IsNoRows(err) {
		return ErrClientNotFound.New(errors.Attributes{
			"client_id": clientID,
		})
	}
	if err != nil {
		return err
	}
	return s.loadAttributes(q, result)
}

// ListByUser finds all the clients an user is collaborator to.
func (s *ClientStore) ListByUser(userID string) ([]types.Client, error) {
	var result []types.Client
	err := s.transact(func(tx *db.Tx) error {
		clientIDs, err := s.userClients(tx, userID)
		if err != nil {
			return err
		}

		for _, clientID := range clientIDs {
			client := s.factory.BuildClient()

			err := s.client(tx, clientID, client)
			if err != nil {
				return err
			}

			result = append(result, client)
		}

		return nil
	})
	return result, err
}

func (s *ClientStore) userClients(q db.QueryContext, userID string) ([]string, error) {
	var clientIDs []string
	err := q.Select(
		&clientIDs,
		`SELECT DISTINCT client_id
			FROM clients_collaborators
			WHERE user_id = $1`,
		userID)
	if !db.IsNoRows(err) && err != nil {
		return nil, err
	}
	return clientIDs, nil
}

// Update updates the client.
func (s *ClientStore) Update(client types.Client) error {
	err := s.transact(func(tx *db.Tx) error {
		return s.update(tx, client)
	})
	return err
}

func (s *ClientStore) update(q db.QueryContext, client types.Client) error {
	cli := client.GetClient()
	cli.UpdatedAt = time.Now()

	_, err := q.NamedExec(
		`UPDATE clients
			SET description = :description, secret = :secret, callback_uri = :callback_uri,
			grants = :grants, rights = :rights, updated_at = :updated_at
			WHERE client_id = :client_id`,
		cli)

	if db.IsNoRows(err) {
		return ErrClientNotFound.New(errors.Attributes{
			"client_id": cli.ClientID,
		})
	}

	if err != nil {
		return err
	}

	return s.writeAttributes(q, client, nil)
}

// Archive disables a Client.
func (s *ClientStore) Archive(clientID string) error {
	return s.archive(s.queryer(), clientID)
}

func (s *ClientStore) archive(q db.QueryContext, clientID string) error {
	var i string
	err := q.SelectOne(
		&i,
		`UPDATE clients
			SET archived_at = $1
			WHERE client_id = $2
			RETURNING client_id`,
		time.Now(),
		clientID)
	if db.IsNoRows(err) {
		return ErrClientNotFound.New(errors.Attributes{
			"client_id": clientID,
		})
	}
	return err
}

func (s *ClientStore) SetClientOfficial(clientID string, official bool) error {
	return s.setClientOfficial(s.queryer(), clientID, official)
}

func (s *ClientStore) setClientOfficial(q db.QueryContext, clientID string, official bool) error {
	var id string
	err := q.SelectOne(
		&id,
		`UPDATE clients
			SET official_labeled = $1, updated_at = $2
			WHERE client_id = $3
			RETURNING client_id`,
		official,
		time.Now(),
		clientID)
	if db.IsNoRows(err) {
		return ErrClientNotFound.New(errors.Attributes{
			"client_id": clientID,
		})
	}
	return err
}

// Reject marks a Client as rejected by the tenant admins, so it cannot be used anymore.
func (s *ClientStore) SetClientState(clientID string, state ttnpb.ClientState) error {
	return s.setClientState(s.queryer(), clientID, state)
}

func (s *ClientStore) setClientState(q db.QueryContext, clientID string, state ttnpb.ClientState) error {
	var id string
	err := q.SelectOne(
		&id,
		`UPDATE clients
			SET state = $1, updated_at = $2
			WHERE client_id = $3
			RETURNING client_id`,
		state,
		time.Now(),
		clientID)
	if db.IsNoRows(err) {
		return ErrClientNotFound.New(errors.Attributes{
			"client_id": clientID,
		})
	}
	return err
}

// LoadAttributes loads the client attributes into result if it is an Attributer.
func (s *ClientStore) LoadAttributes(client types.Client) error {
	return s.transact(func(tx *db.Tx) error {
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
			client.GetClient().ClientID)
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
	return s.transact(func(tx *db.Tx) error {
		return s.writeAttributes(tx, client, result)
	})
}

func (s *ClientStore) writeAttributes(q db.QueryContext, client, result types.Client) error {
	attr, ok := client.(store.Attributer)
	if !ok {
		return nil
	}

	for _, namespace := range attr.Namespaces() {
		query, values := helpers.WriteAttributes(attr, namespace, "clients", "client_id", client.GetClient().ClientID)

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

// SetFactory allows to replace the default ttnpb.Client factory.
func (s *ClientStore) SetFactory(factory factory.ClientFactory) {
	s.factory = factory
}
