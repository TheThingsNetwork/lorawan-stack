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

// Package scheduling implements a packet scheduling that detects and avoids conflicts and enforces regional
// restrictions like duty-cycle and dwell time.
package scheduling

import (
	"context"
	"math"
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/toa"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var (
	// QueueDelay indicates the time the gateway needs to recharge the concentrator between items in the queue.
	// This is a conservative value as implemented in the Semtech UDP Packet Forwarder reference implementation,
	// see https://github.com/Lora-net/packet_forwarder/blob/v4.0.1/lora_pkt_fwd/src/jitqueue.c#L39
	QueueDelay = 30 * time.Millisecond

	// ScheduleTimeShort is a short time to send a downlink message to the gateway before it has to be transmitted.
	// This time is comprised of a lower network latency and QueueDelay. This delay is used for logging a warning message
	// when a message is scheduled shorter before transmission than this value.
	ScheduleTimeShort = 30*time.Millisecond + QueueDelay

	// ScheduleTimeLong is a long time to send a downlink message to the gateway before it has to be transmitted.
	// This time is comprised of a higher network latency and QueueDelay. This delay is used for pseudo-immediate
	// scheduling, see ScheduleAnytime.
	ScheduleTimeLong = 300*time.Millisecond + QueueDelay
)

// NewScheduler instantiates a new Scheduler for the given frequency plan.
func NewScheduler(ctx context.Context, fp *frequencyplans.FrequencyPlan, enforceDutyCycle bool) (*Scheduler, error) {
	s := &Scheduler{
		RolloverClock:     &RolloverClock{},
		respectsDwellTime: fp.RespectsDwellTime,
		timeOffAir:        fp.TimeOffAir,
	}
	if enforceDutyCycle {
		band, err := band.GetByID(fp.BandID)
		if err != nil {
			return nil, err
		}
		for _, subBand := range band.BandDutyCycles {
			sb := NewSubBand(ctx, subBand, s.RolloverClock, nil)
			s.subBands = append(s.subBands, sb)
		}
	} else {
		sb := NewSubBand(ctx, band.DutyCycle{
			MinFrequency: 0,
			MaxFrequency: math.MaxUint64,
			Value:        1,
		}, s.RolloverClock, nil)
		s.subBands = append(s.subBands, sb)
	}
	return s, nil
}

// Scheduler is a packet scheduler that takes time conflicts and sub-band restrictions into account.
type Scheduler struct {
	*RolloverClock
	respectsDwellTime func(isDownlink bool, frequency uint64, duration time.Duration) bool
	timeOffAir        frequencyplans.TimeOffAir
	subBands          []*SubBand
	emissionsMu       sync.Mutex
	emissions         Emissions
}

var errSubBandNotFound = errors.DefineFailedPrecondition("sub_band_not_found", "sub-band not found for frequency `{frequency}` Hz")

func (s *Scheduler) findSubBand(frequency uint64) (*SubBand, error) {
	for _, subBand := range s.subBands {
		if subBand.Comprises(frequency) {
			return subBand, nil
		}
	}
	return nil, errSubBandNotFound.WithAttributes("frequency", frequency)
}

var (
	errDwellTime = errors.DefineFailedPrecondition("dwell_time", "packet exceeds dwell time restriction")
)

func (s *Scheduler) newEmission(payloadSize int, settings ttnpb.TxSettings) (Emission, error) {
	d, err := toa.Compute(payloadSize, settings)
	if err != nil {
		return Emission{}, err
	}
	if !s.respectsDwellTime(true, settings.Frequency, d) {
		return Emission{}, errDwellTime
	}
	var relative ConcentratorTime
	if settings.Time != nil {
		relative = s.RolloverClock.GatewayTime(*settings.Time)
	} else {
		relative = ConcentratorTime(time.Duration(settings.Timestamp) * time.Microsecond)
	}
	return NewEmission(relative, d), nil
}

var errConflict = errors.DefineResourceExhausted("conflict", "scheduling conflict")

// ScheduleAt attempts to schedule the given Tx settings with the given priority.
func (s *Scheduler) ScheduleAt(ctx context.Context, payloadSize int, settings ttnpb.TxSettings, priority ttnpb.TxSchedulePriority) (Emission, error) {
	sb, err := s.findSubBand(settings.Frequency)
	if err != nil {
		return Emission{}, err
	}
	em, err := s.newEmission(payloadSize, settings)
	if err != nil {
		return Emission{}, err
	}
	s.emissionsMu.Lock()
	defer s.emissionsMu.Unlock()
	for _, other := range s.emissions {
		if em.AfterWithOffAir(other, s.timeOffAir)-QueueDelay < 0 && em.BeforeWithOffAir(other, s.timeOffAir)-QueueDelay < 0 {
			return Emission{}, errConflict
		}
	}
	if err := sb.Schedule(em, priority); err != nil {
		return Emission{}, err
	}
	s.emissions = s.emissions.Insert(em)
	return em, nil
}

// ScheduleAnytime attempts to schedule the given Tx settings with the given priority from the time in the settings.
// This method returns the emission.
//
// The scheduler does not support immediate scheduling, i.e. sending a message to the gateway that should be transmitted
// immediately. The reason for this is that this scheduler cannot determine conflicts or enforce duty-cycle when the
// emission time is unknown. Therefore, when the time is set to Immediate, the estimated current concentrator time plus
// ScheduleDelayLong will be used.
func (s *Scheduler) ScheduleAnytime(ctx context.Context, payloadSize int, settings ttnpb.TxSettings, priority ttnpb.TxSchedulePriority) (Emission, error) {
	now := s.RolloverClock.ServerTime(time.Now())
	if settings.Timestamp == 0 && settings.Time == nil {
		settings.Timestamp = uint32((time.Duration(now) + ScheduleTimeLong) / time.Microsecond)
	}
	sb, err := s.findSubBand(settings.Frequency)
	if err != nil {
		return Emission{}, err
	}
	em, err := s.newEmission(payloadSize, settings)
	if err != nil {
		return Emission{}, err
	}
	s.emissionsMu.Lock()
	defer s.emissionsMu.Unlock()
	i := 0
	next := func() ConcentratorTime {
		if len(s.emissions) == 0 {
			// No emissions; schedule at the requested time.
			return em.t
		}
		for i < len(s.emissions)-1 {
			// Find a window between two emissions that does not conflict with either side.
			prevConflicts := s.emissions[i].AfterWithOffAir(em, s.timeOffAir)-QueueDelay < 0
			if prevConflicts {
				// Schedule right after previous to resolve conflict.
				em.t = s.emissions[i].EndsWithOffAir(s.timeOffAir) + ConcentratorTime(QueueDelay)
			}
			nextConflicts := em.BeforeWithOffAir(s.emissions[i+1], s.timeOffAir)-QueueDelay < 0
			if nextConflicts {
				// If it conflicts with the next, try the next window.
				em.t = s.emissions[i+1].EndsWithOffAir(s.timeOffAir) + ConcentratorTime(QueueDelay)
				i++
				continue
			}
			// No conflicts, but advance counter for potential next iteration.
			// A next iteration can be necessary when this emission and priority exceeds a duty-cycle limitation.
			i++
			return em.t
		}
		// No emissions to schedule in between; schedule after last emission.
		return s.emissions[len(s.emissions)-1].EndsWithOffAir(s.timeOffAir) + ConcentratorTime(QueueDelay)
	}
	em, err = sb.ScheduleAnytime(em.d, next, priority)
	if err != nil {
		return Emission{}, err
	}
	if delta := time.Duration(em.Starts() - now); delta < ScheduleTimeShort {
		log.FromContext(ctx).WithField("delta", delta).Warn("The scheduled time is late for transmission")
	}
	s.emissions = s.emissions.Insert(em)
	return em, nil
}
