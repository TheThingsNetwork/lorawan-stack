// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package identityserver

import (
	"context"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/errors/common"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/random"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// ClientGeneratedFields are the fields that are automatically generated.
var ClientGeneratedFields = []string{
	"CreatedAt",
	"UpdatedAt",
	"Grants",
	"Secret",
	"Client.CreatedAt",
	"Client.UpdatedAt",
	"Client.Grants",
	"Client.Secret",
}

type clientService struct {
	*IdentityServer
}

// CreateClient creates a client. The created client has a random secret and the
// refresh_token and authorization_code grants.
func (s *clientService) CreateClient(ctx context.Context, req *ttnpb.CreateClientRequest) (*pbtypes.Empty, error) {
	err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_CLIENTS)
	if err != nil {
		return nil, err
	}

	err = s.store.Transact(func(tx *store.Store) (err error) {
		settings, err := tx.Settings.Get()
		if err != nil {
			return err
		}

		// Check for blacklisted IDs.
		if !settings.IsIDAllowed(req.Client.ClientID) {
			return ErrBlacklistedID.New(errors.Attributes{
				"id": req.Client.ClientID,
			})
		}

		now := time.Now().UTC()

		return tx.Clients.Create(&ttnpb.Client{
			ClientIdentifiers: req.Client.ClientIdentifiers,
			Description:       req.Client.Description,
			RedirectURI:       req.Client.RedirectURI,
			CreatorIDs:        authorizationDataFromContext(ctx).UserIdentifiers(),
			Secret:            random.String(64),
			State:             ttnpb.STATE_PENDING,
			SkipAuthorization: false,
			Grants:            []ttnpb.GrantType{ttnpb.GRANT_AUTHORIZATION_CODE, ttnpb.GRANT_REFRESH_TOKEN},
			Rights:            req.Client.Rights,
			CreatedAt:         now,
			UpdatedAt:         now,
		})
	})

	return ttnpb.Empty, err
}

// GetClient returns the client that matches the identifier.
// It allows to be called without authorization credentials, in this case it
// will only return the publicly information available about the client.
func (s *clientService) GetClient(ctx context.Context, req *ttnpb.ClientIdentifiers) (*ttnpb.Client, error) {
	found, err := s.store.Clients.GetByID(*req, s.specializers.Client)
	if err != nil {
		return nil, err
	}
	client := found.GetClient()

	err = s.enforceUserRights(ctx, ttnpb.RIGHT_USER_CLIENTS)
	if err != nil && common.ErrPermissionDenied.Describes(err) {
		return &ttnpb.Client{
			ClientIdentifiers: client.ClientIdentifiers,
			Description:       client.Description,
			RedirectURI:       client.RedirectURI,
			SkipAuthorization: client.SkipAuthorization,
			Rights:            client.Rights,
		}, nil
	}
	if err != nil {
		return nil, err
	}

	// Ensure the user is the client's creator.
	if !client.CreatorIDs.Equals(authorizationDataFromContext(ctx).UserIdentifiers()) {
		return nil, common.ErrPermissionDenied.New(nil)
	}

	return client, err
}

// ListClients returns all the clients an user has created.
func (s *clientService) ListClients(ctx context.Context, _ *pbtypes.Empty) (*ttnpb.ListClientsResponse, error) {
	err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_CLIENTS)
	if err != nil {
		return nil, err
	}

	found, err := s.store.Clients.ListByUser(authorizationDataFromContext(ctx).UserIdentifiers(), s.specializers.Client)
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
	err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_CLIENTS)
	if err != nil {
		return nil, err
	}

	err = s.store.Transact(func(tx *store.Store) error {
		found, err := tx.Clients.GetByID(req.Client.ClientIdentifiers, s.specializers.Client)
		if err != nil {
			return err
		}
		client := found.GetClient()

		// Ensure the user is the client's creator.
		if !client.CreatorIDs.Equals(authorizationDataFromContext(ctx).UserIdentifiers()) {
			return common.ErrPermissionDenied.New(nil)
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

		client.UpdatedAt = time.Now().UTC()

		return tx.Clients.Update(client)
	})

	return ttnpb.Empty, err
}

// DeleteClient deletes the client that matches the identifier and revokes all
// user authorizations.
func (s *clientService) DeleteClient(ctx context.Context, req *ttnpb.ClientIdentifiers) (*pbtypes.Empty, error) {
	err := s.enforceUserRights(ctx, ttnpb.RIGHT_USER_CLIENTS)
	if err != nil {
		return nil, err
	}

	err = s.store.Transact(func(tx *store.Store) error {
		ids := *req

		found, err := tx.Clients.GetByID(ids, s.specializers.Client)
		if err != nil {
			return err
		}

		// Ensure the user is the client's creator.
		if !found.GetClient().CreatorIDs.Equals(authorizationDataFromContext(ctx).UserIdentifiers()) {
			return common.ErrPermissionDenied.New(nil)
		}

		return tx.Clients.Delete(ids)
	})

	return ttnpb.Empty, err
}
