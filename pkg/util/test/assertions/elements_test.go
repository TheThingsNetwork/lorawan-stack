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
	"reflect"
	"strconv"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	. "go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

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

			a.So(ShouldHaveSameElementsFunc(tc.A, reflect.DeepEqual, tc.B), tc.ShouldFunc)
			a.So(ShouldNotHaveSameElementsFunc(tc.A, reflect.DeepEqual, tc.B), tc.ShouldNotFunc)
			a.So(ShouldHaveSameElementsFunc(tc.A, test.DiffEqual, tc.B), tc.ShouldFunc)
			a.So(ShouldNotHaveSameElementsFunc(tc.A, test.DiffEqual, tc.B), tc.ShouldNotFunc)
			a.So(ShouldHaveSameElementsDeep(tc.A, tc.B), tc.ShouldFunc)
			a.So(ShouldNotHaveSameElementsDeep(tc.A, tc.B), tc.ShouldNotFunc)
			a.So(ShouldHaveSameElementsDiff(tc.A, tc.B), tc.ShouldFunc)
			a.So(ShouldNotHaveSameElementsDiff(tc.A, tc.B), tc.ShouldNotFunc)
		})
	}
}
