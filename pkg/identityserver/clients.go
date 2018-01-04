// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/util"
	"github.com/TheThingsNetwork/ttn/pkg/random"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
)

var _ ttnpb.IsClientServer = new(IdentityServer)

// CreateClient creates a client.
// The created client has a random secret and has set by default as false the
// official labeled flag and has the refresh_token and authorization_code grants.
func (is *IdentityServer) CreateClient(ctx context.Context, req *ttnpb.CreateClientRequest) (*pbtypes.Empty, error) {
	userID, err := is.userCheck(ctx, ttnpb.RIGHT_USER_CLIENTS_CREATE)
	if err != nil {
		return nil, err
	}

	settings, err := is.store.Settings.Get()
	if err != nil {
		return nil, err
	}

	// check for blacklisted ids
	if !util.IsIDAllowed(req.Client.ClientID, settings.BlacklistedIDs) {
		return nil, ErrBlacklistedID.New(errors.Attributes{
			"id": req.Client.ClientID,
		})
	}

	return nil, is.store.Clients.Create(&ttnpb.Client{
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
}

// GetClients returns a client.
func (is *IdentityServer) GetClient(ctx context.Context, req *ttnpb.ClientIdentifier) (*ttnpb.Client, error) {
	userID, err := is.userCheck(ctx, ttnpb.RIGHT_USER_CLIENTS_LIST)
	if err != nil {
		return nil, err
	}

	found, err := is.store.Clients.GetByID(req.ClientID, is.factories.client)
	if err != nil {
		return nil, err
	}

	// ensure the user is the client's creator
	if found.GetClient().Creator.UserID != userID {
		return nil, ErrNotAuthorized
	}

	return found.GetClient(), err
}

// ListClients returns all the clients an user has created.
func (is *IdentityServer) ListClients(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.ListClientsResponse, error) {
	userID, err := is.userCheck(ctx, ttnpb.RIGHT_USER_CLIENTS_LIST)
	if err != nil {
		return nil, err
	}

	found, err := is.store.Clients.ListByUser(userID, is.factories.client)
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
func (is *IdentityServer) UpdateClient(ctx context.Context, req *ttnpb.UpdateClientRequest) (*pbtypes.Empty, error) {
	userID, err := is.userCheck(ctx, ttnpb.RIGHT_USER_CLIENTS_MANAGE)
	if err != nil {
		return nil, err
	}

	found, err := is.store.Clients.GetByID(req.Client.ClientID, is.factories.client)
	if err != nil {
		return nil, err
	}

	// ensure the user is the client's creator
	if found.GetClient().Creator.UserID != userID {
		return nil, ErrNotAuthorized
	}

	for _, path := range req.UpdateMask.Paths {
		switch true {
		case ttnpb.FieldPathClientDescription.MatchString(path):
			found.GetClient().Description = req.Client.Description
		default:
			return nil, ttnpb.ErrInvalidPathUpdateMask.New(errors.Attributes{
				"path": path,
			})
		}
	}

	return nil, is.store.Clients.Update(found)
}

// DeleteClient deletes a client.
func (is *IdentityServer) DeleteClient(ctx context.Context, req *ttnpb.ClientIdentifier) (*pbtypes.Empty, error) {
	userID, err := is.userCheck(ctx, ttnpb.RIGHT_USER_CLIENTS_MANAGE)
	if err != nil {
		return nil, err
	}

	found, err := is.store.Clients.GetByID(req.ClientID, is.factories.client)
	if err != nil {
		return nil, err
	}

	// ensure the user is the client's creator
	if found.GetClient().Creator.UserID != userID {
		return nil, ErrNotAuthorized
	}

	return nil, is.store.Clients.Delete(req.ClientID)
}
