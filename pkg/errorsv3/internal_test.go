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

func TestMessageFormat(t *testing.T) {
	a := assertions.New(t)

	args := messageFormatArguments("Application with ID {app_id} could not be found in namespace { ns } or namespace {   ns } does not exist")
	a.So(args, should.HaveLength, 2) // no duplicates
	a.So(args, should.Contain, "app_id")
	a.So(args, should.Contain, "ns")

	err := Define("test_message_format", "MessageFormat {foo}, {bar}")
	a.So(err.publicAttributes, should.Contain, "foo")
	a.So(err.publicAttributes, should.Contain, "bar")
}
