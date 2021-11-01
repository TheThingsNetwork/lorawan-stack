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

package lbslns

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/ws"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestFromDownlinkMessage(t *testing.T) {
	_, ctx := test.New(t)
	ctx = ws.NewContextWithSession(ctx, &ws.Session{})
	updateSessionID(ctx, 0x11)
	var lbsLNS lbsLNS
	for _, tc := range []struct {
		BandID,
		Name string
		DownlinkMessage         ttnpb.DownlinkMessage
		ExpectedDownlinkMessage DownlinkMessage
	}{
		{
			BandID: band.EU_863_870,
			Name:   "SampleDownlink",
			DownlinkMessage: ttnpb.DownlinkMessage{
				RawPayload: []byte("Ymxhamthc25kJ3M=="),
				EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
					DeviceId: "testdevice",
				},
				Settings: &ttnpb.DownlinkMessage_Scheduled{
					Scheduled: &ttnpb.TxSettings{
						DataRate: ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 10,
									Bandwidth:       125000,
								},
							},
						},
						Frequency: 868500000,
						Downlink: &ttnpb.TxSettings_Downlink{
							AntennaIndex: 2,
						},
						Timestamp: 1553300787,
					},
				},
				CorrelationIds: []string{"correlation1"},
			},
			ExpectedDownlinkMessage: DownlinkMessage{
				DevEUI:      "00-00-00-00-00-00-00-01",
				DeviceClass: 0,
				Diid:        1,
				Pdu:         "596d7868616d74686332356b4a334d3d3d",
				RxDelay:     1,
				Rx1DR:       2,
				Rx1Freq:     868500000,
				RCtx:        2,
				Priority:    25,
				MuxTime:     1554300787.123456,
			},
		},
		{
			BandID: band.EU_863_870,
			Name:   "WithAbsoluteTime",
			DownlinkMessage: ttnpb.DownlinkMessage{
				RawPayload: []byte("Ymxhamthc25kJ3M=="),
				EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
					DeviceId: "testdevice",
				},
				Settings: &ttnpb.DownlinkMessage_Scheduled{
					Scheduled: &ttnpb.TxSettings{
						DataRate: ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 10,
									Bandwidth:       125000,
								},
							},
						},
						Frequency: 869525000,
						Downlink: &ttnpb.TxSettings_Downlink{
							AntennaIndex: 2,
						},
					},
				},
				CorrelationIds: []string{"correlation2"},
			},
			ExpectedDownlinkMessage: DownlinkMessage{
				DevEUI:      "00-00-00-00-00-00-00-01",
				DeviceClass: 0,
				Diid:        2,
				Pdu:         "596d7868616d74686332356b4a334d3d3d",
				RxDelay:     1,
				Rx1DR:       2,
				Rx1Freq:     869525000,
				RCtx:        2,
				Priority:    25,
				MuxTime:     1554300787.123456,
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			raw, err := lbsLNS.FromDownlink(ctx, tc.DownlinkMessage, tc.BandID, 1554300787, time.Unix(1554300787, 123456000))
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			var dnmsg DownlinkMessage
			err = dnmsg.unmarshalJSON(raw)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			dnmsg.XTime = tc.ExpectedDownlinkMessage.XTime
			if !a.So(dnmsg, should.Resemble, tc.ExpectedDownlinkMessage) {
				t.Fatalf("Invalid DownlinkMessage: %v", dnmsg)
			}
		})
	}
}

func TestToDownlinkMessage(t *testing.T) {
	for _, tc := range []struct {
		BandID,
		Name string
		DownlinkMessage         DownlinkMessage
		ExpectedDownlinkMessage *ttnpb.DownlinkMessage
	}{
		{
			BandID: band.EU_863_870,
			Name:   "SampleDownlink",
			DownlinkMessage: DownlinkMessage{
				DeviceClass: 0,
				Pdu:         "Ymxhamthc25kJ3M==",
				RxDelay:     1,
				Rx1DR:       2,
				Rx1Freq:     868500000,
				RCtx:        2,
				Priority:    25,
				XTime:       1554300785,
			},
			ExpectedDownlinkMessage: &ttnpb.DownlinkMessage{
				RawPayload: []byte("Ymxhamthc25kJ3M=="),
				Settings: &ttnpb.DownlinkMessage_Scheduled{
					Scheduled: &ttnpb.TxSettings{
						DataRate: ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 10,
									Bandwidth:       125000,
								},
							},
						},
						Frequency: 868500000,
						Downlink: &ttnpb.TxSettings_Downlink{
							AntennaIndex: 2,
						},
						Timestamp: 1554300785,
					},
				},
			},
		},
		{
			BandID: band.EU_863_870,
			Name:   "WithAbsoluteTime",
			DownlinkMessage: DownlinkMessage{
				DeviceClass: 1,
				Pdu:         "Ymxhamthc25kJ3M==",
				RxDelay:     1,
				Rx1DR:       2,
				Rx1Freq:     869525000,
				RCtx:        2,
				Priority:    25,
			},
			ExpectedDownlinkMessage: &ttnpb.DownlinkMessage{
				RawPayload: []byte("Ymxhamthc25kJ3M=="),
				Settings: &ttnpb.DownlinkMessage_Scheduled{
					Scheduled: &ttnpb.TxSettings{
						DataRate: ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 10,
									Bandwidth:       125000,
								},
							},
						},
						Frequency: 869525000,
						Downlink: &ttnpb.TxSettings_Downlink{
							AntennaIndex: 2,
						},
					},
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			dlMesg, err := tc.DownlinkMessage.ToDownlinkMessage(tc.BandID)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			if !a.So(dlMesg, should.Resemble, tc.ExpectedDownlinkMessage) {
				t.Fatalf("Invalid DownlinkMessage: %v", dlMesg)
			}
		})
	}
}

func TestTransferTime(t *testing.T) {
	a, ctx := test.New(t)

	session := &ws.Session{}
	ctx = ws.NewContextWithSession(ctx, session)

	gtw := &ttnpb.Gateway{
		Ids: &ttnpb.GatewayIdentifiers{
			GatewayId: "eui-1122334455667788",
			Eui:       &types.EUI64{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88},
		},
		FrequencyPlanId: test.EUFrequencyPlanID,
	}

	conn, err := io.NewConnection(ctx, nil, gtw, test.FrequencyPlanStore, true, nil)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	f := (*lbsLNS)(nil)

	// No time sync available.
	b, err := f.TransferTime(ctx, time.Now(), conn)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(b, should.BeNil)

	// Add fictional concentrator sync.
	xTime := 123456
	timeAtSync := time.Now()
	conn.SyncWithGatewayConcentrator(
		uint32(xTime&0xFFFFFFFF),
		timeAtSync,
		scheduling.ConcentratorTime(time.Duration(xTime&0xFFFFFFFFFF)*time.Microsecond),
	)

	// No session ID available.
	b, err = f.TransferTime(ctx, time.Now(), conn)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(b, should.BeNil)

	// Add fictional session ID.
	updateSessionID(ctx, 0x42)

	// Attempt to transfer time.
	timeNow := time.Now()
	b, err = f.TransferTime(ctx, timeNow, conn)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	if a.So(b, should.NotBeNil) {
		var res TimeSyncResponse
		if err := json.Unmarshal(b, &res); !a.So(err, should.BeNil) {
			t.FailNow()
		}
		a.So(res.TxTime, should.Equal, 0.0)
		a.So(res.XTime>>48, should.Equal, 0x42)
		a.So(res.XTime&0xFFFFFFFFFF, should.Equal, int64(xTime)+timeNow.Sub(timeAtSync).Microseconds())
		a.So(res.GPSTime, should.Equal, TimeToGPSTime(timeNow))
		a.So(res.MuxTime, should.Equal, TimeToUnixSeconds(timeNow))
	}
}
