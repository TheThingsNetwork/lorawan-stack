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
	"sort"
	"testing"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var _ ttnpb.IsApplicationServer = new(applicationService)

func TestApplication(t *testing.T) {
	a := assertions.New(t)
	is := newTestIS(t)

	user := newTestUsers()["bob"]

	app := ttnpb.Application{
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "foo-app"},
	}

	ctx := newTestCtx(user.UserIdentifiers)

	_, err := is.applicationService.CreateApplication(ctx, &ttnpb.CreateApplicationRequest{
		Application: app,
	})
	a.So(err, should.BeNil)

	// Can't create applications with blacklisted IDs.
	for _, id := range newTestSettings().BlacklistedIDs {
		_, err = is.applicationService.CreateApplication(ctx, &ttnpb.CreateApplicationRequest{
			Application: ttnpb.Application{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: id},
			},
		})
		a.So(err, should.NotBeNil)
		a.So(err, should.DescribeError, ErrBlacklistedID)
	}

	found, err := is.applicationService.GetApplication(ctx, &ttnpb.ApplicationIdentifiers{ApplicationID: app.ApplicationID})
	a.So(err, should.BeNil)
	a.So(found, should.EqualFieldsWithIgnores(ApplicationGeneratedFields...), app)

	apps, err := is.applicationService.ListApplications(ctx, &ttnpb.ListApplicationsRequest{})
	a.So(err, should.BeNil)
	if a.So(apps.Applications, should.HaveLength, 1) {
		a.So(apps.Applications[0], should.EqualFieldsWithIgnores(ApplicationGeneratedFields...), app)
	}

	app.Description = "foo"
	_, err = is.applicationService.UpdateApplication(ctx, &ttnpb.UpdateApplicationRequest{
		Application: app,
		UpdateMask: pbtypes.FieldMask{
			Paths: []string{"description"},
		},
	})
	a.So(err, should.BeNil)

	found, err = is.applicationService.GetApplication(ctx, &ttnpb.ApplicationIdentifiers{ApplicationID: app.ApplicationID})
	a.So(err, should.BeNil)
	a.So(found, should.EqualFieldsWithIgnores(ApplicationGeneratedFields...), app)

	// Generate a new API key.
	key, err := is.applicationService.GenerateApplicationAPIKey(ctx, &ttnpb.GenerateApplicationAPIKeyRequest{
		ApplicationIdentifiers: app.ApplicationIdentifiers,
		Name:   "foo",
		Rights: ttnpb.AllApplicationRights(),
	})
	a.So(err, should.BeNil)
	a.So(key.Key, should.NotBeEmpty)
	a.So(key.Name, should.Equal, key.Name)
	a.So(key.Rights, should.Resemble, ttnpb.AllApplicationRights())

	// Update API key.
	key.Rights = []ttnpb.Right{ttnpb.Right(10)}
	_, err = is.applicationService.UpdateApplicationAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
		ApplicationIdentifiers: app.ApplicationIdentifiers,
		Name:   key.Name,
		Rights: key.Rights,
	})
	a.So(err, should.BeNil)

	// Can't generate another API key with the same name.
	_, err = is.applicationService.GenerateApplicationAPIKey(ctx, &ttnpb.GenerateApplicationAPIKeyRequest{
		ApplicationIdentifiers: app.ApplicationIdentifiers,
		Name:   key.Name,
		Rights: []ttnpb.Right{ttnpb.Right(1)},
	})
	a.So(err, should.NotBeNil)
	a.So(err, should.DescribeError, store.ErrAPIKeyNameConflict)

	keys, err := is.applicationService.ListApplicationAPIKeys(ctx, &ttnpb.ApplicationIdentifiers{ApplicationID: app.ApplicationID})
	a.So(err, should.BeNil)
	if a.So(keys.APIKeys, should.HaveLength, 1) {
		sort.Slice(keys.APIKeys[0].Rights, func(i, j int) bool { return keys.APIKeys[0].Rights[i] < keys.APIKeys[0].Rights[j] })
		a.So(keys.APIKeys[0], should.Resemble, key)
	}

	_, err = is.applicationService.RemoveApplicationAPIKey(ctx, &ttnpb.RemoveApplicationAPIKeyRequest{
		ApplicationIdentifiers: app.ApplicationIdentifiers,
		Name: key.Name,
	})
	a.So(err, should.BeNil)

	keys, err = is.applicationService.ListApplicationAPIKeys(ctx, &ttnpb.ApplicationIdentifiers{ApplicationID: app.ApplicationID})
	a.So(err, should.BeNil)
	a.So(keys.APIKeys, should.HaveLength, 0)

	// Set a new collaborator with SETTINGS_COLLABORATORS and INFO rights.
	alice := newTestUsers()["alice"]
	collab := &ttnpb.ApplicationCollaborator{
		OrganizationOrUserIdentifiers: ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_UserID{UserID: &alice.UserIdentifiers}},
		ApplicationIdentifiers:        app.ApplicationIdentifiers,
		Rights:                        []ttnpb.Right{ttnpb.RIGHT_APPLICATION_INFO, ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS},
	}

	_, err = is.applicationService.SetApplicationCollaborator(ctx, collab)
	a.So(err, should.BeNil)

	rights, err := is.applicationService.ListApplicationRights(ctx, &ttnpb.ApplicationIdentifiers{ApplicationID: app.ApplicationID})
	a.So(err, should.BeNil)
	a.So(rights.Rights, should.Resemble, ttnpb.AllApplicationRights())

	collabs, err := is.applicationService.ListApplicationCollaborators(ctx, &ttnpb.ApplicationIdentifiers{ApplicationID: app.ApplicationID})
	a.So(err, should.BeNil)
	a.So(collabs.Collaborators, should.HaveLength, 2)
	a.So(collabs.Collaborators, should.Contain, collab)
	a.So(collabs.Collaborators, should.Contain, &ttnpb.ApplicationCollaborator{
		OrganizationOrUserIdentifiers: ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_UserID{UserID: &user.UserIdentifiers}},
		ApplicationIdentifiers:        app.ApplicationIdentifiers,
		Rights:                        ttnpb.AllApplicationRights(),
	})

	// The new collaborator can't grant himself more rights.
	{
		collab.Rights = append(collab.Rights, ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS)

		ctx := newTestCtx(alice.UserIdentifiers)

		_, err = is.applicationService.SetApplicationCollaborator(ctx, collab)
		a.So(err, should.BeNil)

		rights, err := is.applicationService.ListApplicationRights(ctx, &ttnpb.ApplicationIdentifiers{ApplicationID: app.ApplicationID})
		a.So(err, should.BeNil)
		a.So(rights.Rights, should.HaveLength, 2)
		a.So(rights.Rights, should.NotContain, ttnpb.RIGHT_APPLICATION_SETTINGS_API_KEYS)

		// But they can revoke themselves the INFO right.
		collab.Rights = []ttnpb.Right{ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS}
		_, err = is.applicationService.SetApplicationCollaborator(ctx, collab)
		a.So(err, should.BeNil)

		rights, err = is.applicationService.ListApplicationRights(ctx, &ttnpb.ApplicationIdentifiers{ApplicationID: app.ApplicationID})
		a.So(err, should.BeNil)
		a.So(rights.Rights, should.HaveLength, 1)
		a.So(rights.Rights, should.NotContain, ttnpb.RIGHT_APPLICATION_INFO)

		collab.Rights = []ttnpb.Right{ttnpb.RIGHT_APPLICATION_INFO, ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS}
		_, err = is.applicationService.SetApplicationCollaborator(ctx, collab)
		a.So(err, should.BeNil)
	}

	// Trying to unset the main collaborator will result in an error as the application
	// will become unmanageable.
	_, err = is.applicationService.SetApplicationCollaborator(ctx, &ttnpb.ApplicationCollaborator{
		ApplicationIdentifiers:        app.ApplicationIdentifiers,
		OrganizationOrUserIdentifiers: ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_UserID{UserID: &user.UserIdentifiers}},
	})
	a.So(err, should.NotBeNil)
	a.So(err, should.DescribeError, ErrUnmanageableApplication)

	// But we can revoke a shared right between the two collaborators.
	_, err = is.applicationService.SetApplicationCollaborator(ctx, &ttnpb.ApplicationCollaborator{
		ApplicationIdentifiers:        app.ApplicationIdentifiers,
		OrganizationOrUserIdentifiers: ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_UserID{UserID: &user.UserIdentifiers}},
		Rights: ttnpb.DifferenceRights(ttnpb.AllApplicationRights(), []ttnpb.Right{ttnpb.RIGHT_APPLICATION_INFO}),
	})
	a.So(err, should.NotBeNil)

	collabs, err = is.applicationService.ListApplicationCollaborators(ctx, &ttnpb.ApplicationIdentifiers{ApplicationID: app.ApplicationID})
	a.So(err, should.BeNil)
	a.So(collabs.Collaborators, should.HaveLength, 2)

	// Unset the last added collaborator.
	collab.Rights = []ttnpb.Right{}
	_, err = is.applicationService.SetApplicationCollaborator(ctx, collab)
	a.So(err, should.BeNil)

	collabs, err = is.applicationService.ListApplicationCollaborators(ctx, &ttnpb.ApplicationIdentifiers{ApplicationID: app.ApplicationID})
	a.So(err, should.BeNil)
	a.So(collabs.Collaborators, should.HaveLength, 1)

	_, err = is.applicationService.DeleteApplication(ctx, &ttnpb.ApplicationIdentifiers{ApplicationID: app.ApplicationID})
	a.So(err, should.BeNil)
}
