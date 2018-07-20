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
	pbtypes "github.com/gogo/protobuf/types"
)

func NewPopulatedEndDeviceVersion(r randyEndDevice, easy bool) *EndDeviceVersion {
	out := &EndDeviceVersion{}
	out.BrandID = randStringEndDevice(r)
	out.ModelID = randStringEndDevice(r)
	out.HardwareVersion = randStringEndDevice(r)
	out.FirmwareVersion = randStringEndDevice(r)
	out.Photos = []string{randStringEndDevice(r) + ".jpg", randStringEndDevice(r) + ".jpg"}
	if r.Intn(10) != 0 {
		out.DefaultFormatters = *NewPopulatedEndDeviceFormatters(r, easy)
	}
	out.DefaultMACParameters = NewPopulatedMACParameters(r, easy)
	out.MaxFrequency = uint64(r.Uint32())
	out.MinFrequency = uint64(r.Uint32()) % out.MaxFrequency
	out.FCntResets = bool(r.Intn(2) == 0)
	out.Supports32BitFCnt = bool(r.Intn(2) == 0)
	out.DisableJoinNonceCheck = bool(r.Intn(2) == 0)
	out.LoRaWANVersion = MACVersion([]int32{1, 2, 3, 4}[r.Intn(4)])
	out.LoRaWANPHYVersion = PHYVersion([]int32{1, 2, 3, 4, 5, 6}[r.Intn(6)])
	return out
}

func NewPopulatedMACState(r randyEndDevice, easy bool) *MACState {
	out := &MACState{}
	out.DeviceClass = Class([]int32{0, 1, 2}[r.Intn(3)])
	out.LoRaWANVersion = MACVersion([]int32{1, 2, 3, 4}[r.Intn(4)])
	out.PingSlotPeriodicity = PingSlotPeriod([]int32{0, 1, 2, 3, 4, 5, 6, 7}[r.Intn(8)])
	out.NextConfirmedDownlinkAt = pbtypes.NewPopulatedStdTime(r, easy)
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
	out.MACParameters = *NewPopulatedMACParameters(r, easy)
	out.DesiredMACParameters = *NewPopulatedMACParameters(r, easy)
	if r.Intn(10) != 0 {
		out.PendingApplicationDownlink = NewPopulatedApplicationDownlink(r, easy)
	}
	return out
}

func NewPopulatedEndDevice(r randyEndDevice, easy bool) *EndDevice {
	out := &EndDevice{}
	out.EndDeviceIdentifiers = *NewPopulatedEndDeviceIdentifiers(r, easy)
	out.RootKeys = NewPopulatedRootKeys(r, easy)
	out.NextDevNonce = r.Uint32()
	out.UsedDevNonces = make([]uint32, r.Intn(10))
	for i := range out.UsedDevNonces {
		out.UsedDevNonces[i] = r.Uint32()
	}
	out.NextJoinNonce = r.Uint32()
	out.UsedJoinNonces = make([]uint32, r.Intn(10))
	for i := range out.UsedJoinNonces {
		out.UsedJoinNonces[i] = r.Uint32()
	}
	out.NextRJCount0 = r.Uint32()
	out.NextRJCount1 = r.Uint32()
	if r.Intn(10) != 0 {
		out.Session = NewPopulatedSession(r, easy)
	}
	if r.Intn(10) != 0 {
		out.SessionFallback = NewPopulatedSession(r, easy)
	}
	out.BatteryPercentage = r.Float32()
	out.FrequencyPlanID = "EU_863_870"
	out.MACSettings = NewPopulatedMACSettings(r, easy)
	out.MACState = NewPopulatedMACState(r, easy)
	if r.Intn(10) != 0 {
		out.Location = NewPopulatedLocation(r, easy)
	}
	if r.Intn(10) != 0 {
		out.Attributes = pbtypes.NewPopulatedStruct(r, easy)
	}
	out.NetworkServerAddress = randStringEndDevice(r)
	out.ApplicationServerAddress = randStringEndDevice(r)
	out.EndDeviceVersion = *NewPopulatedEndDeviceVersion(r, easy)
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
	out.StatusUpdatedAt = pbtypes.NewPopulatedStdTime(r, easy)
	out.BatteryPercentage = r.Float32()
	out.DownlinkMargin = r.Int31()
	if r.Intn(2) == 0 {
		out.DownlinkMargin *= -1
	}
	out.EndDeviceFormatters = *NewPopulatedEndDeviceFormatters(r, easy)
	out.SupportsJoin = r.Intn(2) == 0
	return out
}

func NewPopulatedMACParameters(r randyEndDevice, _ bool) *MACParameters {
	out := &MACParameters{}
	out.MaxEIRP = r.Float32()
	out.UplinkDwellTime = r.Intn(2) == 0
	out.DownlinkDwellTime = r.Intn(2) == 0
	out.ADRDataRateIndex = r.Uint32()
	out.ADRTxPowerIndex = r.Uint32()
	out.ADRNbTrans = r.Uint32()
	out.ADRAckLimit = r.Uint32()
	out.ADRAckDelay = r.Uint32()
	out.DutyCycle = AggregatedDutyCycle([]int32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}[r.Intn(16)])
	out.Rx1Delay = r.Uint32()
	out.Rx1DataRateOffset = r.Uint32() % 6
	out.Rx2DataRateIndex = r.Uint32() % 16
	out.Rx2Frequency = 868300000
	out.RejoinTimePeriodicity = RejoinTimePeriod([]int32{0, 1, 2, 3, 4, 5, 6, 7}[r.Intn(8)])
	out.PingSlotFrequency = uint64(r.Uint32())
	out.PingSlotDataRateIndex = r.Uint32()
	return out
}
