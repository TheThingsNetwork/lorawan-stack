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

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/email"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/identityserver/emails"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtUpdateClientCollaborator = events.Define(
		"client.collaborator.update", "update client collaborator",
		ttnpb.RIGHT_CLIENT_ALL,
		ttnpb.RIGHT_USER_CLIENTS_LIST,
	)
	evtDeleteClientCollaborator = events.Define(
		"client.collaborator.delete", "delete client collaborator",
		ttnpb.RIGHT_CLIENT_ALL,
		ttnpb.RIGHT_USER_CLIENTS_LIST,
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
	if err := rights.RequireClient(ctx, req.ClientIdentifiers, ttnpb.RIGHT_CLIENT_ALL); err != nil {
		return nil, err
	}
	res := &ttnpb.GetCollaboratorResponse{
		OrganizationOrUserIdentifiers: req.OrganizationOrUserIdentifiers,
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		rights, err := is.getMembershipStore(ctx, db).GetMember(
			ctx,
			&req.OrganizationOrUserIdentifiers,
			req.ClientIdentifiers,
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

func (is *IdentityServer) setClientCollaborator(ctx context.Context, req *ttnpb.SetClientCollaboratorRequest) (*types.Empty, error) {
	// Require that caller has rights to manage collaborators.
	if err := rights.RequireClient(ctx, req.ClientIdentifiers, ttnpb.RIGHT_CLIENT_ALL); err != nil {
		return nil, err
	}
	// Require that caller has at least the rights we're giving to the collaborator.
	if err := rights.RequireClient(ctx, req.ClientIdentifiers, req.Collaborator.Rights...); err != nil {
		return nil, err
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		return is.getMembershipStore(ctx, db).SetMember(
			ctx,
			&req.Collaborator.OrganizationOrUserIdentifiers,
			req.ClientIdentifiers,
			ttnpb.RightsFrom(req.Collaborator.Rights...),
		)
	})
	if err != nil {
		return nil, err
	}
	if len(req.Collaborator.Rights) > 0 {
		events.Publish(evtUpdateClientCollaborator(ctx, ttnpb.CombineIdentifiers(req.ClientIdentifiers, req.Collaborator), nil))
		err = is.SendContactsEmail(ctx, req.EntityIdentifiers(), func(data emails.Data) email.MessageData {
			data.SetEntity(req.EntityIdentifiers())
			return &emails.CollaboratorChanged{Data: data, Collaborator: req.Collaborator}
		})
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Could not send collaborator updated notification email")
		}
	} else {
		events.Publish(evtDeleteClientCollaborator(ctx, ttnpb.CombineIdentifiers(req.ClientIdentifiers, req.Collaborator), nil))
	}
	return ttnpb.Empty, nil
}

func (is *IdentityServer) listClientCollaborators(ctx context.Context, req *ttnpb.ListClientCollaboratorsRequest) (collaborators *ttnpb.Collaborators, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil { // Client collaborators can be seen by all authenticated users.
		return nil, err
	}
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		memberRights, err := is.getMembershipStore(ctx, db).FindMembers(ctx, req.ClientIdentifiers)
		if err != nil {
			return err
		}
		collaborators = &ttnpb.Collaborators{}
		for member, rights := range memberRights {
			collaborators.Collaborators = append(collaborators.Collaborators, &ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *member,
				Rights:                        rights.GetRights(),
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

func (ca *clientAccess) SetCollaborator(ctx context.Context, req *ttnpb.SetClientCollaboratorRequest) (*types.Empty, error) {
	return ca.setClientCollaborator(ctx, req)
}

func (ca *clientAccess) ListCollaborators(ctx context.Context, req *ttnpb.ListClientCollaboratorsRequest) (*ttnpb.Collaborators, error) {
	return ca.listClientCollaborators(ctx, req)
}
