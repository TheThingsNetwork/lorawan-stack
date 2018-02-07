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

func TestOrganizations(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	userID := testUsers()["alice"].UserID
	org := &ttnpb.Organization{
		OrganizationIdentifier: ttnpb.OrganizationIdentifier{"thethingsnetwork"},
		Name:  "The Things Network",
		Email: "foo@bar.org",
	}

	specializer := func(base ttnpb.Organization) store.Organization {
		return &base
	}

	err := s.Organizations.Create(org)
	a.So(err, should.BeNil)

	found, err := s.Organizations.GetByID(org.OrganizationID, specializer)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeOrganizationIgnoringAutoFields, org)

	org.Description = "New description"
	err = s.Organizations.Update(org)
	a.So(err, should.BeNil)

	found, err = s.Organizations.GetByID(org.OrganizationID, specializer)
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeOrganizationIgnoringAutoFields, org)

	member := &ttnpb.OrganizationMember{
		OrganizationIdentifier: org.OrganizationIdentifier,
		UserIdentifier:         ttnpb.UserIdentifier{userID},
		Rights:                 []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)},
	}

	members, err := s.Organizations.ListMembers(org.OrganizationID)
	a.So(err, should.BeNil)
	a.So(members, should.HaveLength, 0)

	err = s.Organizations.SetMember(member)
	a.So(err, should.BeNil)

	members, err = s.Organizations.ListMembers(org.OrganizationID)
	a.So(err, should.BeNil)
	if a.So(members, should.HaveLength, 1) {
		a.So(members[0], should.Resemble, member)
	}

	member.Rights = append(member.Rights, ttnpb.Right(3))
	err = s.Organizations.SetMember(member)
	a.So(err, should.BeNil)

	members, err = s.Organizations.ListMembers(org.OrganizationID)
	a.So(err, should.BeNil)
	if a.So(members, should.HaveLength, 1) {
		a.So(members[0], should.Resemble, member)
	}

	member.Rights = []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)}
	err = s.Organizations.SetMember(member)
	a.So(err, should.BeNil)

	members, err = s.Organizations.ListMembers(org.OrganizationID)
	a.So(err, should.BeNil)
	if a.So(members, should.HaveLength, 1) {
		a.So(members[0], should.Resemble, member)
	}

	yes, err := s.Organizations.HasMemberRights(org.OrganizationID, userID, ttnpb.Right(1))
	a.So(err, should.BeNil)
	a.So(yes, should.BeTrue)

	yes, err = s.Organizations.HasMemberRights(org.OrganizationID, userID, ttnpb.Right(1), ttnpb.Right(99))
	a.So(err, should.BeNil)
	a.So(yes, should.BeFalse)

	organizations, err := s.Organizations.ListByUser(userID, specializer)
	a.So(err, should.BeNil)
	if a.So(organizations, should.HaveLength, 1) {
		a.So(organizations[0], test.ShouldBeOrganizationIgnoringAutoFields, org)
	}

	member.Rights = []ttnpb.Right{}
	err = s.Organizations.SetMember(member)
	a.So(err, should.BeNil)

	members, err = s.Organizations.ListMembers(org.OrganizationID)
	a.So(err, should.BeNil)
	a.So(members, should.HaveLength, 0)

	key := &ttnpb.APIKey{
		Name:   "foo",
		Key:    "bar",
		Rights: []ttnpb.Right{ttnpb.Right(1), ttnpb.Right(2)},
	}

	keys, err := s.Organizations.ListAPIKeys(org.OrganizationID)
	a.So(err, should.BeNil)
	a.So(keys, should.HaveLength, 0)

	err = s.Organizations.SaveAPIKey(org.OrganizationID, key)
	a.So(err, should.BeNil)

	keys, err = s.Organizations.ListAPIKeys(org.OrganizationID)
	a.So(err, should.BeNil)
	if a.So(keys, should.HaveLength, 1) {
		a.So(keys, should.Contain, key)
	}

	foundKey, err := s.Organizations.GetAPIKeyByName(org.OrganizationID, key.Name)
	a.So(err, should.BeNil)
	a.So(foundKey, should.Resemble, key)

	err = s.Organizations.DeleteAPIKey(org.OrganizationID, key.Name)
	a.So(err, should.BeNil)

	keys, err = s.Organizations.ListAPIKeys(org.OrganizationID)
	a.So(err, should.BeNil)
	a.So(keys, should.HaveLength, 0)

	// save it again. Call to `Delete` will handle it
	err = s.Organizations.SaveAPIKey(org.OrganizationID, key)
	a.So(err, should.BeNil)

	err = s.Organizations.Delete(org.OrganizationID)
	a.So(err, should.BeNil)

	found, err = s.Organizations.GetByID(org.OrganizationID, specializer)
	a.So(err, should.NotBeNil)
	a.So(ErrOrganizationNotFound.Describes(err), should.BeTrue)
	a.So(found, should.BeNil)
}
