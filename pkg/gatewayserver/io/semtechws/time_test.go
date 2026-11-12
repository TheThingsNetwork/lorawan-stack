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

package semtechws

import (
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func timePtr(t time.Time) *time.Time { return &t }

func TestTimePtrFromUpInfo(t *testing.T) {
	t.Parallel()

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
			Name:   "OnlyRxTime",
			RxTime: 123.456,

			ExpectedTime: timePtr(time.Unix(123, 456000000).UTC()),
		},
		{
			Name:    "EqualGPSTimeAndRxTime",
			GPSTime: -315964676544000, // microseconds. The timestamp is negative as the UTC epoch precedes the GPS epoch.
			RxTime:  123.456,

			ExpectedTime: timePtr(time.Unix(123, 456000000).UTC()),
		},
		{
			Name:    "OnlyGPSTime",
			GPSTime: 123456000, // microseconds.

			ExpectedTime: timePtr(time.Unix(315964923, 456000000).UTC()),
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			a, _ := test.New(t)
			a.So(TimePtrFromUpInfo(tc.GPSTime, tc.RxTime), should.Resemble, tc.ExpectedTime)
		})
	}
}

func TestIsExtendedXTimeSession(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name     string
		XTime    int64
		Expected bool
	}{
		{
			Name:     "NonExtendedTimestamp",
			XTime:    4290303429,
			Expected: false,
		},
		{
			Name:     "NonExtendedTimestampAfterRollover",
			XTime:    780557,
			Expected: false,
		},
		{
			Name:     "Session",
			XTime:    0x42<<48 | 0x1FFFFFFFF,
			Expected: true,
		},
		{
			Name:     "SessionWithRadioUnit",
			XTime:    0x0142<<48 | 0x1FFFFFFFF,
			Expected: true,
		},
		{
			Name:     "RadioUnitWithoutSession",
			XTime:    0x0100<<48 | 0xFFFFFFFF,
			Expected: false,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			a, _ := test.New(t)
			a.So(IsExtendedXTimeSession(SessionIDFromXTime(tc.XTime)), should.Equal, tc.Expected)
		})
	}
}

func TestConcentratorTimeToTimestampXTime(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name             string
		SessionID        int32
		ConcentratorTime scheduling.ConcentratorTime
		Expected         int64
	}{
		{
			Name:             "BeforeRollover",
			ConcentratorTime: scheduling.ConcentratorTime(4290303429 * time.Microsecond),
			Expected:         4290303429,
		},
		{
			Name:             "AfterRollover",
			ConcentratorTime: scheduling.ConcentratorTime((1<<32 + 780557) * time.Microsecond),
			Expected:         780557,
		},
		{
			Name:             "AfterMultipleRollovers",
			ConcentratorTime: scheduling.ConcentratorTime((3<<32 + 780557) * time.Microsecond),
			Expected:         780557,
		},
		{
			Name:             "Negative",
			ConcentratorTime: scheduling.ConcentratorTime(-500 * time.Millisecond),
			Expected:         1<<32 - 500000,
		},
		{
			Name:             "WithRadioUnit",
			SessionID:        0x0100,
			ConcentratorTime: scheduling.ConcentratorTime((1<<32 + 780557) * time.Microsecond),
			Expected:         0x0100<<48 | 780557,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			a, _ := test.New(t)
			a.So(ConcentratorTimeToTimestampXTime(tc.SessionID, tc.ConcentratorTime), should.Equal, tc.Expected)
		})
	}
}
