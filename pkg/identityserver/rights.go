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

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func allPotentialRights(eIDs *ttnpb.EntityIdentifiers, rights *ttnpb.Rights) *ttnpb.Rights {
	switch eIDs.GetIds().(type) {
	case *ttnpb.EntityIdentifiers_ApplicationIds:
		return ttnpb.AllApplicationRights.Intersect(rights)
	case *ttnpb.EntityIdentifiers_ClientIds:
		return ttnpb.AllClientRights.Intersect(rights)
	case *ttnpb.EntityIdentifiers_GatewayIds:
		return ttnpb.AllGatewayRights.Intersect(rights)
	case *ttnpb.EntityIdentifiers_OrganizationIds:
		return ttnpb.AllEntityRights.Union(ttnpb.AllOrganizationRights).Intersect(rights)
	case *ttnpb.EntityIdentifiers_UserIds:
		return ttnpb.AllEntityRights.Union(ttnpb.AllOrganizationRights, ttnpb.AllUserRights).Intersect(rights)
	}
	return nil
}

func (is *IdentityServer) getRights(ctx context.Context, entityID *ttnpb.EntityIdentifiers) (entityRights, universalRights *ttnpb.Rights, err error) {
	authInfo, err := is.authInfo(ctx)
	if err != nil {
		return nil, nil, err
	}

	authInfoRights := ttnpb.RightsFrom(authInfo.GetRights()...)
	universalRights = allPotentialRights(entityID, authInfo.GetUniversalRights())
	if len(universalRights.GetRights()) == 0 {
		universalRights = nil
	}
	allPotentialRights := allPotentialRights(entityID, authInfoRights)

	// If the rights of the auth do not contain any rights for the entity type,
	// there's nothing more to do.
	if len(allPotentialRights.GetRights()) == 0 {
		return nil, universalRights, nil
	}

	// If the caller is the requested entity,
	// we can directly return the rights of the auth.
	authenticatedAs := authInfo.GetEntityIdentifiers()
	if entityID.EntityType() == authenticatedAs.EntityType() &&
		entityID.IDString() == authenticatedAs.IDString() {
		return authInfoRights, universalRights, nil
	}

	// If the caller is not an organization or user, there's nothing more to do.
	ouID := authInfo.GetOrganizationOrUserIdentifiers()
	if ouID == nil {
		return nil, universalRights, nil
	}

	// If the caller is requesting a user, and they're not that user (see above),
	// they don't have rights on it, so nothing more to do.
	if entityID.GetUserIds() != nil {
		return nil, universalRights, nil
	}

	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		membershipChains, err := st.FindAccountMembershipChains(ctx, ouID, entityID.EntityType(), entityID.IDString())
		if err != nil {
			return err
		}
		for _, chain := range membershipChains {
			entityRights = entityRights.Union(chain.GetRights())
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	entityRights = entityRights.Intersect(authInfoRights)

	return entityRights, universalRights, err
}

// ApplicationRights returns the rights the caller has on the given application.
func (is *IdentityServer) ApplicationRights(ctx context.Context, appIDs *ttnpb.ApplicationIdentifiers) (*ttnpb.Rights, error) {
	entity, universal, err := is.getRights(ctx, appIDs.GetEntityIdentifiers())
	if err != nil {
		return nil, err
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		_, err := st.GetApplication(ctx, appIDs, []string{"ids"})
		return err
	})
	if errors.IsNotFound(err) {
		if is.IsAdmin(ctx) {
			return nil, err
		}
		return &ttnpb.Rights{}, nil
	} else if err != nil {
		return nil, err
	}
	return entity.Union(universal), nil
}

// ClientRights returns the rights the caller has on the given client.
func (is *IdentityServer) ClientRights(ctx context.Context, cliIDs *ttnpb.ClientIdentifiers) (*ttnpb.Rights, error) {
	entity, universal, err := is.getRights(ctx, cliIDs.GetEntityIdentifiers())
	if err != nil {
		return nil, err
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		_, err := st.GetClient(ctx, cliIDs, []string{"ids"})
		return err
	})
	if errors.IsNotFound(err) {
		if is.IsAdmin(ctx) {
			return nil, err
		}
		return &ttnpb.Rights{}, nil
	} else if err != nil {
		return nil, err
	}
	return entity.Union(universal), nil
}

// GatewayRights returns the rights the caller has on the given gateway.
// The query for the gateway only considers the Gateway ID and not the EUI (if provided).
func (is *IdentityServer) GatewayRights(ctx context.Context, gtwIDs *ttnpb.GatewayIdentifiers) (*ttnpb.Rights, error) {
	ids := &ttnpb.GatewayIdentifiers{
		GatewayId: gtwIDs.GatewayId,
	}
	entity, universal, err := is.getRights(ctx, ids.GetEntityIdentifiers())
	if err != nil {
		return nil, err
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) (err error) {
		gtw, err := st.GetGateway(ctx, ids, []string{"ids", "status_public", "location_public"})
		if err != nil {
			return err
		}
		if gtw.StatusPublic {
			entity = entity.Union(ttnpb.RightsFrom(ttnpb.Right_RIGHT_GATEWAY_STATUS_READ))
		}
		if gtw.LocationPublic {
			entity = entity.Union(ttnpb.RightsFrom(ttnpb.Right_RIGHT_GATEWAY_LOCATION_READ))
		}
		return nil
	})
	if errors.IsNotFound(err) {
		if is.IsAdmin(ctx) {
			return nil, err
		}
		return &ttnpb.Rights{}, nil
	} else if err != nil {
		return nil, err
	}
	return entity.Union(universal), nil
}

// OrganizationRights returns the rights the caller has on the given organization.
func (is *IdentityServer) OrganizationRights(ctx context.Context, orgIDs *ttnpb.OrganizationIdentifiers) (*ttnpb.Rights, error) {
	entity, universal, err := is.getRights(ctx, orgIDs.GetEntityIdentifiers())
	if err != nil {
		return nil, err
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		_, err := st.GetOrganization(ctx, orgIDs, []string{"ids"})
		return err
	})
	if errors.IsNotFound(err) {
		if is.IsAdmin(ctx) {
			return nil, err
		}
		return &ttnpb.Rights{}, nil
	} else if err != nil {
		return nil, err
	}
	return entity.Union(universal), nil
}

// UserRights returns the rights the caller has on the given user.
func (is *IdentityServer) UserRights(ctx context.Context, userIDs *ttnpb.UserIdentifiers) (*ttnpb.Rights, error) {
	entity, universal, err := is.getRights(ctx, userIDs.GetEntityIdentifiers())
	if err != nil {
		return nil, err
	}
	err = is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		_, err := st.GetUser(ctx, userIDs, []string{"ids"})
		return err
	})
	if errors.IsNotFound(err) {
		if is.IsAdmin(ctx) {
			return nil, err
		}
		return &ttnpb.Rights{}, nil
	} else if err != nil {
		return nil, err
	}
	return entity.Union(universal), nil
}

var (
	errInsufficientRights = errors.DefinePermissionDenied(
		"insufficient_rights",
		"insufficient rights",
	)
	errSomeGatewaysNotFound = errors.DefineNotFound(
		"some_gateways_not_found",
		"some gateways not found",
	)
)

func (is *IdentityServer) assertGatewayRights(
	ctx context.Context,
	gtwIDs []*ttnpb.GatewayIdentifiers,
	requiredGatewayRights *ttnpb.Rights,
) error {
	authInfo, err := is.authInfo(ctx)
	if err != nil {
		return err
	}

	// If the caller is not an organization or user, there's nothing more to do.
	ouID := authInfo.GetOrganizationOrUserIdentifiers()
	if ouID == nil {
		return errInsufficientRights.New()
	}

	// Check that the caller has the requested rights.
	authInfoRights := ttnpb.RightsFrom(authInfo.GetRights()...)
	if len(requiredGatewayRights.Intersect(authInfoRights).GetRights()) == 0 {
		return errInsufficientRights.New()
	}

	entityIDs := make([]string, 0, len(gtwIDs))
	for _, gtwID := range gtwIDs {
		entityIDs = append(entityIDs, gtwID.GetEntityIdentifiers().IDString())
	}

	return is.store.Transact(ctx, func(ctx context.Context, st store.Store) error {
		gtws, err := st.FindGateways(ctx, gtwIDs, []string{"ids", "status_public", "location_public"})
		if err != nil {
			return err
		}
		if len(gtws) != len(gtwIDs) {
			if is.IsAdmin(ctx) {
				// Return the cause only to the admin.
				return errSomeGatewaysNotFound.New()
			}
			// Some gateways were not found.
			return errInsufficientRights.New()
		}

		if len(requiredGatewayRights.Sub(
			ttnpb.RightsFrom(
				ttnpb.Right_RIGHT_GATEWAY_STATUS_READ,
			),
		).GetRights()) == 0 {
			// Only the public status is requested. Check that all the gateways in the batch allow this.
			for _, gtw := range gtws {
				if !gtw.StatusPublic {
					return errInsufficientRights.New()
				}
			}
			return nil
		}

		if len(requiredGatewayRights.Sub(
			ttnpb.RightsFrom(
				ttnpb.Right_RIGHT_GATEWAY_LOCATION_READ,
			),
		).GetRights()) == 0 {
			// Only public location is requested. Check that all the gateways in the batch allow this.
			for _, gtw := range gtws {
				if !gtw.LocationPublic {
					return errInsufficientRights.New()
				}
			}
			return nil
		}

		membershipChains, err := st.FindAccountMembershipChains(
			ctx,
			ouID,
			"gateway",
			entityIDs...,
		)
		if err != nil {
			return err
		}
		if len(membershipChains) != len(gtwIDs) {
			// Some memberships were not found.
			return errInsufficientRights.New()
		}

		for _, chain := range membershipChains {
			// Make sure that there are no extra rights requested.
			otherRights := requiredGatewayRights.Sub(
				requiredGatewayRights.Intersect(
					chain.GetRights(),
				),
			)
			if len(otherRights.GetRights()) > 0 {
				return errInsufficientRights.New()
			}
		}
		return nil
	})
}
