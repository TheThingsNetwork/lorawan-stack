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
	window Window
	toa    Window
}

func (w packetWindow) withTimeOffAir() Window {
	return Window{
		Start:    w.window.Start,
		Duration: w.window.Duration + w.toa.Duration,
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
			if len(s.schedulingWindows) > windowIndex+1 {
				s.schedulingWindows = append(s.schedulingWindows[:windowIndex], s.schedulingWindows[windowIndex+1:]...)
			} else {
				s.schedulingWindows = s.schedulingWindows[:windowIndex]
			}
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
func (s *subBandScheduling) Schedule(w Window, timeOffAir *ttnpb.FrequencyPlan_TimeOffAir) error {
	s.mu.Lock()
	err := s.schedule(w, timeOffAir)
	s.mu.Unlock()
	return err
}

func (s *subBandScheduling) schedule(w Window, timeOffAir *ttnpb.FrequencyPlan_TimeOffAir) error {
	windowWithTOA := packetWindow{window: w, toa: w.timeOffAir(timeOffAir)}

	emissionWindows := []Window{w}

	for _, window := range s.schedulingWindows {
		emissionWindows = append(emissionWindows, window.window)

		if window.window.Overlaps(w) {
			return ErrOverlap
		}
		if window.withTimeOffAir().Overlaps(windowWithTOA.withTimeOffAir()) {
			return ErrTimeOffAir
		}
	}

	precedingWindowsAirtime := windowDurationSum(emissionWindows, w.End().Add(-1*dutyCycleWindow), w.End())
	prolongingWindowsAirtime := windowDurationSum(emissionWindows, w.Start, w.Start.Add(dutyCycleWindow))

	if prolongingWindowsAirtime > s.dutyCycle.MaxAirTimeDuring(dutyCycleWindow) ||
		precedingWindowsAirtime > s.dutyCycle.MaxAirTimeDuring(dutyCycleWindow) {
		return ErrDutyCycleFull
	}

	s.addScheduling(windowWithTOA)

	return nil
}

// AskScheduling requires a scheduling window if there is no time.Time constraint
func (s *subBandScheduling) AskScheduling(minimum time.Time, d time.Duration, timeOffAir *ttnpb.FrequencyPlan_TimeOffAir) (Window, error) {
	var addMinimum = true
	potentialTimings := []time.Time{}
	emissionWindows := []Window{}

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
		w := Window{Start: potentialTiming, Duration: d}
		err := s.schedule(w, timeOffAir)
		if err != nil {
			continue
		}

		return w, nil
	}

	start := firstMomentConsideringDutyCycle(emissionWindows, s.dutyCycle.DutyCycle, potentialTiming, d)
	w := Window{Start: start, Duration: d}
	err := s.schedule(w, timeOffAir)
	return w, err
}

func firstMomentConsideringDutyCycle(windows []Window, dutyCycle float32, minimum time.Time, duration time.Duration) time.Time {
	maxAirtime := time.Duration(dutyCycle * float32(dutyCycleWindow))
	lastWindow := windows[len(windows)-1]

	precedingWindowsAirtime := windowDurationSum(windows, minimum.Add(-1*dutyCycleWindow).Add(duration), minimum.Add(duration)) + duration

	margin := maxAirtime - (precedingWindowsAirtime - duration)
	minimum = lastWindow.Start.Add(-1 * margin).Add(dutyCycleWindow)
	return minimum
}

func createPacketWindow(start time.Time, duration time.Duration, timeOffAir *ttnpb.FrequencyPlan_TimeOffAir) packetWindow {
	window := Window{Start: start, Duration: duration}
	finalEmissionWindow := packetWindow{window: window, toa: window.timeOffAir(timeOffAir)}
	return finalEmissionWindow
}
