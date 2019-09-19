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
			ttnpb.RIGHT_APPLICATION_LINK,
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

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		a.So(got, should.BeNil)

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			APIKey:                 apiKey,
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		a.So(updated, should.BeNil)
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

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(APIKey, should.BeNil)

		// Choose right that the user does not have and hence cannot add
		right := ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC
		APIKey = applicationAPIKeys(&applicationID).APIKeys[0]

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			APIKey: ttnpb.APIKey{
				ID:     APIKey.ID,
				Name:   APIKey.Name,
				Rights: []ttnpb.Right{right},
			},
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(updated, should.BeNil)

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIdentifiers: applicationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *collaboratorID,
				Rights:                        []ttnpb.Right{right},
			},
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
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

		a.So(err, should.BeNil)
		if a.So(rights, should.NotBeNil) {
			a.So(rights.Rights, should.BeEmpty)
		}

		APIKey, err := reg.GetAPIKey(ctx, &ttnpb.GetApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			KeyID:                  APIKeyID,
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(APIKey, should.BeNil)

		APIKeys, err := reg.ListAPIKeys(ctx, &ttnpb.ListApplicationAPIKeysRequest{
			ApplicationIdentifiers: applicationID,
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(APIKeys, should.BeNil)

		collaborators, err := reg.ListCollaborators(ctx, &ttnpb.ListApplicationCollaboratorsRequest{
			ApplicationIdentifiers: applicationID,
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(collaborators, should.BeNil)

		APIKeyName := "test-application-api-key-name"
		APIKey, err = reg.CreateAPIKey(ctx, &ttnpb.CreateApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			Name:                   APIKeyName,
			Rights:                 []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(APIKey, should.BeNil)

		APIKey = applicationAPIKeys(&applicationID).APIKeys[0]

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			APIKey:                 *APIKey,
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(updated, should.BeNil)

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIdentifiers: applicationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *collaboratorID,
				Rights:                        []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL},
			},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
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

		a.So(err, should.BeNil)
		if a.So(rights, should.NotBeNil) {
			a.So(ttnpb.AllClusterRights.Intersect(ttnpb.AllApplicationRights).Sub(rights).Rights, should.BeEmpty)
		}
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

		a.So(err, should.BeNil)
		if a.So(rights, should.NotBeNil) {
			a.So(rights.Rights, should.NotBeEmpty)
		}

		modifiedApplicationID := applicationID
		modifiedApplicationID.ApplicationID = reverse(modifiedApplicationID.ApplicationID)

		rights, err = reg.ListRights(ctx, &modifiedApplicationID, creds)

		a.So(err, should.BeNil)
		if a.So(rights, should.NotBeNil) {
			a.So(rights.Rights, should.BeEmpty)
		}

		applicationAPIKeys := applicationAPIKeys(&applicationID)
		applicationKey := applicationAPIKeys.APIKeys[0]

		APIKey, err := reg.GetAPIKey(ctx, &ttnpb.GetApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			KeyID:                  applicationKey.ID,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(APIKey, should.NotBeNil) {
			a.So(APIKey.ID, should.Equal, applicationKey.ID)
			a.So(APIKey.Key, should.BeEmpty)
		}

		APIKeys, err := reg.ListAPIKeys(ctx, &ttnpb.ListApplicationAPIKeysRequest{
			ApplicationIdentifiers: applicationID,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(APIKeys, should.NotBeNil) {
			a.So(len(APIKeys.APIKeys), should.Equal, len(applicationAPIKeys.APIKeys))
			for i, APIkey := range APIKeys.APIKeys {
				a.So(APIkey.Name, should.Equal, applicationAPIKeys.APIKeys[i].Name)
				a.So(APIkey.ID, should.Equal, applicationAPIKeys.APIKeys[i].ID)
			}
		}

		collaborators, err := reg.ListCollaborators(ctx, &ttnpb.ListApplicationCollaboratorsRequest{
			ApplicationIdentifiers: applicationID,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(collaborators, should.NotBeNil) {
			a.So(collaborators.Collaborators, should.NotBeEmpty)
		}

		APIKeyName := "test-application-api-key-name"
		APIKey, err = reg.CreateAPIKey(ctx, &ttnpb.CreateApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			Name:                   APIKeyName,
			Rights:                 []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(APIKey, should.NotBeNil) {
			a.So(APIKey.Name, should.Equal, APIKeyName)
		}

		newAPIKeyName := "test-new-api-key"
		APIKey.Name = newAPIKeyName
		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			APIKey:                 *APIKey,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(updated, should.NotBeNil) {
			a.So(updated.Name, should.Equal, newAPIKeyName)
		}

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
		if a.So(res, should.NotBeNil) {
			a.So(res.Rights, should.Resemble, []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL})
		}
	})
}

func TestApplicationAccessRights(t *testing.T) {
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
					ttnpb.RIGHT_APPLICATION_LINK,
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

		APIKey, err := reg.CreateAPIKey(ctx, &ttnpb.CreateApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			Rights:                 []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL},
		}, usrCreds)

		a.So(err, should.BeNil)
		if a.So(APIKey, should.NotBeNil) && a.So(APIKey.Rights, should.NotBeNil) {
			a.So(APIKey.Rights, should.Resemble, []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL})
		}

		// Try revoking rights for the collaborator with RIGHT_APPLICATION_ALL without having it
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIdentifiers: applicationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *removedCollaboratorID,
				Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_LINK,
					ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS,
					ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS,
				},
			},
		}, collaboratorCreds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		// Try revoking rights for the api key with RIGHT_APPLICATION_ALL without having it
		_, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			APIKey: ttnpb.APIKey{
				ID: APIKey.ID,
				Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_LINK,
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

		// Remove RIGHT_APPLICATION_ALL from api key to be removed
		key, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			APIKey: ttnpb.APIKey{
				ID:     APIKey.ID,
				Rights: newRights.Rights,
			},
		}, usrCreds)

		a.So(err, should.BeNil)
		if a.So(key, should.NotBeNil) && a.So(key.Rights, should.NotBeNil) {
			a.So(key.Rights, should.Resemble, newRights.Rights)
		}

		newRights = newRights.Sub(ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_LINK))
		key, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			APIKey: ttnpb.APIKey{
				ID:     APIKey.ID,
				Rights: newRights.Rights,
			},
		}, collaboratorCreds)

		a.So(err, should.BeNil)
		if a.So(key, should.NotBeNil) && a.So(key.Rights, should.NotBeNil) {
			a.So(key.Rights, should.Resemble, newRights.Rights)
		}

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIdentifiers: applicationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *removedCollaboratorID,
				Rights:                        newRights.Rights,
			},
		}, collaboratorCreds)

		a.So(err, should.BeNil)

		// Try revoking RIGHT_APPLICATION_DELETE from collaborator without having it
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIdentifiers: applicationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *removedCollaboratorID,
				Rights:                        newRights.Sub(ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_DELETE)).Rights,
			},
		}, collaboratorCreds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		// Try revoking RIGHT_APPLICATION_DELETE from api key without having it
		_, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			APIKey: ttnpb.APIKey{
				ID:     APIKey.ID,
				Rights: newRights.Sub(ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_DELETE)).Rights,
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

		// Delete api key with more rights
		_, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			APIKey: ttnpb.APIKey{
				ID:     APIKey.ID,
				Rights: []ttnpb.Right{},
			},
		}, collaboratorCreds)

		a.So(err, should.BeNil)

		_, err = reg.GetAPIKey(ctx, &ttnpb.GetApplicationAPIKeyRequest{
			ApplicationIdentifiers: applicationID,
			KeyID:                  APIKey.ID,
		}, collaboratorCreds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
	})
}
