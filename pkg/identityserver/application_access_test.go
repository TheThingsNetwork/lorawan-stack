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
	applicationAccessUser.Admin = false
	applicationAccessUser.State = ttnpb.STATE_APPROVED
	for _, apiKey := range userAPIKeys(&applicationAccessUser.UserIdentifiers).APIKeys {
		apiKey.Rights = []ttnpb.Right{
			ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS,
			ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS,
		}
	}
	appAccessCollaboratorUser.Admin = false
	appAccessCollaboratorUser.State = ttnpb.STATE_APPROVED
	for _, apiKey := range userAPIKeys(&appAccessCollaboratorUser.UserIdentifiers).APIKeys {
		apiKey.Rights = []ttnpb.Right{
			ttnpb.RIGHT_APPLICATION_ALL,
		}
	}
}

func TestApplicationAccessNotFound(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, creds := defaultUser.UserIdentifiers, userCreds(defaultUserIdx)
		applicationID := userApplications(&userID).Applications[0].ApplicationIdentifiers

		reg := ttnpb.NewApplicationAccessClient(cc)

		apiKey := ttnpb.APIKey{
			ID:   "does-not-exist-id",
			Name: "test-application-api-key-name",
		}

		got, err := reg.GetAPIKey(ctx, &ttnpb.GetApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			KeyID:                  apiKey.ID,
		}, creds)

		a.So(got, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsNotFound(err), should.BeTrue)

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			APIKey:                 apiKey,
		}, creds)

		a.So(updated, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsNotFound(err), should.BeTrue)
	})
}

func TestApplicationAccessRightsPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, creds := applicationAccessUser.UserIdentifiers, userCreds(applicationAccessUserIdx)
		applicationID := userApplications(&userID).Applications[0].ApplicationIdentifiers
		collaboratorID := collaboratorUser.UserIdentifiers.OrganizationOrUserIdentifiers()

		reg := ttnpb.NewApplicationAccessClient(cc)

		APIKeyName := "test-application-api-key-name"
		APIKey, err := reg.CreateAPIKey(ctx, &ttnpb.CreateApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			Name:                   APIKeyName,
			Rights:                 []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL},
		}, creds)

		a.So(APIKey, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		APIKey = applicationAPIKeys(&applicationID).APIKeys[0]

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			APIKey:                 *APIKey,
		}, creds)

		a.So(updated, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		// Choose right that the user does not have and hence cannot add
		right := ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIdentifiers: applicationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *collaboratorID,
				Rights:                        []ttnpb.Right{right},
			},
		}, creds)

		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)
	})
}

func TestApplicationAccessPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := defaultUser.UserIdentifiers
		applicationID := userApplications(&userID).Applications[0].ApplicationIdentifiers
		collaboratorID := collaboratorUser.UserIdentifiers.OrganizationOrUserIdentifiers()
		APIKeyID := applicationAPIKeys(&applicationID).APIKeys[0].ID

		reg := ttnpb.NewApplicationAccessClient(cc)

		rights, err := reg.ListRights(ctx, &applicationID)

		a.So(rights, should.NotBeNil)
		a.So(rights.Rights, should.BeEmpty)
		a.So(err, should.BeNil)

		APIKey, err := reg.GetAPIKey(ctx, &ttnpb.GetApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			KeyID:                  APIKeyID,
		})

		a.So(APIKey, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		APIKeys, err := reg.ListAPIKeys(ctx, &ttnpb.ListApplicationAPIKeysRequest{
			ApplicationIdentifiers: applicationID,
		})

		a.So(APIKeys, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		collaborators, err := reg.ListCollaborators(ctx, &ttnpb.ListApplicationCollaboratorsRequest{
			ApplicationIdentifiers: applicationID,
		})

		a.So(collaborators, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		APIKeyName := "test-application-api-key-name"
		APIKey, err = reg.CreateAPIKey(ctx, &ttnpb.CreateApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			Name:                   APIKeyName,
			Rights:                 []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL},
		})

		a.So(APIKey, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		APIKey = applicationAPIKeys(&applicationID).APIKeys[0]

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			APIKey:                 *APIKey,
		})

		a.So(updated, should.BeNil)
		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIdentifiers: applicationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *collaboratorID,
				Rights:                        []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL},
			},
		})

		a.So(err, should.NotBeNil)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)
	})
}

func TestApplicationAccessClusterAuth(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := defaultUser.UserIdentifiers
		applicationID := userApplications(&userID).Applications[0].ApplicationIdentifiers

		reg := ttnpb.NewApplicationAccessClient(cc)

		rights, err := reg.ListRights(ctx, &applicationID, is.WithClusterAuth())

		a.So(rights, should.NotBeNil)
		a.So(ttnpb.AllClusterRights.Intersect(ttnpb.AllApplicationRights).Sub(rights).Rights, should.BeEmpty)
		a.So(err, should.BeNil)
	})
}

func TestApplicationAccessCRUD(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, creds := defaultUser.UserIdentifiers, userCreds(defaultUserIdx)
		applicationID := userApplications(&userID).Applications[0].ApplicationIdentifiers
		collaboratorID := collaboratorUser.UserIdentifiers.OrganizationOrUserIdentifiers()

		reg := ttnpb.NewApplicationAccessClient(cc)

		rights, err := reg.ListRights(ctx, &applicationID, creds)

		a.So(rights, should.NotBeNil)
		a.So(rights.Rights, should.NotBeEmpty)
		a.So(err, should.BeNil)

		modifiedApplicationID := applicationID
		modifiedApplicationID.ApplicationID = reverse(modifiedApplicationID.ApplicationID)

		rights, err = reg.ListRights(ctx, &modifiedApplicationID, creds)
		a.So(rights, should.NotBeNil)
		a.So(rights.Rights, should.BeEmpty)
		a.So(err, should.BeNil)

		applicationAPIKeys := applicationAPIKeys(&applicationID)
		applicationKey := applicationAPIKeys.APIKeys[0]

		APIKey, err := reg.GetAPIKey(ctx, &ttnpb.GetApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			KeyID:                  applicationKey.ID,
		}, creds)

		a.So(APIKey, should.NotBeNil)
		a.So(err, should.BeNil)
		a.So(APIKey.ID, should.Equal, applicationKey.ID)
		a.So(APIKey.Key, should.BeEmpty)

		APIKeys, err := reg.ListAPIKeys(ctx, &ttnpb.ListApplicationAPIKeysRequest{
			ApplicationIdentifiers: applicationID,
		}, creds)

		a.So(APIKeys, should.NotBeNil)
		a.So(err, should.BeNil)
		a.So(len(APIKeys.APIKeys), should.Equal, len(applicationAPIKeys.APIKeys))
		for i, APIkey := range APIKeys.APIKeys {
			a.So(APIkey.Name, should.Equal, applicationAPIKeys.APIKeys[i].Name)
			a.So(APIkey.ID, should.Equal, applicationAPIKeys.APIKeys[i].ID)
		}

		collaborators, err := reg.ListCollaborators(ctx, &ttnpb.ListApplicationCollaboratorsRequest{
			ApplicationIdentifiers: applicationID,
		}, creds)

		a.So(collaborators, should.NotBeNil)
		a.So(collaborators.Collaborators, should.NotBeEmpty)
		a.So(err, should.BeNil)

		APIKeyName := "test-application-api-key-name"
		APIKey, err = reg.CreateAPIKey(ctx, &ttnpb.CreateApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			Name:                   APIKeyName,
			Rights:                 []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL},
		}, creds)

		a.So(APIKey, should.NotBeNil)
		a.So(APIKey.Name, should.Equal, APIKeyName)
		a.So(err, should.BeNil)

		newAPIKeyName := "test-new-api-key"
		APIKey.Name = newAPIKeyName
		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			APIKey:                 *APIKey,
		}, creds)

		a.So(updated, should.NotBeNil)
		a.So(updated.Name, should.Equal, newAPIKeyName)
		a.So(err, should.BeNil)

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIdentifiers: applicationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *collaboratorID,
				Rights:                        []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL},
			},
		}, creds)

		a.So(err, should.BeNil)

		res, err := reg.GetCollaborator(ctx, &ttnpb.GetApplicationCollaboratorRequest{
			ApplicationIdentifiers:        applicationID,
			OrganizationOrUserIdentifiers: *collaboratorID,
		}, creds)

		a.So(err, should.BeNil)
		a.So(res.Rights, should.Resemble, []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL})
	})
}

func TestApplicationAccessCollaboratorRights(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, usrCreds := defaultUser.UserIdentifiers, userCreds(defaultUserIdx)
		applicationID := userApplications(&userID).Applications[0].ApplicationIdentifiers
		collaboratorID := applicationAccessUser.UserIdentifiers.OrganizationOrUserIdentifiers()
		collaboratorCreds := userCreds(applicationAccessUserIdx)
		removedCollaboratorID := appAccessCollaboratorUser.UserIdentifiers.OrganizationOrUserIdentifiers()

		reg := ttnpb.NewApplicationAccessClient(cc)

		_, err := reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIdentifiers: applicationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *collaboratorID,
				Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS,
					ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS,
				},
			},
		}, usrCreds)

		a.So(err, should.BeNil)

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIdentifiers: applicationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *removedCollaboratorID,
				Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_ALL,
				},
			},
		}, usrCreds)

		a.So(err, should.BeNil)

		// Try revoking rights for the collaborator with RIGHT_APPLICATION_ALL without having it
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIdentifiers: applicationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *removedCollaboratorID,
				Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS,
					ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS,
				},
			},
		}, collaboratorCreds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		// Remove RIGHT_APPLICATION_ALL from collaborator to be removed
		newRights := ttnpb.AllApplicationRights.Sub(ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_ALL))
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIdentifiers: applicationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *removedCollaboratorID,
				Rights:                        newRights.Rights,
			},
		}, usrCreds)

		a.So(err, should.BeNil)

		newRights = newRights.Sub(ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS))
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIdentifiers: applicationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *removedCollaboratorID,
				Rights:                        newRights.Rights,
			},
		}, collaboratorCreds)

		a.So(err, should.BeNil)

		// Try revoking RIGHT_APPLICATION_INFO without having it
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIdentifiers: applicationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *removedCollaboratorID,
				Rights:                        newRights.Sub(ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_INFO)).Rights,
			},
		}, collaboratorCreds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		res, err := reg.GetCollaborator(ctx, &ttnpb.GetApplicationCollaboratorRequest{
			ApplicationIdentifiers:        applicationID,
			OrganizationOrUserIdentifiers: *removedCollaboratorID,
		}, collaboratorCreds)

		if a.So(err, should.BeNil) {
			a.So(res.Rights, should.Resemble, newRights.Rights)
		}

		// Delete collaborator with more rights
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIdentifiers: applicationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *removedCollaboratorID,
				Rights:                        []ttnpb.Right{},
			},
		}, collaboratorCreds)

		a.So(err, should.BeNil)

		_, err = reg.GetCollaborator(ctx, &ttnpb.GetApplicationCollaboratorRequest{
			ApplicationIdentifiers:        applicationID,
			OrganizationOrUserIdentifiers: *removedCollaboratorID,
		}, collaboratorCreds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
	})
}
