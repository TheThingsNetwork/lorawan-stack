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
	"strconv"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestSubBandScheduleUnrestricted(t *testing.T) {
	ctx := test.Context()
	band := band.DutyCycle{
		Value: 1,
	}
	clock := &mockClock{}
	sb := scheduling.NewSubBand(ctx, band, clock, nil)
	for i, tc := range []struct {
		Starts            time.Duration
		Duration          time.Duration
		ExpectUtilization float32
	}{
		{
			Starts:            1 * time.Second,
			Duration:          2 * time.Second,
			ExpectUtilization: 0.2,
			// [11                  ]
			//  ^^
		},
		{
			Starts:            4 * time.Second,
			Duration:          1 * time.Second,
			ExpectUtilization: 0.3,
			// [11 2                ]
			//  ^^^^
		},
		{
			Starts:            11 * time.Second,
			Duration:          1 * time.Second,
			ExpectUtilization: 0.3,
			// [11 2      3         ]
			//   ^^^^^^^^^^
		},
		{
			Starts:            13 * time.Second,
			Duration:          1 * time.Second,
			ExpectUtilization: 0.3,
			// [11 2      3 4       ]
			//     ^^^^^^^^^^
		},
		{
			Starts:            15 * time.Second,
			Duration:          3 * time.Second,
			ExpectUtilization: 0.5,
			// [11 2      3 4 555   ]
			//         ^^^^^^^^^^
		},
	} {
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			a := assertions.New(t)

			em := scheduling.NewEmission(tc.Starts, tc.Duration)
			err := sb.Schedule(em, ttnpb.TxSchedulePriority_NORMAL)
			a.So(err, should.BeNil)

			clock.t = tc.Starts + tc.Duration
			utilization := sb.DutyCycleUtilization()
			a.So(utilization, should.Equal, tc.ExpectUtilization)
		})
	}
}

func TestSubBandScheduleRestricted(t *testing.T) {
	ctx := test.Context()
	band := band.DutyCycle{
		Value: 0.5,
	}
	clock := &mockClock{}
	ceilings := map[ttnpb.TxSchedulePriority]float32{
		ttnpb.TxSchedulePriority_NORMAL:  0.5, // Duty-cycle <= 0.25
		ttnpb.TxSchedulePriority_HIGHEST: 1.0, // Duty-cycle <= 0.50
	}
	sb := scheduling.NewSubBand(ctx, band, clock, ceilings)
	for i, tc := range []struct {
		Starts            time.Duration
		Duration          time.Duration
		Priority          ttnpb.TxSchedulePriority
		ExpectError       func(error) bool
		ExpectUtilization float32
	}{
		{
			Starts:            6 * time.Second,
			Duration:          1 * time.Second,
			Priority:          ttnpb.TxSchedulePriority_NORMAL,
			ExpectUtilization: 0.1 / 0.5,
			// [     1              ]
			//  ^^^^^^
		},
		{
			Starts:            14 * time.Second,
			Duration:          1 * time.Second,
			Priority:          ttnpb.TxSchedulePriority_NORMAL,
			ExpectUtilization: 0.2 / 0.5,
			// [     1       2      ]
			//      ^^^^^^^^^^
		},
		{
			Starts:            18 * time.Second,
			Duration:          1 * time.Second,
			Priority:          ttnpb.TxSchedulePriority_NORMAL,
			ExpectUtilization: 0.2 / 0.5,
			// [     1       2   3  ]
			//          ^^^^^^^^^^
		},
		{
			Starts:            11 * time.Second,
			Duration:          1 * time.Second,
			Priority:          ttnpb.TxSchedulePriority_NORMAL,
			ExpectError:       errors.IsResourceExhausted,
			ExpectUtilization: 0.1 / 0.5,
			// [     1    X  2   3  ]
			//   ^^^^^^^^^^
			//            ^^^^^^^^^^
		},
		{
			Starts:            11 * time.Second,
			Duration:          1 * time.Second,
			Priority:          ttnpb.TxSchedulePriority_HIGHEST,
			ExpectUtilization: 0.2 / 0.5,
			// [     1    4  2   3  ]
			//   ^^^^^^^^^^
			//            ^^^^^^^^^^
		},
	} {
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			a := assertions.New(t)

			em := scheduling.NewEmission(tc.Starts, tc.Duration)
			err := sb.Schedule(em, tc.Priority)
			if tc.ExpectError != nil {
				a.So(tc.ExpectError(err), should.BeTrue)
			} else {
				a.So(err, should.BeNil)
			}

			clock.t = tc.Starts + tc.Duration
			utilization := sb.DutyCycleUtilization()
			a.So(utilization, should.Equal, tc.ExpectUtilization)
		})
	}
}

func TestScheduleAnytimeRestricted(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()
	band := band.DutyCycle{
		Value: 0.5,
	}
	clock := &mockClock{}
	ceilings := map[ttnpb.TxSchedulePriority]float32{
		ttnpb.TxSchedulePriority_NORMAL:  0.5, // Duty-cycle <= 0.25
		ttnpb.TxSchedulePriority_HIGHEST: 1.0, // Duty-cycle <= 0.50
	}
	sb := scheduling.NewSubBand(ctx, band, clock, ceilings)

	for _, t := range []time.Duration{6 * time.Second, 14 * time.Second, 18 * time.Second} {
		em := scheduling.NewEmission(t, time.Second)
		err := sb.Schedule(em, ttnpb.TxSchedulePriority_NORMAL)
		a.So(err, should.BeNil)
	}
	// [     1       2   3        ]

	// Step naively 1 second.
	{
		from := 11 * time.Second
		next := func() time.Duration {
			res := from
			from += time.Second
			return res
		}
		em, err := sb.ScheduleAnytime(time.Second, next, ttnpb.TxSchedulePriority_NORMAL)
		a.So(err, should.BeNil)
		a.So(em.Starts(), should.Equal, 16*time.Second)
		// [     1       2 4 3        ]
		//        ^^^^^^^^^^
		//                 ^^^^^^^^^^
	}

	// Get the first available option after all transmissions.
	{
		next := func() time.Duration {
			return 19 * time.Second
		}
		em, err := sb.ScheduleAnytime(time.Second, next, ttnpb.TxSchedulePriority_NORMAL)
		a.So(err, should.BeNil)
		a.So(em.Starts(), should.Equal, 26*time.Second)
		// [     1       2 4 3       5]
		//                  ^^^^^^^^^^
	}

	// Fail when the emission hits any duty-cycle limitation.
	{
		next := func() time.Duration {
			return 19 * time.Second
		}
		_, err := sb.ScheduleAnytime(5*time.Second, next, ttnpb.TxSchedulePriority_NORMAL)
		a.So(err, should.HaveSameErrorDefinitionAs, scheduling.ErrDutyCycle)
	}
}
