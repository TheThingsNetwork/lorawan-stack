// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package store

import (
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

func TestOrganizationStore(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		prepareTest(db, &Account{}, &Organization{}, &Attribute{})
		store := GetOrganizationStore(db)

		created, err := store.CreateOrganization(ctx, &ttnpb.Organization{
			OrganizationIdentifiers: ttnpb.OrganizationIdentifiers{OrganizationID: "foo"},
			Name:                    "Foo Organization",
			Description:             "The Amazing Foo Organization",
			Attributes: map[string]string{
				"foo": "bar",
				"bar": "baz",
				"baz": "qux",
			},
		})

		a.So(err, should.BeNil)
		if a.So(created, should.NotBeNil) {
			a.So(created.OrganizationID, should.Equal, "foo")
			a.So(created.Name, should.Equal, "Foo Organization")
			a.So(created.Description, should.Equal, "The Amazing Foo Organization")
			a.So(created.Attributes, should.HaveLength, 3)
			a.So(created.CreatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
			a.So(created.UpdatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
		}

		got, err := store.GetOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationID: "foo"}, &types.FieldMask{Paths: []string{"name", "attributes"}})

		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.OrganizationID, should.Equal, "foo")
			a.So(got.Name, should.Equal, "Foo Organization")
			a.So(got.Description, should.BeEmpty)
			a.So(got.Attributes, should.HaveLength, 3)
			a.So(got.CreatedAt, should.Equal, created.CreatedAt)
			a.So(got.UpdatedAt, should.Equal, created.UpdatedAt)
		}

		_, err = store.UpdateOrganization(ctx, &ttnpb.Organization{
			OrganizationIdentifiers: ttnpb.OrganizationIdentifiers{OrganizationID: "bar"},
		}, nil)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		updated, err := store.UpdateOrganization(ctx, &ttnpb.Organization{
			OrganizationIdentifiers: ttnpb.OrganizationIdentifiers{OrganizationID: "foo"},
			Name:                    "Foobar Organization",
			Description:             "The Amazing Foobar Organization",
			Attributes: map[string]string{
				"foo": "bar",
				"baz": "baz",
				"qux": "foo",
			},
		}, &types.FieldMask{Paths: []string{"description", "attributes"}})

		a.So(err, should.BeNil)
		if a.So(updated, should.NotBeNil) {
			a.So(updated.Description, should.Equal, "The Amazing Foobar Organization")
			a.So(updated.Attributes, should.HaveLength, 3)
			a.So(updated.CreatedAt, should.Equal, created.CreatedAt)
			a.So(updated.UpdatedAt, should.HappenAfter, created.CreatedAt)
		}

		got, err = store.GetOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationID: "foo"}, nil)

		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.OrganizationID, should.Equal, created.OrganizationID)
			a.So(got.Name, should.Equal, created.Name)
			a.So(got.Description, should.Equal, updated.Description)
			a.So(got.Attributes, should.Resemble, updated.Attributes)
			a.So(got.CreatedAt, should.Equal, created.CreatedAt)
			a.So(got.UpdatedAt, should.Equal, updated.UpdatedAt)
		}

		list, err := store.FindOrganizations(ctx, nil, &types.FieldMask{Paths: []string{"name"}})

		a.So(err, should.BeNil)
		if a.So(list, should.HaveLength, 1) {
			a.So(list[0].Name, should.EndWith, got.Name)
		}

		err = store.DeleteOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationID: "foo"})

		a.So(err, should.BeNil)

		got, err = store.GetOrganization(ctx, &ttnpb.OrganizationIdentifiers{OrganizationID: "foo"}, nil)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		list, err = store.FindOrganizations(ctx, nil, nil)

		a.So(err, should.BeNil)
		a.So(list, should.BeEmpty)
	})
}
