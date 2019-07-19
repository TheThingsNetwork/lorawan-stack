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

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

var errNoApplicationRights = errors.DefinePermissionDenied(
	"no_application_rights",
	"no rights for application `{uid}`",
)

var errInsufficientApplicationRights = errors.DefinePermissionDenied(
	"insufficient_application_rights",
	"insufficient rights for application `{uid}`",
)

var errNoClientRights = errors.DefinePermissionDenied(
	"no_client_rights",
	"no rights for client `{uid}`",
)

var errInsufficientClientRights = errors.DefinePermissionDenied(
	"insufficient_client_rights",
	"insufficient rights for client `{uid}`",
)

var errNoGatewayRights = errors.DefinePermissionDenied(
	"no_gateway_rights",
	"no rights for gateway `{uid}`",
)

var errInsufficientGatewayRights = errors.DefinePermissionDenied(
	"insufficient_gateway_rights",
	"insufficient rights for gateway `{uid}`",
)

var errNoOrganizationRights = errors.DefinePermissionDenied(
	"no_organization_rights",
	"no rights for organization `{uid}`",
)

var errInsufficientOrganizationRights = errors.DefinePermissionDenied(
	"insufficient_organization_rights",
	"insufficient rights for organization `{uid}`",
)

var errNoUserRights = errors.DefinePermissionDenied(
	"no_user_rights",
	"no rights for user `{uid}`",
)

var errInsufficientUserRights = errors.DefinePermissionDenied(
	"insufficient_user_rights",
	"insufficient rights for user `{uid}`",
)

// RequireApplication checks that context contains the required rights for the
// given application ID.
func RequireApplication(ctx context.Context, id ttnpb.ApplicationIdentifiers, required ...ttnpb.Right) (err error) {
	uid := unique.ID(ctx, id)
	rights, err := ListApplication(ctx, id)
	if err != nil {
		return err
	}
	if len(rights.GetRights()) == 0 {
		return errNoApplicationRights.WithAttributes("uid", uid)
	}
	missing := ttnpb.RightsFrom(required...).Sub(rights).GetRights()
	if len(missing) > 0 {
		return errInsufficientApplicationRights.WithAttributes("uid", uid, "missing", missing)
	}
	return nil
}

// RequireClient checks that context contains the required rights for the
// given client ID.
func RequireClient(ctx context.Context, id ttnpb.ClientIdentifiers, required ...ttnpb.Right) (err error) {
	uid := unique.ID(ctx, id)
	rights, err := ListClient(ctx, id)
	if err != nil {
		return err
	}
	if len(rights.GetRights()) == 0 {
		return errNoClientRights.WithAttributes("uid", uid)
	}
	missing := ttnpb.RightsFrom(required...).Sub(rights).GetRights()
	if len(missing) > 0 {
		return errInsufficientClientRights.WithAttributes("uid", uid, "missing", missing)
	}
	return nil
}

// RequireGateway checks that context contains the required rights for the
// given gateway ID.
func RequireGateway(ctx context.Context, id ttnpb.GatewayIdentifiers, required ...ttnpb.Right) (err error) {
	uid := unique.ID(ctx, id)
	rights, err := ListGateway(ctx, id)
	if err != nil {
		return err
	}
	if len(rights.GetRights()) == 0 {
		return errNoGatewayRights.WithAttributes("uid", uid)
	}
	missing := ttnpb.RightsFrom(required...).Sub(rights).GetRights()
	if len(missing) > 0 {
		return errInsufficientGatewayRights.WithAttributes("uid", uid, "missing", missing)
	}
	return nil
}

// RequireOrganization checks that context contains the required rights for the
// given organization ID.
func RequireOrganization(ctx context.Context, id ttnpb.OrganizationIdentifiers, required ...ttnpb.Right) (err error) {
	uid := unique.ID(ctx, id)
	rights, err := ListOrganization(ctx, id)
	if err != nil {
		return err
	}
	if len(rights.GetRights()) == 0 {
		return errNoOrganizationRights.WithAttributes("uid", uid)
	}
	missing := ttnpb.RightsFrom(required...).Sub(rights).GetRights()
	if len(missing) > 0 {
		return errInsufficientOrganizationRights.WithAttributes("uid", uid, "missing", missing)
	}
	return nil
}

// RequireUser checks that context contains the required rights for the
// given user ID.
func RequireUser(ctx context.Context, id ttnpb.UserIdentifiers, required ...ttnpb.Right) (err error) {
	uid := unique.ID(ctx, id)
	rights, err := ListUser(ctx, id)
	if err != nil {
		return err
	}
	if len(rights.GetRights()) == 0 {
		return errNoUserRights.WithAttributes("uid", uid)
	}
	missing := ttnpb.RightsFrom(required...).Sub(rights).GetRights()
	if len(missing) > 0 {
		return errInsufficientUserRights.WithAttributes("uid", uid, "missing", missing)
	}
	return nil
}
