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

package validate

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestID(t *testing.T) {
	a := assertions.New(t)
	a.So(ID("app-test"), should.BeNil)
	a.So(ID("rx1"), should.BeNil)
	a.So(ID("dddsdddjjjjdddsdddsdsdsdsdsdsdsdsdfw"), should.BeNil)
	a.So(ID("-app-test"), should.NotBeNil)
	a.So(ID("app-test-"), should.NotBeNil)
	a.So(ID("app_test"), should.NotBeNil)
	a.So(ID("_dd"), should.NotBeNil)
	a.So(ID("A"), should.NotBeNil)
	a.So(ID("AB"), should.NotBeNil)
	a.So(ID(12), should.NotBeNil)
	a.So(ID(1), should.NotBeNil)
	a.So(ID("dddsdddjjjjdddsdddsdsdsdsdsdsdsdsdf-w"), should.NotBeNil)
	a.So(ID("d-d-d-s-d-d-d-j-j-l-k-j-k-j-k-j-k-j-k-j-k-l-j-k-l-j-k-f-j-d-s-k-f-j-d-s"), should.NotBeNil)

}
