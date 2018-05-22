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

func TestAttributes(t *testing.T) {
	a := assertions.New(t)

	var errInvalidFoo = DefineInvalidArgument("test_attributes_invalid_foo", "Invalid Foo: {foo}", "foo")

	a.So(errInvalidFoo.Attributes(), should.BeEmpty)

	a.So(func() { errInvalidFoo.WithAttributes("only_key") }, should.Panic)

	err1 := errInvalidFoo.WithAttributes("foo", "bar")
	err2 := err1.WithAttributes("bar", "baz")

	a.So(err1, ShouldHaveSameDefinitionAs, errInvalidFoo)
	a.So(err1.Attributes(), should.Resemble, map[string]interface{}{"foo": "bar"})

	a.So(err2, ShouldHaveSameDefinitionAs, err1)
	a.So(err2, ShouldHaveSameDefinitionAs, errInvalidFoo)
	a.So(err2.Attributes(), should.Resemble, map[string]interface{}{"foo": "bar", "bar": "baz"})
}
