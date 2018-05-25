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

package db

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestSliceOfInt32s(t *testing.T) {
	a := assertions.New(t)

	// Empty int32 slice.
	{
		s := Int32Slice{}
		value, err := s.Value()
		a.So(err, should.BeNil)
		a.So(value.(string), should.Equal, "[]")
		a.So(s.Scan("[]"), should.BeNil)
		a.So(s, should.Resemble, s)

	}

	// Filled int32 slice.
	{
		s := Int32Slice{3, 3, 3}
		value, err := s.Value()
		a.So(err, should.BeNil)
		a.So(value.(string), should.Equal, "[3,3,3]")
		a.So(s.Scan("[3,3,3]"), should.BeNil)
		a.So(s, should.Resemble, s)
	}

	type Foo int32

	// Filled int32-like slice.
	{
		s, err := NewInt32Slice([]Foo{1, 2, 3})
		a.So(err, should.BeNil)
		value, err := s.Value()
		a.So(err, should.BeNil)
		a.So(value.(string), should.Equal, "[1,2,3]")
		a.So(s.Scan("[1,2,3]"), should.BeNil)
		a.So(s, should.Resemble, s)

		dest := make([]Foo, 0)
		a.So(s.SetInto(&dest), should.BeNil)
		a.So(dest, should.Resemble, []Foo{1, 2, 3})
	}
}

func TestSliceOfStrings(t *testing.T) {
	a := assertions.New(t)

	// Empty string slice.
	{
		s := StringSlice{}
		value, err := s.Value()
		a.So(err, should.BeNil)
		a.So(value.(string), should.Equal, "[]")
		a.So(s.Scan("[]"), should.BeNil)
		a.So(s, should.Resemble, s)

	}

	// Filled string slice.
	{
		s := StringSlice{"a", "b", "c"}
		value, err := s.Value()
		a.So(err, should.BeNil)
		a.So(value.(string), should.Equal, `["a","b","c"]`)
		a.So(s.Scan(`["a"]`), should.BeNil)
		a.So(s, should.Resemble, StringSlice{"a"})
	}

	type FooStr string

	// Filled string-like slice.
	{
		s, err := NewStringSlice([]FooStr{"bar"})
		a.So(err, should.BeNil)
		value, err := s.Value()
		a.So(err, should.BeNil)
		a.So(value.(string), should.Equal, `["bar"]`)
		a.So(s.Scan(`["qux"]`), should.BeNil)
		a.So(s, should.Resemble, StringSlice{"qux"})

		dest := make([]FooStr, 0)
		a.So(s.SetInto(&dest), should.BeNil)
		a.So(dest, should.Resemble, []FooStr{"qux"})
	}
}
