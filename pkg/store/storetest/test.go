// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package storetest

import (
	"bytes"
	"crypto/rand"
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
func TestTypedStore(t testingT, newStore func() store.TypedStore) {
	a := assertions.New(t)

	s := newStore()

	id1, err := s.Create(make(map[string]interface{}))
	a.So(err, should.BeNil)
	a.So(id1, should.NotBeNil)

	id2, err := s.Create(make(map[string]interface{}))
	a.So(err, should.BeNil)
	a.So(id2, should.NotBeNil)

	a.So(id1, should.NotResemble, id2)

	err = s.Update(id1, make(map[string]interface{}))
	a.So(err, should.BeNil)
	err = s.Update(id2, make(map[string]interface{}))
	a.So(err, should.BeNil)

	// Behavior is implementation-dependent
	a.So(func() { s.Find(id1) }, should.NotPanic)
	a.So(func() { s.Find(id2) }, should.NotPanic)

	err = s.Delete(id1)
	a.So(err, should.BeNil)
	err = s.Delete(id2)
	a.So(err, should.BeNil)

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
				"a.c.b": "acb",
			},
			map[string]interface{}{
				"a.b": nil,
				"a.c": "ac",
			},
			map[string]interface{}{
				"a.a":   1,
				"a.bar": "foo",
				"a.c":   "ac",
			},
			map[string]interface{}{
				"a.a": 1,
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			s := newStore()

			id, err := s.Create(tc.Stored)
			if !a.So(err, should.BeNil) {
				return
			}
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
			if !a.So(err, should.BeNil) {
				return
			}

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
			if !a.So(err, should.BeNil) {
				return
			}

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
func TestByteStore(t testingT, newStore func() store.ByteStore) {
	a := assertions.New(t)

	s := newStore()

	id1, err := s.Create(make(map[string][]byte))
	a.So(err, should.BeNil)
	a.So(id1, should.NotBeNil)

	id2, err := s.Create(make(map[string][]byte))
	a.So(err, should.BeNil)
	a.So(id2, should.NotBeNil)

	a.So(id1, should.NotResemble, id2)

	err = s.Update(id1, make(map[string][]byte))
	a.So(err, should.BeNil)
	err = s.Update(id2, make(map[string][]byte))
	a.So(err, should.BeNil)

	// Behavior is implementation-dependent
	a.So(func() { s.Find(id1) }, should.NotPanic)
	a.So(func() { s.Find(id2) }, should.NotPanic)

	err = s.Delete(id1)
	a.So(err, should.BeNil)
	err = s.Delete(id2)
	a.So(err, should.BeNil)

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
				"qux": []byte("qux"),
				"hey": nil,
			},
			map[string][]byte{
				"foo": []byte("foo"),
				"bar": []byte("bar"),
				"baz": []byte("baz"),
				"qux": []byte("qux"),
			},
			map[string][]byte{
				"foo": []byte("foo"),
				"bar": []byte("bar"),
			},
		},
		{
			map[string][]byte{
				"a.a":   {1},
				"a.bar": []byte("foo"),
				"a.b.a": []byte("1"),
				"a.b.c": []byte("foo"),
				"a.c.b": []byte("acb"),
			},
			map[string][]byte{
				"a.b": nil,
				"a.c": []byte("ac"),
			},
			map[string][]byte{
				"a.a":   {1},
				"a.bar": []byte("foo"),
				"a.c":   []byte("ac"),
			},
			map[string][]byte{
				"a.a": {1},
			},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			s := newStore()

			id, err := s.Create(tc.Stored)
			if !a.So(err, should.BeNil) {
				return
			}
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
			if !a.So(err, should.BeNil) {
				return
			}

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
			if !a.So(err, should.BeNil) {
				return
			}

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

func sameElements(a, b [][]byte) bool {
	bm := make([]bool, len(b))
outer:
	for i := range a {
		for j := range b {
			if !bm[j] && bytes.Equal(a[i], b[j]) {
				bm[j] = true
				continue outer
			}
		}
		return false
	}

	// Check if all values in b have been marked
	for _, v := range bm {
		if !v {
			return false
		}
	}
	return true
}

func randBytes(n int, exclude ...[]byte) []byte {
	rb := make([]byte, n)
	rand.Read(rb)
outer:
	for {
		for _, eb := range exclude {
			if bytes.Equal(rb, eb) {
				rand.Read(rb)
				continue outer
			}
		}
		return rb
	}
}

func TestByteSetStore(t testingT, newStore func() store.ByteSetStore) {
	a := assertions.New(t)

	s := newStore()

	id1, err := s.CreateSet()
	a.So(err, should.BeNil)
	a.So(id1, should.NotBeNil)

	id2, err := s.CreateSet()
	a.So(err, should.BeNil)
	a.So(id2, should.NotBeNil)

	a.So(id1, should.NotResemble, id2)

	// Behavior is implementation-dependent
	a.So(func() { s.FindSet(id1) }, should.NotPanic)
	a.So(func() { s.FindSet(id2) }, should.NotPanic)
	a.So(func() { s.Contains(id1, []byte("non-existent")) }, should.NotPanic)
	a.So(func() { s.Contains(id2, []byte("non-existent")) }, should.NotPanic)

	err = s.Remove(id1)
	a.So(err, should.BeNil)
	err = s.Remove(id2)
	a.So(err, should.BeNil)

	err = s.Delete(id1)
	a.So(err, should.BeNil)
	err = s.Delete(id2)
	a.So(err, should.BeNil)

	for i, tc := range []struct {
		Create      [][]byte
		AfterCreate [][]byte
		Put         [][]byte
		AfterPut    [][]byte
		Remove      [][]byte
		AfterRemove [][]byte
	}{
		{
			[][]byte{[]byte("foo")},
			[][]byte{[]byte("foo")},
			[][]byte{[]byte("bar")},
			[][]byte{[]byte("foo"), []byte("bar")},
			[][]byte{[]byte("foo")},
			[][]byte{[]byte("bar")},
		},
		{
			[][]byte{[]byte("foo"), []byte("foo"), []byte("bar"), []byte("baz"), []byte("bar")},
			[][]byte{[]byte("foo"), []byte("bar"), []byte("baz")},
			[][]byte{[]byte("bar"), []byte("bar"), []byte("baz"), []byte("42")},
			[][]byte{[]byte("foo"), []byte("bar"), []byte("baz"), []byte("42")},
			[][]byte{[]byte("bam"), []byte("bar"), []byte("foo"), []byte("bar"), []byte("baz"), []byte("bar")},
			[][]byte{[]byte("42")},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			s := newStore()

			id, err := s.CreateSet(tc.Create...)
			if !a.So(err, should.BeNil) {
				return
			}
			a.So(id, should.NotBeNil)

			found, err := s.FindSet(id)
			a.So(err, should.BeNil)
			a.So(sameElements(found, tc.AfterCreate), should.BeTrue)

			for _, b := range tc.Create {
				v, err := s.Contains(id, b)
				a.So(err, should.BeNil)
				a.So(v, should.BeTrue)
			}
			v, err := s.Contains(id, randBytes(5, tc.Create...))
			a.So(err, should.BeNil)
			a.So(v, should.BeFalse)

			err = s.Put(id, tc.Put...)
			a.So(err, should.BeNil)
			found, err = s.FindSet(id)
			a.So(err, should.BeNil)
			a.So(sameElements(found, tc.AfterPut), should.BeTrue)

			found, err = s.FindSet(id)
			a.So(err, should.BeNil)
			a.So(sameElements(found, tc.AfterPut), should.BeTrue)

			for _, b := range tc.AfterPut {
				v, err := s.Contains(id, b)
				a.So(err, should.BeNil)
				a.So(v, should.BeTrue)
			}
			v, err = s.Contains(id, randBytes(5, tc.AfterPut...))
			a.So(err, should.BeNil)
			a.So(v, should.BeFalse)

			err = s.Remove(id, tc.Remove...)
			a.So(err, should.BeNil)
			found, err = s.FindSet(id)
			a.So(err, should.BeNil)
			a.So(sameElements(found, tc.AfterRemove), should.BeTrue)

			found, err = s.FindSet(id)
			a.So(err, should.BeNil)
			a.So(sameElements(found, tc.AfterRemove), should.BeTrue)

			for _, b := range tc.AfterRemove {
				v, err := s.Contains(id, b)
				a.So(err, should.BeNil)
				a.So(v, should.BeTrue)
			}
			v, err = s.Contains(id, randBytes(5, tc.AfterRemove...))
			a.So(err, should.BeNil)
			a.So(v, should.BeFalse)

			err = s.Delete(id)
			if !a.So(err, should.BeNil) {
				return
			}
		})
	}
}
