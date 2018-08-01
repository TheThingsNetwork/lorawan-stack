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

// Copy the channel parameters to a new struct.
func (c *FrequencyPlan_Channel) Copy() *FrequencyPlan_Channel {
	return &FrequencyPlan_Channel{
		Frequency: c.Frequency,
		DataRate:  c.DataRate,
	}
}

// Copy the time off air parameters to a new struct.
func (toa *FrequencyPlan_TimeOffAir) Copy() *FrequencyPlan_TimeOffAir {
	return &FrequencyPlan_TimeOffAir{
		Duration: toa.Duration,
		Fraction: toa.Fraction,
	}
}

func (lbt *FrequencyPlan_LBTConfiguration) Copy() *FrequencyPlan_LBTConfiguration {
	return &FrequencyPlan_LBTConfiguration{
		RSSIOffset: lbt.RSSIOffset,
		RSSITarget: lbt.RSSITarget,
		ScanTime:   lbt.ScanTime,
	}
}

// Extend a frequency plan from a frequency plan blueprint
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
	if ext.UplinkDwellTime != nil {
		duration := *ext.UplinkDwellTime
		f.UplinkDwellTime = &duration
	}
	if ext.DownlinkDwellTime != nil {
		duration := *ext.DownlinkDwellTime
		f.DownlinkDwellTime = &duration
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

	return f
}
