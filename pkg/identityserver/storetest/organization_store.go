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

func (st *StoreTest) TestOrganizationStoreCRUD(t *T) {
	usr1 := st.population.NewUser()
	usr2 := st.population.NewUser()

	s, ok := st.PrepareDB(t).(interface {
		Store
		is.OrganizationStore
	})
	defer st.DestroyDB(t, true)
	if !ok {
		t.Skip("Store does not implement OrganizationStore")
	}
	defer s.Close()

	mask := fieldMask(ttnpb.OrganizationFieldPathsTopLevel...)

	var created *ttnpb.Organization

	t.Run("CreateOrganization", func(t *T) {
		a, ctx := test.New(t)
		var err error
		start := time.Now().Truncate(time.Second)

		created, err = s.CreateOrganization(ctx, &ttnpb.Organization{
			Ids:                   &ttnpb.OrganizationIdentifiers{OrganizationId: "foo"},
			Name:                  "Foo Name",
			Description:           "Foo Description",
			Attributes:            attributes,
			AdministrativeContact: usr1.GetOrganizationOrUserIdentifiers(),
			TechnicalContact:      usr2.GetOrganizationOrUserIdentifiers(),
		})

		if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) {
			a.So(created.GetIds().GetOrganizationId(), should.Equal, "foo")
			a.So(created.Name, should.Equal, "Foo Name")
			a.So(created.Description, should.Equal, "Foo Description")
			a.So(created.Attributes, should.Resemble, attributes)
			a.So(created.AdministrativeContact, should.Resemble, usr1.GetOrganizationOrUserIdentifiers())
			a.So(created.TechnicalContact, should.Resemble, usr2.GetOrganizationOrUserIdentifiers())
			a.So(*ttnpb.StdTime(created.CreatedAt), should.HappenWithin, 5*time.Second, start)
			a.So(*ttnpb.StdTime(created.UpdatedAt), should.HappenWithin, 5*time.Second, start)
		}
	})

	t.Run("CreateOrganization_AfterCreate", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.CreateOrganization(ctx, &ttnpb.Organization{
			Ids: &ttnpb.OrganizationIdentifiers{OrganizationId: "foo"},
		})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsAlreadyExists(err), should.BeTrue)
		}
	})

	t.Run("GetOrganization", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.GetOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: "foo"}, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.Resemble, created)
		}
	})

	t.Run("GetOrganization_Other", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.GetOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: "other"}, mask)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// _, err = s.GetOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: ""}, mask)
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("CountOrganizations", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.CountOrganizations(ctx)
		if a.So(err, should.BeNil) {
			a.So(got, should.Equal, 1)
		}
	})

	t.Run("FindOrganizations", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.FindOrganizations(ctx, nil, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 1) {
			a.So(got[0], should.Resemble, created)
		}
	})

	var updated *ttnpb.Organization

	t.Run("UpdateOrganization", func(t *T) {
		a, ctx := test.New(t)
		var err error
		start := time.Now().Truncate(time.Second)

		updated, err = s.UpdateOrganization(ctx, &ttnpb.Organization{
			Ids:                   &ttnpb.OrganizationIdentifiers{OrganizationId: "foo"},
			Name:                  "New Foo Name",
			Description:           "New Foo Description",
			Attributes:            updatedAttributes,
			AdministrativeContact: usr2.GetOrganizationOrUserIdentifiers(),
			TechnicalContact:      usr1.GetOrganizationOrUserIdentifiers(),
		}, mask)
		if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) {
			a.So(updated.GetIds().GetOrganizationId(), should.Equal, "foo")
			a.So(updated.Name, should.Equal, "New Foo Name")
			a.So(updated.Description, should.Equal, "New Foo Description")
			a.So(updated.Attributes, should.Resemble, updatedAttributes)
			a.So(updated.AdministrativeContact, should.Resemble, usr2.GetOrganizationOrUserIdentifiers())
			a.So(updated.TechnicalContact, should.Resemble, usr1.GetOrganizationOrUserIdentifiers())
			a.So(*ttnpb.StdTime(updated.CreatedAt), should.Equal, *ttnpb.StdTime(created.CreatedAt))
			a.So(*ttnpb.StdTime(updated.UpdatedAt), should.HappenWithin, 5*time.Second, start)
		}
	})

	t.Run("UpdateOrganization_Other", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.UpdateOrganization(ctx, &ttnpb.Organization{
			Ids: &ttnpb.OrganizationIdentifiers{OrganizationId: "other"},
		}, mask)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// _, err = s.UpdateOrganization(ctx, &ttnpb.Organization{
		// 	Ids: &ttnpb.OrganizationIdentifiers{OrganizationId: ""},
		// }, mask)
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("GetOrganization_AfterUpdate", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.GetOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: "foo"}, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.Resemble, updated)
		}
	})

	t.Run("DeleteOrganization", func(t *T) {
		a, ctx := test.New(t)
		err := s.DeleteOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: "foo"})
		a.So(err, should.BeNil)
	})

	t.Run("DeleteOrganization_Other", func(t *T) {
		a, ctx := test.New(t)
		err := s.DeleteOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: "other"})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// err = s.DeleteOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: ""})
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("GetOrganization_AfterDelete", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.GetOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: "foo"}, mask)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
	})

	t.Run("FindOrganizations_AfterDelete", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.FindOrganizations(ctx, nil, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.BeEmpty)
		}
	})

	t.Run("GetDeletedOrganization", func(t *T) {
		a, ctx := test.New(t)
		ctx = store.WithSoftDeleted(ctx, true)
		got, err := s.GetOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: "foo"}, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			if a.So(got.DeletedAt, should.NotBeNil) {
				got.DeletedAt = nil // Unset DeletedAt for the should.Resemble below.
			}
			a.So(got, should.Resemble, updated)
		}
	})

	t.Run("FindDeletedOrganizations", func(t *T) {
		a, ctx := test.New(t)
		ctx = store.WithSoftDeleted(ctx, true)
		got, err := s.FindOrganizations(ctx, nil, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 1) {
			if a.So(got[0].DeletedAt, should.NotBeNil) {
				got[0].DeletedAt = nil // Unset DeletedAt for the should.Resemble below.
			}
			a.So(got[0], should.Resemble, updated)
		}
	})

	t.Run("RestoreOrganization", func(t *T) {
		a, ctx := test.New(t)
		err := s.RestoreOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: "foo"})
		a.So(err, should.BeNil)
	})

	t.Run("RestoreOrganization_Other", func(t *T) {
		a, ctx := test.New(t)
		err := s.RestoreOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: "other"})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// err = s.RestoreOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: ""})
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("GetOrganization_AfterRestore", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.GetOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: "foo"}, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.Resemble, updated)
		}
	})

	t.Run("PurgeOrganization", func(t *T) {
		a, ctx := test.New(t)
		err := s.PurgeOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: "foo"})
		a.So(err, should.BeNil)
	})

	t.Run("PurgeOrganization_Other", func(t *T) {
		a, ctx := test.New(t)
		err := s.PurgeOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: "other"})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// err = s.PurgeOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: ""})
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("CreateAfterPurge", func(t *T) {
		for _, itr := range []int{1, 2} {
			t.Run(fmt.Sprintf("Iteration %d", itr), func(t *T) {
				a, ctx := test.New(t)
				var err error
				_, err = s.CreateOrganization(ctx, &ttnpb.Organization{
					Ids: &ttnpb.OrganizationIdentifiers{OrganizationId: "foo"},
				})
				a.So(err, should.BeNil)

				err = s.DeleteOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: "foo"})
				a.So(err, should.BeNil)

				err = s.RestoreOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: "foo"})
				a.So(err, should.BeNil)

				got, err := s.GetOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: "foo"}, mask)
				a.So(err, should.BeNil)
				a.So(got, should.NotBeNil)

				err = s.DeleteOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: "foo"})
				a.So(err, should.BeNil)

				err = s.PurgeOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: "foo"})
				a.So(err, should.BeNil)
			})
		}
	})
}

func (st *StoreTest) TestOrganizationStorePagination(t *T) {
	usr1 := st.population.NewUser()

	var all []*ttnpb.Organization
	for i := 0; i < 7; i++ {
		all = append(all, st.population.NewOrganization(usr1.GetOrganizationOrUserIdentifiers()))
	}

	s, ok := st.PrepareDB(t).(interface {
		Store
		is.OrganizationStore
	})
	defer st.DestroyDB(t, false)
	if !ok {
		t.Skip("Store does not implement OrganizationStore")
	}
	defer s.Close()

	t.Run("FindOrganizations_Paginated", func(t *T) {
		a, ctx := test.New(t)

		var total uint64
		for _, page := range []uint32{1, 2, 3, 4} {
			paginateCtx := store.WithPagination(ctx, 2, page, &total)

			got, err := s.FindOrganizations(paginateCtx, nil, fieldMask(ttnpb.OrganizationFieldPathsTopLevel...))
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
