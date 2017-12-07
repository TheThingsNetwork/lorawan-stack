// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package scheduling offer convenience methods to manage RF packets that must respect scheduling constraints
package scheduling

import (
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/band"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

var (
	// ErrDutyCycleFull is returned is the duty cycle prevents scheduling of a downlink
	ErrDutyCycleFull = errors.New("Duty cycle full")
	// ErrOverlap is returned if there is an already existing scheduling overlapping
	ErrOverlap = errors.New("Window overlap")
	// ErrTimeOffAir is returned if time-off-air constraints prevent scheduling of the new downlink
	ErrTimeOffAir = errors.New("Time-off-air constraints prevent scheduling")
	// ErrNoSubBandFound is returned when an operation fails because there is no sub band for the given channel
	ErrNoSubBandFound = errors.New("No sub band found for the given channel")
	// ErrDwellTime is returned when an operation fails because the packet does not respect the dwell time
	ErrDwellTime = errors.New("Packet time-on-air duration is greater than this band's dwell time")
)

// Scheduler is an abstraction for an entity that manages the packet's time windows.
type Scheduler interface {
	// Schedule adds the requested time window to its internal schedule. If, because of its internal constraints (e.g. for duty cycles, not respecting the duty cycle), it returns ErrScheduleFull. If another error prevents scheduling, it is returned.
	Schedule(w Window, channel uint64) error
	// ScheduleFlexible requires a scheduling window if there is no time.Time constraint
	ScheduleFlexible(minimum time.Time, d time.Duration, channel uint64) (Window, error)
	// RegisterEmission that has happened during that time window, on that specific channel
	RegisterEmission(w Window, channel uint64) error
}

// FrequencyPlanScheduler returns a scheduler based on the frequency plan
func FrequencyPlanScheduler(fp ttnpb.FrequencyPlan) (Scheduler, error) {
	scheduler := &frequencyPlanScheduling{
		dwellTime:  fp.DwellTime,
		timeOffAir: fp.TimeOffAir,
		subBands:   []*subBandScheduling{},
	}

	band, err := band.GetByID(fp.BandID)
	if err != nil {
		return nil, errors.NewWithCause("Could not find band associated to that frequency plan", err)
	}

	for _, subBand := range band.BandDutyCycles {
		scheduler.subBands = append(scheduler.subBands, &subBandScheduling{
			dutyCycle:         subBand,
			schedulingWindows: []packetWindow{},

			mu: sync.Mutex{},
		})
	}

	return scheduler, nil
}

type frequencyPlanScheduling struct {
	dwellTime  *time.Duration
	timeOffAir *ttnpb.FrequencyPlan_TimeOffAir

	subBands []*subBandScheduling
}

func (f frequencyPlanScheduling) findSubBand(channel uint64) (*subBandScheduling, error) {
	for _, subBand := range f.subBands {
		if subBand.dutyCycle.Comprises(channel) {
			return subBand, nil
		}
	}

	return nil, ErrNoSubBandFound
}

func (f frequencyPlanScheduling) Schedule(w Window, channel uint64) error {
	if f.dwellTime != nil && *f.dwellTime < w.Duration {
		return ErrDwellTime
	}

	subBand, err := f.findSubBand(channel)
	if err != nil {
		return err
	}

	return subBand.Schedule(w, f.timeOffAir)
}

func (f frequencyPlanScheduling) ScheduleFlexible(minimum time.Time, d time.Duration, channel uint64) (Window, error) {
	subBand, err := f.findSubBand(channel)
	if err != nil {
		return Window{}, err
	}

	w, err := subBand.ScheduleFlexible(minimum, d, f.timeOffAir)
	return w, err
}

func (f frequencyPlanScheduling) RegisterEmission(w Window, channel uint64) error {
	subBand, err := f.findSubBand(channel)
	if err != nil {
		return err
	}

	subBand.mu.Lock()
	defer subBand.mu.Unlock()
	subBand.addScheduling(packetWindow{window: w, toa: w.timeOffAir(f.timeOffAir)})
	return nil
}
