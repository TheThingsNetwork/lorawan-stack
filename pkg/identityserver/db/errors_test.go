// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"strconv"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestErrorDuplicate(t *testing.T) {
	a := assertions.New(t)
	db := getInstance()

	f := data[0]

	var id int64
	err := db.SelectOne(&id, `INSERT INTO foo (bar, baz, quu) VALUES ($1, $2, $3) RETURNING id`, f.Bar, f.Baz, f.Quu)
	a.So(err, should.BeNil)

	// try to reinsert a record with a duplicated primary key
	_, err = db.Exec(`INSERT INTO foo (id) VALUES ($1)`, id)
	a.So(err, should.NotBeNil)
	//a.So(err, should.Implement, errors.Error)
	//./errors_test.go:23: type "github.com/TheThingsNetwork/ttn/pkg/errors".Error is not an expression
	a.So(err.(errors.Error).Code(), should.Equal, ErrDuplicate.Code)
	a.So(err.(errors.Error).Type(), should.Equal, ErrDuplicate.Type)
	duplicates, yes := IsDuplicate(err)
	a.So(yes, should.BeTrue)
	a.So(duplicates, should.HaveLength, 1)
	a.So(duplicates["id"], should.Equal, strconv.FormatInt(id, 10))

	// delete record
	_, err = db.Exec(`DELETE FROM foo WHERE id = $1`, id)
	a.So(err, should.BeNil)
}
