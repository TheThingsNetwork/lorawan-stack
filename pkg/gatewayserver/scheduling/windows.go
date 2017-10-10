// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package scheduling

import (
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// Window is a time window requested for usage by the consumer entity.
type Window struct {
	Start    time.Time
	Duration time.Duration
}

// End returns the timestamp at which the window ends
func (w Window) End() time.Time {
	return w.Start.Add(w.Duration)
}

// Precedes returns true if a portion of the window is located before the window passed as a parameter
func (w Window) Precedes(newW Window) bool {
	return w.Start.Before(newW.Start)
}

// Contains returns true if the given time is contained between the beginning and the end of this window
func (w Window) Contains(ts time.Time) bool {
	if ts.Before(w.Start) {
		return false
	}
	if ts.After(w.End()) {
		return false
	}
	return true
}

// IsProlongedBy returns true if after the window ends, there is still a portion of the window passed as parameter
func (w Window) IsProlongedBy(newW Window) bool {
	return w.End().Before(newW.End())
}

// Overlaps returns true if the two time windows overlap
func (w Window) Overlaps(newW Window) bool {
	if newW.End().Before(w.Start) || newW.End() == w.Start {
		return false
	}

	if w.End().Before(newW.Start) || w.End() == newW.Start {
		return false
	}

	return true
}

func filterWithinInterval(windows []Window, start, end time.Time) []Window {
	filteredWindows := []Window{}

	for _, window := range windows {
		if window.End().Before(start) || window.Start.After(end) {
			continue
		}

		filteredWindows = append(filteredWindows, window)
	}

	return filteredWindows
}

// windowDurationSum takes an array of non-overlapping windows, a start and an end time, and determines the sum of the duration of these windows within the interval determined by start and end.
func windowDurationSum(windows []Window, start, end time.Time) time.Duration {
	var duration time.Duration

	windows = filterWithinInterval(windows, start, end)
	for _, window := range windows {
		windowInterval := struct {
			start, end time.Time
		}{start: window.Start, end: window.End()}

		if windowInterval.start.Before(start) {
			windowInterval.start = start
		}
		if windowInterval.end.After(end) {
			windowInterval.end = end
		}

		duration = duration + windowInterval.end.Sub(windowInterval.start)
	}

	return duration
}

func (w Window) timeOffAir(timeOffAir *ttnpb.FrequencyPlan_TimeOffAir) (timeOffAirWindow Window) {
	timeOffAirWindow = Window{Start: w.End(), Duration: 0}

	if timeOffAir == nil {
		return
	}

	timeOffAirWindow.Duration = time.Duration(timeOffAir.Fraction * float32(w.Duration))
	if timeOffAir.Duration != nil {
		if *timeOffAir.Duration > timeOffAirWindow.Duration {
			timeOffAirWindow.Duration = *timeOffAir.Duration
		}
	}

	return
}
