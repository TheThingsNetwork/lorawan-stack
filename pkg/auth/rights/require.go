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

package rights

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// Errors for no/insufficient rights.
var (
	ErrNoUniversalRights = errors.DefinePermissionDenied(
		"no_universal_rights",
		"no universal rights",
	)
	ErrInsufficientUniversalRights = errors.DefinePermissionDenied(
		"insufficient_universal_rights",
		"insufficient universal rights",
		"missing",
	)
	ErrNoAdmin = errors.DefinePermissionDenied(
		"no_admin",
		"no admin",
	)
	ErrNoApplicationRights = errors.DefinePermissionDenied(
		"no_application_rights",
		"no rights for application `{uid}`",
	)
	ErrInsufficientApplicationRights = errors.DefinePermissionDenied(
		"insufficient_application_rights",
		"insufficient rights for application `{uid}`",
		"missing",
	)
	ErrNoClientRights = errors.DefinePermissionDenied(
		"no_client_rights",
		"no rights for client `{uid}`",
	)
	ErrInsufficientClientRights = errors.DefinePermissionDenied(
		"insufficient_client_rights",
		"insufficient rights for client `{uid}`",
		"missing",
	)
	ErrNoGatewayRights = errors.DefinePermissionDenied(
		"no_gateway_rights",
		"no rights for gateway `{uid}`",
	)
	ErrInsufficientGatewayRights = errors.DefinePermissionDenied(
		"insufficient_gateway_rights",
		"insufficient rights for gateway `{uid}`",
		"missing",
	)
	ErrNoOrganizationRights = errors.DefinePermissionDenied(
		"no_organization_rights",
		"no rights for organization `{uid}`",
	)
	ErrInsufficientOrganizationRights = errors.DefinePermissionDenied(
		"insufficient_organization_rights",
		"insufficient rights for organization `{uid}`",
		"missing",
	)
	ErrNoUserRights = errors.DefinePermissionDenied(
		"no_user_rights",
		"no rights for user `{uid}`",
	)
	ErrInsufficientUserRights = errors.DefinePermissionDenied(
		"insufficient_user_rights",
		"insufficient rights for user `{uid}`",
		"missing",
	)
	ErrInsufficientRights = errors.DefinePermissionDenied(
		"insufficient_rights",
		"insufficient rights for {entity_type} `{uid}`",
		"missing",
	)
)

func rightsNames(rights ...ttnpb.Right) []string {
	names := make([]string, len(rights))
	for i, right := range rights {
		names[i] = right.String()
	}
	return names
}

// RequireUniversal checks that the context contains the required universal rights.
func RequireUniversal(ctx context.Context, required ...ttnpb.Right) error {
	authInfo, err := AuthInfo(ctx)
	if err != nil {
		return err
	}
	if rights := authInfo.GetUniversalRights(); len(rights.GetRights()) == 0 {
		return ErrNoUniversalRights.New()
	} else if missing := ttnpb.RightsFrom(required...).Sub(rights).GetRights(); len(missing) > 0 {
		return ErrInsufficientUniversalRights.WithAttributes("missing", rightsNames(missing...))
	}
	return nil
}

// RequireIsAdmin checks that the context is authenticated as admin.
func RequireIsAdmin(ctx context.Context) error {
	authInfo, err := AuthInfo(ctx)
	if err != nil {
		return err
	}
	if !authInfo.GetIsAdmin() {
		return ErrNoAdmin.New()
	}
	return nil
}

// RequireApplication checks that context contains the required rights for the
// given application ID.
func RequireApplication(ctx context.Context, id *ttnpb.ApplicationIdentifiers, required ...ttnpb.Right) error {
	uid := unique.ID(ctx, id)
	rights, err := ListApplication(ctx, id)
	if err != nil {
		return err
	}
	if len(rights.GetRights()) == 0 {
		return ErrNoApplicationRights.WithAttributes("uid", uid)
	}
	missing := ttnpb.RightsFrom(required...).Sub(rights).GetRights()
	if len(missing) > 0 {
		return ErrInsufficientApplicationRights.WithAttributes("uid", uid, "missing", rightsNames(missing...))
	}
	return nil
}

// RequireClient checks that context contains the required rights for the
// given client ID.
func RequireClient(ctx context.Context, id *ttnpb.ClientIdentifiers, required ...ttnpb.Right) (err error) {
	uid := unique.ID(ctx, id)
	rights, err := ListClient(ctx, id)
	if err != nil {
		return err
	}
	if len(rights.GetRights()) == 0 {
		return ErrNoClientRights.WithAttributes("uid", uid)
	}
	missing := ttnpb.RightsFrom(required...).Sub(rights).GetRights()
	if len(missing) > 0 {
		return ErrInsufficientClientRights.WithAttributes("uid", uid, "missing", rightsNames(missing...))
	}
	return nil
}

// RequireGateway checks that context contains the required rights for the
// given gateway ID.
func RequireGateway(ctx context.Context, id *ttnpb.GatewayIdentifiers, required ...ttnpb.Right) (err error) {
	uid := unique.ID(ctx, id)
	rights, err := ListGateway(ctx, id)
	if err != nil {
		return err
	}
	if len(rights.GetRights()) == 0 {
		return ErrNoGatewayRights.WithAttributes("uid", uid)
	}
	missing := ttnpb.RightsFrom(required...).Sub(rights).GetRights()
	if len(missing) > 0 {
		return ErrInsufficientGatewayRights.WithAttributes("uid", uid, "missing", rightsNames(missing...))
	}
	return nil
}

// RequireOrganization checks that context contains the required rights for the
// given organization ID.
func RequireOrganization(ctx context.Context, id *ttnpb.OrganizationIdentifiers, required ...ttnpb.Right) (err error) {
	uid := unique.ID(ctx, id)
	rights, err := ListOrganization(ctx, id)
	if err != nil {
		return err
	}
	if len(rights.GetRights()) == 0 {
		return ErrNoOrganizationRights.WithAttributes("uid", uid)
	}
	missing := ttnpb.RightsFrom(required...).Sub(rights).GetRights()
	if len(missing) > 0 {
		return ErrInsufficientOrganizationRights.WithAttributes("uid", uid, "missing", rightsNames(missing...))
	}
	return nil
}

// RequireUser checks that context contains the required rights for the
// given user ID.
func RequireUser(ctx context.Context, id *ttnpb.UserIdentifiers, required ...ttnpb.Right) (err error) {
	uid := unique.ID(ctx, id)
	rights, err := ListUser(ctx, id)
	if err != nil {
		return err
	}
	if len(rights.GetRights()) == 0 {
		return ErrNoUserRights.WithAttributes("uid", uid)
	}
	missing := ttnpb.RightsFrom(required...).Sub(rights).GetRights()
	if len(missing) > 0 {
		return ErrInsufficientUserRights.WithAttributes("uid", uid, "missing", rightsNames(missing...))
	}
	return nil
}

// RequireAny checks that context contains any rights for each of
// the given entity identifiers.
func RequireAny(ctx context.Context, ids ...*ttnpb.EntityIdentifiers) error {
	for _, entityIDs := range ids {
		switch ids := entityIDs.GetIds().(type) {
		case *ttnpb.EntityIdentifiers_ApplicationIds:
			list, err := ListApplication(ctx, ids.ApplicationIds)
			if err != nil {
				return err
			}
			if len(list.GetRights()) == 0 {
				return ErrNoApplicationRights.WithAttributes("uid", unique.ID(ctx, ids.ApplicationIds))
			}
		case *ttnpb.EntityIdentifiers_ClientIds:
			list, err := ListClient(ctx, ids.ClientIds)
			if err != nil {
				return err
			}
			if len(list.GetRights()) == 0 {
				return ErrNoClientRights.WithAttributes("uid", unique.ID(ctx, ids.ClientIds))
			}
		case *ttnpb.EntityIdentifiers_DeviceIds:
			list, err := ListApplication(ctx, ids.DeviceIds.ApplicationIds)
			if err != nil {
				return err
			}
			if len(list.GetRights()) == 0 {
				return ErrNoApplicationRights.WithAttributes("uid", unique.ID(ctx, ids.DeviceIds.ApplicationIds))
			}
		case *ttnpb.EntityIdentifiers_GatewayIds:
			list, err := ListGateway(ctx, ids.GatewayIds)
			if err != nil {
				return err
			}
			if len(list.GetRights()) == 0 {
				return ErrNoGatewayRights.WithAttributes("uid", unique.ID(ctx, ids.GatewayIds))
			}
		case *ttnpb.EntityIdentifiers_OrganizationIds:
			list, err := ListOrganization(ctx, ids.OrganizationIds)
			if err != nil {
				return err
			}
			if len(list.GetRights()) == 0 {
				return ErrNoOrganizationRights.WithAttributes("uid", unique.ID(ctx, ids.OrganizationIds))
			}
		case *ttnpb.EntityIdentifiers_UserIds:
			list, err := ListUser(ctx, ids.UserIds)
			if err != nil {
				return err
			}
			if len(list.GetRights()) == 0 {
				return ErrNoUserRights.WithAttributes("uid", unique.ID(ctx, ids.UserIds))
			}
		}
	}
	return nil
}
