// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package applicationserver

import (
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestValidateStruct(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string

		st       map[string]any
		expected []ValueViolation
	}{
		{
			name: "empty",
		},
		{
			name: "simple object",

			st: map[string]any{
				"foo": 123,
			},
		},
		{
			name: "top level NaN",
			st: map[string]any{
				"foo": "NaN",
			},
			expected: []ValueViolation{
				{
					Type:    ValueTypeNaN,
					Context: ValueContextStruct,
				},
			},
		},
		{
			name: "nested object Infinity",

			st: map[string]any{
				"foo": map[string]any{
					"bar": "Infinity",
				},
			},
			expected: []ValueViolation{
				{
					Type:    ValueTypePosInf,
					Context: ValueContextStruct,
				},
			},
		},
		{
			name: "nested object -Infinity",
			st: map[string]any{
				"foo": []any{"-Infinity"},
			},
			expected: []ValueViolation{
				{
					Type:    ValueTypeNegInf,
					Context: ValueContextList,
				},
			},
		},
		{
			name: "nested object NaN",
			st: map[string]any{
				"foo": []any{
					map[string]any{
						"bar": "NaN",
					},
				},
			},
			expected: []ValueViolation{
				{
					Type:    ValueTypeNaN,
					Context: ValueContextStruct,
				},
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			a := assertions.New(t)

			st, err := structpb.NewStruct(tc.st)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}

			warnings := FindViolations(st)
			a.So(warnings, should.Resemble, tc.expected)
		})
	}
}
