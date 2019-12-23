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

package scheduling_test

import (
	"math"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestRolloverClock(t *testing.T) {
	a := assertions.New(t)
	clock := &scheduling.RolloverClock{}

	clock.SyncWithGateway(10000000, time.Unix(10, 0), time.Unix(0, 0)) // The gateway has no idea of time.
	a.So(clock.FromServerTime(time.Unix(10, 100)), should.Equal, 10000000*time.Microsecond+100)
	a.So(clock.ToServerTime(scheduling.ConcentratorTime(10000000*time.Microsecond+100)), should.Equal, time.Unix(10, 100))

	gatewayTime, ok := clock.FromGatewayTime(time.Unix(0, 100))
	a.So(ok, should.BeTrue)
	a.So(gatewayTime, should.Equal, 10000000*time.Microsecond+100)

	{
		// Test timestamp time without rollover.
		a.So(clock.FromTimestampTime(9999999), should.Equal, 9999999*time.Microsecond)
		a.So(clock.FromTimestampTime(10000000), should.Equal, 10000000*time.Microsecond)
		a.So(clock.FromTimestampTime(10000001), should.Equal, 10000001*time.Microsecond)
	}

	{
		// Test first roll-over to 4299967295 us (math.MaxUint32 + 5000000).
		clock.Sync(math.MaxUint32, time.Unix(10, 0).Add(math.MaxUint32*time.Microsecond))
		passed := time.Microsecond * (math.MaxUint32 + 5000000)
		clock.SyncWithGateway(5000000, time.Unix(10, 0).Add(passed), time.Unix(0, 0).Add(passed))
		a.So(clock.FromServerTime(time.Unix(10, 100).Add(passed)), should.Equal, passed+100)
	}

	{
		// Test second roll-over to 8589934590 us (2 * math.MaxUint32).
		clock.Sync(math.MaxUint32, time.Unix(10, 0).Add(2*math.MaxUint32*time.Microsecond))
		passed := time.Microsecond * 2 * math.MaxUint32
		clock.SyncWithGateway(0, time.Unix(10, 0).Add(passed), time.Unix(0, 0).Add(passed))
		a.So(clock.FromServerTime(time.Unix(10, 100).Add(passed)), should.Equal, passed+100)
	}

	{
		// Test reset of gateway time and rollover.
		passed := time.Microsecond * (2*math.MaxUint32 + 5000000)
		clock.Sync(5000000, time.Unix(10, 0).Add(passed))
		_, ok := clock.FromGatewayTime(time.Unix(0, 100))
		a.So(ok, should.BeFalse)
	}
}
