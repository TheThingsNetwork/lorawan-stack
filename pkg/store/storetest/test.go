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

// Package storetest provides a test suite to test implementations of interfaces defined in store.
package storetest

import (
	"bytes"
	"encoding/gob"
	"math"
	"math/rand"
	"reflect"
	"strconv"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

// IndexedFields represents the fields, which are expected to be indexed in store implementations,
// which support indexed fields.
var IndexedFields = []string{
	"bar",
	"a.a",
	"empty",
}

type testingT interface {
	Error(args ...interface{})
	Run(string, func(t *testing.T)) bool
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

// TestTypedStore executes a black-box test for the given typed store
func TestTypedStore(t testingT, newStore func() store.TypedStore) {
	a := assertions.New(t)

	s := newStore()

	m, err := s.Find(nil)
	a.So(err, should.NotBeNil)
	a.So(m, should.BeNil)

	m, err = s.Find(bytes.NewBufferString("non-existent"))
	a.So(err, should.BeNil)
	a.So(m, should.BeNil)

	filtered, err := s.FindBy(nil)
	a.So(err, should.NotBeNil)
	a.So(filtered, should.BeNil)

	filtered, err = s.FindBy(make(map[string]interface{}))
	a.So(err, should.NotBeNil)
	a.So(filtered, should.BeNil)

	err = s.Update(nil, nil)
	a.So(err, should.NotBeNil)

	err = s.Update(nil, map[string]interface{}{"foo": "bar"})
	a.So(err, should.NotBeNil)

	err = s.Update(bytes.NewBufferString("non-existent"), nil)
	a.So(err, should.BeNil)

	err = s.Update(bytes.NewBufferString("non-existentt"), make(map[string]interface{}))
	a.So(err, should.BeNil)

	id, err := s.Create(make(map[string]interface{}))
	a.So(err, should.BeNil)
	a.So(id, should.NotBeNil)

	id, err = s.Create(nil)
	a.So(err, should.BeNil)
	a.So(id, should.NotBeNil)

	idOther, err := s.Create(nil)
	a.So(err, should.BeNil)
	a.So(id, should.NotBeNil)

	a.So(id, should.NotResemble, idOther)

	m, err = s.Find(id)
	a.So(err, should.BeNil)
	a.So(m, should.BeNil)

	err = s.Delete(id)
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
				"hey": nil,
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
				"a.d":   "hello",
			},
			map[string]interface{}{
				"a.b": nil,
				"a.c": "ac",
			},
			map[string]interface{}{
				"a.a":   1,
				"a.b":   nil,
				"a.bar": "foo",
				"a.c":   "ac",
				"a.d":   "hello",
			},
			map[string]interface{}{
				"a.a": 1,
				"a.d": "hello",
			},
		},
		{
			map[string]interface{}{
				"empty": "",
				"nil":   nil,
			},
			map[string]interface{}{
				"nil.nil": nil,
			},
			map[string]interface{}{
				"empty":   "",
				"nil.nil": nil,
			},
			map[string]interface{}{
				"empty":        "",
				"non-existent": nil,
			},
		},
		{
			map[string]interface{}{
				"empty":   "",
				"nil.nil": nil,
			},
			map[string]interface{}{
				"nil": nil,
			},
			map[string]interface{}{
				"empty": "",
				"nil":   nil,
			},
			map[string]interface{}{
				"empty": "",
				"foo":   nil,
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
			a.So(err, should.BeNil)
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

// TestByteSetStore executes a black-box test for the given byte set store
func TestByteSetStore(t testingT, newStore func() store.ByteSetStore) {
	a := assertions.New(t)

	s := newStore()

	id, err := s.CreateSet()
	a.So(err, should.BeNil)
	a.So(id, should.NotBeNil)

	idOther, err := s.CreateSet()
	a.So(err, should.BeNil)
	a.So(idOther, should.NotBeNil)

	a.So(id, should.NotResemble, idOther)

	// Behavior is implementation-dependent
	a.So(func() { s.FindSet(id) }, should.NotPanic)
	a.So(func() { s.Contains(id, []byte("non-existent")) }, should.NotPanic)
	a.So(func() { s.Remove(id, []byte("non-existent")) }, should.NotPanic)

	err = s.Put(id, []byte("foo"))
	a.So(err, should.BeNil)

	err = s.Put(id)
	a.So(err, should.BeNil)

	err = s.Delete(id)
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
			a.So(test.SameElements(found, tc.AfterCreate), should.BeTrue)

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
			a.So(test.SameElements(found, tc.AfterPut), should.BeTrue)

			found, err = s.FindSet(id)
			a.So(err, should.BeNil)
			a.So(test.SameElements(found, tc.AfterPut), should.BeTrue)

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
			a.So(test.SameElements(found, tc.AfterRemove), should.BeTrue)

			found, err = s.FindSet(id)
			a.So(err, should.BeNil)
			a.So(test.SameElements(found, tc.AfterRemove), should.BeTrue)

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

// TestByteListStore executes a black-box test for the given byte list store
func TestByteListStore(t testingT, newStore func() store.ByteListStore) {
	a := assertions.New(t)

	s := newStore()

	id, err := s.CreateList()
	a.So(err, should.BeNil)
	a.So(id, should.NotBeNil)

	idOther, err := s.CreateList()
	a.So(err, should.BeNil)
	a.So(idOther, should.NotBeNil)

	a.So(id, should.NotResemble, idOther)

	// Behavior is implementation-dependent
	a.So(func() { s.FindList(id) }, should.NotPanic)

	err = s.Append(id, []byte("foo"))
	a.So(err, should.BeNil)

	err = s.Append(id)
	a.So(err, should.NotBeNil)

	err = s.Delete(id)
	a.So(err, should.BeNil)

	for i, tc := range []struct {
		Create    [][]byte
		Append    [][]byte
		Trim      int
		AfterTrim [][]byte
	}{
		{
			[][]byte{[]byte("foo"), []byte("bar")},
			[][]byte{[]byte("bar")},
			2,
			[][]byte{[]byte("bar"), []byte("bar")},
		},
		{
			[][]byte{[]byte("foo"), []byte("bar")},
			[][]byte{[]byte("bar"), []byte("bar")},
			0,
			[][]byte{},
		},
		{
			[][]byte{[]byte("42"), []byte("foo"), []byte("bar")},
			[][]byte{[]byte("42"), []byte("foo"), []byte("bar")},
			4,
			[][]byte{[]byte("bar"), []byte("42"), []byte("foo"), []byte("bar")},
		},
		{
			[][]byte{[]byte("42"), []byte("42")},
			[][]byte{[]byte("42")},
			42,
			[][]byte{[]byte("42"), []byte("42"), []byte("42")},
		},
		{
			[][]byte{[]byte("42"), []byte("42")},
			[][]byte{[]byte("42")},
			math.MaxInt32,
			[][]byte{[]byte("42"), []byte("42"), []byte("42")},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			s := newStore()

			id, err := s.CreateList(tc.Create...)
			if !a.So(err, should.BeNil) {
				return
			}
			a.So(id, should.NotBeNil)

			found, err := s.FindList(id)
			a.So(err, should.BeNil)
			a.So(found, should.Resemble, tc.Create)

			err = s.Append(id, tc.Append...)
			a.So(err, should.BeNil)

			found, err = s.FindList(id)
			a.So(err, should.BeNil)
			a.So(found, should.Resemble, append(tc.Create, tc.Append...))

			if v, ok := s.(store.Trimmer); ok {
				err = v.Trim(id, tc.Trim)
				a.So(err, should.NotBeNil)

				found, err = s.FindList(id)
				a.So(err, should.BeNil)
				a.So(found, should.Resemble, tc.AfterTrim)
			}

			err = s.Delete(id)
			a.So(err, should.BeNil)
		})
	}
}

// GenericStore is a TypedStore adapter for an arbitrary store, which executes methods using reflection
// and converts the supplied/returned values to match function signatures.
type GenericStore struct {
	store reflect.Value

	fromIfaceMap func(map[string]interface{}) interface{}
	toIfaceMap   func(interface{}) map[string]interface{}
}

func reflectValueToError(v reflect.Value) error {
	if !v.IsValid() {
		return nil
	}
	iface := v.Interface()
	if iface == nil {
		return nil
	}
	return iface.(error)
}

// Create is a generic Create.
func (gs GenericStore) Create(fields map[string]interface{}) (store.PrimaryKey, error) {
	ret := gs.store.MethodByName("Create").Call([]reflect.Value{
		reflect.ValueOf(gs.fromIfaceMap(fields)),
	})
	if err := reflectValueToError(ret[1]); err != nil {
		return nil, err
	}
	return ret[0].Interface().(store.PrimaryKey), nil
}

// Find is a generic Find.
func (gs GenericStore) Find(id store.PrimaryKey) (map[string]interface{}, error) {
	rv := reflect.Zero(reflect.TypeOf((*store.PrimaryKey)(nil)).Elem())
	if id != nil {
		rv = reflect.ValueOf(id)
	}
	ret := gs.store.MethodByName("Find").Call([]reflect.Value{
		rv,
	})
	if err := reflectValueToError(ret[1]); err != nil {
		return nil, err
	}
	if store.IsNillableKind(ret[0].Kind()) && ret[0].IsNil() {
		return nil, nil
	}
	return gs.toIfaceMap(ret[0].Interface()), nil
}

// FindBy is a generic FindBy.
func (gs GenericStore) FindBy(filter map[string]interface{}) (map[store.PrimaryKey]map[string]interface{}, error) {
	ret := gs.store.MethodByName("FindBy").Call([]reflect.Value{
		reflect.ValueOf(gs.fromIfaceMap(filter)),
	})
	if err := reflectValueToError(ret[1]); err != nil {
		return nil, err
	}
	if store.IsNillableKind(ret[0].Kind()) && ret[0].IsNil() {
		return nil, nil
	}

	m := make(map[store.PrimaryKey]map[string]interface{}, ret[0].Len())
	for _, k := range ret[0].MapKeys() {
		m[k.Interface().(store.PrimaryKey)] = gs.toIfaceMap(ret[0].MapIndex(k).Interface())
	}
	return m, nil
}

// Update is a generic Update.
func (gs GenericStore) Update(id store.PrimaryKey, diff map[string]interface{}) error {
	rv := reflect.Zero(reflect.TypeOf((*store.PrimaryKey)(nil)).Elem())
	if id != nil {
		rv = reflect.ValueOf(id)
	}
	return reflectValueToError(gs.store.MethodByName("Update").Call([]reflect.Value{
		rv,
		reflect.ValueOf(gs.fromIfaceMap(diff)),
	})[0])
}

// Delete is a generic Delete.
func (gs GenericStore) Delete(id store.PrimaryKey) error {
	rv := reflect.Zero(reflect.TypeOf((*store.PrimaryKey)(nil)).Elem())
	if id != nil {
		rv = reflect.ValueOf(id)
	}
	return reflectValueToError(gs.store.MethodByName("Delete").Call([]reflect.Value{
		rv,
	})[0])
}

// NewGenericStore returns a new generic store given a store implementation s (e.g. a ByteStore),
// fromIfaceMap and toIfaceMap convertors.
// The methods of s are executed using reflection and values are converted if necessary.
func NewGenericStore(s interface{}, fromIfaceMap func(map[string]interface{}) interface{}, toIfaceMap func(interface{}) map[string]interface{}) *GenericStore {
	return &GenericStore{
		store:        reflect.ValueOf(s),
		fromIfaceMap: fromIfaceMap,
		toIfaceMap:   toIfaceMap,
	}
}

// TestByteStore executes a black-box test for the given byte store.
func TestByteStore(t testingT, newStore func() store.ByteStore) {
	TestTypedStore(t, func() store.TypedStore {
		return NewGenericStore(newStore(),
			func(m map[string]interface{}) interface{} {
				if m == nil {
					return (map[string][]byte)(nil)
				}

				ret := make(map[string][]byte, len(m))
				for k, v := range m {
					if v == nil {
						ret[k] = nil
						continue
					}

					gob.Register(v)

					var buf bytes.Buffer
					if err := gob.NewEncoder(&buf).Encode(&v); err != nil {
						panic(errors.Errorf("Failed to gob-encode %s value %s to bytes: %s", k, v, err))
					}
					ret[k] = buf.Bytes()
				}
				return ret
			},
			func(v interface{}) map[string]interface{} {
				if v == nil {
					return (map[string]interface{})(nil)
				}

				m := v.(map[string][]byte)

				ret := make(map[string]interface{}, len(m))
				for k, v := range m {
					if len(v) == 0 {
						ret[k] = nil
						continue
					}

					var dv interface{}
					if err := gob.NewDecoder(bytes.NewReader(v)).Decode(&dv); err != nil {
						panic(errors.Errorf("Failed to gob-decode %s value %s to interface: %s", k, v, err))
					}
					ret[k] = dv
				}
				return ret
			})
	})
}
