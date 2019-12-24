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
	"time"

	"go.thethings.network/lorawan-stack/pkg/gatewayserver/scheduling"
)

type mockTimeSource struct {
	time.Time
}

func (s *mockTimeSource) Now() time.Time {
	return s.Time
}

type mockClock struct {
	t scheduling.ConcentratorTime
}

func (c *mockClock) IsSynced() bool {
	return c.t > 0
}
func (c *mockClock) FromServerTime(_ time.Time) (scheduling.ConcentratorTime, bool) {
	return c.t, true
}
func (c *mockClock) ToServerTime(t scheduling.ConcentratorTime) time.Time {
	return time.Unix(0, 0).Add(time.Duration(t - c.t))
}
func (c *mockClock) FromGatewayTime(t time.Time) (scheduling.ConcentratorTime, bool) {
	return scheduling.ConcentratorTime(t.Sub(time.Unix(0, 0))), true
}
func (c *mockClock) FromTimestampTime(timestamp uint32) scheduling.ConcentratorTime {
	return c.t + scheduling.ConcentratorTime(time.Duration(timestamp)*time.Microsecond)
}

type mockRTTs struct {
	Min,
	Max,
	Median time.Duration
	Count int
}

func (r *mockRTTs) Stats() (min, max, median time.Duration, count int) {
	min = r.Min
	max = r.Max
	median = r.Median
	count = r.Count
	return
}

func boolPtr(v bool) *bool                       { return &v }
func durationPtr(d time.Duration) *time.Duration { return &d }
func timePtr(t time.Time) *time.Time             { return &t }

func init() {
	scheduling.DutyCycleWindow = 10 * time.Second
}
