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
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc"
)

func init() {
	gatewayAccessUser.Admin = false
	gatewayAccessUser.State = ttnpb.STATE_APPROVED
	for _, apiKey := range userAPIKeys(&gatewayAccessUser.UserIdentifiers).APIKeys {
		apiKey.Rights = []ttnpb.Right{
			ttnpb.RIGHT_GATEWAY_SETTINGS_API_KEYS,
			ttnpb.RIGHT_GATEWAY_SETTINGS_COLLABORATORS,
		}
	}
}

func TestGatewayAccessNotFound(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, creds := defaultUser.UserIdentifiers, userCreds(defaultUserIdx)
		gatewayID := userGateways(&userID).Gateways[0].GatewayIdentifiers

		reg := ttnpb.NewGatewayAccessClient(cc)

		apiKey := ttnpb.APIKey{
			ID:   "does-not-exist-id",
			Name: "test-gateway-api-key-name",
		}

		got, err := reg.GetAPIKey(ctx, &ttnpb.GetGatewayAPIKeyRequest{
			GatewayIdentifiers: gatewayID,
			KeyID:              apiKey.ID,
		}, creds)

		a.So(got, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsNotFound(err), should.BeTrue)

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateGatewayAPIKeyRequest{
			GatewayIdentifiers: gatewayID,
			APIKey:             apiKey,
		}, creds)

		a.So(updated, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsNotFound(err), should.BeTrue)
	})
}

func TestGatewayAccessRightsPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, creds := gatewayAccessUser.UserIdentifiers, userCreds(gatewayAccessUserIdx)
		gatewayID := userGateways(&userID).Gateways[0].GatewayIdentifiers
		collaboratorID := collaboratorUser.UserIdentifiers.OrganizationOrUserIdentifiers()

		reg := ttnpb.NewGatewayAccessClient(cc)

		APIKeyName := "test-gateway-api-key-name"
		APIKey, err := reg.CreateAPIKey(ctx, &ttnpb.CreateGatewayAPIKeyRequest{
			GatewayIdentifiers: gatewayID,
			Name:               APIKeyName,
			Rights:             []ttnpb.Right{ttnpb.RIGHT_GATEWAY_ALL},
		}, creds)

		a.So(APIKey, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		APIKey = gatewayAPIKeys(&gatewayID).APIKeys[0]

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateGatewayAPIKeyRequest{
			GatewayIdentifiers: gatewayID,
			APIKey:             *APIKey,
		}, creds)

		a.So(updated, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetGatewayCollaboratorRequest{
			GatewayIdentifiers: gatewayID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *collaboratorID,
				Rights:                        []ttnpb.Right{ttnpb.RIGHT_GATEWAY_ALL},
			},
		}, creds)

		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)
	})
}

func TestGatewayAccessPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := defaultUser.UserIdentifiers
		gatewayID := userGateways(&userID).Gateways[0].GatewayIdentifiers
		collaboratorID := collaboratorUser.UserIdentifiers.OrganizationOrUserIdentifiers()
		APIKeyID := gatewayAPIKeys(&gatewayID).APIKeys[0].ID

		reg := ttnpb.NewGatewayAccessClient(cc)

		rights, err := reg.ListRights(ctx, &gatewayID)

		a.So(rights, should.NotBeNil)
		a.So(rights.Rights, should.BeEmpty)
		a.So(err, should.BeNil)

		APIKey, err := reg.GetAPIKey(ctx, &ttnpb.GetGatewayAPIKeyRequest{
			GatewayIdentifiers: gatewayID,
			KeyID:              APIKeyID,
		})

		a.So(APIKey, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		APIKeys, err := reg.ListAPIKeys(ctx, &ttnpb.ListGatewayAPIKeysRequest{
			GatewayIdentifiers: gatewayID,
		})

		a.So(APIKeys, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		collaborators, err := reg.ListCollaborators(ctx, &ttnpb.ListGatewayCollaboratorsRequest{
			GatewayIdentifiers: gatewayID,
		})

		a.So(collaborators, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		APIKeyName := "test-gateway-api-key-name"
		APIKey, err = reg.CreateAPIKey(ctx, &ttnpb.CreateGatewayAPIKeyRequest{
			GatewayIdentifiers: gatewayID,
			Name:               APIKeyName,
			Rights:             []ttnpb.Right{ttnpb.RIGHT_GATEWAY_ALL},
		})

		a.So(APIKey, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		APIKey = gatewayAPIKeys(&gatewayID).APIKeys[0]

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateGatewayAPIKeyRequest{
			GatewayIdentifiers: gatewayID,
			APIKey:             *APIKey,
		})

		a.So(updated, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetGatewayCollaboratorRequest{
			GatewayIdentifiers: gatewayID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *collaboratorID,
				Rights:                        []ttnpb.Right{ttnpb.RIGHT_GATEWAY_ALL},
			},
		})

		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)
	})
}

func TestGatewayAccessClusterAuth(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := defaultUser.UserIdentifiers
		gatewayID := userGateways(&userID).Gateways[0].GatewayIdentifiers

		reg := ttnpb.NewGatewayAccessClient(cc)

		rights, err := reg.ListRights(ctx, &gatewayID, is.WithClusterAuth())

		a.So(rights, should.NotBeNil)
		a.So(ttnpb.AllClusterRights.Intersect(ttnpb.AllGatewayRights).Sub(rights).Rights, should.BeEmpty)
		a.So(err, should.BeNil)
	})
}

func TestGatewayAccessCRUD(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, userCreds := defaultUser.UserIdentifiers, userCreds(defaultUserIdx)
		gatewayID := userGateways(&userID).Gateways[0].GatewayIdentifiers
		collaboratorID := collaboratorUser.UserIdentifiers.OrganizationOrUserIdentifiers()

		reg := ttnpb.NewGatewayAccessClient(cc)

		rights, err := reg.ListRights(ctx, &gatewayID, userCreds)

		a.So(rights, should.NotBeNil)
		a.So(rights.Rights, should.Contain, ttnpb.RIGHT_GATEWAY_ALL)
		a.So(err, should.BeNil)

		modifiedGatewayID := gatewayID
		modifiedGatewayID.GatewayID += "mod"

		rights, err = reg.ListRights(ctx, &modifiedGatewayID, userCreds)
		a.So(rights, should.NotBeNil)
		a.So(rights.Rights, should.BeEmpty)
		a.So(err, should.BeNil)

		gatewayAPIKeys := gatewayAPIKeys(&gatewayID)
		gatewayKey := gatewayAPIKeys.APIKeys[0]

		APIKey, err := reg.GetAPIKey(ctx, &ttnpb.GetGatewayAPIKeyRequest{
			GatewayIdentifiers: gatewayID,
			KeyID:              gatewayKey.ID,
		}, userCreds)

		a.So(APIKey, should.NotBeNil)
		a.So(err, should.BeNil)
		a.So(APIKey.ID, should.Equal, gatewayKey.ID)
		a.So(APIKey.Key, should.BeEmpty)

		APIKeys, err := reg.ListAPIKeys(ctx, &ttnpb.ListGatewayAPIKeysRequest{
			GatewayIdentifiers: gatewayID,
		}, userCreds)

		a.So(APIKeys, should.NotBeNil)
		a.So(len(APIKeys.APIKeys), should.Equal, len(gatewayAPIKeys.APIKeys))
		a.So(err, should.BeNil)
		for i, APIkey := range APIKeys.APIKeys {
			a.So(APIkey.Name, should.Equal, gatewayAPIKeys.APIKeys[i].Name)
			a.So(APIkey.ID, should.Equal, gatewayAPIKeys.APIKeys[i].ID)
		}

		collaborators, err := reg.ListCollaborators(ctx, &ttnpb.ListGatewayCollaboratorsRequest{
			GatewayIdentifiers: gatewayID,
		}, userCreds)

		a.So(collaborators, should.NotBeNil)
		a.So(collaborators.Collaborators, should.NotBeEmpty)
		a.So(err, should.BeNil)

		APIKeyName := "test-gateway-api-key-name"
		APIKey, err = reg.CreateAPIKey(ctx, &ttnpb.CreateGatewayAPIKeyRequest{
			GatewayIdentifiers: gatewayID,
			Name:               APIKeyName,
			Rights:             []ttnpb.Right{ttnpb.RIGHT_GATEWAY_ALL},
		}, userCreds)

		a.So(APIKey, should.NotBeNil)
		a.So(APIKey.Name, should.Equal, APIKeyName)
		a.So(err, should.BeNil)

		newAPIKeyName := "test-new-gateway-api-key"
		APIKey.Name = newAPIKeyName
		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateGatewayAPIKeyRequest{
			GatewayIdentifiers: gatewayID,
			APIKey:             *APIKey,
		}, userCreds)

		a.So(updated, should.NotBeNil)
		a.So(updated.Name, should.Equal, newAPIKeyName)
		a.So(err, should.BeNil)

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetGatewayCollaboratorRequest{
			GatewayIdentifiers: gatewayID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *collaboratorID,
				Rights:                        []ttnpb.Right{ttnpb.RIGHT_GATEWAY_ALL},
			},
		}, userCreds)

		a.So(err, should.BeNil)
	})
}
