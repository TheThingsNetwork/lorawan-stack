// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/utils"
	ttn_types "github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func testApplications() map[string]*types.DefaultApplication {
	return map[string]*types.DefaultApplication{
		"demo-app": &types.DefaultApplication{
			ID:          "demo-app",
			Description: "Demo application",
			EUIs: []types.AppEUI{
				types.AppEUI(ttn_types.EUI64([8]byte{1, 1, 1, 1, 1, 1, 1, 1})),
				types.AppEUI(ttn_types.EUI64([8]byte{1, 2, 3, 4, 5, 6, 7, 8})),
			},
			APIKeys: []types.ApplicationAPIKey{
				types.ApplicationAPIKey{
					Name: "test-key",
					Key:  "123",
					Rights: []types.Right{
						types.Right("bar"),
						types.Right("foo"),
					},
				},
			},
		},
	}
}

func TestApplicationCreate(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	applications := testApplications()

	for _, application := range applications {
		created, err := s.Applications.Create(application)
		a.So(err, should.BeNil)
		a.So(created, test.ShouldBeApplicationIgnoringAutoFields, application)
	}

	// attempt to recreate them should throw an error
	for _, application := range applications {
		_, err := s.Applications.Create(application)
		a.So(err.Error(), should.Equal, ErrApplicationIDTaken.Error())
	}
}

func TestApplicationRetrieve(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	app := testApplications()["demo-app"]

	found, err := s.Applications.FindByID(app.ID)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeApplicationIgnoringAutoFields, app)
}

func TestApplicationCollaborators(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	user := testUsers()["alice"]
	app := testApplications()["demo-app"]

	// add a new collaborator with Settings and Delete rights
	collaborator := utils.Collaborator(user.Username, []types.Right{types.ApplicationDeleteRight, types.ApplicationSettingsRight})
	err := s.Applications.AddCollaborator(app.ID, collaborator)
	a.So(err, should.BeNil)

	// fetch application collaborators
	{
		collaborators, err := s.Applications.Collaborators(app.ID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 1)
		if len(collaborators) == 1 {
			a.So(collaborators[0], should.Resemble, collaborator)
		}
	}

	// find applications an user is collaborator to
	{
		apps, err := s.Applications.FindByUser(user.Username)
		a.So(err, should.BeNil)
		a.So(apps, should.HaveLength, 1)
		if len(apps) == 1 {
			a.So(apps[0], test.ShouldBeApplicationIgnoringAutoFields, app)
		}
	}

	// revoke Settings right
	{
		err := s.Applications.RevokeRight(app.ID, user.Username, types.ApplicationSettingsRight)
		a.So(err, should.BeNil)
	}

	// fetch user rights
	{
		rights, err := s.Applications.UserRights(app.ID, user.Username)
		a.So(err, should.BeNil)
		a.So(rights, should.HaveLength, 1)
		if len(rights) == 1 {
			a.So(rights[0], should.Equal, types.ApplicationDeleteRight)
		}
	}

	// remove collaborator
	{
		err = s.Applications.RemoveCollaborator(app.ID, user.Username)
		a.So(err, should.BeNil)

		collaborators, err := s.Applications.Collaborators(app.ID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 0)
	}
}

func TestApplicationManagement(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	app := testApplications()["demo-app"]

	// update application
	{
		app.Description = "New description"
		result, err := s.Applications.Update(app)
		a.So(err, should.BeNil)
		a.So(result, test.ShouldBeApplicationIgnoringAutoFields, app)
	}

	// archive an application
	{
		app.ID = "archived-app"

		created, err := s.Applications.Create(app)
		a.So(err, should.BeNil)
		a.So(created, test.ShouldBeApplicationIgnoringAutoFields, app)

		// archive it
		err = s.Applications.Archive(app.ID)
		a.So(err, should.BeNil)
	}

	// attempt to archive a non existent application should return an error
	{
		err := s.Applications.Archive("non-existent")
		a.So(err, should.NotBeNil)
		a.So(err, should.Equal, ErrApplicationNotFound)
	}
}

func TestApplicationManagementAppEUI(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	app := testApplications()["demo-app"]
	eui := types.AppEUI(ttn_types.EUI64([8]byte{2, 2, 2, 2, 2, 2, 2, 2}))

	// add it
	{
		err := s.Applications.AddAppEUI(app.ID, eui)
		a.So(err, should.BeNil)
	}

	// check that indeed has been added
	{
		found, err := s.Applications.FindByID(app.ID)
		a.So(err, should.BeNil)
		a.So(found.GetApplication().EUIs, should.Contain, eui)
	}

	// delete it
	{
		err := s.Applications.DeleteAppEUI(app.ID, eui)
		a.So(err, should.BeNil)
	}

	// check that indeed has been deleted
	{
		found, err := s.Applications.FindByID(app.ID)
		a.So(err, should.BeNil)
		a.So(found.GetApplication().EUIs, should.NotContain, eui)
	}
}

func TestApplicationManagementApplicationAPIKey(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	app := testApplications()["demo-app"]
	key := app.APIKeys[0]
	key.Name = "foo-key"
	key.Key = "1234567"

	// add it
	{
		err := s.Applications.AddApplicationAPIKey(app.ID, key)
		a.So(err, should.BeNil)
	}

	// check that indeed has been added
	{
		found, err := s.Applications.FindByID(app.ID)
		a.So(err, should.BeNil)
		a.So(found.GetApplication().APIKeys, should.Contain, key)
	}

	// delete it
	{
		err := s.Applications.DeleteApplicationAPIKey(app.ID, key.Name)
		a.So(err, should.BeNil)
	}

	// check that indeed has been deleted
	{
		found, err := s.Applications.FindByID(app.ID)
		a.So(err, should.BeNil)
		a.So(found.GetApplication().APIKeys, should.NotContain, key)
	}
}
