// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/util"
	"github.com/TheThingsNetwork/ttn/pkg/random"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
)

type clientService struct {
	*IdentityServer
}

// CreateClient creates a client.
// The created client has a random secret and has set by default as false the
// official labeled flag and has the refresh_token and authorization_code grants.
func (s *clientService) CreateClient(ctx context.Context, req *ttnpb.CreateClientRequest) (*pbtypes.Empty, error) {
	userID, err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_CLIENTS_CREATE)
	if err != nil {
		return nil, err
	}

	err = s.store.Transact(func(tx *store.Store) error {
		settings, err := tx.Settings.Get()
		if err != nil {
			return err
		}

		// check for blacklisted ids
		if !util.IsIDAllowed(req.Client.ClientID, settings.BlacklistedIDs) {
			return ErrBlacklistedID.New(errors.Attributes{
				"id": req.Client.ClientID,
			})
		}

		return tx.Clients.Create(&ttnpb.Client{
			ClientIdentifier: req.Client.ClientIdentifier,
			Description:      req.Client.Description,
			RedirectURI:      req.Client.RedirectURI,
			Creator:          ttnpb.UserIdentifier{userID},
			Secret:           random.String(64),
			State:            ttnpb.STATE_PENDING,
			OfficialLabeled:  false,
			Grants:           []ttnpb.GrantType{ttnpb.GRANT_AUTHORIZATION_CODE, ttnpb.GRANT_REFRESH_TOKEN},
			Rights:           req.Client.Rights,
		})
	})

	return nil, err
}

// GetClient returns a client.
func (s *clientService) GetClient(ctx context.Context, req *ttnpb.ClientIdentifier) (*ttnpb.Client, error) {
	userID, err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_CLIENTS_LIST)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Clients.GetByID(req.ClientID, s.config.Factories.Client)
	if err != nil {
		return nil, err
	}
	client := found.GetClient()

	// ensure the user is the client's creator
	if client.Creator.UserID != userID {
		return nil, ErrNotAuthorized.New(nil)
	}

	return client, err
}

// ListClients returns all the clients an user has created.
func (s *clientService) ListClients(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.ListClientsResponse, error) {
	userID, err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_CLIENTS_LIST)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Clients.ListByUser(userID, s.config.Factories.Client)
	if err != nil {
		return nil, err
	}

	resp := &ttnpb.ListClientsResponse{
		Clients: make([]*ttnpb.Client, 0, len(found)),
	}

	for _, cli := range found {
		resp.Clients = append(resp.Clients, cli.GetClient())
	}

	return resp, nil
}

// UpdateClient updates a client.
// TODO(gomezjdaniel): support to update the RedirectURI and rights (scope).
func (s *clientService) UpdateClient(ctx context.Context, req *ttnpb.UpdateClientRequest) (*pbtypes.Empty, error) {
	userID, err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_CLIENTS_MANAGE)
	if err != nil {
		return nil, err
	}

	err = s.store.Transact(func(tx *store.Store) error {
		found, err := tx.Clients.GetByID(req.Client.ClientID, s.config.Factories.Client)
		if err != nil {
			return err
		}
		client := found.GetClient()

		// ensure the user is the client's creator
		if client.Creator.UserID != userID {
			return ErrNotAuthorized.New(nil)
		}

		for _, path := range req.UpdateMask.Paths {
			switch {
			case ttnpb.FieldPathClientDescription.MatchString(path):
				client.Description = req.Client.Description
			default:
				return ttnpb.ErrInvalidPathUpdateMask.New(errors.Attributes{
					"path": path,
				})
			}
		}

		return tx.Clients.Update(client)
	})

	return nil, err
}

// DeleteClient deletes the client that matches the identifier and revokes all
// user authorizations.
func (s *clientService) DeleteClient(ctx context.Context, req *ttnpb.ClientIdentifier) (*pbtypes.Empty, error) {
	userID, err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_CLIENTS_MANAGE)
	if err != nil {
		return nil, err
	}

	err = s.store.Transact(func(tx *store.Store) error {
		found, err := tx.Clients.GetByID(req.ClientID, s.config.Factories.Client)
		if err != nil {
			return err
		}

		// ensure the user is the client's creator
		if found.GetClient().Creator.UserID != userID {
			return ErrNotAuthorized.New(nil)
		}

		return tx.Clients.Delete(req.ClientID)
	})

	return nil, err
}
