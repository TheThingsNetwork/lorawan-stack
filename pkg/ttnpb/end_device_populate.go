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
	"sort"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/mohae/deepcopy"
	"go.thethings.network/lorawan-stack/pkg/types"
)

func NewPopulatedEndDeviceVersion(r randyEndDevice, easy bool) *EndDeviceVersion {
	out := &EndDeviceVersion{}
	out.EndDeviceVersionIdentifiers = *NewPopulatedEndDeviceVersionIdentifiers(r, easy)
	out.Photos = []string{randStringEndDevice(r) + ".jpg", randStringEndDevice(r) + ".jpg"}
	if r.Intn(10) != 0 {
		out.DefaultFormatters = *NewPopulatedMessagePayloadFormatters(r, easy)
	}
	out.MaxFrequency = uint64(r.Uint32())
	out.MinFrequency = uint64(r.Uint32()) % out.MaxFrequency
	out.ResetsJoinNonces = bool(r.Intn(2) == 0)
	out.DefaultMACSettings = NewPopulatedMACSettings(r, easy)
	return out
}

func NewPopulatedMACState(r randyEndDevice, easy bool) *MACState {
	out := &MACState{}
	out.DeviceClass = Class([]int32{0, 1, 2}[r.Intn(3)])
	out.LoRaWANVersion = MACVersion([]int32{1, 2, 3, 4}[r.Intn(4)])
	out.PingSlotPeriodicity = PingSlotPeriod([]int32{0, 1, 2, 3, 4, 5, 6, 7}[r.Intn(8)])
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

func NewPopulatedEndDevice(r randyEndDevice, easy bool) *EndDevice {
	out := &EndDevice{}
	out.EndDeviceIdentifiers = *NewPopulatedEndDeviceIdentifiers(r, easy)
	out.RootKeys = NewPopulatedRootKeys(r, easy)

	out.LastDevNonce = r.Uint32()
	out.LastJoinNonce = r.Uint32()
	out.LastRJCount0 = r.Uint32()
	out.LastRJCount1 = r.Uint32()

	out.UsedDevNonces = make([]uint32, r.Intn(100))
	for i := range out.UsedDevNonces {
		out.UsedDevNonces[i] = r.Uint32()
	}
	sort.Slice(out.UsedDevNonces, func(i, j int) bool { return out.UsedDevNonces[i] < out.UsedDevNonces[j] })

	if r.Intn(10) != 0 {
		out.Session = NewPopulatedSession(r, easy)
	}
	if out.Session != nil {
		out.EndDeviceIdentifiers.DevAddr = &types.DevAddr{}
		copy(out.EndDeviceIdentifiers.DevAddr[:], out.Session.DevAddr[:])
	}

	out.LoRaWANVersion = MAC_V1_1
	out.LoRaWANPHYVersion = PHY_V1_1_REV_B
	out.FrequencyPlanID = "EU_863_870"
	out.MACSettings = NewPopulatedMACSettings(r, easy)
	out.MACState = NewPopulatedMACState(r, easy)
	out.MACState.CurrentParameters.Channels = []*MACParameters_Channel{
		{
			UplinkFrequency:   868100000,
			DownlinkFrequency: 868100000,
			MinDataRateIndex:  0,
			MaxDataRateIndex:  5,
			EnableUplink:      true,
		},
		{
			UplinkFrequency:   868300000,
			DownlinkFrequency: 868300000,
			MinDataRateIndex:  0,
			MaxDataRateIndex:  5,
			EnableUplink:      true,
		},
		{
			UplinkFrequency:   868500000,
			DownlinkFrequency: 868500000,
			MinDataRateIndex:  0,
			MaxDataRateIndex:  5,
			EnableUplink:      true,
		},
	}
	out.MACState.DesiredParameters.Channels = deepcopy.Copy(out.MACState.CurrentParameters.Channels).([]*MACParameters_Channel)
	out.NetworkServerAddress = randStringEndDevice(r)
	out.ApplicationServerAddress = randStringEndDevice(r)
	out.VersionIDs = NewPopulatedEndDeviceVersionIdentifiers(r, easy)
	if r.Intn(10) != 0 {
		out.RecentUplinks = make([]*UplinkMessage, r.Intn(5))
		for i := range out.RecentUplinks {
			out.RecentUplinks[i] = NewPopulatedUplinkMessage(r, easy)
		}
	}
	if r.Intn(10) != 0 {
		out.RecentDownlinks = make([]*DownlinkMessage, r.Intn(5))
		for i := range out.RecentDownlinks {
			out.RecentDownlinks[i] = NewPopulatedDownlinkMessage(r, easy)
		}
	}
	if r.Intn(10) != 0 {
		out.QueuedApplicationDownlinks = make([]*ApplicationDownlink, r.Intn(5))
		for i := range out.QueuedApplicationDownlinks {
			out.QueuedApplicationDownlinks[i] = NewPopulatedApplicationDownlink(r, easy)
		}
	}
	out.LastDevStatusReceivedAt = pbtypes.NewPopulatedStdTime(r, easy)
	if r.Intn(2) == 0 {
		out.BatteryPercentage = &pbtypes.FloatValue{
			Value: float32(r.Intn(100)) / 100,
		}
	}
	out.DownlinkMargin = r.Int31()
	if r.Intn(2) == 0 {
		out.DownlinkMargin *= -1
	}
	out.Formatters = NewPopulatedMessagePayloadFormatters(r, easy)
	out.SupportsJoin = r.Intn(2) == 0
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
		out.UplinkDwellTime = pbtypes.NewPopulatedBoolValue(r, easy)
	}
	if r.Intn(2) == 0 {
		out.DownlinkDwellTime = pbtypes.NewPopulatedBoolValue(r, easy)
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
	out.Rx1DataRateOffset = r.Uint32() % 6
	out.Rx2DataRateIndex = NewPopulatedDataRateIndex(r, easy)
	out.Rx2Frequency = NewPopulatedFrequency(r, easy)
	out.RejoinTimePeriodicity = RejoinTimeExponent([]int32{0, 1, 2, 3, 4, 5, 6, 7}[r.Intn(8)])
	out.PingSlotFrequency = NewPopulatedFrequency(r, easy)
	out.PingSlotDataRateIndex = NewPopulatedDataRateIndex(r, easy)
	out.BeaconFrequency = NewPopulatedFrequency(r, easy)
	return out
}
