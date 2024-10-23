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
	"sort"
	"strings"
	. "testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	is "go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type apiKeysByID []*ttnpb.APIKey

func (a apiKeysByID) Len() int           { return len(a) }
func (a apiKeysByID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a apiKeysByID) Less(i, j int) bool { return a[i].Id < a[j].Id }

func (st *StoreTest) TestAPIKeyStoreCRUD(t *T) {
	app1 := st.population.NewApplication(nil)
	cli1 := st.population.NewClient(nil)
	gtw1 := st.population.NewGateway(nil)
	org1 := st.population.NewOrganization(nil)
	usr1 := st.population.NewUser()

	s, ok := st.PrepareDB(t).(interface {
		Store

		is.APIKeyStore
	})
	defer st.DestroyDB(t, true, "applications", "clients", "gateways", "organizations", "users", "accounts")
	if !ok {
		t.Skip("Store does not implement APIKeyStore")
	}
	defer s.Close()

	someRights := map[string]*ttnpb.Rights{
		app1.EntityType(): ttnpb.RightsFrom(ttnpb.Right_RIGHT_APPLICATION_ALL),
		cli1.EntityType(): ttnpb.RightsFrom(ttnpb.Right_RIGHT_CLIENT_ALL),
		gtw1.EntityType(): ttnpb.RightsFrom(ttnpb.Right_RIGHT_GATEWAY_ALL),
		org1.EntityType(): ttnpb.RightsFrom(ttnpb.Right_RIGHT_ORGANIZATION_ALL),
		usr1.EntityType(): ttnpb.RightsFrom(ttnpb.Right_RIGHT_USER_ALL),
	}
	allRights := ttnpb.RightsFrom(ttnpb.Right_RIGHT_ALL)

	for _, ids := range []*ttnpb.EntityIdentifiers{
		app1.GetEntityIdentifiers(),
		cli1.GetEntityIdentifiers(),
		gtw1.GetEntityIdentifiers(),
		org1.GetEntityIdentifiers(),
	} {
		t.Run(ids.EntityType(), func(t *T) {
			id := fmt.Sprintf("%sAPIKEY", strings.ToUpper(ids.EntityType()))

			var created *ttnpb.APIKey

			t.Run("CreateAPIKey", func(t *T) {
				a, ctx := test.New(t)
				var err error
				start := time.Now().Truncate(time.Second)

				created, err = s.CreateAPIKey(ctx, ids, &ttnpb.APIKey{
					Id:     id,
					Key:    "Hash",
					Name:   "API Key Name",
					Rights: someRights[ids.EntityType()].GetRights(),
				})
				if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) {
					a.So(created.Id, should.Equal, id)
					a.So(created.Key, should.Equal, "Hash")
					a.So(created.Name, should.Equal, "API Key Name")
					a.So(created.Rights, should.Resemble, someRights[ids.EntityType()].GetRights())
					a.So(*ttnpb.StdTime(created.CreatedAt), should.HappenWithin, 5*time.Second, start)
					a.So(*ttnpb.StdTime(created.UpdatedAt), should.HappenWithin, 5*time.Second, start)
				}
			})

			t.Run("CreateAPIKey_AfterCreate", func(t *T) {
				a, ctx := test.New(t)
				_, err := s.CreateAPIKey(ctx, ids, &ttnpb.APIKey{
					Id: id,
				})
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsAlreadyExists(err), should.BeTrue)
				}
			})

			t.Run("GetAPIKey", func(t *T) {
				a, ctx := test.New(t)
				got, err := s.GetAPIKey(ctx, ids, id)
				if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
					a.So(got, should.Resemble, created)
				}
			})

			t.Run("GetAPIKey_Other", func(t *T) {
				a, ctx := test.New(t)
				_, err := s.GetAPIKey(ctx, ids, "OTHER")
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsNotFound(err), should.BeTrue)
				}
			})

			t.Run("GetAPIKeyByID", func(t *T) {
				a, ctx := test.New(t)
				gotIDs, got, err := s.GetAPIKeyByID(ctx, id)
				if a.So(err, should.BeNil) && a.So(gotIDs, should.NotBeNil) && a.So(got, should.NotBeNil) {
					a.So(gotIDs, should.Resemble, ids)
					a.So(got, should.Resemble, created)
				}
			})

			t.Run("GetAPIKeyByID_Other", func(t *T) {
				a, ctx := test.New(t)
				_, _, err := s.GetAPIKeyByID(ctx, "OTHER")
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsNotFound(err), should.BeTrue)
				}
			})

			t.Run("FindAPIKeys", func(t *T) {
				a, ctx := test.New(t)
				got, err := s.FindAPIKeys(ctx, ids)
				if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 1) {
					a.So(got[0], should.Resemble, created)
				}
			})

			var updated *ttnpb.APIKey

			t.Run("UpdateAPIKey", func(t *T) {
				a, ctx := test.New(t)
				var err error
				start := time.Now().Truncate(time.Second)

				updated, err = s.UpdateAPIKey(ctx, ids, &ttnpb.APIKey{
					Id:        id,
					Name:      "Updated Name",
					Rights:    allRights.GetRights(),
					ExpiresAt: timestamppb.New(start.Add(5 * time.Minute)),
				}, fieldMask("name", "rights", "expires_at"))
				if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) {
					a.So(updated.Name, should.Equal, "Updated Name")
					a.So(updated.Rights, should.Resemble, allRights.GetRights())
					a.So(*ttnpb.StdTime(updated.ExpiresAt), should.Equal, start.Add(5*time.Minute))
					a.So(*ttnpb.StdTime(updated.CreatedAt), should.Equal, *ttnpb.StdTime(created.CreatedAt))
					a.So(*ttnpb.StdTime(updated.UpdatedAt), should.HappenWithin, 5*time.Second, start)
				}
			})

			t.Run("UpdateAPIKey_Other", func(t *T) {
				a, ctx := test.New(t)
				_, err := s.UpdateAPIKey(ctx, ids, &ttnpb.APIKey{
					Id:   "OTHER",
					Name: "Updated Name",
				}, fieldMask("name"))
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsNotFound(err), should.BeTrue)
				}
			})

			t.Run("DeleteAPIKey", func(t *T) {
				a, ctx := test.New(t)
				err := s.DeleteAPIKey(ctx, ids, &ttnpb.APIKey{Id: id})
				a.So(err, should.BeNil)
			})

			t.Run("GetAPIKeyByID_AfterDelete", func(t *T) {
				a, ctx := test.New(t)
				_, _, err := s.GetAPIKeyByID(ctx, id)
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsNotFound(err), should.BeTrue)
				}
			})

			t.Run("FindAPIKeys_AfterDelete", func(t *T) {
				a, ctx := test.New(t)
				got, err := s.FindAPIKeys(ctx, ids)
				if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
					a.So(got, should.BeEmpty)
				}
			})

			t.Run("DeleteEntityAPIKeys", func(t *T) {
				a, ctx := test.New(t)

				_, err := s.CreateAPIKey(ctx, ids, &ttnpb.APIKey{
					Id:     "ALT" + id,
					Key:    "Hash",
					Name:   "API Key Name",
					Rights: someRights[ids.EntityType()].GetRights(),
				})
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}

				err = s.DeleteEntityAPIKeys(ctx, ids)
				a.So(err, should.BeNil)
			})
		})
	}
}

func (st *StoreTest) TestAPIKeyStorePagination(t *T) {
	app1 := st.population.NewApplication(nil)

	var all []*ttnpb.APIKey
	for i := 0; i < 7; i++ {
		_, key := st.population.NewAPIKey(app1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_APPLICATION_ALL)
		key.Name = fmt.Sprintf("Key %d", i)
		all = append(all, key)
	}

	sort.Sort(apiKeysByID(all))

	s, ok := st.PrepareDB(t).(interface {
		Store

		is.APIKeyStore
	})
	defer st.DestroyDB(t, false)
	if !ok {
		t.Skip("Store does not implement APIKeyStore")
	}
	defer s.Close()

	t.Run("FindAPIKeys_Paginated", func(t *T) {
		a, ctx := test.New(t)

		var total uint64
		for _, page := range []uint32{1, 2, 3, 4} {
			paginateCtx := store.WithPagination(ctx, 2, page, &total)

			got, err := s.FindAPIKeys(paginateCtx, app1.GetEntityIdentifiers())
			if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
				if page == 4 {
					a.So(got, should.HaveLength, 1)
				} else {
					a.So(got, should.HaveLength, 2)
				}
				for i, e := range got {
					a.So(e.Name, should.Equal, all[i+2*int(page-1)].Name)
				}
			}

			a.So(total, should.Equal, 7)
		}
	})
}

// TestAPIKeyStorePaginationDefaults tests the default pagination values.
func (st *StoreTest) TestAPIKeyStorePaginationDefaults(t *T) {
	app1 := st.population.NewApplication(nil)

	var all []*ttnpb.APIKey
	for i := 0; i < 102; i++ {
		_, key := st.population.NewAPIKey(app1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_APPLICATION_ALL)
		key.Name = fmt.Sprintf("Key %d", i)
		all = append(all, key)
	}

	sort.Sort(apiKeysByID(all))

	s, ok := st.PrepareDB(t).(interface {
		Store

		is.APIKeyStore
	})
	defer st.DestroyDB(t, false)
	if !ok {
		t.Skip("Store does not implement APIKeyStore")
	}
	defer s.Close()

	t.Run("FindAPIKeys_PageLimit", func(t *T) {
		a, ctx := test.New(t)

		var total uint64
		paginateCtx := store.WithPagination(ctx, 0, 1, &total)
		got, err := s.FindAPIKeys(paginateCtx, app1.GetEntityIdentifiers())
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.HaveLength, 100)
		}
	})
}
