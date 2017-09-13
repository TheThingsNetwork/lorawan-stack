// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/utils"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func testComponents() map[string]*types.DefaultComponent {
	return map[string]*types.DefaultComponent{
		"alice-handler": &types.DefaultComponent{
			ID:   "alice-handler",
			Type: types.Handler,
		},
		"foo-handler": &types.DefaultComponent{
			ID:   "foo-handler",
			Type: types.Handler,
		},
	}
}

func TestComponentCreate(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	components := testComponents()

	for _, component := range components {
		created, err := s.Components.Register(component)
		a.So(err, should.BeNil)
		a.So(created, test.ShouldBeComponentIgnoringAutoFields, component)
	}

	// Attempt to recreate them should throw an error
	for _, component := range components {
		_, err := s.Components.Register(component)
		a.So(err, should.NotBeNil)
		a.So(err.(errors.Error).Code(), should.Equal, 201)
		a.So(err.(errors.Error).Type(), should.Equal, errors.AlreadyExists)
	}
}

func TestComponentEdit(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	component := testComponents()["foo-handler"]
	component.Type = types.Broker

	updated, err := s.Components.Edit(component)
	a.So(err, should.BeNil)
	a.So(updated, test.ShouldBeComponentIgnoringAutoFields, component)
}

func TestComponentFind(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	component := testComponents()["alice-handler"]

	found, err := s.Components.FindByID(component.ID)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeComponentIgnoringAutoFields, component)
}

func TestComponentCollaborators(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	user := testUsers()["alice"]
	rights := []types.Right{types.ComponentDeleteRight}
	component := testComponents()["alice-handler"]

	collaborator := utils.Collaborator(user.Username, rights)

	// add collaborator
	{
		err := s.Components.AddCollaborator(component.ID, collaborator)
		a.So(err, should.BeNil)
	}

	// fetch component collaborators
	{
		collaborators, err := s.Components.ListCollaborators(component.ID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 1)
		if len(collaborators) > 0 {
			a.So(collaborators[0], should.Resemble, collaborator)
		}
	}

	// find which components alice is collaborator
	{
		components, err := s.Components.FindByUser(user.Username)
		a.So(err, should.BeNil)
		a.So(components, should.HaveLength, 1)
		if len(components) > 0 {
			a.So(components[0], test.ShouldBeComponentIgnoringAutoFields, component)
		}
	}

	// right to be granted and revoked
	right := types.ComponentCollaboratorsRight

	// grant a right
	{
		err := s.Components.AddRight(component.ID, user.Username, right)
		a.So(err, should.BeNil)

		rights, err := s.Components.ListUserRights(component.ID, user.Username)
		a.So(err, should.BeNil)
		a.So(rights, should.HaveLength, 2)
		if len(rights) > 0 {
			a.So(rights, should.Contain, right)
		}
	}

	// revoke a right
	{
		err := s.Components.RemoveRight(component.ID, user.Username, right)
		a.So(err, should.BeNil)

		rights, err := s.Components.ListUserRights(component.ID, user.Username)
		a.So(err, should.BeNil)
		a.So(rights, should.HaveLength, 1)
		if len(rights) > 0 {
			a.So(rights, should.NotContain, right)
		}
	}

	// fetch user rights
	{
		rights, err := s.Components.ListUserRights(component.ID, user.Username)
		a.So(err, should.BeNil)
		a.So(collaborator.Rights, should.Resemble, rights)
	}

	// delete collaborator
	{
		err := s.Components.RemoveCollaborator(component.ID, user.Username)
		a.So(err, should.BeNil)

		collaborators, err := s.Components.ListCollaborators(component.ID)
		a.So(err, should.BeNil)
		a.So(collaborators, should.HaveLength, 0)
	}
}

func TestComponentDelete(t *testing.T) {
	a := assertions.New(t)
	s := testStore()

	component := testComponents()["foo-handler"]
	err := s.Components.Delete(component.ID)
	a.So(err, should.BeNil)

	_, err = s.Components.FindByID(component.ID)
	a.So(err, should.NotBeNil)
	a.So(err.(errors.Error).Code(), should.Equal, 200)
	a.So(err.(errors.Error).Type(), should.Equal, errors.NotFound)
}
