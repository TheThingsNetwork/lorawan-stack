// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestCollaboratorStore(t *testing.T) {
	a := assertions.New(t)
	s := cleanStore(t, "collaborator_tests")

	user := testUsers()["alice"]
	app := testApplications()["demo-app"]

	// create application
	err := s.Applications.Create(app)
	a.So(err, should.BeNil)

	c := newCollaboratorStore(
		s,
		applicationsCollaboratorsTable,
		applicationsCollaboratorsForeignKey,
	)

	// check there is no collaborators
	{
		collaborators, err := c.ListCollaborators(app.ApplicationID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 0)
	}

	collaborator := ttnpb.Collaborator{
		UserIdentifier: ttnpb.UserIdentifier{user.UserID},
		Rights: []ttnpb.Right{
			ttnpb.Right(1),
			ttnpb.Right(2),
		},
	}

	// add collaborator
	{
		err := c.SetCollaborator(app.ApplicationID, collaborator)
		a.So(err, should.BeNil)

		collaborators, err := c.ListCollaborators(app.ApplicationID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 1)
		a.So(collaborators, should.Contain, collaborator)
	}

	// modify collaborator
	{
		collaborator.Rights = []ttnpb.Right{ttnpb.Right(5)}

		err := c.SetCollaborator(app.ApplicationID, collaborator)
		a.So(err, should.BeNil)

		collaborators, err := c.ListCollaborators(app.ApplicationID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 1)
		a.So(collaborators, should.Contain, collaborator)

		rights, err := c.ListUserRights(app.ApplicationID, user.UserID)
		a.So(err, should.BeNil)
		a.So(rights, should.Resemble, collaborator.Rights)
	}

	// remove collaborator
	{
		collaborator.Rights = []ttnpb.Right{}

		err := c.SetCollaborator(app.ApplicationID, collaborator)
		a.So(err, should.BeNil)

		collaborators, err := c.ListCollaborators(app.ApplicationID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 0)
	}
}
