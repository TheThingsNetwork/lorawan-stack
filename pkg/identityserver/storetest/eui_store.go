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
	"log"
	. "testing"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	is "go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func (st *StoreTest) TestEUIStore(t *T) {
	app1 := st.population.NewApplication(nil)
	app2 := st.population.NewApplication(nil)

	s, ok := st.PrepareDB(t).(interface {
		Store
		is.EUIStore
	})
	defer st.DestroyDB(t, false)
	defer s.Close()
	if !ok {
		t.Skip("Store does not implement EUIStore")
	}

	t.Run("InitializeDevEUIBlock", func(t *T) {
		a, ctx := test.New(t)
		err := s.CreateEUIBlock(ctx, types.EUI64Prefix{
			EUI64:  types.EUI64{1, 1, 1, 1, 1, 1, 0, 0},
			Length: 62,
		}, 1, "dev_eui")
		a.So(err, should.BeNil)
	})

	t.Run("IssueDevEUIForApplication", func(t *T) {
		a, ctx := test.New(t)
		eui, err := s.IssueDevEUIForApplication(ctx, app1.GetIds(), 1)
		if a.So(err, should.BeNil) && a.So(eui, should.NotBeNil) {
			a.So(*eui, should.Equal, types.EUI64{1, 1, 1, 1, 1, 1, 0, 1})
		}
	})

	t.Run("IssueDevEUIForApplication_Other", func(t *T) {
		a, ctx := test.New(t)
		eui, err := s.IssueDevEUIForApplication(ctx, app2.GetIds(), 1)
		if a.So(err, should.BeNil) && a.So(eui, should.NotBeNil) {
			a.So(*eui, should.Equal, types.EUI64{1, 1, 1, 1, 1, 1, 0, 2})
		}
	})

	t.Run("IssueDevEUIForApplication_Limit", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.IssueDevEUIForApplication(ctx, app1.GetIds(), 1)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsInvalidArgument(err), should.BeTrue)
		}
	})

	t.Run("IssueDevEUIForApplication_BlockLimit", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.IssueDevEUIForApplication(ctx, app1.GetIds(), 3)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		_, err = s.IssueDevEUIForApplication(ctx, app1.GetIds(), 3)
		if a.So(err, should.NotBeNil) {
			log.Println(err)
			a.So(errors.IsInvalidArgument(err), should.BeTrue)
		}
	})

	t.Run("InitializeDevEUIBlock_Extra", func(t *T) {
		a, ctx := test.New(t)
		err := s.CreateEUIBlock(ctx, types.EUI64Prefix{
			EUI64:  types.EUI64{1, 1, 1, 1, 0, 0, 0, 0},
			Length: 32,
		}, 0, "dev_eui")
		a.So(err, should.BeNil)
	})

	t.Run("IssueDevEUIForApplication_Other", func(t *T) {
		a, ctx := test.New(t)
		eui, err := s.IssueDevEUIForApplication(ctx, app1.GetIds(), 5)
		if a.So(err, should.BeNil) && a.So(eui, should.NotBeNil) {
			a.So(*eui, should.Equal, types.EUI64{1, 1, 1, 1, 0, 0, 0, 0})
		}
		a.So(err, should.BeNil)
	})
}
