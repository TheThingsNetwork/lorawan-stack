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
	organizationAccessUser.Admin = false
	organizationAccessUser.State = ttnpb.STATE_APPROVED
	for _, apiKey := range userAPIKeys(&organizationAccessUser.UserIdentifiers).APIKeys {
		apiKey.Rights = []ttnpb.Right{
			ttnpb.RIGHT_ORGANIZATION_SETTINGS_API_KEYS,
			ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS,
		}
	}
}

func TestOrganizationAccessNotFound(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, creds := defaultUser.UserIdentifiers, userCreds(defaultUserIdx)
		organizationID := userOrganizations(&userID).Organizations[0].OrganizationIdentifiers

		reg := ttnpb.NewOrganizationAccessClient(cc)

		apiKey := ttnpb.APIKey{
			ID:   "does-not-exist-id",
			Name: "test-application-api-key-name",
		}

		got, err := reg.GetAPIKey(ctx, &ttnpb.GetOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
			KeyID:                   apiKey.ID,
		}, creds)

		a.So(got, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsNotFound(err), should.BeTrue)

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
			APIKey:                  apiKey,
		}, creds)

		a.So(updated, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsNotFound(err), should.BeTrue)
	})
}

func TestOrganizationAccessRightsPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, creds := organizationAccessUser.UserIdentifiers, userCreds(organizationAccessUserIdx)
		organizationID := userOrganizations(&userID).Organizations[0].OrganizationIdentifiers
		collaboratorID := collaboratorUser.UserIdentifiers.OrganizationOrUserIdentifiers()

		reg := ttnpb.NewOrganizationAccessClient(cc)

		APIKeyName := "test-organization-api-key-name"
		APIKey, err := reg.CreateAPIKey(ctx, &ttnpb.CreateOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
			Name:                    APIKeyName,
			Rights:                  []ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_ALL},
		}, creds)

		a.So(APIKey, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		APIKey = organizationAPIKeys(&organizationID).APIKeys[0]

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
			APIKey:                  *APIKey,
		}, creds)

		a.So(updated, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIdentifiers: organizationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *collaboratorID,
				Rights:                        []ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_ALL},
			},
		}, creds)

		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)
	})
}

func TestOrganizationAccessPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := defaultUser.UserIdentifiers
		organizationID := userOrganizations(&userID).Organizations[0].OrganizationIdentifiers
		collaboratorID := collaboratorUser.UserIdentifiers.OrganizationOrUserIdentifiers()
		APIKeyID := organizationAPIKeys(&organizationID).APIKeys[0].ID

		reg := ttnpb.NewOrganizationAccessClient(cc)

		rights, err := reg.ListRights(ctx, &organizationID)

		a.So(rights, should.NotBeNil)
		a.So(rights.Rights, should.BeEmpty)
		a.So(err, should.BeNil)

		APIKey, err := reg.GetAPIKey(ctx, &ttnpb.GetOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
			KeyID:                   APIKeyID,
		})

		a.So(APIKey, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		APIKeys, err := reg.ListAPIKeys(ctx, &ttnpb.ListOrganizationAPIKeysRequest{
			OrganizationIdentifiers: organizationID,
		})

		a.So(APIKeys, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		collaborators, err := reg.ListCollaborators(ctx, &ttnpb.ListOrganizationCollaboratorsRequest{
			OrganizationIdentifiers: organizationID,
		})

		a.So(collaborators, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		APIKeyName := "test-organization-api-key-name"
		APIKey, err = reg.CreateAPIKey(ctx, &ttnpb.CreateOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
			Name:                    APIKeyName,
			Rights:                  []ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_ALL},
		})

		a.So(APIKey, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		APIKey = organizationAPIKeys(&organizationID).APIKeys[0]

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
			APIKey:                  *APIKey,
		})

		a.So(updated, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIdentifiers: organizationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *collaboratorID,
				Rights:                        []ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_ALL},
			},
		})

		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)
	})
}

func TestOrganizationAccessClusterAuth(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := defaultUser.UserIdentifiers
		organizationID := userOrganizations(&userID).Organizations[0].OrganizationIdentifiers

		reg := ttnpb.NewOrganizationAccessClient(cc)

		rights, err := reg.ListRights(ctx, &organizationID, is.WithClusterAuth())

		a.So(rights, should.NotBeNil)
		a.So(ttnpb.AllClusterRights.Intersect(ttnpb.AllOrganizationRights).Sub(rights).Rights, should.BeEmpty)
		a.So(err, should.BeNil)
	})
}

func TestOrganizationAccessCRUD(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, creds := defaultUser.UserIdentifiers, userCreds(defaultUserIdx)
		organizationID := userOrganizations(&userID).Organizations[0].OrganizationIdentifiers
		collaboratorID := collaboratorUser.UserIdentifiers.OrganizationOrUserIdentifiers()

		reg := ttnpb.NewOrganizationAccessClient(cc)

		rights, err := reg.ListRights(ctx, &organizationID, creds)

		a.So(rights, should.NotBeNil)
		a.So(rights.Rights, should.Contain, ttnpb.RIGHT_ORGANIZATION_ALL)
		a.So(err, should.BeNil)

		modifiedOrganizationID := organizationID
		modifiedOrganizationID.OrganizationID += "mod"

		rights, err = reg.ListRights(ctx, &modifiedOrganizationID, creds)
		a.So(rights, should.NotBeNil)
		a.So(rights.Rights, should.BeEmpty)
		a.So(err, should.BeNil)

		organizationAPIKeys := organizationAPIKeys(&organizationID)
		organizationKey := organizationAPIKeys.APIKeys[0]

		APIKey, err := reg.GetAPIKey(ctx, &ttnpb.GetOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
			KeyID:                   organizationKey.ID,
		}, creds)

		a.So(APIKey, should.NotBeNil)
		a.So(err, should.BeNil)
		a.So(APIKey.ID, should.Equal, organizationKey.ID)
		a.So(APIKey.Key, should.BeEmpty)

		APIKeys, err := reg.ListAPIKeys(ctx, &ttnpb.ListOrganizationAPIKeysRequest{
			OrganizationIdentifiers: organizationID,
		}, creds)

		a.So(APIKeys, should.NotBeNil)
		a.So(len(APIKeys.APIKeys), should.Equal, len(organizationAPIKeys.APIKeys))
		a.So(err, should.BeNil)
		for i, APIkey := range APIKeys.APIKeys {
			a.So(APIkey.Name, should.Equal, organizationAPIKeys.APIKeys[i].Name)
			a.So(APIkey.ID, should.Equal, organizationAPIKeys.APIKeys[i].ID)
		}

		collaborators, err := reg.ListCollaborators(ctx, &ttnpb.ListOrganizationCollaboratorsRequest{
			OrganizationIdentifiers: organizationID,
		}, creds)

		a.So(collaborators, should.NotBeNil)
		a.So(collaborators.Collaborators, should.NotBeEmpty)
		a.So(err, should.BeNil)

		APIKeyName := "test-organization-api-key-name"
		APIKey, err = reg.CreateAPIKey(ctx, &ttnpb.CreateOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
			Name:                    APIKeyName,
			Rights:                  []ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_ALL},
		}, creds)

		a.So(APIKey, should.NotBeNil)
		a.So(APIKey.Name, should.Equal, APIKeyName)
		a.So(err, should.BeNil)

		newAPIKeyName := "test-new-organization-api-key"
		APIKey.Name = newAPIKeyName
		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
			APIKey:                  *APIKey,
		}, creds)

		a.So(updated, should.NotBeNil)
		a.So(updated.Name, should.Equal, newAPIKeyName)
		a.So(err, should.BeNil)

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIdentifiers: organizationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *collaboratorID,
				Rights:                        []ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_ALL},
			},
		}, creds)

		a.So(err, should.BeNil)
	})
}
