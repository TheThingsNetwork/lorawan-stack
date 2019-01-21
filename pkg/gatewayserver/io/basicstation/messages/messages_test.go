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

package messages

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/frequencyplans"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestType(t *testing.T) {
	a := assertions.New(t)
	msg := Version{
		Station:  "test",
		Firmware: "2.0.0",
		Package:  "test",
		Model:    "test",
		Protocol: 2,
	}

	data, err := json.Marshal(msg)
	a.So(err, should.BeNil)

	mt, err := Type(data)
	a.So(err, should.BeNil)
	a.So(mt, should.Equal, TypeUpstreamVersion)
}

func TestIsProduction(t *testing.T) {
	a := assertions.New(t)
	for _, tc := range []struct {
		Name             string
		Message          Version
		ExpectedResponse bool
	}{
		{
			"EmptyMessage",
			Version{},
			false,
		},
		{
			"EmptyMessage1",
			Version{
				Features: []string{""},
			},
			false,
		},
		{
			"NonProduction",
			Version{
				Features: []string{"gps", "rmtsh"},
			},
			false,
		},
		{
			"Production",
			Version{
				Features: []string{"prod"},
			},
			true,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a.So(tc.Message.IsProduction(), should.Equal, tc.ExpectedResponse)
		})
	}

}

func TestGetRouterConfig(t *testing.T) {
	a := assertions.New(t)
	for _, tc := range []struct {
		Name          string
		FP            frequencyplans.FrequencyPlan
		Cfg           RouterConfig
		IsProd        bool
		ExpectedError error
	}{
		{
			"NilFrequencyPlan",
			frequencyplans.FrequencyPlan{},
			RouterConfig{},
			false,
			errFrequencyPlan,
		},
		{
			"InvalidFrequencyPlan",
			frequencyplans.FrequencyPlan{
				BandID: "PinkFloyd",
			},
			RouterConfig{},
			false,
			errFrequencyPlan,
		},
		{
			"ValidFrequencyPlan",
			frequencyplans.FrequencyPlan{
				BandID: "US_902_928",
				Radios: []frequencyplans.Radio{
					{
						Enable:    true,
						ChipType:  "SX1257",
						Frequency: 922300000,
						TxConfiguration: &frequencyplans.RadioTxConfiguration{
							MinFrequency: 909000000,
							MaxFrequency: 925000000,
						},
					},
					{
						Enable:    false,
						ChipType:  "SX1257",
						Frequency: 923000000,
					},
				},
			},
			RouterConfig{
				Region:         "US902",
				HardwareSpec:   "sx1301/1",
				FrequencyRange: []int{909000000, 925000000},
				DataRates: DataRates{
					[3]int{10, 125, 0},
					[3]int{9, 125, 0},
					[3]int{8, 125, 0},
					[3]int{7, 125, 0},
					[3]int{8, 500, 0},
					[3]int{12, 500, 0},
					[3]int{11, 500, 0},
					[3]int{10, 500, 0},
					[3]int{9, 500, 0},
					[3]int{8, 500, 0},
					[3]int{7, 500, 0},
				},
				NoCCA:       true,
				NoDutyCycle: true,
				NoDwellTime: true,
			},
			false,
			nil,
		},
		{
			"ValidFrequencyPlanProd",
			frequencyplans.FrequencyPlan{
				BandID: "US_902_928",
				Radios: []frequencyplans.Radio{
					{
						Enable:    true,
						ChipType:  "SX1257",
						Frequency: 922300000,
						TxConfiguration: &frequencyplans.RadioTxConfiguration{
							MinFrequency: 909000000,
							MaxFrequency: 925000000,
						},
					},
					{
						Enable:    false,
						ChipType:  "SX1257",
						Frequency: 923000000,
					},
				},
			},
			RouterConfig{
				Region:         "US902",
				HardwareSpec:   "sx1301/1",
				FrequencyRange: []int{909000000, 925000000},
				DataRates: DataRates{
					[3]int{10, 125, 0},
					[3]int{9, 125, 0},
					[3]int{8, 125, 0},
					[3]int{7, 125, 0},
					[3]int{8, 500, 0},
					[3]int{12, 500, 0},
					[3]int{11, 500, 0},
					[3]int{10, 500, 0},
					[3]int{9, 500, 0},
					[3]int{8, 500, 0},
					[3]int{7, 500, 0},
				},
			},
			true,
			nil,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			cfg, err := GetRouterConfig(tc.FP, tc.IsProd)
			if !(a.So(err, should.Resemble, tc.ExpectedError)) {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !(a.So(cfg, should.Resemble, tc.Cfg)) {
				t.Fatalf("Invalid config: %v", cfg)
			}
		})
	}
}

func TestGetUplinkMessage(t *testing.T) {
	a := assertions.New(t)
	for _, tc := range []struct {
		Name                  string
		JoinRequest           JoinRequest
		GatewayIDs            ttnpb.GatewayIdentifiers
		FreqPlanID            string
		ExpectedUplinkMessage ttnpb.UplinkMessage
		ExpectedError         error
	}{
		{
			"EmptyJoinRequest",
			JoinRequest{},
			ttnpb.GatewayIdentifiers{GatewayID: "test-gateway"},
			"EU_863_870",
			ttnpb.UplinkMessage{
				Payload: &ttnpb.Message{
					MIC:  []byte{0, 0, 0, 0},
					MHDR: ttnpb.MHDR{MType: ttnpb.MType_JOIN_REQUEST, Major: ttnpb.Major_LORAWAN_R1},
					Payload: &ttnpb.Message_JoinRequestPayload{JoinRequestPayload: &ttnpb.JoinRequestPayload{
						JoinEUI:  types.EUI64{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
						DevEUI:   types.EUI64{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
						DevNonce: [2]byte{0x00, 0x00},
					}}},
				RxMetadata: []*ttnpb.RxMetadata{&ttnpb.RxMetadata{
					GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "test-gateway"},
					Time:               &[]time.Time{time.Unix(0, 0)}[0],
				}},
				Settings: ttnpb.TxSettings{
					CodingRate: "4/5",
					DataRate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
						SpreadingFactor: 12,
						Bandwidth:       125000,
					}}}},
			},
			nil,
		},
		{
			"ValidJoinRequest",
			JoinRequest{
				MHdr:     0,
				DevEUI:   EUI{Prefix: "DevEui", EUI64: types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}},
				JoinEUI:  EUI{Prefix: "JoinEui", EUI64: types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}},
				DevNonce: 18000,
				MIC:      12345678,
				RadioMetaData: RadioMetaData{
					DataRate:  1,
					Frequency: 868300000,
					UpInfo: UpInfo{
						RxTime: 1548059982,
						XTime:  12666373963464220,
						RSSI:   89,
						SNR:    9.25,
					},
				},
			},
			ttnpb.GatewayIdentifiers{GatewayID: "test-gateway"},
			"EU_863_870",
			ttnpb.UplinkMessage{
				Payload: &ttnpb.Message{
					MHDR: ttnpb.MHDR{MType: ttnpb.MType_JOIN_REQUEST, Major: ttnpb.Major_LORAWAN_R1},
					MIC:  []byte{0x4E, 0x61, 0xBC, 0x00},
					Payload: &ttnpb.Message_JoinRequestPayload{JoinRequestPayload: &ttnpb.JoinRequestPayload{
						JoinEUI:  types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22},
						DevEUI:   types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11},
						DevNonce: [2]byte{0x50, 0x46},
					}}},
				RxMetadata: []*ttnpb.RxMetadata{&ttnpb.RxMetadata{
					GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "test-gateway"},
					Time:               &[]time.Time{time.Unix(1548059982, 0)}[0],
					Timestamp:          (uint32)(12666373963464220 & 0xFFFFFFFF),
					RSSI:               89,
					SNR:                9.25,
				},
				},
				Settings: ttnpb.TxSettings{
					CodingRate:    "4/5",
					Frequency:     868300000,
					DataRateIndex: 1,
					DataRate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
						SpreadingFactor: 11,
						Bandwidth:       125000,
					}}},
				},
			},
			nil,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			msg, err := tc.JoinRequest.ToUplinkMessage(tc.GatewayIDs, tc.FreqPlanID)
			if !(a.So(err, should.Resemble, tc.ExpectedError)) {
				t.Fatalf("Unexpected error: %v", err)
			}
			msg.ReceivedAt = time.Time{}
			var payload ttnpb.Message
			a.So(lorawan.UnmarshalMessage(msg.RawPayload, &payload), should.BeNil)
			if !a.So(&payload, should.Resemble, msg.Payload) {
				t.Fatalf("Invalid RawPayload: %v", msg.RawPayload)
			}
			msg.RawPayload = nil
			if !(a.So(msg, should.Resemble, tc.ExpectedUplinkMessage)) {
				t.Fatalf("Invalid UplinkMessage: %s", msg.RawPayload)
			}
		})
	}
}
