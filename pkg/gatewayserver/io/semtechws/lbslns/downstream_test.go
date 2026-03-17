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

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/band"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/semtechws"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/scheduling"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestFromDownlinkMessage(t *testing.T) {
	_, ctx := test.New(t)
	ctx = semtechws.NewContextWithSession(ctx, &semtechws.Session{})
	semtechws.UpdateSessionID(ctx, 0x11)
	var lbsLNS lbsLNS
	downlinkTokens := &io.DownlinkTokens{}
	for _, tc := range []struct {
		BandID,
		Name string
		DownlinkMessage         *ttnpb.DownlinkMessage
		ExpectedDownlinkMessage DownlinkMessage
	}{
		{
			BandID: band.EU_863_870,
			Name:   "SampleDownlink",
			DownlinkMessage: &ttnpb.DownlinkMessage{
				RawPayload: []byte("Ymxhamthc25kJ3M=="),
				EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
					DeviceId: "testdevice",
				},
				Settings: &ttnpb.DownlinkMessage_Scheduled{
					Scheduled: &ttnpb.TxSettings{
						DataRate: &ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 10,
									Bandwidth:       125000,
									CodingRate:      band.Cr4_5,
								},
							},
						},
						Frequency: 868500000,
						Downlink: &ttnpb.TxSettings_Downlink{
							AntennaIndex: 2,
						},
						ConcentratorTimestamp: 1553300787,
					},
				},
				CorrelationIds: []string{"correlation1"},
			},
			ExpectedDownlinkMessage: DownlinkMessage{
				DevEUI:      "00-00-00-00-00-00-00-01",
				DeviceClass: 0,
				Diid:        1,
				Pdu:         "596d7868616d74686332356b4a334d3d3d",
				RCtx:        2,
				Priority:    25,
				MuxTime:     1554300787.123456,
				TimestampDownlinkMessage: &TimestampDownlinkMessage{
					RxDelay: 1,
					Rx1DR:   2,
					Rx1Freq: 868500000,
					XTime:   semtechws.ConcentratorTimeToXTime(0x11, 1553300787) - int64(time.Second/time.Microsecond),
				},
			},
		},
		{
			BandID: band.EU_863_870,
			Name:   "WithAbsoluteTime",
			DownlinkMessage: &ttnpb.DownlinkMessage{
				RawPayload: []byte("Ymxhamthc25kJ3M=="),
				EndDeviceIds: &ttnpb.EndDeviceIdentifiers{
					DeviceId: "testdevice",
				},
				Settings: &ttnpb.DownlinkMessage_Scheduled{
					Scheduled: &ttnpb.TxSettings{
						DataRate: &ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 10,
									Bandwidth:       125000,
									CodingRate:      band.Cr4_5,
								},
							},
						},
						Frequency: 869525000,
						Downlink: &ttnpb.TxSettings_Downlink{
							AntennaIndex: 2,
						},
						Time: timestamppb.New(time.Unix(0x42424242, 0x42424242)),
					},
				},
				CorrelationIds: []string{"correlation2"},
			},
			ExpectedDownlinkMessage: DownlinkMessage{
				DevEUI:      "00-00-00-00-00-00-00-01",
				DeviceClass: 1,
				Diid:        2,
				Pdu:         "596d7868616d74686332356b4a334d3d3d",
				RCtx:        2,
				Priority:    25,
				MuxTime:     1554300787.123456,
				AbsoluteTimeDownlinkMessage: &AbsoluteTimeDownlinkMessage{
					DR:      2,
					Freq:    869525000,
					GPSTime: semtechws.TimeToGPSTime(time.Unix(0x42424242, 0x42424242)),
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)
			raw, err := lbsLNS.FromDownlink(
				ctx,
				tc.DownlinkMessage,
				tc.BandID,
				time.Unix(1554300787, 123456000),
				downlinkTokens,
			)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			var dnmsg DownlinkMessage
			err = dnmsg.unmarshalJSON(raw)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
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
				RCtx:        2,
				Priority:    25,
				TimestampDownlinkMessage: &TimestampDownlinkMessage{
					RxDelay: 1,
					Rx1DR:   2,
					Rx1Freq: 868500000,
					XTime:   1554300785,
				},
			},
			ExpectedDownlinkMessage: &ttnpb.DownlinkMessage{
				RawPayload: []byte("Ymxhamthc25kJ3M=="),
				Settings: &ttnpb.DownlinkMessage_Scheduled{
					Scheduled: &ttnpb.TxSettings{
						DataRate: &ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 10,
									Bandwidth:       125000,
									CodingRate:      band.Cr4_5,
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
				RCtx:        2,
				Priority:    25,
				AbsoluteTimeDownlinkMessage: &AbsoluteTimeDownlinkMessage{
					DR:      2,
					Freq:    869525000,
					GPSTime: semtechws.TimeToGPSTime(time.Unix(0x42424242, 0x42424242)),
				},
			},
			ExpectedDownlinkMessage: &ttnpb.DownlinkMessage{
				RawPayload: []byte("Ymxhamthc25kJ3M=="),
				Settings: &ttnpb.DownlinkMessage_Scheduled{
					Scheduled: &ttnpb.TxSettings{
						DataRate: &ttnpb.DataRate{
							Modulation: &ttnpb.DataRate_Lora{
								Lora: &ttnpb.LoRaDataRate{
									SpreadingFactor: 10,
									Bandwidth:       125000,
									CodingRate:      band.Cr4_5,
								},
							},
						},
						Frequency: 869525000,
						Downlink: &ttnpb.TxSettings_Downlink{
							AntennaIndex: 2,
						},
						Time: timestamppb.New(time.Unix(0x42424242, 0x42424242).Truncate(time.Microsecond)),
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

	ctx = semtechws.NewContextWithSession(ctx, &semtechws.Session{})

	f := (*lbsLNS)(nil)
	now := time.Unix(123, 456)

	// No timesync settings available in the session.
	b, err := f.TransferTime(ctx, now, nil, nil)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(b, should.BeNil)

	// Enable timesync for the session.
	semtechws.UpdateSessionTimeSync(ctx, true)

	// No GPSTime / ConcentratorTime - expect only MuxTime.
	b, err = f.TransferTime(ctx, now, nil, nil)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	if a.So(b, should.NotBeNil) {
		var res TimeSyncResponse
		if err := json.Unmarshal(b, &res); !a.So(err, should.BeNil) {
			t.FailNow()
		}
		a.So(res.TxTime, should.Equal, 0.0)
		a.So(res.XTime, should.Equal, 0)
		a.So(res.GPSTime, should.Equal, 0)
		a.So(res.MuxTime, should.Equal, semtechws.TimeToUnixSeconds(now))
	}

	// Add fictional session ID.
	semtechws.UpdateSessionID(ctx, 0x42)

	gpsTime := time.Unix(456, 678)
	concentratorTime := scheduling.ConcentratorTime(890 * time.Microsecond)

	// Attempt to transfer time.
	b, err = f.TransferTime(ctx, now, &gpsTime, &concentratorTime)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	if a.So(b, should.NotBeNil) {
		var res TimeSyncResponse
		if err := json.Unmarshal(b, &res); !a.So(err, should.BeNil) {
			t.FailNow()
		}
		a.So(res.TxTime, should.Equal, 0.0)
		a.So(semtechws.SessionIDFromXTime(res.XTime), should.Equal, 0x42)
		a.So(semtechws.ConcentratorTimeFromXTime(res.XTime), should.Equal, 890*time.Microsecond)
		a.So(res.GPSTime, should.Equal, semtechws.TimeToGPSTime(gpsTime))
		a.So(res.MuxTime, should.Equal, semtechws.TimeToUnixSeconds(now))
	}
}

func TestTimeSyncDoesNotDisableTransferTime(t *testing.T) {
	a, ctx := test.New(t)
	ctx = semtechws.NewContextWithSession(ctx, &semtechws.Session{})

	f := (*lbsLNS)(nil)
	now := time.Unix(500, 0)

	// Enable timesync for the session (as version.go does on connect).
	semtechws.UpdateSessionTimeSync(ctx, true)
	semtechws.UpdateSessionID(ctx, 0x42)

	// Process a timesync request (as upstream.go HandleUp does).
	// Previously this called UpdateSessionTimeSync(ctx, false), which
	// disabled LNS-initiated time transfers. After the fix, time sync
	// should remain enabled.
	req := TimeSyncRequest{TxTime: 123.456}
	resp := req.Response(now)
	a.So(resp.TxTime, should.Equal, 123.456)

	// Verify time sync is still enabled.
	enabled, ok := semtechws.GetSessionTimeSync(ctx)
	if !a.So(ok, should.BeTrue) {
		t.FailNow()
	}
	a.So(enabled, should.BeTrue)

	// Verify TransferTime still produces output (not suppressed).
	b, err := f.TransferTime(ctx, now, nil, nil)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	if !a.So(b, should.NotBeNil) {
		t.Fatal("TransferTime returned nil after timesync request; time transfers were incorrectly disabled")
	}

	var res TimeSyncResponse
	if err := json.Unmarshal(b, &res); !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(res.MuxTime, should.Equal, semtechws.TimeToUnixSeconds(now))
}

func TestTransferTimeWithUplinkXTime(t *testing.T) {
	a, ctx := test.New(t)
	ctx = semtechws.NewContextWithSession(ctx, &semtechws.Session{})

	f := (*lbsLNS)(nil)
	now := time.Unix(600, 0)

	semtechws.UpdateSessionTimeSync(ctx, true)
	semtechws.UpdateSessionID(ctx, 0x42)

	// Store uplink timing from a GPS-equipped gateway.
	xtime := int64(12666373963464220)
	gpstime := int64(1232095200000000)
	semtechws.UpdateLastUplink(ctx, xtime, gpstime, now)

	// Retrieve and use in time transfer (as ws.go ticker does).
	gotXTime, gotGPSTime, rxAt, ok := semtechws.GetLastUplink(ctx)
	if !a.So(ok, should.BeTrue) {
		t.FailNow()
	}
	a.So(gotXTime, should.Equal, xtime)
	a.So(gotGPSTime, should.Equal, gpstime)
	a.So(rxAt, should.Equal, now)

	// Build TransferTime args from uplink data (simulating the ws.go ticker logic).
	ct := semtechws.ConcentratorTimeFromXTime(gotXTime)
	gt := semtechws.TimeFromGPSTime(gotGPSTime)

	b, err := f.TransferTime(ctx, now, &gt, &ct)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	if a.So(b, should.NotBeNil) {
		var res TimeSyncResponse
		if err := json.Unmarshal(b, &res); !a.So(err, should.BeNil) {
			t.FailNow()
		}
		a.So(res.GPSTime, should.Equal, semtechws.TimeToGPSTime(gt))
		a.So(res.XTime, should.NotEqual, 0)
		a.So(res.MuxTime, should.Equal, semtechws.TimeToUnixSeconds(now))
	}

	// Non-GPS gateway: gpstime=0, only MuxTime should be present.
	// TransferTime requires both gpsTime and concentratorTime to produce
	// xtime+gpstime; with nil gpsTime it falls back to MuxTime only.
	semtechws.UpdateLastUplink(ctx, xtime, 0, now)
	_, gotGPSTime, _, ok = semtechws.GetLastUplink(ctx)
	if !a.So(ok, should.BeTrue) {
		t.FailNow()
	}
	a.So(gotGPSTime, should.Equal, 0)

	b, err = f.TransferTime(ctx, now, nil, &ct)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	if a.So(b, should.NotBeNil) {
		var res TimeSyncResponse
		if err := json.Unmarshal(b, &res); !a.So(err, should.BeNil) {
			t.FailNow()
		}
		a.So(res.GPSTime, should.Equal, 0)
		a.So(res.XTime, should.Equal, 0)
		a.So(res.MuxTime, should.Equal, semtechws.TimeToUnixSeconds(now))
	}
}
