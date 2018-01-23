// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/store"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestInvitations(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	invitation := &store.InvitationData{
		Token:     "123",
		Email:     "foo@bar.com",
		IssuedAt:  time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(time.Duration(24) * time.Hour),
	}

	err := s.Invitations.Save(invitation)
	a.So(err, should.BeNil)

	found, err := s.Invitations.List()
	a.So(err, should.BeNil)
	if a.So(found, should.HaveLength, 1) {
		i := found[0]

		a.So(i.Email, should.Equal, invitation.Email)
		a.So(i.Token, should.Equal, invitation.Token)
		a.So(invitation.IssuedAt.Equal(i.IssuedAt), should.BeTrue)
		a.So(invitation.ExpiresAt.Equal(i.ExpiresAt), should.BeTrue)
	}

	// reissue invitation
	invitation.Token = "123456"
	invitation.ExpiresAt = invitation.ExpiresAt.Add(time.Hour)
	err = s.Invitations.Save(invitation)
	a.So(err, should.BeNil)

	found, err = s.Invitations.List()
	a.So(err, should.BeNil)
	if a.So(found, should.HaveLength, 1) {
		i := found[0]

		a.So(i.Email, should.Equal, invitation.Email)
		a.So(i.Token, should.Equal, invitation.Token)
		a.So(invitation.IssuedAt.Equal(i.IssuedAt), should.BeTrue)
		a.So(invitation.ExpiresAt.Equal(i.ExpiresAt), should.BeTrue)
	}

	err = s.Invitations.Use(invitation.Email, invitation.Token)
	a.So(err, should.BeNil)

	err = s.Invitations.Use(invitation.Email, invitation.Token)
	a.So(err, should.NotBeNil)
	a.So(ErrInvitationNotFound.Describes(err), should.BeTrue)

	found, err = s.Invitations.List()
	a.So(err, should.BeNil)
	a.So(found, should.HaveLength, 0)

	err = s.Invitations.Save(invitation)
	a.So(err, should.BeNil)

	found, err = s.Invitations.List()
	a.So(err, should.BeNil)
	a.So(found, should.HaveLength, 1)

	err = s.Invitations.Delete(invitation.Email)
	a.So(err, should.BeNil)

	found, err = s.Invitations.List()
	a.So(err, should.BeNil)
	a.So(found, should.HaveLength, 0)
}
