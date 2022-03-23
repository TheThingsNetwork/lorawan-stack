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
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/blocklist"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var (
	evtCreateClient = events.Define(
		"client.create", "create OAuth client",
		events.WithVisibility(ttnpb.Right_RIGHT_CLIENT_ALL),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUpdateClient = events.Define(
		"client.update", "update OAuth client",
		events.WithVisibility(ttnpb.Right_RIGHT_CLIENT_ALL),
		events.WithUpdatedFieldsDataType(),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteClient = events.Define(
		"client.delete", "delete OAuth client",
		events.WithVisibility(ttnpb.Right_RIGHT_CLIENT_ALL),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtRestoreClient = events.Define(
		"client.restore", "restore OAuth client",
		events.WithVisibility(ttnpb.Right_RIGHT_CLIENT_ALL),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtPurgeClient = events.Define(
		"client.purge", "purge client",
		events.WithVisibility(ttnpb.Right_RIGHT_CLIENT_ALL),
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
	if err = blocklist.Check(ctx, req.Client.GetIds().GetClientId()); err != nil {
		return nil, err
	}
	if usrIDs := req.GetCollaborator().GetUserIds(); usrIDs != nil {
		if !createdByAdmin && !is.configFromContext(ctx).UserRights.CreateClients {
			return nil, errAdminsCreateClients.New()
		}
		if err = rights.RequireUser(ctx, usrIDs, ttnpb.Right_RIGHT_USER_CLIENTS_CREATE); err != nil {
			return nil, err
		}
	} else if orgIDs := req.GetCollaborator().GetOrganizationIds(); orgIDs != nil {
		if err = rights.RequireOrganization(ctx, orgIDs, ttnpb.Right_RIGHT_ORGANIZATION_CLIENTS_CREATE); err != nil {
			return nil, err
		}
	}

	if req.Client.AdministrativeContact == nil {
		req.Client.AdministrativeContact = req.Collaborator
	} else if err := validateCollaboratorEqualsContact(req.Collaborator, req.Client.AdministrativeContact); err != nil {
		return nil, err
	}
	if req.Client.TechnicalContact == nil {
		req.Client.TechnicalContact = req.Collaborator
	} else if err := validateCollaboratorEqualsContact(req.Collaborator, req.Client.TechnicalContact); err != nil {
		return nil, err
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
		req.Client.State = ttnpb.State_STATE_REQUESTED
		req.Client.StateDescription = "admin approval required"
		req.Client.SkipAuthorization = false
		req.Client.Endorsed = false
	}

	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		cli, err = st.CreateClient(ctx, req.Client)
		if err != nil {
			return err
		}
		if err = st.SetMember(
			ctx,
			req.GetCollaborator(),
			cli.GetIds().GetEntityIdentifiers(),
			ttnpb.RightsFrom(ttnpb.Right_RIGHT_ALL),
		); err != nil {
			return err
		}
		if len(req.Client.ContactInfo) > 0 {
			cleanContactInfo(req.Client.ContactInfo)
			cli.ContactInfo, err = st.SetContactInfo(ctx, cli.GetIds(), req.Client.ContactInfo)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if cli.State == ttnpb.State_STATE_REQUESTED {
		go is.notifyAdminsInternal(ctx, &ttnpb.CreateNotificationRequest{
			EntityIds:        req.GetClient().GetIds().GetEntityIdentifiers(),
			NotificationType: "client_requested",
			Data:             ttnpb.MustMarshalAny(req),
			Email:            true,
		})
	}

	cli.Secret = secret // Return the unhashed secret, in case it was generated.

	events.Publish(evtCreateClient.NewWithIdentifiersAndData(ctx, req.Client.GetIds(), nil))
	return cli, nil
}

func (is *IdentityServer) getClient(ctx context.Context, req *ttnpb.GetClientRequest) (cli *ttnpb.Client, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.ClientFieldPathsNested, req.FieldMask, getPaths, nil)
	if err = rights.RequireClient(ctx, req.GetClientIds(), ttnpb.Right_RIGHT_CLIENT_ALL); err != nil {
		if !ttnpb.HasOnlyAllowedFields(req.FieldMask.GetPaths(), ttnpb.PublicClientFields...) {
			return nil, err
		}
		defer func() { cli = cli.PublicSafe() }()
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		cli, err = st.GetClient(ctx, req.GetClientIds(), req.FieldMask.GetPaths())
		if err != nil {
			return err
		}
		if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info") {
			cli.ContactInfo, err = st.GetContactInfo(ctx, cli.GetIds())
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

	authInfo, err := is.authInfo(ctx)
	if err != nil {
		return nil, err
	}
	callerAccountID := authInfo.GetOrganizationOrUserIdentifiers()
	var includeIndirect bool
	if req.Collaborator == nil {
		req.Collaborator = callerAccountID
		includeIndirect = true
	}
	if req.Collaborator == nil {
		return &ttnpb.Clients{}, nil
	}

	if usrIDs := req.Collaborator.GetUserIds(); usrIDs != nil {
		if err = rights.RequireUser(ctx, usrIDs, ttnpb.Right_RIGHT_USER_CLIENTS_LIST); err != nil {
			return nil, err
		}
	} else if orgIDs := req.Collaborator.GetOrganizationIds(); orgIDs != nil {
		if err = rights.RequireOrganization(ctx, orgIDs, ttnpb.Right_RIGHT_ORGANIZATION_CLIENTS_LIST); err != nil {
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
	var callerMemberships store.MembershipChains

	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		ids, err := st.FindMemberships(paginateCtx, req.Collaborator, "client", includeIndirect)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return nil
		}
		callerMemberships, err = st.FindAccountMembershipChains(ctx, callerAccountID, "client", idStrings(ids...)...)
		if err != nil {
			return err
		}
		cliIDs := make([]*ttnpb.ClientIdentifiers, 0, len(ids))
		for _, id := range ids {
			if cliID := id.GetEntityIdentifiers().GetClientIds(); cliID != nil {
				cliIDs = append(cliIDs, cliID)
			}
		}
		clis.Clients, err = st.FindClients(ctx, cliIDs, req.FieldMask.GetPaths())
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	for i, cli := range clis.Clients {
		entityRights := callerMemberships.GetRights(callerAccountID, cli.GetIds()).Union(authInfo.GetUniversalRights())
		if !entityRights.IncludesAll(ttnpb.Right_RIGHT_CLIENT_ALL) {
			clis.Clients[i] = cli.PublicSafe()
		}
	}

	return clis, nil
}

var errUpdateClientAdminField = errors.DefinePermissionDenied("client_update_admin_field", "only admins can update the `{field}` field")

func (is *IdentityServer) updateClient(ctx context.Context, req *ttnpb.UpdateClientRequest) (cli *ttnpb.Client, err error) {
	if err = rights.RequireClient(ctx, req.Client.GetIds(), ttnpb.Right_RIGHT_CLIENT_ALL); err != nil {
		return nil, err
	}
	req.FieldMask = cleanFieldMaskPaths(ttnpb.ClientFieldPathsNested, req.FieldMask, nil, getPaths)
	if len(req.FieldMask.GetPaths()) == 0 {
		req.FieldMask = ttnpb.FieldMask(updatePaths...)
	}
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info") {
		if err := validateContactInfo(req.Client.ContactInfo); err != nil {
			return nil, err
		}
	}
	req.FieldMask.Paths = ttnpb.FlattenPaths(req.FieldMask.Paths, []string{"administrative_contact", "technical_contact"})
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
			req.Client.StateDescription = ""
		}
	}

	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		if err := validateContactIsCollaborator(ctx, st, req.Client.AdministrativeContact, req.Client.GetEntityIdentifiers()); err != nil {
			return err
		}
		if err := validateContactIsCollaborator(ctx, st, req.Client.TechnicalContact, req.Client.GetEntityIdentifiers()); err != nil {
			return err
		}
		cli, err = st.UpdateClient(ctx, req.Client, req.FieldMask.GetPaths())
		if err != nil {
			return err
		}
		if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "contact_info") {
			cleanContactInfo(req.Client.ContactInfo)
			cli.ContactInfo, err = st.SetContactInfo(ctx, cli.Ids, req.Client.ContactInfo)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtUpdateClient.NewWithIdentifiersAndData(ctx, req.Client.GetIds(), req.FieldMask.GetPaths()))
	if ttnpb.HasAnyField(req.FieldMask.GetPaths(), "state") {
		go is.notifyInternal(ctx, &ttnpb.CreateNotificationRequest{
			EntityIds:        cli.GetIds().GetEntityIdentifiers(),
			NotificationType: "entity_state_changed",
			Data: ttnpb.MustMarshalAny(&ttnpb.EntityStateChangedNotification{
				State:            cli.State,
				StateDescription: cli.StateDescription,
			}),
			Receivers: []ttnpb.NotificationReceiver{ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_ADMINISTRATIVE_CONTACT},
			Email:     true,
		})
	}
	return cli, nil
}

func (is *IdentityServer) deleteClient(ctx context.Context, ids *ttnpb.ClientIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireClient(ctx, ids, ttnpb.Right_RIGHT_CLIENT_ALL); err != nil {
		return nil, err
	}
	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		return st.DeleteClient(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtDeleteClient.NewWithIdentifiersAndData(ctx, ids, nil))
	return ttnpb.Empty, nil
}

func (is *IdentityServer) restoreClient(ctx context.Context, ids *ttnpb.ClientIdentifiers) (*pbtypes.Empty, error) {
	if err := rights.RequireClient(store.WithSoftDeleted(ctx, false), ids, ttnpb.Right_RIGHT_CLIENT_ALL); err != nil {
		return nil, err
	}
	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		cli, err := st.GetClient(store.WithSoftDeleted(ctx, true), ids, softDeleteFieldMask)
		if err != nil {
			return err
		}
		deletedAt := ttnpb.StdTime(cli.DeletedAt)
		if deletedAt == nil {
			panic("store.WithSoftDeleted(ctx, true) returned result that is not deleted")
		}
		if time.Since(*deletedAt) > is.configFromContext(ctx).Delete.Restore {
			return errRestoreWindowExpired.New()
		}
		return st.RestoreClient(ctx, ids)
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtRestoreClient.NewWithIdentifiersAndData(ctx, ids, nil))
	return ttnpb.Empty, nil
}

func (is *IdentityServer) purgeClient(ctx context.Context, ids *ttnpb.ClientIdentifiers) (*pbtypes.Empty, error) {
	if !is.IsAdmin(ctx) {
		return nil, errAdminsPurgeClients.New()
	}
	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		// delete related authorizations before purging the client
		err := st.DeleteClientAuthorizations(ctx, ids)
		if err != nil {
			return err
		}
		// delete related memberships before purging the client
		err = st.DeleteEntityMembers(ctx, ids.GetEntityIdentifiers())
		if err != nil {
			return err
		}
		// delete related contact info before purging the client
		err = st.DeleteEntityContactInfo(ctx, ids)
		if err != nil {
			return err
		}
		return st.PurgeClient(ctx, ids)
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
