// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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
	"strings"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/email"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/blacklist"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/emails"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	evtCreateClient = events.Define(
		"client.create", "create OAuth client",
		events.WithVisibility(ttnpb.RIGHT_CLIENT_ALL),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUpdateClient = events.Define(
		"client.update", "update OAuth client",
		events.WithVisibility(ttnpb.RIGHT_CLIENT_ALL),
		events.WithUpdatedFieldsDataType(),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteClient = events.Define(
		"client.delete", "delete OAuth client",
		events.WithVisibility(ttnpb.RIGHT_CLIENT_ALL),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtRestoreClient = events.Define(
		"client.restore", "restore OAuth client",
		events.WithVisibility(ttnpb.RIGHT_CLIENT_ALL),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtPurgeClient = events.Define(
		"client.purge", "purge client",
		events.WithVisibility(ttnpb.RIGHT_CLIENT_ALL),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
)

var (
	errAdminsCreateClients = errors.DefinePermissionDenied("admins_create_clients", "OAuth clients may only be created by admins, or in organizations")
	errAdminsPurgeClients  = errors.DefinePermissionDenied("admins_purge_clients", "OAuth clients may only be purged by admins")
)

func (is *IdentityServer) createClient(ctx context.Context, req *ttnpb.CreateClientRequest) (cli *ttnpb.Client, err error) {
	createdByAdmin := is.IsAdmin(ctx)
	if err = blacklist.Check(ctx, req.ClientId); err != nil {
		return nil, err
	}
	if usrIDs := req.Collaborator.GetUserIds(); usrIDs != nil {
		if !createdByAdmin && !is.configFromContext(ctx).UserRights.CreateClients {
			return nil, errAdminsCreateClients
		}
		if err = rights.RequireUser(ctx, *usrIDs, ttnpb.RIGHT_USER_CLIENTS_CREATE); err != nil {
			return nil, err
		}
	} else if orgIDs := req.Collaborator.GetOrganizationIds(); orgIDs != nil {
		if err = rights.RequireOrganization(ctx, *orgIDs, ttnpb.RIGHT_ORGANIZATION_CLIENTS_CREATE); err != nil {
			return nil, err
		}
	}

	if err := validateContactInfo(req.Client.ContactInfo); err != nil {
		return nil, err
	}

	secret := req.Client.Secret
	if secret == "" {
		secret, err = auth.GenerateKey(ctx)
		if err != nil {
			return nil, err
		}
	}
	hashedSecret, err := auth.Hash(ctx, secret)
	if err != nil {
		return nil, err
	}
	req.Client.Secret = hashedSecret

	if !createdByAdmin {
		req.Client.State = ttnpb.STATE_REQUESTED
		req.Client.StateDescription = "admin approval required"
		req.Client.SkipAuthorization = false
		req.Client.Endorsed = false
	}

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		cli, err = store.GetClientStore(db).CreateClient(ctx, &req.Client)
		if err != nil {
			return err
		}
		if err = is.getMembershipStore(ctx, db).SetMember(
			ctx,
			&req.Collaborator,
			cli.ClientIdentifiers.GetEntityIdentifiers(),
			ttnpb.RightsFrom(ttnpb.RIGHT_ALL),
		); err != nil {
			return err
		}
		if len(req.ContactInfo) > 0 {
			cleanContactInfo(req.ContactInfo)
			cli.ContactInfo, err = store.GetContactInfoStore(db).SetContactInfo(ctx, cli.ClientIdentifiers, req.ContactInfo)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if cli.State == ttnpb.STATE_REQUESTED {
		err = is.SendAdminsEmail(ctx, func(data emails.Data) email.MessageData {
			data.Entity.Type, data.Entity.ID = "client", cli.ClientId
			return &emails.ClientRequested{
				Data:         data,
				Client:       cli,
				Collaborator: &req.Collaborator,
			}
		})
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Could not send client requested email")
		}
	}

	cli.Secret = secret // Return the unhashed secret, in case it was generated.

	events.Publish(evtCreateClient.NewWithIdentifiersAndData(ctx, &req.ClientIdentifiers, nil))
	return cli, nil
}

func (is *IdentityServer) getClient(ctx context.Context, req *ttnpb.GetClientRequest) (cli *ttnpb.Client, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.ClientFieldPathsNested, req.FieldMask, getPaths, nil)
	if err = rights.RequireClient(ctx, req.ClientIdentifiers, ttnpb.RIGHT_CLIENT_ALL); err != nil {
		if ttnpb.HasOnlyAllowedFields(req.FieldMask.GetPaths(), ttnpb.PublicClientFields...) {
			defer func() { cli = cli.PublicSafe() }()
		} else {
			return nil, err
		}
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		cli, err = store.GetClientStore(db).GetClient(ctx, &req.ClientIdentifiers, req.FieldMask)
		if err != nil {
			return err
		}
		if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info") {
			cli.ContactInfo, err = store.GetContactInfoStore(db).GetContactInfo(ctx, cli.ClientIdentifiers)
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
	req.FieldMask = cleanFieldMaskPaths(ttnpb.ClientFieldPathsNested, req.FieldMask, getPaths, nil)
	var includeIndirect bool
	if req.Collaborator == nil {
		authInfo, err := is.authInfo(ctx)
		if err != nil {
			return nil, err
		}
		collaborator := authInfo.GetOrganizationOrUserIdentifiers()
		if collaborator == nil {
			return &ttnpb.Clients{}, nil
		}
		req.Collaborator = collaborator
		includeIndirect = true
	}
	if usrIDs := req.Collaborator.GetUserIds(); usrIDs != nil {
		if err = rights.RequireUser(ctx, *usrIDs, ttnpb.RIGHT_USER_CLIENTS_LIST); err != nil {
			return nil, err
		}
	} else if orgIDs := req.Collaborator.GetOrganizationIds(); orgIDs != nil {
		if err = rights.RequireOrganization(ctx, *orgIDs, ttnpb.RIGHT_ORGANIZATION_CLIENTS_LIST); err != nil {
			return nil, err
		}
	}
	if req.Deleted {
		ctx = store.WithSoftDeleted(ctx, true)
	}
	ctx = store.WithOrder(ctx, req.Order)
	var total uint64
	paginateCtx := store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	clis = &ttnpb.Clients{}
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		ids, err := is.getMembershipStore(ctx, db).FindMemberships(paginateCtx, req.Collaborator, "client", includeIndirect)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		cliIDs := make([]*ttnpb.ClientIdentifiers, 0, len(ids))
		for _, id := range ids {
			if cliID := id.GetEntityIdentifiers().GetClientIds(); cliID != nil {
				cliIDs = append(cliIDs, cliID)
			}
		}
		clis.Clients, err = store.GetClientStore(db).FindClients(ctx, cliIDs, req.FieldMask)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	for i, cli := range clis.Clients {
		if rights.RequireClient(ctx, cli.ClientIdentifiers, ttnpb.RIGHT_CLIENT_ALL) != nil {
			clis.Clients[i] = cli.PublicSafe()
		}
	}

	return clis, nil
}

var errUpdateClientAdminField = errors.DefinePermissionDenied("client_update_admin_field", "only admins can update the `{field}` field")

func (is *IdentityServer) updateClient(ctx context.Context, req *ttnpb.UpdateClientRequest) (cli *ttnpb.Client, err error) {
	if err = rights.RequireClient(ctx, req.ClientIdentifiers, ttnpb.RIGHT_CLIENT_ALL); err != nil {
		return nil, err
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.ClientFieldPathsNested, req.FieldMask, nil, getPaths)
	if len(req.FieldMask.GetPaths()) == 0 {
		req.FieldMask = &pbtypes.FieldMask{Paths: updatePaths}
	}
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info") {
		if err := validateContactInfo(req.Client.ContactInfo); err != nil {
			return nil, err
		}
	}
	updatedByAdmin := is.IsAdmin(ctx)

	if !updatedByAdmin {
		for _, path := range req.FieldMask.Paths {
			switch path {
			case "state", "state_description", "skip_authorization", "endorsed", "grants":
				return nil, errUpdateUserAdminField.WithAttributes("field", path)
			}
		}
	}

	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "state") {
		if !ttnpb.HasAnyField(req.FieldMask.GetPaths(), "state_description") {
			req.FieldMask.Paths = append(req.FieldMask.GetPaths(), "state_description")
			req.StateDescription = ""
		}
	}

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		cli, err = store.GetClientStore(db).UpdateClient(ctx, &req.Client, req.FieldMask)
		if err != nil {
			return err
		}
		if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info") {
			cleanContactInfo(req.ContactInfo)
			cli.ContactInfo, err = store.GetContactInfoStore(db).SetContactInfo(ctx, cli.ClientIdentifiers, req.ContactInfo)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtUpdateClient.NewWithIdentifiersAndData(ctx, &req.ClientIdentifiers, req.FieldMask.GetPaths()))
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "state") {
		err = is.SendContactsEmail(ctx, req, func(data emails.Data) email.MessageData {
			data.SetEntity(req)
			return &emails.EntityStateChanged{
				Data:             data,
				State:            strings.ToLower(strings.TrimPrefix(cli.State.String(), "STATE_")),
				StateDescription: cli.StateDescription,
			}
		})
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Could not send state change notification email")
		}
	}
	return cli, nil
}

func (is *IdentityServer) deleteClient(ctx context.Context, ids *ttnpb.ClientIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireClient(ctx, *ids, ttnpb.RIGHT_CLIENT_ALL); err != nil {
		return nil, err
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		return store.GetClientStore(db).DeleteClient(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtDeleteClient.NewWithIdentifiersAndData(ctx, ids, nil))
	return ttnpb.Empty, nil
}

func (is *IdentityServer) restoreClient(ctx context.Context, ids *ttnpb.ClientIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireClient(store.WithSoftDeleted(ctx, false), *ids, ttnpb.RIGHT_CLIENT_ALL); err != nil {
		return nil, err
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		cliStore := store.GetClientStore(db)
		cli, err := cliStore.GetClient(store.WithSoftDeleted(ctx, true), ids, softDeleteFieldMask)
		if err != nil {
			return err
		}
		if cli.DeletedAt == nil {
			panic("store.WithSoftDeleted(ctx, true) returned result that is not deleted")
		}
		if time.Since(*cli.DeletedAt) > is.configFromContext(ctx).Delete.Restore {
			return errRestoreWindowExpired.New()
		}
		return cliStore.RestoreClient(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtRestoreClient.NewWithIdentifiersAndData(ctx, ids, nil))
	return ttnpb.Empty, nil
}

func (is *IdentityServer) purgeClient(ctx context.Context, ids *ttnpb.ClientIdentifiers) (*pbtypes.Empty, error) {
	if !is.IsAdmin(ctx) {
		return nil, errAdminsPurgeClients
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		// delete related authorizations before purging the client
		err := store.GetOAuthStore(db).DeleteClientAuthorizations(ctx, ids)
		if err != nil {
			return err
		}
		// delete related memberships before purging the client
		err = store.GetMembershipStore(db).DeleteEntityMembers(ctx, ids.GetEntityIdentifiers())
		if err != nil {
			return err
		}
		// delete related contact info before purging the client
		err = store.GetContactInfoStore(db).DeleteEntityContactInfo(ctx, ids)
		if err != nil {
			return err
		}
		return store.GetClientStore(db).PurgeClient(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtPurgeClient.NewWithIdentifiersAndData(ctx, ids, nil))
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

func (cr *clientRegistry) Delete(ctx context.Context, req *ttnpb.ClientIdentifiers) (*pbtypes.Empty, error) {
	return cr.deleteClient(ctx, req)
}

func (cr *clientRegistry) Purge(ctx context.Context, req *ttnpb.ClientIdentifiers) (*pbtypes.Empty, error) {
	return cr.purgeClient(ctx, req)
}

func (cr *clientRegistry) Restore(ctx context.Context, req *ttnpb.ClientIdentifiers) (*pbtypes.Empty, error) {
	return cr.restoreClient(ctx, req)
}
