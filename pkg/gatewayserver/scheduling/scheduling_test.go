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
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/band"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func emptyEUFrequencyPlan() *frequencyplans.FrequencyPlan {
	return &frequencyplans.FrequencyPlan{
		BandID: string(band.EU_863_870),
	}
}

func TestEmptyScheduler(t *testing.T) {
	a := assertions.New(t)

	s, err := scheduling.FrequencyPlanScheduler(test.Context(), emptyEUFrequencyPlan())
	a.So(err, should.BeNil)

	askingTime := scheduling.SystemTime(time.Now().Add(time.Minute))
	askingDuration := time.Second
	_, err = s.ScheduleAnytime(askingTime, askingDuration, 0)
	a.So(err, should.NotBeNil)

	span := scheduling.Span{Start: askingTime, Duration: askingDuration}
	err = s.RegisterEmission(span, 0)
	a.So(err, should.NotBeNil)

	err = s.ScheduleAt(span, 0)
	a.So(err, should.NotBeNil)
}

func TestNotExistingBand(t *testing.T) {
	a := assertions.New(t)

	_, err := scheduling.FrequencyPlanScheduler(test.Context(), &frequencyplans.FrequencyPlan{
		BandID: "not-existing",
	})
	a.So(err, should.NotBeNil)
}

func TestDwellTimeBlocking(t *testing.T) {
	a := assertions.New(t)
	dwellTimeDuration := 400 * time.Millisecond

	dtDownlinks := true
	s, err := scheduling.FrequencyPlanScheduler(test.Context(), &frequencyplans.FrequencyPlan{
		BandID: string(band.EU_863_870),
		DwellTime: frequencyplans.DwellTime{
			Downlinks: &dtDownlinks,
			Duration:  &dwellTimeDuration,
		},
	})
	a.So(err, should.BeNil)

	err = s.ScheduleAt(scheduling.Span{
		Start:    scheduling.SystemTime(time.Now()),
		Duration: 3 * dwellTimeDuration,
	}, 0)
	a.So(errors.IsFailedPrecondition(err), should.BeTrue)
}

func TestScheduleAnytime(t *testing.T) {
	a := assertions.New(t)

	s, err := scheduling.FrequencyPlanScheduler(test.Context(), &frequencyplans.FrequencyPlan{
		BandID: string(band.EU_863_870),
	})
	a.So(err, should.BeNil)

	askingTime := scheduling.SystemTime(time.Now())
	err = s.ScheduleAt(scheduling.Span{Start: askingTime, Duration: time.Microsecond}, 863000000)
	a.So(err, should.BeNil)

	schedule, err := s.ScheduleAnytime(askingTime.Add(time.Hour), time.Microsecond, 863000000)
	a.So(err, should.BeNil)
	a.So(schedule.Start.Equal(askingTime.Add(time.Hour)), should.BeTrue)
	a.So(schedule.Duration, should.Equal, time.Microsecond)

	_, err = s.ScheduleAnytime(askingTime.Add(time.Hour).Add(-1*time.Microsecond), time.Minute, 863000000)
	a.So(err, should.NotBeNil)
}

func TestScheduleAnytime2(t *testing.T) {
	a := assertions.New(t)

	s, err := scheduling.FrequencyPlanScheduler(test.Context(), &frequencyplans.FrequencyPlan{
		BandID: string(band.EU_863_870),
	})
	a.So(err, should.BeNil)

	askingTime := scheduling.SystemTime(time.Now())
	err = s.ScheduleAt(scheduling.Span{Start: askingTime, Duration: time.Microsecond}, 863000000)
	a.So(err, should.BeNil)

	schedule, err := s.ScheduleAnytime(askingTime.Add(time.Hour), time.Microsecond, 863000000)
	a.So(err, should.BeNil)
	a.So(schedule.Start.Equal(askingTime.Add(time.Hour)), should.BeTrue)
	a.So(schedule.Duration, should.Equal, time.Microsecond)

	schedule2, err := s.ScheduleAnytime(askingTime.Add(time.Hour), time.Microsecond, 863000000)
	a.So(err, should.BeNil)
	a.So(askingTime.Add(time.Hour).Add(time.Microsecond).Equal(schedule2.Start), should.BeTrue)
	a.So(schedule2.Duration, should.Equal, time.Microsecond)
}

func TestScheduleAnytimeFullDutyCycle(t *testing.T) {
	a := assertions.New(t)

	s, err := scheduling.FrequencyPlanScheduler(test.Context(), &frequencyplans.FrequencyPlan{
		BandID: string(band.EU_863_870),
	})
	a.So(err, should.BeNil)

	askingTime := scheduling.SystemTime(time.Now())
	scheduleDuration := time.Duration(180 * time.Millisecond)
	err = s.ScheduleAt(scheduling.Span{Start: askingTime, Duration: scheduleDuration}, 863000000)
	a.So(err, should.BeNil)

	schedule, err := s.ScheduleAnytime(askingTime, scheduleDuration, 863000000)
	a.So(err, should.BeNil)
	expectedSchedule2Time := askingTime.Add(5 * time.Minute).Add(-120 * time.Millisecond)
	a.So(schedule.Start.Equal(expectedSchedule2Time), should.BeTrue)
	a.So(schedule.Duration, should.Equal, scheduleDuration)

	schedule, err = s.ScheduleAnytime(askingTime, scheduleDuration, 863000000)
	a.So(err, should.BeNil)
	expectedSchedule3Time := expectedSchedule2Time.Add(5 * time.Minute).Add(-120 * time.Millisecond)
	a.So(schedule.Start.Equal(expectedSchedule3Time), should.BeTrue)
	a.So(schedule.Duration, should.Equal, scheduleDuration)
}

func TestScheduleAnytimeFullDutyCycleAfterRegisteredEmission(t *testing.T) {
	a := assertions.New(t)

	s, err := scheduling.FrequencyPlanScheduler(test.Context(), &frequencyplans.FrequencyPlan{
		BandID: string(band.EU_863_870),
	})
	a.So(err, should.BeNil)

	askingTime := scheduling.SystemTime(time.Now())
	scheduleDuration := time.Duration(180 * time.Millisecond)
	err = s.RegisterEmission(scheduling.Span{Start: askingTime, Duration: scheduleDuration}, 863000000)
	a.So(err, should.BeNil)

	schedule, err := s.ScheduleAnytime(askingTime, scheduleDuration, 863000000)
	a.So(err, should.BeNil)
	expectedSchedule2Time := askingTime.Add(5 * time.Minute).Add(-120 * time.Millisecond)
	a.So(schedule.Start.Equal(expectedSchedule2Time), should.BeTrue)
	a.So(schedule.Duration, should.Equal, scheduleDuration)
}

func TestScheduleFullDutyCycle(t *testing.T) {
	a := assertions.New(t)

	s, err := scheduling.FrequencyPlanScheduler(test.Context(), &frequencyplans.FrequencyPlan{
		BandID: string(band.EU_863_870),
	})
	a.So(err, should.BeNil)

	askingTime := scheduling.SystemTime(time.Now())
	scheduleDuration := time.Duration(180 * time.Millisecond)
	err = s.ScheduleAt(scheduling.Span{Start: askingTime, Duration: scheduleDuration}, 863000000)
	a.So(err, should.BeNil)

	err = s.ScheduleAt(scheduling.Span{Start: askingTime, Duration: scheduleDuration}, 863000000)
	a.So(err, should.NotBeNil)

	err = s.ScheduleAt(scheduling.Span{Start: askingTime.Add(200 * time.Millisecond), Duration: scheduleDuration}, 863000000)
	a.So(err, should.NotBeNil)

	err = s.ScheduleAt(scheduling.Span{Start: askingTime.Add(5 * time.Minute).Add(-120 * time.Millisecond), Duration: scheduleDuration}, 863000000)
	a.So(err, should.BeNil)
}

func TestScheduleOrdering(t *testing.T) {
	a := assertions.New(t)

	s, err := scheduling.FrequencyPlanScheduler(test.Context(), &frequencyplans.FrequencyPlan{
		BandID: string(band.EU_863_870),
	})
	a.So(err, should.BeNil)

	askingTime := scheduling.SystemTime(time.Now().Add(time.Minute))
	scheduleDuration := time.Duration(time.Millisecond)

	schedule, err := s.ScheduleAnytime(askingTime, scheduleDuration, 863000000)
	a.So(err, should.BeNil)
	a.So(schedule.Start.Equal(askingTime), should.BeTrue)

	err = s.ScheduleAt(scheduling.Span{Start: askingTime.Add(time.Second), Duration: scheduleDuration}, 863000000)
	a.So(err, should.BeNil)

	err = s.ScheduleAt(scheduling.Span{Start: askingTime.Add(50 * scheduleDuration), Duration: scheduleDuration}, 863000000)
	a.So(err, should.BeNil)
}

func TestTimeOffAirError(t *testing.T) {
	a := assertions.New(t)

	toa := time.Minute
	s, err := scheduling.FrequencyPlanScheduler(test.Context(), &frequencyplans.FrequencyPlan{
		BandID: string(band.EU_863_870),
		TimeOffAir: frequencyplans.TimeOffAir{
			Duration: toa,
		},
	})
	a.So(err, should.BeNil)

	askingTime := scheduling.SystemTime(time.Now())
	scheduleDuration := time.Duration(60 * time.Millisecond)
	err = s.ScheduleAt(scheduling.Span{Start: askingTime, Duration: scheduleDuration}, 863000000)
	a.So(err, should.BeNil)

	err = s.ScheduleAt(scheduling.Span{Start: askingTime.Add(90 * time.Millisecond), Duration: scheduleDuration}, 863000000)
	a.So(err, should.NotBeNil)
}
