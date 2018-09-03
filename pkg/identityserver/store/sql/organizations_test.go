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

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/identityserver"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestOrganizations(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	userID := alice.UserIdentifiers
	org := &ttnpb.Organization{
		OrganizationIdentifiers: ttnpb.OrganizationIdentifiers{OrganizationID: "thethingsnetwork"},
		Name:                    "The Things Network",
		Email:                   "foo@bar.org",
	}

	specializer := func(base ttnpb.Organization) store.Organization {
		return &base
	}

	err := s.Organizations.Create(org)
	a.So(err, should.BeNil)

	found, err := s.Organizations.GetByID(org.OrganizationIdentifiers, specializer)
	a.So(err, should.BeNil)
	a.So(found, should.EqualFieldsWithIgnores(identityserver.OrganizationGeneratedFields...), org)

	org.Description = "New description"
	err = s.Organizations.Update(org)
	a.So(err, should.BeNil)

	found, err = s.Organizations.GetByID(org.OrganizationIdentifiers, specializer)
	a.So(err, should.BeNil)
	a.So(found, should.EqualFieldsWithIgnores(identityserver.OrganizationGeneratedFields...), org)

	member := ttnpb.OrganizationMember{
		OrganizationIdentifiers: org.OrganizationIdentifiers,
		UserIdentifiers:         userID,
		Rights:                  []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)},
	}

	members, err := s.Organizations.ListMembers(org.OrganizationIdentifiers)
	a.So(err, should.BeNil)
	a.So(members, should.HaveLength, 0)

	err = s.Organizations.SetMember(member)
	a.So(err, should.BeNil)

	members, err = s.Organizations.ListMembers(org.OrganizationIdentifiers)
	a.So(err, should.BeNil)
	if a.So(members, should.HaveLength, 1) {
		a.So(members[0], should.Resemble, member)
	}

	member.Rights = append(member.Rights, ttnpb.Right(3))
	err = s.Organizations.SetMember(member)
	a.So(err, should.BeNil)

	members, err = s.Organizations.ListMembers(org.OrganizationIdentifiers)
	a.So(err, should.BeNil)
	if a.So(members, should.HaveLength, 1) {
		a.So(members[0], should.Resemble, member)
	}

	member.Rights = []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)}
	err = s.Organizations.SetMember(member)
	a.So(err, should.BeNil)

	members, err = s.Organizations.ListMembers(org.OrganizationIdentifiers)
	a.So(err, should.BeNil)
	if a.So(members, should.HaveLength, 1) {
		a.So(members[0], should.Resemble, member)
	}

	members, err = s.Organizations.ListMembers(org.OrganizationIdentifiers, ttnpb.Right(0))
	a.So(err, should.BeNil)
	a.So(members, should.HaveLength, 0)

	members, err = s.Organizations.ListMembers(org.OrganizationIdentifiers, member.Rights[0])
	a.So(err, should.BeNil)
	if a.So(members, should.HaveLength, 1) {
		a.So(members[0], should.Resemble, member)
	}

	members, err = s.Organizations.ListMembers(org.OrganizationIdentifiers, append(member.Rights, ttnpb.Right(0))...)
	a.So(err, should.BeNil)
	a.So(members, should.HaveLength, 0)

	yes, err := s.Organizations.HasMemberRights(org.OrganizationIdentifiers, userID, ttnpb.Right(1))
	a.So(err, should.BeNil)
	a.So(yes, should.BeTrue)

	yes, err = s.Organizations.HasMemberRights(org.OrganizationIdentifiers, userID, ttnpb.Right(1), ttnpb.Right(99))
	a.So(err, should.BeNil)
	a.So(yes, should.BeFalse)

	organizations, err := s.Organizations.ListByUser(userID, specializer)
	a.So(err, should.BeNil)
	if a.So(organizations, should.HaveLength, 1) {
		a.So(organizations[0], should.EqualFieldsWithIgnores(identityserver.OrganizationGeneratedFields...), org)
	}

	// Test applications rights inheritance.
	{
		application := &ttnpb.Application{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "org-app"},
		}
		err = s.Applications.Create(application)
		a.So(err, should.BeNil)

		collaborator := ttnpb.ApplicationCollaborator{
			ApplicationIdentifiers:        application.ApplicationIdentifiers,
			OrganizationOrUserIdentifiers: ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_OrganizationID{OrganizationID: &org.OrganizationIdentifiers}},
			Rights:                        ttnpb.AllApplicationRights(),
		}
		err = s.Applications.SetCollaborator(collaborator)
		a.So(err, should.BeNil)

		rights, err := s.Applications.ListCollaboratorRights(application.ApplicationIdentifiers, collaborator.OrganizationOrUserIdentifiers)
		a.So(err, should.BeNil)
		a.So(rights, should.Resemble, collaborator.Rights)

		userIDs := ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_UserID{UserID: &userID}}

		urights, err := s.Applications.ListCollaboratorRights(application.ApplicationIdentifiers, userIDs)
		a.So(err, should.BeNil)
		a.So(rights, should.Resemble, collaborator.Rights)
		a.So(rights, should.Resemble, urights)

		has, err := s.Applications.HasCollaboratorRights(application.ApplicationIdentifiers, userIDs, append(ttnpb.AllApplicationRights(), ttnpb.Right(0))...)
		a.So(err, should.BeNil)
		a.So(has, should.BeFalse)

		has, err = s.Applications.HasCollaboratorRights(application.ApplicationIdentifiers, userIDs, ttnpb.AllApplicationRights()...)
		a.So(err, should.BeNil)
		a.So(has, should.BeTrue)

		applications, err := s.Applications.ListByOrganizationOrUser(userIDs, applicationSpecializer)
		a.So(err, should.BeNil)
		if a.So(applications, should.HaveLength, 1) {
			a.So(applications[0], should.EqualFieldsWithIgnores(identityserver.ApplicationGeneratedFields...), application)
		}
	}

	// Test gateways rights inheritance.
	// TODO(gomezjdaniel): if Antennas, Attributes and Radios are removed tests
	// fails because <nil> =\ empty slice.
	{
		gateway := &ttnpb.Gateway{
			GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "org-gtw"},
			Antennas:           []ttnpb.GatewayAntenna{},
			Radios:             []ttnpb.GatewayRadio{},
			Attributes:         make(map[string]string),
		}
		err = s.Gateways.Create(gateway)
		a.So(err, should.BeNil)

		collaborator := ttnpb.GatewayCollaborator{
			GatewayIdentifiers:            gateway.GatewayIdentifiers,
			OrganizationOrUserIdentifiers: ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_OrganizationID{OrganizationID: &org.OrganizationIdentifiers}},
			Rights:                        ttnpb.AllGatewayRights(),
		}
		err = s.Gateways.SetCollaborator(collaborator)
		a.So(err, should.BeNil)

		rights, err := s.Gateways.ListCollaboratorRights(gateway.GatewayIdentifiers, collaborator.OrganizationOrUserIdentifiers)
		a.So(err, should.BeNil)
		a.So(rights, should.Resemble, collaborator.Rights)

		userIDs := ttnpb.OrganizationOrUserIdentifiers{ID: &ttnpb.OrganizationOrUserIdentifiers_UserID{UserID: &userID}}

		urights, err := s.Gateways.ListCollaboratorRights(gateway.GatewayIdentifiers, userIDs)
		a.So(err, should.BeNil)
		a.So(rights, should.Resemble, collaborator.Rights)
		a.So(rights, should.Resemble, urights)

		has, err := s.Gateways.HasCollaboratorRights(gateway.GatewayIdentifiers, userIDs, append(ttnpb.AllGatewayRights(), ttnpb.Right(0))...)
		a.So(err, should.BeNil)
		a.So(has, should.BeFalse)

		has, err = s.Gateways.HasCollaboratorRights(gateway.GatewayIdentifiers, userIDs, ttnpb.AllGatewayRights()...)
		a.So(err, should.BeNil)
		a.So(has, should.BeTrue)

		gateways, err := s.Gateways.ListByOrganizationOrUser(userIDs, gatewaySpecializer)
		a.So(err, should.BeNil)
		if a.So(gateways, should.HaveLength, 1) {
			a.So(gateways[0], should.EqualFieldsWithIgnores(identityserver.GatewayGeneratedFields...), gateway)
		}
	}

	member.Rights = []ttnpb.Right{}
	err = s.Organizations.SetMember(member)
	a.So(err, should.BeNil)

	members, err = s.Organizations.ListMembers(org.OrganizationIdentifiers)
	a.So(err, should.BeNil)
	a.So(members, should.HaveLength, 0)

	key := ttnpb.APIKey{
		Name:   "foo",
		Key:    "bar",
		Rights: []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)},
	}

	keys, err := s.Organizations.ListAPIKeys(org.OrganizationIdentifiers)
	a.So(err, should.BeNil)
	a.So(keys, should.HaveLength, 0)

	err = s.Organizations.SaveAPIKey(org.OrganizationIdentifiers, key)
	a.So(err, should.BeNil)

	keys, err = s.Organizations.ListAPIKeys(org.OrganizationIdentifiers)
	a.So(err, should.BeNil)
	if a.So(keys, should.HaveLength, 1) {
		a.So(keys, should.Contain, key)
	}

	key.Rights = []ttnpb.Right{ttnpb.Right(1)}
	err = s.Organizations.UpdateAPIKeyRights(org.OrganizationIdentifiers, key.Name, key.Rights)
	a.So(err, should.BeNil)

	ids, foundKey, err := s.Organizations.GetAPIKey(key.Key)
	a.So(err, should.BeNil)
	a.So(ids, should.Resemble, org.OrganizationIdentifiers)
	a.So(foundKey, should.Resemble, key)

	foundKey, err = s.Organizations.GetAPIKeyByName(org.OrganizationIdentifiers, key.Name)
	a.So(err, should.BeNil)
	a.So(foundKey, should.Resemble, key)

	err = s.Organizations.DeleteAPIKey(org.OrganizationIdentifiers, key.Name)
	a.So(err, should.BeNil)

	keys, err = s.Organizations.ListAPIKeys(org.OrganizationIdentifiers)
	a.So(err, should.BeNil)
	a.So(keys, should.HaveLength, 0)

	// Save it again. Call to `Delete` will handle it.
	err = s.Organizations.SaveAPIKey(org.OrganizationIdentifiers, key)
	a.So(err, should.BeNil)

	err = s.Organizations.Delete(org.OrganizationIdentifiers)
	a.So(err, should.BeNil)

	found, err = s.Organizations.GetByID(org.OrganizationIdentifiers, specializer)
	a.So(err, should.NotBeNil)
	a.So(err, should.DescribeError, store.ErrOrganizationNotFound)
	a.So(found, should.BeNil)
}
