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
	"fmt"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

var testDescriptor = &errors.ErrDescriptor{
	MessageFormat: "Test error",
	Code:          42,
}

func init() {
	testDescriptor.Register()
}

func TestShouldDescribe(t *testing.T) {
	a := assertions.New(t)

	// Happy flow.
	a.So(ShouldDescribe(testDescriptor.New(nil), testDescriptor), should.BeEmpty)
	a.So(ShouldNotDescribe(testDescriptor.New(nil), testDescriptor), should.NotBeEmpty)

	// Unknown error.
	a.So(ShouldDescribe(fmt.Errorf("unknown error"), testDescriptor), should.NotBeEmpty)
	a.So(ShouldNotDescribe(fmt.Errorf("unknown error"), testDescriptor), should.BeEmpty)

	// Wrong namespace or code.
	a.So(ShouldDescribe(errors.New("test"), testDescriptor), should.NotBeEmpty)
	a.So(ShouldNotDescribe(errors.New("test"), testDescriptor), should.BeEmpty)
}
