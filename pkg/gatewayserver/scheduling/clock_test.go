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
	"strconv"
	"testing"
	"time"

	"github.com/smarty/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestRolloverClock(t *testing.T) {
	a := assertions.New(t)
	clock := &RolloverClock{}

	for i, stc := range []struct {
		Absolute ConcentratorTime
		Relative uint32
	}{
		{
			Absolute: ConcentratorTime(20 * time.Minute),
			Relative: uint32(20 * time.Minute / time.Microsecond),
		},
		{
			// 1 rollover.
			Absolute: ConcentratorTime(1<<32*time.Microsecond) + ConcentratorTime(5*time.Second),
			Relative: uint32(5000000),
		},
		{
			// 3 rollovers (1 existing + 2 server time rollovers).
			Absolute: ConcentratorTime(3<<32*time.Microsecond) + ConcentratorTime(30*time.Minute),
			Relative: uint32(30 * time.Minute / time.Microsecond),
		},
		{
			// 5 rollovers (3 existing + 1 concentrator timestamp rollover + 1 server time rollover).
			Absolute: ConcentratorTime(5<<32*time.Microsecond) + ConcentratorTime(1*time.Second),
			Relative: uint32(1000000),
		},
		{
			// 5 rollovers (5 existing) and advance to end of concentrator time epoch.
			Absolute: ConcentratorTime(6<<32*time.Microsecond) - ConcentratorTime(1*time.Second),
			Relative: uint32(4293967296),
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			serverTime := time.Unix(0, 0).Add(time.Duration(stc.Absolute))
			// Run twice; once synchronizing with rollover detection, and once synchronizing with the known concentrator time.
			for i := 0; i < 2; i++ {
				t.Run([]string{"DetectRollover", "ResetAbsolute"}[i], func(t *testing.T) {
					if i == 0 {
						clock.Sync(stc.Relative, serverTime)
					} else {
						clock.SyncWithGatewayConcentrator(stc.Relative, serverTime, nil, stc.Absolute)
					}

					for _, tc := range []struct {
						D        time.Duration
						Rollover bool
					}{
						{
							D:        -5 * time.Second,
							Rollover: false,
						},
						{
							D:        5 * time.Second,
							Rollover: false,
						},
						{
							D:        30 * time.Minute,
							Rollover: false,
						},
						{
							D:        2 * time.Hour,
							Rollover: true,
						},
					} {
						t.Run(tc.D.String(), func(t *testing.T) {
							a := assertions.New(t)

							v, ok := clock.FromServerTime(serverTime.Add(tc.D))
							a.So(ok, should.BeTrue)
							a.So(v, should.Equal, stc.Absolute+ConcentratorTime(tc.D))

							d := tc.D / time.Microsecond
							rollover := d > math.MaxUint32/2 || d < -math.MaxUint32/2
							a.So(rollover, should.Equal, tc.Rollover)

							if !rollover {
								ts := uint32(time.Duration(stc.Relative) + tc.D/time.Microsecond)
								v = clock.FromTimestampTime(ts)
								a.So(v, should.Equal, stc.Absolute+ConcentratorTime(tc.D))
							}
						})
					}
				})
			}
		})
	}

	// Test Rollovers where the absolute diff low.
	const rfc3339Micro = "2006-01-02T15:04:05.999999Z07:00"
	now, _ := time.Parse(rfc3339Micro, "2021-08-27T09:06:21.001774Z")

	timestamps := []int64{
		53761741353978084, // Close to a rollover
		53761744960482044, // Next one arrives an hour later
	}
	var (
		prev      *int64
		sessionID int64
	)
	for i, xtimeIn := range timestamps {
		xtimeIn := xtimeIn
		diff := int64(0)
		if prev != nil {
			diff = xtimeIn - *prev
		}
		prev = &xtimeIn

		timestamp := uint32(xtimeIn & 0xFFFFFFFF)
		serverTime := now
		if i != 0 {
			serverTime, _ = time.Parse(rfc3339Micro, "2021-08-27T10:06:27.487028Z")
		}

		if i == 0 {
			t.Log("Synchronizing gateway concentrator")
			sessionID = xtimeIn >> 48
			clock.SyncWithGatewayConcentrator(timestamp, serverTime, nil, ConcentratorTime(time.Duration(xtimeIn&0xFFFFFFFFFFFF)*time.Microsecond))
		}
		rx := clock.Sync(timestamp, serverTime)
		tx := clock.FromTimestampTime(timestamp)

		t.Logf("xtimeIn=%016X tmst=%08X concentrator=%016X received=%v diff=%d", xtimeIn, timestamp, tx/1000, serverTime, diff)

		a.So(tx-rx, should.BeZeroValue)

		xtimeOut := sessionID<<48 | (int64(tx) / int64(time.Microsecond) & 0xFFFFFFFFFF)
		a.So(time.Duration(xtimeOut-xtimeIn)*time.Microsecond, should.BeZeroValue)
	}
}

func TestSyncWithGatewayConcentrator(t *testing.T) {
	a := assertions.New(t)

	clock := &RolloverClock{}
	clock.SyncWithGatewayConcentrator(0x496054D6, time.Now(), nil, ConcentratorTime(0xAA496054D6)*ConcentratorTime(time.Microsecond))
	v := int64(clock.FromTimestampTime(0x499D5DD6)) / int64(time.Microsecond)
	a.So(v, should.Equal, int64(0xAA499D5DD6))
}

// TestIssue2581 is a test case for resolving https://github.com/TheThingsNetwork/lorawan-stack/issues/2581.
func TestIssue2581(t *testing.T) {
	a := assertions.New(t)

	clock := &RolloverClock{}

	timestamps := []int64{
		63331869818403100,
		63331869827372300,
		63331869837372200,
		63331869847372000,
		63331869857377000,
		63331869867371600,
		63331869877371600,
		63331869878424700,
		// snip
		63331870307366100,
		63331870317365900,
		63331870327365900,
		63331870337365800,
		63331870337361700, // Before the previous one
		63331870347365700,
		63331870357365600,
		63331870367365500,
	}
	var (
		prev      *int64
		sessionID int64
	)
	for i, xtimeIn := range timestamps {
		xtimeIn := xtimeIn
		diff := int64(0)
		if prev != nil {
			diff = xtimeIn - *prev
		}
		prev = &xtimeIn

		// timestamp can go back, but serverTime is always increasing
		timestamp := uint32(xtimeIn & 0xFFFFFFFF)
		serverTime := time.Now()

		if i == 0 {
			t.Log("Synchronizing gateway concentrator")
			sessionID = xtimeIn >> 48
			clock.SyncWithGatewayConcentrator(timestamp, serverTime, nil, ConcentratorTime(time.Duration(xtimeIn&0xFFFFFFFFFFFF)*time.Microsecond))
		}
		rx := clock.Sync(timestamp, serverTime)
		tx := clock.FromTimestampTime(timestamp)

		t.Logf("xtimeIn=%016X tmst=%08X concentrator=%016X received=%v diff=%d", xtimeIn, timestamp, tx/1000, serverTime, diff)

		a.So(tx-rx, should.BeZeroValue)

		xtimeOut := sessionID<<48 | (int64(tx) / int64(time.Microsecond) & 0xFFFFFFFFFF)
		a.So(time.Duration(xtimeOut-xtimeIn)*time.Microsecond, should.BeZeroValue)
	}
}
