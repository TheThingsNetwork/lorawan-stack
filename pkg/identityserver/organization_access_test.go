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
	organizationAccessUser.Admin = false
	organizationAccessUser.State = ttnpb.State_STATE_APPROVED
	for _, apiKey := range userAPIKeys(organizationAccessUser.GetIds()).ApiKeys {
		apiKey.Rights = []ttnpb.Right{
			ttnpb.Right_RIGHT_APPLICATION_LINK,
			ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_API_KEYS,
			ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_MEMBERS,
		}
	}

	orgAccessCollaboratorUser.Admin = false
	orgAccessCollaboratorUser.State = ttnpb.State_STATE_APPROVED
	for _, apiKey := range userAPIKeys(orgAccessCollaboratorUser.GetIds()).ApiKeys {
		apiKey.Rights = []ttnpb.Right{
			ttnpb.Right_RIGHT_ORGANIZATION_ALL,
		}
	}
}

func TestOrganizationAccessNotFound(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, creds := defaultUser.GetIds(), userCreds(defaultUserIdx)
		organizationID := userOrganizations(userID).Organizations[0].GetIds()

		reg := ttnpb.NewOrganizationAccessClient(cc)

		apiKey := ttnpb.APIKey{
			Id:   "does-not-exist-id",
			Name: "test-application-api-key-name",
		}

		got, err := reg.GetAPIKey(ctx, &ttnpb.GetOrganizationAPIKeyRequest{
			OrganizationIds: organizationID,
			KeyId:           apiKey.Id,
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		a.So(got, should.BeNil)

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
			OrganizationIds: organizationID,
			ApiKey:          &apiKey,
			FieldMask:       &pbtypes.FieldMask{Paths: []string{"name"}},
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		a.So(updated, should.BeNil)
	})
}

func TestOrganizationAccessRightsPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, creds := organizationAccessUser.GetIds(), userCreds(organizationAccessUserIdx)
		organizationID := userOrganizations(userID).Organizations[0].GetIds()
		collaboratorID := collaboratorUser.GetIds().OrganizationOrUserIdentifiers()

		reg := ttnpb.NewOrganizationAccessClient(cc)

		apiKeyName := "test-organization-api-key-name"
		apiKey, err := reg.CreateAPIKey(ctx, &ttnpb.CreateOrganizationAPIKeyRequest{
			OrganizationIds: organizationID,
			Name:            apiKeyName,
			Rights:          []ttnpb.Right{ttnpb.Right_RIGHT_ORGANIZATION_ALL},
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(apiKey, should.BeNil)

		// Choose right that the user does not have and hence cannot add
		right := ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_BASIC
		apiKey = organizationAPIKeys(organizationID).ApiKeys[0]

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
			OrganizationIds: organizationID,
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

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIds: organizationID,
			Collaborator: &ttnpb.Collaborator{
				Ids:    collaboratorID,
				Rights: []ttnpb.Right{ttnpb.Right_RIGHT_ORGANIZATION_ALL},
			},
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
	})
}

func TestOrganizationAccessPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := defaultUser.GetIds()
		organizationID := userOrganizations(userID).Organizations[0].GetIds()
		collaboratorID := collaboratorUser.GetIds().OrganizationOrUserIdentifiers()
		apiKeyID := organizationAPIKeys(organizationID).ApiKeys[0].Id

		reg := ttnpb.NewOrganizationAccessClient(cc)

		rights, err := reg.ListRights(ctx, organizationID)

		a.So(err, should.BeNil)
		if a.So(rights, should.NotBeNil) {
			a.So(rights.Rights, should.BeEmpty)
		}

		apiKey, err := reg.GetAPIKey(ctx, &ttnpb.GetOrganizationAPIKeyRequest{
			OrganizationIds: organizationID,
			KeyId:           apiKeyID,
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(apiKey, should.BeNil)

		apiKeys, err := reg.ListAPIKeys(ctx, &ttnpb.ListOrganizationAPIKeysRequest{
			OrganizationIds: organizationID,
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(apiKeys, should.BeNil)

		collaborators, err := reg.ListCollaborators(ctx, &ttnpb.ListOrganizationCollaboratorsRequest{
			OrganizationIds: organizationID,
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsUnauthenticated(err), should.BeTrue)
		}
		a.So(collaborators, should.BeNil)

		apiKeyName := "test-organization-api-key-name"
		apiKey, err = reg.CreateAPIKey(ctx, &ttnpb.CreateOrganizationAPIKeyRequest{
			OrganizationIds: organizationID,
			Name:            apiKeyName,
			Rights:          []ttnpb.Right{ttnpb.Right_RIGHT_ORGANIZATION_ALL},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(apiKey, should.BeNil)

		apiKey = organizationAPIKeys(organizationID).ApiKeys[0]

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
			OrganizationIds: organizationID,
			ApiKey:          apiKey,
			FieldMask:       &pbtypes.FieldMask{Paths: []string{"rights", "name"}},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(updated, should.BeNil)

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIds: organizationID,
			Collaborator: &ttnpb.Collaborator{
				Ids:    collaboratorID,
				Rights: []ttnpb.Right{ttnpb.Right_RIGHT_ORGANIZATION_ALL},
			},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
	})
}

func TestOrganizationAccessClusterAuth(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := defaultUser.GetIds()
		organizationID := userOrganizations(userID).Organizations[0].GetIds()

		reg := ttnpb.NewOrganizationAccessClient(cc)

		rights, err := reg.ListRights(ctx, organizationID, is.WithClusterAuth())

		a.So(err, should.BeNil)
		if a.So(rights, should.NotBeNil) {
			a.So(ttnpb.AllClusterRights.Intersect(ttnpb.AllOrganizationRights).Sub(rights).Rights, should.BeEmpty)
		}
	})
}

func TestOrganizationAccessCRUD(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, creds := defaultUser.GetIds(), userCreds(defaultUserIdx)
		organizationID := userOrganizations(userID).Organizations[0].GetIds()
		collaboratorID := collaboratorUser.GetIds().OrganizationOrUserIdentifiers()

		reg := ttnpb.NewOrganizationAccessClient(cc)

		rights, err := reg.ListRights(ctx, organizationID, creds)

		a.So(err, should.BeNil)
		if a.So(rights, should.NotBeNil) {
			a.So(rights.Rights, should.Contain, ttnpb.Right_RIGHT_ORGANIZATION_ALL)
		}

		modifiedOrganizationID := &ttnpb.OrganizationIdentifiers{OrganizationId: reverse(organizationID.GetOrganizationId())}

		rights, err = reg.ListRights(ctx, modifiedOrganizationID, creds)

		a.So(err, should.BeNil)
		if a.So(rights, should.NotBeNil) {
			a.So(rights.Rights, should.BeEmpty)
		}

		organizationAPIKeys := organizationAPIKeys(organizationID)
		organizationKey := organizationAPIKeys.ApiKeys[0]

		apiKey, err := reg.GetAPIKey(ctx, &ttnpb.GetOrganizationAPIKeyRequest{
			OrganizationIds: organizationID,
			KeyId:           organizationKey.Id,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(apiKey, should.NotBeNil) {
			a.So(apiKey.Id, should.Equal, organizationKey.Id)
			a.So(apiKey.Key, should.BeEmpty)
		}

		apiKeys, err := reg.ListAPIKeys(ctx, &ttnpb.ListOrganizationAPIKeysRequest{
			OrganizationIds: organizationID,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(apiKeys, should.NotBeNil) {
			a.So(len(apiKeys.ApiKeys), should.Equal, len(organizationAPIKeys.ApiKeys))
			for i, APIkey := range apiKeys.ApiKeys {
				a.So(APIkey.Name, should.Equal, organizationAPIKeys.ApiKeys[i].Name)
				a.So(APIkey.Id, should.Equal, organizationAPIKeys.ApiKeys[i].Id)
			}
		}

		collaborators, err := reg.ListCollaborators(ctx, &ttnpb.ListOrganizationCollaboratorsRequest{
			OrganizationIds: organizationID,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(collaborators, should.NotBeNil) {
			a.So(collaborators.Collaborators, should.NotBeEmpty)
		}

		apiKeyName := "test-organization-api-key-name"
		apiKey, err = reg.CreateAPIKey(ctx, &ttnpb.CreateOrganizationAPIKeyRequest{
			OrganizationIds: organizationID,
			Name:            apiKeyName,
			Rights:          []ttnpb.Right{ttnpb.Right_RIGHT_ORGANIZATION_ALL},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(apiKey, should.NotBeNil) {
			a.So(apiKey.Name, should.Equal, apiKeyName)
		}

		newAPIKeyName := "test-new-organization-api-key"
		apiKey.Name = newAPIKeyName
		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
			OrganizationIds: organizationID,
			ApiKey:          apiKey,
			FieldMask:       &pbtypes.FieldMask{Paths: []string{"name"}},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(updated, should.NotBeNil) {
			a.So(updated.Name, should.Equal, newAPIKeyName)
		}

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIds: organizationID,
			Collaborator: &ttnpb.Collaborator{
				Ids:    collaboratorID,
				Rights: []ttnpb.Right{ttnpb.Right_RIGHT_ORGANIZATION_ALL},
			},
		}, creds)

		a.So(err, should.BeNil)

		res, err := reg.GetCollaborator(ctx, &ttnpb.GetOrganizationCollaboratorRequest{
			OrganizationIds: organizationID,
			Collaborator:    collaboratorID,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(res, should.NotBeNil) {
			a.So(res.Rights, should.Resemble, []ttnpb.Right{ttnpb.Right_RIGHT_ORGANIZATION_ALL})
		}

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIds: organizationID,
			Collaborator: &ttnpb.Collaborator{
				Ids: collaboratorID,
			},
		}, creds)

		a.So(err, should.BeNil)

		res, err = reg.GetCollaborator(ctx, &ttnpb.GetOrganizationCollaboratorRequest{
			OrganizationIds: organizationID,
			Collaborator:    collaboratorID,
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
	})
}

func TestOrganizationAccessRights(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, usrCreds := defaultUser.GetIds(), userCreds(defaultUserIdx)
		organizationID := userOrganizations(userID).Organizations[0].GetIds()
		collaboratorID := organizationAccessUser.GetIds().OrganizationOrUserIdentifiers()
		collaboratorCreds := userCreds(organizationAccessUserIdx)
		removedCollaboratorID := orgAccessCollaboratorUser.GetIds().OrganizationOrUserIdentifiers()

		reg := ttnpb.NewOrganizationAccessClient(cc)

		_, err := reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIds: organizationID,
			Collaborator: &ttnpb.Collaborator{
				Ids: collaboratorID,
				Rights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_LINK,
					ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_API_KEYS,
					ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_MEMBERS,
				},
			},
		}, usrCreds)

		a.So(err, should.BeNil)

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIds: organizationID,
			Collaborator: &ttnpb.Collaborator{
				Ids: removedCollaboratorID,
				Rights: []ttnpb.Right{
					ttnpb.Right_RIGHT_ORGANIZATION_ALL,
				},
			},
		}, usrCreds)

		a.So(err, should.BeNil)

		apiKey, err := reg.CreateAPIKey(ctx, &ttnpb.CreateOrganizationAPIKeyRequest{
			OrganizationIds: organizationID,
			Rights:          []ttnpb.Right{ttnpb.Right_RIGHT_ORGANIZATION_ALL},
		}, usrCreds)

		a.So(err, should.BeNil)
		if a.So(apiKey, should.NotBeNil) && a.So(apiKey.Rights, should.NotBeNil) {
			a.So(apiKey.Rights, should.Resemble, []ttnpb.Right{ttnpb.Right_RIGHT_ORGANIZATION_ALL})
		}

		// Try revoking rights for the collaborator with RIGHT_ORGANIZATION_ALL without having it
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIds: organizationID,
			Collaborator: &ttnpb.Collaborator{
				Ids: removedCollaboratorID,
				Rights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_LINK,
					ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_API_KEYS,
					ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_MEMBERS,
				},
			},
		}, collaboratorCreds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		// Try revoking rights for the api key with RIGHT_ORGANIZATION_ALL without having it
		_, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
			OrganizationIds: organizationID,
			ApiKey: &ttnpb.APIKey{
				Id: apiKey.Id,
				Rights: []ttnpb.Right{
					ttnpb.Right_RIGHT_APPLICATION_LINK,
					ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_API_KEYS,
					ttnpb.Right_RIGHT_ORGANIZATION_SETTINGS_MEMBERS,
				},
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"rights"}},
		}, collaboratorCreds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		// Remove RIGHT_ORGANIZATION_ALL from collaborator to be removed
		newRights := ttnpb.AllOrganizationRights.Sub(ttnpb.RightsFrom(ttnpb.Right_RIGHT_ORGANIZATION_ALL))
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIds: organizationID,
			Collaborator: &ttnpb.Collaborator{
				Ids:    removedCollaboratorID,
				Rights: newRights.Rights,
			},
		}, usrCreds)

		a.So(err, should.BeNil)

		// Remove RIGHT_ORGANIZATION_ALL from api key to be removed
		key, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
			OrganizationIds: organizationID,
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

		newRights = newRights.Sub(ttnpb.RightsFrom(ttnpb.Right_RIGHT_APPLICATION_LINK))
		key, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
			OrganizationIds: organizationID,
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

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIds: organizationID,
			Collaborator: &ttnpb.Collaborator{
				Ids:    removedCollaboratorID,
				Rights: newRights.Rights,
			},
		}, collaboratorCreds)

		a.So(err, should.BeNil)

		// Try revoking RIGHT_ORGANIZATION_DELETE without having it
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIds: organizationID,
			Collaborator: &ttnpb.Collaborator{
				Ids:    removedCollaboratorID,
				Rights: newRights.Sub(ttnpb.RightsFrom(ttnpb.Right_RIGHT_ORGANIZATION_DELETE)).Rights,
			},
		}, collaboratorCreds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		// Try revoking RIGHT_ORGANIZATION_DELETE from api key without having it
		_, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
			OrganizationIds: organizationID,
			ApiKey: &ttnpb.APIKey{
				Id:     apiKey.Id,
				Rights: newRights.Sub(ttnpb.RightsFrom(ttnpb.Right_RIGHT_ORGANIZATION_DELETE)).Rights,
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"rights"}},
		}, collaboratorCreds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		res, err := reg.GetCollaborator(ctx, &ttnpb.GetOrganizationCollaboratorRequest{
			OrganizationIds: organizationID,
			Collaborator:    removedCollaboratorID,
		}, collaboratorCreds)

		if a.So(err, should.BeNil) {
			a.So(res.Rights, should.Resemble, newRights.Rights)
		}

		// Delete collaborator with more rights
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIds: organizationID,
			Collaborator: &ttnpb.Collaborator{
				Ids:    removedCollaboratorID,
				Rights: []ttnpb.Right{},
			},
		}, collaboratorCreds)

		a.So(err, should.BeNil)

		_, err = reg.GetCollaborator(ctx, &ttnpb.GetOrganizationCollaboratorRequest{
			OrganizationIds: organizationID,
			Collaborator:    removedCollaboratorID,
		}, collaboratorCreds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		// Delete api key with more rights
		_, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
			OrganizationIds: organizationID,
			ApiKey: &ttnpb.APIKey{
				Id:     apiKey.Id,
				Rights: []ttnpb.Right{},
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"rights"}},
		}, collaboratorCreds)

		a.So(err, should.BeNil)

		_, err = reg.GetAPIKey(ctx, &ttnpb.GetOrganizationAPIKeyRequest{
			OrganizationIds: organizationID,
			KeyId:           apiKey.Id,
		}, collaboratorCreds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
	})
}
