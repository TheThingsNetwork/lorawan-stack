// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package javascript_test

import (
	"context"
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/scripting"
	. "github.com/TheThingsNetwork/ttn/pkg/scripting/javascript"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestRun(t *testing.T) {
	a := assertions.New(t)

	ctx := context.Background()

	script := `
		(function () {
			return {
				x: 42
			}
		})()
	`

	e := New(scripting.DefaultOptions)
	output, err := e.Run(ctx, script, nil)
	a.So(err, should.BeNil)
	a.So(output, should.HaveSameTypeAs, map[string]interface{}{})
	a.So(output.(map[string]interface{})["x"], should.Equal, 42)
}

func TestRunError(t *testing.T) {
	a := assertions.New(t)

	ctx := context.Background()

	script := `
		(function () {
			throw Error("something didn't work")
		})()
	`

	e := New(scripting.DefaultOptions)
	_, err := e.Run(ctx, script, nil)
	a.So(err, should.NotBeNil)
	a.So(errors.GetType(err), should.Equal, errors.External)
}

func TestRunStackOverflow(t *testing.T) {
	a := assertions.New(t)

	ctx := context.Background()

	script := `
		(function () {
			var obj = {foo: "bar"};
			obj.ob = obj;
			return obj;
		})()
	`

	e := New(scripting.DefaultOptions)
	_, err := e.Run(ctx, script, nil)
	a.So(err, should.NotBeNil)
}

func TestRunTimeout(t *testing.T) {
	a := assertions.New(t)

	ctx := context.Background()

	script := `
		(function () {
			while (true) { }
			return {};
		})()
	`

	e := New(scripting.DefaultOptions)
	_, err := e.Run(ctx, script, nil)
	a.So(err, should.NotBeNil)
	a.So(errors.GetCode(err), should.Equal, scripting.ErrRuntime.Code)
	a.So(errors.GetType(errors.Cause(err)), should.Equal, errors.Timeout)
}
