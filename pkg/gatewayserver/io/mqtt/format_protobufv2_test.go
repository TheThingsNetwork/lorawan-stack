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

package mqtt_test

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	ttnpbv2 "go.thethings.network/lorawan-stack-legacy/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver/io/mqtt"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestProtobufV2Downlink(t *testing.T) {
	a := assertions.New(t)
	pld, _ := base64.RawStdEncoding.DecodeString("YHBhYUoAAgABj9/clY414A")
	ids := ttnpb.GatewayIdentifiers{
		GatewayID: "gateway-id",
	}
	input := &ttnpb.DownlinkMessage{
		RawPayload: pld,
		Payload:    &ttnpb.Message{},
		Settings: &ttnpb.DownlinkMessage_Scheduled{
			Scheduled: &ttnpb.TxSettings{
				DataRate: ttnpb.DataRate{
					Modulation: &ttnpb.DataRate_LoRa{
						LoRa: &ttnpb.LoRaDataRate{
							Bandwidth:       125000,
							SpreadingFactor: 12,
						},
					},
				},
				CodingRate:    "4/5",
				DataRateIndex: 0,
				Frequency:     863000000,
				Downlink: &ttnpb.TxSettings_Downlink{
					TxPower: 16.15,
				},
				Timestamp: 12000,
			},
		},
	}
	err := lorawan.UnmarshalMessage(pld, input.Payload)
	if !a.So(err, should.BeNil) {
		t.Fatal("Could not unmarshal the downlink payload in the v3 payload struct")
	}

	Expected := &ttnpbv2.DownlinkMessage{
		Payload: pld,
		GatewayConfiguration: ttnpbv2.GatewayTxConfiguration{
			Frequency:             863000000,
			Power:                 14,
			PolarizationInversion: true,
			Timestamp:             12000,
		},
		ProtocolConfiguration: ttnpbv2.ProtocolTxConfiguration{
			LoRaWAN: &ttnpbv2.LoRaWANTxConfiguration{
				CodingRate: "4/5",
				DataRate:   "SF12BW125",
				FCnt:       2,
				Modulation: ttnpbv2.Modulation_LORA,
			},
		},
	}
	expectedBuf, err := Expected.Marshal()
	if !a.So(err, should.BeNil) {
		t.Fatal("Could not marshal the v2 struct")
	}

	actualBuf, err := mqtt.ProtobufV2.FromDownlink(input, ids)
	if !a.So(err, should.BeNil) {
		t.Fatal("Could not marshal the input v3 struct")
	}
	a.So(actualBuf, should.Resemble, expectedBuf)
}

func TestProtobufV2Uplinks(t *testing.T) {
	validV2Settings := ttnpbv2.ProtocolRxMetadata{
		LoRaWAN: &ttnpbv2.LoRaWANMetadata{
			CodingRate:    "4/5",
			DataRate:      "SF7BW125",
			FrequencyPlan: 0,
			Modulation:    ttnpbv2.Modulation_LORA,
		},
	}
	validV3Settings := ttnpb.TxSettings{
		DataRate: ttnpb.DataRate{
			Modulation: &ttnpb.DataRate_LoRa{
				LoRa: &ttnpb.LoRaDataRate{
					Bandwidth:       125000,
					SpreadingFactor: 7,
				},
			},
		},
		CodingRate:    "4/5",
		DataRateIndex: 5,
	}
	validV2Metadata := ttnpbv2.GatewayRxMetadata{
		GatewayID: "gateway-id",
		RSSI:      -2,
		SNR:       -75,
		Timestamp: 1000,
	}
	validV2RSigMetadata := ttnpbv2.GatewayRxMetadata{
		GatewayID: validV2Metadata.GatewayID,
		Antennas: []*ttnpbv2.GatewayRxMetadata_Antenna{
			{
				RSSI: validV2Metadata.RSSI,
				SNR:  validV2Metadata.SNR,
			},
		},
		Timestamp: validV2Metadata.Timestamp,
	}
	nilTime := time.Unix(0, 0)
	ids := ttnpb.GatewayIdentifiers{
		GatewayID: "gateway-id",
	}
	validV3Metadata := []*ttnpb.RxMetadata{
		{
			GatewayIdentifiers: ids,
			RSSI:               -2,
			SNR:                -75,
			Time:               &nilTime,
			Timestamp:          1000,
		},
	}
	validRawPayload := []byte{
		0x00, 0x31, 0x46, 0x52, 0x41, 0x44, 0xB2, 0x18, 0x00, 0x74, 0x0A, 0x00,
		0x00, 0x00, 0xB2, 0x18, 0x00, 0xD7, 0x43, 0x9A, 0xF3, 0xA5, 0x9B,
	}

	for _, tc := range []struct {
		Name           string
		Input          *ttnpbv2.UplinkMessage
		InputPayload   []byte
		Expected       *ttnpb.UplinkMessage
		ErrorAssertion func(error) bool
	}{
		{
			Name:           "empty Input",
			Input:          &ttnpbv2.UplinkMessage{},
			ErrorAssertion: func(err error) bool { return err != nil },
		},
		{
			Name: "correct Input",
			Input: &ttnpbv2.UplinkMessage{
				GatewayMetadata:  validV2Metadata,
				ProtocolMetadata: validV2Settings,
			},
			InputPayload: validRawPayload,
			Expected: &ttnpb.UplinkMessage{
				Settings:   validV3Settings,
				RxMetadata: validV3Metadata,
			},
		},
		{
			Name: "correct Input with Rsig",
			Input: &ttnpbv2.UplinkMessage{
				GatewayMetadata:  validV2RSigMetadata,
				ProtocolMetadata: validV2Settings,
			},
			InputPayload: validRawPayload,
			Expected: &ttnpb.UplinkMessage{
				Settings:   validV3Settings,
				RxMetadata: validV3Metadata,
			},
		},
		{
			Name: "incorrect data rate",
			Input: &ttnpbv2.UplinkMessage{
				GatewayMetadata: validV2Metadata,
				ProtocolMetadata: ttnpbv2.ProtocolRxMetadata{
					LoRaWAN: &ttnpbv2.LoRaWANMetadata{
						CodingRate:    validV2Settings.LoRaWAN.CodingRate,
						DataRate:      "SF3BW000",
						FrequencyPlan: validV2Settings.LoRaWAN.FrequencyPlan,
						Modulation:    validV2Settings.LoRaWAN.Modulation,
					},
				},
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
		{
			Name: "incorrect modulation",
			Input: &ttnpbv2.UplinkMessage{
				GatewayMetadata: validV2Metadata,
				ProtocolMetadata: ttnpbv2.ProtocolRxMetadata{
					LoRaWAN: &ttnpbv2.LoRaWANMetadata{
						CodingRate:    validV2Settings.LoRaWAN.CodingRate,
						DataRate:      validV2Settings.LoRaWAN.DataRate,
						FrequencyPlan: validV2Settings.LoRaWAN.FrequencyPlan,
						Modulation:    ttnpbv2.Modulation(3252),
					},
				},
			},
			ErrorAssertion: errors.IsInvalidArgument,
		},
	} {
		tcok := t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			buf, err := tc.Input.Marshal()
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			res, err := mqtt.ProtobufV2.ToUplink(buf, ids)
			if tc.ErrorAssertion != nil {
				if !a.So(tc.ErrorAssertion(err), should.BeTrue) {
					t.FailNow()
				}
				return
			}
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(res, should.HaveEmptyDiff, tc.Expected)
		})
		if !tcok {
			t.FailNow()
		}
	}
}

func TestProtobufV2Status(t *testing.T) {
	ids := ttnpb.GatewayIdentifiers{
		GatewayID: "gateway-id",
	}
	for _, tc := range []struct {
		Name           string
		input          *ttnpbv2.StatusMessage
		Expected       *ttnpb.GatewayStatus
		ErrorAssertion func(error) bool
	}{
		{
			Name: "Simple",
			input: &ttnpbv2.StatusMessage{
				TxIn: 5,
				TxOk: 3,
				RxIn: 15,
				RxOk: 14,
			},
			Expected: &ttnpb.GatewayStatus{
				BootTime: time.Unix(0, 0),
				Time:     time.Unix(0, 0),
				Metrics: map[string]float32{
					"lmnw": 0.0,
					"lmst": 0.0,
					"lmok": 0.0,
					"lpps": 0.0,
					"rxin": 15.0,
					"rxok": 14.0,
					"txin": 5.0,
					"txok": 3.0,
				},
			},
		},
		{
			Name: "With versions",
			input: &ttnpbv2.StatusMessage{
				DSP:  3,
				HAL:  "v1.1",
				FPGA: 4,
			},
			Expected: &ttnpb.GatewayStatus{
				BootTime: time.Unix(0, 0),
				Time:     time.Unix(0, 0),
				Metrics: map[string]float32{
					"lmnw": 0.0,
					"lmst": 0.0,
					"lmok": 0.0,
					"lpps": 0.0,
					"rxin": 0.0,
					"rxok": 0.0,
					"txin": 0.0,
					"txok": 0.0,
				},
				Versions: map[string]string{
					"dsp":  "3",
					"fpga": "4",
					"hal":  "v1.1",
				},
			},
		},
		{
			Name: "With metrics",
			input: &ttnpbv2.StatusMessage{
				OS: &ttnpbv2.StatusMessage_OSMetrics{
					CPUPercentage:    10.0,
					Load_1:           30.0,
					Load_5:           40.0,
					Load_15:          50.0,
					MemoryPercentage: 20.0,
					Temperature:      30.0,
				},
				RTT: 3,
			},
			Expected: &ttnpb.GatewayStatus{
				BootTime: time.Unix(0, 0),
				Time:     time.Unix(0, 0),
				Metrics: map[string]float32{
					"lmnw":              0.0,
					"lmst":              0.0,
					"lmok":              0.0,
					"lpps":              0.0,
					"rxin":              0.0,
					"rxok":              0.0,
					"txin":              0.0,
					"txok":              0.0,
					"cpu_percentage":    10.0,
					"load_1":            30.0,
					"load_5":            40.0,
					"load_15":           50.0,
					"memory_percentage": 20.0,
					"temp":              30.0,
					"rtt_ms":            3.0,
				},
			},
		},
		{
			Name: "With location",
			input: &ttnpbv2.StatusMessage{
				Location: &ttnpbv2.LocationMetadata{
					Latitude:  10,
					Longitude: 10,
					Altitude:  10,
					Source:    ttnpbv2.LocationMetadata_GPS,
				},
			},
			Expected: &ttnpb.GatewayStatus{
				BootTime: time.Unix(0, 0),
				Time:     time.Unix(0, 0),
				Metrics: map[string]float32{
					"lmnw": 0.0,
					"lmst": 0.0,
					"lmok": 0.0,
					"lpps": 0.0,
					"rxin": 0.0,
					"rxok": 0.0,
					"txin": 0.0,
					"txok": 0.0,
				},
				AntennaLocations: []*ttnpb.Location{
					{
						Altitude:  10,
						Latitude:  10.0,
						Longitude: 10.0,
						Source:    ttnpb.SOURCE_GPS,
					},
				},
			},
		},
	} {
		tcok := t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			buf, err := tc.input.Marshal()
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			res, err := mqtt.ProtobufV2.ToStatus(buf, ids)
			if tc.ErrorAssertion != nil {
				if !a.So(tc.ErrorAssertion(err), should.BeTrue) {
					t.FailNow()
				}
				return
			}
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			a.So(res, should.HaveEmptyDiff, tc.Expected)
		})
		if !tcok {
			t.FailNow()
		}
	}
}
