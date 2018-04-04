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

package scheduling

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestWindowTiming(t *testing.T) {
	a := assertions.New(t)

	testingTime := SystemTime(time.Now())
	s := Span{
		Start:    testingTime,
		Duration: time.Second,
	}

	a.So(s.End().Equal(testingTime.Add(time.Second)), should.BeTrue)
}

func TestWindowContains(t *testing.T) {
	a := assertions.New(t)

	testingTime := SystemTime(time.Now())
	s := Span{
		Start:    testingTime,
		Duration: time.Second,
	}

	comparisons := map[Timestamp]bool{
		testingTime: true,

		testingTime.Add(time.Second):      true,
		testingTime.Add(time.Microsecond): true,

		testingTime.Add(2 * time.Second):       false,
		testingTime.Add(-1 * time.Millisecond): false,
	}

	for compared, result := range comparisons {
		a.So(s.Contains(compared), should.Equal, result)
	}
}

func TestPrecedingComparison(t *testing.T) {
	a := assertions.New(t)

	testingTime := SystemTime(time.Now())
	s := Span{
		Start:    testingTime,
		Duration: time.Second,
	}

	comparisons := map[Span]bool{
		{Start: testingTime.Add(-1 * time.Second), Duration: 2 * time.Second}:  true,
		{Start: testingTime.Add(-1 * time.Second), Duration: time.Millisecond}: true,
		{Start: testingTime.Add(-1 * time.Second), Duration: time.Minute}:      true,

		{Start: testingTime, Duration: time.Millisecond}: false,
		{Start: testingTime, Duration: time.Second}:      false,
		{Start: testingTime, Duration: time.Minute}:      false,

		{Start: testingTime.Add(time.Millisecond), Duration: time.Millisecond}: false,
		{Start: testingTime.Add(time.Millisecond), Duration: time.Second}:      false,
		{Start: testingTime.Add(time.Millisecond), Duration: time.Minute}:      false,

		{Start: testingTime.Add(2 * time.Second), Duration: time.Millisecond}: false,
		{Start: testingTime.Add(2 * time.Second), Duration: time.Second}:      false,
		{Start: testingTime.Add(2 * time.Second), Duration: time.Minute}:      false,
	}

	for compared, result := range comparisons {
		a.So(compared.StartsBefore(s), should.Equal, result)
	}
}

func TestIsProlongedByComparison(t *testing.T) {
	a := assertions.New(t)

	testingTime := SystemTime(time.Now())
	s := Span{
		Start:    testingTime,
		Duration: time.Second,
	}

	comparisons := map[Span]bool{
		{Start: testingTime.Add(-1 * time.Second), Duration: 2 * time.Second}:  false,
		{Start: testingTime.Add(-1 * time.Second), Duration: time.Millisecond}: false,
		{Start: testingTime.Add(-1 * time.Second), Duration: time.Minute}:      true,

		{Start: testingTime, Duration: time.Millisecond}: false,
		{Start: testingTime, Duration: time.Second}:      false,
		{Start: testingTime, Duration: time.Minute}:      true,

		{Start: testingTime.Add(time.Millisecond), Duration: time.Millisecond}: false,
		{Start: testingTime.Add(time.Millisecond), Duration: time.Second}:      true,
		{Start: testingTime.Add(time.Millisecond), Duration: time.Minute}:      true,

		{Start: testingTime.Add(2 * time.Second), Duration: time.Millisecond}: true,
		{Start: testingTime.Add(2 * time.Second), Duration: time.Second}:      true,
		{Start: testingTime.Add(2 * time.Second), Duration: time.Minute}:      true,
	}

	for compared, result := range comparisons {
		a.So(s.IsProlongedBy(compared), should.Equal, result)
	}
}

func TestOverlap(t *testing.T) {
	a := assertions.New(t)

	testingTime := SystemTime(time.Now())
	s := Span{
		Start:    testingTime,
		Duration: time.Second,
	}

	comparisons := map[Span]bool{
		{Start: testingTime.Add(-1 * time.Second), Duration: time.Second}:      false,
		{Start: testingTime.Add(-1 * time.Second), Duration: 2 * time.Second}:  true,
		{Start: testingTime.Add(-1 * time.Second), Duration: time.Millisecond}: false,
		{Start: testingTime.Add(-1 * time.Second), Duration: time.Minute}:      true,

		{Start: testingTime, Duration: time.Millisecond}: true,
		{Start: testingTime, Duration: time.Second}:      true,
		{Start: testingTime, Duration: time.Minute}:      true,

		{Start: testingTime.Add(time.Millisecond), Duration: time.Millisecond}: true,
		{Start: testingTime.Add(time.Millisecond), Duration: time.Second}:      true,
		{Start: testingTime.Add(time.Millisecond), Duration: time.Minute}:      true,

		{Start: testingTime.Add(2 * time.Second), Duration: time.Millisecond}: false,
		{Start: testingTime.Add(2 * time.Second), Duration: time.Second}:      false,
		{Start: testingTime.Add(2 * time.Second), Duration: time.Minute}:      false,
	}

	for compared, result := range comparisons {
		a.So(s.Overlaps(compared), should.Equal, result)
	}
}

func TestTimeOffAir(t *testing.T) {
	a := assertions.New(t)

	testingTime := SystemTime(time.Now())
	s := Span{
		Start:    testingTime,
		Duration: time.Second,
	}

	comparisons := map[ttnpb.FrequencyPlan_TimeOffAir]time.Duration{
		{Fraction: 1}:   time.Second,
		{Fraction: 2}:   time.Second * 2,
		{Fraction: 0.1}: time.Millisecond * 100,
	}

	for compared, result := range comparisons {
		timeOffAirWindow := s.timeOffAir(&compared)
		a.So(timeOffAirWindow.Start.Equal(s.End()), should.BeTrue)
		a.So(timeOffAirWindow.Duration, should.Equal, result)
	}

	timeOffAirDurations := []time.Duration{time.Millisecond, time.Second, time.Minute}
	for _, timeOffAirDuration := range timeOffAirDurations {
		toa := &ttnpb.FrequencyPlan_TimeOffAir{Duration: &timeOffAirDuration}
		a.So(s.timeOffAir(toa).Duration, should.Equal, timeOffAirDuration)
	}

	nilTOAWindow := s.timeOffAir(nil)
	a.So(nilTOAWindow.Duration, should.Equal, time.Duration(0))
}
