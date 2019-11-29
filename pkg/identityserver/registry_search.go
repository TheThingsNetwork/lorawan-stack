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

	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

type registrySearch struct {
	*IdentityServer
}

var errSearchForbidden = errors.DefinePermissionDenied("search_forbidden", "search is forbidden")

func (rs *registrySearch) memberForSearch(ctx context.Context) (*ttnpb.OrganizationOrUserIdentifiers, error) {
	authInfo, err := rs.authInfo(ctx)
	if err != nil {
		return nil, err
	}
	if authInfo.IsAdmin {
		return nil, nil
	}
	member := authInfo.GetOrganizationOrUserIdentifiers()
	if member != nil {
		return member, nil
	}
	return nil, errSearchForbidden
}

func (rs *registrySearch) SearchApplications(ctx context.Context, req *ttnpb.SearchEntitiesRequest) (*ttnpb.Applications, error) {
	member, err := rs.memberForSearch(ctx)
	if err != nil {
		return nil, err
	}
	req.FieldMask.Paths = cleanFieldMaskPaths(ttnpb.ApplicationFieldPathsNested, req.FieldMask.Paths, getPaths, nil)
	res := &ttnpb.Applications{}
	err = rs.withDatabase(ctx, func(db *gorm.DB) error {
		entityIDs, err := store.GetEntitySearch(db).FindEntities(ctx, member, req, "application")
		if err != nil {
			return err
		}
		var ids []*ttnpb.ApplicationIdentifiers
		for _, id := range entityIDs {
			id := id.Identifiers().(*ttnpb.ApplicationIdentifiers)
			if rights.RequireApplication(ctx, *id, ttnpb.RIGHT_APPLICATION_INFO) == nil {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			return nil
		}
		res.Applications, err = store.GetApplicationStore(db).FindApplications(ctx, ids, &req.FieldMask)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (rs *registrySearch) SearchClients(ctx context.Context, req *ttnpb.SearchEntitiesRequest) (*ttnpb.Clients, error) {
	member, err := rs.memberForSearch(ctx)
	if err != nil {
		return nil, err
	}
	req.FieldMask.Paths = cleanFieldMaskPaths(ttnpb.ClientFieldPathsNested, req.FieldMask.Paths, getPaths, nil)
	res := &ttnpb.Clients{}
	err = rs.withDatabase(ctx, func(db *gorm.DB) error {
		entityIDs, err := store.GetEntitySearch(db).FindEntities(ctx, member, req, "client")
		if err != nil {
			return err
		}
		var ids []*ttnpb.ClientIdentifiers
		for _, id := range entityIDs {
			id := id.Identifiers().(*ttnpb.ClientIdentifiers)
			if rights.RequireClient(ctx, *id, ttnpb.RIGHT_CLIENT_ALL) == nil {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			return nil
		}
		res.Clients, err = store.GetClientStore(db).FindClients(ctx, ids, &req.FieldMask)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (rs *registrySearch) SearchGateways(ctx context.Context, req *ttnpb.SearchEntitiesRequest) (*ttnpb.Gateways, error) {
	member, err := rs.memberForSearch(ctx)
	if err != nil {
		return nil, err
	}
	req.FieldMask.Paths = cleanFieldMaskPaths(ttnpb.GatewayFieldPathsNested, req.FieldMask.Paths, getPaths, nil)
	res := &ttnpb.Gateways{}
	err = rs.withDatabase(ctx, func(db *gorm.DB) error {
		entityIDs, err := store.GetEntitySearch(db).FindEntities(ctx, member, req, "gateway")
		if err != nil {
			return err
		}
		var ids []*ttnpb.GatewayIdentifiers
		for _, id := range entityIDs {
			id := id.Identifiers().(*ttnpb.GatewayIdentifiers)
			if rights.RequireGateway(ctx, *id, ttnpb.RIGHT_GATEWAY_INFO) == nil {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			return nil
		}
		res.Gateways, err = store.GetGatewayStore(db).FindGateways(ctx, ids, &req.FieldMask)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (rs *registrySearch) SearchOrganizations(ctx context.Context, req *ttnpb.SearchEntitiesRequest) (*ttnpb.Organizations, error) {
	member, err := rs.memberForSearch(ctx)
	if err != nil {
		return nil, err
	}
	req.FieldMask.Paths = cleanFieldMaskPaths(ttnpb.OrganizationFieldPathsNested, req.FieldMask.Paths, getPaths, nil)
	res := &ttnpb.Organizations{}
	err = rs.withDatabase(ctx, func(db *gorm.DB) error {
		entityIDs, err := store.GetEntitySearch(db).FindEntities(ctx, member, req, "organization")
		if err != nil {
			return err
		}
		var ids []*ttnpb.OrganizationIdentifiers
		for _, id := range entityIDs {
			id := id.Identifiers().(*ttnpb.OrganizationIdentifiers)
			if rights.RequireOrganization(ctx, *id, ttnpb.RIGHT_ORGANIZATION_INFO) == nil {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			return nil
		}
		res.Organizations, err = store.GetOrganizationStore(db).FindOrganizations(ctx, ids, &req.FieldMask)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (rs *registrySearch) SearchUsers(ctx context.Context, req *ttnpb.SearchEntitiesRequest) (*ttnpb.Users, error) {
	member, err := rs.memberForSearch(ctx)
	if err != nil {
		return nil, err
	}
	req.FieldMask.Paths = cleanFieldMaskPaths(ttnpb.UserFieldPathsNested, req.FieldMask.Paths, getPaths, nil)
	res := &ttnpb.Users{}
	err = rs.withDatabase(ctx, func(db *gorm.DB) error {
		entityIDs, err := store.GetEntitySearch(db).FindEntities(ctx, member, req, "user")
		if err != nil {
			return err
		}
		var ids []*ttnpb.UserIdentifiers
		for _, id := range entityIDs {
			id := id.Identifiers().(*ttnpb.UserIdentifiers)
			if rights.RequireUser(ctx, *id, ttnpb.RIGHT_USER_INFO) == nil {
				ids = append(ids, id)
			}
		}
		if len(ids) == 0 {
			return nil
		}
		res.Users, err = store.GetUserStore(db).FindUsers(ctx, ids, &req.FieldMask)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
