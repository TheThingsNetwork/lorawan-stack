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

package assertions

import (
	"fmt"
	"io"
	"reflect"
	"strconv"
	"sync"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestSameElements(t *testing.T) {
	for _, tc := range []struct {
		A    interface{}
		B    interface{}
		Same bool
	}{
		{
			[][]byte{{42}, {43}},
			[][]byte{{43}, {44}},
			false,
		},
		{
			[][]byte{{43}, {43}},
			[][]byte{{43}, {44}},
			false,
		},
		{
			[][]byte{{43}, {43}, {43}},
			[][]byte{{43}, {44}},
			false,
		},
		{
			[][]byte{{42}, {43}, {43}},
			[][]byte{{43}, {42}, {43}},
			true,
		},
		{
			[][]byte{},
			[][]byte{{43}, {42}, {43}},
			false,
		},
		{
			[][]byte{{43}, {42}, {43}},
			[][]byte{},
			false,
		},
		{
			[]string{"a", "b"},
			[][]byte{{'a'}, {'b'}},
			false,
		},
		{
			map[string]interface{}{"a": 42, "b": 77},
			map[string]int{"a": 42, "b": 77},
			true,
		},
		{
			map[string]io.Writer{},
			[0]int{},
			true,
		},
		{
			func() *sync.Map { m := &sync.Map{}; m.Store("42", 42); m.Store("77", "b"); return m }(),
			map[string]interface{}{"42": 42, "77": "b"},
			true,
		},
		{
			func() *sync.Map { m := &sync.Map{}; m.Store("42", 42); m.Store("77", "b"); return m }(),
			map[string]interface{}{"42": 42.2, "77": "b"},
			false,
		},
		{
			[]int{42},
			[]int{42},
			true,
		},
		{
			[]byte("ttn"),
			"ttn",
			true,
		},
		{
			[]byte("foo"),
			"bar",
			false,
		},
		{
			[2]int{42, 43},
			[]int{43, 42},
			true,
		},
		{
			[3]int{42, 43, 43},
			[]int{43, 42},
			false,
		},
		{
			"hello",
			"olleh",
			true,
		},
		{
			"foo",
			"fof",
			false,
		},
	} {
		t.Run(fmt.Sprintf("%v/%v", tc.A, tc.B), func(t *testing.T) {
			a := assertions.New(t)

			a.So(sameElements(reflect.DeepEqual, tc.A, tc.B), should.Equal, tc.Same)
			a.So(sameElements(diffEqual, tc.A, tc.B), should.Equal, tc.Same)
		})
	}
}

func TestShouldHaveSameElements(t *testing.T) {
	for i, tc := range []struct {
		A             interface{}
		B             interface{}
		ShouldFunc    func(actual interface{}, expected ...interface{}) string
		ShouldNotFunc func(actual interface{}, expected ...interface{}) string
	}{
		{
			[][]byte{{42}, {43}},
			[][]byte{{43}, {44}},
			should.NotBeEmpty,
			should.BeEmpty,
		},
		{
			[][]byte{{43}, {43}},
			[][]byte{{43}, {44}},
			should.NotBeEmpty,
			should.BeEmpty,
		},
		{
			[][]byte{{43}, {43}, {43}},
			[][]byte{{43}, {44}},
			should.NotBeEmpty,
			should.BeEmpty,
		},
		{
			[][]byte{{42}, {43}, {43}},
			[][]byte{{43}, {42}, {43}},
			should.BeEmpty,
			should.NotBeEmpty,
		},
		{
			[][]byte{},
			[][]byte{{43}, {42}, {43}},
			should.NotBeEmpty,
			should.BeEmpty,
		},
		{
			[][]byte{{43}, {42}, {43}},
			[][]byte{},
			should.NotBeEmpty,
			should.BeEmpty,
		},
		{
			[]int{42},
			[]int{42},
			should.BeEmpty,
			should.NotBeEmpty,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)

			a.So(ShouldHaveSameElementsFunc(tc.A, tc.B, reflect.DeepEqual), tc.ShouldFunc)
			a.So(ShouldNotHaveSameElementsFunc(tc.A, tc.B, reflect.DeepEqual), tc.ShouldNotFunc)
			a.So(ShouldHaveSameElementsFunc(tc.A, tc.B, diffEqual), tc.ShouldFunc)
			a.So(ShouldNotHaveSameElementsFunc(tc.A, tc.B, diffEqual), tc.ShouldNotFunc)
			a.So(ShouldHaveSameElementsDeep(tc.A, tc.B), tc.ShouldFunc)
			a.So(ShouldNotHaveSameElementsDeep(tc.A, tc.B), tc.ShouldNotFunc)
			a.So(ShouldHaveSameElementsDiff(tc.A, tc.B), tc.ShouldFunc)
			a.So(ShouldNotHaveSameElementsDiff(tc.A, tc.B), tc.ShouldNotFunc)
		})
	}
}
