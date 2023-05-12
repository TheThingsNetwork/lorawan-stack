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

package javascript_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/goproto"
	"go.thethings.network/lorawan-stack/v3/pkg/scripting"
	"go.thethings.network/lorawan-stack/v3/pkg/scripting/javascript"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestRun(t *testing.T) {
	a := assertions.New(t)

	ctx := test.Context()

	script := `
		function test() {
			return {
				x: 42
			}
		}
	`
	e := javascript.New(scripting.DefaultOptions)
	as, err := e.Run(ctx, script, "test")
	a.So(err, should.BeNil)

	var output struct {
		X int `json:"x"`
	}
	err = as(&output)
	a.So(err, should.BeNil)
	a.So(output.X, should.Equal, 42)
}

func TestRunError(t *testing.T) {
	a := assertions.New(t)

	ctx := test.Context()

	script := `
		function test() {
			throw Error("something didn't work")
		}
	`

	e := javascript.New(scripting.DefaultOptions)
	_, err := e.Run(ctx, script, "test")
	a.So(err, should.NotBeNil)
}

func TestRunStackOverflow(t *testing.T) {
	a := assertions.New(t)

	ctx := test.Context()

	script := `
		function test() {
			var obj = {foo: "bar"};
			obj.ob = obj;
			return obj;
		}
	`

	e := javascript.New(scripting.DefaultOptions)
	_, err := e.Run(ctx, script, "test")
	a.So(err, should.BeNil)
}

func TestRunTimeout(t *testing.T) {
	a := assertions.New(t)

	ctx := test.Context()

	script := `
		function test() {
			while (true) { }
			return {};
		}
	`

	e := javascript.New(scripting.DefaultOptions)
	_, err := e.Run(ctx, script, "test")
	a.So(err, should.NotBeNil)
	a.So(errors.IsDeadlineExceeded(err), should.BeTrue)
}

func TestStackOverflow(t *testing.T) {
	t.Parallel()
	a, ctx := test.New(t)

	script := `
		function test(input) {
			var attempt = {};
			attempt.test = attempt;
			return attempt;
		}
	`
	e := javascript.New(scripting.DefaultOptions)
	as, err := e.Run(ctx, script, "test")
	a.So(err, should.BeNil)

	m := make(map[string]any)
	err = as(&m)
	a.So(err, should.BeNil)

	_, err = goproto.Struct(m)
	a.So(err, should.NotBeNil)

	_, err = goproto.Struct(m)
	a.So(err, should.NotBeNil)
}
