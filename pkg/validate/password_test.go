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

package validate

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestPassword(t *testing.T) {
	a := assertions.New(t)
	a.So(Password("_Foo1__BaR"), should.BeNil)
	a.So(Password("_Foo1."), should.NotBeNil)
	a.So(Password("hhHiHIHIii1555"), should.BeNil)
	a.So(Password("Hi12//i12ddddd"), should.BeNil)
	a.So(Password(1), should.NotBeNil)
	a.So(Password(""), should.NotBeNil)
}
