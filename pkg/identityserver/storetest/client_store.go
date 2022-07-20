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
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	is "go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func (st *StoreTest) TestClientStoreCRUD(t *T) {
	usr1 := st.population.NewUser()
	org1 := st.population.NewOrganization(usr1.GetOrganizationOrUserIdentifiers())

	s, ok := st.PrepareDB(t).(interface {
		Store
		is.ClientStore
	})
	defer st.DestroyDB(t, true)
	defer s.Close()
	if !ok {
		t.Fatal("Store does not implement ClientStore")
	}

	mask := fieldMask(ttnpb.ClientFieldPathsTopLevel...)

	var created *ttnpb.Client

	t.Run("CreateClient", func(t *T) {
		a, ctx := test.New(t)
		var err error
		start := time.Now().Truncate(time.Second)

		created, err = s.CreateClient(ctx, &ttnpb.Client{
			Ids:                   &ttnpb.ClientIdentifiers{ClientId: "foo"},
			Name:                  "Foo Name",
			Description:           "Foo Description",
			Attributes:            attributes,
			AdministrativeContact: usr1.GetOrganizationOrUserIdentifiers(),
			TechnicalContact:      org1.GetOrganizationOrUserIdentifiers(),
		})

		if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) {
			a.So(created.GetIds().GetClientId(), should.Equal, "foo")
			a.So(created.Name, should.Equal, "Foo Name")
			a.So(created.Description, should.Equal, "Foo Description")
			a.So(created.Attributes, should.Resemble, attributes)
			a.So(created.AdministrativeContact, should.Resemble, usr1.GetOrganizationOrUserIdentifiers())
			a.So(created.TechnicalContact, should.Resemble, org1.GetOrganizationOrUserIdentifiers())
			a.So(*ttnpb.StdTime(created.CreatedAt), should.HappenWithin, 5*time.Second, start)
			a.So(*ttnpb.StdTime(created.UpdatedAt), should.HappenWithin, 5*time.Second, start)
		}
	})

	t.Run("CreateClient_AfterCreate", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.CreateClient(ctx, &ttnpb.Client{
			Ids: &ttnpb.ClientIdentifiers{ClientId: "foo"},
		})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsAlreadyExists(err), should.BeTrue)
		}
	})

	t.Run("GetClient", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.GetClient(ctx, &ttnpb.ClientIdentifiers{ClientId: "foo"}, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.Resemble, created)
		}
	})

	t.Run("GetClient_Other", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.GetClient(ctx, &ttnpb.ClientIdentifiers{ClientId: "other"}, mask)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// _, err = s.GetClient(ctx, &ttnpb.ClientIdentifiers{ClientId: ""}, mask)
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("CountClients", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.CountClients(ctx)
		if a.So(err, should.BeNil) {
			a.So(got, should.Equal, 1)
		}
	})

	t.Run("FindClients", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.FindClients(ctx, nil, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 1) {
			a.So(got[0], should.Resemble, created)
		}
	})

	var updated *ttnpb.Client

	t.Run("UpdateClient", func(t *T) {
		a, ctx := test.New(t)
		var err error
		start := time.Now().Truncate(time.Second)

		updated, err = s.UpdateClient(ctx, &ttnpb.Client{
			Ids:                   &ttnpb.ClientIdentifiers{ClientId: "foo"},
			Name:                  "New Foo Name",
			Description:           "New Foo Description",
			Attributes:            updatedAttributes,
			AdministrativeContact: org1.GetOrganizationOrUserIdentifiers(),
			TechnicalContact:      usr1.GetOrganizationOrUserIdentifiers(),
		}, mask)
		if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) {
			a.So(updated.GetIds().GetClientId(), should.Equal, "foo")
			a.So(updated.Name, should.Equal, "New Foo Name")
			a.So(updated.Description, should.Equal, "New Foo Description")
			a.So(updated.Attributes, should.Resemble, updatedAttributes)
			a.So(updated.AdministrativeContact, should.Resemble, org1.GetOrganizationOrUserIdentifiers())
			a.So(updated.TechnicalContact, should.Resemble, usr1.GetOrganizationOrUserIdentifiers())
			a.So(*ttnpb.StdTime(updated.CreatedAt), should.Equal, *ttnpb.StdTime(created.CreatedAt))
			a.So(*ttnpb.StdTime(updated.UpdatedAt), should.HappenWithin, 5*time.Second, start)
		}
	})

	t.Run("UpdateClient_Other", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.UpdateClient(ctx, &ttnpb.Client{
			Ids: &ttnpb.ClientIdentifiers{ClientId: "other"},
		}, mask)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// _, err = s.UpdateClient(ctx, &ttnpb.Client{
		// 	Ids: &ttnpb.ClientIdentifiers{ClientId: ""},
		// }, mask)
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("GetClient_AfterUpdate", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.GetClient(ctx, &ttnpb.ClientIdentifiers{ClientId: "foo"}, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.Resemble, updated)
		}
	})

	t.Run("DeleteClient", func(t *T) {
		a, ctx := test.New(t)
		err := s.DeleteClient(ctx, &ttnpb.ClientIdentifiers{ClientId: "foo"})
		a.So(err, should.BeNil)
	})

	t.Run("DeleteClient_Other", func(t *T) {
		a, ctx := test.New(t)
		err := s.DeleteClient(ctx, &ttnpb.ClientIdentifiers{ClientId: "other"})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// err = s.DeleteClient(ctx, &ttnpb.ClientIdentifiers{ClientId: ""})
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("GetClient_AfterDelete", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.GetClient(ctx, &ttnpb.ClientIdentifiers{ClientId: "foo"}, mask)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
	})

	t.Run("FindClients_AfterDelete", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.FindClients(ctx, nil, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.BeEmpty)
		}
	})

	t.Run("GetDeletedClient", func(t *T) {
		a, ctx := test.New(t)
		ctx = store.WithSoftDeleted(ctx, true)
		got, err := s.GetClient(ctx, &ttnpb.ClientIdentifiers{ClientId: "foo"}, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			if a.So(got.DeletedAt, should.NotBeNil) {
				got.DeletedAt = nil // Unset DeletedAt for the should.Resemble below.
			}
			a.So(got, should.Resemble, updated)
		}
	})

	t.Run("FindDeletedClients", func(t *T) {
		a, ctx := test.New(t)
		ctx = store.WithSoftDeleted(ctx, true)
		got, err := s.FindClients(ctx, nil, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 1) {
			if a.So(got[0].DeletedAt, should.NotBeNil) {
				got[0].DeletedAt = nil // Unset DeletedAt for the should.Resemble below.
			}
			a.So(got[0], should.Resemble, updated)
		}
	})

	t.Run("RestoreClient", func(t *T) {
		a, ctx := test.New(t)
		err := s.RestoreClient(ctx, &ttnpb.ClientIdentifiers{ClientId: "foo"})
		a.So(err, should.BeNil)
	})

	t.Run("RestoreClient_Other", func(t *T) {
		a, ctx := test.New(t)
		err := s.RestoreClient(ctx, &ttnpb.ClientIdentifiers{ClientId: "other"})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// err = s.RestoreClient(ctx, &ttnpb.ClientIdentifiers{ClientId: ""})
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("GetClient_AfterRestore", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.GetClient(ctx, &ttnpb.ClientIdentifiers{ClientId: "foo"}, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.Resemble, updated)
		}
	})

	t.Run("PurgeClient", func(t *T) {
		a, ctx := test.New(t)
		err := s.PurgeClient(ctx, &ttnpb.ClientIdentifiers{ClientId: "foo"})
		a.So(err, should.BeNil)
	})

	t.Run("PurgeClient_Other", func(t *T) {
		a, ctx := test.New(t)
		err := s.PurgeClient(ctx, &ttnpb.ClientIdentifiers{ClientId: "other"})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// err = s.PurgeClient(ctx, &ttnpb.ClientIdentifiers{ClientId: ""})
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})
}

func (st *StoreTest) TestClientStorePagination(t *T) {
	usr1 := st.population.NewUser()

	var all []*ttnpb.Client
	for i := 0; i < 7; i++ {
		all = append(all, st.population.NewClient(usr1.GetOrganizationOrUserIdentifiers()))
	}

	s, ok := st.PrepareDB(t).(interface {
		Store
		is.ClientStore
	})
	defer st.DestroyDB(t, false)
	defer s.Close()
	if !ok {
		t.Fatal("Store does not implement ClientStore")
	}

	t.Run("FindClients_Paginated", func(t *T) {
		a, ctx := test.New(t)

		var total uint64
		for _, page := range []uint32{1, 2, 3, 4} {
			paginateCtx := store.WithPagination(ctx, 2, page, &total)

			got, err := s.FindClients(paginateCtx, nil, fieldMask(ttnpb.ClientFieldPathsTopLevel...))
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
