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
	out.FCntResets = r.Intn(2) == 0
	out.FCntIs16Bit = r.Intn(2) == 0
	if r.Intn(10) != 0 {
		out.Session = NewPopulatedSession(r, easy)
	}
	if r.Intn(10) != 0 {
		out.SessionFallback = NewPopulatedSession(r, easy)
	}
	switch r.Intn(6) {
	case 0:
		out.LoRaWANVersion = MAC_V1_0
		out.LoRaWANPHYVersion = PHY_V1_0
	case 1:
		out.LoRaWANVersion = MAC_V1_0_1
		out.LoRaWANPHYVersion = PHY_V1_0_1
	case 2:
		out.LoRaWANVersion = MAC_V1_0_2
		out.LoRaWANPHYVersion = PHY_V1_0_2_REV_A
	case 3:
		out.LoRaWANVersion = MAC_V1_0_2
		out.LoRaWANPHYVersion = PHY_V1_0_2_REV_B
	case 4:
		out.LoRaWANVersion = MAC_V1_1
		out.LoRaWANPHYVersion = PHY_V1_1_REV_A
	case 5:
		out.LoRaWANVersion = MAC_V1_1
		out.LoRaWANPHYVersion = PHY_V1_1_REV_B
	}
	out.FrequencyPlanID = "EU_863_870"
	out.MinFrequency = uint64(r.Uint32())
	out.MaxFrequency = uint64(r.Uint32())
	out.MaxEIRP = r.Float32()
	out.MACSettings = NewPopulatedMACSettings(r, easy)
	out.MACInfo = NewPopulatedMACInfo(r, easy)
	out.MACState = NewPopulatedMACState(r, easy)
	out.MACStateDesired = NewPopulatedMACState(r, easy)
	if r.Intn(10) != 0 {
		out.Location = NewPopulatedLocation(r, easy)
	}
	if r.Intn(10) != 0 {
		out.Attributes = pbtypes.NewPopulatedStruct(r, easy)
	}
	out.DisableJoinNonceCheck = r.Intn(2) == 0
	out.NetworkServerAddress = randStringEndDevice(r)
	out.ApplicationServerAddress = randStringEndDevice(r)
	out.EndDeviceVersion = NewPopulatedEndDeviceVersion(r, easy)
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
		out.QueuedMACResponses = make([]*MACCommand, r.Intn(5))
		for i := range out.QueuedMACResponses {
			out.QueuedMACResponses[i] = NewPopulatedMACCommand(r, easy)
		}
	}
	if r.Intn(10) != 0 {
		out.PendingMACRequests = make([]*MACCommand, r.Intn(5))
		for i := range out.PendingMACRequests {
			out.PendingMACRequests[i] = NewPopulatedMACCommand(r, easy)
		}
	}
	if r.Intn(10) != 0 {
		out.QueuedApplicationDownlinks = make([]*ApplicationDownlink, r.Intn(5))
		for i := range out.QueuedApplicationDownlinks {
			out.QueuedApplicationDownlinks[i] = NewPopulatedApplicationDownlink(r, easy)
		}
	}
	out.DeviceFormatters = *NewPopulatedDeviceFormatters(r, easy)
	return out
}

func NewPopulatedMACState(r randyEndDevice, _ bool) *MACState {
	out := &MACState{}
	out.MaxEIRP = r.Float32()
	out.UplinkDwellTime = r.Intn(2) == 0
	out.DownlinkDwellTime = r.Intn(2) == 0
	out.ADRDataRateIndex = r.Uint32()
	out.ADRTXPowerIndex = r.Uint32()
	out.ADRNbTrans = r.Uint32()
	out.ADRAckLimit = r.Uint32()
	out.ADRAckDelay = r.Uint32()
	out.DutyCycle = AggregatedDutyCycle([]int32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}[r.Intn(16)])
	out.RxDelay = r.Uint32()
	out.Rx1DataRateOffset = r.Uint32() % 6
	out.Rx2DataRateIndex = r.Uint32() % 16
	out.Rx2Frequency = 868300000
	out.RejoinTimer = r.Uint32()
	out.RejoinCounter = r.Uint32()
	out.PingSlotFrequency = uint64(r.Uint32())
	out.PingSlotDataRateIndex = r.Uint32()
	return out
}
