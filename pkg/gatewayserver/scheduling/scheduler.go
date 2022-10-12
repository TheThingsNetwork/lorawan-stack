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

	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/toa"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
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

	// scheduleMinRTTCount is the minimum number of observed round-trip times that are taken into account before using
	// using their statistics for calculating an absolute time or determining whether scheduling is too late.
	scheduleMinRTTCount = 5

	// scheduleLateRTTPercentile is the percentile of round-trip times that is considered for determining whether
	// scheduling is too late.
	scheduleLateRTTPercentile = 90
)

// TimeSource is a source for getting a current time.
type TimeSource interface {
	Now() time.Time
}

type systemTimeSource struct{}

// Now implements TimeSource.
func (systemTimeSource) Now() time.Time { return time.Now() }

// SystemTimeSource is a TimeSource that uses the local system time.
var SystemTimeSource = &systemTimeSource{}

// RTTs provides round-trip times.
type RTTs interface {
	Stats(percentile int, ref time.Time) (min, max, median, np time.Duration, count int)
}

var (
	errFrequencyPlansTimeOffAir     = errors.DefineInvalidArgument("frequency_plans_time_off_air", "frequency plans must have the same time off air value")
	errFrequencyPlansOverlapSubBand = errors.DefineInvalidArgument("frequency_plans_overlap_sub_band", "frequency plans must not have overlapping sub bands")
)

// NewScheduler instantiates a new Scheduler for the given frequency plan.
// If no time source is specified, the system time is used.
func NewScheduler(
	ctx context.Context,
	fps map[string]*frequencyplans.FrequencyPlan,
	enforceDutyCycle bool,
	dutyCycleStyle DutyCycleStyle,
	scheduleAnytimeDelay *time.Duration,
	timeSource TimeSource,
) (*Scheduler, error) {
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

	var toa *frequencyplans.TimeOffAir
	for _, fp := range fps {
		if toa != nil && fp.TimeOffAir != *toa {
			return nil, errFrequencyPlansTimeOffAir.New()
		}
		toa = &fp.TimeOffAir
	}

	if toa.Duration < QueueDelay {
		toa.Duration = QueueDelay
	}

	s := &Scheduler{
		clock:                &RolloverClock{},
		timeOffAir:           *toa,
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
					sb := NewSubBand(params, s.clock, nil, dutyCycleStyle)
					var isIdentical bool
					for _, subBand := range s.subBands {
						if subBand.IsIdentical(sb) {
							isIdentical = true
							break
						}
						if subBand.HasOverlap(sb) {
							return nil, errFrequencyPlansOverlapSubBand.New()
						}
					}
					if !isIdentical {
						s.subBands = append(s.subBands, sb)
					}
				}
			} else {
				band, err := band.GetLatest(fp.BandID)
				if err != nil {
					return nil, err
				}
				for _, subBand := range band.SubBands {
					params := SubBandParameters{
						MinFrequency: subBand.MinFrequency,
						MaxFrequency: subBand.MaxFrequency,
						DutyCycle:    subBand.DutyCycle,
					}
					sb := NewSubBand(params, s.clock, nil, dutyCycleStyle)
					var isIdentical bool
					for _, subBand := range s.subBands {
						if subBand.IsIdentical(sb) {
							isIdentical = true
							break
						}
						if subBand.HasOverlap(sb) {
							return nil, errFrequencyPlansOverlapSubBand.New()
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
		sb := NewSubBand(noDutyCycleParams, s.clock, nil, dutyCycleStyle)
		s.subBands = append(s.subBands, sb)
	}
	go s.gc(ctx)
	return s, nil
}

// Scheduler is a packet scheduler that takes time conflicts and sub-band restrictions into account.
type Scheduler struct {
	clock                *RolloverClock
	fps                  map[string]*frequencyplans.FrequencyPlan
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

func (s *Scheduler) gc(ctx context.Context) error {
	ticker := time.NewTicker(DutyCycleWindow / 2)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			s.mu.RLock()
			serverTime, ok := s.clock.FromServerTime(s.timeSource.Now())
			s.mu.RUnlock()
			if !ok {
				continue
			}
			to := serverTime - ConcentratorTime(DutyCycleWindow)
			for _, subBand := range s.subBands {
				subBand.gc(to)
			}
			s.mu.Lock()
			s.emissions = s.emissions.GreaterThan(to)
			s.mu.Unlock()
		}
	}
}

var errDwellTime = errors.DefineFailedPrecondition("dwell_time", "packet exceeds dwell time restriction")

func (s *Scheduler) newEmission(payloadSize int, settings *ttnpb.TxSettings, starts ConcentratorTime) (Emission, error) {
	d, err := toa.Compute(payloadSize, settings)
	if err != nil {
		return Emission{}, err
	}
	for _, fp := range s.fps {
		if fp.RespectsDwellTime(true, settings.Frequency, d) {
			return NewEmission(starts, d), nil
		}
	}
	return Emission{}, errDwellTime.New()
}

// SubBandCount returns the number of sub bands in the scheduler.
func (s *Scheduler) SubBandCount() int {
	return len(s.subBands)
}

// syncWithUplinkToken synchronizes the clock using the given token.
// If the given token does not provide enough information or if the latest clock sync is more recent, this method returns false.
// This method assumes that the mutex is held.
func (s *Scheduler) syncWithUplinkToken(token *ttnpb.UplinkToken) bool {
	if token.GetServerTime() == nil || token.GetConcentratorTime() == 0 {
		return false
	}
	if lastSync, ok := s.clock.SyncTime(); ok && lastSync.After(*ttnpb.StdTime(token.ServerTime)) {
		return false
	}
	s.clock.SyncWithGatewayConcentrator(token.Timestamp, *ttnpb.StdTime(token.ServerTime), ttnpb.StdTime(token.GatewayTime), ConcentratorTime(token.ConcentratorTime))
	return true
}

var (
	errConflict              = errors.DefineAlreadyExists("conflict", "scheduling conflict")
	errTooLate               = errors.DefineFailedPrecondition("too_late", "too late to transmission scheduled time (delta is `{delta}`, min is `{min}`)")
	errNoClockSync           = errors.DefineUnavailable("no_clock_sync", "no clock sync")
	errNoAbsoluteGatewayTime = errors.DefineAborted("no_absolute_gateway_time", "no absolute gateway time")
	errNoServerTime          = errors.DefineAborted("no_server_time", "no server time")
)

// Options define options for scheduling downlink.
type Options struct {
	PayloadSize int
	*ttnpb.TxSettings
	RTTs        RTTs
	Priority    ttnpb.TxSchedulePriority
	UplinkToken *ttnpb.UplinkToken
}

// ScheduleAt attempts to schedule the given Tx settings with the given priority.
// If there are round-trip times available, the nth percentile (n = scheduleLateRTTPercentile) value will be used instead of ScheduleTimeShort.
func (s *Scheduler) ScheduleAt(ctx context.Context, opts Options) (res Emission, now ConcentratorTime, err error) {
	defer trace.StartRegion(ctx, "schedule transmission").End()

	s.mu.Lock()
	defer s.mu.Unlock()
	if opts.UplinkToken != nil {
		s.syncWithUplinkToken(opts.UplinkToken)
	}
	if !s.clock.IsSynced() {
		return Emission{}, 0, errNoClockSync.New()
	}
	minScheduleTime := ScheduleTimeShort
	var medianRTT *time.Duration
	if opts.RTTs != nil {
		if _, _, median, np, n := opts.RTTs.Stats(scheduleLateRTTPercentile, s.timeSource.Now()); n >= scheduleMinRTTCount {
			minScheduleTime = np/2 + QueueDelay
			medianRTT = &median
		}
	}
	log.FromContext(ctx).WithFields(log.Fields(
		"median_rtt", medianRTT,
		"min_schedule_time", minScheduleTime,
	)).Debug("Computed scheduling delays")
	var starts ConcentratorTime
	now, ok := s.clock.FromServerTime(s.timeSource.Now())
	if !ok {
		panic("clock is synced without server time")
	}
	if opts.Time != nil {
		var ok bool
		starts, ok = s.clock.FromGatewayTime(*ttnpb.StdTime(opts.Time))
		if !ok {
			if medianRTT == nil {
				return Emission{}, 0, errNoAbsoluteGatewayTime.New()
			}
			serverTime, ok := s.clock.FromServerTime(*ttnpb.StdTime(opts.Time))
			if !ok {
				return Emission{}, 0, errNoServerTime.New()
			}
			starts = serverTime - ConcentratorTime(*medianRTT/2)
		}
	} else {
		starts = s.clock.FromTimestampTime(opts.Timestamp)
	}
	delay := time.Duration(starts - now)
	if delay < minScheduleTime {
		return Emission{}, 0, errTooLate.WithAttributes(
			"delay", delay,
			"min", minScheduleTime,
		)
	}
	log.FromContext(ctx).WithFields(log.Fields(
		"now", now,
		"starts", starts,
		"delay", delay,
	)).Debug("Computed downlink start timestamp")
	sb, err := s.findSubBand(opts.Frequency)
	if err != nil {
		return Emission{}, 0, err
	}
	em, err := s.newEmission(opts.PayloadSize, opts.TxSettings, starts)
	if err != nil {
		return Emission{}, 0, err
	}
	for _, other := range s.emissions {
		if em.OverlapsWithOffAir(other, s.timeOffAir) {
			return Emission{}, 0, errConflict.New()
		}
	}
	if err := sb.Schedule(em, opts.Priority); err != nil {
		return Emission{}, 0, err
	}
	s.emissions = s.emissions.Insert(em)
	return em, now, nil
}

// ScheduleAnytime attempts to schedule the given Tx settings with the given priority from the time in the settings.
// If there are round-trip times available, the maximum value will be used instead of ScheduleTimeShort.
// This method returns the emission.
//
// The scheduler does not support immediate scheduling, i.e. sending a message to the gateway that should be transmitted
// immediately. The reason for this is that this scheduler cannot determine conflicts or enforce duty-cycle when the
// emission time is unknown. Therefore, when the time is set to Immediate, the estimated current concentrator time plus
// ScheduleDelayLong will be used.
func (s *Scheduler) ScheduleAnytime(ctx context.Context, opts Options) (res Emission, now ConcentratorTime, err error) {
	defer trace.StartRegion(ctx, "schedule transmission at any time").End()

	s.mu.Lock()
	defer s.mu.Unlock()
	if opts.UplinkToken != nil {
		s.syncWithUplinkToken(opts.UplinkToken)
	}
	if !s.clock.IsSynced() {
		return Emission{}, 0, errNoClockSync.New()
	}
	minScheduleTime := ScheduleTimeShort
	if opts.RTTs != nil {
		if _, _, _, np, n := opts.RTTs.Stats(scheduleLateRTTPercentile, s.timeSource.Now()); n >= scheduleMinRTTCount {
			minScheduleTime = np/2 + QueueDelay
		}
	}
	var starts ConcentratorTime
	now, ok := s.clock.FromServerTime(s.timeSource.Now())
	if !ok {
		panic("clock is synced without server time")
	}
	if opts.Timestamp == 0 {
		starts = now + ConcentratorTime(s.scheduleAnytimeDelay)
		opts.Timestamp = uint32(time.Duration(starts) / time.Microsecond)
	} else {
		starts = s.clock.FromTimestampTime(opts.Timestamp)
		if delta := minScheduleTime - time.Duration(starts-now); delta > 0 {
			starts += ConcentratorTime(delta)
			opts.Timestamp += uint32(delta / time.Microsecond)
		}
	}
	sb, err := s.findSubBand(opts.Frequency)
	if err != nil {
		return Emission{}, 0, err
	}
	em, err := s.newEmission(opts.PayloadSize, opts.TxSettings, starts)
	if err != nil {
		return Emission{}, 0, err
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
	em, err = sb.ScheduleAnytime(em.d, next, opts.Priority)
	if err != nil {
		return Emission{}, 0, err
	}
	s.emissions = s.emissions.Insert(em)
	return em, now, nil
}

// Sync synchronizes the clock with the given concentrator time v and the server time.
func (s *Scheduler) Sync(v uint32, server time.Time) ConcentratorTime {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.clock.Sync(v, server)
}

// SyncWithGatewayAbsolute synchronizes the clock with the given concentrator timestamp, the server time and the
// absolute gateway time that corresponds to the given timestamp.
func (s *Scheduler) SyncWithGatewayAbsolute(timestamp uint32, server, gateway time.Time) ConcentratorTime {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.clock.SyncWithGatewayAbsolute(timestamp, server, gateway)
}

// SyncWithGatewayConcentrator synchronizes the clock with the given concentrator timestamp, the server time and the
// relative gateway time that corresponds to the given timestamp.
func (s *Scheduler) SyncWithGatewayConcentrator(timestamp uint32, server time.Time, gateway *time.Time, concentrator ConcentratorTime) ConcentratorTime {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.clock.SyncWithGatewayConcentrator(timestamp, server, gateway, concentrator)
}

// IsGatewayTimeSynced reports whether scheduler clock is synchronized with gateway time.
func (s *Scheduler) IsGatewayTimeSynced() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clock.IsSynced() && s.clock.gateway != nil
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

// TimeFromServerTime returns an indication of the provided timestamp in concentrator time.
// This method returns false if the clock is not synced with the server.
func (s *Scheduler) TimeFromServerTime(t time.Time) (ConcentratorTime, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if !s.clock.IsSynced() {
		return 0, false
	}
	return s.clock.FromServerTime(t)
}

// SubBandStats returns a map with the usage stats of each sub band.
func (s *Scheduler) SubBandStats() []*ttnpb.GatewayConnectionStats_SubBand {
	var res []*ttnpb.GatewayConnectionStats_SubBand

	for _, sb := range s.subBands {
		res = append(res, &ttnpb.GatewayConnectionStats_SubBand{
			MaxFrequency:             sb.MaxFrequency,
			MinFrequency:             sb.MinFrequency,
			DownlinkUtilizationLimit: sb.DutyCycle,
			DownlinkUtilization:      sb.DutyCycleUtilization(),
		})
	}

	return res
}
