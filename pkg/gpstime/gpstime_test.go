// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package gpstime_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/gpstime"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var (
	epoch          = time.Date(1980, time.January, 6, 0, 0, 0, 0, time.UTC)
	leap1          = time.Duration(time.Date(1981, time.June, 30, 23, 59, 59, 0, time.UTC).UnixNano()-epoch.UnixNano()) + time.Second
	leap5          = time.Duration(time.Date(1987, time.December, 31, 23, 59, 59, 0, time.UTC).UnixNano()-epoch.UnixNano()) + 5*time.Second
	now            = time.Date(2017, time.October, 24, 23, 53, 30, 0, time.UTC)
	nowLeaps int64 = 18
)

func TestGPSConversion(t *testing.T) {
	t.Logf("Leap 1: %d Leap 5: %d", leap1, leap5)

	for i, tc := range []struct {
		GPS  time.Duration
		Time time.Time
	}{
		{
			// From LoRaWAN 1.1 specification
			GPS:  1139322288 * time.Second,
			Time: time.Date(2016, time.February, 12, 14, 24, 31, 0, time.UTC),
		},
		{
			GPS:  time.Duration(now.Unix()-epoch.Unix()+nowLeaps) * time.Second,
			Time: now,
		},
		{
			GPS:  42 * time.Nanosecond,
			Time: epoch.Add(42 * time.Nanosecond),
		},
		{
			GPS:  42 * time.Second,
			Time: epoch.Add(42 * time.Second),
		},
		{
			Time: epoch,
		},
		{
			GPS:  -1 * time.Second,
			Time: epoch.Add(-1 * time.Second),
		},

		{
			GPS:  leap1 - 2*time.Second,
			Time: epoch.Add(leap1 - 2*time.Second),
		},
		{
			GPS:  leap1 - time.Second,
			Time: epoch.Add(leap1 - time.Second),
		},
		{
			GPS:  leap1,
			Time: epoch.Add(leap1),
		},
		{
			GPS:  leap1 + time.Microsecond,
			Time: epoch.Add(leap1 + time.Microsecond),
		},
		{
			GPS:  leap1 + time.Millisecond,
			Time: epoch.Add(leap1 + time.Millisecond),
		},
		{
			GPS:  leap1 + time.Second,
			Time: epoch.Add(leap1),
		},
		{
			GPS:  leap1 + time.Second + time.Nanosecond,
			Time: epoch.Add(leap1 + time.Nanosecond),
		},
		{
			GPS:  leap1 + time.Second + time.Millisecond,
			Time: epoch.Add(leap1 + time.Millisecond),
		},
		{
			GPS:  leap1 + 2*time.Second,
			Time: epoch.Add(leap1 + time.Second),
		},

		{
			GPS:  leap5 - 2*time.Second,
			Time: epoch.Add(leap5 - 6*time.Second),
		},
		{
			GPS:  leap5 - time.Second,
			Time: epoch.Add(leap5 - 5*time.Second),
		},
		{
			GPS:  leap5,
			Time: epoch.Add(leap5 - 4*time.Second),
		},
		{
			GPS:  leap5 + time.Second,
			Time: epoch.Add(leap5 - 4*time.Second),
		},
		{
			GPS:  leap5 + 2*time.Second,
			Time: epoch.Add(leap5 - 3*time.Second),
		},
	} {
		t.Run(fmt.Sprintf("%d/Time:%s/UnixNano:%d/GPS:%d", i, tc.Time, tc.Time.UnixNano(), tc.GPS), func(t *testing.T) {
			a := assertions.New(t)
			a.So(Parse(tc.GPS).UnixNano(), should.Resemble, tc.Time.UnixNano())
			if IsLeapSecond(tc.GPS) {
				a.So(ToGPS(tc.Time), should.Equal, tc.GPS+time.Second)
			} else {
				a.So(ToGPS(tc.Time), should.Equal, tc.GPS)
			}
		})
	}
}
