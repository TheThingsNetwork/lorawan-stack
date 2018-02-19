// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package identityserver

import (
	"sort"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/util"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var _ ttnpb.IsApplicationServer = new(applicationService)

func TestApplication(t *testing.T) {
	a := assertions.New(t)
	is := getIS(t)

	user := testUsers()["bob"]

	app := ttnpb.Application{
		ApplicationIdentifier: ttnpb.ApplicationIdentifier{"foo-app"},
	}

	ctx := testCtx(user.UserID)

	_, err := is.applicationService.CreateApplication(ctx, &ttnpb.CreateApplicationRequest{
		Application: app,
	})
	a.So(err, should.BeNil)

	// can't create applications with blacklisted ids
	for _, id := range testSettings().BlacklistedIDs {
		_, err := is.applicationService.CreateApplication(ctx, &ttnpb.CreateApplicationRequest{
			Application: ttnpb.Application{
				ApplicationIdentifier: ttnpb.ApplicationIdentifier{id},
			},
		})
		a.So(err, should.NotBeNil)
		a.So(ErrBlacklistedID.Describes(err), should.BeTrue)
	}

	found, err := is.applicationService.GetApplication(ctx, &ttnpb.ApplicationIdentifier{app.ApplicationID})
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeApplicationIgnoringAutoFields, app)

	apps, err := is.applicationService.ListApplications(ctx, &ttnpb.ListApplicationsRequest{})
	a.So(err, should.BeNil)
	if a.So(apps.Applications, should.HaveLength, 1) {
		a.So(apps.Applications[0], test.ShouldBeApplicationIgnoringAutoFields, app)
	}

	app.Description = "foo"
	_, err = is.applicationService.UpdateApplication(ctx, &ttnpb.UpdateApplicationRequest{
		Application: app,
		UpdateMask: pbtypes.FieldMask{
			Paths: []string{"description"},
		},
	})
	a.So(err, should.BeNil)

	found, err = is.applicationService.GetApplication(ctx, &ttnpb.ApplicationIdentifier{app.ApplicationID})
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeApplicationIgnoringAutoFields, app)

	// generate a new API key
	key, err := is.applicationService.GenerateApplicationAPIKey(ctx, &ttnpb.GenerateApplicationAPIKeyRequest{
		ApplicationIdentifier: app.ApplicationIdentifier,
		Name:   "foo",
		Rights: ttnpb.AllApplicationRights(),
	})
	a.So(err, should.BeNil)
	a.So(key.Key, should.NotBeEmpty)
	a.So(key.Name, should.Equal, key.Name)
	a.So(key.Rights, should.Resemble, ttnpb.AllApplicationRights())

	// update api key
	key.Rights = []ttnpb.Right{ttnpb.Right(10)}
	_, err = is.applicationService.UpdateApplicationAPIKey(ctx, &ttnpb.UpdateApplicationAPIKeyRequest{
		ApplicationIdentifier: app.ApplicationIdentifier,
		Name:   key.Name,
		Rights: key.Rights,
	})
	a.So(err, should.BeNil)

	// can't generate another API Key with the same name
	_, err = is.applicationService.GenerateApplicationAPIKey(ctx, &ttnpb.GenerateApplicationAPIKeyRequest{
		ApplicationIdentifier: app.ApplicationIdentifier,
		Name:   key.Name,
		Rights: []ttnpb.Right{ttnpb.Right(1)},
	})
	a.So(err, should.NotBeNil)
	a.So(sql.ErrAPIKeyNameConflict.Describes(err), should.BeTrue)

	keys, err := is.applicationService.ListApplicationAPIKeys(ctx, &ttnpb.ApplicationIdentifier{app.ApplicationID})
	a.So(err, should.BeNil)
	if a.So(keys.APIKeys, should.HaveLength, 1) {
		sort.Slice(keys.APIKeys[0].Rights, func(i, j int) bool { return keys.APIKeys[0].Rights[i] < keys.APIKeys[0].Rights[j] })
		a.So(keys.APIKeys[0], should.Resemble, key)
	}

	_, err = is.applicationService.RemoveApplicationAPIKey(ctx, &ttnpb.RemoveApplicationAPIKeyRequest{
		ApplicationIdentifier: app.ApplicationIdentifier,
		Name: key.Name,
	})

	keys, err = is.applicationService.ListApplicationAPIKeys(ctx, &ttnpb.ApplicationIdentifier{app.ApplicationID})
	a.So(err, should.BeNil)
	a.So(keys.APIKeys, should.HaveLength, 0)

	// set a new collaborator with SETTINGS_COLLABORATOR and INFO rights
	alice := testUsers()["alice"]
	collab := &ttnpb.ApplicationCollaborator{
		OrganizationOrUserIdentifier: ttnpb.OrganizationOrUserIdentifier{ID: &ttnpb.OrganizationOrUserIdentifier_UserID{alice.UserID}},
		ApplicationIdentifier:        app.ApplicationIdentifier,
		Rights:                       []ttnpb.Right{ttnpb.RIGHT_APPLICATION_INFO, ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS},
	}

	_, err = is.applicationService.SetApplicationCollaborator(ctx, collab)
	a.So(err, should.BeNil)

	rights, err := is.applicationService.ListApplicationRights(ctx, &ttnpb.ApplicationIdentifier{app.ApplicationID})
	a.So(err, should.BeNil)
	a.So(rights.Rights, should.Resemble, ttnpb.AllApplicationRights())

	collabs, err := is.applicationService.ListApplicationCollaborators(ctx, &ttnpb.ApplicationIdentifier{app.ApplicationID})
	a.So(err, should.BeNil)
	a.So(collabs.Collaborators, should.HaveLength, 2)
	a.So(collabs.Collaborators, should.Contain, collab)
	a.So(collabs.Collaborators, should.Contain, &ttnpb.ApplicationCollaborator{
		OrganizationOrUserIdentifier: ttnpb.OrganizationOrUserIdentifier{ID: &ttnpb.OrganizationOrUserIdentifier_UserID{user.UserID}},
		ApplicationIdentifier:        app.ApplicationIdentifier,
		Rights:                       ttnpb.AllApplicationRights(),
	})

	// the new collaborator can't grant himself more rights
	{
		collab.Rights = append(collab.Rights, ttnpb.RIGHT_APPLICATION_SETTINGS_KEYS)

		ctx := testCtx(alice.UserID)

		_, err = is.applicationService.SetApplicationCollaborator(ctx, collab)
		a.So(err, should.BeNil)

		rights, err := is.applicationService.ListApplicationRights(ctx, &ttnpb.ApplicationIdentifier{app.ApplicationID})
		a.So(err, should.BeNil)
		a.So(rights.Rights, should.HaveLength, 2)
		a.So(rights.Rights, should.NotContain, ttnpb.RIGHT_APPLICATION_SETTINGS_KEYS)

		// but it can't revoke itself the INFO right
		collab.Rights = []ttnpb.Right{ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS}
		_, err = is.applicationService.SetApplicationCollaborator(ctx, collab)
		a.So(err, should.BeNil)

		rights, err = is.applicationService.ListApplicationRights(ctx, &ttnpb.ApplicationIdentifier{app.ApplicationID})
		a.So(err, should.BeNil)
		a.So(rights.Rights, should.HaveLength, 1)
		a.So(rights.Rights, should.NotContain, ttnpb.RIGHT_APPLICATION_INFO)

		collab.Rights = []ttnpb.Right{ttnpb.RIGHT_APPLICATION_INFO, ttnpb.RIGHT_APPLICATION_SETTINGS_COLLABORATORS}
		_, err = is.applicationService.SetApplicationCollaborator(ctx, collab)
		a.So(err, should.BeNil)
	}

	// try to unset the main collaborator will result in error as the application
	// will become unmanageable
	_, err = is.applicationService.SetApplicationCollaborator(ctx, &ttnpb.ApplicationCollaborator{
		ApplicationIdentifier:        app.ApplicationIdentifier,
		OrganizationOrUserIdentifier: ttnpb.OrganizationOrUserIdentifier{ID: &ttnpb.OrganizationOrUserIdentifier_UserID{user.UserID}},
	})
	a.So(err, should.NotBeNil)
	a.So(ErrSetApplicationCollaboratorFailed.Describes(err), should.BeTrue)

	// but we can revoke a shared right between the two collaborators
	_, err = is.applicationService.SetApplicationCollaborator(ctx, &ttnpb.ApplicationCollaborator{
		ApplicationIdentifier:        app.ApplicationIdentifier,
		OrganizationOrUserIdentifier: ttnpb.OrganizationOrUserIdentifier{ID: &ttnpb.OrganizationOrUserIdentifier_UserID{user.UserID}},
		Rights: util.RightsDifference(ttnpb.AllApplicationRights(), []ttnpb.Right{ttnpb.RIGHT_APPLICATION_INFO}),
	})
	a.So(err, should.NotBeNil)

	collabs, err = is.applicationService.ListApplicationCollaborators(ctx, &ttnpb.ApplicationIdentifier{app.ApplicationID})
	a.So(err, should.BeNil)
	a.So(collabs.Collaborators, should.HaveLength, 2)

	// unset the last added collaborator
	collab.Rights = []ttnpb.Right{}
	_, err = is.applicationService.SetApplicationCollaborator(ctx, collab)
	a.So(err, should.BeNil)

	collabs, err = is.applicationService.ListApplicationCollaborators(ctx, &ttnpb.ApplicationIdentifier{app.ApplicationID})
	a.So(err, should.BeNil)
	a.So(collabs.Collaborators, should.HaveLength, 1)

	_, err = is.applicationService.DeleteApplication(ctx, &ttnpb.ApplicationIdentifier{app.ApplicationID})
	a.So(err, should.BeNil)
}
