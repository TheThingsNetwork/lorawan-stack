// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package storetest

import (
	"strconv"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

type testingT interface {
	Error(args ...interface{})
	Run(string, func(t *testing.T)) bool
}

// TestTypedStore executes a black-box test for the given typed store
func TestTypedStore(t testingT, s store.TypedStore) {
	a := assertions.New(t)

	id1, err := s.Create(map[string]interface{}{})
	a.So(err, should.BeNil)
	a.So(id1, should.NotBeNil)

	id2, err := s.Create(map[string]interface{}{})
	a.So(err, should.BeNil)
	a.So(id2, should.NotBeNil)

	a.So(id1, should.NotResemble, id2)

	for i, tc := range []struct {
		Stored      map[string]interface{}
		Updated     map[string]interface{}
		AfterUpdate map[string]interface{}
		FindBy      map[string]interface{}
	}{
		{
			map[string]interface{}{
				"foo": "foo",
				"bar": "bar",
				"baz": "baz",
				"hey": "there",
			},
			map[string]interface{}{
				"foo": "baz",
				"bar": "bar",
				"qux": "qux",
				"hey": nil,
			},
			map[string]interface{}{
				"foo": "baz",
				"bar": "bar",
				"baz": "baz",
				"qux": "qux",
			},
			map[string]interface{}{
				"bar": "bar",
			},
		},
		{
			map[string]interface{}{
				"a.a":   1,
				"a.bar": "foo",
				"a.b.a": "1",
				"a.b.c": "foo",
			},
			map[string]interface{}{
				"a.b": nil,
			},
			map[string]interface{}{
				"a.a":   1,
				"a.bar": "foo",
			},
			map[string]interface{}{
				"a.a": 1,
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			id, err := s.Create(tc.Stored)
			a.So(err, should.BeNil)
			a.So(id, should.NotBeNil)

			found, err := s.Find(id)
			a.So(err, should.BeNil)
			a.So(found, should.Resemble, tc.Stored)

			matches, err := s.FindBy(tc.Stored)
			a.So(err, should.BeNil)
			if a.So(matches, should.HaveLength, 1) {
				for _, v := range matches {
					a.So(v, should.Resemble, tc.Stored)
				}
			}

			matches, err = s.FindBy(tc.FindBy)
			a.So(err, should.BeNil)
			if a.So(matches, should.HaveLength, 1) {
				for _, v := range matches {
					a.So(v, should.Resemble, tc.Stored)
				}
			}

			err = s.Update(id, tc.Updated)
			a.So(err, should.BeNil)

			found, err = s.Find(id)
			a.So(err, should.BeNil)
			a.So(found, should.Resemble, tc.AfterUpdate)

			matches, err = s.FindBy(tc.AfterUpdate)
			a.So(err, should.BeNil)
			if a.So(matches, should.HaveLength, 1) {
				for _, v := range matches {
					a.So(v, should.Resemble, tc.AfterUpdate)
				}
			}

			matches, err = s.FindBy(tc.FindBy)
			a.So(err, should.BeNil)
			if a.So(matches, should.HaveLength, 1) {
				for _, v := range matches {
					a.So(v, should.Resemble, tc.AfterUpdate)
				}
			}

			err = s.Delete(id)
			a.So(err, should.BeNil)

			found, err = s.Find(id)
			a.So(err, should.Equal, store.ErrNotFound)
			a.So(found, should.Equal, nil)

			matches, err = s.FindBy(tc.AfterUpdate)
			a.So(err, should.BeNil)
			a.So(matches, should.HaveLength, 0)

			matches, err = s.FindBy(tc.FindBy)
			a.So(err, should.BeNil)
			a.So(matches, should.HaveLength, 0)
		})
	}
}

// TestByteStore executes a black-box test for the given byte store
func TestByteStore(t testingT, s store.ByteStore) {
	a := assertions.New(t)

	id1, err := s.Create(map[string][]byte{})
	a.So(err, should.BeNil)
	a.So(id1, should.NotBeNil)

	id2, err := s.Create(map[string][]byte{})
	a.So(err, should.BeNil)
	a.So(id2, should.NotBeNil)

	a.So(id1, should.NotResemble, id2)

	for i, tc := range []struct {
		Stored      map[string][]byte
		Updated     map[string][]byte
		AfterUpdate map[string][]byte
		FindBy      map[string][]byte
	}{
		{
			map[string][]byte{
				"foo": []byte("foo"),
				"bar": []byte("bar"),
				"baz": []byte("baz"),
				"hey": []byte("there"),
			},
			map[string][]byte{
				"foo": []byte("foo"),
				"bar": []byte("bar"),
				"qux": []byte("qux"),
				"hey": nil,
			},
			map[string][]byte{
				"foo": []byte("baz"),
				"bar": []byte("bar"),
				"qux": []byte("qux"),
			},
			map[string][]byte{
				"foo": []byte("baz"),
				"bar": []byte("bar"),
			},
		},
		{
			map[string][]byte{
				"a.a":   {1},
				"a.b.a": []byte("1"),
				"a.b.c": []byte("foo"),
				"a.c.a": []byte("bar"),
			},
			map[string][]byte{
				"a.b": nil,
				"a.c": nil,
			},
			map[string][]byte{
				"a.a": {1},
			},
			map[string][]byte{
				"a.a": {1},
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			id, err := s.Create(tc.Stored)
			a.So(err, should.BeNil)
			a.So(id, should.NotBeNil)

			found, err := s.Find(id)
			a.So(err, should.BeNil)
			a.So(found, should.Resemble, tc.Stored)

			matches, err := s.FindBy(tc.Stored)
			a.So(err, should.BeNil)
			if a.So(matches, should.HaveLength, 1) {
				for _, v := range matches {
					a.So(v, should.Resemble, tc.Stored)
				}
			}

			matches, err = s.FindBy(tc.FindBy)
			a.So(err, should.BeNil)
			if a.So(matches, should.HaveLength, 1) {
				for _, v := range matches {
					a.So(v, should.Resemble, tc.Stored)
				}
			}

			err = s.Update(id, tc.Updated)
			a.So(err, should.BeNil)

			found, err = s.Find(id)
			a.So(err, should.BeNil)
			a.So(found, should.Resemble, tc.AfterUpdate)

			matches, err = s.FindBy(tc.AfterUpdate)
			a.So(err, should.BeNil)
			if a.So(matches, should.HaveLength, 1) {
				for _, v := range matches {
					a.So(v, should.Resemble, tc.AfterUpdate)
				}
			}

			matches, err = s.FindBy(tc.FindBy)
			a.So(err, should.BeNil)
			if a.So(matches, should.HaveLength, 1) {
				for _, v := range matches {
					a.So(v, should.Resemble, tc.AfterUpdate)
				}
			}

			err = s.Delete(id)
			a.So(err, should.BeNil)

			found, err = s.Find(id)
			a.So(err, should.Equal, store.ErrNotFound)
			a.So(found, should.Equal, nil)

			matches, err = s.FindBy(tc.AfterUpdate)
			a.So(err, should.BeNil)
			a.So(matches, should.HaveLength, 0)

			matches, err = s.FindBy(tc.FindBy)
			a.So(err, should.BeNil)
			a.So(matches, should.HaveLength, 0)
		})
	}
}
