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

func TestCause(t *testing.T) {
	a := assertions.New(t)

	var cause = New("cause")

	var errInvalidFoo = DefineInvalidArgument("test_cause_invalid_foo", "Invalid Foo: {foo}", "foo")

	a.So(errInvalidFoo.Cause(), ShouldEqual, nil)
	a.So(RootCause(errInvalidFoo), ShouldEqual, errInvalidFoo)

	err1 := errInvalidFoo.WithCause(cause)

	a.So(func() {
		err1.WithCause(cause)
	}, should.Panic)

	a.So(err1, ShouldHaveSameDefinitionAs, errInvalidFoo)
	a.So(err1.Cause(), ShouldEqual, &cause)
	a.So(RootCause(err1), ShouldEqual, &cause)
	a.So(Stack(err1), should.Resemble, []error{err1, &cause})

	var errInvalidBar = DefineInvalidArgument("test_cause_invalid_bar", "Invalid Bar")
	err2 := errInvalidBar.WithCause(&err1)

	a.So(err2, ShouldHaveSameDefinitionAs, errInvalidBar)
	a.So(err2.Cause(), ShouldEqual, &err1)
	a.So(RootCause(err2), ShouldEqual, &cause)
	a.So(Stack(err2), should.Resemble, []error{err2, &err1, &cause})
}
