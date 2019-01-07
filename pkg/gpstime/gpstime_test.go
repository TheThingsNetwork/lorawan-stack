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

package gpstime_test

import (
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/gpstime"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var (
	epoch          = time.Date(1980, time.January, 6, 0, 0, 0, 0, time.UTC)
	leap1          = time.Date(1981, time.June, 30, 23, 59, 59, 0, time.UTC).Unix() - epoch.Unix() + 1
	leap5          = time.Date(1987, time.December, 31, 23, 59, 59, 0, time.UTC).Unix() - epoch.Unix() + 1 + 4
	now            = time.Date(2017, time.October, 24, 23, 53, 30, 0, time.UTC)
	nowLeaps int64 = 18
)

func TestGPSConversion(t *testing.T) {
	t.Logf("Leap 1: %d Leap 5: %d", leap1, leap5)

	for _, tc := range []struct {
		GPS  int64
		Time time.Time
	}{
		{
			// From LoRaWAN 1.1 specification
			1139322288,
			time.Date(2016, time.February, 12, 14, 24, 31, 0, time.UTC),
		},
		{
			now.Unix() - epoch.Unix() + nowLeaps,
			now,
		},
		{
			42,
			epoch.Add(42 * time.Second),
		},
		{
			0,
			epoch,
		},
		{
			-1,
			epoch.Add(-1 * time.Second),
		},

		{
			leap1 - 2,
			epoch.Add(time.Second * time.Duration(leap1-2)),
		},
		{
			leap1 - 1,
			epoch.Add(time.Second * time.Duration(leap1-1)),
		},
		{
			leap1,
			epoch.Add(time.Second * time.Duration(leap1)),
		},
		{
			leap1 + 1,
			epoch.Add(time.Second * time.Duration(leap1)),
		},
		{
			leap1 + 2,
			epoch.Add(time.Second * time.Duration(leap1+1)),
		},

		{
			leap5 - 2,
			epoch.Add(time.Second * time.Duration(leap5-6)),
		},
		{
			leap5 - 1,
			epoch.Add(time.Second * time.Duration(leap5-5)),
		},
		{
			leap5,
			epoch.Add(time.Second * time.Duration(leap5-4)),
		},
		{
			leap5 + 1,
			epoch.Add(time.Second * time.Duration(leap5-4)),
		},
		{
			leap5 + 2,
			epoch.Add(time.Second * time.Duration(leap5-3)),
		},
	} {
		a := assertions.New(t)
		a.So(Parse(tc.GPS).UnixNano(), should.Resemble, tc.Time.UnixNano())
		if IsLeap(tc.GPS) {
			a.So(ToGPS(tc.Time), should.Equal, tc.GPS+1)
		} else {
			a.So(ToGPS(tc.Time), should.Equal, tc.GPS)
		}
		if a.Failed() {
			t.Errorf("Time: %s, Unix: %d, GPS: %d", tc.Time, tc.Time.Unix(), tc.GPS)
		}
	}
}
