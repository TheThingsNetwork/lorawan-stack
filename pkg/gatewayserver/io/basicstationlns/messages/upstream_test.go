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
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/basicstation"
	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

type message interface {
	MarshalJSON() ([]byte, error)
}

func TestMarshalJSON(t *testing.T) {
	for _, tc := range []struct {
		Name     string
		Message  message
		Expected []byte
	}{
		{
			Name: "JoinRequest",
			Message: JoinRequest{
				MHdr:     0,
				DevEUI:   basicstation.EUI{EUI64: types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}},
				JoinEUI:  basicstation.EUI{EUI64: types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}},
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
			Expected: []byte(`{"msgtype":"jreq","MHdr":0,"JoinEui":"2222:2222:2222:2222","DevEui":"1111:1111:1111:1111","DevNonce":18000,"MIC":12345678,"RefTime":0,"RadioMetaData":{"DR":1,"Freq":868300000,"upinfo":{"rxtime":1548059982,"rtcx":0,"xtime":12666373963464220,"gpstime":0,"rssi":89,"snr":9.25}}}`),
		},
		{
			Name: "UplinkDataFrame",
			Message: UplinkDataFrame{
				MHdr:       0x40,
				DevAddr:    0x11223344,
				FCtrl:      0x30,
				FPort:      0x00,
				FCnt:       25,
				FOpts:      "FD",
				FRMPayload: "Ymxhamthc25kJ3M=",
				MIC:        12345678,
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
			Expected: []byte(`{"msgtype":"updf","MHdr":64,"DevAddr":287454020,"FCtrl":48,"Fcnt":25,"FOpts":"FD","FPort":0,"FRMPayload":"Ymxhamthc25kJ3M=","MIC":12345678,"RefTime":0,"RadioMetaData":{"DR":1,"Freq":868300000,"upinfo":{"rxtime":1548059982,"rtcx":0,"xtime":12666373963464220,"gpstime":0,"rssi":89,"snr":9.25}}}`),
		},
		{
			Name: "TxConfirmation",
			Message: TxConfirmation{
				Diid:    35,
				DevEUI:  basicstation.EUI{EUI64: types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}},
				XTime:   1552906698,
				TxTime:  1552906698,
				GpsTime: 1552906698,
			},
			Expected: []byte(`{"msgtype":"dntxed","diid":35,"DevEui":"1111:1111:1111:1111","rctx":0,"xtime":1552906698,"txtime":1552906698,"gpstime":1552906698}`),
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			msg, err := tc.Message.MarshalJSON()
			if !a.So(err, should.Resemble, nil) {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !a.So(msg, should.Resemble, tc.Expected) {
				t.Fatalf("Unexpected message: %v", msg)
			}
		})
	}
}

func TestJoinRequest(t *testing.T) {
	for _, tc := range []struct {
		Name                  string
		JoinRequest           JoinRequest
		GatewayIDs            ttnpb.GatewayIdentifiers
		BandID                string
		ExpectedUplinkMessage ttnpb.UplinkMessage
		ErrorAssertion        func(err error) bool
	}{
		{
			Name: "InvalidBandID",
			JoinRequest: JoinRequest{
				MHdr:     0,
				DevEUI:   basicstation.EUI{EUI64: types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}},
				JoinEUI:  basicstation.EUI{EUI64: types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}},
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
			GatewayIDs:     ttnpb.GatewayIdentifiers{GatewayID: "eui-1122334455667788"},
			BandID:         "EU_86_870",
			ErrorAssertion: errors.IsNotFound,
		},
		{
			Name: "InvalidMhdr",
			JoinRequest: JoinRequest{
				MHdr:     25,
				DevEUI:   basicstation.EUI{EUI64: types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}},
				JoinEUI:  basicstation.EUI{EUI64: types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}},
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
			GatewayIDs:     ttnpb.GatewayIdentifiers{GatewayID: "eui-1122334455667788"},
			BandID:         "EU_868_870",
			ErrorAssertion: errors.IsNotFound,
		},
		{
			Name:        "EmptyJoinRequest",
			JoinRequest: JoinRequest{},
			GatewayIDs:  ttnpb.GatewayIdentifiers{GatewayID: "eui-1122334455667788"},
			BandID:      "EU_863_870",
			ExpectedUplinkMessage: ttnpb.UplinkMessage{
				Payload: &ttnpb.Message{
					MIC:  []byte{0, 0, 0, 0},
					MHDR: ttnpb.MHDR{MType: ttnpb.MType_JOIN_REQUEST, Major: ttnpb.Major_LORAWAN_R1},
					Payload: &ttnpb.Message_JoinRequestPayload{JoinRequestPayload: &ttnpb.JoinRequestPayload{
						JoinEUI:  types.EUI64{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
						DevEUI:   types.EUI64{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
						DevNonce: [2]byte{0x00, 0x00},
					}}},
				RxMetadata: []*ttnpb.RxMetadata{{
					GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "eui-1122334455667788",
						EUI: &types.EUI64{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88},
					},
					UplinkToken: []byte{10, 24, 10, 22, 10, 20, 101, 117, 105, 45, 49, 49, 50, 50, 51, 51, 52, 52, 53, 53, 54, 54, 55, 55, 56, 56},
				}},
				Settings: ttnpb.TxSettings{
					CodingRate: "4/5",
					DataRate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
						SpreadingFactor: 12,
						Bandwidth:       125000,
					}}}},
			},
		},
		{
			Name: "ValidJoinRequest",
			JoinRequest: JoinRequest{
				MHdr:     0,
				DevEUI:   basicstation.EUI{EUI64: types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}},
				JoinEUI:  basicstation.EUI{EUI64: types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}},
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
			GatewayIDs: ttnpb.GatewayIdentifiers{GatewayID: "eui-1122334455667788"},
			BandID:     "EU_863_870",
			ExpectedUplinkMessage: ttnpb.UplinkMessage{
				Payload: &ttnpb.Message{
					MHDR: ttnpb.MHDR{MType: ttnpb.MType_JOIN_REQUEST, Major: ttnpb.Major_LORAWAN_R1},
					MIC:  []byte{0x4E, 0x61, 0xBC, 0x00},
					Payload: &ttnpb.Message_JoinRequestPayload{JoinRequestPayload: &ttnpb.JoinRequestPayload{
						JoinEUI:  types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22},
						DevEUI:   types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11},
						DevNonce: [2]byte{0x50, 0x46},
					}}},
				RxMetadata: []*ttnpb.RxMetadata{{
					GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "eui-1122334455667788",
						EUI: &types.EUI64{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88},
					},
					Time:        &[]time.Time{time.Unix(1548059982, 0)}[0],
					Timestamp:   (uint32)(12666373963464220 & 0xFFFFFFFF),
					RSSI:        89,
					SNR:         9.25,
					UplinkToken: []byte{10, 24, 10, 22, 10, 20, 101, 117, 105, 45, 49, 49, 50, 50, 51, 51, 52, 52, 53, 53, 54, 54, 55, 55, 56, 56, 16, 156, 252, 188, 5},
				},
				},
				Settings: ttnpb.TxSettings{
					Frequency:  868300000,
					Time:       &[]time.Time{time.Unix(1548059982, 0)}[0],
					Timestamp:  (uint32)(12666373963464220 & 0xFFFFFFFF),
					CodingRate: "4/5",
					DataRate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
						SpreadingFactor: 11,
						Bandwidth:       125000,
					}}},
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			msg, err := tc.JoinRequest.ToUplinkMessage(tc.GatewayIDs, tc.BandID, time.Time{})
			if err != nil {
				if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else if tc.ErrorAssertion != nil {
				t.Fatalf("Expected error")
			} else {
				var payload ttnpb.Message
				a.So(lorawan.UnmarshalMessage(msg.RawPayload, &payload), should.BeNil)
				if !a.So(&payload, should.Resemble, msg.Payload) {
					t.Fatalf("Invalid RawPayload: %v", msg.RawPayload)
				}
				msg.RawPayload = nil
				if !a.So(*msg, should.Resemble, tc.ExpectedUplinkMessage) {
					t.Fatalf("Invalid UplinkMessage: %s", msg.RawPayload)
				}
			}
		})
	}
}

func TestUplinkDataFrame(t *testing.T) {
	for _, tc := range []struct {
		Name                  string
		UplinkDataFrame       UplinkDataFrame
		GatewayIDs            ttnpb.GatewayIdentifiers
		FrequencyPlanID       string
		ExpectedUplinkMessage ttnpb.UplinkMessage
		ErrorAssertion        func(err error) bool
	}{
		{
			Name:                  "Empty",
			UplinkDataFrame:       UplinkDataFrame{},
			GatewayIDs:            ttnpb.GatewayIdentifiers{GatewayID: "eui-1122334455667788"},
			FrequencyPlanID:       "EU_863_870",
			ExpectedUplinkMessage: ttnpb.UplinkMessage{},
			ErrorAssertion: func(err error) bool {
				return errors.Resemble(err, errUplinkDataFrame)
			},
		},
		{
			Name: "ValidFrame",
			UplinkDataFrame: UplinkDataFrame{
				MHdr:       0x40,
				DevAddr:    0x11223344,
				FCtrl:      0x30,
				FPort:      0x00,
				FCnt:       25,
				FOpts:      "FD",
				FRMPayload: "5fcc",
				MIC:        12345678,
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
			GatewayIDs:      ttnpb.GatewayIdentifiers{GatewayID: "eui-1122334455667788"},
			FrequencyPlanID: "EU_863_870",
			ExpectedUplinkMessage: ttnpb.UplinkMessage{
				Payload: &ttnpb.Message{
					MHDR: ttnpb.MHDR{MType: ttnpb.MType_UNCONFIRMED_UP, Major: ttnpb.Major_LORAWAN_R1},
					MIC:  []byte{0x4E, 0x61, 0xBC, 0x00},
					Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{
						FPort:      0,
						FRMPayload: []byte{0x5F, 0xCC},
						FHDR: ttnpb.FHDR{
							DevAddr: [4]byte{0x11, 0x22, 0x33, 0x44},
							FCtrl: ttnpb.FCtrl{
								Ack:    true,
								ClassB: true,
							},
							FCnt:  25,
							FOpts: []byte{0xFD},
						},
					}}},
				RxMetadata: []*ttnpb.RxMetadata{{
					GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "eui-1122334455667788",
						EUI: &types.EUI64{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88},
					},
					Time:        &[]time.Time{time.Unix(1548059982, 0)}[0],
					Timestamp:   (uint32)(12666373963464220 & 0xFFFFFFFF),
					RSSI:        89,
					SNR:         9.25,
					UplinkToken: []byte{10, 34, 10, 32, 10, 20, 101, 117, 105, 45, 49, 49, 50, 50, 51, 51, 52, 52, 53, 53, 54, 54, 55, 55, 56, 56, 18, 8, 17, 34, 51, 68, 85, 102, 119, 136, 16, 156, 252, 188, 5},
				},
				},
				Settings: ttnpb.TxSettings{
					Timestamp:  (uint32)(12666373963464220 & 0xFFFFFFFF),
					Time:       &[]time.Time{time.Unix(1548059982, 0)}[0],
					CodingRate: "4/5",
					Frequency:  868300000,
					DataRate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
						SpreadingFactor: 11,
						Bandwidth:       125000,
					}}},
				},
			},
		},
		{
			Name: "NegativeFPort",
			UplinkDataFrame: UplinkDataFrame{
				MHdr:       0x40,
				DevAddr:    0x11223344,
				FCtrl:      0x30,
				FPort:      -1,
				FCnt:       25,
				FOpts:      "FD",
				FRMPayload: "5fcc",
				MIC:        12345678,
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
			GatewayIDs:      ttnpb.GatewayIdentifiers{GatewayID: "eui-1122334455667788"},
			FrequencyPlanID: "EU_863_870",
			ExpectedUplinkMessage: ttnpb.UplinkMessage{
				Payload: &ttnpb.Message{
					MHDR: ttnpb.MHDR{MType: ttnpb.MType_UNCONFIRMED_UP, Major: ttnpb.Major_LORAWAN_R1},
					MIC:  []byte{0x4E, 0x61, 0xBC, 0x00},
					Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{
						FPort:      0,
						FRMPayload: []byte{0x5F, 0xCC},
						FHDR: ttnpb.FHDR{
							DevAddr: [4]byte{0x11, 0x22, 0x33, 0x44},
							FCtrl: ttnpb.FCtrl{
								Ack:    true,
								ClassB: true,
							},
							FCnt:  25,
							FOpts: []byte{0xFD},
						},
					}}},
				RxMetadata: []*ttnpb.RxMetadata{{
					GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "eui-1122334455667788",
						EUI: &types.EUI64{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88},
					},
					Time:        &[]time.Time{time.Unix(1548059982, 0)}[0],
					Timestamp:   (uint32)(12666373963464220 & 0xFFFFFFFF),
					RSSI:        89,
					SNR:         9.25,
					UplinkToken: []byte{10, 34, 10, 32, 10, 20, 101, 117, 105, 45, 49, 49, 50, 50, 51, 51, 52, 52, 53, 53, 54, 54, 55, 55, 56, 56, 18, 8, 17, 34, 51, 68, 85, 102, 119, 136, 16, 156, 252, 188, 5},
				},
				},
				Settings: ttnpb.TxSettings{
					Frequency:  868300000,
					Timestamp:  (uint32)(12666373963464220 & 0xFFFFFFFF),
					Time:       &[]time.Time{time.Unix(1548059982, 0)}[0],
					CodingRate: "4/5",
					DataRate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
						SpreadingFactor: 11,
						Bandwidth:       125000,
					}}},
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			msg, err := tc.UplinkDataFrame.ToUplinkMessage(tc.GatewayIDs, tc.FrequencyPlanID, time.Time{})
			if err != nil {
				if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else if tc.ErrorAssertion != nil {
				t.Fatalf("Expected error")
			} else {
				msg.RawPayload = nil
				if !a.So(*msg, should.Resemble, tc.ExpectedUplinkMessage) {
					t.Fatalf("Invalid UplinkMessage: %s", msg.RawPayload)
				}
			}
		})
	}
}

func TestFromUplinkDataFrame(t *testing.T) {
	for _, tc := range []struct {
		Name                    string
		UplinkMessage           ttnpb.UplinkMessage
		GatewayIDs              ttnpb.GatewayIdentifiers
		FrequencyPlanID         string
		ExpectedUplinkDataFrame UplinkDataFrame
		ErrorAssertion          func(err error) bool
	}{
		{
			Name:                    "Empty",
			ExpectedUplinkDataFrame: UplinkDataFrame{},
			FrequencyPlanID:         "EU_863_870",
			UplinkMessage:           ttnpb.UplinkMessage{},
			ErrorAssertion: func(err error) bool {
				return errors.Resemble(err, errUplinkMessage)
			},
		},
		{
			Name: "ValidFrame",
			UplinkMessage: ttnpb.UplinkMessage{
				RawPayload: []byte{0x40, 0xff, 0xff, 0xff, 0x42, 0xb2, 0x42, 0xff, 0xfe, 0xff, 0x42, 0xfe, 0xff, 0x42, 0xff, 0xff, 0x0f},
				Payload: &ttnpb.Message{
					MHDR: ttnpb.MHDR{MType: ttnpb.MType_UNCONFIRMED_UP, Major: 0},
					Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{
						FHDR: ttnpb.FHDR{
							DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
							FCtrl: ttnpb.FCtrl{
								ADR:       true,
								ADRAckReq: false,
								Ack:       true,
								ClassB:    true,
								FPending:  false,
							},
							FCnt:  0xff42,
							FOpts: []byte{0xfe, 0xff},
						},
						FPort:      0x42,
						FRMPayload: []byte{0xfe, 0xff},
					}},
					MIC: []byte{0x42, 0xff, 0xff, 0x0f}},
				RxMetadata: []*ttnpb.RxMetadata{{
					GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "eui-1122334455667788"},
					Time:               &[]time.Time{time.Unix(1548059982, 0)}[0],
					Timestamp:          (uint32)(12666373963464220 & 0xFFFFFFFF),
					RSSI:               89,
					SNR:                9.25,
					UplinkToken:        []byte{10, 16, 10, 14, 10, 12, 116, 101, 115, 116, 45, 103, 97, 116, 101, 119, 97, 121, 16, 156, 252, 188, 5},
				},
				},
				Settings: ttnpb.TxSettings{
					Frequency: 868300000,
					DataRate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
						SpreadingFactor: 11,
						Bandwidth:       125000,
					}}},
				},
			},
			FrequencyPlanID: "EU_863_870",
			ExpectedUplinkDataFrame: UplinkDataFrame{
				MHdr:       0x40,
				DevAddr:    0x42ffffff,
				FCtrl:      0xb0,
				FPort:      0x42,
				FCnt:       0xff42,
				FOpts:      "feff",
				FRMPayload: "feff",
				MIC:        268435266,
				RadioMetaData: RadioMetaData{
					DataRate:  1,
					Frequency: 868300000,
					UpInfo: UpInfo{
						RxTime: 1548059982,
						XTime:  (12666373963464220 & 0xFFFFFFFF),
						RSSI:   89,
						SNR:    9.25,
					},
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			var updf UplinkDataFrame
			err := updf.FromUplinkMessage(&tc.UplinkMessage, tc.FrequencyPlanID)
			if err != nil {
				if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else if tc.ErrorAssertion != nil {
				t.Fatalf("Expected error")
			} else {
				if !a.So(updf, should.Resemble, tc.ExpectedUplinkDataFrame) {
					t.Fatalf("Invalid UplinkMessage: %v", updf)
				}
			}
		})
	}
}

func TestJreqFromUplinkDataFrame(t *testing.T) {
	for _, tc := range []struct {
		Name                string
		UplinkMessage       ttnpb.UplinkMessage
		FrequencyPlanID     string
		ExpectedJoinRequest JoinRequest
		ErrorAssertion      func(err error) bool
	}{
		{
			Name:                "Empty",
			ExpectedJoinRequest: JoinRequest{},
			FrequencyPlanID:     "EU_863_870",
			UplinkMessage:       ttnpb.UplinkMessage{},
			ErrorAssertion: func(err error) bool {
				return errors.Resemble(err, errUplinkMessage)
			},
		},
		{
			Name: "ValidFrame",
			UplinkMessage: ttnpb.UplinkMessage{
				RawPayload: []byte{0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x42, 0xff, 0x42, 0x42, 0xff, 0xff, 0x0f},
				Payload: &ttnpb.Message{
					MHDR: ttnpb.MHDR{MType: ttnpb.MType_JOIN_REQUEST, Major: 0},
					Payload: &ttnpb.Message_JoinRequestPayload{JoinRequestPayload: &ttnpb.JoinRequestPayload{
						JoinEUI:  types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
						DevEUI:   types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
						DevNonce: types.DevNonce{0x42, 0xff},
					}},
					MIC: []byte{0x42, 0xff, 0xff, 0x0f}},
				RxMetadata: []*ttnpb.RxMetadata{{
					GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "eui-1122334455667788"},
					Time:               &[]time.Time{time.Unix(1548059982, 0)}[0],
					Timestamp:          (uint32)(12666373963464220 & 0xFFFFFFFF),
					RSSI:               89,
					SNR:                9.25,
					UplinkToken:        []byte{10, 16, 10, 14, 10, 12, 116, 101, 115, 116, 45, 103, 97, 116, 101, 119, 97, 121, 16, 156, 252, 188, 5},
				},
				},
				Settings: ttnpb.TxSettings{
					Frequency: 868300000,
					DataRate: ttnpb.DataRate{Modulation: &ttnpb.DataRate_LoRa{LoRa: &ttnpb.LoRaDataRate{
						SpreadingFactor: 11,
						Bandwidth:       125000,
					}}},
				},
			},
			FrequencyPlanID: "EU_863_870",
			ExpectedJoinRequest: JoinRequest{
				MHdr:     0,
				DevEUI:   basicstation.EUI{EUI64: types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
				JoinEUI:  basicstation.EUI{EUI64: types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
				DevNonce: 65346,
				MIC:      268435266,
				RadioMetaData: RadioMetaData{
					DataRate:  1,
					Frequency: 868300000,
					UpInfo: UpInfo{
						RxTime: 1548059982,
						XTime:  (12666373963464220 & 0xFFFFFFFF),
						RSSI:   89,
						SNR:    9.25,
					},
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			var jreq JoinRequest
			err := jreq.FromUplinkMessage(&tc.UplinkMessage, tc.FrequencyPlanID)
			if err != nil {
				if tc.ErrorAssertion == nil || !a.So(tc.ErrorAssertion(err), should.BeTrue) {
					t.Fatalf("Unexpected error: %v", err)
				}
			} else if tc.ErrorAssertion != nil {
				t.Fatalf("Expected error")
			} else {
				if !a.So(jreq, should.Resemble, tc.ExpectedJoinRequest) {
					t.Fatalf("Invalid UplinkMessage: %v", jreq)
				}
			}
		})
	}
}

func TestTxAck(t *testing.T) {
	a := assertions.New(t)
	correlationIDs := []string{"i3N84kvunPAS8wOmiEKbhsP62wNMRdmn", "deK3h59wUZhR0xb17eumTkauGQxoB5xn"}
	res := ToTxAcknowledgment(correlationIDs)

	if !a.So(res, should.Resemble, ttnpb.TxAcknowledgment{
		CorrelationIDs: correlationIDs,
		Result:         ttnpb.TxAcknowledgment_SUCCESS,
	}) {
		t.Fatalf("Unexpected TxAck: %v", res)
	}

}
