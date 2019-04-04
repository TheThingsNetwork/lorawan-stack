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
	_, band, err := getDeviceBandVersion(dev, fps)
	if err != nil {
		return 0, err
	}
	for i, bDR := range band.DataRates {
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

func resetMACState(dev *ttnpb.EndDevice, fps *frequencyplans.Store, defaults ttnpb.MACSettings) error {
	fp, band, err := getDeviceBandVersion(dev, fps)
	if err != nil {
		return err
	}

	dev.MACState = &ttnpb.MACState{
		LoRaWANVersion: dev.LoRaWANVersion,
		DeviceClass:    ttnpb.CLASS_A,
	}

	dev.MACState.CurrentParameters.MaxEIRP = band.DefaultMaxEIRP
	dev.MACState.DesiredParameters.MaxEIRP = dev.MACState.CurrentParameters.MaxEIRP
	if fp.MaxEIRP != nil && *fp.MaxEIRP > 0 && *fp.MaxEIRP < dev.MACState.CurrentParameters.MaxEIRP {
		dev.MACState.DesiredParameters.MaxEIRP = *fp.MaxEIRP
	}

	dev.MACState.CurrentParameters.UplinkDwellTime = false
	dev.MACState.DesiredParameters.UplinkDwellTime = fp.DwellTime.GetUplinks()

	dev.MACState.CurrentParameters.DownlinkDwellTime = false
	dev.MACState.DesiredParameters.DownlinkDwellTime = fp.DwellTime.GetDownlinks()

	dev.MACState.CurrentParameters.ADRDataRateIndex = ttnpb.DATA_RATE_0
	dev.MACState.DesiredParameters.ADRDataRateIndex = dev.MACState.CurrentParameters.ADRDataRateIndex

	dev.MACState.CurrentParameters.ADRTxPowerIndex = 0
	dev.MACState.DesiredParameters.ADRTxPowerIndex = dev.MACState.CurrentParameters.ADRTxPowerIndex

	dev.MACState.CurrentParameters.ADRNbTrans = 1
	dev.MACState.DesiredParameters.ADRNbTrans = dev.MACState.CurrentParameters.ADRNbTrans

	dev.MACState.CurrentParameters.ADRAckLimit = uint32(band.ADRAckLimit)
	dev.MACState.DesiredParameters.ADRAckLimit = dev.MACState.CurrentParameters.ADRAckLimit

	dev.MACState.CurrentParameters.ADRAckDelay = uint32(band.ADRAckDelay)
	dev.MACState.DesiredParameters.ADRAckDelay = dev.MACState.CurrentParameters.ADRAckDelay

	dev.MACState.CurrentParameters.Rx1Delay = ttnpb.RxDelay(band.ReceiveDelay1.Seconds())
	if dev.GetMACSettings().GetRx1Delay() != nil {
		dev.MACState.CurrentParameters.Rx1Delay = dev.MACSettings.Rx1Delay.Value
	} else if defaults.Rx1Delay != nil {
		dev.MACState.CurrentParameters.Rx1Delay = defaults.Rx1Delay.Value
	}
	dev.MACState.DesiredParameters.Rx1Delay = dev.MACState.CurrentParameters.Rx1Delay
	if dev.GetMACSettings().GetDesiredRx1Delay() != nil {
		dev.MACState.DesiredParameters.Rx1Delay = dev.MACSettings.DesiredRx1Delay.Value
	} else if defaults.DesiredRx1Delay != nil {
		dev.MACState.DesiredParameters.Rx1Delay = defaults.DesiredRx1Delay.Value
	}

	dev.MACState.CurrentParameters.Rx1DataRateOffset = 0
	if dev.GetMACSettings().GetRx1DataRateOffset() != nil {
		dev.MACState.CurrentParameters.Rx1DataRateOffset = dev.MACSettings.Rx1DataRateOffset.Value
	} else if defaults.Rx1DataRateOffset != nil {
		dev.MACState.CurrentParameters.Rx1DataRateOffset = defaults.Rx1DataRateOffset.Value
	}
	dev.MACState.DesiredParameters.Rx1DataRateOffset = dev.MACState.CurrentParameters.Rx1DataRateOffset
	if dev.GetMACSettings().GetDesiredRx1DataRateOffset() != nil {
		dev.MACState.DesiredParameters.Rx1DataRateOffset = dev.MACSettings.DesiredRx1DataRateOffset.Value
	} else if defaults.DesiredRx1DataRateOffset != nil {
		dev.MACState.DesiredParameters.Rx1DataRateOffset = defaults.DesiredRx1DataRateOffset.Value
	}

	dev.MACState.CurrentParameters.Rx2DataRateIndex = band.DefaultRx2Parameters.DataRateIndex
	if dev.GetMACSettings().GetRx2DataRateIndex() != nil {
		dev.MACState.CurrentParameters.Rx2DataRateIndex = dev.MACSettings.Rx2DataRateIndex.Value
	} else if defaults.Rx2DataRateIndex != nil {
		dev.MACState.CurrentParameters.Rx2DataRateIndex = defaults.Rx2DataRateIndex.Value
	}
	dev.MACState.DesiredParameters.Rx2DataRateIndex = dev.MACState.CurrentParameters.Rx2DataRateIndex
	if dev.GetMACSettings().GetDesiredRx2DataRateIndex() != nil {
		dev.MACState.DesiredParameters.Rx2DataRateIndex = dev.MACSettings.DesiredRx2DataRateIndex.Value
	} else if fp.DefaultRx2DataRate != nil {
		dev.MACState.DesiredParameters.Rx2DataRateIndex = ttnpb.DataRateIndex(*fp.DefaultRx2DataRate)
	} else if defaults.DesiredRx2DataRateIndex != nil {
		dev.MACState.DesiredParameters.Rx2DataRateIndex = defaults.DesiredRx2DataRateIndex.Value
	}

	dev.MACState.CurrentParameters.Rx2Frequency = band.DefaultRx2Parameters.Frequency
	if dev.GetMACSettings().GetRx2Frequency() != nil && dev.MACSettings.Rx2Frequency.Value != 0 {
		dev.MACState.CurrentParameters.Rx2Frequency = dev.MACSettings.Rx2Frequency.Value
	} else if defaults.Rx2Frequency != nil && dev.MACSettings.Rx2Frequency.Value != 0 {
		dev.MACState.CurrentParameters.Rx2Frequency = defaults.Rx2Frequency.Value
	}
	dev.MACState.DesiredParameters.Rx2Frequency = dev.MACState.CurrentParameters.Rx2Frequency
	if dev.GetMACSettings().GetDesiredRx2Frequency() != nil && dev.MACSettings.Rx2Frequency.Value != 0 {
		dev.MACState.DesiredParameters.Rx2Frequency = dev.MACSettings.DesiredRx2Frequency.Value
	} else if fp.Rx2Channel != nil {
		dev.MACState.DesiredParameters.Rx2Frequency = fp.Rx2Channel.Frequency
	} else if defaults.DesiredRx2Frequency != nil && defaults.DesiredRx2Frequency.Value != 0 {
		dev.MACState.DesiredParameters.Rx2Frequency = defaults.DesiredRx2Frequency.Value
	}

	dev.MACState.CurrentParameters.MaxDutyCycle = ttnpb.DUTY_CYCLE_1
	if dev.GetMACSettings().GetMaxDutyCycle() != nil {
		dev.MACState.CurrentParameters.MaxDutyCycle = dev.MACSettings.MaxDutyCycle.Value
	}
	dev.MACState.DesiredParameters.MaxDutyCycle = dev.MACState.CurrentParameters.MaxDutyCycle

	dev.MACState.CurrentParameters.RejoinTimePeriodicity = ttnpb.REJOIN_TIME_0
	dev.MACState.DesiredParameters.RejoinTimePeriodicity = dev.MACState.CurrentParameters.RejoinTimePeriodicity

	dev.MACState.CurrentParameters.RejoinCountPeriodicity = ttnpb.REJOIN_COUNT_16
	dev.MACState.DesiredParameters.RejoinCountPeriodicity = dev.MACState.CurrentParameters.RejoinCountPeriodicity

	dev.MACState.CurrentParameters.PingSlotFrequency = 0
	if dev.GetMACSettings().GetPingSlotFrequency() != nil && dev.MACSettings.PingSlotFrequency.Value != 0 {
		dev.MACState.CurrentParameters.PingSlotFrequency = dev.MACSettings.PingSlotFrequency.Value
	} else if defaults.PingSlotFrequency != nil && defaults.PingSlotFrequency.Value != 0 {
		dev.MACState.CurrentParameters.PingSlotFrequency = defaults.PingSlotFrequency.Value
	}
	dev.MACState.DesiredParameters.PingSlotFrequency = dev.MACState.CurrentParameters.PingSlotFrequency
	if fp.PingSlot != nil && fp.PingSlot.Frequency != 0 {
		dev.MACState.DesiredParameters.PingSlotFrequency = fp.PingSlot.Frequency
	}

	dev.MACState.CurrentParameters.PingSlotDataRateIndex = ttnpb.DATA_RATE_0
	if dev.GetMACSettings().GetPingSlotDataRateIndex() != nil {
		dev.MACState.CurrentParameters.PingSlotDataRateIndex = dev.MACSettings.PingSlotDataRateIndex.Value
	} else if defaults.PingSlotDataRateIndex != nil {
		dev.MACState.CurrentParameters.PingSlotDataRateIndex = defaults.PingSlotDataRateIndex.Value
	}
	dev.MACState.DesiredParameters.PingSlotDataRateIndex = dev.MACState.CurrentParameters.PingSlotDataRateIndex
	if fp.DefaultPingSlotDataRate != nil {
		dev.MACState.DesiredParameters.PingSlotDataRateIndex = ttnpb.DataRateIndex(*fp.DefaultPingSlotDataRate)
	}

	dev.MACState.CurrentParameters.BeaconFrequency = 0
	dev.MACState.DesiredParameters.BeaconFrequency = dev.MACState.CurrentParameters.BeaconFrequency

	if len(band.DownlinkChannels) > len(band.UplinkChannels) || len(fp.DownlinkChannels) > len(fp.UplinkChannels) ||
		len(band.UplinkChannels) > int(band.MaxUplinkChannels) || len(band.DownlinkChannels) > int(band.MaxDownlinkChannels) ||
		len(fp.UplinkChannels) > int(band.MaxUplinkChannels) || len(fp.DownlinkChannels) > int(band.MaxDownlinkChannels) {
		// NOTE: In case the spec changes and this assumption is not valid anymore,
		// the implementation of this function won't be valid and has to be changed.
		panic("uplink/downlink channel length is inconsistent")
	}

	if len(dev.GetMACSettings().GetFactoryPresetFrequencies()) > 0 {
		dev.MACState.CurrentParameters.Channels = make([]*ttnpb.MACParameters_Channel, 0, len(dev.MACSettings.FactoryPresetFrequencies))
		for _, freq := range dev.MACSettings.FactoryPresetFrequencies {
			dev.MACState.CurrentParameters.Channels = append(dev.MACState.CurrentParameters.Channels, &ttnpb.MACParameters_Channel{
				MinDataRateIndex:  0,
				MaxDataRateIndex:  ttnpb.DATA_RATE_15,
				UplinkFrequency:   freq,
				DownlinkFrequency: freq,
				EnableUplink:      true,
			})
		}
	} else if len(defaults.GetFactoryPresetFrequencies()) > 0 {
		dev.MACState.CurrentParameters.Channels = make([]*ttnpb.MACParameters_Channel, 0, len(defaults.FactoryPresetFrequencies))
		for _, freq := range defaults.FactoryPresetFrequencies {
			dev.MACState.CurrentParameters.Channels = append(dev.MACState.CurrentParameters.Channels, &ttnpb.MACParameters_Channel{
				MinDataRateIndex:  0,
				MaxDataRateIndex:  ttnpb.DATA_RATE_15,
				UplinkFrequency:   freq,
				DownlinkFrequency: freq,
				EnableUplink:      true,
			})
		}
	} else {
		dev.MACState.CurrentParameters.Channels = make([]*ttnpb.MACParameters_Channel, 0, len(band.UplinkChannels))
		for _, upCh := range band.UplinkChannels {
			dev.MACState.CurrentParameters.Channels = append(dev.MACState.CurrentParameters.Channels, &ttnpb.MACParameters_Channel{
				MinDataRateIndex: upCh.MinDataRate,
				MaxDataRateIndex: upCh.MaxDataRate,
				UplinkFrequency:  upCh.Frequency,
				EnableUplink:     true,
			})
		}
		for i, downCh := range band.DownlinkChannels {
			dev.MACState.CurrentParameters.Channels[i].DownlinkFrequency = downCh.Frequency
		}
	}

	dev.MACState.DesiredParameters.Channels = make([]*ttnpb.MACParameters_Channel, 0, len(band.UplinkChannels)+len(fp.UplinkChannels))
	for _, upCh := range band.UplinkChannels {
		dev.MACState.DesiredParameters.Channels = append(dev.MACState.DesiredParameters.Channels, &ttnpb.MACParameters_Channel{
			MinDataRateIndex: upCh.MinDataRate,
			MaxDataRateIndex: upCh.MaxDataRate,
			UplinkFrequency:  upCh.Frequency,
		})
	}
	for i, downCh := range band.DownlinkChannels {
		dev.MACState.DesiredParameters.Channels[i].DownlinkFrequency = downCh.Frequency
	}

outerUp:
	for _, upCh := range fp.UplinkChannels {
		for _, ch := range dev.MACState.DesiredParameters.Channels {
			if ch.UplinkFrequency == upCh.Frequency {
				ch.MinDataRateIndex = ttnpb.DataRateIndex(upCh.MinDataRate)
				ch.MaxDataRateIndex = ttnpb.DataRateIndex(upCh.MaxDataRate)
				ch.EnableUplink = true
				continue outerUp
			}
		}

		dev.MACState.DesiredParameters.Channels = append(dev.MACState.DesiredParameters.Channels, &ttnpb.MACParameters_Channel{
			MinDataRateIndex: ttnpb.DataRateIndex(upCh.MinDataRate),
			MaxDataRateIndex: ttnpb.DataRateIndex(upCh.MaxDataRate),
			UplinkFrequency:  upCh.Frequency,
			EnableUplink:     true,
		})
	}

outerDown:
	for _, downCh := range fp.DownlinkChannels {
		for _, ch := range dev.MACState.DesiredParameters.Channels {
			switch ch.DownlinkFrequency {
			case 0:
				ch.DownlinkFrequency = downCh.Frequency
				continue outerDown

			case downCh.Frequency:
				continue outerDown
			}
		}
		panic("uplink/downlink channel length is inconsistent")
	}

	return nil
}
