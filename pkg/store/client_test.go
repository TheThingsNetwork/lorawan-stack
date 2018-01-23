// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package store_test

import (
	"reflect"
	"strconv"
	"testing"

	. "github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/TheThingsNetwork/ttn/pkg/store/mapstore"
	"github.com/kr/pretty"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

type First struct {
	A int
	B uint
	C float64
	D string
	E map[string]interface{}
	F map[string][]byte
	G []int
	H []struct {
		Foo struct{ Bar string }
		Int int
	}
}

type Second struct {
	Slice       []First
	SlicePtr    []*First
	Map         map[string]First
	MapPtr      map[string]*First
	SliceMap    map[string][]First
	SliceMapPtr map[string][]*First
}

type Third struct {
	Slice       []Second
	SlicePtr    []*Second
	Map         map[string]Second
	MapPtr      map[string]*Second
	SliceMap    map[string][]Second
	SliceMapPtr map[string][]*Second
}

var firstVal = First{
	A: 42,
	B: 42,
	C: 42,
	D: "42",
	E: map[string]interface{}{
		"foo": "bar",
		"42":  "42",
	},
	F: map[string][]byte{
		"foo": []byte("bar"),
		"42":  []byte("42"),
	},
	G: []int{42, 42, 42},
}

var secondVal = Second{
	Slice: []First{
		firstVal,
	},
	SlicePtr: []*First{
		&firstVal,
	},
	Map: map[string]First{
		"first": firstVal,
	},
	MapPtr: map[string]*First{
		"first": &firstVal,
	},
	SliceMap: map[string][]First{
		"first": []First{
			firstVal,
		},
	},
	SliceMapPtr: map[string][]*First{
		"first": []*First{
			&firstVal,
		},
	},
}

func TestTypedClient(t *testing.T) {
	for i, tc := range []struct {
		Stored      interface{}
		Updated     interface{}
		AfterUpdate interface{}
		Fields      []string
	}{
		{
			&firstVal,
			&First{
				A: 43,
				B: 43,
				C: 43,
				D: "43",
				E: map[string]interface{}{
					"hey": "there",
				},
				G: []int{41, 43},
			},
			&First{
				A: 43,
				B: 43,
				C: 43,
				D: "43",
				E: map[string]interface{}{
					"foo": "bar",
					"42":  "42",
				},
				F: map[string][]byte{
					"foo": []byte("bar"),
					"42":  []byte("42"),
				},
				G: []int{41, 43, 42},
			},
			[]string{"A", "B", "C", "D", "G.0", "G.1"},
		},
		{
			&secondVal,
			&Second{
				Slice: []First{
					First{A: 42},
					First{A: 42},
					firstVal,
				},
				SlicePtr: []*First{
					nil,
					&firstVal,
				},
				SliceMap:    nil,
				SliceMapPtr: nil,
			},
			&Second{
				Slice: []First{
					First{A: 42},
					First{A: 42},
					firstVal,
				},
				SlicePtr: []*First{
					&firstVal,
					&firstVal,
				},
				Map: map[string]First{
					"first": firstVal,
				},
				MapPtr: map[string]*First{
					"first": &firstVal,
				},
				SliceMap: nil,
				SliceMapPtr: map[string][]*First{
					"first": []*First{
						&firstVal,
					},
				},
			},
			[]string{"Slice", "SlicePtr.1", "SliceMap"},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)
			cl := NewTypedStoreClient(mapstore.New())
			if !a.So(cl, should.NotBeNil) {
				return
			}

			var newResult NewResultFunc = func() interface{} {
				return reflect.New(reflect.Indirect(reflect.ValueOf(tc.Stored)).Type()).Interface()
			}

			k, err := cl.Create(tc.Stored)
			if !a.So(err, should.BeNil) || !a.So(k, should.NotBeNil) {
				return
			}

			v := newResult()
			err = cl.Find(k, v)
			a.So(err, should.BeNil)
			if !a.So(pretty.Diff(v, tc.Stored), should.BeNil) {
				return
			}

			m, err := cl.FindBy(v, newResult)
			if a.So(err, should.BeNil) {
				for mk, mv := range m {
					a.So(mk, should.Resemble, k)
					a.So(pretty.Diff(mv, tc.Stored), should.BeNil)
				}
			}

			err = cl.Update(k, tc.Updated, tc.Fields...)
			a.So(err, should.BeNil)

			v = newResult()
			err = cl.Find(k, v)
			a.So(err, should.BeNil)
			if !a.So(pretty.Diff(v, tc.AfterUpdate), should.BeNil) {
				pretty.Println(v)
				return
			}

			m, err = cl.FindBy(v, newResult)
			if a.So(err, should.BeNil) {
				for mk, mv := range m {
					a.So(mk, should.Resemble, k)
					a.So(pretty.Diff(mv, tc.AfterUpdate), should.BeNil)
				}
			}

			err = cl.Delete(k)
			a.So(err, should.BeNil)

			v = newResult()
			err = cl.Find(k, v)
			a.So(err, should.NotBeNil)

			m, err = cl.FindBy(tc.AfterUpdate, newResult)
			a.So(err, should.BeNil)
		})
	}
}
