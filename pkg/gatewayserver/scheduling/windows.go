// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package scheduling

import (
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

type realTime time.Time

func (r realTime) Add(d time.Duration) Timestamp { return realTime(time.Time(r).Add(d)) }
func (r realTime) Sub(s Timestamp) time.Duration { return time.Time(r).Sub(time.Time(s.(realTime))) }
func (r realTime) After(s Timestamp) bool        { return time.Time(r).After(time.Time(s.(realTime))) }
func (r realTime) Before(s Timestamp) bool       { return time.Time(r).Before(time.Time(s.(realTime))) }
func (r realTime) Equal(s Timestamp) bool        { return time.Time(r) == time.Time(s.(realTime)) }

// FromSystemTimestamp returns a Timestamp value from a system timestamp.
func FromSystemTimestamp(t time.Time) Timestamp { return realTime(t) }

type concentratorTimestamp uint64

func (c concentratorTimestamp) Add(d time.Duration) Timestamp {
	return concentratorTimestamp(uint64(c) + uint64(d))
}
func (c concentratorTimestamp) Sub(s Timestamp) time.Duration {
	return time.Duration(uint64(c) - uint64(s.(concentratorTimestamp)))
}
func (c concentratorTimestamp) After(s Timestamp) bool {
	return uint64(c) > uint64(s.(concentratorTimestamp))
}
func (c concentratorTimestamp) Before(s Timestamp) bool {
	return uint64(c) < uint64(s.(concentratorTimestamp))
}
func (c concentratorTimestamp) Equal(s Timestamp) bool {
	return uint64(c) == uint64(s.(concentratorTimestamp))
}

// FromConcentratorTimestamp returns a Timestamp value from a concentrator timestamp.
func FromConcentratorTimestamp(t uint64) Timestamp { return concentratorTimestamp(t) }

// Timestamp represents a temporal value. Using two Timestamp value of different origins (concentrator, system) in the same scheduler will result in a panic.
type Timestamp interface {
	Add(time.Duration) Timestamp
	Sub(Timestamp) time.Duration

	After(Timestamp) bool
	Before(Timestamp) bool
	Equal(Timestamp) bool
}

// Span is a time window requested for usage by the consumer entity.
type Span struct {
	Start    Timestamp
	Duration time.Duration
}

// End returns the timestamp at which the timespan ends
func (s Span) End() Timestamp {
	return s.Start.Add(s.Duration)
}

// StartsBefore returns true if a portion of the timespan is located before the span passed as a parameter.
func (s Span) StartsBefore(other Span) bool {
	return s.Start.Before(other.Start)
}

// Contains returns true if the given time is contained between the beginning and the end of this timespan
func (s Span) Contains(ts Timestamp) bool {
	if ts.Before(s.Start) {
		return false
	}
	if ts.After(s.End()) {
		return false
	}
	return true
}

// IsProlongedBy returns true if after the span ends, there is still a portion of the span passed as parameter.
func (s Span) IsProlongedBy(other Span) bool {
	return s.End().Before(other.End())
}

// Overlaps returns true if the two timespans overlap.
func (s Span) Overlaps(other Span) bool {
	if other.End().Before(s.Start) || other.End() == s.Start {
		return false
	}

	if s.End().Before(other.Start) || s.End() == other.Start {
		return false
	}

	return true
}

func filterWithinInterval(spans []Span, start, end Timestamp) []Span {
	filteredSpans := []Span{}

	for _, span := range spans {
		if span.End().Before(start) || span.Start.After(end) {
			continue
		}

		filteredSpans = append(filteredSpans, span)
	}

	return filteredSpans
}

// sumWithinInterval takes an array of non-overlapping spans, a start and an end time, and determines the sum of the duration of these spans within the interval determined by start and end.
func sumWithinInterval(spans []Span, start, end Timestamp) time.Duration {
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

		duration = duration + spanEnd.Sub(spanStart)
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
