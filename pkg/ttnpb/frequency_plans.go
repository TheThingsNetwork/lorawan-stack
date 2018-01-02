// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import time "time"

// Extend a frequency plan from a frequency plan blueprint
func (f FrequencyPlan) Extend(ext FrequencyPlan) FrequencyPlan {
	extended := f

	extended.copyChannels(ext.Channels, ext.FSKChannel, ext.LoraStandardChannel)
	extended.copyDwellTime(ext.DwellTime)
	extended.copyLBT(ext.LBT)
	extended.copyTimeOffAir(ext.TimeOffAir)

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

func (f *FrequencyPlan) copyDwellTime(dwellTime *time.Duration) {
	if dwellTime != nil {
		duration := *dwellTime
		f.DwellTime = &duration
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
