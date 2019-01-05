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

package assertions

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

func TestShouldHaveSameErrorDefinition(t *testing.T) {
	a := assertions.New(t)

	errDef := errors.Define("test_error_assertions", "Error Assertions Test")
	errOtherDef := errors.Define("test_error_assertions_other", "Other Error Assertions Test")

	// Happy flow.
	a.So(ShouldHaveSameErrorDefinitionAs(errDef.WithAttributes("k", "v"), errDef.WithAttributes("foo", "bar")), should.BeEmpty)
	a.So(ShouldEqualErrorOrDefinition(errDef.WithAttributes("k", "v"), errDef.WithAttributes("k", "v")), should.BeEmpty)
	a.So(ShouldEqualErrorOrDefinition(errDef, errDef), should.BeEmpty)

	// Not same.
	a.So(ShouldHaveSameErrorDefinitionAs(errDef.WithAttributes("k", "v"), errOtherDef.WithAttributes("k", "v")), should.NotBeEmpty)
	a.So(ShouldEqualErrorOrDefinition(errDef.WithAttributes("k", "v"), errOtherDef.WithAttributes("k", "v")), should.NotBeEmpty)
	a.So(ShouldEqualErrorOrDefinition(errDef, errDef.WithAttributes("k", "v")), should.NotBeEmpty)
	a.So(ShouldEqualErrorOrDefinition(errDef.WithAttributes("k", "v"), errDef), should.NotBeEmpty)
	a.So(ShouldEqualErrorOrDefinition(errDef, errOtherDef), should.NotBeEmpty)
}
