// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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

	testingTime := time.Now()
	w := Window{
		Start:    testingTime,
		Duration: time.Second,
	}

	a.So(w.End(), should.Equal, testingTime.Add(time.Second))
}

func TestWindowContains(t *testing.T) {
	a := assertions.New(t)

	testingTime := time.Now()
	w := Window{
		Start:    testingTime,
		Duration: time.Second,
	}

	comparisons := map[time.Time]bool{
		testingTime: true,

		testingTime.Add(time.Second):      true,
		testingTime.Add(time.Microsecond): true,

		testingTime.Add(2 * time.Second):       false,
		testingTime.Add(-1 * time.Millisecond): false,
	}

	for compared, result := range comparisons {
		a.So(w.Contains(compared), should.Equal, result)
	}
}

func TestPrecedingComparison(t *testing.T) {
	a := assertions.New(t)

	testingTime := time.Now()
	w := Window{
		Start:    testingTime,
		Duration: time.Second,
	}

	comparisons := map[Window]bool{
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
		a.So(compared.Precedes(w), should.Equal, result)
	}
}

func TestIsProlongedByComparison(t *testing.T) {
	a := assertions.New(t)

	testingTime := time.Now()
	w := Window{
		Start:    testingTime,
		Duration: time.Second,
	}

	comparisons := map[Window]bool{
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
		a.So(w.IsProlongedBy(compared), should.Equal, result)
	}
}

func TestOverlap(t *testing.T) {
	a := assertions.New(t)

	testingTime := time.Now()
	w := Window{
		Start:    testingTime,
		Duration: time.Second,
	}

	comparisons := map[Window]bool{
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
		a.So(w.Overlaps(compared), should.Equal, result)
	}
}

func TestTimeOffAir(t *testing.T) {
	a := assertions.New(t)

	testingTime := time.Now()
	w := Window{
		Start:    testingTime,
		Duration: time.Second,
	}

	comparisons := map[ttnpb.FrequencyPlan_TimeOffAir]time.Duration{
		{Fraction: 1}:   time.Second,
		{Fraction: 2}:   time.Second * 2,
		{Fraction: 0.1}: time.Millisecond * 100,
	}

	for compared, result := range comparisons {
		toaWindow := w.timeOffAir(&compared)
		a.So(toaWindow.Start, should.Equal, w.End())
		a.So(toaWindow.Duration, should.Equal, result)
	}

	toaDurations := []time.Duration{time.Millisecond, time.Second, time.Minute}
	for _, toaDuration := range toaDurations {
		toa := &ttnpb.FrequencyPlan_TimeOffAir{Duration: &toaDuration}
		a.So(w.timeOffAir(toa).Duration, should.Equal, toaDuration)
	}

	nilTOAWindow := w.timeOffAir(nil)
	a.So(nilTOAWindow.Duration, should.Equal, time.Duration(0))
}
