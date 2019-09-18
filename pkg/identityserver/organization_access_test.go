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
			ttnpb.RIGHT_APPLICATION_LINK,
			ttnpb.RIGHT_ORGANIZATION_SETTINGS_API_KEYS,
			ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS,
		}
	}

	orgAccessCollaboratorUser.Admin = false
	orgAccessCollaboratorUser.State = ttnpb.STATE_APPROVED
	for _, apiKey := range userAPIKeys(&orgAccessCollaboratorUser.UserIdentifiers).APIKeys {
		apiKey.Rights = []ttnpb.Right{
			ttnpb.RIGHT_ORGANIZATION_ALL,
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

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		a.So(got, should.BeNil)

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
			APIKey:                  apiKey,
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

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(APIKey, should.BeNil)

		// Choose right that the user does not have and hence cannot add
		right := ttnpb.RIGHT_ORGANIZATION_SETTINGS_BASIC
		APIKey = organizationAPIKeys(&organizationID).APIKeys[0]

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
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

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIdentifiers: organizationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *collaboratorID,
				Rights:                        []ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_ALL},
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
		userID := defaultUser.UserIdentifiers
		organizationID := userOrganizations(&userID).Organizations[0].OrganizationIdentifiers
		collaboratorID := collaboratorUser.UserIdentifiers.OrganizationOrUserIdentifiers()
		APIKeyID := organizationAPIKeys(&organizationID).APIKeys[0].ID

		reg := ttnpb.NewOrganizationAccessClient(cc)

		rights, err := reg.ListRights(ctx, &organizationID)

		a.So(err, should.BeNil)
		if a.So(rights, should.NotBeNil) {
			a.So(rights.Rights, should.BeEmpty)
		}

		APIKey, err := reg.GetAPIKey(ctx, &ttnpb.GetOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
			KeyID:                   APIKeyID,
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(APIKey, should.BeNil)

		APIKeys, err := reg.ListAPIKeys(ctx, &ttnpb.ListOrganizationAPIKeysRequest{
			OrganizationIdentifiers: organizationID,
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(APIKeys, should.BeNil)

		collaborators, err := reg.ListCollaborators(ctx, &ttnpb.ListOrganizationCollaboratorsRequest{
			OrganizationIdentifiers: organizationID,
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(collaborators, should.BeNil)

		APIKeyName := "test-organization-api-key-name"
		APIKey, err = reg.CreateAPIKey(ctx, &ttnpb.CreateOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
			Name:                    APIKeyName,
			Rights:                  []ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_ALL},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(APIKey, should.BeNil)

		APIKey = organizationAPIKeys(&organizationID).APIKeys[0]

		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
			APIKey:                  *APIKey,
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
		a.So(updated, should.BeNil)

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIdentifiers: organizationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *collaboratorID,
				Rights:                        []ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_ALL},
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
		userID := defaultUser.UserIdentifiers
		organizationID := userOrganizations(&userID).Organizations[0].OrganizationIdentifiers

		reg := ttnpb.NewOrganizationAccessClient(cc)

		rights, err := reg.ListRights(ctx, &organizationID, is.WithClusterAuth())

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
		userID, creds := defaultUser.UserIdentifiers, userCreds(defaultUserIdx)
		organizationID := userOrganizations(&userID).Organizations[0].OrganizationIdentifiers
		collaboratorID := collaboratorUser.UserIdentifiers.OrganizationOrUserIdentifiers()

		reg := ttnpb.NewOrganizationAccessClient(cc)

		rights, err := reg.ListRights(ctx, &organizationID, creds)

		a.So(err, should.BeNil)
		if a.So(rights, should.NotBeNil) {
			a.So(rights.Rights, should.Contain, ttnpb.RIGHT_ORGANIZATION_ALL)
		}

		modifiedOrganizationID := organizationID
		modifiedOrganizationID.OrganizationID = reverse(modifiedOrganizationID.OrganizationID)

		rights, err = reg.ListRights(ctx, &modifiedOrganizationID, creds)

		a.So(err, should.BeNil)
		if a.So(rights, should.NotBeNil) {
			a.So(rights.Rights, should.BeEmpty)
		}

		organizationAPIKeys := organizationAPIKeys(&organizationID)
		organizationKey := organizationAPIKeys.APIKeys[0]

		APIKey, err := reg.GetAPIKey(ctx, &ttnpb.GetOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
			KeyID:                   organizationKey.ID,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(APIKey, should.NotBeNil) {
			a.So(APIKey.ID, should.Equal, organizationKey.ID)
			a.So(APIKey.Key, should.BeEmpty)
		}

		APIKeys, err := reg.ListAPIKeys(ctx, &ttnpb.ListOrganizationAPIKeysRequest{
			OrganizationIdentifiers: organizationID,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(APIKeys, should.NotBeNil) {
			a.So(len(APIKeys.APIKeys), should.Equal, len(organizationAPIKeys.APIKeys))
			for i, APIkey := range APIKeys.APIKeys {
				a.So(APIkey.Name, should.Equal, organizationAPIKeys.APIKeys[i].Name)
				a.So(APIkey.ID, should.Equal, organizationAPIKeys.APIKeys[i].ID)
			}
		}

		collaborators, err := reg.ListCollaborators(ctx, &ttnpb.ListOrganizationCollaboratorsRequest{
			OrganizationIdentifiers: organizationID,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(collaborators, should.NotBeNil) {
			a.So(collaborators.Collaborators, should.NotBeEmpty)
		}

		APIKeyName := "test-organization-api-key-name"
		APIKey, err = reg.CreateAPIKey(ctx, &ttnpb.CreateOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
			Name:                    APIKeyName,
			Rights:                  []ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_ALL},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(APIKey, should.NotBeNil) {
			a.So(APIKey.Name, should.Equal, APIKeyName)
		}

		newAPIKeyName := "test-new-organization-api-key"
		APIKey.Name = newAPIKeyName
		updated, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
			APIKey:                  *APIKey,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(updated, should.NotBeNil) {
			a.So(updated.Name, should.Equal, newAPIKeyName)
		}

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIdentifiers: organizationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *collaboratorID,
				Rights:                        []ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_ALL},
			},
		}, creds)

		a.So(err, should.BeNil)

		res, err := reg.GetCollaborator(ctx, &ttnpb.GetOrganizationCollaboratorRequest{
			OrganizationIdentifiers:       organizationID,
			OrganizationOrUserIdentifiers: *collaboratorID,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(res, should.NotBeNil) {
			a.So(res.Rights, should.Resemble, []ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_ALL})
		}
	})
}
func TestOrganizationAccessRights(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID, usrCreds := defaultUser.UserIdentifiers, userCreds(defaultUserIdx)
		organizationID := userOrganizations(&userID).Organizations[0].OrganizationIdentifiers
		collaboratorID := organizationAccessUser.UserIdentifiers.OrganizationOrUserIdentifiers()
		collaboratorCreds := userCreds(organizationAccessUserIdx)
		removedCollaboratorID := orgAccessCollaboratorUser.UserIdentifiers.OrganizationOrUserIdentifiers()

		reg := ttnpb.NewOrganizationAccessClient(cc)

		_, err := reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIdentifiers: organizationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *collaboratorID,
				Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_LINK,
					ttnpb.RIGHT_ORGANIZATION_SETTINGS_API_KEYS,
					ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS,
				},
			},
		}, usrCreds)

		a.So(err, should.BeNil)

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIdentifiers: organizationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *removedCollaboratorID,
				Rights: []ttnpb.Right{
					ttnpb.RIGHT_ORGANIZATION_ALL,
				},
			},
		}, usrCreds)

		a.So(err, should.BeNil)

		APIKey, err := reg.CreateAPIKey(ctx, &ttnpb.CreateOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
			Rights:                  []ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_ALL},
		}, usrCreds)

		a.So(err, should.BeNil)
		if a.So(APIKey, should.NotBeNil) && a.So(APIKey.Rights, should.NotBeNil) {
			a.So(APIKey.Rights, should.Resemble, []ttnpb.Right{ttnpb.RIGHT_ORGANIZATION_ALL})
		}

		// Try revoking rights for the collaborator with RIGHT_ORGANIZATION_ALL without having it
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIdentifiers: organizationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *removedCollaboratorID,
				Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_LINK,
					ttnpb.RIGHT_ORGANIZATION_SETTINGS_API_KEYS,
					ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS,
				},
			},
		}, collaboratorCreds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		// Try revoking rights for the api key with RIGHT_ORGANIZATION_ALL without having it
		_, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
			APIKey: ttnpb.APIKey{
				ID: APIKey.ID,
				Rights: []ttnpb.Right{
					ttnpb.RIGHT_APPLICATION_LINK,
					ttnpb.RIGHT_ORGANIZATION_SETTINGS_API_KEYS,
					ttnpb.RIGHT_ORGANIZATION_SETTINGS_MEMBERS,
				},
			},
		}, collaboratorCreds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		// Remove RIGHT_ORGANIZATION_ALL from collaborator to be removed
		newRights := ttnpb.AllOrganizationRights.Sub(ttnpb.RightsFrom(ttnpb.RIGHT_ORGANIZATION_ALL))
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIdentifiers: organizationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *removedCollaboratorID,
				Rights:                        newRights.Rights,
			},
		}, usrCreds)

		a.So(err, should.BeNil)

		// Remove RIGHT_ORGANIZATION_ALL from api key to be removed
		key, err := reg.UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
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
		key, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
			APIKey: ttnpb.APIKey{
				ID:     APIKey.ID,
				Rights: newRights.Rights,
			},
		}, collaboratorCreds)

		a.So(err, should.BeNil)
		if a.So(key, should.NotBeNil) && a.So(key.Rights, should.NotBeNil) {
			a.So(key.Rights, should.Resemble, newRights.Rights)
		}

		_, err = reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIdentifiers: organizationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *removedCollaboratorID,
				Rights:                        newRights.Rights,
			},
		}, collaboratorCreds)

		a.So(err, should.BeNil)

		// Try revoking RIGHT_ORGANIZATION_DELETE without having it
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIdentifiers: organizationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *removedCollaboratorID,
				Rights:                        newRights.Sub(ttnpb.RightsFrom(ttnpb.RIGHT_ORGANIZATION_DELETE)).Rights,
			},
		}, collaboratorCreds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		// Try revoking RIGHT_ORGANIZATION_DELETE from api key without having it
		_, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
			APIKey: ttnpb.APIKey{
				ID:     APIKey.ID,
				Rights: newRights.Sub(ttnpb.RightsFrom(ttnpb.RIGHT_ORGANIZATION_DELETE)).Rights,
			},
		}, collaboratorCreds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		res, err := reg.GetCollaborator(ctx, &ttnpb.GetOrganizationCollaboratorRequest{
			OrganizationIdentifiers:       organizationID,
			OrganizationOrUserIdentifiers: *removedCollaboratorID,
		}, collaboratorCreds)

		if a.So(err, should.BeNil) {
			a.So(res.Rights, should.Resemble, newRights.Rights)
		}

		// Delete collaborator with more rights
		_, err = reg.SetCollaborator(ctx, &ttnpb.SetOrganizationCollaboratorRequest{
			OrganizationIdentifiers: organizationID,
			Collaborator: ttnpb.Collaborator{
				OrganizationOrUserIdentifiers: *removedCollaboratorID,
				Rights:                        []ttnpb.Right{},
			},
		}, collaboratorCreds)

		a.So(err, should.BeNil)

		_, err = reg.GetCollaborator(ctx, &ttnpb.GetOrganizationCollaboratorRequest{
			OrganizationIdentifiers:       organizationID,
			OrganizationOrUserIdentifiers: *removedCollaboratorID,
		}, collaboratorCreds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		// Delete api key with more rights
		_, err = reg.UpdateAPIKey(ctx, &ttnpb.UpdateOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
			APIKey: ttnpb.APIKey{
				ID:     APIKey.ID,
				Rights: []ttnpb.Right{},
			},
		}, collaboratorCreds)

		a.So(err, should.BeNil)

		_, err = reg.GetAPIKey(ctx, &ttnpb.GetOrganizationAPIKeyRequest{
			OrganizationIdentifiers: organizationID,
			KeyID:                   APIKey.ID,
		}, collaboratorCreds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
	})
}
