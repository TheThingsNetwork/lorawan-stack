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

	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	evtUpdateClientCollaborator = events.Define(
		"client.collaborator.update", "update client collaborator",
		events.WithVisibility(
			ttnpb.Right_RIGHT_CLIENT_SETTINGS_COLLABORATORS,
			ttnpb.Right_RIGHT_USER_CLIENTS_LIST,
		),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteClientCollaborator = events.Define(
		"client.collaborator.delete", "delete client collaborator",
		events.WithVisibility(
			ttnpb.Right_RIGHT_CLIENT_SETTINGS_COLLABORATORS,
			ttnpb.Right_RIGHT_USER_CLIENTS_LIST,
		),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
)

func (*IdentityServer) listClientRights(ctx context.Context, ids *ttnpb.ClientIdentifiers) (*ttnpb.Rights, error) {
	cliRights, err := rights.ListClient(ctx, ids)
	if err != nil {
		return nil, err
	}
	return cliRights.Intersect(ttnpb.AllClientRights), nil
}

func (is *IdentityServer) getClientCollaborator(
	ctx context.Context,
	req *ttnpb.GetClientCollaboratorRequest,
) (*ttnpb.GetCollaboratorResponse, error) {
	err := rights.RequireClient(ctx, req.GetClientIds(), ttnpb.Right_RIGHT_CLIENT_SETTINGS_COLLABORATORS)
	if err != nil {
		return nil, err
	}
	res := &ttnpb.GetCollaboratorResponse{
		Ids: req.GetCollaborator(),
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		rights, err := st.GetMember(
			ctx,
			req.GetCollaborator(),
			req.GetClientIds().GetEntityIdentifiers(),
		)
		if err != nil {
			return err
		}
		res.Rights = rights.GetRights()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

var errClientNeedsCollaborator = errors.DefineFailedPrecondition(
	"client_needs_collaborator",
	"every client needs at least one collaborator with all rights",
)

func (is *IdentityServer) setClientCollaborator(
	ctx context.Context, req *ttnpb.SetClientCollaboratorRequest,
) (_ *emptypb.Empty, err error) {
	// Require that caller has rights to manage collaborators.
	err = rights.RequireClient(ctx, req.GetClientIds(), ttnpb.Right_RIGHT_CLIENT_SETTINGS_COLLABORATORS)
	if err != nil {
		return nil, err
	}

	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		existingRights, err := st.GetMember(
			ctx,
			req.GetCollaborator().GetIds(),
			req.GetClientIds().GetEntityIdentifiers(),
		)
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
		existingRights = existingRights.Implied()
		newRights := ttnpb.RightsFrom(req.Collaborator.Rights...).Implied()
		addedRights := newRights.Sub(existingRights)
		removedRights := existingRights.Sub(newRights)

		// Require the caller to have all added rights.
		if len(addedRights.GetRights()) > 0 {
			if err := rights.RequireClient(ctx, req.GetClientIds(), addedRights.GetRights()...); err != nil {
				return err
			}
		}

		// Unless we're deleting the collaborator, require the caller to have all removed rights.
		if len(newRights.GetRights()) > 0 && len(removedRights.GetRights()) > 0 {
			if err := rights.RequireClient(ctx, req.GetClientIds(), removedRights.GetRights()...); err != nil {
				return err
			}
		}

		if removedRights.IncludesAll(ttnpb.Right_RIGHT_CLIENT_ALL) {
			memberRights, err := st.FindMembers(ctx, req.GetClientIds().GetEntityIdentifiers())
			if err != nil {
				return err
			}
			var hasOtherOwner bool
			for _, v := range memberRights {
				member, rights := v.Ids, v.Rights
				if unique.ID(ctx, member) == unique.ID(ctx, req.GetCollaborator().GetIds()) {
					continue
				}
				if rights.Implied().IncludesAll(ttnpb.Right_RIGHT_CLIENT_ALL) {
					hasOtherOwner = true
					break
				}
			}
			if !hasOtherOwner {
				return errClientNeedsCollaborator.New()
			}
		}

		if len(req.Collaborator.Rights) == 0 {
			// TODO: Remove delete capability (https://github.com/TheThingsNetwork/lorawan-stack/issues/6488).
			return st.DeleteMember(ctx, req.GetCollaborator().GetIds(), req.GetClientIds().GetEntityIdentifiers())
		}

		return st.SetMember(
			ctx,
			req.GetCollaborator().GetIds(),
			req.GetClientIds().GetEntityIdentifiers(),
			ttnpb.RightsFrom(req.Collaborator.Rights...),
		)
	})
	if err != nil {
		return nil, err
	}
	if len(req.Collaborator.Rights) > 0 {
		events.Publish(
			evtUpdateClientCollaborator.New(
				ctx,
				events.WithIdentifiers(req.GetClientIds(), req.GetCollaborator().GetIds()),
				events.WithData(req.GetCollaborator()),
			),
		)
		go is.notifyInternal(ctx, &ttnpb.CreateNotificationRequest{
			EntityIds:        req.GetClientIds().GetEntityIdentifiers(),
			NotificationType: ttnpb.GetNotificationTypeString(ttnpb.NotificationType_COLLABORATOR_CHANGED),
			Data:             ttnpb.MustMarshalAny(req.GetCollaborator()),
			Receivers: []ttnpb.NotificationReceiver{
				ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_ADMINISTRATIVE_CONTACT,
			},
		})
	} else {
		events.Publish(evtDeleteClientCollaborator.New(
			ctx, events.WithIdentifiers(req.GetClientIds(), req.GetCollaborator().GetIds()),
		))
	}
	return ttnpb.Empty, nil
}

func (is *IdentityServer) listClientCollaborators(
	ctx context.Context,
	req *ttnpb.ListClientCollaboratorsRequest,
) (collaborators *ttnpb.Collaborators, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	if err = rights.RequireClient(
		ctx, req.GetClientIds(), ttnpb.Right_RIGHT_CLIENT_SETTINGS_COLLABORATORS,
	); err != nil {
		defer func() { collaborators = collaborators.PublicSafe() }()
	}

	ctx = store.WithOrder(ctx, req.Order)
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		memberRights, err := st.FindMembers(ctx, req.GetClientIds().GetEntityIdentifiers())
		if err != nil {
			return err
		}
		collaborators = &ttnpb.Collaborators{}
		for _, v := range memberRights {
			member, rights := v.Ids, v.Rights
			collaborators.Collaborators = append(collaborators.Collaborators, &ttnpb.Collaborator{
				Ids:    member,
				Rights: rights.GetRights(),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return collaborators, nil
}

func (is *IdentityServer) deleteClientCollaborator(
	ctx context.Context, req *ttnpb.DeleteClientCollaboratorRequest,
) (*emptypb.Empty, error) {
	err := rights.RequireClient(ctx, req.GetClientIds(), ttnpb.Right_RIGHT_CLIENT_SETTINGS_COLLABORATORS)
	if err != nil {
		return ttnpb.Empty, err
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		removedRights, err := st.GetMember(ctx, req.GetCollaboratorIds(), req.GetClientIds().GetEntityIdentifiers())
		if err != nil {
			return err
		}
		clt, err := st.GetClient(
			ctx,
			req.GetClientIds(),
			store.FieldMask([]string{"administrative_contact", "technical_contact"}),
		)
		if err != nil {
			return err
		}
		if proto.Equal(clt.GetAdministrativeContact(), req.GetCollaboratorIds()) ||
			proto.Equal(clt.GetTechnicalContact(), req.GetCollaboratorIds()) {
			return errCollaboratorIsContact.WithAttributes("collaborator_id", req.GetCollaboratorIds().IDString())
		}
		if removedRights.Implied().IncludesAll(ttnpb.Right_RIGHT_CLIENT_ALL) {
			memberRights, err := st.FindMembers(ctx, req.GetClientIds().GetEntityIdentifiers())
			if err != nil {
				return err
			}
			var hasOtherOwner bool
			for _, v := range memberRights {
				member, rights := v.Ids, v.Rights
				if unique.ID(ctx, member) == unique.ID(ctx, req.GetCollaboratorIds()) {
					continue
				}
				if rights.Implied().IncludesAll(ttnpb.Right_RIGHT_CLIENT_ALL) {
					hasOtherOwner = true
					break
				}
			}
			if !hasOtherOwner {
				return errClientNeedsCollaborator.New()
			}
		}
		return st.DeleteMember(
			ctx,
			req.GetCollaboratorIds(),
			req.GetClientIds().GetEntityIdentifiers(),
		)
	})
	if err != nil {
		return ttnpb.Empty, err
	}
	events.Publish(evtDeleteClientCollaborator.New(
		ctx,
		events.WithIdentifiers(req.GetClientIds(), req.GetCollaboratorIds()),
	))
	return ttnpb.Empty, nil
}

type clientAccess struct {
	ttnpb.UnimplementedClientAccessServer

	*IdentityServer
}

func (ca *clientAccess) ListRights(
	ctx context.Context, req *ttnpb.ClientIdentifiers,
) (*ttnpb.Rights, error) {
	return ca.listClientRights(ctx, req)
}

func (ca *clientAccess) GetCollaborator(
	ctx context.Context,
	req *ttnpb.GetClientCollaboratorRequest,
) (*ttnpb.GetCollaboratorResponse, error) {
	return ca.getClientCollaborator(ctx, req)
}

func (ca *clientAccess) SetCollaborator(
	ctx context.Context,
	req *ttnpb.SetClientCollaboratorRequest,
) (*emptypb.Empty, error) {
	return ca.setClientCollaborator(ctx, req)
}

func (ca *clientAccess) ListCollaborators(
	ctx context.Context,
	req *ttnpb.ListClientCollaboratorsRequest,
) (*ttnpb.Collaborators, error) {
	return ca.listClientCollaborators(ctx, req)
}

func (ca *clientAccess) DeleteCollaborator(
	ctx context.Context, req *ttnpb.DeleteClientCollaboratorRequest,
) (*emptypb.Empty, error) {
	return ca.deleteClientCollaborator(ctx, req)
}
