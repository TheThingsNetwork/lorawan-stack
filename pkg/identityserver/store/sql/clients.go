// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// ClientStore implements store.ClientStore.
type ClientStore struct {
	storer
	*extraAttributesStore
}

func NewClientStore(store storer) *ClientStore {
	return &ClientStore{
		storer:               store,
		extraAttributesStore: newExtraAttributesStore(store, "client"),
	}
}

// Create creates a client.
func (s *ClientStore) Create(client types.Client) error {
	err := s.transact(func(tx *db.Tx) error {
		err := s.create(tx, client)
		if err != nil {
			return err
		}

		return s.storeAttributes(tx, client.GetClient().ClientID, client, nil)
	})
	return err
}

func (s *ClientStore) create(q db.QueryContext, client types.Client) error {
	var cli struct {
		*ttnpb.Client
		GrantsConverted db.Int32Slice
		RightsConverted db.Int32Slice
	}
	cli.Client = client.GetClient()

	rights, err := db.NewInt32Slice(cli.Client.Rights)
	if err != nil {
		return err
	}
	cli.RightsConverted = rights

	grants, err := db.NewInt32Slice(cli.Client.Grants)
	if err != nil {
		return err
	}
	cli.GrantsConverted = grants

	_, err = q.NamedExec(
		`INSERT
			INTO clients (
				client_id,
				description,
				secret,
				redirect_uri,
				grants,
				state,
				rights,
				official_labeled)
			VALUES (
				:client_id,
				:description,
				:secret,
				:redirect_uri,
				:grants_converted,
				:state,
				:rights_converted,
				:official_labeled)`,
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

		return s.loadAttributes(tx, clientID, result)
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *ClientStore) getByID(q db.QueryContext, clientID string, result types.Client) error {
	var res struct {
		*ttnpb.Client
		GrantsConverted db.Int32Slice
		RightsConverted db.Int32Slice
	}

	err := q.SelectOne(&res,
		`SELECT
				client_id,
				description,
				secret,
				redirect_uri,
				grants AS grants_converted,
				state,
				rights AS rights_converted,
				official_labeled,
				created_at,
				updated_at,
				archived_at
			FROM clients
			WHERE client_id = $1`,
		clientID)
	if db.IsNoRows(err) {
		return ErrClientNotFound.New(errors.Attributes{
			"client_id": clientID,
		})
	}

	res.RightsConverted.SetInto(&res.Client.Rights)
	res.GrantsConverted.SetInto(&res.Client.Grants)
	*(result.GetClient()) = *res.Client

	return err
}

// Update updates the client.
func (s *ClientStore) Update(client types.Client) error {
	err := s.transact(func(tx *db.Tx) error {
		err := s.update(tx, client)
		if err != nil {
			return err
		}

		return s.storeAttributes(tx, client.GetClient().ClientID, client, nil)
	})
	return err
}

func (s *ClientStore) update(q db.QueryContext, client types.Client) error {
	var cli struct {
		*ttnpb.Client
		GrantsConverted db.Int32Slice
		RightsConverted db.Int32Slice
	}
	cli.Client = client.GetClient()

	rights, err := db.NewInt32Slice(cli.Client.Rights)
	if err != nil {
		return err
	}
	cli.RightsConverted = rights

	grants, err := db.NewInt32Slice(cli.Client.Grants)
	if err != nil {
		return err
	}
	cli.GrantsConverted = grants

	_, err = q.NamedExec(
		`UPDATE clients
			SET
				description = :description,
				secret = :secret,
				redirect_uri = :redirect_uri,
				grants = :grants_converted,
				state = :state,
				official_labeled = :official_labeled,
				rights = :rights_converted,
				updated_at = current_timestamp()
			WHERE client_id = :client_id`,
		cli)

	if db.IsNoRows(err) {
		return ErrClientNotFound.New(errors.Attributes{
			"client_id": cli.ClientID,
		})
	}

	return err
}

// LoadAttributes loads the extra attributes in cli if it is a store.Attributer.
func (s *ClientStore) LoadAttributes(clientID string, cli types.Client) error {
	return s.loadAttributes(s.queryer(), clientID, cli)
}

func (s *ClientStore) loadAttributes(q db.QueryContext, clientID string, cli types.Client) error {
	attr, ok := cli.(store.Attributer)
	if ok {
		return s.extraAttributesStore.loadAttributes(q, clientID, attr)
	}

	return nil
}

// StoreAttributes store the extra attributes of cli if it is a store.Attributer
// and writes the resulting application in result.
func (s *ClientStore) StoreAttributes(clientID string, cli, result types.Client) error {
	return s.storeAttributes(s.queryer(), clientID, cli, result)
}

func (s *ClientStore) storeAttributes(q db.QueryContext, clientID string, cli, result types.Client) error {
	attr, ok := cli.(store.Attributer)
	if ok {
		res, ok := result.(store.Attributer)
		if result == nil || !ok {
			return s.extraAttributesStore.storeAttributes(q, clientID, attr, nil)
		}

		return s.extraAttributesStore.storeAttributes(q, clientID, attr, res)
	}

	return nil
}
