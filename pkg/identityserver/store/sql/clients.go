// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/db"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/satori/go.uuid"
)

type client struct {
	ID uuid.UUID
	ttnpb.Client
}

// ClientStore implements store.ClientStore.
type ClientStore struct {
	storer
	*extraAttributesStore
}

// NewClientStore retuens a ClientStore.
func NewClientStore(store storer) *ClientStore {
	return &ClientStore{
		storer:               store,
		extraAttributesStore: newExtraAttributesStore(store, "client"),
	}
}

func (s *ClientStore) getClientIdentifiersFromID(q db.QueryContext, id uuid.UUID) (res ttnpb.ClientIdentifiers, err error) {
	err = q.SelectOne(
		&res,
		`SELECT
				client_id
			FROM clients
			WHERE id = $1`,
		id)
	return
}

func (s *ClientStore) getClientID(q db.QueryContext, ids ttnpb.ClientIdentifiers) (res uuid.UUID, err error) {
	err = q.SelectOne(
		&res,
		`SELECT
				id
			FROM clients
			WHERE client_id = $1`,
		ids.ClientID)
	if db.IsNoRows(err) {
		err = ErrClientNotFound.New(nil)
	}
	return
}

// Create creates a client.
func (s *ClientStore) Create(client store.Client) error {
	err := s.transact(func(tx *db.Tx) error {
		cli := client.GetClient()

		userID, err := s.store().Users.(*UserStore).getUserID(tx, cli.CreatorIDs)
		if err != nil {
			return err
		}

		clientID, err := s.create(tx, userID, cli)
		if err != nil {
			return err
		}

		return s.storeAttributes(tx, clientID, client)
	})
	return err
}

func (s *ClientStore) create(q db.QueryContext, userID uuid.UUID, data *ttnpb.Client) (id uuid.UUID, err error) {
	var cli struct {
		*ttnpb.Client
		CreatorID       uuid.UUID
		GrantsConverted db.Int32Slice
		RightsConverted db.Int32Slice
	}
	cli.Client = data
	cli.CreatorID = userID

	cli.RightsConverted, err = db.NewInt32Slice(cli.Client.Rights)
	if err != nil {
		return
	}

	cli.GrantsConverted, err = db.NewInt32Slice(cli.Client.Grants)
	if err != nil {
		return
	}

	err = q.NamedSelectOne(
		&id,
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
				:official_labeled)
			RETURNING id`,
		cli)

	if _, yes := db.IsDuplicate(err); yes {
		err = ErrClientIDTaken.New(nil)
	}

	return
}

// GetByID finds a client by ID and retrieves it.
func (s *ClientStore) GetByID(ids ttnpb.ClientIdentifiers, specializer store.ClientSpecializer) (result store.Client, err error) {
	err = s.transact(func(tx *db.Tx) error {
		clientID, err := s.getClientID(tx, ids)
		if err != nil {
			return err
		}

		client, err := s.getByID(tx, clientID)
		if err != nil {
			return err
		}

		result = specializer(client)

		return s.loadAttributes(tx, clientID, result)
	})

	return
}

func (s *ClientStore) getByID(q db.QueryContext, id uuid.UUID) (client ttnpb.Client, err error) {
	var res struct {
		ttnpb.Client
		CreatorID       uuid.UUID
		GrantsConverted db.Int32Slice
		RightsConverted db.Int32Slice
	}

	err = q.SelectOne(
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
			WHERE id = $1`,
		id)

	if db.IsNoRows(err) {
		err = ErrClientNotFound.New(nil)
	}
	if err != nil {
		return
	}

	res.RightsConverted.SetInto(&res.Client.Rights)
	res.GrantsConverted.SetInto(&res.Client.Grants)
	res.Client.CreatorIDs, err = s.store().Users.(*UserStore).getUserIdentifiersFromID(q, res.CreatorID)
	client = res.Client

	return
}

// List returns all the clients.
func (s *ClientStore) List(specializer store.ClientSpecializer) ([]store.Client, error) {
	var res []store.Client
	err := s.transact(func(tx *db.Tx) error {
		found, err := s.list(tx)
		if err != nil {
			return err
		}

		res = make([]store.Client, 0, len(found))

		for _, client := range found {
			cli := specializer(client.Client)

			err := s.loadAttributes(tx, client.ID, cli)
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

func (s *ClientStore) list(q db.QueryContext) ([]client, error) {
	var res []struct {
		client
		CreatorID       uuid.UUID
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

	clients := make([]client, 0, len(res))
	for _, client := range res {
		client.RightsConverted.SetInto(&client.client.Client.Rights)
		client.GrantsConverted.SetInto(&client.client.Client.Grants)

		userID, err := s.store().Users.(*UserStore).getUserIdentifiersFromID(q, client.CreatorID)
		if err != nil {
			return nil, err
		}
		client.client.Client.CreatorIDs = userID

		clients = append(clients, client.client)
	}

	return clients, nil
}

// ListByUser returns all the clients created by the client.
func (s *ClientStore) ListByUser(ids ttnpb.UserIdentifiers, specializer store.ClientSpecializer) (result []store.Client, err error) {
	err = s.transact(func(tx *db.Tx) error {
		userID, err := s.store().Users.(*UserStore).getUserID(tx, ids)
		if err != nil {
			return err
		}

		clients, err := s.userClients(tx, userID)
		if err != nil {
			return err
		}

		for _, client := range clients {
			specialized := specializer(client.Client)

			err := s.loadAttributes(tx, client.ID, specialized)
			if err != nil {
				return err
			}

			result = append(result, specialized)
		}

		return nil
	})
	return
}

func (s *ClientStore) userClients(q db.QueryContext, userID uuid.UUID) ([]client, error) {
	var res []struct {
		client
		CreatorID       uuid.UUID
		GrantsConverted db.Int32Slice
		RightsConverted db.Int32Slice
	}
	err := q.Select(
		&res,
		`SELECT
				id,
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

	clients := make([]client, 0, len(res))
	for _, client := range res {
		client.RightsConverted.SetInto(&client.client.Client.Rights)
		client.GrantsConverted.SetInto(&client.client.Client.Grants)
		userID, err := s.store().Users.(*UserStore).getUserIdentifiersFromID(q, client.CreatorID)
		if err != nil {
			return nil, err
		}
		client.client.Client.CreatorIDs = userID

		clients = append(clients, client.client)
	}

	return clients, nil
}

// Update updates the client.
func (s *ClientStore) Update(client store.Client) error {
	err := s.transact(func(tx *db.Tx) error {
		cli := client.GetClient()

		clientID, err := s.getClientID(tx, cli.ClientIdentifiers)
		if err != nil {
			return err
		}

		err = s.update(tx, clientID, cli)
		if err != nil {
			return err
		}

		return s.storeAttributes(tx, clientID, client)
	})
	return err
}

func (s *ClientStore) update(q db.QueryContext, clientID uuid.UUID, data *ttnpb.Client) (err error) {
	var input struct {
		client
		CreatorID       uuid.UUID
		GrantsConverted db.Int32Slice
		RightsConverted db.Int32Slice
	}
	input.client = client{
		ID:     clientID,
		Client: *data,
	}

	input.CreatorID, err = s.store().Users.(*UserStore).getUserID(q, input.CreatorIDs)
	if err != nil {
		return err
	}

	rights, err := db.NewInt32Slice(input.Client.Rights)
	if err != nil {
		return err
	}
	input.RightsConverted = rights

	grants, err := db.NewInt32Slice(input.Client.Grants)
	if err != nil {
		return err
	}
	input.GrantsConverted = grants

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
			WHERE id = :id`,
		input)

	if db.IsNoRows(err) {
		err = ErrClientNotFound.New(nil)
	}

	return
}

// Delete deletes a client.
func (s *ClientStore) Delete(ids ttnpb.ClientIdentifiers) error {
	err := s.transact(func(tx *db.Tx) error {
		clientID, err := s.getClientID(tx, ids)
		if err != nil {
			return err
		}

		err = s.store().OAuth.(*OAuthStore).deleteAuthorizationCodesByClient(tx, clientID)
		if err != nil {
			return err
		}

		err = s.store().OAuth.(*OAuthStore).deleteAccessTokensByClient(tx, clientID)
		if err != nil {
			return err
		}

		err = s.store().OAuth.(*OAuthStore).deleteRefreshTokensByClient(tx, clientID)
		if err != nil {
			return err
		}

		return s.delete(tx, clientID)
	})

	return err
}

// delete deletes the client itself. All rows in other tables that references
// this entity must be deleted before this one gets deleted.
func (s *ClientStore) delete(q db.QueryContext, clientID uuid.UUID) (err error) {
	id := new(string)
	err = q.SelectOne(
		id,
		`DELETE
			FROM clients
			WHERE id = $1
			RETURNING client_id`,
		clientID)
	if db.IsNoRows(err) {
		err = ErrClientNotFound.New(nil)
	}
	return
}
