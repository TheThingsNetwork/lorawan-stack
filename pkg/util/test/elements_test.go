// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package test_test

import (
	"fmt"
	"io"
	"reflect"
	"sync"
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestSetRelations(t *testing.T) {
	for _, tc := range []struct {
		A            any
		B            any
		AIsSubsetOfB bool
		BIsSubsetOfA bool
	}{
		{
			AIsSubsetOfB: true,
			BIsSubsetOfA: true,
		},
		{
			A: [][]byte{{42}, {43}},

			BIsSubsetOfA: true,
		},
		{
			A: [][]byte{{42}, {43}},
			B: [][]byte{{43}, {44}},
		},
		{
			A: [][]byte{{43}, {43}},
			B: [][]byte{{43}, {44}},
		},
		{
			A: [][]byte{{43}, {43}, {43}},
			B: [][]byte{{43}, {44}},
		},
		{
			A: [][]byte{{42}, {43}, {43}},
			B: [][]byte{{43}, {42}, {43}},

			AIsSubsetOfB: true,
			BIsSubsetOfA: true,
		},
		{
			A: [][]byte{},
			B: [][]byte{{43}, {42}, {43}},

			AIsSubsetOfB: true,
		},
		{
			A: [][]byte{{43}, {42}, {43}},
			B: [][]byte{},

			BIsSubsetOfA: true,
		},
		{
			A: []string{"a", "b"},
			B: [][]byte{{'a'}, {'b'}},
		},
		{
			A: map[string]any{"a": 42, "d": 77},
			B: map[string]int{"a": 42, "b": 77},

			AIsSubsetOfB: true,
			BIsSubsetOfA: true,
		},
		{
			A: map[string]any{"a": 42, "d": 77},
			B: []int{42, 77},

			AIsSubsetOfB: true,
			BIsSubsetOfA: true,
		},
		{
			A: map[string]io.Writer{},
			B: [0]int{},

			AIsSubsetOfB: true,
			BIsSubsetOfA: true,
		},
		{
			A: func() *sync.Map { m := &sync.Map{}; m.Store("42", 42); m.Store("77", "b"); return m }(),
			B: map[string]any{"42": 42, "77": "b"},

			AIsSubsetOfB: true,
			BIsSubsetOfA: true,
		},
		{
			A: func() *sync.Map { m := &sync.Map{}; m.Store("42", 42); m.Store("77", "b"); return m }(),
			B: map[string]any{"42": 42.2, "77": "b"},
		},
		{
			A: []int{42},
			B: []int{42},

			AIsSubsetOfB: true,
			BIsSubsetOfA: true,
		},
		{
			A: []byte("ttn"),
			B: "ttn",

			AIsSubsetOfB: true,
			BIsSubsetOfA: true,
		},
		{
			A: []byte("foo"),
			B: "bar",
		},
		{
			A: [2]int{42, 43},
			B: []int{43, 42},

			AIsSubsetOfB: true,
			BIsSubsetOfA: true,
		},
		{
			A: [3]int{42, 43, 43},
			B: []int{43, 42},

			BIsSubsetOfA: true,
		},
		{
			A: "hello",
			B: "olleh",

			AIsSubsetOfB: true,
			BIsSubsetOfA: true,
		},
		{
			A: "foo",
			B: "fof",
		},
	} {
		t.Run(fmt.Sprintf("%v/%v", tc.A, tc.B), func(t *testing.T) {
			for _, eq := range []any{
				reflect.DeepEqual,
				DiffEqual,
			} {
				t.Run("IsSubsetOfElements", func(t *testing.T) {
					a := assertions.New(t)
					a.So(IsSubsetOfElements(eq, tc.A, tc.B), should.Equal, tc.AIsSubsetOfB)
					a.So(IsSubsetOfElements(eq, tc.B, tc.A), should.Equal, tc.BIsSubsetOfA)
				})
				t.Run("IsProperSubsetOfElements", func(t *testing.T) {
					a := assertions.New(t)
					a.So(IsProperSubsetOfElements(eq, tc.A, tc.B), should.Equal, tc.AIsSubsetOfB && !tc.BIsSubsetOfA)
					a.So(IsProperSubsetOfElements(eq, tc.B, tc.A), should.Equal, tc.BIsSubsetOfA && !tc.AIsSubsetOfB)
				})
				t.Run("SameElements", func(t *testing.T) {
					a := assertions.New(t)
					a.So(SameElements(eq, tc.A, tc.B), should.Equal, tc.AIsSubsetOfB && tc.BIsSubsetOfA)
				})
			}
		})
	}
}
