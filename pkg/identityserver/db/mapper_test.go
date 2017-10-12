// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package db

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestNameMapper(t *testing.T) {
	a := assertions.New(t)
	a.So(nameMapper("ArchivedAt"), should.Equal, "archived_at")
	a.So(nameMapper("RedirectURI"), should.Equal, "redirect_uri")
	a.So(nameMapper("FrequencyPlanID"), should.Equal, "frequency_plan_id")
	a.So(nameMapper("Name"), should.Equal, "name")
	a.So(nameMapper("Device_Id"), should.Equal, "device_id")
}
