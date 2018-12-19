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
	"strconv"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/identityserver/blacklist"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	evtCreateClient = events.Define("client.create", "Create OAuth client")
	evtUpdateClient = events.Define("client.update", "Update OAuth client")
	evtDeleteClient = events.Define("client.delete", "Delete OAuth client")
)

func (is *IdentityServer) createClient(ctx context.Context, req *ttnpb.CreateClientRequest) (cli *ttnpb.Client, err error) {
	createdByAdmin := is.UniversalRights(ctx).IncludesAll(ttnpb.RIGHT_USER_ALL)

	if err := blacklist.Check(ctx, req.ClientID); err != nil {
		return nil, err
	}
	if usrIDs := req.Collaborator.GetUserIDs(); usrIDs != nil {
		if err = rights.RequireUser(ctx, *usrIDs, ttnpb.RIGHT_USER_CLIENTS_CREATE); err != nil {
			return nil, err
		}
	} else if orgIDs := req.Collaborator.GetOrganizationIDs(); orgIDs != nil {
		if err = rights.RequireOrganization(ctx, *orgIDs, ttnpb.RIGHT_ORGANIZATION_CLIENTS_CREATE); err != nil {
			return nil, err
		}
	}

	secret := req.Client.Secret
	if secret == "" {
		secret, err = auth.GenerateKey(ctx)
		if err != nil {
			return nil, err
		}
	}
	hashedSecret, err := auth.Hash(secret)
	if err != nil {
		return nil, err
	}
	req.Client.Secret = string(hashedSecret)

	if !createdByAdmin {
		req.Client.State = ttnpb.STATE_REQUESTED
		req.Client.SkipAuthorization = false
		req.Client.Endorsed = false
	}

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		cliStore := store.GetClientStore(db)
		cli, err = cliStore.CreateClient(ctx, &req.Client)
		if err != nil {
			return err
		}
		memberStore := store.GetMembershipStore(db)
		if err = memberStore.SetMember(ctx, &req.Collaborator, cli.ClientIdentifiers.EntityIdentifiers(), ttnpb.RightsFrom(ttnpb.RIGHT_CLIENT_ALL)); err != nil {
			return err
		}
		if len(req.ContactInfo) > 0 {
			cleanContactInfo(req.ContactInfo)
			cli.ContactInfo, err = store.GetContactInfoStore(db).SetContactInfo(ctx, cli.EntityIdentifiers(), req.ContactInfo)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	cli.Secret = secret // Return the unhashed secret, in case it was generated.

	events.Publish(evtCreateClient(ctx, req.ClientIdentifiers, nil))
	is.invalidateCachedMembershipsForAccount(ctx, &req.Collaborator)
	return cli, nil
}

func (is *IdentityServer) getClient(ctx context.Context, req *ttnpb.GetClientRequest) (cli *ttnpb.Client, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	if err = rights.RequireClient(ctx, req.ClientIdentifiers, ttnpb.RIGHT_CLIENT_ALL); err != nil {
		if hasOnlyAllowedFields(topLevelFields(req.FieldMask.Paths), ttnpb.PublicClientFields) {
			defer func() { cli = cli.PublicSafe() }()
		} else {
			return nil, err
		}
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		cliStore := store.GetClientStore(db)
		cli, err = cliStore.GetClient(ctx, &req.ClientIdentifiers, &req.FieldMask)
		if err != nil {
			return err
		}
		if hasField(req.FieldMask.Paths, "contact_info") {
			cli.ContactInfo, err = store.GetContactInfoStore(db).GetContactInfo(ctx, cli.EntityIdentifiers())
			if err != nil {
				return err
			}
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	return cli, nil
}

func (is *IdentityServer) listClients(ctx context.Context, req *ttnpb.ListClientsRequest) (clis *ttnpb.Clients, err error) {
	var cliRights map[string]*ttnpb.Rights
	if req.Collaborator == nil {
		callerRights, err := is.getRights(ctx)
		if err != nil {
			return nil, err
		}
		cliRights = make(map[string]*ttnpb.Rights, len(callerRights))
		for ids, rights := range callerRights {
			if ids := ids.GetClientIDs(); ids != nil {
				cliRights[unique.ID(ctx, ids)] = rights
			}
		}
		if len(cliRights) == 0 {
			return &ttnpb.Clients{}, nil
		}
	}
	if usrIDs := req.Collaborator.GetUserIDs(); usrIDs != nil {
		if err = rights.RequireUser(ctx, *usrIDs, ttnpb.RIGHT_USER_CLIENTS_LIST); err != nil {
			return nil, err
		}
	} else if orgIDs := req.Collaborator.GetOrganizationIDs(); orgIDs != nil {
		if err = rights.RequireOrganization(ctx, *orgIDs, ttnpb.RIGHT_ORGANIZATION_CLIENTS_LIST); err != nil {
			return nil, err
		}
	}
	var total uint64
	ctx = store.SetTotalCount(ctx, &total)
	defer func() {
		if err == nil {
			grpc.SetHeader(ctx, metadata.Pairs("x-total-count", strconv.FormatUint(total, 10)))
		}
	}()
	clis = &ttnpb.Clients{}
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		if cliRights == nil {
			memberStore := store.GetMembershipStore(db)
			rights, err := memberStore.FindMemberRights(ctx, req.Collaborator, "client")
			if err != nil {
				return err
			}
			cliRights = make(map[string]*ttnpb.Rights, len(rights))
			for ids, rights := range rights {
				cliRights[unique.ID(ctx, ids)] = rights
			}
		}
		if len(cliRights) == 0 {
			return nil
		}
		cliIDs := make([]*ttnpb.ClientIdentifiers, 0, len(cliRights))
		for uid := range cliRights {
			cliID, err := unique.ToClientID(uid)
			if err != nil {
				continue
			}
			cliIDs = append(cliIDs, &cliID)
		}
		cliStore := store.GetClientStore(db)
		clis.Clients, err = cliStore.FindClients(ctx, cliIDs, &req.FieldMask)
		if err != nil {
			return err
		}
		for _, cli := range clis.Clients {
			if !cliRights[unique.ID(ctx, cli.ClientIdentifiers)].IncludesAll(ttnpb.RIGHT_CLIENT_ALL) {
				cli = cli.PublicSafe()
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return clis, nil
}

var (
	errUpdateClientAdminField = errors.DefinePermissionDenied("client_update_admin_field", "only admins can update the `{field}` field")
)

func (is *IdentityServer) updateClient(ctx context.Context, req *ttnpb.UpdateClientRequest) (cli *ttnpb.Client, err error) {
	if err = rights.RequireClient(ctx, req.ClientIdentifiers, ttnpb.RIGHT_CLIENT_ALL); err != nil {
		return nil, err
	}
	updatedByAdmin := is.UniversalRights(ctx).IncludesAll(ttnpb.RIGHT_USER_ALL)

	if hasField(req.FieldMask.Paths, "grants") && !updatedByAdmin {
		return nil, errUpdateClientAdminField.WithAttributes("field", "grants")
	}

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		cliStore := store.GetClientStore(db)
		cli, err = cliStore.UpdateClient(ctx, &req.Client, &req.FieldMask)
		if err != nil {
			return err
		}
		if hasField(req.FieldMask.Paths, "contact_info") {
			cleanContactInfo(req.ContactInfo)
			cli.ContactInfo, err = store.GetContactInfoStore(db).SetContactInfo(ctx, cli.EntityIdentifiers(), req.ContactInfo)
			if err != nil {
				return err
			}
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtUpdateClient(ctx, req.ClientIdentifiers, req.FieldMask.Paths))
	return cli, nil
}

func (is *IdentityServer) deleteClient(ctx context.Context, ids *ttnpb.ClientIdentifiers) (*types.Empty, error) {
	if err := rights.RequireClient(ctx, *ids, ttnpb.RIGHT_CLIENT_ALL); err != nil {
		return nil, err
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		cliStore := store.GetClientStore(db)
		return cliStore.DeleteClient(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtDeleteClient(ctx, ids, nil))
	// TODO: Invalidate rights of members
	return ttnpb.Empty, nil
}

type clientRegistry struct {
	*IdentityServer
}

func (cr *clientRegistry) Create(ctx context.Context, req *ttnpb.CreateClientRequest) (*ttnpb.Client, error) {
	return cr.createClient(ctx, req)
}
func (cr *clientRegistry) Get(ctx context.Context, req *ttnpb.GetClientRequest) (*ttnpb.Client, error) {
	return cr.getClient(ctx, req)
}
func (cr *clientRegistry) List(ctx context.Context, req *ttnpb.ListClientsRequest) (*ttnpb.Clients, error) {
	return cr.listClients(ctx, req)
}
func (cr *clientRegistry) Update(ctx context.Context, req *ttnpb.UpdateClientRequest) (*ttnpb.Client, error) {
	return cr.updateClient(ctx, req)
}
func (cr *clientRegistry) Delete(ctx context.Context, req *ttnpb.ClientIdentifiers) (*types.Empty, error) {
	return cr.deleteClient(ctx, req)
}
