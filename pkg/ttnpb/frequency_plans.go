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

import "time"

// Extend a frequency plan from a frequency plan blueprint
func (f FrequencyPlan) Extend(ext FrequencyPlan) FrequencyPlan {
	extended := f

	extended.copyChannels(ext.Channels, ext.FSKChannel, ext.LoraStandardChannel)
	extended.copyDwellTime(ext.UplinkDwellTime, ext.DownlinkDwellTime)
	extended.copyLBT(ext.LBT)
	extended.copyTimeOffAir(ext.TimeOffAir)
	extended.copyPingSlot(ext.PingSlot)
	extended.copyRX2(ext.RX2)
	extended.copyMaxEIRP(ext.MaxEIRP)

	return extended
}

func (f *FrequencyPlan) copyChannels(channels []*FrequencyPlan_Channel, fskChannel *FrequencyPlan_Channel, stdChannel *FrequencyPlan_Channel) {
	if channels != nil {
		f.Channels = make([]*FrequencyPlan_Channel, 0)
		for _, channel := range channels {
			f.Channels = append(f.Channels, &FrequencyPlan_Channel{
				Frequency: channel.Frequency,
				DataRate:  channel.DataRate,
			})
		}
	}

	if fskChannel != nil {
		f.FSKChannel = &FrequencyPlan_Channel{Frequency: fskChannel.Frequency, DataRate: fskChannel.DataRate}
	}

	if stdChannel != nil {
		f.LoraStandardChannel = &FrequencyPlan_Channel{Frequency: stdChannel.Frequency, DataRate: stdChannel.DataRate}
	}
}

func (f *FrequencyPlan) copyDwellTime(uplink, downlink *time.Duration) {
	if uplink != nil {
		duration := *uplink
		f.UplinkDwellTime = &duration
	}
	if downlink != nil {
		duration := *downlink
		f.DownlinkDwellTime = &duration
	}
}

func (f *FrequencyPlan) copyTimeOffAir(timeoff *FrequencyPlan_TimeOffAir) {
	if timeoff != nil {
		f.TimeOffAir = &FrequencyPlan_TimeOffAir{
			Duration: timeoff.Duration,
			Fraction: timeoff.Fraction,
		}
	}
}

func (f *FrequencyPlan) copyLBT(lbt *FrequencyPlan_LBTConfiguration) {
	if lbt != nil {
		f.LBT = &FrequencyPlan_LBTConfiguration{
			RSSIOffset: lbt.RSSIOffset,
			RSSITarget: lbt.RSSITarget,
			ScanTime:   lbt.ScanTime,
		}
	}
}

func (f *FrequencyPlan) copyPingSlot(pingSlot *FrequencyPlan_Channel) {
	if pingSlot != nil {
		f.PingSlot = &FrequencyPlan_Channel{
			Frequency: pingSlot.Frequency,
			DataRate:  pingSlot.DataRate,
		}
	}
}

func (f *FrequencyPlan) copyRX2(rx2 *FrequencyPlan_Channel) {
	if rx2 != nil {
		f.RX2 = &FrequencyPlan_Channel{
			Frequency: rx2.Frequency,
			DataRate:  rx2.DataRate,
		}
	}
}

func (f *FrequencyPlan) copyMaxEIRP(maxEIRP float32) {
	if maxEIRP != 0.0 {
		f.MaxEIRP = maxEIRP
	}
}
