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

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"google.golang.org/grpc"
)

func init() {
	applicationAccessUser.Admin = false
	applicationAccessUser.State = ttnpb.State_STATE_APPROVED
	for _, apiKey := range userAPIKeys(applicationAccessUser.GetIds()).ApiKeys {
		apiKey.Rights = []ttnpb.Right{
			ttnpb.RIGHT_APPLICATION_LINK,
			ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS,
			ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS,
		}
	}
	appAccessCollaboratorUser.Admin = false
	appAccessCollaboratorUser.State = ttnpb.State_STATE_APPROVED
	for _, apiKey := range userAPIKeys(appAccessCollaboratorUser.GetIds()).ApiKeys {
		apiKey.Rights = []ttnpb.Right{
			ttnpb.RIGHT_APPLICATION_ALL,
		}
	}
}

func TestApplicationAccessNotFound(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, creds := defaultUser.GetIds(), userCreds(defaultUserIdx)
		applicationID := userApplications(userID).Applications[0].GetIds()

		reg := ttnpb.NewApplicationAccessClient(cc)

		apiKey := ttnpb.APIKey{
			Id:   "does-not-exist-id",
			Name: "test-application-api-key-name",
		}

		got, err := reg.GetAPIKey(ctx, &ttnpb.GetApplicationAPIKeyRequest{
			ApplicationIds: applicationID,
			KeyId:          apiKey.Id,
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		a.So(got, should.BeNil)

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIds: applicationID,
			ApiKey:         &apiKey,
			FieldMask:      &pbtypes.FieldMask{Paths: []string{"rights", "name"}},
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
		userID, creds := applicationAccessUser.GetIds(), userCreds(applicationAccessUserIdx)
		applicationID := userApplications(userID).Applications[0].GetIds()
		collaboratorID := collaboratorUser.GetIds().OrganizationOrUserIdentifiers()

		reg := ttnpb.NewApplicationAccessClient(cc)

		apiKeyName := "test-application-api-key-name"
		apiKey, err := reg.CreateAPIKey(ctx, &ttnpb.CreateApplicationAPIKeyRequest{
			ApplicationIds: applicationID,
			Name:           apiKeyName,
			Rights:         []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL},
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(apiKey, should.BeNil)

		// Choose right that the user does not have and hence cannot add
		right := ttnpb.RIGHT_APPLICATION_SETTINGS_BASIC
		apiKey = applicationAPIKeys(applicationID).ApiKeys[0]

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIds: applicationID,
			ApiKey: &ttnpb.APIKey{
				Id:     apiKey.Id,
				Name:   apiKey.Name,
				Rights: []ttnpb.Right{right},
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"rights", "name"}},
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(updated, should.BeNil)

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIds: applicationID,
			Collaborator: &ttnpb.Collaborator{
				Ids:    collaboratorID,
				Rights: []ttnpb.Right{right},
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
		userID := defaultUser.GetIds()
		applicationID := userApplications(userID).Applications[0].GetIds()
		collaboratorID := collaboratorUser.GetIds().OrganizationOrUserIdentifiers()
		apiKeyID := applicationAPIKeys(applicationID).ApiKeys[0].Id

		reg := ttnpb.NewApplicationAccessClient(cc)

		rights, err := reg.ListRights(ctx, applicationID)

		a.So(err, should.BeNil)
		if a.So(rights, should.NotBeNil) {
			a.So(rights.Rights, should.BeEmpty)
		}

		apiKey, err := reg.GetAPIKey(ctx, &ttnpb.GetApplicationAPIKeyRequest{
			ApplicationIds: applicationID,
			KeyId:          apiKeyID,
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(apiKey, should.BeNil)

		apiKeys, err := reg.ListAPIKeys(ctx, &ttnpb.ListApplicationAPIKeysRequest{
			ApplicationIds: applicationID,
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(apiKeys, should.BeNil)

		collaborators, err := reg.ListCollaborators(ctx, &ttnpb.ListApplicationCollaboratorsRequest{
			ApplicationIds: applicationID,
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsUnauthenticated(err), should.BeTrue)
		}
		a.So(collaborators, should.BeNil)

		apiKeyName := "test-application-api-key-name"
		apiKey, err = reg.CreateAPIKey(ctx, &ttnpb.CreateApplicationAPIKeyRequest{
			ApplicationIds: applicationID,
			Name:           apiKeyName,
			Rights:         []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(apiKey, should.BeNil)

		apiKey = applicationAPIKeys(applicationID).ApiKeys[0]

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIds: applicationID,
			ApiKey:         apiKey,
			FieldMask:      &pbtypes.FieldMask{Paths: []string{"name", "rights"}},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(updated, should.BeNil)

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIds: applicationID,
			Collaborator: &ttnpb.Collaborator{
				Ids:    collaboratorID,
				Rights: []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL},
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
		userID := defaultUser.GetIds()
		applicationID := userApplications(userID).Applications[0].GetIds()

		reg := ttnpb.NewApplicationAccessClient(cc)

		rights, err := reg.ListRights(ctx, applicationID, is.WithClusterAuth())

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
		userID, creds := defaultUser.GetIds(), userCreds(defaultUserIdx)
		applicationID := userApplications(userID).Applications[0].GetIds()
		collaboratorID := collaboratorUser.GetIds().OrganizationOrUserIdentifiers()

		reg := ttnpb.NewApplicationAccessClient(cc)

		rights, err := reg.ListRights(ctx, applicationID, creds)

		a.So(err, should.BeNil)
		if a.So(rights, should.NotBeNil) {
			a.So(rights.Rights, should.Contain, ttnpb.RIGHT_APPLICATION_ALL)
		}

		modifiedApplicationID := &ttnpb.ApplicationIdentifiers{ApplicationId: reverse(applicationID.GetApplicationId())}

		rights, err = reg.ListRights(ctx, modifiedApplicationID, creds)

		a.So(err, should.BeNil)
		if a.So(rights, should.NotBeNil) {
			a.So(rights.Rights, should.BeEmpty)
		}

		applicationAPIKeys := applicationAPIKeys(applicationID)
		applicationKey := applicationAPIKeys.ApiKeys[0]

		apiKey, err := reg.GetAPIKey(ctx, &ttnpb.GetApplicationAPIKeyRequest{
			ApplicationIds: applicationID,
			KeyId:          applicationKey.Id,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(apiKey, should.NotBeNil) {
			a.So(apiKey.Id, should.Equal, applicationKey.Id)
			a.So(apiKey.Key, should.BeEmpty)
		}

		apiKeys, err := reg.ListAPIKeys(ctx, &ttnpb.ListApplicationAPIKeysRequest{
			ApplicationIds: applicationID,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(apiKeys, should.NotBeNil) {
			a.So(len(apiKeys.ApiKeys), should.Equal, len(applicationAPIKeys.ApiKeys))
			for i, APIkey := range apiKeys.ApiKeys {
				a.So(APIkey.Name, should.Equal, applicationAPIKeys.ApiKeys[i].Name)
				a.So(APIkey.Id, should.Equal, applicationAPIKeys.ApiKeys[i].Id)
			}
		}

		collaborators, err := reg.ListCollaborators(ctx, &ttnpb.ListApplicationCollaboratorsRequest{
			ApplicationIds: applicationID,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(collaborators, should.NotBeNil) {
			a.So(collaborators.Collaborators, should.NotBeEmpty)
		}

		apiKeyName := "test-application-api-key-name"
		apiKey, err = reg.CreateAPIKey(ctx, &ttnpb.CreateApplicationAPIKeyRequest{
			ApplicationIds: applicationID,
			Name:           apiKeyName,
			Rights:         []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(apiKey, should.NotBeNil) {
			a.So(apiKey.Name, should.Equal, apiKeyName)
		}

		newAPIKeyName := "test-new-api-key"
		apiKey.Name = newAPIKeyName
		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIds: applicationID,
			ApiKey:         apiKey,
			FieldMask:      &pbtypes.FieldMask{Paths: []string{"name"}},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(updated, should.NotBeNil) {
			a.So(updated.Name, should.Equal, newAPIKeyName)
		}

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIds: applicationID,
			Collaborator: &ttnpb.Collaborator{
				Ids:    collaboratorID,
				Rights: []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL},
			},
		}, creds)

		a.So(err, should.BeNil)

		res, err := reg.GetCollaborator(ctx, &ttnpb.GetApplicationCollaboratorRequest{
			ApplicationIds: applicationID,
			Collaborator:   collaboratorID,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(res, should.NotBeNil) {
			a.So(res.Rights, should.Resemble, []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL})
		}

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIds: applicationID,
			Collaborator: &ttnpb.Collaborator{
				Ids: collaboratorID,
			},
		}, creds)

		a.So(err, should.BeNil)

		res, err = reg.GetCollaborator(ctx, &ttnpb.GetApplicationCollaboratorRequest{
			ApplicationIds: applicationID,
			Collaborator:   collaboratorID,
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
	})
}

func TestApplicationAccessRights(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, usrCreds := defaultUser.GetIds(), userCreds(defaultUserIdx)
		applicationID := userApplications(userID).Applications[0].GetIds()
		collaboratorID := applicationAccessUser.GetIds().OrganizationOrUserIdentifiers()
		collaboratorCreds := userCreds(applicationAccessUserIdx)
		removedCollaboratorID := appAccessCollaboratorUser.GetIds().OrganizationOrUserIdentifiers()

		reg := ttnpb.NewApplicationAccessClient(cc)

		_, err := reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIds: applicationID,
			Collaborator: &ttnpb.Collaborator{
				Ids: collaboratorID,
				Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_LINK,
					ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS,
					ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS,
				},
			},
		}, usrCreds)

		a.So(err, should.BeNil)

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIds: applicationID,
			Collaborator: &ttnpb.Collaborator{
				Ids: removedCollaboratorID,
				Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_ALL,
				},
			},
		}, usrCreds)

		a.So(err, should.BeNil)

		apiKey, err := reg.CreateAPIKey(ctx, &ttnpb.CreateApplicationAPIKeyRequest{
			ApplicationIds: applicationID,
			Rights:         []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL},
		}, usrCreds)

		a.So(err, should.BeNil)
		if a.So(apiKey, should.NotBeNil) && a.So(apiKey.Rights, should.NotBeNil) {
			a.So(apiKey.Rights, should.Resemble, []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL})
		}

		// Try revoking rights for the collaborator with RIGHT_APPLICATION_ALL without having it
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIds: applicationID,
			Collaborator: &ttnpb.Collaborator{
				Ids: removedCollaboratorID,
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
			ApplicationIds: applicationID,
			ApiKey: &ttnpb.APIKey{
				Id: apiKey.Id,
				Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_LINK,
					ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS,
					ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS,
				},
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"rights"}},
		}, collaboratorCreds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		// Remove RIGHT_APPLICATION_ALL from collaborator to be removed
		newRights := ttnpb.AllApplicationRights.Sub(ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_ALL))
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIds: applicationID,
			Collaborator: &ttnpb.Collaborator{
				Ids:    removedCollaboratorID,
				Rights: newRights.Rights,
			},
		}, usrCreds)

		a.So(err, should.BeNil)

		// Remove RIGHT_APPLICATION_ALL from api key to be removed
		key, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIds: applicationID,
			ApiKey: &ttnpb.APIKey{
				Id:     apiKey.Id,
				Rights: newRights.Rights,
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"rights"}},
		}, usrCreds)

		a.So(err, should.BeNil)
		if a.So(key, should.NotBeNil) && a.So(key.Rights, should.NotBeNil) {
			a.So(key.Rights, should.Resemble, newRights.Rights)
		}

		newRights = newRights.Sub(ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_LINK))
		key, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIds: applicationID,
			ApiKey: &ttnpb.APIKey{
				Id:     apiKey.Id,
				Rights: newRights.Rights,
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"rights"}},
		}, collaboratorCreds)

		a.So(err, should.BeNil)
		if a.So(key, should.NotBeNil) && a.So(key.Rights, should.NotBeNil) {
			a.So(key.Rights, should.Resemble, newRights.Rights)
		}

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIds: applicationID,
			Collaborator: &ttnpb.Collaborator{
				Ids:    removedCollaboratorID,
				Rights: newRights.Rights,
			},
		}, collaboratorCreds)

		a.So(err, should.BeNil)

		// Try revoking RIGHT_APPLICATION_DELETE from collaborator without having it
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIds: applicationID,
			Collaborator: &ttnpb.Collaborator{
				Ids:    removedCollaboratorID,
				Rights: newRights.Sub(ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_DELETE)).Rights,
			},
		}, collaboratorCreds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		// Try revoking RIGHT_APPLICATION_DELETE from api key without having it
		_, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIds: applicationID,
			ApiKey: &ttnpb.APIKey{
				Id:     apiKey.Id,
				Rights: newRights.Sub(ttnpb.RightsFrom(ttnpb.RIGHT_APPLICATION_DELETE)).Rights,
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"rights"}},
		}, collaboratorCreds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		res, err := reg.GetCollaborator(ctx, &ttnpb.GetApplicationCollaboratorRequest{
			ApplicationIds: applicationID,
			Collaborator:   removedCollaboratorID,
		}, collaboratorCreds)

		if a.So(err, should.BeNil) {
			a.So(res.Rights, should.Resemble, newRights.Rights)
		}

		// Delete collaborator with more rights
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetApplicationCollaboratorRequest{
			ApplicationIds: applicationID,
			Collaborator: &ttnpb.Collaborator{
				Ids:    removedCollaboratorID,
				Rights: []ttnpb.Right{},
			},
		}, collaboratorCreds)

		a.So(err, should.BeNil)

		_, err = reg.GetCollaborator(ctx, &ttnpb.GetApplicationCollaboratorRequest{
			ApplicationIds: applicationID,
			Collaborator:   removedCollaboratorID,
		}, collaboratorCreds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		// Delete api key with more rights
		_, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
			ApplicationIds: applicationID,
			ApiKey: &ttnpb.APIKey{
				Id:     apiKey.Id,
				Rights: []ttnpb.Right{},
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"rights"}},
		}, collaboratorCreds)

		a.So(err, should.BeNil)

		_, err = reg.GetAPIKey(ctx, &ttnpb.GetApplicationAPIKeyRequest{
			ApplicationIds: applicationID,
			KeyId:          apiKey.Id,
		}, collaboratorCreds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
	})
}
