// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package storetest

import (
	"fmt"
	. "testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	is "go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func (st *StoreTest) TestInvitationStore(t *T) {
	usr1 := st.population.NewUser()

	s, ok := st.PrepareDB(t).(interface {
		Store
		is.InvitationStore
	})
	defer st.DestroyDB(t, false)
	defer s.Close()
	if !ok {
		t.Skip("Store does not implement InvitationStore")
	}

	start := time.Now().Truncate(time.Second)

	var created *ttnpb.Invitation

	t.Run("CreateInvitation", func(t *T) {
		a, ctx := test.New(t)
		var err error
		start := time.Now().Truncate(time.Second)

		created, err = s.CreateInvitation(ctx, &ttnpb.Invitation{
			Email:     "foo@example.com",
			Token:     "TOKEN",
			ExpiresAt: ttnpb.ProtoTimePtr(start.Add(5 * time.Minute)),
		})

		if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) {
			a.So(created.Email, should.Equal, "foo@example.com")
			a.So(created.Token, should.Equal, "TOKEN")
			a.So(*ttnpb.StdTime(created.ExpiresAt), should.Equal, start.Add(5*time.Minute))
			a.So(*ttnpb.StdTime(created.CreatedAt), should.HappenWithin, 5*time.Second, start)
			a.So(*ttnpb.StdTime(created.UpdatedAt), should.HappenWithin, 5*time.Second, start)
		}
	})

	t.Run("CreateInvitation_AfterCreate", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.CreateInvitation(ctx, &ttnpb.Invitation{
			Email: "foo@example.com",
			Token: "TOKEN",
		})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsAlreadyExists(err), should.BeTrue)
		}
	})

	t.Run("GetInvitation", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.GetInvitation(ctx, "TOKEN")
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.Resemble, created)
		}
	})

	t.Run("GetInvitation_Other", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.GetInvitation(ctx, "OTHER")
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// _, err = s.GetInvitation(ctx, "")
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("FindInvitations", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.FindInvitations(ctx)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 1) {
			a.So(got[0], should.Resemble, created)
		}
	})

	t.Run("SetInvitationAcceptedBy", func(t *T) {
		a, ctx := test.New(t)
		err := s.SetInvitationAcceptedBy(ctx, "TOKEN", usr1.GetIds())
		a.So(err, should.BeNil)
	})

	t.Run("SetInvitationAcceptedBy_Again", func(t *T) {
		a, ctx := test.New(t)
		err := s.SetInvitationAcceptedBy(ctx, "TOKEN", usr1.GetIds())
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsFailedPrecondition(err), should.BeTrue)
		}
	})

	t.Run("SetInvitationAcceptedBy_Other", func(t *T) {
		a, ctx := test.New(t)
		err := s.SetInvitationAcceptedBy(ctx, "OTHER", usr1.GetIds())
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// err = s.SetInvitationAcceptedBy(ctx, "", usr1.GetIds())
		// a.So(err, should.BeNil)
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	var accepted *ttnpb.Invitation

	t.Run("GetInvitation_AfterUpdate", func(t *T) {
		a, ctx := test.New(t)
		var err error
		accepted, err = s.GetInvitation(ctx, "TOKEN")
		if a.So(err, should.BeNil) && a.So(accepted, should.NotBeNil) {
			if a.So(accepted.AcceptedAt, should.NotBeNil) {
				a.So(*ttnpb.StdTime(accepted.AcceptedAt), should.HappenWithin, 5*time.Second, start)
			}
			if a.So(accepted.AcceptedBy, should.NotBeNil) {
				a.So(accepted.AcceptedBy, should.Resemble, usr1.GetIds())
			}
		}
	})

	t.Run("FindInvitations_AfterUpdate", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.FindInvitations(ctx)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 1) {
			a.So(got[0], should.Resemble, accepted)
		}
	})

	t.Run("DeleteInvitation", func(t *T) {
		a, ctx := test.New(t)
		err := s.DeleteInvitation(ctx, "foo@example.com")
		a.So(err, should.BeNil)
	})

	t.Run("DeleteInvitation_Other", func(t *T) {
		a, ctx := test.New(t)
		err := s.DeleteInvitation(ctx, "bar@example.com")
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// err = s.DeleteInvitation(ctx, "")
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("GetInvitation_AfterDelete", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.GetInvitation(ctx, "TOKEN")
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
	})

	t.Run("FindInvitations_AfterDelete", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.FindInvitations(ctx)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.BeEmpty)
		}
	})

	t.Run("SetInvitationAcceptedBy_Expired", func(t *T) {
		a, ctx := test.New(t)

		_, err := s.CreateInvitation(ctx, &ttnpb.Invitation{
			Email:     "expired@example.com",
			Token:     "EXPIRED_TOKEN",
			ExpiresAt: ttnpb.ProtoTimePtr(start.Add(-1 * time.Minute)),
		})
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		err = s.SetInvitationAcceptedBy(ctx, "EXPIRED_TOKEN", usr1.GetIds())
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsFailedPrecondition(err), should.BeTrue)
		}

		err = s.DeleteInvitation(ctx, "expired@example.com")
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
	})
}

func (st *StoreTest) TestInvitationStorePagination(t *T) {
	a, ctx := test.New(t)
	start := time.Now().Truncate(time.Second)

	s, ok := st.PrepareDB(t).(interface {
		Store
		is.InvitationStore
	})
	defer st.DestroyDB(t, false)
	defer s.Close()
	if !ok {
		t.Skip("Store does not implement InvitationStore")
	}

	var all []*ttnpb.Invitation
	for i := 0; i < 7; i++ {
		created, err := s.CreateInvitation(ctx, &ttnpb.Invitation{
			Email:     fmt.Sprintf("user%d@example.com", i+1),
			Token:     fmt.Sprintf("TOKEN%d", i+1),
			ExpiresAt: ttnpb.ProtoTimePtr(start.Add(time.Minute)),
		})
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		all = append(all, created)
	}

	t.Run("FindInvitations_Paginated", func(t *T) {
		a, ctx := test.New(t)

		var total uint64
		for _, page := range []uint32{1, 2, 3, 4} {
			paginateCtx := store.WithPagination(store.WithOrder(ctx, "email"), 2, page, &total)

			got, err := s.FindInvitations(paginateCtx)
			if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
				if page == 4 {
					a.So(got, should.HaveLength, 1)
				} else {
					a.So(got, should.HaveLength, 2)
				}
				for i, e := range got {
					a.So(e, should.Resemble, all[i+2*int(page-1)])
				}
			}

			a.So(total, should.Equal, 7)
		}
	})
}
