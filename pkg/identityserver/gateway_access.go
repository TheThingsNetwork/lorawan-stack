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
	"go.thethings.network/lorawan-stack/pkg/unique"
)

var (
	evtCreateGatewayAPIKey       = events.Define("gateway.api-key.create", "create gateway API key")
	evtUpdateGatewayAPIKey       = events.Define("gateway.api-key.update", "update gateway API key")
	evtDeleteGatewayAPIKey       = events.Define("gateway.api-key.delete", "delete gateway API key")
	evtUpdateGatewayCollaborator = events.Define("gateway.collaborator.update", "update gateway collaborator")
	evtDeleteGatewayCollaborator = events.Define("gateway.collaborator.delete", "delete gateway collaborator")
)

func (is *IdentityServer) listGatewayRights(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (*ttnpb.Rights, error) {
	rights, ok := rights.FromContext(ctx)
	if !ok {
		return &ttnpb.Rights{}, nil
	}
	gtwRights, ok := rights.GatewayRights[unique.ID(ctx, ids)]
	if !ok || gtwRights == nil {
		return &ttnpb.Rights{}, nil
	}
	return gtwRights.Intersect(ttnpb.AllGatewayRights), nil
}

func (is *IdentityServer) createGatewayAPIKey(ctx context.Context, req *ttnpb.CreateGatewayAPIKeyRequest) (key *ttnpb.APIKey, err error) {
	// Require that caller has rights to manage API keys.
	if err = rights.RequireGateway(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}
	// Require that caller has at least the rights of the API key.
	if err = rights.RequireGateway(ctx, req.GatewayIdentifiers, req.Rights...); err != nil {
		return nil, err
	}
	key, token, err := generateAPIKey(ctx, req.Name, req.Rights...)
	if err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		return store.GetAPIKeyStore(db).CreateAPIKey(ctx, req.GatewayIdentifiers.EntityIdentifiers(), key)
	})
	if err != nil {
		return nil, err
	}
	key.Key = token
	events.Publish(evtCreateGatewayAPIKey(ctx, req.GatewayIdentifiers, nil))
	err = is.SendContactsEmail(ctx, req.EntityIdentifiers(), func(data emails.Data) email.MessageData {
		data.SetEntity(req.EntityIdentifiers())
		return &emails.APIKeyCreated{Data: data, Identifier: key.PrettyName(), Rights: key.Rights}
	})
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("Could not send API key creation notification email")
	}
	return key, nil
}

func (is *IdentityServer) listGatewayAPIKeys(ctx context.Context, req *ttnpb.ListGatewayAPIKeysRequest) (keys *ttnpb.APIKeys, err error) {
	if err = rights.RequireGateway(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	keys = &ttnpb.APIKeys{}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		keys.APIKeys, err = store.GetAPIKeyStore(db).FindAPIKeys(ctx, req.EntityIdentifiers())
		return err
	})
	if err != nil {
		return nil, err
	}
	for _, key := range keys.APIKeys {
		key.Key = ""
	}
	return keys, nil
}

func (is *IdentityServer) getGatewayAPIKey(ctx context.Context, req *ttnpb.GetGatewayAPIKeyRequest) (key *ttnpb.APIKey, err error) {
	if err = rights.RequireGateway(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		_, key, err = store.GetAPIKeyStore(db).GetAPIKey(ctx, req.KeyID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	key.Key = ""
	return key, nil
}

func (is *IdentityServer) updateGatewayAPIKey(ctx context.Context, req *ttnpb.UpdateGatewayAPIKeyRequest) (key *ttnpb.APIKey, err error) {
	// Require that caller has rights to manage API keys.
	if err = rights.RequireGateway(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}
	// Require that caller has at least the rights of the API key.
	if err = rights.RequireGateway(ctx, req.GatewayIdentifiers, req.Rights...); err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		key, err = store.GetAPIKeyStore(db).UpdateAPIKey(ctx, req.GatewayIdentifiers.EntityIdentifiers(), &req.APIKey)
		return err
	})
	if err != nil {
		return nil, err
	}
	if key == nil {
		return &ttnpb.APIKey{}, nil
	}
	key.Key = ""
	if len(req.Rights) > 0 {
		events.Publish(evtUpdateGatewayAPIKey(ctx, req.GatewayIdentifiers, nil))
		err = is.SendContactsEmail(ctx, req.EntityIdentifiers(), func(data emails.Data) email.MessageData {
			data.SetEntity(req.EntityIdentifiers())
			return &emails.APIKeyChanged{Data: data, Identifier: key.PrettyName(), Rights: key.Rights}
		})
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Could not send API key update notification email")
		}
	} else {
		events.Publish(evtDeleteGatewayAPIKey(ctx, req.GatewayIdentifiers, nil))
	}
	return key, nil
}

func (is *IdentityServer) setGatewayCollaborator(ctx context.Context, req *ttnpb.SetGatewayCollaboratorRequest) (*types.Empty, error) {
	// Require that caller has rights to manage collaborators.
	if err := rights.RequireGateway(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_SETTINGS_COLLABORATORS); err != nil {
		return nil, err
	}
	// Require that caller has at least the rights we're giving to the collaborator.
	if err := rights.RequireGateway(ctx, req.GatewayIdentifiers, req.Collaborator.Rights...); err != nil {
		return nil, err
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		return store.GetMembershipStore(db).SetMember(
			ctx,
			&req.Collaborator.OrganizationOrUserIdentifiers,
			req.GatewayIdentifiers.EntityIdentifiers(),
			ttnpb.RightsFrom(req.Collaborator.Rights...),
		)
	})
	if err != nil {
		return nil, err
	}
	if len(req.Collaborator.Rights) > 0 {
		events.Publish(evtUpdateGatewayCollaborator(ctx, req.GatewayIdentifiers, nil))
		err = is.SendContactsEmail(ctx, req.EntityIdentifiers(), func(data emails.Data) email.MessageData {
			data.SetEntity(req.EntityIdentifiers())
			return &emails.CollaboratorChanged{Data: data, Collaborator: req.Collaborator}
		})
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Could not send collaborator updated notification email")
		}
	} else {
		events.Publish(evtDeleteGatewayCollaborator(ctx, req.GatewayIdentifiers, nil))
	}
	is.invalidateCachedMembershipsForAccount(ctx, &req.Collaborator.OrganizationOrUserIdentifiers)
	return ttnpb.Empty, nil
}

func (is *IdentityServer) listGatewayCollaborators(ctx context.Context, req *ttnpb.ListGatewayCollaboratorsRequest) (collaborators *ttnpb.Collaborators, err error) {
	if err = rights.RequireGateway(ctx, req.GatewayIdentifiers, ttnpb.RIGHT_GATEWAY_SETTINGS_COLLABORATORS); err != nil {
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
		memberRights, err := store.GetMembershipStore(db).FindMembers(ctx, req.EntityIdentifiers())
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

type gatewayAccess struct {
	*IdentityServer
}

func (ga *gatewayAccess) ListRights(ctx context.Context, req *ttnpb.GatewayIdentifiers) (*ttnpb.Rights, error) {
	return ga.listGatewayRights(ctx, req)
}
func (ga *gatewayAccess) CreateAPIKey(ctx context.Context, req *ttnpb.CreateGatewayAPIKeyRequest) (*ttnpb.APIKey, error) {
	return ga.createGatewayAPIKey(ctx, req)
}
func (ga *gatewayAccess) ListAPIKeys(ctx context.Context, req *ttnpb.ListGatewayAPIKeysRequest) (*ttnpb.APIKeys, error) {
	return ga.listGatewayAPIKeys(ctx, req)
}
func (ga *gatewayAccess) GetAPIKey(ctx context.Context, req *ttnpb.GetGatewayAPIKeyRequest) (*ttnpb.APIKey, error) {
	return ga.getGatewayAPIKey(ctx, req)
}
func (ga *gatewayAccess) UpdateAPIKey(ctx context.Context, req *ttnpb.UpdateGatewayAPIKeyRequest) (*ttnpb.APIKey, error) {
	return ga.updateGatewayAPIKey(ctx, req)
}
func (ga *gatewayAccess) SetCollaborator(ctx context.Context, req *ttnpb.SetGatewayCollaboratorRequest) (*types.Empty, error) {
	return ga.setGatewayCollaborator(ctx, req)
}
func (ga *gatewayAccess) ListCollaborators(ctx context.Context, req *ttnpb.ListGatewayCollaboratorsRequest) (*ttnpb.Collaborators, error) {
	return ga.listGatewayCollaborators(ctx, req)
}
