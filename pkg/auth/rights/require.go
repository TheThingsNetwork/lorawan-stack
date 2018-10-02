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

package rights

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

var errNoApplicationRights = errors.DefinePermissionDenied(
	"no_application_rights",
	"no rights for application `{uid} specified`",
)

var errNoClientRights = errors.DefinePermissionDenied(
	"no_client_rights",
	"no rights for client `{uid} specified`",
)

var errNoGatewayRights = errors.DefinePermissionDenied(
	"no_gateway_rights",
	"no rights for gateway `{uid}` specified",
)

var errNoOrganizationRights = errors.DefinePermissionDenied(
	"no_organization_rights",
	"no rights for organization `{uid}` specified",
)

var errNoUserRights = errors.DefinePermissionDenied(
	"no_user_rights",
	"no rights for user `{uid}` specified",
)

// RequireApplication checks that context contains the required rights for the
// given application ID.
func RequireApplication(ctx context.Context, appID ttnpb.ApplicationIdentifiers, required ...ttnpb.Right) error {
	appUID := unique.ID(ctx, appID)
	rights, ok := FromContext(ctx)
	if !ok {
		fetcher, ok := fetcherFromContext(ctx)
		if !ok {
			panic(errNoFetcher)
		}
		appRights, err := fetcher.ApplicationRights(ctx, appID)
		switch {
		case err == nil, errors.IsPermissionDenied(err):
			break
		default:
			return err
		}
		rights.ApplicationRights = map[string]*ttnpb.Rights{appUID: appRights}
	}
	if !rights.IncludesApplicationRights(appUID, required...) {
		return errNoApplicationRights.WithAttributes("uid", appUID)
	}
	return nil
}

// RequireClient checks that context contains the required rights for the
// given client ID.
func RequireClient(ctx context.Context, cliID ttnpb.ClientIdentifiers, required ...ttnpb.Right) error {
	cliUID := unique.ID(ctx, cliID)
	rights, ok := FromContext(ctx)
	if !ok {
		fetcher, ok := fetcherFromContext(ctx)
		if !ok {
			panic(errNoFetcher)
		}
		cliRights, err := fetcher.ClientRights(ctx, cliID)
		switch {
		case err == nil, errors.IsPermissionDenied(err):
			break
		default:
			return err
		}
		rights.ClientRights = map[string]*ttnpb.Rights{cliUID: cliRights}
	}
	if !rights.IncludesClientRights(cliUID, required...) {
		return errNoClientRights.WithAttributes("uid", cliUID)
	}
	return nil
}

// RequireGateway checks that context contains the required rights for the
// given gateway ID.
func RequireGateway(ctx context.Context, gtwID ttnpb.GatewayIdentifiers, required ...ttnpb.Right) error {
	gtwUID := unique.ID(ctx, gtwID)
	rights, ok := FromContext(ctx)
	if !ok {
		fetcher, ok := fetcherFromContext(ctx)
		if !ok {
			panic(errNoFetcher)
		}
		gtwRights, err := fetcher.GatewayRights(ctx, gtwID)
		switch {
		case err == nil, errors.IsPermissionDenied(err):
			break
		default:
			return err
		}
		rights.GatewayRights = map[string]*ttnpb.Rights{gtwUID: gtwRights}
	}
	if !rights.IncludesGatewayRights(gtwUID, required...) {
		return errNoGatewayRights.WithAttributes("uid", gtwUID)
	}
	return nil
}

// RequireOrganization checks that context contains the required rights for the
// given organization ID.
func RequireOrganization(ctx context.Context, orgID ttnpb.OrganizationIdentifiers, required ...ttnpb.Right) error {
	orgUID := unique.ID(ctx, orgID)
	rights, ok := FromContext(ctx)
	if !ok {
		fetcher, ok := fetcherFromContext(ctx)
		if !ok {
			panic(errNoFetcher)
		}
		orgRights, err := fetcher.OrganizationRights(ctx, orgID)
		switch {
		case err == nil, errors.IsPermissionDenied(err):
			break
		default:
			return err
		}
		rights.OrganizationRights = map[string]*ttnpb.Rights{orgUID: orgRights}
	}
	if !rights.IncludesOrganizationRights(orgUID, required...) {
		return errNoOrganizationRights.WithAttributes("uid", orgUID)
	}
	return nil
}

// RequireUser checks that context contains the required rights for the
// given user ID.
func RequireUser(ctx context.Context, usrID ttnpb.UserIdentifiers, required ...ttnpb.Right) error {
	usrUID := unique.ID(ctx, usrID)
	rights, ok := FromContext(ctx)
	if !ok {
		fetcher, ok := fetcherFromContext(ctx)
		if !ok {
			panic(errNoFetcher)
		}
		usrRights, err := fetcher.UserRights(ctx, usrID)
		switch {
		case err == nil, errors.IsPermissionDenied(err):
			break
		default:
			return err
		}
		rights.UserRights = map[string]*ttnpb.Rights{usrUID: usrRights}
	}
	if !rights.IncludesUserRights(usrUID, required...) {
		return errNoUserRights.WithAttributes("uid", usrUID)
	}
	return nil
}
