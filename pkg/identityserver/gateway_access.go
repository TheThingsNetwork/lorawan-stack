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
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	evtCreateGatewayAPIKey = events.Define(
		"gateway.api-key.create", "create gateway API key",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_SETTINGS_API_KEYS),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUpdateGatewayAPIKey = events.Define(
		"gateway.api-key.update", "update gateway API key",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_SETTINGS_API_KEYS),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteGatewayAPIKey = events.Define(
		"gateway.api-key.delete", "delete gateway API key",
		events.WithVisibility(ttnpb.Right_RIGHT_GATEWAY_SETTINGS_API_KEYS),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtUpdateGatewayCollaborator = events.Define(
		"gateway.collaborator.update", "update gateway collaborator",
		events.WithVisibility(
			ttnpb.Right_RIGHT_GATEWAY_SETTINGS_COLLABORATORS,
			ttnpb.Right_RIGHT_USER_GATEWAYS_LIST,
		),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
	evtDeleteGatewayCollaborator = events.Define(
		"gateway.collaborator.delete", "delete gateway collaborator",
		events.WithVisibility(
			ttnpb.Right_RIGHT_GATEWAY_SETTINGS_COLLABORATORS,
			ttnpb.Right_RIGHT_USER_GATEWAYS_LIST,
		),
		events.WithAuthFromContext(),
		events.WithClientInfoFromContext(),
	)
)

func (*IdentityServer) listGatewayRights(ctx context.Context, ids *ttnpb.GatewayIdentifiers) (*ttnpb.Rights, error) {
	gtwRights, err := rights.ListGateway(ctx, ids)
	if err != nil {
		return nil, err
	}
	return gtwRights.Intersect(ttnpb.AllGatewayRights), nil
}

func (is *IdentityServer) createGatewayAPIKey(
	ctx context.Context, req *ttnpb.CreateGatewayAPIKeyRequest,
) (key *ttnpb.APIKey, err error) {
	// Require that caller has rights to manage API keys.
	if err = rights.RequireGateway(ctx, req.GetGatewayIds(), ttnpb.Right_RIGHT_GATEWAY_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}
	// Require that caller has at least the rights of the API key.
	if err = rights.RequireGateway(ctx, req.GetGatewayIds(), req.Rights...); err != nil {
		return nil, err
	}
	key, token, err := GenerateAPIKey(ctx, req.Name, ttnpb.StdTime(req.ExpiresAt), req.Rights...)
	if err != nil {
		return nil, err
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		key, err = st.CreateAPIKey(ctx, req.GetGatewayIds().GetEntityIdentifiers(), key)
		return err
	})
	if err != nil {
		return nil, err
	}
	key.Key = ""

	events.Publish(evtCreateGatewayAPIKey.NewWithIdentifiersAndData(ctx, req.GetGatewayIds(), key))
	go is.notifyInternal(ctx, &ttnpb.CreateNotificationRequest{
		EntityIds:        req.GetGatewayIds().GetEntityIdentifiers(),
		NotificationType: "api_key_created",
		Data:             ttnpb.MustMarshalAny(key),
		Receivers: []ttnpb.NotificationReceiver{
			ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_ADMINISTRATIVE_CONTACT,
		},
		Email: true,
	})

	key.Key = token
	return key, nil
}

func (is *IdentityServer) listGatewayAPIKeys(
	ctx context.Context, req *ttnpb.ListGatewayAPIKeysRequest,
) (keys *ttnpb.APIKeys, err error) {
	if err = rights.RequireGateway(ctx, req.GetGatewayIds(), ttnpb.Right_RIGHT_GATEWAY_SETTINGS_API_KEYS); err != nil {
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
		keys.ApiKeys, err = st.FindAPIKeys(ctx, req.GetGatewayIds().GetEntityIdentifiers())
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

func (is *IdentityServer) getGatewayAPIKey(
	ctx context.Context, req *ttnpb.GetGatewayAPIKeyRequest,
) (key *ttnpb.APIKey, err error) {
	if err = rights.RequireGateway(ctx, req.GetGatewayIds(), ttnpb.Right_RIGHT_GATEWAY_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}

	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		key, err = st.GetAPIKey(ctx, req.GetGatewayIds().GetEntityIdentifiers(), req.KeyId)
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

func (is *IdentityServer) updateGatewayAPIKey(
	ctx context.Context, req *ttnpb.UpdateGatewayAPIKeyRequest,
) (key *ttnpb.APIKey, err error) {
	// Require that caller has rights to manage API keys.
	if err = rights.RequireGateway(ctx, req.GetGatewayIds(), ttnpb.Right_RIGHT_GATEWAY_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}

	// Backwards compatibility for older clients.
	if len(req.FieldMask.GetPaths()) == 0 {
		req.FieldMask = ttnpb.FieldMask("rights", "name")
	}

	apiKey := req.GetApiKey()
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		if len(apiKey.Rights) > 0 {
			key, err := st.GetAPIKey(ctx, req.GetGatewayIds().GetEntityIdentifiers(), apiKey.Id)
			if err != nil {
				return err
			}

			newRights := ttnpb.RightsFrom(apiKey.Rights...)
			existingRights := ttnpb.RightsFrom(key.Rights...)

			// Require the caller to have all added rights.
			err = rights.RequireGateway(ctx, req.GetGatewayIds(), newRights.Sub(existingRights).GetRights()...)
			if err != nil {
				return err
			}
			// Require the caller to have all removed rights.
			err = rights.RequireGateway(ctx, req.GetGatewayIds(), existingRights.Sub(newRights).GetRights()...)
			if err != nil {
				return err
			}
		}

		if len(req.ApiKey.Rights) == 0 && ttnpb.HasAnyField(req.GetFieldMask().GetPaths(), "rights") {
			// TODO: Remove delete capability (https://github.com/TheThingsNetwork/lorawan-stack/issues/6488).
			return st.DeleteAPIKey(ctx, req.GetGatewayIds().GetEntityIdentifiers(), req.ApiKey)
		}

		key, err = st.UpdateAPIKey(ctx, req.GetGatewayIds().GetEntityIdentifiers(), apiKey, req.FieldMask.GetPaths())
		return err
	})
	if err != nil {
		return nil, err
	}
	if key == nil { // API key was deleted.
		events.Publish(evtDeleteGatewayAPIKey.NewWithIdentifiersAndData(ctx, req.GetGatewayIds(), nil))
		return &ttnpb.APIKey{}, nil
	}
	key.Key = ""

	events.Publish(evtUpdateGatewayAPIKey.NewWithIdentifiersAndData(ctx, req.GetGatewayIds(), key))
	go is.notifyInternal(ctx, &ttnpb.CreateNotificationRequest{
		EntityIds:        req.GetGatewayIds().GetEntityIdentifiers(),
		NotificationType: "api_key_changed",
		Data:             ttnpb.MustMarshalAny(key),
		Receivers: []ttnpb.NotificationReceiver{
			ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_ADMINISTRATIVE_CONTACT,
		},
		Email: true,
	})

	return key, nil
}

func (is *IdentityServer) deleteGatewayAPIKey(
	ctx context.Context, req *ttnpb.DeleteGatewayAPIKeyRequest,
) (*emptypb.Empty, error) {
	// Require that caller has rights to manage API keys.
	if err := rights.RequireGateway(ctx, req.GetGatewayIds(), ttnpb.Right_RIGHT_GATEWAY_SETTINGS_API_KEYS); err != nil {
		return nil, err
	}

	err := is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		return st.DeleteAPIKey(ctx, req.GetGatewayIds().GetEntityIdentifiers(), &ttnpb.APIKey{Id: req.KeyId})
	})
	if err != nil {
		return nil, err
	}
	events.Publish(evtDeleteGatewayAPIKey.New(ctx, events.WithIdentifiers(req.GetGatewayIds())))
	return ttnpb.Empty, nil
}

func (is *IdentityServer) getGatewayCollaborator(
	ctx context.Context, req *ttnpb.GetGatewayCollaboratorRequest,
) (*ttnpb.GetCollaboratorResponse, error) {
	err := rights.RequireGateway(ctx, req.GetGatewayIds(), ttnpb.Right_RIGHT_GATEWAY_SETTINGS_COLLABORATORS)
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
			req.GetGatewayIds().GetEntityIdentifiers(),
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

var errGatewayNeedsCollaborator = errors.DefineFailedPrecondition(
	"gateway_needs_collaborator", "every gateway needs at least one collaborator with all rights",
)

func (is *IdentityServer) setGatewayCollaborator(
	ctx context.Context, req *ttnpb.SetGatewayCollaboratorRequest,
) (_ *emptypb.Empty, err error) {
	// Require that caller has rights to manage collaborators.
	err = rights.RequireGateway(ctx, req.GetGatewayIds(), ttnpb.Right_RIGHT_GATEWAY_SETTINGS_COLLABORATORS)
	if err != nil {
		return nil, err
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		existingRights, err := st.GetMember(
			ctx,
			req.GetCollaborator().GetIds(),
			req.GetGatewayIds().GetEntityIdentifiers(),
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
			if err := rights.RequireGateway(ctx, req.GetGatewayIds(), addedRights.GetRights()...); err != nil {
				return err
			}
		}

		// Unless we're deleting the collaborator, require the caller to have all removed rights.
		if len(newRights.GetRights()) > 0 && len(removedRights.GetRights()) > 0 {
			if err := rights.RequireGateway(ctx, req.GetGatewayIds(), removedRights.GetRights()...); err != nil {
				return err
			}
		}

		if removedRights.IncludesAll(ttnpb.Right_RIGHT_GATEWAY_ALL) {
			memberRights, err := st.FindMembers(ctx, req.GetGatewayIds().GetEntityIdentifiers())
			if err != nil {
				return err
			}
			var hasOtherOwner bool
			for _, v := range memberRights {
				member, rights := v.Ids, v.Rights
				if unique.ID(ctx, member) == unique.ID(ctx, req.GetCollaborator().GetIds()) {
					continue
				}
				if rights.Implied().IncludesAll(ttnpb.Right_RIGHT_GATEWAY_ALL) {
					hasOtherOwner = true
					break
				}
			}
			if !hasOtherOwner {
				return errGatewayNeedsCollaborator.New()
			}
		}

		if len(req.GetCollaborator().GetRights()) == 0 {
			// TODO: Remove delete capability (https://github.com/TheThingsNetwork/lorawan-stack/issues/6488).
			return st.DeleteMember(ctx, req.GetCollaborator().GetIds(), req.GetGatewayIds().GetEntityIdentifiers())
		}

		return st.SetMember(
			ctx,
			req.GetCollaborator().GetIds(),
			req.GetGatewayIds().GetEntityIdentifiers(),
			ttnpb.RightsFrom(req.GetCollaborator().GetRights()...),
		)
	})
	if err != nil {
		return nil, err
	}
	if len(req.GetCollaborator().GetRights()) > 0 {
		events.Publish(evtUpdateGatewayCollaborator.New(
			ctx,
			events.WithIdentifiers(req.GetGatewayIds(), req.GetCollaborator().GetIds()),
			events.WithData(req.GetCollaborator()),
		))
		go is.notifyInternal(ctx, &ttnpb.CreateNotificationRequest{
			EntityIds:        req.GetGatewayIds().GetEntityIdentifiers(),
			NotificationType: "collaborator_changed",
			Data:             ttnpb.MustMarshalAny(req.GetCollaborator()),
			Receivers: []ttnpb.NotificationReceiver{
				ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_ADMINISTRATIVE_CONTACT,
			},
			Email: false,
		})
	} else {
		events.Publish(evtDeleteGatewayCollaborator.New(
			ctx,
			events.WithIdentifiers(req.GetGatewayIds(), req.GetCollaborator().GetIds()),
		))
	}
	return ttnpb.Empty, nil
}

func (is *IdentityServer) listGatewayCollaborators(
	ctx context.Context, req *ttnpb.ListGatewayCollaboratorsRequest,
) (collaborators *ttnpb.Collaborators, err error) {
	if err = is.RequireAuthenticated(ctx); err != nil {
		return nil, err
	}
	err = rights.RequireGateway(ctx, req.GetGatewayIds(), ttnpb.Right_RIGHT_GATEWAY_SETTINGS_COLLABORATORS)
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
		memberRights, err := st.FindMembers(ctx, req.GetGatewayIds().GetEntityIdentifiers())
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

func (is *IdentityServer) deleteGatewayCollaborator(
	ctx context.Context, req *ttnpb.DeleteGatewayCollaboratorRequest,
) (*emptypb.Empty, error) {
	err := rights.RequireGateway(ctx, req.GetGatewayIds(), ttnpb.Right_RIGHT_GATEWAY_SETTINGS_COLLABORATORS)
	if err != nil {
		return ttnpb.Empty, err
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		removedRights, err := st.GetMember(ctx, req.GetCollaboratorIds(), req.GetGatewayIds().GetEntityIdentifiers())
		if err != nil {
			return err
		}
		if removedRights.Implied().IncludesAll(ttnpb.Right_RIGHT_GATEWAY_ALL) {
			memberRights, err := st.FindMembers(ctx, req.GetGatewayIds().GetEntityIdentifiers())
			if err != nil {
				return err
			}
			var hasOtherOwner bool
			for _, v := range memberRights {
				member, rights := v.Ids, v.Rights
				if unique.ID(ctx, member) == unique.ID(ctx, req.GetCollaboratorIds()) {
					continue
				}
				if rights.Implied().IncludesAll(ttnpb.Right_RIGHT_GATEWAY_ALL) {
					hasOtherOwner = true
					break
				}
			}
			if !hasOtherOwner {
				return errGatewayNeedsCollaborator.New()
			}
		}

		return st.DeleteMember(
			ctx,
			req.GetCollaboratorIds(),
			req.GetGatewayIds().GetEntityIdentifiers(),
		)
	})
	if err != nil {
		return ttnpb.Empty, err
	}
	events.Publish(evtDeleteGatewayCollaborator.New(
		ctx,
		events.WithIdentifiers(req.GetGatewayIds(), req.GetCollaboratorIds()),
	))
	return ttnpb.Empty, nil
}

type gatewayAccess struct {
	ttnpb.UnimplementedGatewayAccessServer

	*IdentityServer
}

func (ga *gatewayAccess) ListRights(
	ctx context.Context, req *ttnpb.GatewayIdentifiers,
) (*ttnpb.Rights, error) {
	return ga.listGatewayRights(ctx, req)
}

func (ga *gatewayAccess) CreateAPIKey(
	ctx context.Context, req *ttnpb.CreateGatewayAPIKeyRequest,
) (*ttnpb.APIKey, error) {
	return ga.createGatewayAPIKey(ctx, req)
}

func (ga *gatewayAccess) ListAPIKeys(
	ctx context.Context, req *ttnpb.ListGatewayAPIKeysRequest,
) (*ttnpb.APIKeys, error) {
	return ga.listGatewayAPIKeys(ctx, req)
}

func (ga *gatewayAccess) GetAPIKey(
	ctx context.Context, req *ttnpb.GetGatewayAPIKeyRequest,
) (*ttnpb.APIKey, error) {
	return ga.getGatewayAPIKey(ctx, req)
}

func (ga *gatewayAccess) UpdateAPIKey(
	ctx context.Context, req *ttnpb.UpdateGatewayAPIKeyRequest,
) (*ttnpb.APIKey, error) {
	return ga.updateGatewayAPIKey(ctx, req)
}

func (ga *gatewayAccess) DeleteAPIKey(
	ctx context.Context, req *ttnpb.DeleteGatewayAPIKeyRequest,
) (*emptypb.Empty, error) {
	return ga.deleteGatewayAPIKey(ctx, req)
}

func (ga *gatewayAccess) GetCollaborator(
	ctx context.Context, req *ttnpb.GetGatewayCollaboratorRequest,
) (*ttnpb.GetCollaboratorResponse, error) {
	return ga.getGatewayCollaborator(ctx, req)
}

func (ga *gatewayAccess) SetCollaborator(
	ctx context.Context, req *ttnpb.SetGatewayCollaboratorRequest,
) (*emptypb.Empty, error) {
	return ga.setGatewayCollaborator(ctx, req)
}

func (ga *gatewayAccess) ListCollaborators(
	ctx context.Context, req *ttnpb.ListGatewayCollaboratorsRequest,
) (*ttnpb.Collaborators, error) {
	return ga.listGatewayCollaborators(ctx, req)
}

func (ga *gatewayAccess) DeleteCollaborator(
	ctx context.Context, req *ttnpb.DeleteGatewayCollaboratorRequest,
) (*emptypb.Empty, error) {
	return ga.deleteGatewayCollaborator(ctx, req)
}

type gatewayBatchAccess struct {
	ttnpb.UnimplementedGatewayBatchAccessServer

	*IdentityServer
}

var (
	errEmptyRequest = errors.DefineInvalidArgument(
		"empty_request",
		"empty request",
	)
	errNonGatewayRights = errors.DefineInvalidArgument(
		"non_gateway_rights",
		"non-gateway rights in request",
	)
)

// AssertRights implements ttnpb.GatewayBatchAccessServer.
func (gba *gatewayBatchAccess) AssertRights(
	ctx context.Context,
	req *ttnpb.AssertGatewayRightsRequest,
) (*emptypb.Empty, error) {
	// Sanitize request.
	required := req.Required.Unique()
	if len(required.GetRights()) == 0 {
		return nil, errEmptyRequest.New()
	}

	// Check that the request is checking only gateway rights.
	if !ttnpb.AllGatewayRights.IncludesAll(required.GetRights()...) {
		return nil, errNonGatewayRights.New()
	}

	err := gba.assertGatewayRights(ctx, req.GatewayIds, required)
	if err != nil {
		return nil, err
	}

	return ttnpb.Empty, nil
}
