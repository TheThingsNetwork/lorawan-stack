// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package storetest

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

type testingT interface {
	Error(args ...interface{})
}

// TestTypedStore executes a black-box test for the given typed store
func TestTypedStore(t *testing.T, s store.TypedStore) {
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

	id, err := s.Create(old)
	a.So(err, should.BeNil)

	other := map[string]interface{}{
		"hello": "world",
	}
	otherID, _ := s.Create(other)

	a.So(otherID, should.NotEqual, id)

	fields, err := s.Find(id)
	a.So(err, should.BeNil)
	a.So(fields, should.Resemble, old)

	fields, err = s.Find(otherID)
	a.So(err, should.BeNil)
	a.So(fields, should.Resemble, other)

	err = s.Update(id, store.Diff(new, old))
	a.So(err, should.BeNil)

	fields, err = s.Find(id)
	a.So(err, should.BeNil)
	a.So(fields, should.Resemble, new)

	matches, err := s.FindBy(map[string]interface{}{
		"foo": "baz",
		"bar": "bar",
	})
	a.So(err, should.BeNil)
	if a.So(matches, should.HaveLength, 1) {
		for _, m := range matches {
			a.So(m, should.Resemble, new)
		}
	}

	matches, err = s.FindBy(map[string]interface{}{
		"foo": "foo",
		"bar": "bar",
	})
	a.So(err, should.BeNil)
	a.So(matches, should.HaveLength, 0)

	err = s.Delete(id)
	a.So(err, should.BeNil)

	fields, err = s.Find(id)
	a.So(err, should.Equal, store.ErrNotFound)
	a.So(fields, should.Equal, nil)
}

func TestByteStore(t testingT, s store.ByteStore) {
	a := assertions.New(t)

	old := map[string][]byte{
		"foo": []byte("foo"),
		"bar": []byte("bar"),
		"baz": []byte("baz"),
	}
	new := map[string][]byte{
		"foo": []byte("baz"),
		"bar": []byte("bar"),
		"qux": []byte("qux"),
	}

	id, err := s.Create(old)
	a.So(err, should.BeNil)

	other := map[string][]byte{
		"hello": []byte("world"),
	}
	otherID, _ := s.Create(other)

	a.So(otherID, should.NotEqual, id)

	fields, err := s.Find(id)
	a.So(err, should.BeNil)
	a.So(fields, should.Resemble, old)

	fields, err = s.Find(otherID)
	a.So(err, should.BeNil)
	a.So(fields, should.Resemble, other)

	err = s.Update(id, store.ByteDiff(new, old))
	a.So(err, should.BeNil)

	fields, err = s.Find(id)
	a.So(err, should.BeNil)
	a.So(fields, should.Resemble, new)

	matches, err := s.FindBy(map[string][]byte{
		"foo": []byte("baz"),
		"bar": []byte("bar"),
	})
	a.So(err, should.BeNil)
	if a.So(matches, should.HaveLength, 1) {
		for _, m := range matches {
			a.So(m, should.Resemble, new)
		}
	}

	matches, err = s.FindBy(map[string][]byte{
		"foo": []byte("foo"),
		"bar": []byte("bar"),
	})
	a.So(err, should.BeNil)
	a.So(matches, should.HaveLength, 0)

	err = s.Delete(id)
	a.So(err, should.BeNil)

	fields, err = s.Find(id)
	a.So(err, should.Equal, store.ErrNotFound)
	a.So(fields, should.Equal, nil)
}
