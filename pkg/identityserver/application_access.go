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
	evtCreateApplicationAPIKey = events.Define(
		"application.api-key.create", "create application API key",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_SETTINGS_API_KEYS),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUpdateApplicationAPIKey = events.Define(
		"application.api-key.update", "update application API key",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_SETTINGS_API_KEYS),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteApplicationAPIKey = events.Define(
		"application.api-key.delete", "delete application API key",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_SETTINGS_API_KEYS),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUpdateApplicationCollaborator = events.Define(
		"application.collaborator.update", "update application collaborator",
		events.WithVisibility(
			ttnpb.Right_RIGHT_APPLICATION_SETTINGS_COLLABORATORS,
			ttnpb.Right_RIGHT_USER_APPLICATIONS_LIST,
		),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteApplicationCollaborator = events.Define(
		"application.collaborator.delete", "delete application collaborator",
		events.WithVisibility(
			ttnpb.Right_RIGHT_APPLICATION_SETTINGS_COLLABORATORS,
			ttnpb.Right_RIGHT_USER_APPLICATIONS_LIST,
		),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
)

func (*IdentityServer) listApplicationRights(
	ctx context.Context, ids *ttnpb.ApplicationIdentifiers,
) (*ttnpb.Rights, error) {
	appRights, err := rights.ListApplication(ctx, ids)
	if err != nil {
		return nil, err
	}
	return appRights.Intersect(ttnpb.AllApplicationRights), nil
}

func (is *IdentityServer) createApplicationAPIKey(
	ctx context.Context, req *ttnpb.CreateApplicationAPIKeyRequest,
) (key *ttnpb.APIKey, err error) {
	// Require that caller has rights to manage API keys.
	err = rights.RequireApplication(ctx, req.GetApplicationIds(), ttnpb.Right_RIGHT_APPLICATION_SETTINGS_API_KEYS)
	if err != nil {
		return nil, err
	}
	// Require that caller has at least the rights of the API key.
	if err = rights.RequireApplication(ctx, req.GetApplicationIds(), req.Rights...); err != nil {
		return nil, err
	}
	key, token, err := GenerateAPIKey(ctx, req.Name, ttnpb.StdTime(req.ExpiresAt), req.Rights...)
	if err != nil {
		return nil, err
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		key, err = st.CreateAPIKey(ctx, req.GetApplicationIds().GetEntityIdentifiers(), key)
		return err
	})
	if err != nil {
		return nil, err
	}
	key.Key = ""

	events.Publish(evtCreateApplicationAPIKey.NewWithIdentifiersAndData(ctx, req.GetApplicationIds(), key))
	go is.notifyInternal(ctx, &ttnpb.CreateNotificationRequest{
		EntityIds:        req.GetApplicationIds().GetEntityIdentifiers(),
		NotificationType: ttnpb.GetNotificationTypeString(ttnpb.NotificationType_API_KEY_CREATED),
		Data:             ttnpb.MustMarshalAny(key),
		Receivers: []ttnpb.NotificationReceiver{
			ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_ADMINISTRATIVE_CONTACT,
		},
	})

	key.Key = token
	return key, nil
}

func (is *IdentityServer) listApplicationAPIKeys(
	ctx context.Context, req *ttnpb.ListApplicationAPIKeysRequest,
) (keys *ttnpb.APIKeys, err error) {
	err = rights.RequireApplication(ctx, req.GetApplicationIds(), ttnpb.Right_RIGHT_APPLICATION_SETTINGS_API_KEYS)
	if err != nil {
		return nil, err
	}
	ctx = store.WithOrder(ctx, req.Order)
	var total uint64
	ctx = store.WithPagination(ctx, req.Limit, req.Page, &total)
	defer func() {
		if err == nil {
			setTotalHeader(ctx, total)
		}
	}()
	keys = &ttnpb.APIKeys{}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		keys.ApiKeys, err = st.FindAPIKeys(ctx, req.GetApplicationIds().GetEntityIdentifiers())
		return err
	})
	if err != nil {
		return nil, err
	}
	for _, key := range keys.ApiKeys {
		key.Key = ""
	}
	return keys, nil
}

func (is *IdentityServer) getApplicationAPIKey(
	ctx context.Context, req *ttnpb.GetApplicationAPIKeyRequest,
) (key *ttnpb.APIKey, err error) {
	err = rights.RequireApplication(ctx, req.GetApplicationIds(), ttnpb.Right_RIGHT_APPLICATION_SETTINGS_API_KEYS)
	if err != nil {
		return nil, err
	}

	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		key, err = st.GetAPIKey(ctx, req.GetApplicationIds().GetEntityIdentifiers(), req.KeyId)
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

func (is *IdentityServer) updateApplicationAPIKey(
	ctx context.Context, req *ttnpb.UpdateApplicationAPIKeyRequest,
) (key *ttnpb.APIKey, err error) {
	// Require that caller has rights to manage API keys.
	err = rights.RequireApplication(ctx, req.GetApplicationIds(), ttnpb.Right_RIGHT_APPLICATION_SETTINGS_API_KEYS)
	if err != nil {
		return nil, err
	}

	// Backwards compatibility for older clients.
	if len(req.FieldMask.GetPaths()) == 0 {
		req.FieldMask = ttnpb.FieldMask("rights", "name")
	}

	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		if len(req.ApiKey.Rights) > 0 {
			key, err := st.GetAPIKey(ctx, req.GetApplicationIds().GetEntityIdentifiers(), req.ApiKey.Id)
			if err != nil {
				return err
			}

			newRights := ttnpb.RightsFrom(req.ApiKey.Rights...)
			existingRights := ttnpb.RightsFrom(key.Rights...)

			// Require the caller to have all added rights.
			if err := rights.RequireApplication(
				ctx, req.GetApplicationIds(), newRights.Sub(existingRights).GetRights()...,
			); err != nil {
				return err
			}
			// Require the caller to have all removed rights.
			if err := rights.RequireApplication(
				ctx, req.GetApplicationIds(), existingRights.Sub(newRights).GetRights()...,
			); err != nil {
				return err
			}
		}

		if len(req.ApiKey.Rights) == 0 && ttnpb.HasAnyField(req.GetFieldMask().GetPaths(), "rights") {
			// TODO: Remove delete capability (https://github.com/TheThingsNetwork/lorawan-stack/issues/6488).
			return st.DeleteAPIKey(ctx, req.ApplicationIds.GetEntityIdentifiers(), req.ApiKey)
		}

		key, err = st.UpdateAPIKey(ctx, req.ApplicationIds.GetEntityIdentifiers(), req.ApiKey, req.FieldMask.GetPaths())
		return err
	})
	if err != nil {
		return nil, err
	}
	if key == nil { // API key was deleted.
		events.Publish(evtDeleteApplicationAPIKey.NewWithIdentifiersAndData(ctx, req.GetApplicationIds(), nil))
		return &ttnpb.APIKey{}, nil
	}
	key.Key = ""

	events.Publish(evtUpdateApplicationAPIKey.NewWithIdentifiersAndData(ctx, req.GetApplicationIds(), key))
	go is.notifyInternal(ctx, &ttnpb.CreateNotificationRequest{
		EntityIds:        req.GetApplicationIds().GetEntityIdentifiers(),
		NotificationType: ttnpb.GetNotificationTypeString(ttnpb.NotificationType_API_KEY_CHANGED),
		Data:             ttnpb.MustMarshalAny(key),
		Receivers: []ttnpb.NotificationReceiver{
			ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_ADMINISTRATIVE_CONTACT,
		},
	})

	return key, nil
}

func (is *IdentityServer) deleteApplicationAPIKey(
	ctx context.Context, req *ttnpb.DeleteApplicationAPIKeyRequest,
) (*emptypb.Empty, error) {
	// Require that caller has rights to manage API keys.
	err := rights.RequireApplication(ctx, req.GetApplicationIds(), ttnpb.Right_RIGHT_APPLICATION_SETTINGS_API_KEYS)
	if err != nil {
		return nil, err
	}

	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		return st.DeleteAPIKey(ctx, req.ApplicationIds.GetEntityIdentifiers(), &ttnpb.APIKey{Id: req.KeyId})
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtDeleteApplicationAPIKey.NewWithIdentifiersAndData(ctx, req.GetApplicationIds(), nil))
	return ttnpb.Empty, nil
}

func (is *IdentityServer) getApplicationCollaborator(
	ctx context.Context, req *ttnpb.GetApplicationCollaboratorRequest,
) (_ *ttnpb.GetCollaboratorResponse, err error) {
	err = rights.RequireApplication(ctx, req.GetApplicationIds(), ttnpb.Right_RIGHT_APPLICATION_SETTINGS_COLLABORATORS)
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
			req.GetApplicationIds().GetEntityIdentifiers(),
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

var errApplicationNeedsCollaborator = errors.DefineFailedPrecondition(
	"application_needs_collaborator", "every application needs at least one collaborator with all rights",
)

func (is *IdentityServer) setApplicationCollaborator(
	ctx context.Context, req *ttnpb.SetApplicationCollaboratorRequest,
) (_ *emptypb.Empty, err error) {
	// Require that caller has rights to manage collaborators.
	err = rights.RequireApplication(ctx, req.GetApplicationIds(), ttnpb.Right_RIGHT_APPLICATION_SETTINGS_COLLABORATORS)
	if err != nil {
		return nil, err
	}

	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		existingRights, err := st.GetMember(
			ctx,
			req.GetCollaborator().GetIds(),
			req.GetApplicationIds().GetEntityIdentifiers(),
		)
		if err != nil && !errors.IsNotFound(err) {
			return err
		}
		existingRights = existingRights.Implied()
		newRights := ttnpb.RightsFrom(req.GetCollaborator().GetRights()...).Implied()
		addedRights := newRights.Sub(existingRights)
		removedRights := existingRights.Sub(newRights)

		// Require the caller to have all added rights.
		if len(addedRights.GetRights()) > 0 {
			if err := rights.RequireApplication(ctx, req.GetApplicationIds(), addedRights.GetRights()...); err != nil {
				return err
			}
		}

		// Unless we're deleting the collaborator, require the caller to have all removed rights.
		if len(newRights.GetRights()) > 0 && len(removedRights.GetRights()) > 0 {
			if err := rights.RequireApplication(ctx, req.GetApplicationIds(), removedRights.GetRights()...); err != nil {
				return err
			}
		}

		if removedRights.IncludesAll(ttnpb.Right_RIGHT_APPLICATION_ALL) {
			memberRights, err := st.FindMembers(ctx, req.GetApplicationIds().GetEntityIdentifiers())
			if err != nil {
				return err
			}
			var hasOtherOwner bool
			for _, v := range memberRights {
				member, rights := v.Ids, v.Rights

				if unique.ID(ctx, member) == unique.ID(ctx, req.GetCollaborator().GetIds()) {
					continue
				}
				if rights.Implied().IncludesAll(ttnpb.Right_RIGHT_APPLICATION_ALL) {
					hasOtherOwner = true
					break
				}
			}
			if !hasOtherOwner {
				return errApplicationNeedsCollaborator.New()
			}
		}

		if len(req.Collaborator.Rights) == 0 {
			// TODO: Remove delete capability (https://github.com/TheThingsNetwork/lorawan-stack/issues/6488).
			return st.DeleteMember(ctx, req.GetCollaborator().GetIds(), req.GetApplicationIds().GetEntityIdentifiers())
		}

		return st.SetMember(
			ctx,
			req.GetCollaborator().GetIds(),
			req.GetApplicationIds().GetEntityIdentifiers(),
			ttnpb.RightsFrom(req.Collaborator.Rights...),
		)
	})
	if err != nil {
		return nil, err
	}
	if len(req.GetCollaborator().GetRights()) > 0 {
		events.Publish(evtUpdateApplicationCollaborator.New(
			ctx,
			events.WithIdentifiers(req.GetApplicationIds(), req.GetCollaborator().GetIds()),
			events.WithData(req.GetCollaborator()),
		))
		go is.notifyInternal(ctx, &ttnpb.CreateNotificationRequest{
			EntityIds:        req.GetApplicationIds().GetEntityIdentifiers(),
			NotificationType: ttnpb.GetNotificationTypeString(ttnpb.NotificationType_COLLABORATOR_CHANGED),
			Data:             ttnpb.MustMarshalAny(req.GetCollaborator()),
			Receivers: []ttnpb.NotificationReceiver{
				ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_ADMINISTRATIVE_CONTACT,
			},
		})
	} else {
		events.Publish(evtDeleteApplicationCollaborator.New(
			ctx, events.WithIdentifiers(req.GetApplicationIds(), req.GetCollaborator().GetIds()),
		))
	}
	return ttnpb.Empty, nil
}

func (is *IdentityServer) listApplicationCollaborators(
	ctx context.Context, req *ttnpb.ListApplicationCollaboratorsRequest,
) (collaborators *ttnpb.Collaborators, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	err = rights.RequireApplication(ctx, req.GetApplicationIds(), ttnpb.Right_RIGHT_APPLICATION_SETTINGS_COLLABORATORS)
	if err != nil {
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
		memberRights, err := st.FindMembers(ctx, req.GetApplicationIds().GetEntityIdentifiers())
		if err != nil {
			return err
		}
		collaborators = &ttnpb.Collaborators{
			Collaborators: make([]*ttnpb.Collaborator, len(memberRights)),
		}
		for i, v := range memberRights {
			member, rights := v.Ids, v.Rights
			collaborators.Collaborators[i] = &ttnpb.Collaborator{
				Ids:    member,
				Rights: rights.GetRights(),
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return collaborators, nil
}

func (is *IdentityServer) deleteApplicationCollaborator(
	ctx context.Context, req *ttnpb.DeleteApplicationCollaboratorRequest,
) (*emptypb.Empty, error) {
	err := rights.RequireApplication(ctx, req.GetApplicationIds(), ttnpb.Right_RIGHT_APPLICATION_SETTINGS_COLLABORATORS)
	if err != nil {
		return nil, err
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		removedRights, err := st.GetMember(
			ctx, req.GetCollaboratorIds(), req.GetApplicationIds().GetEntityIdentifiers(),
		)
		if err != nil {
			return err
		}
		app, err := st.GetApplication(
			ctx,
			req.GetApplicationIds(),
			store.FieldMask([]string{"administrative_contact", "technical_contact"}),
		)
		if err != nil {
			return err
		}
		if proto.Equal(app.GetAdministrativeContact(), req.GetCollaboratorIds()) ||
			proto.Equal(app.GetTechnicalContact(), req.GetCollaboratorIds()) {
			return errCollaboratorIsContact.WithAttributes("collaborator_id", req.GetCollaboratorIds().IDString())
		}
		if removedRights.Implied().IncludesAll(ttnpb.Right_RIGHT_APPLICATION_ALL) {
			memberRights, err := st.FindMembers(ctx, req.GetApplicationIds().GetEntityIdentifiers())
			if err != nil {
				return err
			}
			var hasOtherOwner bool
			for _, v := range memberRights {
				member, rights := v.Ids, v.Rights
				if unique.ID(ctx, member) == unique.ID(ctx, req.GetCollaboratorIds()) {
					continue
				}
				if rights.Implied().IncludesAll(ttnpb.Right_RIGHT_APPLICATION_ALL) {
					hasOtherOwner = true
					break
				}
			}
			if !hasOtherOwner {
				return errOrganizationNeedsCollaborator.New()
			}
		}

		return st.DeleteMember(
			ctx,
			req.GetCollaboratorIds(),
			req.GetApplicationIds().GetEntityIdentifiers(),
		)
	})
	if err != nil {
		return ttnpb.Empty, err
	}
	events.Publish(evtDeleteApplicationCollaborator.New(
		ctx,
		events.WithIdentifiers(req.GetApplicationIds(), req.GetCollaboratorIds()),
	))
	return ttnpb.Empty, nil
}

type applicationAccess struct {
	ttnpb.UnimplementedApplicationAccessServer

	*IdentityServer
}

func (aa *applicationAccess) ListRights(
	ctx context.Context, req *ttnpb.ApplicationIdentifiers,
) (*ttnpb.Rights, error) {
	return aa.listApplicationRights(ctx, req)
}

func (aa *applicationAccess) CreateAPIKey(
	ctx context.Context, req *ttnpb.CreateApplicationAPIKeyRequest,
) (*ttnpb.APIKey, error) {
	return aa.createApplicationAPIKey(ctx, req)
}

func (aa *applicationAccess) ListAPIKeys(
	ctx context.Context, req *ttnpb.ListApplicationAPIKeysRequest,
) (*ttnpb.APIKeys, error) {
	return aa.listApplicationAPIKeys(ctx, req)
}

func (aa *applicationAccess) GetAPIKey(
	ctx context.Context, req *ttnpb.GetApplicationAPIKeyRequest,
) (*ttnpb.APIKey, error) {
	return aa.getApplicationAPIKey(ctx, req)
}

func (aa *applicationAccess) UpdateAPIKey(
	ctx context.Context, req *ttnpb.UpdateApplicationAPIKeyRequest,
) (*ttnpb.APIKey, error) {
	return aa.updateApplicationAPIKey(ctx, req)
}

func (aa *applicationAccess) DeleteAPIKey(
	ctx context.Context, req *ttnpb.DeleteApplicationAPIKeyRequest,
) (*emptypb.Empty, error) {
	return aa.deleteApplicationAPIKey(ctx, req)
}

func (aa *applicationAccess) GetCollaborator(
	ctx context.Context, req *ttnpb.GetApplicationCollaboratorRequest,
) (*ttnpb.GetCollaboratorResponse, error) {
	return aa.getApplicationCollaborator(ctx, req)
}

func (aa *applicationAccess) SetCollaborator(
	ctx context.Context, req *ttnpb.SetApplicationCollaboratorRequest,
) (*emptypb.Empty, error) {
	return aa.setApplicationCollaborator(ctx, req)
}

func (aa *applicationAccess) ListCollaborators(
	ctx context.Context, req *ttnpb.ListApplicationCollaboratorsRequest,
) (*ttnpb.Collaborators, error) {
	return aa.listApplicationCollaborators(ctx, req)
}

func (aa *applicationAccess) DeleteCollaborator(
	ctx context.Context, req *ttnpb.DeleteApplicationCollaboratorRequest,
) (*emptypb.Empty, error) {
	return aa.deleteApplicationCollaborator(ctx, req)
}
