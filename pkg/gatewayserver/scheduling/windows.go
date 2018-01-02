// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package scheduling

import (
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// Span is a time window requested for usage by the consumer entity.
type Span struct {
	Start    time.Time
	Duration time.Duration
}

// End returns the timestamp at which the timespan ends
func (s Span) End() time.Time {
	return s.Start.Add(s.Duration)
}

// Precedes returns true if a portion of the timespan is located before the span passed as a parameter
func (s Span) Precedes(other Span) bool {
	return s.Start.Before(other.Start)
}

// Contains returns true if the given time is contained between the beginning and the end of this timespan
func (s Span) Contains(ts time.Time) bool {
	if ts.Before(s.Start) {
		return false
	}
	if ts.After(s.End()) {
		return false
	}
	return true
}

// IsProlongedBy returns true if after the span ends, there is still a portion of the span passed as parameter
func (s Span) IsProlongedBy(other Span) bool {
	return s.End().Before(other.End())
}

// Overlaps returns true if the two timespans overlap
func (s Span) Overlaps(other Span) bool {
	if other.End().Before(s.Start) || other.End() == s.Start {
		return false
	}

	if s.End().Before(other.Start) || s.End() == other.Start {
		return false
	}

	return true
}

func filterWithinInterval(spans []Span, start, end time.Time) []Span {
	filteredSpans := []Span{}

	for _, span := range spans {
		if span.End().Before(start) || span.Start.After(end) {
			continue
		}

		filteredSpans = append(filteredSpans, span)
	}

	return filteredSpans
}

// spanDurationSum takes an array of non-overlapping spans, a start and an end time, and determines the sum of the duration of these spans within the interval determined by start and end.
func spanDurationSum(spans []Span, start, end time.Time) time.Duration {
	var duration time.Duration

	spans = filterWithinInterval(spans, start, end)
	for _, span := range spans {
		spanStart := span.Start
		spanEnd := span.End()

		if spanStart.Before(start) {
			spanStart = start
		}
		if spanEnd.After(end) {
			spanEnd = end
		}

		duration = duration + spanStart.Sub(spanEnd)
	}

	return duration
}

func (s Span) timeOffAir(timeOffAir *ttnpb.FrequencyPlan_TimeOffAir) (timeOffAirSpan Span) {
	timeOffAirSpan = Span{Start: s.End(), Duration: 0}

	if timeOffAir == nil {
		return
	}

	timeOffAirSpan.Duration = time.Duration(timeOffAir.Fraction * float32(s.Duration))
	if timeOffAir.Duration != nil && *timeOffAir.Duration > timeOffAirSpan.Duration {
		timeOffAirSpan.Duration = *timeOffAir.Duration
	}

	return
}
