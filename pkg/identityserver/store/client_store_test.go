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

func TestClientStore(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		prepareTest(db, &Client{}, &Attribute{})
		store := GetClientStore(db)

		created, err := store.CreateClient(ctx, &ttnpb.Client{
			ClientIdentifiers: ttnpb.ClientIdentifiers{ClientID: "foo"},
			Name:              "Foo Client",
			Description:       "The Amazing Foo Client",
			Attributes: map[string]string{
				"foo": "bar",
				"bar": "baz",
				"baz": "qux",
			},
		})
		a.So(err, should.BeNil)
		a.So(created.ClientID, should.Equal, "foo")
		a.So(created.Name, should.Equal, "Foo Client")
		a.So(created.Description, should.Equal, "The Amazing Foo Client")
		a.So(created.Attributes, should.HaveLength, 3)
		a.So(created.CreatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
		a.So(created.UpdatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))

		got, err := store.GetClient(ctx, &ttnpb.ClientIdentifiers{ClientID: "foo"}, &types.FieldMask{Paths: []string{"name", "attributes"}})
		a.So(err, should.BeNil)
		a.So(got.ClientID, should.Equal, "foo")
		a.So(got.Name, should.Equal, "Foo Client")
		a.So(got.Description, should.BeEmpty)
		a.So(got.Attributes, should.HaveLength, 3)
		a.So(got.CreatedAt, should.Equal, created.CreatedAt)
		a.So(got.UpdatedAt, should.Equal, created.UpdatedAt)

		_, err = store.UpdateClient(ctx, &ttnpb.Client{
			ClientIdentifiers: ttnpb.ClientIdentifiers{ClientID: "bar"},
		}, nil)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		updated, err := store.UpdateClient(ctx, &ttnpb.Client{
			ClientIdentifiers: ttnpb.ClientIdentifiers{ClientID: "foo"},
			Name:              "Foobar Client",
			Description:       "The Amazing Foobar Client",
			Attributes: map[string]string{
				"foo": "bar",
				"baz": "baz",
				"qux": "foo",
			},
		}, &types.FieldMask{Paths: []string{"description", "attributes"}})
		a.So(err, should.BeNil)
		a.So(updated.Description, should.Equal, "The Amazing Foobar Client")
		a.So(updated.Attributes, should.HaveLength, 3)
		a.So(updated.CreatedAt, should.Equal, created.CreatedAt)
		a.So(updated.UpdatedAt, should.HappenAfter, created.CreatedAt)

		got, err = store.GetClient(ctx, &ttnpb.ClientIdentifiers{ClientID: "foo"}, nil)
		a.So(err, should.BeNil)
		a.So(got.ClientID, should.Equal, created.ClientID)
		a.So(got.Name, should.Equal, created.Name)
		a.So(got.Description, should.Equal, updated.Description)
		a.So(got.Attributes, should.Resemble, updated.Attributes)
		a.So(got.CreatedAt, should.Equal, created.CreatedAt)
		a.So(got.UpdatedAt, should.Equal, updated.UpdatedAt)

		list, err := store.FindClients(ctx, nil, &types.FieldMask{Paths: []string{"name"}})
		a.So(err, should.BeNil)
		if a.So(list, should.HaveLength, 1) {
			a.So(list[0].Name, should.EndWith, got.Name)
		}

		err = store.DeleteClient(ctx, &ttnpb.ClientIdentifiers{ClientID: "foo"})
		a.So(err, should.BeNil)

		got, err = store.GetClient(ctx, &ttnpb.ClientIdentifiers{ClientID: "foo"}, nil)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		list, err = store.FindClients(ctx, nil, nil)
		a.So(err, should.BeNil)
		a.So(list, should.BeEmpty)
	})
}
