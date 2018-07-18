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

package store_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/kr/pretty"
	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/store"
	"go.thethings.network/lorawan-stack/pkg/store/mapstore"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
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

var firstVal = First{
	A: 42,
	B: 42,
	C: 42,
	D: "42",
	E: map[string]interface{}{
		"42": "42",
	},
	F: map[string][]byte{
		"42": []byte("42"),
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
		"first": {
			firstVal,
		},
	},
	SliceMapPtr: map[string][]*First{
		"first": {
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
					"42": "42",
				},
				F: map[string][]byte{
					"42": []byte("42"),
				},
				G: []int{41, 43},
			},
			[]string{"A", "B", "C", "D", "G"},
		},
		{
			&secondVal,
			&Second{
				Slice: []First{
					{A: 42},
					{A: 42},
					firstVal,
				},
				SlicePtr: []*First{
					&firstVal,
				},
				SliceMap:    nil,
				SliceMapPtr: nil,
			},
			&Second{
				Slice: []First{
					{A: 42},
					{A: 42},
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
				SliceMap: nil,
				SliceMapPtr: map[string][]*First{
					"first": {
						&firstVal,
					},
				},
			},
			[]string{"Slice", "SlicePtr", "SliceMap"},
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)
			cl := NewTypedMapStoreClient(mapstore.New())
			if !a.So(cl, should.NotBeNil) {
				return
			}

			var newResult NewResultFunc = func() interface{} {
				return reflect.New(reflect.Indirect(reflect.ValueOf(tc.Stored)).Type()).Interface()
			}

			key, err := cl.Create(tc.Stored)
			if !a.So(err, should.BeNil) || !a.So(key, should.NotBeNil) {
				return
			}

			v := newResult()
			err = cl.Find(key, v)
			a.So(err, should.BeNil)
			if !a.So(pretty.Diff(v, tc.Stored), should.BeEmpty) {
				t.FailNow()
			}

			i := 0
			total, err := cl.Range(v, newResult, "", 1, 0, func(k PrimaryKey, v interface{}) bool {
				i++
				a.So(k, should.Resemble, key)
				a.So(pretty.Diff(v, tc.Stored), should.BeEmpty)
				return true
			})
			a.So(err, should.BeNil)
			a.So(total, should.Equal, 1)
			a.So(i, should.Equal, 1)

			err = cl.Update(key, tc.Updated, tc.Fields...)
			a.So(err, should.BeNil)

			v = newResult()
			err = cl.Find(key, v)
			a.So(err, should.BeNil)
			if !a.So(v, should.Resemble, tc.AfterUpdate) {
				pretty.Ldiff(t, v, tc.AfterUpdate)
				return
			}

			i = 0
			total, err = cl.Range(v, newResult, "", 1, 0, func(k PrimaryKey, v interface{}) bool {
				i++
				a.So(k, should.Resemble, key)
				a.So(pretty.Diff(v, tc.AfterUpdate), should.BeEmpty)
				return true
			})
			a.So(err, should.BeNil)
			a.So(total, should.Equal, 1)
			a.So(i, should.Equal, 1)

			err = cl.Delete(key)
			a.So(err, should.BeNil)

			v = newResult()
			err = cl.Find(key, v)
			a.So(err, should.BeNil)

			i = 0
			total, err = cl.Range(v, newResult, "", 1, 0, func(k PrimaryKey, v interface{}) bool { i++; return true })
			a.So(err, should.BeNil)
			a.So(total, should.Equal, 0)
			a.So(i, should.Equal, 0)
		})
	}
}
