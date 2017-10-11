// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package oauth

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestParseScope(t *testing.T) {
	a := assertions.New(t)

	// valid
	{
		rights, err := ParseScope("RIGHT_APPLICATION_INFO RIGHT_APPLICATION_TRAFFIC_READ")
		a.So(err, should.BeNil)
		a.So(rights, should.Resemble, []ttnpb.Right{
			ttnpb.RIGHT_APPLICATION_INFO,
			ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
		})
	}

	// invalid
	{
		rights, err := ParseScope("RIGHT_APPLICATION_TRAFFIC_READ RIGHT_WEIRD")
		a.So(err, should.NotBeNil)
		a.So(rights, should.BeNil)
	}
}
