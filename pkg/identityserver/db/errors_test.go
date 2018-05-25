// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package db

import (
	"strconv"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestErrorDuplicate(t *testing.T) {
	a := assertions.New(t)
	db := getInstance(t)

	var id int64
	err := db.SelectOne(&id, `INSERT INTO foo (bar) VALUES ($1) RETURNING id`, "bar")
	a.So(err, should.BeNil)

	// Try to reinsert a record with a duplicated primary key.
	_, err = db.Exec(`INSERT INTO foo (id) VALUES ($1)`, id)
	a.So(err, should.NotBeNil)
	duplicates, ok := IsDuplicate(err)
	a.So(ok, should.BeTrue)
	a.So(duplicates, should.HaveLength, 1)
	a.So(duplicates["id"], should.Equal, strconv.FormatInt(id, 10))

	// Delete record.
	_, err = db.Exec(`DELETE FROM foo WHERE id = $1`, id)
	a.So(err, should.BeNil)
}
