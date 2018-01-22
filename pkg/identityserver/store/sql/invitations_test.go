// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestInvitations(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	invitation := &store.InvitationData{
		Token: "123",
		Email: "foo@bar.com",
		TTL:   uint32(3600),
	}

	userID := testUsers()["alice"].UserID

	err := s.Invitations.Save(invitation)
	a.So(err, should.BeNil)

	found, err := s.Invitations.List()
	a.So(err, should.BeNil)
	id := ""
	if a.So(found, should.HaveLength, 1) {
		i := found[0]
		id = i.ID

		a.So(i.ID, should.NotBeEmpty)
		a.So(i.Email, should.Equal, invitation.Email)
		if a.So(i.SentAt, should.NotBeNil) {
			a.So(i.SentAt.IsZero(), should.BeFalse)
		}
		a.So(i.UsedAt, should.BeNil)
		a.So(i.GetUserID(), should.BeEmpty)
	}

	err = s.Invitations.Use(invitation.Token, userID)
	a.So(err, should.BeNil)

	err = s.Invitations.Use(invitation.Token, userID)
	a.So(err, should.NotBeNil)
	a.So(ErrInvitationAlreadyUsed.Describes(err), should.BeTrue)

	found, err = s.Invitations.List()
	a.So(err, should.BeNil)
	if a.So(found, should.HaveLength, 1) {
		i := found[0]

		a.So(i.ID, should.NotBeEmpty)
		a.So(i.Email, should.Equal, invitation.Email)
		if a.So(i.SentAt, should.NotBeNil) {
			a.So(i.SentAt.IsZero(), should.BeFalse)
		}
		if a.So(i.UsedAt, should.NotBeNil) {
			a.So(i.UsedAt.IsZero(), should.BeFalse)
		}
		a.So(i.GetUserID(), should.Equal, userID)
	}

	err = s.Invitations.Delete(id)
	a.So(err, should.BeNil)

	err = s.Invitations.Use(invitation.Token, userID)
	a.So(err, should.NotBeNil)
	a.So(ErrInvitationNotFound.Describes(err), should.BeTrue)

	found, err = s.Invitations.List()
	a.So(err, should.BeNil)
	a.So(found, should.HaveLength, 0)
}
