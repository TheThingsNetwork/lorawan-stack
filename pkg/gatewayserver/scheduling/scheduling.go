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

// Package scheduling offer convenience methods to manage RF packets that must respect scheduling constraints.
package scheduling

import (
	"context"
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/pkg/band"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	errDutyCycleFull = errors.DefineResourceExhausted(
		"duty_cycle_full",
		"duty cycle between {min_frequency} and {max_frequency} full, exceeded quota of {quota}",
	)
	errOverlap                = errors.DefineResourceExhausted("window_overlap", "window overlap")
	errTimeOffAir             = errors.DefineUnavailable("time_off_air_required", "time-off-air constraints prevent scheduling")
	errNoSubBandFound         = errors.DefineNotFound("no_sub_band_found", "no sub-band found for frequency {frequency} Hz")
	errExceededDwellTime      = errors.DefineFailedPrecondition("exceeded_dwell_time", "packet exceeded dwell time restrictions")
	errCouldNotRetrieveFPBand = errors.DefineCorruption("retrieve_fp_band", "could not retrieve the band associated with the frequency plan")
	errNegativeDuration       = errors.DefineInternal("negative_duration", "duration cannot be negative")
)

// Scheduler is an abstraction for an entity that manages the packet's timespans.
type Scheduler interface {
	// ScheduleAt adds the requested timespan to its internal schedule. If, because of its internal constraints (e.g. for duty cycles, not respecting the duty cycle), it returns errScheduleFull. If another error prevents scheduling, it is returned.
	ScheduleAt(s Span, channel uint64) error
	// ScheduleAnytime requires a scheduling window if there is no timestamp constraint.
	ScheduleAnytime(minimum Timestamp, d time.Duration, channel uint64) (Span, error)
	// RegisterEmission that has happened during that timespan, on that specific channel.
	RegisterEmission(s Span, channel uint64) error
}

// FrequencyPlanScheduler returns a scheduler based on the frequency plan, and starts a goroutine for cleanup. The scheduler is based on the dwell time, time off air, and the frequency plan's band. Assumption is made that no two duty cycles on a given band overlap.
func FrequencyPlanScheduler(ctx context.Context, fp ttnpb.FrequencyPlan) (Scheduler, error) {
	scheduler := &frequencyPlanScheduling{
		respectsDwellTime: fp.RespectsDwellTime,
		timeOffAir:        fp.TimeOffAir,
		subBands:          []*subBandScheduling{},
	}

	band, err := band.GetByID(fp.BandID)
	if err != nil {
		return nil, errCouldNotRetrieveFPBand.WithCause(err)
	}
	for _, subBand := range band.BandDutyCycles {
		scheduling := &subBandScheduling{
			dutyCycle:         subBand,
			schedulingWindows: []schedulingWindow{},

			mu: sync.Mutex{},
		}
		scheduler.subBands = append(scheduler.subBands, scheduling)
		go scheduling.bgCleanup(ctx)
	}

	return scheduler, nil
}

type frequencyPlanScheduling struct {
	respectsDwellTime func(isDownlink bool, frequency uint64, duration time.Duration) bool
	timeOffAir        *ttnpb.FrequencyPlan_TimeOffAir

	subBands []*subBandScheduling
}

func (f frequencyPlanScheduling) findSubBand(channel uint64) (*subBandScheduling, error) {
	for _, subBand := range f.subBands {
		if subBand.dutyCycle.Comprises(channel) {
			return subBand, nil
		}
	}

	return nil, errNoSubBandFound.WithAttributes("frequency", channel)
}

func (f frequencyPlanScheduling) ScheduleAt(s Span, channel uint64) error {
	if s.Duration <= 0 {
		return errNegativeDuration
	}

	if !f.respectsDwellTime(true, channel, s.Duration) {
		return errExceededDwellTime.WithAttributes("packet_duration", s.Duration.String())
	}

	subBand, err := f.findSubBand(channel)
	if err != nil {
		return err
	}

	return subBand.ScheduleAt(s, f.timeOffAir)
}

func (f frequencyPlanScheduling) ScheduleAnytime(minimum Timestamp, d time.Duration, channel uint64) (Span, error) {
	if d <= 0 {
		return Span{}, errNegativeDuration
	}

	subBand, err := f.findSubBand(channel)
	if err != nil {
		return Span{}, err
	}

	return subBand.ScheduleAnytime(minimum, d, f.timeOffAir)
}

func (f frequencyPlanScheduling) RegisterEmission(s Span, channel uint64) error {
	if s.Duration <= 0 {
		return errNegativeDuration
	}

	subBand, err := f.findSubBand(channel)
	if err != nil {
		return err
	}

	subBand.RegisterEmission(packetWindow{window: s, timeOffAir: s.timeOffAir(f.timeOffAir)})
	return nil
}
