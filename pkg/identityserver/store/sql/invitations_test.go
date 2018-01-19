// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestInvitations(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t)

	token := "123"
	email := "foo@bar.com"
	ttl := time.Duration(time.Hour * 24)

	userID := testUsers()["alice"].UserID

	err := s.Invitations.Save(token, email, uint32(ttl.Seconds()))
	a.So(err, should.BeNil)

	found, err := s.Invitations.List()
	a.So(err, should.BeNil)
	id := ""
	if a.So(found, should.HaveLength, 1) {
		i := found[0]
		id = i.ID

		a.So(i.ID, should.NotBeEmpty)
		a.So(i.Email, should.Equal, email)
		if a.So(i.SentAt, should.NotBeNil) {
			a.So(i.SentAt.IsZero(), should.BeFalse)
		}
		a.So(i.UsedAt, should.BeNil)
		a.So(i.UserID, should.BeEmpty)
	}

	err = s.Invitations.Use(token, userID)
	a.So(err, should.BeNil)

	err = s.Invitations.Use(token, userID)
	a.So(err, should.NotBeNil)
	a.So(ErrInvitationAlreadyUsed.Describes(err), should.BeTrue)

	found, err = s.Invitations.List()
	a.So(err, should.BeNil)
	if a.So(found, should.HaveLength, 1) {
		i := found[0]

		a.So(i.ID, should.NotBeEmpty)
		a.So(i.Email, should.Equal, email)
		if a.So(i.SentAt, should.NotBeNil) {
			a.So(i.SentAt.IsZero(), should.BeFalse)
		}
		if a.So(i.UsedAt, should.NotBeNil) {
			a.So(i.UsedAt.IsZero(), should.BeFalse)
		}
		a.So(i.UserID, should.Equal, userID)
	}

	err = s.Invitations.Delete(id)
	a.So(err, should.BeNil)

	err = s.Invitations.Use(token, userID)
	a.So(err, should.NotBeNil)
	a.So(ErrInvitationNotFound.Describes(err), should.BeTrue)

	found, err = s.Invitations.List()
	a.So(err, should.BeNil)
	a.So(found, should.HaveLength, 0)
}
