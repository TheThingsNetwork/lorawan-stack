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
	"reflect"
	"strconv"
	"testing"

	ssassertions "github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/util/test"
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
			a := ssassertions.New(t)

			a.So(ShouldHaveSameElements(tc.A, tc.B, reflect.DeepEqual), tc.ShouldFunc)
			a.So(ShouldNotHaveSameElements(tc.A, tc.B, reflect.DeepEqual), tc.ShouldNotFunc)
			a.So(ShouldHaveSameElements(tc.A, tc.B, test.DiffEqual), tc.ShouldFunc)
			a.So(ShouldNotHaveSameElements(tc.A, tc.B, test.DiffEqual), tc.ShouldNotFunc)
			a.So(ShouldHaveSameElementsDeep(tc.A, tc.B), tc.ShouldFunc)
			a.So(ShouldNotHaveSameElementsDeep(tc.A, tc.B), tc.ShouldNotFunc)
			a.So(ShouldHaveSameElementsDiff(tc.A, tc.B), tc.ShouldFunc)
			a.So(ShouldNotHaveSameElementsDiff(tc.A, tc.B), tc.ShouldNotFunc)
		})
	}
}
