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
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

type message interface {
	MarshalJSON() ([]byte, error)
}

func TestMarshalJSON(t *testing.T) {

	a := assertions.New(t)
	for _, tc := range []struct {
		Name             string
		Message          message
		MarshaledMessage []byte
	}{
		{
			"JoinRequest",
			JoinRequest{
				MHdr:     0,
				DevEUI:   basicstation.EUI{Prefix: "DevEui", EUI64: types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}},
				JoinEUI:  basicstation.EUI{Prefix: "JoinEui", EUI64: types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}},
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
			[]byte("{\"msgtype\":\"jreq\",\"MHdr\":0,\"JoinEui\":\"joineui-2222:2222:2222:2222\",\"DevEui\":\"deveui-1111:1111:1111:1111\",\"DevNonce\":18000,\"MIC\":12345678,\"RefTime\":0,\"RadioMetaData\":{\"DR\":1,\"Freq\":868300000,\"upinfo\":{\"rxtime\":1548059982,\"rtcx\":0,\"xtime\":12666373963464220,\"gpstime\":0,\"rssi\":89,\"snr\":9.25}}}"),
		},
		{
			"UplinkDataFrame",
			UplinkDataFrame{
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
			[]byte("{\"msgtype\":\"updf\",\"MHdr\":64,\"DevAddr\":287454020,\"FCtrl\":48,\"Fcnt\":25,\"FOpts\":\"FD\",\"FPort\":0,\"FRMPayload\":\"Ymxhamthc25kJ3M=\",\"MIC\":12345678,\"RefTime\":0,\"RadioMetaData\":{\"DR\":1,\"Freq\":868300000,\"upinfo\":{\"rxtime\":1548059982,\"rtcx\":0,\"xtime\":12666373963464220,\"gpstime\":0,\"rssi\":89,\"snr\":9.25}}}"),
		},
		{
			"TxConfirmation",
			TxConfirmation{
				Diid:    35,
				DevEUI:  basicstation.EUI{Prefix: "DevEui", EUI64: types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}},
				XTime:   1552906698,
				TxTime:  1552906698,
				GpsTime: 1552906698,
			},
			[]byte("{\"msgtype\":\"dntxed\",\"diid\":35,\"DevEui\":\"deveui-1111:1111:1111:1111\",\"rctx\":0,\"xtime\":1552906698,\"txtime\":1552906698,\"gpstime\":1552906698}"),
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			msg, err := tc.Message.MarshalJSON()
			if !(a.So(err, should.Resemble, nil)) {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !(a.So(msg, should.Resemble, tc.MarshaledMessage)) {
				t.Fatalf("Unexpected message: %v", msg)
			}
		})
	}
}

func TestJoinRequest(t *testing.T) {
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
				RxMetadata: []*ttnpb.RxMetadata{{
					GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "test-gateway"},
					Time:               &[]time.Time{time.Unix(0, 0)}[0],
					UplinkToken:        []byte{10, 16, 10, 14, 10, 12, 116, 101, 115, 116, 45, 103, 97, 116, 101, 119, 97, 121},
				}},
				Settings: ttnpb.TxSettings{
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
				DevEUI:   basicstation.EUI{Prefix: "DevEui", EUI64: types.EUI64{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}},
				JoinEUI:  basicstation.EUI{Prefix: "JoinEui", EUI64: types.EUI64{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}},
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
				RxMetadata: []*ttnpb.RxMetadata{{
					GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "test-gateway"},
					Time:               &[]time.Time{time.Unix(1548059982, 0)}[0],
					Timestamp:          (uint32)(12666373963464220 & 0xFFFFFFFF),
					RSSI:               89,
					SNR:                9.25,
					UplinkToken:        []byte{10, 16, 10, 14, 10, 12, 116, 101, 115, 116, 45, 103, 97, 116, 101, 119, 97, 121, 16, 156, 252, 188, 5},
				},
				},
				Settings: ttnpb.TxSettings{
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

func TestUplinkDataFrame(t *testing.T) {
	a := assertions.New(t)
	for _, tc := range []struct {
		Name                  string
		UplinkDataFrame       UplinkDataFrame
		GatewayIDs            ttnpb.GatewayIdentifiers
		FreqPlanID            string
		ExpectedUplinkMessage ttnpb.UplinkMessage
		ExpectedError         error
	}{
		{
			"Empty",
			UplinkDataFrame{},
			ttnpb.GatewayIdentifiers{GatewayID: "test-gateway"},
			"EU_863_870",
			ttnpb.UplinkMessage{},
			errUplinkDataFrame,
		},
		{
			"ValidFrame",
			UplinkDataFrame{
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
			ttnpb.GatewayIdentifiers{GatewayID: "test-gateway"},
			"EU_863_870",
			ttnpb.UplinkMessage{
				Payload: &ttnpb.Message{
					MHDR: ttnpb.MHDR{MType: ttnpb.MType_UNCONFIRMED_UP, Major: ttnpb.Major_LORAWAN_R1},
					MIC:  []byte{0x4E, 0x61, 0xBC, 0x00},
					Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{
						FPort:      0,
						FRMPayload: []byte("Ymxhamthc25kJ3M="),
						FHDR: ttnpb.FHDR{
							DevAddr: [4]byte{0x44, 0x33, 0x22, 0x11},
							FCtrl: ttnpb.FCtrl{
								Ack:    true,
								ClassB: true,
							},
							FCnt:  25,
							FOpts: []byte("FD"),
						},
					}}},
				RxMetadata: []*ttnpb.RxMetadata{{
					GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "test-gateway"},
					Time:               &[]time.Time{time.Unix(1548059982, 0)}[0],
					Timestamp:          (uint32)(12666373963464220 & 0xFFFFFFFF),
					RSSI:               89,
					SNR:                9.25,
					UplinkToken:        []byte{10, 16, 10, 14, 10, 12, 116, 101, 115, 116, 45, 103, 97, 116, 101, 119, 97, 121, 16, 156, 252, 188, 5},
				},
				},
				Settings: ttnpb.TxSettings{
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
		{
			"NegativeFPort",
			UplinkDataFrame{
				MHdr:       0x40,
				DevAddr:    0x11223344,
				FCtrl:      0x30,
				FPort:      -1,
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
			ttnpb.GatewayIdentifiers{GatewayID: "test-gateway"},
			"EU_863_870",
			ttnpb.UplinkMessage{
				Payload: &ttnpb.Message{
					MHDR: ttnpb.MHDR{MType: ttnpb.MType_UNCONFIRMED_UP, Major: ttnpb.Major_LORAWAN_R1},
					MIC:  []byte{0x4E, 0x61, 0xBC, 0x00},
					Payload: &ttnpb.Message_MACPayload{MACPayload: &ttnpb.MACPayload{
						FPort:      0,
						FRMPayload: []byte("Ymxhamthc25kJ3M="),
						FHDR: ttnpb.FHDR{
							DevAddr: [4]byte{0x44, 0x33, 0x22, 0x11},
							FCtrl: ttnpb.FCtrl{
								Ack:    true,
								ClassB: true,
							},
							FCnt:  25,
							FOpts: []byte("FD"),
						},
					}}},
				RxMetadata: []*ttnpb.RxMetadata{{
					GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "test-gateway"},
					Time:               &[]time.Time{time.Unix(1548059982, 0)}[0],
					Timestamp:          (uint32)(12666373963464220 & 0xFFFFFFFF),
					RSSI:               89,
					SNR:                9.25,
					UplinkToken:        []byte{10, 16, 10, 14, 10, 12, 116, 101, 115, 116, 45, 103, 97, 116, 101, 119, 97, 121, 16, 156, 252, 188, 5},
				},
				},
				Settings: ttnpb.TxSettings{
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
			msg, err := tc.UplinkDataFrame.ToUplinkMessage(tc.GatewayIDs, tc.FreqPlanID)
			if !(a.So(err, should.Resemble, tc.ExpectedError)) {
				t.Fatalf("Unexpected error: %v", err)
			}
			msg.ReceivedAt = time.Time{}
			msg.RawPayload = nil
			if !(a.So(msg, should.Resemble, tc.ExpectedUplinkMessage)) {
				t.Fatalf("Invalid UplinkMessage: %s", msg.RawPayload)
			}
		})
	}
}

func TestTxAck(t *testing.T) {
	a := assertions.New(t)
	correlationIDs := []string{"i3N84kvunPAS8wOmiEKbhsP62wNMRdmn", "deK3h59wUZhR0xb17eumTkauGQxoB5xn"}
	res := ToTxAcknowledgment(correlationIDs)

	if !(a.So(res, should.Resemble, ttnpb.TxAcknowledgment{
		CorrelationIDs: correlationIDs,
		Result:         ttnpb.TxAcknowledgment_SUCCESS,
	})) {
		t.Fatalf("Unexpected TxAck: %v", res)
	}

}
