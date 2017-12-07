// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package scheduling

import (
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/band"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

const dutyCycleWindow = 5 * time.Minute

type packetWindow struct {
	window     Span
	timeOffAir Span
}

func (w packetWindow) withTimeOffAir() Span {
	return Span{
		Start:    w.window.Start,
		Duration: w.window.Duration + w.timeOffAir.Duration,
	}
}

type subBandScheduling struct {
	dutyCycle         band.DutyCycle
	schedulingWindows []packetWindow

	mu sync.Mutex
}

func (s *subBandScheduling) removeOldScheduling(w packetWindow) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for windowIndex, window := range s.schedulingWindows {
		if w == window {
			s.schedulingWindows = append(s.schedulingWindows[:windowIndex], s.schedulingWindows[windowIndex+1:]...)
			return
		}
	}
}

func (s *subBandScheduling) swoopOldScheduling(w packetWindow) {
	time.Sleep(w.window.End().Sub(time.Now()) + dutyCycleWindow)
	s.removeOldScheduling(w)
}

func (s *subBandScheduling) addScheduling(w packetWindow) {
	go s.swoopOldScheduling(w)

	for i, window := range s.schedulingWindows {
		if w.window.Precedes(window.window) {
			s.schedulingWindows = append(s.schedulingWindows[:i], append([]packetWindow{w}, s.schedulingWindows[i:]...)...)
			return
		}
	}
	s.schedulingWindows = append(s.schedulingWindows, w)
}

// Schedule adds the requested time window to its internal schedule. If, because of its internal constraints (e.g. for duty cycles, not respecting the duty cycle), it returns ErrScheduleFull. If another error prevents scheduling, it is returned.
func (s *subBandScheduling) Schedule(w Span, timeOffAir *ttnpb.FrequencyPlan_TimeOffAir) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.schedule(w, timeOffAir)
	return err
}

func (s *subBandScheduling) schedule(w Span, timeOffAir *ttnpb.FrequencyPlan_TimeOffAir) error {
	windowWithTimeOffAir := packetWindow{window: w, timeOffAir: w.timeOffAir(timeOffAir)}

	emissionWindows := []Span{w}

	for _, scheduledWindow := range s.schedulingWindows {
		emissionWindows = append(emissionWindows, scheduledWindow.window)

		if scheduledWindow.window.Overlaps(w) {
			return ErrOverlap
		}
		if scheduledWindow.withTimeOffAir().Overlaps(windowWithTimeOffAir.withTimeOffAir()) {
			return ErrTimeOffAir
		}
	}

	precedingWindowsAirtime := spanDurationSum(emissionWindows, w.End().Add(-1*dutyCycleWindow), w.End())
	prolongingWindowsAirtime := spanDurationSum(emissionWindows, w.Start, w.Start.Add(dutyCycleWindow))

	if prolongingWindowsAirtime > s.dutyCycle.MaxEmissionDuring(dutyCycleWindow) ||
		precedingWindowsAirtime > s.dutyCycle.MaxEmissionDuring(dutyCycleWindow) {
		return ErrDutyCycleFull
	}

	s.addScheduling(windowWithTimeOffAir)

	return nil
}

// ScheduleFlexible requires a scheduling window if there is no time.Time constraint
func (s *subBandScheduling) ScheduleFlexible(minimum time.Time, d time.Duration, timeOffAir *ttnpb.FrequencyPlan_TimeOffAir) (Span, error) {
	var addMinimum = true
	potentialTimings := []time.Time{}
	emissionWindows := []Span{}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, window := range s.schedulingWindows {
		emissionWindows = append(emissionWindows, window.window)
		windowWithTimeOffAir := window.withTimeOffAir()
		if addMinimum && windowWithTimeOffAir.Contains(minimum) {
			addMinimum = false
		}

		windowEnd := windowWithTimeOffAir.End()
		if windowEnd.After(minimum) {
			potentialTimings = append(potentialTimings, windowEnd)
		}
	}

	if addMinimum {
		potentialTimings = append(potentialTimings, minimum)
	}

	var potentialTiming time.Time
	for _, potentialTiming = range potentialTimings {
		w := Span{Start: potentialTiming, Duration: d}
		err := s.schedule(w, timeOffAir)
		if err != nil {
			continue
		}

		return w, nil
	}

	start := firstMomentConsideringDutyCycle(emissionWindows, s.dutyCycle.DutyCycle, potentialTiming, d)
	w := Span{Start: start, Duration: d}
	err := s.schedule(w, timeOffAir)
	return w, err
}

func firstMomentConsideringDutyCycle(spans []Span, dutyCycle float32, minimum time.Time, duration time.Duration) time.Time {
	maxAirtime := time.Duration(dutyCycle * float32(dutyCycleWindow))
	lastWindow := spans[len(spans)-1]

	precedingWindowsAirtime := spanDurationSum(spans, minimum.Add(-1*dutyCycleWindow).Add(duration), minimum.Add(duration)) + duration

	margin := maxAirtime - (precedingWindowsAirtime - duration)
	minimum = lastWindow.Start.Add(-1 * margin).Add(dutyCycleWindow)
	return minimum
}

func createPacketWindow(start time.Time, duration time.Duration, timeOffAir *ttnpb.FrequencyPlan_TimeOffAir) packetWindow {
	window := Span{Start: start, Duration: duration}
	finalEmissionWindow := packetWindow{window: window, timeOffAir: window.timeOffAir(timeOffAir)}
	return finalEmissionWindow
}
