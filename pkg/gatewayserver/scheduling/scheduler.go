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
	"context"
	"math"
	"runtime/trace"
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
	// This time is comprised of a lower network latency and QueueDelay. This delay is used to block scheduling when the
	// schedule time to the estimated concentrator time is less than this value, see ScheduleAt.
	ScheduleTimeShort = 100*time.Millisecond + QueueDelay

	// ScheduleTimeLong is a long time to send a downlink message to the gateway before it has to be transmitted.
	// This time is comprised of a higher network latency and QueueDelay. This delay is used for pseudo-immediate
	// scheduling, see ScheduleAnytime.
	ScheduleTimeLong = 500*time.Millisecond + QueueDelay
)

// TimeSource is a source for getting a current time.
type TimeSource interface {
	Now() time.Time
}

type systemTimeSource struct {
}

// Now implements TimeSource.
func (systemTimeSource) Now() time.Time { return time.Now() }

// SystemTimeSource is a TimeSource that uses the local system time.
var SystemTimeSource = &systemTimeSource{}

// RTTs provides round-trip times.
type RTTs interface {
	Stats() (min, max, median time.Duration, count int)
}

var errFrequencyPlansTimeOffAir = errors.DefineInvalidArgument("frequency_plans_timeoffair", "frequency plans must have the same TimeOffAir value")
var errFrequencyPlansOverlapSubBand = errors.DefineInvalidArgument("frequency_plans_overlap_subband", "frequency plans must not have overlapping sub bands")

// NewScheduler instantiates a new Scheduler for the given frequency plan.
// If no time source is specified, the system time is used.
func NewScheduler(ctx context.Context, fps []*frequencyplans.FrequencyPlan, enforceDutyCycle bool, scheduleAnytimeDelay *time.Duration, timeSource TimeSource) (*Scheduler, error) {
	logger := log.FromContext(ctx)
	if timeSource == nil {
		timeSource = SystemTimeSource
	}

	if scheduleAnytimeDelay == nil || *scheduleAnytimeDelay == 0 {
		scheduleAnytimeDelay = &ScheduleTimeLong
	} else if *scheduleAnytimeDelay < ScheduleTimeShort {
		logger.WithFields(log.Fields(
			"minimum", ScheduleTimeShort,
			"requested", *scheduleAnytimeDelay,
		)).Info("Requested scheduling delay is too small")
		scheduleAnytimeDelay = &ScheduleTimeShort
	}

	for i := 0; i < len(fps)-1; i++ {
		if fps[i].TimeOffAir != fps[i+1].TimeOffAir {
			return nil, errFrequencyPlansTimeOffAir
		}
	}

	toa := fps[0].TimeOffAir
	if toa.Duration < QueueDelay {
		toa.Duration = QueueDelay
	}

	s := &Scheduler{
		clock:                &RolloverClock{},
		timeOffAir:           toa,
		fps:                  fps,
		timeSource:           timeSource,
		scheduleAnytimeDelay: *scheduleAnytimeDelay,
	}
	if enforceDutyCycle {
		for _, fp := range fps {
			if subBands := fp.SubBands; len(subBands) > 0 {
				for _, subBand := range subBands {
					params := SubBandParameters{
						MinFrequency: subBand.MinFrequency,
						MaxFrequency: subBand.MaxFrequency,
						DutyCycle:    subBand.DutyCycle,
					}
					sb := NewSubBand(ctx, params, s.clock, nil)
					var isIdentical bool
					for _, subBand := range s.subBands {
						if subBand.IsIdentical(sb) {
							isIdentical = true
							break
						}
						if subBand.HasOverlap(sb) {
							return nil, errFrequencyPlansOverlapSubBand
						}
					}
					if !isIdentical {
						s.subBands = append(s.subBands, sb)
					}
				}
			} else {
				band, err := band.GetByID(fp.BandID)
				if err != nil {
					return nil, err
				}
				for _, subBand := range band.SubBands {
					params := SubBandParameters{
						MinFrequency: subBand.MinFrequency,
						MaxFrequency: subBand.MaxFrequency,
						DutyCycle:    subBand.DutyCycle,
					}
					sb := NewSubBand(ctx, params, s.clock, nil)
					var isIdentical bool
					for _, subBand := range s.subBands {
						if subBand.IsIdentical(sb) {
							isIdentical = true
							break
						}
						if subBand.HasOverlap(sb) {
							return nil, errFrequencyPlansOverlapSubBand
						}
					}
					if !isIdentical {
						s.subBands = append(s.subBands, sb)
					}
				}
			}
		}
	} else {
		noDutyCycleParams := SubBandParameters{
			MinFrequency: 0,
			MaxFrequency: math.MaxUint64,
		}
		sb := NewSubBand(ctx, noDutyCycleParams, s.clock, nil)
		s.subBands = append(s.subBands, sb)
	}
	return s, nil
}

// Scheduler is a packet scheduler that takes time conflicts and sub-band restrictions into account.
type Scheduler struct {
	clock                *RolloverClock
	fps                  []*frequencyplans.FrequencyPlan
	timeOffAir           frequencyplans.TimeOffAir
	timeSource           TimeSource
	subBands             []*SubBand
	mu                   sync.RWMutex
	emissions            Emissions
	scheduleAnytimeDelay time.Duration
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

var errDwellTime = errors.DefineFailedPrecondition("dwell_time", "packet exceeds dwell time restriction")

func (s *Scheduler) newEmission(payloadSize int, settings ttnpb.TxSettings, starts ConcentratorTime) (Emission, error) {
	d, err := toa.Compute(payloadSize, settings)
	if err != nil {
		return Emission{}, err
	}
	for _, fp := range s.fps {
		if fp.RespectsDwellTime(true, settings.Frequency, d) {
			return NewEmission(starts, d), nil
		}
	}
	return Emission{}, errDwellTime
}

// NoOfSubBands returns the number of sub bands in the scheduler
func (s *Scheduler) NoOfSubBands() int {
	return len(s.subBands)
}

var (
	errConflict              = errors.DefineResourceExhausted("conflict", "scheduling conflict")
	errTooLate               = errors.DefineFailedPrecondition("too_late", "too late to transmission scheduled time (delta is `{delta}`)")
	errNoClockSync           = errors.DefineUnavailable("no_clock_sync", "no clock sync")
	errNoAbsoluteGatewayTime = errors.DefineAborted("no_absolute_gateway_time", "no absolute gateway time")
	errNoServerTime          = errors.DefineAborted("no_server_time", "no server time")
)

// ScheduleAt attempts to schedule the given Tx settings with the given priority.
// If there are round-trip times available, the maximum value will be used instead of ScheduleTimeShort.
func (s *Scheduler) ScheduleAt(ctx context.Context, payloadSize int, settings ttnpb.TxSettings, rtts RTTs, priority ttnpb.TxSchedulePriority) (Emission, error) {
	defer trace.StartRegion(ctx, "schedule transmission").End()

	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.clock.IsSynced() {
		return Emission{}, errNoClockSync
	}
	minScheduleTime := ScheduleTimeShort
	var medianRTT *time.Duration
	if rtts != nil {
		if _, max, median, n := rtts.Stats(); n > 0 {
			minScheduleTime = max + QueueDelay
			medianRTT = &median
		}
	}
	var starts ConcentratorTime
	now, ok := s.clock.FromServerTime(s.timeSource.Now())
	if settings.Time != nil {
		var ok bool
		starts, ok = s.clock.FromGatewayTime(*settings.Time)
		if !ok {
			if medianRTT != nil {
				serverTime, ok := s.clock.FromServerTime(*settings.Time)
				if !ok {
					return Emission{}, errNoServerTime
				}
				starts = serverTime - ConcentratorTime(*medianRTT/2)
			} else {
				return Emission{}, errNoAbsoluteGatewayTime
			}
		}
		// Assume that the absolute time is the time of arrival, not time of transmission.
		toa, err := toa.Compute(payloadSize, settings)
		if err != nil {
			return Emission{}, err
		}
		starts -= ConcentratorTime(toa)
	} else {
		starts = s.clock.FromTimestampTime(settings.Timestamp)
	}
	if ok {
		if delta := time.Duration(starts - now); delta < minScheduleTime {
			return Emission{}, errTooLate.WithAttributes("delta", delta)
		}
	}
	sb, err := s.findSubBand(settings.Frequency)
	if err != nil {
		return Emission{}, err
	}
	em, err := s.newEmission(payloadSize, settings, starts)
	if err != nil {
		return Emission{}, err
	}
	for _, other := range s.emissions {
		if em.OverlapsWithOffAir(other, s.timeOffAir) {
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
// If there are round-trip times available, the maximum value will be used instead of ScheduleTimeShort.
// This method returns the emission.
//
// The scheduler does not support immediate scheduling, i.e. sending a message to the gateway that should be transmitted
// immediately. The reason for this is that this scheduler cannot determine conflicts or enforce duty-cycle when the
// emission time is unknown. Therefore, when the time is set to Immediate, the estimated current concentrator time plus
// ScheduleDelayLong will be used.
func (s *Scheduler) ScheduleAnytime(ctx context.Context, payloadSize int, settings ttnpb.TxSettings, rtts RTTs, priority ttnpb.TxSchedulePriority) (Emission, error) {
	defer trace.StartRegion(ctx, "schedule transmission at any time").End()

	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.clock.IsSynced() {
		return Emission{}, errNoClockSync
	}
	minScheduleTime := ScheduleTimeShort
	if rtts != nil {
		if _, max, _, n := rtts.Stats(); n > 0 {
			minScheduleTime = max + QueueDelay
		}
	}
	var starts ConcentratorTime
	now, ok := s.clock.FromServerTime(s.timeSource.Now())
	if !ok {
		return Emission{}, errNoServerTime
	}
	if settings.Timestamp == 0 {
		starts = now + ConcentratorTime(s.scheduleAnytimeDelay)
		settings.Timestamp = uint32(time.Duration(starts) / time.Microsecond)
	} else {
		starts = s.clock.FromTimestampTime(settings.Timestamp)
		if delta := minScheduleTime - time.Duration(starts-now); delta > 0 {
			starts += ConcentratorTime(delta)
			settings.Timestamp += uint32(delta / time.Microsecond)
		}
	}
	sb, err := s.findSubBand(settings.Frequency)
	if err != nil {
		return Emission{}, err
	}
	em, err := s.newEmission(payloadSize, settings, starts)
	if err != nil {
		return Emission{}, err
	}
	i := 0
	next := func() ConcentratorTime {
		if len(s.emissions) == 0 {
			// No emissions; schedule at the requested time.
			return em.t
		}
		for i < len(s.emissions)-1 {
			// Find a window between two emissions that does not conflict with either side.
			if em.OverlapsWithOffAir(s.emissions[i], s.timeOffAir) {
				// Schedule right after previous to resolve conflict.
				em.t = s.emissions[i].EndsWithOffAir(s.timeOffAir)
			}
			if em.OverlapsWithOffAir(s.emissions[i+1], s.timeOffAir) {
				// Schedule right after next to resolve conflict.
				em.t = s.emissions[i+1].EndsWithOffAir(s.timeOffAir)
				i++
				continue
			}
			// No conflicts, but advance counter for potential next iteration.
			// A next iteration can be necessary when this emission and priority exceeds a duty-cycle limitation.
			i++
			return em.t
		}
		// No emissions to schedule in between; schedule at timestamp or last transmission, whichever comes first.
		afterLast := s.emissions[len(s.emissions)-1].EndsWithOffAir(s.timeOffAir)
		if afterLast > em.t {
			return afterLast
		}
		return em.t
	}
	em, err = sb.ScheduleAnytime(em.d, next, priority)
	if err != nil {
		return Emission{}, err
	}
	s.emissions = s.emissions.Insert(em)
	return em, nil
}

// Sync synchronizes the clock with the given concentrator time v and the server time.
func (s *Scheduler) Sync(v uint32, server time.Time) {
	s.mu.Lock()
	s.clock.Sync(v, server)
	s.mu.Unlock()
}

// SyncWithGatewayAbsolute synchronizes the clock with the given concentrator timestamp, the server time and the
// absolute gateway time that corresponds to the given timestamp.
func (s *Scheduler) SyncWithGatewayAbsolute(timestamp uint32, server, gateway time.Time) {
	s.mu.Lock()
	s.clock.SyncWithGatewayAbsolute(timestamp, server, gateway)
	s.mu.Unlock()
}

// SyncWithGatewayConcentrator synchronizes the clock with the given concentrator timestamp, the server time and the
// relative gateway time that corresponds to the given timestamp.
func (s *Scheduler) SyncWithGatewayConcentrator(timestamp uint32, server time.Time, concentrator ConcentratorTime) {
	s.mu.Lock()
	s.clock.SyncWithGatewayConcentrator(timestamp, server, concentrator)
	s.mu.Unlock()
}

// Now returns an indication of the current concentrator time.
// This method returns false if the clock is not synced with the server.
func (s *Scheduler) Now() (ConcentratorTime, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.clock.IsSynced() {
		return 0, false
	}
	return s.clock.FromServerTime(s.timeSource.Now())
}

// TimeFromTimestampTime returns the concentrator time by the given timestamp.
// This method returns false if the clock is not synced with the server.
func (s *Scheduler) TimeFromTimestampTime(t uint32) (ConcentratorTime, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.clock.IsSynced() {
		return 0, false
	}
	return s.clock.FromTimestampTime(t), true
}
