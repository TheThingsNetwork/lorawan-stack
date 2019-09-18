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
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/events"
	"go.thethings.network/lorawan-stack/pkg/identityserver/emails"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	evtCreateOrganizationAPIKey = events.Define(
		"organization.api-key.create", "create organization API key",
		ttnpb.RIGHT_ORGANIZATION_SETTINGS_API_KEYS,
	)
	evtUpdateOrganizationAPIKey = events.Define(
		"organization.api-key.update", "update organization API key",
		ttnpb.RIGHT_ORGANIZATION_SETTINGS_API_KEYS,
	)
	evtDeleteOrganizationAPIKey = events.Define(
		"organization.api-key.delete", "delete organization API key",
		ttnpb.RIGHT_ORGANIZATION_SETTINGS_API_KEYS,
	)
	evtUpdateOrganizationCollaborator = events.Define(
		"organization.collaborator.update", "update organization collaborator",
		ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS,
		ttnpb.RIGHT_USER_ORGANIZATIONS_LIST,
	)
	evtDeleteOrganizationCollaborator = events.Define(
		"organization.collaborator.delete", "delete organization collaborator",
		ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS,
		ttnpb.RIGHT_USER_ORGANIZATIONS_LIST,
	)
)

func (is *IdentityServer) listOrganizationRights(ctx context.Context, ids *ttnpb.OrganizationIdentifiers) (*ttnpb.Rights, error) {
	orgRights, err := rights.ListOrganization(ctx, *ids)
	if err != nil {
		return nil, err
	}
	return orgRights.Intersect(ttnpb.AllEntityRights.Union(ttnpb.AllOrganizationRights)), nil
}

func (is *IdentityServer) createOrganizationAPIKey(ctx context.Context, req *ttnpb.CreateOrganizationAPIKeyRequest) (key *ttnpb.APIKey, err error) {
	// Require that caller has rights to manage API keys.
	if err = rights.RequireOrganization(ctx, req.OrganizationIdentifiers, ttnpb.RIGHT_ORGANIZATION_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}
	// Require that caller has at least the rights of the API key.
	if err = rights.RequireOrganization(ctx, req.OrganizationIdentifiers, req.Rights...); err != nil {
		return nil, err
	}
	key, token, err := generateAPIKey(ctx, req.Name, req.Rights...)
	if err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		return store.GetAPIKeyStore(db).CreateAPIKey(ctx, req.OrganizationIdentifiers, key)
	})
	if err != nil {
		return nil, err
	}
	key.Key = token
	events.Publish(evtCreateOrganizationAPIKey(ctx, req.OrganizationIdentifiers, nil))
	err = is.SendContactsEmail(ctx, req.EntityIdentifiers(), func(data emails.Data) email.MessageData {
		data.SetEntity(req.EntityIdentifiers())
		return &emails.APIKeyCreated{Data: data, Identifier: key.PrettyName(), Rights: key.Rights}
	})
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("Could not send API key creation notification email")
	}
	return key, nil
}

func (is *IdentityServer) listOrganizationAPIKeys(ctx context.Context, req *ttnpb.ListOrganizationAPIKeysRequest) (keys *ttnpb.APIKeys, err error) {
	if err = rights.RequireOrganization(ctx, req.OrganizationIdentifiers, ttnpb.RIGHT_ORGANIZATION_SETTINGS_API_KEYS); err != nil {
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
		keys.APIKeys, err = store.GetAPIKeyStore(db).FindAPIKeys(ctx, req.OrganizationIdentifiers)
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

func (is *IdentityServer) getOrganizationAPIKey(ctx context.Context, req *ttnpb.GetOrganizationAPIKeyRequest) (key *ttnpb.APIKey, err error) {
	if err = rights.RequireOrganization(ctx, req.OrganizationIdentifiers, ttnpb.RIGHT_ORGANIZATION_SETTINGS_API_KEYS); err != nil {
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

func (is *IdentityServer) updateOrganizationAPIKey(ctx context.Context, req *ttnpb.UpdateOrganizationAPIKeyRequest) (key *ttnpb.APIKey, err error) {
	// Require that caller has rights to manage API keys.
	if err = rights.RequireOrganization(ctx, req.OrganizationIdentifiers, ttnpb.RIGHT_ORGANIZATION_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}
	// Require that caller has at least the rights of the API key.
	if err = rights.RequireOrganization(ctx, req.OrganizationIdentifiers, req.Rights...); err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		key, err = store.GetAPIKeyStore(db).UpdateAPIKey(ctx, req.OrganizationIdentifiers, &req.APIKey)
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
		events.Publish(evtUpdateOrganizationAPIKey(ctx, req.OrganizationIdentifiers, nil))
		err = is.SendContactsEmail(ctx, req.EntityIdentifiers(), func(data emails.Data) email.MessageData {
			data.SetEntity(req.EntityIdentifiers())
			return &emails.APIKeyChanged{Data: data, Identifier: key.PrettyName(), Rights: key.Rights}
		})
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Could not send API key update notification email")
		}
	} else {
		events.Publish(evtDeleteOrganizationAPIKey(ctx, req.OrganizationIdentifiers, nil))
	}
	return key, nil
}

func (is *IdentityServer) getOrganizationCollaborator(ctx context.Context, req *ttnpb.GetOrganizationCollaboratorRequest) (*ttnpb.GetCollaboratorResponse, error) {
	if err := rights.RequireOrganization(ctx, req.OrganizationIdentifiers, ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS); err != nil {
		return nil, err
	}
	res := &ttnpb.GetCollaboratorResponse{
		OrganizationOrUserIdentifiers: req.OrganizationOrUserIdentifiers,
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		rights, err := is.getMembershipStore(ctx, db).GetMember(
			ctx,
			&req.OrganizationOrUserIdentifiers,
			req.OrganizationIdentifiers,
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

func (is *IdentityServer) setOrganizationCollaborator(ctx context.Context, req *ttnpb.SetOrganizationCollaboratorRequest) (*types.Empty, error) {
	// Require that caller has rights to manage collaborators.
	if err := rights.RequireOrganization(ctx, req.OrganizationIdentifiers, ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS); err != nil {
		return nil, err
	}

	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		store := is.getMembershipStore(ctx, db)

		if len(req.Collaborator.Rights) > 0 {
			newRights := ttnpb.RightsFrom(req.Collaborator.Rights...)
			existingRights, err := store.GetMember(
				ctx,
				&req.Collaborator.OrganizationOrUserIdentifiers,
				req.OrganizationIdentifiers,
			)

			if err != nil && !errors.IsNotFound(err) {
				return err
			}
			// Require the caller to have all added rights.
			if err := rights.RequireOrganization(ctx, req.OrganizationIdentifiers, newRights.Sub(existingRights).GetRights()...); err != nil {
				return err
			}
			// Require the caller to have all removed rights.
			if err := rights.RequireOrganization(ctx, req.OrganizationIdentifiers, existingRights.Sub(newRights).GetRights()...); err != nil {
				return err
			}
		}

		return store.SetMember(
			ctx,
			&req.Collaborator.OrganizationOrUserIdentifiers,
			req.OrganizationIdentifiers,
			ttnpb.RightsFrom(req.Collaborator.Rights...),
		)
	})
	if err != nil {
		return nil, err
	}
	if len(req.Collaborator.Rights) > 0 {
		events.Publish(evtUpdateOrganizationCollaborator(ctx, ttnpb.CombineIdentifiers(req.OrganizationIdentifiers, req.Collaborator), nil))
		err = is.SendContactsEmail(ctx, req.EntityIdentifiers(), func(data emails.Data) email.MessageData {
			data.SetEntity(req.EntityIdentifiers())
			return &emails.CollaboratorChanged{Data: data, Collaborator: req.Collaborator}
		})
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Could not send collaborator updated notification email")
		}
	} else {
		events.Publish(evtDeleteOrganizationCollaborator(ctx, ttnpb.CombineIdentifiers(req.OrganizationIdentifiers, req.Collaborator), nil))
	}
	return ttnpb.Empty, nil
}

func (is *IdentityServer) listOrganizationCollaborators(ctx context.Context, req *ttnpb.ListOrganizationCollaboratorsRequest) (collaborators *ttnpb.Collaborators, err error) {
	if err = rights.RequireOrganization(ctx, req.OrganizationIdentifiers, ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS); err != nil {
		return nil, err
	}
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		memberRights, err := is.getMembershipStore(ctx, db).FindMembers(ctx, req.OrganizationIdentifiers)
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

type organizationAccess struct {
	*IdentityServer
}

func (oa *organizationAccess) ListRights(ctx context.Context, req *ttnpb.OrganizationIdentifiers) (*ttnpb.Rights, error) {
	return oa.listOrganizationRights(ctx, req)
}

func (oa *organizationAccess) CreateAPIKey(ctx context.Context, req *ttnpb.CreateOrganizationAPIKeyRequest) (*ttnpb.APIKey, error) {
	return oa.createOrganizationAPIKey(ctx, req)
}

func (oa *organizationAccess) ListAPIKeys(ctx context.Context, req *ttnpb.ListOrganizationAPIKeysRequest) (*ttnpb.APIKeys, error) {
	return oa.listOrganizationAPIKeys(ctx, req)
}

func (oa *organizationAccess) GetAPIKey(ctx context.Context, req *ttnpb.GetOrganizationAPIKeyRequest) (*ttnpb.APIKey, error) {
	return oa.getOrganizationAPIKey(ctx, req)
}

func (oa *organizationAccess) UpdateAPIKey(ctx context.Context, req *ttnpb.UpdateOrganizationAPIKeyRequest) (*ttnpb.APIKey, error) {
	return oa.updateOrganizationAPIKey(ctx, req)
}

func (oa *organizationAccess) GetCollaborator(ctx context.Context, req *ttnpb.GetOrganizationCollaboratorRequest) (*ttnpb.GetCollaboratorResponse, error) {
	return oa.getOrganizationCollaborator(ctx, req)
}

func (oa *organizationAccess) SetCollaborator(ctx context.Context, req *ttnpb.SetOrganizationCollaboratorRequest) (*types.Empty, error) {
	return oa.setOrganizationCollaborator(ctx, req)
}

func (oa *organizationAccess) ListCollaborators(ctx context.Context, req *ttnpb.ListOrganizationCollaboratorsRequest) (*ttnpb.Collaborators, error) {
	return oa.listOrganizationCollaborators(ctx, req)
}
