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

package ttnpb

import (
	"fmt"
	"time"
)

// Validate returns an error if the frequency plan is invalid.
func (f *FrequencyPlan) Validate() error {
	if f.DwellTime != nil {
		dt := f.DwellTime
		if dt.Duration == nil && (dt.Uplinks || dt.Downlinks) {
			return errNoDwellTimeDuration
		}
	}
	for i, channel := range f.Channels {
		if channel == nil || channel.DwellTime == nil {
			continue
		}
		if f.DwellTime.GetDuration() == nil && channel.DwellTime.Duration == nil &&
			(channel.DwellTime.Uplinks || channel.DwellTime.Downlinks) {
			return errInvalidFrequencyPlanChannel.WithAttributes("index", i).WithCause(errNoDwellTimeDuration)
		}
	}
	return nil
}

// RespectsDwellTime returns whether the transmission respects the frequency plan's dwell time restrictions.
func (f *FrequencyPlan) RespectsDwellTime(isDownlink bool, frequency uint64, duration time.Duration) bool {
	isUplink := !isDownlink
	channels := append(f.Channels, f.LoraStandardChannel, f.FSKChannel)
	for _, ch := range channels {
		if ch == nil || ch.Frequency != frequency {
			continue
		}
		dtConfig := FrequencyPlan_DwellTime{
			Uplinks:   ch.DwellTime.GetUplinks() || f.DwellTime.GetUplinks(),
			Downlinks: ch.DwellTime.GetDownlinks() || f.DwellTime.GetDownlinks(),
		}
		if isDownlink && !dtConfig.Downlinks || isUplink && !dtConfig.Uplinks {
			return true
		}
		var dtDuration time.Duration
		switch {
		case ch.DwellTime.GetDuration() != nil:
			dtDuration = *ch.DwellTime.Duration
		case f.DwellTime.GetDuration() != nil:
			dtDuration = *f.DwellTime.Duration
		default:
			panic(fmt.Sprintf("frequency plan has dwell time enabled, but no dwell time duration set for channel %d", frequency))
		}
		return duration <= dtDuration
	}
	if isDownlink && f.DwellTime.GetDownlinks() || isUplink && f.DwellTime.GetUplinks() {
		return duration <= *f.DwellTime.Duration
	}
	return true
}

// Copy copies the dwell time to a new structure.
func (dt *FrequencyPlan_DwellTime) Copy() *FrequencyPlan_DwellTime {
	var duration *time.Duration
	if dt.Duration != nil {
		copyDuration := *dt.Duration
		duration = &copyDuration
	}
	return &FrequencyPlan_DwellTime{
		Uplinks:   dt.Uplinks,
		Downlinks: dt.Downlinks,
		Duration:  duration,
	}
}

// Copy copies the channel parameters to a new struct.
func (c *FrequencyPlan_Channel) Copy() *FrequencyPlan_Channel {
	return &FrequencyPlan_Channel{
		Frequency: c.Frequency,
		DataRate:  c.DataRate,
		DwellTime: c.DwellTime.Copy(),
	}
}

// Copy copies the time off air parameters to a new struct.
func (toa *FrequencyPlan_TimeOffAir) Copy() *FrequencyPlan_TimeOffAir {
	return &FrequencyPlan_TimeOffAir{
		Duration: toa.Duration,
		Fraction: toa.Fraction,
	}
}

// Copy copies the listen-before-talk configuration.
func (lbt *FrequencyPlan_LBTConfiguration) Copy() *FrequencyPlan_LBTConfiguration {
	return &FrequencyPlan_LBTConfiguration{
		RSSIOffset: lbt.RSSIOffset,
		RSSITarget: lbt.RSSITarget,
		ScanTime:   lbt.ScanTime,
	}
}

// Extend returns a new frequency plan with f's values, overridden by ext.
func (f FrequencyPlan) Extend(ext FrequencyPlan) FrequencyPlan {
	if channels := ext.Channels; channels != nil {
		f.Channels = make([]*FrequencyPlan_Channel, 0)
		for _, channel := range channels {
			f.Channels = append(f.Channels, channel.Copy())
		}
	}
	if ext.FSKChannel != nil {
		f.FSKChannel = ext.FSKChannel.Copy()
	}
	if ext.LoraStandardChannel != nil {
		f.LoraStandardChannel = ext.LoraStandardChannel.Copy()
	}
	if ext.LBT != nil {
		f.LBT = ext.LBT.Copy()
	}
	if ext.TimeOffAir != nil {
		f.TimeOffAir = ext.TimeOffAir.Copy()
	}
	if ext.PingSlot != nil {
		f.PingSlot = ext.PingSlot.Copy()
	}
	if ext.Rx2 != nil {
		f.Rx2 = ext.Rx2.Copy()
	}
	if ext.MaxEIRP != 0.0 {
		f.MaxEIRP = ext.MaxEIRP
	}
	if ext.DwellTime != nil {
		f.DwellTime = ext.DwellTime.Copy()
	}

	return f
}
