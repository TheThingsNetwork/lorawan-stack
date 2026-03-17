// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestLastUplink(t *testing.T) {
	a, ctx := test.New(t)
	ctx = NewContextWithSession(ctx, &Session{})

	// No uplink data initially.
	_, _, _, ok := GetLastUplink(ctx)
	a.So(ok, should.BeFalse)

	// Store uplink timing info.
	now := time.Now().UTC()
	xtime := int64(12666373963464220)
	gpstime := int64(1232095200000000)
	UpdateLastUplink(ctx, xtime, gpstime, now)

	// Retrieve and verify.
	gotXTime, gotGPSTime, gotReceivedAt, ok := GetLastUplink(ctx)
	if !a.So(ok, should.BeTrue) {
		t.FailNow()
	}
	a.So(gotXTime, should.Equal, xtime)
	a.So(gotGPSTime, should.Equal, gpstime)
	a.So(gotReceivedAt, should.Equal, now)

	// Update with new uplink.
	later := now.Add(5 * time.Second)
	xtime2 := int64(12666373963564220)
	UpdateLastUplink(ctx, xtime2, gpstime, later)

	gotXTime, _, gotReceivedAt, ok = GetLastUplink(ctx)
	if !a.So(ok, should.BeTrue) {
		t.FailNow()
	}
	a.So(gotXTime, should.Equal, xtime2)
	a.So(gotReceivedAt, should.Equal, later)

	// Non-GPS gateway: gpstime=0 is stored and retrieved.
	UpdateLastUplink(ctx, xtime, 0, now)
	_, gotGPSTime, _, ok = GetLastUplink(ctx)
	if !a.So(ok, should.BeTrue) {
		t.FailNow()
	}
	a.So(gotGPSTime, should.Equal, 0)
}
