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

package errors

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestFormatTypes(t *testing.T) {
	a := assertions.New(t)

	format := "{foo} - {bar} - {nil} - {list} - {map} - {complex} - {ptr}"
	{
		val := 10
		res := Format(format, Attributes{
			"foo":     10,
			"bar":     "bar",
			"nil":     nil,
			"list":    []int{1, 2, 3},
			"map":     map[string]int{"ok": 1},
			"complex": 3 + 4i,
			"ptr":     &val,
		})
		a.So(res, should.Equal, "10 - bar - <nil> - [1 2 3] - map[ok:1] - (3+4i) - 10")
	}
}

func TestFormat(t *testing.T) {
	a := assertions.New(t)

	format := "Found {foo, plural, =0 {no foos} =1 {# foo} other {# foos}}"
	{
		res := Format(format, Attributes{
			"foo": 1,
		})
		a.So(res, should.Equal, "Found 1 foo")
	}
	{
		res := Format(format, Attributes{
			"foo": 0,
		})
		a.So(res, should.Equal, "Found no foos")
	}
}

type errorsCounter uint8

func (c *errorsCounter) Errorf(_ string, _ ...interface{}) { *c = *c + 1 }

func (c *errorsCounter) WithError(_ error, msg string) { c.Errorf(msg) }

func TestAttributeType(t *testing.T) {
	a := assertions.New(t)

	counter := errorsCounter(0)
	oldErrorSignaler := FormatErrorSignaler
	FormatErrorSignaler = &counter

	format := "Found {foo} results"

	// Non-primitive types
	{
		initCounter := counter
		Format(format, Attributes{
			"foo": struct {
				Number int
			}{3},
		})
		a.So(counter, should.Equal, initCounter+1)
	}

	// Primitive types
	{
		initCounter := counter
		res := Format(format, Attributes{
			"foo": 0,
		})
		a.So(res, should.Equal, "Found 0 results")
		a.So(counter, should.Equal, initCounter)
	}

	FormatErrorSignaler = oldErrorSignaler
}
