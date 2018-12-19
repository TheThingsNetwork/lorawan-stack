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

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

var (
	evtCreateApplicationAPIKey       = events.Define("application.api-key.create", "Create application API key")
	evtUpdateApplicationAPIKey       = events.Define("application.api-key.update", "Update application API key")
	evtDeleteApplicationAPIKey       = events.Define("application.api-key.delete", "Delete application API key")
	evtUpdateApplicationCollaborator = events.Define("application.collaborator.update", "Update application collaborator")
	evtDeleteApplicationCollaborator = events.Define("application.collaborator.delete", "Delete application collaborator")
)

func (is *IdentityServer) listApplicationRights(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*ttnpb.Rights, error) {
	rights, ok := rights.FromContext(ctx)
	if !ok {
		return &ttnpb.Rights{}, nil
	}
	appRights, ok := rights.ApplicationRights[unique.ID(ctx, ids)]
	if !ok || appRights == nil {
		return &ttnpb.Rights{}, nil
	}
	return appRights, nil
}

func (is *IdentityServer) createApplicationAPIKey(ctx context.Context, req *ttnpb.CreateApplicationAPIKeyRequest) (key *ttnpb.APIKey, err error) {
	if err = rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}
	if err = rights.RequireApplication(ctx, req.ApplicationIdentifiers, req.Rights...); err != nil {
		return nil, err
	}
	key, token, err := generateAPIKey(ctx, req.Name, req.Rights...)
	if err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		keyStore := store.GetAPIKeyStore(db)
		return keyStore.CreateAPIKey(ctx, req.ApplicationIdentifiers.EntityIdentifiers(), key)
	})
	if err != nil {
		return nil, err
	}
	key.Key = token
	events.Publish(evtCreateApplicationAPIKey(ctx, req.ApplicationIdentifiers, nil))
	return key, nil
}

func (is *IdentityServer) listApplicationAPIKeys(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (keys *ttnpb.APIKeys, err error) {
	if err = rights.RequireApplication(ctx, *ids, ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}
	keys = new(ttnpb.APIKeys)
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		keyStore := store.GetAPIKeyStore(db)
		keys.APIKeys, err = keyStore.FindAPIKeys(ctx, ids.EntityIdentifiers())
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

func (is *IdentityServer) updateApplicationAPIKey(ctx context.Context, req *ttnpb.UpdateApplicationAPIKeyRequest) (key *ttnpb.APIKey, err error) {
	if err = rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}
	if err = rights.RequireApplication(ctx, req.ApplicationIdentifiers, req.Rights...); err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		keyStore := store.GetAPIKeyStore(db)
		key, err = keyStore.UpdateAPIKey(ctx, req.ApplicationIdentifiers.EntityIdentifiers(), &req.APIKey)
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
		events.Publish(evtUpdateApplicationAPIKey(ctx, req.ApplicationIdentifiers, nil))
	} else {
		events.Publish(evtDeleteApplicationAPIKey(ctx, req.ApplicationIdentifiers, nil))
	}
	return key, nil
}

func (is *IdentityServer) setApplicationCollaborator(ctx context.Context, req *ttnpb.SetApplicationCollaboratorRequest) (*types.Empty, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS); err != nil {
		return nil, err
	}
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, req.Collaborator.Rights...); err != nil {
		return nil, err
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		memberStore := store.GetMembershipStore(db)
		return memberStore.SetMember(
			ctx,
			&req.Collaborator.OrganizationOrUserIdentifiers,
			req.ApplicationIdentifiers.EntityIdentifiers(),
			ttnpb.RightsFrom(req.Collaborator.Rights...),
		)
	})
	if err != nil {
		return nil, err
	}
	if len(req.Collaborator.Rights) > 0 {
		events.Publish(evtUpdateApplicationCollaborator(ctx, req.ApplicationIdentifiers, nil))
	} else {
		events.Publish(evtDeleteApplicationCollaborator(ctx, req.ApplicationIdentifiers, nil))
	}
	is.invalidateCachedMembershipsForAccount(ctx, &req.Collaborator.OrganizationOrUserIdentifiers)
	return ttnpb.Empty, nil
}

func (is *IdentityServer) listApplicationCollaborators(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (collaborators *ttnpb.Collaborators, err error) {
	if err = rights.RequireApplication(ctx, *ids, ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS); err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		memberStore := store.GetMembershipStore(db)
		memberRights, err := memberStore.FindMembers(ctx, ids.EntityIdentifiers())
		if err != nil {
			return err
		}
		collaborators = new(ttnpb.Collaborators)
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

type applicationAccess struct {
	*IdentityServer
}

func (aa *applicationAccess) ListRights(ctx context.Context, req *ttnpb.ApplicationIdentifiers) (*ttnpb.Rights, error) {
	return aa.listApplicationRights(ctx, req)
}
func (aa *applicationAccess) CreateAPIKey(ctx context.Context, req *ttnpb.CreateApplicationAPIKeyRequest) (*ttnpb.APIKey, error) {
	return aa.createApplicationAPIKey(ctx, req)
}
func (aa *applicationAccess) ListAPIKeys(ctx context.Context, req *ttnpb.ApplicationIdentifiers) (*ttnpb.APIKeys, error) {
	return aa.listApplicationAPIKeys(ctx, req)
}
func (aa *applicationAccess) UpdateAPIKey(ctx context.Context, req *ttnpb.UpdateApplicationAPIKeyRequest) (*ttnpb.APIKey, error) {
	return aa.updateApplicationAPIKey(ctx, req)
}
func (aa *applicationAccess) SetCollaborator(ctx context.Context, req *ttnpb.SetApplicationCollaboratorRequest) (*types.Empty, error) {
	return aa.setApplicationCollaborator(ctx, req)
}
func (aa *applicationAccess) ListCollaborators(ctx context.Context, req *ttnpb.ApplicationIdentifiers) (*ttnpb.Collaborators, error) {
	return aa.listApplicationCollaborators(ctx, req)
}
