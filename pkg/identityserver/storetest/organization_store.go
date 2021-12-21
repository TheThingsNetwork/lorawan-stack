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

func (st *StoreTest) TestOrganizationStoreCRUD(t *T) {
	s, ok := st.PrepareDB(t).(interface {
		Store
		is.OrganizationStore
	})
	defer st.DestroyDB(t, true)
	defer s.Close()
	if !ok {
		t.Fatal("Store does not implement OrganizationStore")
	}

	mask := fieldMask(ttnpb.OrganizationFieldPathsTopLevel...)

	var created *ttnpb.Organization

	t.Run("CreateOrganization", func(t *T) {
		a, ctx := test.New(t)
		var err error
		start := time.Now().Truncate(time.Second)

		created, err = s.CreateOrganization(ctx, &ttnpb.Organization{
			Ids:         &ttnpb.OrganizationIdentifiers{OrganizationId: "foo"},
			Name:        "Foo Name",
			Description: "Foo Description",
			Attributes: map[string]string{
				"foo": "bar",
				"bar": "baz",
				"baz": "qux",
			},
		})

		if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) {
			a.So(created.GetIds().GetOrganizationId(), should.Equal, "foo")
			a.So(created.Name, should.Equal, "Foo Name")
			a.So(created.Description, should.Equal, "Foo Description")
			a.So(created.Attributes, should.Resemble, map[string]string{
				"foo": "bar",
				"bar": "baz",
				"baz": "qux",
			})
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
		// _, err = s.GetOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: ""}, mask)
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
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
			Ids:         &ttnpb.OrganizationIdentifiers{OrganizationId: "foo"},
			Name:        "New Foo Name",
			Description: "New Foo Description",
			Attributes: map[string]string{
				"attribute": "new",
			},
		}, mask)
		if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) {
			a.So(updated.GetIds().GetOrganizationId(), should.Equal, "foo")
			a.So(updated.Name, should.Equal, "New Foo Name")
			a.So(updated.Description, should.Equal, "New Foo Description")
			a.So(updated.Attributes, should.Resemble, map[string]string{
				"attribute": "new",
			})
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
		// err = s.PurgeOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: ""})
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})
}

// TODO: Test Pagination
