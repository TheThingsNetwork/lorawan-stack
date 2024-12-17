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
	evtCreateOrganizationAPIKey = events.Define(
		"organization.api-key.create", "create organization API key",
		events.WithVisibility(ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_API_KEYS),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUpdateOrganizationAPIKey = events.Define(
		"organization.api-key.update", "update organization API key",
		events.WithVisibility(ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_API_KEYS),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteOrganizationAPIKey = events.Define(
		"organization.api-key.delete", "delete organization API key",
		events.WithVisibility(ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_API_KEYS),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUpdateOrganizationCollaborator = events.Define(
		"organization.collaborator.update", "update organization collaborator",
		events.WithVisibility(
			ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_MEMBERS,
			ttnpb.Right_RIGHT_USER_ORGANIZATIONS_LIST,
		),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteOrganizationCollaborator = events.Define(
		"organization.collaborator.delete", "delete organization collaborator",
		events.WithVisibility(
			ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_MEMBERS,
			ttnpb.Right_RIGHT_USER_ORGANIZATIONS_LIST,
		),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)

	errOrganizationNeedsCollaborator = errors.DefineFailedPrecondition(
		"organization_needs_collaborator", "every organization needs at least one collaborator with all rights",
	)
)

func (*IdentityServer) listOrganizationRights(
	ctx context.Context, ids *ttnpb.OrganizationIdentifiers,
) (*ttnpb.Rights, error) {
	orgRights, err := rights.ListOrganization(ctx, ids)
	if err != nil {
		return nil, err
	}
	return orgRights.Intersect(ttnpb.AllEntityRights.Union(ttnpb.AllOrganizationRights)), nil
}

func (is *IdentityServer) createOrganizationAPIKey(
	ctx context.Context, req *ttnpb.CreateOrganizationAPIKeyRequest,
) (key *ttnpb.APIKey, err error) {
	// Require that caller has rights to manage API keys.
	err = rights.RequireOrganization(ctx, req.GetOrganizationIds(), ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_API_KEYS)
	if err != nil {
		return nil, err
	}
	// Require that caller has at least the rights of the API key.
	if err = rights.RequireOrganization(ctx, req.GetOrganizationIds(), req.Rights...); err != nil {
		return nil, err
	}
	key, token, err := GenerateAPIKey(ctx, req.Name, ttnpb.StdTime(req.ExpiresAt), req.Rights...)
	if err != nil {
		return nil, err
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		key, err = st.CreateAPIKey(ctx, req.GetOrganizationIds().GetEntityIdentifiers(), key)
		return err
	})
	if err != nil {
		return nil, err
	}
	key.Key = ""

	events.Publish(evtCreateOrganizationAPIKey.NewWithIdentifiersAndData(ctx, req.GetOrganizationIds(), key))
	go is.notifyInternal(ctx, &ttnpb.CreateNotificationRequest{
		EntityIds:        req.GetOrganizationIds().GetEntityIdentifiers(),
		NotificationType: ttnpb.GetNotificationTypeString(ttnpb.NotificationType_API_KEY_CREATED),
		Data:             ttnpb.MustMarshalAny(key),
		Receivers: []ttnpb.NotificationReceiver{
			ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_ADMINISTRATIVE_CONTACT,
		},
	})

	key.Key = token
	return key, nil
}

func (is *IdentityServer) listOrganizationAPIKeys(
	ctx context.Context, req *ttnpb.ListOrganizationAPIKeysRequest,
) (keys *ttnpb.APIKeys, err error) {
	err = rights.RequireOrganization(ctx, req.GetOrganizationIds(), ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_API_KEYS)
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
		keys.ApiKeys, err = st.FindAPIKeys(ctx, req.GetOrganizationIds().GetEntityIdentifiers())
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

func (is *IdentityServer) getOrganizationAPIKey(
	ctx context.Context, req *ttnpb.GetOrganizationAPIKeyRequest,
) (key *ttnpb.APIKey, err error) {
	err = rights.RequireOrganization(ctx, req.GetOrganizationIds(), ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_API_KEYS)
	if err != nil {
		return nil, err
	}

	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		key, err = st.GetAPIKey(ctx, req.GetOrganizationIds().GetEntityIdentifiers(), req.KeyId)
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

func (is *IdentityServer) updateOrganizationAPIKey(
	ctx context.Context, req *ttnpb.UpdateOrganizationAPIKeyRequest,
) (key *ttnpb.APIKey, err error) {
	// Require that caller has rights to manage API keys.
	err = rights.RequireOrganization(ctx, req.GetOrganizationIds(), ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_API_KEYS)
	if err != nil {
		return nil, err
	}

	// Backwards compatibility for older clients.
	if len(req.FieldMask.GetPaths()) == 0 {
		req.FieldMask = ttnpb.FieldMask("rights", "name")
	}

	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		if len(req.ApiKey.Rights) > 0 {
			key, err = st.GetAPIKey(ctx, req.GetOrganizationIds().GetEntityIdentifiers(), req.ApiKey.Id)
			if err != nil {
				return err
			}

			newRights := ttnpb.RightsFrom(req.ApiKey.Rights...)
			existingRights := ttnpb.RightsFrom(key.Rights...)

			// Require the caller to have all added rights.
			err = rights.RequireOrganization(ctx, req.GetOrganizationIds(), newRights.Sub(existingRights).GetRights()...)
			if err != nil {
				return err
			}
			// Require the caller to have all removed rights.
			err = rights.RequireOrganization(ctx, req.GetOrganizationIds(), existingRights.Sub(newRights).GetRights()...)
			if err != nil {
				return err
			}
		}

		if len(req.ApiKey.Rights) == 0 && ttnpb.HasAnyField(req.GetFieldMask().GetPaths(), "rights") {
			// TODO: Remove delete capability (https://github.com/TheThingsNetwork/lorawan-stack/issues/6488).
			return st.DeleteAPIKey(ctx, req.GetOrganizationIds().GetEntityIdentifiers(), req.ApiKey)
		}

		key, err = st.UpdateAPIKey(
			ctx, req.GetOrganizationIds().GetEntityIdentifiers(), req.ApiKey, req.FieldMask.GetPaths(),
		)
		return err
	})
	if err != nil {
		return nil, err
	}
	if key == nil { // API key was deleted.
		events.Publish(evtDeleteOrganizationAPIKey.NewWithIdentifiersAndData(ctx, req.GetOrganizationIds(), nil))
		return &ttnpb.APIKey{}, nil
	}
	key.Key = ""

	events.Publish(evtUpdateOrganizationAPIKey.NewWithIdentifiersAndData(ctx, req.GetOrganizationIds(), key))
	go is.notifyInternal(ctx, &ttnpb.CreateNotificationRequest{
		EntityIds:        req.GetOrganizationIds().GetEntityIdentifiers(),
		NotificationType: ttnpb.GetNotificationTypeString(ttnpb.NotificationType_API_KEY_CHANGED),
		Data:             ttnpb.MustMarshalAny(key),
		Receivers: []ttnpb.NotificationReceiver{
			ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_ADMINISTRATIVE_CONTACT,
		},
	})

	return key, nil
}

func (is *IdentityServer) deleteOrganizationAPIKey(
	ctx context.Context, req *ttnpb.DeleteOrganizationAPIKeyRequest,
) (*emptypb.Empty, error) {
	// Require that caller has rights to manage API keys.
	err := rights.RequireOrganization(ctx, req.GetOrganizationIds(), ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_API_KEYS)
	if err != nil {
		return ttnpb.Empty, err
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		return st.DeleteAPIKey(ctx, req.GetOrganizationIds().GetEntityIdentifiers(), &ttnpb.APIKey{Id: req.KeyId})
	})
	if err != nil {
		return ttnpb.Empty, err
	}
	events.Publish(evtDeleteUserAPIKey.New(ctx, events.WithIdentifiers(req.GetOrganizationIds())))
	return ttnpb.Empty, nil
}

func (is *IdentityServer) getOrganizationCollaborator(
	ctx context.Context, req *ttnpb.GetOrganizationCollaboratorRequest,
) (*ttnpb.GetCollaboratorResponse, error) {
	err := rights.RequireOrganization(ctx, req.GetOrganizationIds(), ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_MEMBERS)
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
			req.GetOrganizationIds().GetEntityIdentifiers(),
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

func (is *IdentityServer) setOrganizationCollaborator( //nolint:gocyclo
	ctx context.Context, req *ttnpb.SetOrganizationCollaboratorRequest,
) (_ *emptypb.Empty, err error) {
	// Require that caller has rights to manage collaborators.
	err = rights.RequireOrganization(ctx, req.GetOrganizationIds(), ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_MEMBERS)
	if err != nil {
		return nil, err
	}

	if req.GetCollaborator().GetIds().EntityType() == "organization" {
		return nil, errNestedOrganizations.New()
	}

	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		existingRights, err := st.GetMember(
			ctx,
			req.GetCollaborator().GetIds(),
			req.GetOrganizationIds().GetEntityIdentifiers(),
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
			err := rights.RequireOrganization(ctx, req.GetOrganizationIds(), addedRights.GetRights()...)
			if err != nil {
				return err
			}
		}

		// Unless we're deleting the collaborator, require the caller to have all removed rights.
		if len(newRights.GetRights()) > 0 && len(removedRights.GetRights()) > 0 {
			err := rights.RequireOrganization(ctx, req.GetOrganizationIds(), removedRights.GetRights()...)
			if err != nil {
				return err
			}
		}

		if removedRights.IncludesAll(ttnpb.Right_RIGHT_ORGANIZATION_ALL) {
			memberRights, err := st.FindMembers(ctx, req.GetOrganizationIds().GetEntityIdentifiers())
			if err != nil {
				return err
			}
			var hasOtherOwner bool
			for _, v := range memberRights {
				member, rights := v.Ids, v.Rights
				if unique.ID(ctx, member) == unique.ID(ctx, req.GetCollaborator().GetIds()) {
					continue
				}
				if rights.Implied().IncludesAll(ttnpb.Right_RIGHT_ORGANIZATION_ALL) {
					hasOtherOwner = true
					break
				}
			}
			if !hasOtherOwner {
				return errOrganizationNeedsCollaborator.New()
			}
		}

		if len(req.Collaborator.Rights) == 0 {
			// TODO: Remove delete capability (https://github.com/TheThingsNetwork/lorawan-stack/issues/6488).
			return st.DeleteMember(ctx, req.GetCollaborator().GetIds(), req.GetOrganizationIds().GetEntityIdentifiers())
		}

		return st.SetMember(
			ctx,
			req.GetCollaborator().GetIds(),
			req.GetOrganizationIds().GetEntityIdentifiers(),
			ttnpb.RightsFrom(req.GetCollaborator().GetRights()...),
		)
	})
	if err != nil {
		return nil, err
	}
	if len(req.GetCollaborator().GetRights()) > 0 {
		events.Publish(evtUpdateOrganizationCollaborator.New(
			ctx,
			events.WithIdentifiers(req.GetOrganizationIds(), req.GetCollaborator().GetIds()),
			events.WithData(req.GetCollaborator()),
		))
		go is.notifyInternal(ctx, &ttnpb.CreateNotificationRequest{
			EntityIds:        req.GetOrganizationIds().GetEntityIdentifiers(),
			NotificationType: ttnpb.GetNotificationTypeString(ttnpb.NotificationType_COLLABORATOR_CHANGED),
			Data:             ttnpb.MustMarshalAny(req.GetCollaborator()),
			Receivers: []ttnpb.NotificationReceiver{
				ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_ADMINISTRATIVE_CONTACT,
			},
		})
	} else {
		events.Publish(evtDeleteOrganizationCollaborator.New(
			ctx, events.WithIdentifiers(req.GetOrganizationIds(), req.GetCollaborator().GetIds()),
		))
	}
	return ttnpb.Empty, nil
}

func (is *IdentityServer) listOrganizationCollaborators(
	ctx context.Context, req *ttnpb.ListOrganizationCollaboratorsRequest,
) (collaborators *ttnpb.Collaborators, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	err = rights.RequireOrganization(ctx, req.GetOrganizationIds(), ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_MEMBERS)
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
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		memberRights, err := st.FindMembers(ctx, req.GetOrganizationIds().GetEntityIdentifiers())
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

func (is *IdentityServer) deleteOrganizationCollaborator(
	ctx context.Context, req *ttnpb.DeleteOrganizationCollaboratorRequest,
) (*emptypb.Empty, error) {
	err := rights.RequireOrganization(ctx, req.GetOrganizationIds(), ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_MEMBERS)
	if err != nil {
		return ttnpb.Empty, err
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		r, err := st.GetMember(ctx, req.GetCollaboratorIds(), req.GetOrganizationIds().GetEntityIdentifiers())
		if err != nil {
			return err
		}
		org, err := st.GetOrganization(
			ctx,
			req.GetOrganizationIds(),
			store.FieldMask([]string{"administrative_contact", "technical_contact"}),
		)
		if err != nil {
			return err
		}
		if proto.Equal(org.GetAdministrativeContact(), req.GetCollaboratorIds()) ||
			proto.Equal(org.GetTechnicalContact(), req.GetCollaboratorIds()) {
			return errCollaboratorIsContact.WithAttributes("collaborator_id", req.GetCollaboratorIds().IDString())
		}
		if r.Implied().IncludesAll(ttnpb.Right_RIGHT_ORGANIZATION_ALL) {
			memberRights, err := st.FindMembers(ctx, req.GetOrganizationIds().GetEntityIdentifiers())
			if err != nil {
				return err
			}
			var hasOtherOwner bool
			for _, v := range memberRights {
				member, rights := v.Ids, v.Rights
				if unique.ID(ctx, member) == unique.ID(ctx, req.GetCollaboratorIds()) {
					continue
				}
				if rights.Implied().IncludesAll(ttnpb.Right_RIGHT_ORGANIZATION_ALL) {
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
			req.GetOrganizationIds().GetEntityIdentifiers(),
		)
	})
	if err != nil {
		return ttnpb.Empty, err
	}
	events.Publish(evtDeleteOrganizationCollaborator.New(
		ctx,
		events.WithIdentifiers(req.GetOrganizationIds(), req.GetCollaboratorIds()),
	))
	return ttnpb.Empty, nil
}

type organizationAccess struct {
	ttnpb.UnimplementedOrganizationAccessServer

	*IdentityServer
}

func (oa *organizationAccess) ListRights(
	ctx context.Context, req *ttnpb.OrganizationIdentifiers,
) (*ttnpb.Rights, error) {
	return oa.listOrganizationRights(ctx, req)
}

func (oa *organizationAccess) CreateAPIKey(
	ctx context.Context, req *ttnpb.CreateOrganizationAPIKeyRequest,
) (*ttnpb.APIKey, error) {
	return oa.createOrganizationAPIKey(ctx, req)
}

func (oa *organizationAccess) ListAPIKeys(
	ctx context.Context, req *ttnpb.ListOrganizationAPIKeysRequest,
) (*ttnpb.APIKeys, error) {
	return oa.listOrganizationAPIKeys(ctx, req)
}

func (oa *organizationAccess) GetAPIKey(
	ctx context.Context, req *ttnpb.GetOrganizationAPIKeyRequest,
) (*ttnpb.APIKey, error) {
	return oa.getOrganizationAPIKey(ctx, req)
}

func (oa *organizationAccess) UpdateAPIKey(
	ctx context.Context, req *ttnpb.UpdateOrganizationAPIKeyRequest,
) (*ttnpb.APIKey, error) {
	return oa.updateOrganizationAPIKey(ctx, req)
}

func (oa *organizationAccess) DeleteAPIKey(
	ctx context.Context, req *ttnpb.DeleteOrganizationAPIKeyRequest,
) (*emptypb.Empty, error) {
	return oa.deleteOrganizationAPIKey(ctx, req)
}

func (oa *organizationAccess) GetCollaborator(
	ctx context.Context, req *ttnpb.GetOrganizationCollaboratorRequest,
) (*ttnpb.GetCollaboratorResponse, error) {
	return oa.getOrganizationCollaborator(ctx, req)
}

func (oa *organizationAccess) SetCollaborator(
	ctx context.Context, req *ttnpb.SetOrganizationCollaboratorRequest,
) (*emptypb.Empty, error) {
	return oa.setOrganizationCollaborator(ctx, req)
}

func (oa *organizationAccess) ListCollaborators(
	ctx context.Context, req *ttnpb.ListOrganizationCollaboratorsRequest,
) (*ttnpb.Collaborators, error) {
	return oa.listOrganizationCollaborators(ctx, req)
}

func (oa *organizationAccess) DeleteCollaborator(
	ctx context.Context, req *ttnpb.DeleteOrganizationCollaboratorRequest,
) (*emptypb.Empty, error) {
	return oa.deleteOrganizationCollaborator(ctx, req)
}
