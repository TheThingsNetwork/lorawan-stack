// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package test

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func app() *ttnpb.Application {
	return &ttnpb.Application{
		ApplicationIdentifier: ttnpb.ApplicationIdentifier{"demo-app"},
		Description:           "Demo application",
	}
}

func TestShouldBeApplication(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeApplication(app(), app()), should.Equal, success)

	modified := app()
	modified.CreatedAt = time.Now()

	a.So(ShouldBeApplication(modified, app()), should.NotEqual, success)
}

func TestShouldBeApplicationIgnoringAutoFields(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeApplicationIgnoringAutoFields(app(), app()), should.Equal, success)

	modified := app()
	modified.Description = "foo"

	a.So(ShouldBeApplicationIgnoringAutoFields(modified, app()), should.NotEqual, success)
}
