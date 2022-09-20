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

package udp_test

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/kr/pretty"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/gpstime"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb/udp"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/datarate"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/version"
)

var ids = &ttnpb.GatewayIdentifiers{GatewayId: "test-gateway"}

func uint32Ptr(v uint32) *uint32 { return &v }
func int32Ptr(v int32) *int32    { return &v }

func TestStatusRaw(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	raw := []byte(`{"stat":{"rxfw":0,"hal":"5.1.0","fpga":2,"dsp":31,"lpps":2,"lmnw":3,"lmst":1,"lmok":3,"temp":30.5,"lati":52.34223,"long":5.29685,"txnb":0,"dwnb":0,"alti":66,"rxok":0,"boot":"2017-06-07 09:40:42 GMT","time":"2017-06-08 09:40:42 GMT","rxnb":0,"ackr":0.0}}`) //nolint:lll
	var statusData udp.Data
	err := json.Unmarshal(raw, &statusData)
	a.So(err, should.BeNil)

	upstream, err := udp.ToGatewayUp(statusData, udp.UpstreamMetadata{
		IP: "127.0.0.1",
		ID: ids,
	})
	a.So(err, should.BeNil)

	status := upstream.GatewayStatus
	a.So(status, should.NotBeNil)

	a.So(status.AntennaLocations, should.NotBeNil)
	a.So(len(status.AntennaLocations), should.Equal, 1)
	a.So(status.AntennaLocations[0].Longitude, should.AlmostEqual, 5.29685, 0.0001)
	a.So(status.AntennaLocations[0].Latitude, should.AlmostEqual, 52.34223, 0.0001)
	a.So(status.AntennaLocations[0].Altitude, should.AlmostEqual, 66)

	a.So(status.Versions, should.NotBeNil)
	a.So(status.Metrics, should.NotBeNil)

	a.So(status.Versions["ttn-lw-gateway-server"], should.Equal, version.TTN)
	a.So(status.Versions["hal"], should.Equal, "5.1.0")
	a.So(status.Versions["fpga"], should.Equal, "2")
	a.So(status.Versions["dsp"], should.Equal, "31")

	a.So(status.Metrics["rxfw"], should.AlmostEqual, 0)
	a.So(status.Metrics["txnb"], should.AlmostEqual, 0)
	a.So(status.Metrics["dwnb"], should.AlmostEqual, 0)
	a.So(status.Metrics["rxok"], should.AlmostEqual, 0)
	a.So(status.Metrics["rxnb"], should.AlmostEqual, 0)
	a.So(status.Metrics["ackr"], should.AlmostEqual, 0)
	a.So(status.Metrics["temp"], should.AlmostEqual, 30.5)
	a.So(status.Metrics["lpps"], should.AlmostEqual, 2)
	a.So(status.Metrics["lmnw"], should.AlmostEqual, 3)
	a.So(status.Metrics["lmst"], should.AlmostEqual, 1)
	a.So(status.Metrics["lmok"], should.AlmostEqual, 3)

	a.So(status.BootTime, should.NotBeNil)
	a.So(status.Time, should.NotBeNil)
	currentTime := time.Date(2017, 6, 8, 9, 40, 42, 0, time.UTC)
	if a.So(status.Time, should.NotBeNil) {
		a.So(*ttnpb.StdTime(status.Time), should.Equal, currentTime)
	}
	bootTime := time.Date(2017, 6, 7, 9, 40, 42, 0, time.UTC)
	if a.So(status.BootTime, should.NotBeNil) {
		a.So(*ttnpb.StdTime(status.BootTime), should.Equal, bootTime)
	}
}

func TestToGatewayUp(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	absoluteTime := time.Now().UTC().Truncate(time.Millisecond)
	gpsTime := uint64(gpstime.ToGPS(absoluteTime) / time.Millisecond)

	p := udp.Packet{
		GatewayEUI:      &types.EUI64{0xAA, 0xEE, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		ProtocolVersion: udp.Version1,
		Token:           [2]byte{0x11, 0x00},
		Data: &udp.Data{
			RxPacket: []*udp.RxPacket{
				{
					Freq: 868.0,
					Chan: 2,
					Modu: "LORA",
					DatR: datarate.DR{
						DataRate: &ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 10,
									Bandwidth:       125000,
									CodingRate:      band.Cr4_7,
								},
							},
						},
					},
					CodR:  band.Cr4_7,
					Data:  "QCkuASaAAAAByFaF53Iu+vzmwQ==",
					Size:  19,
					Tmst:  1000,
					Tmms:  &gpsTime,
					FTime: uint32Ptr(12345678),
					FOff:  int32Ptr(-42),
				},
			},
		},
		PacketType: udp.PushData,
	}

	upstream, err := udp.ToGatewayUp(*p.Data, udp.UpstreamMetadata{ID: ids})
	a.So(err, should.BeNil)

	msg := upstream.UplinkMessages[0]
	dr := msg.Settings.DataRate.GetLora()
	a.So(dr, should.NotBeNil)
	a.So(dr.SpreadingFactor, should.Equal, 10)
	a.So(dr.Bandwidth, should.Equal, 125000)
	a.So(msg.Settings.DataRate.GetLora().CodingRate, should.Equal, band.Cr4_7)
	a.So(msg.Settings.Frequency, should.Equal, 868000000)
	a.So(msg.Settings.Timestamp, should.Equal, 1000)
	a.So(*ttnpb.StdTime(msg.Settings.Time), should.Resemble, absoluteTime)
	a.So(msg.RxMetadata[0].Timestamp, should.Equal, 1000)
	a.So(ttnpb.StdTime(msg.RxMetadata[0].Time), should.Resemble, &absoluteTime)
	a.So(msg.RxMetadata[0].FineTimestamp, should.Equal, 12345678)
	a.So(msg.RxMetadata[0].FrequencyOffset, should.Equal, -42)
	a.So(msg.RawPayload, should.Resemble, []byte{0x40, 0x29, 0x2e, 0x01, 0x26, 0x80, 0x00, 0x00, 0x01, 0xc8, 0x56, 0x85, 0xe7, 0x72, 0x2e, 0xfa, 0xfc, 0xe6, 0xc1})
}

func TestToGatewayUpLRFHSS(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	p := udp.Packet{
		GatewayEUI:      &types.EUI64{0xAA, 0xEE, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		ProtocolVersion: udp.Version1,
		Token:           [2]byte{0x11, 0x00},
		Data: &udp.Data{
			RxPacket: []*udp.RxPacket{
				{
					Freq: 868.0,
					Chan: 2,
					Modu: "LR-FHSS",
					DatR: datarate.DR{
						DataRate: &ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lrfhss{
								Lrfhss: &ttnpb.LRFHSSDataRate{
									ModulationType:        0,
									OperatingChannelWidth: 125,
									CodingRate:            band.Cr4_6,
								},
							},
						},
					},
					CodR: band.Cr4_6,
					Data: "QCkuASaAAAAByFaF53Iu+vzmwQ==",
					Size: 19,
					Tmst: 1000,
					Hpw:  8,
					RSig: []udp.RSig{
						{
							FOff: 125000,
							Fdri: 25000,
						},
					},
				},
			},
		},
		PacketType: udp.PushData,
	}

	upstream, err := udp.ToGatewayUp(*p.Data, udp.UpstreamMetadata{ID: ids})
	a.So(err, should.BeNil)

	msg := upstream.UplinkMessages[0]
	dr := msg.Settings.DataRate.GetLrfhss()
	a.So(dr, should.NotBeNil)
	a.So(dr.ModulationType, should.Equal, 0)
	a.So(dr.OperatingChannelWidth, should.Equal, 125)
	a.So(dr.CodingRate, should.Equal, band.Cr4_6)
	a.So(msg.Settings.Frequency, should.Equal, 868000000)
	a.So(msg.Settings.Timestamp, should.Equal, 1000)
	a.So(msg.RxMetadata[0].Timestamp, should.Equal, 1000)
	a.So(msg.RxMetadata[0].HoppingWidth, should.Equal, 8)
	a.So(msg.RxMetadata[0].FrequencyDrift, should.Equal, 25000)
	a.So(msg.RawPayload, should.Resemble, []byte{0x40, 0x29, 0x2e, 0x01, 0x26, 0x80, 0x00, 0x00, 0x01, 0xc8, 0x56, 0x85, 0xe7, 0x72, 0x2e, 0xfa, 0xfc, 0xe6, 0xc1}) //nolint:lll
}

func TestToGatewayUpRoundtrip(t *testing.T) {
	t.Parallel()
	expectedMd := udp.UpstreamMetadata{
		ID: &ttnpb.GatewayIdentifiers{
			Eui: types.EUI64{0xAA, 0xEE, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}.Bytes(),
		},
		IP: "1.1.1.1",
	}

	for _, tc := range []struct {
		Name       string
		Data       *udp.Data
		PacketType udp.PacketType
	}{
		{
			Name: "Uplink",
			Data: &udp.Data{
				RxPacket: []*udp.RxPacket{
					{
						Freq: 868.0,
						Chan: 2,
						Modu: "LORA",
						DatR: datarate.DR{
							DataRate: &ttnpb.DataRate{
								Modulation: &ttnpb.DataRate_Lora{
									Lora: &ttnpb.LoRaDataRate{
										SpreadingFactor: 10,
										Bandwidth:       125000,
										CodingRate:      band.Cr4_7,
									},
								},
							},
						},
						CodR:  band.Cr4_7,
						Data:  "QCkuASaAAAAByFaF53Iu+vzmwQ==",
						Size:  19,
						Tmst:  1000,
						FTime: uint32Ptr(12345678),
						FOff:  int32Ptr(-42),
					},
				},
			},
			PacketType: udp.PushData,
		},
		{
			Name: "TxAcknowledgment",
			Data: &udp.Data{
				TxPacketAck: &udp.TxPacketAck{
					Error: udp.TxErrNone,
				},
			},
			PacketType: udp.TxAck,
		},
	} {
		a := assertions.New(t)

		expected := udp.Packet{
			ProtocolVersion: udp.Version1,
			Token:           [2]byte{0x11, 0x00},
			Data:            tc.Data,
			PacketType:      tc.PacketType,
		}
		up, err := udp.ToGatewayUp(*expected.Data, expectedMd)
		a.So(err, should.BeNil)

		actual := udp.Packet{
			ProtocolVersion: udp.Version1,
			Token:           [2]byte{0x11, 0x00},
			Data:            &udp.Data{},
			PacketType:      tc.PacketType,
		}
		actual.Data.RxPacket, actual.Data.Stat, actual.Data.TxPacketAck = udp.FromGatewayUp(up)
		a.So(pretty.Diff(actual, expected), should.BeEmpty)
	}
}

func TestToGatewayUpRaw(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	raw := []byte(`{"rxpk":[{"tmst":368384825,"chan":0,"rfch":0,"freq":868.100000,"stat":1,"modu":"LORA","datr":"SF7BW125","codr":"4/5","lsnr":-11,"rssi":-107,"size":108,"data":"Wqish6GVYpKy6o9WFHingeTJ1oh+ABc8iALBvwz44yxZP+BKDocaC5VQT5Y6dDdUaBILVjRMz0Ynzow1U/Kkts9AoZh3Ja3DX+DyY27exB+BKpSx2rXJ2vs9svm/EKYIsPF0RG1E+7lBYaD9"}]}`) //nolint:lll
	var rxData udp.Data
	err := json.Unmarshal(raw, &rxData)
	a.So(err, should.BeNil)

	upstream, err := udp.ToGatewayUp(rxData, udp.UpstreamMetadata{ID: ids})
	a.So(err, should.BeNil)

	a.So(len(upstream.UplinkMessages), should.Equal, 1)
	msg := upstream.UplinkMessages[0]
	dr := msg.Settings.DataRate.GetLora()
	a.So(dr, should.NotBeNil)
	a.So(dr.SpreadingFactor, should.Equal, 7)
	a.So(dr.Bandwidth, should.Equal, 125000)
	a.So(msg.Settings.DataRate.GetLora().CodingRate, should.Equal, band.Cr4_5)
	a.So(msg.Settings.Frequency, should.Equal, 868100000)
	a.So(msg.RxMetadata[0].Timestamp, should.Equal, 368384825)
	a.So(len(msg.RawPayload), should.Equal, base64.StdEncoding.DecodedLen(len("Wqish6GVYpKy6o9WFHingeTJ1oh+ABc8iALBvwz44yxZP+BKDocaC5VQT5Y6dDdUaBILVjRMz0Ynzow1U/Kkts9AoZh3Ja3DX+DyY27exB+BKpSx2rXJ2vs9svm/EKYIsPF0RG1E+7lBYaD9"))) //nolint:lll
}

func TestToGatewayUpRawLRFHSS(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	raw := []byte(`{"rxpk":[{"tmst":368384825,"chan":0,"rfch":0,"freq":868.100000,"stat":1,"modu":"LR-FHSS","datr":"M0CW125","codr":"4/6","hpw":52,"rssi":-107,"size":108,"data":"Wqish6GVYpKy6o9WFHingeTJ1oh+ABc8iALBvwz44yxZP+BKDocaC5VQT5Y6dDdUaBILVjRMz0Ynzow1U/Kkts9AoZh3Ja3DX+DyY27exB+BKpSx2rXJ2vs9svm/EKYIsPF0RG1E+7lBYaD9"}]}`) //nolint:lll
	var rxData udp.Data
	err := json.Unmarshal(raw, &rxData)
	a.So(err, should.BeNil)

	upstream, err := udp.ToGatewayUp(rxData, udp.UpstreamMetadata{ID: ids})
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	a.So(len(upstream.UplinkMessages), should.Equal, 1)
	msg := upstream.UplinkMessages[0]
	dr := msg.Settings.DataRate.GetLrfhss()
	a.So(dr, should.NotBeNil)
	a.So(dr.ModulationType, should.Equal, 0)
	a.So(dr.OperatingChannelWidth, should.Equal, 125000)
	a.So(dr.CodingRate, should.Equal, "2/3")
	a.So(msg.Settings.Frequency, should.Equal, 868100000)
	a.So(msg.RxMetadata[0].Timestamp, should.Equal, 368384825)
	a.So(msg.RxMetadata[0].HoppingWidth, should.Equal, 52)
	a.So(len(msg.RawPayload), should.Equal, base64.StdEncoding.DecodedLen(len("Wqish6GVYpKy6o9WFHingeTJ1oh+ABc8iALBvwz44yxZP+BKDocaC5VQT5Y6dDdUaBILVjRMz0Ynzow1U/Kkts9AoZh3Ja3DX+DyY27exB+BKpSx2rXJ2vs9svm/EKYIsPF0RG1E+7lBYaD9"))) //nolint:lll
}

func TestToGatewayUpRawMultiAntenna(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	rx := []byte(`{
		"rxpk": [{
			"tmst": 879148780,
			"time": "2017-07-04T13:51:17.997099Z",
			"rfch": 0,
			"freq": 868.500000,
			"stat": 1,
			"modu": "LORA",
			"datr": "SF7BW125",
			"codr": "4/5",
			"size": 24,
			"data": "gM+AMQcAvgQBlohnlJqUGOJKTDuTscQD",
			"aesk": 42,
			"rsig": [{
				"ant": 0,
				"chan": 7,
				"lsnr": 14.0,
				"etime": "42QMzOlYSSPMMeqVPrY0fQ==",
				"rssis": -92,
				"rssic": -95,
				"rssisd": 0,
				"ftime": 1255738435,
				"foff": -8898,
				"ft2d": -251,
				"rfbsb": 100,
				"rs2s1": 97
			}, {
				"ant": 1,
				"chan": 23,
				"lsnr": 14.0,
				"etime": "djGiSzOC+gCT7vRPv7+Asw==",
				"rssis": -88,
				"rssic": -93,
				"rssisd": 0,
				"ftime": 1252538436,
				"foff": -8898,
				"ft2d": -187,
				"rfbsb": 100,
				"rs2s1": 104
			}]
		}]
	}`)
	var rxData udp.Data
	err := json.Unmarshal(rx, &rxData)
	a.So(err, should.BeNil)

	utcTime := time.Date(2017, 7, 4, 13, 51, 17, 997099000, time.UTC)
	const timestamp = 879148780

	up, err := udp.ToGatewayUp(rxData, udp.UpstreamMetadata{ID: ids})
	a.So(err, should.BeNil)
	a.So(up, should.Resemble, &ttnpb.GatewayUp{
		UplinkMessages: []*ttnpb.UplinkMessage{
			{
				RawPayload: []byte{0x80, 0xcf, 0x80, 0x31, 0x07, 0x00, 0xbe, 0x04, 0x01, 0x96, 0x88, 0x67, 0x94, 0x9a, 0x94, 0x18, 0xe2, 0x4a, 0x4c, 0x3b, 0x93, 0xb1, 0xc4, 0x03}, //nolint:lll
				Settings: &ttnpb.TxSettings{
					DataRate: &ttnpb.DataRate{
						Modulation: &ttnpb.DataRate_Lora{
							Lora: &ttnpb.LoRaDataRate{
								SpreadingFactor: 7,
								Bandwidth:       125000,
								CodingRate:      "4/5",
							},
						},
					},
					Frequency: 868500000,
					Time:      ttnpb.ProtoTimePtr(utcTime),
					Timestamp: timestamp,
				},
				RxMetadata: []*ttnpb.RxMetadata{
					{
						GatewayIds: &ttnpb.GatewayIdentifiers{
							GatewayId: "test-gateway",
						},
						AntennaIndex:                0,
						ChannelIndex:                7,
						Time:                        ttnpb.ProtoTimePtr(utcTime),
						Timestamp:                   timestamp,
						FineTimestamp:               1255738435,
						EncryptedFineTimestamp:      []byte{0xe3, 0x64, 0x0c, 0xcc, 0xe9, 0x58, 0x49, 0x23, 0xcc, 0x31, 0xea, 0x95, 0x3e, 0xb6, 0x34, 0x7d}, //nolint:lll
						EncryptedFineTimestampKeyId: "42",
						Rssi:                        -95,
						SignalRssi:                  &pbtypes.FloatValue{Value: -92},
						ChannelRssi:                 -95,
						RssiStandardDeviation:       0,
						Snr:                         14.0,
						FrequencyOffset:             -8898,
					},
					{
						GatewayIds: &ttnpb.GatewayIdentifiers{
							GatewayId: "test-gateway",
						},
						AntennaIndex:                1,
						ChannelIndex:                23,
						Time:                        ttnpb.ProtoTimePtr(utcTime),
						Timestamp:                   timestamp,
						FineTimestamp:               1252538436,
						EncryptedFineTimestamp:      []byte{0x76, 0x31, 0xa2, 0x4b, 0x33, 0x82, 0xfa, 0x00, 0x93, 0xee, 0xf4, 0x4f, 0xbf, 0xbf, 0x80, 0xb3}, //nolint:lll
						EncryptedFineTimestampKeyId: "42",
						Rssi:                        -93,
						SignalRssi:                  &pbtypes.FloatValue{Value: -88},
						ChannelRssi:                 -93,
						RssiStandardDeviation:       0,
						Snr:                         14.0,
						FrequencyOffset:             -8898,
					},
				},
			},
		},
	})
}

func TestFromDownlinkMessageLoRa(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	msg := &ttnpb.DownlinkMessage{
		Settings: &ttnpb.DownlinkMessage_Scheduled{
			Scheduled: &ttnpb.TxSettings{
				Frequency: 925700000,
				DataRate: &ttnpb.DataRate{
					Modulation: &ttnpb.DataRate_Lora{
						Lora: &ttnpb.LoRaDataRate{
							SpreadingFactor: 10,
							Bandwidth:       500000,
						},
					},
				},
				Downlink: &ttnpb.TxSettings_Downlink{
					TxPower:            20,
					InvertPolarization: true,
				},
				Timestamp: 1886440700,
			},
		},
		RawPayload: []byte{0x7d, 0xf3, 0x8e},
	}
	tx, err := udp.FromDownlinkMessage(msg)
	a.So(err, should.BeNil)
	a.So(tx.DatR, should.Resemble, datarate.DR{
		DataRate: &ttnpb.DataRate{
			Modulation: &ttnpb.DataRate_Lora{
				Lora: &ttnpb.LoRaDataRate{
					Bandwidth:       500000,
					SpreadingFactor: 10,
				},
			},
		},
	})
	a.So(tx.Tmst, should.Equal, 1886440700)
	a.So(tx.NCRC, should.Equal, true)
	a.So(tx.Data, should.Equal, "ffOO")
}

func TestFromDownlinkMessageFSK(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	msg := &ttnpb.DownlinkMessage{
		Settings: &ttnpb.DownlinkMessage_Scheduled{
			Scheduled: &ttnpb.TxSettings{
				Frequency: 925700000,
				DataRate: &ttnpb.DataRate{
					Modulation: &ttnpb.DataRate_Fsk{
						Fsk: &ttnpb.FSKDataRate{
							BitRate: 50000,
						},
					},
				},
				Downlink: &ttnpb.TxSettings_Downlink{
					TxPower: 20,
				},
				Timestamp: 1886440700,
			},
		},
		RawPayload: []byte{0x7d, 0xf3, 0x8e},
	}
	tx, err := udp.FromDownlinkMessage(msg)
	a.So(err, should.BeNil)
	a.So(tx.DatR, should.Resemble, datarate.DR{
		DataRate: &ttnpb.DataRate{
			Modulation: &ttnpb.DataRate_Fsk{
				Fsk: &ttnpb.FSKDataRate{
					BitRate: 50000,
				},
			},
		},
	})
	a.So(tx.Tmst, should.Equal, 1886440700)
	a.So(tx.FDev, should.Equal, 25000)
	a.So(tx.Data, should.Equal, "ffOO")
}

func TestDownlinkRoundtrip(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)
	expected := &ttnpb.DownlinkMessage{
		Settings: &ttnpb.DownlinkMessage_Scheduled{
			Scheduled: &ttnpb.TxSettings{
				Frequency: 925700000,
				DataRate: &ttnpb.DataRate{
					Modulation: &ttnpb.DataRate_Lora{
						Lora: &ttnpb.LoRaDataRate{
							SpreadingFactor: 10,
							Bandwidth:       500000,
						},
					},
				},
				Downlink: &ttnpb.TxSettings_Downlink{
					TxPower:            16.15,
					InvertPolarization: true,
				},
				Timestamp: 188700000,
			},
		},
		RawPayload: []byte{0x7d, 0xf3, 0x8e},
	}
	tx, err := udp.FromDownlinkMessage(expected)
	a.So(err, should.BeNil)

	actual, err := udp.ToDownlinkMessage(tx)
	a.So(err, should.BeNil)

	a.So(actual, should.HaveEmptyDiff, expected)
}
