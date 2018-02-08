// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var applicationSpecializer = func(base ttnpb.Application) store.Application {
	return &base
}

func testApplications() map[string]*ttnpb.Application {
	return map[string]*ttnpb.Application{
		"demo-app": {
			ApplicationIdentifier: ttnpb.ApplicationIdentifier{ApplicationID: "demo-app"},
			Description:           "Demo application",
		},
	}
}

func TestApplicationCreate(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	applications := testApplications()

	for _, application := range applications {
		err := s.Applications.Create(application)
		a.So(err, should.BeNil)
	}

	// Attempt to recreate them should throw an error
	for _, application := range applications {
		err := s.Applications.Create(application)
		a.So(err, should.NotBeNil)
		a.So(ErrApplicationIDTaken.Describes(err), should.BeTrue)
	}
}

func TestApplicationAPIKeys(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	appID := testApplications()["demo-app"].ApplicationID
	key := &ttnpb.APIKey{
		Key:    "abcabcabc",
		Name:   "foo",
		Rights: []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)},
	}

	list, err := s.Applications.ListAPIKeys(appID)
	a.So(err, should.BeNil)
	a.So(list, should.HaveLength, 0)

	err = s.Applications.SaveAPIKey(appID, key)
	a.So(err, should.BeNil)

	key2 := &ttnpb.APIKey{
		Key:    "123abc",
		Name:   "foo",
		Rights: []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)},
	}
	err = s.Applications.SaveAPIKey(appID, key2)
	a.So(err, should.NotBeNil)
	a.So(ErrAPIKeyNameConflict.Describes(err), should.BeTrue)

	found, err := s.Applications.GetAPIKeyByName(appID, key.Name)
	a.So(err, should.BeNil)
	a.So(found, should.Resemble, key)

	key.Rights = append(key.Rights, ttnpb.Right(5))
	err = s.Applications.UpdateAPIKeyRights(appID, key.Name, key.Rights)
	a.So(err, should.BeNil)

	list, err = s.Applications.ListAPIKeys(appID)
	a.So(err, should.BeNil)
	if a.So(list, should.HaveLength, 1) {
		a.So(list[0], should.Resemble, key)
	}

	err = s.Applications.DeleteAPIKey(appID, key.Name)
	a.So(err, should.BeNil)

	found, err = s.Applications.GetAPIKeyByName(appID, key.Name)
	a.So(err, should.NotBeNil)
	a.So(ErrAPIKeyNotFound.Describes(err), should.BeTrue)
	a.So(found, should.BeNil)
}

func TestApplicationRetrieve(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	app := testApplications()["demo-app"]

	found, err := s.Applications.GetByID(app.ApplicationID, applicationSpecializer)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeApplicationIgnoringAutoFields, app)
}

func TestApplicationCollaborators(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	user := testUsers()["alice"]
	app := testApplications()["demo-app"]

	// check indeed that application has no collaborator
	{
		collaborators, err := s.Applications.ListCollaborators(app.ApplicationID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 0)
	}

	collaborator := &ttnpb.ApplicationCollaborator{
		ApplicationIdentifier:        ttnpb.ApplicationIdentifier{ApplicationID: app.ApplicationID},
		OrganizationOrUserIdentifier: ttnpb.OrganizationOrUserIdentifier{ID: &ttnpb.OrganizationOrUserIdentifier_UserID{user.UserID}},
		Rights: []ttnpb.Right{
			ttnpb.Right(1),
			ttnpb.Right(2),
		},
	}

	// add one
	{
		err := s.Applications.SetCollaborator(collaborator)
		a.So(err, should.BeNil)
	}

	// test HasCollaboratorRights method
	{
		yes, err := s.Applications.HasCollaboratorRights(app.ApplicationID, user.UserID, ttnpb.Right(0))
		a.So(yes, should.BeFalse)
		a.So(err, should.BeNil)

		yes, err = s.Applications.HasCollaboratorRights(app.ApplicationID, user.UserID, collaborator.Rights...)
		a.So(yes, should.BeTrue)
		a.So(err, should.BeNil)

		yes, err = s.Applications.HasCollaboratorRights(app.ApplicationID, user.UserID, append(collaborator.Rights, ttnpb.Right(0))...)
		a.So(yes, should.BeFalse)
		a.So(err, should.BeNil)
	}

	// check that it was added
	{
		collaborators, err := s.Applications.ListCollaborators(app.ApplicationID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 1)
		a.So(collaborators, should.Contain, collaborator)
	}

	// test ListCollaborators filter
	{
		collaborators, err := s.Applications.ListCollaborators(app.ApplicationID, ttnpb.Right(999))
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 0)

		collaborators, err = s.Applications.ListCollaborators(app.ApplicationID, ttnpb.Right(1))
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 1)
		a.So(collaborators, should.Contain, collaborator)

		collaborators, err = s.Applications.ListCollaborators(app.ApplicationID, ttnpb.Right(1), ttnpb.Right(3))
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 0)

	}

	// fetch applications where Alice is collaborator
	{
		apps, err := s.Applications.ListByOrganizationOrUser(user.UserID, applicationSpecializer)
		a.So(err, should.BeNil)
		if a.So(apps, should.HaveLength, 1) {
			a.So(apps[0].GetApplication().ApplicationID, should.Equal, app.ApplicationID)
		}
	}

	// modify rights
	{
		collaborator.Rights = append(collaborator.Rights, ttnpb.Right(3))
		err := s.Applications.SetCollaborator(collaborator)
		a.So(err, should.BeNil)

		collaborators, err := s.Applications.ListCollaborators(app.ApplicationID)
		a.So(err, should.BeNil)
		if a.So(collaborators, should.HaveLength, 1) {
			a.So(collaborators[0].Rights, should.Resemble, collaborator.Rights)
		}
	}

	// fetch user rights
	{
		rights, err := s.Applications.ListCollaboratorRights(app.ApplicationID, user.UserID)
		a.So(err, should.BeNil)
		if a.So(rights, should.HaveLength, 3) {
			a.So(rights, should.Resemble, collaborator.Rights)
		}
	}

	// remove collaborator
	{
		collaborator.Rights = []ttnpb.Right{}
		err := s.Applications.SetCollaborator(collaborator)
		a.So(err, should.BeNil)

		collaborators, err := s.Applications.ListCollaborators(app.ApplicationID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 0)
	}
}

func TestApplicationUpdate(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	app := testApplications()["demo-app"]
	app.Description = "New description"

	err := s.Applications.Update(app)
	a.So(err, should.BeNil)

	found, err := s.Applications.GetByID(app.ApplicationID, applicationSpecializer)
	a.So(err, should.BeNil)
	a.So(found.GetApplication().Description, should.Equal, app.Description)
}

func TestApplicationDelete(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	userID := testUsers()["bob"].UserID
	appID := "delete-test"

	testApplicationDeleteFeedDatabase(t, userID, appID)

	err := s.Applications.Delete(appID)
	a.So(err, should.BeNil)

	found, err := s.Applications.GetByID(appID, applicationSpecializer)
	a.So(err, should.NotBeNil)
	a.So(ErrApplicationNotFound.Describes(err), should.BeTrue)
	a.So(found, should.BeNil)
}

func testApplicationDeleteFeedDatabase(t *testing.T, userID, appID string) {
	a := assertions.New(t)
	s := testStore(t, database)

	app := &ttnpb.Application{
		ApplicationIdentifier: ttnpb.ApplicationIdentifier{ApplicationID: appID},
	}
	err := s.Applications.Create(app)
	a.So(err, should.BeNil)

	collaborator := &ttnpb.ApplicationCollaborator{
		ApplicationIdentifier:        app.ApplicationIdentifier,
		OrganizationOrUserIdentifier: ttnpb.OrganizationOrUserIdentifier{ID: &ttnpb.OrganizationOrUserIdentifier_UserID{userID}},
		Rights: []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)},
	}
	err = s.Applications.SetCollaborator(collaborator)
	a.So(err, should.BeNil)

	key := &ttnpb.APIKey{
		Name:   "foo",
		Key:    "123",
		Rights: []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)},
	}
	err = s.Applications.SaveAPIKey(appID, key)
	a.So(err, should.BeNil)
}
