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
	evtCreateApplicationAPIKey       = events.Define("application.api-key.create", "create application API key")
	evtUpdateApplicationAPIKey       = events.Define("application.api-key.update", "update application API key")
	evtDeleteApplicationAPIKey       = events.Define("application.api-key.delete", "delete application API key")
	evtUpdateApplicationCollaborator = events.Define("application.collaborator.update", "update application collaborator")
	evtDeleteApplicationCollaborator = events.Define("application.collaborator.delete", "delete application collaborator")
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
	return appRights.Intersect(ttnpb.AllApplicationRights), nil
}

func (is *IdentityServer) createApplicationAPIKey(ctx context.Context, req *ttnpb.CreateApplicationAPIKeyRequest) (key *ttnpb.APIKey, err error) {
	// Require that caller has rights to manage API keys.
	if err = rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}
	// Require that caller has at least the rights of the API key.
	if err = rights.RequireApplication(ctx, req.ApplicationIdentifiers, req.Rights...); err != nil {
		return nil, err
	}
	key, token, err := generateAPIKey(ctx, req.Name, req.Rights...)
	if err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		return store.GetAPIKeyStore(db).CreateAPIKey(ctx, req.ApplicationIdentifiers, key)
	})
	if err != nil {
		return nil, err
	}
	key.Key = token
	events.Publish(evtCreateApplicationAPIKey(ctx, req.ApplicationIdentifiers, nil))
	err = is.SendContactsEmail(ctx, req.EntityIdentifiers(), func(data emails.Data) email.MessageData {
		data.SetEntity(req.EntityIdentifiers())
		return &emails.APIKeyCreated{Data: data, Identifier: key.PrettyName(), Rights: key.Rights}
	})
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("Could not send API key creation notification email")
	}
	return key, nil
}

func (is *IdentityServer) listApplicationAPIKeys(ctx context.Context, req *ttnpb.ListApplicationAPIKeysRequest) (keys *ttnpb.APIKeys, err error) {
	if err = rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS); err != nil {
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
		keys.APIKeys, err = store.GetAPIKeyStore(db).FindAPIKeys(ctx, req.ApplicationIdentifiers)
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

func (is *IdentityServer) getApplicationAPIKey(ctx context.Context, req *ttnpb.GetApplicationAPIKeyRequest) (key *ttnpb.APIKey, err error) {
	if err = rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS); err != nil {
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

func (is *IdentityServer) updateApplicationAPIKey(ctx context.Context, req *ttnpb.UpdateApplicationAPIKeyRequest) (key *ttnpb.APIKey, err error) {
	// Require that caller has rights to manage API keys.
	if err = rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}
	// Require that caller has at least the rights of the API key.
	if err = rights.RequireApplication(ctx, req.ApplicationIdentifiers, req.Rights...); err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		key, err = store.GetAPIKeyStore(db).UpdateAPIKey(ctx, req.ApplicationIdentifiers, &req.APIKey)
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
		err = is.SendContactsEmail(ctx, req.EntityIdentifiers(), func(data emails.Data) email.MessageData {
			data.SetEntity(req.EntityIdentifiers())
			return &emails.APIKeyChanged{Data: data, Identifier: key.PrettyName(), Rights: key.Rights}
		})
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Could not send API key update notification email")
		}
	} else {
		events.Publish(evtDeleteApplicationAPIKey(ctx, req.ApplicationIdentifiers, nil))
	}
	return key, nil
}

func (is *IdentityServer) setApplicationCollaborator(ctx context.Context, req *ttnpb.SetApplicationCollaboratorRequest) (*types.Empty, error) {
	// Require that caller has rights to manage collaborators.
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS); err != nil {
		return nil, err
	}
	// Require that caller has at least the rights we're giving to the collaborator.
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, req.Collaborator.Rights...); err != nil {
		return nil, err
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		return store.GetMembershipStore(db).SetMember(
			ctx,
			&req.Collaborator.OrganizationOrUserIdentifiers,
			req.ApplicationIdentifiers,
			ttnpb.RightsFrom(req.Collaborator.Rights...),
		)
	})
	if err != nil {
		return nil, err
	}
	if len(req.Collaborator.Rights) > 0 {
		events.Publish(evtUpdateApplicationCollaborator(ctx, req.ApplicationIdentifiers, nil))
		err = is.SendContactsEmail(ctx, req.EntityIdentifiers(), func(data emails.Data) email.MessageData {
			data.SetEntity(req.EntityIdentifiers())
			return &emails.CollaboratorChanged{Data: data, Collaborator: req.Collaborator}
		})
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Could not send collaborator updated notification email")
		}
	} else {
		events.Publish(evtDeleteApplicationCollaborator(ctx, req.ApplicationIdentifiers, nil))
	}
	is.invalidateCachedMembershipsForAccount(ctx, &req.Collaborator.OrganizationOrUserIdentifiers)
	return ttnpb.Empty, nil
}

func (is *IdentityServer) listApplicationCollaborators(ctx context.Context, req *ttnpb.ListApplicationCollaboratorsRequest) (collaborators *ttnpb.Collaborators, err error) {
	if err = rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS); err != nil {
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
		memberRights, err := store.GetMembershipStore(db).FindMembers(ctx, req.ApplicationIdentifiers)
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

type applicationAccess struct {
	*IdentityServer
}

func (aa *applicationAccess) ListRights(ctx context.Context, req *ttnpb.ApplicationIdentifiers) (*ttnpb.Rights, error) {
	return aa.listApplicationRights(ctx, req)
}

func (aa *applicationAccess) CreateAPIKey(ctx context.Context, req *ttnpb.CreateApplicationAPIKeyRequest) (*ttnpb.APIKey, error) {
	return aa.createApplicationAPIKey(ctx, req)
}

func (aa *applicationAccess) ListAPIKeys(ctx context.Context, req *ttnpb.ListApplicationAPIKeysRequest) (*ttnpb.APIKeys, error) {
	return aa.listApplicationAPIKeys(ctx, req)
}

func (aa *applicationAccess) GetAPIKey(ctx context.Context, req *ttnpb.GetApplicationAPIKeyRequest) (*ttnpb.APIKey, error) {
	return aa.getApplicationAPIKey(ctx, req)
}

func (aa *applicationAccess) UpdateAPIKey(ctx context.Context, req *ttnpb.UpdateApplicationAPIKeyRequest) (*ttnpb.APIKey, error) {
	return aa.updateApplicationAPIKey(ctx, req)
}

func (aa *applicationAccess) SetCollaborator(ctx context.Context, req *ttnpb.SetApplicationCollaboratorRequest) (*types.Empty, error) {
	return aa.setApplicationCollaborator(ctx, req)
}

func (aa *applicationAccess) ListCollaborators(ctx context.Context, req *ttnpb.ListApplicationCollaboratorsRequest) (*ttnpb.Collaborators, error) {
	return aa.listApplicationCollaborators(ctx, req)
}
