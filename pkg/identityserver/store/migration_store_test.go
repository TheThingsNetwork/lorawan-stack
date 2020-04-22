// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

	"github.com/jinzhu/gorm"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

func TestMigrationStore(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		prepareTest(db, &Migration{})
		store := GetMigrationStore(db)

		err := store.CreateMigration(ctx, &Migration{
			Name: "foo",
		})

		a.So(err, should.BeNil)

		got, err := store.GetMigration(ctx, "foo")

		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.Name, should.Equal, "foo")
			a.So(got.CreatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
			a.So(got.UpdatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
		}

		list, err := store.FindMigrations(ctx)

		a.So(err, should.BeNil)
		if a.So(list, should.HaveLength, 1) {
			a.So(list[0].Name, should.EndWith, got.Name)
		}

		err = store.DeleteMigration(ctx, "foo")

		a.So(err, should.BeNil)

		got, err = store.GetMigration(ctx, "foo")

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		list, err = store.FindMigrations(ctx)

		a.So(err, should.BeNil)
		a.So(list, should.BeEmpty)
	})
}
