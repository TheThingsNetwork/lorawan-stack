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
	"time"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/identityserver/store"
)

func TestInvitations(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	now := time.Now().UTC()

	invitation := store.InvitationData{
		Token:     "123",
		Email:     "foo@bar.com",
		IssuedAt:  now,
		ExpiresAt: now.Add(time.Duration(24) * time.Hour),
	}

	err := s.Invitations.Save(invitation)
	a.So(err, should.BeNil)

	found, err := s.Invitations.List()
	a.So(err, should.BeNil)
	if a.So(found, should.HaveLength, 1) {
		i := found[0]

		a.So(i.Email, should.Equal, invitation.Email)
		a.So(i.Token, should.Equal, invitation.Token)
		a.So(i.IssuedAt.IsZero(), should.BeFalse)
		a.So(i.ExpiresAt.IsZero(), should.BeFalse)
	}

	// Re-issue invitation.
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
		a.So(i.IssuedAt.IsZero(), should.BeFalse)
		a.So(i.ExpiresAt.IsZero(), should.BeFalse)
	}

	err = s.Invitations.Use(invitation.Email, invitation.Token)
	a.So(err, should.BeNil)

	err = s.Invitations.Use(invitation.Email, invitation.Token)
	a.So(err, should.NotBeNil)
	a.So(store.ErrInvitationNotFound.Describes(err), should.BeTrue)

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
