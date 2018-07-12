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

	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var errMissingOrganizationRights = errors.DefinePermissionDenied(
	"missing_organization_rights",
	"missing rights for organization `{uid}`",
)

// RequireOrganization checks that context contains the specified rights for the organization.
// This only works if the Rights hook previously added the rights on the context,
// using the auth data in the metadata, and the request data of the RPC.
func RequireOrganization(ctx context.Context, rights ...ttnpb.Right) error {
	if ad := fromContext(ctx); !ttnpb.IncludesRights(ad.rights, rights...) {
		return errMissingOrganizationRights.WithAttributes("uid", ad.uid)
	}
	return nil
}

var errMissingApplicationRights = errors.DefinePermissionDenied(
	"missing_application_rights",
	"missing rights for application `{uid}`",
)

// RequireApplication checks that context contains the specified rights for the application.
// This only works if the Rights hook previously added the rights on the context,
// using the auth data in the metadata, and the request data of the RPC.
func RequireApplication(ctx context.Context, rights ...ttnpb.Right) error {
	if ad := fromContext(ctx); !ttnpb.IncludesRights(ad.rights, rights...) {
		return errMissingApplicationRights.WithAttributes("uid", ad.uid)
	}
	return nil
}

var errMissingGatewayRights = errors.DefinePermissionDenied(
	"missing_gateway_rights",
	"missing rights for gateway `{uid}`",
)

// RequireGateway checks that context contains the specified rights for the gateway.
// This only works if the Rights hook previously added the rights on the context,
// using the auth data in the metadata, and the request data of the RPC.
func RequireGateway(ctx context.Context, rights ...ttnpb.Right) error {
	if ad := fromContext(ctx); !ttnpb.IncludesRights(ad.rights, rights...) {
		return errMissingGatewayRights.WithAttributes("uid", ad.uid)
	}
	return nil
}
