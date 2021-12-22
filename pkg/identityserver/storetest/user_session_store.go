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
	. "testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	is "go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func (st *StoreTest) TestUserSessionStore(t *T) {
	usr1 := st.population.NewUser()

	s, ok := st.PrepareDB(t).(interface {
		Store
		is.UserSessionStore
	})
	defer st.DestroyDB(t, true, "users", "accounts")
	defer s.Close()
	if !ok {
		t.Fatal("Store does not implement UserSessionStore")
	}

	var created *ttnpb.UserSession

	t.Run("CreateSession", func(t *T) {
		a, ctx := test.New(t)
		var err error
		start := time.Now().Truncate(time.Second)

		created, err = s.CreateSession(ctx, &ttnpb.UserSession{
			UserIds:       usr1.GetIds(),
			SessionSecret: "secret",
			ExpiresAt:     ttnpb.ProtoTimePtr(start.Add(5 * time.Minute)),
		})
		if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) {
			a.So(created.UserIds, should.Resemble, usr1.GetIds())
			a.So(created.SessionId, should.NotBeBlank)
			a.So(created.SessionSecret, should.Equal, "secret")
			a.So(*ttnpb.StdTime(created.ExpiresAt), should.Equal, start.Add(5*time.Minute))
			a.So(*ttnpb.StdTime(created.CreatedAt), should.HappenWithin, 5*time.Second, start)
			a.So(*ttnpb.StdTime(created.UpdatedAt), should.HappenWithin, 5*time.Second, start)
		}
	})

	t.Run("GetSession", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.GetSession(ctx, usr1.GetIds(), created.SessionId)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.Resemble, created)
		}
	})

	t.Run("GetSession_Other", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.GetSession(ctx, usr1.GetIds(), "857c66da-304a-4378-b71f-03f2e94ff947")
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// _, err = s.GetSession(ctx, usr1.GetIds(), "")
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("GetSessionByID", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.GetSessionByID(ctx, created.SessionId)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.Resemble, created)
		}
	})

	t.Run("GetSessionByID_Other", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.GetSessionByID(ctx, "857c66da-304a-4378-b71f-03f2e94ff947")
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// _, err = s.GetSessionByID(ctx, "")
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("FindSessions", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.FindSessions(ctx, usr1.GetIds())
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 1) {
			a.So(got[0], should.Resemble, created)
		}
	})

	t.Run("DeleteSession", func(t *T) {
		a, ctx := test.New(t)
		err := s.DeleteSession(ctx, usr1.GetIds(), created.SessionId)
		a.So(err, should.BeNil)
	})

	t.Run("DeleteSession_Other", func(t *T) {
		a, ctx := test.New(t)
		err := s.DeleteSession(ctx, usr1.GetIds(), "857c66da-304a-4378-b71f-03f2e94ff947")
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// err = s.DeleteSession(ctx, usr1.GetIds(), "")
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("GetSession_AfterDelete", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.GetSession(ctx, usr1.GetIds(), created.SessionId)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
	})

	t.Run("GetSessionByID_AfterDelete", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.GetSessionByID(ctx, created.SessionId)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
	})

	t.Run("FindSessions_AfterDelete", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.FindSessions(ctx, usr1.GetIds())
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.BeEmpty)
		}
	})

	t.Run("DeleteAllUserSessions", func(t *T) {
		a, ctx := test.New(t)
		created, err := s.CreateSession(ctx, &ttnpb.UserSession{
			UserIds:       usr1.GetIds(),
			SessionSecret: "secret",
		})
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		err = s.DeleteAllUserSessions(ctx, usr1.GetIds())
		a.So(err, should.BeNil)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		_, err = s.GetSessionByID(ctx, created.SessionId)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
	})
}
