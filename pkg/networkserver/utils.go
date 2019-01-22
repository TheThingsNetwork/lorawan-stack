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
	"math"

	"github.com/mohae/deepcopy"
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

func resetMACState(dev *ttnpb.EndDevice, fps *frequencyplans.Store) error {
	fp, band, err := getDeviceBandVersion(dev, fps)
	if err != nil {
		return err
	}

	dev.MACState = &ttnpb.MACState{
		DeviceClass:    dev.DefaultClass,
		LoRaWANVersion: dev.LoRaWANVersion,
		CurrentParameters: ttnpb.MACParameters{
			ADRAckDelay:      uint32(band.ADRAckDelay),
			ADRAckLimit:      uint32(band.ADRAckLimit),
			ADRNbTrans:       1,
			MaxDutyCycle:     ttnpb.DUTY_CYCLE_1,
			MaxEIRP:          band.DefaultMaxEIRP,
			Rx1Delay:         ttnpb.RxDelay(band.ReceiveDelay1.Seconds()),
			Rx2DataRateIndex: band.DefaultRx2Parameters.DataRateIndex,
			Rx2Frequency:     band.DefaultRx2Parameters.Frequency,
		},
	}

	// NOTE: dev.MACState.CurrentParameters must not contain pointer values at this point.
	dev.MACState.DesiredParameters = dev.MACState.CurrentParameters

	if len(band.DownlinkChannels) > len(band.UplinkChannels) || len(fp.DownlinkChannels) > len(fp.UplinkChannels) ||
		len(band.UplinkChannels) > int(band.MaxUplinkChannels) || len(band.DownlinkChannels) > int(band.MaxDownlinkChannels) ||
		len(fp.UplinkChannels) > int(band.MaxUplinkChannels) || len(fp.DownlinkChannels) > int(band.MaxDownlinkChannels) {
		// NOTE: In case the spec changes and this assumption is not valid anymore,
		// the implementation of this function won't be valid and has to be changed.
		panic("uplink/downlink channel length is inconsistent")
	}

	dev.MACState.CurrentParameters.Channels = make([]*ttnpb.MACParameters_Channel, len(band.UplinkChannels))
	dev.MACState.DesiredParameters.Channels = make([]*ttnpb.MACParameters_Channel, 0, len(band.UplinkChannels)+len(fp.UplinkChannels))

	for i, upCh := range band.UplinkChannels {
		ch := &ttnpb.MACParameters_Channel{
			MinDataRateIndex: upCh.MinDataRate,
			MaxDataRateIndex: upCh.MaxDataRate,
			UplinkFrequency:  upCh.Frequency,
			EnableUplink:     true,
		}
		dev.MACState.CurrentParameters.Channels[i] = ch
	}

	for i, downCh := range band.DownlinkChannels {
		// NOTE: len(band.DownlinkChannels) <= len(band.UplinkChannels) => i < len(dev.MACState.CurrentParameters.Channels)
		dev.MACState.CurrentParameters.Channels[i].DownlinkFrequency = downCh.Frequency
	}

	for _, ch := range dev.MACState.CurrentParameters.Channels {
		chCopy := *ch
		chCopy.EnableUplink = false
		dev.MACState.DesiredParameters.Channels = append(dev.MACState.DesiredParameters.Channels, &chCopy)
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

	dev.MACState.DesiredParameters.UplinkDwellTime = fp.DwellTime.GetUplinks()
	dev.MACState.DesiredParameters.DownlinkDwellTime = fp.DwellTime.GetDownlinks()

	if fp.Rx2Channel != nil {
		dev.MACState.DesiredParameters.Rx2Frequency = fp.Rx2Channel.Frequency
	}
	if fp.DefaultRx2DataRate != nil {
		dev.MACState.DesiredParameters.Rx2DataRateIndex = ttnpb.DataRateIndex(*fp.DefaultRx2DataRate)
	}

	if fp.PingSlot != nil {
		dev.MACState.DesiredParameters.PingSlotFrequency = fp.PingSlot.Frequency
	}
	if fp.DefaultPingSlotDataRate != nil {
		dev.MACState.DesiredParameters.PingSlotDataRateIndex = ttnpb.DataRateIndex(*fp.DefaultPingSlotDataRate)
	}

	if fp.MaxEIRP != nil && *fp.MaxEIRP > 0 {
		dev.MACState.DesiredParameters.MaxEIRP = float32(math.Min(float64(dev.MACState.CurrentParameters.MaxEIRP), float64(*fp.MaxEIRP)))
	}

	if dev.DefaultMACParameters != nil {
		dev.MACState.CurrentParameters = deepcopy.Copy(*dev.DefaultMACParameters).(ttnpb.MACParameters)
	}

	return nil
}
