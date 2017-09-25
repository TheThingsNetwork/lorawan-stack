// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
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
	s := testStore(t)

	applications := testApplications()

	for _, application := range applications {
		created, err := s.Applications.Register(application)
		a.So(err, should.BeNil)
		a.So(created, test.ShouldBeApplicationIgnoringAutoFields, application)
	}

	// Attempt to recreate them should throw an error
	for _, application := range applications {
		_, err := s.Applications.Register(application)
		a.So(err, should.NotBeNil)
		a.So(err.(errors.Error).Code(), should.Equal, 2)
		a.So(err.(errors.Error).Type(), should.Equal, errors.AlreadyExists)
	}
}

func TestApplicationAppEUIManagement(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	app := testApplications()["demo-app"]
	eui := types.AppEUI(ttn_types.EUI64([8]byte{2, 2, 2, 2, 2, 2, 2, 2}))

	// Fetch AppEUIs
	{
		euis, err := s.Applications.ListAppEUIs(app.ID)
		a.So(err, should.BeNil)
		if a.So(euis, should.HaveLength, 2) {
			a.So(euis, should.Resemble, app.EUIs)
		}
	}

	// Add AppEUI
	{
		err := s.Applications.AddAppEUI(app.ID, eui)
		a.So(err, should.BeNil)

		euis, err := s.Applications.ListAppEUIs(app.ID)
		a.So(err, should.BeNil)
		if a.So(euis, should.HaveLength, 3) {
			a.So(euis, should.Contain, eui)
		}
	}

	// Delete AppEUI
	{
		err := s.Applications.RemoveAppEUI(app.ID, eui)
		a.So(err, should.BeNil)

		euis, err := s.Applications.ListAppEUIs(app.ID)
		a.So(err, should.BeNil)
		if a.So(euis, should.HaveLength, 2) {
			a.So(euis, should.NotContain, eui)
		}
	}
}

func TestApplicationAPIKeyManagement(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	app := testApplications()["demo-app"]
	key := types.ApplicationAPIKey{
		Name: "foo-key",
		Key:  "1234567",
		Rights: []types.Right{
			types.Right("bar"),
			types.Right("foo"),
			types.Right("zzzz"),
		},
	}

	// Fetch APIKeys
	{
		keys, err := s.Applications.ListAPIKeys(app.ID)
		a.So(err, should.BeNil)
		if a.So(keys, should.HaveLength, 1) {
			a.So(keys, should.Resemble, app.APIKeys)
		}
	}

	// Add APIKey
	{
		err := s.Applications.AddAPIKey(app.ID, key)
		a.So(err, should.BeNil)

		keys, err := s.Applications.ListAPIKeys(app.ID)
		a.So(err, should.BeNil)
		if a.So(keys, should.HaveLength, 2) {
			a.So(keys, should.Contain, key)
		}
	}

	// Delete APIKey
	{
		err := s.Applications.RemoveAPIKey(app.ID, key.Name)
		a.So(err, should.BeNil)

		keys, err := s.Applications.ListAPIKeys(app.ID)
		a.So(err, should.BeNil)
		if a.So(keys, should.HaveLength, 1) {
			a.So(keys, should.NotContain, key)
		}
	}
}

func TestApplicationRetrieve(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	app := testApplications()["demo-app"]

	found, err := s.Applications.FindByID(app.ID)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeApplicationIgnoringAutoFields, app)
}

func TestApplicationCollaborators(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	user := testUsers()["alice"]
	app := testApplications()["demo-app"]

	// Add a new collaborator with Settings and Delete rights
	collaborator := utils.Collaborator(user.Username, []types.Right{types.ApplicationDeleteRight, types.ApplicationSettingsRight})
	err := s.Applications.AddCollaborator(app.ID, collaborator)
	a.So(err, should.BeNil)

	// Fetch application collaborators
	{
		collaborators, err := s.Applications.ListCollaborators(app.ID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 1)
		if len(collaborators) == 1 {
			a.So(collaborators[0], should.Resemble, collaborator)
		}
	}

	// Fetch applications where Alice is collaborator
	{
		apps, err := s.Applications.FindByUser(user.Username)
		a.So(err, should.BeNil)
		a.So(apps, should.HaveLength, 1)
		if len(apps) == 1 {
			a.So(apps[0], test.ShouldBeApplicationIgnoringAutoFields, app)
		}
	}

	// Revoke Settings right
	{
		err := s.Applications.RemoveRight(app.ID, user.Username, types.ApplicationSettingsRight)
		a.So(err, should.BeNil)
	}

	// Fetch user rights
	{
		rights, err := s.Applications.ListUserRights(app.ID, user.Username)
		a.So(err, should.BeNil)
		a.So(rights, should.HaveLength, 1)
		if len(rights) == 1 {
			a.So(rights[0], should.Equal, types.ApplicationDeleteRight)
		}
	}

	// Remove collaborator
	{
		err = s.Applications.RemoveCollaborator(app.ID, user.Username)
		a.So(err, should.BeNil)

		collaborators, err := s.Applications.ListCollaborators(app.ID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 0)
	}
}

func TestApplicationEdit(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	app := testApplications()["demo-app"]
	app.Description = "New description"

	updated, err := s.Applications.Edit(app)
	a.So(err, should.BeNil)
	a.So(updated, test.ShouldBeApplicationIgnoringAutoFields, app)
}

func TestApplicationArchive(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	app := testApplications()["demo-app"]
	app.ID = "archived-app"

	// Create new application
	{
		created, err := s.Applications.Register(app)
		a.So(err, should.BeNil)
		a.So(created, test.ShouldBeApplicationIgnoringAutoFields, app)
	}

	// Archive it
	{
		err := s.Applications.Archive(app.ID)
		a.So(err, should.BeNil)

		found, err := s.Applications.FindByID(app.ID)
		a.So(err, should.BeNil)
		a.So(found.GetApplication().Archived, should.NotBeNil)
	}
}
