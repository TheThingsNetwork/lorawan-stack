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

// Extend a frequency plan from a frequency plan blueprint
func (f FrequencyPlan) Extend(ext FrequencyPlan) FrequencyPlan {
	if channels := ext.Channels; channels != nil {
		f.Channels = make([]*FrequencyPlan_Channel, 0)
		for _, channel := range channels {
			f.Channels = append(f.Channels, &FrequencyPlan_Channel{
				Frequency: channel.Frequency,
				DataRate:  channel.DataRate,
			})
		}
	}
	if ext.FSKChannel != nil {
		f.FSKChannel = &FrequencyPlan_Channel{Frequency: ext.FSKChannel.Frequency, DataRate: ext.FSKChannel.DataRate}
	}
	if ext.LoraStandardChannel != nil {
		f.LoraStandardChannel = &FrequencyPlan_Channel{Frequency: ext.LoraStandardChannel.Frequency, DataRate: ext.LoraStandardChannel.DataRate}
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
		f.LBT = &FrequencyPlan_LBTConfiguration{
			RSSIOffset: ext.LBT.RSSIOffset,
			RSSITarget: ext.LBT.RSSITarget,
			ScanTime:   ext.LBT.ScanTime,
		}
	}
	if ext.TimeOffAir != nil {
		f.TimeOffAir = &FrequencyPlan_TimeOffAir{
			Duration: ext.TimeOffAir.Duration,
			Fraction: ext.TimeOffAir.Fraction,
		}
	}
	if ext.PingSlot != nil {
		f.PingSlot = &FrequencyPlan_Channel{
			Frequency: ext.PingSlot.Frequency,
			DataRate:  ext.PingSlot.DataRate,
		}
	}
	if ext.Rx2 != nil {
		f.Rx2 = &FrequencyPlan_Channel{
			Frequency: ext.Rx2.Frequency,
			DataRate:  ext.Rx2.DataRate,
		}
	}
	if ext.MaxEIRP != 0.0 {
		f.MaxEIRP = ext.MaxEIRP
	}

	return f
}
