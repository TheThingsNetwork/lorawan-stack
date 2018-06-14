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

package errors_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestCause(t *testing.T) {
	a := assertions.New(t)

	var cause = errors.New("cause")

	var errInvalidFoo = errors.DefineInvalidArgument("test_cause_invalid_foo", "Invalid Foo: {foo}", "foo")

	a.So(errInvalidFoo.Cause(), should.EqualErrorOrDefinition, nil)
	a.So(errors.RootCause(errInvalidFoo), should.EqualErrorOrDefinition, errInvalidFoo)

	err1 := errInvalidFoo.WithCause(cause)

	a.So(func() {
		err1.WithCause(cause)
	}, should.Panic)

	a.So(err1, should.HaveSameErrorDefinitionAs, errInvalidFoo)
	a.So(err1.Cause(), should.EqualErrorOrDefinition, &cause)
	a.So(errors.RootCause(err1), should.EqualErrorOrDefinition, &cause)
	a.So(errors.Stack(err1), should.Resemble, []error{err1, &cause})

	var errInvalidBar = errors.DefineInvalidArgument("test_cause_invalid_bar", "Invalid Bar")
	err2 := errInvalidBar.WithCause(&err1)

	a.So(err2, should.HaveSameErrorDefinitionAs, errInvalidBar)
	a.So(err2.Cause(), should.EqualErrorOrDefinition, &err1)
	a.So(errors.RootCause(err2), should.EqualErrorOrDefinition, &cause)
	a.So(errors.Stack(err2), should.Resemble, []error{err2, &err1, &cause})
}
