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

package networkserver

import (
	"go.thethings.network/lorawan-stack/pkg/band"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func getDeviceBandVersion(dev *ttnpb.EndDevice, fps *frequencyplans.Store) (*frequencyplans.FrequencyPlan, band.Band, error) {
	fp, err := fps.GetByID(dev.FrequencyPlanID)
	if err != nil {
		return nil, band.Band{}, err
	}
	b, err := band.GetByID(fp.BandID)
	if err != nil {
		return nil, band.Band{}, err
	}
	b, err = b.Version(dev.LoRaWANPHYVersion)
	if err != nil {
		return nil, band.Band{}, err
	}
	return fp, b, nil
}

func searchDataRate(dr ttnpb.DataRate, dev *ttnpb.EndDevice, fps *frequencyplans.Store) (ttnpb.DataRateIndex, error) {
	_, phy, err := getDeviceBandVersion(dev, fps)
	if err != nil {
		return 0, err
	}
	for i, bDR := range phy.DataRates {
		if bDR.Rate.Equal(dr) {
			return ttnpb.DataRateIndex(i), nil
		}
	}
	return 0, errDataRateNotFound.WithAttributes("data_rate", dr)
}

func searchUplinkChannel(freq uint64, dev *ttnpb.EndDevice) (uint8, error) {
	for i, ch := range dev.MACState.CurrentParameters.Channels {
		if ch.UplinkFrequency == freq {
			return uint8(i), nil
		}
	}
	return 0, errUplinkChannelNotFound.WithAttributes("frequency", freq)
}

func newMACState(dev *ttnpb.EndDevice, fps *frequencyplans.Store, defaults ttnpb.MACSettings) (*ttnpb.MACState, error) {
	fp, phy, err := getDeviceBandVersion(dev, fps)
	if err != nil {
		return nil, err
	}

	st := &ttnpb.MACState{
		LoRaWANVersion: dev.LoRaWANVersion,
		DeviceClass:    ttnpb.CLASS_A,
	}

	st.CurrentParameters.MaxEIRP = phy.DefaultMaxEIRP
	st.DesiredParameters.MaxEIRP = st.CurrentParameters.MaxEIRP
	if fp.MaxEIRP != nil && *fp.MaxEIRP > 0 && *fp.MaxEIRP < st.CurrentParameters.MaxEIRP {
		st.DesiredParameters.MaxEIRP = *fp.MaxEIRP
	}

	st.CurrentParameters.UplinkDwellTime = false
	st.DesiredParameters.UplinkDwellTime = fp.DwellTime.GetUplinks()

	st.CurrentParameters.DownlinkDwellTime = false
	st.DesiredParameters.DownlinkDwellTime = fp.DwellTime.GetDownlinks()

	st.CurrentParameters.ADRDataRateIndex = ttnpb.DATA_RATE_0
	st.DesiredParameters.ADRDataRateIndex = st.CurrentParameters.ADRDataRateIndex

	st.CurrentParameters.ADRTxPowerIndex = 0
	st.DesiredParameters.ADRTxPowerIndex = st.CurrentParameters.ADRTxPowerIndex

	st.CurrentParameters.ADRNbTrans = 1
	st.DesiredParameters.ADRNbTrans = st.CurrentParameters.ADRNbTrans

	st.CurrentParameters.ADRAckLimit = uint32(phy.ADRAckLimit)
	st.DesiredParameters.ADRAckLimit = st.CurrentParameters.ADRAckLimit

	st.CurrentParameters.ADRAckDelay = uint32(phy.ADRAckDelay)
	st.DesiredParameters.ADRAckDelay = st.CurrentParameters.ADRAckDelay

	st.CurrentParameters.Rx1Delay = ttnpb.RxDelay(phy.ReceiveDelay1.Seconds())
	if dev.GetMACSettings().GetRx1Delay() != nil {
		st.CurrentParameters.Rx1Delay = dev.MACSettings.Rx1Delay.Value
	} else if defaults.Rx1Delay != nil {
		st.CurrentParameters.Rx1Delay = defaults.Rx1Delay.Value
	}
	st.DesiredParameters.Rx1Delay = st.CurrentParameters.Rx1Delay
	if dev.GetMACSettings().GetDesiredRx1Delay() != nil {
		st.DesiredParameters.Rx1Delay = dev.MACSettings.DesiredRx1Delay.Value
	} else if defaults.DesiredRx1Delay != nil {
		st.DesiredParameters.Rx1Delay = defaults.DesiredRx1Delay.Value
	}

	st.CurrentParameters.Rx1DataRateOffset = 0
	if dev.GetMACSettings().GetRx1DataRateOffset() != nil {
		st.CurrentParameters.Rx1DataRateOffset = dev.MACSettings.Rx1DataRateOffset.Value
	} else if defaults.Rx1DataRateOffset != nil {
		st.CurrentParameters.Rx1DataRateOffset = defaults.Rx1DataRateOffset.Value
	}
	st.DesiredParameters.Rx1DataRateOffset = st.CurrentParameters.Rx1DataRateOffset
	if dev.GetMACSettings().GetDesiredRx1DataRateOffset() != nil {
		st.DesiredParameters.Rx1DataRateOffset = dev.MACSettings.DesiredRx1DataRateOffset.Value
	} else if defaults.DesiredRx1DataRateOffset != nil {
		st.DesiredParameters.Rx1DataRateOffset = defaults.DesiredRx1DataRateOffset.Value
	}

	st.CurrentParameters.Rx2DataRateIndex = phy.DefaultRx2Parameters.DataRateIndex
	if dev.GetMACSettings().GetRx2DataRateIndex() != nil {
		st.CurrentParameters.Rx2DataRateIndex = dev.MACSettings.Rx2DataRateIndex.Value
	} else if defaults.Rx2DataRateIndex != nil {
		st.CurrentParameters.Rx2DataRateIndex = defaults.Rx2DataRateIndex.Value
	}
	st.DesiredParameters.Rx2DataRateIndex = st.CurrentParameters.Rx2DataRateIndex
	if dev.GetMACSettings().GetDesiredRx2DataRateIndex() != nil {
		st.DesiredParameters.Rx2DataRateIndex = dev.MACSettings.DesiredRx2DataRateIndex.Value
	} else if fp.DefaultRx2DataRate != nil {
		st.DesiredParameters.Rx2DataRateIndex = ttnpb.DataRateIndex(*fp.DefaultRx2DataRate)
	} else if defaults.DesiredRx2DataRateIndex != nil {
		st.DesiredParameters.Rx2DataRateIndex = defaults.DesiredRx2DataRateIndex.Value
	}

	st.CurrentParameters.Rx2Frequency = phy.DefaultRx2Parameters.Frequency
	if dev.GetMACSettings().GetRx2Frequency() != nil && dev.MACSettings.Rx2Frequency.Value != 0 {
		st.CurrentParameters.Rx2Frequency = dev.MACSettings.Rx2Frequency.Value
	} else if defaults.Rx2Frequency != nil && dev.MACSettings.Rx2Frequency.Value != 0 {
		st.CurrentParameters.Rx2Frequency = defaults.Rx2Frequency.Value
	}
	st.DesiredParameters.Rx2Frequency = st.CurrentParameters.Rx2Frequency
	if dev.GetMACSettings().GetDesiredRx2Frequency() != nil && dev.MACSettings.Rx2Frequency.Value != 0 {
		st.DesiredParameters.Rx2Frequency = dev.MACSettings.DesiredRx2Frequency.Value
	} else if fp.Rx2Channel != nil {
		st.DesiredParameters.Rx2Frequency = fp.Rx2Channel.Frequency
	} else if defaults.DesiredRx2Frequency != nil && defaults.DesiredRx2Frequency.Value != 0 {
		st.DesiredParameters.Rx2Frequency = defaults.DesiredRx2Frequency.Value
	}

	st.CurrentParameters.MaxDutyCycle = ttnpb.DUTY_CYCLE_1
	if dev.GetMACSettings().GetMaxDutyCycle() != nil {
		st.CurrentParameters.MaxDutyCycle = dev.MACSettings.MaxDutyCycle.Value
	}
	st.DesiredParameters.MaxDutyCycle = st.CurrentParameters.MaxDutyCycle

	st.CurrentParameters.RejoinTimePeriodicity = ttnpb.REJOIN_TIME_0
	st.DesiredParameters.RejoinTimePeriodicity = st.CurrentParameters.RejoinTimePeriodicity

	st.CurrentParameters.RejoinCountPeriodicity = ttnpb.REJOIN_COUNT_16
	st.DesiredParameters.RejoinCountPeriodicity = st.CurrentParameters.RejoinCountPeriodicity

	st.CurrentParameters.PingSlotFrequency = 0
	if dev.GetMACSettings().GetPingSlotFrequency() != nil && dev.MACSettings.PingSlotFrequency.Value != 0 {
		st.CurrentParameters.PingSlotFrequency = dev.MACSettings.PingSlotFrequency.Value
	} else if defaults.PingSlotFrequency != nil && defaults.PingSlotFrequency.Value != 0 {
		st.CurrentParameters.PingSlotFrequency = defaults.PingSlotFrequency.Value
	}
	st.DesiredParameters.PingSlotFrequency = st.CurrentParameters.PingSlotFrequency
	if fp.PingSlot != nil && fp.PingSlot.Frequency != 0 {
		st.DesiredParameters.PingSlotFrequency = fp.PingSlot.Frequency
	}

	st.CurrentParameters.PingSlotDataRateIndex = ttnpb.DATA_RATE_0
	if dev.GetMACSettings().GetPingSlotDataRateIndex() != nil {
		st.CurrentParameters.PingSlotDataRateIndex = dev.MACSettings.PingSlotDataRateIndex.Value
	} else if defaults.PingSlotDataRateIndex != nil {
		st.CurrentParameters.PingSlotDataRateIndex = defaults.PingSlotDataRateIndex.Value
	}
	st.DesiredParameters.PingSlotDataRateIndex = st.CurrentParameters.PingSlotDataRateIndex
	if fp.DefaultPingSlotDataRate != nil {
		st.DesiredParameters.PingSlotDataRateIndex = ttnpb.DataRateIndex(*fp.DefaultPingSlotDataRate)
	}

	st.CurrentParameters.BeaconFrequency = 0
	st.DesiredParameters.BeaconFrequency = st.CurrentParameters.BeaconFrequency

	if len(phy.DownlinkChannels) > len(phy.UplinkChannels) || len(fp.DownlinkChannels) > len(fp.UplinkChannels) ||
		len(phy.UplinkChannels) > int(phy.MaxUplinkChannels) || len(phy.DownlinkChannels) > int(phy.MaxDownlinkChannels) ||
		len(fp.UplinkChannels) > int(phy.MaxUplinkChannels) || len(fp.DownlinkChannels) > int(phy.MaxDownlinkChannels) {
		// NOTE: In case the spec changes and this assumption is not valid anymore,
		// the implementation of this function won't be valid and has to be changed.
		panic("uplink/downlink channel length is inconsistent")
	}

	if len(dev.GetMACSettings().GetFactoryPresetFrequencies()) > 0 {
		st.CurrentParameters.Channels = make([]*ttnpb.MACParameters_Channel, 0, len(dev.MACSettings.FactoryPresetFrequencies))
		for _, freq := range dev.MACSettings.FactoryPresetFrequencies {
			st.CurrentParameters.Channels = append(st.CurrentParameters.Channels, &ttnpb.MACParameters_Channel{
				MinDataRateIndex:  0,
				MaxDataRateIndex:  ttnpb.DATA_RATE_15,
				UplinkFrequency:   freq,
				DownlinkFrequency: freq,
				EnableUplink:      true,
			})
		}
	} else if len(defaults.GetFactoryPresetFrequencies()) > 0 {
		st.CurrentParameters.Channels = make([]*ttnpb.MACParameters_Channel, 0, len(defaults.FactoryPresetFrequencies))
		for _, freq := range defaults.FactoryPresetFrequencies {
			st.CurrentParameters.Channels = append(st.CurrentParameters.Channels, &ttnpb.MACParameters_Channel{
				MinDataRateIndex:  0,
				MaxDataRateIndex:  ttnpb.DATA_RATE_15,
				UplinkFrequency:   freq,
				DownlinkFrequency: freq,
				EnableUplink:      true,
			})
		}
	} else {
		st.CurrentParameters.Channels = make([]*ttnpb.MACParameters_Channel, 0, len(phy.UplinkChannels))
		for i, upCh := range phy.UplinkChannels {
			channel := &ttnpb.MACParameters_Channel{
				MinDataRateIndex: upCh.MinDataRate,
				MaxDataRateIndex: upCh.MaxDataRate,
				UplinkFrequency:  upCh.Frequency,
				EnableUplink:     true,
			}
			channel.DownlinkFrequency = phy.DownlinkChannels[i%len(phy.DownlinkChannels)].Frequency
			st.CurrentParameters.Channels = append(st.CurrentParameters.Channels, channel)
		}
	}

	st.DesiredParameters.Channels = make([]*ttnpb.MACParameters_Channel, 0, len(phy.UplinkChannels)+len(fp.UplinkChannels))
	for i, upCh := range phy.UplinkChannels {
		channel := &ttnpb.MACParameters_Channel{
			MinDataRateIndex: upCh.MinDataRate,
			MaxDataRateIndex: upCh.MaxDataRate,
			UplinkFrequency:  upCh.Frequency,
		}
		channel.DownlinkFrequency = phy.DownlinkChannels[i%len(phy.DownlinkChannels)].Frequency
		st.DesiredParameters.Channels = append(st.DesiredParameters.Channels, channel)
	}

outerUp:
	for _, upCh := range fp.UplinkChannels {
		for _, ch := range st.DesiredParameters.Channels {
			if ch.UplinkFrequency == upCh.Frequency {
				ch.MinDataRateIndex = ttnpb.DataRateIndex(upCh.MinDataRate)
				ch.MaxDataRateIndex = ttnpb.DataRateIndex(upCh.MaxDataRate)
				ch.EnableUplink = true
				continue outerUp
			}
		}

		st.DesiredParameters.Channels = append(st.DesiredParameters.Channels, &ttnpb.MACParameters_Channel{
			MinDataRateIndex: ttnpb.DataRateIndex(upCh.MinDataRate),
			MaxDataRateIndex: ttnpb.DataRateIndex(upCh.MaxDataRate),
			UplinkFrequency:  upCh.Frequency,
			EnableUplink:     true,
		})
	}

	if len(fp.DownlinkChannels) > 0 {
		for i, ch := range st.DesiredParameters.Channels {
			downCh := fp.DownlinkChannels[i%len(fp.DownlinkChannels)]
			if downCh.Frequency != 0 {
				ch.DownlinkFrequency = downCh.Frequency
			}
		}
	}

	return st, nil
}
