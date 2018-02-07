// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
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
func (s *ClientStore) Create(client store.Client) error {
	err := s.transact(func(tx *db.Tx) error {
		err := s.create(tx, client)
		if err != nil {
			return err
		}

		return s.storeAttributes(tx, client.GetClient().ClientID, client, nil)
	})
	return err
}

func (s *ClientStore) create(q db.QueryContext, client store.Client) error {
	var cli struct {
		*ttnpb.Client
		CreatorID       string
		GrantsConverted db.Int32Slice
		RightsConverted db.Int32Slice
	}
	cli.Client = client.GetClient()
	cli.CreatorID = cli.Creator.UserID

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
				creator_id,
				official_labeled)
			VALUES (
				:client_id,
				:description,
				:secret,
				:redirect_uri,
				:grants_converted,
				:state,
				:rights_converted,
				:creator_id,
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
func (s *ClientStore) GetByID(clientID string, factory store.ClientSpecializer) (result store.Client, err error) {
	err = s.transact(func(tx *db.Tx) error {
		client, err := s.getByID(tx, clientID)
		if err != nil {
			return err
		}

		result = factory(*client)

		return s.loadAttributes(tx, clientID, result)
	})

	return
}

func (s *ClientStore) getByID(q db.QueryContext, clientID string) (*ttnpb.Client, error) {
	var res struct {
		*ttnpb.Client
		CreatorID       string
		GrantsConverted db.Int32Slice
		RightsConverted db.Int32Slice
	}

	err := q.SelectOne(
		&res,
		`SELECT
				client_id,
				description,
				secret,
				redirect_uri,
				grants AS grants_converted,
				state,
				rights AS rights_converted,
				official_labeled,
				creator_id,
				created_at,
				updated_at
			FROM clients
			WHERE client_id = $1`,
		clientID)
	if db.IsNoRows(err) {
		return nil, ErrClientNotFound.New(errors.Attributes{
			"client_id": clientID,
		})
	}

	if err != nil {
		return nil, err
	}

	res.RightsConverted.SetInto(&res.Client.Rights)
	res.GrantsConverted.SetInto(&res.Client.Grants)
	res.Client.Creator.UserID = res.CreatorID

	return res.Client, nil
}

// List returns all the clients.
func (s *ClientStore) List(factory store.ClientSpecializer) ([]store.Client, error) {
	var res []store.Client
	err := s.transact(func(tx *db.Tx) error {
		found, err := s.list(tx)
		if err != nil {
			return err
		}

		res = make([]store.Client, 0, len(found))

		for _, client := range found {
			cli := factory(client)

			err := s.loadAttributes(tx, client.ClientID, cli)
			if err != nil {
				return err
			}

			res = append(res, cli)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *ClientStore) list(q db.QueryContext) ([]ttnpb.Client, error) {
	var res []struct {
		ttnpb.Client
		CreatorID       string
		GrantsConverted db.Int32Slice
		RightsConverted db.Int32Slice
	}
	err := q.Select(
		&res,
		`SELECT
				client_id,
				description,
				secret,
				redirect_uri,
				grants AS grants_converted,
				state,
				rights AS rights_converted,
				official_labeled,
				creator_id,
				created_at,
				updated_at
			FROM clients`)
	if err != nil {
		return nil, err
	}

	clients := make([]ttnpb.Client, 0, len(res))
	for _, client := range res {
		client.RightsConverted.SetInto(&client.Client.Rights)
		client.GrantsConverted.SetInto(&client.Client.Grants)
		client.Client.Creator.UserID = client.CreatorID

		clients = append(clients, client.Client)
	}

	return clients, nil
}

// ListByUser returns all the clients created by the client.
func (s *ClientStore) ListByUser(userID string, factory store.ClientSpecializer) ([]store.Client, error) {
	var result []store.Client

	err := s.transact(func(tx *db.Tx) error {
		clients, err := s.userClients(tx, userID)
		if err != nil {
			return err
		}

		for _, client := range clients {
			cli := factory(client)

			err := s.loadAttributes(tx, cli.GetClient().ClientID, cli)
			if err != nil {
				return err
			}

			result = append(result, cli)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *ClientStore) userClients(q db.QueryContext, userID string) ([]ttnpb.Client, error) {
	var clients []struct {
		CreatorID       string
		GrantsConverted db.Int32Slice
		RightsConverted db.Int32Slice
		ttnpb.Client
	}
	err := q.Select(
		&clients,
		`SELECT
				client_id,
				description,
				secret,
				redirect_uri,
				grants AS grants_converted,
				state,
				rights AS rights_converted,
				official_labeled,
				creator_id,
				created_at,
				updated_at
			FROM clients
			WHERE	creator_id = $1`,
		userID)

	if err != nil {
		return nil, err
	}

	if len(clients) == 0 {
		return make([]ttnpb.Client, 0), nil
	}

	res := make([]ttnpb.Client, 0, len(clients))
	for _, client := range clients {
		client.RightsConverted.SetInto(&client.Client.Rights)
		client.GrantsConverted.SetInto(&client.Client.Grants)
		client.Client.Creator.UserID = client.CreatorID

		res = append(res, client.Client)
	}

	return res, nil
}

// Update updates the client.
func (s *ClientStore) Update(client store.Client) error {
	err := s.transact(func(tx *db.Tx) error {
		err := s.update(tx, client)
		if err != nil {
			return err
		}

		return s.storeAttributes(tx, client.GetClient().ClientID, client, nil)
	})
	return err
}

func (s *ClientStore) update(q db.QueryContext, client store.Client) error {
	var cli struct {
		*ttnpb.Client
		CreatorID       string
		GrantsConverted db.Int32Slice
		RightsConverted db.Int32Slice
	}
	cli.Client = client.GetClient()
	cli.CreatorID = cli.Creator.UserID

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
				creator_id = :creator_id,
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

// Delete deletes a client.
func (s *ClientStore) Delete(clientID string) error {
	err := s.transact(func(tx *db.Tx) error {
		oauth, ok := s.store().OAuth.(*OAuthStore)
		if !ok {
			return errors.Errorf("Expected ptr to OAuthStore but got %T", s.store().OAuth)
		}

		err := oauth.deleteAuthorizationCodesByClient(tx, clientID)
		if err != nil {
			return err
		}

		err = oauth.deleteAccessTokensByClient(tx, clientID)
		if err != nil {
			return err
		}

		err = oauth.deleteRefreshTokensByClient(tx, clientID)
		if err != nil {
			return err
		}

		return s.delete(tx, clientID)
	})

	return err
}

// delete deletes the client itself. All rows in other tables that references
// this entity must be deleted before this one gets deleted.
func (s *ClientStore) delete(q db.QueryContext, clientID string) error {
	id := new(string)
	err := q.SelectOne(
		id,
		`DELETE
			FROM clients
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

// LoadAttributes loads the extra attributes in cli if it is a store.Attributer.
func (s *ClientStore) LoadAttributes(clientID string, cli store.Client) error {
	return s.loadAttributes(s.queryer(), clientID, cli)
}

func (s *ClientStore) loadAttributes(q db.QueryContext, clientID string, cli store.Client) error {
	attr, ok := cli.(store.Attributer)
	if ok {
		return s.extraAttributesStore.loadAttributes(q, clientID, attr)
	}

	return nil
}

// StoreAttributes store the extra attributes of cli if it is a store.Attributer
// and writes the resulting client in result.
func (s *ClientStore) StoreAttributes(clientID string, cli, result store.Client) error {
	return s.storeAttributes(s.queryer(), clientID, cli, result)
}

func (s *ClientStore) storeAttributes(q db.QueryContext, clientID string, cli, result store.Client) error {
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
