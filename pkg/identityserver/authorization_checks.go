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

package identityserver

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/auth"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

// enforceUserRights is a hook that checks whether if the given authorization
// credentials are allowed to perform an action given the set of passed rights and
// returns the User ID attached to the credentials.
func (is *IdentityServer) enforceUserRights(ctx context.Context, rights ...ttnpb.Right) error {
	ad := authorizationDataFromContext(ctx)

	if ad.UserIdentifiers().IsZero() {
		return ErrNotAuthorized.New(nil)
	}

	if !ad.HasRights(rights...) {
		return ErrNotAuthorized.New(nil)
	}

	return nil
}

// enforceAdmin checks whether the given credentials are enough to access an admin resource.
func (is *IdentityServer) enforceAdmin(ctx context.Context) error {
	err := is.enforceUserRights(ctx, ttnpb.RIGHT_USER_ADMIN)
	if err != nil {
		return err
	}

	found, err := is.store.Users.GetByID(authorizationDataFromContext(ctx).UserIdentifiers(), is.specializers.User)
	if err != nil {
		return err
	}

	if !found.GetUser().Admin {
		return ErrNotAuthorized.New(nil)
	}

	return nil
}

// enforceApplicationRights is a hook that checks whether if the given authorization
// credentials are allowed to access the application with the given rights.
func (is *IdentityServer) enforceApplicationRights(ctx context.Context, ids ttnpb.ApplicationIdentifiers, rights ...ttnpb.Right) error {
	ad := authorizationDataFromContext(ctx)

	if !ad.HasRights(rights...) {
		return ErrNotAuthorized.New(nil)
	}

	var authorized bool
	switch ad.Source {
	case auth.Key:
		kids := ad.ApplicationIdentifiers()
		if kids.IsZero() {
			break
		}

		authorized = kids.Contains(ids)
	case auth.Token:
		uids := ad.UserIdentifiers()
		if uids.IsZero() {
			break
		}

		var err error
		authorized, err = is.store.Applications.HasCollaboratorRights(ids, organizationOrUserIDsUserIDs(uids), rights...)
		if err != nil {
			return err
		}
	}

	if !authorized {
		return ErrNotAuthorized.New(nil)
	}

	return nil
}

// enforceGatewayRights is a hook that checks whether if the given authorization
// credentials are allowed to access the gateway with the given rights.
func (is *IdentityServer) enforceGatewayRights(ctx context.Context, ids ttnpb.GatewayIdentifiers, rights ...ttnpb.Right) error {
	ad := authorizationDataFromContext(ctx)

	if !ad.HasRights(rights...) {
		return ErrNotAuthorized.New(nil)
	}

	var authorized bool
	switch ad.Source {
	case auth.Key:
		kids := ad.GatewayIdentifiers()
		if kids.IsZero() {
			break
		}

		authorized = kids.Contains(ids)
	case auth.Token:
		uids := ad.UserIdentifiers()
		if uids.IsZero() {
			break
		}

		var err error
		authorized, err = is.store.Gateways.HasCollaboratorRights(ids, organizationOrUserIDsUserIDs(uids), rights...)
		if err != nil {
			return err
		}
	}

	if !authorized {
		return ErrNotAuthorized.New(nil)
	}

	return nil
}

// enforceOrganizationRights is a hook that checks whether if the given authorization
// credentials are allowed to access the organization with the given rights.
func (is *IdentityServer) enforceOrganizationRights(ctx context.Context, ids ttnpb.OrganizationIdentifiers, rights ...ttnpb.Right) error {
	ad := authorizationDataFromContext(ctx)

	if !ad.HasRights(rights...) {
		return ErrNotAuthorized.New(nil)
	}

	var authorized bool
	switch ad.Source {
	case auth.Key:
		kids := ad.OrganizationIdentifiers()
		if kids.IsZero() {
			break
		}

		authorized = kids.Contains(ids)
	case auth.Token:
		uids := ad.UserIdentifiers()
		if uids.IsZero() {
			break
		}

		var err error
		authorized, err = is.store.Organizations.HasMemberRights(ids, uids, rights...)
		if err != nil {
			return err
		}
	}

	if !authorized {
		return ErrNotAuthorized.New(nil)
	}

	return nil
}
