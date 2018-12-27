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

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func (is *IdentityServer) getRights(ctx context.Context) (map[*ttnpb.EntityIdentifiers]*ttnpb.Rights, error) {
	authInfo, err := is.authInfo(ctx)
	if err != nil {
		return nil, err
	}
	entityRights, err := is.entityRights(ctx, authInfo)
	return entityRights, err
}

// ApplicationRights returns the rights the caller has on the given application.
func (is *IdentityServer) ApplicationRights(ctx context.Context, appIDs ttnpb.ApplicationIdentifiers) (*ttnpb.Rights, error) {
	rights, err := is.getRights(ctx)
	if err != nil {
		return nil, err
	}
	for ids, rights := range rights {
		if ids := ids.GetApplicationIDs(); ids != nil && ids.ApplicationID == appIDs.ApplicationID {
			return rights, nil
		}
	}
	return &ttnpb.Rights{}, nil
}

// ClientRights returns the rights the caller has on the given client.
func (is *IdentityServer) ClientRights(ctx context.Context, cliIDs ttnpb.ClientIdentifiers) (*ttnpb.Rights, error) {
	rights, err := is.getRights(ctx)
	if err != nil {
		return nil, err
	}
	for ids, rights := range rights {
		if ids := ids.GetClientIDs(); ids != nil && ids.ClientID == cliIDs.ClientID {
			return rights, nil
		}
	}
	return &ttnpb.Rights{}, nil
}

// GatewayRights returns the rights the caller has on the given gateway.
func (is *IdentityServer) GatewayRights(ctx context.Context, gtwIDs ttnpb.GatewayIdentifiers) (*ttnpb.Rights, error) {
	rights, err := is.getRights(ctx)
	if err != nil {
		return nil, err
	}
	for ids, rights := range rights {
		if ids := ids.GetGatewayIDs(); ids != nil && ids.GatewayID == gtwIDs.GatewayID {
			return rights, nil
		}
	}
	return &ttnpb.Rights{}, nil
}

// OrganizationRights returns the rights the caller has on the given organization.
func (is *IdentityServer) OrganizationRights(ctx context.Context, orgIDs ttnpb.OrganizationIdentifiers) (*ttnpb.Rights, error) {
	rights, err := is.getRights(ctx)
	if err != nil {
		return nil, err
	}
	for ids, rights := range rights {
		if ids := ids.GetOrganizationIDs(); ids != nil && ids.OrganizationID == orgIDs.OrganizationID {
			return rights, nil
		}
	}
	return &ttnpb.Rights{}, nil
}

// UserRights returns the rights the caller has on the given user.
func (is *IdentityServer) UserRights(ctx context.Context, userIDs ttnpb.UserIdentifiers) (*ttnpb.Rights, error) {
	rights, err := is.getRights(ctx)
	if err != nil {
		return nil, err
	}
	for ids, rights := range rights {
		if ids := ids.GetUserIDs(); ids != nil && ids.UserID == userIDs.UserID {
			return rights, nil
		}
	}
	return &ttnpb.Rights{}, nil
}
