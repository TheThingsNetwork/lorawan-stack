// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"strconv"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestErrorDuplicate(t *testing.T) {
	a := assertions.New(t)
	db := getInstance(t)

	var id int64
	err := db.SelectOne(&id, `INSERT INTO foo (bar) VALUES ($1) RETURNING id`, "bar")
	a.So(err, should.BeNil)

	// try to reinsert a record with a duplicated primary key
	_, err = db.Exec(`INSERT INTO foo (id) VALUES ($1)`, id)
	a.So(err, should.NotBeNil)
	duplicates, ok := IsDuplicate(err)
	a.So(ok, should.BeTrue)
	a.So(duplicates, should.HaveLength, 1)
	a.So(duplicates["id"], should.Equal, strconv.FormatInt(id, 10))

	// delete record
	_, err = db.Exec(`DELETE FROM foo WHERE id = $1`, id)
	a.So(err, should.BeNil)
}
