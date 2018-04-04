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

package sql_test

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	. "github.com/TheThingsNetwork/ttn/pkg/identityserver/store/sql"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver/test"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var applicationSpecializer = func(base ttnpb.Application) store.Application {
	return &base
}

func TestApplications(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	application := &ttnpb.Application{
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "demo-app"},
		Description:            "Demo application",
	}

	err := s.Applications.Create(application)
	a.So(err, should.BeNil)

	err = s.Applications.Create(application)
	a.So(err, should.NotBeNil)
	a.So(ErrApplicationIDTaken.Describes(err), should.BeTrue)

	found, err := s.Applications.GetByID(application.ApplicationIdentifiers, applicationSpecializer)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeApplicationIgnoringAutoFields, application)

	application.Description = ""
	err = s.Applications.Update(application)
	a.So(err, should.BeNil)

	found, err = s.Applications.GetByID(application.ApplicationIdentifiers, applicationSpecializer)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeApplicationIgnoringAutoFields, application)

	collaborator := ttnpb.ApplicationCollaborator{
		ApplicationIdentifiers:        application.ApplicationIdentifiers,
		OrganizationOrUserIdentifiers: ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_UserID{UserID: &alice.UserIdentifiers}},
		Rights: []ttnpb.Right{ttnpb.RIGHT_APPLICATION_INFO},
	}

	collaborators, err := s.Applications.ListCollaborators(application.ApplicationIdentifiers)
	a.So(err, should.BeNil)
	a.So(collaborators, should.HaveLength, 0)

	err = s.Applications.SetCollaborator(collaborator)
	a.So(err, should.BeNil)

	collaborators, err = s.Applications.ListCollaborators(application.ApplicationIdentifiers)
	a.So(err, should.BeNil)
	a.So(collaborators, should.HaveLength, 1)
	a.So(collaborators, should.Contain, collaborator)

	collaborators, err = s.Applications.ListCollaborators(application.ApplicationIdentifiers, ttnpb.Right(0))
	a.So(err, should.BeNil)
	a.So(collaborators, should.HaveLength, 0)

	collaborators, err = s.Applications.ListCollaborators(application.ApplicationIdentifiers, collaborator.Rights...)
	a.So(err, should.BeNil)
	a.So(collaborators, should.HaveLength, 1)
	a.So(collaborators, should.Contain, collaborator)

	rights, err := s.Applications.ListCollaboratorRights(application.ApplicationIdentifiers, collaborator.OrganizationOrUserIdentifiers)
	a.So(err, should.BeNil)
	a.So(rights, should.Resemble, collaborator.Rights)

	has, err := s.Applications.HasCollaboratorRights(application.ApplicationIdentifiers, collaborator.OrganizationOrUserIdentifiers, ttnpb.Right(0))
	a.So(err, should.BeNil)
	a.So(has, should.BeFalse)

	has, err = s.Applications.HasCollaboratorRights(application.ApplicationIdentifiers, collaborator.OrganizationOrUserIdentifiers, collaborator.Rights...)
	a.So(err, should.BeNil)
	a.So(has, should.BeTrue)

	has, err = s.Applications.HasCollaboratorRights(application.ApplicationIdentifiers, collaborator.OrganizationOrUserIdentifiers, ttnpb.RIGHT_APPLICATION_INFO, ttnpb.Right(0))
	a.So(err, should.BeNil)
	a.So(has, should.BeFalse)

	collaborator.Rights = []ttnpb.Right{ttnpb.RIGHT_APPLICATION_TRAFFIC_READ}

	err = s.Applications.SetCollaborator(collaborator)
	a.So(err, should.BeNil)

	collaborators, err = s.Applications.ListCollaborators(application.ApplicationIdentifiers)
	a.So(err, should.BeNil)
	a.So(collaborators, should.HaveLength, 1)
	a.So(collaborators, should.Contain, collaborator)

	// Add a second collaborator.
	collaborator2 := ttnpb.ApplicationCollaborator{
		ApplicationIdentifiers:        application.ApplicationIdentifiers,
		OrganizationOrUserIdentifiers: ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_UserID{UserID: &bob.UserIdentifiers}},
		Rights: []ttnpb.Right{ttnpb.RIGHT_APPLICATION_INFO},
	}
	err = s.Applications.SetCollaborator(collaborator2)
	a.So(err, should.BeNil)

	collaborators, err = s.Applications.ListCollaborators(application.ApplicationIdentifiers)
	a.So(err, should.BeNil)
	a.So(collaborators, should.HaveLength, 2)
	a.So(collaborators, should.Contain, collaborator)
	a.So(collaborators, should.Contain, collaborator2)

	// Unset the collaborator.
	collaborator2.Rights = []ttnpb.Right{}
	err = s.Applications.SetCollaborator(collaborator2)
	a.So(err, should.BeNil)

	collaborators, err = s.Applications.ListCollaborators(application.ApplicationIdentifiers)
	a.So(err, should.BeNil)
	a.So(collaborators, should.HaveLength, 1)
	a.So(collaborators, should.Contain, collaborator)

	key := ttnpb.APIKey{
		Name:   "foo",
		Key:    "bar",
		Rights: []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)},
	}

	keys, err := s.Applications.ListAPIKeys(application.ApplicationIdentifiers)
	a.So(err, should.BeNil)
	a.So(keys, should.HaveLength, 0)

	err = s.Applications.SaveAPIKey(application.ApplicationIdentifiers, key)
	a.So(err, should.BeNil)

	keys, err = s.Applications.ListAPIKeys(application.ApplicationIdentifiers)
	a.So(err, should.BeNil)
	if a.So(keys, should.HaveLength, 1) {
		a.So(keys, should.Contain, key)
	}

	key.Rights = []ttnpb.Right{ttnpb.Right(1)}
	err = s.Applications.UpdateAPIKeyRights(application.ApplicationIdentifiers, key.Name, key.Rights)
	a.So(err, should.BeNil)

	ids, foundKey, err := s.Applications.GetAPIKey(key.Key)
	a.So(err, should.BeNil)
	a.So(ids, should.Resemble, application.ApplicationIdentifiers)
	a.So(foundKey, should.Resemble, key)

	foundKey, err = s.Applications.GetAPIKeyByName(application.ApplicationIdentifiers, key.Name)
	a.So(err, should.BeNil)
	a.So(foundKey, should.Resemble, key)

	err = s.Applications.DeleteAPIKey(application.ApplicationIdentifiers, key.Name)
	a.So(err, should.BeNil)

	keys, err = s.Applications.ListAPIKeys(application.ApplicationIdentifiers)
	a.So(err, should.BeNil)
	a.So(keys, should.HaveLength, 0)

	// save it again. Call to `Delete` will handle it
	err = s.Applications.SaveAPIKey(application.ApplicationIdentifiers, key)
	a.So(err, should.BeNil)

	err = s.Applications.Delete(application.ApplicationIdentifiers)
	a.So(err, should.BeNil)

	_, err = s.Applications.GetByID(application.ApplicationIdentifiers, applicationSpecializer)
	a.So(err, should.NotBeNil)
	a.So(ErrApplicationNotFound.Describes(err), should.BeTrue)
}
