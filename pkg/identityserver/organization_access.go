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
	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

func (is *IdentityServer) listOrganizationRights(ctx context.Context, ids *ttnpb.OrganizationIdentifiers) (*ttnpb.Rights, error) {
	rights, ok := rights.FromContext(ctx)
	if !ok {
		return &ttnpb.Rights{}, nil
	}
	orgRights, ok := rights.OrganizationRights[unique.ID(ctx, ids)]
	if !ok || orgRights == nil {
		return &ttnpb.Rights{}, nil
	}
	return orgRights, nil
}

func (is *IdentityServer) createOrganizationAPIKey(ctx context.Context, req *ttnpb.CreateOrganizationAPIKeyRequest) (key *ttnpb.APIKey, err error) {
	err = rights.RequireOrganization(ctx, req.OrganizationIdentifiers, ttnpb.RIGHT_ORGANIZATION_SETTINGS_API_KEYS)
	if err != nil {
		return nil, err
	}
	err = rights.RequireOrganization(ctx, req.OrganizationIdentifiers, req.Rights...)
	if err != nil {
		return nil, err
	}
	id, err := auth.GenerateID(ctx)
	if err != nil {
		return nil, err
	}
	token, err := auth.APIKey.Generate(ctx, id)
	if err != nil {
		return nil, err
	}
	key = &ttnpb.APIKey{
		ID:     id,
		Key:    token,
		Name:   req.Name,
		Rights: req.Rights,
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		keyStore := store.GetAPIKeyStore(db)
		err = keyStore.CreateAPIKey(ctx, req.OrganizationIdentifiers.EntityIdentifiers(), key)
		return err
	})
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (is *IdentityServer) listOrganizationAPIKeys(ctx context.Context, ids *ttnpb.OrganizationIdentifiers) (keys *ttnpb.APIKeys, err error) {
	err = rights.RequireOrganization(ctx, *ids, ttnpb.RIGHT_ORGANIZATION_SETTINGS_API_KEYS)
	if err != nil {
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
	return keys, nil
}

func (is *IdentityServer) updateOrganizationAPIKey(ctx context.Context, req *ttnpb.UpdateOrganizationAPIKeyRequest) (key *ttnpb.APIKey, err error) {
	err = rights.RequireOrganization(ctx, req.OrganizationIdentifiers, ttnpb.RIGHT_ORGANIZATION_SETTINGS_API_KEYS)
	if err != nil {
		return nil, err
	}
	err = rights.RequireOrganization(ctx, req.OrganizationIdentifiers, req.Rights...)
	if err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		keyStore := store.GetAPIKeyStore(db)
		key, err = keyStore.UpdateAPIKey(ctx, req.OrganizationIdentifiers.EntityIdentifiers(), &req.APIKey)
		return err
	})
	if err != nil {
		return nil, err
	}
	if key == nil {
		return &ttnpb.APIKey{}, nil
	}
	return key, nil
}

func (is *IdentityServer) setOrganizationCollaborator(ctx context.Context, req *ttnpb.SetOrganizationCollaboratorRequest) (*types.Empty, error) {
	err := rights.RequireOrganization(ctx, req.OrganizationIdentifiers, ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS)
	if err != nil {
		return nil, err
	}
	err = rights.RequireOrganization(ctx, req.OrganizationIdentifiers, req.Collaborator.Rights...)
	if err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		memberStore := store.GetMembershipStore(db)
		err = memberStore.SetMember(ctx, &req.Collaborator.OrganizationOrUserIdentifiers, req.OrganizationIdentifiers.EntityIdentifiers(), ttnpb.RightsFrom(req.Collaborator.Rights...))
		return err
	})
	if err != nil {
		return nil, err
	}
	return ttnpb.Empty, nil
}

func (is *IdentityServer) listOrganizationCollaborators(ctx context.Context, ids *ttnpb.OrganizationIdentifiers) (collaborators *ttnpb.Collaborators, err error) {
	err = rights.RequireOrganization(ctx, *ids, ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS)
	if err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
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

type organizationAccess struct {
	*IdentityServer
}

func (oa *organizationAccess) ListRights(ctx context.Context, req *ttnpb.OrganizationIdentifiers) (*ttnpb.Rights, error) {
	return oa.listOrganizationRights(ctx, req)
}
func (oa *organizationAccess) CreateAPIKey(ctx context.Context, req *ttnpb.CreateOrganizationAPIKeyRequest) (*ttnpb.APIKey, error) {
	return oa.createOrganizationAPIKey(ctx, req)
}
func (oa *organizationAccess) ListAPIKeys(ctx context.Context, req *ttnpb.OrganizationIdentifiers) (*ttnpb.APIKeys, error) {
	return oa.listOrganizationAPIKeys(ctx, req)
}
func (oa *organizationAccess) UpdateAPIKey(ctx context.Context, req *ttnpb.UpdateOrganizationAPIKeyRequest) (*ttnpb.APIKey, error) {
	return oa.updateOrganizationAPIKey(ctx, req)
}
func (oa *organizationAccess) SetCollaborator(ctx context.Context, req *ttnpb.SetOrganizationCollaboratorRequest) (*types.Empty, error) {
	return oa.setOrganizationCollaborator(ctx, req)
}
func (oa *organizationAccess) ListCollaborators(ctx context.Context, req *ttnpb.OrganizationIdentifiers) (*ttnpb.Collaborators, error) {
	return oa.listOrganizationCollaborators(ctx, req)
}
