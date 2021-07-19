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
)

var (
	evtCreateApplicationAPIKey = events.Define(
		"application.api-key.create", "create application API key",
		events.WithVisibility(ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUpdateApplicationAPIKey = events.Define(
		"application.api-key.update", "update application API key",
		events.WithVisibility(ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteApplicationAPIKey = events.Define(
		"application.api-key.delete", "delete application API key",
		events.WithVisibility(ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUpdateApplicationCollaborator = events.Define(
		"application.collaborator.update", "update application collaborator",
		events.WithVisibility(
			ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS,
			ttnpb.RIGHT_USER_APPLICATIONS_LIST,
		),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteApplicationCollaborator = events.Define(
		"application.collaborator.delete", "delete application collaborator",
		events.WithVisibility(
			ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS,
			ttnpb.RIGHT_USER_APPLICATIONS_LIST,
		),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
)

func (is *IdentityServer) listApplicationRights(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*ttnpb.Rights, error) {
	appRights, err := rights.ListApplication(ctx, *ids)
	if err != nil {
		return nil, err
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
	key, token, err := GenerateAPIKey(ctx, req.Name, req.ExpiresAt, req.Rights...)
	if err != nil {
		return nil, err
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		key, err = store.GetAPIKeyStore(db).CreateAPIKey(ctx, req.GetEntityIdentifiers(), key)
		return err
	})
	if err != nil {
		return nil, err
	}
	key.Key = token
	events.Publish(evtCreateApplicationAPIKey.NewWithIdentifiersAndData(ctx, &req.ApplicationIdentifiers, nil))
	err = is.SendContactsEmail(ctx, req, func(data emails.Data) email.MessageData {
		data.SetEntity(req)
		return &emails.APIKeyCreated{Data: data, Key: key, Rights: key.Rights}
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
		keys.APIKeys, err = store.GetAPIKeyStore(db).FindAPIKeys(ctx, req.ApplicationIdentifiers.GetEntityIdentifiers())
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
		_, key, err = store.GetAPIKeyStore(db).GetAPIKey(ctx, req.KeyId)
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
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}

	err = is.withDatabase(ctx, func(db *gorm.DB) (err error) {
		if len(req.APIKey.Rights) > 0 {
			_, key, err := store.GetAPIKeyStore(db).GetAPIKey(ctx, req.APIKey.ID)
			if err != nil {
				return err
			}

			newRights := ttnpb.RightsFrom(req.APIKey.Rights...)
			existingRights := ttnpb.RightsFrom(key.Rights...)

			// Require the caller to have all added rights.
			if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, newRights.Sub(existingRights).GetRights()...); err != nil {
				return err
			}
			// Require the caller to have all removed rights.
			if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, existingRights.Sub(newRights).GetRights()...); err != nil {
				return err
			}
		}

		key, err = store.GetAPIKeyStore(db).UpdateAPIKey(ctx, req.ApplicationIdentifiers.GetEntityIdentifiers(), &req.APIKey, req.FieldMask)
		return err
	})
	if err != nil {
		return nil, err
	}
	if key == nil { // API key was deleted.
		events.Publish(evtDeleteApplicationAPIKey.NewWithIdentifiersAndData(ctx, &req.ApplicationIdentifiers, nil))
		return &ttnpb.APIKey{}, nil
	}
	key.Key = ""
	events.Publish(evtUpdateApplicationAPIKey.NewWithIdentifiersAndData(ctx, &req.ApplicationIdentifiers, nil))
	err = is.SendContactsEmail(ctx, req, func(data emails.Data) email.MessageData {
		data.SetEntity(req)
		return &emails.APIKeyChanged{Data: data, Key: key, Rights: key.Rights}
	})
	if err != nil {
		log.FromContext(ctx).WithError(err).Error("Could not send API key update notification email")
	}

	return key, nil
}

func (is *IdentityServer) getApplicationCollaborator(ctx context.Context, req *ttnpb.GetApplicationCollaboratorRequest) (*ttnpb.GetCollaboratorResponse, error) {
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS); err != nil {
		return nil, err
	}
	res := &ttnpb.GetCollaboratorResponse{
		OrganizationOrUserIdentifiers: req.OrganizationOrUserIdentifiers,
	}
	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		rights, err := is.getMembershipStore(ctx, db).GetMember(
			ctx,
			&req.OrganizationOrUserIdentifiers,
			req.ApplicationIdentifiers.GetEntityIdentifiers(),
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

func (is *IdentityServer) setApplicationCollaborator(ctx context.Context, req *ttnpb.SetApplicationCollaboratorRequest) (*pbtypes.Empty, error) {
	// Require that caller has rights to manage collaborators.
	if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS); err != nil {
		return nil, err
	}

	err := is.withDatabase(ctx, func(db *gorm.DB) error {
		store := is.getMembershipStore(ctx, db)

		if len(req.Collaborator.Rights) > 0 {
			newRights := ttnpb.RightsFrom(req.Collaborator.Rights...)
			existingRights, err := store.GetMember(
				ctx,
				&req.Collaborator.OrganizationOrUserIdentifiers,
				req.ApplicationIdentifiers.GetEntityIdentifiers(),
			)

			if err != nil && !errors.IsNotFound(err) {
				return err
			}
			// Require the caller to have all added rights.
			if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, newRights.Sub(existingRights).GetRights()...); err != nil {
				return err
			}
			// Require the caller to have all removed rights.
			if err := rights.RequireApplication(ctx, req.ApplicationIdentifiers, existingRights.Sub(newRights).GetRights()...); err != nil {
				return err
			}
		}

		return store.SetMember(
			ctx,
			&req.Collaborator.OrganizationOrUserIdentifiers,
			req.ApplicationIdentifiers.GetEntityIdentifiers(),
			ttnpb.RightsFrom(req.Collaborator.Rights...),
		)
	})
	if err != nil {
		return nil, err
	}
	if len(req.Collaborator.Rights) > 0 {
		events.Publish(evtUpdateApplicationCollaborator.New(ctx, events.WithIdentifiers(&req.ApplicationIdentifiers, &req.Collaborator)))
		err = is.SendContactsEmail(ctx, req, func(data emails.Data) email.MessageData {
			data.SetEntity(req)
			return &emails.CollaboratorChanged{Data: data, Collaborator: req.Collaborator}
		})
		if err != nil {
			log.FromContext(ctx).WithError(err).Error("Could not send collaborator updated notification email")
		}
	} else {
		events.Publish(evtDeleteApplicationCollaborator.New(ctx, events.WithIdentifiers(&req.ApplicationIdentifiers, &req.Collaborator)))
	}
	return ttnpb.Empty, nil
}

func (is *IdentityServer) listApplicationCollaborators(ctx context.Context, req *ttnpb.ListApplicationCollaboratorsRequest) (collaborators *ttnpb.Collaborators, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	if err = rights.RequireApplication(ctx, req.ApplicationIdentifiers, ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS); err != nil {
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
		memberRights, err := is.getMembershipStore(ctx, db).FindMembers(ctx, req.ApplicationIdentifiers.GetEntityIdentifiers())
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

func (aa *applicationAccess) GetCollaborator(ctx context.Context, req *ttnpb.GetApplicationCollaboratorRequest) (*ttnpb.GetCollaboratorResponse, error) {
	return aa.getApplicationCollaborator(ctx, req)
}

func (aa *applicationAccess) SetCollaborator(ctx context.Context, req *ttnpb.SetApplicationCollaboratorRequest) (*pbtypes.Empty, error) {
	return aa.setApplicationCollaborator(ctx, req)
}

func (aa *applicationAccess) ListCollaborators(ctx context.Context, req *ttnpb.ListApplicationCollaboratorsRequest) (*ttnpb.Collaborators, error) {
	return aa.listApplicationCollaborators(ctx, req)
}
