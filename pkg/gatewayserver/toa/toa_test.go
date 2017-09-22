// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package toa

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/types"
	. "github.com/smartystreets/assertions"
)

func TestComputeLoRa(t *testing.T) {
	a := New(t)

	var toa time.Duration
	var err error

	_, err = ComputeLoRa(10, types.DataRate{LoRa: "SFUT"}, "4/5")
	a.So(err, ShouldNotBeNil)

	_, err = ComputeLoRa(10, types.DataRate{LoRa: "SF10BW125"}, "1/9")
	a.So(err, ShouldNotBeNil)

	// Test different SFs
	sfTests := map[types.DataRate]uint{
		types.DataRate{LoRa: "SF7BW125"}:  41216,
		types.DataRate{LoRa: "SF8BW125"}:  72192,
		types.DataRate{LoRa: "SF9BW125"}:  144384,
		types.DataRate{LoRa: "SF10BW125"}: 288768,
		types.DataRate{LoRa: "SF11BW125"}: 577536,
		types.DataRate{LoRa: "SF12BW125"}: 991232,
	}
	for dr, us := range sfTests {
		toa, err = ComputeLoRa(10, dr, "4/5")
		a.So(err, ShouldBeNil)
		a.So(toa, ShouldAlmostEqual, time.Duration(us)*time.Microsecond)
	}

	// Test different BWs
	bwTests := map[types.DataRate]uint{
		types.DataRate{LoRa: "SF7BW125"}: 41216,
		types.DataRate{LoRa: "SF7BW250"}: 20608,
		types.DataRate{LoRa: "SF7BW500"}: 10304,
	}
	for dr, us := range bwTests {
		toa, err = ComputeLoRa(10, dr, "4/5")
		a.So(err, ShouldBeNil)
		a.So(toa, ShouldAlmostEqual, time.Duration(us)*time.Microsecond)
	}

	// Test different CRs
	crTests := map[string]uint{
		"4/5": 41216,
		"4/6": 45312,
		"4/7": 49408,
		"4/8": 53504,
	}
	for cr, us := range crTests {
		toa, err = ComputeLoRa(10, types.DataRate{LoRa: "SF7BW125"}, cr)
		a.So(err, ShouldBeNil)
		a.So(toa, ShouldAlmostEqual, time.Duration(us)*time.Microsecond)
	}

	// Test different payload sizes
	plTests := map[uint]uint{
		13: 46336,
		14: 46336,
		15: 46336,
		16: 51456,
		17: 51456,
		18: 51456,
		19: 51456,
	}
	for size, us := range plTests {
		toa, err = ComputeLoRa(size, types.DataRate{LoRa: "SF7BW125"}, "4/5")
		a.So(err, ShouldBeNil)
		a.So(toa, ShouldAlmostEqual, time.Duration(us)*time.Microsecond)
	}

}

func TestComputeFSK(t *testing.T) {
	a := New(t)
	toa, err := ComputeFSK(200, 50000)
	a.So(err, ShouldBeNil)
	a.So(toa, ShouldAlmostEqual, 33760*time.Microsecond)
}
