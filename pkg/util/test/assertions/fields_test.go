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
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

type assertion func(actual interface{}, expected ...interface{}) string

func TestShouldEqualFields(t *testing.T) {
	for _, tc := range []struct {
		Name          string
		A             interface{}
		B             interface{}
		MakeAssertion func() assertion
	}{
		{
			"Equal",
			struct {
				Foo int
			}{
				42,
			},
			struct {
				Foo int
			}{
				42,
			},
			func() assertion { return ShouldEqualFields },
		},
		{
			"NotEqual",
			struct {
				Foo int
				Bar struct{ Desc, Lang string }
			}{
				42,
				struct{ Desc, Lang string }{"hallo", "nl"},
			},
			struct {
				Foo int
			}{
				42,
			},
			func() assertion { return ShouldNotEqualFields },
		},
		{
			"EqualWithIgnore",
			struct {
				Foo int
				Bar struct{ Desc, Lang string }
			}{
				42,
				struct{ Desc, Lang string }{"hallo", "nl"},
			},
			struct {
				Foo int
				Bar struct{ Desc, Lang string }
			}{
				42,
				struct{ Desc, Lang string }{"hallo", "de"},
			},
			func() assertion { return ShouldEqualFieldsWithIgnores("Bar.Lang") },
		},
		{
			"NotEqualWithIgnore",
			struct {
				Foo int
				Bar struct{ Desc, Lang string }
			}{
				42,
				struct{ Desc, Lang string }{"hallo", "nl"},
			},
			struct {
				Foo int
				Bar struct{ Desc, Lang string }
			}{
				42,
				struct{ Desc, Lang string }{"bonjour", "fr"},
			},
			func() assertion { return ShouldNotEqualFieldsWithIgnores("Bar.Lang") },
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			a.So(tc.MakeAssertion()(tc.A, tc.B), should.BeEmpty)
		})
	}
}
