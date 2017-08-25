// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package mapstore

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestMapStore(t *testing.T) {
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

	s := New()

	id, err := s.Create(old)
	a.So(err, should.BeNil)

	obj, err := s.Find(id)
	a.So(err, should.BeNil)
	a.So(obj, should.Resemble, old)

	err = s.Update(id, new, old)
	a.So(err, should.BeNil)

	obj, err = s.Find(id)
	a.So(err, should.BeNil)
	a.So(obj, should.Resemble, new)

	matches, err := s.FindBy(map[string]interface{}{
		"bar": "bar",
	})
	a.So(err, should.BeNil)
	a.So(matches, should.HaveLength, 1)

	err = s.Delete(id)
	a.So(err, should.BeNil)

	obj, err = s.Find(id)
	a.So(err, should.Equal, store.ErrNotFound)

	err = s.Update(id, new, old)
	a.So(err, should.Equal, store.ErrNotFound)
}
