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

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/v3/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/v3/pkg/email"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/emails"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

var (
	evtUpdateClientCollaborator = events.Define(
		"client.collaborator.update", "update client collaborator",
		events.WithVisibility(
			ttnpb.Right_RIGHT_CLIENT_ALL,
			ttnpb.Right_RIGHT_USER_CLIENTS_LIST,
		),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteClientCollaborator = events.Define(
		"client.collaborator.delete", "delete client collaborator",
		events.WithVisibility(
			ttnpb.Right_RIGHT_CLIENT_ALL,
			ttnpb.Right_RIGHT_USER_CLIENTS_LIST,
		),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
)

func (is *IdentityServer) listClientRights(ctx context.Context, ids *ttnpb.ClientIdentifiers) (*ttnpb.Rights, error) {
	cliRights, err := rights.ListClient(ctx, *ids)
	if err != nil {
		return nil, err
	}
	return cliRights.Intersect(ttnpb.AllClientRights), nil
}

func (is *IdentityServer) getClientCollaborator(ctx context.Context, req *ttnpb.GetClientCollaboratorRequest) (*ttnpb.GetCollaboratorResponse, error) {
	if err := rights.RequireClient(ctx, *req.GetClientIds(), ttnpb.Right_RIGHT_CLIENT_ALL); err != nil {
		return nil, err
	}
	res := &ttnpb.GetCollaboratorResponse{
		Ids: req.GetCollaborator(),
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		rights, err := is.getMembershipStore(ctx, db).GetMember(
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

var errClientNeedsCollaborator = errors.DefineFailedPrecondition("client_needs_collaborator", "every client needs at least one collaborator with all rights")

func (is *IdentityServer) setClientCollaborator(ctx context.Context, req *ttnpb.SetClientCollaboratorRequest) (*pbtypes.Empty, error) {
	// Require that caller has rights to manage collaborators.
	if err := rights.RequireClient(ctx, *req.GetClientIds(), ttnpb.Right_RIGHT_CLIENT_ALL); err != nil {
		return nil, err
	}

	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		store := is.getMembershipStore(ctx, db)

		existingRights, err := store.GetMember(
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
			if err := rights.RequireClient(ctx, *req.GetClientIds(), addedRights.GetRights()...); err != nil {
				return err
			}
		}

		// Unless we're deleting the collaborator, require the caller to have all removed rights.
		if len(newRights.GetRights()) > 0 && len(removedRights.GetRights()) > 0 {
			if err := rights.RequireClient(ctx, *req.GetClientIds(), removedRights.GetRights()...); err != nil {
				return err
			}
		}

		if removedRights.IncludesAll(ttnpb.Right_RIGHT_CLIENT_ALL) {
			memberRights, err := is.getMembershipStore(ctx, db).FindMembers(ctx, req.GetClientIds().GetEntityIdentifiers())
			if err != nil {
				return err
			}
			var hasOtherOwner bool
			for member, rights := range memberRights {
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

		return store.SetMember(
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
		events.Publish(evtUpdateClientCollaborator.New(ctx, events.WithIdentifiers(req.GetClientIds(), req.GetCollaborator().GetIds())))
		err = is.SendContactsEmail(ctx, req, func(data emails.Data) email.MessageData {
			data.SetEntity(req)
			return &emails.CollaboratorChanged{Data: data, Collaborator: *req.GetCollaborator()}
		})
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Could not send collaborator updated notification email")
		}
	} else {
		events.Publish(evtDeleteClientCollaborator.New(ctx, events.WithIdentifiers(req.GetClientIds(), req.GetCollaborator().GetIds())))
	}
	return ttnpb.Empty, nil
}

func (is *IdentityServer) listClientCollaborators(ctx context.Context, req *ttnpb.ListClientCollaboratorsRequest) (collaborators *ttnpb.Collaborators, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	if err = rights.RequireClient(ctx, *req.GetClientIds(), ttnpb.Right_RIGHT_CLIENT_ALL); err != nil {
		defer func() { collaborators = collaborators.PublicSafe() }()
	}
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		memberRights, err := is.getMembershipStore(ctx, db).FindMembers(ctx, req.GetClientIds().GetEntityIdentifiers())
		if err != nil {
			return err
		}
		collaborators = &ttnpb.Collaborators{}
		for member, rights := range memberRights {
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

type clientAccess struct {
	*IdentityServer
}

func (ca *clientAccess) ListRights(ctx context.Context, req *ttnpb.ClientIdentifiers) (*ttnpb.Rights, error) {
	return ca.listClientRights(ctx, req)
}

func (ca *clientAccess) GetCollaborator(ctx context.Context, req *ttnpb.GetClientCollaboratorRequest) (*ttnpb.GetCollaboratorResponse, error) {
	return ca.getClientCollaborator(ctx, req)
}

func (ca *clientAccess) SetCollaborator(ctx context.Context, req *ttnpb.SetClientCollaboratorRequest) (*pbtypes.Empty, error) {
	return ca.setClientCollaborator(ctx, req)
}

func (ca *clientAccess) ListCollaborators(ctx context.Context, req *ttnpb.ListClientCollaboratorsRequest) (*ttnpb.Collaborators, error) {
	return ca.listClientCollaborators(ctx, req)
}
