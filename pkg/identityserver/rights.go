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

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func (is *IdentityServer) getRights(ctx context.Context) (entity map[ttnpb.Identifiers]*ttnpb.Rights, universal *ttnpb.Rights, err error) {
	authInfo, err := is.authInfo(ctx)
	if err != nil {
		return nil, nil, err
	}
	entity, err = is.entityRights(ctx, authInfo)
	if err != nil {
		return nil, nil, err
	}
	universal = authInfo.UniversalRights
	return
}

// ApplicationRights returns the rights the caller has on the given application.
func (is *IdentityServer) ApplicationRights(ctx context.Context, appIDs ttnpb.ApplicationIdentifiers) (*ttnpb.Rights, error) {
	entity, universal, err := is.getRights(ctx)
	if err != nil {
		return nil, err
	}
	for ids, rights := range entity {
		if ids.EntityType() == "application" && ids.IDString() == appIDs.IDString() {
			return rights.Union(universal), nil
		}
	}
	if universal == nil {
		return &ttnpb.Rights{}, nil
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		_, err := store.GetApplicationStore(db).GetApplication(ctx, &appIDs, &types.FieldMask{Paths: []string{"ids"}})
		return err
	})
	if err != nil {
		return nil, err
	}
	return universal, nil
}

// ClientRights returns the rights the caller has on the given client.
func (is *IdentityServer) ClientRights(ctx context.Context, cliIDs ttnpb.ClientIdentifiers) (*ttnpb.Rights, error) {
	entity, universal, err := is.getRights(ctx)
	if err != nil {
		return nil, err
	}
	for ids, rights := range entity {
		if ids.EntityType() == "client" && ids.IDString() == cliIDs.IDString() {
			return rights.Union(universal), nil
		}
	}
	if universal == nil {
		return &ttnpb.Rights{}, nil
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		_, err := store.GetClientStore(db).GetClient(ctx, &cliIDs, &types.FieldMask{Paths: []string{"ids"}})
		return err
	})
	if err != nil {
		return nil, err
	}
	return universal, nil
}

// GatewayRights returns the rights the caller has on the given gateway.
func (is *IdentityServer) GatewayRights(ctx context.Context, gtwIDs ttnpb.GatewayIdentifiers) (*ttnpb.Rights, error) {
	entity, universal, err := is.getRights(ctx)
	if err != nil {
		return nil, err
	}
	for ids, rights := range entity {
		if ids.EntityType() == "gateway" && ids.IDString() == gtwIDs.IDString() {
			return rights.Union(universal), nil
		}
	}
	if universal == nil {
		return &ttnpb.Rights{}, nil
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		_, err := store.GetGatewayStore(db).GetGateway(ctx, &gtwIDs, &types.FieldMask{Paths: []string{"ids"}})
		return err
	})
	if err != nil {
		return nil, err
	}
	return universal, nil
}

// OrganizationRights returns the rights the caller has on the given organization.
func (is *IdentityServer) OrganizationRights(ctx context.Context, orgIDs ttnpb.OrganizationIdentifiers) (*ttnpb.Rights, error) {
	entity, universal, err := is.getRights(ctx)
	if err != nil {
		return nil, err
	}
	for ids, rights := range entity {
		if ids.EntityType() == "organization" && ids.IDString() == orgIDs.IDString() {
			return rights.Union(universal), nil
		}
	}
	if universal == nil {
		return &ttnpb.Rights{}, nil
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		_, err := store.GetOrganizationStore(db).GetOrganization(ctx, &orgIDs, &types.FieldMask{Paths: []string{"ids"}})
		return err
	})
	if err != nil {
		return nil, err
	}
	return universal, nil
}

// UserRights returns the rights the caller has on the given user.
func (is *IdentityServer) UserRights(ctx context.Context, userIDs ttnpb.UserIdentifiers) (*ttnpb.Rights, error) {
	entity, universal, err := is.getRights(ctx)
	if err != nil {
		return nil, err
	}
	for ids, rights := range entity {
		if ids.EntityType() == "user" && ids.IDString() == userIDs.IDString() {
			return rights.Union(universal), nil
		}
	}
	if universal == nil {
		return &ttnpb.Rights{}, nil
	}
	err = is.withDatabase(ctx, func(db *gorm.DB) error {
		_, err := store.GetUserStore(db).GetUser(ctx, &userIDs, &types.FieldMask{Paths: []string{"ids"}})
		return err
	})
	if err != nil {
		return nil, err
	}
	return universal, nil
}
