// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package lbslns

import (
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func timePtr(t time.Time) *time.Time { return &t }

func TestTimePtrFromUpInfo(t *testing.T) {
	for _, tc := range []struct {
		Name          string
		GPSTime       int64
		RxTime        float64
		ReferenceTime time.Time
		ExpectedTime  *time.Time
	}{
		{
			Name:          "NoTimestamps",
			ReferenceTime: time.Unix(0, 456),
		},
		{
			Name:          "OnlyRxTime",
			RxTime:        123.456,
			ReferenceTime: time.Unix(0, 456),

			ExpectedTime: timePtr(time.Unix(123, 456000000).UTC()),
		},
		{
			Name:          "EqualGPSTimeAndRxTime",
			GPSTime:       -315964676544, // The timestamp is negative as the UTC epoch precedes the GPS epoch.
			RxTime:        123.456,
			ReferenceTime: time.Unix(123, 456),

			ExpectedTime: timePtr(time.Unix(123, 456000000).UTC()),
		},
		{
			Name:          "OnlyGPSTime",
			GPSTime:       -315964676544, // The timestamp is negative as the UTC epoch precedes the GPS epoch.
			ReferenceTime: time.Unix(0, 456),

			ExpectedTime: timePtr(time.Unix(123, 456000000).UTC()),
		},
		{
			Name:          "MillisecondGPSTime",
			GPSTime:       1321619791991, // This timestamp is in milliseconds, instead of microseconds.
			RxTime:        1637584573,
			ReferenceTime: time.Unix(1637584483, 999974502), // 2021-11-22T12:34:43.999974502Z

			ExpectedTime: timePtr(time.Unix(1637584573, 991000000).UTC()),
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a, _ := test.New(t)
			a.So(TimePtrFromUpInfo(tc.GPSTime, tc.RxTime, tc.ReferenceTime), should.Resemble, tc.ExpectedTime)
		})
	}
}
