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

package errors_test

import (
	gerrors "errors"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

type detail string

func (s *detail) Reset()        { *s = "" }
func (s detail) String() string { return string(s) }
func (detail) ProtoMessage()    {}

func stringDetail(s string) *detail {
	return (*detail)(&s)
}

func TestDetails(t *testing.T) {
	a := assertions.New(t)

	var errInvalidFoo = errors.DefineInvalidArgument("test_details_invalid_foo", "Invalid Foo")

	a.So(errInvalidFoo.Details(), should.BeEmpty)

	a.So(errors.Details(errInvalidFoo), should.BeEmpty)
	a.So(errors.Details(errInvalidFoo.GRPCStatus().Err()), should.BeEmpty)
	a.So(errors.Details(gerrors.New("go stdlib error")), should.BeEmpty)

	err1 := errInvalidFoo.WithDetails(stringDetail("foo"), stringDetail("bar"))
	err2 := err1.WithDetails(stringDetail("baz"))

	a.So(err1, should.HaveSameErrorDefinitionAs, errInvalidFoo)
	a.So(err1.Details(), should.Resemble, []proto.Message{stringDetail("foo"), stringDetail("bar")})

	a.So(err2, should.HaveSameErrorDefinitionAs, err1)
	a.So(err2, should.HaveSameErrorDefinitionAs, errInvalidFoo)
	a.So(err2.Details(), should.Resemble, []proto.Message{stringDetail("foo"), stringDetail("bar"), stringDetail("baz")})
}
