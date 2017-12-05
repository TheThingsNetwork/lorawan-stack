// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql/helpers"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
)

// ClientStore implements store.ClientStore.
type ClientStore struct {
	storer
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
	SafeAttributes: []string{
		"client_id",
	},
}

// ErrClientIDTaken is returned when trying to create a new client with an ID.
// that already exists
var ErrClientIDTaken = &errors.ErrDescriptor{
	MessageFormat: "Client id `{client_id}` is already taken",
	Code:          21,
	Type:          errors.AlreadyExists,
	SafeAttributes: []string{
		"client_id",
	},
}

func NewClientStore(store storer) *ClientStore {
	return &ClientStore{
		storer: store,
	}
}

// Create creates a client.
func (s *ClientStore) Create(client types.Client) error {
	err := s.transact(func(tx *db.Tx) error {
		err := s.create(tx, client)
		if err != nil {
			return err
		}

		return s.writeAttributes(tx, client, nil)
	})
	return err
}

func (s *ClientStore) create(q db.QueryContext, client types.Client) error {
	cli := client.GetClient()
	_, err := q.NamedExec(
		`INSERT
			INTO clients (
				client_id,
				description,
				secret,
				redirect_uri,
				grants,
				state,
				rights,
				archived_at)
			VALUES (
				:client_id,
				:description,
				:secret,
				:redirect_uri,
				:grants,
				:state,
				:rights,
				:archived_at)`,
		cli)

	if _, yes := db.IsDuplicate(err); yes {
		return ErrClientIDTaken.New(errors.Attributes{
			"client_id": cli.ClientID,
		})
	}

	return err
}

// GetByID finds a client by ID and retrieves it.
func (s *ClientStore) GetByID(clientID string, factory store.ClientFactory) (types.Client, error) {
	result := factory()

	err := s.transact(func(tx *db.Tx) error {
		err := s.getByID(tx, clientID, result)
		if err != nil {
			return err
		}

		return s.loadAttributes(tx, result)
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *ClientStore) getByID(q db.QueryContext, clientID string, result types.Client) error {
	err := q.SelectOne(result,
		`SELECT *
			FROM clients
			WHERE client_id = $1`,
		clientID)
	if db.IsNoRows(err) {
		return ErrClientNotFound.New(errors.Attributes{
			"client_id": clientID,
		})
	}
	return err
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

	_, err := q.NamedExec(
		`UPDATE clients
			SET description = :description, secret = :secret, redirect_uri = :redirect_uri,
			grants = :grants, rights = :rights, updated_at = current_timestamp()
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
			SET archived_at = current_timestamp()
			WHERE client_id = $1
			RETURNING client_id`,
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
