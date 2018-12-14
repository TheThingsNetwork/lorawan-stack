// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

type mockClock struct {
	t time.Duration
}

func (c *mockClock) Now() time.Duration {
	return c.t
}
func (c *mockClock) ConcentratorTime(t time.Time) time.Duration {
	return t.Sub(time.Unix(0, 0))
}

func boolPtr(v bool) *bool                       { return &v }
func durationPtr(d time.Duration) *time.Duration { return &d }

func init() {
	scheduling.DutyCycleWindow = 10 * time.Second
}
