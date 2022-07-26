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

	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	is "go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func (st *StoreTest) TestDeletedEntities(t *T) {
	_, ctx := test.New(t)

	app1 := st.population.NewApplication(nil)
	app2 := st.population.NewApplication(nil)
	st.population.NewApplication(nil)
	cli1 := st.population.NewClient(nil)
	cli2 := st.population.NewClient(nil)
	st.population.NewClient(nil)
	gtw1 := st.population.NewGateway(nil)
	gtw2 := st.population.NewGateway(nil)
	st.population.NewGateway(nil)
	org1 := st.population.NewOrganization(nil)
	org2 := st.population.NewOrganization(nil)
	st.population.NewOrganization(nil)
	usr1 := st.population.NewUser()
	usr2 := st.population.NewUser()
	st.population.NewUser()

	s, ok := st.PrepareDB(t).(interface {
		Store
		is.ApplicationStore
		is.ClientStore
		is.GatewayStore
		is.OrganizationStore
		is.UserStore
	})
	defer st.DestroyDB(t, false)
	if !ok {
		t.Skip("Store does not implement entity store interfaces")
	}
	defer s.Close()

	start := time.Now().UTC().Truncate(time.Millisecond)
	time.Sleep(10 * test.Delay)

	for _, err := range []error{
		s.DeleteApplication(ctx, app1.GetIds()),
		s.DeleteClient(ctx, cli1.GetIds()),
		s.DeleteGateway(ctx, gtw1.GetIds()),
		s.DeleteOrganization(ctx, org1.GetIds()),
		s.DeleteUser(ctx, usr1.GetIds()),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}

	time.Sleep(10 * test.Delay)
	mid := time.Now().UTC().Truncate(time.Millisecond)
	time.Sleep(10 * test.Delay)

	for _, err := range []error{
		s.DeleteApplication(ctx, app2.GetIds()),
		s.DeleteClient(ctx, cli2.GetIds()),
		s.DeleteGateway(ctx, gtw2.GetIds()),
		s.DeleteOrganization(ctx, org2.GetIds()),
		s.DeleteUser(ctx, usr2.GetIds()),
	} {
		if err != nil {
			t.Fatal(err)
		}
	}

	time.Sleep(10 * test.Delay)
	end := time.Now().UTC().Truncate(time.Millisecond)

	t.Run("WithDeleted", func(t *T) {
		a, ctx := test.New(t)
		gotApplications, err := s.FindApplications(store.WithSoftDeleted(ctx, false), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotApplications, should.NotBeNil) {
			a.So(gotApplications, should.HaveLength, 3)
		}
		gotClients, err := s.FindClients(store.WithSoftDeleted(ctx, false), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotClients, should.NotBeNil) {
			a.So(gotClients, should.HaveLength, 3)
		}
		gotGateways, err := s.FindGateways(store.WithSoftDeleted(ctx, false), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotGateways, should.NotBeNil) {
			a.So(gotGateways, should.HaveLength, 3)
		}
		gotOrganizations, err := s.FindOrganizations(store.WithSoftDeleted(ctx, false), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotOrganizations, should.NotBeNil) {
			a.So(gotOrganizations, should.HaveLength, 3)
		}
		gotUsers, err := s.FindUsers(store.WithSoftDeleted(ctx, false), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotUsers, should.NotBeNil) {
			a.So(gotUsers, should.HaveLength, 3)
		}
	})

	t.Run("OnlyDeleted", func(t *T) {
		a, ctx := test.New(t)
		gotApplications, err := s.FindApplications(store.WithSoftDeleted(ctx, true), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotApplications, should.NotBeNil) {
			a.So(gotApplications, should.HaveLength, 2)
		}
		gotClients, err := s.FindClients(store.WithSoftDeleted(ctx, true), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotClients, should.NotBeNil) {
			a.So(gotClients, should.HaveLength, 2)
		}
		gotGateways, err := s.FindGateways(store.WithSoftDeleted(ctx, true), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotGateways, should.NotBeNil) {
			a.So(gotGateways, should.HaveLength, 2)
		}
		gotOrganizations, err := s.FindOrganizations(store.WithSoftDeleted(ctx, true), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotOrganizations, should.NotBeNil) {
			a.So(gotOrganizations, should.HaveLength, 2)
		}
		gotUsers, err := s.FindUsers(store.WithSoftDeleted(ctx, true), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotUsers, should.NotBeNil) {
			a.So(gotUsers, should.HaveLength, 2)
		}
	})

	t.Run("DeletedBeforeStart", func(t *T) {
		a, ctx := test.New(t)
		gotApplications, err := s.FindApplications(store.WithSoftDeletedBetween(ctx, nil, &start), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotApplications, should.NotBeNil) {
			a.So(gotApplications, should.BeEmpty)
		}
		gotClients, err := s.FindClients(store.WithSoftDeletedBetween(ctx, nil, &start), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotClients, should.NotBeNil) {
			a.So(gotClients, should.BeEmpty)
		}
		gotGateways, err := s.FindGateways(store.WithSoftDeletedBetween(ctx, nil, &start), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotGateways, should.NotBeNil) {
			a.So(gotGateways, should.BeEmpty)
		}
		gotOrganizations, err := s.FindOrganizations(store.WithSoftDeletedBetween(ctx, nil, &start), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotOrganizations, should.NotBeNil) {
			a.So(gotOrganizations, should.BeEmpty)
		}
		gotUsers, err := s.FindUsers(store.WithSoftDeletedBetween(ctx, nil, &start), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotUsers, should.NotBeNil) {
			a.So(gotUsers, should.BeEmpty)
		}
	})

	t.Run("DeletedFirstBatch", func(t *T) {
		a, ctx := test.New(t)
		gotApplications, err := s.FindApplications(store.WithSoftDeletedBetween(ctx, &start, &mid), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotApplications, should.NotBeNil) && a.So(gotApplications, should.HaveLength, 1) {
			a.So(gotApplications[0].GetIds(), should.Resemble, app1.GetIds())
		}
		gotClients, err := s.FindClients(store.WithSoftDeletedBetween(ctx, &start, &mid), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotClients, should.NotBeNil) && a.So(gotClients, should.HaveLength, 1) {
			a.So(gotClients[0].GetIds(), should.Resemble, cli1.GetIds())
		}
		gotGateways, err := s.FindGateways(store.WithSoftDeletedBetween(ctx, &start, &mid), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotGateways, should.NotBeNil) && a.So(gotGateways, should.HaveLength, 1) {
			a.So(gotGateways[0].GetIds(), should.Resemble, gtw1.GetIds())
		}
		gotOrganizations, err := s.FindOrganizations(store.WithSoftDeletedBetween(ctx, &start, &mid), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotOrganizations, should.NotBeNil) && a.So(gotOrganizations, should.HaveLength, 1) {
			a.So(gotOrganizations[0].GetIds(), should.Resemble, org1.GetIds())
		}
		gotUsers, err := s.FindUsers(store.WithSoftDeletedBetween(ctx, &start, &mid), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotUsers, should.NotBeNil) && a.So(gotUsers, should.HaveLength, 1) {
			a.So(gotUsers[0].GetIds(), should.Resemble, usr1.GetIds())
		}

		gotApplications, err = s.FindApplications(store.WithSoftDeletedBetween(ctx, nil, &mid), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotApplications, should.NotBeNil) && a.So(gotApplications, should.HaveLength, 1) {
			a.So(gotApplications[0].GetIds(), should.Resemble, app1.GetIds())
		}
		gotClients, err = s.FindClients(store.WithSoftDeletedBetween(ctx, nil, &mid), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotClients, should.NotBeNil) && a.So(gotClients, should.HaveLength, 1) {
			a.So(gotClients[0].GetIds(), should.Resemble, cli1.GetIds())
		}
		gotGateways, err = s.FindGateways(store.WithSoftDeletedBetween(ctx, nil, &mid), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotGateways, should.NotBeNil) && a.So(gotGateways, should.HaveLength, 1) {
			a.So(gotGateways[0].GetIds(), should.Resemble, gtw1.GetIds())
		}
		gotOrganizations, err = s.FindOrganizations(store.WithSoftDeletedBetween(ctx, nil, &mid), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotOrganizations, should.NotBeNil) && a.So(gotOrganizations, should.HaveLength, 1) {
			a.So(gotOrganizations[0].GetIds(), should.Resemble, org1.GetIds())
		}
		gotUsers, err = s.FindUsers(store.WithSoftDeletedBetween(ctx, nil, &mid), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotUsers, should.NotBeNil) && a.So(gotUsers, should.HaveLength, 1) {
			a.So(gotUsers[0].GetIds(), should.Resemble, usr1.GetIds())
		}
	})

	t.Run("DeletedSecondBatch", func(t *T) {
		a, ctx := test.New(t)
		gotApplications, err := s.FindApplications(store.WithSoftDeletedBetween(ctx, &mid, &end), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotApplications, should.NotBeNil) && a.So(gotApplications, should.HaveLength, 1) {
			a.So(gotApplications[0].GetIds(), should.Resemble, app2.GetIds())
		}
		gotClients, err := s.FindClients(store.WithSoftDeletedBetween(ctx, &mid, &end), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotClients, should.NotBeNil) && a.So(gotClients, should.HaveLength, 1) {
			a.So(gotClients[0].GetIds(), should.Resemble, cli2.GetIds())
		}
		gotGateways, err := s.FindGateways(store.WithSoftDeletedBetween(ctx, &mid, &end), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotGateways, should.NotBeNil) && a.So(gotGateways, should.HaveLength, 1) {
			a.So(gotGateways[0].GetIds(), should.Resemble, gtw2.GetIds())
		}
		gotOrganizations, err := s.FindOrganizations(store.WithSoftDeletedBetween(ctx, &mid, &end), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotOrganizations, should.NotBeNil) && a.So(gotOrganizations, should.HaveLength, 1) {
			a.So(gotOrganizations[0].GetIds(), should.Resemble, org2.GetIds())
		}
		gotUsers, err := s.FindUsers(store.WithSoftDeletedBetween(ctx, &mid, &end), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotUsers, should.NotBeNil) && a.So(gotUsers, should.HaveLength, 1) {
			a.So(gotUsers[0].GetIds(), should.Resemble, usr2.GetIds())
		}

		gotApplications, err = s.FindApplications(store.WithSoftDeletedBetween(ctx, &mid, nil), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotApplications, should.NotBeNil) && a.So(gotApplications, should.HaveLength, 1) {
			a.So(gotApplications[0].GetIds(), should.Resemble, app2.GetIds())
		}
		gotClients, err = s.FindClients(store.WithSoftDeletedBetween(ctx, &mid, nil), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotClients, should.NotBeNil) && a.So(gotClients, should.HaveLength, 1) {
			a.So(gotClients[0].GetIds(), should.Resemble, cli2.GetIds())
		}
		gotGateways, err = s.FindGateways(store.WithSoftDeletedBetween(ctx, &mid, nil), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotGateways, should.NotBeNil) && a.So(gotGateways, should.HaveLength, 1) {
			a.So(gotGateways[0].GetIds(), should.Resemble, gtw2.GetIds())
		}
		gotOrganizations, err = s.FindOrganizations(store.WithSoftDeletedBetween(ctx, &mid, nil), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotOrganizations, should.NotBeNil) && a.So(gotOrganizations, should.HaveLength, 1) {
			a.So(gotOrganizations[0].GetIds(), should.Resemble, org2.GetIds())
		}
		gotUsers, err = s.FindUsers(store.WithSoftDeletedBetween(ctx, &mid, nil), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotUsers, should.NotBeNil) && a.So(gotUsers, should.HaveLength, 1) {
			a.So(gotUsers[0].GetIds(), should.Resemble, usr2.GetIds())
		}
	})

	t.Run("DeletedAfterEnd", func(t *T) {
		a, ctx := test.New(t)
		gotApplications, err := s.FindApplications(store.WithSoftDeletedBetween(ctx, &end, nil), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotApplications, should.NotBeNil) {
			a.So(gotApplications, should.BeEmpty)
		}
		gotClients, err := s.FindClients(store.WithSoftDeletedBetween(ctx, &end, nil), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotClients, should.NotBeNil) {
			a.So(gotClients, should.BeEmpty)
		}
		gotGateways, err := s.FindGateways(store.WithSoftDeletedBetween(ctx, &end, nil), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotGateways, should.NotBeNil) {
			a.So(gotGateways, should.BeEmpty)
		}
		gotOrganizations, err := s.FindOrganizations(store.WithSoftDeletedBetween(ctx, &end, nil), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotOrganizations, should.NotBeNil) {
			a.So(gotOrganizations, should.BeEmpty)
		}
		gotUsers, err := s.FindUsers(store.WithSoftDeletedBetween(ctx, &end, nil), nil, fieldMask("ids"))
		if a.So(err, should.BeNil) && a.So(gotUsers, should.NotBeNil) {
			a.So(gotUsers, should.BeEmpty)
		}
	})
}
