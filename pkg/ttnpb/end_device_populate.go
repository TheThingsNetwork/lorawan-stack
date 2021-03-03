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

package ttnpb

import (
	pbtypes "github.com/gogo/protobuf/types"
)

func NewPopulatedMACState(r randyEndDevice, easy bool) *MACState {
	out := &MACState{}
	out.DeviceClass = Class([]int32{0, 1, 2}[r.Intn(3)])
	out.LoRaWANVersion = MACVersion([]int32{1, 2, 3, 4}[r.Intn(4)])
	if r.Intn(2) == 0 {
		out.PingSlotPeriodicity = &PingSlotPeriodValue{
			Value: PingSlotPeriod([]int32{0, 1, 2, 3, 4, 5, 6, 7}[r.Intn(8)]),
		}
	}
	out.LastConfirmedDownlinkAt = pbtypes.NewPopulatedStdTime(r, easy)
	if r.Intn(10) != 0 {
		out.QueuedResponses = make([]*MACCommand, r.Intn(5))
		for i := range out.QueuedResponses {
			out.QueuedResponses[i] = NewPopulatedMACCommand(r, easy)
		}
	}
	if r.Intn(10) != 0 {
		out.PendingRequests = make([]*MACCommand, r.Intn(5))
		for i := range out.PendingRequests {
			out.PendingRequests[i] = NewPopulatedMACCommand(r, easy)
		}
	}
	out.CurrentParameters = *NewPopulatedMACParameters(r, easy)
	out.DesiredParameters = *NewPopulatedMACParameters(r, easy)
	if r.Intn(10) != 0 {
		out.PendingApplicationDownlink = NewPopulatedApplicationDownlink(r, easy)
	}
	return out
}

func NewPopulatedMACParameters_Channel(r randyEndDevice, easy bool) *MACParameters_Channel {
	drMin := NewPopulatedDataRateIndex(r, easy)
	drMax := NewPopulatedDataRateIndex(r, easy)
	if drMax < drMin {
		drMax, drMin = drMin, drMax
	}
	return &MACParameters_Channel{
		UplinkFrequency:   NewPopulatedFrequency(r, easy),
		DownlinkFrequency: NewPopulatedFrequency(r, easy),
		MinDataRateIndex:  drMin,
		MaxDataRateIndex:  drMax,
	}
}

func NewPopulatedMACParameters(r randyEndDevice, easy bool) *MACParameters {
	out := &MACParameters{}
	out.MaxEIRP = r.Float32()
	if r.Intn(2) == 0 {
		out.UplinkDwellTime = &BoolValue{Value: r.Uint32()%2 == 0}
	}
	if r.Intn(2) == 0 {
		out.DownlinkDwellTime = &BoolValue{Value: r.Uint32()%2 == 0}
	}
	out.ADRDataRateIndex = NewPopulatedDataRateIndex(r, easy)
	out.ADRTxPowerIndex = r.Uint32() % 16
	out.ADRNbTrans = 1 + r.Uint32()%15
	out.ADRAckLimitExponent = NewPopulatedADRAckLimitExponentValue(r, easy)
	out.ADRAckDelayExponent = NewPopulatedADRAckDelayExponentValue(r, easy)
	out.MaxDutyCycle = AggregatedDutyCycle([]int32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}[r.Intn(16)])
	out.Channels = make([]*MACParameters_Channel, 1+r.Intn(254))
	for i := range out.Channels {
		out.Channels[i] = NewPopulatedMACParameters_Channel(r, easy)
	}
	out.Rx1Delay = RxDelay(r.Uint32() % 16)
	out.Rx1DataRateOffset = DataRateOffset(r.Uint32() % 8)
	out.Rx2DataRateIndex = NewPopulatedDataRateIndex(r, easy)
	out.Rx2Frequency = NewPopulatedFrequency(r, easy)
	out.RejoinTimePeriodicity = RejoinTimeExponent([]int32{0, 1, 2, 3, 4, 5, 6, 7}[r.Intn(8)])
	out.PingSlotFrequency = NewPopulatedFrequency(r, easy)
	if r.Intn(2) == 0 {
		out.PingSlotDataRateIndexValue = &DataRateIndexValue{
			Value: NewPopulatedDataRateIndex(r, easy),
		}
		out.PingSlotDataRateIndex = out.PingSlotDataRateIndexValue.Value
	}
	out.BeaconFrequency = NewPopulatedFrequency(r, easy)
	return out
}
