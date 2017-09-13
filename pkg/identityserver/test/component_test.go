// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package test

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/identityserver/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func component() *types.DefaultComponent {
	return &types.DefaultComponent{
		ID:   "alice-handler",
		Type: types.Handler,
	}
}

func TestShouldBeComponent(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeComponent(component(), component()), should.Equal, success)

	modified := component()
	modified.Created = time.Now()

	a.So(ShouldBeComponent(modified, component()), should.NotEqual, success)
}

func TestShouldBeComponentIgnoringAutoFields(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeComponentIgnoringAutoFields(component(), component()), should.Equal, success)

	modified := component()
	modified.ID = "foo"

	a.So(ShouldBeComponentIgnoringAutoFields(modified, component()), should.NotEqual, success)
}
