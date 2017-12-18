// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var applicationFactory = func() types.Application {
	return &ttnpb.Application{}

}

func testApplications() map[string]*ttnpb.Application {
	return map[string]*ttnpb.Application{
		"demo-app": {
			ApplicationIdentifier: ttnpb.ApplicationIdentifier{"demo-app"},
			Description:           "Demo application",
		},
	}
}

func TestApplicationCreate(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

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
	s := testStore(t)

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

	found, err := s.Applications.GetAPIKey(appID, key.Name)
	a.So(err, should.BeNil)
	a.So(found, should.Resemble, key)

	key.Rights = append(key.Rights, ttnpb.Right(5))
	err = s.Applications.UpdateAPIKey(appID, key)
	a.So(err, should.BeNil)

	list, err = s.Applications.ListAPIKeys(appID)
	a.So(err, should.BeNil)
	if a.So(list, should.HaveLength, 1) {
		a.So(list[0], should.Resemble, key)
	}

	err = s.Applications.DeleteAPIKey(appID, key.Name)
	a.So(err, should.BeNil)

	found, err = s.Applications.GetAPIKey(appID, key.Name)
	a.So(err, should.NotBeNil)
	a.So(ErrAPIKeyNotFound.Describes(err), should.BeTrue)
}

func TestApplicationRetrieve(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	app := testApplications()["demo-app"]

	found, err := s.Applications.GetByID(app.ApplicationID, applicationFactory)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeApplicationIgnoringAutoFields, app)
}

func TestApplicationCollaborators(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	user := testUsers()["alice"]
	app := testApplications()["demo-app"]

	// check indeed that application has no collaborator
	{
		collaborators, err := s.Applications.ListCollaborators(app.ApplicationID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 0)
	}

	collaborator := ttnpb.ApplicationCollaborator{
		ApplicationIdentifier: ttnpb.ApplicationIdentifier{app.ApplicationID},
		UserIdentifier:        ttnpb.UserIdentifier{user.UserID},
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

	// check that it was added
	{
		collaborators, err := s.Applications.ListCollaborators(app.ApplicationID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 1)
		a.So(collaborators, should.Contain, collaborator)
	}

	// fetch applications where Alice is collaborator
	{
		apps, err := s.Applications.ListByUser(user.UserID, applicationFactory)
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
		rights, err := s.Applications.ListUserRights(app.ApplicationID, user.UserID)
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
	s := testStore(t)

	app := testApplications()["demo-app"]
	app.Description = "New description"

	err := s.Applications.Update(app)
	a.So(err, should.BeNil)

	found, err := s.Applications.GetByID(app.ApplicationID, applicationFactory)
	a.So(err, should.BeNil)
	a.So(found.GetApplication().Description, should.Equal, app.Description)
}
