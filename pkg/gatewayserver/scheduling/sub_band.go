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

package scheduling

import (
	"fmt"
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// DutyCycleWindow is the window in which duty-cycle is enforced.
// A lower value results in balancing capacity in time, while a higher value allows for bursts.
var DutyCycleWindow = 1 * time.Hour

// DutyCycleStyle represents the of duty cycle algorithm to be used by a sub band.
type DutyCycleStyle int

const (
	// DefaultDutyCycleStyle is the default duty cycle style.
	DefaultDutyCycleStyle = DutyCycleStyleSimpleWindow

	// DutyCycleStyleSimpleWindow uses a rolling window in order to compute the usage of a sub band.
	DutyCycleStyleSimpleWindow DutyCycleStyle = iota
	// DutyCycleStyleBlockingWindow uses a rolling window in order to compute the usage of a sub band.
	// The sub band is also blocked for the inverse duration of the non prioritized duty cycle.
	DutyCycleStyleBlockingWindow
)

// DutyCycleCeilings contains the upper limits per schedule priority.
// The limit is a fraction of the duty-cycle.
type DutyCycleCeilings map[ttnpb.TxSchedulePriority]float32

// DefaultDutyCycleCeilings contains the default duty-cycle ceilings per schedule priority.
var DefaultDutyCycleCeilings DutyCycleCeilings = map[ttnpb.TxSchedulePriority]float32{
	ttnpb.TxSchedulePriority_LOWEST:       0.40,
	ttnpb.TxSchedulePriority_LOW:          0.50,
	ttnpb.TxSchedulePriority_BELOW_NORMAL: 0.60,
	ttnpb.TxSchedulePriority_NORMAL:       0.70,
	ttnpb.TxSchedulePriority_ABOVE_NORMAL: 0.80,
	ttnpb.TxSchedulePriority_HIGH:         0.90,
	ttnpb.TxSchedulePriority_HIGHEST:      1.00,
}

// SubBandParameters defines the sub-band frequency bounds and duty-cycle value.
type SubBandParameters struct {
	MinFrequency,
	MaxFrequency uint64
	DutyCycle float32
}

// SubBand tracks the utilization and controls the duty-cycle of a sub-band.
type SubBand struct {
	SubBandParameters
	mu        sync.RWMutex
	clock     Clock
	ceilings  DutyCycleCeilings
	style     DutyCycleStyle
	emissions Emissions
}

// NewSubBand returns a new SubBand considering the given duty-cycle, clock and optionally duty-cycle ceilings.
func NewSubBand(
	params SubBandParameters, clock Clock, ceilings DutyCycleCeilings, style DutyCycleStyle,
) *SubBand {
	if ceilings == nil {
		ceilings = DefaultDutyCycleCeilings
	}
	sb := &SubBand{
		SubBandParameters: params,
		clock:             clock,
		ceilings:          ceilings,
		style:             style,
	}
	if sb.DutyCycle == 0 {
		sb.DutyCycle = 1
	}
	return sb
}

func (sb *SubBand) gc(to ConcentratorTime) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	sb.emissions = sb.emissions.GreaterThan(to)
}

// Comprises returns whether the given frequency falls in the sub-band.
func (sb SubBandParameters) Comprises(frequency uint64) bool {
	return frequency >= sb.MinFrequency && frequency <= sb.MaxFrequency
}

// sum returns the total emission durations in the given window.
// This method requires the read lock to be held.
func (sb *SubBand) sum(from, to ConcentratorTime) time.Duration {
	total := time.Duration(0)
	for _, em := range sb.emissions {
		total += em.Within(from, to)
	}
	return total
}

// DutyCycleUtilization returns the utilization as a fraction of the available duty-cycle.
func (sb *SubBand) DutyCycleUtilization() float32 {
	now, ok := sb.clock.FromServerTime(time.Now())
	if !ok {
		return 0
	}
	sb.mu.RLock()
	val := sb.sum(now-ConcentratorTime(DutyCycleWindow), now)
	sb.mu.RUnlock()
	return float32(val) / float32(DutyCycleWindow)
}

// prioritizedDutyCycle returns the duty-cycle given the scheduling priority.
// This is calculated as the available duty-cycle for the sub-band times the priority ceiling.
func (sb *SubBand) prioritizedDutyCycle(p ttnpb.TxSchedulePriority) float32 {
	ceiling := float32(1)
	if c, ok := sb.ceilings[p]; ok {
		ceiling = c
	}
	return sb.DutyCycle * ceiling
}

var (
	errDutyCycle = errors.DefineResourceExhausted(
		"duty_cycle",
		"utilization `{used}%` would be higher than the available `{usable}%` for priority `{priority}`",
	)
	errBlocked = errors.DefineResourceExhausted(
		"blocked",
		"sub band is blocked for `{duration}`",
	)
)

// checkDutyCycle verifies if the emission complies with the duty cycle limitations, based on the style.
//
// For the simple window style, it verifies if the emission will not exceed the usable amount of the duty
// cycle in the [t + toa - window, t + toa] and [t, t + window] windows, where `t` is the start of the
// emission and `toa` is the duration of the emission.
//
// For the blocking window style, in addition to the above mentioned check it verifies if the inverse of
// the duty cycle has passed from the last emission.
// Given a duty cycle of p%, and an emission of d time units at timestamp t, the duty cycle check will
// allow the next transmission to occur only after the t + d * (100%/p%) timestamp.
//
// This method requires that the read lock is held.
func (sb *SubBand) checkDutyCycle(em Emission, p ttnpb.TxSchedulePriority) error {
	usable := sb.prioritizedDutyCycle(p)
	for _, to := range []ConcentratorTime{em.Ends(), em.t + ConcentratorTime(DutyCycleWindow)} {
		used := float32(sb.sum(to-ConcentratorTime(DutyCycleWindow), to)+em.d) / float32(DutyCycleWindow)
		if used <= usable {
			continue
		}
		return errDutyCycle.WithAttributes(
			"used", fmt.Sprintf("%.1f", used*100),
			"usable", fmt.Sprintf("%.1f", usable*100),
			"priority", fmt.Sprintf("%v", p),
		)
	}
	switch sb.style {
	case DutyCycleStyleSimpleWindow:
	case DutyCycleStyleBlockingWindow:
		if len(sb.emissions) == 0 {
			return nil
		}
		lastEmission := sb.emissions[len(sb.emissions)-1]
		// NOTE: The priority is intentionally elided here, as the blocking algorithm does not consider
		// the emission priority for duty cycle purposes.
		blockedUntil := lastEmission.t + ConcentratorTime(lastEmission.d*time.Duration(1.0/sb.DutyCycle))
		if em.t < blockedUntil {
			return errBlocked.WithAttributes(
				"duration", time.Duration(blockedUntil-em.t),
			)
		}
	default:
		panic("unreachable")
	}
	return nil
}

// Schedule schedules the given emission with the priority.
// If there is no time available due to duty-cycle limitations, an error with code ResourceExhausted is returned.
func (sb *SubBand) Schedule(em Emission, p ttnpb.TxSchedulePriority) error {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	if sb.DutyCycle >= 1 {
		sb.emissions = sb.emissions.Insert(em)
		return nil
	}
	if err := sb.checkDutyCycle(em, p); err != nil {
		return err
	}
	sb.emissions = sb.emissions.Insert(em)
	return nil
}

// ScheduleAnytime schedules the given duration at a time when there is availability by accounting for duty-cycle.
// The given next callback should return the next option that does not conflict with other scheduled downlinks.
// If there is no duty-cycle limitation, this method returns the first option.
func (sb *SubBand) ScheduleAnytime(d time.Duration, next func() ConcentratorTime, p ttnpb.TxSchedulePriority) (Emission, error) {
	sb.mu.Lock()
	defer sb.mu.Unlock()
	em := NewEmission(next(), d)
	if sb.DutyCycle >= 1 {
		sb.emissions = sb.emissions.Insert(em)
		return em, nil
	}
	usable := sb.prioritizedDutyCycle(p)
	used := float32(em.d) / float32(DutyCycleWindow)
	if used > usable {
		return Emission{}, errDutyCycle.WithAttributes(
			"used", fmt.Sprintf("%.1f", used*100),
			"usable", fmt.Sprintf("%.1f", usable*100),
			"priority", fmt.Sprintf("%v", p),
		)
	}
	for {
		if err := sb.checkDutyCycle(em, p); err == nil {
			break
		}
		if t := next(); t != em.t {
			em.t = t
			continue
		}
		// The caller has no later option; find the last emission after which we consider the duty-cycle window.
		for i := len(sb.emissions) - 1; i >= 0; i-- {
			other := sb.emissions[i]
			used += float32(other.d) / float32(DutyCycleWindow)
			if used > usable {
				em.t = other.Ends() + ConcentratorTime(DutyCycleWindow) - ConcentratorTime(em.d)
				break
			}
		}
		break
	}
	sb.emissions = sb.emissions.Insert(em)
	return em, nil
}

// HasOverlap checks if the two sub bands have an overlap.
func (sb *SubBand) HasOverlap(subBand *SubBand) bool {
	return subBand.MaxFrequency > sb.MinFrequency && subBand.MinFrequency < sb.MaxFrequency ||
		subBand.MinFrequency < sb.MaxFrequency && subBand.MaxFrequency > sb.MaxFrequency
}

// IsIdentical checks if the two sub bands are identical.
func (sb *SubBand) IsIdentical(subBand *SubBand) bool {
	return sb.MinFrequency == subBand.MinFrequency && sb.MaxFrequency == subBand.MaxFrequency
}
