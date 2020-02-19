// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package assertions_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	. "go.thethings.network/lorawan-stack/pkg/util/test/assertions"
)

func TestHaveEmptyDiff(t *testing.T) {
	for _, tc := range []struct {
		A                interface{}
		B                interface{}
		Assertion        func(interface{}, ...interface{}) string
		InverseAssertion func(interface{}, ...interface{}) string
	}{
		{
			A:                "test",
			B:                "test",
			Assertion:        should.BeEmpty,
			InverseAssertion: should.NotBeEmpty,
		},
		{
			A:                "test",
			B:                "test1",
			Assertion:        should.NotBeEmpty,
			InverseAssertion: should.BeEmpty,
		},
		{
			A:                42,
			B:                42,
			Assertion:        should.BeEmpty,
			InverseAssertion: should.NotBeEmpty,
		},
		{
			A:                1,
			B:                2,
			Assertion:        should.NotBeEmpty,
			InverseAssertion: should.BeEmpty,
		},
		{
			A: struct {
				Foo int
				Bar int
			}{42, 43},
			B: struct {
				Foo int
				Bar int
			}{42, 43},
			Assertion:        should.BeEmpty,
			InverseAssertion: should.NotBeEmpty,
		},
		{
			A:                nil,
			B:                0,
			Assertion:        should.NotBeEmpty,
			InverseAssertion: should.BeEmpty,
		},
		{
			A:                nil,
			B:                "test",
			Assertion:        should.NotBeEmpty,
			InverseAssertion: should.BeEmpty,
		},
		{
			A:                []string{},
			B:                []string(nil),
			Assertion:        should.BeEmpty,
			InverseAssertion: should.NotBeEmpty,
		},
		{
			A:                map[int]int{},
			B:                map[int]int(nil),
			Assertion:        should.BeEmpty,
			InverseAssertion: should.NotBeEmpty,
		},
	} {
		t.Run(fmt.Sprintf("%v/%v", tc.A, tc.B), func(t *testing.T) {
			a := assertions.New(t)
			a.So(ShouldHaveEmptyDiff(tc.A, tc.B), tc.Assertion)
			a.So(ShouldNotHaveEmptyDiff(tc.A, tc.B), tc.InverseAssertion)
		})
	}
}

type testSetFielder struct {
	A int
	B string
	C []string
}

func (dst *testSetFielder) SetFields(src *testSetFielder, paths ...string) error {
	for _, p := range paths {
		switch p {
		case "a":
			dst.A = src.A
		case "b":
			dst.B = src.B
		case "c":
			dst.C = src.C
		default:
			return fmt.Errorf("invalid path '%s'", p)
		}
	}
	return nil
}

func TestShouldResembleFields(t *testing.T) {
	for _, tc := range []struct {
		A                interface{}
		B                interface{}
		Paths            []interface{}
		Assertion        func(interface{}, ...interface{}) string
		InverseAssertion func(interface{}, ...interface{}) string
	}{
		{
			A:         &testSetFielder{},
			B:         &testSetFielder{},
			Assertion: should.BeEmpty,
		},
		{
			A: &testSetFielder{
				A: 42,
			},
			B:         &testSetFielder{},
			Assertion: should.NotBeEmpty,
		},
		{
			A: &testSetFielder{
				A: 42,
			},
			B: &testSetFielder{
				A: 42,
			},
			Assertion: should.BeEmpty,
		},
		{
			A: &testSetFielder{
				A: 42,
			},
			B: &testSetFielder{
				A: 42,
			},
			Paths:     []interface{}{"a"},
			Assertion: should.BeEmpty,
		},
		{
			A: &testSetFielder{
				A: 42,
			},
			B: &testSetFielder{
				A: 42,
			},
			Paths:     []interface{}{[]string{"a"}},
			Assertion: should.BeEmpty,
		},
		{
			A: &testSetFielder{
				A: 42,
			},
			B: &testSetFielder{
				A: 42,
			},
			Paths:     []interface{}{[1]string{"a"}},
			Assertion: should.BeEmpty,
		},
		{
			A: &testSetFielder{
				A: 42,
			},
			B: &testSetFielder{
				A: 42,
			},
			Paths:     []interface{}{"b"},
			Assertion: should.BeEmpty,
		},
		{
			A: &testSetFielder{
				A: 42,
			},
			B: &testSetFielder{
				A: 42,
			},
			Paths:     []interface{}{"a", "b"},
			Assertion: should.BeEmpty,
		},
		{
			A: &testSetFielder{
				A: 42,
			},
			B: &testSetFielder{
				A: 43,
			},
			Paths:     []interface{}{[]string{"b"}, "c"},
			Assertion: should.BeEmpty,
		},
		{
			A: &testSetFielder{
				A: 42,
			},
			B: &testSetFielder{
				A: 43,
			},
			Paths:     []interface{}{"a", "b"},
			Assertion: should.NotBeEmpty,
		},
		{
			A: &testSetFielder{
				A: 42,
			},
			B: &testSetFielder{
				A: 43,
			},
			Paths:     []interface{}{[2]string{"a", "a"}, []string{"b"}, "c"},
			Assertion: should.NotBeEmpty,
		},
	} {
		t.Run(fmt.Sprintf("%+v/%+v/%+v", tc.A, tc.B, tc.Paths), func(t *testing.T) {
			a := assertions.New(t)
			a.So(ShouldResembleFields(tc.A, append([]interface{}{tc.B}, tc.Paths...)...), tc.Assertion)
			if reflect.DeepEqual(tc.A, tc.B) {
				a.So(ShouldResembleFields(tc.A, tc.B), should.BeEmpty)
			} else {
				a.So(ShouldResembleFields(tc.A, tc.B), should.NotBeEmpty)
			}
		})
	}
}
