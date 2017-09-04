// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package storetest

import (
	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

type testingT interface {
	Error(args ...interface{})
}

// TestStore executes a black-box test for the given store
func TestStore(t testingT, v store.Interface) {
	a := assertions.New(t)

	old := map[string]interface{}{
		"foo": "foo",
		"bar": "bar",
		"baz": "baz",
	}
	new := map[string]interface{}{
		"foo": "baz",
		"bar": "bar",
		"qux": "qux",
	}

	id, err := v.Create(old)
	a.So(err, should.BeNil)

	other := map[string]interface{}{
		"hello": "world",
	}
	otherID, _ := v.Create(other)

	a.So(otherID, should.NotEqual, id)

	fields, err := v.Find(id)
	a.So(err, should.BeNil)
	a.So(fields, should.Resemble, old)

	fields, err = v.Find(otherID)
	a.So(err, should.BeNil)
	a.So(fields, should.Resemble, other)

	err = v.Update(id, store.Diff(new, old))
	a.So(err, should.BeNil)

	fields, err = v.Find(id)
	a.So(err, should.BeNil)
	a.So(fields, should.Resemble, new)

	matches, err := v.FindBy(map[string]interface{}{
		"foo": "baz",
		"bar": "bar",
	})
	a.So(err, should.BeNil)
	a.So(matches, should.HaveLength, 1)

	matches, err = v.FindBy(map[string]interface{}{
		"foo": "foo",
		"bar": "bar",
	})
	a.So(err, should.BeNil)
	a.So(matches, should.HaveLength, 0)

	err = v.Delete(id)
	a.So(err, should.BeNil)

	fields, err = v.Find(id)
	a.So(err, should.Equal, store.ErrNotFound)
	a.So(fields, should.Equal, nil)
}
